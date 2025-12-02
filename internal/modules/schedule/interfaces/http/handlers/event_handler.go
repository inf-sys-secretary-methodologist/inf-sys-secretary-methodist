// Package handlers contains HTTP request handlers for the schedule module.
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// EventHandler handles HTTP requests for event endpoints.
type EventHandler struct {
	usecase   *usecases.EventUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewEventHandler creates a new event handler.
func NewEventHandler(usecase *usecases.EventUseCase) *EventHandler {
	return &EventHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Create handles event creation
// @Summary Create a new event
// @Tags Events
// @Accept json
// @Produce json
// @Param input body dto.CreateEventInput true "Event data"
// @Success 201 {object} response.Response{data=dto.EventOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/events [post]
func (h *EventHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var input dto.CreateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса: " + err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Title = h.sanitizer.SanitizeString(input.Title)
	if input.Description != nil {
		desc := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &desc
	}
	if input.Location != nil {
		loc := h.sanitizer.SanitizeString(*input.Location)
		input.Location = &loc
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	event, err := h.usecase.Create(ctx, input, userID.(int64))
	if err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(event)
	c.JSON(http.StatusCreated, resp)
}

// Update handles event update
// @Summary Update an event
// @Tags Events
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param input body dto.UpdateEventInput true "Event data"
// @Success 200 {object} response.Response{data=dto.EventOutput}
// @Failure 400,404 {object} response.Response
// @Router /api/v1/events/{id} [put]
func (h *EventHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	if input.Title != nil {
		title := h.sanitizer.SanitizeString(*input.Title)
		input.Title = &title
	}
	if input.Description != nil {
		desc := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &desc
	}
	if input.Location != nil {
		loc := h.sanitizer.SanitizeString(*input.Location)
		input.Location = &loc
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	event, err := h.usecase.Update(ctx, id, input, userID.(int64))
	if err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(event)
	c.JSON(http.StatusOK, resp)
}

// Delete handles event deletion
// @Summary Delete an event
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} response.Response
// @Failure 400,404 {object} response.Response
// @Router /api/v1/events/{id} [delete]
func (h *EventHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Delete(ctx, id, userID.(int64)); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// GetByID handles getting an event by ID
// @Summary Get an event by ID
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} response.Response{data=dto.EventOutput}
// @Failure 404 {object} response.Response
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	event, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		resp := response.NotFound("Событие не найдено")
		c.JSON(http.StatusNotFound, resp)
		return
	}

	resp := response.Success(event)
	c.JSON(http.StatusOK, resp)
}

// List handles listing events with filters
// @Summary List events
// @Tags Events
// @Param organizer_id query int false "Organizer ID"
// @Param event_type query string false "Event type"
// @Param status query string false "Event status"
// @Param start_from query string false "Start from (RFC3339)"
// @Param start_to query string false "Start to (RFC3339)"
// @Param search query string false "Search query"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} response.Response{data=dto.EventListOutput}
// @Router /api/v1/events [get]
func (h *EventHandler) List(c *gin.Context) {
	var input dto.EventFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		resp := response.BadRequest("Неверные параметры запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	ctx := c.Request.Context()
	result, err := h.usecase.List(ctx, input)
	if err != nil {
		resp := response.InternalError("Ошибка при получении списка событий")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetByDateRange handles getting events by date range
// @Summary Get events by date range
// @Tags Events
// @Param start query string true "Start date (RFC3339)"
// @Param end query string true "End date (RFC3339)"
// @Success 200 {object} response.Response{data=[]dto.EventOutput}
// @Router /api/v1/events/range [get]
func (h *EventHandler) GetByDateRange(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		resp := response.BadRequest("Параметры start и end обязательны")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		resp := response.BadRequest("Неверный формат даты start")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		resp := response.BadRequest("Неверный формат даты end")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var userID *int64
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(int64)
		userID = &id
	}

	ctx := c.Request.Context()
	events, err := h.usecase.GetByDateRange(ctx, start, end, userID)
	if err != nil {
		resp := response.InternalError("Ошибка при получении событий")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(events)
	c.JSON(http.StatusOK, resp)
}

// GetUpcoming handles getting upcoming events
// @Summary Get upcoming events
// @Tags Events
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response{data=[]dto.EventOutput}
// @Router /api/v1/events/upcoming [get]
func (h *EventHandler) GetUpcoming(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := c.Request.Context()
	events, err := h.usecase.GetUpcoming(ctx, userID.(int64), limit)
	if err != nil {
		resp := response.InternalError("Ошибка при получении событий")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(events)
	c.JSON(http.StatusOK, resp)
}

// Cancel handles event cancellation
// @Summary Cancel an event
// @Tags Events
// @Param id path int true "Event ID"
// @Success 200 {object} response.Response{data=dto.EventOutput}
// @Router /api/v1/events/{id}/cancel [post]
func (h *EventHandler) Cancel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	event, err := h.usecase.Cancel(ctx, id, userID.(int64))
	if err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(event)
	c.JSON(http.StatusOK, resp)
}

// Reschedule handles event rescheduling
// @Summary Reschedule an event
// @Tags Events
// @Param id path int true "Event ID"
// @Param input body object{start_time=string,end_time=string} true "New times"
// @Success 200 {object} response.Response{data=dto.EventOutput}
// @Router /api/v1/events/{id}/reschedule [post]
func (h *EventHandler) Reschedule(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input struct {
		StartTime time.Time  `json:"start_time" binding:"required"`
		EndTime   *time.Time `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	event, err := h.usecase.Reschedule(ctx, id, input.StartTime, input.EndTime, userID.(int64))
	if err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(event)
	c.JSON(http.StatusOK, resp)
}

// AddParticipants handles adding participants to an event
// @Summary Add participants to an event
// @Tags Events
// @Param id path int true "Event ID"
// @Param input body dto.AddParticipantsInput true "Participants"
// @Success 200 {object} response.Response
// @Router /api/v1/events/{id}/participants [post]
func (h *EventHandler) AddParticipants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.AddParticipantsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.AddParticipants(ctx, id, input, userID.(int64)); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// RemoveParticipant handles removing a participant from an event
// @Summary Remove a participant from an event
// @Tags Events
// @Param id path int true "Event ID"
// @Param user_id path int true "User ID"
// @Success 200 {object} response.Response
// @Router /api/v1/events/{id}/participants/{user_id} [delete]
func (h *EventHandler) RemoveParticipant(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	participantID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID участника")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.RemoveParticipant(ctx, eventID, participantID, userID.(int64)); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// UpdateParticipantStatus handles updating participant response status
// @Summary Update participant status (accept/decline)
// @Tags Events
// @Param id path int true "Event ID"
// @Param input body dto.UpdateParticipantStatusInput true "Status"
// @Success 200 {object} response.Response
// @Router /api/v1/events/{id}/respond [post]
func (h *EventHandler) UpdateParticipantStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID события")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateParticipantStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.UpdateParticipantStatus(ctx, id, input, userID.(int64)); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// GetPendingInvitations handles getting pending invitations
// @Summary Get pending event invitations
// @Tags Events
// @Success 200 {object} response.Response{data=[]dto.EventOutput}
// @Router /api/v1/events/invitations [get]
func (h *EventHandler) GetPendingInvitations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	ctx := c.Request.Context()
	events, err := h.usecase.GetPendingInvitations(ctx, userID.(int64))
	if err != nil {
		resp := response.InternalError("Ошибка при получении приглашений")
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := response.Success(events)
	c.JSON(http.StatusOK, resp)
}
