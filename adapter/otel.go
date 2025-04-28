// Package adapter provides adapters for plugin-shm integration with external systems.
package adapter

// OTelAdapter integrates OpenTelemetry metrics and tracing for plugins.
type OTelAdapter interface {
	RecordMetric(name string, value float64) error
	StartSpan(name string) (Span, error)
}

// Span is a minimal interface for OTel spans (example).
type Span interface {
	End()
}

// ExampleOTelAdapter is a sample implementation of OTelAdapter.
type ExampleOTelAdapter struct{}

// RecordMetric records a metric (example implementation).
func (a *ExampleOTelAdapter) RecordMetric(name string, value float64) error {
	// TODO: integrate with OpenTelemetry metrics
	return nil
}

// StartSpan starts a new span (example implementation).
func (a *ExampleOTelAdapter) StartSpan(name string) (Span, error) {
	// TODO: integrate with OpenTelemetry tracing
	return &exampleSpan{}, nil
}

type exampleSpan struct{}

func (s *exampleSpan) End() {}
