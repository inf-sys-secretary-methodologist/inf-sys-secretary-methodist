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

type fakeUpdatePort struct {
	called     bool
	gotActor   int64
	gotIsAdmin bool
	gotInput   curUsecases.UpdateCurriculumInput
	out        *entities.Curriculum
	err        error
}

func (f *fakeUpdatePort) Execute(_ context.Context, actorID int64, isAdmin bool, in curUsecases.UpdateCurriculumInput) (*entities.Curriculum, error) {
	f.called = true
	f.gotActor = actorID
	f.gotIsAdmin = isAdmin
	f.gotInput = in
	return f.out, f.err
}

func setupUpdateRouter(update handlers.UpdateCurriculumPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		&fakeCreatePort{}, stubGetPort{}, stubListPort{}, update,
		stubSubmitPort{}, stubApprovePort{}, stubRejectPort{},
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
	r.PUT("/api/curriculum/:id", h.Update)
	return r
}

func doUpdate(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	if s, ok := body.(string); ok {
		req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(s))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(http.MethodPut, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCurriculumHandler_Update_HappyPath_Methodist(t *testing.T) {
	update := &fakeUpdatePort{out: builtCurriculum(t, 7)}
	r := setupUpdateRouter(update, "methodist", 42)

	body := map[string]any{
		"title":       "New Title",
		"code":        "NEW-2026",
		"specialty":   "New Specialty",
		"year":        2026,
		"description": "new desc",
	}
	rec := doUpdate(t, r, "/api/curriculum/7", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, update.called)
	assert.Equal(t, int64(42), update.gotActor)
	assert.False(t, update.gotIsAdmin, "methodist must not pass isAdmin=true")
	assert.Equal(t, int64(7), update.gotInput.ID)
	assert.Equal(t, "New Title", update.gotInput.Title)
}

func TestCurriculumHandler_Update_HappyPath_AdminPassesIsAdminTrue(t *testing.T) {
	update := &fakeUpdatePort{out: builtCurriculum(t, 7)}
	r := setupUpdateRouter(update, "system_admin", 99)

	body := map[string]any{
		"title": "T", "code": "C", "specialty": "S", "year": 2026,
	}
	rec := doUpdate(t, r, "/api/curriculum/7", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, update.gotIsAdmin, "system_admin must propagate isAdmin=true to use case")
	assert.Equal(t, int64(99), update.gotActor)
}

func TestCurriculumHandler_Update_RejectsNonWriteRoles(t *testing.T) {
	cases := []string{"teacher", "academic_secretary", "student", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			update := &fakeUpdatePort{}
			r := setupUpdateRouter(update, role, 42)

			body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
			rec := doUpdate(t, r, "/api/curriculum/7", body)
			assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			assert.False(t, update.called)
		})
	}
}

func TestCurriculumHandler_Update_MissingContextReturns401(t *testing.T) {
	update := &fakeUpdatePort{}
	r := setupUpdateRouter(update, "", 0)

	body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
	rec := doUpdate(t, r, "/api/curriculum/7", body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Update_BadPathIDReturns400(t *testing.T) {
	cases := []string{"abc", "0", "-1", "1.5"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			update := &fakeUpdatePort{}
			r := setupUpdateRouter(update, "methodist", 42)

			body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
			rec := doUpdate(t, r, "/api/curriculum/"+raw, body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, update.called)
		})
	}
}

func TestCurriculumHandler_Update_MalformedBodyReturns400(t *testing.T) {
	update := &fakeUpdatePort{}
	r := setupUpdateRouter(update, "methodist", 42)

	rec := doUpdate(t, r, "/api/curriculum/7", "not-json")
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.False(t, update.called)
}

func TestCurriculumHandler_Update_DomainErrorMappings(t *testing.T) {
	cases := []struct {
		name  string
		ucErr error
		want  int
	}{
		{"forbidden → 403", entities.ErrCurriculumScopeForbidden, http.StatusForbidden},
		{"not editable → 422", entities.ErrCannotEditApproved, http.StatusUnprocessableEntity},
		{"invariant → 422", entities.ErrInvalidCurriculum, http.StatusUnprocessableEntity},
		{"code conflict → 409", repositories.ErrCurriculumCodeExists, http.StatusConflict},
		// v0.157.0 #269 ADR-2 — lost-update race surfaces as 409
		// VERSION_CONFLICT (mirror section_handler.go precedent).
		{"version conflict → 409", repositories.ErrCurriculumVersionConflict, http.StatusConflict},
		{"not found → 404", repositories.ErrCurriculumNotFound, http.StatusNotFound},
		{"transport → 500", errors.New("conn refused"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			update := &fakeUpdatePort{err: tc.ucErr}
			r := setupUpdateRouter(update, "methodist", 42)

			body := map[string]any{"title": "T", "code": "C", "specialty": "S", "year": 2026}
			rec := doUpdate(t, r, "/api/curriculum/7", body)
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
		})
	}
}
