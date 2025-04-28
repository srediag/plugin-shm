// Package api defines public API contracts for plugin-shm.
package api

// Transport defines the interface for plugin communication transports.
type Transport interface {
	Send(data []byte) error
	Receive() ([]byte, error)
}

// ExampleTransport is a sample implementation of Transport.
type ExampleTransport struct{}

func (t *ExampleTransport) Send(data []byte) error {
	// TODO: implement send logic
	return nil
}
func (t *ExampleTransport) Receive() ([]byte, error) {
	// TODO: implement receive logic
	return nil, nil
}
