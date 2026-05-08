package entities

import (
	"encoding/base32"
	"errors"
	"fmt"
)

// MFASecret is a 160-bit TOTP shared secret encoded as 32-character Base32
// (RFC 4648 alphabet, no padding). The unexported field forces construction
// through NewMFASecret so invariants stay enforced inside the domain.
type MFASecret struct {
	encoded string
}

// ErrInvalidMFASecret indicates a malformed, empty, or wrong-length secret.
var ErrInvalidMFASecret = errors.New("invalid MFA secret")

// ErrMFAAlreadyEnabled is returned when enrollment is attempted on an account
// that already has MFA enabled.
var ErrMFAAlreadyEnabled = errors.New("MFA is already enabled")

// ErrMFANotEnabled is returned when disable is attempted on an account that
// does not have MFA enabled.
var ErrMFANotEnabled = errors.New("MFA is not enabled")

// ErrMFANotPending is returned when ConfirmEnrollment is called without a
// preceding BeginEnrollment (no pending secret on file).
var ErrMFANotPending = errors.New("MFA enrollment was not started")

// ErrInvalidMFACode is returned when a user-supplied TOTP code does not
// match the stored secret within the drift window.
var ErrInvalidMFACode = errors.New("invalid MFA code")

// MFASecretLength is the canonical Base32 length of a 160-bit TOTP secret.
const MFASecretLength = 32

// NewMFASecret validates and wraps a Base32-encoded 160-bit secret. Length
// must be MFASecretLength (32) and the input must decode under the canonical
// uppercase RFC 4648 Base32 alphabet without padding.
func NewMFASecret(encoded string) (MFASecret, error) {
	if len(encoded) != MFASecretLength {
		return MFASecret{}, ErrInvalidMFASecret
	}
	if _, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(encoded); err != nil {
		return MFASecret{}, ErrInvalidMFASecret
	}
	return MFASecret{encoded: encoded}, nil
}

// String returns the Base32 form for persistence and otpauth URIs.
func (s MFASecret) String() string {
	return s.encoded
}

// Decode returns the raw 20-byte secret used for HMAC-SHA1.
func (s MFASecret) Decode() ([]byte, error) {
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s.encoded)
	if err != nil {
		return nil, fmt.Errorf("decode mfa secret: %w", err)
	}
	return raw, nil
}

// BeginMFAEnrollment stores a pending TOTP secret on the account but keeps
// MFAEnabled=false until the first code is confirmed. Returns
// ErrMFAAlreadyEnabled if the account already has MFA active. Re-calling
// while a previous secret is pending intentionally overwrites it so the
// user can restart enrollment.
//
// Timestamp updates are intentionally left to the caller (use case) — the
// domain method enforces invariants only; clock injection is a use-case
// concern, so we don't double-set UpdatedAt here.
func (u *User) BeginMFAEnrollment(secret MFASecret) error {
	if u.MFAEnabled {
		return ErrMFAAlreadyEnabled
	}
	u.MFASecret = &secret
	u.MFAEnabled = false
	return nil
}

// EnableMFA flips the account to fully enrolled. Returns ErrMFAAlreadyEnabled
// if MFA is already on. Caller is responsible for bumping UpdatedAt.
func (u *User) EnableMFA(secret MFASecret) error {
	if u.MFAEnabled {
		return ErrMFAAlreadyEnabled
	}
	u.MFASecret = &secret
	u.MFAEnabled = true
	return nil
}

// DisableMFA clears MFA state. Returns ErrMFANotEnabled when the account is
// not enrolled. Caller is responsible for bumping UpdatedAt.
func (u *User) DisableMFA() error {
	if !u.MFAEnabled {
		return ErrMFANotEnabled
	}
	u.MFASecret = nil
	u.MFAEnabled = false
	return nil
}
