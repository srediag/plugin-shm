// Package health provides interfaces and helpers for plugin heartbeat, liveness, and health monitoring.
package health

// HealthProvider defines the interface for plugin health and liveness monitoring.
type HealthProvider interface {
	// Heartbeat records a plugin heartbeat event.
	Heartbeat(pluginID string) error
	// LivenessCheck checks if a plugin is alive.
	LivenessCheck(pluginID string) (bool, error)
	// ReportHealth reports plugin health status.
	ReportHealth(pluginID string, status string) error
}

// TODO: Migration Checklist
//   - Migrate and refactor stats.go (plugin stats, health metrics)
//   - Migrate and refactor heartbeat/liveness logic from session.go, session_manager.go
//
// TODO: New Features/Interfaces
//   - Define health/liveness API for plugins
//   - Add OTel health/liveness metrics and tracing
//   - Add plugin health/liveness callback interfaces
//   - Integrate with orchestration/monitoring systems
//   - Add health check CLI and test utilities
