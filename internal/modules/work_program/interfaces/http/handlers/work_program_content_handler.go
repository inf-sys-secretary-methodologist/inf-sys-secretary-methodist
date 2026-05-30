package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ===== Request DTOs =====

// AddGoalRequest is the JSON body for add/update of a Goal. OrderIndex is
// optional (zero is a valid first position); Text carries the invariant
// and is required at the transport boundary — deeper rules (status gate,
// trimming) stay in the domain.
type AddGoalRequest struct {
	Text       string `json:"text"        binding:"required"`
	OrderIndex int    `json:"order_index"`
}

// CompetenceContentRequest is the JSON body for add/update of a Competence.
type CompetenceContentRequest struct {
	Code        string `json:"code"        binding:"required"`
	Type        string `json:"type"        binding:"required"`
	Description string `json:"description" binding:"required"`
}

// TopicContentRequest is the JSON body for add/update of a Topic. Only the
// identity-ish fields (kind, title) are required at the boundary; hours /
// week / outcomes / order are optional and validated by the domain.
type TopicContentRequest struct {
	Kind             string `json:"kind"              binding:"required"`
	Title            string `json:"title"             binding:"required"`
	Hours            int    `json:"hours"`
	WeekNumber       *int   `json:"week_number"`
	LearningOutcomes string `json:"learning_outcomes"`
	OrderIndex       int    `json:"order_index"`
}

func (r TopicContentRequest) toInput() wpUsecases.TopicContentInput {
	return wpUsecases.TopicContentInput{
		Kind:             r.Kind,
		Title:            r.Title,
		Hours:            r.Hours,
		WeekNumber:       r.WeekNumber,
		LearningOutcomes: r.LearningOutcomes,
		OrderIndex:       r.OrderIndex,
	}
}

// AssessmentContentRequest is the JSON body for add/update of an
// AssessmentCriterion (ФОС item). Type + Description are required at the
// boundary; MaxScore range + per-item rules stay in the domain.
type AssessmentContentRequest struct {
	Type             string   `json:"type"              binding:"required"`
	Description      string   `json:"description"       binding:"required"`
	MaxScore         int      `json:"max_score"`
	ExampleQuestions []string `json:"example_questions"`
}

func (r AssessmentContentRequest) toInput() wpUsecases.AssessmentContentInput {
	return wpUsecases.AssessmentContentInput{
		Type:             r.Type,
		Description:      r.Description,
		MaxScore:         r.MaxScore,
		ExampleQuestions: r.ExampleQuestions,
	}
}

// ReferenceContentRequest is the JSON body for add/update of a Reference.
type ReferenceContentRequest struct {
	Kind       string `json:"kind"        binding:"required"`
	Citation   string `json:"citation"    binding:"required"`
	Year       *int   `json:"year"`
	ISBN       string `json:"isbn"`
	URL        string `json:"url"`
	OrderIndex int    `json:"order_index"`
}

func (r ReferenceContentRequest) toInput() wpUsecases.ReferenceContentInput {
	return wpUsecases.ReferenceContentInput{
		Kind:       r.Kind,
		Citation:   r.Citation,
		Year:       r.Year,
		ISBN:       r.ISBN,
		URL:        r.URL,
		OrderIndex: r.OrderIndex,
	}
}

// ===== Port =====

