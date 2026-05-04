// Package handlers contains HTTP handlers for the assignments module.
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

// SaveGradeUseCasePort is the narrow port through which the handler
// invokes the use case. Defining it here (rather than importing the
// concrete *SaveGradeUseCase) keeps handler tests free of fake
// repositories and audit loggers — only the use-case behaviour is
// stubbed.
type SaveGradeUseCasePort interface {
	Execute(ctx context.Context, teacherID int64, in assignUsecases.SaveGradeInput) error
}

// GradeHandler handles HTTP requests for the assignments grading flow.
type GradeHandler struct {
	usecase SaveGradeUseCasePort
}

// NewGradeHandler wires the handler. The use case is required (non-nil):
// a nil use case would let requests reach a panic deeper in the call
// stack instead of failing during DI wiring. This matches the
// failure-closed posture established for analytics in v0.108.3.
func NewGradeHandler(usecase SaveGradeUseCasePort) *GradeHandler {
	if usecase == nil {
		panic("assignments: NewGradeHandler requires non-nil usecase")
	}
	return &GradeHandler{usecase: usecase}
}

// saveGradeRequest is the JSON body schema for POST /api/assignments/:id/grades.
type saveGradeRequest struct {
	StudentID int64  `json:"student_id"`
	Value     int    `json:"value"`
	Feedback  string `json:"feedback"`
}

// SaveGrade records a teacher's grade on a student's submission.
// @Summary Save a grade for a student's submission
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path int true "Assignment ID"
// @Param body body saveGradeRequest true "Grade payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id}/grades [post]
func (h *GradeHandler) SaveGrade(c *gin.Context) {
	assignmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}

	var body saveGradeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	if body.StudentID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("student_id is required and must be positive"))
		return
	}

	teacherID, ok := teacherIDFromContext(c)
	if !ok {
		// Auth middleware misconfiguration. We refuse to fall back to a
		// silent admin/system identity — failing here surfaces routing
		// bugs immediately instead of leaking a grade under an unknown
		// actor.
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	in := assignUsecases.SaveGradeInput{
		AssignmentID: assignmentID,
		StudentID:    body.StudentID,
		Value:        body.Value,
		Feedback:     body.Feedback,
	}
	if err := h.usecase.Execute(c.Request.Context(), teacherID, in); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"assignment_id": assignmentID,
		"student_id":    body.StudentID,
		"value":         body.Value,
	}))
}

// teacherIDFromContext extracts the authenticated user id from the gin
// context. Handles both int64 and float64 (gin's default JWT claim
// numeric type) so that swapping the auth middleware does not silently
// break the handler.
func teacherIDFromContext(c *gin.Context) (int64, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case float64:
		return int64(id), true
	default:
		return 0, false
	}
}

// handleError maps domain errors to HTTP responses. Every domain
// sentinel is matched explicitly via errors.Is BEFORE the generic
// MapDomainError fallback — otherwise the generic mapper falls through
// to 500 and clients lose the actionable distinction between e.g.
// "wrong teacher" (403) and "internal error" (500). Same pattern as
// analytics_handler in v0.108.3.
func (h *GradeHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrAssignmentNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("assignment"))
		return
	case errors.Is(err, entities.ErrAssignmentScopeForbidden):
		c.JSON(http.StatusForbidden, response.Forbidden("only the assignment author can grade"))
		return
	case errors.Is(err, entities.ErrInvalidScore),
		errors.Is(err, entities.ErrInvalidAssignment):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	case errors.Is(err, entities.ErrAlreadyGraded):
		c.JSON(http.StatusConflict,
			response.ErrorResponse("ALREADY_GRADED", "this submission is already graded"))
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
