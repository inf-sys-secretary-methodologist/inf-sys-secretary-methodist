package database

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func TestMapPostgresError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError error
		expectedCode  string
	}{
		{
			name:          "nil error returns nil",
			err:           nil,
			expectedError: nil,
		},
		{
			name:          "sql.ErrNoRows maps to ErrNotFound",
			err:           sql.ErrNoRows,
			expectedError: domainErrors.ErrNotFound,
		},
		{
			name:          "unique_violation maps to ErrAlreadyExists",
			err:           &pq.Error{Code: "23505"},
			expectedError: domainErrors.ErrAlreadyExists,
		},
		{
			name:         "foreign_key_violation maps to custom error",
			err:          &pq.Error{Code: "23503"},
			expectedCode: "FOREIGN_KEY_VIOLATION",
		},
		{
			name:          "not_null_violation maps to ErrRequiredField",
			err:           &pq.Error{Code: "23502"},
			expectedError: domainErrors.ErrRequiredField,
		},
		{
			name:          "string_data_right_truncation maps to ErrInvalidLength",
			err:           &pq.Error{Code: "22001"},
			expectedError: domainErrors.ErrInvalidLength,
		},
		{
			name:          "invalid_text_representation maps to ErrInvalidFormat",
			err:           &pq.Error{Code: "22P02"},
			expectedError: domainErrors.ErrInvalidFormat,
		},
		{
			name:         "check_violation maps to custom error",
			err:          &pq.Error{Code: "23514"},
			expectedCode: "CHECK_VIOLATION",
		},
		{
			name:         "unknown pq error maps to DATABASE_ERROR",
			err:          &pq.Error{Code: "99999"},
			expectedCode: "DATABASE_ERROR",
		},
		{
			name:         "generic error maps to DATABASE_ERROR",
			err:          errors.New("some database error"),
			expectedCode: "DATABASE_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapPostgresError(tt.err)

			if tt.expectedError != nil {
				assert.True(t, errors.Is(result, tt.expectedError),
					"Expected error %v, got %v", tt.expectedError, result)
			} else if tt.expectedCode != "" {
				var domainErr *domainErrors.DomainError
				assert.True(t, errors.As(result, &domainErr),
					"Expected DomainError, got %T", result)
				if domainErr != nil {
					assert.Equal(t, tt.expectedCode, domainErr.Code)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "sql.ErrNoRows is not found",
			err:      sql.ErrNoRows,
			expected: true,
		},
		{
			name:     "domainErrors.ErrNotFound is not found",
			err:      domainErrors.ErrNotFound,
			expected: true,
		},
		{
			name:     "generic error is not not found",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil is not not found",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAlreadyExistsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "domainErrors.ErrAlreadyExists is already exists",
			err:      domainErrors.ErrAlreadyExists,
			expected: true,
		},
		{
			name:     "unique_violation is already exists",
			err:      &pq.Error{Code: "23505"},
			expected: true,
		},
		{
			name:     "generic error is not already exists",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil is not already exists",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlreadyExistsError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
