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
	"testing"
	"unsafe"

	"github.com/stretchr/testify/suite"
)

type BufferSliceSuite struct {
	suite.Suite
}

func (s *BufferSliceSuite) TestBufferSlice_ReadWrite() {
	s.T().Logf("[START] TestBufferSlice_ReadWrite")
	size := 8192
	slice := newBufferSlice(nil, make([]byte, size), 0, false)

	for i := 0; i < size; i++ {
		n := slice.append(byte(i))
		s.Equal(1, n)
	}
	n := slice.append(byte(10))
	s.Equal(0, n)

	data, err := slice.read(size * 10)
	s.Equal(size, len(data))
	s.Equal(err, ErrNotEnoughData)

	//verify read data.
	for i := 0; i < size; i++ {
		s.Equal(byte(i), data[i])
	}
	s.T().Logf("[END] TestBufferSlice_ReadWrite")
}

func (s *BufferSliceSuite) TestBufferSlice_Skip() {
	s.T().Logf("[START] TestBufferSlice_Skip")
	slice := newBufferSlice(nil, make([]byte, 8192), 0, false)
	slice.append(make([]byte, slice.cap)...)
	remain := int(slice.cap)

	n := slice.skip(10)
	remain -= n
	s.Equal(remain, slice.size())

	n = slice.skip(100)
	remain -= n
	s.Equal(remain, slice.size())

	_ = slice.skip(10000)
	s.Equal(0, slice.size())
	s.T().Logf("[END] TestBufferSlice_Skip")
}

func (s *BufferSliceSuite) TestBufferSlice_Reserve() {
	s.T().Logf("[START] TestBufferSlice_Reserve")
	size := 8192
	slice := newBufferSlice(nil, make([]byte, size), 0, false)
	data1, err := slice.reserve(100)
	s.Equal(nil, err)
	s.Equal(100, len(data1))

	data2, err := slice.reserve(size)
	s.Equal(ErrNoMoreBuffer, err)
	s.Equal(0, len(data2))

	for i := range data1 {
		data1[i] = byte(i)
	}
	for i := range data2 {
		data2[i] = byte(i)
	}

	readData, err := slice.read(100)
	s.Equal(nil, err)
	s.Equal(len(data1), len(readData))

	for i := 0; i < len(data1); i++ {
		s.Equal(data1[i], readData[i])
	}

	readData, err = slice.read(10000)
	s.Equal(ErrNotEnoughData, err)
	s.Equal(len(data2), len(readData))

	for i := 0; i < len(data2); i++ {
		s.Equal(data2[i], readData[i])
	}
	s.T().Logf("[END] TestBufferSlice_Reserve")
}

func (s *BufferSliceSuite) TestBufferSlice_Update() {
	s.T().Logf("[START] TestBufferSlice_Update")
	size := 8192
	header := make([]byte, bufferHeaderSize)
	*(*uint32)(unsafe.Pointer(&header[bufferCapOffset])) = uint32(size)
	slice := newBufferSlice(header, make([]byte, size), 0, true)

	n := slice.append(make([]byte, size)...)
	s.Equal(size, n)
	slice.update()
	s.Equal(size, int(*(*uint32)(unsafe.Pointer(&slice.bufferHeader[bufferSizeOffset]))))
	s.T().Logf("[END] TestBufferSlice_Update")
}

