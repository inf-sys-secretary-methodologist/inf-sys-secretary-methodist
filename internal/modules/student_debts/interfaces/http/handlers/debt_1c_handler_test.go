package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/interfaces/http/handlers"
)

type fakeImport1C struct {
	res      sdUsecases.ImportResult
	err      error
	gotActor int64
	gotRole  string
	called   bool
}

func (f *fakeImport1C) Execute(_ context.Context, actorID int64, role string) (sdUsecases.ImportResult, error) {
	f.called = true
	f.gotActor, f.gotRole = actorID, role
	return f.res, f.err
}

func new1CRouter(i1c *fakeImport1C, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	for _, h := range mw {
		api.Use(h)
	}
	h := handlers.NewStudentDebt1CImportHandler(i1c)
	handlers.RegisterStudentDebt1CImportRoutes(api, h)
	return r
}

func post1C(r *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/student-debts/import-1c", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestStudentDebt1CImport_HappyPath(t *testing.T) {
	i1c := &fakeImport1C{res: sdUsecases.ImportResult{Created: 3, Updated: 1, Skipped: 0}}
	r := new1CRouter(i1c, withAuth(7, "methodist"))

	w := post1C(r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, i1c.called)
	assert.Equal(t, int64(7), i1c.gotActor)
	assert.Equal(t, "methodist", i1c.gotRole)

	var env importEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, 3, env.Data.Created)
	assert.Equal(t, 1, env.Data.Updated)
}

func TestStudentDebt1CImport_EmptyErrorsIsArray(t *testing.T) {
	i1c := &fakeImport1C{res: sdUsecases.ImportResult{Created: 1}}
	r := new1CRouter(i1c, withAuth(7, "methodist"))

	w := post1C(r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"errors":[]`)
}

func TestStudentDebt1CImport_ForbiddenIs403(t *testing.T) {
	i1c := &fakeImport1C{err: entities.ErrDebtAccessForbidden}
	r := new1CRouter(i1c, withAuth(9, "teacher"))

	w := post1C(r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudentDebt1CImport_UpstreamErrorIs502(t *testing.T) {
	// A non-forbidden use-case error is a 1С transport/parse failure — an
	// upstream-dependency error, not a client error.
	i1c := &fakeImport1C{err: errors.New("student_debts: 1С fetch: connection refused")}
	r := new1CRouter(i1c, withAuth(7, "methodist"))

	w := post1C(r)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestStudentDebt1CImport_MissingAuthIs401(t *testing.T) {
	i1c := &fakeImport1C{}
	r := new1CRouter(i1c) // no auth middleware

	w := post1C(r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, i1c.called, "use case must not run without an authenticated actor")
}
