package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/interfaces/http/handlers"
)

// --- fakes ------------------------------------------------------------------

type fakeImport struct {
	res      sdUsecases.ImportResult
	err      error
	gotActor int64
	gotRole  string
	gotBody  []byte
}

func (f *fakeImport) Execute(_ context.Context, actorID int64, role string, src io.Reader) (sdUsecases.ImportResult, error) {
	f.gotActor, f.gotRole = actorID, role
	f.gotBody, _ = io.ReadAll(src)
	return f.res, f.err
}

type fakeExport struct {
	data      []byte
	err       error
	gotActor  int64
	gotRole   string
	gotFilter repositories.StudentDebtListFilter
}

func (f *fakeExport) Execute(_ context.Context, actorID int64, role string, filter repositories.StudentDebtListFilter) ([]byte, error) {
	f.gotActor, f.gotRole, f.gotFilter = actorID, role, filter
	return f.data, f.err
}

func newTransferRouter(imp *fakeImport, exp *fakeExport, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	for _, h := range mw {
		api.Use(h)
	}
	h := handlers.NewStudentDebtTransferHandler(imp, exp)
	handlers.RegisterStudentDebtTransferRoutes(api, h)
	return r
}

// doUpload posts a single-file multipart request under fieldName.
func doUpload(t *testing.T, r *gin.Engine, path, fieldName, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// importEnvelope decodes the import response body for field assertions.
type importEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		Created int `json:"created"`
		Updated int `json:"updated"`
		Skipped int `json:"skipped"`
		Errors  []struct {
			Row      int    `json:"row"`
			Identity string `json:"identity"`
			Message  string `json:"message"`
		} `json:"errors"`
	} `json:"data"`
}

// --- Import -----------------------------------------------------------------

func TestStudentDebtTransferHandler_Import_HappyPath(t *testing.T) {
	imp := &fakeImport{res: sdUsecases.ImportResult{
		Created: 2, Updated: 1, Skipped: 3,
		Errors: []sdUsecases.ImportRowError{{Row: 4, Identity: "ИВТ-21 / Иванов / БД / сем. 3", Message: "service id 99 not found"}},
	}}
	r := newTransferRouter(imp, &fakeExport{}, withAuth(7, "methodist"))

	w := doUpload(t, r, "/api/v1/student-debts/import", "file", "debts.xlsx", []byte("xlsx-bytes"))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(7), imp.gotActor)
	assert.Equal(t, "methodist", imp.gotRole)
	assert.Equal(t, []byte("xlsx-bytes"), imp.gotBody)

	var env importEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, 2, env.Data.Created)
	assert.Equal(t, 1, env.Data.Updated)
	assert.Equal(t, 3, env.Data.Skipped)
	require.Len(t, env.Data.Errors, 1)
	assert.Equal(t, 4, env.Data.Errors[0].Row)
	assert.Equal(t, "ИВТ-21 / Иванов / БД / сем. 3", env.Data.Errors[0].Identity)
	assert.Equal(t, "service id 99 not found", env.Data.Errors[0].Message)
}

func TestStudentDebtTransferHandler_Import_EmptyErrorsIsArray(t *testing.T) {
	// No row errors must serialize as [] (not null) so the frontend can map over it.
	imp := &fakeImport{res: sdUsecases.ImportResult{Created: 1}}
	r := newTransferRouter(imp, &fakeExport{}, withAuth(7, "methodist"))

	w := doUpload(t, r, "/api/v1/student-debts/import", "file", "debts.xlsx", []byte("x"))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"errors":[]`)
}

func TestStudentDebtTransferHandler_Import_ForbiddenIs403(t *testing.T) {
	imp := &fakeImport{err: entities.ErrDebtAccessForbidden}
	r := newTransferRouter(imp, &fakeExport{}, withAuth(9, "teacher"))

	w := doUpload(t, r, "/api/v1/student-debts/import", "file", "debts.xlsx", []byte("x"))
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudentDebtTransferHandler_Import_MalformedDocumentIs400(t *testing.T) {
	// A non-forbidden use-case error is a parse-level (bad document) failure.
	imp := &fakeImport{err: errors.New("student_debts: import parse: corrupt xlsx")}
	r := newTransferRouter(imp, &fakeExport{}, withAuth(7, "methodist"))

	w := doUpload(t, r, "/api/v1/student-debts/import", "file", "debts.xlsx", []byte("x"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentDebtTransferHandler_Import_MissingFileIs400(t *testing.T) {
	imp := &fakeImport{}
	r := newTransferRouter(imp, &fakeExport{}, withAuth(7, "methodist"))

	// Upload under the wrong field name → c.FormFile("file") fails.
	w := doUpload(t, r, "/api/v1/student-debts/import", "wrong_field", "debts.xlsx", []byte("x"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, imp.gotBody, "use case must not run when no file was uploaded")
}

func TestStudentDebtTransferHandler_Import_MissingAuthIs401(t *testing.T) {
	r := newTransferRouter(&fakeImport{}, &fakeExport{})
	w := doUpload(t, r, "/api/v1/student-debts/import", "file", "debts.xlsx", []byte("x"))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Export -----------------------------------------------------------------

func TestStudentDebtTransferHandler_Export_HappyPath(t *testing.T) {
	exp := &fakeExport{data: []byte("XLSX-DOCUMENT-BYTES")}
	r := newTransferRouter(&fakeImport{}, exp, withAuth(7, "methodist"))

	w := doGET(t, r, "/api/v1/student-debts/export?group_name=ИВТ-21&status=open&semester=3")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []byte("XLSX-DOCUMENT-BYTES"), w.Body.Bytes())
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="student-debts.xlsx"`, w.Header().Get("Content-Disposition"))
	assert.Equal(t, "19", w.Header().Get("Content-Length"))

	// Filter is forwarded to the use case.
	assert.Equal(t, "ИВТ-21", exp.gotFilter.GroupName)
	require.NotNil(t, exp.gotFilter.Status)
	assert.Equal(t, entities.DebtStatusOpen, *exp.gotFilter.Status)
	require.NotNil(t, exp.gotFilter.Semester)
	assert.Equal(t, 3, *exp.gotFilter.Semester)
}

func TestStudentDebtTransferHandler_Export_ForbiddenIs403(t *testing.T) {
	exp := &fakeExport{err: entities.ErrDebtAccessForbidden}
	r := newTransferRouter(&fakeImport{}, exp, withAuth(9, "student"))

	w := doGET(t, r, "/api/v1/student-debts/export")
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudentDebtTransferHandler_Export_RepoErrorIs500(t *testing.T) {
	exp := &fakeExport{err: errors.New("student_debts: export: list: db down")}
	r := newTransferRouter(&fakeImport{}, exp, withAuth(7, "methodist"))

	w := doGET(t, r, "/api/v1/student-debts/export")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentDebtTransferHandler_Export_MissingAuthIs401(t *testing.T) {
	r := newTransferRouter(&fakeImport{}, &fakeExport{})
	w := doGET(t, r, "/api/v1/student-debts/export")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- construction -----------------------------------------------------------

func TestNewStudentDebtTransferHandler_PanicsOnNilPort(t *testing.T) {
	imp, exp := &fakeImport{}, &fakeExport{}
	cases := map[string]func(){
		"nil import": func() { handlers.NewStudentDebtTransferHandler(nil, exp) },
		"nil export": func() { handlers.NewStudentDebtTransferHandler(imp, nil) },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic on nil port")
				}
			}()
			build()
		})
	}
}