func (s *BufferSliceSuite) TestBufferSlice_linkedNext() {
	s.T().Logf("[START] TestBufferSlice_linkedNext")
	// size := 8192
	// sliceNum := 100

	// slices := make([]*bufferSlice, 0, sliceNum)
	// mem := make([]byte, 10<<20)
	// bm, err := createBufferManager([]*SizePercentPair{
	//  {Size: uint32(size), Percent: 100},
	// }, "", mem, 0)
	// s.Equal(nil, err)
	// writeDataArray := make([][]byte, 0, sliceNum)

	// TODO: Enable this test when shared memory support is implemented
	// for i := 0; i < sliceNum; i++ {
	//  slice, err := bm.AllocShmBuffer(uint32(size))
	//  s.Equal(nil, err, "i:%d", i)
	//  data := make([]byte, size)
	//  _, _ = rand.Read(data)
	//  writeDataArray = append(writeDataArray, data)
	//  s.Equal(uint32(size), slice.cap)
	//  s.Equal(size, slice.append(data...))
	//  slice.update()
	//  slices = append(slices, slice)
	// }

	// for i := 0; i <= len(slices)-2; i++ {
	//  slices[i].linkNext(slices[i+1].offsetInShm)
	// }

	// next := slices[0].offsetInShm
	// for i := 0; i < sliceNum; i++ {
	//  slice, err := bm.ReadBufferSlice(next)
	//  s.Equal(nil, err)
	//  s.Equal(uint32(size), slice.cap)
	//  s.Equal(size, slice.size())
	//  readData, err := slice.read(size)
	//  s.Equal(nil, err, "i:%d offset:%d", i, next)
	//  s.Equal(readData, writeDataArray[i])
	//  isLastSlice := i == sliceNum-1
	//  s.Equal(!isLastSlice, slice.hasNext())
	//  next = slice.nextBufferOffset()
	// }
	s.T().Logf("[END] TestBufferSlice_linkedNext")
}

func (s *BufferSliceSuite) TestSliceList_PushPop() {
	//1. twice push , twice pop
	l := newSliceList()
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	s.Equal(l.front(), l.back())
	s.Equal(1, l.size())

	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	s.Equal(2, l.size())
	s.Equal(false, l.front() == l.back())

	s.Equal(l.front(), l.popFront())
	s.Equal(1, l.size())
	s.Equal(l.front(), l.back())

	s.Equal(l.front(), l.popFront())
	s.Equal(0, l.size())
	s.Equal(true, l.front() == nil)
	s.Equal(true, l.back() == nil)

	// multi push and pop
	const iterCount = 100
	for i := 0; i < iterCount; i++ {
		l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
		s.Equal(i+1, l.size())
	}
	for i := 0; i < iterCount; i++ {
		l.popFront()
		s.Equal(iterCount-(i+1), l.size())
	}
	s.Equal(0, l.size())
	s.Equal(true, l.front() == nil)
	s.Equal(true, l.back() == nil)
}

func (s *BufferSliceSuite) TestSliceList_SplitFromWrite() {
	//1. sliceList's size == 1
	l := newSliceList()
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	l.writeSlice = l.front()
	s.Equal(true, l.splitFromWrite() == nil)
	s.Equal(1, l.size())
	s.Equal(l.front(), l.back())
	s.Equal(l.back(), l.writeSlice)

	//2. sliceList's size == 2, writeSlice's index is 0
	l = newSliceList()
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	l.writeSlice = l.front()
	s.Equal(l.back(), l.splitFromWrite())
	s.Equal(1, l.size())
	s.Equal(l.front(), l.back())
	s.Equal(l.back(), l.writeSlice)

	//2. sliceList's size == 2, writeSlice's index is 1
	l = newSliceList()
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
	l.writeSlice = l.back()
	s.Equal(true, l.splitFromWrite() == nil)
	s.Equal(2, l.size())
	s.Equal(l.back(), l.writeSlice)

	//3. sliceList's size == 2, writeSlice's index is 50
	l = newSliceList()
	for i := 0; i < 100; i++ {
		l.pushBack(newBufferSlice(nil, make([]byte, 1024), 0, false))
		if i == 50 {
			l.writeSlice = l.back()
		}
	}
	l.splitFromWrite()
	s.Equal(l.writeSlice, l.back())
	s.Equal(51, l.size())
}

func TestBufferSliceSuite(t *testing.T) {
	suite.Run(t, new(BufferSliceSuite))
}