// WorkProgramContentPort is the narrow port for the manual collection-edit
// use case (slice 12). It exposes all five collections —
// goals / competences / topics (12b-1) and assessments / references (12b-2).
type WorkProgramContentPort interface {
	AddGoal(ctx context.Context, actorID int64, actorRole string, wpID int64, text string, orderIndex int) (*entities.WorkProgram, error)
	UpdateGoal(ctx context.Context, actorID int64, actorRole string, wpID, goalID int64, text string, orderIndex int) (*entities.WorkProgram, error)
	RemoveGoal(ctx context.Context, actorID int64, actorRole string, wpID, goalID int64) (*entities.WorkProgram, error)
	AddCompetence(ctx context.Context, actorID int64, actorRole string, wpID int64, code, ctype, description string) (*entities.WorkProgram, error)
	UpdateCompetence(ctx context.Context, actorID int64, actorRole string, wpID, compID int64, code, ctype, description string) (*entities.WorkProgram, error)
	RemoveCompetence(ctx context.Context, actorID int64, actorRole string, wpID, compID int64) (*entities.WorkProgram, error)
	AddTopic(ctx context.Context, actorID int64, actorRole string, wpID int64, in wpUsecases.TopicContentInput) (*entities.WorkProgram, error)
	UpdateTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64, in wpUsecases.TopicContentInput) (*entities.WorkProgram, error)
	RemoveTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64) (*entities.WorkProgram, error)
	AddAssessment(ctx context.Context, actorID int64, actorRole string, wpID int64, in wpUsecases.AssessmentContentInput) (*entities.WorkProgram, error)
	UpdateAssessment(ctx context.Context, actorID int64, actorRole string, wpID, assessmentID int64, in wpUsecases.AssessmentContentInput) (*entities.WorkProgram, error)
	RemoveAssessment(ctx context.Context, actorID int64, actorRole string, wpID, assessmentID int64) (*entities.WorkProgram, error)
	AddReference(ctx context.Context, actorID int64, actorRole string, wpID int64, in wpUsecases.ReferenceContentInput) (*entities.WorkProgram, error)
	UpdateReference(ctx context.Context, actorID int64, actorRole string, wpID, referenceID int64, in wpUsecases.ReferenceContentInput) (*entities.WorkProgram, error)
	RemoveReference(ctx context.Context, actorID int64, actorRole string, wpID, referenceID int64) (*entities.WorkProgram, error)
}

// WorkProgramContentHandler exposes the manual collection-edit endpoints.
type WorkProgramContentHandler struct {
	content WorkProgramContentPort
}

// NewWorkProgramContentHandler wires the handler. The port is required —
// a nil dependency fails loudly at DI time rather than under load.
func NewWorkProgramContentHandler(content WorkProgramContentPort) *WorkProgramContentHandler {
	if content == nil {
		panic("work_program: NewWorkProgramContentHandler requires non-nil port")
	}
	return &WorkProgramContentHandler{content: content}
}

// respondContent maps the use-case outcome to HTTP. Collection edits are
// resource-scoped on :id, so a non-admin author-scope denial collapses to
// 404 (IDOR mitigation) — mirroring the transition endpoints.
func (h *WorkProgramContentHandler) respondContent(c *gin.Context, role string, wp *entities.WorkProgram, err error) {
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// ===== Goals =====

// AddGoal handles POST /api/v1/work-programs/:id/goals.
// @Summary Add a goal to a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int            true "Work program ID"
// @Param   body body AddGoalRequest true "Goal payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/goals [post]
func (h *WorkProgramContentHandler) AddGoal(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	var body AddGoalRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.AddGoal(c.Request.Context(), actorID, role, wpID, body.Text, body.OrderIndex)
	h.respondContent(c, role, wp, err)
}

// UpdateGoal handles PUT /api/v1/work-programs/:id/goals/:childId.
// @Summary Update a goal of a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id      path int            true "Work program ID"
// @Param   childId path int            true "Goal ID"
// @Param   body    body AddGoalRequest true "Goal payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/goals/{childId} [put]
func (h *WorkProgramContentHandler) UpdateGoal(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	goalID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid goal id"))
		return
	}
	var body AddGoalRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.UpdateGoal(c.Request.Context(), actorID, role, wpID, goalID, body.Text, body.OrderIndex)
	h.respondContent(c, role, wp, err)
}

// RemoveGoal handles DELETE /api/v1/work-programs/:id/goals/:childId.
// @Summary Remove a goal from a work program (РПД)
// @Tags    work-programs
// @Produce json
// @Param   id      path int true "Work program ID"
// @Param   childId path int true "Goal ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/goals/{childId} [delete]
func (h *WorkProgramContentHandler) RemoveGoal(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	goalID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid goal id"))
		return
	}
	wp, err := h.content.RemoveGoal(c.Request.Context(), actorID, role, wpID, goalID)
	h.respondContent(c, role, wp, err)
}

