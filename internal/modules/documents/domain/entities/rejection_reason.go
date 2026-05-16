package entities

import (
	"errors"
	"fmt"
	"strings"
)

// ErrRejectionReasonInvalid signals an attempt to construct a
// RejectionReason от a string that does не satisfy the length
// invariant (10..500 characters after trimming). Sentinel so handler
// layer can errors.Is it and map к a stable 422 response.
//
// Issue: #227
var ErrRejectionReasonInvalid = errors.New("document: rejection reason invalid (length must be 10..500)")

const (
	rejectionReasonMinLen = 10
	rejectionReasonMaxLen = 500
)

// RejectionReason is a value object wrapping the obligatory reason
// text a methodist/secretary/admin attaches when rejecting a
// document. Invariant: length ≥10 и ≤500 characters (Unicode rune
// count) after trimming leading/trailing whitespace.
//
// Zero-value is intentionally invalid — Reject() refuses it via the
// IsZero check to defend against handler-layer omissions.
//
// Issue: #227
type RejectionReason struct {
	value string
}

// NewRejectionReason validates the raw string + returns the VO.
// Trims leading/trailing whitespace before measuring length.
// Returns ErrRejectionReasonInvalid for empty / whitespace-only /
// too short / too long inputs.
func NewRejectionReason(raw string) (RejectionReason, error) {
	trimmed := strings.TrimSpace(raw)
	length := len([]rune(trimmed))
	if length < rejectionReasonMinLen || length > rejectionReasonMaxLen {
		return RejectionReason{}, fmt.Errorf("%w: got %d chars", ErrRejectionReasonInvalid, length)
	}
	return RejectionReason{value: trimmed}, nil
}

// MustRejectionReason panics on validation failure. Convenience for
// test code that constructs known-valid reasons inline.
func MustRejectionReason(raw string) RejectionReason {
	r, err := NewRejectionReason(raw)
	if err != nil {
		panic(fmt.Sprintf("MustRejectionReason: invalid raw %q: %v", raw, err))
	}
	return r
}

// String returns the canonical (trimmed) reason text.
func (r RejectionReason) String() string { return r.value }

// IsZero reports whether the VO is the zero-value (empty string).
// Reject() uses this to enforce the reason-required invariant at the
// domain boundary в дополнение к handler-layer validation.
func (r RejectionReason) IsZero() bool { return strings.TrimSpace(r.value) == "" }
