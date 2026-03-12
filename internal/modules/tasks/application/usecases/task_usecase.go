// Package usecases provides application use cases for the tasks module.
package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Use case errors.
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrCannotModifyTask  = errors.New("cannot modify task")
	ErrInvalidInput      = errors.New("invalid input")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrChecklistNotFound = errors.New("checklist not found")
)

// TaskUseCase provides task management operations.
type TaskUseCase struct {
	taskRepo            repositories.TaskRepository
	projectRepo         repositories.ProjectRepository
	auditLogger         *logging.AuditLogger
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewTaskUseCase creates a new TaskUseCase.
func NewTaskUseCase(
	taskRepo repositories.TaskRepository,
	projectRepo repositories.ProjectRepository,
	auditLogger *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *TaskUseCase {
	return &TaskUseCase{
		taskRepo:            taskRepo,
		projectRepo:         projectRepo,
		auditLogger:         auditLogger,
		notificationUseCase: notificationUseCase,
	}
}

// Create creates a new task.
func (uc *TaskUseCase) Create(ctx context.Context, userID int64, input dto.CreateTaskInput) (*entities.Task, error) {
	task := entities.NewTask(input.Title, userID)
	task.Description = input.Description
	task.ProjectID = input.ProjectID
	task.DocumentID = input.DocumentID
	task.DueDate = input.DueDate
	task.StartDate = input.StartDate
	task.EstimatedHours = input.EstimatedHours
	task.Tags = input.Tags

	if input.Priority != nil {
		priority := domain.TaskPriority(*input.Priority)
		if !priority.IsValid() {
			return nil, fmt.Errorf("%w: invalid priority", ErrInvalidInput)
		}
		task.Priority = priority
	}

	if input.AssigneeID != nil {
		if err := task.Assign(*input.AssigneeID); err != nil {
			return nil, err
		}
	}

	if input.Metadata != nil {
		metadata, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid metadata", ErrInvalidInput)
		}
		task.Metadata = metadata
	}

	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	uc.logAudit(ctx, userID, "task.created", task.ID)
	return task, nil
}

// GetByID retrieves a task by ID.
func (uc *TaskUseCase) GetByID(ctx context.Context, id int64) (*entities.Task, error) {
	task, err := uc.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// Update updates a task.
func (uc *TaskUseCase) Update(ctx context.Context, userID, taskID int64, input dto.UpdateTaskInput) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if !task.CanEdit() {
		return nil, ErrCannotModifyTask
	}

	// Track changes for history
	changes := make(map[string][2]*string)

	if input.Title != nil && *input.Title != task.Title {
		oldVal := task.Title
		task.Title = *input.Title
		changes["title"] = [2]*string{&oldVal, input.Title}
	}

	if input.Description != nil {
		oldVal := ""
		if task.Description != nil {
			oldVal = *task.Description
		}
		task.Description = input.Description
		changes["description"] = [2]*string{&oldVal, input.Description}
	}

	if input.ProjectID != nil {
		task.ProjectID = input.ProjectID
	}

	if input.Priority != nil {
		priority := domain.TaskPriority(*input.Priority)
		if !priority.IsValid() {
			return nil, fmt.Errorf("%w: invalid priority", ErrInvalidInput)
		}
		oldVal := string(task.Priority)
		task.SetPriority(priority)
		changes["priority"] = [2]*string{&oldVal, input.Priority}
	}

	if input.DueDate != nil {
		task.SetDueDate(input.DueDate)
	}

	if input.StartDate != nil {
		task.StartDate = input.StartDate
	}

	if input.Progress != nil {
		if err := task.SetProgress(*input.Progress); err != nil {
			return nil, err
		}
	}

	if input.EstimatedHours != nil {
		task.EstimatedHours = input.EstimatedHours
	}

	if input.ActualHours != nil {
		task.ActualHours = input.ActualHours
	}

	if input.Tags != nil {
		task.Tags = input.Tags
	}

	if input.Metadata != nil {
		metadata, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid metadata", ErrInvalidInput)
		}
		task.Metadata = metadata
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Record history
	for field, vals := range changes {
		history := entities.NewTaskHistory(taskID, &userID, field, vals[0], vals[1])
		_ = uc.taskRepo.AddHistory(ctx, history)
	}

	uc.logAudit(ctx, userID, "task.updated", taskID)
	return task, nil
}

// Delete deletes a task.
func (uc *TaskUseCase) Delete(ctx context.Context, userID, taskID int64) error {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Only author can delete
	if task.AuthorID != userID {
		return ErrUnauthorized
	}

	if err := uc.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	uc.logAudit(ctx, userID, "task.deleted", taskID)
	return nil
}

// List lists tasks with filters.
func (uc *TaskUseCase) List(ctx context.Context, input dto.TaskFilterInput) (*dto.TaskListOutput, error) {
	filter := input.ToTaskFilter()

	tasks, err := uc.taskRepo.List(ctx, filter, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	total, err := uc.taskRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	output := &dto.TaskListOutput{
		Tasks:  make([]dto.TaskOutput, 0, len(tasks)),
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	for _, task := range tasks {
		output.Tasks = append(output.Tasks, dto.ToTaskOutput(task))
	}

	return output, nil
}

// Assign assigns a task to a user.
func (uc *TaskUseCase) Assign(ctx context.Context, userID, taskID int64, input dto.AssignTaskInput) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldAssignee := task.AssigneeID
	if err := task.Assign(input.AssigneeID); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to assign task: %w", err)
	}

	// Record history
	var oldVal, newVal *string
	if oldAssignee != nil {
		s := fmt.Sprintf("%d", *oldAssignee)
		oldVal = &s
	}
	s := fmt.Sprintf("%d", input.AssigneeID)
	newVal = &s
	history := entities.NewTaskHistory(taskID, &userID, "assignee_id", oldVal, newVal)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.assigned", taskID)

	// Send notification to the new assignee
	if uc.notificationUseCase != nil && input.AssigneeID != userID {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			link := fmt.Sprintf("/tasks/%d", taskID)
			_ = uc.notificationUseCase.SendTaskNotification(
				context.Background(),
				input.AssigneeID,
				"Новая задача",
				fmt.Sprintf("Вам назначена задача: %s", task.Title),
				link,
			)
		}()
	}

	return task, nil
}

