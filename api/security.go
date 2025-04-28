// Package api defines public API contracts for plugin-shm.
package api

// Security defines the interface for plugin security operations.
type Security interface {
	ValidateSignature(pluginPath string) error
	SecureChannel(peer string) error
}

// ExampleSecurity is a sample implementation of Security.
type ExampleSecurity struct{}

func (s *ExampleSecurity) ValidateSignature(pluginPath string) error {
	// TODO: implement signature validation
	return nil
}
func (s *ExampleSecurity) SecureChannel(peer string) error {
	// TODO: implement secure channel negotiation
	return nil
}
