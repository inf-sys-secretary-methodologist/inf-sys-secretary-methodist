package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// SignatureRepositoryPG implements usecases.SignatureRepository using PostgreSQL.
type SignatureRepositoryPG struct {
	db *sql.DB
}

// NewSignatureRepositoryPG creates a new PostgreSQL signature repository.
func NewSignatureRepositoryPG(db *sql.DB) *SignatureRepositoryPG {
	return &SignatureRepositoryPG{db: db}
}

// sigColumns is the ordered column list shared by the read queries.
const sigColumns = `id, document_id, document_version, signer_id, signer_name,
		signature_algorithm, digest_hex, signature_der, certificate_pem,
		signed_at, created_at`

// Save persists a new signature and returns its generated id.
func (r *SignatureRepositoryPG) Save(ctx context.Context, sig *entities.DocumentSignature) (int64, error) {
	const q = `INSERT INTO document_signatures (
			document_id, document_version, signer_id, signer_name,
			signature_algorithm, digest_hex, signature_der, certificate_pem,
			signed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, q,
		sig.DocumentID,
		sig.DocumentVersion,
		sig.SignerID,
		sig.SignerName,
		string(sig.Algorithm),
		sig.DigestHex,
		sig.SignatureDER,
		sig.CertificatePEM,
		sig.SignedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("document signatures: save: %w", err)
	}
	return id, nil
}

// ListByDocument returns all signatures for a document, oldest first.
func (r *SignatureRepositoryPG) ListByDocument(ctx context.Context, documentID int64) ([]*entities.DocumentSignature, error) {
	q := `SELECT ` + sigColumns + `
		FROM document_signatures
		WHERE document_id = $1
		ORDER BY signed_at ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, q, documentID)
	if err != nil {
		return nil, fmt.Errorf("document signatures: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.DocumentSignature
	for rows.Next() {
		sig, err := scanSignature(rows)
		if err != nil {
			return nil, fmt.Errorf("document signatures: list: scan: %w", err)
		}
		out = append(out, sig)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("document signatures: list: rows: %w", err)
	}
	return out, nil
}

// GetByID returns a single signature or repositories.ErrSignatureNotFound.
func (r *SignatureRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DocumentSignature, error) {
	q := `SELECT ` + sigColumns + `
		FROM document_signatures
		WHERE id = $1`

	sig, err := scanSignature(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrSignatureNotFound
		}
		return nil, fmt.Errorf("document signatures: get: %w", err)
	}
	return sig, nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanSignature maps one row into a reconstituted domain signature.
func scanSignature(s rowScanner) (*entities.DocumentSignature, error) {
	var (
		id, documentID, signerID int64
		documentVersion          int
		signerName, algorithm    string
		digestHex, certPEM       string
		signatureDER             []byte
		signedAt, createdAt      time.Time
	)
	if err := s.Scan(
		&id, &documentID, &documentVersion, &signerID, &signerName,
		&algorithm, &digestHex, &signatureDER, &certPEM,
		&signedAt, &createdAt,
	); err != nil {
		return nil, err
	}
	return entities.ReconstituteDocumentSignature(
		id, documentID, documentVersion, signerID, signerName,
		entities.SignatureAlgorithm(algorithm), digestHex, signatureDER, certPEM,
		signedAt, createdAt,
	), nil
}
