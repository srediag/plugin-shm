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
	"io"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/valyala/bytebufferpool"
)

var (
	_ BufferWriter = &linkedBuffer{}
	_ BufferReader = &linkedBuffer{}
)

// BufferWriter used to write data to stream.
//
// NOTE: BufferWriter and BufferReader are NOT safe for concurrent use by multiple goroutines.
// If you need to use a buffer concurrently, synchronize all access externally.
//
// Always call ReleasePreviousRead() when it is safe to drop all previous results of ReadBytes and Peek,
// otherwise shared memory may leak. See method docs for details.
type BufferWriter interface {
	//Len() return the current wrote size of buffer.
	//It will traverse all underlying slices to compute the unread size, please don't call frequently.
	Len() int
	io.ByteWriter
	//Reserve `size` byte share memory space, user could use it implement zero copy write.
	Reserve(size int) ([]byte, error)
	//Copy data to share memory.
	//return value: `n` is the written size
	//return value: `err`, is nil mean that succeed, otherwise failure.
	WriteBytes(data []byte) (n int, err error)
	//Copy string to share memory
	WriteString(string) error
}

// BufferReader used to read data from stream.
//
// WARNING: BufferReader is NOT safe for concurrent use by multiple goroutines.
// Always synchronize access externally if used concurrently.
//
// After calling ReleasePreviousRead(), all previously returned slices from ReadBytes/Peek become invalid.
type BufferReader interface {
	io.ByteReader

	//Len() return the current unread size of buffer.
	//It will traverse all underlying slices to compute the unread size, please don't call frequently.
	Len() int

	//Read `size` bytes from share memory, which maybe block if size is greater than Len().
	//Notice: when ReleasePreviousRead() was called, the results of previous ReadBytes() will be invalid.
	ReadBytes(size int) ([]byte, error)

	//Peek `size` bytes from share memory. the different between Peek() and ReadBytes() is that
	//Peek() don't influence the return value of Len(), but the ReadBytes() will decrease the unread size.
	//eg: the buffer is [0,1,2,3]
	//1. after Peek(2), the buffer is also [0,1,2,3], and the Len() is 4.
	//2. after ReadBytes(3), the buffer is [3], and the Len() is 1.
	//Notice: when ReleasePreviousRead was called, the results of previous Peek call is invalid .
	Peek(size int) ([]byte, error)

	//Drop data of given length. If there's no that much data, will block until the data is enough to discard
	Discard(size int) (int, error)

	/* Call ReleasePreviousRead when it is safe to drop all previous result of ReadBytes and Peek, otherwise shm memory will leak.
	  eg:
	    buf, err := BufferReader.ReadBytes(size) // or Buffer.
		//do
	*/
	ReleasePreviousRead()

	//If you would like to read string from the buffer, ReadString(size) is better than string(ReadBytes(size)).
	ReadString(size int) (string, error)
}

type linkedBuffer struct {
	mu            sync.Mutex
	recycleMux    sync.Mutex
	sliceList     *sliceList
	bufferManager *bufferManager
	stream        *Stream
	// Already read slices dropped by ReadBytes will be saved here instead of recycled instantly.
	// Slices inside be recycled when ReleasePreviousRead is called.
	pinnedList *sliceList
	// if sliceList.Front() is pinned(initialized with false, be turned to true when ReadByte and Peek)
	currentPinned bool
	endStream     bool
	isFromShm     bool
	len           int
}

func newEmptyLinkedBuffer(manager *bufferManager) *linkedBuffer {
	l := &linkedBuffer{
		sliceList:     newSliceList(),
		pinnedList:    newSliceList(),
		bufferManager: manager,
		isFromShm:     true,
	}
	return l
}

func (l *linkedBuffer) Len() int {
	return l.len
}

func (l *linkedBuffer) copyWriteAndFlush(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}
	written, err := l.WriteBytes(data)
	if err != nil {
		return 0, err
	}

	err = l.stream.Flush(false)
	return written, err
}

func (l *linkedBuffer) WriteByte(b byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sliceList.writeSlice == nil {
		l.alloc(1)
		l.sliceList.writeSlice = l.sliceList.front()
	}
	n := l.sliceList.writeSlice.append(b)
	if n == 1 {
		l.len++
		return nil
	}
	l.alloc(1)
	l.sliceList.writeSlice = l.sliceList.writeSlice.next()
	l.sliceList.writeSlice.append(b)
	l.len++
	return nil
}

