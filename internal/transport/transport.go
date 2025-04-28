// Package transport contains internal helpers for plugin communication transports.
package transport

// SocketHelper provides low-level socket operations for transports.
type SocketHelper interface {
	SetNonblock(fd int) error
	SetSocketOptions(fd int) error
}

// TODO: Migration Checklist
//   - Migrate low-level transport helpers from pkg/plugin (if any)
//   - Migrate platform-specific event loop or transport code
//
// TODO: New Helpers
//   - Add optimized buffer management for transports
//   - Implement platform-specific event loops and optimizations
//   - Provide internal hooks for transport diagnostics and debugging
