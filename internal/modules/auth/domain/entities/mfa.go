package entities

import "errors"

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

// MFASecretLength is the canonical Base32 length of a 160-bit TOTP secret.
const MFASecretLength = 32

// NewMFASecret validates and wraps a Base32-encoded 160-bit secret.
func NewMFASecret(_ string) (MFASecret, error) {
	return MFASecret{}, nil // RED stub
}

// String returns the Base32 form for persistence and otpauth URIs.
func (s MFASecret) String() string {
	return s.encoded
}

// Decode returns the raw 20-byte secret used for HMAC-SHA1.
func (s MFASecret) Decode() ([]byte, error) {
	return nil, nil // RED stub
}

// EnableMFA enrolls the user with the provided secret. Returns
// ErrMFAAlreadyEnabled if the account already has MFA active.
func (u *User) EnableMFA(_ MFASecret) error {
	return nil // RED stub
}

// DisableMFA clears MFA state. Returns ErrMFANotEnabled when the account is
// not enrolled.
func (u *User) DisableMFA() error {
	return nil // RED stub
}
