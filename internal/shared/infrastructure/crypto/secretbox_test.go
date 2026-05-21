package crypto_test

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
)

// genKey returns a fresh 32-byte AES-256 key for the test case.
func genKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

// TestEncryptDecryptRoundTrip pins the v0.159.0 ADR-4 invariants:
// encrypt → decrypt under the same key recovers the plaintext byte
// for byte, decrypt under a different key fails, and tampering with
// the ciphertext fails the GCM auth tag check. Issue #279.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Run("round-trip preserves plaintext", func(t *testing.T) {
		key := genKey(t)
		plaintext := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP" // canonical Base32 MFA secret

		ct, err := crypto.EncryptString(plaintext, key)
		require.NoError(t, err)
		require.NotEqual(t, plaintext, ct, "ciphertext must not equal plaintext")

		recovered, err := crypto.DecryptString(ct, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, recovered)
	})

	t.Run("ciphertexts under same key + plaintext differ (fresh nonce)", func(t *testing.T) {
		key := genKey(t)
		plaintext := "stable-input"

		ct1, err := crypto.EncryptString(plaintext, key)
		require.NoError(t, err)
		ct2, err := crypto.EncryptString(plaintext, key)
		require.NoError(t, err)
		assert.NotEqual(t, ct1, ct2, "GCM nonce must be unique per call — repeated encryption must yield distinct ciphertexts")
	})

	t.Run("wrong key fails to decrypt", func(t *testing.T) {
		key1 := genKey(t)
		key2 := genKey(t)
		ct, err := crypto.EncryptString("plaintext", key1)
		require.NoError(t, err)

		_, err = crypto.DecryptString(ct, key2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, crypto.ErrInvalidCiphertext) || strings.Contains(err.Error(), "invalid"),
			"wrong-key decrypt must surface a recognizable error")
	})

	t.Run("tampered ciphertext fails auth tag check", func(t *testing.T) {
		key := genKey(t)
		ct, err := crypto.EncryptString("plaintext", key)
		require.NoError(t, err)

		// Flip a byte mid-blob to corrupt either nonce, ciphertext, or tag.
		corrupted := []byte(ct)
		corrupted[len(corrupted)/2] = corrupted[len(corrupted)/2] ^ 0x01

		_, err = crypto.DecryptString(string(corrupted), key)
		assert.Error(t, err, "corrupted ciphertext must not decrypt")
	})

	t.Run("encrypt with wrong-length key returns ErrInvalidKEKLength", func(t *testing.T) {
		shortKey := []byte("too-short")
		_, err := crypto.EncryptString("plaintext", shortKey)
		assert.ErrorIs(t, err, crypto.ErrInvalidKEKLength)
	})

	t.Run("decrypt with wrong-length key returns ErrInvalidKEKLength", func(t *testing.T) {
		shortKey := []byte("too-short")
		_, err := crypto.DecryptString("anything", shortKey)
		assert.ErrorIs(t, err, crypto.ErrInvalidKEKLength)
	})
}

// TestParseKEKHex pins the KEK env-spec parser: 64-hex-char input → 32
// bytes; empty input → ErrEmptyKEK so callers can branch on "KEK not
// configured" vs "KEK malformed". Issue #279 ADR-4.
func TestParseKEKHex(t *testing.T) {
	t.Run("64-char hex returns 32 bytes", func(t *testing.T) {
		raw := make([]byte, 32)
		_, err := rand.Read(raw)
		require.NoError(t, err)
		s := hex.EncodeToString(raw)

		got, err := crypto.ParseKEKHex(s)
		require.NoError(t, err)
		assert.Equal(t, raw, got)
	})

	t.Run("empty string returns ErrEmptyKEK", func(t *testing.T) {
		_, err := crypto.ParseKEKHex("")
		assert.ErrorIs(t, err, crypto.ErrEmptyKEK)
	})

	t.Run("non-hex string returns ErrInvalidKEKLength", func(t *testing.T) {
		_, err := crypto.ParseKEKHex("not-hex-at-all")
		assert.ErrorIs(t, err, crypto.ErrInvalidKEKLength)
	})

	t.Run("wrong length returns ErrInvalidKEKLength", func(t *testing.T) {
		_, err := crypto.ParseKEKHex(hex.EncodeToString([]byte{0x01, 0x02, 0x03}))
		assert.ErrorIs(t, err, crypto.ErrInvalidKEKLength)
	})
}
