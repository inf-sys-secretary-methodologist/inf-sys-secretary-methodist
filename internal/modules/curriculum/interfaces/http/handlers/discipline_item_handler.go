package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// CreateDisciplineItemPort is the narrow port for the create use case.
type CreateDisciplineItemPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.CreateDisciplineItemInput) (*entities.DisciplineItem, error)
}

// GetDisciplineItemPort is the narrow port for the read use case.
type GetDisciplineItemPort interface {
	Execute(ctx context.Context, id int64) (*entities.DisciplineItem, error)
}

// ListDisciplineItemsPort is the narrow port for the list-by-section use case.
type ListDisciplineItemsPort interface {
	Execute(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error)
}

// UpdateDisciplineItemPort is the narrow port for the update use case.
type UpdateDisciplineItemPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.UpdateDisciplineItemInput) (*entities.DisciplineItem, error)
}

// DeleteDisciplineItemPort is the narrow port for the delete use case.
type DeleteDisciplineItemPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, itemID int64) error
}

// DisciplineItemHandler exposes 5 endpoints over HTTP.
//
//	POST   /api/sections/:sectionID/items   — create
//	GET    /api/sections/:sectionID/items   — list
//	GET    /api/items/:id                   — get
//	PUT    /api/items/:id                   — update
//	DELETE /api/items/:id                   — delete
type DisciplineItemHandler struct {
	create CreateDisciplineItemPort
	get    GetDisciplineItemPort
	list   ListDisciplineItemsPort
	update UpdateDisciplineItemPort
	del    DeleteDisciplineItemPort
}

// NewDisciplineItemHandler wires the handler. Failure-closed nil-panic.
func NewDisciplineItemHandler(
	create CreateDisciplineItemPort,
	get GetDisciplineItemPort,
	list ListDisciplineItemsPort,
	update UpdateDisciplineItemPort,
	del DeleteDisciplineItemPort,
) *DisciplineItemHandler {
	if create == nil || get == nil || list == nil || update == nil || del == nil {
		panic("discipline_item: NewDisciplineItemHandler requires non-nil ports (create / get / list / update / delete)")
	}
	return &DisciplineItemHandler{create: create, get: get, list: list, update: update, del: del}
}

