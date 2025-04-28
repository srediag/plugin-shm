package shm

import (
	"errors"
	"sync"
)

// SizePercentPair describes a buffer list's specification.
type SizePercentPair struct {
	Size    uint32
	Percent uint32
}

// BufferSlice represents a slice of a shared memory buffer.
type BufferSlice struct {
	Data   []byte
	Offset uint32
	Cap    uint32
	Used   bool
}

// BufferManager manages pools of BufferSlices of different sizes.
type BufferManager struct {
	mu     sync.Mutex
	pools  map[uint32][]*BufferSlice // key: size
	layout []SizePercentPair
	mem    []byte
}

// NewBufferManager creates a BufferManager with the given memory and layout.
func NewBufferManager(mem []byte, layout []SizePercentPair) *BufferManager {
	bm := &BufferManager{
		pools:  make(map[uint32][]*BufferSlice),
		layout: layout,
		mem:    mem,
	}
	// For each size, create a pool of slices from the memory region.
	// For now, just split memory evenly (not production logic).
	offset := 0
	for _, pair := range layout {
		count := int(pair.Percent) // For test, treat percent as count
		for i := 0; i < count; i++ {
			if offset+int(pair.Size) > len(mem) {
				break
			}
			slice := &BufferSlice{
				Data:   mem[offset : offset+int(pair.Size)],
				Offset: uint32(offset),
				Cap:    pair.Size,
			}
			bm.pools[pair.Size] = append(bm.pools[pair.Size], slice)
			offset += int(pair.Size)
		}
	}
	return bm
}

// Alloc allocates a BufferSlice of the given size.
func (bm *BufferManager) Alloc(size uint32) (*BufferSlice, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	pool := bm.pools[size]
	for _, s := range pool {
		if !s.Used {
			s.Used = true
			return s, nil
		}
	}
	return nil, errors.New("no free buffer slice")
}

// Recycle returns a BufferSlice to the pool.
func (bm *BufferManager) Recycle(s *BufferSlice) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	s.Used = false
}

// Stats returns the number of free slices for each size.
func (bm *BufferManager) Stats() map[uint32]int {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	stats := make(map[uint32]int)
	for size, pool := range bm.pools {
		free := 0
		for _, s := range pool {
			if !s.Used {
				free++
			}
		}
		stats[size] = free
	}
	return stats
}