func (l *linkedBuffer) WriteBytes(data []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(data) == 0 {
		return
	}
	if l.sliceList.writeSlice == nil {
		l.alloc(uint32(len(data) - n))
		l.sliceList.writeSlice = l.sliceList.front()
	}
	for {
		n += l.sliceList.writeSlice.append(data[n:]...)
		if n < len(data) {
			// l.sliceList.write slice must be used out
			if l.sliceList.writeSlice.next() == nil {
				// which means no allocated bufferSlice is left
				l.alloc(uint32(len(data) - n))
			}
			l.sliceList.writeSlice = l.sliceList.writeSlice.next()
		} else {
			// n equals len(data)
			break
		}
	}
	l.len += n
	return
}

// 1. if cur slice can contain the size, then reserve and return it
// 2. if the next slice can contain the size, then reserve and return it
// 3. alloc a new slice which can contain the size
func (l *linkedBuffer) Reserve(size int) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 1. use current slice
	if l.sliceList.writeSlice == nil {
		l.alloc(uint32(size))
		l.sliceList.writeSlice = l.sliceList.front()
	}
	ret, err := l.sliceList.writeSlice.reserve(size)
	if err == nil {
		l.len += size
		return ret, err
	}

	// 2. use next slice
	if e := l.sliceList.writeSlice.next(); e != nil {
		ret, err = e.reserve(size)
		if err == nil {
			l.sliceList.writeSlice = e
			l.len += size
			return ret, err
		}
	}

	// 3. alloc a new slice
	// TODO: This requires shared memory support (allocShmBuffer)
	// buf, err := l.bufferManager.allocShmBuffer(uint32(size))
	// if err == nil {
	// 	l.sliceList.pushBack(buf)
	// } else {
	// 	// fallback
	// 	allocSize := size
	// 	if allocSize < defaultSingleBufferSize {
	// 		allocSize = defaultSingleBufferSize
	// 	}
	// 	// Use bytebufferpool for fallback []byte allocation (non-shared-memory buffer)
	// 	bb := bytebufferpool.Get()
	// 	if cap(bb.B) < allocSize {
	// 		bb.B = make([]byte, allocSize)
	// 	}
	// 	l.sliceList.pushBack(newBufferSlice(nil, bb.B[:allocSize], 0, false))
	// 	l.isFromShm = false
	// }
	l.sliceList.writeSlice = l.sliceList.back()
	l.len += size
	return l.sliceList.writeSlice.reserve(size)
}

func (l *linkedBuffer) WriteString(str string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.WriteBytes(string2bytesZeroCopy(str))
	return err
}

func (l *linkedBuffer) recycle() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recycleMux.Lock()
	defer l.recycleMux.Unlock()
	for l.sliceList.size() > 0 {
		slice := l.sliceList.popFront()
		// TODO: This requires shared memory support (recycleBuffer)
		// if slice.isFromShm {
		// 	l.bufferManager.recycleBuffer(slice)
		// } else {
		if slice.Data != nil {
			// Return pooled []byte to bytebufferpool for non-shared-memory buffer
			bb := &bytebufferpool.ByteBuffer{B: slice.Data}
			bytebufferpool.Put(bb)
		}
		putBackBufferSlice(slice)
		// }
	}
	l.clean()
}

func (l *linkedBuffer) rootBufOffset() uint32 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.sliceList.front().offsetInShm
}

func (l *linkedBuffer) done(endStream bool) BufferReader {
	_ = endStream
	// todo endStream
	if l.isFromShm {
		for slice := l.sliceList.front(); slice != nil; slice = slice.nextSlice {
			slice.update()
			if slice == l.sliceList.writeSlice {
				break
			}
		}
		// recycle unused slice
		if l.sliceList.writeSlice != nil && l.sliceList.writeSlice.next() != nil {
			// TODO: This requires shared memory support (recycleBuffer)
			// for slice := head; slice != nil; {
			// 	next := slice.nextSlice
			// 	l.bufferManager.recycleBuffer(slice)
			// 	slice = next
			// }
		}
	}
	return l
}

