package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// BulkEditDisciplineItemsPort is the narrow port for the bulk-edit
// use case. Mirror к other ports — handler's view of the use case is
// just Execute.
type BulkEditDisciplineItemsPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.BulkEditDisciplineItemsInput) (*curUsecases.BulkEditDisciplineItemsResult, error)
}

// BulkDisciplineItemsHandler exposes the bulk-edit endpoint.
//
//	POST /api/sections/:sectionID/items/bulk — combined creates / updates / deletes
//
// Separate handler from DisciplineItemHandler — bulk endpoint has different
// request/response shapes и different error mapping (collect-all conflicts
// vs single-item denial).
type BulkDisciplineItemsHandler struct {
	bulk BulkEditDisciplineItemsPort
}

// NewBulkDisciplineItemsHandler wires the handler. Failure-closed nil-panic.
func NewBulkDisciplineItemsHandler(bulk BulkEditDisciplineItemsPort) *BulkDisciplineItemsHandler {
	if bulk == nil {
		panic("bulk_discipline_items: NewBulkDisciplineItemsHandler requires non-nil bulk port")
	}
	return &BulkDisciplineItemsHandler{bulk: bulk}
}

// ===== Request / Response DTOs =====

// BulkCreateItemRequest is the per-item create payload в bulk request.
// SectionID inherited from path :sectionID — не в body.
type BulkCreateItemRequest struct {
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

// BulkUpdateItemRequest is the per-item update payload в bulk request.
// Carries item id + new field values; version intentionally NOT in DTO
// (repo loads server-side fresh entity, optimistic-lock SQL guards race).
type BulkUpdateItemRequest struct {
	ID            int64  `json:"id"`
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

// BulkEditRequest is the JSON body schema для POST.
type BulkEditRequest struct {
	Creates []BulkCreateItemRequest `json:"creates"`
	Updates []BulkUpdateItemRequest `json:"updates"`
	Deletes []int64                 `json:"deletes"`
}

// BulkEditSuccessResponse is the 200 success body shape.
type BulkEditSuccessResponse struct {
	Created []DisciplineItemDTO `json:"created"`
	Updated []DisciplineItemDTO `json:"updated"`
	Deleted []int64             `json:"deleted"`
}

// BulkEditConflictItem is one optimistic-lock conflict entry в 409
// response. Mirror к ADR-12 shape.
type BulkEditConflictItem struct {
	ID              int64 `json:"id"`
	ExpectedVersion int   `json:"expected_version"`
	CurrentVersion  int   `json:"current_version"`
}

// BulkEditConflictResponse is the 409 VERSION_CONFLICT response shape
// per ADR-12 — collect-all conflicts reported в one render для UI merge.
type BulkEditConflictResponse struct {
	Error     string                 `json:"error"`
	Conflicts []BulkEditConflictItem `json:"conflicts"`
}

// ===== Handler method =====

// BulkEdit handles POST /api/sections/:sectionID/items/bulk.
// @Summary Atomically apply creates+updates+deletes to a section's items
// @Tags    discipline-items
// @Accept  json
// @Produce json
// @Param   sectionID path int true "Section ID"
// @Param   body body BulkEditRequest true "Bulk edit payload"
// @Success 200 {object} response.Response
// @Failure 409 {object} BulkEditConflictResponse
// @Security BearerAuth
// @Router  /api/sections/{sectionID}/items/bulk [post]
func (h *BulkDisciplineItemsHandler) BulkEdit(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only academic_secretary or system_admin may bulk-edit discipline items"))
		return
	}
	sectionID, ok := parsePositiveID(c.Param("sectionID"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid section id"))
		return
	}
	var body BulkEditRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	in := curUsecases.BulkEditDisciplineItemsInput{
		SectionID: sectionID,
		Creates:   make([]curUsecases.BulkCreateItem, 0, len(body.Creates)),
		Updates:   make([]curUsecases.BulkUpdateItem, 0, len(body.Updates)),
		Deletes:   body.Deletes,
	}
	for _, c := range body.Creates {
		in.Creates = append(in.Creates, curUsecases.BulkCreateItem{
			Title:         c.Title,
			HoursLectures: c.HoursLectures,
			HoursPractice: c.HoursPractice,
			HoursLab:      c.HoursLab,
			HoursSelf:     c.HoursSelf,
			ControlForm:   entities.ControlForm(c.ControlForm),
			Credits:       c.Credits,
			Semester:      c.Semester,
			OrderIndex:    c.OrderIndex,
		})
	}
	for _, u := range body.Updates {
		in.Updates = append(in.Updates, curUsecases.BulkUpdateItem{
			ID:            u.ID,
			Title:         u.Title,
			HoursLectures: u.HoursLectures,
			HoursPractice: u.HoursPractice,
			HoursLab:      u.HoursLab,
			HoursSelf:     u.HoursSelf,
			ControlForm:   entities.ControlForm(u.ControlForm),
			Credits:       u.Credits,
			Semester:      u.Semester,
			OrderIndex:    u.OrderIndex,
		})
	}

	res, err := h.bulk.Execute(c.Request.Context(), actorID, isAdminRole(role), in)
	if err != nil {
		// Version-conflict path: 409 с per-item conflict details (collect-all).
		if errors.Is(err, curUsecases.ErrBulkVersionConflict) && res != nil {
			conflicts := make([]BulkEditConflictItem, 0, len(res.Conflicts))
			for _, conflict := range res.Conflicts {
				conflicts = append(conflicts, BulkEditConflictItem{
					ID:              conflict.ID,
					ExpectedVersion: conflict.ExpectedVersion,
					CurrentVersion:  conflict.CurrentVersion,
				})
			}
			c.JSON(http.StatusConflict, BulkEditConflictResponse{
				Error:     "VERSION_CONFLICT",
				Conflicts: conflicts,
			})
			return
		}
		mapBulkEditError(c, err)
		return
	}

	created := make([]DisciplineItemDTO, 0, len(res.Created))
	for _, d := range res.Created {
		created = append(created, mapDisciplineItem(d))
	}
	updated := make([]DisciplineItemDTO, 0, len(res.Updated))
	for _, d := range res.Updated {
		updated = append(updated, mapDisciplineItem(d))
	}
	c.JSON(http.StatusOK, response.Success(BulkEditSuccessResponse{
		Created: created,
		Updated: updated,
		Deleted: res.Deleted,
	}))
}

// mapBulkEditError maps bulk-edit-specific sentinels к HTTP statuses.
// VERSION_CONFLICT handled inline в BulkEdit because it needs response
// body shape с conflict details.
func mapBulkEditError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, curUsecases.ErrEmptyBulkInput):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("EMPTY_BULK_INPUT", "bulk edit input must contain at least one create, update, or delete"))
		return
	case errors.Is(err, curUsecases.ErrCrossSectionBulkEdit):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("CROSS_SECTION_BULK_EDIT", "bulk edit target item belongs to a different section"))
		return
	case errors.Is(err, curUsecases.ErrBulkVersionConflict):
		// Defensive fallback: usecase returns this paired with non-nil
		// result; BulkEdit handler renders 409 с conflict details inline.
		// This case fires only on contract violation (nil result) — emit
		// 409 без details rather than 500 для consistent client semantics.
		c.JSON(http.StatusConflict, BulkEditConflictResponse{Error: "VERSION_CONFLICT"})
		return
	case errors.Is(err, repositories.ErrSectionNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("section"))
		return
	case errors.Is(err, repositories.ErrDisciplineItemNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("discipline_item"))
		return
	case errors.Is(err, entities.ErrCannotEditDisciplineItem):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_EDITABLE", "curriculum is not in an editable state"))
		return
	case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
		c.JSON(http.StatusForbidden,
			response.Forbidden("only the curriculum's author or admin may bulk-edit its items"))
		return
	case errors.Is(err, entities.ErrInvalidDisciplineItem):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
	}
}
