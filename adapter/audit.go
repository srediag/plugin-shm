// Package adapter provides adapters for plugin-shm integration with external systems.
package adapter

// AuditAdapter sends audit events to external loggers or compliance systems.
type AuditAdapter interface {
	LogEvent(event string, details map[string]interface{}) error
}

// ExampleAuditAdapter is a sample implementation of AuditAdapter.
type ExampleAuditAdapter struct{}

// LogEvent logs an audit event (example implementation).
func (a *ExampleAuditAdapter) LogEvent(event string, details map[string]interface{}) error {
	// TODO: send event to external audit system
	return nil
}
