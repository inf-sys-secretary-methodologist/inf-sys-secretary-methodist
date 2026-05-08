package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	authhttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
)

// stubMFAService records the last call and returns scripted results so each
// handler test can assert wiring without spinning up the full use-case stack.
type stubMFAService struct {
	beginURI    string
	beginSecret string
	beginErr    error

	confirmErr error
	disableErr error

	lastBeginUserID   int64
	lastConfirmUserID int64
	lastConfirmCode   string
	lastDisableUserID int64
	lastDisableCode   string
}

func (s *stubMFAService) BeginEnrollment(_ context.Context, userID int64) (string, string, error) {
	s.lastBeginUserID = userID
	return s.beginURI, s.beginSecret, s.beginErr
}
func (s *stubMFAService) ConfirmEnrollment(_ context.Context, userID int64, code string) error {
	s.lastConfirmUserID = userID
	s.lastConfirmCode = code
	return s.confirmErr
}
func (s *stubMFAService) Disable(_ context.Context, userID int64, code string) error {
	s.lastDisableUserID = userID
	s.lastDisableCode = code
	return s.disableErr
}

// withAuthedAdmin installs middleware that pretends JWTMiddleware ran and
// stamped a system_admin user_id on the gin context.
func withAuthedAdmin(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", "system_admin")
		c.Next()
	}
}

func newRouter(svc *stubMFAService, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := authhttp.NewMFAHandler(svc)
	g := r.Group("/api/auth/mfa", mw...)
	g.POST("/begin", h.Begin)
	g.POST("/confirm", h.Confirm)
	g.POST("/disable", h.Disable)
	return r
}

func performJSON(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// --- Begin -------------------------------------------------------------------

func TestMFAHandler_Begin(t *testing.T) {
	t.Run("200 on success returns otpauth uri and secret", func(t *testing.T) {
		svc := &stubMFAService{
			beginURI:    "otpauth://totp/inf:admin@x?secret=ABCD",
			beginSecret: "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
		}
		r := newRouter(svc, withAuthedAdmin(99))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/begin", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status: want 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Data struct {
				OTPAuthURI string `json:"otpauth_uri"`
				Secret     string `json:"secret"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Data.OTPAuthURI != svc.beginURI {
			t.Errorf("otpauth_uri: want %q, got %q", svc.beginURI, resp.Data.OTPAuthURI)
		}
		if resp.Data.Secret != svc.beginSecret {
			t.Errorf("secret: want %q, got %q", svc.beginSecret, resp.Data.Secret)
		}
		if svc.lastBeginUserID != 99 {
			t.Errorf("must forward user_id from context; got %d", svc.lastBeginUserID)
		}
	})

	t.Run("409 when MFA already enabled", func(t *testing.T) {
		svc := &stubMFAService{beginErr: entities.ErrMFAAlreadyEnabled}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/begin", "")
		if rec.Code != http.StatusConflict {
			t.Errorf("status: want 409, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("500 on opaque error", func(t *testing.T) {
		svc := &stubMFAService{beginErr: errors.New("redis down")}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/begin", "")
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status: want 500, got %d", rec.Code)
		}
	})

	t.Run("401 when user_id missing from context", func(t *testing.T) {
		svc := &stubMFAService{}
		r := newRouter(svc) // no auth middleware
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/begin", "")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status: want 401, got %d", rec.Code)
		}
	})
}

// --- Confirm -----------------------------------------------------------------

func TestMFAHandler_Confirm(t *testing.T) {
	t.Run("200 on success forwards code", func(t *testing.T) {
		svc := &stubMFAService{}
		r := newRouter(svc, withAuthedAdmin(7))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/confirm", `{"code":"123456"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: want 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastConfirmCode != "123456" {
			t.Errorf("code: want 123456, got %q", svc.lastConfirmCode)
		}
		if svc.lastConfirmUserID != 7 {
			t.Errorf("user_id: want 7, got %d", svc.lastConfirmUserID)
		}
	})

	t.Run("400 when body missing or code empty", func(t *testing.T) {
		cases := []string{"", `{}`, `{"code":""}`, `{"code":"abc"}`}
		for _, body := range cases {
			t.Run(body, func(t *testing.T) {
				svc := &stubMFAService{}
				r := newRouter(svc, withAuthedAdmin(1))
				rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/confirm", body)
				if rec.Code != http.StatusBadRequest {
					t.Errorf("status: want 400, got %d (body %q)", rec.Code, body)
				}
			})
		}
	})

	t.Run("422 on invalid code", func(t *testing.T) {
		svc := &stubMFAService{confirmErr: entities.ErrInvalidMFACode}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/confirm", `{"code":"000000"}`)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status: want 422, got %d", rec.Code)
		}
	})

	t.Run("409 on ErrMFANotPending", func(t *testing.T) {
		svc := &stubMFAService{confirmErr: entities.ErrMFANotPending}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/confirm", `{"code":"123456"}`)
		if rec.Code != http.StatusConflict {
			t.Errorf("status: want 409, got %d", rec.Code)
		}
	})

	t.Run("409 on ErrMFAAlreadyEnabled", func(t *testing.T) {
		svc := &stubMFAService{confirmErr: entities.ErrMFAAlreadyEnabled}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/confirm", `{"code":"123456"}`)
		if rec.Code != http.StatusConflict {
			t.Errorf("status: want 409, got %d", rec.Code)
		}
	})
}

// --- Disable -----------------------------------------------------------------

func TestMFAHandler_Disable(t *testing.T) {
	t.Run("200 on success forwards code", func(t *testing.T) {
		svc := &stubMFAService{}
		r := newRouter(svc, withAuthedAdmin(7))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/disable", `{"code":"654321"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: want 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastDisableCode != "654321" {
			t.Errorf("code: want 654321, got %q", svc.lastDisableCode)
		}
		if svc.lastDisableUserID != 7 {
			t.Errorf("user_id: want 7, got %d", svc.lastDisableUserID)
		}
	})

	t.Run("400 on malformed code", func(t *testing.T) {
		svc := &stubMFAService{}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/disable", `{"code":""}`)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status: want 400, got %d", rec.Code)
		}
	})

	t.Run("409 on ErrMFANotEnabled", func(t *testing.T) {
		svc := &stubMFAService{disableErr: entities.ErrMFANotEnabled}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/disable", `{"code":"123456"}`)
		if rec.Code != http.StatusConflict {
			t.Errorf("status: want 409, got %d", rec.Code)
		}
	})

	t.Run("422 on invalid code", func(t *testing.T) {
		svc := &stubMFAService{disableErr: entities.ErrInvalidMFACode}
		r := newRouter(svc, withAuthedAdmin(1))
		rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/disable", `{"code":"123456"}`)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("status: want 422, got %d", rec.Code)
		}
	})
}

// --- Constructor guard -------------------------------------------------------

func TestNewMFAHandler_NilSvcPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on nil svc")
		}
	}()
	_ = authhttp.NewMFAHandler(nil)
}

// Sanity check that response body is JSON (not plain text) so client UIs can
// parse error envelopes consistently across endpoints.
func TestMFAHandler_ResponsesAreJSON(t *testing.T) {
	svc := &stubMFAService{beginErr: entities.ErrMFAAlreadyEnabled}
	r := newRouter(svc, withAuthedAdmin(1))
	rec := performJSON(t, r, http.MethodPost, "/api/auth/mfa/begin", "")
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("expected JSON content-type, got %q", rec.Header().Get("Content-Type"))
	}
}
