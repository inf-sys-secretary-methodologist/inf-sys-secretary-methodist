package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDomainError_NotFound(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrNotFound)

	assert.Equal(t, http.StatusNotFound, httpErr.Status)
	assert.False(t, httpErr.Response.Success)
	assert.Equal(t, "NOT_FOUND", httpErr.Response.Error.Code)
}

func TestMapDomainError_AlreadyExists(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrAlreadyExists)

	assert.Equal(t, http.StatusConflict, httpErr.Status)
	assert.Equal(t, "ALREADY_EXISTS", httpErr.Response.Error.Code)
}

func TestMapDomainError_InvalidInput(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrInvalidInput)

	assert.Equal(t, http.StatusBadRequest, httpErr.Status)
	assert.Equal(t, "BAD_REQUEST", httpErr.Response.Error.Code)
}

func TestMapDomainError_ValidationFailed(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrValidationFailed)

	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.Status)
	assert.Equal(t, "VALIDATION_ERROR", httpErr.Response.Error.Code)
}

func TestMapDomainError_Unauthorized(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, httpErr.Status)
	assert.Equal(t, "UNAUTHORIZED", httpErr.Response.Error.Code)
}

func TestMapDomainError_Forbidden(t *testing.T) {
	httpErr := MapDomainError(domainErrors.ErrForbidden)

	assert.Equal(t, http.StatusForbidden, httpErr.Status)
	assert.Equal(t, "FORBIDDEN", httpErr.Response.Error.Code)
}

func TestMapDomainError_DomainErrorWithCode(t *testing.T) {
	tests := []struct {
		name           string
		domainErr      *domainErrors.DomainError
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "NOT_FOUND code",
			domainErr:      domainErrors.NewDomainError("NOT_FOUND", "User not found", nil),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "ALREADY_EXISTS code",
			domainErr:      domainErrors.NewDomainError("ALREADY_EXISTS", "Email already exists", nil),
			expectedStatus: http.StatusConflict,
			expectedCode:   "ALREADY_EXISTS",
		},
		{
			name:           "VALIDATION_ERROR code",
			domainErr:      domainErrors.NewDomainError("VALIDATION_ERROR", "Invalid data", nil),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name:           "UNAUTHORIZED code",
			domainErr:      domainErrors.NewDomainError("UNAUTHORIZED", "Token expired", nil),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "FORBIDDEN code",
			domainErr:      domainErrors.NewDomainError("FORBIDDEN", "Access denied", nil),
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "BAD_REQUEST code",
			domainErr:      domainErrors.NewDomainError("BAD_REQUEST", "Invalid JSON", nil),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Unknown code",
			domainErr:      domainErrors.NewDomainError("UNKNOWN", "Unknown error", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := MapDomainError(tt.domainErr)
			assert.Equal(t, tt.expectedStatus, httpErr.Status)
			assert.Equal(t, tt.expectedCode, httpErr.Response.Error.Code)
		})
	}
}

func TestMapDomainError_NilError(t *testing.T) {
	httpErr := MapDomainError(nil)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Status)
	assert.Equal(t, "INTERNAL_ERROR", httpErr.Response.Error.Code)
}

func TestMapDomainError_UnknownError(t *testing.T) {
	unknownErr := errors.New("some unknown error")
	httpErr := MapDomainError(unknownErr)

	assert.Equal(t, http.StatusInternalServerError, httpErr.Status)
	assert.Equal(t, "INTERNAL_ERROR", httpErr.Response.Error.Code)
}

func TestMapDomainError_WrappedError(t *testing.T) {
	wrappedErr := errors.Join(domainErrors.ErrNotFound, errors.New("additional context"))
	httpErr := MapDomainError(wrappedErr)

	assert.Equal(t, http.StatusNotFound, httpErr.Status)
	assert.Equal(t, "NOT_FOUND", httpErr.Response.Error.Code)
}

func TestHandleError(t *testing.T) {
	w := httptest.NewRecorder()
	err := domainErrors.ErrNotFound

	handleErr := HandleError(w, err)
	require.NoError(t, handleErr)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "NOT_FOUND")
}

func TestHandleErrorWithRequestID(t *testing.T) {
	w := httptest.NewRecorder()
	err := domainErrors.ErrUnauthorized
	requestID := "req-123-456"

	handleErr := HandleErrorWithRequestID(w, err, requestID)
	require.NoError(t, handleErr)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
	assert.Contains(t, w.Body.String(), requestID)
}

func TestMapDomainError_AllStandardErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "ErrNotFound",
			err:            domainErrors.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
		{
			name:           "ErrAlreadyExists",
			err:            domainErrors.ErrAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "ALREADY_EXISTS",
		},
		{
			name:           "ErrInvalidInput",
			err:            domainErrors.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "ErrUnauthorized",
			err:            domainErrors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "ErrForbidden",
			err:            domainErrors.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "FORBIDDEN",
		},
		{
			name:           "ErrValidationFailed",
			err:            domainErrors.ErrValidationFailed,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name:           "ErrRequiredField",
			err:            domainErrors.ErrRequiredField,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "ErrInvalidFormat",
			err:            domainErrors.ErrInvalidFormat,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "ErrInvalidLength",
			err:            domainErrors.ErrInvalidLength,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := MapDomainError(tt.err)
			assert.Equal(t, tt.expectedStatus, httpErr.Status, "Status code mismatch")
			assert.Equal(t, tt.expectedCode, httpErr.Response.Error.Code, "Error code mismatch")
		})
	}
}