// Unassign removes the assignee from a task.
func (uc *TaskUseCase) Unassign(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := task.Unassign(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to unassign task: %w", err)
	}

	uc.logAudit(ctx, userID, "task.unassigned", taskID)
	return task, nil
}

// StartWork starts work on a task.
func (uc *TaskUseCase) StartWork(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(task.Status)
	if err := task.StartWork(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to start work: %w", err)
	}

	newStatus := string(task.Status)
	history := entities.NewTaskHistory(taskID, &userID, "status", &oldStatus, &newStatus)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.started", taskID)
	return task, nil
}

// SubmitForReview submits a task for review.
func (uc *TaskUseCase) SubmitForReview(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(task.Status)
	if err := task.SubmitForReview(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to submit for review: %w", err)
	}

	newStatus := string(task.Status)
	history := entities.NewTaskHistory(taskID, &userID, "status", &oldStatus, &newStatus)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.submitted_for_review", taskID)

	// Notify author about task submitted for review
	if uc.notificationUseCase != nil && task.AuthorID != userID {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			link := fmt.Sprintf("/tasks/%d", taskID)
			_ = uc.notificationUseCase.SendTaskNotification(
				context.Background(),
				task.AuthorID,
				"Задача на проверке",
				fmt.Sprintf("Задача «%s» отправлена на проверку", task.Title),
				link,
			)
		}()
	}

	return task, nil
}

// Complete completes a task.
func (uc *TaskUseCase) Complete(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(task.Status)
	if err := task.Complete(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	newStatus := string(task.Status)
	history := entities.NewTaskHistory(taskID, &userID, "status", &oldStatus, &newStatus)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.completed", taskID)

	// Send notification to the author about task completion
	if uc.notificationUseCase != nil && task.AuthorID != userID {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			link := fmt.Sprintf("/tasks/%d", taskID)
			_ = uc.notificationUseCase.SendTaskNotification(
				context.Background(),
				task.AuthorID,
				"Задача выполнена",
				fmt.Sprintf("Задача «%s» была завершена", task.Title),
				link,
			)
		}()
	}

	return task, nil
}