// ===== Competences =====

// AddCompetence handles POST /api/v1/work-programs/:id/competences.
// @Summary Add a competence to a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                       true "Work program ID"
// @Param   body body CompetenceContentRequest true "Competence payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/competences [post]
func (h *WorkProgramContentHandler) AddCompetence(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	var body CompetenceContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.AddCompetence(c.Request.Context(), actorID, role, wpID, body.Code, body.Type, body.Description)
	h.respondContent(c, role, wp, err)
}

// UpdateCompetence handles PUT /api/v1/work-programs/:id/competences/:childId.
// @Summary Update a competence of a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id      path int                       true "Work program ID"
// @Param   childId path int                       true "Competence ID"
// @Param   body    body CompetenceContentRequest true "Competence payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/competences/{childId} [put]
func (h *WorkProgramContentHandler) UpdateCompetence(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	compID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid competence id"))
		return
	}
	var body CompetenceContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.UpdateCompetence(c.Request.Context(), actorID, role, wpID, compID, body.Code, body.Type, body.Description)
	h.respondContent(c, role, wp, err)
}

// RemoveCompetence handles DELETE /api/v1/work-programs/:id/competences/:childId.
// @Summary Remove a competence from a work program (РПД)
// @Tags    work-programs
// @Produce json
// @Param   id      path int true "Work program ID"
// @Param   childId path int true "Competence ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/competences/{childId} [delete]
func (h *WorkProgramContentHandler) RemoveCompetence(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	compID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid competence id"))
		return
	}
	wp, err := h.content.RemoveCompetence(c.Request.Context(), actorID, role, wpID, compID)
	h.respondContent(c, role, wp, err)
}

// ===== Topics =====

