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

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/interfaces/http/handlers"
)

type fakeResubmitUseCase struct {
	err        error
	called     bool
	gotActorID int64
	gotInput   assignUsecases.ResubmitSubmissionInput
}

func (f *fakeResubmitUseCase) Execute(ctx context.Context, actorID int64, in assignUsecases.ResubmitSubmissionInput) error {
	f.called = true
	f.gotActorID = actorID
	f.gotInput = in
	return f.err
}

func setupResubmitRouter(uc handlers.ResubmitSubmissionUseCasePort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewResubmitHandler(uc)
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("role", role)
			c.Next()
		})
	}
	r.POST("/api/assignments/:id/resubmit", h.Resubmit)
	return r
}

func doResubmitRequest(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// TestResubmitHandler_RoleWhitelist documents the failure-closed role
// matrix: ONLY "student" is permitted to resubmit; every other role
// (including missing auth context) falls through to 401. The use-case
// must NOT be invoked when access is denied — defence in depth that
// mirrors the read-side AssignmentsHandler / GradeHandler / ReturnHandler.
func TestResubmitHandler_RoleWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		role     string // "" means no auth middleware at all
		wantCode int
		wantUC   bool
	}{
		{name: "student allowed", role: "student", wantCode: http.StatusOK, wantUC: true},
		{name: "teacher rejected", role: "teacher", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "methodist rejected", role: "methodist", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "academic_secretary rejected", role: "academic_secretary", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "system_admin rejected", role: "system_admin", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "unknown role rejected", role: "auditor", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "empty-string role rejected", role: "garbage-no-auth", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "no auth middleware → no user_id → 401", role: "", wantCode: http.StatusUnauthorized, wantUC: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeResubmitUseCase{}
			r := setupResubmitRouter(uc, tc.role, 7)

			rec := doResubmitRequest(t, r, "/api/assignments/10/resubmit")

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			assert.Equal(t, tc.wantUC, uc.called,
				"use case invocation must align with whitelist outcome")
		})
	}
}

// TestResubmitHandler_InputValidationAndErrorMapping exercises both the
// HTTP-layer parsing gates (bad path id) and the domain-error → HTTP
// status mapping for every sentinel the use case can surface. Generic
// errors fall through to 500 via response.MapDomainError. Mirrors
// TestReturnHandler_InputValidationAndErrorMapping in shape.
func TestResubmitHandler_InputValidationAndErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		ucErr    error
		wantCode int
	}{
		// Input validation — short-circuit before the use case is touched.
		{name: "non-numeric assignment id → 400",
			path: "/api/assignments/abc/resubmit", wantCode: http.StatusBadRequest},
		{name: "zero assignment id → 400",
			path: "/api/assignments/0/resubmit", wantCode: http.StatusBadRequest},
		{name: "negative assignment id → 400",
			path: "/api/assignments/-1/resubmit", wantCode: http.StatusBadRequest},

		// Error mapping (use case returns sentinel)
		{name: "ErrAssignmentNotFound → 404",
			path:  "/api/assignments/10/resubmit",
			ucErr: repositories.ErrAssignmentNotFound, wantCode: http.StatusNotFound},
		{name: "ErrSubmissionNotFound → 404",
			path:  "/api/assignments/10/resubmit",
			ucErr: repositories.ErrSubmissionNotFound, wantCode: http.StatusNotFound},
		{name: "ErrSubmissionOwnerOnly → 403",
			path:  "/api/assignments/10/resubmit",
			ucErr: entities.ErrSubmissionOwnerOnly, wantCode: http.StatusForbidden},
		{name: "ErrNotReturned → 409",
			path:  "/api/assignments/10/resubmit",
			ucErr: entities.ErrNotReturned, wantCode: http.StatusConflict},
		{name: "generic error → 500",
			path:  "/api/assignments/10/resubmit",
			ucErr: errors.New("boom"), wantCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeResubmitUseCase{err: tc.ucErr}
			r := setupResubmitRouter(uc, "student", 7)

			rec := doResubmitRequest(t, r, tc.path)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
		})
	}
}

// TestResubmitHandler_NilUseCasePanics ensures the constructor fails loud
// at wiring time rather than letting requests reach a nil-deref panic
// deeper in the call stack. Mirrors TestReturnHandler_NilUseCasePanics.
func TestResubmitHandler_NilUseCasePanics(t *testing.T) {
	assert.Panics(t, func() {
		handlers.NewResubmitHandler(nil)
	}, "NewResubmitHandler must panic on nil usecase to surface DI mistakes early")
}

// TestResubmitHandler_SuccessResponseEchoesPayload locks down the success
// response body so the frontend can reflect the just-resubmitted ids
// without an extra round-trip. Mirrors the same lock-in on the return
// side.
func TestResubmitHandler_SuccessResponseEchoesPayload(t *testing.T) {
	uc := &fakeResubmitUseCase{}
	r := setupResubmitRouter(uc, "student", 7)

	rec := doResubmitRequest(t, r, "/api/assignments/10/resubmit")

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			AssignmentID int64 `json:"assignment_id"`
			StudentID    int64 `json:"student_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(10), resp.Data.AssignmentID)
	assert.Equal(t, int64(7), resp.Data.StudentID)
}

// TestResubmitHandler_HappyPath asserts the core student-driven request:
// authenticated as student, valid assignment id in path, no body needed
// (the student supplies no input — they are simply re-submitting their
// own returned work). The handler must derive the studentID from the
// JWT context (= actorID) and invoke the use case with both ids set
// consistently. A nil error from the use case yields HTTP 200.
func TestResubmitHandler_HappyPath(t *testing.T) {
	uc := &fakeResubmitUseCase{}
	r := setupResubmitRouter(uc, "student", 7)

	rec := doResubmitRequest(t, r, "/api/assignments/10/resubmit")

	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.True(t, uc.called, "use case must be invoked on happy path")
	assert.Equal(t, int64(7), uc.gotActorID)
	assert.Equal(t, int64(10), uc.gotInput.AssignmentID)
	assert.Equal(t, int64(7), uc.gotInput.StudentID,
		"student_id must be derived from JWT context, not from request body")
}
