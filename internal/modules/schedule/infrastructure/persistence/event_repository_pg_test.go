package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

func newEventRepoMock(t *testing.T) (*EventRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewEventRepositoryPG(db), mock
}

func newParticipantRepoMock(t *testing.T) (*EventParticipantRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewEventParticipantRepositoryPG(db), mock
}

func newReminderRepoMock(t *testing.T) (*EventReminderRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewEventReminderRepositoryPG(db), mock
}

var eventCols = []string{
	"id", "title", "description", "event_type", "status",
	"start_time", "end_time", "all_day", "timezone", "location",
	"organizer_id", "is_recurring", "recurrence_rule",
	"parent_event_id", "recurrence_id", "color", "priority",
	"metadata", "external_id", "created_at", "updated_at", "deleted_at",
}

func addEventRow(rows *sqlmock.Rows, id int64, title string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, title, nil, "meeting", "scheduled",
		now, nil, false, "Europe/Moscow", nil,
		int64(1), false, nil,
		nil, nil, nil, 3,
		nil, nil, now, now, nil,
	)
}

func addEventRowWithJSON(rows *sqlmock.Rows, id int64, title string, recurrence, metadata []byte) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, title, nil, "meeting", "scheduled",
		now, nil, false, "Europe/Moscow", nil,
		int64(1), true, recurrence,
		nil, nil, nil, 3,
		metadata, nil, now, now, nil,
	)
}

var participantCols = []string{"id", "event_id", "user_id", "response_status", "role", "notified_at", "responded_at", "created_at"}
var reminderCols = []string{"id", "event_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}

// ====== Event Repository ======

func TestEventCreate_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Now()
	event := &entities.Event{
		Title: "Meeting", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: now, AllDay: false, Timezone: "Europe/Moscow", OrganizerID: 1, Priority: 3,
		CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(
			event.Title, event.Description, event.EventType, event.Status,
			event.StartTime, event.EndTime, event.AllDay, event.Timezone, event.Location,
			event.OrganizerID, event.IsRecurring, nil,
			event.ParentEventID, event.RecurrenceID, event.Color, event.Priority,
			nil, event.ExternalID, event.CreatedAt, event.UpdatedAt,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), event)
	require.NoError(t, err)
	assert.Equal(t, int64(1), event.ID)
}

func TestEventCreate_WithRecurrenceAndMetadata(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Now()
	rule := &entities.RecurrenceRule{Frequency: entities.FrequencyWeekly, Interval: 1}
	meta := map[string]interface{}{"key": "val"}
	event := &entities.Event{
		Title: "Weekly", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled,
		StartTime: now, Timezone: "UTC", OrganizerID: 1, Priority: 3,
		IsRecurring: true, RecurrenceRule: rule, Metadata: meta,
		CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))

	err := repo.Create(context.Background(), event)
	require.NoError(t, err)
	assert.Equal(t, int64(2), event.ID)
}

func TestEventCreate_Error(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	event := &entities.Event{Title: "Meeting", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), event)
	assert.Error(t, err)
}

func TestEventUpdate_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	event := &entities.Event{ID: 1, Title: "Updated", EventType: entities.EventTypeMeeting, Status: entities.EventStatusScheduled, Timezone: "UTC", OrganizerID: 1, Priority: 3}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), event)
	require.NoError(t, err)
}

func TestEventUpdate_WithRecurrence(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	rule := &entities.RecurrenceRule{Frequency: entities.FrequencyDaily, Interval: 1}
	meta := map[string]interface{}{"k": "v"}
	event := &entities.Event{ID: 1, Title: "R", RecurrenceRule: rule, Metadata: meta}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), event)
	require.NoError(t, err)
}

func TestEventUpdate_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	event := &entities.Event{ID: 999, Title: "X"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event not found")
}

func TestEventUpdate_Error(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	event := &entities.Event{ID: 1, Title: "X"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.Update(context.Background(), event)
	assert.Error(t, err)
}

func TestEventDelete_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM events")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestEventDelete_Error(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM events")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestEventSoftDelete_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete(context.Background(), 1)
	require.NoError(t, err)
}

func TestEventSoftDelete_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.SoftDelete(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event not found")
}

func TestEventSoftDelete_Error(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE events SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.SoftDelete(context.Background(), 1)
	assert.Error(t, err)
}

