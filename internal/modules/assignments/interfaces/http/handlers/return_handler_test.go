package handlers_test

import (
	"bytes"
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

type fakeReturnUseCase struct {
	err        error
	called     bool
	gotActorID int64
	gotInput   assignUsecases.ReturnSubmissionInput
}

func (f *fakeReturnUseCase) Execute(ctx context.Context, actorID int64, in assignUsecases.ReturnSubmissionInput) error {
	f.called = true
	f.gotActorID = actorID
	f.gotInput = in
	return f.err
}

func setupReturnRouter(uc handlers.ReturnSubmissionUseCasePort, role string, userID int64) *gin.Engine {
	r := gin.New()
	h := handlers.NewReturnHandler(uc)
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("role", role)
			c.Next()
		})
	}
	r.POST("/api/assignments/:id/returns", h.Return)
	return r
}

func doReturnRequest(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
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

func TestReturnHandler_HappyPath(t *testing.T) {
	uc := &fakeReturnUseCase{}
	r := setupReturnRouter(uc, "teacher", 42)

	body := map[string]any{"student_id": 7, "reason": "revisit derivation"}
	rec := doReturnRequest(t, r, "/api/assignments/10/returns", body)

	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.True(t, uc.called, "use case must be invoked on happy path")
	assert.Equal(t, int64(42), uc.gotActorID)
	assert.Equal(t, int64(10), uc.gotInput.AssignmentID)
	assert.Equal(t, int64(7), uc.gotInput.StudentID)
	assert.Equal(t, "revisit derivation", uc.gotInput.Reason)
}

// TestReturnHandler_RoleWhitelist documents the failure-closed role
// matrix: only the four EDIT_ROLES are permitted to return submissions;
// anything else (including missing auth context) falls through to 401.
// The use-case must NOT be invoked when access is denied — defence in
// depth that mirrors the read-side AssignmentsHandler / GradeHandler.
func TestReturnHandler_RoleWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		role     string // "" means no auth middleware at all
		wantCode int
		wantUC   bool // whether use case should be invoked
	}{
		{name: "teacher allowed", role: "teacher", wantCode: http.StatusOK, wantUC: true},
		{name: "methodist allowed", role: "methodist", wantCode: http.StatusOK, wantUC: true},
		{name: "academic_secretary allowed", role: "academic_secretary", wantCode: http.StatusOK, wantUC: true},
		{name: "system_admin allowed", role: "system_admin", wantCode: http.StatusOK, wantUC: true},
		{name: "student rejected", role: "student", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "auditor rejected", role: "auditor", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "empty-string role rejected", role: "garbage-no-auth", wantCode: http.StatusUnauthorized, wantUC: false},
		{name: "no auth middleware → no user_id → 401", role: "", wantCode: http.StatusUnauthorized, wantUC: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeReturnUseCase{}
			r := setupReturnRouter(uc, tc.role, 42)

			body := map[string]any{"student_id": 7, "reason": "x"}
			rec := doReturnRequest(t, r, "/api/assignments/10/returns", body)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
			assert.Equal(t, tc.wantUC, uc.called,
				"use case invocation must align with whitelist outcome")
		})
	}
}

// TestReturnHandler_InputValidationAndErrorMapping exercises both the
// HTTP-layer parsing/validation gates (bad path id, malformed body,
// missing/non-positive student_id) and the domain-error → HTTP-status
// mapping for every sentinel the use case can surface. Generic errors
// must fall through to 500 via response.MapDomainError.
func TestReturnHandler_InputValidationAndErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		body     any
		ucErr    error
		wantCode int
	}{
		// Input validation
		{name: "non-numeric assignment id → 400",
			path: "/api/assignments/abc/returns",
			body: map[string]any{"student_id": 7, "reason": "x"}, wantCode: http.StatusBadRequest},
		{name: "malformed json body → 400",
			path: "/api/assignments/10/returns",
			body: "not-json", wantCode: http.StatusBadRequest},
		{name: "missing student_id → 400",
			path: "/api/assignments/10/returns",
			body: map[string]any{"reason": "x"}, wantCode: http.StatusBadRequest},
		{name: "negative student_id → 400",
			path: "/api/assignments/10/returns",
			body: map[string]any{"student_id": -1, "reason": "x"}, wantCode: http.StatusBadRequest},
		{name: "zero student_id → 400",
			path: "/api/assignments/10/returns",
			body: map[string]any{"student_id": 0, "reason": "x"}, wantCode: http.StatusBadRequest},

		// Error mapping (use case returns sentinel)
		{name: "ErrAssignmentNotFound → 404",
			path:  "/api/assignments/10/returns",
			body:  map[string]any{"student_id": 7, "reason": "x"},
			ucErr: repositories.ErrAssignmentNotFound, wantCode: http.StatusNotFound},
		{name: "ErrAssignmentScopeForbidden → 403",
			path:  "/api/assignments/10/returns",
			body:  map[string]any{"student_id": 7, "reason": "x"},
			ucErr: entities.ErrAssignmentScopeForbidden, wantCode: http.StatusForbidden},
		{name: "ErrAlreadyReturned → 409",
			path:  "/api/assignments/10/returns",
			body:  map[string]any{"student_id": 7, "reason": "x"},
			ucErr: entities.ErrAlreadyReturned, wantCode: http.StatusConflict},
		{name: "ErrInvalidReturn → 422",
			path:  "/api/assignments/10/returns",
			body:  map[string]any{"student_id": 7, "reason": ""},
			ucErr: entities.ErrInvalidReturn, wantCode: http.StatusUnprocessableEntity},
		{name: "generic error → 500",
			path:  "/api/assignments/10/returns",
			body:  map[string]any{"student_id": 7, "reason": "x"},
			ucErr: errors.New("boom"), wantCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := &fakeReturnUseCase{err: tc.ucErr}
			r := setupReturnRouter(uc, "teacher", 42)

			rec := doReturnRequest(t, r, tc.path, tc.body)

			assert.Equal(t, tc.wantCode, rec.Code, rec.Body.String())
		})
	}
}

// TestReturnHandler_NilUseCasePanics ensures the constructor fails loud
// at wiring time rather than letting requests reach a nil-deref panic
// deeper in the call stack. Mirrors TestSaveGradeHandler_NilUseCasePanics.
func TestReturnHandler_NilUseCasePanics(t *testing.T) {
	assert.Panics(t, func() {
		handlers.NewReturnHandler(nil)
	}, "NewReturnHandler must panic on nil usecase to surface DI mistakes early")
}

// TestReturnHandler_SuccessResponseEchoesPayload locks down the success
// response body so the frontend can reflect the just-returned assignment
// without an extra round-trip. T11 reviewer should-fix.
func TestReturnHandler_SuccessResponseEchoesPayload(t *testing.T) {
	uc := &fakeReturnUseCase{}
	r := setupReturnRouter(uc, "teacher", 42)

	body := map[string]any{"student_id": 7, "reason": "revisit derivation"}
	rec := doReturnRequest(t, r, "/api/assignments/10/returns", body)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			AssignmentID int64  `json:"assignment_id"`
			StudentID    int64  `json:"student_id"`
			Reason       string `json:"reason"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(10), resp.Data.AssignmentID)
	assert.Equal(t, int64(7), resp.Data.StudentID)
	assert.Equal(t, "revisit derivation", resp.Data.Reason)
}