func (l *linkedBuffer) underlyingData() [][]byte {
	l.mu.Lock()
	defer l.mu.Unlock()
	data := make([][]byte, 0, 4)
	for slice := l.sliceList.front(); slice != nil; slice = slice.next() {
		data = append(data, slice.Data[slice.readIndex:slice.writeIndex])
		if slice == l.sliceList.writeSlice {
			break
		}
	}
	return data
}

func (l *linkedBuffer) read(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	size := len(p)
	if size <= 0 {
		return
	}
	written := 0
	if l.len < 1 {
		if err = l.stream.readMore(1); err != nil {
			return 0, err
		}
	}
	for front := l.sliceList.front(); front != nil && size > written; {
		b, err := front.read(size - written)
		written += copy(p[written:], b)
		if err == nil {
			break
		} else if err == ErrNotEnoughData {
			l.readNextSlice()
			front = l.sliceList.front()
		}
	}
	l.len -= written
	return written, nil
}

func (l *linkedBuffer) ReadByte() (byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.len < 1 {
		if err := l.stream.readMore(1); err != nil {
			return 0, err
		}
	}
	r, err := l.sliceList.front().read(1)
	if err == nil {
		l.len--
		return r[0], nil
	}
	l.readNextSlice()
	r, _ = l.sliceList.front().read(1)
	l.len--
	return r[0], nil
}

func (l *linkedBuffer) ReadBytes(size int) (result []byte, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if size <= 0 {
		return
	}
	if l.len < size {
		if err = l.stream.readMore(size); err != nil {
			return nil, err
		}
	}

	if l.sliceList.front().size() == 0 {
		l.readNextSlice()
	}

	if l.sliceList.front().size() >= size {
		l.currentPinned = true
		l.len -= size
		return l.sliceList.front().read(size)
	}
	// slow path
	l.len -= size
	result = make([]byte, 0, size)

	for size > 0 {
		readData, _ := l.sliceList.front().read(size)
		result = append(result, readData...) // since we only copy the data, there's no need to mark the node as pinned
		if len(readData) != size {
			l.readNextSlice()
		}
		size -= len(readData)
	}
	return
}

func (l *linkedBuffer) ReadString(size int) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if size <= 0 {
		return "", nil
	}
	if l.len < size {
		if err := l.stream.readMore(size); err != nil {
			return "", err
		}
	}

	if l.sliceList.front().size() >= size {
		data, _ := l.sliceList.front().read(size)
		l.len -= size
		return string(data), nil
	}

	//slow path, the sized buffer cross multi buffer slice
	s := make([]byte, size)

	written := 0
	for written < size {
		if l.sliceList.front().size() == 0 {
			l.readNextSlice()
		}
		readData, _ := l.sliceList.front().read(size - written)
		written += copy(s[written:], readData)
	}
	l.len -= size
	return *(*string)(unsafe.Pointer(&s)), nil
}

// Peek isn't influence l.Len()
func (l *linkedBuffer) Peek(size int) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if size <= 0 {
		return nil, nil
	}
	if l.len < size {
		if err := l.stream.readMore(size); err != nil {
			return nil, err
		}
	}

	readBytes, _ := l.sliceList.front().peek(size)
	if len(readBytes) == size {
		l.currentPinned = true
		return readBytes, nil
	}

	// slow path
	result := make([]byte, 0, size)

	result = append(result, readBytes...)
	size -= len(readBytes)
	for e := l.sliceList.front().next(); size > 0 && e != nil; e = e.next() {
		readBytes, _ := e.peek(size)
		result = append(result, readBytes...)
		size -= len(readBytes)
	}

	return result, nil
}

func (l *linkedBuffer) Discard(size int) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.len < size {
		if err = l.stream.readMore(size); err != nil {
			return
		}
	}
	for {
		skip := l.sliceList.front().skip(size)
		n += skip
		size -= skip
		if size == 0 {
			break
		}
		l.readNextSlice()
	}
	l.len -= n
	return
}

// cleanPinnedList recycles all pinned slices. Only called internally.
//
// Slices are pinned when returned by ReadBytes/Peek and not yet released.
func (l *linkedBuffer) cleanPinnedList() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.pinnedList.size() == 0 {
		return
	}
	l.currentPinned = false
	for l.pinnedList.size() > 0 {
		slice := l.pinnedList.popFront()
		// TODO: This requires shared memory support (recycleBuffer)
		// if slice.isFromShm {
		// 	l.bufferManager.recycleBuffer(slice)
		// } else {
		putBackBufferSlice(slice)
		// }
	}
}