func TestEventGetByID_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "Meeting")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	event, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Meeting", event.Title)
}

func TestEventGetByID_WithJSONFields(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	recurrence, _ := json.Marshal(&entities.RecurrenceRule{Frequency: entities.FrequencyWeekly, Interval: 1})
	metadata, _ := json.Marshal(map[string]interface{}{"key": "val"})

	rows := sqlmock.NewRows(eventCols)
	addEventRowWithJSON(rows, 1, "Weekly", recurrence, metadata)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	event, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, event.RecurrenceRule)
	assert.NotNil(t, event.Metadata)
}

func TestEventGetByID_NotFound(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event not found")
}

func TestEventGetByID_DBError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get event")
}

func TestEventGetByID_BadRecurrenceJSON(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRowWithJSON(rows, 1, "Bad", []byte(`{invalid`), nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal recurrence")
}

func TestEventGetByID_BadMetadataJSON(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRowWithJSON(rows, 1, "Bad", nil, []byte(`{invalid`))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal metadata")
}

func TestEventList_NoFilter(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WillReturnRows(rows)

	events, total, err := repo.List(context.Background(), repositories.EventFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, events, 1)
}

func TestEventList_AllFilters(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Now()
	orgID := int64(1)
	evType := entities.EventTypeMeeting
	status := entities.EventStatusScheduled
	isRecurring := true
	search := "test"

	filter := repositories.EventFilter{
		OrganizerID:    &orgID,
		EventType:      &evType,
		Status:         &status,
		StartFrom:      &now,
		StartTo:        &now,
		IsRecurring:    &isRecurring,
		SearchQuery:    &search,
		IncludeDeleted: true,
		Limit:          5,
		Offset:         0,
		OrderBy:        "start_time DESC",
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(orgID, evType, status, now, now, isRecurring, "%test%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WithArgs(orgID, evType, status, now, now, isRecurring, "%test%", 5, 0).
		WillReturnRows(rows)

	events, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, events, 1)
}

func TestEventList_CountError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), repositories.EventFilter{})
	assert.Error(t, err)
}

func TestEventList_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.List(context.Background(), repositories.EventFilter{})
	assert.Error(t, err)
}

func TestEventList_ScanError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, _, err := repo.List(context.Background(), repositories.EventFilter{})
	assert.Error(t, err)
}

func TestEventGetByDateRange_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	start := time.Now()
	end := start.Add(24 * time.Hour)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT e.id")).
		WithArgs(start, end).
		WillReturnRows(rows)

	events, err := repo.GetByDateRange(context.Background(), start, end, nil)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetByDateRange_WithUserID(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	start := time.Now()
	end := start.Add(24 * time.Hour)
	userID := int64(1)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT e.id")).
		WithArgs(start, end, userID).
		WillReturnRows(rows)

	events, err := repo.GetByDateRange(context.Background(), start, end, &userID)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetByDateRange_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT e.id")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByDateRange(context.Background(), time.Now(), time.Now(), nil)
	assert.Error(t, err)
}

func TestEventGetByOrganizer_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE organizer_id = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	events, err := repo.GetByOrganizer(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetByOrganizer_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE organizer_id = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByOrganizer(context.Background(), 1, 10, 0)
	assert.Error(t, err)
}

func TestEventGetByParticipant_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("ep.user_id = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	events, err := repo.GetByParticipant(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetByParticipant_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("ep.user_id = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByParticipant(context.Background(), 1, 10, 0)
	assert.Error(t, err)
}

func TestEventGetUpcoming_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "E1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT e.id")).
		WithArgs(int64(1), 10).
		WillReturnRows(rows)

	events, err := repo.GetUpcoming(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetUpcoming_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT e.id")).
		WithArgs(int64(1), 10).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetUpcoming(context.Background(), 1, 10)
	assert.Error(t, err)
}

func TestEventGetRecurringEvents_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 1, "Weekly")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_recurring = true")).
		WillReturnRows(rows)

	events, err := repo.GetRecurringEvents(context.Background())
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetRecurringEvents_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_recurring = true")).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetRecurringEvents(context.Background())
	assert.Error(t, err)
}

