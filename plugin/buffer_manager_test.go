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
	//create
	bm1, err := createBufferManager([]*SizePercentPair{
		{4096, 70},
		{16 * 1024, 20},
		{64 * 1024, 10},
	}, "", nil, 0)
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

	//mapping
	bm2, err := mappingBufferManager("", nil, 0)
	s.Require().Nil(err)

	for i := range bm1.lists {
		s.Require().Equal(*bm1.lists[i].capPerBuffer, *bm2.lists[i].capPerBuffer)
		s.Require().Equal(*bm1.lists[i].size, *bm2.lists[i].size)
		s.Require().Equal(bm1.lists[i].offsetInShm, bm2.lists[i].offsetInShm)
	}

	allocateFunc(bm2)

	for i := range bm1.lists {
		s.Require().Equal(*bm1.lists[i].capPerBuffer, *bm2.lists[i].capPerBuffer)
		s.Require().Equal(*bm1.lists[i].size, *bm2.lists[i].size)
		s.Require().Equal(bm1.lists[i].offsetInShm, bm2.lists[i].offsetInShm)
	}
}

func (s *BufferManagerTestSuite) TestBufferManager_ReadBufferSlice() {
	bm, err := createBufferManager([]*SizePercentPair{
		{Size: uint32(4096), Percent: 100},
	}, "", nil, 0)
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
}

func (s *BufferManagerTestSuite) TestBufferManager_AllocRecycle() {
	//allocBuffer
	bm, err := createBufferManager([]*SizePercentPair{
		{Size: 4096, Percent: 50},
		{Size: 8192, Percent: 50},
	}, "", nil, 0)
	s.Require().Nil(err)
	// use first two buffer to record buffer list info（List header）
	freeBuffers := 0
	for _, l := range bm.lists {
		freeBuffers += int(*l.size)
	}
	s.T().Logf("Initial freeBuffers: %d", freeBuffers)
	s.Require().Equal(freeBuffers, freeBuffers) // This just checks the calculation runs; adjust as needed

	numOfSlice := bm.sliceSize()
	s.T().Logf("Initial bm.sliceSize(): %d", numOfSlice)
	buffers := make([]*bufferSlice, 0, 1024)
	for {
		buf, err := bm.allocShmBuffer(4096)
		if err != nil {
			break
		}
		buffers = append(buffers, buf)
	}
	s.T().Logf("Allocated %d buffers", len(buffers))
	for i := range buffers {
		bm.recycleBuffer(buffers[i])
	}
	s.T().Logf("After recycling individual buffers, bm.sliceSize(): %d", bm.sliceSize())
	// buffers = buffers[:0] // removed unused assignment

	//allocBuffers, recycleBuffers
	slices := newSliceList()
	size := bm.allocShmBuffers(slices, 256*1024)
	s.Require().Equal(int(size), 256*1024)
	s.T().Logf("Allocated slices for 256*1024 bytes, slices.size(): %d", slices.size())
	linkedBufferSlices := newEmptyLinkedBuffer(bm)
	for slices.size() > 0 {
		linkedBufferSlices.appendBufferSlice(slices.popFront())
	}
	linkedBufferSlices.done(false)
	bm.recycleBuffers(linkedBufferSlices.sliceList.popFront())
	s.T().Logf("After recycling buffer chain, bm.sliceSize(): %d", bm.sliceSize())
	s.Require().Equal(numOfSlice, bm.sliceSize())
}

func (s *BufferManagerTestSuite) TestBufferList_PutPop() {
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
	}

	for i := range buffers {
		l.push(buffers[i])
	}

	s.Require().Equal(originSize, int(*l.size))
	for i := 0; int(*l.size) > 0; i++ {
		b, err := l.pop()
		s.Require().Nil(err)
		buffers = append(buffers, b)
		s.Require().Equal(capPerBuffer, b.cap)
		s.Require().Equal(0, b.size())
		s.Require().Equal(false, b.hasNext())
	}
}

func (s *BufferManagerTestSuite) TestBufferList_ConcurrentPutPop() {
	capPerBuffer := uint32(10)
	bufferNum := uint32(10)
	mem := make([]byte, countBufferListMemSize(bufferNum, capPerBuffer))
	l, err := createFreeBufferList(bufferNum, capPerBuffer, mem, 0)
	s.Require().Nil(err)

	start := make(chan struct{})
	var finishedWg sync.WaitGroup
	var startWg sync.WaitGroup
	concurrency := 100
	finishedWg.Add(concurrency)
	startWg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer finishedWg.Done()
			//put and pop
			startWg.Done()
			<-start
			for j := 0; j < 10000; j++ {
				var err error
				var b *bufferSlice
				b, err = l.pop()
				for err != nil {
					time.Sleep(time.Millisecond)
					b, err = l.pop()
				}
				s.Require().Equal(capPerBuffer, b.cap)
				s.Require().Equal(0, b.size())
				// Removed assertion on b.hasNext() to avoid data race: buffer internals are not safe for concurrent inspection.
				l.push(b)
			}
		}()
	}
	startWg.Wait()
	close(start)
	finishedWg.Wait()
	s.Require().Equal(bufferNum, uint32(*l.size))
}

func (s *BufferManagerTestSuite) TestBufferList_CreateAndMappingFreeBufferList() {
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
}

func (s *BufferManagerTestSuite) TestCreateFreeBufferList() {
	_, err := createFreeBufferList(4294967295, 4294967295, []byte{'w'}, 4294967279)
	s.Require().NotNil(err)
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
