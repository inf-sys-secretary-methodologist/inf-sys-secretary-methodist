package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidDocumentSignature signals an attempt to construct a DocumentSignature
// whose fields violate a domain invariant (bad identity, malformed digest,
// empty signature/certificate, ...). Sentinel so the HTTP layer can errors.Is
// it and map to a stable 422 response.
//
// Issue: #140
var ErrInvalidDocumentSignature = errors.New("document: invalid document signature")

// signingDigestDomain is a versioned prefix mixed into the canonical signing
// payload. Bumping it deliberately invalidates every prior signature's digest
// derivation, should the canonical form ever need to change.
const signingDigestDomain = "docsig.v1"

// sha256HexLen is the length of a SHA-256 digest rendered as lowercase hex.
const sha256HexLen = 64

// DocumentSignature is a cryptographic electronic signature applied to a
// specific version of a document by a specific signer. It records the signed
// digest, the raw ECDSA signature (DER), and the signer's X.509 certificate so
// the signature can later be verified independently of any server-held key.
//
// A document may carry several signatures (multiple signers, or re-signing a
// new version). DocumentVersion is captured so verification can detect that the
// underlying document changed after signing.
type DocumentSignature struct {
	ID              int64
	DocumentID      int64
	DocumentVersion int
	SignerID        int64
	SignerName      string
	Algorithm       SignatureAlgorithm
	// DigestHex is the lowercase-hex SHA-256 digest that was actually signed —
	// the output of ComputeSigningDigest over the canonical payload.
	DigestHex string
	// SignatureDER is the raw ECDSA signature in ASN.1 DER form.
	SignatureDER []byte
	// CertificatePEM is the signer's self-signed X.509 certificate (PEM) whose
	// public key verifies SignatureDER.
	CertificatePEM string
	SignedAt       time.Time
	CreatedAt      time.Time
}

const certificatePEMPrefix = "-----BEGIN CERTIFICATE-----"

// NewDocumentSignature constructs a validated signature record. signerName is
// trimmed. All invariant violations collapse to ErrInvalidDocumentSignature,
// except an unsupported algorithm which surfaces ErrInvalidSignatureAlgorithm.
func NewDocumentSignature(
	documentID int64,
	documentVersion int,
	signerID int64,
	signerName string,
	algo SignatureAlgorithm,
	digestHex string,
	signatureDER []byte,
	certificatePEM string,
	signedAt time.Time,
) (*DocumentSignature, error) {
	if documentID <= 0 || documentVersion < 1 || signerID <= 0 {
		return nil, fmt.Errorf("%w: identity", ErrInvalidDocumentSignature)
	}
	name := strings.TrimSpace(signerName)
	if name == "" {
		return nil, fmt.Errorf("%w: signer name", ErrInvalidDocumentSignature)
	}
	if err := algo.Validate(); err != nil {
		return nil, err
	}
	if !isSHA256Hex(digestHex) {
		return nil, fmt.Errorf("%w: digest must be 64-char lowercase hex", ErrInvalidDocumentSignature)
	}
	if len(signatureDER) == 0 {
		return nil, fmt.Errorf("%w: empty signature", ErrInvalidDocumentSignature)
	}
	cert := strings.TrimSpace(certificatePEM)
	if !strings.HasPrefix(cert, certificatePEMPrefix) {
		return nil, fmt.Errorf("%w: certificate must be PEM", ErrInvalidDocumentSignature)
	}
	if signedAt.IsZero() {
		return nil, fmt.Errorf("%w: signed-at timestamp", ErrInvalidDocumentSignature)
	}

	return &DocumentSignature{
		DocumentID:      documentID,
		DocumentVersion: documentVersion,
		SignerID:        signerID,
		SignerName:      name,
		Algorithm:       algo,
		DigestHex:       digestHex,
		SignatureDER:    signatureDER,
		CertificatePEM:  cert,
		SignedAt:        signedAt,
	}, nil
}

// ReconstituteDocumentSignature rehydrates a signature from persistence without
// re-validating invariants (the row was valid when it was written).
func ReconstituteDocumentSignature(
	id, documentID int64,
	documentVersion int,
	signerID int64,
	signerName string,
	algo SignatureAlgorithm,
	digestHex string,
	signatureDER []byte,
	certificatePEM string,
	signedAt, createdAt time.Time,
) *DocumentSignature {
	return &DocumentSignature{
		ID:              id,
		DocumentID:      documentID,
		DocumentVersion: documentVersion,
		SignerID:        signerID,
		SignerName:      signerName,
		Algorithm:       algo,
		DigestHex:       digestHex,
		SignatureDER:    signatureDER,
		CertificatePEM:  certificatePEM,
		SignedAt:        signedAt,
		CreatedAt:       createdAt,
	}
}

// ComputeSigningDigest is the canonical contract for WHAT a signature commits
// to. It binds the document identity (id + version), the signer, the signing
// timestamp, and a hash of the document body into a single deterministic
// SHA-256 digest (returned as lowercase hex). contentSHA256Hex is the hash of
// the document's file bytes (or text content); it must already be a 64-char
// lowercase hex digest. Any change to any input yields a different digest, so a
// verifier recomputing it can detect a swapped signer, a bumped version, or a
// mutated document body.
//
// signedAtUnix is the signing time in WHOLE SECONDS (time.Time.Unix(), never
// UnixNano): a verifier MUST recompute the digest with SignedAt.Unix() to match.
//
// The identity invariants mirror NewDocumentSignature exactly (documentID>=1,
// documentVersion>=1, signerID>=1) so the two domain functions never disagree
// on what a valid signature identity is.
func ComputeSigningDigest(documentID int64, documentVersion int, signerID int64, signedAtUnix int64, contentSHA256Hex string) (string, error) {
	if documentID <= 0 || documentVersion < 1 || signerID <= 0 {
		return "", fmt.Errorf("%w: identity", ErrInvalidDocumentSignature)
	}
	if !isSHA256Hex(contentSHA256Hex) {
		return "", fmt.Errorf("%w: content hash must be 64-char lowercase hex", ErrInvalidDocumentSignature)
	}
	canonical := fmt.Sprintf("%s|doc=%d|ver=%d|signer=%d|ts=%d|content=%s",
		signingDigestDomain, documentID, documentVersion, signerID, signedAtUnix, contentSHA256Hex)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:]), nil
}

// isSHA256Hex reports whether s is exactly 64 lowercase hexadecimal characters.
func isSHA256Hex(s string) bool {
	if len(s) != sha256HexLen {
		return false
	}
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f':
		default:
			return false
		}
	}
	return true
}
