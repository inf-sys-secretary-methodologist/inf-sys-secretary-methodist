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

// mockPasswordResetTokenRepo is a hand-rolled mock for the
// PasswordResetTokenRepository interface. Defined locally to keep the RED
// test self-contained — same convention as logout_usecase_test.go.
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
// PasswordResetUseCase only needs SendPasswordResetEmail, so we constrain
// the dependency to that single method via an interface defined on the
// usecase side (Interface Segregation Principle).
type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error {
	args := m.Called(ctx, recipientEmail, resetToken)
	return args.Error(0)
}

// mockUserLookupRepo provides only the GetByEmail method needed for the
// RequestReset flow. Reusing the full UserRepository mock would drag in
// many unrelated methods.
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

// TestPasswordResetUseCase_RequestReset_KnownEmail_StoresTokenAndSendsEmail
// verifies the happy path: a known-active user gets a fresh secret token
// stored in the repository (TTL ≈ 1 hour) and an email containing that
// exact token sent through the email service.
func TestPasswordResetUseCase_RequestReset_KnownEmail_StoresTokenAndSendsEmail(t *testing.T) {
	user := &entities.User{
		ID:     42,
		Email:  "alice@example.com",
		Name:   "Alice",
		Status: entities.UserStatusActive,
		Role:   domain.RoleStudent,
	}

	userRepo := new(mockUserLookupRepo)
	userRepo.On("GetByEmail", mock.Anything, "alice@example.com").Return(user, nil)

	tokenRepo := new(mockPasswordResetTokenRepo)
	// Capture the token the usecase generates so we can assert it gets
	// passed verbatim to the email sender — a leak would mean the user
	// receives a different token than what is stored, breaking confirm.
	var storedToken string
	tokenRepo.On("Store", mock.Anything, mock.AnythingOfType("string"), int64(42),
		mock.MatchedBy(func(d time.Duration) bool {
			diff := d - time.Hour
			if diff < 0 {
				diff = -diff
			}
			return diff <= 2*time.Second
		}),
	).Run(func(args mock.Arguments) {
		storedToken = args.String(1)
	}).Return(nil)

	emailer := new(mockEmailSender)
	emailer.On("SendPasswordResetEmail", mock.Anything, "alice@example.com",
		mock.AnythingOfType("string"),
	).Return(nil)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.RequestReset(context.Background(), "alice@example.com")

	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
	emailer.AssertExpectations(t)

	// Same token in storage and email — otherwise the user can never
	// confirm; this is the invariant that ties the two calls together.
	emailCallToken := emailer.Calls[0].Arguments.String(2)
	assert.Equal(t, storedToken, emailCallToken,
		"token sent in email must match token stored in repo")
	assert.NotEmpty(t, storedToken, "generated token must be non-empty")
	assert.GreaterOrEqual(t, len(storedToken), 32,
		"reset token must be unguessable: at least 32 chars of entropy")
}

