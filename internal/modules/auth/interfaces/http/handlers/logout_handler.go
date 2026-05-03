package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// LogoutHandler exposes POST /api/auth/logout. It pulls the access token
// out of the Authorization header and asks LogoutUseCase to add its JTI
// to the revoked-token set. The client is expected to discard both
// access and refresh tokens; this endpoint is responsible for the
// access-token side via the JTI blacklist.
type LogoutHandler struct {
	usecase *usecases.LogoutUseCase
}

// NewLogoutHandler wires a LogoutHandler.
func NewLogoutHandler(usecase *usecases.LogoutUseCase) *LogoutHandler {
	return &LogoutHandler{usecase: usecase}
}

// Logout handles POST /api/auth/logout.
//
// On success: 204 No Content. On invalid/missing token: 401.
func (h *LogoutHandler) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется токен авторизации"))
		return
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == header || token == "" {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется Bearer токен"))
		return
	}

	if err := h.usecase.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Не удалось выполнить выход"))
		return
	}
	c.Status(http.StatusNoContent)
}
