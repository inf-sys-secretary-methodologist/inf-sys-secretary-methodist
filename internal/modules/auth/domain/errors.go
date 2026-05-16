package domain

import "errors"

// ErrPasswordResetTokenNotFound is returned by PasswordResetTokenRepository.LookupUser
// when the token is absent or has expired. Exposed as a domain sentinel so callers can
// distinguish "invalid/expired token" from a transport/storage failure via errors.Is,
// without parsing strings.
var ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