// Cancel cancels a task.
func (uc *TaskUseCase) Cancel(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(task.Status)
	if err := task.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	newStatus := string(task.Status)
	history := entities.NewTaskHistory(taskID, &userID, "status", &oldStatus, &newStatus)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.cancelled", taskID)

	// Notify assignee about task cancellation
	if uc.notificationUseCase != nil && task.AssigneeID != nil && *task.AssigneeID != userID {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			link := fmt.Sprintf("/tasks/%d", taskID)
			_ = uc.notificationUseCase.SendTaskNotification(
				context.Background(),
				*task.AssigneeID,
				"Задача отменена",
				fmt.Sprintf("Задача «%s» была отменена", task.Title),
				link,
			)
		}()
	}

	return task, nil
}

// Reopen reopens a task.
func (uc *TaskUseCase) Reopen(ctx context.Context, userID, taskID int64) (*entities.Task, error) {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(task.Status)
	if err := task.Reopen(); err != nil {
		return nil, err
	}

	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to reopen task: %w", err)
	}

	newStatus := string(task.Status)
	history := entities.NewTaskHistory(taskID, &userID, "status", &oldStatus, &newStatus)
	_ = uc.taskRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, userID, "task.reopened", taskID)
	return task, nil
}

// AddWatcher adds a watcher to a task.
func (uc *TaskUseCase) AddWatcher(ctx context.Context, userID, taskID, watcherID int64) error {
	task, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	watching, err := uc.taskRepo.IsWatching(ctx, task.ID, watcherID)
	if err != nil {
		return fmt.Errorf("failed to check watcher: %w", err)
	}
	if watching {
		return nil // Already watching
	}

	watcher := entities.NewTaskWatcher(taskID, watcherID)
	if err := uc.taskRepo.AddWatcher(ctx, watcher); err != nil {
		return fmt.Errorf("failed to add watcher: %w", err)
	}

	uc.logAudit(ctx, userID, "task.watcher_added", taskID)
	return nil
}

// RemoveWatcher removes a watcher from a task.
func (uc *TaskUseCase) RemoveWatcher(ctx context.Context, userID, taskID, watcherID int64) error {
	if err := uc.taskRepo.RemoveWatcher(ctx, taskID, watcherID); err != nil {
		return fmt.Errorf("failed to remove watcher: %w", err)
	}

	uc.logAudit(ctx, userID, "task.watcher_removed", taskID)
	return nil
}

// GetWatchers gets watchers of a task.
func (uc *TaskUseCase) GetWatchers(ctx context.Context, taskID int64) ([]*entities.TaskWatcher, error) {
	return uc.taskRepo.GetWatchers(ctx, taskID)
}

// AddComment adds a comment to a task.
func (uc *TaskUseCase) AddComment(ctx context.Context, userID, taskID int64, input dto.AddCommentInput) (*entities.TaskComment, error) {
	_, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	comment := entities.NewTaskComment(taskID, userID, input.Content)
	if input.ParentCommentID != nil {
		comment.SetParent(*input.ParentCommentID)
	}

	if err := uc.taskRepo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	uc.logAudit(ctx, userID, "task.comment_added", taskID)

	// Notify author and assignee about new comment
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			task, err := uc.taskRepo.GetByID(context.Background(), taskID)
			if err != nil || task == nil {
				return
			}
			link := fmt.Sprintf("/tasks/%d", taskID)
			// Notify author if not the commenter
			if task.AuthorID != userID {
				_ = uc.notificationUseCase.SendTaskNotification(
					context.Background(),
					task.AuthorID,
					"Новый комментарий",
					fmt.Sprintf("Новый комментарий к задаче «%s»", task.Title),
					link,
				)
			}
			// Notify assignee if different from author and commenter
			if task.AssigneeID != nil && *task.AssigneeID != userID && *task.AssigneeID != task.AuthorID {
				_ = uc.notificationUseCase.SendTaskNotification(
					context.Background(),
					*task.AssigneeID,
					"Новый комментарий",
					fmt.Sprintf("Новый комментарий к задаче «%s»", task.Title),
					link,
				)
			}
		}()
	}

	return comment, nil
}

// UpdateComment updates a comment.
func (uc *TaskUseCase) UpdateComment(ctx context.Context, userID, commentID int64, input dto.UpdateCommentInput) (*entities.TaskComment, error) {
	comment, err := uc.taskRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}

	if comment.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	comment.Update(input.Content)

	if err := uc.taskRepo.UpdateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	uc.logAudit(ctx, userID, "task.comment_updated", comment.TaskID)
	return comment, nil
}

