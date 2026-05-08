package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

const validSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP" // 32 Base32 chars = 20 bytes

func TestNewMFASecret(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid 32-char Base32", validSecret, nil},
		{"empty string rejected", "", entities.ErrInvalidMFASecret},
		{"too short rejected", "JBSWY3DPEHPK3PXP", entities.ErrInvalidMFASecret},
		{"too long rejected", validSecret + "AAAA", entities.ErrInvalidMFASecret},
		{"non-Base32 chars rejected", strings.Repeat("1", 32), entities.ErrInvalidMFASecret},
		{"lowercase rejected (canonical Base32 is upper)", strings.ToLower(validSecret), entities.ErrInvalidMFASecret},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := entities.NewMFASecret(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("NewMFASecret(%q): want %v, got %v", tc.input, tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tc.input {
				t.Errorf("String(): want %q, got %q", tc.input, got.String())
			}
		})
	}
}

func TestMFASecret_Decode(t *testing.T) {
	s, err := entities.NewMFASecret(validSecret)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	raw, err := s.Decode()
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(raw) != 20 {
		t.Errorf("Decode len: want 20, got %d", len(raw))
	}
}

func TestUser_BeginMFAEnrollment(t *testing.T) {
	secret, err := entities.NewMFASecret(validSecret)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("on un-enrolled user stores pending secret without flipping enabled", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.BeginMFAEnrollment(secret); err != nil {
			t.Fatalf("BeginMFAEnrollment: %v", err)
		}
		if u.MFAEnabled {
			t.Errorf("MFAEnabled must remain false until ConfirmEnrollment")
		}
		if u.MFASecret == nil || u.MFASecret.String() != validSecret {
			t.Errorf("pending secret must be stored; got %v", u.MFASecret)
		}
	})

	t.Run("on already-enrolled user returns ErrMFAAlreadyEnabled", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.EnableMFA(secret); err != nil {
			t.Fatalf("first enable: %v", err)
		}
		err := u.BeginMFAEnrollment(secret)
		if !errors.Is(err, entities.ErrMFAAlreadyEnabled) {
			t.Errorf("BeginMFAEnrollment: want ErrMFAAlreadyEnabled, got %v", err)
		}
	})

	t.Run("re-call replaces previously pending secret", func(t *testing.T) {
		const otherEncoded = "MFRGGZDFMZTWQ2LKMFRGGZDFMZTWQ2LK" // 32-char Base32
		other, err := entities.NewMFASecret(otherEncoded)
		if err != nil {
			t.Fatalf("setup: NewMFASecret(other): %v", err)
		}
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.BeginMFAEnrollment(secret); err != nil {
			t.Fatalf("first BeginMFAEnrollment: %v", err)
		}
		if err := u.BeginMFAEnrollment(other); err != nil {
			t.Fatalf("second BeginMFAEnrollment: %v", err)
		}
		if u.MFASecret == nil || u.MFASecret.String() != otherEncoded {
			t.Errorf("re-call must overwrite pending secret; got %v", u.MFASecret)
		}
		if u.MFAEnabled {
			t.Errorf("MFAEnabled must remain false after re-call")
		}
	})
}

func TestUser_EnableMFA(t *testing.T) {
	secret, err := entities.NewMFASecret(validSecret)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("on user without MFA enrolls and flags enabled", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.EnableMFA(secret); err != nil {
			t.Fatalf("EnableMFA: %v", err)
		}
		if !u.MFAEnabled {
			t.Errorf("MFAEnabled should be true after EnableMFA")
		}
		if u.MFASecret == nil || u.MFASecret.String() != validSecret {
			t.Errorf("MFASecret should be stored; got %v", u.MFASecret)
		}
	})

	t.Run("on user with MFA already enabled rejects", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.EnableMFA(secret); err != nil {
			t.Fatalf("first enable: %v", err)
		}
		err := u.EnableMFA(secret)
		if !errors.Is(err, entities.ErrMFAAlreadyEnabled) {
			t.Errorf("second EnableMFA: want ErrMFAAlreadyEnabled, got %v", err)
		}
	})
}

func TestUser_DisableMFA(t *testing.T) {
	secret, err := entities.NewMFASecret(validSecret)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("on enrolled user clears state", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		if err := u.EnableMFA(secret); err != nil {
			t.Fatalf("setup: %v", err)
		}
		if err := u.DisableMFA(); err != nil {
			t.Fatalf("DisableMFA: %v", err)
		}
		if u.MFAEnabled {
			t.Errorf("MFAEnabled should be false after DisableMFA")
		}
		if u.MFASecret != nil {
			t.Errorf("MFASecret should be cleared, got %v", u.MFASecret)
		}
	})

	t.Run("on non-enrolled user rejects", func(t *testing.T) {
		u := entities.NewUser("a@b.c", "hash", "Alice", domain.RoleType("system_admin"))
		err := u.DisableMFA()
		if !errors.Is(err, entities.ErrMFANotEnabled) {
			t.Errorf("DisableMFA: want ErrMFANotEnabled, got %v", err)
		}
	})
}
