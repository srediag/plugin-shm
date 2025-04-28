// Package adapter provides adapters for plugin-shm integration with external systems.
package adapter

// HealthAdapter integrates plugin health/liveness with external monitoring systems.
type HealthAdapter interface {
	ReportHealth(pluginID string, status string) error
}

// ExampleHealthAdapter is a sample implementation of HealthAdapter.
type ExampleHealthAdapter struct{}

// ReportHealth reports plugin health status (example implementation).
func (a *ExampleHealthAdapter) ReportHealth(pluginID string, status string) error {
	// TODO: integrate with external health monitoring
	return nil
}
