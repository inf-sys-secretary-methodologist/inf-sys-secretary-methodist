// Package handlers exposes HTTP endpoints для the extracurricular
// events bounded context.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	extUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Narrow ports — handler stays test-friendly + swaps к fakes easily.

// CreateEventPort is the narrow port для CreateEvent usecase.
type CreateEventPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in extUsecases.CreateEventInput) (*entities.ExtracurricularEvent, error)
}

// UpdateEventPort is the narrow port для UpdateEvent usecase.
type UpdateEventPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, in extUsecases.UpdateEventInput) (*entities.ExtracurricularEvent, error)
}

// DeleteEventPort is the narrow port для DeleteEvent usecase.
type DeleteEventPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, isAdmin bool, eventID int64) error
}

// GetEventPort is the narrow port для GetEvent usecase.
type GetEventPort interface {
	Execute(ctx context.Context, actorRole string, isAdmin bool, eventID int64) (*entities.ExtracurricularEvent, error)
}

// ListEventsPort is the narrow port для ListEvents usecase.
type ListEventsPort interface {
	Execute(ctx context.Context, actorRole string, isAdmin bool, in extUsecases.ListEventsInput) (repositories.EventListResult, error)
}

// RegisterParticipantPort is the narrow port для RegisterParticipant.
type RegisterParticipantPort interface {
	Execute(ctx context.Context, actorID int64, eventID int64) error
}

// UnregisterParticipantPort is the narrow port для UnregisterParticipant.
type UnregisterParticipantPort interface {
	Execute(ctx context.Context, actorID int64, eventID int64) error
}

// EventHandler exposes the 7 extracurricular endpoints over HTTP.
type EventHandler struct {
	create     CreateEventPort
	update     UpdateEventPort
	del        DeleteEventPort
	get        GetEventPort
	list       ListEventsPort
	register   RegisterParticipantPort
	unregister UnregisterParticipantPort
}

// NewEventHandler wires the handler. All ports required (non-nil) —
// nil dependencies would surface as nil-pointer panics under load
// instead of failing at DI wiring (mirror к failure-closed posture
// в curriculum handler).
func NewEventHandler(
	create CreateEventPort,
	update UpdateEventPort,
	del DeleteEventPort,
	get GetEventPort,
	list ListEventsPort,
	register RegisterParticipantPort,
	unregister UnregisterParticipantPort,
) *EventHandler {
	if create == nil || update == nil || del == nil || get == nil ||
		list == nil || register == nil || unregister == nil {
		panic("extracurricular: NewEventHandler requires non-nil ports")
	}
	return &EventHandler{
		create: create, update: update, del: del, get: get,
		list: list, register: register, unregister: unregister,
	}
}

