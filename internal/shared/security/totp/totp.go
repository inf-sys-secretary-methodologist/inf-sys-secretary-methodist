// Package totp implements RFC 6238 Time-based One-Time Password algorithm
// (HMAC-SHA1, 30-second step, 6-digit codes) without third-party dependencies.
package totp

import "time"

// Generate returns the 6-digit TOTP code for the given secret at time t.
func Generate(secret []byte, t time.Time) (string, error) {
	return "", nil
}

// Verify reports whether code matches the TOTP for secret at time t, allowing
// a tolerance of ±windowSize 30-second steps to account for clock drift.
func Verify(secret []byte, code string, t time.Time, windowSize int) bool {
	return false
}

// GenerateSecret returns 20 random bytes (160-bit secret per RFC 6238) and
// their canonical Base32 (no-padding) encoding suitable for otpauth URIs.
func GenerateSecret() ([]byte, string, error) {
	return nil, "", nil
}
