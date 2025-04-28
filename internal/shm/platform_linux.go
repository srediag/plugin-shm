//go:build linux

package shm

import (
	"context"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// MapRegion maps or creates a shared memory region (Linux implementation).
func MapRegion(ctx context.Context, opts MapOptions) (*MappedRegion, error) {
	flags := unix.O_RDWR
	if opts.Create {
		flags |= unix.O_CREAT
	}
	shmPath := filepath.Join("/dev/shm", opts.Name)
	fd, err := unix.Open(shmPath, flags, 0600)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if opts.Create {
		if err := unix.Ftruncate(fd, int64(opts.Size)); err != nil {
			_ = unix.Close(fd)
			return nil, fmt.Errorf("ftruncate: %w", err)
		}
	}
	addr, err := unix.Mmap(fd, 0, opts.Size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("mmap: %w", err)
	}
	return &MappedRegion{
		Addr: addr,
		// Store fd for cleanup
		// Add more fields if needed
	}, nil
}

// UnmapRegion unmaps and closes the shared memory region (Linux implementation).
func UnmapRegion(ctx context.Context, region *MappedRegion) error {
	if region == nil || region.Addr == nil {
		return nil
	}
	if err := unix.Munmap(region.Addr); err != nil {
		return fmt.Errorf("munmap: %w", err)
	}
	// TODO: close fd if tracked
	return nil
}