// DisciplineItemDTO is the public response shape.
type DisciplineItemDTO struct {
	ID            int64  `json:"id"`
	SectionID     int64  `json:"section_id"`
	Title         string `json:"title"`
	HoursLectures int    `json:"hours_lectures"`
	HoursPractice int    `json:"hours_practice"`
	HoursLab      int    `json:"hours_lab"`
	HoursSelf     int    `json:"hours_self"`
	ControlForm   string `json:"control_form"`
	Credits       int    `json:"credits"`
	Semester      int    `json:"semester"`
	OrderIndex    int    `json:"order_index"`
	Version       int    `json:"version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// mapDisciplineItem projects the domain entity to its public DTO.
func mapDisciplineItem(d *entities.DisciplineItem) DisciplineItemDTO {
	return DisciplineItemDTO{
		ID:            d.ID,
		SectionID:     d.SectionID(),
		Title:         d.Title(),
		HoursLectures: d.HoursLectures(),
		HoursPractice: d.HoursPractice(),
		HoursLab:      d.HoursLab(),
		HoursSelf:     d.HoursSelf(),
		ControlForm:   string(d.ControlForm()),
		Credits:       d.Credits(),
		Semester:      d.Semester(),
		OrderIndex:    d.OrderIndex(),
		Version:       d.Version(),
		CreatedAt:     d.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     d.UpdatedAt().Format(time.RFC3339),
	}
}

// CreateDisciplineItemRequest is the JSON body schema для POST.
type CreateDisciplineItemRequest struct {
	Title         string `json:"title"          example:"Математический анализ"`
	HoursLectures int    `json:"hours_lectures" example:"36"`
	HoursPractice int    `json:"hours_practice" example:"36"`
	HoursLab      int    `json:"hours_lab"      example:"0"`
	HoursSelf     int    `json:"hours_self"     example:"72"`
	ControlForm   string `json:"control_form"   example:"exam"`
	Credits       int    `json:"credits"        example:"4"`
	Semester      int    `json:"semester"       example:"1"`
	OrderIndex    int    `json:"order_index"    example:"0"`
}

// UpdateDisciplineItemRequest is the JSON body schema для PUT.
type UpdateDisciplineItemRequest struct {
	Title         string `json:"title"`
	HoursLectures int    `json:"hours_lectures"`
	HoursPractice int    `json:"hours_practice"`
	HoursLab      int    `json:"hours_lab"`
	HoursSelf     int    `json:"hours_self"`
	ControlForm   string `json:"control_form"`
	Credits       int    `json:"credits"`
	Semester      int    `json:"semester"`
	OrderIndex    int    `json:"order_index"`
}

// DisciplineItemsListResponse is the response shape для list endpoint.
type DisciplineItemsListResponse struct {
	Items []DisciplineItemDTO `json:"items"`
}

// Create handles POST /api/sections/:sectionID/items.
// @Summary Create a discipline item in a section
// @Tags    discipline-items
// @Accept  json
// @Produce json
// @Param   sectionID path int true "Section ID"
// @Param   body body CreateDisciplineItemRequest true "Item payload"
// @Success 201 {object} response.Response
// @Security BearerAuth
// @Router  /api/sections/{sectionID}/items [post]
func (h *DisciplineItemHandler) Create(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may create discipline items"))
		return
	}
	sectionID, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	var body CreateDisciplineItemRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	d, err := h.create.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.CreateDisciplineItemInput{
			SectionID:     sectionID,
			Title:         body.Title,
			HoursLectures: body.HoursLectures,
			HoursPractice: body.HoursPractice,
			HoursLab:      body.HoursLab,
			HoursSelf:     body.HoursSelf,
			ControlForm:   entities.ControlForm(body.ControlForm),
			Credits:       body.Credits,
			Semester:      body.Semester,
			OrderIndex:    body.OrderIndex,
		})
	if err != nil {
		mapDisciplineItemError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapDisciplineItem(d)))
}

// Get handles GET /api/items/:id.
// @Summary Fetch a discipline item by id
// @Tags    discipline-items
// @Produce json
// @Security BearerAuth
// @Router  /api/items/{id} [get]
func (h *DisciplineItemHandler) Get(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this item view"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid item id"))
		return
	}
	d, err := h.get.Execute(c.Request.Context(), id)
	if err != nil {
		mapDisciplineItemError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapDisciplineItem(d)))
}

// List handles GET /api/sections/:sectionID/items.
// @Summary List all discipline items in a section
// @Tags    discipline-items
// @Produce json
// @Security BearerAuth
// @Router  /api/sections/{sectionID}/items [get]
func (h *DisciplineItemHandler) List(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this item view"))
		return
	}
	sectionID, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	items, err := h.list.Execute(c.Request.Context(), sectionID)
	if err != nil {
		mapDisciplineItemError(c, err)
		return
	}
	dtos := make([]DisciplineItemDTO, 0, len(items))
	for _, d := range items {
		dtos = append(dtos, mapDisciplineItem(d))
	}
	c.JSON(http.StatusOK, response.Success(DisciplineItemsListResponse{Items: dtos}))
}

// Update handles PUT /api/items/:id.
// @Summary Update a discipline item
// @Tags    discipline-items
// @Accept  json
// @Produce json
// @Security BearerAuth
// @Router  /api/items/{id} [put]
func (h *DisciplineItemHandler) Update(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may edit items"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid item id"))
		return
	}
	var body UpdateDisciplineItemRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	d, err := h.update.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.UpdateDisciplineItemInput{
			ID:            id,
			Title:         body.Title,
			HoursLectures: body.HoursLectures,
			HoursPractice: body.HoursPractice,
			HoursLab:      body.HoursLab,
			HoursSelf:     body.HoursSelf,
			ControlForm:   entities.ControlForm(body.ControlForm),
			Credits:       body.Credits,
			Semester:      body.Semester,
			OrderIndex:    body.OrderIndex,
		})
	if err != nil {
		mapDisciplineItemError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapDisciplineItem(d)))
}

// Delete handles DELETE /api/items/:id.
// @Summary Delete a discipline item
// @Tags    discipline-items
// @Produce json
// @Security BearerAuth
// @Router  /api/items/{id} [delete]
func (h *DisciplineItemHandler) Delete(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may delete items"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid item id"))
		return
	}
	if err := h.del.Execute(c.Request.Context(), actorID, isAdminRole(role), id); err != nil {
		mapDisciplineItemError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// mapDisciplineItemError maps domain / repository sentinels к HTTP statuses.
// Includes section + curriculum sentinels (cross-aggregate use cases
// surface them).
func mapDisciplineItemError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrDisciplineItemNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("discipline_item"))
		return
	case errors.Is(err, repositories.ErrSectionNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("section"))
		return
	case errors.Is(err, repositories.ErrCurriculumNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("curriculum"))
		return
	case errors.Is(err, repositories.ErrDisciplineItemVersionConflict):
		c.JSON(http.StatusConflict,
			response.ErrorResponse("VERSION_CONFLICT",
				"discipline item was modified by another request; reload and retry"))
		return
	case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
		c.JSON(http.StatusForbidden,
			response.Forbidden("only the curriculum author or an administrator may operate on this item"))
		return
	case errors.Is(err, entities.ErrCannotEditDisciplineItem):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_EDITABLE", "curriculum is not in an editable state"))
		return
	case errors.Is(err, entities.ErrInvalidDisciplineItem):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
		return
	}
}
