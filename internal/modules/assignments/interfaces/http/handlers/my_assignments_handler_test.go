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

// --- Fakes for the two narrow use-case ports ---

type fakeListMyAssignmentsUC struct {
	out      []views.StudentAssignmentView
	err      error
	called   bool
	gotInput assignUsecases.ListMyAssignmentsInput
}

func (f *fakeListMyAssignmentsUC) Execute(ctx context.Context, in assignUsecases.ListMyAssignmentsInput) ([]views.StudentAssignmentView, error) {
	f.called = true
	f.gotInput = in
	return f.out, f.err
}

type fakeGetMyAssignmentDetailUC struct {
	out      *views.StudentAssignmentView
	err      error
	called   bool
	gotInput assignUsecases.GetMyAssignmentDetailInput
}

func (f *fakeGetMyAssignmentDetailUC) Execute(ctx context.Context, in assignUsecases.GetMyAssignmentDetailInput) (*views.StudentAssignmentView, error) {
	f.called = true
	f.gotInput = in
	return f.out, f.err
}

func setupMyAssignmentsRouter(t *testing.T, listUC handlers.ListMyAssignmentsUseCasePort, detailUC handlers.GetMyAssignmentDetailUseCasePort, role string, userID int64) *gin.Engine {
	t.Helper()
	r := gin.New()
	h := handlers.NewMyAssignmentsHandler(listUC, detailUC)
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("role", role)
			c.Next()
		})
	}
	r.GET("/api/assignments/my", h.List)
	r.GET("/api/assignments/:id/my", h.Detail)
	return r
}

func doGetMy(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// TestMyAssignmentsHandler_RoleWhitelist — only "student" reaches the use
// case. Every other role (or missing context) → 401, defense-in-depth on
// top of the studentAssignmentsGroup RequireRole("student") middleware.
func TestMyAssignmentsHandler_RoleWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		path     string
		wantCode int
		wantUC   bool
	}{
		{name: "student list allowed", role: "student", path: "/api/assignments/my", wantCode: http.StatusOK, wantUC: true},
		{name: "student detail allowed", role: "student", path: "/api/assignments/10/my", wantCode: http.StatusOK, wantUC: true},
		{name: "teacher list rejected", role: "teacher", path: "/api/assignments/my", wantCode: http.StatusUnauthorized},
		{name: "teacher detail rejected", role: "teacher", path: "/api/assignments/10/my", wantCode: http.StatusUnauthorized},
		{name: "system_admin list rejected", role: "system_admin", path: "/api/assignments/my", wantCode: http.StatusUnauthorized},
		{name: "case-mismatched 'Student' rejected", role: "Student", path: "/api/assignments/my", wantCode: http.StatusUnauthorized},
		{name: "no auth context list", role: "", path: "/api/assignments/my", wantCode: http.StatusUnauthorized},
		{name: "no auth context detail", role: "", path: "/api/assignments/10/my", wantCode: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			listUC := &fakeListMyAssignmentsUC{out: []views.StudentAssignmentView{}}
			detailUC := &fakeGetMyAssignmentDetailUC{out: &views.StudentAssignmentView{AssignmentID: 10, StudentID: 7, Status: entities.StatusPending}}
			r := setupMyAssignmentsRouter(t, listUC, detailUC, tc.role, 7)

			rec := doGetMy(t, r, tc.path)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			anyCalled := listUC.called || detailUC.called
			assert.Equal(t, tc.wantUC, anyCalled, "use case invocation must align with whitelist outcome")
		})
	}
}

// TestMyAssignmentsHandler_DetailInputValidation — the detail endpoint
// rejects malformed assignment ids before touching the use case.
func TestMyAssignmentsHandler_DetailInputValidation(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{name: "non-numeric id → 400", path: "/api/assignments/abc/my", wantCode: http.StatusBadRequest},
		{name: "zero id → 400", path: "/api/assignments/0/my", wantCode: http.StatusBadRequest},
		{name: "negative id → 400", path: "/api/assignments/-1/my", wantCode: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			detailUC := &fakeGetMyAssignmentDetailUC{}
			r := setupMyAssignmentsRouter(t, &fakeListMyAssignmentsUC{}, detailUC, "student", 7)

			rec := doGetMy(t, r, tc.path)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			assert.False(t, detailUC.called, "use case must not be invoked on bad path")
		})
	}
}

// TestMyAssignmentsHandler_DetailErrorMapping — sentinel-first mapping,
// generic falls through to 500.
func TestMyAssignmentsHandler_DetailErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		ucErr    error
		wantCode int
	}{
		{name: "ErrAssignmentNotFound → 404", ucErr: repositories.ErrAssignmentNotFound, wantCode: http.StatusNotFound},
		{name: "ErrSubmissionNotFound → 404", ucErr: repositories.ErrSubmissionNotFound, wantCode: http.StatusNotFound},
		{name: "ErrSubmissionOwnerOnly → 403", ucErr: entities.ErrSubmissionOwnerOnly, wantCode: http.StatusForbidden},
		{name: "generic → 500", ucErr: errors.New("boom"), wantCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			detailUC := &fakeGetMyAssignmentDetailUC{err: tc.ucErr}
			r := setupMyAssignmentsRouter(t, &fakeListMyAssignmentsUC{}, detailUC, "student", 7)

			rec := doGetMy(t, r, "/api/assignments/10/my")
			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
		})
	}
}

