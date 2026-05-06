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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

type fakeGetPort struct {
	called bool
	gotID  int64
	out    *entities.Curriculum
	err    error
}

func (f *fakeGetPort) Execute(_ context.Context, id int64) (*entities.Curriculum, error) {
	f.called = true
	f.gotID = id
	return f.out, f.err
}

func setupGetRouter(get handlers.GetCurriculumPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	// fakeCreatePort lives in curriculum_handler_test.go (same package);
	// reuse its zero value to satisfy the constructor's nil guard.
	h := handlers.NewCurriculumHandler(&fakeCreatePort{}, get, stubListPort{}, stubUpdatePort{})
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
	r.GET("/api/curriculum/:id", h.Get)
	return r
}

func doGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCurriculumHandler_Get_HappyPath_AllNonStudentRoles(t *testing.T) {
	roles := []string{"methodist", "system_admin", "academic_secretary", "teacher"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			get := &fakeGetPort{out: builtCurriculum(t, 7)}
			r := setupGetRouter(get, role, 42)

			rec := doGet(t, r, "/api/curriculum/7")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.True(t, get.called)
			assert.Equal(t, int64(7), get.gotID)

			var resp struct {
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.True(t, resp.Success)
			assert.EqualValues(t, 7, resp.Data["id"])
			assert.Equal(t, "ИВТ-2026", resp.Data["title"])
		})
	}
}

func TestCurriculumHandler_Get_RejectsStudent(t *testing.T) {
	get := &fakeGetPort{}
	r := setupGetRouter(get, "student", 42)

	rec := doGet(t, r, "/api/curriculum/7")
	assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	assert.False(t, get.called)
}

func TestCurriculumHandler_Get_MissingContextReturns401(t *testing.T) {
	get := &fakeGetPort{}
	r := setupGetRouter(get, "", 0)

	rec := doGet(t, r, "/api/curriculum/7")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Get_BadIDReturns400(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{"non-numeric", "/api/curriculum/abc"},
		{"negative", "/api/curriculum/-1"},
		{"zero", "/api/curriculum/0"},
		{"fractional rejected", "/api/curriculum/1.5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			get := &fakeGetPort{}
			r := setupGetRouter(get, "methodist", 42)

			rec := doGet(t, r, tc.path)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, get.called, "use case must not be invoked on bad id")
		})
	}
}

func TestCurriculumHandler_Get_NotFoundReturns404(t *testing.T) {
	get := &fakeGetPort{err: repositories.ErrCurriculumNotFound}
	r := setupGetRouter(get, "methodist", 42)

	rec := doGet(t, r, "/api/curriculum/999")
	assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Get_TransportErrorReturns500(t *testing.T) {
	get := &fakeGetPort{err: errors.New("conn refused")}
	r := setupGetRouter(get, "methodist", 42)

	rec := doGet(t, r, "/api/curriculum/7")
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}
