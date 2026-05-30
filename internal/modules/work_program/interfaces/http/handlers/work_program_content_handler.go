package handlers

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

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
	AddTopic(ctx context.Context, actorID int64, actorRole string, wpID int64, in usecases.TopicContentInput) (*entities.WorkProgram, error)
	UpdateTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64, in usecases.TopicContentInput) (*entities.WorkProgram, error)
	RemoveTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64) (*entities.WorkProgram, error)
}

// WorkProgramContentHandler exposes the manual collection-edit endpoints.
type WorkProgramContentHandler struct{}

// NewWorkProgramContentHandler wires the handler. The port is required —
// a nil dependency fails loudly at DI time rather than under load.
func NewWorkProgramContentHandler(content WorkProgramContentPort) *WorkProgramContentHandler {
	if content == nil {
		panic("work_program: NewWorkProgramContentHandler requires non-nil port")
	}
	return &WorkProgramContentHandler{}
}

// RegisterWorkProgramContentRoutes mounts the collection-edit endpoints.
// Routes land in the GREEN step.
func RegisterWorkProgramContentRoutes(rg *gin.RouterGroup, h *WorkProgramContentHandler) {
}
