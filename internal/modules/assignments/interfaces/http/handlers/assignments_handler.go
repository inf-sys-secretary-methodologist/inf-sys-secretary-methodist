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

// ListAssignments handles GET /api/assignments.
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
	scope, ok := callerScopeFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	out, err := h.listUC.Execute(c.Request.Context(), assignUsecases.ListAssignmentsInput{
		Caller:    scope,
		Subject:   c.Query("subject"),
		GroupName: c.Query("group_name"),
		Limit:     parseQueryInt(c, "page_size"),
		Offset:    parseQueryInt(c, "offset"),
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"items": mapAssignments(out.Items),
		"total": out.Total,
	}))
}

// GetAssignment handles GET /api/assignments/:id.
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}
	scope, ok := callerScopeFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}

	a, err := h.getUC.Execute(c.Request.Context(), assignUsecases.GetAssignmentInput{
		Caller:       scope,
		AssignmentID: id,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(mapAssignment(a)))
}

// ListSubmissions handles GET /api/assignments/:id/submissions.
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assignment id"))
		return
	}
	scope, ok := callerScopeFromContext(c)
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

	items, err := h.listSubsUC.Execute(c.Request.Context(), assignUsecases.ListSubmissionsInput{
		Caller:       scope,
		AssignmentID: id,
		Status:       statusFilter,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"items": mapSubmissionViews(items),
	}))
}

// Recognised role values, mirrored from auth/domain/permission.go but
// duplicated here to avoid an HTTP layer importing the auth domain.
// Updating this list when a new role is introduced is a deliberate
// step — the failure-closed default is "unknown role → no access."
const (
	roleTeacher           = "teacher"
	roleMethodist         = "methodist"
	roleAcademicSecretary = "academic_secretary"
	roleSystemAdmin       = "system_admin"
	roleStudent           = "student"
)

// callerScopeFromContext converts the gin auth context (user_id +
// role) into a use-case CallerScope. Failure-closed twice over:
//   - missing user_id or role → ok=false → 401.
//   - role not on the explicit whitelist (teacher / methodist /
//     academic_secretary / system_admin) → ok=false → 401, even
//     when the upstream middleware has already validated the JWT.
//
// Defence-in-depth: a future engineer who removes RequireNonStudent
// or adds a new role to RequireRole must NOT silently get
// "unrestricted" data access via this handler. The whitelist is the
// last gate; unknown roles fall through to the same 401 path as
// missing context.
func callerScopeFromContext(c *gin.Context) (assignUsecases.CallerScope, bool) {
	userID, ok := teacherIDFromContext(c)
	if !ok {
		return assignUsecases.CallerScope{}, false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return assignUsecases.CallerScope{}, false
	}
	roleStr, _ := roleVal.(string)
	switch roleStr {
	case roleMethodist, roleAcademicSecretary, roleSystemAdmin:
		return assignUsecases.CallerScope{UserID: userID, Unrestricted: true}, true
	case roleTeacher:
		return assignUsecases.CallerScope{UserID: userID, Unrestricted: false}, true
	default:
		return assignUsecases.CallerScope{}, false
	}
}

func parseQueryInt(c *gin.Context, key string) int {
	raw := c.Query(key)
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return v
}

func mapAssignment(a *entities.Assignment) AssignmentDTO {
	return AssignmentDTO{
		ID:          a.ID,
		Title:       a.Title(),
		Description: a.Description(),
		TeacherID:   a.TeacherID(),
		GroupName:   a.GroupName(),
		Subject:     a.Subject(),
		MaxScore:    a.MaxScore(),
		DueDate:     a.DueDate(),
		CreatedAt:   a.CreatedAt(),
		UpdatedAt:   a.UpdatedAt(),
	}
}

func mapAssignments(in []*entities.Assignment) []AssignmentDTO {
	out := make([]AssignmentDTO, 0, len(in))
	for _, a := range in {
		out = append(out, mapAssignment(a))
	}
	return out
}

func mapSubmissionViews(in []views.SubmissionView) []SubmissionViewDTO {
	out := make([]SubmissionViewDTO, 0, len(in))
	for _, v := range in {
		out = append(out, SubmissionViewDTO{
			ID:           v.ID,
			AssignmentID: v.AssignmentID,
			StudentID:    v.StudentID,
			StudentName:  v.StudentName,
			GradeValue:   v.GradeValue,
			Feedback:     v.Feedback,
			GradedBy:     v.GradedBy,
			GradedAt:     v.GradedAt,
			Status:       string(v.Status),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
		})
	}
	return out
}

// handleError centralises domain → HTTP mapping for the read handlers.
// Same explicit-sentinel-first pattern as grade_handler so module
// errors get their actionable status (403, 404) instead of falling
// through to the generic mapper's 500.
func (h *AssignmentsHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrAssignmentNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("assignment"))
		return
	case errors.Is(err, entities.ErrAssignmentScopeForbidden):
		c.JSON(http.StatusForbidden, response.Forbidden("you cannot access this assignment"))
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