// EventDTO is the public response shape для one event. Timestamps
// encoded as RFC 3339 strings.
type EventDTO struct {
	ID               int64            `json:"id"`
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	Category         string           `json:"category"`
	TargetAudience   string           `json:"target_audience"`
	Status           string           `json:"status"`
	Location         string           `json:"location"`
	StartAt          string           `json:"start_at"`
	EndAt            string           `json:"end_at"`
	MaxCapacity      *int             `json:"max_capacity,omitempty"`
	OrganizerID      int64            `json:"organizer_id"`
	Participants     []ParticipantDTO `json:"participants,omitempty"`
	ParticipantCount int              `json:"participant_count"`
	Version          int              `json:"version"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

// ParticipantDTO is the projected shape of one participant row.
type ParticipantDTO struct {
	UserID       int64  `json:"user_id"`
	RegisteredAt string `json:"registered_at"`
}

// EventSummaryDTO is the projected shape для one row в list response.
type EventSummaryDTO struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Category         string `json:"category"`
	TargetAudience   string `json:"target_audience"`
	Status           string `json:"status"`
	Location         string `json:"location"`
	StartAt          string `json:"start_at"`
	EndAt            string `json:"end_at"`
	MaxCapacity      *int   `json:"max_capacity,omitempty"`
	OrganizerID      int64  `json:"organizer_id"`
	ParticipantCount int    `json:"participant_count"`
	Version          int    `json:"version"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// EventsListResponse is the page response shape.
type EventsListResponse struct {
	Items []EventSummaryDTO `json:"items"`
	Total int               `json:"total"`
}

// CreateEventRequest is the JSON body для POST /events. binding tags
// per CLAUDE.md feedback (NOT `validate:`).
type CreateEventRequest struct {
	Title          string `json:"title"           binding:"required"`
	Description    string `json:"description"`
	Category       string `json:"category"        binding:"required"`
	TargetAudience string `json:"target_audience" binding:"required"`
	Location       string `json:"location"`
	StartAt        string `json:"start_at"        binding:"required"`
	EndAt          string `json:"end_at"          binding:"required"`
	MaxCapacity    *int   `json:"max_capacity"`
}

// UpdateEventRequest is the JSON body для PUT /events/:id.
type UpdateEventRequest struct {
	Title          string `json:"title"           binding:"required"`
	Description    string `json:"description"`
	Category       string `json:"category"        binding:"required"`
	TargetAudience string `json:"target_audience" binding:"required"`
	Location       string `json:"location"`
	StartAt        string `json:"start_at"        binding:"required"`
	EndAt          string `json:"end_at"          binding:"required"`
	MaxCapacity    *int   `json:"max_capacity"`
}

// mapEvent maps a domain event to the public DTO с participants.
func mapEvent(e *entities.ExtracurricularEvent) EventDTO {
	parts := e.Participants()
	out := make([]ParticipantDTO, 0, len(parts))
	for _, p := range parts {
		out = append(out, ParticipantDTO{
			UserID:       p.UserID,
			RegisteredAt: p.RegisteredAt.Format(time.RFC3339),
		})
	}
	return EventDTO{
		ID:               e.ID,
		Title:            e.Title(),
		Description:      e.Description(),
		Category:         string(e.Category()),
		TargetAudience:   string(e.TargetAudience()),
		Status:           string(e.Status()),
		Location:         e.Location(),
		StartAt:          e.StartAt().Format(time.RFC3339),
		EndAt:            e.EndAt().Format(time.RFC3339),
		MaxCapacity:      e.MaxCapacity(),
		OrganizerID:      e.OrganizerID(),
		Participants:     out,
		ParticipantCount: len(parts),
		Version:          e.Version(),
		CreatedAt:        e.CreatedAt().Format(time.RFC3339),
		UpdatedAt:        e.UpdatedAt().Format(time.RFC3339),
	}
}

func mapSummary(s repositories.EventSummary) EventSummaryDTO {
	return EventSummaryDTO{
		ID:               s.ID,
		Title:            s.Title,
		Category:         s.Category,
		TargetAudience:   s.TargetAudience,
		Status:           s.Status,
		Location:         s.Location,
		StartAt:          s.StartAt,
		EndAt:            s.EndAt,
		MaxCapacity:      s.MaxCapacity,
		OrganizerID:      s.OrganizerID,
		ParticipantCount: s.ParticipantCount,
		Version:          s.Version,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

// ===== Auth + parse helpers (module-local copies, mirror curriculum) =====

func authContext(c *gin.Context) (userID int64, role string, ok bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}
	switch v := uid.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		return 0, "", false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return 0, "", false
	}
	roleStr, _ := roleVal.(string)
	if roleStr == "" {
		return 0, "", false
	}
	return userID, roleStr, true
}

func parsePositiveID(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func isAdminRole(role string) bool { return role == "system_admin" }

// parseRFC3339OrFail parses the JSON string OR responds with 400.
// Returns (time, true) on success; (zero, false) and writes 400 response.
func parseRFC3339OrFail(c *gin.Context, field, raw string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid "+field+": expected RFC 3339 timestamp"))
		return time.Time{}, false
	}
	return t, true
}

// mapEventError maps domain / repository sentinels к HTTP status.
func mapEventError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrEventNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("event"))
	case errors.Is(err, repositories.ErrEventVersionConflict):
		c.JSON(http.StatusConflict, response.ErrorResponse("VERSION_CONFLICT", "event was modified concurrently; reload + retry"))
	case errors.Is(err, entities.ErrEventScopeForbidden):
		c.JSON(http.StatusForbidden, response.Forbidden("not authorized to operate on this event"))
	case errors.Is(err, entities.ErrCannotEditEvent):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("CANNOT_EDIT", err.Error()))
	case errors.Is(err, entities.ErrInvalidEvent):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("INVALID_EVENT", err.Error()))
	case errors.Is(err, entities.ErrParticipantExists):
		c.JSON(http.StatusConflict, response.ErrorResponse("ALREADY_REGISTERED", "already registered for this event"))
	case errors.Is(err, entities.ErrParticipantNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("participant"))
	case errors.Is(err, entities.ErrEventFull):
		c.JSON(http.StatusConflict, response.ErrorResponse("EVENT_FULL", "event at full capacity"))
	case errors.Is(err, entities.ErrEventNotOpenForRegistration):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("REGISTRATION_CLOSED", "event not open for registration"))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError("internal error"))
	}
}

// ===== Endpoints =====

// Create handles POST /api/v1/extracurricular/events.
// @Summary Create extracurricular event
// @Tags    extracurricular
// @Accept  json
// @Produce json
// @Param   body body CreateEventRequest true "Event payload"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events [post]
func (h *EventHandler) Create(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	var body CreateEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	startAt, ok := parseRFC3339OrFail(c, "start_at", body.StartAt)
	if !ok {
		return
	}
	endAt, ok := parseRFC3339OrFail(c, "end_at", body.EndAt)
	if !ok {
		return
	}
	e, err := h.create.Execute(c.Request.Context(), actorID, role, isAdminRole(role),
		extUsecases.CreateEventInput{
			Title:          body.Title,
			Description:    body.Description,
			Category:       entities.Category(body.Category),
			TargetAudience: entities.TargetAudience(body.TargetAudience),
			Location:       body.Location,
			StartAt:        startAt,
			EndAt:          endAt,
			MaxCapacity:    body.MaxCapacity,
		})
	if err != nil {
		mapEventError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapEvent(e)))
}

