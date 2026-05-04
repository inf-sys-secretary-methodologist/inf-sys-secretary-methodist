package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
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
