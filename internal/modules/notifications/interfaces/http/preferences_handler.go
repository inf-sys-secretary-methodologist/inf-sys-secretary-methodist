// Package http contains HTTP handlers for the notifications module.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
)

// PreferencesHandler handles notification preferences HTTP requests
type PreferencesHandler struct {
	preferencesUseCase *usecases.PreferencesUseCase
	validate           *validator.Validate
}

// NewPreferencesHandler creates a new preferences handler
func NewPreferencesHandler(preferencesUseCase *usecases.PreferencesUseCase) *PreferencesHandler {
	return &PreferencesHandler{
		preferencesUseCase: preferencesUseCase,
		validate:           validator.New(),
	}
}

// Get godoc
// @Summary Get notification preferences
// @Description Get notification preferences for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} dto.PreferencesOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/preferences [get]
func (h *PreferencesHandler) Get(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.preferencesUseCase.Get(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update godoc
// @Summary Update notification preferences
// @Description Update notification preferences for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param preferences body dto.PreferencesInput true "Preferences data"
// @Success 200 {object} dto.PreferencesOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/preferences [put]
func (h *PreferencesHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input dto.PreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.preferencesUseCase.Update(c.Request.Context(), userID.(int64), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ToggleChannel godoc
// @Summary Toggle notification channel
// @Description Enable or disable a specific notification channel
// @Tags notifications
// @Accept json
// @Produce json
// @Param toggle body dto.ChannelToggleInput true "Channel toggle data"
// @Success 200 {object} dto.PreferencesOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/preferences/channel [put]
func (h *PreferencesHandler) ToggleChannel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input dto.ChannelToggleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.preferencesUseCase.ToggleChannel(c.Request.Context(), userID.(int64), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateQuietHours godoc
// @Summary Update quiet hours
// @Description Update quiet hours settings for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param quiet_hours body dto.QuietHoursInput true "Quiet hours data"
// @Success 200 {object} dto.PreferencesOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/preferences/quiet-hours [put]
func (h *PreferencesHandler) UpdateQuietHours(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input dto.QuietHoursInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.preferencesUseCase.UpdateQuietHours(c.Request.Context(), userID.(int64), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Reset godoc
// @Summary Reset notification preferences
// @Description Reset notification preferences to defaults for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} dto.PreferencesOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/preferences/reset [post]
func (h *PreferencesHandler) Reset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.preferencesUseCase.Reset(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTimezones godoc
// @Summary Get available timezones
// @Description Get a list of available timezones for quiet hours configuration
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Router /api/notifications/timezones [get]
func (h *PreferencesHandler) GetTimezones(c *gin.Context) {
	timezones := h.preferencesUseCase.GetAvailableTimezones()
	c.JSON(http.StatusOK, gin.H{"timezones": timezones})
}
