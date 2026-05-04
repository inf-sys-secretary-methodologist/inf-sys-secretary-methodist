// Package handlers contains HTTP handlers for the assignments module.
package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
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

// SaveGrade records a teacher's grade on a student's submission.
// @Summary Save a grade for a student's submission
// @Tags Assignments
// @Router /api/assignments/{id}/grades [post]
func (h *GradeHandler) SaveGrade(c *gin.Context) {
	// stub for RED — always returns 200 empty Success
	c.JSON(http.StatusOK, response.Success(gin.H{}))
}
