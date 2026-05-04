package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ListAssignmentsUseCasePort, GetAssignmentUseCasePort and
// ListSubmissionsUseCasePort decouple the handler from the concrete
// use-case structs so handler tests stub only the relevant Execute
// signature instead of the full repository graph.
type (
	ListAssignmentsUseCasePort interface {
		Execute(ctx context.Context, in assignUsecases.ListAssignmentsInput) (assignUsecases.ListAssignmentsOutput, error)
	}
	GetAssignmentUseCasePort interface {
		Execute(ctx context.Context, in assignUsecases.GetAssignmentInput) (*entities.Assignment, error)
	}
	ListSubmissionsUseCasePort interface {
		Execute(ctx context.Context, in assignUsecases.ListSubmissionsInput) ([]views.SubmissionView, error)
	}
)

// AssignmentsHandler serves the read-side HTTP surface for the
// assignments module.
type AssignmentsHandler struct {
	listUC     ListAssignmentsUseCasePort
	getUC      GetAssignmentUseCasePort
	listSubsUC ListSubmissionsUseCasePort
}

// NewAssignmentsHandler wires the handler. All three use cases are
// required (non-nil) — failure-closed posture: a nil dependency would
// surface as a request-time panic deeper in the stack instead of
// failing during DI wiring, where it can be caught by smoke tests.
func NewAssignmentsHandler(
	listUC ListAssignmentsUseCasePort,
	getUC GetAssignmentUseCasePort,
	listSubsUC ListSubmissionsUseCasePort,
) *AssignmentsHandler {
	if listUC == nil || getUC == nil || listSubsUC == nil {
		panic("assignments: NewAssignmentsHandler requires non-nil use cases")
	}
	return &AssignmentsHandler{
		listUC:     listUC,
		getUC:      getUC,
		listSubsUC: listSubsUC,
	}
}

// AssignmentDTO is the JSON shape returned to the frontend. Kept
// separate from the domain entity so the entity can change its
// internal layout without breaking the HTTP contract.
type AssignmentDTO struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	TeacherID   int64      `json:"teacher_id"`
	GroupName   string     `json:"group_name"`
	Subject     string     `json:"subject"`
	MaxScore    int        `json:"max_score"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SubmissionViewDTO is the JSON shape returned for a submission row in
// the grading list.
type SubmissionViewDTO struct {
	ID           int64      `json:"id"`
	AssignmentID int64      `json:"assignment_id"`
	StudentID    int64      `json:"student_id"`
	StudentName  string     `json:"student_name"`
	GradeValue   *int       `json:"grade_value,omitempty"`
	Feedback     string     `json:"feedback,omitempty"`
	GradedBy     *int64     `json:"graded_by,omitempty"`
	GradedAt     *time.Time `json:"graded_at,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ListAssignments handles GET /api/assignments. Stub during the RED
// stage of cycle 5 — replaced with the real implementation in the
// GREEN commit once the failing tests are in place.
//
// @Summary List assignments visible to the caller
// @Tags assignments
// @Produce json
// @Param subject query string false "Subject filter"
// @Param group_name query string false "Group filter"
// @Param page_size query int false "Page size (1..100, default 50)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments [get]
func (h *AssignmentsHandler) ListAssignments(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.ErrorResponse("NOT_IMPLEMENTED", "ListAssignments not implemented"))
}

// GetAssignment handles GET /api/assignments/:id. Stub during RED.
//
// @Summary Get a single assignment by id
// @Tags assignments
// @Produce json
// @Param id path int true "Assignment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id} [get]
func (h *AssignmentsHandler) GetAssignment(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.ErrorResponse("NOT_IMPLEMENTED", "GetAssignment not implemented"))
}

// ListSubmissions handles GET /api/assignments/:id/submissions. Stub
// during RED.
//
// @Summary List submissions for an assignment
// @Tags assignments
// @Produce json
// @Param id path int true "Assignment ID"
// @Param status query string false "Filter by status: pending, graded, returned"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/assignments/{id}/submissions [get]
func (h *AssignmentsHandler) ListSubmissions(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, response.ErrorResponse("NOT_IMPLEMENTED", "ListSubmissions not implemented"))
}

// All shared helpers (callerScopeFromContext, mapAssignment, etc.) and
// the centralised handleError live in the GREEN commit alongside the
// real handler bodies — they would be dead code in the RED commit.
