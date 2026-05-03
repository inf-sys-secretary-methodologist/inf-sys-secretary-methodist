package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

// --- local mocks (kept here so the RED is self-contained — same
// convention as logout_usecase_test.go) ---

type mockPasswordResetTokenRepo struct {
	mock.Mock
}

func (m *mockPasswordResetTokenRepo) Store(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	args := m.Called(ctx, token, userID, ttl)
	return args.Error(0)
}

func (m *mockPasswordResetTokenRepo) LookupUser(ctx context.Context, token string) (int64, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPasswordResetTokenRepo) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// mockEmailSender is a tiny stand-in for the notifications EmailService —
// PasswordResetUseCase only needs SendPasswordResetEmail, so the
// dependency stays narrow (Interface Segregation).
type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error {
	args := m.Called(ctx, recipientEmail, resetToken)
	return args.Error(0)
}

// mockUserLookupRepo provides only the methods the password-reset flow
// needs. Reusing the full UserRepository mock would drag in unrelated
// methods.
type mockUserLookupRepo struct {
	mock.Mock
}

func (m *mockUserLookupRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*entities.User)
	return user, args.Error(1)
}

func (m *mockUserLookupRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*entities.User)
	return user, args.Error(1)
}

func (m *mockUserLookupRepo) Save(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// mkUser is a fixture helper. It calls the domain factory
// ReconstituteUser instead of constructing &entities.User{...} directly
// — the project's red-flag rule forbids raw composite literals of
// domain types outside the domain package.
func mkUser(id int64, email string, status entities.UserStatus) *entities.User {
	now := time.Now()
	return entities.ReconstituteUser(
		id,
		email,
		"old-bcrypt-hash",
		"Test User",
		domain.RoleStudent,
		status,
		now.Add(-time.Hour),
		now.Add(-time.Hour),
	)
}

// =================== RequestReset ===================

// TestPasswordResetUseCase_RequestReset is table-driven over the three
// observable cases the request step must guarantee:
//   - happy path: token stored AND emailed, both with the same token
//     value (otherwise the user can never confirm);
//   - unknown email: nil error AND no I/O (anti-enumeration — the
//     caller cannot tell from the response whether the user exists);
//   - blocked user: same shape as unknown — a blocked account cannot
//     reclaim itself via reset.
func TestPasswordResetUseCase_RequestReset(t *testing.T) {
	cases := []struct {
		name string
		// setup populates mocks and returns the user GetByEmail should
		// resolve to (or nil + error to model "not found").
		setup func(t *testing.T,
			userRepo *mockUserLookupRepo,
			tokenRepo *mockPasswordResetTokenRepo,
			emailer *mockEmailSender)
		email      string
		wantErr    bool
		wantStore  bool
		wantEmail  bool
		extraCheck func(t *testing.T,
			tokenRepo *mockPasswordResetTokenRepo,
			emailer *mockEmailSender)
	}{
		{
			name:  "known active email -> token stored AND emailed (same token)",
			email: "alice@example.com",
			setup: func(t *testing.T, userRepo *mockUserLookupRepo, tokenRepo *mockPasswordResetTokenRepo, emailer *mockEmailSender) {
				user := mkUser(42, "alice@example.com", entities.UserStatusActive)
				userRepo.On("GetByEmail", mock.Anything, "alice@example.com").Return(user, nil)
				tokenRepo.On("Store", mock.Anything,
					mock.AnythingOfType("string"),
					int64(42),
					mock.MatchedBy(func(d time.Duration) bool {
						diff := d - time.Hour
						if diff < 0 {
							diff = -diff
						}
						return diff <= 2*time.Second
					}),
				).Return(nil)
				emailer.On("SendPasswordResetEmail", mock.Anything,
					"alice@example.com",
					mock.AnythingOfType("string")).Return(nil)
			},
			wantStore: true,
			wantEmail: true,
			extraCheck: func(t *testing.T, tokenRepo *mockPasswordResetTokenRepo, emailer *mockEmailSender) {
				storeToken := tokenRepo.Calls[0].Arguments.String(1)
				emailToken := emailer.Calls[0].Arguments.String(2)
				assert.Equal(t, storeToken, emailToken,
					"token sent in email must match token stored in repo")
				assert.GreaterOrEqual(t, len(storeToken), 32,
					"reset token must carry ≥256 bits of entropy")
			},
		},
		{
			name:  "unknown email -> 204-shape success, no I/O (anti-enumeration)",
			email: "ghost@example.com",
			setup: func(t *testing.T, userRepo *mockUserLookupRepo, _ *mockPasswordResetTokenRepo, _ *mockEmailSender) {
				userRepo.On("GetByEmail", mock.Anything, "ghost@example.com").
					Return((*entities.User)(nil), errors.New("user not found"))
			},
		},
		{
			name:  "blocked user -> same anti-enumeration shape, no token issued",
			email: "blocked@example.com",
			setup: func(t *testing.T, userRepo *mockUserLookupRepo, _ *mockPasswordResetTokenRepo, _ *mockEmailSender) {
				user := mkUser(7, "blocked@example.com", entities.UserStatusBlocked)
				userRepo.On("GetByEmail", mock.Anything, "blocked@example.com").Return(user, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(mockUserLookupRepo)
			tokenRepo := new(mockPasswordResetTokenRepo)
			emailer := new(mockEmailSender)
			tc.setup(t, userRepo, tokenRepo, emailer)

			uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
			err := uc.RequestReset(context.Background(), tc.email)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if !tc.wantStore {
				tokenRepo.AssertNotCalled(t, "Store",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
			if !tc.wantEmail {
				emailer.AssertNotCalled(t, "SendPasswordResetEmail",
					mock.Anything, mock.Anything, mock.Anything)
			}
			if tc.extraCheck != nil {
				tc.extraCheck(t, tokenRepo, emailer)
			}
		})
	}
}

// =================== ConfirmReset ===================

// TestPasswordResetUseCase_ConfirmReset is table-driven over the four
// observable cases of the confirm step:
//   - happy path: stored hash actually validates against the new
//     plaintext, token deleted to enforce single use;
//   - invalid/expired token: ErrInvalidResetToken (errors.Is reachable
//     so handlers can map to a stable HTTP code), no Save / Delete;
//   - weak password: ErrWeakResetPassword BEFORE any I/O so a leaked
//     token cannot be spent on a 1-char password;
//   - user vanished: same opaque ErrInvalidResetToken (do not leak
//     "user existed but is gone"), no Save / Delete.
// savedSlot captures the *entities.User passed to Save so the table
// case can inspect the rotated hash post-call without reaching into
// testify mock internals.
type savedSlot struct{ user *entities.User }

func TestPasswordResetUseCase_ConfirmReset(t *testing.T) {
	const goodPassword = "NewStrongPass1!"

	cases := []struct {
		name        string
		token       string
		password    string
		setup       func(userRepo *mockUserLookupRepo, tokenRepo *mockPasswordResetTokenRepo, slot *savedSlot)
		wantErrIs   error
		wantSave    bool
		wantDelete  bool
		assertSaved func(t *testing.T, saved *entities.User)
	}{
		{
			name:     "valid token + strong password -> Save + Delete, hash validates",
			token:    "valid-token",
			password: goodPassword,
			setup: func(userRepo *mockUserLookupRepo, tokenRepo *mockPasswordResetTokenRepo, slot *savedSlot) {
				user := mkUser(42, "alice@example.com", entities.UserStatusActive)
				tokenRepo.On("LookupUser", mock.Anything, "valid-token").Return(int64(42), nil)
				userRepo.On("GetByID", mock.Anything, int64(42)).Return(user, nil)
				userRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.User")).
					Run(func(args mock.Arguments) {
						slot.user = args.Get(1).(*entities.User)
					}).Return(nil)
				tokenRepo.On("Delete", mock.Anything, "valid-token").Return(nil)
			},
			wantSave:   true,
			wantDelete: true,
			assertSaved: func(t *testing.T, saved *entities.User) {
				require.NotNil(t, saved)
				assert.NotEqual(t, "old-bcrypt-hash", saved.Password)
				assert.NoError(t,
					bcrypt.CompareHashAndPassword([]byte(saved.Password), []byte(goodPassword)),
					"stored hash must match the new plaintext")
			},
		},
		{
			name:     "invalid/expired token -> ErrInvalidResetToken",
			token:    "bad-token",
			password: goodPassword,
			setup: func(_ *mockUserLookupRepo, tokenRepo *mockPasswordResetTokenRepo, _ *savedSlot) {
				tokenRepo.On("LookupUser", mock.Anything, "bad-token").
					Return(int64(0), repositories.ErrPasswordResetTokenNotFound)
			},
			wantErrIs: ErrInvalidResetToken,
		},
		{
			name:      "weak password -> ErrWeakResetPassword BEFORE any I/O",
			token:     "any-token",
			password:  "short",
			setup:     func(_ *mockUserLookupRepo, _ *mockPasswordResetTokenRepo, _ *savedSlot) {},
			wantErrIs: ErrWeakResetPassword,
		},
		{
			name:     "user vanished -> opaque ErrInvalidResetToken, no Save / Delete",
			token:    "ghost-token",
			password: goodPassword,
			setup: func(userRepo *mockUserLookupRepo, tokenRepo *mockPasswordResetTokenRepo, _ *savedSlot) {
				tokenRepo.On("LookupUser", mock.Anything, "ghost-token").Return(int64(99), nil)
				userRepo.On("GetByID", mock.Anything, int64(99)).
					Return((*entities.User)(nil), errors.New("user not found"))
			},
			wantErrIs: ErrInvalidResetToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := new(mockUserLookupRepo)
			tokenRepo := new(mockPasswordResetTokenRepo)
			emailer := new(mockEmailSender)
			slot := &savedSlot{}
			tc.setup(userRepo, tokenRepo, slot)

			uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
			err := uc.ConfirmReset(context.Background(), tc.token, tc.password)

			if tc.wantErrIs != nil {
				assert.ErrorIs(t, err, tc.wantErrIs)
			} else {
				assert.NoError(t, err)
			}
			if !tc.wantSave {
				userRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
			}
			if !tc.wantDelete {
				tokenRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
			}
			if tc.assertSaved != nil {
				tc.assertSaved(t, slot.user)
			}
		})
	}
}

// =================== VerifyToken ===================

// TestPasswordResetUseCase_VerifyToken pins the read-only validity
// probe used by the frontend before rendering the new-password form.
// Only two cases (the gate "≥3 = table-driven" does not apply);
// importantly the valid case also asserts the token is NOT consumed —
// otherwise the user could not confirm right after verifying.
func TestPasswordResetUseCase_VerifyToken_Valid(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	tokenRepo := new(mockPasswordResetTokenRepo)
	tokenRepo.On("LookupUser", mock.Anything, "good-token").Return(int64(42), nil)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.VerifyToken(context.Background(), "good-token")

	assert.NoError(t, err)
	tokenRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// TestPasswordResetUseCase_VerifyToken_Invalid — unknown / expired
// token surfaces ErrInvalidResetToken (errors.Is reachable). Same
// shape as ConfirmReset so the frontend can use one error mapping.
func TestPasswordResetUseCase_VerifyToken_Invalid(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	tokenRepo := new(mockPasswordResetTokenRepo)
	tokenRepo.On("LookupUser", mock.Anything, "bad-token").
		Return(int64(0), repositories.ErrPasswordResetTokenNotFound)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.VerifyToken(context.Background(), "bad-token")

	assert.ErrorIs(t, err, ErrInvalidResetToken)
}
