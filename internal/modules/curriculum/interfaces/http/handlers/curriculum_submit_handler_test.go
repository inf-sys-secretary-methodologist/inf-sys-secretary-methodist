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

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

type fakeSubmitPort struct {
	called     bool
	gotActor   int64
	gotIsAdmin bool
	gotInput   curUsecases.SubmitForApprovalInput
	out        *entities.Curriculum
	err        error
}

func (f *fakeSubmitPort) Execute(_ context.Context, actorID int64, isAdmin bool, in curUsecases.SubmitForApprovalInput) (*entities.Curriculum, error) {
	f.called = true
	f.gotActor = actorID
	f.gotIsAdmin = isAdmin
	f.gotInput = in
	return f.out, f.err
}

// Stub ports for the not-under-test endpoints. Defined here (alongside
// fakeSubmitPort) so the v0.117.0 stubs live next to each other; other
// test files in this package reference these types when they wire a
// CurriculumHandler that doesn't exercise the lifecycle paths.
type stubSubmitPort struct{}

func (stubSubmitPort) Execute(context.Context, int64, bool, curUsecases.SubmitForApprovalInput) (*entities.Curriculum, error) {
	return nil, errors.New("stub")
}

type stubApprovePort struct{}

func (stubApprovePort) Execute(context.Context, int64, curUsecases.ApproveCurriculumInput) (*entities.Curriculum, error) {
	return nil, errors.New("stub")
}

type stubRejectPort struct{}

func (stubRejectPort) Execute(context.Context, int64, curUsecases.RejectCurriculumInput) (*entities.Curriculum, error) {
	return nil, errors.New("stub")
}

func setupSubmitRouter(submit handlers.SubmitForApprovalPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		&fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{},
		submit, stubApprovePort{}, stubRejectPort{},
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
	r.POST("/api/curriculum/:id/submit", h.Submit)
	return r
}

func doSubmit(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCurriculumHandler_Submit_HappyPath_Methodist(t *testing.T) {
	submit := &fakeSubmitPort{out: builtCurriculum(t, 7)}
	r := setupSubmitRouter(submit, "methodist", 42)

	rec := doSubmit(t, r, "/api/curriculum/7/submit")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, submit.called)
	assert.Equal(t, int64(42), submit.gotActor)
	assert.False(t, submit.gotIsAdmin, "methodist must not pass isAdmin=true")
	assert.Equal(t, int64(7), submit.gotInput.ID)

	var resp struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestCurriculumHandler_Submit_HappyPath_AdminPropagatesIsAdmin(t *testing.T) {
	submit := &fakeSubmitPort{out: builtCurriculum(t, 7)}
	r := setupSubmitRouter(submit, "system_admin", 99)

	rec := doSubmit(t, r, "/api/curriculum/7/submit")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, submit.gotIsAdmin, "system_admin must propagate isAdmin=true")
}

func TestCurriculumHandler_Submit_RejectsNonWriteRoles(t *testing.T) {
	cases := []string{"teacher", "academic_secretary", "student", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			submit := &fakeSubmitPort{}
			r := setupSubmitRouter(submit, role, 42)

			rec := doSubmit(t, r, "/api/curriculum/7/submit")
			assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			assert.False(t, submit.called)
		})
	}
}

func TestCurriculumHandler_Submit_MissingContextReturns401(t *testing.T) {
	submit := &fakeSubmitPort{}
	r := setupSubmitRouter(submit, "", 0)

	rec := doSubmit(t, r, "/api/curriculum/7/submit")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Submit_BadIDReturns400(t *testing.T) {
	cases := []string{"abc", "0", "-1", "1.5"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			submit := &fakeSubmitPort{}
			r := setupSubmitRouter(submit, "methodist", 42)

			rec := doSubmit(t, r, "/api/curriculum/"+raw+"/submit")
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, submit.called)
		})
	}
}

func TestCurriculumHandler_Submit_DomainErrorMappings(t *testing.T) {
	cases := []struct {
		name  string
		ucErr error
		want  int
	}{
		{"forbidden → 403", entities.ErrCurriculumScopeForbidden, http.StatusForbidden},
		{"not draft → 422", entities.ErrCannotSubmit, http.StatusUnprocessableEntity},
		{"not found → 404", repositories.ErrCurriculumNotFound, http.StatusNotFound},
		{"transport → 500", errors.New("conn refused"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			submit := &fakeSubmitPort{err: tc.ucErr}
			r := setupSubmitRouter(submit, "methodist", 42)

			rec := doSubmit(t, r, "/api/curriculum/7/submit")
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
		})
	}
}
