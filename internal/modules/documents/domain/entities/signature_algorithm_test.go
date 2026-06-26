package entities

import (
	"errors"
	"testing"
)

// TestSignatureAlgorithm_TableDriven pins the SignatureAlgorithm enum used by
// the cryptographic electronic-signature feature (#140):
//   - the single supported algorithm is ECDSA over P-256 with SHA-256 digest;
//   - IsValid/Validate accept it and reject anything else with the
//     ErrInvalidSignatureAlgorithm sentinel so the HTTP layer can map a stable
//     422 response;
//   - String round-trips the wire literal stored in the DB.
//
// Issue: #140
func TestSignatureAlgorithm_TableDriven(t *testing.T) {
	cases := []struct {
		name      string
		algo      SignatureAlgorithm
		wantValid bool
	}{
		{name: "ecdsa p256 sha256", algo: SignatureAlgorithmECDSAP256SHA256, wantValid: true},
		{name: "empty", algo: SignatureAlgorithm(""), wantValid: false},
		{name: "unknown", algo: SignatureAlgorithm("RSA_SHA1"), wantValid: false},
		{name: "wrong case", algo: SignatureAlgorithm("ecdsa_p256_sha256"), wantValid: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.algo.IsValid(); got != tc.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tc.wantValid)
			}
			err := tc.algo.Validate()
			if tc.wantValid {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			} else if !errors.Is(err, ErrInvalidSignatureAlgorithm) {
				t.Errorf("Validate() = %v, want ErrInvalidSignatureAlgorithm", err)
			}
		})
	}
}

func TestSignatureAlgorithm_String(t *testing.T) {
	if got := SignatureAlgorithmECDSAP256SHA256.String(); got != "ECDSA_P256_SHA256" {
		t.Errorf("String() = %q, want %q", got, "ECDSA_P256_SHA256")
	}
}
