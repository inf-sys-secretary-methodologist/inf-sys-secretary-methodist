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

// --- otpauth URI escaping ----------------------------------------------------

// TestBuildOTPAuthURI_LabelEscape verifies that issuer/email containing
// characters with special meaning inside the otpauth label segment
// (':', '/', non-ASCII, spaces) round-trip safely. The label format is
// `<issuer>:<email>`, so any unescaped colon inside the issuer would
// fool authenticator apps into splitting at the wrong position.
func TestBuildOTPAuthURI_LabelEscape(t *testing.T) {
	type tcase struct {
		name   string
		issuer string
		email  string
	}
	tests := []tcase{
		{"colon in issuer escaped", "App: Prod", "user@example.com"},
		{"slash in issuer escaped", "Acme/Org", "user@v"},
		{"non-ASCII in issuer escaped", "Café", "user@example.com"},
		{"colon in email escaped", "App", "user:weird@example.com"},
		{"plain ASCII passes through", "inf-sys-test", "admin@example.local"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			audit := &fakeAuditEmitter{}
			user := &entities.User{ID: 1, Email: tc.email}
			repo := &stubUserRepo{
				getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) { return user, nil },
			}
			uc := usecases.NewMFAUseCaseWithClock(repo, audit, tc.issuer, func() time.Time { return frozenTime })
			uri, _, err := uc.BeginEnrollment(context.Background(), 1)
			if err != nil {
				t.Fatalf("BeginEnrollment: %v", err)
			}

			// Split on `?` to isolate the label.
			parts := strings.SplitN(strings.TrimPrefix(uri, "otpauth://totp/"), "?", 2)
			if len(parts) != 2 {
				t.Fatalf("URI not in expected shape: %q", uri)
			}
			label := parts[0]
			// Exactly one ':' separator between issuer and email — any
			// extra unescaped ':' from issuer/email would break parsing.
			if strings.Count(label, ":") != 1 {
				t.Errorf("label %q must contain exactly one ':' separator; got %d", label, strings.Count(label, ":"))
			}
			// '/' anywhere in the label would terminate the path segment
			// and confuse authenticators that re-parse the URI.
			if strings.Contains(label, "/") {
				t.Errorf("label %q must not contain unescaped '/'", label)
			}
		})
	}
}

// --- Audit log emission ------------------------------------------------------

type recordedAuditEvent struct {
	action string
	userID int64
}

type fakeAuditEmitter struct {
	mu     sync.Mutex
	events []recordedAuditEvent
}

func (f *fakeAuditEmitter) LogAuditEvent(_ context.Context, action, _ string, fields map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, _ := fields["user_id"].(int64)
	f.events = append(f.events, recordedAuditEvent{action: action, userID: id})
}

func TestMFAUseCase_AuditLog(t *testing.T) {
	const enrolled = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

	t.Run("BeginEnrollment success emits mfa_enrollment_begin", func(t *testing.T) {
		audit := &fakeAuditEmitter{}
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return freshAdminUser(), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, audit, testIssuer, func() time.Time { return frozenTime })
		if _, _, err := uc.BeginEnrollment(context.Background(), 42); err != nil {
			t.Fatalf("BeginEnrollment: %v", err)
		}
		if len(audit.events) != 1 || audit.events[0].action != usecases.AuditActionMFAEnrollmentBegin || audit.events[0].userID != 42 {
			t.Errorf("audit events: want [{begin, 42}], got %+v", audit.events)
		}
	})

	t.Run("ConfirmEnrollment success emits mfa_enrollment_confirm", func(t *testing.T) {
		audit := &fakeAuditEmitter{}
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminWithPendingSecret(t, enrolled), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, audit, testIssuer, func() time.Time { return frozenTime })
		if err := uc.ConfirmEnrollment(context.Background(), 42, codeAt(t, enrolled, frozenTime)); err != nil {
			t.Fatalf("ConfirmEnrollment: %v", err)
		}
		if len(audit.events) != 1 || audit.events[0].action != usecases.AuditActionMFAEnrollmentConfirm {
			t.Errorf("audit events: want [{confirm, ...}], got %+v", audit.events)
		}
	})

	t.Run("Disable success emits mfa_disabled", func(t *testing.T) {
		audit := &fakeAuditEmitter{}
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, enrolled), nil
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, audit, testIssuer, func() time.Time { return frozenTime })
		if err := uc.Disable(context.Background(), 42, codeAt(t, enrolled, frozenTime)); err != nil {
			t.Fatalf("Disable: %v", err)
		}
		if len(audit.events) != 1 || audit.events[0].action != usecases.AuditActionMFADisabled {
			t.Errorf("audit events: want [{disabled, ...}], got %+v", audit.events)
		}
	})

	t.Run("BeginEnrollment failure emits no audit event", func(t *testing.T) {
		audit := &fakeAuditEmitter{}
		repo := &stubUserRepo{
			getByIDForAuth: func(_ context.Context, _ int64) (*entities.User, error) {
				return adminEnrolled(t, enrolled), nil // already enrolled → fails
			},
		}
		uc := usecases.NewMFAUseCaseWithClock(repo, audit, testIssuer, func() time.Time { return frozenTime })
		_, _, _ = uc.BeginEnrollment(context.Background(), 42)
		if len(audit.events) != 0 {
			t.Errorf("audit events: want none on failure, got %+v", audit.events)
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
