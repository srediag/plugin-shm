# pkg/transport

This package will provide all transport-layer logic for plugin communication, including shared memory, TCP, and future QUIC support.

## TODO: Migration from `pkg/plugin`

- [ ] Migrate and refactor `stream.go` (core stream transport logic)
- [ ] Migrate and refactor `queue.go` (queue-based transport)
- [ ] Migrate and refactor `listener.go` and `net_listener.go` (network listeners)
- [ ] Migrate and refactor `protocol_manager.go`, `protocol_event.go`, `protocol_initializer.go` (protocol negotiation and events)
- [ ] Migrate and refactor `epoll_linux.go`, `event_dispatcher_linux.go`, `event_dispatcher_race_linux.go` (Linux event/epoll dispatch)

## TODO: New Features/Interfaces

- [ ] Define a pluggable `Transport` interface (for SHM, TCP, QUIC, etc.)
- [ ] Add OpenTelemetry metrics/tracing for all transports
- [ ] Add QUIC transport support (future)
- [ ] Add secure channel support (TLS, mTLS)
- [ ] Add connection pooling and backpressure handling
- [ ] Add transport-level error handling and recovery
