// Package transport provides interfaces and implementations for plugin communication transports (e.g., shared memory, TCP, QUIC).
package transport

// Transport defines the interface for all plugin communication transports.
type Transport interface {
	// Start the transport (e.g., listen, connect, etc.)
	Start() error
	// Stop the transport and clean up resources.
	Stop() error
	// Send data over the transport.
	Send(data []byte) error
	// Receive data from the transport.
	Receive() ([]byte, error)
}

// TODO: Migration Checklist
//   - Migrate and refactor stream.go (core stream transport logic)
//   - Migrate and refactor queue.go (queue-based transport)
//   - Migrate and refactor listener.go and net_listener.go (network listeners)
//   - Migrate and refactor protocol_manager.go, protocol_event.go, protocol_initializer.go (protocol negotiation and events)
//   - Migrate and refactor epoll_linux.go, event_dispatcher_linux.go, event_dispatcher_race_linux.go (Linux event/epoll dispatch)
//
// TODO: New Features/Interfaces
//   - Define a pluggable Transport interface (for SHM, TCP, QUIC, etc.)
//   - Add OpenTelemetry metrics/tracing for all transports
//   - Add QUIC transport support (future)
//   - Add secure channel support (TLS, mTLS)
//   - Add connection pooling and backpressure handling
//   - Add transport-level error handling and recovery
