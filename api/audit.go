// Package api defines public API contracts for plugin-shm.
package api

// Audit defines the interface for plugin audit and governance.
type Audit interface {
	LogEvent(event string, details map[string]interface{}) error
}

// ExampleAudit is a sample implementation of Audit.
type ExampleAudit struct{}

func (a *ExampleAudit) LogEvent(event string, details map[string]interface{}) error {
	// TODO: implement audit event logging
	return nil
}
