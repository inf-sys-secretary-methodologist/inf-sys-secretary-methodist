package usecases_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/security/totp"
)

// stubUserRepo captures the last Save call so assertions can inspect what the
// use case persisted, while letting tests script GetByIDForAuth's response.
type stubUserRepo struct {
	mu             sync.Mutex
	getByIDForAuth func(ctx context.Context, id int64) (*entities.User, error)
	lastSaved      *entities.User
	saveErr        error
}

func (s *stubUserRepo) Create(_ context.Context, _ *entities.User) error { return nil }
func (s *stubUserRepo) Save(_ context.Context, u *entities.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSaved = u
	return s.saveErr
}
func (s *stubUserRepo) GetByID(_ context.Context, _ int64) (*entities.User, error) { return nil, nil }
func (s *stubUserRepo) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetByEmailForAuth(_ context.Context, _ string) (*entities.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetByIDForAuth(ctx context.Context, id int64) (*entities.User, error) {
	if s.getByIDForAuth != nil {
		return s.getByIDForAuth(ctx, id)
	}
	return nil, errors.New("getByIDForAuth not stubbed")
}
func (s *stubUserRepo) Delete(_ context.Context, _ int64) error                    { return nil }
func (s *stubUserRepo) List(_ context.Context, _, _ int) ([]*entities.User, error) { return nil, nil }

const (
	testIssuer = "inf-sys-test"
	testEmail  = "admin@example.local"
)

// frozenTime returns a fixed timestamp so TOTP codes are deterministic
// across the BeginEnrollment → ConfirmEnrollment / Disable sequence.
var frozenTime = time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)

// freshAdminUser builds an admin user with no MFA configured.
func freshAdminUser() *entities.User {
	return &entities.User{
		ID:     42,
		Email:  testEmail,
		Role:   domain.RoleSystemAdmin,
		Status: entities.UserStatusActive,
	}
}

// adminWithPendingSecret builds an admin in the "Begin done, Confirm pending"
// state — secret stored, mfa_enabled still false.
func adminWithPendingSecret(t *testing.T, encoded string) *entities.User {
	t.Helper()
	secret, err := entities.NewMFASecret(encoded)
	if err != nil {
		t.Fatalf("setup: NewMFASecret: %v", err)
	}
	u := freshAdminUser()
	u.MFASecret = &secret
	u.MFAEnabled = false
	return u
}

// adminEnrolled builds a fully enrolled admin (Begin + Confirm done).
func adminEnrolled(t *testing.T, encoded string) *entities.User {
	t.Helper()
	u := adminWithPendingSecret(t, encoded)
	u.MFAEnabled = true
	return u
}

// codeAt computes the TOTP code for the given Base32 secret at frozenTime so
// tests can hand it to ConfirmEnrollment / Disable without hardcoding values.
func codeAt(t *testing.T, encoded string, at time.Time) string {
	t.Helper()
	secret, err := entities.NewMFASecret(encoded)
	if err != nil {
		t.Fatalf("setup: NewMFASecret: %v", err)
	}
	raw, err := secret.Decode()
	if err != nil {
		t.Fatalf("setup: Decode: %v", err)
	}
	code, err := totp.Generate(raw, at)
	if err != nil {
		t.Fatalf("setup: Generate: %v", err)
	}
	return code
}

// --- BeginEnrollment ---------------------------------------------------------

