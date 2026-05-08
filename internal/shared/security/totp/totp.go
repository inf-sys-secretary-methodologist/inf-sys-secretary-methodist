// Package totp implements RFC 6238 Time-based One-Time Password algorithm
// (HMAC-SHA1, 30-second step, 6-digit codes) without third-party dependencies.
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- RFC 6238 mandates HMAC-SHA1 for TOTP interop
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	timeStepSeconds = 30
	codeDigits      = 6
	secretBytes     = 20 // 160-bit secret per RFC 6238 §5.1
)

// ErrEmptySecret is returned when Generate/Verify is called without a secret.
var ErrEmptySecret = errors.New("totp: empty secret")

// Generate returns the 6-digit TOTP code for secret at time t.
func Generate(secret []byte, t time.Time) (string, error) {
	return generateAtStep(secret, t.Unix()/timeStepSeconds)
}

// Verify reports whether code matches secret at time t within ±windowSize
// 30-second steps. windowSize=0 enforces exact-step match; positive values
// tolerate clock drift between client and server.
func Verify(secret []byte, code string, t time.Time, windowSize int) bool {
	if len(code) != codeDigits {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}

	currentStep := t.Unix() / timeStepSeconds
	for delta := -windowSize; delta <= windowSize; delta++ {
		want, err := generateAtStep(secret, currentStep+int64(delta))
		if err != nil {
			return false
		}
		if hmac.Equal([]byte(want), []byte(code)) {
			return true
		}
	}
	return false
}

// GenerateSecret returns 20 random bytes and their RFC 4648 Base32 (no
// padding) encoding suitable for embedding in otpauth:// URIs.
func GenerateSecret() ([]byte, string, error) {
	buf := make([]byte, secretBytes)
	if _, err := rand.Read(buf); err != nil {
		return nil, "", fmt.Errorf("totp: read random bytes: %w", err)
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	return buf, encoded, nil
}

// generateAtStep computes the TOTP code for an explicit time-step counter,
// shared between Generate and Verify so drift-window iteration stays cheap.
func generateAtStep(secret []byte, step int64) (string, error) {
	if len(secret) == 0 {
		return "", ErrEmptySecret
	}

	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))

	mac := hmac.New(sha1.New, secret)
	mac.Write(counter[:])
	digest := mac.Sum(nil)

	offset := digest[len(digest)-1] & 0x0F
	bin := (uint32(digest[offset]&0x7F) << 24) |
		(uint32(digest[offset+1]) << 16) |
		(uint32(digest[offset+2]) << 8) |
		uint32(digest[offset+3])

	mod := uint32(1)
	for range codeDigits {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", codeDigits, bin%mod), nil
}
