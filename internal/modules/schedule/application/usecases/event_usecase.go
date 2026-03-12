// Package usecases contains business logic for the schedule module.
package usecases

import (
	"context"
	"fmt"
	"time"

	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// EventUseCase handles event business logic
type EventUseCase struct {
	eventRepo           repositories.EventRepository
	participantRepo     repositories.EventParticipantRepository
	reminderRepo        repositories.EventReminderRepository
	auditLog            *logging.AuditLogger
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewEventUseCase creates a new event use case
func NewEventUseCase(
	eventRepo repositories.EventRepository,
	participantRepo repositories.EventParticipantRepository,
	reminderRepo repositories.EventReminderRepository,
	auditLog *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *EventUseCase {
	return &EventUseCase{
		eventRepo:           eventRepo,
		participantRepo:     participantRepo,
		reminderRepo:        reminderRepo,
		auditLog:            auditLog,
		notificationUseCase: notificationUseCase,
	}
}

// Create creates a new event
func (uc *EventUseCase) Create(ctx context.Context, input dto.CreateEventInput, organizerID int64) (*dto.EventOutput, error) {
	// Validate time
	if input.EndTime != nil && input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("время окончания не может быть раньше времени начала")
	}

	// Create event entity
	event := entities.NewEvent(input.Title, entities.EventType(input.EventType), input.StartTime, organizerID)
	event.Description = input.Description
	event.EndTime = input.EndTime
	event.AllDay = input.AllDay
	event.Location = input.Location
	event.Color = input.Color

	if input.Timezone != "" {
		event.Timezone = input.Timezone
	}

	if input.Priority != nil {
		event.Priority = *input.Priority
	}

	// Set recurrence
	if input.IsRecurring && input.RecurrenceRule != nil {
		event.SetRecurrence(dto.ToRecurrenceRule(input.RecurrenceRule))
	}

	// Save event
	if err := uc.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("не удалось создать событие: %w", err)
	}

	// Add participants
	if len(input.ParticipantIDs) > 0 {
		err := uc.participantRepo.AddParticipants(ctx, event.ID, input.ParticipantIDs, entities.ParticipantRoleRequired)
		if err != nil {
			return nil, fmt.Errorf("не удалось добавить участников: %w", err)
		}

		// Notify participants about new event
		if uc.notificationUseCase != nil {
			go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
				for _, userID := range input.ParticipantIDs {
					_ = uc.notificationUseCase.SendSystemNotification(
						context.Background(),
						userID,
						"Приглашение на событие",
						fmt.Sprintf("Вы приглашены на «%s» (%s)", event.Title, event.StartTime.Format("02.01.2006 15:04")),
					)
				}
			}()
		}
	}

	// Create reminders
	if len(input.Reminders) > 0 {
		for _, r := range input.Reminders {
			reminder := entities.NewEventReminder(event.ID, organizerID, entities.ReminderType(r.ReminderType), r.MinutesBefore)
			if err := uc.reminderRepo.Create(ctx, reminder); err != nil {
				return nil, fmt.Errorf("не удалось создать напоминание: %w", err)
			}
		}
	} else {
		// Create default reminders
		if err := uc.reminderRepo.CreateDefault(ctx, event.ID, organizerID); err != nil {
			// Non-critical error, log but continue
		}
	}

	// Log audit event
	uc.logAudit(ctx, "event_created", "event", map[string]interface{}{
		"event_id":          event.ID,
		"title":             event.Title,
		"event_type":        string(event.EventType),
		"organizer_id":      organizerID,
		"is_recurring":      event.IsRecurring,
		"participant_count": len(input.ParticipantIDs),
	})

	return uc.buildEventOutput(ctx, event)
}

