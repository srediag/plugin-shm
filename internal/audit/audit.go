// Package audit contains internal helpers for plugin governance, audit logging, and compliance.
package audit

// AuditEventHelper provides helpers for audit event formatting and compliance.
type AuditEventHelper interface {
	FormatEvent(event string, details map[string]interface{}) (string, error)
	CheckCompliance(event string) error
}

// TODO: Migration Checklist
//   - Migrate audit event formatting from pkg/plugin/session.go, protocol_manager.go, event_dispatcher.go
//   - Migrate compliance and policy enforcement helpers
//
// TODO: New Helpers
//   - Provide test utilities and mocks for audit logging
//   - Add internal error handling and recovery helpers
