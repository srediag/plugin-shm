// Package security contains internal helpers for secure plugin loading and cryptographic operations.
package security

// CryptoHelper provides cryptographic primitives for plugin security.
type CryptoHelper interface {
	Hash(data []byte) ([]byte, error)
	VerifySignature(data, sig []byte) error
}

// TODO: Migration Checklist
//   - Migrate cryptographic helpers from pkg/plugin/util.go, etc.
//   - Migrate secure random and platform-specific security features
//
// TODO: New Helpers
//   - Add secure random number generation utilities
//   - Integrate platform-specific security features (e.g., seccomp, sandbox)
//   - Provide test utilities and mocks for security features
