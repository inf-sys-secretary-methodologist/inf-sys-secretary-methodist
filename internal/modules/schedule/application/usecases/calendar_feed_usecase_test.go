package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ---- fakes ---------------------------------------------------------------

type fakeTokenRepo struct {
	byUser  map[int64]*entities.CalendarFeedToken
	byToken map[string]*entities.CalendarFeedToken
	saves   int
	deletes []int64
	nextID  int64
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{
		byUser:  map[int64]*entities.CalendarFeedToken{},
		byToken: map[string]*entities.CalendarFeedToken{},
	}
}

func (r *fakeTokenRepo) Save(_ context.Context, t *entities.CalendarFeedToken) error {
	r.saves++
	r.nextID++
	t.ID = r.nextID
	// Upsert semantics: a user has at most one token.
	if prev, ok := r.byUser[t.UserID]; ok {
		delete(r.byToken, prev.Token)
	}
	r.byUser[t.UserID] = t
	r.byToken[t.Token] = t
	return nil
}

func (r *fakeTokenRepo) GetByUserID(_ context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	if t, ok := r.byUser[userID]; ok {
		return t, nil
	}
	return nil, entities.ErrCalendarFeedTokenNotFound
}

func (r *fakeTokenRepo) GetByToken(_ context.Context, token string) (*entities.CalendarFeedToken, error) {
	if t, ok := r.byToken[token]; ok {
		return t, nil
	}
	return nil, entities.ErrCalendarFeedTokenNotFound
}

func (r *fakeTokenRepo) DeleteByUserID(_ context.Context, userID int64) error {
	r.deletes = append(r.deletes, userID)
	if prev, ok := r.byUser[userID]; ok {
		delete(r.byToken, prev.Token)
		delete(r.byUser, userID)
	}
	return nil
}

// recordingLessonRepo embeds the shared mock and records GetTimetable filters,
// returning teacher- or group-scoped lessons depending on the filter.
type recordingLessonRepo struct {
	*MockLessonRepository
	byTeacher []*entities.Lesson
	byGroup   []*entities.Lesson
	filters   []LessonFilter
}

func (r *recordingLessonRepo) GetTimetable(_ context.Context, f LessonFilter) ([]*entities.Lesson, error) {
	r.filters = append(r.filters, f)
	switch {
	case f.TeacherID != nil:
		return r.byTeacher, nil
	case f.GroupID != nil:
		return r.byGroup, nil
	default:
		return nil, nil
	}
}

type fakeGroupResolver struct {
	groupID int64
	err     error
}

func (f fakeGroupResolver) ResolveGroupID(_ context.Context, _ int64) (int64, error) {
	return f.groupID, f.err
}

// ---- helpers -------------------------------------------------------------

func feedCfg() CalendarFeedConfig {
	return CalendarFeedConfig{
		ProdID:       "-//Secretary Methodist//Calendar Feed//EN",
		CalendarName: "Расписание",
		TZID:         "Europe/Moscow",
		UIDDomain:    "methodist",
		PastWindow:   30 * 24 * time.Hour,
		FutureWindow: 365 * 24 * time.Hour,
	}
}

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC) }
}

// ---- token management ----------------------------------------------------

func TestEnsureToken_CreatesWhenAbsent(t *testing.T) {
	repo := newFakeTokenRepo()
	uc := NewCalendarFeedUseCase(repo, NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg(),
		WithClock(fixedClock()),
		WithTokenGenerator(func() (string, error) { return "generated-token", nil }))

	tok, err := uc.EnsureToken(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "generated-token" || tok.UserID != 42 {
		t.Errorf("unexpected token %+v", tok)
	}
	if repo.saves != 1 {
		t.Errorf("expected 1 save, got %d", repo.saves)
	}
}

func TestEnsureToken_ReturnsExisting(t *testing.T) {
	repo := newFakeTokenRepo()
	existing := &entities.CalendarFeedToken{ID: 1, UserID: 42, Token: "existing"}
	repo.byUser[42] = existing
	repo.byToken["existing"] = existing

	gen := 0
	uc := NewCalendarFeedUseCase(repo, NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg(),
		WithTokenGenerator(func() (string, error) { gen++; return "new", nil }))

	tok, err := uc.EnsureToken(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "existing" {
		t.Errorf("expected existing token, got %q", tok.Token)
	}
	if gen != 0 {
		t.Errorf("generator must not run when a token exists")
	}
}

func TestRotateToken_AlwaysIssuesNew(t *testing.T) {
	repo := newFakeTokenRepo()
	existing := &entities.CalendarFeedToken{ID: 1, UserID: 42, Token: "old"}
	repo.byUser[42] = existing
	repo.byToken["old"] = existing

	uc := NewCalendarFeedUseCase(repo, NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg(),
		WithClock(fixedClock()),
		WithTokenGenerator(func() (string, error) { return "rotated", nil }))

	tok, err := uc.RotateToken(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "rotated" {
		t.Errorf("expected rotated token, got %q", tok.Token)
	}
	if _, err := repo.GetByToken(context.Background(), "old"); !errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		t.Error("old token must be invalidated after rotation")
	}
}

func TestGetToken_NotFound(t *testing.T) {
	uc := NewCalendarFeedUseCase(newFakeTokenRepo(), NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg())
	_, err := uc.GetToken(context.Background(), 42)
	if !errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		t.Errorf("expected not-found, got %v", err)
	}
}

func TestDeleteToken(t *testing.T) {
	repo := newFakeTokenRepo()
	uc := NewCalendarFeedUseCase(repo, NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg())
	if err := uc.DeleteToken(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.deletes) != 1 || repo.deletes[0] != 42 {
		t.Errorf("expected delete for user 42, got %v", repo.deletes)
	}
}

// ---- feed rendering ------------------------------------------------------

func TestRenderFeed_UnknownTokenErrors(t *testing.T) {
	uc := NewCalendarFeedUseCase(newFakeTokenRepo(), NewMockLessonRepository(), &MockEventRepository{}, fakeGroupResolver{}, feedCfg())
	_, err := uc.RenderFeed(context.Background(), "missing")
	if !errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		t.Errorf("expected not-found, got %v", err)
	}
}