// AddTopic handles POST /api/v1/work-programs/:id/topics.
// @Summary Add a topic to a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                 true "Work program ID"
// @Param   body body TopicContentRequest true "Topic payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/topics [post]
func (h *WorkProgramContentHandler) AddTopic(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	var body TopicContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.AddTopic(c.Request.Context(), actorID, role, wpID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// UpdateTopic handles PUT /api/v1/work-programs/:id/topics/:childId.
// @Summary Update a topic of a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id      path int                 true "Work program ID"
// @Param   childId path int                 true "Topic ID"
// @Param   body    body TopicContentRequest true "Topic payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/topics/{childId} [put]
func (h *WorkProgramContentHandler) UpdateTopic(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	topicID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid topic id"))
		return
	}
	var body TopicContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.UpdateTopic(c.Request.Context(), actorID, role, wpID, topicID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// RemoveTopic handles DELETE /api/v1/work-programs/:id/topics/:childId.
// @Summary Remove a topic from a work program (РПД)
// @Tags    work-programs
// @Produce json
// @Param   id      path int true "Work program ID"
// @Param   childId path int true "Topic ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/topics/{childId} [delete]
func (h *WorkProgramContentHandler) RemoveTopic(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	topicID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid topic id"))
		return
	}
	wp, err := h.content.RemoveTopic(c.Request.Context(), actorID, role, wpID, topicID)
	h.respondContent(c, role, wp, err)
}

// ===== Assessments (ФОС) =====

// AddAssessment handles POST /api/v1/work-programs/:id/assessments.
// @Summary Add an assessment criterion (ФОС) to a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                       true "Work program ID"
// @Param   body body AssessmentContentRequest true "Assessment payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/assessments [post]
func (h *WorkProgramContentHandler) AddAssessment(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	var body AssessmentContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.AddAssessment(c.Request.Context(), actorID, role, wpID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// UpdateAssessment handles PUT /api/v1/work-programs/:id/assessments/:childId.
// @Summary Update an assessment criterion (ФОС) of a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id      path int                       true "Work program ID"
// @Param   childId path int                       true "Assessment ID"
// @Param   body    body AssessmentContentRequest true "Assessment payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/assessments/{childId} [put]
func (h *WorkProgramContentHandler) UpdateAssessment(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	assessmentID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assessment id"))
		return
	}
	var body AssessmentContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.UpdateAssessment(c.Request.Context(), actorID, role, wpID, assessmentID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// RemoveAssessment handles DELETE /api/v1/work-programs/:id/assessments/:childId.
// @Summary Remove an assessment criterion (ФОС) from a work program (РПД)
// @Tags    work-programs
// @Produce json
// @Param   id      path int true "Work program ID"
// @Param   childId path int true "Assessment ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/assessments/{childId} [delete]
func (h *WorkProgramContentHandler) RemoveAssessment(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	assessmentID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid assessment id"))
		return
	}
	wp, err := h.content.RemoveAssessment(c.Request.Context(), actorID, role, wpID, assessmentID)
	h.respondContent(c, role, wp, err)
}

// ===== References =====

// AddReference handles POST /api/v1/work-programs/:id/references.
// @Summary Add a reference to a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                     true "Work program ID"
// @Param   body body ReferenceContentRequest true "Reference payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/references [post]
func (h *WorkProgramContentHandler) AddReference(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	var body ReferenceContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.AddReference(c.Request.Context(), actorID, role, wpID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// UpdateReference handles PUT /api/v1/work-programs/:id/references/:childId.
// @Summary Update a reference of a work program (РПД)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id      path int                     true "Work program ID"
// @Param   childId path int                     true "Reference ID"
// @Param   body    body ReferenceContentRequest true "Reference payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/references/{childId} [put]
func (h *WorkProgramContentHandler) UpdateReference(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	referenceID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid reference id"))
		return
	}
	var body ReferenceContentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.content.UpdateReference(c.Request.Context(), actorID, role, wpID, referenceID, body.toInput())
	h.respondContent(c, role, wp, err)
}

// RemoveReference handles DELETE /api/v1/work-programs/:id/references/:childId.
// @Summary Remove a reference from a work program (РПД)
// @Tags    work-programs
// @Produce json
// @Param   id      path int true "Work program ID"
// @Param   childId path int true "Reference ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/references/{childId} [delete]
func (h *WorkProgramContentHandler) RemoveReference(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	wpID, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	referenceID, ok := parsePositiveID(c.Param("childId"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid reference id"))
		return
	}
	wp, err := h.content.RemoveReference(c.Request.Context(), actorID, role, wpID, referenceID)
	h.respondContent(c, role, wp, err)
}

// RegisterWorkProgramContentRoutes mounts the 15 manual collection-edit
// endpoints (goals / competences / topics / assessments / references) under
// /work-programs/:id.
func RegisterWorkProgramContentRoutes(rg *gin.RouterGroup, h *WorkProgramContentHandler) {
	wp := rg.Group("/work-programs")

	wp.POST("/:id/goals", h.AddGoal)
	wp.PUT("/:id/goals/:childId", h.UpdateGoal)
	wp.DELETE("/:id/goals/:childId", h.RemoveGoal)

	wp.POST("/:id/competences", h.AddCompetence)
	wp.PUT("/:id/competences/:childId", h.UpdateCompetence)
	wp.DELETE("/:id/competences/:childId", h.RemoveCompetence)

	wp.POST("/:id/topics", h.AddTopic)
	wp.PUT("/:id/topics/:childId", h.UpdateTopic)
	wp.DELETE("/:id/topics/:childId", h.RemoveTopic)

	wp.POST("/:id/assessments", h.AddAssessment)
	wp.PUT("/:id/assessments/:childId", h.UpdateAssessment)
	wp.DELETE("/:id/assessments/:childId", h.RemoveAssessment)

	wp.POST("/:id/references", h.AddReference)
	wp.PUT("/:id/references/:childId", h.UpdateReference)
	wp.DELETE("/:id/references/:childId", h.RemoveReference)
}