// TestMyAssignmentsHandler_ListErrorMapping — list endpoint surfaces
// generic errors as 500 (no specific sentinels in the read path beyond
// repository-layer errors).
func TestMyAssignmentsHandler_ListErrorMapping(t *testing.T) {
	listUC := &fakeListMyAssignmentsUC{err: errors.New("boom")}
	r := setupMyAssignmentsRouter(t, listUC, &fakeGetMyAssignmentDetailUC{}, "student", 7)

	rec := doGetMy(t, r, "/api/assignments/my")
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

// TestMyAssignmentsHandler_NilDepsPanics — failure-closed DI.
func TestMyAssignmentsHandler_NilDepsPanics(t *testing.T) {
	listUC := &fakeListMyAssignmentsUC{}
	detailUC := &fakeGetMyAssignmentDetailUC{}

	assert.Panics(t, func() { handlers.NewMyAssignmentsHandler(nil, detailUC) })
	assert.Panics(t, func() { handlers.NewMyAssignmentsHandler(listUC, nil) })
}

// TestMyAssignmentsHandler_ListSuccessEcho — happy path returns a JSON
// envelope with items + total. Status query parameter, when supplied,
// flows into the use-case input as a typed pointer.
func TestMyAssignmentsHandler_ListSuccessEcho(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	listUC := &fakeListMyAssignmentsUC{
		out: []views.StudentAssignmentView{
			{AssignmentID: 10, Title: "Lab 1", Subject: "Math", GroupName: "БСБО-01-22", MaxScore: 100, SubmissionID: 1, StudentID: 7, Status: entities.StatusPending, AssignmentCreatedAt: now, AssignmentUpdatedAt: now, SubmissionCreatedAt: now, SubmissionUpdatedAt: now},
			{AssignmentID: 11, Title: "Lab 2", Subject: "Math", GroupName: "БСБО-01-22", MaxScore: 50, SubmissionID: 2, StudentID: 7, Status: entities.StatusReturned, ReturnReason: "redo", AssignmentCreatedAt: now, AssignmentUpdatedAt: now, SubmissionCreatedAt: now, SubmissionUpdatedAt: now},
		},
	}
	r := setupMyAssignmentsRouter(t, listUC, &fakeGetMyAssignmentDetailUC{}, "student", 7)

	rec := doGetMy(t, r, "/api/assignments/my?status=returned")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.True(t, listUC.called)
	assert.Equal(t, int64(7), listUC.gotInput.StudentID)
	require.NotNil(t, listUC.gotInput.Status)
	assert.Equal(t, entities.StatusReturned, *listUC.gotInput.Status)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data.Items, 2)
	assert.Equal(t, 2, resp.Data.Total)
}

// TestMyAssignmentsHandler_ListUnknownStatusRejected — typed enum
// validation: unknown status → 400, use case never invoked.
func TestMyAssignmentsHandler_ListUnknownStatusRejected(t *testing.T) {
	listUC := &fakeListMyAssignmentsUC{}
	r := setupMyAssignmentsRouter(t, listUC, &fakeGetMyAssignmentDetailUC{}, "student", 7)

	rec := doGetMy(t, r, "/api/assignments/my?status=bogus")
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.False(t, listUC.called)
}

// TestMyAssignmentsHandler_DetailSuccessEcho — happy path returns the
// denormalised view fields as JSON. Locks down the contract the
// frontend depends on.
func TestMyAssignmentsHandler_DetailSuccessEcho(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	gradedAt := now
	gv := 85
	gb := int64(42)
	detailUC := &fakeGetMyAssignmentDetailUC{
		out: &views.StudentAssignmentView{
			AssignmentID:        10,
			Title:               "Lab 1",
			Description:         "Solve A",
			Subject:             "Math",
			GroupName:           "БСБО-01-22",
			MaxScore:            100,
			AssignmentCreatedAt: now,
			AssignmentUpdatedAt: now,
			SubmissionID:        1,
			StudentID:           7,
			GradeValue:          &gv,
			Feedback:            "good",
			GradedBy:            &gb,
			GradedAt:            &gradedAt,
			Status:              entities.StatusGraded,
			SubmissionCreatedAt: now,
			SubmissionUpdatedAt: now,
		},
	}
	r := setupMyAssignmentsRouter(t, &fakeListMyAssignmentsUC{}, detailUC, "student", 7)

	rec := doGetMy(t, r, "/api/assignments/10/my")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.True(t, detailUC.called)
	assert.Equal(t, int64(7), detailUC.gotInput.StudentID, "studentID must come from JWT, not path")
	assert.Equal(t, int64(10), detailUC.gotInput.AssignmentID)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			AssignmentID int64  `json:"assignment_id"`
			Title        string `json:"title"`
			Status       string `json:"status"`
			GradeValue   *int   `json:"grade_value"`
			Feedback     string `json:"feedback"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(10), resp.Data.AssignmentID)
	assert.Equal(t, "Lab 1", resp.Data.Title)
	assert.Equal(t, "graded", resp.Data.Status)
	require.NotNil(t, resp.Data.GradeValue)
	assert.Equal(t, 85, *resp.Data.GradeValue)
	assert.Equal(t, "good", resp.Data.Feedback)
}
