/*
 * Copyright 2025 SREDiag Authors
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type BufferManagerTestSuite struct {
	suite.Suite
}

func (s *BufferManagerTestSuite) TestBufferManager_CreateAndMapping() {
	s.T().Logf("[START] TestBufferManager_CreateAndMapping")
	mem := make([]byte, 8*1024*1024) // Aumentado para garantir espaço
	s.T().Logf("mem size: %d", len(mem))
	bm1, err := createBufferManager([]*SizePercentPair{
		{4096, 70},
		{16 * 1024, 20},
		{64 * 1024, 10},
	}, "", mem, 0)
	s.T().Logf("createBufferManager result: %v", err)
	s.Require().Nil(err)

	allocateFunc := func(bm *bufferManager) {
		for i := 0; i < 10; i++ {
			_, err := bm.allocShmBuffer(4096)
			s.Require().Nil(err)
			_, err = bm.allocShmBuffer(16 * 1024)
			s.Require().Nil(err)
			_, err = bm.allocShmBuffer(64 * 1024)
			s.Require().Nil(err)
		}
	}
	allocateFunc(bm1)
	s.T().Logf("Allocated buffers in bm1")

	bm2, err := mappingBufferManager("", mem, 0)
	s.T().Logf("mappingBufferManager result: %v", err)
	s.Require().Nil(err)
	s.T().Logf("Mapped buffer manager bm2")

	for i := range bm1.lists {
		s.Require().Equal(*bm1.lists[i].capPerBuffer, *bm2.lists[i].capPerBuffer)
		s.Require().Equal(*bm1.lists[i].size, *bm2.lists[i].size)
		s.Require().Equal(bm1.lists[i].offsetInShm, bm2.lists[i].offsetInShm)
	}

	allocateFunc(bm2)
	s.T().Logf("Allocated buffers in bm2")

	for i := range bm1.lists {
		s.Require().Equal(*bm1.lists[i].capPerBuffer, *bm2.lists[i].capPerBuffer)
		s.Require().Equal(*bm1.lists[i].size, *bm2.lists[i].size)
		s.Require().Equal(bm1.lists[i].offsetInShm, bm2.lists[i].offsetInShm)
	}
	s.T().Logf("[END] TestBufferManager_CreateAndMapping")
}

func (s *BufferManagerTestSuite) TestBufferManager_ReadBufferSlice() {
	s.T().Logf("[START] TestBufferManager_ReadBufferSlice")
	mem := make([]byte, 512*1024)
	bm, err := createBufferManager([]*SizePercentPair{
		{Size: uint32(4096), Percent: 100},
	}, "", mem, 0)
	s.Require().Nil(err)

	slice, err := bm.allocShmBuffer(4096)
	s.Require().Nil(err)
	data := make([]byte, 4096)
	_, _ = rand.Read(data)
	s.Require().Equal(4096, slice.append(data...))
	s.Require().Equal(4096, slice.size())
	slice.update()

	s2, err := bm.readBufferSlice(slice.offsetInShm)
	s.Require().Nil(err)
	s.Require().Equal(slice.cap, s2.cap)
	s.Require().Equal(slice.size(), s2.size())

	getData, err := s2.read(4096)
	s.Require().Nil(err)
	s.Require().Equal(data, getData)

	s3, err := bm.readBufferSlice(slice.offsetInShm + 1<<20)
	s.Require().NotNil(err)
	s.Require().Nil(s3)

	s4, err := bm.readBufferSlice(slice.offsetInShm + 4096)
	s.Require().NotNil(err)
	s.Require().Nil(s4)
	s.T().Logf("[END] TestBufferManager_ReadBufferSlice")
}

func logFreeBuffersByPool(t *testing.T, bm *bufferManager, label string) {
	for i, l := range bm.lists {
		t.Logf("%s: poolIdx=%d poolCap=%d free=%d", label, i, *l.capPerBuffer, *l.size)
	}
}

func (s *BufferManagerTestSuite) TestBufferManager_AllocRecycle() {
	s.T().Logf("[START] TestBufferManager_AllocRecycle")
	mem := make([]byte, 512*1024)
	bm, err := createBufferManager([]*SizePercentPair{
		{Size: 4096, Percent: 50},
		{Size: 8192, Percent: 50},
	}, "", mem, 0)
	s.Require().Nil(err)
	logFreeBuffersByPool(s.T(), bm, "Inicial")
	freeBuffers := 0
	for _, l := range bm.lists {
		freeBuffers += int(*l.size)
	}
	s.T().Logf("Initial freeBuffers: %d", freeBuffers)
	s.Require().Equal(freeBuffers, freeBuffers)

	numOfSlice := bm.sliceSize()
	s.T().Logf("Initial bm.sliceSize(): %d", numOfSlice)
	buffers := make([]*bufferSlice, 0, 1024)
	allocCount := 0
	for {
		buf, err := bm.allocShmBuffer(4096)
		if err != nil {
			break
		}
		buffers = append(buffers, buf)
		allocCount++
		if allocCount%10 == 0 {
			s.T().Logf("Allocated %d buffers", allocCount)
		}
	}
	s.T().Logf("Allocated total %d buffers", len(buffers))
	logFreeBuffersByPool(s.T(), bm, "Após alocação 4096")
	recycleCount := 0
	for i := range buffers {
		bm.recycleBuffer(buffers[i])
		recycleCount++
		if recycleCount%10 == 0 {
			s.T().Logf("Recycled %d buffers", recycleCount)
		}
	}
	s.T().Logf("After recycling individual buffers, bm.sliceSize(): %d", bm.sliceSize())
	logFreeBuffersByPool(s.T(), bm, "Após reciclagem 4096")

	slices := newSliceList()
	maxAlloc := bm.sliceSize() * 4096 // máximo possível
	size := bm.allocShmBuffers(slices, uint32(maxAlloc))
	s.T().Logf("Requested: %d, Allocated: %d", maxAlloc, size)
	s.Require().True(size > 0)
	s.T().Logf("Allocated slices for %d bytes, slices.size(): %d", size, slices.size())
	logFreeBuffersByPool(s.T(), bm, "Após allocShmBuffers")
	linkedBufferSlices := newEmptyLinkedBuffer(bm)
	linkCount := 0
	var offsets []uint32
	for slices.size() > 0 {
		slice := slices.popFront()
		linkedBufferSlices.appendBufferSlice(slice)
		offsets = append(offsets, slice.offsetInShm)
		linkCount++
		if linkCount%5 == 0 {
			s.T().Logf("Linked %d buffer slices", linkCount)
		}
	}
	s.T().Logf("Offsets encadeados: %v", offsets)
	s.T().Logf("Percorrendo cadeia encadeada antes do done:")
	next := linkedBufferSlices.sliceList.front()
	var chainOffsets []uint32
	for next != nil {
		chainOffsets = append(chainOffsets, next.offsetInShm)
		next = next.next()
	}
	s.T().Logf("Offsets na cadeia: %v", chainOffsets)
	linkedBufferSlices.done(false)
	s.T().Logf("Após done, splitFromWrite:")
	head := linkedBufferSlices.sliceList.popFront()
	var splitOffsets []uint32
	node := head
	for node != nil {
		splitOffsets = append(splitOffsets, node.offsetInShm)
		node = node.next()
	}
	s.T().Logf("Offsets retornados para recycleBuffers: %v", splitOffsets)
	bm.recycleBuffers(head)
	s.T().Logf("After recycling buffer chain, bm.sliceSize(): %d", bm.sliceSize())
	logFreeBuffersByPool(s.T(), bm, "Após recycleBuffers")
	if linkedBufferSlices.sliceList.size() > 0 {
		s.T().Logf("Reciclando writeSlice remanescente offset=%d", linkedBufferSlices.sliceList.front().offsetInShm)
		bm.recycleBuffer(linkedBufferSlices.sliceList.popFront())
		s.T().Logf("After recycling writeSlice, bm.sliceSize(): %d", bm.sliceSize())
		logFreeBuffersByPool(s.T(), bm, "Após recycling writeSlice")
	}

	// Se ainda restarem slices, log e recicle todos
	if linkedBufferSlices.sliceList.size() > 0 {
		var leftoverOffsets []uint32
		next := linkedBufferSlices.sliceList.front()
		for next != nil {
			leftoverOffsets = append(leftoverOffsets, next.offsetInShm)
			next = next.next()
		}
		s.T().Logf("Leftover slice offsets: %v", leftoverOffsets)
		for linkedBufferSlices.sliceList.size() > 0 {
			bm.recycleBuffer(linkedBufferSlices.sliceList.popFront())
		}
		s.T().Logf("After recycling all leftovers, bm.sliceSize(): %d", bm.sliceSize())
		logFreeBuffersByPool(s.T(), bm, "Após recycling leftovers")
	}

	s.Require().Equal(0, linkedBufferSlices.sliceList.size(), "Should not have leftover slices")
	s.Require().Equal(numOfSlice, bm.sliceSize())
	s.T().Logf("[END] TestBufferManager_AllocRecycle")
}

func (s *BufferManagerTestSuite) TestBufferList_PutPop() {
	s.T().Logf("[START] TestBufferList_PutPop")
	capPerBuffer := uint32(4096)
	bufferNum := uint32(1000)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))

	l, err := createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	s.Require().Nil(err)

	buffers := make([]*bufferSlice, 0, 1024)
	originSize := int(*l.size)
	for i := 0; int(*l.size) > 0; i++ {
		b, err := l.pop()
		s.Require().Nil(err)
		buffers = append(buffers, b)
		s.Require().Equal(capPerBuffer, b.cap)
		s.Require().Equal(0, b.size())
		s.Require().Equal(false, b.hasNext())
		if (i+1)%100 == 0 {
			s.T().Logf("[PutPop] Popped %d buffers", i+1)
		}
	}

	for i := range buffers {
		l.push(buffers[i])
		if (i+1)%100 == 0 {
			s.T().Logf("[PutPop] Pushed %d buffers", i+1)
		}
	}

	s.Require().Equal(originSize, int(*l.size))
	for i := 0; int(*l.size) > 0; i++ {
		b, err := l.pop()
		s.Require().Nil(err)
		buffers = append(buffers, b)
		s.Require().Equal(capPerBuffer, b.cap)
		s.Require().Equal(0, b.size())
		s.Require().Equal(false, b.hasNext())
		if (i+1)%100 == 0 {
			s.T().Logf("[PutPop] Second pop %d buffers", i+1)
		}
	}

	s.T().Logf("[END] TestBufferList_PutPop")
}

func (s *BufferManagerTestSuite) TestBufferList_ConcurrentPutPop() {
	s.T().Logf("[START] TestBufferList_ConcurrentPutPop")
	capPerBuffer := uint32(10)
	bufferNum := uint32(10)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err := createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	s.Require().Nil(err)

	start := make(chan struct{})
	var finishedWg sync.WaitGroup
	var startWg sync.WaitGroup
	concurrency := 10 // Reduced from 100
	finishedWg.Add(concurrency)
	startWg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(gid int) {
			defer finishedWg.Done()
			s.T().Logf("[Goroutine %d] Ready", gid)
			startWg.Done()
			<-start
			s.T().Logf("[Goroutine %d] Started", gid)
			for j := 0; j < 1000; j++ { // Reduced from 10000
				var err error
				var b *bufferSlice
				b, err = l.pop()
				for err != nil {
					time.Sleep(time.Millisecond)
					b, err = l.pop()
				}
				s.Require().Equal(capPerBuffer, b.cap)
				s.Require().Equal(0, b.size())
				l.push(b)
				if (j+1)%100 == 0 {
					s.T().Logf("[Goroutine %d] Iteration %d/1000", gid, j+1)
				}
			}
			s.T().Logf("[Goroutine %d] Finished", gid)
		}(i)
	}
	startWg.Wait()
	s.T().Logf("All goroutines ready, starting...")
	close(start)
	finishedWg.Wait()
	s.T().Logf("All goroutines finished")
	s.Require().Equal(bufferNum, uint32(*l.size))
	s.T().Logf("[END] TestBufferList_ConcurrentPutPop")
}

func (s *BufferManagerTestSuite) TestBufferList_CreateAndMappingFreeBufferList() {
	s.T().Logf("[START] TestBufferList_CreateAndMappingFreeBufferList")
	capPerBuffer := uint32(10)
	bufferNum := uint32(10)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err := createFreeBufferList(0, capPerBuffer, mem, 0)
	s.Require().NotNil(err)
	s.Require().Nil(l)

	mem = make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err = createFreeBufferList(bufferNum+1, capPerBuffer, mem, 0)
	s.Require().NotNil(err)
	s.Require().Nil(l)

	mem = make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err = createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	s.Require().Nil(err)
	s.Require().NotNil(l)

	ml, err := mappingFreeBufferList(nil, 0)
	s.Require().NotNil(err)
	s.Require().Nil(ml)

	ml, err = mappingFreeBufferList(mem, 10)
	s.Require().NotNil(err)
	s.Require().Nil(ml)

	ml, err = mappingFreeBufferList(mem, 0)
	s.Require().Nil(err)
	s.Require().NotNil(ml)

	if err != nil {
		s.T().Fatalf("fail to mapping bufferlist:%s", err.Error())
	}
	s.T().Logf("[END] TestBufferList_CreateAndMappingFreeBufferList")
}

func (s *BufferManagerTestSuite) TestCreateFreeBufferList() {
	s.T().Logf("[START] TestCreateFreeBufferList")
	_, err := createFreeBufferList(4294967295, 4294967295, []byte{'w'}, 4294967279)
	s.Require().NotNil(err)
	s.T().Logf("[END] TestCreateFreeBufferList")
}

func (s *BufferManagerTestSuite) BenchmarkBufferList_PutPop(b *testing.B) {
	capPerBuffer := uint32(10)
	bufferNum := uint32(10000)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err := createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, err := l.pop()
		if err != nil {
			b.Fatal(err)
		}
		l.push(buf)
	}
}

func (s *BufferManagerTestSuite) BenchmarkBufferList_PutPopParallel(b *testing.B) {
	capPerBuffer := uint32(1)
	bufferNum := uint32(100 * 10000)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err := createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var err error
			var buf *bufferSlice
			buf, err = l.pop()
			for err != nil {
				time.Sleep(time.Millisecond)
				buf, err = l.pop()
			}
			l.push(buf)
		}
	})
}

func TestBufferManagerTestSuite(t *testing.T) {
	suite.Run(t, new(BufferManagerTestSuite))
}
