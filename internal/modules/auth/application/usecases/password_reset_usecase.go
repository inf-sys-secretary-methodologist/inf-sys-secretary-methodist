package usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

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

// generateResetToken returns a cryptographically random URL-safe token.
func generateResetToken() (string, error) {
	b := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
