package response

import (
	"errors"
	"net/http"

	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

// HTTPError объединяет HTTP status и Response
type HTTPError struct {
	Status   int
	Response Response
}

// MapDomainError преобразует domain error в HTTP error response
func MapDomainError(err error) HTTPError {
	if err == nil {
		return HTTPError{
			Status:   http.StatusInternalServerError,
			Response: InternalError("Unknown error"),
		}
	}

	// Проверяем domain-specific ошибки
	switch {
	case errors.Is(err, domainErrors.ErrNotFound):
		return HTTPError{
			Status:   http.StatusNotFound,
			Response: NotFound("Resource"),
		}

	case errors.Is(err, domainErrors.ErrAlreadyExists):
		return HTTPError{
			Status:   http.StatusConflict,
			Response: ErrorResponse("ALREADY_EXISTS", "Resource already exists"),
		}

	case errors.Is(err, domainErrors.ErrInvalidInput):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Invalid input provided"),
		}

	case errors.Is(err, domainErrors.ErrValidationFailed):
		return HTTPError{
			Status:   http.StatusUnprocessableEntity,
			Response: ErrorResponse("VALIDATION_ERROR", "Validation failed"),
		}

	case errors.Is(err, domainErrors.ErrUnauthorized):
		return HTTPError{
			Status:   http.StatusUnauthorized,
			Response: Unauthorized(""),
		}

	case errors.Is(err, domainErrors.ErrForbidden):
		return HTTPError{
			Status:   http.StatusForbidden,
			Response: Forbidden(""),
		}

	case errors.Is(err, domainErrors.ErrRequiredField):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Required field is missing"),
		}

	case errors.Is(err, domainErrors.ErrInvalidFormat):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Invalid format"),
		}

	case errors.Is(err, domainErrors.ErrInvalidLength):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Invalid length"),
		}
	}

	// Проверяем DomainError с кодом
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		return mapDomainErrorByCode(domainErr)
	}

	// По умолчанию - internal server error
	// В production не показываем технические детали
	return HTTPError{
		Status:   http.StatusInternalServerError,
		Response: InternalError(""),
	}
}

// mapDomainErrorByCode маппит DomainError по коду
func mapDomainErrorByCode(err *domainErrors.DomainError) HTTPError {
	switch err.Code {
	case "NOT_FOUND":
		return HTTPError{
			Status:   http.StatusNotFound,
			Response: ErrorResponse(err.Code, err.Message),
		}

	case "ALREADY_EXISTS", "CONFLICT":
		return HTTPError{
			Status:   http.StatusConflict,
			Response: ErrorResponse(err.Code, err.Message),
		}

	case "VALIDATION_ERROR", "INVALID_INPUT":
		return HTTPError{
			Status:   http.StatusUnprocessableEntity,
			Response: ErrorResponse(err.Code, err.Message),
		}

	case "UNAUTHORIZED":
		return HTTPError{
			Status:   http.StatusUnauthorized,
			Response: ErrorResponse(err.Code, err.Message),
		}

	case "FORBIDDEN":
		return HTTPError{
			Status:   http.StatusForbidden,
			Response: ErrorResponse(err.Code, err.Message),
		}

	case "BAD_REQUEST":
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: ErrorResponse(err.Code, err.Message),
		}

	default:
		return HTTPError{
			Status:   http.StatusInternalServerError,
			Response: InternalError(err.Message),
		}
	}
}

// HandleError маппит и записывает error response
func HandleError(w http.ResponseWriter, err error) error {
	httpErr := MapDomainError(err)
	return WriteError(w, httpErr.Status, httpErr.Response)
}

// HandleErrorWithRequestID маппит и записывает error response с Request ID
func HandleErrorWithRequestID(w http.ResponseWriter, err error, requestID string) error {
	httpErr := MapDomainError(err)
	resp := WithRequestID(httpErr.Response, requestID)
	return WriteError(w, httpErr.Status, resp)
}
