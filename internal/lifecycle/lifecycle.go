// Package lifecycle contains internal helpers for plugin lifecycle management, hot-reload, and orchestration.
package lifecycle

// StateManager provides helpers for plugin state management and persistence.
type StateManager interface {
	SaveState(pluginID string, state string) error
	LoadState(pluginID string) (string, error)
}

// TODO: Migration Checklist
//   - Migrate plugin state management from pkg/plugin/session.go, session_manager.go
//   - Migrate hot-reload routines and orchestration integration
//
// TODO: New Helpers
//   - Provide test utilities and mocks for lifecycle management
//   - Add internal error handling and recovery helpers
