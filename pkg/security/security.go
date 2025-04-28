// Package security provides interfaces and helpers for secure plugin loading, validation, and cryptographic operations.
package security

// SecurityProvider defines the interface for plugin security operations.
type SecurityProvider interface {
	// ValidateSignature checks the plugin's signature (e.g., SHA256, cosign).
	ValidateSignature(pluginPath string) error
	// SecureChannel negotiates a secure channel (TLS, mTLS).
	SecureChannel(peer string) error
	// EnforcePolicy applies security policies at admission/runtime.
	EnforcePolicy(pluginID string) error
}

// TODO: Migration Checklist
//   - Migrate and refactor config.go, config_test.go (plugin config, security options)
//   - Migrate and refactor errors.go (security-related errors)
//   - Migrate and refactor util.go, util_test.go (utility functions, secure random, etc.)
//   - Migrate and refactor debug.go, debug_test.go (debug/trace hooks, if security-relevant)
//
// TODO: New Features/Interfaces
//   - Define plugin signature validation (SHA256, cosign, etc.)
//   - Add secure channel negotiation (TLS, mTLS)
//   - Add OTel security event integration
//   - Add policy enforcement hooks (admission, runtime checks)
//   - Add plugin sandboxing and isolation interfaces
