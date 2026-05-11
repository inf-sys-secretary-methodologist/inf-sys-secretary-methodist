package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/interfaces/http/handlers"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeGeneratePort records call args + returns canned response.
type fakeGeneratePort struct {
	called  bool
	gotYear int
	gotID   int64
	result  []byte
	err     error
}

func (f *fakeGeneratePort) Generate(_ context.Context, in usecases.GenerateAnnualReportInput) ([]byte, error) {
	f.called = true
	f.gotYear = in.Year
	f.gotID = in.ActorID
	return f.result, f.err
}

// withAuth mirrors production middleware c.Set("user_id"/"role") keys.
func withAuth(uid int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != 0 {
			c.Set("user_id", uid)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
}

func setupRouter(port handlers.GenerateAnnualReportPort, uid int64, role string) *gin.Engine {
	r := gin.New()
	h := handlers.NewAnnualReportHandler(port)
	r.Use(withAuth(uid, role))
	r.GET("/api/reports/annual", h.Generate)
	return r
}

func doGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- happy path --------------------------------------------------------

func TestAnnualReportHandler_HappyPath_Methodist(t *testing.T) {
	port := &fakeGeneratePort{result: []byte{0x50, 0x4B, 'D', 'O', 'C', 'X'}}
	r := setupRouter(port, 42, "methodist")

	rec := doGet(t, r, "/api/reports/annual?year=2026")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", rec.Header().Get("Content-Type"))
	require.Equal(t, `attachment; filename="annual_report_2026.docx"`, rec.Header().Get("Content-Disposition"))
	require.Equal(t, []byte{0x50, 0x4B, 'D', 'O', 'C', 'X'}, rec.Body.Bytes())

	require.True(t, port.called)
	require.Equal(t, 2026, port.gotYear)
	require.Equal(t, int64(42), port.gotID)
}

func TestAnnualReportHandler_HappyPath_SystemAdmin(t *testing.T) {
	port := &fakeGeneratePort{result: []byte("DOCX")}
	r := setupRouter(port, 1, "system_admin")

	rec := doGet(t, r, "/api/reports/annual?year=2026")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// --- bad year (422) ----------------------------------------------------

func TestAnnualReportHandler_BadYear_Returns422(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"missing year param", "/api/reports/annual"},
		{"empty year param", "/api/reports/annual?year="},
		{"non-numeric year", "/api/reports/annual?year=abc"},
		{"year too small", "/api/reports/annual?year=1999"},
		{"year too large", "/api/reports/annual?year=2101"},
		{"negative year", "/api/reports/annual?year=-2026"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			port := &fakeGeneratePort{}
			r := setupRouter(port, 42, "methodist")

			rec := doGet(t, r, tc.path)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			require.False(t, port.called, "usecase must NOT be called for bad input")
		})
	}
}

// --- role gate (403) ---------------------------------------------------

func TestAnnualReportHandler_NonMethodistRole_Returns403(t *testing.T) {
	cases := []string{"teacher", "student", "academic_secretary", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			port := &fakeGeneratePort{}
			r := setupRouter(port, 42, role)

			rec := doGet(t, r, "/api/reports/annual?year=2026")
			require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			require.False(t, port.called)
		})
	}
}

// --- missing auth context (401) ----------------------------------------

func TestAnnualReportHandler_MissingUserID_Returns401(t *testing.T) {
	port := &fakeGeneratePort{}
	r := setupRouter(port, 0, "methodist")

	rec := doGet(t, r, "/api/reports/annual?year=2026")
	require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
	require.False(t, port.called)
}

func TestAnnualReportHandler_MissingRole_Returns403(t *testing.T) {
	port := &fakeGeneratePort{}
	r := setupRouter(port, 42, "")

	rec := doGet(t, r, "/api/reports/annual?year=2026")
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	require.False(t, port.called)
}

// --- usecase failure (500) ---------------------------------------------

func TestAnnualReportHandler_UseCaseError_Returns500(t *testing.T) {
	port := &fakeGeneratePort{err: errors.New("internal boom")}
	r := setupRouter(port, 42, "methodist")

	rec := doGet(t, r, "/api/reports/annual?year=2026")
	require.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
	require.True(t, port.called)
}
