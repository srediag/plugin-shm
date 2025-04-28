// Package api defines public API contracts for plugin-shm.
package api

// Health defines the interface for plugin health and liveness.
type Health interface {
	Heartbeat(pluginID string) error
	LivenessCheck(pluginID string) (bool, error)
}

// ExampleHealth is a sample implementation of Health.
type ExampleHealth struct{}

func (h *ExampleHealth) Heartbeat(pluginID string) error {
	// TODO: implement heartbeat logic
	return nil
}
func (h *ExampleHealth) LivenessCheck(pluginID string) (bool, error) {
	// TODO: implement liveness check
	return true, nil
}
