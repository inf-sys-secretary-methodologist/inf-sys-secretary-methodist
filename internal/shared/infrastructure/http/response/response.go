package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response базовая структура всех API ответов
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Meta    Meta        `json:"meta"`
}

// Error структура для ошибок в API ответах
type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta метаданные ответа
type Meta struct {
	Timestamp  string      `json:"timestamp"`
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination информация о пагинации для списков
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ValidationError детали ошибок валидации
type ValidationError struct {
	Fields map[string][]string `json:"fields"`
}

// Success создает успешный ответ
func Success(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// SuccessWithMeta создает успешный ответ с дополнительными метаданными
func SuccessWithMeta(data interface{}, meta Meta) Response {
	meta.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// List создает успешный ответ для списка с пагинацией
func List(data interface{}, pagination Pagination) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta: Meta{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Pagination: &pagination,
		},
	}
}

// ErrorResponse создает ответ с ошибкой
func ErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// ErrorWithDetails создает ответ с ошибкой и дополнительными деталями
func ErrorWithDetails(code, message string, details interface{}) Response {
	return Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// ValidationErrorResponse создает ответ для ошибок валидации
func ValidationErrorResponse(fields map[string][]string) Response {
	return ErrorWithDetails(
		"VALIDATION_ERROR",
		"Ошибка валидации",
		ValidationError{Fields: fields},
	)
}

// NotFound создает ответ для ошибки "не найдено"
func NotFound(entity string) Response {
	return ErrorResponse(
		"NOT_FOUND",
		entity+" не найден",
	)
}

// Unauthorized создает ответ для ошибки "не авторизован"
func Unauthorized(message string) Response {
	if message == "" {
		message = "Доступ не авторизован"
	}
	return ErrorResponse("UNAUTHORIZED", message)
}

// Forbidden создает ответ для ошибки "доступ запрещен"
func Forbidden(message string) Response {
	if message == "" {
		message = "Доступ запрещен"
	}
	return ErrorResponse("FORBIDDEN", message)
}

// InternalError создает ответ для внутренней ошибки сервера
func InternalError(message string) Response {
	if message == "" {
		message = "Внутренняя ошибка сервера"
	}
	return ErrorResponse("INTERNAL_ERROR", message)
}

// BadRequest создает ответ для ошибки "неверный запрос"
func BadRequest(message string) Response {
	return ErrorResponse("BAD_REQUEST", message)
}

// WriteJSON записывает JSON ответ в http.ResponseWriter
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteSuccess записывает успешный ответ
func WriteSuccess(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusOK, Success(data))
}

// WriteCreated записывает ответ для созданного ресурса
func WriteCreated(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusCreated, Success(data))
}

// WriteNoContent записывает ответ без контента
func WriteNoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// WriteError записывает ответ с ошибкой
func WriteError(w http.ResponseWriter, status int, resp Response) error {
	return WriteJSON(w, status, resp)
}

// WithRequestID добавляет Request ID в мета-данные
func WithRequestID(resp Response, requestID string) Response {
	resp.Meta.RequestID = requestID
	return resp
}
