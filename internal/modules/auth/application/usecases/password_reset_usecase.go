package usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

const (
	// passwordResetTokenTTL bounds how long a reset link is usable. Long
	// enough that a user can switch to email and back; short enough that
	// a leaked link expires quickly.
	passwordResetTokenTTL = time.Hour

	// passwordResetTokenBytes — 256 bits of entropy. Encoded with
	// base64.RawURLEncoding the resulting string is 43 chars, safe for
	// query parameters without further escaping.
	passwordResetTokenBytes = 32

	// passwordResetMinLength — backend floor; the frontend enforces a
	// stronger composition policy (upper/lower/digit/special). The
	// backend cannot trust frontend-only checks, so it pins the minimum
	// here to refuse trivially short passwords end-to-end.
	passwordResetMinLength = 8
)

var (
	// ErrInvalidResetToken is returned by ConfirmReset when the supplied
	// token is unknown / expired or its referenced user has vanished.
	// The two cases are deliberately collapsed to one error so callers
	// cannot distinguish them and probe the user table.
	ErrInvalidResetToken = errors.New("invalid or expired password reset token")

	// ErrWeakResetPassword is returned by ConfirmReset when the new
	// password fails the backend minimum-length check. Distinct from
	// ErrInvalidResetToken so handlers can produce different HTTP
	// responses (validation vs auth).
	ErrWeakResetPassword = errors.New("password does not meet minimum strength requirements")
)

// EmailSender is the narrow outbound contract PasswordResetUseCase
// needs. Defined here (Interface Segregation) so the test can stub one
// method instead of the whole notifications EmailService.
type EmailSender interface {
	SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error
}

// userLookup is the slice of UserRepository the password-reset flow
// uses. Defined where consumed (DIP); the production user repository
// satisfies it implicitly via Go's structural typing.
type userLookup interface {
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	Save(ctx context.Context, user *entities.User) error
}

// PasswordResetUseCase orchestrates the password recovery flow:
// request a reset (issue token + email), then confirm with the token to
// set a new password.
type PasswordResetUseCase struct {
	userRepo  userLookup
	tokenRepo repositories.PasswordResetTokenRepository
	emailer   EmailSender
}

// NewPasswordResetUseCase wires the dependencies.
func NewPasswordResetUseCase(
	userRepo userLookup,
	tokenRepo repositories.PasswordResetTokenRepository,
	emailer EmailSender,
) *PasswordResetUseCase {
	return &PasswordResetUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		emailer:   emailer,
	}
}

// RequestReset issues a password reset link for the user with the given
// email.
//
// Anti-enumeration: returns nil for an unknown email or a non-active
// user, so a caller cannot tell from the response whether the email
// exists in the system. Errors are only returned for genuine system
// faults (token storage, email delivery).
func (u *PasswordResetUseCase) RequestReset(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil
	}
	if user.Status != entities.UserStatusActive {
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}
	if err := u.tokenRepo.Store(ctx, token, user.ID, passwordResetTokenTTL); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}
	if err := u.emailer.SendPasswordResetEmail(ctx, user.Email, token); err != nil {
		return fmt.Errorf("send reset email: %w", err)
	}
	return nil
}

// VerifyToken reports whether token is currently a valid reset token
// without consuming it. Used by the frontend to decide whether to
// render the new-password form. Returns ErrInvalidResetToken for
// unknown / expired tokens; storage faults bubble up wrapped.
func (u *PasswordResetUseCase) VerifyToken(ctx context.Context, token string) error {
	if _, err := u.tokenRepo.LookupUser(ctx, token); err != nil {
		if errors.Is(err, repositories.ErrPasswordResetTokenNotFound) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("lookup reset token: %w", err)
	}
	return nil
}

// ConfirmReset consumes a previously issued reset token and replaces
// the target user's password.
//
// Order matters: the password is validated first, before any I/O, so a
// leaked token cannot be used to set a trivially weak password (the
// invalid attempt is rejected without burning the token). On success
// the token is deleted to enforce single use; failure to delete is
// treated as fatal because a re-usable token would defeat the bound.
//
// ErrInvalidResetToken collapses "token unknown/expired" and "user
// vanished" into one shape so callers cannot use the response to probe
// the user table.
func (u *PasswordResetUseCase) ConfirmReset(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < passwordResetMinLength {
		return ErrWeakResetPassword
	}

	userID, err := u.tokenRepo.LookupUser(ctx, token)
	if err != nil {
		if errors.Is(err, repositories.ErrPasswordResetTokenNotFound) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("lookup reset token: %w", err)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ErrInvalidResetToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	user.UpdatePassword(string(hashed))

	if err := u.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	if err := u.tokenRepo.Delete(ctx, token); err != nil {
		return fmt.Errorf("delete reset token: %w", err)
	}
	return nil
}

// generateResetToken returns a cryptographically random URL-safe token.
func generateResetToken() (string, error) {
	b := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
