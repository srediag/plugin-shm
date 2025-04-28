// Package shm provides a portable, lock-free shared memory buffer for inter-process communication (IPC).
//
// This package exposes a generic, advanced API for concurrent, lock-free access to a shared memory ring buffer.
// It is portable across Linux and Windows, and is instrumented with OpenTelemetry metrics and tracing (OTel Go SDK v1.30.0).
//
// Example usage:
//
//	cfg := shm.Config{
//	  Name: "myshm",
//	  Size: 65536,
//	  Meter: myMeter,
//	  Tracer: myTracer,
//	}
//	buf, err := shm.NewBuffer(cfg)
//	// ...
//
// See README.md for details.
package shm
