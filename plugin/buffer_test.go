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
	rand "crypto/rand"
	"fmt"
	mathrand "math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	shmPath = "/tmp/ipc.test"
)

// BufferTestSuite is the testify suite for buffer tests
type BufferTestSuite struct {
	suite.Suite
	bm *bufferManager
}

func (s *BufferTestSuite) SetupTest() {
	os.Remove(shmPath) //nolint:errcheck // test cleanup, error ignored intentionally
	shmSize := 10 << 20
	mem := make([]byte, shmSize)
	var err error
	s.bm, err = createBufferManager([]*SizePercentPair{
		{defaultSingleBufferSize, 70},
		{16 * 1024, 20},
		{64 * 1024, 10},
	}, "", mem, 0)
	s.Require().NoError(err)
}

func (s *BufferTestSuite) newLinkedBufferWithSlice(slice *bufferSlice) *linkedBuffer {
	l := &linkedBuffer{
		sliceList:     newSliceList(),
		pinnedList:    newSliceList(),
		bufferManager: s.bm,
		endStream:     true,
		isFromShm:     slice.isFromShm,
	}
	l.sliceList.pushBack(slice)
	l.sliceList.writeSlice = l.sliceList.back()
	return l
}

func (s *BufferTestSuite) TestLinkedBuffer_ReadWrite() {
	factory := func() *linkedBuffer {
		buf, err := s.bm.allocShmBuffer(1024)
		s.Require().NoError(err)
		return s.newLinkedBufferWithSlice(buf)
	}
	s.testLinkedBufferReadBytes(factory)
}

func (s *BufferTestSuite) TestLinkedBuffer_ReleasePreviousRead() {
	slice, err := s.bm.allocShmBuffer(1024)
	s.Require().NoError(err)
	buf := s.newLinkedBufferWithSlice(slice)
	sliceNum := 100
	for i := 0; i < sliceNum*4096; i++ {
		s.Require().NoError(buf.WriteByte(byte(i)))
	}
	buf.done(true)

	for i := 0; i < sliceNum/2; i++ {
		r, err := buf.ReadBytes(4096)
		s.T().Logf("Iteration %d: len(r)=%d, buf.Len()=%d, err=%v", i, len(r), buf.Len(), err)
		s.Require().Equal(4096, len(r), "buf.Len():%d", buf.Len())
		s.Require().NoError(err)
	}
	s.T().Logf("After half reads: pinnedList.size()=%d", buf.pinnedList.size())
	s.Require().Equal((sliceNum/2)-1, buf.pinnedList.size())
	_, _ = buf.Discard(buf.Len())

	buf.releasePreviousReadAndReserve()
	s.T().Logf("After releasePreviousReadAndReserve: pinnedList.size()=%d, buf.Len()=%d", buf.pinnedList.size(), buf.Len())
	s.Require().Equal(0, buf.pinnedList.size())
	s.Require().Equal(0, buf.Len())
	//the last slice shouldn't release
	s.T().Logf("After release: sliceList.size()=%d, writeSlice!=nil? %v", buf.sliceList.size(), buf.sliceList.writeSlice != nil)
	s.Require().Equal(1, buf.sliceList.size())
	s.Require().True(buf.sliceList.writeSlice != nil)

	buf.ReleasePreviousRead()
	s.T().Logf("After ReleasePreviousRead: sliceList.size()=%d, writeSlice==nil? %v", buf.sliceList.size(), buf.sliceList.writeSlice == nil)
	s.Require().Equal(0, buf.sliceList.size())
	s.Require().True(buf.sliceList.writeSlice == nil)
}

func (s *BufferTestSuite) TestLinkedBuffer_FallbackWhenWrite() {
	mem := make([]byte, 10*1024)
	bm, err := createBufferManager([]*SizePercentPair{
		{Size: 1024, Percent: 100},
	}, "", mem, 0)
	s.Require().NoError(err)

	buf, err := bm.allocShmBuffer(1024)
	s.Require().NoError(err)
	writer := s.newLinkedBufferWithSlice(buf)
	dataSize := 1024
	mockDataArray := make([][]byte, 100)
	for i := range mockDataArray {
		mockDataArray[i] = make([]byte, dataSize)
		_, _ = rand.Read(mockDataArray[i])
		n, err := writer.WriteBytes(mockDataArray[i])
		s.Require().Equal(dataSize, n)
		s.Require().NoError(err)
		s.Require().Equal(dataSize*(i+1), writer.Len())
	}
	s.Require().False(writer.isFromShm)

	reader := writer.done(false)
	all := dataSize * len(mockDataArray)
	s.Require().Equal(all, writer.Len())

	for i := range mockDataArray {
		s.Require().Equal(all-i*dataSize, reader.Len())
		get, err := reader.ReadBytes(dataSize)
		if err != nil {
			s.T().Fatalf("reader.ReadBytes error %v %d", err, i)
		}
		s.Require().Equal(mockDataArray[i], get)
	}
}

func (s *BufferTestSuite) TestBufferReadString() {
	buf, err := s.bm.allocShmBuffer(16 << 10)
	s.Require().NoError(err)
	writer := s.newLinkedBufferWithSlice(buf)
	oneSliceSize := 16 << 10
	strBytesArray := make([][]byte, 100)
	for i := 0; i < len(strBytesArray); i++ {
		strBytesArray[i] = make([]byte, oneSliceSize)
		_, _ = rand.Read(strBytesArray[i])
		_, err := writer.WriteBytes(strBytesArray[i])
		s.Require().NoError(err)
	}

	reader := writer.done(false)
	for i := 0; i < len(strBytesArray); i++ {
		str, err := reader.ReadString(oneSliceSize)
		s.Require().NoError(err)
		s.Require().Equal(string(strBytesArray[i]), str)
	}
}

