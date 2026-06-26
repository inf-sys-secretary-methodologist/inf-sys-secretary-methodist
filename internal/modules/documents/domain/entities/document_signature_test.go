package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// validDigestHex is a syntactically valid SHA-256 hex digest (64 lowercase hex
// chars) used across signature construction tests.
const validDigestHex = "ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12"

const fakeCertPEM = "-----BEGIN CERTIFICATE-----\nMIIBfakebytes\n-----END CERTIFICATE-----\n"

func validSignedAt() time.Time {
	return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
}

// TestNewDocumentSignature_HappyPath pins that a well-formed signature record
// constructs and exposes its fields verbatim.
//
// Issue: #140
func TestNewDocumentSignature_HappyPath(t *testing.T) {
	der := []byte{0x30, 0x44, 0x01, 0x02}
	sig, err := NewDocumentSignature(
		42,                                // documentID
		3,                                 // documentVersion
		7,                                 // signerID
		"  Иванов И.И.  ",                 // signerName (trimmed)
		SignatureAlgorithmECDSAP256SHA256, // algo
		validDigestHex,                    // digestHex
		der,                               // signatureDER
		fakeCertPEM,                       // certificatePEM
		validSignedAt(),                   // signedAt
	)
	if err != nil {
		t.Fatalf("NewDocumentSignature() = %v, want nil", err)
	}
	if sig.DocumentID != 42 || sig.DocumentVersion != 3 || sig.SignerID != 7 {
		t.Errorf("identity mismatch: %+v", sig)
	}
	if sig.SignerName != "Иванов И.И." {
		t.Errorf("SignerName = %q, want trimmed %q", sig.SignerName, "Иванов И.И.")
	}
	if sig.Algorithm != SignatureAlgorithmECDSAP256SHA256 {
		t.Errorf("Algorithm = %q", sig.Algorithm)
	}
	if sig.DigestHex != validDigestHex {
		t.Errorf("DigestHex = %q", sig.DigestHex)
	}
	if string(sig.SignatureDER) != string(der) {
		t.Errorf("SignatureDER not preserved")
	}
	if !sig.SignedAt.Equal(validSignedAt()) {
		t.Errorf("SignedAt = %v", sig.SignedAt)
	}
}

// TestNewDocumentSignature_Invalid pins every construction invariant. All
// failures collapse to ErrInvalidDocumentSignature except the unsupported
// algorithm, which surfaces ErrInvalidSignatureAlgorithm.
func TestNewDocumentSignature_Invalid(t *testing.T) {
	der := []byte{0x30, 0x44}
	cases := []struct {
		name    string
		docID   int64
		version int
		signer  int64
		sName   string
		algo    SignatureAlgorithm
		digest  string
		der     []byte
		cert    string
		at      time.Time
		wantErr error
	}{
		{name: "zero document id", docID: 0, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "zero version", docID: 1, version: 0, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "zero signer", docID: 1, version: 1, signer: 0, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "blank signer name", docID: 1, version: 1, signer: 7, sName: "   ", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "bad algorithm", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithm("RSA"), digest: validDigestHex, der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidSignatureAlgorithm},
		{name: "digest too short", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: "abcd", der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "digest uppercase", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: strings.ToUpper(validDigestHex), der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "digest non-hex", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: strings.Repeat("zz", 32), der: der, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "empty signature der", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: nil, cert: fakeCertPEM, at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "blank cert", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: "  ", at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "cert not pem", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: "not a certificate", at: validSignedAt(), wantErr: ErrInvalidDocumentSignature},
		{name: "zero signed at", docID: 1, version: 1, signer: 7, sName: "X", algo: SignatureAlgorithmECDSAP256SHA256, digest: validDigestHex, der: der, cert: fakeCertPEM, at: time.Time{}, wantErr: ErrInvalidDocumentSignature},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewDocumentSignature(tc.docID, tc.version, tc.signer, tc.sName, tc.algo, tc.digest, tc.der, tc.cert, tc.at)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// TestComputeSigningDigest pins the canonical signing-digest contract: it is the
// deterministic SHA-256 over (documentID, version, signerID, signedAt,
// content-hash). The result is a 64-char lowercase hex digest, it is stable for
// identical inputs, and it changes when ANY input changes (so a verifier can
// detect a swapped signer, version, or mutated document body).
func TestComputeSigningDigest(t *testing.T) {
	contentHash := validDigestHex
	base, err := ComputeSigningDigest(42, 3, 7, validSignedAt().Unix(), contentHash)
	if err != nil {
		t.Fatalf("ComputeSigningDigest() = %v, want nil", err)
	}
	if len(base) != 64 || strings.ToLower(base) != base {
		t.Errorf("digest = %q, want 64-char lowercase hex", base)
	}

	// Determinism.
	again, _ := ComputeSigningDigest(42, 3, 7, validSignedAt().Unix(), contentHash)
	if again != base {
		t.Errorf("non-deterministic: %q != %q", again, base)
	}

	// Sensitivity to each field.
	variants := map[string]string{}
	variants["doc"], _ = ComputeSigningDigest(43, 3, 7, validSignedAt().Unix(), contentHash)
	variants["ver"], _ = ComputeSigningDigest(42, 4, 7, validSignedAt().Unix(), contentHash)
	variants["signer"], _ = ComputeSigningDigest(42, 3, 8, validSignedAt().Unix(), contentHash)
	variants["ts"], _ = ComputeSigningDigest(42, 3, 7, validSignedAt().Add(time.Second).Unix(), contentHash)
	variants["content"], _ = ComputeSigningDigest(42, 3, 7, validSignedAt().Unix(), strings.Repeat("cd", 32))
	for field, d := range variants {
		if d == base {
			t.Errorf("digest not sensitive to %s change", field)
		}
	}
}

func TestComputeSigningDigest_InvalidContentHash(t *testing.T) {
	if _, err := ComputeSigningDigest(1, 1, 1, validSignedAt().Unix(), "deadbeef"); !errors.Is(err, ErrInvalidDocumentSignature) {
		t.Errorf("err = %v, want ErrInvalidDocumentSignature", err)
	}
}

// TestReconstituteDocumentSignature pins the persistence rehydration path:
// it assigns all fields verbatim without re-validating invariants.
func TestReconstituteDocumentSignature(t *testing.T) {
	created := validSignedAt().Add(time.Minute)
	sig := ReconstituteDocumentSignature(99, 42, 3, 7, "Иванов И.И.", SignatureAlgorithmECDSAP256SHA256, validDigestHex, []byte{0x30}, fakeCertPEM, validSignedAt(), created)
	if sig.ID != 99 || sig.DocumentID != 42 || sig.SignerID != 7 {
		t.Errorf("reconstitute identity mismatch: %+v", sig)
	}
	if !sig.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want %v", sig.CreatedAt, created)
	}
}