// DeleteComment deletes a comment.
func (uc *TaskUseCase) DeleteComment(ctx context.Context, userID, commentID int64) error {
	comment, err := uc.taskRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	if comment.AuthorID != userID {
		return ErrUnauthorized
	}

	if err := uc.taskRepo.DeleteComment(ctx, commentID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	uc.logAudit(ctx, userID, "task.comment_deleted", comment.TaskID)
	return nil
}

// GetComments gets comments of a task.
func (uc *TaskUseCase) GetComments(ctx context.Context, taskID int64) ([]*entities.TaskComment, error) {
	return uc.taskRepo.GetComments(ctx, taskID)
}

// AddChecklist adds a checklist to a task.
func (uc *TaskUseCase) AddChecklist(ctx context.Context, userID, taskID int64, input dto.AddChecklistInput) (*entities.TaskChecklist, error) {
	_, err := uc.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	checklists, err := uc.taskRepo.GetChecklists(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checklists: %w", err)
	}

	checklist := entities.NewTaskChecklist(taskID, input.Title, len(checklists))

	if err := uc.taskRepo.AddChecklist(ctx, checklist); err != nil {
		return nil, fmt.Errorf("failed to add checklist: %w", err)
	}

	uc.logAudit(ctx, userID, "task.checklist_added", taskID)
	return checklist, nil
}

// DeleteChecklist deletes a checklist.
func (uc *TaskUseCase) DeleteChecklist(ctx context.Context, userID, checklistID int64) error {
	if err := uc.taskRepo.DeleteChecklist(ctx, checklistID); err != nil {
		return fmt.Errorf("failed to delete checklist: %w", err)
	}

	uc.logAudit(ctx, userID, "task.checklist_deleted", checklistID)
	return nil
}

// GetChecklists gets checklists of a task.
func (uc *TaskUseCase) GetChecklists(ctx context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	checklists, err := uc.taskRepo.GetChecklists(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Load items for each checklist
	for _, checklist := range checklists {
		items, err := uc.taskRepo.GetChecklistItems(ctx, checklist.ID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			checklist.Items = append(checklist.Items, *item)
		}
	}

	return checklists, nil
}

// AddChecklistItem adds an item to a checklist.
func (uc *TaskUseCase) AddChecklistItem(ctx context.Context, userID, checklistID int64, input dto.AddChecklistItemInput) (*entities.TaskChecklistItem, error) {
	items, err := uc.taskRepo.GetChecklistItems(ctx, checklistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checklist items: %w", err)
	}

	item := entities.NewTaskChecklistItem(checklistID, input.Title, len(items))

	if err := uc.taskRepo.AddChecklistItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add checklist item: %w", err)
	}

	uc.logAudit(ctx, userID, "task.checklist_item_added", checklistID)
	return item, nil
}

// ToggleChecklistItem toggles a checklist item completion.
func (uc *TaskUseCase) ToggleChecklistItem(ctx context.Context, userID, itemID int64) (*entities.TaskChecklistItem, error) {
	items, err := uc.taskRepo.GetChecklistItems(ctx, itemID)
	if err != nil || len(items) == 0 {
		return nil, fmt.Errorf("failed to get checklist item: %w", err)
	}

	// Find item by ID (we need to iterate since GetChecklistItems returns items by checklist)
	// For simplicity, we'll need to add a direct GetChecklistItemByID method
	// For now, assume item toggle works
	return nil, errors.New("not implemented")
}

// DeleteChecklistItem deletes a checklist item.
func (uc *TaskUseCase) DeleteChecklistItem(ctx context.Context, userID, itemID int64) error {
	if err := uc.taskRepo.DeleteChecklistItem(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete checklist item: %w", err)
	}

	uc.logAudit(ctx, userID, "task.checklist_item_deleted", itemID)
	return nil
}

// GetHistory gets the history of a task.
func (uc *TaskUseCase) GetHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	return uc.taskRepo.GetHistory(ctx, taskID, limit, offset)
}

// logAudit logs an audit event.
func (uc *TaskUseCase) logAudit(ctx context.Context, userID int64, action string, resourceID int64) {
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, action, "task", map[string]interface{}{
			"user_id":     userID,
			"resource_id": resourceID,
		})
	}
}
