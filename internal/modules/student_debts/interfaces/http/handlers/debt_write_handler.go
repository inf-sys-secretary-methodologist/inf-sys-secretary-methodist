package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// ScheduleResitPort schedules a resit on a debt (EDIT_ROLES only).
type ScheduleResitPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in sdUsecases.ScheduleResitInput) (*entities.StudentDebt, error)
}

// RecordResitResultPort records the outcome of a scheduled resit.
type RecordResitResultPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in sdUsecases.RecordResitResultInput) (*entities.StudentDebt, error)
}

// ScheduleResitRequest is the POST /:id/resit body.
type ScheduleResitRequest struct {
	ScheduledDate string `json:"scheduled_date" binding:"required"`
	Examiner      string `json:"examiner" binding:"required"`
}

// RecordResitResultRequest is the POST /:id/attempts/:n/result body.
type RecordResitResultRequest struct {
	Result string `json:"result" binding:"required"`
	Grade  *int   `json:"grade"`
}

// StudentDebtWriteHandler serves the debt write endpoints (resit lifecycle).
type StudentDebtWriteHandler struct {
	schedule ScheduleResitPort
	record   RecordResitResultPort
}

// NewStudentDebtWriteHandler wires the handler. All ports are required.
func NewStudentDebtWriteHandler(schedule ScheduleResitPort, record RecordResitResultPort) *StudentDebtWriteHandler {
	if schedule == nil || record == nil {
		panic("student_debts: NewStudentDebtWriteHandler requires non-nil ports")
	}
	return &StudentDebtWriteHandler{schedule: schedule, record: record}
}

// RegisterStudentDebtWriteRoutes mounts the write endpoints under
// /student-debts. The caller applies authentication + the EDIT_ROLES gate
// is enforced inside the use cases (denial → 403, pre-existence).
func RegisterStudentDebtWriteRoutes(rg *gin.RouterGroup, h *StudentDebtWriteHandler) {
	g := rg.Group("/student-debts")
	g.POST("/:id/resit", h.ScheduleResit)
	g.POST("/:id/attempts/:n/result", h.RecordResult)
}

// ScheduleResit handles POST /student-debts/:id/resit.
func (h *StudentDebtWriteHandler) ScheduleResit(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid student debt id"))
		return
	}
	var req ScheduleResitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("scheduled_date and examiner are required"))
		return
	}
	scheduledDate, err := time.Parse(time.RFC3339, req.ScheduledDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("scheduled_date must be an RFC3339 timestamp"))
		return
	}

	debt, err := h.schedule.Execute(c.Request.Context(), actorID, role, sdUsecases.ScheduleResitInput{
		DebtID:        id,
		ScheduledDate: scheduledDate,
		Examiner:      req.Examiner,
	})
	if err != nil {
		// Write denial is role-based and pre-existence — a true 403, no IDOR.
		mapDebtError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapDebt(debt)))
}

// RecordResult handles POST /student-debts/:id/attempts/:n/result.
func (h *StudentDebtWriteHandler) RecordResult(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid student debt id"))
		return
	}
	attemptNo, ok := parsePositiveInt(c.Param("n"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid attempt number"))
		return
	}
	var req RecordResitResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("result is required"))
		return
	}

	debt, err := h.record.Execute(c.Request.Context(), actorID, role, sdUsecases.RecordResitResultInput{
		DebtID:    id,
		AttemptNo: attemptNo,
		Result:    entities.ResitResult(req.Result),
		Grade:     req.Grade,
	})
	if err != nil {
		mapDebtError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapDebt(debt)))
}

// parsePositiveInt parses a strictly-positive int path parameter (the
// attempt number).
func parsePositiveInt(raw string) (int, bool) {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
