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

// CreateSectionPort is the narrow port for the create use case.
type CreateSectionPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.CreateSectionInput) (*entities.Section, error)
}

// GetSectionPort is the narrow port for the read use case.
type GetSectionPort interface {
	Execute(ctx context.Context, id int64) (*entities.Section, error)
}

// ListSectionsPort is the narrow port for the list-by-curriculum use case.
type ListSectionsPort interface {
	Execute(ctx context.Context, curriculumID int64) ([]*entities.Section, error)
}

// UpdateSectionPort is the narrow port for the update use case.
type UpdateSectionPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.UpdateSectionInput) (*entities.Section, error)
}

// DeleteSectionPort is the narrow port for the delete use case.
type DeleteSectionPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, sectionID int64) error
}

// SectionHandler exposes the five section endpoints over HTTP. Routes:
//
//	POST   /api/curricula/:curriculumID/sections   — create
//	GET    /api/curricula/:curriculumID/sections   — list
//	GET    /api/sections/:id                       — get
//	PUT    /api/sections/:id                       — update
//	DELETE /api/sections/:id                       — delete
type SectionHandler struct {
	create CreateSectionPort
	get    GetSectionPort
	list   ListSectionsPort
	update UpdateSectionPort
	del    DeleteSectionPort
}

// NewSectionHandler wires the handler. All five ports are required
// (non-nil): nil dependencies would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring (mirror к
// NewCurriculumHandler failure-closed posture).
func NewSectionHandler(
	create CreateSectionPort,
	get GetSectionPort,
	list ListSectionsPort,
	update UpdateSectionPort,
	del DeleteSectionPort,
) *SectionHandler {
	if create == nil || get == nil || list == nil || update == nil || del == nil {
		panic("section: NewSectionHandler requires non-nil ports (create / get / list / update / delete)")
	}
	return &SectionHandler{create: create, get: get, list: list, update: update, del: del}
}

// SectionDTO is the public response shape for a section row.
// Timestamps encoded as RFC 3339 strings so frontend clients don't
// depend on Go time-marshal quirks.
type SectionDTO struct {
	ID           int64  `json:"id"`
	CurriculumID int64  `json:"curriculum_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	OrderIndex   int    `json:"order_index"`
	Version      int    `json:"version"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// mapSection projects the domain entity to its public DTO.
func mapSection(s *entities.Section) SectionDTO {
	return SectionDTO{
		ID:           s.ID,
		CurriculumID: s.CurriculumID(),
		Title:        s.Title(),
		Description:  s.Description(),
		OrderIndex:   s.OrderIndex(),
		Version:      s.Version(),
		CreatedAt:    s.CreatedAt().Format(time.RFC3339),
		UpdatedAt:    s.UpdatedAt().Format(time.RFC3339),
	}
}

// CreateSectionRequest is the JSON body schema for POST
// /api/curricula/:curriculumID/sections. Exported so swag generates
// the schema in the OpenAPI spec.
type CreateSectionRequest struct {
	Title       string `json:"title"       example:"Базовая часть"`
	Description string `json:"description" example:"Дисциплины обязательной части программы"`
	OrderIndex  int    `json:"order_index" example:"0"`
}

// UpdateSectionRequest is the JSON body schema for PUT /api/sections/:id.
type UpdateSectionRequest struct {
	Title       string `json:"title"       example:"Вариативная часть"`
	Description string `json:"description" example:"Дисциплины по выбору"`
	OrderIndex  int    `json:"order_index" example:"1"`
}

// SectionsListResponse is the response shape for the list endpoint.
type SectionsListResponse struct {
	Items []SectionDTO `json:"items"`
}

// Create handles POST /api/curricula/:curriculumID/sections.
// @Summary Create a section in a curriculum
// @Tags    sections
// @Accept  json
// @Produce json
// @Param   curriculumID path int                  true "Curriculum ID"
// @Param   body         body CreateSectionRequest true "Section payload"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curricula/{curriculumID}/sections [post]
func (h *SectionHandler) Create(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only academic_secretary or system_admin may create sections"))
		return
	}
	curID, ok := parsePositiveID(c.Param("curriculumID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}
	var body CreateSectionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	section, err := h.create.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.CreateSectionInput{
			CurriculumID: curID,
			Title:        body.Title,
			Description:  body.Description,
			OrderIndex:   body.OrderIndex,
		})
	if err != nil {
		mapSectionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapSection(section)))
}

