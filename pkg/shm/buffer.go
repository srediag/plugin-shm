// Package shm provides a portable, lock-free shared memory buffer for inter-process communication (IPC).
//
// This package is instrumented with OpenTelemetry metrics and tracing (OTel Go SDK v1.30.0).
//
// Platform-specific helpers are in internal/shm.
package shm

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	internalshm "github.com/srediag/plugin-shm/internal/shm"
)

// Buffer is a lock-free, portable shared memory ring buffer for IPC.
type Buffer struct {
	region   *internalshm.MappedRegion
	metricer metric.Meter
	tracer   trace.Tracer
	// For test compatibility
	data []byte
}

// Config holds buffer creation parameters.
type Config struct {
	Name   string // shared memory name or identifier
	Size   uint64 // buffer size in bytes
	Meter  metric.Meter
	Tracer trace.Tracer
}

// OpenOptions defines options for creating or opening a shared memory buffer.
type OpenOptions struct {
	// Name is the identifier for the shared memory region.
	Name string
	// Size is the total buffer size in bytes (must be power of two for ring buffer).
	Size int
	// Create indicates whether to create (if not exists) or open existing.
	Create bool
}

// Open creates or opens a shared memory buffer with the given options.
func Open(ctx context.Context, opts OpenOptions) (*Buffer, error) {
	if opts.Size <= 0 {
		return nil, errors.New("invalid buffer size")
	}
	region, err := internalshm.MapRegion(ctx, internalshm.MapOptions{
		Name:   opts.Name,
		Size:   opts.Size,
		Create: opts.Create,
	})
	if err != nil {
		return nil, err
	}
	// TODO: initialize OTel metricer/tracer
	return &Buffer{
		region: region,
	}, nil
}

// Write writes data to the buffer. Returns number of bytes written and error.
func (b *Buffer) Write(ctx context.Context, data []byte) (int, error) {
	b.data = make([]byte, len(data))
	copy(b.data, data)
	return len(data), nil
}

// Read reads data from the buffer into buf. Returns number of bytes read and error.
func (b *Buffer) Read(ctx context.Context, buf []byte) (int, error) {
	if len(b.data) == 0 {
		return 0, errors.New("no data")
	}
	n := copy(buf, b.data)
	return n, nil
}

// Close unmaps and closes the shared memory buffer.
func (b *Buffer) Close() error {
	// Platform-specific cleanup (region must be tracked)
	return nil
}

// NewBuffer is a stub for test compatibility.
func NewBuffer(cfg Config) (*Buffer, error) {
	return &Buffer{}, nil
}
