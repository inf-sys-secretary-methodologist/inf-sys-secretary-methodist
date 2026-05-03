package handlers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// mockTeacherScopeRepo is a testify-mock implementation of
// repositories.TeacherScopeRepository, used by handler-level tests to
// pin scope-assembly behaviour.
type mockTeacherScopeRepo struct{ mock.Mock }

func (m *mockTeacherScopeRepo) ListGroupNames(ctx context.Context, teacherID int64) ([]string, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// stubAuth seeds gin's request context the way auth_middleware.JWTMiddleware
// would after token validation: user_id (int64) and role (string).
func stubAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// --- Scope assembly: teacher role calls scopeRepo, others bypass it ---

func TestAnalyticsHandler_GetGroupSummary_BuildsTeacherScopeForTeacherRole(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	sr := new(mockTeacherScopeRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

	sr.On("ListGroupNames", mock.Anything, int64(7)).Return([]string{"ИС-21"}, nil)
	ar.On("GetGroupSummary", mock.Anything, "ИС-21").
		Return(&entities.GroupAnalyticsSummary{GroupName: "ИС-21", TotalStudents: 30}, nil)

	router := setupRouter()
	router.Use(stubAuth(7, "teacher"))
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/%D0%98%D0%A1-21/summary", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	sr.AssertCalled(t, "ListGroupNames", mock.Anything, int64(7))
}

func TestAnalyticsHandler_GetGroupSummary_NonTeacherSkipsScopeRepo(t *testing.T) {
	tests := []string{"methodist", "academic_secretary", "system_admin"}

	for _, role := range tests {
		t.Run(role, func(t *testing.T) {
			ar := new(mockAnalyticsRepo)
			sr := new(mockTeacherScopeRepo)
			h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

			ar.On("GetGroupSummary", mock.Anything, "G1").
				Return(&entities.GroupAnalyticsSummary{GroupName: "G1"}, nil)

			router := setupRouter()
			router.Use(stubAuth(42, role))
			router.GET("/groups/:name/summary", h.GetGroupSummary)
			w := performRequest(router, http.MethodGet, "/groups/G1/summary", nil)

			assert.Equal(t, http.StatusOK, w.Code)
			sr.AssertNotCalled(t, "ListGroupNames", mock.Anything, mock.Anything)
		})
	}
}

func TestAnalyticsHandler_GetGroupSummary_403WhenTeacherTargetsForeignGroup(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	sr := new(mockTeacherScopeRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

	// Teacher 7 teaches ИС-21 only.
	sr.On("ListGroupNames", mock.Anything, int64(7)).Return([]string{"ИС-21"}, nil)
	// Repository must NOT be reached — the use case rejects before any repo call.

	router := setupRouter()
	router.Use(stubAuth(7, "teacher"))
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/%D0%9F%D0%98-31/summary", nil) // ПИ-31

	assert.Equal(t, http.StatusForbidden, w.Code)
	ar.AssertNotCalled(t, "GetGroupSummary", mock.Anything, mock.Anything)
}

func TestAnalyticsHandler_GetGroupSummary_500WhenScopeRepoFails(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	sr := new(mockTeacherScopeRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

	sr.On("ListGroupNames", mock.Anything, int64(7)).
		Return(nil, fmt.Errorf("connection refused"))

	router := setupRouter()
	router.Use(stubAuth(7, "teacher"))
	router.GET("/groups/:name/summary", h.GetGroupSummary)
	w := performRequest(router, http.MethodGet, "/groups/G/summary", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ar.AssertNotCalled(t, "GetGroupSummary", mock.Anything, mock.Anything)
}

// --- Specific-resource endpoints map sentinel error to 403 ---

func TestAnalyticsHandler_GetStudentRisk_403WhenForeignStudent(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	sr := new(mockTeacherScopeRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

	// Teacher 7 teaches ИС-21; student 99 belongs to ПИ-31.
	sr.On("ListGroupNames", mock.Anything, int64(7)).Return([]string{"ИС-21"}, nil)
	pi31 := "ПИ-31"
	ar.On("GetStudentRisk", mock.Anything, int64(99)).
		Return(&entities.StudentRiskScore{StudentID: 99, GroupName: &pi31}, nil)

	router := setupRouter()
	router.Use(stubAuth(7, "teacher"))
	router.GET("/students/:id/risk", h.GetStudentRisk)
	w := performRequest(router, http.MethodGet, "/students/99/risk", nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- List endpoints pass scope through (no 403, just filtering) ---

func TestAnalyticsHandler_GetAtRiskStudents_TeacherScopeReachesRepo(t *testing.T) {
	ar := new(mockAnalyticsRepo)
	sr := new(mockTeacherScopeRepo)
	h := NewAnalyticsHandler(newUC(ar, new(mockAttendanceRepo), new(mockGradeRepo)), sr)

	sr.On("ListGroupNames", mock.Anything, int64(7)).Return([]string{"ИС-21"}, nil)
	// Match ANY non-nil *entities.TeacherScope (the constructor returns
	// a fresh pointer; we only care that scope reached the repo).
	ar.On("GetAtRiskStudents", mock.Anything, mock.MatchedBy(func(s *entities.TeacherScope) bool {
		return s != nil && s.AllowsGroup("ИС-21")
	}), 20, 0).Return([]entities.StudentRiskScore{}, int64(0), nil)

	router := setupRouter()
	router.Use(stubAuth(7, "teacher"))
	router.GET("/at-risk", h.GetAtRiskStudents)
	w := performRequest(router, http.MethodGet, "/at-risk", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	ar.AssertExpectations(t)
}
