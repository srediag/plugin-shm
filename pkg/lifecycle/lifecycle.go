// Package lifecycle provides interfaces and helpers for plugin lifecycle management, hot-reload, and orchestration.
package lifecycle

// LifecycleManager defines the interface for plugin lifecycle management.
type LifecycleManager interface {
	// StartPlugin starts a plugin instance.
	StartPlugin(pluginID string) error
	// StopPlugin stops a plugin instance.
	StopPlugin(pluginID string) error
	// ReloadPlugin reloads a plugin instance (hot-reload).
	ReloadPlugin(pluginID string) error
	// GetState returns the current state of a plugin.
	GetState(pluginID string) (string, error)
}

// TODO: Migration Checklist
//   - Migrate and refactor lifecycle logic from session.go, session_manager.go
//   - Migrate and refactor hot-reload logic
//   - Migrate and refactor block_io.go (if lifecycle-relevant)
//
// TODO: New Features/Interfaces
//   - Define lifecycle API and interfaces (start, stop, reload, etc.)
//   - Add hot-reload hooks and orchestration integration
//   - Add plugin state management and persistence
//   - Add OTel lifecycle event integration
//   - Add lifecycle event hooks for external orchestration
