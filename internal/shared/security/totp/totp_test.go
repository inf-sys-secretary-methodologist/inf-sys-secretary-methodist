package totp_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/security/totp"
)

// rfc6238Secret is the ASCII string "12345678901234567890" (20 bytes), the
// reference secret from RFC 6238 Appendix B test vectors.
var rfc6238Secret = []byte("12345678901234567890")

// TestGenerate_RFC6238Vectors validates Generate against the public RFC 6238
// Appendix B test vectors, truncated from 8 digits to 6 (mod 10^6) since this
// implementation uses the RFC 4226 default 6-digit code length.
func TestGenerate_RFC6238Vectors(t *testing.T) {
	tests := []struct {
		name    string
		unixSec int64
		want    string
	}{
		{"T=59", 59, "287082"},                 // RFC vector 94287082 mod 1e6
		{"T=1111111109", 1111111109, "081804"}, // 07081804 mod 1e6
		{"T=1111111111", 1111111111, "050471"}, // 14050471 mod 1e6
		{"T=1234567890", 1234567890, "005924"}, // 89005924 mod 1e6
		{"T=2000000000", 2000000000, "279037"}, // 69279037 mod 1e6
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := totp.Generate(rfc6238Secret, time.Unix(tc.unixSec, 0))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Generate at T=%d: got %q want %q", tc.unixSec, got, tc.want)
			}
		})
	}

	// Sanity: hex-decoded RFC secret 31323334...3930 matches ASCII-derived secret.
	hexSecret, err := hex.DecodeString("3132333435363738393031323334353637383930")
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}
	got, err := totp.Generate(hexSecret, time.Unix(59, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "287082" {
		t.Errorf("Generate with hex-decoded RFC secret at T=59: got %q want 287082", got)
	}
}

// TestVerify covers exact-match, drift-window, out-of-window and malformed-code cases.
func TestVerify(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		unixSec    int64
		windowSize int
		want       bool
	}{
		{"matching code at exact time, window=0", "287082", 59, 0, true},
		{"wrong code at exact time", "999999", 59, 0, false},
		// At T=89 (step 2) with window=1, we check steps {1,2,3}; step 1 = "287082".
		{"previous-step code accepted within window=1", "287082", 89, 1, true},
		{"previous-step code rejected when window=0", "287082", 89, 0, false},
		{"empty code rejected", "", 59, 0, false},
		{"non-numeric code rejected", "abcdef", 59, 0, false},
		{"wrong length code rejected", "12345", 59, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := totp.Verify(rfc6238Secret, tc.code, time.Unix(tc.unixSec, 0), tc.windowSize)
			if got != tc.want {
				t.Errorf("Verify(%q at T=%d, window=%d): got %v want %v",
					tc.code, tc.unixSec, tc.windowSize, got, tc.want)
			}
		})
	}
}

// TestGenerateSecret enforces 160-bit length, non-empty Base32 encoding, and
// distinct values across calls (entropy sanity, not statistical).
func TestGenerateSecret(t *testing.T) {
	raw, b32, err := totp.GenerateSecret()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) != 20 {
		t.Errorf("expected 20-byte secret (RFC 6238 minimum), got %d bytes", len(raw))
	}
	if b32 == "" {
		t.Errorf("Base32 representation must be non-empty")
	}

	_, b32b, err := totp.GenerateSecret()
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if b32 == b32b {
		t.Errorf("GenerateSecret must produce unique secrets; got duplicate %q", b32)
	}
}
