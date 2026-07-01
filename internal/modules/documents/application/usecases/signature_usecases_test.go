package usecases_test

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

const contentHashHex = "ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12"

func fixedNow() time.Time { return time.Date(2026, 6, 26, 12, 0, 0, 30, time.UTC) }

// --- fakes ----------------------------------------------------------------

type fakeSigRepo struct {
	saved   *entities.DocumentSignature
	saveID  int64
	saveErr error
	listOut []*entities.DocumentSignature
	listErr error
	getOut  *entities.DocumentSignature
	getErr  error
}

func (f *fakeSigRepo) Save(_ context.Context, sig *entities.DocumentSignature) (int64, error) {
	f.saved = sig
	return f.saveID, f.saveErr
}
func (f *fakeSigRepo) ListByDocument(_ context.Context, _ int64) ([]*entities.DocumentSignature, error) {
	return f.listOut, f.listErr
}
func (f *fakeSigRepo) GetByID(_ context.Context, _ int64) (*entities.DocumentSignature, error) {
	return f.getOut, f.getErr
}

type fakeView struct {
	version int
	hash    string
	err     error
}

func (f *fakeView) GetForSigning(_ context.Context, _ int64) (int, string, error) {
	return f.version, f.hash, f.err
}

type fakeEngine struct {
	der        []byte
	cert       string
	signErr    error
	verifyOK   bool
	verifyErr  error
	signedHash []byte
}

func (f *fakeEngine) SignDigest(_ context.Context, _ int64, _ string, digest []byte) ([]byte, string, error) {
	f.signedHash = digest
	return f.der, f.cert, f.signErr
}
func (f *fakeEngine) Verify(_ string, _ []byte, _ []byte) (bool, error) {
	return f.verifyOK, f.verifyErr
}

type recordingAudit struct{ actions []string }

func (a *recordingAudit) LogAuditEvent(_ context.Context, action, _ string, _ map[string]any) {
	a.actions = append(a.actions, action)
}

const fakeCert = "-----BEGIN CERTIFICATE-----\nMIIBfake\n-----END CERTIFICATE-----\n"

// --- SignDocument ---------------------------------------------------------

func TestSignDocumentUseCase_HappyPath(t *testing.T) {
	repo := &fakeSigRepo{saveID: 99}
	view := &fakeView{version: 3, hash: contentHashHex}
	engine := &fakeEngine{der: []byte{0x30, 0x44, 0x01}, cert: fakeCert}
	audit := &recordingAudit{}
	uc := usecases.NewSignDocumentUseCase(repo, view, engine, audit, fixedNow)

	sig, err := uc.Execute(context.Background(), 42, 7, "Иванов И.И.", "methodist")
	require.NoError(t, err)
	assert.Equal(t, int64(99), sig.ID)
	assert.Equal(t, int64(42), sig.DocumentID)
	assert.Equal(t, 3, sig.DocumentVersion)
	assert.Equal(t, entities.SignatureAlgorithmECDSAP256SHA256, sig.Algorithm)

	// The digest signed by the engine must equal ComputeSigningDigest over the
	// view's version/hash and the whole-seconds timestamp.
	wantDigest, derr := entities.ComputeSigningDigest(42, 3, 7, fixedNow().Truncate(time.Second).Unix(), contentHashHex)
	require.NoError(t, derr)
	assert.Equal(t, wantDigest, sig.DigestHex)
	assert.Equal(t, wantDigest, hex.EncodeToString(engine.signedHash))
	assert.Same(t, sig, repo.saved)
}

func TestSignDocumentUseCase_StudentDenied(t *testing.T) {
	uc := usecases.NewSignDocumentUseCase(&fakeSigRepo{}, &fakeView{version: 1, hash: contentHashHex}, &fakeEngine{}, &recordingAudit{}, fixedNow)
	_, err := uc.Execute(context.Background(), 42, 7, "Студент", "student")
	assert.True(t, errors.Is(err, entities.ErrDocumentEditDenied), "want ErrDocumentEditDenied, got %v", err)
}