func TestEventGetRecurrenceInstances_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	start := time.Now()
	end := start.Add(24 * time.Hour)

	rows := sqlmock.NewRows(eventCols)
	addEventRow(rows, 2, "Instance")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_event_id = $1")).
		WithArgs(int64(1), start, end).
		WillReturnRows(rows)

	events, err := repo.GetRecurrenceInstances(context.Background(), 1, start, end)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}

func TestEventGetRecurrenceInstances_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_event_id = $1")).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetRecurrenceInstances(context.Background(), 1, time.Now(), time.Now())
	assert.Error(t, err)
}

func TestEventCreateRecurrenceInstance(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Now()
	event := &entities.Event{Title: "Instance", CreatedAt: now, UpdatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))

	err := repo.CreateRecurrenceInstance(context.Background(), event)
	require.NoError(t, err)
}

func TestEventGetRecurrenceExceptions_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT exception_date")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exception_date"}).AddRow(now))

	exceptions, err := repo.GetRecurrenceExceptions(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, exceptions, 1)
}

func TestEventGetRecurrenceExceptions_QueryError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT exception_date")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetRecurrenceExceptions(context.Background(), 1)
	assert.Error(t, err)
}

func TestEventGetRecurrenceExceptions_ScanError(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT exception_date")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"exception_date"}).AddRow("bad"))

	_, err := repo.GetRecurrenceExceptions(context.Background(), 1)
	assert.Error(t, err)
}

func TestEventAddRecurrenceException_Success(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_recurrence_exceptions")).
		WithArgs(int64(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AddRecurrenceException(context.Background(), 1, time.Now())
	require.NoError(t, err)
}

func TestEventAddRecurrenceException_Error(t *testing.T) {
	repo, mock := newEventRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_recurrence_exceptions")).
		WithArgs(int64(1), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.AddRecurrenceException(context.Background(), 1, time.Now())
	assert.Error(t, err)
}

// ====== Participant Repository ======

func TestParticipantCreate_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()
	p := &entities.EventParticipant{EventID: 1, UserID: 2, ResponseStatus: entities.ParticipantStatusPending, Role: entities.ParticipantRoleRequired, CreatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO event_participants")).
		WithArgs(p.EventID, p.UserID, p.ResponseStatus, p.Role, p.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
}

func TestParticipantCreate_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	p := &entities.EventParticipant{EventID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO event_participants")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), p)
	assert.Error(t, err)
}

func TestParticipantUpdate_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	p := &entities.EventParticipant{ID: 1, ResponseStatus: entities.ParticipantStatusAccepted, Role: entities.ParticipantRoleRequired}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_participants SET")).
		WithArgs(p.ResponseStatus, p.Role, p.NotifiedAt, p.RespondedAt, p.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), p)
	require.NoError(t, err)
}

func TestParticipantUpdate_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	p := &entities.EventParticipant{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_participants SET")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.Update(context.Background(), p)
	assert.Error(t, err)
}

func TestParticipantDelete_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestParticipantDelete_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestParticipantGetByID_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(participantCols).AddRow(int64(1), int64(10), int64(2), "pending", "required", nil, nil, now))

	p, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(10), p.EventID)
}

func TestParticipantGetByID_NotFound(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "participant not found")
}

func TestParticipantGetByID_DBError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestParticipantGetByEventID_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(participantCols).
			AddRow(int64(1), int64(1), int64(2), "pending", "required", nil, nil, now))

	ps, err := repo.GetByEventID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, ps, 1)
}

func TestParticipantGetByEventID_QueryError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByEventID(context.Background(), 1)
	assert.Error(t, err)
}

func TestParticipantGetByUserID_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(2), 10, 0).
		WillReturnRows(sqlmock.NewRows(participantCols).
			AddRow(int64(1), int64(1), int64(2), "accepted", "required", nil, nil, now))

	ps, err := repo.GetByUserID(context.Background(), 2, 10, 0)
	require.NoError(t, err)
	assert.Len(t, ps, 1)
}

func TestParticipantGetByUserID_QueryError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(2), 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByUserID(context.Background(), 2, 10, 0)
	assert.Error(t, err)
}

func TestParticipantGetByEventAndUser_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows(participantCols).
			AddRow(int64(1), int64(1), int64(2), "pending", "required", nil, nil, now))

	p, err := repo.GetByEventAndUser(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.UserID)
}