// TestPasswordResetUseCase_RequestReset_UnknownEmail_NoEnumeration verifies
// the anti-enumeration property: requesting a reset for an email that does
// not exist must return success and must NOT touch the token repo or the
// email service. Otherwise an attacker can enumerate valid users by
// observing differences in response timing / error / 2xx vs 4xx.
func TestPasswordResetUseCase_RequestReset_UnknownEmail_NoEnumeration(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	userRepo.On("GetByEmail", mock.Anything, "ghost@example.com").
		Return((*entities.User)(nil), errors.New("user not found"))

	tokenRepo := new(mockPasswordResetTokenRepo)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.RequestReset(context.Background(), "ghost@example.com")

	assert.NoError(t, err, "unknown email must not leak existence via error")
	tokenRepo.AssertNotCalled(t, "Store", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	emailer.AssertNotCalled(t, "SendPasswordResetEmail", mock.Anything, mock.Anything, mock.Anything)
}

// TestPasswordResetUseCase_RequestReset_BlockedUser_NoTokenIssued — a
// blocked account must not be able to recover access via password reset.
// Issuing a token would let a removed/blocked user reclaim the account.
func TestPasswordResetUseCase_RequestReset_BlockedUser_NoTokenIssued(t *testing.T) {
	user := &entities.User{
		ID:     7,
		Email:  "blocked@example.com",
		Status: entities.UserStatusBlocked,
		Role:   domain.RoleStudent,
	}

	userRepo := new(mockUserLookupRepo)
	userRepo.On("GetByEmail", mock.Anything, "blocked@example.com").Return(user, nil)

	tokenRepo := new(mockPasswordResetTokenRepo)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.RequestReset(context.Background(), "blocked@example.com")

	// Same anti-enumeration shape: caller can't tell "blocked" from "unknown".
	assert.NoError(t, err)
	tokenRepo.AssertNotCalled(t, "Store", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	emailer.AssertNotCalled(t, "SendPasswordResetEmail", mock.Anything, mock.Anything, mock.Anything)
}

// TestPasswordResetUseCase_ConfirmReset_ValidToken_UpdatesPasswordAndDeletesToken
// is the happy path: a stored token resolves to a user, the new
// password is hashed (bcrypt comparable), the user is saved with the
// new hash, and the token is deleted to enforce single use.
func TestPasswordResetUseCase_ConfirmReset_ValidToken_UpdatesPasswordAndDeletesToken(t *testing.T) {
	user := &entities.User{
		ID:       42,
		Email:    "alice@example.com",
		Password: "old-bcrypt-hash",
		Status:   entities.UserStatusActive,
		Role:     domain.RoleStudent,
	}

	userRepo := new(mockUserLookupRepo)
	userRepo.On("GetByID", mock.Anything, int64(42)).Return(user, nil)
	var savedUser *entities.User
	userRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.User")).
		Run(func(args mock.Arguments) {
			savedUser = args.Get(1).(*entities.User)
		}).Return(nil)

	tokenRepo := new(mockPasswordResetTokenRepo)
	tokenRepo.On("LookupUser", mock.Anything, "valid-token").Return(int64(42), nil)
	tokenRepo.On("Delete", mock.Anything, "valid-token").Return(nil)

	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.ConfirmReset(context.Background(), "valid-token", "NewStrongPass1!")

	require.NoError(t, err)
	require.NotNil(t, savedUser, "Save must be called with the user")
	assert.NotEqual(t, "old-bcrypt-hash", savedUser.Password, "stored password hash must change")
	// The hash must validate against the plaintext we passed in — otherwise
	// the user cannot log in with the new password.
	assert.NoError(t,
		bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte("NewStrongPass1!")),
		"stored hash must match the new plaintext password")
	tokenRepo.AssertExpectations(t)
}

// TestPasswordResetUseCase_ConfirmReset_InvalidToken_ReturnsErrInvalidResetToken
// — a token absent from storage (expired or never issued) yields a
// distinguishable error and triggers no Save / Delete. errors.Is must
// find ErrInvalidResetToken so handlers can map it to a stable HTTP code.
func TestPasswordResetUseCase_ConfirmReset_InvalidToken_ReturnsErrInvalidResetToken(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	tokenRepo := new(mockPasswordResetTokenRepo)
	tokenRepo.On("LookupUser", mock.Anything, "bad-token").
		Return(int64(0), repositories.ErrPasswordResetTokenNotFound)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.ConfirmReset(context.Background(), "bad-token", "NewStrongPass1!")

	assert.ErrorIs(t, err, ErrInvalidResetToken)
	userRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	tokenRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// TestPasswordResetUseCase_ConfirmReset_WeakPassword_RejectsBeforeLookup
// — password must clear a minimum length bar before we even consult the
// token store. Otherwise a leaked token could be used to set an empty
// or 1-char password and lock the account into a trivially guessable
// state. Order matters: validation before any I/O.
func TestPasswordResetUseCase_ConfirmReset_WeakPassword_RejectsBeforeLookup(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	tokenRepo := new(mockPasswordResetTokenRepo)
	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.ConfirmReset(context.Background(), "any-token", "short")

	assert.ErrorIs(t, err, ErrWeakResetPassword)
	tokenRepo.AssertNotCalled(t, "LookupUser", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// TestPasswordResetUseCase_ConfirmReset_UserVanished_ReturnsErrInvalidResetToken
// — token was valid at issue time, but the user has since been deleted.
// Surface the same ErrInvalidResetToken to the caller (do not leak the
// distinction); ensure no stale Save / Delete occurs.
func TestPasswordResetUseCase_ConfirmReset_UserVanished_ReturnsErrInvalidResetToken(t *testing.T) {
	userRepo := new(mockUserLookupRepo)
	userRepo.On("GetByID", mock.Anything, int64(99)).
		Return((*entities.User)(nil), errors.New("user not found"))

	tokenRepo := new(mockPasswordResetTokenRepo)
	tokenRepo.On("LookupUser", mock.Anything, "ghost-token").Return(int64(99), nil)

	emailer := new(mockEmailSender)

	uc := NewPasswordResetUseCase(userRepo, tokenRepo, emailer)
	err := uc.ConfirmReset(context.Background(), "ghost-token", "NewStrongPass1!")

	assert.ErrorIs(t, err, ErrInvalidResetToken)
	userRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	tokenRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
