package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/handlers"
)

// fakeReminderRepo replicates the repository contract for these
// integration tests без a real PG. The pattern mirrors the use case
// unit tests in application/usecases.
type fakeReminderRepo struct {
	mu     sync.Mutex
	rows   map[int64]*entities.TaskReminder
	nextID int64
}

func newFakeReminderRepo() *fakeReminderRepo {
	return &fakeReminderRepo{rows: map[int64]*entities.TaskReminder{}, nextID: 100}
}

func (r *fakeReminderRepo) Create(_ context.Context, rem *entities.TaskReminder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	*rem = *entities.HydrateFromPersistence(
		r.nextID, rem.TaskID(), rem.UserID(), rem.ReminderType(),
		rem.MinutesBefore(), rem.IsSent(), rem.SentAt(), rem.CreatedAt(),
	)
	r.rows[r.nextID] = rem
	return nil
}

func (r *fakeReminderRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rows[id]; !ok {
		// Mirror the PG repo sentinel via local var so test code
		// doesn't need to import the persistence package.
		return errReminderNotFound
	}
	delete(r.rows, id)
	return nil
}

func (r *fakeReminderRepo) GetByID(_ context.Context, id int64) (*entities.TaskReminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rem, ok := r.rows[id]
	if !ok {
		return nil, errReminderNotFound
	}
	return rem, nil
}

func (r *fakeReminderRepo) ListByTaskAndUser(_ context.Context, taskID, userID int64) ([]*entities.TaskReminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []*entities.TaskReminder{}
	for _, rem := range r.rows {
		if rem.TaskID() == taskID && rem.UserID() == userID {
			out = append(out, rem)
		}
	}
	return out, nil
}

func (r *fakeReminderRepo) GetPendingReminders(_ context.Context, _ time.Time) ([]*entities.TaskReminder, error) {
	return nil, nil
}

func (r *fakeReminderRepo) MarkSentBatch(_ context.Context, _ []int64, _ time.Time) error {
	return nil
}

// errReminderNotFound stands in for the PG repo's
// ErrTaskReminderNotFound — both wrap к 404 in the handler so the
// test cares only about behavior, not the exact sentinel identity.
// The handler's switch chain catches the matching path via
// errors.Is against the PG sentinel directly; this fake's returns
// surface as the default 500 path в the handler. Because the
// integration test only exercises happy paths + privacy boundary +
// CORS + 400 (bad body), we never trip the not-found branch from
// the fake — the route-extractor test pins the surface, not the
// 404 mapping (covered separately by the use case unit tests).
var errReminderNotFound = stubError("reminder not found in fake repo")

type stubError string

func (e stubError) Error() string { return string(e) }

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

// withAuth mirrors production JWTMiddleware: writes user_id +
// role context keys so downstream handler code can read them.
func withAuth(uid int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != 0 {
			c.Set("user_id", uid)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
}

// newTestEngine assembles a production-shaped router: an /api
// protected group with withAuth applied, then
// RegisterTaskReminderRoutes mounts the reminder routes. The route
// extractor's contract is exercised end-to-end.
func newTestEngine(t *testing.T, uid int64, role string, repo *fakeReminderRepo) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.Use(withAuth(uid, role))

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	setUC := usecases.NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	listUC := usecases.NewListTaskRemindersUseCase(repo)
	delUC := usecases.NewDeleteReminderUseCase(repo, nil)
	handler := handlers.NewTaskReminderHandler(setUC, listUC, delUC)

	RegisterTaskReminderRoutes(api, handler)
	return r
}

// TestRegisterTaskReminderRoutes_Create_201 pins the POST surface.
// Caller posts a valid body, receives 201 + the new reminder DTO.
func TestRegisterTaskReminderRoutes_Create_201(t *testing.T) {
	repo := newFakeReminderRepo()
	r := newTestEngine(t, 7, "methodist", repo)

	body, _ := json.Marshal(dto.CreateTaskReminderRequest{
		ReminderType:  "telegram",
		MinutesBefore: 15,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/42/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())
	var resp dto.TaskReminderResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(101), resp.ID)
	assert.Equal(t, int64(42), resp.TaskID)
	assert.Equal(t, int64(7), resp.UserID, "ActorUserID derives from JWT, not body")
	assert.Equal(t, "telegram", resp.ReminderType)
	assert.Equal(t, 15, resp.MinutesBefore)
	assert.False(t, resp.IsSent)
}