// Update updates an existing event
func (uc *EventUseCase) Update(ctx context.Context, id int64, input dto.UpdateEventInput, userID int64) (*dto.EventOutput, error) {
	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("событие не найдено")
	}

	if event.IsDeleted() {
		return nil, fmt.Errorf("событие удалено")
	}

	// Check permission (only organizer can edit)
	if event.OrganizerID != userID {
		uc.logAudit(ctx, "event_update_denied", "event", map[string]interface{}{
			"event_id":     id,
			"user_id":      userID,
			"organizer_id": event.OrganizerID,
			"reason":       "not organizer",
		})
		return nil, fmt.Errorf("только организатор может редактировать событие")
	}

	// Apply updates
	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Description != nil {
		event.Description = input.Description
	}
	if input.EventType != nil {
		event.EventType = entities.EventType(*input.EventType)
	}
	if input.Status != nil {
		event.Status = entities.EventStatus(*input.Status)
	}
	if input.StartTime != nil {
		event.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		event.EndTime = input.EndTime
	}
	if input.AllDay != nil {
		event.AllDay = *input.AllDay
	}
	if input.Timezone != nil {
		event.Timezone = *input.Timezone
	}
	if input.Location != nil {
		event.Location = input.Location
	}
	if input.Color != nil {
		event.Color = input.Color
	}
	if input.Priority != nil {
		event.Priority = *input.Priority
	}

	// Update recurrence
	if input.IsRecurring != nil {
		if *input.IsRecurring && input.RecurrenceRule != nil {
			event.SetRecurrence(dto.ToRecurrenceRule(input.RecurrenceRule))
		} else if !*input.IsRecurring {
			event.SetRecurrence(nil)
		}
	}

	// Validate time
	if event.EndTime != nil && event.EndTime.Before(event.StartTime) {
		return nil, fmt.Errorf("время окончания не может быть раньше времени начала")
	}

	event.UpdatedAt = time.Now()

	if err := uc.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("не удалось обновить событие: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "event_updated", "event", map[string]interface{}{
		"event_id": event.ID,
		"user_id":  userID,
	})

	// Notify participants about event update
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			participants, err := uc.participantRepo.GetByEventID(context.Background(), event.ID)
			if err == nil {
				for _, p := range participants {
					_ = uc.notificationUseCase.SendSystemNotification(
						context.Background(),
						p.UserID,
						"Событие изменено",
						fmt.Sprintf("Событие «%s» было обновлено", event.Title),
					)
				}
			}
		}()
	}

	return uc.buildEventOutput(ctx, event)
}

// Delete soft-deletes an event
func (uc *EventUseCase) Delete(ctx context.Context, id int64, userID int64) error {
	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("событие не найдено")
	}

	// Check permission
	if event.OrganizerID != userID {
		uc.logAudit(ctx, "event_delete_denied", "event", map[string]interface{}{
			"event_id":     id,
			"user_id":      userID,
			"organizer_id": event.OrganizerID,
			"reason":       "not organizer",
		})
		return fmt.Errorf("только организатор может удалить событие")
	}

	if err := uc.eventRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("не удалось удалить событие: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "event_deleted", "event", map[string]interface{}{
		"event_id": id,
		"user_id":  userID,
		"title":    event.Title,
	})

	return nil
}

// GetByID retrieves an event by ID
func (uc *EventUseCase) GetByID(ctx context.Context, id int64) (*dto.EventOutput, error) {
	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("событие не найдено")
	}

	if event.IsDeleted() {
		return nil, fmt.Errorf("событие удалено")
	}

	return uc.buildEventOutput(ctx, event)
}

