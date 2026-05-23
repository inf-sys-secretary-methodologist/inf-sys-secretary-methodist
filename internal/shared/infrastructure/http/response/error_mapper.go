// Package response provides HTTP response utilities and error mapping.
package response

import (
	"errors"
	"net/http"

	usersDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain"
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
			Response: InternalError("Неизвестная ошибка"),
		}
	}

	// Проверяем domain-specific ошибки
	switch {
	case errors.Is(err, domainErrors.ErrNotFound):
		return HTTPError{
			Status:   http.StatusNotFound,
			Response: NotFound("Ресурс"),
		}

	case errors.Is(err, domainErrors.ErrAlreadyExists):
		return HTTPError{
			Status:   http.StatusConflict,
			Response: ErrorResponse("ALREADY_EXISTS", "Ресурс уже существует"),
		}

	case errors.Is(err, domainErrors.ErrInvalidInput):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Предоставлены неверные данные"),
		}

	case errors.Is(err, domainErrors.ErrValidationFailed):
		return HTTPError{
			Status:   http.StatusUnprocessableEntity,
			Response: ErrorResponse("VALIDATION_ERROR", "Ошибка валидации"),
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
			Response: BadRequest("Отсутствует обязательное поле"),
		}

	case errors.Is(err, domainErrors.ErrInvalidFormat):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Неверный формат"),
		}

	case errors.Is(err, domainErrors.ErrInvalidLength):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: BadRequest("Неверная длина"),
		}

	// users module sentinels — closes #283 reviewer T0-1 (sentinel→HTTP
	// status mapping). Without these arms все four landed как 500
	// despite ADR-1..4 contracts promising 403/400/409/409.
	case errors.Is(err, usersDomain.ErrProfileEditForbidden):
		return HTTPError{
			Status:   http.StatusForbidden,
			Response: ErrorResponse("PROFILE_EDIT_FORBIDDEN", "Редактирование чужого профиля запрещено"),
		}

	case errors.Is(err, usersDomain.ErrInvalidAvatarKey):
		return HTTPError{
			Status:   http.StatusBadRequest,
			Response: ErrorResponse("INVALID_AVATAR_KEY", "Недопустимый ключ аватара"),
		}

	case errors.Is(err, usersDomain.ErrCannotDeleteSelf):
		return HTTPError{
			Status:   http.StatusConflict,
			Response: ErrorResponse("CANNOT_DELETE_SELF", "Нельзя удалить собственный аккаунт"),
		}

	case errors.Is(err, usersDomain.ErrLastAdminProtected):
		return HTTPError{
			Status:   http.StatusConflict,
			Response: ErrorResponse("LAST_ADMIN_PROTECTED", "Нельзя удалить или заблокировать последнего администратора"),
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
