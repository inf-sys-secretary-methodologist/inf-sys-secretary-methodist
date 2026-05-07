package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ListMyAssignmentsUseCasePort is the narrow port through which the
// handler invokes the student's "my assignments" list use case.
// Defining the ports here keeps handler tests free of repo / view
// fixtures — only the use-case Execute signature is stubbed. Same
// pattern as the grading-side ports above.
//
// GetMyAssignmentDetailUseCasePort below follows the same shape for
// the single-row read.
type (
	// ListMyAssignmentsUseCasePort declares the Execute signature used
	// by GET /api/assignments/my.
	ListMyAssignmentsUseCasePort interface {
		Execute(ctx context.Context, in assignUsecases.ListMyAssignmentsInput) ([]views.StudentAssignmentView, error)
	}
	// GetMyAssignmentDetailUseCasePort declares the Execute signature
	// used by GET /api/assignments/:id/my.
	GetMyAssignmentDetailUseCasePort interface {
		Execute(ctx context.Context, in assignUsecases.GetMyAssignmentDetailInput) (*views.StudentAssignmentView, error)
	}
)

// MyAssignmentsHandler serves the student-facing read surface for the
// assignments module. Routes mount under studentAssignmentsGroup
// behind RequireRole("student"); the handler additionally whitelists
// "student" via studentIDFromContext as defense in depth.
type MyAssignmentsHandler struct {
	listUC   ListMyAssignmentsUseCasePort
	detailUC GetMyAssignmentDetailUseCasePort
}

// NewMyAssignmentsHandler wires the handler. Failure-closed: any nil
// dependency panics during construction so DI mistakes surface in
// smoke tests, not in production traffic.
func NewMyAssignmentsHandler(listUC ListMyAssignmentsUseCasePort, detailUC GetMyAssignmentDetailUseCasePort) *MyAssignmentsHandler {
	if listUC == nil || detailUC == nil {
		panic("assignments: NewMyAssignmentsHandler requires non-nil use cases")
	}
	return &MyAssignmentsHandler{listUC: listUC, detailUC: detailUC}
}

// StudentAssignmentDTO is the JSON shape returned to the student
// frontend. Mirrors StudentAssignmentView 1:1 with snake_case JSON
// tags. Kept separate from the view so domain / view changes do not
// silently break the HTTP contract.
type StudentAssignmentDTO struct {
	AssignmentID        int64      `json:"assignment_id"`
	Title               string     `json:"title"`
	Description         string     `json:"description,omitempty"`
	Subject             string     `json:"subject"`
	GroupName           string     `json:"group_name"`
	MaxScore            int        `json:"max_score"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	AssignmentCreatedAt time.Time  `json:"assignment_created_at"`
	AssignmentUpdatedAt time.Time  `json:"assignment_updated_at"`

	SubmissionID        int64      `json:"submission_id"`
	StudentID           int64      `json:"student_id"`
	GradeValue          *int       `json:"grade_value,omitempty"`
	Feedback            string     `json:"feedback,omitempty"`
	GradedBy            *int64     `json:"graded_by,omitempty"`
	GradedAt            *time.Time `json:"graded_at,omitempty"`
	ReturnReason        string     `json:"return_reason,omitempty"`
	ReturnedBy          *int64     `json:"returned_by,omitempty"`
	ReturnedAt          *time.Time `json:"returned_at,omitempty"`
	Status              string     `json:"status"`
	SubmissionCreatedAt time.Time  `json:"submission_created_at"`
	SubmissionUpdatedAt time.Time  `json:"submission_updated_at"`
}

// List handles GET /api/assignments/my.
//
// @Summary List the student's own assignments
// @Tags assignments
// @Produce json
// @Param status query string false "Filter by status: pending, graded, returned"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/my [get]
func (h *MyAssignmentsHandler) List(c *gin.Context) {
	studentID, ok := studentIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	var statusFilter *entities.SubmissionStatus
	if raw := c.Query("status"); raw != "" {
		s := entities.SubmissionStatus(raw)
		if !s.IsValid() {
			c.JSON(http.StatusBadRequest, response.BadRequest("invalid status: "+raw))
			return
		}
		statusFilter = &s
	}

	out, err := h.listUC.Execute(c.Request.Context(), assignUsecases.ListMyAssignmentsInput{
		StudentID: studentID,
		Status:    statusFilter,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	dtos := mapStudentAssignmentViews(out)
	c.JSON(http.StatusOK, response.Success(gin.H{
		"items": dtos,
		"total": len(dtos),
	}))
}

// Detail handles GET /api/assignments/:id/my.
//
// @Summary Get the student's submission view for one assignment
// @Tags assignments
// @Produce json
// @Param id path int true "Assignment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id}/my [get]
func (h *MyAssignmentsHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}
	studentID, ok := studentIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	view, err := h.detailUC.Execute(c.Request.Context(), assignUsecases.GetMyAssignmentDetailInput{
		AssignmentID: id,
		StudentID:    studentID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapStudentAssignmentView(view)))
}

func mapStudentAssignmentView(v *views.StudentAssignmentView) StudentAssignmentDTO {
	return StudentAssignmentDTO{
		AssignmentID:        v.AssignmentID,
		Title:               v.Title,
		Description:         v.Description,
		Subject:             v.Subject,
		GroupName:           v.GroupName,
		MaxScore:            v.MaxScore,
		DueDate:             v.DueDate,
		AssignmentCreatedAt: v.AssignmentCreatedAt,
		AssignmentUpdatedAt: v.AssignmentUpdatedAt,
		SubmissionID:        v.SubmissionID,
		StudentID:           v.StudentID,
		GradeValue:          v.GradeValue,
		Feedback:            v.Feedback,
		GradedBy:            v.GradedBy,
		GradedAt:            v.GradedAt,
		ReturnReason:        v.ReturnReason,
		ReturnedBy:          v.ReturnedBy,
		ReturnedAt:          v.ReturnedAt,
		Status:              string(v.Status),
		SubmissionCreatedAt: v.SubmissionCreatedAt,
		SubmissionUpdatedAt: v.SubmissionUpdatedAt,
	}
}

func mapStudentAssignmentViews(in []views.StudentAssignmentView) []StudentAssignmentDTO {
	out := make([]StudentAssignmentDTO, 0, len(in))
	for i := range in {
		out = append(out, mapStudentAssignmentView(&in[i]))
	}
	return out
}

// handleError mirrors AssignmentsHandler.handleError but adds the
// student-side ownership sentinel. Sentinel-first via errors.Is BEFORE
// the generic MapDomainError fallback.
func (h *MyAssignmentsHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrAssignmentNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("assignment"))
		return
	case errors.Is(err, repositories.ErrSubmissionNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("submission"))
		return
	case errors.Is(err, entities.ErrSubmissionOwnerOnly):
		c.JSON(http.StatusForbidden, response.Forbidden("you can only view your own submission"))
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
