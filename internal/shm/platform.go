// Package shm contains platform-specific helpers for shared memory buffer implementation.
package shm

// MappedRegion represents a memory-mapped shared region.
type MappedRegion struct {
	Addr []byte
	// platform-specific fields (fd, handle, etc.)
}

// MapOptions defines options for mapping shared memory.
type MapOptions struct {
	Name   string
	Size   int
	Create bool
}

// Function implementations are provided in platform-specific files (e.g., platform_linux.go, platform_windows.go).
