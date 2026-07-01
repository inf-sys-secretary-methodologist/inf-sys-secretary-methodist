package usecases

import (
	"context"
)

// SignatureEngine signs digests with per-user keys and verifies signatures.
// Declared here (consumer package) per DIP; implemented by
// infrastructure/signing.Service.
//
// Issue: #140
type SignatureEngine interface {
	// SignDigest signs a 32-byte digest with the signer's key (creating it on
	// first use) and returns the ASN.1 DER signature + the signer cert (PEM).
	SignDigest(ctx context.Context, userID int64, subjectCommonName string, digest []byte) (signatureDER []byte, certPEM string, err error)
	// Verify reports whether signatureDER verifies over digest under certPEM.
	Verify(certPEM string, digest []byte, signatureDER []byte) (bool, error)
}

// DocumentSigningView exposes the minimum a signer/verifier needs about a
// document: its current version and a SHA-256 hex of its canonical body (file
// bytes if present, else text content). Implemented by an adapter over the
// document repository + object storage (PR5). Returns
// ErrDocumentNotFound when the document does not exist.
type DocumentSigningView interface {
	GetForSigning(ctx context.Context, documentID int64) (version int, contentSHA256Hex string, err error)
}

// SignatureVerdict is the outcome of verifying a stored signature against the
// document's current state.
type SignatureVerdict struct {
	SignatureID int64
	// Valid is the overall verdict: DigestMatch && CryptoValid.
	Valid bool
	// DigestMatch is true when the digest recomputed from the document's
	// current body equals the digest that was signed (body unchanged).
	DigestMatch bool
	// CryptoValid is true when the ECDSA signature verifies under the signer's
	// certificate (only evaluated when DigestMatch holds).
	CryptoValid bool
	// VersionChanged is informational: the document's current version differs
	// from the version that was signed.
	VersionChanged bool
}
