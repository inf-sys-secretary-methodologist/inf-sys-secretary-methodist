package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/messages"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// PasswordResetHandler exposes the three endpoints of the password
// recovery flow: request a reset (email + store token), verify a token
// (read-only check before showing the form), and confirm a token to
// rotate the password.
type PasswordResetHandler struct {
	usecase *usecases.PasswordResetUseCase
}

// NewPasswordResetHandler wires a PasswordResetHandler.
func NewPasswordResetHandler(usecase *usecases.PasswordResetUseCase) *PasswordResetHandler {
	return &PasswordResetHandler{usecase: usecase}
}

// requestResetBody is the JSON shape for POST /password-reset/request.
type requestResetBody struct {
	Email string `json:"email" binding:"required,email"`
}

// confirmResetBody is the JSON shape for POST /password-reset/confirm.
type confirmResetBody struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RequestReset handles POST /api/auth/password-reset/request.
//
// Always returns 204 for a well-formed body, regardless of whether the
// email exists. Anti-enumeration is enforced by the usecase; the
// handler must not introduce a difference at the HTTP layer.
func (h *PasswordResetHandler) RequestReset(c *gin.Context) {
	var body requestResetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(messages.PasswordResetEmailRequired))
		return
	}
	if err := h.usecase.RequestReset(c.Request.Context(), body.Email); err != nil {
		// Genuine system fault (storage, email transport). Keep the
		// response generic — do not leak which dependency failed.
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
		return
	}
	c.Status(http.StatusNoContent)
}

// VerifyResetToken handles GET /api/auth/password-reset/verify/:token.
//
// Read-only check the frontend uses before rendering the new-password
// form. 204 if the token is currently usable, 410 Gone otherwise.
func (h *PasswordResetHandler) VerifyResetToken(c *gin.Context) {
	token := c.Param("token")
	if err := h.usecase.VerifyToken(c.Request.Context(), token); err != nil {
		if errors.Is(err, usecases.ErrInvalidResetToken) {
			c.JSON(http.StatusGone, response.ErrorResponse("RESET_TOKEN_EXPIRED",
				messages.PasswordResetTokenExpired))
			return
		}
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
		return
	}
	c.Status(http.StatusNoContent)
}

// ConfirmReset handles POST /api/auth/password-reset/confirm.
//
// Maps usecase errors to distinct HTTP codes so the frontend can
// render the right message:
//   - ErrWeakResetPassword -> 400 (user can fix and retry)
//   - ErrInvalidResetToken -> 410 Gone (link is dead, must re-request)
//   - other -> 500
func (h *PasswordResetHandler) ConfirmReset(c *gin.Context) {
	var body confirmResetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(messages.PasswordResetMalformedRequest))
		return
	}

	err := h.usecase.ConfirmReset(c.Request.Context(), body.Token, body.Password)
	if err == nil {
		c.Status(http.StatusNoContent)
		return
	}
	switch {
	case errors.Is(err, usecases.ErrWeakResetPassword):
		c.JSON(http.StatusBadRequest, response.BadRequest(messages.PasswordResetWeakPassword))
	case errors.Is(err, usecases.ErrInvalidResetToken):
		c.JSON(http.StatusGone, response.ErrorResponse("RESET_TOKEN_EXPIRED",
			messages.PasswordResetTokenExpired))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError(""))
	}
}
