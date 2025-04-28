// Package api defines public API contracts for plugin-shm.
package api

// Lifecycle defines the interface for plugin lifecycle management.
type Lifecycle interface {
	StartPlugin(pluginID string) error
	StopPlugin(pluginID string) error
	ReloadPlugin(pluginID string) error
}

// ExampleLifecycle is a sample implementation of Lifecycle.
type ExampleLifecycle struct{}

func (l *ExampleLifecycle) StartPlugin(pluginID string) error {
	// TODO: implement start logic
	return nil
}
func (l *ExampleLifecycle) StopPlugin(pluginID string) error {
	// TODO: implement stop logic
	return nil
}
func (l *ExampleLifecycle) ReloadPlugin(pluginID string) error {
	// TODO: implement reload logic
	return nil
}
