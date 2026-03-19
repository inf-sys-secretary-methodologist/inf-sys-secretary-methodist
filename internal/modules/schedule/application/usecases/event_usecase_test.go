package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// MockEventRepository is a mock implementation of EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *entities.Event) error {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		event.ID = 1
	}
	return args.Error(0)
}

func (m *MockEventRepository) Update(ctx context.Context, event *entities.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id int64) (*entities.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Event), args.Error(1)
}

func (m *MockEventRepository) List(ctx context.Context, filter repositories.EventFilter) ([]*entities.Event, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Event), args.Get(1).(int64), args.Error(2)
}

func (m *MockEventRepository) GetByDateRange(ctx context.Context, start, end time.Time, userID *int64) ([]*entities.Event, error) {
	args := m.Called(ctx, start, end, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) GetByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, error) {
	args := m.Called(ctx, organizerID, limit, offset)
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) GetByParticipant(ctx context.Context, userID int64, limit, offset int) ([]*entities.Event, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) GetUpcoming(ctx context.Context, userID int64, limit int) ([]*entities.Event, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) GetRecurringEvents(ctx context.Context) ([]*entities.Event, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) GetRecurrenceInstances(ctx context.Context, parentEventID int64, start, end time.Time) ([]*entities.Event, error) {
	args := m.Called(ctx, parentEventID, start, end)
	return args.Get(0).([]*entities.Event), args.Error(1)
}

func (m *MockEventRepository) CreateRecurrenceInstance(ctx context.Context, event *entities.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetRecurrenceExceptions(ctx context.Context, parentEventID int64) ([]time.Time, error) {
	args := m.Called(ctx, parentEventID)
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *MockEventRepository) AddRecurrenceException(ctx context.Context, parentEventID int64, exceptionDate time.Time) error {
	args := m.Called(ctx, parentEventID, exceptionDate)
	return args.Error(0)
}

// MockEventParticipantRepository is a mock implementation of EventParticipantRepository
type MockEventParticipantRepository struct {
	mock.Mock
}

func (m *MockEventParticipantRepository) Create(ctx context.Context, participant *entities.EventParticipant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) Update(ctx context.Context, participant *entities.EventParticipant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) GetByID(ctx context.Context, id int64) (*entities.EventParticipant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EventParticipant), args.Error(1)
}

func (m *MockEventParticipantRepository) GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventParticipant, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.EventParticipant), args.Error(1)
}

func (m *MockEventParticipantRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.EventParticipant, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.EventParticipant), args.Error(1)
}

func (m *MockEventParticipantRepository) GetByEventAndUser(ctx context.Context, eventID, userID int64) (*entities.EventParticipant, error) {
	args := m.Called(ctx, eventID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EventParticipant), args.Error(1)
}

func (m *MockEventParticipantRepository) AddParticipants(ctx context.Context, eventID int64, userIDs []int64, role entities.ParticipantRole) error {
	args := m.Called(ctx, eventID, userIDs, role)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) RemoveParticipants(ctx context.Context, eventID int64, userIDs []int64) error {
	args := m.Called(ctx, eventID, userIDs)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) RemoveAllParticipants(ctx context.Context, eventID int64) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) UpdateStatus(ctx context.Context, eventID, userID int64, status entities.ParticipantStatus) error {
	args := m.Called(ctx, eventID, userID, status)
	return args.Error(0)
}

func (m *MockEventParticipantRepository) GetPendingInvitations(ctx context.Context, userID int64) ([]*entities.EventParticipant, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.EventParticipant), args.Error(1)
}

// MockEventReminderRepository is a mock implementation of EventReminderRepository
type MockEventReminderRepository struct {
	mock.Mock
}

func (m *MockEventReminderRepository) Create(ctx context.Context, reminder *entities.EventReminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockEventReminderRepository) Update(ctx context.Context, reminder *entities.EventReminder) error {
	args := m.Called(ctx, reminder)
	return args.Error(0)
}

func (m *MockEventReminderRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventReminderRepository) GetByID(ctx context.Context, id int64) (*entities.EventReminder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EventReminder), args.Error(1)
}

func (m *MockEventReminderRepository) GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventReminder, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.EventReminder), args.Error(1)
}

func (m *MockEventReminderRepository) GetByUserID(ctx context.Context, userID int64) ([]*entities.EventReminder, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.EventReminder), args.Error(1)
}

func (m *MockEventReminderRepository) GetByEventAndUser(ctx context.Context, eventID, userID int64) ([]*entities.EventReminder, error) {
	args := m.Called(ctx, eventID, userID)
	return args.Get(0).([]*entities.EventReminder), args.Error(1)
}

func (m *MockEventReminderRepository) GetPendingReminders(ctx context.Context, beforeTime time.Time) ([]*entities.EventReminder, error) {
	args := m.Called(ctx, beforeTime)
	return args.Get(0).([]*entities.EventReminder), args.Error(1)
}

