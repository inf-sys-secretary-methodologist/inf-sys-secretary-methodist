package handlers

// v0.153.9 Phase 6 backfill — closes task_reminder_handler.go 0% funcs:
// Create / List / Delete plus actorID / pathInt64 / projectReminder /
// handleError. Constructs three use cases against an in-memory fake
// TaskReminderRepository (mirror к existing usecases test patterns).
// No production change.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tasksUC "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	tasksEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	tasksPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/infrastructure/persistence"
)

func init() { gin.SetMode(gin.TestMode) }

// reminderFakeRepo implements TaskReminderRepository в-памяти. Only
// the methods touched by the three use cases backing the handler
// (Create / GetByID / Delete / ListByTaskAndUser) drive behavior;
// GetPendingReminders + MarkSentBatch are no-ops here.
type reminderFakeRepo struct {
	stored map[int64]*tasksEntities.TaskReminder
	nextID int64
}

func newReminderFakeRepo() *reminderFakeRepo {
	return &reminderFakeRepo{stored: map[int64]*tasksEntities.TaskReminder{}, nextID: 100}
}

func (r *reminderFakeRepo) Create(_ context.Context, rem *tasksEntities.TaskReminder) error {
	r.nextID++
	*rem = *tasksEntities.HydrateFromPersistence(
		r.nextID, rem.TaskID(), rem.UserID(), rem.ReminderType(),
		rem.MinutesBefore(), rem.IsSent(), rem.SentAt(),
		time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
	)
	r.stored[r.nextID] = rem
	return nil
}

func (r *reminderFakeRepo) Delete(_ context.Context, id int64) error {
	delete(r.stored, id)
	return nil
}

func (r *reminderFakeRepo) GetByID(_ context.Context, id int64) (*tasksEntities.TaskReminder, error) {
	rem, ok := r.stored[id]
	if !ok {
		return nil, tasksPersistence.ErrTaskReminderNotFound
	}
	return rem, nil
}

func (r *reminderFakeRepo) ListByTaskAndUser(_ context.Context, taskID, userID int64) ([]*tasksEntities.TaskReminder, error) {
	out := []*tasksEntities.TaskReminder{}
	for _, rem := range r.stored {
		if rem.TaskID() == taskID && rem.UserID() == userID {
			out = append(out, rem)
		}
	}
	return out, nil
}

func (r *reminderFakeRepo) GetPendingReminders(_ context.Context, _ time.Time) ([]*tasksEntities.TaskReminder, error) {
	return nil, nil
}

func (r *reminderFakeRepo) MarkSentBatch(_ context.Context, _ []int64, _ time.Time) error {
	return nil
}

func buildReminderRouter(t *testing.T, repo *reminderFakeRepo, userID int64) *gin.Engine {
	t.Helper()
	setUC := tasksUC.NewSetReminderUseCase(repo, nil, nil)
	listUC := tasksUC.NewListTaskRemindersUseCase(repo)
	deleteUC := tasksUC.NewDeleteReminderUseCase(repo, nil)
	h := NewTaskReminderHandler(setUC, listUC, deleteUC)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	r.POST("/api/tasks/:id/reminders", h.Create)
	r.GET("/api/tasks/:id/reminders", h.List)
	r.DELETE("/api/tasks/:id/reminders/:reminderID", h.Delete)
	return r
}

// ===== Create =====

func TestTaskReminderHandler_Create_HappyPath(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)

	body, _ := json.Marshal(map[string]any{
		"reminder_type":  "in_app",
		"minutes_before": 15,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.Len(t, repo.stored, 1, "reminder persisted")
}

func TestTaskReminderHandler_Create_Unauthorized(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 0)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/reminders",
		strings.NewReader(`{"reminder_type":"in_app","minutes_before":15}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskReminderHandler_Create_InvalidTaskID(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/abc/reminders",
		strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskReminderHandler_Create_InvalidJSON(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/reminders",
		strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskReminderHandler_Create_DomainValidationError_Returns422(t *testing.T) {
	// minutes_before=0 violates ErrInvalidMinutesBefore.
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	body, _ := json.Marshal(map[string]any{
		"reminder_type":  "in_app",
		"minutes_before": 0,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/7/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== List =====

func TestTaskReminderHandler_List_HappyPath(t *testing.T) {
	repo := newReminderFakeRepo()
	rem, err := tasksEntities.NewTaskReminder(7, 42,
		tasksEntities.ReminderTypeInApp, 15, time.Now())
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), rem))

	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/7/reminders", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "in_app")
}

func TestTaskReminderHandler_List_Unauthorized(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 0)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/7/reminders", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskReminderHandler_List_InvalidTaskID(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/abc/reminders", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===== Delete =====

func TestTaskReminderHandler_Delete_HappyPath(t *testing.T) {
	repo := newReminderFakeRepo()
	rem, err := tasksEntities.NewTaskReminder(7, 42,
		tasksEntities.ReminderTypeInApp, 15, time.Now())
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), rem))
	reminderID := rem.ID()

	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete,
		"/api/tasks/7/reminders/"+intToStr(reminderID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	_, ok := repo.stored[reminderID]
	assert.False(t, ok, "reminder removed from store")
}

func TestTaskReminderHandler_Delete_NotFound(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/7/reminders/999", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTaskReminderHandler_Delete_WrongOwner_Returns403(t *testing.T) {
	repo := newReminderFakeRepo()
	// Create reminder as user 42
	rem, err := tasksEntities.NewTaskReminder(7, 42,
		tasksEntities.ReminderTypeInApp, 15, time.Now())
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), rem))
	reminderID := rem.ID()

	// Different user 99 tries to delete — wrong owner
	r := buildReminderRouter(t, repo, 99)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete,
		"/api/tasks/7/reminders/"+intToStr(reminderID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTaskReminderHandler_Delete_Unauthorized(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 0)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/7/reminders/100", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskReminderHandler_Delete_InvalidReminderID(t *testing.T) {
	repo := newReminderFakeRepo()
	r := buildReminderRouter(t, repo, 42)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/7/reminders/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===== actorID type-switch branches =====

func TestActorID_FloatType(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", float64(99))
	id, ok := actorID(c)
	assert.True(t, ok)
	assert.Equal(t, int64(99), id)
}

func TestActorID_IntType(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", int(7))
	id, ok := actorID(c)
	assert.True(t, ok)
	assert.Equal(t, int64(7), id)
}

func TestActorID_UnsupportedType_Returns401(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "string-not-supported")
	_, ok := actorID(c)
	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ===== NewTaskReminderHandler nil-panic guards =====

func TestNewTaskReminderHandler_NilUseCase_Panics(t *testing.T) {
	repo := newReminderFakeRepo()
	setUC := tasksUC.NewSetReminderUseCase(repo, nil, nil)
	listUC := tasksUC.NewListTaskRemindersUseCase(repo)
	deleteUC := tasksUC.NewDeleteReminderUseCase(repo, nil)

	assert.Panics(t, func() { NewTaskReminderHandler(nil, listUC, deleteUC) })
	assert.Panics(t, func() { NewTaskReminderHandler(setUC, nil, deleteUC) })
	assert.Panics(t, func() { NewTaskReminderHandler(setUC, listUC, nil) })
}

func intToStr(n int64) string {
	return formatInt(n)
}

// formatInt avoids strconv import bloat for a one-off — same shape as
// strconv.FormatInt, written locally so the import set stays tight.
func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
