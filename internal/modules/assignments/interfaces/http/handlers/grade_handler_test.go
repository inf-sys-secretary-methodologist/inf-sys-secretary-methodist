package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/interfaces/http/handlers"
)

func init() { gin.SetMode(gin.TestMode) }

type fakeSaveGradeUseCase struct {
	err          error
	called       bool
	gotTeacherID int64
	gotInput     assignUsecases.SaveGradeInput
}

func (f *fakeSaveGradeUseCase) Execute(ctx context.Context, teacherID int64, in assignUsecases.SaveGradeInput) error {
	f.called = true
	f.gotTeacherID = teacherID
	f.gotInput = in
	return f.err
}

func setupRouter(uc handlers.SaveGradeUseCasePort, withAuth bool, teacherID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewGradeHandler(uc)
	if withAuth {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", teacherID)
			c.Set("role", "teacher")
			c.Next()
		})
	}
	r.POST("/api/assignments/:id/grades", h.SaveGrade)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestSaveGradeHandler_HappyPath(t *testing.T) {
	uc := &fakeSaveGradeUseCase{}
	r := setupRouter(uc, true, 42)

	body := map[string]any{"student_id": 7, "value": 85, "feedback": "good"}
	rec := doRequest(t, r, "/api/assignments/10/grades", body)

	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.True(t, uc.called)
	assert.Equal(t, int64(42), uc.gotTeacherID)
	assert.Equal(t, int64(10), uc.gotInput.AssignmentID)
	assert.Equal(t, int64(7), uc.gotInput.StudentID)
	assert.Equal(t, 85, uc.gotInput.Value)
	assert.Equal(t, "good", uc.gotInput.Feedback)
}

func TestSaveGradeHandler_InputValidation(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		body     any
		withAuth bool
		want     int
	}{
		{name: "non-numeric assignment id → 400", path: "/api/assignments/abc/grades",
			body: map[string]any{"student_id": 7, "value": 50}, withAuth: true, want: http.StatusBadRequest},
		{name: "malformed json body → 400", path: "/api/assignments/10/grades",
			body: "not-json-but-string", withAuth: true, want: http.StatusBadRequest},
		{name: "missing student_id → 400", path: "/api/assignments/10/grades",
			body: map[string]any{"value": 50}, withAuth: true, want: http.StatusBadRequest},
		{name: "negative student_id → 400", path: "/api/assignments/10/grades",
			body: map[string]any{"student_id": -1, "value": 50}, withAuth: true, want: http.StatusBadRequest},
		{name: "no auth context → 401", path: "/api/assignments/10/grades",
			body: map[string]any{"student_id": 7, "value": 50}, withAuth: false, want: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeSaveGradeUseCase{}
			r := setupRouter(uc, tc.withAuth, 42)

			var rec *httptest.ResponseRecorder
			if s, ok := tc.body.(string); ok {
				req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(s))
				req.Header.Set("Content-Type", "application/json")
				rec = httptest.NewRecorder()
				r.ServeHTTP(rec, req)
			} else {
				rec = doRequest(t, r, tc.path, tc.body)
			}

			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
			assert.False(t, uc.called, "use case must not be invoked when input validation fails")
		})
	}
}

func TestSaveGradeHandler_DomainErrorMapping(t *testing.T) {
	tests := []struct {
		name      string
		ucErr     error
		wantCode  int
		wantBody  string
	}{
		{name: "ErrAssignmentNotFound → 404",
			ucErr: repositories.ErrAssignmentNotFound, wantCode: http.StatusNotFound, wantBody: "NOT_FOUND"},
		{name: "ErrAssignmentScopeForbidden → 403",
			ucErr: entities.ErrAssignmentScopeForbidden, wantCode: http.StatusForbidden, wantBody: "FORBIDDEN"},
		{name: "ErrInvalidScore → 422",
			ucErr: entities.ErrInvalidScore, wantCode: http.StatusUnprocessableEntity, wantBody: "INVALID_INPUT"},
		{name: "ErrInvalidAssignment → 422",
			ucErr: entities.ErrInvalidAssignment, wantCode: http.StatusUnprocessableEntity, wantBody: "INVALID_INPUT"},
		{name: "ErrAlreadyGraded → 409",
			ucErr: entities.ErrAlreadyGraded, wantCode: http.StatusConflict, wantBody: "ALREADY_GRADED"},
		{name: "unknown error → 500",
			ucErr: errors.New("kaboom"), wantCode: http.StatusInternalServerError, wantBody: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeSaveGradeUseCase{err: tc.ucErr}
			r := setupRouter(uc, true, 42)

			body := map[string]any{"student_id": 7, "value": 50}
			rec := doRequest(t, r, "/api/assignments/10/grades", body)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				assert.Contains(t, rec.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestSaveGradeHandler_NilUseCasePanics(t *testing.T) {
	assert.Panics(t, func() {
		handlers.NewGradeHandler(nil)
	}, "constructor must panic on nil usecase to make wiring failures loud")
}