func (s *BufferTestSuite) TestLinkedBuffer_Reserve() {
	// alloc 3 buffer slice
	buffer := newLinkedBuffer(s.bm, (64+64+64)*1024)
	s.Require().Equal(3, buffer.sliceList.size())
	s.Require().True(buffer.isFromShm)
	s.Require().Equal(buffer.sliceList.front(), buffer.sliceList.writeSlice)

	// reserve a buf in first slice
	ret, err := buffer.Reserve(60 * 1024)
	s.Require().NoError(err)
	s.Require().Equal(len(ret), 60*1024)
	s.Require().Equal(3, buffer.sliceList.size())
	s.Require().True(buffer.isFromShm)
	s.Require().Equal(buffer.sliceList.front(), buffer.sliceList.writeSlice)

	// reserve a buf in the second slice when the first one is not enough
	ret, err = buffer.Reserve(6 * 1024)
	s.Require().NoError(err)
	s.Require().Equal(len(ret), 6*1024)
	s.Require().Equal(3, buffer.sliceList.size())
	s.Require().True(buffer.isFromShm)
	s.Require().Equal(buffer.sliceList.front().next(), buffer.sliceList.writeSlice)

	// reserve a buf in a new allocated slice
	ret, err = buffer.Reserve(128 * 1024)
	s.Require().NoError(err)
	s.Require().Equal(len(ret), 128*1024)
	s.Require().Equal(4, buffer.sliceList.size())
	s.Require().False(buffer.isFromShm)
	s.Require().Equal(buffer.sliceList.back(), buffer.sliceList.writeSlice)
}

func (s *BufferTestSuite) TestLinkedBuffer_Done() {
	mockDataSize := 128 * 1024
	mockData := make([]byte, mockDataSize)
	_, _ = rand.Read(mockData)
	// alloc 3 buffer slice
	buffer := newLinkedBuffer(s.bm, (64+64+64)*1024)
	s.Require().Equal(3, buffer.sliceList.size())

	// write data to full 2 slice, remove one
	_, _ = buffer.WriteBytes(mockData)
	reader := buffer.done(true)
	s.Require().Equal(2, buffer.sliceList.size())
	getBytes, _ := reader.ReadBytes(mockDataSize)
	s.Require().Equal(mockData, getBytes)
}

func (s *BufferTestSuite) TestLinkedBufferReadBytes() {
	factory := func() *linkedBuffer {
		buf, err := s.bm.allocShmBuffer(1024)
		s.Require().NoError(err)
		return s.newLinkedBufferWithSlice(buf)
	}
	s.testLinkedBufferReadBytes(factory)
}

func TestBufferTestSuite(t *testing.T) {
	suite.Run(t, new(BufferTestSuite))
}

func newLinkedBuffer(manager *bufferManager, size uint32) *linkedBuffer {
	l := &linkedBuffer{
		sliceList:     newSliceList(),
		pinnedList:    newSliceList(),
		bufferManager: manager,
		isFromShm:     true,
	}
	l.alloc(size)
	l.sliceList.writeSlice = l.sliceList.front()
	return l
}

func (s *BufferTestSuite) testLinkedBufferReadBytes(createBufferWriter func() *linkedBuffer) {
	writeAndRead := func(buf *linkedBuffer) {
		//2 MB
		size := 1 << 21
		data := make([]byte, size)
		_, _ = rand.Read(data)
		for buf.Len() < size {
			oneWriteSize := mathrand.Intn(size / 10)
			if buf.Len()+oneWriteSize > size {
				oneWriteSize = size - buf.Len()
			}
			n, err := buf.WriteBytes(data[buf.Len() : buf.Len()+oneWriteSize])
			s.Require().True(err == nil, err)
			s.Require().Equal(oneWriteSize, n)
		}

		buf.done(false)
		read := 0
		for buf.Len() > 0 {
			oneReadSize := mathrand.Intn(size / 10000)
			if read+oneReadSize > buf.Len() {
				oneReadSize = buf.Len()
			}
			//do nothing
			_, _ = buf.Peek(oneReadSize)

			readData, err := buf.ReadBytes(oneReadSize)
			s.Require().True(err == nil, err)
			if len(readData) == 0 {
				s.Require().Equal(oneReadSize, 0)
			} else {
				if !s.Equal(data[read:read+oneReadSize], readData) {
					exp := data[read : read+oneReadSize]
					act := readData
					fmt.Printf("Mismatch at read=%d, oneReadSize=%d\nExpected (len=%d): % x...\nActual   (len=%d): % x...\n", read, oneReadSize, len(exp), exp[:min(16, len(exp))], len(act), act[:min(16, len(act))])
				}
				s.Require().Equal(data[read:read+oneReadSize], readData)
			}
			read += oneReadSize
		}
		s.Require().Equal(1<<21, read)
		_, _ = buf.ReadBytes(-10)
		_, _ = buf.ReadBytes(0)
		buf.ReleasePreviousRead()
	}

	for i := 0; i < 100; i++ {
		writeAndRead(createBufferWriter())
	}
}

// Add a min helper for slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
