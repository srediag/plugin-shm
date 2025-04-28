// Package api defines public API contracts for plugin-shm.
package api

// Plugin defines the interface for a plugin instance.
type Plugin interface {
	Start() error
	Stop() error
	Reload() error
}

// ExamplePlugin is a sample implementation of Plugin.
type ExamplePlugin struct{}

func (p *ExamplePlugin) Start() error {
	// TODO: implement plugin start logic
	return nil
}
func (p *ExamplePlugin) Stop() error {
	// TODO: implement plugin stop logic
	return nil
}
func (p *ExamplePlugin) Reload() error {
	// TODO: implement plugin reload logic
	return nil
}
