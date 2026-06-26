package entities

import "errors"

// ErrInvalidSignatureAlgorithm signals an attempt to use a signature algorithm
// the system does not support. Sentinel so the HTTP layer can errors.Is it and
// map to a stable 422 response.
//
// Issue: #140
var ErrInvalidSignatureAlgorithm = errors.New("document: unsupported signature algorithm")

// SignatureAlgorithm enumerates the cryptographic schemes the document
// e-signature feature supports. The literal is stored verbatim in the DB
// (document_signatures.signature_algorithm), so the wire form must stay stable.
type SignatureAlgorithm string

const (
	// SignatureAlgorithmECDSAP256SHA256 is ECDSA over the NIST P-256 curve with
	// a SHA-256 message digest — the single supported scheme (ADR #140).
	SignatureAlgorithmECDSAP256SHA256 SignatureAlgorithm = "ECDSA_P256_SHA256"
)

// IsValid reports whether a is a supported signature algorithm.
func (a SignatureAlgorithm) IsValid() bool {
	return a == SignatureAlgorithmECDSAP256SHA256
}

// Validate returns ErrInvalidSignatureAlgorithm when a is not supported.
func (a SignatureAlgorithm) Validate() error {
	if !a.IsValid() {
		return ErrInvalidSignatureAlgorithm
	}
	return nil
}

// String returns the wire literal of the algorithm.
func (a SignatureAlgorithm) String() string {
	return string(a)
}