// TestRegisterTaskReminderRoutes_Create_InvalidReminderType_422
// verifies that invalid reminder_type maps к 422 (domain validation).
func TestRegisterTaskReminderRoutes_Create_InvalidReminderType_422(t *testing.T) {
	repo := newFakeReminderRepo()
	r := newTestEngine(t, 7, "methodist", repo)

	body, _ := json.Marshal(dto.CreateTaskReminderRequest{
		ReminderType:  "slack", // not в the enum
		MinutesBefore: 15,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/42/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code, "body: %s", w.Body.String())
}

// TestRegisterTaskReminderRoutes_Create_InvalidMinutes_422 verifies
// minutes_before <= 0 → 422.
func TestRegisterTaskReminderRoutes_Create_InvalidMinutes_422(t *testing.T) {
	repo := newFakeReminderRepo()
	r := newTestEngine(t, 7, "methodist", repo)

	body, _ := json.Marshal(dto.CreateTaskReminderRequest{
		ReminderType:  "email",
		MinutesBefore: 0,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/42/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// TestRegisterTaskReminderRoutes_Create_NoAuth_401 verifies the
// handler refuses requests без user_id в context (e.g., if JWT
// middleware is bypassed).
func TestRegisterTaskReminderRoutes_Create_NoAuth_401(t *testing.T) {
	repo := newFakeReminderRepo()
	r := newTestEngine(t, 0, "", repo) // uid=0 → withAuth writes nothing

	body, _ := json.Marshal(dto.CreateTaskReminderRequest{
		ReminderType:  "email",
		MinutesBefore: 15,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/42/reminders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRegisterTaskReminderRoutes_List_FiltersByCaller pins the
// privacy boundary: user 7 sees own 2 reminders; user 8's reminder
// on the same task is invisible.
func TestRegisterTaskReminderRoutes_List_FiltersByCaller(t *testing.T) {
	repo := newFakeReminderRepo()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	for _, in := range []usecases.SetReminderInput{
		{TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15},
		{TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeTelegram, MinutesBefore: 30},
		{TaskID: 42, ActorUserID: 8, ReminderType: entities.ReminderTypeInApp, MinutesBefore: 60},
	} {
		setUC := usecases.NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
		_, err := setUC.Execute(context.Background(), in)
		require.NoError(t, err)
	}

	r := newTestEngine(t, 7, "methodist", repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/42/reminders", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp []dto.TaskReminderResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp, 2, "user 7 sees own 2 reminders, not user 8's")
	for _, r := range resp {
		assert.Equal(t, int64(7), r.UserID)
	}
}

// TestRegisterTaskReminderRoutes_Delete_204 — owner deletes own.
func TestRegisterTaskReminderRoutes_Delete_204(t *testing.T) {
	repo := newFakeReminderRepo()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	setUC := usecases.NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), usecases.SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	r := newTestEngine(t, 7, "methodist", repo)
	w := httptest.NewRecorder()
	url := "/api/tasks/42/reminders/" + int64ToStr(rem.ID())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code, "body: %s", w.Body.String())
}

// TestRegisterTaskReminderRoutes_Delete_WrongOwner_403 pins the
// ownership gate — user 99 deletes user 7's reminder → 403.
func TestRegisterTaskReminderRoutes_Delete_WrongOwner_403(t *testing.T) {
	repo := newFakeReminderRepo()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	setUC := usecases.NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), usecases.SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	r := newTestEngine(t, 99, "student", repo) // different user
	w := httptest.NewRecorder()
	url := "/api/tasks/42/reminders/" + int64ToStr(rem.ID())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code, "non-owner delete → 403")
}

// TestRegisterTaskReminderRoutes_Delete_WrongTask_404 — reminder
// exists но belongs to a different task. 404 без leaking task_id.
func TestRegisterTaskReminderRoutes_Delete_WrongTask_404(t *testing.T) {
	repo := newFakeReminderRepo()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	setUC := usecases.NewSetReminderUseCase(repo, fakeClock{now: now}, nil)
	rem, err := setUC.Execute(context.Background(), usecases.SetReminderInput{
		TaskID: 42, ActorUserID: 7, ReminderType: entities.ReminderTypeEmail, MinutesBefore: 15,
	})
	require.NoError(t, err)

	r := newTestEngine(t, 7, "methodist", repo)
	w := httptest.NewRecorder()
	// Wrong task path: 99 instead of 42.
	url := "/api/tasks/99/reminders/" + int64ToStr(rem.ID())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

// TestRegisterTaskReminderRoutes_OptionsCORS_204 pins the CORS
// preflight surface for both /reminders and /reminders/:id endpoints.
func TestRegisterTaskReminderRoutes_OptionsCORS_204(t *testing.T) {
	repo := newFakeReminderRepo()
	r := newTestEngine(t, 7, "methodist", repo)
	cases := []string{
		"/api/tasks/42/reminders",
		"/api/tasks/42/reminders/101",
	}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodOptions, path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNoContent, w.Code, "OPTIONS %s must respond 204", path)
		})
	}
}

// int64ToStr — tiny helper to keep tests dependency-free.
func int64ToStr(n int64) string {
	// 64-bit fits в a 20-char buffer.
	var buf [20]byte
	i := len(buf)
	if n == 0 {
		return "0"
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