// Update handles PUT /api/v1/extracurricular/events/:id.
// @Summary Update extracurricular event
// @Tags    extracurricular
// @Accept  json
// @Produce json
// @Param   id   path int                 true "Event ID"
// @Param   body body UpdateEventRequest  true "Event payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events/{id} [put]
func (h *EventHandler) Update(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid event id"))
		return
	}
	var body UpdateEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	startAt, ok := parseRFC3339OrFail(c, "start_at", body.StartAt)
	if !ok {
		return
	}
	endAt, ok := parseRFC3339OrFail(c, "end_at", body.EndAt)
	if !ok {
		return
	}
	e, err := h.update.Execute(c.Request.Context(), actorID, role, isAdminRole(role),
		extUsecases.UpdateEventInput{
			ID:             id,
			Title:          body.Title,
			Description:    body.Description,
			Category:       entities.Category(body.Category),
			TargetAudience: entities.TargetAudience(body.TargetAudience),
			Location:       body.Location,
			StartAt:        startAt,
			EndAt:          endAt,
			MaxCapacity:    body.MaxCapacity,
		})
	if err != nil {
		mapEventError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapEvent(e)))
}

// Delete handles DELETE /api/v1/extracurricular/events/:id.
// @Summary Delete extracurricular event
// @Tags    extracurricular
// @Produce json
// @Param   id path int true "Event ID"
// @Success 204
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events/{id} [delete]
func (h *EventHandler) Delete(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid event id"))
		return
	}
	if err := h.del.Execute(c.Request.Context(), actorID, role, isAdminRole(role), id); err != nil {
		mapEventError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Get handles GET /api/v1/extracurricular/events/:id.
// @Summary Fetch event by id
// @Tags    extracurricular
// @Produce json
// @Param   id path int true "Event ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events/{id} [get]
func (h *EventHandler) Get(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid event id"))
		return
	}
	e, err := h.get.Execute(c.Request.Context(), role, isAdminRole(role), id)
	if err != nil {
		mapEventError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapEvent(e)))
}

// List handles GET /api/v1/extracurricular/events.
// @Summary List extracurricular events
// @Tags    extracurricular
// @Produce json
// @Param   status      query string false "Lifecycle status filter"
// @Param   category    query string false "Category filter"
// @Param   organizer_id query int    false "Organizer user id"
// @Param   from        query string false "Date lower bound (YYYY-MM-DD)"
// @Param   to          query string false "Date upper bound (YYYY-MM-DD)"
// @Param   limit       query int    false "Page size (default 100, max 200)"
// @Param   offset      query int    false "Page offset"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events [get]
func (h *EventHandler) List(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	if limit > 200 {
		limit = 200
	}
	organizer, _ := strconv.ParseInt(c.Query("organizer_id"), 10, 64)

	res, err := h.list.Execute(c.Request.Context(), role, isAdminRole(role),
		extUsecases.ListEventsInput{
			Status:      c.Query("status"),
			Category:    c.Query("category"),
			OrganizerID: organizer,
			FromDate:    c.Query("from"),
			ToDate:      c.Query("to"),
			Limit:       limit,
			Offset:      offset,
		})
	if err != nil {
		mapEventError(c, err)
		return
	}
	out := EventsListResponse{Items: make([]EventSummaryDTO, 0, len(res.Items)), Total: res.Total}
	for _, s := range res.Items {
		out.Items = append(out.Items, mapSummary(s))
	}
	c.JSON(http.StatusOK, response.Success(out))
}

// Register handles POST /api/v1/extracurricular/events/:id/register.
// Self-register only.
// @Summary Register caller для event
// @Tags    extracurricular
// @Produce json
// @Param   id path int true "Event ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events/{id}/register [post]
func (h *EventHandler) Register(c *gin.Context) {
	actorID, _, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid event id"))
		return
	}
	if err := h.register.Execute(c.Request.Context(), actorID, id); err != nil {
		mapEventError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"event_id": id, "status": "registered"}))
}

// Unregister handles DELETE /api/v1/extracurricular/events/:id/register.
// @Summary Unregister caller от event
// @Tags    extracurricular
// @Produce json
// @Param   id path int true "Event ID"
// @Success 204
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/extracurricular/events/{id}/register [delete]
func (h *EventHandler) Unregister(c *gin.Context) {
	actorID, _, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid event id"))
		return
	}
	if err := h.unregister.Execute(c.Request.Context(), actorID, id); err != nil {
		mapEventError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterExtracurricularRoutes mounts all 7 endpoints under
// /api/v1/extracurricular. Caller must apply auth middleware к the
// group before passing it in (no anonymous access — every endpoint
// requires authContext).
func RegisterExtracurricularRoutes(rg *gin.RouterGroup, h *EventHandler) {
	events := rg.Group("/extracurricular/events")
	events.POST("", h.Create)
	events.GET("", h.List)
	events.GET("/:id", h.Get)
	events.PUT("/:id", h.Update)
	events.DELETE("/:id", h.Delete)
	events.POST("/:id/register", h.Register)
	events.DELETE("/:id/register", h.Unregister)
}
