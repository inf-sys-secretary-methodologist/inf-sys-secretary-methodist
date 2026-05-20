package handlers_test

import (
	"context"
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

type fakeApprovePort struct {
	called   bool
	gotAdmin int64
	gotID    int64
	out      *entities.Curriculum
	err      error
}

func (f *fakeApprovePort) Execute(_ context.Context, adminID int64, in curUsecases.ApproveCurriculumInput) (*entities.Curriculum, error) {
	f.called = true
	f.gotAdmin = adminID
	f.gotID = in.ID
	return f.out, f.err
}

func setupApproveRouter(approve handlers.ApproveCurriculumPort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewCurriculumHandler(
		&fakeCreatePort{}, stubGetPort{}, stubListPort{}, stubUpdatePort{},
		stubSubmitPort{}, approve, stubRejectPort{},
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
	r.POST("/api/curriculum/:id/approve", h.Approve)
	return r
}

func doApprove(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCurriculumHandler_Approve_HappyPath_AuthorizedRoles(t *testing.T) {
	// v0.158.0: methodist + system_admin are the authorized approvers
	// per the role matrix (academic_secretary authors, methodist approves,
	// admin retains emergency override).
	cases := []string{"methodist", "system_admin"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			approve := &fakeApprovePort{out: builtCurriculum(t, 7)}
			r := setupApproveRouter(approve, role, 99)

			rec := doApprove(t, r, "/api/curriculum/7/approve")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.True(t, approve.called)
			assert.Equal(t, int64(99), approve.gotAdmin)
			assert.Equal(t, int64(7), approve.gotID)
		})
	}
}

func TestCurriculumHandler_Approve_RejectsNonApproverRoles(t *testing.T) {
	// Defense in depth: even if RequireRole(Methodist, SystemAdmin)
	// middleware were stripped from the route, the handler-level
	// whitelist still rejects every role that is not an approver.
	// v0.158.0: academic_secretary creates curricula, methodist approves —
	// methodist is now an authorized approver (covered by HappyPath test).
	cases := []string{"teacher", "academic_secretary", "student", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			approve := &fakeApprovePort{}
			r := setupApproveRouter(approve, role, 42)

			rec := doApprove(t, r, "/api/curriculum/7/approve")
			assert.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			assert.False(t, approve.called)
		})
	}
}

func TestCurriculumHandler_Approve_MissingContextReturns401(t *testing.T) {
	approve := &fakeApprovePort{}
	r := setupApproveRouter(approve, "", 0)

	rec := doApprove(t, r, "/api/curriculum/7/approve")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func TestCurriculumHandler_Approve_BadIDReturns400(t *testing.T) {
	cases := []string{"abc", "0", "-1", "1.5"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			approve := &fakeApprovePort{}
			r := setupApproveRouter(approve, "system_admin", 99)

			rec := doApprove(t, r, "/api/curriculum/"+raw+"/approve")
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.False(t, approve.called)
		})
	}
}

func TestCurriculumHandler_Approve_DomainErrorMappings(t *testing.T) {
	cases := []struct {
		name  string
		ucErr error
		want  int
	}{
		{"not pending → 422", entities.ErrCannotApprove, http.StatusUnprocessableEntity},
		{"not found → 404", repositories.ErrCurriculumNotFound, http.StatusNotFound},
		{"transport → 500", errors.New("conn refused"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			approve := &fakeApprovePort{err: tc.ucErr}
			r := setupApproveRouter(approve, "system_admin", 99)

			rec := doApprove(t, r, "/api/curriculum/7/approve")
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
		})
	}
}
