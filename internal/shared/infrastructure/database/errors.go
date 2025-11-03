package database

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

// MapPostgresError maps PostgreSQL errors to domain errors
func MapPostgresError(err error) error {
	if err == nil {
		return nil
	}

	// Handle sql.ErrNoRows
	if errors.Is(err, sql.ErrNoRows) {
		return domainErrors.ErrNotFound
	}

	// Handle PostgreSQL errors
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // unique_violation
			return domainErrors.ErrAlreadyExists
		case "23503": // foreign_key_violation
			return domainErrors.NewDomainError(
				"FOREIGN_KEY_VIOLATION",
				"Referenced resource does not exist",
				err,
			)
		case "23502": // not_null_violation
			return domainErrors.ErrRequiredField
		case "22001": // string_data_right_truncation
			return domainErrors.ErrInvalidLength
		case "22P02": // invalid_text_representation
			return domainErrors.ErrInvalidFormat
		case "23514": // check_violation
			return domainErrors.NewDomainError(
				"CHECK_VIOLATION",
				"Value does not satisfy check constraint",
				err,
			)
		default:
			return domainErrors.NewDomainError(
				"DATABASE_ERROR",
				"Database operation failed",
				err,
			)
		}
	}

	// Return original error wrapped as domain error
	return domainErrors.NewDomainError(
		"DATABASE_ERROR",
		"Database operation failed",
		err,
	)
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, domainErrors.ErrNotFound) || errors.Is(err, sql.ErrNoRows)
}

// IsAlreadyExistsError checks if error is an already exists error
func IsAlreadyExistsError(err error) bool {
	if errors.Is(err, domainErrors.ErrAlreadyExists) {
		return true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" // unique_violation
	}

	return false
}
