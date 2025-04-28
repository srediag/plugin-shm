// Package audit provides interfaces and helpers for plugin governance, audit logging, and compliance.
package audit

// AuditLogger defines the interface for plugin audit logging and governance.
type AuditLogger interface {
	// LogEvent records an audit event.
	LogEvent(event string, details map[string]interface{}) error
	// SetCompliancePolicy sets the compliance policy for audit logging.
	SetCompliancePolicy(policy string) error
}

// TODO: Migration Checklist
//   - Migrate and refactor audit log hooks from session.go, protocol_manager.go, event_dispatcher.go
//
// TODO: New Features/Interfaces
//   - Define audit event API and interfaces
//   - Add OTel audit event integration
//   - Add compliance and governance hooks
//   - Add pluggable audit log backends (file, syslog, cloud, etc.)
//   - Add audit event filtering and policy enforcement