// List retrieves events with filtering and pagination
func (uc *EventUseCase) List(ctx context.Context, input dto.EventFilterInput) (*dto.EventListOutput, error) {
	filter := repositories.EventFilter{
		Limit:  input.PageSize,
		Offset: (input.Page - 1) * input.PageSize,
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if input.OrganizerID != nil {
		filter.OrganizerID = input.OrganizerID
	}
	if input.EventType != nil {
		eventType := entities.EventType(*input.EventType)
		filter.EventType = &eventType
	}
	if input.Status != nil {
		status := entities.EventStatus(*input.Status)
		filter.Status = &status
	}
	if input.Search != nil {
		filter.SearchQuery = input.Search
	}
	if input.IsRecurring != nil {
		filter.IsRecurring = input.IsRecurring
	}
	if input.StartFrom != nil {
		t, err := time.Parse(time.RFC3339, *input.StartFrom)
		if err == nil {
			filter.StartFrom = &t
		}
	}
	if input.StartTo != nil {
		t, err := time.Parse(time.RFC3339, *input.StartTo)
		if err == nil {
			filter.StartTo = &t
		}
	}
	if input.OrderBy != nil {
		filter.OrderBy = *input.OrderBy
	}

	events, total, err := uc.eventRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список событий: %w", err)
	}

	outputs := make([]*dto.EventOutput, len(events))
	for i, event := range events {
		output, err := uc.buildEventOutput(ctx, event)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return &dto.EventListOutput{
		Events:     outputs,
		Total:      total,
		Page:       input.Page,
		PageSize:   filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetByDateRange retrieves events in a date range
func (uc *EventUseCase) GetByDateRange(ctx context.Context, start, end time.Time, userID *int64) ([]*dto.EventOutput, error) {
	events, err := uc.eventRepo.GetByDateRange(ctx, start, end, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить события: %w", err)
	}

	outputs := make([]*dto.EventOutput, len(events))
	for i, event := range events {
		output, err := uc.buildEventOutput(ctx, event)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	return outputs, nil
}

// GetUpcoming retrieves upcoming events for a user
func (uc *EventUseCase) GetUpcoming(ctx context.Context, userID int64, limit int) ([]*dto.EventOutput, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	events, err := uc.eventRepo.GetUpcoming(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить предстоящие события: %w", err)
	}

	outputs := make([]*dto.EventOutput, len(events))
	for i, event := range events {
		output, err := uc.buildEventOutput(ctx, event)
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}

	return outputs, nil
}

// Cancel cancels an event
func (uc *EventUseCase) Cancel(ctx context.Context, id int64, userID int64) (*dto.EventOutput, error) {
	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("событие не найдено")
	}

	if event.OrganizerID != userID {
		uc.logAudit(ctx, "event_cancel_denied", "event", map[string]interface{}{
			"event_id":     id,
			"user_id":      userID,
			"organizer_id": event.OrganizerID,
			"reason":       "not organizer",
		})
		return nil, fmt.Errorf("только организатор может отменить событие")
	}

	event.Cancel()

	if err := uc.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("не удалось отменить событие: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "event_cancelled", "event", map[string]interface{}{
		"event_id": id,
		"user_id":  userID,
		"title":    event.Title,
	})

	// Notify participants about event cancellation
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			participants, err := uc.participantRepo.GetByEventID(context.Background(), event.ID)
			if err == nil {
				for _, p := range participants {
					_ = uc.notificationUseCase.SendSystemNotification(
						context.Background(),
						p.UserID,
						"Событие отменено",
						fmt.Sprintf("Событие «%s» (%s) было отменено", event.Title, event.StartTime.Format("02.01.2006 15:04")),
					)
				}
			}
		}()
	}

	return uc.buildEventOutput(ctx, event)
}

// Reschedule reschedules an event
func (uc *EventUseCase) Reschedule(ctx context.Context, id int64, newStart time.Time, newEnd *time.Time, userID int64) (*dto.EventOutput, error) {
	event, err := uc.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("событие не найдено")
	}

	if event.OrganizerID != userID {
		uc.logAudit(ctx, "event_reschedule_denied", "event", map[string]interface{}{
			"event_id":     id,
			"user_id":      userID,
			"organizer_id": event.OrganizerID,
			"reason":       "not organizer",
		})
		return nil, fmt.Errorf("только организатор может перенести событие")
	}

	if newEnd != nil && newEnd.Before(newStart) {
		return nil, fmt.Errorf("время окончания не может быть раньше времени начала")
	}

	oldStart := event.StartTime
	event.Reschedule(newStart, newEnd)

	if err := uc.eventRepo.Update(ctx, event); err != nil {
		return nil, fmt.Errorf("не удалось перенести событие: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "event_rescheduled", "event", map[string]interface{}{
		"event_id":       id,
		"user_id":        userID,
		"old_start_time": oldStart,
		"new_start_time": newStart,
	})

	return uc.buildEventOutput(ctx, event)
}

