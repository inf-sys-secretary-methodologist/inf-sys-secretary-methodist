// Package handlers exposes HTTP endpoints for the work_program
// (рабочая программа дисциплины / РПД) bounded context.
package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Narrow ports — handler depends on the methods it calls, not the
// concrete use-case structs. Keeps the handler test double-friendly
// and the DI seam explicit.

// CreateWorkProgramPort is the narrow port for CreateWorkProgramUseCase.
type CreateWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.CreateWorkProgramInput) (*entities.WorkProgram, error)
}

// GetWorkProgramPort is the narrow port for GetWorkProgramUseCase.
type GetWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.GetWorkProgramInput) (*entities.WorkProgram, error)
}

// ListWorkProgramsPort is the narrow port for ListWorkProgramsUseCase.
type ListWorkProgramsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.ListWorkProgramsInput) (wpUsecases.ListWorkProgramsResult, error)
}

// WorkProgramHandler exposes the read + create endpoints over HTTP
// (PR 4a). Transition endpoints (submit/approve/reject/discard) join in
// PR 4b — the ctor grows then.
type WorkProgramHandler struct {
	create CreateWorkProgramPort
	get    GetWorkProgramPort
	list   ListWorkProgramsPort
}

// NewWorkProgramHandler wires the handler. All ports are required —
// a nil dependency would surface as a nil-pointer panic deep under load
// instead of failing loudly at DI wiring time (mirror к extracurricular
// + curriculum failure-closed posture).
func NewWorkProgramHandler(
	create CreateWorkProgramPort,
	get GetWorkProgramPort,
	list ListWorkProgramsPort,
) *WorkProgramHandler {
	if create == nil || get == nil || list == nil {
		panic("work_program: NewWorkProgramHandler requires non-nil ports")
	}
	return &WorkProgramHandler{create: create, get: get, list: list}
}

// CreateWorkProgramRequest is the JSON body for POST /work-programs.
// binding tags per CLAUDE.md feedback (NOT `validate:`). The author is
// derived from the JWT subject server-side — never from the body.
type CreateWorkProgramRequest struct {
	DisciplineID       int64  `json:"discipline_id"        binding:"required"`
	SpecialtyCode      string `json:"specialty_code"       binding:"required"`
	ApplicableFromYear int    `json:"applicable_from_year" binding:"required"`
	Title              string `json:"title"                binding:"required"`
	Annotation         string `json:"annotation"`
}

// ===== Endpoints (PR 4a stubs — implemented in following GREEN commits) =====

// Create handles POST /api/v1/work-programs.
func (h *WorkProgramHandler) Create(c *gin.Context) { c.Status(http.StatusNotImplemented) }

// Get handles GET /api/v1/work-programs/:id.
func (h *WorkProgramHandler) Get(c *gin.Context) { c.Status(http.StatusNotImplemented) }

// List handles GET /api/v1/work-programs.
func (h *WorkProgramHandler) List(c *gin.Context) { c.Status(http.StatusNotImplemented) }

// RegisterWorkProgramRoutes mounts the read + create endpoints under
// /work-programs. Caller must apply auth middleware to the group before
// passing it in — every endpoint requires an authenticated context.
func RegisterWorkProgramRoutes(rg *gin.RouterGroup, h *WorkProgramHandler) {
	wp := rg.Group("/work-programs")
	wp.POST("", h.Create)
	wp.GET("", h.List)
	wp.GET("/:id", h.Get)
}
