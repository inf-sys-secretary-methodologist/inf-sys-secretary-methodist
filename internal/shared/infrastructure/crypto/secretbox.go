// Package crypto provides at-rest encryption primitives used by the
// auth and other modules to wrap secrets before persisting them. The
// scheme is AES-256-GCM with a 96-bit nonce prepended to the ciphertext
// (RFC 5288 §3.2 nonce shape), the whole blob base64 (StdEncoding)
// encoded so it round-trips through TEXT columns without escaping.
// Issue #279 ADR-4.
//
// gcmNonceSize and the AEAD helper live in this package's GREEN pair
// (see EncryptString / DecryptString implementation); they are not
// exported.
package crypto

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// ErrEmptyKEK is returned when callers pass a nil / zero-length key.
var ErrEmptyKEK = errors.New("crypto: empty KEK")

// ErrInvalidKEKLength is returned when a key is not 32 bytes (AES-256).
var ErrInvalidKEKLength = errors.New("crypto: KEK must be 32 bytes (AES-256)")

// ErrInvalidCiphertext is returned for malformed input (decode / length
// failure) — the surface kept narrow so callers can map to a single
// "decrypt failed" outcome without leaking the specific failure mode.
var ErrInvalidCiphertext = errors.New("crypto: invalid ciphertext")

// EncryptString seals plaintext under key (AES-256, 32-byte KEK) and
// returns base64.StdEncoding(nonce || ciphertext || tag). Returns
// ErrInvalidKEKLength when the key is the wrong length and propagates
// underlying AES / random failures (rare). RED stub returns an error.
func EncryptString(_ string, _ []byte) (string, error) {
	return "", errors.New("crypto: not implemented")
}

// DecryptString reverses EncryptString. Returns ErrInvalidKEKLength
// when the key is the wrong length, ErrInvalidCiphertext for any decode
// / auth-tag failure (wrong key, tampered blob, truncated input). RED
// stub returns an error.
func DecryptString(_ string, _ []byte) (string, error) {
	return "", errors.New("crypto: not implemented")
}

// ParseKEKHex decodes a 64-hex-char string into a 32-byte AES-256 key.
// Empty input returns ErrEmptyKEK so callers can branch on "no KEK
// configured" vs "malformed KEK". Other parse failures return
// ErrInvalidKEKLength wrapped with the hex.DecodeString error so logs
// surface the cause without leaking the (sensitive) key material.
func ParseKEKHex(s string) ([]byte, error) {
	if s == "" {
		return nil, ErrEmptyKEK
	}
	raw, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidKEKLength, err)
	}
	if len(raw) != 32 {
		return nil, ErrInvalidKEKLength
	}
	return raw, nil
}
