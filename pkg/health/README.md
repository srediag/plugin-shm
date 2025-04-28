# pkg/health

This package provides interfaces and helpers for plugin heartbeat, liveness, and health monitoring.

## TODO: Migration from pkg/plugin

- Migrate and refactor:
  - `stats.go` (plugin statistics and health)
  - Heartbeat/liveness logic from plugin main loop (if present)

## TODO: New Features

- Integrate OpenTelemetry health/liveness metrics
- Add liveness and readiness probe interfaces
- Provide health check API for plugins and host
- Add alerting and notification hooks
- Write comprehensive unit and integration tests
