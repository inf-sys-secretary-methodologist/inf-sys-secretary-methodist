# HTTP Response Package

Стандартизированная структура для всех HTTP API ответов.

## Структура ответов

### Success Response

```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "Test"
  },
  "meta": {
    "timestamp": "2025-11-03T12:00:00Z",
    "request_id": "req-abc-123"
  }
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  },
  "meta": {
    "timestamp": "2025-11-03T12:00:00Z"
  }
}
```

### Validation Error Response

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": {
        "email": ["Email is required", "Invalid email format"],
        "password": ["Password must be at least 8 characters"]
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-03T12:00:00Z"
  }
}
```

### List Response (с пагинацией)

```json
{
  "success": true,
  "data": [
    {"id": "1", "name": "Item 1"},
    {"id": "2", "name": "Item 2"}
  ],
  "meta": {
    "timestamp": "2025-11-03T12:00:00Z",
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 150,
      "total_pages": 8
    }
  }
}
```

## Использование

### Success Response

```go
import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"

func (h *Handler) GetUser(c *gin.Context) {
    user := &User{ID: "123", Name: "John"}

    resp := response.Success(user)
    c.JSON(http.StatusOK, resp)
}

// Или короче:
func (h *Handler) GetUser(c *gin.Context) {
    user := &User{ID: "123", Name: "John"}
    response.WriteSuccess(c.Writer, user)
}
```

### Error Response

```go
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(ctx, id)
    if err != nil {
        // Автоматический маппинг domain error → HTTP error
        httpErr := response.MapDomainError(err)
        c.JSON(httpErr.Status, httpErr.Response)
        return
    }

    c.JSON(http.StatusOK, response.Success(user))
}

// Или еще проще:
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(ctx, id)
    if err != nil {
        response.HandleError(c.Writer, err)
        return
    }

    response.WriteSuccess(c.Writer, user)
}
```

### Validation Errors

```go
func (h *Handler) CreateUser(c *gin.Context) {
    validationErrors := map[string][]string{
        "email":    []string{"Email is required", "Invalid format"},
        "password": []string{"Password too short"},
    }

    resp := response.ValidationErrorResponse(validationErrors)
    c.JSON(http.StatusUnprocessableEntity, resp)
}
```

### List с пагинацией

```go
func (h *Handler) ListUsers(c *gin.Context) {
    users := []User{...}

    pagination := response.Pagination{
        Page:       1,
        PerPage:    20,
        Total:      150,
        TotalPages: 8,
    }

    resp := response.List(users, pagination)
    c.JSON(http.StatusOK, resp)
}
```

### Request ID

```go
func (h *Handler) GetUser(c *gin.Context) {
    requestID := c.GetString("request_id") // из middleware

    user := &User{ID: "123"}
    resp := response.Success(user)
    resp = response.WithRequestID(resp, requestID)

    c.JSON(http.StatusOK, resp)
}
```

## Domain Error Mapping

Package автоматически маппит domain errors в HTTP статус коды:

| Domain Error              | HTTP Status | Error Code           |
|---------------------------|-------------|----------------------|
| `ErrNotFound`             | 404         | `NOT_FOUND`          |
| `ErrAlreadyExists`        | 409         | `ALREADY_EXISTS`     |
| `ErrInvalidInput`         | 400         | `BAD_REQUEST`        |
| `ErrValidationFailed`     | 422         | `VALIDATION_ERROR`   |
| `ErrUnauthorized`         | 401         | `UNAUTHORIZED`       |
| `ErrForbidden`            | 403         | `FORBIDDEN`          |
| `ErrRequiredField`        | 400         | `BAD_REQUEST`        |
| `ErrInvalidFormat`        | 400         | `BAD_REQUEST`        |
| `ErrInvalidLength`        | 400         | `BAD_REQUEST`        |
| `ErrInternalError`        | 500         | `INTERNAL_ERROR`     |
| Неизвестная ошибка        | 500         | `INTERNAL_ERROR`     |

## Helper Functions

### Success Helpers
- `Success(data)` - успешный ответ
- `List(data, pagination)` - список с пагинацией
- `SuccessWithMeta(data, meta)` - с кастомными метаданными

### Error Helpers
- `ErrorResponse(code, message)` - базовая ошибка
- `ErrorWithDetails(code, message, details)` - с деталями
- `ValidationErrorResponse(fields)` - ошибка валидации
- `NotFound(entity)` - 404
- `Unauthorized(message)` - 401
- `Forbidden(message)` - 403
- `InternalError(message)` - 500
- `BadRequest(message)` - 400

### HTTP Writers
- `WriteJSON(w, status, data)` - базовая запись JSON
- `WriteSuccess(w, data)` - запись success (200)
- `WriteCreated(w, data)` - запись created (201)
- `WriteNoContent(w)` - пустой ответ (204)
- `WriteError(w, status, resp)` - запись ошибки

### Error Mappers
- `MapDomainError(err)` - маппинг domain → HTTP
- `HandleError(w, err)` - маппинг + запись
- `HandleErrorWithRequestID(w, err, requestID)` - с Request ID

## Best Practices

1. **Всегда используйте `response.Success()` для успешных ответов**
   - Обеспечивает консистентность
   - Автоматически добавляет timestamp

2. **Используйте `response.MapDomainError()` для ошибок**
   - Правильный HTTP status
   - Не показывает внутренние детали в production

3. **Добавляйте Request ID из middleware**
   ```go
   resp = response.WithRequestID(resp, requestID)
   ```

4. **Для списков всегда используйте пагинацию**
   ```go
   response.List(items, pagination)
   ```

5. **Не показывайте технические детали ошибок клиенту**
   - В production: "Internal server error"
   - В dev: можно добавить stack trace в details

## Тестирование

Package имеет **96.5% покрытие** unit тестами.

Запуск тестов:
```bash
go test ./internal/shared/infrastructure/http/response/... -v
go test ./internal/shared/infrastructure/http/response/... -cover
```

## Примеры интеграции

См. `internal/modules/auth/interfaces/http/handlers/auth_handler.go` для примера использования.
