// Package adapter provides adapters for plugin-shm integration with external systems.
package adapter

// HotReloadAdapter handles plugin hot-reload and orchestration integration.
type HotReloadAdapter interface {
	ReloadPlugin(pluginID string) error
}

// ExampleHotReloadAdapter is a sample implementation of HotReloadAdapter.
type ExampleHotReloadAdapter struct{}

// ReloadPlugin reloads the plugin (example implementation).
func (a *ExampleHotReloadAdapter) ReloadPlugin(pluginID string) error {
	// TODO: implement real hot-reload logic
	return nil
}
