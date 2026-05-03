package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

// --- local mocks (kept here so the handler test is self-contained;
// same convention as logout_handler_test.go) ---

type fakeResetTokenRepo struct {
	mock.Mock
}

func (f *fakeResetTokenRepo) Store(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	return f.Called(ctx, token, userID, ttl).Error(0)
}

func (f *fakeResetTokenRepo) LookupUser(ctx context.Context, token string) (int64, error) {
	args := f.Called(ctx, token)
	return args.Get(0).(int64), args.Error(1)
}

func (f *fakeResetTokenRepo) Delete(ctx context.Context, token string) error {
	return f.Called(ctx, token).Error(0)
}

type fakeUserRepoForResetHandler struct {
	mock.Mock
}

func (f *fakeUserRepoForResetHandler) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := f.Called(ctx, email)
	user, _ := args.Get(0).(*entities.User)
	return user, args.Error(1)
}

func (f *fakeUserRepoForResetHandler) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	args := f.Called(ctx, id)
	user, _ := args.Get(0).(*entities.User)
	return user, args.Error(1)
}

func (f *fakeUserRepoForResetHandler) Save(ctx context.Context, user *entities.User) error {
	return f.Called(ctx, user).Error(0)
}

type fakeEmailerForResetHandler struct {
	mock.Mock
}

func (f *fakeEmailerForResetHandler) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	return f.Called(ctx, email, token).Error(0)
}

// buildResetHandler wires a real PasswordResetUseCase with the fake
// dependencies — handler tests assert HTTP-layer behavior end-to-end
// through the usecase, mirroring logout_handler_test.go.
func buildResetHandler(t *testing.T) (
	*PasswordResetHandler,
	*fakeUserRepoForResetHandler,
	*fakeResetTokenRepo,
	*fakeEmailerForResetHandler,
) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	userRepo := new(fakeUserRepoForResetHandler)
	tokenRepo := new(fakeResetTokenRepo)
	emailer := new(fakeEmailerForResetHandler)
	uc := usecases.NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	return NewPasswordResetHandler(uc), userRepo, tokenRepo, emailer
}

// =================== POST /password-reset/request ===================

