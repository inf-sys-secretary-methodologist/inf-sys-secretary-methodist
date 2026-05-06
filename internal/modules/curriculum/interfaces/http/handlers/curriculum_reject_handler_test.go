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

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

type fakeRejectPort struct {
	called   bool
	gotAdmin int64
	gotInput curUsecases.RejectCurriculumInput
	out      *entities.Curriculum
	err      error
}

func (f *fakeRejectPort) Execute(_ context.Context, adminID int64, in curUsecases.RejectCurriculumInput) (*entities.Curriculum, error) {
	f.called = true
	f.gotAdmin = adminID
	f.gotInput = in
	return f.out, f.err
}

func setupRejectRouter(reject handlers.RejectCurriculumPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		&fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{},
		stubSubmitPort{}, stubApprovePort{}, reject,
	)
	if role != "" || userID != 0 {
		r.Use(func(c *gin.Context) {
			if userID != 0 {
				c.Set("user_id", userID)
			}
			if role != "" {
				c.Set("role", role)
			}
			c.Next()
		})
	}
	r.POST("/api/curriculum/:id/reject", h.Reject)
	return r
}

func doReject(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	if s, ok := body.(string); ok {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(s))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
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

func TestCurriculumHandler_Reject_HappyPath_AdminWithReason(t *testing.T) {
	reject := &fakeRejectPort{out: builtCurriculum(t, 7)}
	r := setupRejectRouter(reject, "system_admin", 99)

	body := map[string]any{"reason": "Не соответствует ФГОС"}
	rec := doReject(t, r, "/api/curriculum/7/reject", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, reject.called)
	assert.Equal(t, int64(99), reject.gotAdmin)
	assert.Equal(t, int64(7), reject.gotInput.ID)
	assert.Equal(t, "Не соответствует ФГОС", reject.gotInput.Reason)
}

func TestCurriculumHandler_Reject_RejectsNonAdminRoles(t *testing.T) {
	cases := []string{"methodist", "teacher", "academic_secretary", "student", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			reject := &fakeRejectPort{}
			r := setupRejectRouter(reject, role, 42)

			rec := doReject(t, r, "/api/curriculum/7/reject", map[string]any{"reason": "x"})
			assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			assert.False(t, reject.called)
		})
	}
}

func TestCurriculumHandler_Reject_MissingContextReturns401(t *testing.T) {
	reject := &fakeRejectPort{}
	r := setupRejectRouter(reject, "", 0)

	rec := doReject(t, r, "/api/curriculum/7/reject", map[string]any{"reason": "x"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Reject_BadIDReturns400(t *testing.T) {
	cases := []string{"abc", "0", "-1", "1.5"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			reject := &fakeRejectPort{}
			r := setupRejectRouter(reject, "system_admin", 99)

			rec := doReject(t, r, "/api/curriculum/"+raw+"/reject",
				map[string]any{"reason": "x"})
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, reject.called)
		})
	}
}

func TestCurriculumHandler_Reject_MalformedBodyReturns400(t *testing.T) {
	reject := &fakeRejectPort{}
	r := setupRejectRouter(reject, "system_admin", 99)

	rec := doReject(t, r, "/api/curriculum/7/reject", "not-json")
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.False(t, reject.called)
}

func TestCurriculumHandler_Reject_RequiresNonEmptyReason(t *testing.T) {
	cases := []struct {
		name string
		body any
	}{
		{"missing reason key", map[string]any{}},
		{"empty reason", map[string]any{"reason": ""}},
		{"whitespace reason", map[string]any{"reason": "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reject := &fakeRejectPort{}
			r := setupRejectRouter(reject, "system_admin", 99)

			rec := doReject(t, r, "/api/curriculum/7/reject", tc.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String(),
				"reason must be required and non-empty after trim")
			assert.False(t, reject.called,
				"use case must not be invoked without a reason")
		})
	}
}

func TestCurriculumHandler_Reject_DomainErrorMappings(t *testing.T) {
	cases := []struct {
		name  string
		ucErr error
		want  int
	}{
		{"not pending → 422", entities.ErrCannotReject, http.StatusUnprocessableEntity},
		{"not found → 404", repositories.ErrCurriculumNotFound, http.StatusNotFound},
		{"transport → 500", errors.New("conn refused"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reject := &fakeRejectPort{err: tc.ucErr}
			r := setupRejectRouter(reject, "system_admin", 99)

			body := map[string]any{"reason": "x"}
			rec := doReject(t, r, "/api/curriculum/7/reject", body)
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
		})
	}
}
