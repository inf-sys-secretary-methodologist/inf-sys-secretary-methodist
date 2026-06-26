package usecases

import (
	"context"
	"encoding/hex"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// VerifySignatureUseCase re-checks a stored signature against the document's
// current state: it recomputes the canonical digest from the current body and
// verifies the ECDSA signature under the stored certificate.
type VerifySignatureUseCase struct {
	repo   SignatureRepository
	view   DocumentSigningView
	engine SignatureEngine
}

// NewVerifySignatureUseCase wires the use case.
func NewVerifySignatureUseCase(repo SignatureRepository, view DocumentSigningView, engine SignatureEngine) *VerifySignatureUseCase {
	if repo == nil || view == nil || engine == nil {
		panic("documents: NewVerifySignatureUseCase requires non-nil repo, view and engine")
	}
	return &VerifySignatureUseCase{repo: repo, view: view, engine: engine}
}

// Execute verifies the signature identified by signatureID.
func (uc *VerifySignatureUseCase) Execute(ctx context.Context, signatureID int64) (SignatureVerdict, error) {
	sig, err := uc.repo.GetByID(ctx, signatureID)
	if err != nil {
		return SignatureVerdict{}, err
	}

	version, contentHash, err := uc.view.GetForSigning(ctx, sig.DocumentID)
	if err != nil {
		return SignatureVerdict{}, err
	}

	verdict := SignatureVerdict{
		SignatureID:    sig.ID,
		VersionChanged: version != sig.DocumentVersion,
	}

	// Recompute the digest with the SIGNED version (stored on the signature)
	// and the document's CURRENT body hash. A body mutation flips DigestMatch.
	expected, err := entities.ComputeSigningDigest(sig.DocumentID, sig.DocumentVersion, sig.SignerID, sig.SignedAt.Unix(), contentHash)
	if err != nil {
		return SignatureVerdict{}, err
	}
	verdict.DigestMatch = expected == sig.DigestHex

	if verdict.DigestMatch {
		digest, derr := hex.DecodeString(sig.DigestHex)
		if derr != nil {
			return SignatureVerdict{}, derr
		}
		ok, verr := uc.engine.Verify(sig.CertificatePEM, digest, sig.SignatureDER)
		if verr != nil {
			return SignatureVerdict{}, verr
		}
		verdict.CryptoValid = ok
	}

	verdict.Valid = verdict.DigestMatch && verdict.CryptoValid
	return verdict, nil
}
