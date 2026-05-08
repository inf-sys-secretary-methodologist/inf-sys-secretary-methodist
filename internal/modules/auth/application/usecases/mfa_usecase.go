package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// MFAUseCase orchestrates TOTP enrollment, confirmation, and disable flows.
type MFAUseCase struct {
	userRepo    repositories.UserRepository
	auditLogger *logging.AuditLogger
	issuer      string
	now         func() time.Time
}

// NewMFAUseCase builds the use case. issuer is embedded in the otpauth URI
// so the user's authenticator app shows a recognizable label.
func NewMFAUseCase(userRepo repositories.UserRepository, auditLogger *logging.AuditLogger, issuer string) *MFAUseCase {
	return NewMFAUseCaseWithClock(userRepo, auditLogger, issuer, time.Now)
}

// NewMFAUseCaseWithClock is the same as NewMFAUseCase but accepts an
// injectable clock so tests can pin TOTP verification to a deterministic
// timestamp.
func NewMFAUseCaseWithClock(userRepo repositories.UserRepository, auditLogger *logging.AuditLogger, issuer string, now func() time.Time) *MFAUseCase {
	if userRepo == nil {
		panic("mfa usecase: userRepo is nil")
	}
	if now == nil {
		now = time.Now
	}
	return &MFAUseCase{
		userRepo:    userRepo,
		auditLogger: auditLogger,
		issuer:      issuer,
		now:         now,
	}
}

// BeginEnrollment generates a fresh TOTP secret, persists it on the user
// (mfa_enabled stays false until confirmation), and returns the otpauth URI
// the frontend renders as a QR code plus the raw Base32 secret for manual
// entry. Discards any previously pending secret.
func (uc *MFAUseCase) BeginEnrollment(_ context.Context, _ int64) (string, string, error) {
	return "", "", nil // RED stub
}

// ConfirmEnrollment verifies a user-provided 6-digit code against the pending
// secret. On success flips mfa_enabled to true.
func (uc *MFAUseCase) ConfirmEnrollment(_ context.Context, _ int64, _ string) error {
	return nil // RED stub
}

// Disable verifies a user-provided 6-digit code against the active secret and,
// on success, clears MFA state.
func (uc *MFAUseCase) Disable(_ context.Context, _ int64, _ string) error {
	return nil // RED stub
}