func TestSignDocumentUseCase_DocumentNotFound(t *testing.T) {
	view := &fakeView{err: usecases.ErrDocumentNotFound}
	uc := usecases.NewSignDocumentUseCase(&fakeSigRepo{}, view, &fakeEngine{}, &recordingAudit{}, fixedNow)
	_, err := uc.Execute(context.Background(), 42, 7, "X", "methodist")
	assert.True(t, errors.Is(err, usecases.ErrDocumentNotFound))
}

// --- ListSignatures -------------------------------------------------------

func TestListSignaturesUseCase(t *testing.T) {
	s1 := &entities.DocumentSignature{ID: 1, DocumentID: 42}
	repo := &fakeSigRepo{listOut: []*entities.DocumentSignature{s1}}
	uc := usecases.NewListSignaturesUseCase(repo)
	got, err := uc.Execute(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(1), got[0].ID)
}

// --- VerifySignature ------------------------------------------------------

func storedSignature(t *testing.T) *entities.DocumentSignature {
	t.Helper()
	digest, err := entities.ComputeSigningDigest(42, 3, 7, fixedNow().Truncate(time.Second).Unix(), contentHashHex)
	require.NoError(t, err)
	return entities.ReconstituteDocumentSignature(
		5, 42, 3, 7, "Иванов И.И.", entities.SignatureAlgorithmECDSAP256SHA256,
		digest, []byte{0x30, 0x44}, fakeCert,
		fixedNow().Truncate(time.Second), fixedNow(),
	)
}

func TestVerifySignatureUseCase_Valid(t *testing.T) {
	sig := storedSignature(t)
	repo := &fakeSigRepo{getOut: sig}
	view := &fakeView{version: 3, hash: contentHashHex}
	engine := &fakeEngine{verifyOK: true}
	uc := usecases.NewVerifySignatureUseCase(repo, view, engine)

	v, err := uc.Execute(context.Background(), 5)
	require.NoError(t, err)
	assert.True(t, v.Valid)
	assert.True(t, v.DigestMatch)
	assert.True(t, v.CryptoValid)
	assert.False(t, v.VersionChanged)
}

func TestVerifySignatureUseCase_DocumentModified(t *testing.T) {
	sig := storedSignature(t)
	repo := &fakeSigRepo{getOut: sig}
	// Different content hash => recomputed digest will not match the stored one.
	view := &fakeView{version: 3, hash: "cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34"}
	engine := &fakeEngine{verifyOK: true}
	uc := usecases.NewVerifySignatureUseCase(repo, view, engine)

	v, err := uc.Execute(context.Background(), 5)
	require.NoError(t, err)
	assert.False(t, v.Valid)
	assert.False(t, v.DigestMatch)
}

func TestVerifySignatureUseCase_CryptoInvalid(t *testing.T) {
	sig := storedSignature(t)
	repo := &fakeSigRepo{getOut: sig}
	view := &fakeView{version: 3, hash: contentHashHex}
	engine := &fakeEngine{verifyOK: false} // signature does not match cert
	uc := usecases.NewVerifySignatureUseCase(repo, view, engine)

	v, err := uc.Execute(context.Background(), 5)
	require.NoError(t, err)
	assert.True(t, v.DigestMatch)
	assert.False(t, v.CryptoValid)
	assert.False(t, v.Valid)
}

func TestVerifySignatureUseCase_NotFound(t *testing.T) {
	repo := &fakeSigRepo{getErr: usecases.ErrSignatureNotFound}
	uc := usecases.NewVerifySignatureUseCase(repo, &fakeView{}, &fakeEngine{})
	_, err := uc.Execute(context.Background(), 123)
	assert.True(t, errors.Is(err, usecases.ErrSignatureNotFound))
}
