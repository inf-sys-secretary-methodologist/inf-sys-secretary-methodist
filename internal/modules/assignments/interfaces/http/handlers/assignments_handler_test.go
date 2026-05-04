package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/interfaces/http/handlers"
)

// --- fake use cases --------------------------------------------------------

type fakeListAssignmentsUC struct {
	out    assignUsecases.ListAssignmentsOutput
	err    error
	called bool
	got    assignUsecases.ListAssignmentsInput
}

func (f *fakeListAssignmentsUC) Execute(ctx context.Context, in assignUsecases.ListAssignmentsInput) (assignUsecases.ListAssignmentsOutput, error) {
	f.called = true
	f.got = in
	return f.out, f.err
}

type fakeGetAssignmentUC struct {
	out    *entities.Assignment
	err    error
	called bool
	got    assignUsecases.GetAssignmentInput
}

func (f *fakeGetAssignmentUC) Execute(ctx context.Context, in assignUsecases.GetAssignmentInput) (*entities.Assignment, error) {
	f.called = true
	f.got = in
	return f.out, f.err
}

type fakeListSubmissionsUC struct {
	out    []views.SubmissionView
	err    error
	called bool
	got    assignUsecases.ListSubmissionsInput
}

func (f *fakeListSubmissionsUC) Execute(ctx context.Context, in assignUsecases.ListSubmissionsInput) ([]views.SubmissionView, error) {
	f.called = true
	f.got = in
	return f.out, f.err
}

// --- helpers ---------------------------------------------------------------

type readAuth struct {
	withUserID bool
	withRole   bool
	userID     int64
	role       string
}

func setupReadRouter(
	listUC handlers.ListAssignmentsUseCasePort,
	getUC handlers.GetAssignmentUseCasePort,
	listSubsUC handlers.ListSubmissionsUseCasePort,
	auth readAuth,
) *gin.Engine {
	r := gin.New()
	h := handlers.NewAssignmentsHandler(listUC, getUC, listSubsUC)
	r.Use(func(c *gin.Context) {
		if auth.withUserID {
			c.Set("user_id", auth.userID)
		}
		if auth.withRole {
			c.Set("role", auth.role)
		}
		c.Next()
	})
	r.GET("/api/assignments", h.ListAssignments)
	r.GET("/api/assignments/:id", h.GetAssignment)
	r.GET("/api/assignments/:id/submissions", h.ListSubmissions)
	return r
}

func makeAssignment(t *testing.T) *entities.Assignment {
	t.Helper()
	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "L1", TeacherID: 42, GroupName: "ИС-21",
		Subject: "Algo", MaxScore: 100, Now: time.Now(),
	})
	require.NoError(t, err)
	a.ID = 10
	return a
}

func doGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- ListAssignments ------------------------------------------------------

func TestListAssignmentsHandler(t *testing.T) {
	tests := []struct {
		name           string
		auth           readAuth
		path           string
		ucOut          assignUsecases.ListAssignmentsOutput
		ucErr          error
		wantStatus     int
		wantUCCalled   bool
		wantUnrestrict bool
		wantSubject    string
	}{
		{
			name:           "teacher gets restricted scope",
			auth:           readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:           "/api/assignments",
			ucOut:          assignUsecases.ListAssignmentsOutput{Items: []*entities.Assignment{makeAssignment(t)}, Total: 1},
			wantStatus:     http.StatusOK,
			wantUCCalled:   true,
			wantUnrestrict: false,
		},
		{
			name:           "methodist gets unrestricted scope",
			auth:           readAuth{withUserID: true, withRole: true, userID: 99, role: "methodist"},
			path:           "/api/assignments",
			ucOut:          assignUsecases.ListAssignmentsOutput{Total: 0},
			wantStatus:     http.StatusOK,
			wantUCCalled:   true,
			wantUnrestrict: true,
		},
		{
			name:           "subject query passed to use case",
			auth:           readAuth{withUserID: true, withRole: true, userID: 1, role: "system_admin"},
			path:           "/api/assignments?subject=Algo",
			ucOut:          assignUsecases.ListAssignmentsOutput{},
			wantStatus:     http.StatusOK,
			wantUCCalled:   true,
			wantUnrestrict: true,
			wantSubject:    "Algo",
		},
		{
			name:       "missing user_id is 401",
			auth:       readAuth{withRole: true, role: "teacher"},
			path:       "/api/assignments",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing role is 401",
			auth:       readAuth{withUserID: true, userID: 42},
			path:       "/api/assignments",
			wantStatus: http.StatusUnauthorized,
		},
		{
			// Defence-in-depth: even if RequireNonStudent middleware were
			// ever bypassed, the handler itself must refuse "student"
			// rather than fall through to "unrestricted" via a negative
			// rule like `role != "teacher"`.
			name:       "role 'student' is 401 (defence-in-depth)",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "student"},
			path:       "/api/assignments",
			wantStatus: http.StatusUnauthorized,
		},
		{
			// Same defence-in-depth: an unrecognised role must NOT be
			// silently treated as unrestricted. Forces the handler to
			// whitelist the four valid non-student roles instead of
			// blacklisting only "teacher".
			name:       "unknown role is 401",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "auditor"},
			path:       "/api/assignments",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:           "academic_secretary gets unrestricted scope",
			auth:           readAuth{withUserID: true, withRole: true, userID: 5, role: "academic_secretary"},
			path:           "/api/assignments",
			ucOut:          assignUsecases.ListAssignmentsOutput{Total: 0},
			wantStatus:     http.StatusOK,
			wantUCCalled:   true,
			wantUnrestrict: true,
		},
		{
			name:           "system_admin gets unrestricted scope",
			auth:           readAuth{withUserID: true, withRole: true, userID: 1, role: "system_admin"},
			path:           "/api/assignments",
			ucOut:          assignUsecases.ListAssignmentsOutput{Total: 0},
			wantStatus:     http.StatusOK,
			wantUCCalled:   true,
			wantUnrestrict: true,
		},
		{
			name:       "use case error becomes 500",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:       "/api/assignments",
			ucErr:      errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantUCCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			listUC := &fakeListAssignmentsUC{out: tc.ucOut, err: tc.ucErr}
			r := setupReadRouter(listUC, &fakeGetAssignmentUC{}, &fakeListSubmissionsUC{}, tc.auth)

			rec := doGet(t, r, tc.path)
			assert.Equal(t, tc.wantStatus, rec.Code, rec.Body.String())

			assert.Equal(t, tc.wantUCCalled, listUC.called)
			if !tc.wantUCCalled {
				return
			}
			assert.Equal(t, tc.wantUnrestrict, listUC.got.Caller.Unrestricted)
			if tc.wantSubject != "" {
				assert.Equal(t, tc.wantSubject, listUC.got.Subject)
			}
		})
	}
}