// TestPasswordResetHandler_Request — table-driven over the four
// observable cases of the request endpoint. The contract: malformed
// input gets 400; any well-formed body returns 204 regardless of
// whether the email exists, so the response shape never leaks user
// existence.
func TestPasswordResetHandler_Request(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		setup          func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo, em *fakeEmailerForResetHandler)
		wantStatus     int
		wantEmailSent  bool
		wantTokenStore bool
	}{
		{
			name: "known email -> 204 + email + token stored",
			body: `{"email":"alice@example.com"}`,
			setup: func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo, em *fakeEmailerForResetHandler) {
				u.On("GetByEmail", mock.Anything, "alice@example.com").Return(&entities.User{
					ID: 1, Email: "alice@example.com", Status: entities.UserStatusActive, Role: domain.RoleStudent,
				}, nil)
				tr.On("Store", mock.Anything, mock.AnythingOfType("string"), int64(1), mock.AnythingOfType("time.Duration")).Return(nil)
				em.On("SendPasswordResetEmail", mock.Anything, "alice@example.com", mock.AnythingOfType("string")).Return(nil)
			},
			wantStatus:     http.StatusNoContent,
			wantEmailSent:  true,
			wantTokenStore: true,
		},
		{
			name: "unknown email -> 204 (anti-enumeration), no email, no token",
			body: `{"email":"ghost@example.com"}`,
			setup: func(u *fakeUserRepoForResetHandler, _ *fakeResetTokenRepo, _ *fakeEmailerForResetHandler) {
				u.On("GetByEmail", mock.Anything, "ghost@example.com").Return((*entities.User)(nil), assertNotFound())
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "empty email field -> 400",
			body:       `{"email":""}`,
			setup:      func(u *fakeUserRepoForResetHandler, _ *fakeResetTokenRepo, _ *fakeEmailerForResetHandler) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed JSON -> 400",
			body:       `{not-json`,
			setup:      func(u *fakeUserRepoForResetHandler, _ *fakeResetTokenRepo, _ *fakeEmailerForResetHandler) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, userRepo, tokenRepo, emailer := buildResetHandler(t)
			tc.setup(userRepo, tokenRepo, emailer)

			router := gin.New()
			router.POST("/api/auth/password-reset/request", h.RequestReset)

			req := httptest.NewRequest(http.MethodPost,
				"/api/auth/password-reset/request",
				strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			if !tc.wantEmailSent {
				emailer.AssertNotCalled(t, "SendPasswordResetEmail",
					mock.Anything, mock.Anything, mock.Anything)
			}
			if !tc.wantTokenStore {
				tokenRepo.AssertNotCalled(t, "Store",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

// =================== GET /password-reset/verify/:token ===================

// TestPasswordResetHandler_Verify — read-only token validity probe.
// Valid token -> 204, invalid -> 410 Gone. The frontend uses 410 to
// render "this link has expired" before showing the password form.
func TestPasswordResetHandler_Verify(t *testing.T) {
	cases := []struct {
		name       string
		token      string
		setup      func(tr *fakeResetTokenRepo)
		wantStatus int
	}{
		{
			name:  "valid stored token -> 204",
			token: "valid-tok",
			setup: func(tr *fakeResetTokenRepo) {
				tr.On("LookupUser", mock.Anything, "valid-tok").Return(int64(1), nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:  "expired/unknown token -> 410 Gone",
			token: "bad-tok",
			setup: func(tr *fakeResetTokenRepo) {
				tr.On("LookupUser", mock.Anything, "bad-tok").
					Return(int64(0), repositories.ErrPasswordResetTokenNotFound)
			},
			wantStatus: http.StatusGone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _, tokenRepo, _ := buildResetHandler(t)
			tc.setup(tokenRepo)

			router := gin.New()
			router.GET("/api/auth/password-reset/verify/:token", h.VerifyResetToken)

			req := httptest.NewRequest(http.MethodGet,
				"/api/auth/password-reset/verify/"+tc.token, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// =================== POST /password-reset/confirm ===================

// TestPasswordResetHandler_Confirm — the consume step. Distinct error
// codes so the frontend can render the right message: 400 for weak
// password (user can fix it), 410 for invalid token (link is dead and
// must be re-requested), 204 on success.
func TestPasswordResetHandler_Confirm(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		setup      func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo)
		wantStatus int
	}{
		{
			name: "valid token + strong password -> 204, password rotated",
			body: `{"token":"good-tok","password":"NewStrongPass1!"}`,
			setup: func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo) {
				tr.On("LookupUser", mock.Anything, "good-tok").Return(int64(1), nil)
				u.On("GetByID", mock.Anything, int64(1)).Return(&entities.User{
					ID: 1, Email: "alice@example.com", Password: "old", Status: entities.UserStatusActive,
				}, nil)
				u.On("Save", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				tr.On("Delete", mock.Anything, "good-tok").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "weak password -> 400 (no token consume)",
			body: `{"token":"good-tok","password":"short"}`,
			setup: func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo) {
				// No mock setup — usecase rejects before any I/O.
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid token -> 410 Gone",
			body: `{"token":"bad-tok","password":"NewStrongPass1!"}`,
			setup: func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo) {
				tr.On("LookupUser", mock.Anything, "bad-tok").
					Return(int64(0), repositories.ErrPasswordResetTokenNotFound)
			},
			wantStatus: http.StatusGone,
		},
		{
			name:       "missing token field -> 400",
			body:       `{"password":"NewStrongPass1!"}`,
			setup:      func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed JSON -> 400",
			body:       `{nope`,
			setup:      func(u *fakeUserRepoForResetHandler, tr *fakeResetTokenRepo) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, userRepo, tokenRepo, _ := buildResetHandler(t)
			tc.setup(userRepo, tokenRepo)

			router := gin.New()
			router.POST("/api/auth/password-reset/confirm", h.ConfirmReset)

			req := httptest.NewRequest(http.MethodPost,
				"/api/auth/password-reset/confirm",
				strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

// TestPasswordResetHandler_Confirm_RotatedPasswordIsBcrypt verifies
// the end-to-end happy path also lands a real bcrypt hash on the user
// (defensive — the usecase test already pins this, but this is the
// handler path under realistic JSON wiring).
func TestPasswordResetHandler_Confirm_RotatedPasswordIsBcrypt(t *testing.T) {
	h, userRepo, tokenRepo, _ := buildResetHandler(t)

	user := &entities.User{
		ID: 1, Email: "alice@example.com", Password: "old",
		Status: entities.UserStatusActive,
	}
	tokenRepo.On("LookupUser", mock.Anything, "tok").Return(int64(1), nil)
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)
	var saved *entities.User
	userRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.User")).
		Run(func(args mock.Arguments) { saved = args.Get(1).(*entities.User) }).
		Return(nil)
	tokenRepo.On("Delete", mock.Anything, "tok").Return(nil)

	router := gin.New()
	router.POST("/api/auth/password-reset/confirm", h.ConfirmReset)
	req := httptest.NewRequest(http.MethodPost,
		"/api/auth/password-reset/confirm",
		strings.NewReader(`{"token":"tok","password":"NewStrongPass1!"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.NotNil(t, saved)
	assert.NoError(t,
		bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte("NewStrongPass1!")))
}

// assertNotFound returns a sentinel-ish error to feed into mocks where
// only "not nil" matters; keeps test bodies compact.
func assertNotFound() error {
	return errStub("user not found")
}

type errStub string

func (e errStub) Error() string { return string(e) }
