package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// MFAService is the narrow interface MFAHandler depends on; the concrete
// implementation is *usecases.MFAUseCase. Keeping it local to the handler
// keeps test mocks small and follows DIP.
type MFAService interface {
	BeginEnrollment(ctx context.Context, userID int64) (otpAuthURI string, secret string, err error)
	ConfirmEnrollment(ctx context.Context, userID int64, code string) error
	Disable(ctx context.Context, userID int64, code string) error
}

// MFAHandler exposes /api/auth/mfa/{begin,confirm,disable} endpoints.
type MFAHandler struct {
	svc MFAService
}

// NewMFAHandler builds a handler. Panics on nil dependency (failure-closed).
func NewMFAHandler(svc MFAService) *MFAHandler {
	if svc == nil {
		panic("mfa handler: svc is nil")
	}
	return &MFAHandler{svc: svc}
}

type mfaCodeRequest struct {
	Code string `json:"code"`
}

// Begin handles POST /api/auth/mfa/begin.
func (h *MFAHandler) Begin(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется авторизация"))
		return
	}

	uri, secret, err := h.svc.BeginEnrollment(c.Request.Context(), userID)
	if err != nil {
		respondMFAError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"otpauth_uri": uri,
		"secret":      secret,
	}))
}

// Confirm handles POST /api/auth/mfa/confirm with body {"code": "123456"}.
func (h *MFAHandler) Confirm(c *gin.Context) {
	h.handleCodeAction(c, h.svc.ConfirmEnrollment)
}

// Disable handles POST /api/auth/mfa/disable with body {"code": "123456"}.
func (h *MFAHandler) Disable(c *gin.Context) {
	h.handleCodeAction(c, h.svc.Disable)
}

func (h *MFAHandler) handleCodeAction(c *gin.Context, action func(context.Context, int64, string) error) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется авторизация"))
		return
	}

	var req mfaCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Неверный формат запроса"))
		return
	}
	if !isValidMFACode(req.Code) {
		c.JSON(http.StatusBadRequest, response.BadRequest("Код должен состоять из 6 цифр"))
		return
	}

	if err := action(c.Request.Context(), userID, req.Code); err != nil {
		respondMFAError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"ok": true}))
}

// userIDFromContext extracts the authenticated user_id stamped onto the gin
// context by JWTMiddleware. Returns (0, false) if missing — the caller maps
// that to 401 so unauthenticated traffic never reaches the use case.
func userIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	if !ok || id == 0 {
		return 0, false
	}
	return id, true
}

// isValidMFACode enforces the standard 6-digit numeric form before invoking
// the use case, so malformed input fails fast at the boundary instead of
// burning a TOTP verification round-trip on obviously bad data.
func isValidMFACode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// respondMFAError maps known domain errors to HTTP status codes; opaque
// errors fall through to 500 to avoid leaking implementation details.
func respondMFAError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entities.ErrMFAAlreadyEnabled),
		errors.Is(err, entities.ErrMFANotEnabled),
		errors.Is(err, entities.ErrMFANotPending):
		c.JSON(http.StatusConflict, response.ErrorResponse("MFA_STATE_CONFLICT", err.Error()))
	case errors.Is(err, entities.ErrInvalidMFACode):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("INVALID_MFA_CODE", err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError("Не удалось обработать запрос"))
	}
}
