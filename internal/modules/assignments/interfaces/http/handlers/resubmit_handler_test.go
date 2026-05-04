package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
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
