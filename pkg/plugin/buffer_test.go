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
	"context"
	rand "crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	shmPath = "/tmp/ipc.test"
)

// BufferTestSuite is the testify suite for buffer tests
type BufferTestSuite struct {
	suite.Suite
	bm *BufferManager
}

func (s *BufferTestSuite) SetupTest() {
	os.Remove(shmPath)                       //nolint:errcheck // test cleanup, error ignored intentionally
	s.bm = NewBufferManager(4096, 100, true) // 100 buffers of 4KiB, fallback enabled
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

// Remove or comment out all tests and code that depend on the old bufferManager, bufferSlice, or linkedBuffer types, except for BufferTestSuite and TestBufferReadString.
// Leave a TODO comment for porting or rewriting the other tests to the new API.

func (s *BufferTestSuite) TestLinkedBuffer_ReadWrite() {
	// TODO: Implement this test
}

func (s *BufferTestSuite) TestLinkedBuffer_ReleasePreviousRead() {
	// TODO: Implement this test
}

func (s *BufferTestSuite) TestLinkedBuffer_FallbackWhenWrite() {
	// TODO: Implement this test
}

func (s *BufferTestSuite) TestBufferReadString() {
	s.T().Logf("[START] TestBufferReadString")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	buf, err := s.bm.Get(ctx)
	s.Require().NoError(err)
	oneSliceSize := 4096
	strBytesArray := make([][]byte, 10)
	for i := 0; i < len(strBytesArray); i++ {
		strBytesArray[i] = make([]byte, oneSliceSize)
		_, _ = rand.Read(strBytesArray[i])
		copy(buf.Data, strBytesArray[i])
		s.T().Logf("Wrote %d/%d slices", i+1, len(strBytesArray))
	}
	s.Require().NoError(s.bm.Put(buf))
	// Simulate reading back
	buf, err = s.bm.Get(ctx)
	s.Require().NoError(err)
	for i := 0; i < len(strBytesArray); i++ {
		read := buf.Data[:oneSliceSize]
		s.T().Logf("ReadString %d/%d: got %d bytes", i+1, len(strBytesArray), len(read))
		s.Require().Equal(string(strBytesArray[i]), string(read))
	}
	s.Require().NoError(s.bm.Put(buf))
	s.T().Logf("[END] TestBufferReadString")
}

func (s *BufferTestSuite) TestLinkedBuffer_Reserve() {
	// TODO: Implement this test
}

func (s *BufferTestSuite) TestLinkedBuffer_Done() {
	// TODO: Implement this test
}

func (s *BufferTestSuite) TestLinkedBufferReadBytes() {
	// TODO: Implement this test
}

func TestBufferTestSuite(t *testing.T) {
	suite.Run(t, new(BufferTestSuite))
}

func newLinkedBuffer(manager *BufferManager, size uint32) *linkedBuffer {
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
	// TODO: Implement this test
}

// Add a min helper for slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func logFreeBuffersByPool(t *testing.T, bm *BufferManager, label string) {
	// TODO: Implement this function
}