func TestMFAUseCase_BeginEnrollment(t *testing.T) {
	t.Run("un-enrolled admin: persists pending secret + returns otpauth URI", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return freshAdminUser(), nil
			},
		}
		uc := usecases.NewMFAUseCase(repo, nil, testIssuer)
		uri, secret, err := uc.BeginEnrollment(context.Background(), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(uri, "otpauth://totp/") {
			t.Errorf("uri must start with otpauth://totp/, got %q", uri)
		}
		if !strings.Contains(uri, "secret="+secret) {
			t.Errorf("uri must embed Base32 secret, got %q", uri)
		}
		if !strings.Contains(uri, "issuer="+testIssuer) {
			t.Errorf("uri must embed issuer %q, got %q", testIssuer, uri)
		}
		if len(secret) != entities.MFASecretLength {
			t.Errorf("secret length: want %d, got %d", entities.MFASecretLength, len(secret))
		}
		if repo.lastSaved == nil {
			t.Fatalf("expected Save to be called")
		}
		if repo.lastSaved.MFASecret == nil || repo.lastSaved.MFASecret.String() != secret {
			t.Errorf("Save must persist the generated secret; got %v", repo.lastSaved.MFASecret)
		}
		if repo.lastSaved.MFAEnabled {
			t.Errorf("MFAEnabled must remain false until confirmation")
		}
	})

	t.Run("already-enrolled admin returns ErrMFAAlreadyEnabled", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"), nil
			},
		}
		uc := usecases.NewMFAUseCase(repo, nil, testIssuer)
		_, _, err := uc.BeginEnrollment(context.Background(), 42)
		if !errors.Is(err, entities.ErrMFAAlreadyEnabled) {
			t.Errorf("want ErrMFAAlreadyEnabled, got %v", err)
		}
		if repo.lastSaved != nil {
			t.Errorf("must not call Save on already-enrolled user")
		}
	})

	t.Run("GetByIDForAuth error propagates", func(t *testing.T) {
		dbErr := errors.New("connection refused")
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) { return nil, dbErr },
		}
		uc := usecases.NewMFAUseCase(repo, nil, testIssuer)
		_, _, err := uc.BeginEnrollment(context.Background(), 42)
		if !errors.Is(err, dbErr) {
			t.Errorf("want wrapped db error, got %v", err)
		}
	})
}

// --- ConfirmEnrollment -------------------------------------------------------

func TestMFAUseCase_ConfirmEnrollment(t *testing.T) {
	const pending = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

	t.Run("valid code flips MFAEnabled to true", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminWithPendingSecret(t, pending), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.ConfirmEnrollment(context.Background(), 42, codeAt(t, pending, frozenTime))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.lastSaved == nil {
			t.Fatalf("Save must be called")
		}
		if !repo.lastSaved.MFAEnabled {
			t.Errorf("MFAEnabled must be true after confirmation")
		}
	})

	t.Run("invalid code returns ErrInvalidMFACode", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminWithPendingSecret(t, pending), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.ConfirmEnrollment(context.Background(), 42, "999999")
		if !errors.Is(err, entities.ErrInvalidMFACode) {
			t.Errorf("want ErrInvalidMFACode, got %v", err)
		}
		if repo.lastSaved != nil {
			t.Errorf("Save must not be called on invalid code")
		}
	})

	t.Run("missing pending secret returns ErrMFANotPending", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return freshAdminUser(), nil // no MFASecret
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.ConfirmEnrollment(context.Background(), 42, "123456")
		if !errors.Is(err, entities.ErrMFANotPending) {
			t.Errorf("want ErrMFANotPending, got %v", err)
		}
	})

	t.Run("already-enrolled returns ErrMFAAlreadyEnabled", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, pending), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.ConfirmEnrollment(context.Background(), 42, codeAt(t, pending, frozenTime))
		if !errors.Is(err, entities.ErrMFAAlreadyEnabled) {
			t.Errorf("want ErrMFAAlreadyEnabled, got %v", err)
		}
	})
}

// --- Disable -----------------------------------------------------------------

func TestMFAUseCase_Disable(t *testing.T) {
	const enrolled = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

	t.Run("valid code clears MFA state", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, enrolled), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.Disable(context.Background(), 42, codeAt(t, enrolled, frozenTime))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.lastSaved == nil {
			t.Fatalf("Save must be called")
		}
		if repo.lastSaved.MFAEnabled {
			t.Errorf("MFAEnabled must be false after disable")
		}
		if repo.lastSaved.MFASecret != nil {
			t.Errorf("MFASecret must be nil after disable, got %v", repo.lastSaved.MFASecret)
		}
	})

	t.Run("invalid code returns ErrInvalidMFACode", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, enrolled), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.Disable(context.Background(), 42, "000000")
		if !errors.Is(err, entities.ErrInvalidMFACode) {
			t.Errorf("want ErrInvalidMFACode, got %v", err)
		}
	})

	t.Run("non-enrolled returns ErrMFANotEnabled", func(t *testing.T) {
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return freshAdminUser(), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, nil, testIssuer, func() time.Time { return frozenTime })
		err := uc.Disable(context.Background(), 42, "123456")
		if !errors.Is(err, entities.ErrMFANotEnabled) {
			t.Errorf("want ErrMFANotEnabled, got %v", err)
		}
	})
}

// --- Constructor guards ------------------------------------------------------

func TestNewMFAUseCase_NilRepoPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when userRepo is nil")
		}
	}()
	_ = usecases.NewMFAUseCase(nil, nil, testIssuer)
}