func TestParticipantGetByEventAndUser_NotFound(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByEventAndUser(context.Background(), 1, 999)
	assert.Error(t, err)
}

func TestParticipantGetByEventAndUser_DBError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(2)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByEventAndUser(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestParticipantAddParticipants_Empty(t *testing.T) {
	repo, _ := newParticipantRepoMock(t)
	err := repo.AddParticipants(context.Background(), 1, []int64{}, entities.ParticipantRoleRequired)
	require.NoError(t, err)
}

func TestParticipantAddParticipants_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_participants")).
		WithArgs(int64(1), sqlmock.AnyArg(), entities.ParticipantStatusPending, entities.ParticipantRoleRequired).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.AddParticipants(context.Background(), 1, []int64{2, 3}, entities.ParticipantRoleRequired)
	require.NoError(t, err)
}

func TestParticipantAddParticipants_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_participants")).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.AddParticipants(context.Background(), 1, []int64{2}, entities.ParticipantRoleRequired)
	assert.Error(t, err)
}

func TestParticipantRemoveParticipants_Empty(t *testing.T) {
	repo, _ := newParticipantRepoMock(t)
	err := repo.RemoveParticipants(context.Background(), 1, []int64{})
	require.NoError(t, err)
}

func TestParticipantRemoveParticipants_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE event_id = $1 AND user_id = ANY")).
		WithArgs(int64(1), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.RemoveParticipants(context.Background(), 1, []int64{2, 3})
	require.NoError(t, err)
}

func TestParticipantRemoveParticipants_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE event_id = $1 AND user_id = ANY")).
		WithArgs(int64(1), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.RemoveParticipants(context.Background(), 1, []int64{2})
	assert.Error(t, err)
}

func TestParticipantRemoveAllParticipants_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE event_id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	err := repo.RemoveAllParticipants(context.Background(), 1)
	require.NoError(t, err)
}

func TestParticipantRemoveAllParticipants_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_participants WHERE event_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.RemoveAllParticipants(context.Background(), 1)
	assert.Error(t, err)
}

func TestParticipantUpdateStatus_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_participants SET response_status")).
		WithArgs(entities.ParticipantStatusAccepted, sqlmock.AnyArg(), int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), 1, 2, entities.ParticipantStatusAccepted)
	require.NoError(t, err)
}

func TestParticipantUpdateStatus_NotFound(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_participants SET response_status")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1), int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateStatus(context.Background(), 1, 999, entities.ParticipantStatusAccepted)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "participant not found")
}

func TestParticipantUpdateStatus_Error(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_participants SET response_status")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1), int64(2)).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.UpdateStatus(context.Background(), 1, 2, entities.ParticipantStatusAccepted)
	assert.Error(t, err)
}

func TestParticipantGetPendingInvitations_Success(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("response_status = 'pending'")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows(participantCols).
			AddRow(int64(1), int64(10), int64(2), "pending", "required", nil, nil, now))

	ps, err := repo.GetPendingInvitations(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, ps, 1)
}

func TestParticipantGetPendingInvitations_QueryError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("response_status = 'pending'")).
		WithArgs(int64(2)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetPendingInvitations(context.Background(), 2)
	assert.Error(t, err)
}

func TestParticipantScanError(t *testing.T) {
	repo, mock := newParticipantRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetByEventID(context.Background(), 1)
	assert.Error(t, err)
}

// ====== Reminder Repository ======

func TestReminderCreate_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()
	r := &entities.EventReminder{EventID: 1, UserID: 2, ReminderType: entities.ReminderTypeInApp, MinutesBefore: 15, IsSent: false, CreatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO event_reminders")).
		WithArgs(r.EventID, r.UserID, r.ReminderType, r.MinutesBefore, r.IsSent, r.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, int64(1), r.ID)
}

func TestReminderCreate_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	r := &entities.EventReminder{EventID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO event_reminders")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), r)
	assert.Error(t, err)
}

func TestReminderUpdate_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	r := &entities.EventReminder{ID: 1, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 30, IsSent: true}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET")).
		WithArgs(r.ReminderType, r.MinutesBefore, r.IsSent, r.SentAt, r.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), r)
	require.NoError(t, err)
}