// Get handles GET /api/sections/:id.
// @Summary  Fetch a section by id
// @Tags     sections
// @Produce  json
// @Param    id path int true "Section ID"
// @Success  200 {object} response.Response
// @Failure  400 {object} response.Response
// @Failure  401 {object} response.Response
// @Failure  403 {object} response.Response
// @Failure  404 {object} response.Response
// @Security BearerAuth
// @Router   /api/sections/{id} [get]
func (h *SectionHandler) Get(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this section view"))
		return
	}
	id, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	section, err := h.get.Execute(c.Request.Context(), id)
	if err != nil {
		mapSectionError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapSection(section)))
}

// List handles GET /api/curricula/:curriculumID/sections.
// @Summary  List all sections in a curriculum
// @Tags     sections
// @Produce  json
// @Param    curriculumID path int true "Curriculum ID"
// @Success  200 {object} response.Response
// @Failure  400 {object} response.Response
// @Failure  401 {object} response.Response
// @Failure  403 {object} response.Response
// @Security BearerAuth
// @Router   /api/curricula/{curriculumID}/sections [get]
func (h *SectionHandler) List(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this section view"))
		return
	}
	curID, ok := parsePositiveID(c.Param("curriculumID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}
	sections, err := h.list.Execute(c.Request.Context(), curID)
	if err != nil {
		mapSectionError(c, err)
		return
	}
	dtos := make([]SectionDTO, 0, len(sections))
	for _, s := range sections {
		dtos = append(dtos, mapSection(s))
	}
	c.JSON(http.StatusOK, response.Success(SectionsListResponse{Items: dtos}))
}

// Update handles PUT /api/sections/:id.
// @Summary  Update a section
// @Tags     sections
// @Accept   json
// @Produce  json
// @Param    id   path int                  true "Section ID"
// @Param    body body UpdateSectionRequest true "Section payload"
// @Success  200 {object} response.Response
// @Failure  400 {object} response.Response
// @Failure  401 {object} response.Response
// @Failure  403 {object} response.Response
// @Failure  404 {object} response.Response
// @Failure  409 {object} response.Response
// @Failure  422 {object} response.Response
// @Security BearerAuth
// @Router   /api/sections/{id} [put]
func (h *SectionHandler) Update(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only academic_secretary or system_admin may edit sections"))
		return
	}
	id, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	var body UpdateSectionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	section, err := h.update.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.UpdateSectionInput{
			ID:          id,
			Title:       body.Title,
			Description: body.Description,
			OrderIndex:  body.OrderIndex,
		})
	if err != nil {
		mapSectionError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapSection(section)))
}

// Delete handles DELETE /api/sections/:id.
// @Summary  Delete a section
// @Tags     sections
// @Produce  json
// @Param    id path int true "Section ID"
// @Success  204
// @Failure  400 {object} response.Response
// @Failure  401 {object} response.Response
// @Failure  403 {object} response.Response
// @Failure  404 {object} response.Response
// @Failure  422 {object} response.Response
// @Security BearerAuth
// @Router   /api/sections/{id} [delete]
func (h *SectionHandler) Delete(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only academic_secretary or system_admin may delete sections"))
		return
	}
	id, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	if err := h.del.Execute(c.Request.Context(), actorID, isAdminRole(role), id); err != nil {
		mapSectionError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// mapSectionError maps domain / repository sentinels surfaced by the
// section use cases to the matching HTTP status. Every sentinel is
// matched explicitly via errors.Is. The curriculum sentinels are
// included because cross-aggregate use cases (Create / Update /
// Delete) load a curriculum and may surface ErrCurriculumNotFound.
//
// Status mapping:
//   - ErrSectionNotFound, ErrCurriculumNotFound          → 404
//   - ErrSectionVersionConflict                          → 409
//   - ErrSectionScopeForbidden                           → 403
//   - ErrCannotEditSection                               → 422 (NOT_EDITABLE)
//   - ErrInvalidSection                                  → 422 (INVALID_INPUT)
//   - default                                            → 500 generic
func mapSectionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrSectionNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("section"))
		return
	case errors.Is(err, repositories.ErrCurriculumNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("curriculum"))
		return
	case errors.Is(err, repositories.ErrSectionVersionConflict):
		c.JSON(http.StatusConflict,
			response.ErrorResponse("VERSION_CONFLICT",
				"section was modified by another request; reload and retry"))
		return
	case errors.Is(err, entities.ErrSectionScopeForbidden):
		c.JSON(http.StatusForbidden,
			response.Forbidden("only the curriculum author or an administrator may operate on this section"))
		return
	case errors.Is(err, entities.ErrCannotEditSection):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_EDITABLE", "curriculum is not in an editable state"))
		return
	case errors.Is(err, entities.ErrInvalidSection):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
		return
	}
}
