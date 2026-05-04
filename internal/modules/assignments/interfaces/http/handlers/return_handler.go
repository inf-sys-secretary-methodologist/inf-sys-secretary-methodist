package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ReturnSubmissionUseCasePort is the narrow port through which the
// handler invokes the use case. Defining it here (rather than importing
// the concrete *ReturnSubmissionUseCase) keeps handler tests free of
// fake repositories / audit sinks — only the use-case behaviour is
// stubbed. Same pattern as SaveGradeUseCasePort.
type ReturnSubmissionUseCasePort interface {
	Execute(ctx context.Context, actorID int64, in assignUsecases.ReturnSubmissionInput) error
}

// ReturnHandler handles HTTP requests for returning a submission for
// revision (status: pending|graded → returned).
type ReturnHandler struct {
	usecase ReturnSubmissionUseCasePort
}

// NewReturnHandler wires the handler. The use case is required (non-nil):
// a nil use case would let requests reach a panic deeper in the call
// stack instead of failing during DI wiring. Failure-closed posture
// established for analytics in v0.108.3 and grading in v0.109.0.
func NewReturnHandler(usecase ReturnSubmissionUseCasePort) *ReturnHandler {
	if usecase == nil {
		panic("assignments: NewReturnHandler requires non-nil usecase")
	}
	return &ReturnHandler{usecase: usecase}
}

// ReturnSubmissionRequest is the JSON body schema for
// POST /api/assignments/:id/returns. Exported so swag can generate the
// schema in the OpenAPI spec.
type ReturnSubmissionRequest struct {
	StudentID int64  `json:"student_id" example:"7"`
	Reason    string `json:"reason"     example:"revisit derivation"`
}

// Return marks a student's submission as returned for revision.
// @Summary Return a student's submission for revision
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Assignment ID"
// @Param body body ReturnSubmissionRequest true "Return payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id}/returns [post]
func (h *ReturnHandler) Return(c *gin.Context) {
	assignmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}

	var body ReturnSubmissionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	if body.StudentID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("student_id is required and must be positive"))
		return
	}

	actorID, ok := actorIDFromContext(c)
	if !ok {
		// Failure-closed: missing or unrecognised role context surfaces
		// as 401 rather than falling through to a silent admin identity.
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	in := assignUsecases.ReturnSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    body.StudentID,
		Reason:       body.Reason,
	}
	if err := h.usecase.Execute(c.Request.Context(), actorID, in); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"assignment_id": assignmentID,
		"student_id":    body.StudentID,
		"reason":        body.Reason,
	}))
}

// actorIDFromContext extracts the user_id and role from gin context,
// applies the failure-closed role whitelist (teacher / methodist /
// academic_secretary / system_admin), and returns the user_id for the
// use case to authorise. Defence in depth: a future engineer who
// removes RequireNonStudent or adds a new role to RequireRole must NOT
// silently get write access via this handler. Unknown role → ok=false →
// 401, identical posture as the read-side AssignmentsHandler.
func actorIDFromContext(c *gin.Context) (int64, bool) {
	userID, ok := userIDFromContext(c)
	if !ok {
		return 0, false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return 0, false
	}
	roleStr, _ := roleVal.(string)
	switch roleStr {
	case roleTeacher, roleMethodist, roleAcademicSecretary, roleSystemAdmin:
		return userID, true
	default:
		return 0, false
	}
}

// handleError maps domain sentinels to HTTP responses. Explicit
// sentinel-first via errors.Is BEFORE the generic MapDomainError
// fallback — same pattern as grade_handler / assignments_handler.
func (h *ReturnHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrAssignmentNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("assignment"))
		return
	case errors.Is(err, entities.ErrAssignmentScopeForbidden):
		c.JSON(http.StatusForbidden, response.Forbidden("you cannot return submissions on this assignment"))
		return
	case errors.Is(err, entities.ErrAlreadyReturned):
		c.JSON(http.StatusConflict, response.ErrorResponse("ALREADY_RETURNED",
			"this submission is already returned"))
		return
	case errors.Is(err, entities.ErrInvalidReturn):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