// ReleasePreviousRead releases all pinned slices and recycles them if possible.
//
// Call this method ONLY when it is safe to drop all previous results of ReadBytes and Peek.
// Failing to do so will leak shared memory.
func (l *linkedBuffer) ReleasePreviousRead() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cleanPinnedList()

	if l.sliceList.size() == 0 {
		return
	}

	if l.sliceList.front().size() == 0 && l.sliceList.front() == l.sliceList.writeSlice {
		// TODO: This requires shared memory support (recycleBuffer)
		// l.bufferManager.recycleBuffer(l.sliceList.popFront())
		l.sliceList.writeSlice = nil
	}
}

func (l *linkedBuffer) releasePreviousReadAndReserve() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cleanPinnedList()
	//try reserve a buffer slice  in long-stream mode for improving performance.
	//we could use read buffer as next write buffer, to avoiding share memory allocate and recycle.
	if l.len == 0 && l.sliceList.size() == 1 {
		if l.sliceList.front().isFromShm {
			l.sliceList.front().reset()
		} else {
			putBackBufferSlice(l.sliceList.popFront())
		}
	}
}

// readNextSlice advances to the next buffer slice, recycling or pinning as needed.
//
// If the current slice is pinned, it is added to pinnedList for later recycling.
// Otherwise, it is recycled immediately.
func (l *linkedBuffer) readNextSlice() {
	l.mu.Lock()
	defer l.mu.Unlock()
	// slice := l.sliceList.popFront()
	// TODO: This requires shared memory support (recycleBuffer)
	// if slice.isFromShm {
	// 	if l.currentPinned {
	// 		l.pinnedList.pushBack(slice)
	// 	} else {
	// 		l.bufferManager.recycleBuffer(slice)
	// 	}
	// }
	l.currentPinned = false
}

// alloc allocates a new buffer slice, preferring shared memory but falling back to heap if needed.
func (l *linkedBuffer) alloc(size uint32) {
	l.mu.Lock()
	defer l.mu.Unlock()
	remain := int64(size)
	// TODO: This requires shared memory support (allocShmBuffer)
	// buf, err := l.bufferManager.allocShmBuffer(size)
	// if err == nil {
	// 	l.sliceList.pushBack(buf)
	// 	return
	// }
	// allocSize := l.bufferManager.allocShmBuffers(l.sliceList, size)
	// remain -= allocSize
	// fallback. alloc memory buffer (not shm)
	if remain > 0 {
		if remain < defaultSingleBufferSize {
			remain = defaultSingleBufferSize
		}
		// Use bytebufferpool for fallback []byte allocation (non-shared-memory buffer)
		bb := bytebufferpool.Get()
		if cap(bb.B) < int(remain) {
			bb.B = make([]byte, int(remain))
		}
		l.sliceList.pushBack(newBufferSlice(nil, bb.B[:int(remain)], 0, false))
		l.isFromShm = false
		// in unit test, l.stream maybe is nil
		if l.stream != nil {
			atomic.AddUint64(&l.stream.session.stats.allocShmErrorCount, 1)
		}
	}
}

func (l *linkedBuffer) isFromShareMemory() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isFromShm
}

func (l *linkedBuffer) appendBufferSlice(slice *bufferSlice) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if slice == nil {
		return
	}
	l.sliceList.pushBack(slice)
	if !slice.isFromShm {
		l.isFromShm = false
	}
	l.len += slice.size()
	l.sliceList.writeSlice = slice
}

// todo
func (l *linkedBuffer) clean() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.sliceList.size() > 0 {
		slice := l.sliceList.popFront()
		if slice != nil && slice.Data != nil && !slice.isFromShm {
			// Return pooled []byte to bytebufferpool for non-shared-memory buffer
			bb := &bytebufferpool.ByteBuffer{B: slice.Data}
			bytebufferpool.Put(bb)
		}
		putBackBufferSlice(slice)
	}
	l.sliceList.writeSlice = nil
	l.isFromShm = true
	l.endStream = false
	l.currentPinned = false
	l.len = 0
}

func (l *linkedBuffer) bindStream(s *Stream) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stream = s
}
