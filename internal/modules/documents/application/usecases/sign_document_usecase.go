package usecases

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// roleStudent is the role string denied from signing documents.
const roleStudent = "student"

// signatureAuditResource is the audit resource tag for signature events.
const signatureAuditResource = "document_signature"

// SignDocumentUseCase applies a cryptographic signature to a document: it
// computes the canonical signing digest, has the engine sign it with the
// signer's key, and persists the resulting signature.
type SignDocumentUseCase struct {
	repo   SignatureRepository
	view   DocumentSigningView
	engine SignatureEngine
	audit  AuditSink
	now    func() time.Time
}

// NewSignDocumentUseCase wires the use case. audit may be nil.
func NewSignDocumentUseCase(repo SignatureRepository, view DocumentSigningView, engine SignatureEngine, audit AuditSink, now func() time.Time) *SignDocumentUseCase {
	if repo == nil || view == nil || engine == nil || now == nil {
		panic("documents: NewSignDocumentUseCase requires non-nil repo, view, engine and clock")
	}
	return &SignDocumentUseCase{repo: repo, view: view, engine: engine, audit: audit, now: now}
}

// Execute signs documentID on behalf of the signer. Students are denied.
func (uc *SignDocumentUseCase) Execute(ctx context.Context, documentID, signerID int64, signerName, signerRole string) (*entities.DocumentSignature, error) {
	if signerRole == roleStudent {
		uc.logEvent(ctx, "document_sign_denied", signerID, documentID)
		return nil, entities.ErrDocumentEditDenied
	}

	version, contentHash, err := uc.view.GetForSigning(ctx, documentID)
	if err != nil {
		return nil, err
	}

	signedAt := uc.now().Truncate(time.Second)
	digestHex, err := entities.ComputeSigningDigest(documentID, version, signerID, signedAt.Unix(), contentHash)
	if err != nil {
		return nil, err
	}
	digest, err := hex.DecodeString(digestHex)
	if err != nil {
		return nil, err
	}

	der, certPEM, err := uc.engine.SignDigest(ctx, signerID, signerName, digest)
	if err != nil {
		return nil, err
	}

	sig, err := entities.NewDocumentSignature(
		documentID, version, signerID, signerName,
		entities.SignatureAlgorithmECDSAP256SHA256,
		digestHex, der, certPEM, signedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := uc.repo.Save(ctx, sig)
	if err != nil {
		return nil, err
	}
	sig.ID = id

	uc.logEvent(ctx, "document_signed", signerID, documentID)
	return sig, nil
}

func (uc *SignDocumentUseCase) logEvent(ctx context.Context, action string, signerID, documentID int64) {
	if uc.audit == nil {
		return
	}
	uc.audit.LogAuditEvent(ctx, action, signatureAuditResource, map[string]any{
		"actor_user_id": signerID,
		"document_id":   documentID,
	})
}
