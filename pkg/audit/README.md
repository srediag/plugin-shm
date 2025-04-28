# pkg/audit

This package will provide all governance, audit logging, and compliance logic for plugin operations.

## TODO: Migration from `pkg/plugin`

- [ ] Migrate and refactor audit log hooks from `session.go`, `protocol_manager.go`, `event_dispatcher.go`

## TODO: New Features/Interfaces

- [ ] Define audit event API and interfaces
- [ ] Add OTel audit event integration
- [ ] Add compliance and governance hooks
- [ ] Add pluggable audit log backends (file, syslog, cloud, etc.)
- [ ] Add audit event filtering and policy enforcement

## TODO: Migration from pkg/plugin

- Migrate and refactor:
  - Audit log hooks and event logging (if present)
  - Any governance-related logic in plugin main loop

## TODO: New Features

- Integrate OpenTelemetry for audit/governance events
- Provide governance and compliance API
- Add pluggable audit log backends (file, syslog, remote)
- Add compliance and policy enforcement hooks
- Write comprehensive unit and integration tests
