package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	docHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
)

// --- fakes ---------------------------------------------------------------

type fakeSign struct {
	out *entities.DocumentSignature
	err error
	got struct {
		documentID, signerID int64
		signerName, role     string
	}
}

func (f *fakeSign) Execute(_ context.Context, documentID, signerID int64, signerName, role string) (*entities.DocumentSignature, error) {
	f.got.documentID, f.got.signerID, f.got.signerName, f.got.role = documentID, signerID, signerName, role
	return f.out, f.err
}

type fakeList struct {
	out []*entities.DocumentSignature
	err error
}

func (f *fakeList) Execute(_ context.Context, _ int64) ([]*entities.DocumentSignature, error) {
	return f.out, f.err
}

type fakeVerify struct {
	out usecases.SignatureVerdict
	err error
}

func (f *fakeVerify) Execute(_ context.Context, _ int64) (usecases.SignatureVerdict, error) {
	return f.out, f.err
}

type fakeNames struct{ name string }

func (f *fakeNames) FullName(_ context.Context, _ int64) (string, error) { return f.name, nil }

func signatureFixture() *entities.DocumentSignature {
	const digest = "ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12"
	return entities.ReconstituteDocumentSignature(
		7, 42, 3, 9, "Иванов И.И.", entities.SignatureAlgorithmECDSAP256SHA256,
		digest, []byte{0x30, 0x44}, "-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----\n",
		time.Now(), time.Now(),
	)
}

func newSigEngine(t *testing.T, sign docHttp.SignDocumentPort, list docHttp.ListSignaturesPort, verify docHttp.VerifySignaturePort, names docHttp.SignerNameResolver, uid int64, role string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if uid != 0 {
			c.Set("user_id", uid)
			c.Set("role", role)
		}
		c.Next()
	})
	h := docHttp.NewSignatureHandler(sign, list, verify, names)
	docHttp.RegisterSignatureRoutes(r.Group("/api"), h)
	return r
}

// --- Sign ----------------------------------------------------------------

func TestSignatureHandler_Sign_Created(t *testing.T) {
	sign := &fakeSign{out: signatureFixture()}
	r := newSigEngine(t, sign, &fakeList{}, &fakeVerify{}, &fakeNames{name: "Иванов И.И."}, 9, "methodist")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/documents/42/sign", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, int64(42), sign.got.documentID)
	assert.Equal(t, int64(9), sign.got.signerID)
	assert.Equal(t, "Иванов И.И.", sign.got.signerName)
	assert.Equal(t, "methodist", sign.got.role)
}

func TestSignatureHandler_Sign_Unauthorized(t *testing.T) {
	r := newSigEngine(t, &fakeSign{}, &fakeList{}, &fakeVerify{}, &fakeNames{}, 0, "")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/documents/42/sign", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSignatureHandler_Sign_Forbidden(t *testing.T) {
	sign := &fakeSign{err: entities.ErrDocumentEditDenied}
	r := newSigEngine(t, sign, &fakeList{}, &fakeVerify{}, &fakeNames{name: "S"}, 5, "student")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/documents/42/sign", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSignatureHandler_Sign_DocumentNotFound(t *testing.T) {
	sign := &fakeSign{err: repositories.ErrDocumentNotFound}
	r := newSigEngine(t, sign, &fakeList{}, &fakeVerify{}, &fakeNames{name: "X"}, 9, "methodist")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/documents/42/sign", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- List ----------------------------------------------------------------

func TestSignatureHandler_List_OK(t *testing.T) {
	list := &fakeList{out: []*entities.DocumentSignature{signatureFixture()}}
	r := newSigEngine(t, &fakeSign{}, list, &fakeVerify{}, &fakeNames{}, 9, "teacher")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/documents/42/signatures", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Иванов")
}

// --- Verify --------------------------------------------------------------

func TestSignatureHandler_Verify_OK(t *testing.T) {
	verify := &fakeVerify{out: usecases.SignatureVerdict{SignatureID: 7, Valid: true, DigestMatch: true, CryptoValid: true}}
	r := newSigEngine(t, &fakeSign{}, &fakeList{}, verify, &fakeNames{}, 9, "teacher")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/documents/42/signatures/7/verify", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "valid")
}

func TestSignatureHandler_Verify_NotFound(t *testing.T) {
	verify := &fakeVerify{err: repositories.ErrSignatureNotFound}
	r := newSigEngine(t, &fakeSign{}, &fakeList{}, verify, &fakeNames{}, 9, "teacher")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/documents/42/signatures/7/verify", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
