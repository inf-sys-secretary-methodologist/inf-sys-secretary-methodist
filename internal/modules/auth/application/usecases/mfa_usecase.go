package usecases

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/security/totp"
)

// totpDriftWindow is the number of 30-second steps tolerated either side of
// the server clock. ±1 step (≈30s) is enough for typical phone clock drift
// without widening the brute-force surface meaningfully.
const totpDriftWindow = 1

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
func (uc *MFAUseCase) BeginEnrollment(ctx context.Context, userID int64) (string, string, error) {
	user, err := uc.userRepo.GetByIDForAuth(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("mfa: load user: %w", err)
	}
	if user.MFAEnabled {
		return "", "", entities.ErrMFAAlreadyEnabled
	}

	_, encoded, err := totp.GenerateSecret()
	if err != nil {
		return "", "", fmt.Errorf("mfa: generate secret: %w", err)
	}
	secret, err := entities.NewMFASecret(encoded)
	if err != nil {
		return "", "", fmt.Errorf("mfa: validate generated secret: %w", err)
	}

	user.MFASecret = &secret
	user.MFAEnabled = false
	user.UpdatedAt = uc.now()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return "", "", fmt.Errorf("mfa: save pending secret: %w", err)
	}

	uc.logAudit(ctx, "mfa_enrollment_begin", user.ID)
	return buildOTPAuthURI(uc.issuer, user.Email, encoded), encoded, nil
}

// ConfirmEnrollment verifies a user-provided 6-digit code against the pending
// secret. On success flips mfa_enabled to true.
func (uc *MFAUseCase) ConfirmEnrollment(ctx context.Context, userID int64, code string) error {
	user, err := uc.userRepo.GetByIDForAuth(ctx, userID)
	if err != nil {
		return fmt.Errorf("mfa: load user: %w", err)
	}
	if user.MFAEnabled {
		return entities.ErrMFAAlreadyEnabled
	}
	if user.MFASecret == nil {
		return entities.ErrMFANotPending
	}

	if err := uc.verifyCode(*user.MFASecret, code); err != nil {
		return err
	}

	if err := user.EnableMFA(*user.MFASecret); err != nil {
		return fmt.Errorf("mfa: enable: %w", err)
	}
	user.UpdatedAt = uc.now()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("mfa: persist enabled state: %w", err)
	}

	uc.logAudit(ctx, "mfa_enrollment_confirm", user.ID)
	return nil
}

// Disable verifies a user-provided 6-digit code against the active secret and,
// on success, clears MFA state.
func (uc *MFAUseCase) Disable(ctx context.Context, userID int64, code string) error {
	user, err := uc.userRepo.GetByIDForAuth(ctx, userID)
	if err != nil {
		return fmt.Errorf("mfa: load user: %w", err)
	}
	if !user.MFAEnabled || user.MFASecret == nil {
		return entities.ErrMFANotEnabled
	}

	if err := uc.verifyCode(*user.MFASecret, code); err != nil {
		return err
	}

	if err := user.DisableMFA(); err != nil {
		return fmt.Errorf("mfa: disable: %w", err)
	}
	user.UpdatedAt = uc.now()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("mfa: persist disabled state: %w", err)
	}

	uc.logAudit(ctx, "mfa_disabled", user.ID)
	return nil
}

func (uc *MFAUseCase) verifyCode(secret entities.MFASecret, code string) error {
	raw, err := secret.Decode()
	if err != nil {
		return fmt.Errorf("mfa: decode secret: %w", err)
	}
	if !totp.Verify(raw, code, uc.now(), totpDriftWindow) {
		return entities.ErrInvalidMFACode
	}
	return nil
}

func (uc *MFAUseCase) logAudit(ctx context.Context, action string, userID int64) {
	if uc.auditLogger == nil {
		return
	}
	uc.auditLogger.LogAuditEvent(ctx, action, "auth", map[string]any{
		"user_id": userID,
	})
}

// buildOTPAuthURI returns the standard otpauth:// URI consumed by Google
// Authenticator, Authy, 1Password, etc. Fixed parameters: SHA1, 6 digits,
// 30-second period (RFC 4226 + RFC 6238 defaults).
func buildOTPAuthURI(issuer, email, secret string) string {
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", "30")
	return fmt.Sprintf("otpauth://totp/%s:%s?%s", issuer, url.QueryEscape(email), q.Encode())
}
