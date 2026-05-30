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

// ===== Port =====

// WorkProgramContentPort is the narrow port for the manual collection-edit
// use case (slice 12). It exposes the goals / competences / topics methods
// the handler calls — assessments / references arrive in 12b-2.
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

// RegisterWorkProgramContentRoutes mounts the 9 manual collection-edit
// endpoints (goals / competences / topics) under /work-programs/:id.
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
}
