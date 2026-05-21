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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

// gcmNonceSize is the standard 96-bit nonce GCM is parameterised for.
const gcmNonceSize = 12

// EncryptString seals plaintext under key (AES-256, 32-byte KEK) and
// returns base64.StdEncoding(nonce || ciphertext || tag). Returns
// ErrInvalidKEKLength when the key is the wrong length and propagates
// underlying AES / random failures (rare).
func EncryptString(plaintext string, key []byte) (string, error) {
	aead, err := gcmCipher(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("crypto: read nonce: %w", err)
	}
	sealed := aead.Seal(nil, nonce, []byte(plaintext), nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString reverses EncryptString. Returns ErrInvalidKEKLength
// when the key is the wrong length, ErrInvalidCiphertext for any decode
// / auth-tag failure (wrong key, tampered blob, truncated input).
func DecryptString(ciphertext string, key []byte) (string, error) {
	aead, err := gcmCipher(key)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	if len(raw) < gcmNonceSize+aead.Overhead() {
		return "", ErrInvalidCiphertext
	}
	nonce := raw[:gcmNonceSize]
	sealed := raw[gcmNonceSize:]
	plain, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidCiphertext, err)
	}
	return string(plain), nil
}

// gcmCipher constructs the AEAD from the key, shared by encrypt and
// decrypt — kept private because callers must use the higher-level
// EncryptString / DecryptString which also handle nonce + encoding.
func gcmCipher(key []byte) (cipher.AEAD, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKEKLength
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: aes.NewCipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: cipher.NewGCM: %w", err)
	}
	return aead, nil
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