// --- GetAssignment --------------------------------------------------------

func TestGetAssignmentHandler(t *testing.T) {
	a := makeAssignment(t)

	tests := []struct {
		name       string
		auth       readAuth
		path       string
		ucOut      *entities.Assignment
		ucErr      error
		wantStatus int
	}{
		{
			name:       "200 returns assignment json",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:       "/api/assignments/10",
			ucOut:      a,
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 on non-numeric id",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:       "/api/assignments/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "401 missing user_id",
			auth:       readAuth{withRole: true, role: "teacher"},
			path:       "/api/assignments/10",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "404 on ErrAssignmentNotFound",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "methodist"},
			path:       "/api/assignments/999",
			ucErr:      repositories.ErrAssignmentNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "403 on ErrAssignmentScopeForbidden",
			auth:       readAuth{withUserID: true, withRole: true, userID: 99, role: "teacher"},
			path:       "/api/assignments/10",
			ucErr:      entities.ErrAssignmentScopeForbidden,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getUC := &fakeGetAssignmentUC{out: tc.ucOut, err: tc.ucErr}
			r := setupReadRouter(&fakeListAssignmentsUC{}, getUC, &fakeListSubmissionsUC{}, tc.auth)

			rec := doGet(t, r, tc.path)
			assert.Equal(t, tc.wantStatus, rec.Code, rec.Body.String())

			if tc.wantStatus == http.StatusOK {
				var body struct {
					Data struct {
						ID    int64  `json:"id"`
						Title string `json:"title"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, int64(10), body.Data.ID)
				assert.Equal(t, "L1", body.Data.Title)
			}
		})
	}
}

// --- ListSubmissions ------------------------------------------------------

func TestListSubmissionsHandler(t *testing.T) {
	subs := []views.SubmissionView{
		{ID: 1, AssignmentID: 10, StudentID: 7, StudentName: "Иван", Status: entities.StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, AssignmentID: 10, StudentID: 8, StudentName: "Анна", Status: entities.StatusGraded, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	tests := []struct {
		name         string
		auth         readAuth
		path         string
		ucOut        []views.SubmissionView
		ucErr        error
		wantStatus   int
		wantStatusFP *entities.SubmissionStatus
		wantUCCalled bool
	}{
		{
			name:         "200 returns submissions list",
			auth:         readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:         "/api/assignments/10/submissions",
			ucOut:        subs,
			wantStatus:   http.StatusOK,
			wantUCCalled: true,
		},
		{
			name:         "200 with status=pending filter forwards typed enum",
			auth:         readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:         "/api/assignments/10/submissions?status=pending",
			ucOut:        subs[:1],
			wantStatus:   http.StatusOK,
			wantStatusFP: ptrStatus(entities.StatusPending),
			wantUCCalled: true,
		},
		{
			name:       "400 on invalid status",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:       "/api/assignments/10/submissions?status=bogus",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 on non-numeric id",
			auth:       readAuth{withUserID: true, withRole: true, userID: 42, role: "teacher"},
			path:       "/api/assignments/abc/submissions",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "401 missing user_id",
			auth:       readAuth{withRole: true, role: "teacher"},
			path:       "/api/assignments/10/submissions",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "404 on ErrAssignmentNotFound",
			auth:         readAuth{withUserID: true, withRole: true, userID: 42, role: "methodist"},
			path:         "/api/assignments/999/submissions",
			ucErr:        repositories.ErrAssignmentNotFound,
			wantStatus:   http.StatusNotFound,
			wantUCCalled: true,
		},
		{
			name:         "403 on ErrAssignmentScopeForbidden",
			auth:         readAuth{withUserID: true, withRole: true, userID: 99, role: "teacher"},
			path:         "/api/assignments/10/submissions",
			ucErr:        entities.ErrAssignmentScopeForbidden,
			wantStatus:   http.StatusForbidden,
			wantUCCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			listSubsUC := &fakeListSubmissionsUC{out: tc.ucOut, err: tc.ucErr}
			r := setupReadRouter(&fakeListAssignmentsUC{}, &fakeGetAssignmentUC{}, listSubsUC, tc.auth)

			rec := doGet(t, r, tc.path)
			assert.Equal(t, tc.wantStatus, rec.Code, rec.Body.String())

			assert.Equal(t, tc.wantUCCalled, listSubsUC.called)
			if tc.wantStatusFP != nil {
				require.NotNil(t, listSubsUC.got.Status)
				assert.Equal(t, *tc.wantStatusFP, *listSubsUC.got.Status)
			}
		})
	}
}

func ptrStatus(s entities.SubmissionStatus) *entities.SubmissionStatus { return &s }
