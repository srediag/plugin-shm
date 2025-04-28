# internal/transport

This directory contains internal helpers and utilities for plugin communication transports.

## TODO

- [ ] Implement low-level socket helpers (e.g., non-blocking I/O, platform-specific tuning)
- [ ] Add platform-specific optimizations for shared memory, TCP, and QUIC
- [ ] Provide test utilities and mocks for transport layer
- [ ] Add internal error handling and recovery helpers

## TODO: Migration from pkg/plugin

- Migrate and refactor:
  - Low-level transport helpers from `pkg/plugin` (if any)
  - Platform-specific event loop or transport code

## TODO: New Helpers

- Add optimized buffer management for transports
- Implement platform-specific event loops and optimizations
- Provide internal hooks for transport diagnostics and debugging