func lessonFixture(id int64) *entities.Lesson {
	return &entities.Lesson{
		ID:         id,
		DayOfWeek:  domain.Monday,
		TimeStart:  "09:00:00",
		TimeEnd:    "10:40:00",
		WeekType:   domain.WeekTypeAll,
		DateStart:  time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		DateEnd:    time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC),
		Discipline: &entities.Discipline{Name: "Матанализ"},
	}
}

func TestRenderFeed_TeacherLessonsAndEvents(t *testing.T) {
	repo := newFakeTokenRepo()
	tok := &entities.CalendarFeedToken{ID: 1, UserID: 42, Token: "feedtok"}
	repo.byUser[42] = tok
	repo.byToken["feedtok"] = tok

	lessons := &recordingLessonRepo{
		MockLessonRepository: NewMockLessonRepository(),
		byTeacher:            []*entities.Lesson{lessonFixture(7)},
	}
	eventsMock := &MockEventRepository{}
	evEnd := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC)
	eventsMock.On("GetByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]*entities.Event{{
			ID: 11, Title: "Педсовет", Status: entities.EventStatusScheduled,
			StartTime: time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC), EndTime: &evEnd,
			UpdatedAt: time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC),
		}}, nil)

	uc := NewCalendarFeedUseCase(repo, lessons, eventsMock, fakeGroupResolver{err: ErrStudentGroupNotFound}, feedCfg(),
		WithClock(fixedClock()))

	out, err := uc.RenderFeed(context.Background(), "feedtok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "UID:lesson-7@methodist") {
		t.Errorf("feed missing teacher lesson:\n%s", out)
	}
	if !strings.Contains(out, "UID:event-11@methodist") {
		t.Errorf("feed missing event:\n%s", out)
	}
	// Teacher path must scope by TeacherID and must not query by group.
	for _, f := range lessons.filters {
		if f.GroupID != nil {
			t.Error("must not query group lessons when the user has no group")
		}
	}
	eventsMock.AssertExpectations(t)
}

func TestRenderFeed_StudentGroupLessons(t *testing.T) {
	repo := newFakeTokenRepo()
	tok := &entities.CalendarFeedToken{ID: 1, UserID: 50, Token: "stok"}
	repo.byUser[50] = tok
	repo.byToken["stok"] = tok

	lessons := &recordingLessonRepo{
		MockLessonRepository: NewMockLessonRepository(),
		byGroup:              []*entities.Lesson{lessonFixture(21)},
	}
	eventsMock := &MockEventRepository{}
	eventsMock.On("GetByDateRange", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]*entities.Event{}, nil)

	uc := NewCalendarFeedUseCase(repo, lessons, eventsMock, fakeGroupResolver{groupID: 5}, feedCfg(),
		WithClock(fixedClock()))

	out, err := uc.RenderFeed(context.Background(), "stok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "UID:lesson-21@methodist") {
		t.Errorf("feed missing group lesson:\n%s", out)
	}
	sawGroupFilter := false
	for _, f := range lessons.filters {
		if f.GroupID != nil && *f.GroupID == 5 {
			sawGroupFilter = true
		}
	}
	if !sawGroupFilter {
		t.Error("expected a GetTimetable call scoped to the resolved group")
	}
}
