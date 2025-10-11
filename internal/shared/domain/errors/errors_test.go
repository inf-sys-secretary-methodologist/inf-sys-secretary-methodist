package errors

import (
	"errors"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DomainError
		expected string
	}{
		{
			name: "error without wrapped error",
			err: &DomainError{
				Code:    "TEST_ERROR",
				Message: "test error message",
				Err:     nil,
			},
			expected: "test error message",
		},
		{
			name: "error with wrapped error",
			err: &DomainError{
				Code:    "TEST_ERROR",
				Message: "test error message",
				Err:     errors.New("wrapped error"),
			},
			expected: "test error message: wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	domainErr := &DomainError{
		Code:    "TEST_ERROR",
		Message: "test error",
		Err:     wrappedErr,
	}

	if got := domainErr.Unwrap(); !errors.Is(got, wrappedErr) {
		t.Errorf("Unwrap() = %v, want %v", got, wrappedErr)
	}
}

func TestNewDomainError(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	domainErr := NewDomainError("TEST_CODE", "test message", wrappedErr)

	if domainErr.Code != "TEST_CODE" {
		t.Errorf("expected Code 'TEST_CODE', got '%s'", domainErr.Code)
	}
	if domainErr.Message != "test message" {
		t.Errorf("expected Message 'test message', got '%s'", domainErr.Message)
	}
	if !errors.Is(domainErr.Err, wrappedErr) {
		t.Errorf("expected wrapped error to be preserved")
	}
}