func (m *MockEventReminderRepository) MarkAsSent(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventReminderRepository) MarkMultipleAsSent(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockEventReminderRepository) DeleteByEventID(ctx context.Context, eventID int64) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockEventReminderRepository) CreateDefault(ctx context.Context, eventID, userID int64) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

// helper to create a standard test usecase with fresh mocks
func newTestUseCase() (*EventUseCase, *MockEventRepository, *MockEventParticipantRepository, *MockEventReminderRepository) {
	er := new(MockEventRepository)
	pr := new(MockEventParticipantRepository)
	rr := new(MockEventReminderRepository)
	uc := NewEventUseCase(er, pr, rr, nil, nil)
	return uc, er, pr, rr
}

// helper to set up standard buildEventOutput mocks (empty participants and reminders)
func setupBuildOutputMocks(pr *MockEventParticipantRepository, rr *MockEventReminderRepository, eventID int64) {
	pr.On("GetByEventID", mock.Anything, eventID).Return([]*entities.EventParticipant{}, nil).Maybe()
	rr.On("GetByEventID", mock.Anything, eventID).Return([]*entities.EventReminder{}, nil).Maybe()
}

// ─── Create ─────────────────────────────────────────────────────────────────

func TestEventUseCase_Create(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	startTime := time.Now().Add(24 * time.Hour)
	input := dto.CreateEventInput{
		Title:     "Test Meeting",
		EventType: "meeting",
		StartTime: startTime,
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	assert.Equal(t, "Test Meeting", result.Title)
	assert.Equal(t, "meeting", result.EventType)
	er.AssertExpectations(t)
}

func TestEventUseCase_Create_WithParticipants(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	startTime := time.Now().Add(24 * time.Hour)
	input := dto.CreateEventInput{
		Title:          "Team Meeting",
		EventType:      "meeting",
		StartTime:      startTime,
		ParticipantIDs: []int64{2, 3, 4},
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	pr.On("AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRoleRequired).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	pr.AssertCalled(t, "AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRoleRequired)
}

func TestEventUseCase_Create_InvalidTime(t *testing.T) {
	ctx := context.Background()
	uc, _, _, _ := newTestUseCase()

	startTime := time.Now().Add(24 * time.Hour)
	endTime := time.Now().Add(-24 * time.Hour)
	input := dto.CreateEventInput{
		Title:     "Invalid Meeting",
		EventType: "meeting",
		StartTime: startTime,
		EndTime:   &endTime,
	}

	result, err := uc.Create(ctx, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "время окончания не может быть раньше")
}

func TestEventUseCase_Create_WithAllOptions(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)
	desc := "A detailed description"
	loc := "Room 101"
	color := "#ff0000"
	priority := 5

	input := dto.CreateEventInput{
		Title:       "Full Event",
		Description: &desc,
		EventType:   "task",
		StartTime:   startTime,
		EndTime:     &endTime,
		AllDay:      true,
		Timezone:    "UTC",
		Location:    &loc,
		Color:       &color,
		Priority:    &priority,
		IsRecurring: true,
		RecurrenceRule: &dto.RecurrenceRuleInput{
			Frequency: "weekly",
			Interval:  1,
			ByWeekday: []string{"MO", "WE"},
		},
		Reminders: []dto.ReminderInput{
			{ReminderType: "email", MinutesBefore: 30},
			{ReminderType: "push", MinutesBefore: 10},
		},
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("Create", ctx, mock.AnythingOfType("*entities.EventReminder")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	assert.Equal(t, "Full Event", result.Title)
	assert.True(t, result.IsRecurring)
	assert.Equal(t, 5, result.Priority)
	assert.Equal(t, "UTC", result.Timezone)
	assert.Equal(t, &desc, result.Description)
	assert.Equal(t, &loc, result.Location)
	assert.Equal(t, &color, result.Color)
}

func TestEventUseCase_Create_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.CreateEventInput{
		Title:     "Failing Event",
		EventType: "meeting",
		StartTime: time.Now().Add(24 * time.Hour),
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(errors.New("db error"))

	result, err := uc.Create(ctx, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось создать событие")
}

func TestEventUseCase_Create_AddParticipantsError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	input := dto.CreateEventInput{
		Title:          "Event with bad participants",
		EventType:      "meeting",
		StartTime:      time.Now().Add(24 * time.Hour),
		ParticipantIDs: []int64{2, 3},
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	pr.On("AddParticipants", ctx, int64(1), []int64{2, 3}, entities.ParticipantRoleRequired).Return(errors.New("participant error"))

	result, err := uc.Create(ctx, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось добавить участников")
}

func TestEventUseCase_Create_ReminderCreateError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	input := dto.CreateEventInput{
		Title:     "Event with bad reminder",
		EventType: "meeting",
		StartTime: time.Now().Add(24 * time.Hour),
		Reminders: []dto.ReminderInput{
			{ReminderType: "email", MinutesBefore: 30},
		},
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("Create", ctx, mock.AnythingOfType("*entities.EventReminder")).Return(errors.New("reminder error"))
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось создать напоминание")
}

func TestEventUseCase_Create_DefaultReminderError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	input := dto.CreateEventInput{
		Title:     "Event default reminder fail",
		EventType: "meeting",
		StartTime: time.Now().Add(24 * time.Hour),
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(errors.New("non-critical error"))
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	// Default reminder errors are non-critical, should still succeed
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestEventUseCase_Create_WithRecurrenceNoRule(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	input := dto.CreateEventInput{
		Title:       "Recurring without rule",
		EventType:   "meeting",
		StartTime:   time.Now().Add(24 * time.Hour),
		IsRecurring: true,
		// RecurrenceRule is nil
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	// IsRecurring should be false because SetRecurrence was not called (rule is nil)
	assert.False(t, result.IsRecurring)
}

func TestEventUseCase_Create_EmptyTimezone(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	input := dto.CreateEventInput{
		Title:     "Default TZ",
		EventType: "meeting",
		StartTime: time.Now().Add(24 * time.Hour),
		Timezone:  "", // empty means use default
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	assert.Equal(t, "Europe/Moscow", result.Timezone)
}

// ─── Update ─────────────────────────────────────────────────────────────────

func TestEventUseCase_Update(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	existingEvent := &entities.Event{
		ID: 1, Title: "Old Title", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	newTitle := "New Title"
	input := dto.UpdateEventInput{Title: &newTitle}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Update(ctx, 1, input, 1)

	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
}

func TestEventUseCase_Update_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	existingEvent := &entities.Event{ID: 1, Title: "Meeting", OrganizerID: 1}
	newTitle := "Hacked Title"
	input := dto.UpdateEventInput{Title: &newTitle}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	result, err := uc.Update(ctx, 1, input, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	result, err := uc.Update(ctx, 999, dto.UpdateEventInput{}, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_Update_DeletedEvent(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	now := time.Now()
	existingEvent := &entities.Event{
		ID: 1, OrganizerID: 1, DeletedAt: &now,
	}
	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	result, err := uc.Update(ctx, 1, dto.UpdateEventInput{}, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие удалено")
}

func TestEventUseCase_Update_AllFields(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	existingEvent := &entities.Event{
		ID: 1, Title: "Old", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	newTitle := "New Title"
	newDesc := "New desc"
	newType := "task"
	newStatus := "ongoing"
	newStart := time.Now().Add(48 * time.Hour)
	newEnd := newStart.Add(1 * time.Hour)
	newAllDay := true
	newTZ := "UTC"
	newLoc := "Room 42"
	newColor := "#00ff00"
	newPriority := 5

	input := dto.UpdateEventInput{
		Title:       &newTitle,
		Description: &newDesc,
		EventType:   &newType,
		Status:      &newStatus,
		StartTime:   &newStart,
		EndTime:     &newEnd,
		AllDay:      &newAllDay,
		Timezone:    &newTZ,
		Location:    &newLoc,
		Color:       &newColor,
		Priority:    &newPriority,
	}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Update(ctx, 1, input, 1)

	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	assert.Equal(t, "task", result.EventType)
	assert.Equal(t, "ongoing", result.Status)
	assert.Equal(t, "UTC", result.Timezone)
	assert.Equal(t, 5, result.Priority)
	assert.True(t, result.AllDay)
}

func TestEventUseCase_Update_RecurrenceEnable(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	existingEvent := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	isRecurring := true
	input := dto.UpdateEventInput{
		IsRecurring: &isRecurring,
		RecurrenceRule: &dto.RecurrenceRuleInput{
			Frequency: "daily",
			Interval:  2,
		},
	}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Update(ctx, 1, input, 1)

	require.NoError(t, err)
	assert.True(t, result.IsRecurring)
}

func TestEventUseCase_Update_RecurrenceDisable(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	existingEvent := &entities.Event{
		ID: 1, Title: "Recurring", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
		IsRecurring: true, RecurrenceRule: &entities.RecurrenceRule{Frequency: entities.FrequencyDaily},
	}

	isRecurring := false
	input := dto.UpdateEventInput{
		IsRecurring: &isRecurring,
	}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Update(ctx, 1, input, 1)

	require.NoError(t, err)
	assert.False(t, result.IsRecurring)
}

func TestEventUseCase_Update_InvalidTime(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	startTime := time.Now().Add(48 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour) // before start
	existingEvent := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: startTime,
	}

	input := dto.UpdateEventInput{
		EndTime: &endTime,
	}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	result, err := uc.Update(ctx, 1, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "время окончания не может быть раньше")
}

func TestEventUseCase_Update_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	existingEvent := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	newTitle := "Updated"
	input := dto.UpdateEventInput{Title: &newTitle}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(errors.New("update failed"))

	result, err := uc.Update(ctx, 1, input, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось обновить событие")
}

// ─── Delete ─────────────────────────────────────────────────────────────────

func TestEventUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	existingEvent := &entities.Event{ID: 1, OrganizerID: 1}

	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("SoftDelete", ctx, int64(1)).Return(nil)

	err := uc.Delete(ctx, 1, 1)

	assert.NoError(t, err)
	er.AssertCalled(t, "SoftDelete", ctx, int64(1))
}

func TestEventUseCase_Delete_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	existingEvent := &entities.Event{ID: 1, OrganizerID: 1}
	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	err := uc.Delete(ctx, 1, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	err := uc.Delete(ctx, 999, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_Delete_SoftDeleteError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	existingEvent := &entities.Event{ID: 1, OrganizerID: 1}
	er.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	er.On("SoftDelete", ctx, int64(1)).Return(errors.New("delete error"))

	err := uc.Delete(ctx, 1, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить событие")
}

// ─── GetByID ────────────────────────────────────────────────────────────────

func TestEventUseCase_GetByID(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Test Event", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Event", result.Title)
}

func TestEventUseCase_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	result, err := uc.GetByID(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_GetByID_Deleted(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	now := time.Now()
	event := &entities.Event{
		ID: 1, Title: "Deleted Event", OrganizerID: 1, DeletedAt: &now,
	}
	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.GetByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие удалено")
}

// ─── List ───────────────────────────────────────────────────────────────────

func TestEventUseCase_List(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	events := []*entities.Event{
		{ID: 1, Title: "Event 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now()},
		{ID: 2, Title: "Event 2", EventType: entities.EventTypeDeadline, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now()},
	}

	input := dto.EventFilterInput{Page: 1, PageSize: 20}

	er.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return(events, int64(2), nil)
	setupBuildOutputMocks(pr, rr, 1)
	setupBuildOutputMocks(pr, rr, 2)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Events, 2)
	assert.Equal(t, 1, result.TotalPages)
}

func TestEventUseCase_List_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.EventFilterInput{Page: 1, PageSize: 0} // zero -> default 20

	er.On("List", ctx, mock.MatchedBy(func(f repositories.EventFilter) bool {
		return f.Limit == 20
	})).Return([]*entities.Event{}, int64(0), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, 20, result.PageSize)
}

func TestEventUseCase_List_NegativeLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.EventFilterInput{Page: 1, PageSize: -5} // negative -> default 20

	er.On("List", ctx, mock.MatchedBy(func(f repositories.EventFilter) bool {
		return f.Limit == 20
	})).Return([]*entities.Event{}, int64(0), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, 20, result.PageSize)
}

func TestEventUseCase_List_MaxLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.EventFilterInput{Page: 1, PageSize: 200} // >100 -> capped to 100

	er.On("List", ctx, mock.MatchedBy(func(f repositories.EventFilter) bool {
		return f.Limit == 100
	})).Return([]*entities.Event{}, int64(0), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, 100, result.PageSize)
}

func TestEventUseCase_List_WithAllFilters(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	organizerID := int64(5)
	eventType := "meeting"
	status := "scheduled"
	search := "important"
	isRecurring := true
	startFrom := time.Now().Format(time.RFC3339)
	startTo := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	orderBy := "start_time ASC"

	input := dto.EventFilterInput{
		OrganizerID: &organizerID,
		EventType:   &eventType,
		Status:      &status,
		Search:      &search,
		IsRecurring: &isRecurring,
		StartFrom:   &startFrom,
		StartTo:     &startTo,
		OrderBy:     &orderBy,
		Page:        1,
		PageSize:    20,
	}

	er.On("List", ctx, mock.MatchedBy(func(f repositories.EventFilter) bool {
		return f.OrganizerID != nil && *f.OrganizerID == int64(5) &&
			f.EventType != nil && *f.EventType == entities.EventTypeMeeting &&
			f.Status != nil && *f.Status == entities.EventStatusScheduled &&
			f.SearchQuery != nil && *f.SearchQuery == "important" &&
			f.IsRecurring != nil && *f.IsRecurring &&
			f.StartFrom != nil && f.StartTo != nil &&
			f.OrderBy == "start_time ASC"
	})).Return([]*entities.Event{}, int64(0), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
}

func TestEventUseCase_List_InvalidDateFormats(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	badFrom := "not-a-date"
	badTo := "also-bad"
	input := dto.EventFilterInput{
		StartFrom: &badFrom,
		StartTo:   &badTo,
		Page:      1,
		PageSize:  20,
	}

	// Bad dates should be silently ignored (StartFrom/StartTo remain nil)
	er.On("List", ctx, mock.MatchedBy(func(f repositories.EventFilter) bool {
		return f.StartFrom == nil && f.StartTo == nil
	})).Return([]*entities.Event{}, int64(0), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
}

func TestEventUseCase_List_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.EventFilterInput{Page: 1, PageSize: 20}

	er.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return(nil, int64(0), errors.New("db error"))

	result, err := uc.List(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось получить список событий")
}

func TestEventUseCase_List_TotalPagesRounding(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	events := make([]*entities.Event, 3)
	for i := 0; i < 3; i++ {
		events[i] = &entities.Event{
			ID: int64(i + 1), Title: "Event", EventType: entities.EventTypeMeeting,
			Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now(),
		}
		setupBuildOutputMocks(pr, rr, int64(i+1))
	}

	input := dto.EventFilterInput{Page: 1, PageSize: 2} // 5 total / 2 per page = 3 pages

	er.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return(events, int64(5), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalPages) // ceil(5/2) = 3
}

func TestEventUseCase_List_ExactPage(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	input := dto.EventFilterInput{Page: 1, PageSize: 5} // 10 total / 5 per page = 2 pages exact

	er.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return([]*entities.Event{}, int64(10), nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalPages)
}

// ─── GetByDateRange ─────────────────────────────────────────────────────────

func TestEventUseCase_GetByDateRange(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	start := time.Now()
	end := start.Add(7 * 24 * time.Hour)
	userID := int64(1)

	events := []*entities.Event{
		{ID: 1, Title: "Event 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: start.Add(1 * time.Hour)},
	}

	er.On("GetByDateRange", ctx, start, end, &userID).Return(events, nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.GetByDateRange(ctx, start, end, &userID)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestEventUseCase_GetByDateRange_NoUserFilter(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	start := time.Now()
	end := start.Add(7 * 24 * time.Hour)

	events := []*entities.Event{
		{ID: 1, Title: "Public Event", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: start.Add(1 * time.Hour)},
	}

	er.On("GetByDateRange", ctx, start, end, (*int64)(nil)).Return(events, nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.GetByDateRange(ctx, start, end, nil)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestEventUseCase_GetByDateRange_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	start := time.Now()
	end := start.Add(24 * time.Hour)

	er.On("GetByDateRange", ctx, start, end, (*int64)(nil)).Return(nil, errors.New("db error"))

	result, err := uc.GetByDateRange(ctx, start, end, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось получить события")
}

// ─── GetUpcoming ────────────────────────────────────────────────────────────

func TestEventUseCase_GetUpcoming(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	events := []*entities.Event{
		{ID: 1, Title: "Upcoming 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour)},
		{ID: 2, Title: "Upcoming 2", EventType: entities.EventTypeDeadline, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(2 * time.Hour)},
	}

	er.On("GetUpcoming", ctx, int64(1), 10).Return(events, nil)
	setupBuildOutputMocks(pr, rr, 1)
	setupBuildOutputMocks(pr, rr, 2)

	result, err := uc.GetUpcoming(ctx, 1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestEventUseCase_GetUpcoming_DefaultLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetUpcoming", ctx, int64(1), 10).Return([]*entities.Event{}, nil)

	result, err := uc.GetUpcoming(ctx, 1, 0) // 0 -> default 10

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestEventUseCase_GetUpcoming_NegativeLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetUpcoming", ctx, int64(1), 10).Return([]*entities.Event{}, nil)

	result, err := uc.GetUpcoming(ctx, 1, -5) // negative -> default 10

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestEventUseCase_GetUpcoming_MaxLimit(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetUpcoming", ctx, int64(1), 50).Return([]*entities.Event{}, nil)

	result, err := uc.GetUpcoming(ctx, 1, 100) // >50 -> capped to 50

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestEventUseCase_GetUpcoming_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetUpcoming", ctx, int64(1), 10).Return(nil, errors.New("db error"))

	result, err := uc.GetUpcoming(ctx, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось получить предстоящие события")
}

// ─── Cancel ─────────────────────────────────────────────────────────────────

func TestEventUseCase_Cancel(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting to Cancel", OrganizerID: 1,
		Status: entities.EventStatusScheduled, EventType: entities.EventTypeMeeting,
		StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Cancel(ctx, 1, 1)

	require.NoError(t, err)
	assert.Equal(t, string(entities.EventStatusCancelled), result.Status)
}

func TestEventUseCase_Cancel_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled,
	}
	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Cancel(ctx, 1, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Cancel_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	result, err := uc.Cancel(ctx, 999, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_Cancel_UpdateError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, EventType: entities.EventTypeMeeting,
		StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(errors.New("update failed"))

	result, err := uc.Cancel(ctx, 1, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось отменить событие")
}

// ─── Reschedule ─────────────────────────────────────────────────────────────

func TestEventUseCase_Reschedule(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, EventType: entities.EventTypeMeeting,
		StartTime: time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)
	newEnd := newStart.Add(1 * time.Hour)

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Reschedule(ctx, 1, newStart, &newEnd, 1)

	require.NoError(t, err)
	assert.Equal(t, newStart.Unix(), result.StartTime.Unix())
}

func TestEventUseCase_Reschedule_NoEndTime(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, EventType: entities.EventTypeMeeting,
		StartTime: time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Reschedule(ctx, 1, newStart, nil, 1)

	require.NoError(t, err)
	assert.Nil(t, result.EndTime)
}

func TestEventUseCase_Reschedule_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Reschedule(ctx, 1, time.Now().Add(48*time.Hour), nil, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Reschedule_InvalidTime(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, StartTime: time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)
	newEnd := time.Now().Add(24 * time.Hour) // end before start

	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Reschedule(ctx, 1, newStart, &newEnd, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "время окончания")
}

func TestEventUseCase_Reschedule_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	result, err := uc.Reschedule(ctx, 999, time.Now(), nil, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_Reschedule_UpdateError(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		Status: entities.EventStatusScheduled, EventType: entities.EventTypeMeeting,
		StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	er.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(errors.New("update failed"))

	result, err := uc.Reschedule(ctx, 1, time.Now().Add(48*time.Hour), nil, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось перенести событие")
}

// ─── AddParticipants ────────────────────────────────────────────────────────

func TestEventUseCase_AddParticipants(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Team Meeting", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	input := dto.AddParticipantsInput{UserIDs: []int64{2, 3, 4}, Role: "required"}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRole("required")).Return(nil)

	err := uc.AddParticipants(ctx, 1, input, 1)

	assert.NoError(t, err)
}

func TestEventUseCase_AddParticipants_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{ID: 1, Title: "Team Meeting", OrganizerID: 1}
	input := dto.AddParticipantsInput{UserIDs: []int64{3, 4}, Role: "required"}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	err := uc.AddParticipants(ctx, 1, input, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_AddParticipants_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	err := uc.AddParticipants(ctx, 999, dto.AddParticipantsInput{UserIDs: []int64{2}, Role: "required"}, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_AddParticipants_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Meeting", OrganizerID: 1,
		EventType: entities.EventTypeMeeting, StartTime: time.Now().Add(24 * time.Hour),
	}
	input := dto.AddParticipantsInput{UserIDs: []int64{2}, Role: "optional"}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("AddParticipants", ctx, int64(1), []int64{2}, entities.ParticipantRole("optional")).Return(errors.New("add error"))

	err := uc.AddParticipants(ctx, 1, input, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось добавить участников")
}

// ─── RemoveParticipant ──────────────────────────────────────────────────────

func TestEventUseCase_RemoveParticipant_ByOrganizer(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	event := &entities.Event{ID: 1, Title: "Team Meeting", OrganizerID: 1}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("RemoveParticipants", ctx, int64(1), []int64{2}).Return(nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 1) // organizer removes user 2

	assert.NoError(t, err)
}

func TestEventUseCase_RemoveParticipant_SelfRemoval(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	event := &entities.Event{ID: 1, Title: "Team Meeting", OrganizerID: 1}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("RemoveParticipants", ctx, int64(1), []int64{2}).Return(nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 2) // user 2 removes themselves

	assert.NoError(t, err)
}

func TestEventUseCase_RemoveParticipant_NotAllowed(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	event := &entities.Event{ID: 1, Title: "Team Meeting", OrganizerID: 1}
	er.On("GetByID", ctx, int64(1)).Return(event, nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 3) // user 3 tries to remove user 2

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "недостаточно прав")
}

func TestEventUseCase_RemoveParticipant_NotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	err := uc.RemoveParticipant(ctx, 999, 2, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_RemoveParticipant_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, _ := newTestUseCase()

	event := &entities.Event{ID: 1, Title: "Meeting", OrganizerID: 1}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("RemoveParticipants", ctx, int64(1), []int64{2}).Return(errors.New("remove error"))

	err := uc.RemoveParticipant(ctx, 1, 2, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось удалить участника")
}

// ─── UpdateParticipantStatus ────────────────────────────────────────────────

func TestEventUseCase_UpdateParticipantStatus(t *testing.T) {
	ctx := context.Background()
	uc, _, pr, _ := newTestUseCase()

	participant := &entities.EventParticipant{
		ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending,
	}

	input := dto.UpdateParticipantStatusInput{Status: "accepted"}

	pr.On("GetByEventAndUser", ctx, int64(1), int64(2)).Return(participant, nil)
	pr.On("UpdateStatus", ctx, int64(1), int64(2), entities.ParticipantStatusAccepted).Return(nil)

	err := uc.UpdateParticipantStatus(ctx, 1, input, 2)

	assert.NoError(t, err)
}

func TestEventUseCase_UpdateParticipantStatus_NotParticipant(t *testing.T) {
	ctx := context.Background()
	uc, _, pr, _ := newTestUseCase()

	input := dto.UpdateParticipantStatusInput{Status: "accepted"}
	pr.On("GetByEventAndUser", ctx, int64(1), int64(99)).Return(nil, errors.New("not found"))

	err := uc.UpdateParticipantStatus(ctx, 1, input, 99)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не являетесь участником")
}

func TestEventUseCase_UpdateParticipantStatus_UpdateError(t *testing.T) {
	ctx := context.Background()
	uc, _, pr, _ := newTestUseCase()

	participant := &entities.EventParticipant{
		ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending,
	}

	input := dto.UpdateParticipantStatusInput{Status: "declined"}

	pr.On("GetByEventAndUser", ctx, int64(1), int64(2)).Return(participant, nil)
	pr.On("UpdateStatus", ctx, int64(1), int64(2), entities.ParticipantStatusDeclined).Return(errors.New("update error"))

	err := uc.UpdateParticipantStatus(ctx, 1, input, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось обновить статус")
}

// ─── GetPendingInvitations ──────────────────────────────────────────────────

func TestEventUseCase_GetPendingInvitations(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	participations := []*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
		{ID: 2, EventID: 3, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
	}

	event1 := &entities.Event{
		ID: 1, Title: "Meeting 1", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour),
	}
	event3 := &entities.Event{
		ID: 3, Title: "Meeting 3", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(3 * time.Hour),
	}

	pr.On("GetPendingInvitations", ctx, int64(2)).Return(participations, nil)
	er.On("GetByID", ctx, int64(1)).Return(event1, nil)
	er.On("GetByID", ctx, int64(3)).Return(event3, nil)
	setupBuildOutputMocks(pr, rr, 1)
	setupBuildOutputMocks(pr, rr, 3)

	result, err := uc.GetPendingInvitations(ctx, 2)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Meeting 1", result[0].Title)
	assert.Equal(t, "Meeting 3", result[1].Title)
}

func TestEventUseCase_GetPendingInvitations_Empty(t *testing.T) {
	ctx := context.Background()
	uc, _, pr, _ := newTestUseCase()

	pr.On("GetPendingInvitations", ctx, int64(2)).Return([]*entities.EventParticipant{}, nil)

	result, err := uc.GetPendingInvitations(ctx, 2)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestEventUseCase_GetPendingInvitations_RepoError(t *testing.T) {
	ctx := context.Background()
	uc, _, pr, _ := newTestUseCase()

	pr.On("GetPendingInvitations", ctx, int64(2)).Return(nil, errors.New("db error"))

	result, err := uc.GetPendingInvitations(ctx, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "не удалось получить приглашения")
}

func TestEventUseCase_GetPendingInvitations_EventNotFound(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	participations := []*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
		{ID: 2, EventID: 99, UserID: 2, ResponseStatus: entities.ParticipantStatusPending}, // event not found
		{ID: 3, EventID: 3, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
	}

	event1 := &entities.Event{
		ID: 1, Title: "Meeting 1", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour),
	}
	event3 := &entities.Event{
		ID: 3, Title: "Meeting 3", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(3 * time.Hour),
	}

	pr.On("GetPendingInvitations", ctx, int64(2)).Return(participations, nil)
	er.On("GetByID", ctx, int64(1)).Return(event1, nil)
	er.On("GetByID", ctx, int64(99)).Return(nil, errors.New("not found"))
	er.On("GetByID", ctx, int64(3)).Return(event3, nil)
	setupBuildOutputMocks(pr, rr, 1)
	setupBuildOutputMocks(pr, rr, 3)

	result, err := uc.GetPendingInvitations(ctx, 2)

	require.NoError(t, err)
	assert.Len(t, result, 2) // one event skipped
}

func TestEventUseCase_GetPendingInvitations_BuildOutputError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	participations := []*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
	}

	event1 := &entities.Event{
		ID: 1, Title: "Meeting 1", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour),
	}

	pr.On("GetPendingInvitations", ctx, int64(2)).Return(participations, nil)
	er.On("GetByID", ctx, int64(1)).Return(event1, nil)
	// buildEventOutput always succeeds currently (it ignores participant/reminder errors)
	// but let's verify the happy path with actual participants & reminders
	pr.On("GetByEventID", ctx, int64(1)).Return(nil, errors.New("participant load error"))
	rr.On("GetByEventID", ctx, int64(1)).Return(nil, errors.New("reminder load error"))

	result, err := uc.GetPendingInvitations(ctx, 2)

	// buildEventOutput doesn't return errors for participant/reminder load failures
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// ─── buildEventOutput ───────────────────────────────────────────────────────

func TestEventUseCase_BuildEventOutput_WithParticipantsAndReminders(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Full Event", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now(),
		IsRecurring: true, RecurrenceRule: &entities.RecurrenceRule{
			Frequency: entities.FrequencyWeekly, Interval: 1,
			ByWeekday: []entities.Weekday{entities.WeekdayMonday},
		},
	}

	participants := []*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusAccepted, Role: entities.ParticipantRoleRequired},
		{ID: 2, EventID: 1, UserID: 3, ResponseStatus: entities.ParticipantStatusPending, Role: entities.ParticipantRoleOptional},
	}

	reminders := []*entities.EventReminder{
		{ID: 1, EventID: 1, UserID: 1, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 30},
		{ID: 2, EventID: 1, UserID: 1, ReminderType: entities.ReminderTypePush, MinutesBefore: 10},
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("GetByEventID", ctx, int64(1)).Return(participants, nil)
	rr.On("GetByEventID", ctx, int64(1)).Return(reminders, nil)

	result, err := uc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, result.Participants, 2)
	assert.Equal(t, int64(2), result.Participants[0].UserID)
	assert.Equal(t, "accepted", result.Participants[0].ResponseStatus)
	assert.Equal(t, "required", result.Participants[0].Role)
	assert.Len(t, result.Reminders, 2)
	assert.Equal(t, "email", result.Reminders[0].ReminderType)
	assert.Equal(t, 30, result.Reminders[0].MinutesBefore)
	assert.True(t, result.IsRecurring)
	assert.NotNil(t, result.RecurrenceRule)
}

func TestEventUseCase_BuildEventOutput_ParticipantLoadError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Event", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("GetByEventID", ctx, int64(1)).Return(nil, errors.New("load error"))
	rr.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Nil(t, result.Participants)
}

func TestEventUseCase_BuildEventOutput_ReminderLoadError(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	event := &entities.Event{
		ID: 1, Title: "Event", EventType: entities.EventTypeMeeting,
		Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now(),
	}

	er.On("GetByID", ctx, int64(1)).Return(event, nil)
	pr.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	rr.On("GetByEventID", ctx, int64(1)).Return(nil, errors.New("load error"))

	result, err := uc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.Nil(t, result.Reminders)
}

// ─── logAudit ───────────────────────────────────────────────────────────────

func TestEventUseCase_LogAudit_NilAuditLog(t *testing.T) {
	// This test ensures logAudit handles nil auditLog gracefully.
	// All existing tests already use nil auditLog, but this makes it explicit.
	uc := NewEventUseCase(nil, nil, nil, nil, nil)
	// Should not panic
	uc.logAudit(context.Background(), "test_action", "test_resource", map[string]interface{}{"key": "value"})
}

// ─── NewEventUseCase ────────────────────────────────────────────────────────

func TestNewEventUseCase(t *testing.T) {
	er := new(MockEventRepository)
	pr := new(MockEventParticipantRepository)
	rr := new(MockEventReminderRepository)

	uc := NewEventUseCase(er, pr, rr, nil, nil)

	assert.NotNil(t, uc)
	assert.Equal(t, er, uc.eventRepo)
	assert.Equal(t, pr, uc.participantRepo)
	assert.Equal(t, rr, uc.reminderRepo)
	assert.Nil(t, uc.auditLog)
	assert.Nil(t, uc.notificationUseCase)
}

// ─── List buildEventOutput error propagation ────────────────────────────────

func TestEventUseCase_List_BuildOutputErrorPropagation(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	// This test is slightly unusual: buildEventOutput doesn't actually return errors
	// in the current implementation (it swallows participant/reminder load errors).
	// But the List function does propagate buildEventOutput errors.
	// We verify that the List function works correctly with events that have
	// participants and reminders.
	events := []*entities.Event{
		{ID: 1, Title: "E1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now()},
	}

	input := dto.EventFilterInput{Page: 1, PageSize: 20}

	er.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return(events, int64(1), nil)
	pr.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusAccepted, Role: entities.ParticipantRoleRequired},
	}, nil)
	rr.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{
		{ID: 1, EventID: 1, UserID: 1, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15},
	}, nil)

	result, err := uc.List(ctx, input)

	require.NoError(t, err)
	assert.Len(t, result.Events, 1)
	assert.Len(t, result.Events[0].Participants, 1)
	assert.Len(t, result.Events[0].Reminders, 1)
}

// ─── GetByDateRange buildEventOutput error propagation ──────────────────────

func TestEventUseCase_GetByDateRange_EmptyResult(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	start := time.Now()
	end := start.Add(24 * time.Hour)

	er.On("GetByDateRange", ctx, start, end, (*int64)(nil)).Return([]*entities.Event{}, nil)

	result, err := uc.GetByDateRange(ctx, start, end, nil)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// ─── GetUpcoming empty result ───────────────────────────────────────────────

func TestEventUseCase_GetUpcoming_EmptyResult(t *testing.T) {
	ctx := context.Background()
	uc, er, _, _ := newTestUseCase()

	er.On("GetUpcoming", ctx, int64(1), 10).Return([]*entities.Event{}, nil)

	result, err := uc.GetUpcoming(ctx, 1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

// ─── Create with NilPriority (no priority set, uses default) ────────────────

func TestEventUseCase_Create_NilPriority(t *testing.T) {
	ctx := context.Background()
	uc, er, pr, rr := newTestUseCase()

	input := dto.CreateEventInput{
		Title:     "Default Priority",
		EventType: "meeting",
		StartTime: time.Now().Add(24 * time.Hour),
		// Priority is nil
	}

	er.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	rr.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	setupBuildOutputMocks(pr, rr, 1)

	result, err := uc.Create(ctx, input, 1)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Priority) // default priority
}
