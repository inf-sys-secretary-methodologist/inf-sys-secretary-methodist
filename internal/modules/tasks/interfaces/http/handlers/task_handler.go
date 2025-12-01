// Package handlers provides HTTP handlers for the tasks module.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
)

// TaskHandler handles HTTP requests for tasks.
type TaskHandler struct {
	taskUseCase *usecases.TaskUseCase
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(taskUseCase *usecases.TaskUseCase) *TaskHandler {
	return &TaskHandler{taskUseCase: taskUseCase}
}

// getUserID extracts user ID from context.
func (h *TaskHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 0, false
	}
	return userID.(int64), true
}

// getIDParam extracts ID parameter from URL.
func (h *TaskHandler) getIDParam(c *gin.Context, param string) (int64, bool) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, false
	}
	return id, true
}

// handleError handles use case errors.
func (h *TaskHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
	case errors.Is(err, usecases.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	case errors.Is(err, usecases.ErrCannotModifyTask):
		c.JSON(http.StatusConflict, gin.H{"error": "cannot modify task"})
	case errors.Is(err, usecases.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, usecases.ErrCommentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
	case errors.Is(err, usecases.ErrChecklistNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "checklist not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// Create creates a new task.
func (h *TaskHandler) Create(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskUseCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToTaskOutput(task))
}

// GetByID retrieves a task by ID.
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Update updates a task.
func (h *TaskHandler) Update(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskUseCase.Update(c.Request.Context(), userID, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Delete deletes a task.
func (h *TaskHandler) Delete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.taskUseCase.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// List lists tasks with filters.
func (h *TaskHandler) List(c *gin.Context) {
	var input dto.TaskFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Limit == 0 {
		input.Limit = 20
	}

	output, err := h.taskUseCase.List(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

// Assign assigns a task to a user.
func (h *TaskHandler) Assign(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.AssignTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskUseCase.Assign(c.Request.Context(), userID, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Unassign removes the assignee from a task.
func (h *TaskHandler) Unassign(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.Unassign(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// StartWork starts work on a task.
func (h *TaskHandler) StartWork(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.StartWork(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// SubmitForReview submits a task for review.
func (h *TaskHandler) SubmitForReview(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.SubmitForReview(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Complete completes a task.
func (h *TaskHandler) Complete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.Complete(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Cancel cancels a task.
func (h *TaskHandler) Cancel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.Cancel(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// Reopen reopens a task.
func (h *TaskHandler) Reopen(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.taskUseCase.Reopen(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskOutput(task))
}

// AddWatcher adds a watcher to a task.
func (h *TaskHandler) AddWatcher(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input struct {
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskUseCase.AddWatcher(c.Request.Context(), userID, id, input.UserID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "watcher added"})
}

// RemoveWatcher removes a watcher from a task.
func (h *TaskHandler) RemoveWatcher(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	taskID, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	watcherID, ok := h.getIDParam(c, "watcher_id")
	if !ok {
		return
	}

	if err := h.taskUseCase.RemoveWatcher(c.Request.Context(), userID, taskID, watcherID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetWatchers gets watchers of a task.
func (h *TaskHandler) GetWatchers(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	watchers, err := h.taskUseCase.GetWatchers(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var output []dto.UserOutput
	for _, w := range watchers {
		if w.User != nil {
			output = append(output, dto.UserOutput{
				ID:    w.User.ID,
				Name:  w.User.Name,
				Email: w.User.Email,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"watchers": output})
}

// AddComment adds a comment to a task.
func (h *TaskHandler) AddComment(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.AddCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.taskUseCase.AddComment(c.Request.Context(), userID, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToTaskCommentOutput(comment))
}

// UpdateComment updates a comment.
func (h *TaskHandler) UpdateComment(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	commentID, ok := h.getIDParam(c, "comment_id")
	if !ok {
		return
	}

	var input dto.UpdateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.taskUseCase.UpdateComment(c.Request.Context(), userID, commentID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskCommentOutput(comment))
}

// DeleteComment deletes a comment.
func (h *TaskHandler) DeleteComment(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	commentID, ok := h.getIDParam(c, "comment_id")
	if !ok {
		return
	}

	if err := h.taskUseCase.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetComments gets comments of a task.
func (h *TaskHandler) GetComments(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	comments, err := h.taskUseCase.GetComments(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var output []dto.TaskCommentOutput
	for _, comment := range comments {
		output = append(output, dto.ToTaskCommentOutput(comment))
	}

	c.JSON(http.StatusOK, gin.H{"comments": output})
}

// AddChecklist adds a checklist to a task.
func (h *TaskHandler) AddChecklist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.AddChecklistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checklist, err := h.taskUseCase.AddChecklist(c.Request.Context(), userID, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToTaskChecklistOutput(checklist))
}

// DeleteChecklist deletes a checklist.
func (h *TaskHandler) DeleteChecklist(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	checklistID, ok := h.getIDParam(c, "checklist_id")
	if !ok {
		return
	}

	if err := h.taskUseCase.DeleteChecklist(c.Request.Context(), userID, checklistID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetChecklists gets checklists of a task.
func (h *TaskHandler) GetChecklists(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	checklists, err := h.taskUseCase.GetChecklists(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var output []dto.TaskChecklistOutput
	for _, checklist := range checklists {
		output = append(output, dto.ToTaskChecklistOutput(checklist))
	}

	c.JSON(http.StatusOK, gin.H{"checklists": output})
}

// AddChecklistItem adds an item to a checklist.
func (h *TaskHandler) AddChecklistItem(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	checklistID, ok := h.getIDParam(c, "checklist_id")
	if !ok {
		return
	}

	var input dto.AddChecklistItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.taskUseCase.AddChecklistItem(c.Request.Context(), userID, checklistID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.TaskChecklistItemOutput{
		ID:          item.ID,
		ChecklistID: item.ChecklistID,
		Title:       item.Title,
		IsCompleted: item.IsCompleted,
		Position:    item.Position,
		CreatedAt:   item.CreatedAt,
	})
}

// DeleteChecklistItem deletes a checklist item.
func (h *TaskHandler) DeleteChecklistItem(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	itemID, ok := h.getIDParam(c, "item_id")
	if !ok {
		return
	}

	if err := h.taskUseCase.DeleteChecklistItem(c.Request.Context(), userID, itemID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetHistory gets the history of a task.
func (h *TaskHandler) GetHistory(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	history, err := h.taskUseCase.GetHistory(c.Request.Context(), id, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var output []dto.TaskHistoryOutput
	for _, h := range history {
		output = append(output, dto.ToTaskHistoryOutput(h))
	}

	c.JSON(http.StatusOK, gin.H{"history": output})
}
