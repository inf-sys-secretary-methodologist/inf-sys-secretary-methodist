package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ResubmitSubmissionUseCasePort is the narrow port through which the
// handler invokes the use case. Defining it here (rather than importing
// the concrete *ResubmitSubmissionUseCase) keeps handler tests free of
// fake repositories / audit sinks — only the use-case behavior is
// stubbed. Same pattern as ReturnSubmissionUseCasePort.
type ResubmitSubmissionUseCasePort interface {
	Execute(ctx context.Context, actorID int64, in assignUsecases.ResubmitSubmissionInput) error
}

// ResubmitHandler handles HTTP requests for a student resubmitting
// their own returned work (status: returned → pending). Counterpart to
// ReturnHandler on the student side of the academic loop.
type ResubmitHandler struct {
	usecase ResubmitSubmissionUseCasePort
}

// NewResubmitHandler wires the handler. The use case is required (non-nil):
// a nil use case would let requests reach a panic deeper in the call
// stack instead of failing during DI wiring. Failure-closed posture
// established for analytics in v0.108.3, grading in v0.109.0, and
// returning in v0.111.0.
func NewResubmitHandler(usecase ResubmitSubmissionUseCasePort) *ResubmitHandler {
	if usecase == nil {
		panic("assignments: NewResubmitHandler requires non-nil usecase")
	}
	return &ResubmitHandler{usecase: usecase}
}

// Resubmit transitions the authenticated student's submission back to
// pending so they can supply revisions.
//
// @Summary Student resubmits their own returned submission
// @Tags assignments
// @Accept json
// @Produce json
// @Param id path int true "Assignment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id}/resubmit [post]
func (h *ResubmitHandler) Resubmit(c *gin.Context) {
	assignmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || assignmentID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}

	// No request body — the student supplies no input. They are simply
	// re-submitting their own returned work; the assignment id in the
	// path plus the authenticated student id is enough to identify the
	// row. Skipping body parsing eliminates any chance of a body
	// student_id mismatching the JWT subject.

	studentID, ok := studentIDFromContext(c)
	if !ok {
		// Failure-closed: missing or non-student role context surfaces
		// as 401 rather than falling through to a silent admin identity.
		// Defense in depth on top of RequireRole("student") middleware.
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	in := assignUsecases.ResubmitSubmissionInput{
		AssignmentID: assignmentID,
		StudentID:    studentID,
	}
	if err := h.usecase.Execute(c.Request.Context(), studentID, in); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"assignment_id": assignmentID,
		"student_id":    studentID,
	}))
}

// studentIDFromContext extracts the user_id from gin context and applies
// the student-only role whitelist — the only role permitted on the
// resubmit endpoint. Defense in depth: a future engineer who removes
// RequireRole("student") at the route level must NOT silently let
// teachers, methodists, etc. resubmit on a student's behalf via this
// handler. Unknown role → ok=false → 401, identical posture to
// actorIDFromContext on the return side.
func studentIDFromContext(c *gin.Context) (int64, bool) {
	userID, ok := userIDFromContext(c)
	if !ok {
		return 0, false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return 0, false
	}
	roleStr, _ := roleVal.(string)
	if roleStr == roleStudent {
		return userID, true
	}
	return 0, false
}

// handleError maps domain sentinels to HTTP responses. Explicit
// sentinel-first via errors.Is BEFORE the generic MapDomainError
// fallback — same pattern as grade_handler / return_handler /
// assignments_handler.
func (h *ResubmitHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, assignUsecases.ErrAssignmentNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("assignment"))
		return
	case errors.Is(err, assignUsecases.ErrSubmissionNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("submission"))
		return
	case errors.Is(err, entities.ErrSubmissionOwnerOnly):
		c.JSON(http.StatusForbidden, response.Forbidden("you can only resubmit your own submission"))
		return
	case errors.Is(err, entities.ErrNotReturned):
		c.JSON(http.StatusConflict, response.ErrorResponse("NOT_RETURNED",
			"this submission is not in returned state"))
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
