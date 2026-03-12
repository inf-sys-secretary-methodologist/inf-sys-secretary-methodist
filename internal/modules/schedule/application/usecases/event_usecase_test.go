package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
	return args.Get(0).([]*entities.Event), args.Get(1).(int64), args.Error(2)
}

func (m *MockEventRepository) GetByDateRange(ctx context.Context, start, end time.Time, userID *int64) ([]*entities.Event, error) {
	args := m.Called(ctx, start, end, userID)
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

func TestEventUseCase_Create(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	startTime := time.Now().Add(24 * time.Hour)
	input := dto.CreateEventInput{
		Title:     "Test Meeting",
		EventType: "meeting",
		StartTime: startTime,
	}

	mockEventRepo.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	mockReminderRepo.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.Create(ctx, input, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Meeting", result.Title)
	assert.Equal(t, "meeting", result.EventType)
	mockEventRepo.AssertExpectations(t)
}

func TestEventUseCase_Create_WithParticipants(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	startTime := time.Now().Add(24 * time.Hour)
	input := dto.CreateEventInput{
		Title:          "Team Meeting",
		EventType:      "meeting",
		StartTime:      startTime,
		ParticipantIDs: []int64{2, 3, 4},
	}

	mockEventRepo.On("Create", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	mockParticipantRepo.On("AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRoleRequired).Return(nil)
	mockReminderRepo.On("CreateDefault", ctx, int64(1), int64(1)).Return(nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.Create(ctx, input, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockParticipantRepo.AssertCalled(t, "AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRoleRequired)
}

func TestEventUseCase_Create_InvalidTime(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	startTime := time.Now().Add(24 * time.Hour)
	endTime := time.Now().Add(-24 * time.Hour) // End before start
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

func TestEventUseCase_Update(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	existingEvent := &entities.Event{
		ID:          1,
		Title:       "Old Title",
		OrganizerID: 1,
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		StartTime:   time.Now().Add(24 * time.Hour),
	}

	newTitle := "New Title"
	input := dto.UpdateEventInput{
		Title: &newTitle,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	mockEventRepo.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.Update(ctx, 1, input, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Title", result.Title)
}

func TestEventUseCase_Update_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	existingEvent := &entities.Event{
		ID:          1,
		Title:       "Meeting",
		OrganizerID: 1, // Organizer is user 1
	}

	newTitle := "Hacked Title"
	input := dto.UpdateEventInput{
		Title: &newTitle,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	result, err := uc.Update(ctx, 1, input, 2) // User 2 tries to update

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Delete(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	existingEvent := &entities.Event{
		ID:          1,
		OrganizerID: 1,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)
	mockEventRepo.On("SoftDelete", ctx, int64(1)).Return(nil)

	err := uc.Delete(ctx, 1, 1)

	assert.NoError(t, err)
	mockEventRepo.AssertCalled(t, "SoftDelete", ctx, int64(1))
}

func TestEventUseCase_GetByID(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Test Event",
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		OrganizerID: 1,
		StartTime:   time.Now(),
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Event", result.Title)
}

func TestEventUseCase_List(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	events := []*entities.Event{
		{ID: 1, Title: "Event 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now()},
		{ID: 2, Title: "Event 2", EventType: entities.EventTypeDeadline, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now()},
	}

	input := dto.EventFilterInput{
		Page:     1,
		PageSize: 20,
	}

	mockEventRepo.On("List", ctx, mock.AnythingOfType("repositories.EventFilter")).Return(events, int64(2), nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.List(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Events, 2)
}

func TestEventUseCase_Cancel(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Meeting to Cancel",
		OrganizerID: 1,
		Status:      entities.EventStatusScheduled,
		EventType:   entities.EventTypeMeeting,
		StartTime:   time.Now(),
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockEventRepo.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.Cancel(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(entities.EventStatusCancelled), result.Status)
}

func TestEventUseCase_Reschedule(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Meeting",
		OrganizerID: 1,
		Status:      entities.EventStatusScheduled,
		EventType:   entities.EventTypeMeeting,
		StartTime:   time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)
	newEnd := time.Now().Add(49 * time.Hour)

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockEventRepo.On("Update", ctx, mock.AnythingOfType("*entities.Event")).Return(nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.Reschedule(ctx, 1, newStart, &newEnd, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newStart.Unix(), result.StartTime.Unix())
}

func TestEventUseCase_UpdateParticipantStatus(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	participant := &entities.EventParticipant{
		ID:             1,
		EventID:        1,
		UserID:         2,
		ResponseStatus: entities.ParticipantStatusPending,
	}

	input := dto.UpdateParticipantStatusInput{
		Status: "accepted",
	}

	mockParticipantRepo.On("GetByEventAndUser", ctx, int64(1), int64(2)).Return(participant, nil)
	mockParticipantRepo.On("UpdateStatus", ctx, int64(1), int64(2), entities.ParticipantStatusAccepted).Return(nil)

	err := uc.UpdateParticipantStatus(ctx, 1, input, 2)

	assert.NoError(t, err)
	mockParticipantRepo.AssertCalled(t, "UpdateStatus", ctx, int64(1), int64(2), entities.ParticipantStatusAccepted)
}

func TestEventUseCase_GetUpcoming(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	events := []*entities.Event{
		{ID: 1, Title: "Upcoming 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour)},
		{ID: 2, Title: "Upcoming 2", EventType: entities.EventTypeDeadline, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(2 * time.Hour)},
	}

	mockEventRepo.On("GetUpcoming", ctx, int64(1), 10).Return(events, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetUpcoming(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Upcoming 1", result[0].Title)
}

func TestEventUseCase_GetByDateRange(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	start := time.Now()
	end := time.Now().Add(7 * 24 * time.Hour)
	userID := int64(1)

	events := []*entities.Event{
		{ID: 1, Title: "Event 1", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour)},
		{ID: 2, Title: "Event 2", EventType: entities.EventTypeDeadline, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(3 * 24 * time.Hour)},
	}

	mockEventRepo.On("GetByDateRange", ctx, start, end, &userID).Return(events, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(2)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetByDateRange(ctx, start, end, &userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Event 1", result[0].Title)
	assert.Equal(t, "Event 2", result[1].Title)
}

func TestEventUseCase_GetByDateRange_NoUserFilter(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	start := time.Now()
	end := time.Now().Add(7 * 24 * time.Hour)

	events := []*entities.Event{
		{ID: 1, Title: "Public Event", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, OrganizerID: 1, StartTime: time.Now().Add(1 * time.Hour)},
	}

	mockEventRepo.On("GetByDateRange", ctx, start, end, (*int64)(nil)).Return(events, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetByDateRange(ctx, start, end, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestEventUseCase_AddParticipants(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		OrganizerID: 1,
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		StartTime:   time.Now().Add(24 * time.Hour),
	}

	input := dto.AddParticipantsInput{
		UserIDs: []int64{2, 3, 4},
		Role:    "required",
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockParticipantRepo.On("AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRole("required")).Return(nil)

	err := uc.AddParticipants(ctx, 1, input, 1)

	assert.NoError(t, err)
	mockParticipantRepo.AssertCalled(t, "AddParticipants", ctx, int64(1), []int64{2, 3, 4}, entities.ParticipantRole("required"))
}

func TestEventUseCase_AddParticipants_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		OrganizerID: 1, // Organizer is user 1
	}

	input := dto.AddParticipantsInput{
		UserIDs: []int64{3, 4},
		Role:    "required",
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)

	err := uc.AddParticipants(ctx, 1, input, 2) // User 2 tries to add

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_RemoveParticipant_ByOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		OrganizerID: 1,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockParticipantRepo.On("RemoveParticipants", ctx, int64(1), []int64{2}).Return(nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 1) // Organizer (user 1) removes participant (user 2)

	assert.NoError(t, err)
	mockParticipantRepo.AssertCalled(t, "RemoveParticipants", ctx, int64(1), []int64{2})
}

func TestEventUseCase_RemoveParticipant_SelfRemoval(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		OrganizerID: 1,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)
	mockParticipantRepo.On("RemoveParticipants", ctx, int64(1), []int64{2}).Return(nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 2) // User 2 removes themselves

	assert.NoError(t, err)
	mockParticipantRepo.AssertCalled(t, "RemoveParticipants", ctx, int64(1), []int64{2})
}

func TestEventUseCase_RemoveParticipant_NotAllowed(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		OrganizerID: 1, // Organizer is user 1
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)

	err := uc.RemoveParticipant(ctx, 1, 2, 3) // User 3 tries to remove user 2

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "недостаточно прав")
}

func TestEventUseCase_GetPendingInvitations(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	participations := []*entities.EventParticipant{
		{ID: 1, EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
		{ID: 2, EventID: 3, UserID: 2, ResponseStatus: entities.ParticipantStatusPending},
	}

	event1 := &entities.Event{
		ID:          1,
		Title:       "Meeting 1",
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		OrganizerID: 1,
		StartTime:   time.Now().Add(1 * time.Hour),
	}
	event3 := &entities.Event{
		ID:          3,
		Title:       "Meeting 3",
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		OrganizerID: 1,
		StartTime:   time.Now().Add(3 * time.Hour),
	}

	mockParticipantRepo.On("GetPendingInvitations", ctx, int64(2)).Return(participations, nil)
	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event1, nil)
	mockEventRepo.On("GetByID", ctx, int64(3)).Return(event3, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventParticipant{}, nil)
	mockParticipantRepo.On("GetByEventID", ctx, int64(3)).Return([]*entities.EventParticipant{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(1)).Return([]*entities.EventReminder{}, nil)
	mockReminderRepo.On("GetByEventID", ctx, int64(3)).Return([]*entities.EventReminder{}, nil)

	result, err := uc.GetPendingInvitations(ctx, 2)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Meeting 1", result[0].Title)
	assert.Equal(t, "Meeting 3", result[1].Title)
}

func TestEventUseCase_GetPendingInvitations_Empty(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	mockParticipantRepo.On("GetPendingInvitations", ctx, int64(2)).Return([]*entities.EventParticipant{}, nil)

	result, err := uc.GetPendingInvitations(ctx, 2)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestEventUseCase_Delete_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	existingEvent := &entities.Event{
		ID:          1,
		OrganizerID: 1, // Organizer is user 1
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(existingEvent, nil)

	err := uc.Delete(ctx, 1, 2) // User 2 tries to delete

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Cancel_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Meeting",
		OrganizerID: 1, // Organizer is user 1
		Status:      entities.EventStatusScheduled,
	}

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Cancel(ctx, 1, 2) // User 2 tries to cancel

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Reschedule_NotOrganizer(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Meeting",
		OrganizerID: 1, // Organizer is user 1
		Status:      entities.EventStatusScheduled,
		StartTime:   time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Reschedule(ctx, 1, newStart, nil, 2) // User 2 tries to reschedule

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "только организатор")
}

func TestEventUseCase_Reschedule_InvalidTime(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	event := &entities.Event{
		ID:          1,
		Title:       "Meeting",
		OrganizerID: 1,
		Status:      entities.EventStatusScheduled,
		StartTime:   time.Now(),
	}

	newStart := time.Now().Add(48 * time.Hour)
	newEnd := time.Now().Add(24 * time.Hour) // End before start

	mockEventRepo.On("GetByID", ctx, int64(1)).Return(event, nil)

	result, err := uc.Reschedule(ctx, 1, newStart, &newEnd, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "время окончания")
}

func TestEventUseCase_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	mockEventRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	result, err := uc.GetByID(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "событие не найдено")
}

func TestEventUseCase_UpdateParticipantStatus_NotParticipant(t *testing.T) {
	ctx := context.Background()
	mockEventRepo := new(MockEventRepository)
	mockParticipantRepo := new(MockEventParticipantRepository)
	mockReminderRepo := new(MockEventReminderRepository)

	uc := NewEventUseCase(mockEventRepo, mockParticipantRepo, mockReminderRepo, nil, nil)

	input := dto.UpdateParticipantStatusInput{
		Status: "accepted",
	}

	mockParticipantRepo.On("GetByEventAndUser", ctx, int64(1), int64(99)).Return(nil, errors.New("not found"))

	err := uc.UpdateParticipantStatus(ctx, 1, input, 99) // User 99 is not a participant

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не являетесь участником")
}
