package signing

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appcrypto "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
)

// testKEK is a deterministic 32-byte AES-256 key for at-rest encryption tests.
var testKEK = []byte("0123456789abcdef0123456789abcdef")

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC) }
}

func newServiceMock(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, testKEK, fixedClock()), mock
}

// digest32 is a fixed 32-byte SHA-256-sized digest to be signed.
func digest32() []byte {
	d := make([]byte, 32)
	for i := range d {
		d[i] = byte(i + 1)
	}
	return d
}

func TestNewService_PanicsOnBadKEK(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.Panics(t, func() { NewService(db, []byte("short"), fixedClock()) })
}

// TestService_SignDigest_FirstUse_GeneratesKeyThenVerifies pins the
// first-signature path: no existing key row -> generate ECDSA P-256 key +
// self-signed X.509 cert -> persist (encrypted private key) -> sign. The
// returned certificate must verify the returned signature over the digest.
func TestService_SignDigest_FirstUse_GeneratesKeyThenVerifies(t *testing.T) {
	svc, mock := newServiceMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM signing_keys")).
		WithArgs(int64(7)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO signing_keys")).
		WithArgs(int64(7), "ECDSA_P256_SHA256", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	der, certPEM, err := svc.SignDigest(context.Background(), 7, "Иванов И.И.", digest32())
	require.NoError(t, err)
	require.NotEmpty(t, der)
	require.Contains(t, certPEM, "BEGIN CERTIFICATE")

	ok, err := svc.Verify(certPEM, digest32(), der)
	require.NoError(t, err)
	assert.True(t, ok, "freshly produced signature must verify")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestService_Verify_RejectsTamperedDigest pins that a signature does not
// verify against a different digest (document body changed after signing).
func TestService_Verify_RejectsTamperedDigest(t *testing.T) {
	svc, mock := newServiceMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM signing_keys")).WithArgs(int64(7)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO signing_keys")).WillReturnResult(sqlmock.NewResult(1, 1))

	der, certPEM, err := svc.SignDigest(context.Background(), 7, "X", digest32())
	require.NoError(t, err)

	tampered := digest32()
	tampered[0] ^= 0xFF
	ok, err := svc.Verify(certPEM, tampered, der)
	require.NoError(t, err)
	assert.False(t, ok, "tampered digest must not verify")
}

// TestService_SignDigest_ReusesExistingKey pins that when a key row already
// exists, the service decrypts it and signs WITHOUT inserting a new key, and
// the stored certificate verifies the new signature.
func TestService_SignDigest_ReusesExistingKey(t *testing.T) {
	svc, mock := newServiceMock(t)
	certPEM, encPriv := makeStoredKey(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM signing_keys")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"certificate_pem", "encrypted_private_key"}).
			AddRow(certPEM, encPriv))
	// No INSERT expected.

	der, gotCert, err := svc.SignDigest(context.Background(), 7, "X", digest32())
	require.NoError(t, err)
	assert.Equal(t, certPEM, gotCert)

	ok, err := svc.Verify(certPEM, digest32(), der)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Verify_BadCertificate(t *testing.T) {
	svc, _ := newServiceMock(t)
	_, err := svc.Verify("not a certificate", digest32(), []byte{0x30})
	assert.Error(t, err)
}

// --- test helpers ---------------------------------------------------------

// makeStoredKey produces a (certificatePEM, encryptedPrivateKey) pair as the
// service itself would persist, so the reuse path can be exercised.
func makeStoredKey(t *testing.T) (string, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:     time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))

	privDER, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER}))
	enc, err := appcrypto.EncryptString(privPEM, testKEK)
	require.NoError(t, err)
	return certPEM, enc
}
