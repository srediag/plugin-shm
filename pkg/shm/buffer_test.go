package shm

import (
	"context"
	"testing"
)

type nopMeter struct{}
type nopTracer struct{}

func TestBufferAPI(t *testing.T) {
	cfg := Config{
		Name:   "testshm",
		Size:   4096,
		Meter:  nil, // TODO: use real or mock OTel meter
		Tracer: nil, // TODO: use real or mock OTel tracer
	}
	buf, err := NewBuffer(cfg)
	if err == nil {
		defer buf.Close()
	}
	if err != nil {
		t.Skipf("platform not implemented: %v", err)
	}
	ctx := context.Background()
	wdata := []byte("hello world")
	n, err := buf.Write(ctx, wdata)
	if err != nil && n == 0 {
		t.Errorf("Write failed: %v", err)
	}
	rbuf := make([]byte, len(wdata))
	n, err = buf.Read(ctx, rbuf)
	if err != nil && n == 0 {
		t.Errorf("Read failed: %v", err)
	}
	if string(rbuf) != string(wdata) {
		t.Errorf("Read data mismatch: got %q, want %q", rbuf, wdata)
	}
}
