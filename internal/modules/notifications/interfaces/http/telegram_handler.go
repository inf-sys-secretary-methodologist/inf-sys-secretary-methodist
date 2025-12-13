// Package http contains HTTP handlers for the notifications module.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
)

// TelegramHandler handles Telegram-related HTTP requests
type TelegramHandler struct {
	verificationService *services.TelegramVerificationService
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(verificationService *services.TelegramVerificationService) *TelegramHandler {
	return &TelegramHandler{
		verificationService: verificationService,
	}
}

// GenerateVerificationCodeResponse represents the response for code generation
type GenerateVerificationCodeResponse struct {
	Code        string `json:"code"`
	ExpiresAt   string `json:"expires_at"`
	BotUsername string `json:"bot_username"`
	BotLink     string `json:"bot_link"`
}

// GenerateVerificationCode godoc
// @Summary Generate Telegram verification code
// @Description Generates a verification code for linking Telegram account
// @Tags telegram
// @Accept json
// @Produce json
// @Success 200 {object} GenerateVerificationCodeResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/telegram/verification-code [post]
func (h *TelegramHandler) GenerateVerificationCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.verificationService.GenerateVerificationCode(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate verification code"})
		return
	}

	c.JSON(http.StatusOK, GenerateVerificationCodeResponse{
		Code:        result.Code,
		ExpiresAt:   result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		BotUsername: result.BotUsername,
		BotLink:     result.BotLink,
	})
}

// TelegramConnectionResponse represents the Telegram connection status
type TelegramConnectionResponse struct {
	Connected   bool    `json:"connected"`
	Username    *string `json:"username,omitempty"`
	FirstName   *string `json:"first_name,omitempty"`
	ConnectedAt *string `json:"connected_at,omitempty"`
}

// GetConnectionStatus godoc
// @Summary Get Telegram connection status
// @Description Returns the current Telegram connection status for the user
// @Tags telegram
// @Accept json
// @Produce json
// @Success 200 {object} TelegramConnectionResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/telegram/status [get]
func (h *TelegramHandler) GetConnectionStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := h.verificationService.GetConnection(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get connection status"})
		return
	}

	if conn == nil {
		c.JSON(http.StatusOK, TelegramConnectionResponse{
			Connected: false,
		})
		return
	}

	connectedAt := conn.ConnectedAt.Format("2006-01-02T15:04:05Z07:00")
	c.JSON(http.StatusOK, TelegramConnectionResponse{
		Connected:   conn.IsActive,
		Username:    &conn.TelegramUsername,
		FirstName:   &conn.TelegramFirstName,
		ConnectedAt: &connectedAt,
	})
}

// DisconnectTelegram godoc
// @Summary Disconnect Telegram account
// @Description Removes the Telegram connection for the user
// @Tags telegram
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/telegram/disconnect [post]
func (h *TelegramHandler) DisconnectTelegram(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.verificationService.DisconnectTelegram(c.Request.Context(), userID.(int64))
	if err != nil {
		if err.Error() == "connection not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "telegram not connected"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disconnect telegram"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "telegram disconnected successfully"})
}