func TestReminderUpdate_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	r := &entities.EventReminder{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.Update(context.Background(), r)
	assert.Error(t, err)
}

func TestReminderDelete_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_reminders WHERE id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestReminderDelete_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_reminders WHERE id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestReminderGetByID_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("FROM event_reminders WHERE id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(reminderCols).AddRow(int64(1), int64(10), int64(2), "in_app", 15, false, nil, now))

	r, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 15, r.MinutesBefore)
}

func TestReminderGetByID_NotFound(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM event_reminders WHERE id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reminder not found")
}

func TestReminderGetByID_DBError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM event_reminders WHERE id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestReminderGetByEventID_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(reminderCols).AddRow(int64(1), int64(1), int64(2), "in_app", 15, false, nil, now))

	rs, err := repo.GetByEventID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, rs, 1)
}

func TestReminderGetByEventID_QueryError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByEventID(context.Background(), 1)
	assert.Error(t, err)
}

func TestReminderGetByUserID_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows(reminderCols).AddRow(int64(1), int64(10), int64(2), "in_app", 60, false, nil, now))

	rs, err := repo.GetByUserID(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, rs, 1)
}

func TestReminderGetByUserID_QueryError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(2)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByUserID(context.Background(), 2)
	assert.Error(t, err)
}

func TestReminderGetByEventAndUser_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows(reminderCols).AddRow(int64(1), int64(1), int64(2), "in_app", 15, false, nil, now))

	rs, err := repo.GetByEventAndUser(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Len(t, rs, 1)
}

func TestReminderGetByEventAndUser_QueryError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(2)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByEventAndUser(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestReminderGetPendingReminders_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE r.is_sent = false")).
		WithArgs(now).
		WillReturnRows(sqlmock.NewRows(reminderCols).AddRow(int64(1), int64(10), int64(2), "in_app", 15, false, nil, now))

	rs, err := repo.GetPendingReminders(context.Background(), now)
	require.NoError(t, err)
	assert.Len(t, rs, 1)
}

func TestReminderGetPendingReminders_QueryError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE r.is_sent = false")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetPendingReminders(context.Background(), time.Now())
	assert.Error(t, err)
}

func TestReminderMarkAsSent_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET is_sent = true")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkAsSent(context.Background(), 1)
	require.NoError(t, err)
}

func TestReminderMarkAsSent_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET is_sent = true")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.MarkAsSent(context.Background(), 1)
	assert.Error(t, err)
}

func TestReminderMarkMultipleAsSent_Empty(t *testing.T) {
	repo, _ := newReminderRepoMock(t)
	err := repo.MarkMultipleAsSent(context.Background(), []int64{})
	require.NoError(t, err)
}

func TestReminderMarkMultipleAsSent_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET is_sent = true")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.MarkMultipleAsSent(context.Background(), []int64{1, 2})
	require.NoError(t, err)
}

func TestReminderMarkMultipleAsSent_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_reminders SET is_sent = true")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.MarkMultipleAsSent(context.Background(), []int64{1})
	assert.Error(t, err)
}

func TestReminderDeleteByEventID_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_reminders WHERE event_id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.DeleteByEventID(context.Background(), 1)
	require.NoError(t, err)
}

func TestReminderDeleteByEventID_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_reminders WHERE event_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.DeleteByEventID(context.Background(), 1)
	assert.Error(t, err)
}

func TestReminderCreateDefault_Success(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_reminders")).
		WithArgs(int64(1), int64(2), entities.ReminderTypeInApp, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.CreateDefault(context.Background(), 1, 2)
	require.NoError(t, err)
}

func TestReminderCreateDefault_Error(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_reminders")).
		WithArgs(int64(1), int64(2), entities.ReminderTypeInApp, sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.CreateDefault(context.Background(), 1, 2)
	assert.Error(t, err)
}

func TestReminderScanError(t *testing.T) {
	repo, mock := newReminderRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE event_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetByEventID(context.Background(), 1)
	assert.Error(t, err)
}

func TestEventList_EmptySearchQuery(t *testing.T) {
	repo, mock := newEventRepoMock(t)
	emptySearch := ""

	filter := repositories.EventFilter{
		SearchQuery: &emptySearch,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title")).
		WillReturnRows(sqlmock.NewRows(eventCols))

	events, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, events, 0)
}