// AddParticipants adds participants to an event
func (uc *EventUseCase) AddParticipants(ctx context.Context, eventID int64, input dto.AddParticipantsInput, userID int64) error {
	event, err := uc.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("событие не найдено")
	}

	if event.OrganizerID != userID {
		uc.logAudit(ctx, "participant_add_denied", "event", map[string]interface{}{
			"event_id":     eventID,
			"user_id":      userID,
			"organizer_id": event.OrganizerID,
			"reason":       "not organizer",
		})
		return fmt.Errorf("только организатор может добавлять участников")
	}

	role := entities.ParticipantRole(input.Role)
	if err := uc.participantRepo.AddParticipants(ctx, eventID, input.UserIDs, role); err != nil {
		return fmt.Errorf("не удалось добавить участников: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "participants_added", "event", map[string]interface{}{
		"event_id":          eventID,
		"user_id":           userID,
		"participant_ids":   input.UserIDs,
		"participant_count": len(input.UserIDs),
		"role":              input.Role,
	})

	// Notify added participants
	if uc.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			for _, uid := range input.UserIDs {
				_ = uc.notificationUseCase.SendSystemNotification(
					context.Background(),
					uid,
					"Приглашение на событие",
					fmt.Sprintf("Вы приглашены на «%s» (%s)", event.Title, event.StartTime.Format("02.01.2006 15:04")),
				)
			}
		}()
	}

	return nil
}

// RemoveParticipant removes a participant from an event
func (uc *EventUseCase) RemoveParticipant(ctx context.Context, eventID, participantUserID, userID int64) error {
	event, err := uc.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("событие не найдено")
	}

	// Organizer can remove anyone, participant can only remove themselves
	if event.OrganizerID != userID && participantUserID != userID {
		uc.logAudit(ctx, "participant_remove_denied", "event", map[string]interface{}{
			"event_id":            eventID,
			"user_id":             userID,
			"participant_user_id": participantUserID,
			"organizer_id":        event.OrganizerID,
			"reason":              "not organizer and not self",
		})
		return fmt.Errorf("недостаточно прав для удаления участника")
	}

	if err := uc.participantRepo.RemoveParticipants(ctx, eventID, []int64{participantUserID}); err != nil {
		return fmt.Errorf("не удалось удалить участника: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "participant_removed", "event", map[string]interface{}{
		"event_id":            eventID,
		"user_id":             userID,
		"participant_user_id": participantUserID,
		"self_removal":        participantUserID == userID,
	})

	return nil
}

// UpdateParticipantStatus updates participant response status
func (uc *EventUseCase) UpdateParticipantStatus(ctx context.Context, eventID int64, input dto.UpdateParticipantStatusInput, userID int64) error {
	_, err := uc.participantRepo.GetByEventAndUser(ctx, eventID, userID)
	if err != nil {
		return fmt.Errorf("вы не являетесь участником этого события")
	}

	status := entities.ParticipantStatus(input.Status)
	if err := uc.participantRepo.UpdateStatus(ctx, eventID, userID, status); err != nil {
		return fmt.Errorf("не удалось обновить статус: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "participant_status_updated", "event", map[string]interface{}{
		"event_id":   eventID,
		"user_id":    userID,
		"new_status": input.Status,
	})

	return nil
}

// GetPendingInvitations retrieves pending invitations for a user
func (uc *EventUseCase) GetPendingInvitations(ctx context.Context, userID int64) ([]*dto.EventOutput, error) {
	participations, err := uc.participantRepo.GetPendingInvitations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить приглашения: %w", err)
	}

	outputs := make([]*dto.EventOutput, 0, len(participations))
	for _, p := range participations {
		event, err := uc.eventRepo.GetByID(ctx, p.EventID)
		if err != nil {
			continue
		}
		output, err := uc.buildEventOutput(ctx, event)
		if err != nil {
			continue
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// buildEventOutput builds full event output with participants and reminders
func (uc *EventUseCase) buildEventOutput(ctx context.Context, event *entities.Event) (*dto.EventOutput, error) {
	output := dto.ToEventOutput(event)

	// Load participants
	participants, err := uc.participantRepo.GetByEventID(ctx, event.ID)
	if err == nil && len(participants) > 0 {
		output.Participants = make([]dto.ParticipantOutput, len(participants))
		for i, p := range participants {
			output.Participants[i] = *dto.ToParticipantOutput(p)
		}
	}

	// Load reminders for organizer
	reminders, err := uc.reminderRepo.GetByEventID(ctx, event.ID)
	if err == nil && len(reminders) > 0 {
		output.Reminders = make([]dto.ReminderOutput, len(reminders))
		for i, r := range reminders {
			output.Reminders[i] = *dto.ToReminderOutput(r)
		}
	}

	return output, nil
}

// logAudit safely logs an audit event with nil check
func (uc *EventUseCase) logAudit(ctx context.Context, action, resourceType string, details map[string]interface{}) {
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
