package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

type fakeGenerateOrderRevisions struct {
	result   wpUsecases.GenerateOrderRevisionsResult
	err      error
	called   bool
	gotActor int64
	gotRole  string
	gotID    int64
}

func (f *fakeGenerateOrderRevisions) Execute(
	_ context.Context, actorID int64, role string, orderID int64,
) (wpUsecases.GenerateOrderRevisionsResult, error) {
	f.called = true
	f.gotActor = actorID
	f.gotRole = role
	f.gotID = orderID
	return f.result, f.err
}

func newGenRouter(gen GenerateOrderRevisionsPort, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if gen == nil {
		gen = &fakeGenerateOrderRevisions{}
	}
	h := NewGenerateOrderRevisionsHandler(gen)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterGenerateOrderRevisionsRoutes(api, h)
	return r
}

func TestGenerateOrderRevisionsHandler_HappyPath(t *testing.T) {
	gen := &fakeGenerateOrderRevisions{
		result: wpUsecases.GenerateOrderRevisionsResult{Generated: 3, Skipped: 1, Failures: 0},
	}
	r := newGenRouter(gen, withAuth(42, "methodist"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders/50/generate-revisions", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, gen.called)
	assert.Equal(t, int64(42), gen.gotActor, "actor derives from JWT, not body")
	assert.Equal(t, "methodist", gen.gotRole)
	assert.Equal(t, int64(50), gen.gotID)

	var env map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	data := env["data"].(map[string]any)
	assert.Equal(t, float64(3), data["generated"])
	assert.Equal(t, float64(1), data["skipped"])
	assert.Equal(t, float64(0), data["failures"])
}

func TestGenerateOrderRevisionsHandler_Unauthorized(t *testing.T) {
	r := newGenRouter(nil) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders/50/generate-revisions", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGenerateOrderRevisionsHandler_BadID(t *testing.T) {
	gen := &fakeGenerateOrderRevisions{}
	r := newGenRouter(gen, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders/abc/generate-revisions", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, gen.called)
}

func TestGenerateOrderRevisionsHandler_ErrorMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"forbidden", domain.ErrMinobrnaukiOrderScopeForbidden, http.StatusForbidden},
		{"not_found", repositories.ErrMinobrnaukiOrderNotFound, http.StatusNotFound},
		{"rate_limited", domain.ErrGenerationRateLimited, http.StatusTooManyRequests},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gen := &fakeGenerateOrderRevisions{err: tc.err}
			r := newGenRouter(gen, withAuth(42, "methodist"))
			w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders/50/generate-revisions", nil)
			assert.Equal(t, tc.want, w.Code)
		})
	}
}

func TestNewGenerateOrderRevisionsHandler_PanicsOnNilPort(t *testing.T) {
	assert.Panics(t, func() { NewGenerateOrderRevisionsHandler(nil) })
}
