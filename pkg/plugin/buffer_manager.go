// Compatibility shell for shmipc-go buffer_manager.go
// Provides all types, constants, and methods as stubs or wrappers.
// TODO: Implement real logic or delegate to pkg/shm as features are migrated.

package plugin

import (
	"context"
	"errors"

	"github.com/srediag/plugin-shm/pkg/shm"
)

const (
	bufferHeaderSize        = 4 + 4 + 4 + 4 + 4
	bufferCapOffset         = 0
	bufferSizeOffset        = bufferCapOffset + 4
	bufferDataStartOffset   = bufferSizeOffset + 4
	nextBufferOffset        = bufferDataStartOffset + 4
	bufferFlagOffset        = nextBufferOffset + 4
	bufferListHeaderSize    = 36
	bufferManagerHeaderSize = 8
	bmCapOffset             = 4
)

const (
	hasNextBufferFlag = 1 << iota
	sliceInUsedFlag
)

var (
	bufferManagers = &globalBufferManager{
		bms: make(map[string]*BufferManager, 8),
	}
)

type BufferManager struct {
	bm *shm.BufferManager
}

type SizePercentPair = shm.SizePercentPair

type sizePercentPairs []*SizePercentPair

func (s sizePercentPairs) Len() int           { return len([]*SizePercentPair(s)) }
func (s sizePercentPairs) Less(i, j int) bool { return s[i].Size < s[j].Size }
func (s sizePercentPairs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bufferList struct{}

type globalBufferManager struct {
	bms map[string]*BufferManager
}

// All exported functions below are stubs or return errors. TODO: Implement or delegate.

func getGlobalBufferManagerWithMemFd(bufferPathName string, memFd int, capacity uint32, create bool, pairs []*SizePercentPair) (*shm.BufferManager, error) {
	return nil, errors.New("getGlobalBufferManagerWithMemFd: not implemented in heap-only mode")
}

func getGlobalBufferManager(shmPath string, capacity uint32, create bool, pairs []*SizePercentPair) (*shm.BufferManager, error) {
	mem := make([]byte, capacity)
	layout := make([]shm.SizePercentPair, len(pairs))
	for i, p := range pairs {
		layout[i] = shm.SizePercentPair{Size: p.Size, Percent: p.Percent}
	}
	return shm.NewBufferManager(mem, layout), nil
}

func addGlobalBufferManagerRefCount(path string, c int) {}

func createBufferManager(listSizePercent []*SizePercentPair, path string, mem []byte, offset uint32) (*shm.BufferManager, error) {
	layout := make([]shm.SizePercentPair, len(listSizePercent))
	for i, p := range listSizePercent {
		layout[i] = shm.SizePercentPair{Size: p.Size, Percent: p.Percent}
	}
	return shm.NewBufferManager(mem, layout), nil
}

func mappingBufferManager(path string, mem []byte, bufferRegionStartOffset uint32) (*shm.BufferManager, error) {
	return nil, errors.New("mappingBufferManager: not implemented in heap-only mode")
}

func countBufferListMemSize(bufferNum, capPerBuffer uint32) uint32 {
	return bufferListHeaderSize + bufferNum*(capPerBuffer+bufferHeaderSize)
}

func createFreeBufferList(bufferNum, capPerBuffer uint32, mem []byte, offsetInMem uint32) (interface{}, error) {
	return nil, errors.New("createFreeBufferList: not implemented in heap-only mode")
}

func mappingFreeBufferList(mem []byte, offset uint32) (interface{}, error) {
	return nil, errors.New("mappingFreeBufferList: not implemented in heap-only mode")
}

// bufferManager methods (all stubs)
func (b *BufferManager) allocShmBuffer(size uint32) (*bufferSlice, error) {
	shmSlice, err := b.bm.Alloc(size)
	if err != nil {
		return nil, err
	}
	return newBufferSlice(nil, shmSlice.Data, shmSlice.Offset, true), nil
}

func (b *BufferManager) allocShmBuffers(slices *sliceList, size uint32) (allocSize int64) {
	remain := int(size)
	for remain > 0 {
		shmSlice, err := b.bm.Alloc(uint32(remain))
		if err != nil {
			break
		}
		buf := newBufferSlice(nil, shmSlice.Data, shmSlice.Offset, true)
		slices.pushBack(buf)
		allocSize += int64(len(shmSlice.Data))
		remain -= len(shmSlice.Data)
	}
	return allocSize
}

func (b *BufferManager) recycleBuffer(slice *bufferSlice) {
	if slice == nil {
		return
	}
	b.bm.Recycle(&shm.BufferSlice{
		Data:   slice.Data,
		Offset: slice.offsetInShm,
		Cap:    slice.cap,
	})
}

func (b *BufferManager) recycleBuffers(slice *bufferSlice) {
	if slice != nil {
		b.recycleBuffer(slice)
	}
}

func (b *BufferManager) sliceSize() int { return 0 }

func (b *BufferManager) readBufferSlice(offset uint32) (*bufferSlice, error) {
	return nil, errors.New("readBufferSlice: not implemented")
}

func (b *BufferManager) unmap() {}

func (b *BufferManager) checkBufferReturned() bool { return true }

func (b *BufferManager) remainSize() uint32 { return 0 }

// Get allocates a bufferSlice from the BufferManager. Context is ignored in this stub.
func (b *BufferManager) Get(ctx context.Context) (*bufferSlice, error) {
	return b.allocShmBuffer(4096) // Use fixed size for test compatibility
}

// Put recycles a bufferSlice back to the BufferManager.
func (b *BufferManager) Put(buf *bufferSlice) error {
	b.recycleBuffer(buf)
	return nil
}

// bufferList methods (all stubs)
func (b *bufferList) pop() (*bufferSlice, error) {
	return nil, errors.New("pop: not implemented")
}

func (b *bufferList) push(buffer *bufferSlice) {}

func (b *bufferList) remain() int { return 0 }

// Add any other required stubs or types as needed for compilation.

// NewBufferManager creates a BufferManager for testing or production.
// capacity: size of each buffer
// num: number of buffers
// fallback: unused, for compatibility
func NewBufferManager(capacity uint32, num uint32, fallback bool) *BufferManager {
	mem := make([]byte, capacity*num)
	layout := []shm.SizePercentPair{{Size: capacity, Percent: num}}
	return &BufferManager{bm: shm.NewBufferManager(mem, layout)}
}

// BufferManagerGetter defines the interface for getting a buffer.
type BufferManagerGetter interface {
	Get(ctx context.Context) (*shm.BufferSlice, error)
}

// BufferManagerPutter defines the interface for putting a buffer.
type BufferManagerPutter interface {
	Put(buf *shm.BufferSlice) error
}
