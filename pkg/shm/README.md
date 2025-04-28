# pkg/shm

This package provides a portable, lock-free shared memory buffer implementation for inter-process communication (IPC).

- Designed for use across Linux and Windows.
- Exposes a generic, advanced API for efficient, concurrent access.
- Instrumented with OpenTelemetry metrics and tracing (OTel Go SDK v1.30.0).

> **Note:** All code in this directory avoids naming collisions with Go standard library packages.
