package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

const sigDigestHex = "ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12"
const sigCertPEM = "-----BEGIN CERTIFICATE-----\nMIIBfake\n-----END CERTIFICATE-----\n"

var sigCols = []string{
	"id", "document_id", "document_version", "signer_id", "signer_name",
	"signature_algorithm", "digest_hex", "signature_der", "certificate_pem",
	"signed_at", "created_at",
}

func newSigRepoMock(t *testing.T) (*SignatureRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewSignatureRepositoryPG(db), mock
}

func freshSignature(t *testing.T) *entities.DocumentSignature {
	t.Helper()
	sig, err := entities.NewDocumentSignature(
		42, 3, 7, "Иванов И.И.",
		entities.SignatureAlgorithmECDSAP256SHA256,
		sigDigestHex, []byte{0x30, 0x44, 0x01}, sigCertPEM,
		time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	return sig
}

// Compile-time proof the PG implementation satisfies the usecase port.
var _ usecases.SignatureRepository = (*SignatureRepositoryPG)(nil)

func TestNewSignatureRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewSignatureRepositoryPG(db))
}

func TestSignatureRepositoryPG_Save_HappyPath(t *testing.T) {
	repo, mock := newSigRepoMock(t)
	sig := freshSignature(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_signatures")).
		WithArgs(
			int64(42), 3, int64(7), "Иванов И.И.",
			"ECDSA_P256_SHA256", sigDigestHex, []byte{0x30, 0x44, 0x01}, sigCertPEM,
			sig.SignedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))

	id, err := repo.Save(context.Background(), sig)
	require.NoError(t, err)
	assert.Equal(t, int64(99), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignatureRepositoryPG_ListByDocument(t *testing.T) {
	repo, mock := newSigRepoMock(t)
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows(sigCols).
		AddRow(int64(1), int64(42), 3, int64(7), "Иванов И.И.", "ECDSA_P256_SHA256", sigDigestHex, []byte{0x30}, sigCertPEM, now, now).
		AddRow(int64(2), int64(42), 3, int64(8), "Петров П.П.", "ECDSA_P256_SHA256", sigDigestHex, []byte{0x31}, sigCertPEM, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM document_signatures")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	got, err := repo.ListByDocument(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(1), got[0].ID)
	assert.Equal(t, "Петров П.П.", got[1].SignerName)
	assert.Equal(t, entities.SignatureAlgorithmECDSAP256SHA256, got[0].Algorithm)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignatureRepositoryPG_GetByID_HappyPath(t *testing.T) {
	repo, mock := newSigRepoMock(t)
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows(sigCols).
		AddRow(int64(99), int64(42), 3, int64(7), "Иванов И.И.", "ECDSA_P256_SHA256", sigDigestHex, []byte{0x30}, sigCertPEM, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM document_signatures")).
		WithArgs(int64(99)).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 99)
	require.NoError(t, err)
	assert.Equal(t, int64(99), got.ID)
	assert.Equal(t, int64(42), got.DocumentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignatureRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newSigRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_signatures")).
		WithArgs(int64(123)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 123)
	assert.True(t, errors.Is(err, repositories.ErrSignatureNotFound), "want ErrSignatureNotFound, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
