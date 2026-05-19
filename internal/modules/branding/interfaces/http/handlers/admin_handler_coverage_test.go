package handlers

// v0.153.9 Phase 6 backfill — closes branch gaps in admin_handler.go:
// GetBranding repo-error path, UpdateBranding invalid-body / non-domain
// error → 500 paths, actorIDFromContext int / float64 / absent branches,
// domainErrorCode exhaustive switch, NewAdminBrandingHandler nil-panic
// guards. No production change.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

// errorRepo is a BrandSettingsRepository that returns sentinel errors
// — used to exercise the 500 fallback in GetBranding + UpdateBranding.
type errorRepo struct {
	getErr    error
	updateErr error
	settings  *entities.BrandSettings
}

func (r *errorRepo) Get(_ context.Context) (*entities.BrandSettings, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.settings, nil
}

func (r *errorRepo) Update(_ context.Context, s *entities.BrandSettings) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.settings = s
	return nil
}

func TestAdminBrandingHandler_Get_RepoError_Returns500(t *testing.T) {
	repo := &errorRepo{getErr: fmt.Errorf("db down")}
	r := newTestEngine(t, repo, fakeClock{now: time.Now()}, &fakeAudit{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/branding", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdminBrandingHandler_Update_InvalidJSON_Returns400(t *testing.T) {
	bs, _ := entities.NewBrandSettings("Old", "", "", "", "", "", time.Now())
	r := newTestEngine(t, &fakeRepo{settings: bs}, fakeClock{now: time.Now()}, &fakeAudit{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminBrandingHandler_Update_RepoError_Returns500(t *testing.T) {
	// repo.Update fails with non-domain-validation error → 500 branch.
	bs, _ := entities.NewBrandSettings("Old", "", "", "", "", "", time.Now())
	repo := &errorRepo{
		settings:  bs,
		updateErr: fmt.Errorf("db write failed"),
	}
	r := newTestEngine(t, repo, fakeClock{now: time.Now()}, &fakeAudit{})
	body, _ := json.Marshal(map[string]string{
		"app_name":        "Brand",
		"primary_color":   "#aabbcc",
		"secondary_color": "#445566",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== actorIDFromContext type-switch branches =====

func TestActorIDFromContext_AbsentReturnsZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Equal(t, int64(0), actorIDFromContext(c))
}

func TestActorIDFromContext_Int64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", int64(42))
	assert.Equal(t, int64(42), actorIDFromContext(c))
}

func TestActorIDFromContext_Int(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", int(7))
	assert.Equal(t, int64(7), actorIDFromContext(c))
}

func TestActorIDFromContext_Float64(t *testing.T) {
	// JSON unmarshalling sometimes hydrates ids as float64 — handler
	// recovers gracefully.
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", float64(99))
	assert.Equal(t, int64(99), actorIDFromContext(c))
}

func TestActorIDFromContext_UnsupportedType_Returns0(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", "not-a-number")
	assert.Equal(t, int64(0), actorIDFromContext(c))
}

// ===== domainErrorCode exhaustive =====

func TestDomainErrorCode_Exhaustive(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{entities.ErrInvalidAppName, "INVALID_APP_NAME"},
		{entities.ErrInvalidTagline, "INVALID_TAGLINE"},
		{entities.ErrInvalidColor, "INVALID_COLOR"},
		{entities.ErrInvalidURL, "INVALID_URL"},
		{fmt.Errorf("random"), "INVALID_INPUT"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, domainErrorCode(tc.err), "err=%v", tc.err)
	}
}

func TestIsDomainValidationError(t *testing.T) {
	assert.True(t, isDomainValidationError(entities.ErrInvalidAppName))
	assert.True(t, isDomainValidationError(entities.ErrInvalidTagline))
	assert.True(t, isDomainValidationError(entities.ErrInvalidColor))
	assert.True(t, isDomainValidationError(entities.ErrInvalidURL))
	assert.False(t, isDomainValidationError(errors.New("random")))
}

// ===== NewAdminBrandingHandler nil-panic guards =====

func TestNewAdminBrandingHandler_NilGetUC_Panics(t *testing.T) {
	repo := &fakeRepo{}
	updateUC := usecases.NewUpdateBrandingUseCase(repo, fakeClock{now: time.Now()}, &fakeAudit{})
	assert.PanicsWithValue(t, "branding: nil GetBrandingUseCase", func() {
		NewAdminBrandingHandler(nil, updateUC)
	})
}

func TestNewAdminBrandingHandler_NilUpdateUC_Panics(t *testing.T) {
	repo := &fakeRepo{}
	getUC := usecases.NewGetBrandingUseCase(repo)
	assert.PanicsWithValue(t, "branding: nil UpdateBrandingUseCase", func() {
		NewAdminBrandingHandler(getUC, nil)
	})
}

// Sanity guard — ensures the import chain still works.
var _ = require.NotNil

// ===== PublicBrandingHandler error branch (GetBranding 60% → 100%) =====

func TestPublicBrandingHandler_Get_RepoError_Returns500(t *testing.T) {
	repo := &errorRepo{getErr: fmt.Errorf("db down")}
	getUC := usecases.NewGetBrandingUseCase(repo)
	h := NewPublicBrandingHandler(getUC)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/public/branding", h.GetBranding)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/public/branding", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
