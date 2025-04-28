// Package adapter provides adapters for plugin-shm integration with external systems.
package adapter

// NetworkAdapter provides TCP/QUIC transport integration for plugins.
type NetworkAdapter interface {
	Dial(address string) error
	Listen(address string) error
}

// ExampleNetworkAdapter is a sample implementation of NetworkAdapter.
type ExampleNetworkAdapter struct{}

// Dial connects to a remote address (example implementation).
func (a *ExampleNetworkAdapter) Dial(address string) error {
	// TODO: implement TCP/QUIC dial logic
	return nil
}

// Listen starts listening on an address (example implementation).
func (a *ExampleNetworkAdapter) Listen(address string) error {
	// TODO: implement TCP/QUIC listen logic
	return nil
}
