package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	tasksPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/infrastructure/persistence"
)

// TaskReminderHandler serves the per-user reminder CRUD under
// /api/tasks/:id/reminders. All endpoints sit behind protectedGroup
// auth; per-user privacy comes from each use case taking
// ActorUserID = JWT subject (not URL parameter).
type TaskReminderHandler struct {
	setUC    *usecases.SetReminderUseCase
	listUC   *usecases.ListTaskRemindersUseCase
	deleteUC *usecases.DeleteReminderUseCase
}

// NewTaskReminderHandler wires the three use case dependencies.
// Panics on nil so misconfigured DI fails at construction.
func NewTaskReminderHandler(
	setUC *usecases.SetReminderUseCase,
	listUC *usecases.ListTaskRemindersUseCase,
	deleteUC *usecases.DeleteReminderUseCase,
) *TaskReminderHandler {
	if setUC == nil || listUC == nil || deleteUC == nil {
		panic("tasks: NewTaskReminderHandler requires all three use cases non-nil")
	}
	return &TaskReminderHandler{setUC: setUC, listUC: listUC, deleteUC: deleteUC}
}

// Create handles POST /api/tasks/:id/reminders. Body:
// {reminder_type, minutes_before}. Returns 201 + reminder DTO.
// Domain validation errors → 422 with typed codes.
func (h *TaskReminderHandler) Create(c *gin.Context) {
	userID, ok := actorID(c)
	if !ok {
		return
	}
	taskID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	var req dto.CreateTaskReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid body: " + err.Error()})
		return
	}
	reminder, err := h.setUC.Execute(c.Request.Context(), usecases.SetReminderInput{
		TaskID:        taskID,
		ActorUserID:   userID,
		ReminderType:  entities.ReminderType(req.ReminderType),
		MinutesBefore: req.MinutesBefore,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, projectReminder(reminder))
}

// List handles GET /api/tasks/:id/reminders. Returns the caller's
// reminders only — per-user privacy applied at use-case layer.
func (h *TaskReminderHandler) List(c *gin.Context) {
	userID, ok := actorID(c)
	if !ok {
		return
	}
	taskID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	reminders, err := h.listUC.Execute(c.Request.Context(), usecases.ListTaskRemindersInput{
		TaskID:      taskID,
		ActorUserID: userID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	out := make([]dto.TaskReminderResponse, 0, len(reminders))
	for _, r := range reminders {
		out = append(out, projectReminder(r))
	}
	c.JSON(http.StatusOK, out)
}

// Delete handles DELETE /api/tasks/:id/reminders/:reminderID.
// Three failure modes mapped: 404 not-found, 404 wrong task path,
// 403 wrong owner.
func (h *TaskReminderHandler) Delete(c *gin.Context) {
	userID, ok := actorID(c)
	if !ok {
		return
	}
	taskID, ok := pathInt64(c, "id")
	if !ok {
		return
	}
	reminderID, ok := pathInt64(c, "reminderID")
	if !ok {
		return
	}
	err := h.deleteUC.Execute(c.Request.Context(), usecases.DeleteReminderInput{
		ReminderID:  reminderID,
		TaskID:      taskID,
		ActorUserID: userID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// handleError maps domain + use-case sentinels to HTTP status codes.
func (h *TaskReminderHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entities.ErrInvalidTaskID),
		errors.Is(err, entities.ErrInvalidUserID),
		errors.Is(err, entities.ErrInvalidReminderType),
		errors.Is(err, entities.ErrInvalidMinutesBefore):
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: err.Error()})
	case errors.Is(err, tasksPersistence.ErrTaskReminderNotFound),
		errors.Is(err, usecases.ErrReminderNotFoundForTask):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "reminder not found"})
	case errors.Is(err, usecases.ErrReminderOwnerOnly):
		c.JSON(http.StatusForbidden, gin.H{errorKey: "not reminder owner"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
	}
}

// actorID reads user_id from context (production JWTMiddleware
// writes it). Mirror к task_handler.go:getUserID shape but as a
// package-level helper so the reminder handler does not duplicate
// it across methods.
func actorID(c *gin.Context) (int64, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "user not authenticated"})
		return 0, false
	}
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	}
	c.JSON(http.StatusUnauthorized, gin.H{errorKey: "user_id has unexpected type"})
	return 0, false
}

// pathInt64 parses an int64 URL path parameter с the supplied name.
// Emits 400 with a descriptive message on failure.
func pathInt64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid " + name})
		return 0, false
	}
	return v, true
}

// projectReminder converts a domain entity к the JSON DTO shape.
func projectReminder(r *entities.TaskReminder) dto.TaskReminderResponse {
	return dto.TaskReminderResponse{
		ID:            r.ID(),
		TaskID:        r.TaskID(),
		UserID:        r.UserID(),
		ReminderType:  string(r.ReminderType()),
		MinutesBefore: r.MinutesBefore(),
		IsSent:        r.IsSent(),
		SentAt:        r.SentAt(),
		CreatedAt:     r.CreatedAt(),
	}
}
