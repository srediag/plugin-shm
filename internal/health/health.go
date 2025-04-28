// Package health contains internal helpers for plugin heartbeat, liveness, and health monitoring.
package health

// HealthCheckHelper provides routines for health checks and liveness probes.
type HealthCheckHelper interface {
	CheckHeartbeat(pluginID string) error
	CheckLiveness(pluginID string) (bool, error)
}

// TODO: Migration Checklist
//   - Migrate health check routines from pkg/plugin/session.go, session_manager.go
//   - Migrate platform-specific health/liveness integrations
//
// TODO: New Helpers
//   - Provide test utilities and mocks for health monitoring
//   - Add internal error handling and recovery helpers
