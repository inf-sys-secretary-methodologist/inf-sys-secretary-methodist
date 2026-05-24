package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/domain/repositories"
)

// ===== Fake repository — implements the full EventRepository surface =====

type fakeRepo struct {
	saveCalls         int
	saveErr           error
	saveAssignID      int64
	updateCalls       int
	updateErr         error
	deleteCalls       int
	deleteErr         error
	getByIDErr        error
	getByIDResult     *entities.ExtracurricularEvent
	listErr           error
	listResult        repositories.EventListResult
	listFilterCapture repositories.EventListFilter
	addPartErr        error
	addPartCalls      int
	addPartCapture    struct {
		eventID, userID int64
		at              time.Time
	}
	removePartErr     error
	removePartCalls   int
	removePartCapture struct{ eventID, userID int64 }
}

func (r *fakeRepo) Save(ctx context.Context, e *entities.ExtracurricularEvent) error {
	r.saveCalls++
	if r.saveErr != nil {
		return r.saveErr
	}
	if r.saveAssignID != 0 {
		e.ID = r.saveAssignID
	}
	return nil
}
func (r *fakeRepo) GetByID(_ context.Context, _ int64) (*entities.ExtracurricularEvent, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	return r.getByIDResult, nil
}
func (r *fakeRepo) Update(_ context.Context, _ *entities.ExtracurricularEvent) error {
	r.updateCalls++
	return r.updateErr
}
func (r *fakeRepo) Delete(_ context.Context, _ int64) error {
	r.deleteCalls++
	return r.deleteErr
}
func (r *fakeRepo) List(_ context.Context, f repositories.EventListFilter) (repositories.EventListResult, error) {
	r.listFilterCapture = f
	if r.listErr != nil {
		return repositories.EventListResult{}, r.listErr
	}
	return r.listResult, nil
}
func (r *fakeRepo) AddParticipant(_ context.Context, eventID, userID int64, at time.Time) error {
	r.addPartCalls++
	r.addPartCapture.eventID = eventID
	r.addPartCapture.userID = userID
	r.addPartCapture.at = at
	return r.addPartErr
}
func (r *fakeRepo) RemoveParticipant(_ context.Context, eventID, userID int64) error {
	r.removePartCalls++
	r.removePartCapture.eventID = eventID
	r.removePartCapture.userID = userID
	return r.removePartErr
}

// recordingAudit captures every emitted event for assertion.
type recordingAudit struct {
	events []auditCapture
}
type auditCapture struct {
	action   string
	resource string
	fields   map[string]any
}

func (r *recordingAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	r.events = append(r.events, auditCapture{action: action, resource: resource, fields: fields})
}

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func validCreateInput() CreateEventInput {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	return CreateEventInput{
		Title:          "Концерт",
		Description:    "desc",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		Location:       "loc",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
	}
}

func newSampleEvent(t *testing.T, organizerID int64, status entities.Status) *entities.ExtracurricularEvent {
	t.Helper()
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, err := entities.NewExtracurricularEvent(entities.NewExtracurricularEventParams{
		Title:          "Концерт",
		Description:    "desc",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		Location:       "loc",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
		OrganizerID:    organizerID,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("setup NewExtracurricularEvent: %v", err)
	}
	if status == entities.StatusPublished {
		if err := e.Publish(now); err != nil {
			t.Fatalf("setup Publish: %v", err)
		}
	}
	e.ID = 99
	return e
}

// ===== CreateEvent =====

func TestCreateEventUseCase_HappyPath(t *testing.T) {
	repo := &fakeRepo{saveAssignID: 99}
	audit := &recordingAudit{}
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	uc := NewCreateEventUseCase(repo, audit, fixedClock(now))
	e, err := uc.Execute(context.Background(), 42, "methodist", false, validCreateInput())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if e == nil || e.ID != 99 {
		t.Fatalf("event.ID = %v, want 99", e)
	}
	if e.OrganizerID() != 42 {
		t.Errorf("OrganizerID = %d, want 42 (actor)", e.OrganizerID())
	}
	if repo.saveCalls != 1 {
		t.Errorf("saveCalls = %d, want 1", repo.saveCalls)
	}
	if len(audit.events) == 0 || audit.events[0].action != "extracurricular.event_created" {
		t.Errorf("audit events = %+v, want extracurricular.event_created", audit.events)
	}
}

func TestCreateEventUseCase_DeniedForTeacher(t *testing.T) {
	repo := &fakeRepo{saveAssignID: 99}
	audit := &recordingAudit{}
	uc := NewCreateEventUseCase(repo, audit, fixedClock(time.Now()))
	_, err := uc.Execute(context.Background(), 42, "teacher", false, validCreateInput())
	if err == nil {
		t.Fatal("expected denial, got nil")
	}
	if !errors.Is(err, entities.ErrEventScopeForbidden) {
		t.Errorf("err = %v, want ErrEventScopeForbidden", err)
	}
	if repo.saveCalls != 0 {
		t.Errorf("saveCalls = %d, want 0 (denied before persistence)", repo.saveCalls)
	}
	hasDenial := false
	for _, ev := range audit.events {
		if ev.action == "extracurricular.event_create_denied" {
			hasDenial = true
			if ev.fields["code"] != "forbidden" {
				t.Errorf("denial code = %v, want forbidden", ev.fields["code"])
			}
		}
	}
	if !hasDenial {
		t.Errorf("missing denial audit event, got: %+v", audit.events)
	}
}

func TestCreateEventUseCase_InvalidInput(t *testing.T) {
	repo := &fakeRepo{}
	audit := &recordingAudit{}
	uc := NewCreateEventUseCase(repo, audit, fixedClock(time.Now()))
	in := validCreateInput()
	in.Title = ""
	_, err := uc.Execute(context.Background(), 42, "methodist", false, in)
	if !errors.Is(err, entities.ErrInvalidEvent) {
		t.Errorf("err = %v, want ErrInvalidEvent", err)
	}
	if repo.saveCalls != 0 {
		t.Errorf("saveCalls = %d, want 0", repo.saveCalls)
	}
}

// ===== UpdateEvent =====

func TestUpdateEventUseCase_HappyPath(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusDraft)
	repo := &fakeRepo{getByIDResult: existing}
	audit := &recordingAudit{}
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC)
	uc := NewUpdateEventUseCase(repo, audit, nil, fixedClock(now))
	in := UpdateEventInput{
		ID:             99,
		Title:          "Обновлённый",
		Description:    "new desc",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		Location:       "new loc",
		StartAt:        now.Add(48 * time.Hour),
		EndAt:          now.Add(50 * time.Hour),
	}
	_, err := uc.Execute(context.Background(), 42, "methodist", false, in)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.updateCalls != 1 {
		t.Errorf("updateCalls = %d, want 1", repo.updateCalls)
	}
	if existing.Title() != "Обновлённый" {
		t.Errorf("entity not mutated; Title = %q", existing.Title())
	}
}

func TestUpdateEventUseCase_DeniedForOtherMethodist(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusDraft)
	repo := &fakeRepo{getByIDResult: existing}
	audit := &recordingAudit{}
	uc := NewUpdateEventUseCase(repo, audit, nil, fixedClock(time.Now()))
	in := UpdateEventInput{
		ID:             99,
		Title:          "x",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		StartAt:        time.Now().Add(48 * time.Hour),
		EndAt:          time.Now().Add(50 * time.Hour),
	}
	_, err := uc.Execute(context.Background(), 100, "methodist", false, in)
	if !errors.Is(err, entities.ErrEventScopeForbidden) {
		t.Errorf("err = %v, want ErrEventScopeForbidden", err)
	}
}

func TestUpdateEventUseCase_VersionConflictPropagates(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusDraft)
	repo := &fakeRepo{getByIDResult: existing, updateErr: repositories.ErrEventVersionConflict}
	uc := NewUpdateEventUseCase(repo, nil, nil, fixedClock(time.Now()))
	in := UpdateEventInput{
		ID:             99,
		Title:          "x",
		Category:       entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceAll,
		StartAt:        time.Now().Add(48 * time.Hour),
		EndAt:          time.Now().Add(50 * time.Hour),
	}
	_, err := uc.Execute(context.Background(), 42, "methodist", false, in)
	if !errors.Is(err, repositories.ErrEventVersionConflict) {
		t.Errorf("err = %v, want ErrEventVersionConflict", err)
	}
}

// ===== DeleteEvent =====

func TestDeleteEventUseCase_HappyPath(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusDraft)
	repo := &fakeRepo{getByIDResult: existing}
	audit := &recordingAudit{}
	uc := NewDeleteEventUseCase(repo, audit)
	if err := uc.Execute(context.Background(), 42, "methodist", false, 99); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Errorf("deleteCalls = %d, want 1", repo.deleteCalls)
	}
}

func TestDeleteEventUseCase_NotFound(t *testing.T) {
	repo := &fakeRepo{getByIDErr: repositories.ErrEventNotFound}
	uc := NewDeleteEventUseCase(repo, nil)
	err := uc.Execute(context.Background(), 42, "methodist", false, 404)
	if !errors.Is(err, repositories.ErrEventNotFound) {
		t.Errorf("err = %v, want ErrEventNotFound", err)
	}
}

// ===== GetEvent =====

func TestGetEventUseCase_AdminViewsAnyAudience(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusPublished)
	// even staff-only event visible к admin via isAdmin flag
	repo := &fakeRepo{getByIDResult: existing}
	uc := NewGetEventUseCase(repo)
	e, err := uc.Execute(context.Background(), "system_admin", true, 99)
	if err != nil || e == nil {
		t.Fatalf("admin Execute: err=%v e=%v", err, e)
	}
}

func TestGetEventUseCase_StudentCannotViewTeachersAudience(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	e, _ := entities.NewExtracurricularEvent(entities.NewExtracurricularEventParams{
		Title: "x", Category: entities.CategoryCultural,
		TargetAudience: entities.TargetAudienceTeachers,
		StartAt:        now.Add(48 * time.Hour), EndAt: now.Add(50 * time.Hour),
		OrganizerID: 42, Now: now,
	})
	e.ID = 99
	repo := &fakeRepo{getByIDResult: e}
	uc := NewGetEventUseCase(repo)
	_, err := uc.Execute(context.Background(), "student", false, 99)
	if !errors.Is(err, repositories.ErrEventNotFound) {
		t.Errorf("err = %v, want ErrEventNotFound (hide existence)", err)
	}
}

// ===== ListEvents =====

func TestListEventsUseCase_StudentSeesAllAndStudentsAudiences(t *testing.T) {
	repo := &fakeRepo{listResult: repositories.EventListResult{Total: 3}}
	uc := NewListEventsUseCase(repo)
	_, err := uc.Execute(context.Background(), "student", false, ListEventsInput{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := repo.listFilterCapture.AudienceIn
	if !containsAll(got, []string{"all", "students"}) {
		t.Errorf("AudienceIn = %v, want subset {all, students}", got)
	}
}

func TestListEventsUseCase_AdminSeesAllAudiences(t *testing.T) {
	repo := &fakeRepo{listResult: repositories.EventListResult{Total: 7}}
	uc := NewListEventsUseCase(repo)
	_, err := uc.Execute(context.Background(), "system_admin", true, ListEventsInput{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(repo.listFilterCapture.AudienceIn) != 0 {
		t.Errorf("admin AudienceIn = %v, want empty (no audience filter)", repo.listFilterCapture.AudienceIn)
	}
}

func containsAll(s []string, want []string) bool {
	m := map[string]bool{}
	for _, v := range s {
		m[v] = true
	}
	for _, w := range want {
		if !m[w] {
			return false
		}
	}
	return true
}

// ===== RegisterParticipant =====

func TestRegisterParticipantUseCase_HappyPath(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusPublished)
	repo := &fakeRepo{getByIDResult: existing}
	audit := &recordingAudit{}
	now := time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC)
	uc := NewRegisterParticipantUseCase(repo, audit, fixedClock(now))
	if err := uc.Execute(context.Background(), 101, 99); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.addPartCalls != 1 {
		t.Fatalf("addPartCalls = %d, want 1", repo.addPartCalls)
	}
	if repo.addPartCapture.eventID != 99 || repo.addPartCapture.userID != 101 {
		t.Errorf("addPart capture wrong: %+v", repo.addPartCapture)
	}
	if !repo.addPartCapture.at.Equal(now) {
		t.Errorf("registeredAt = %v, want clock %v", repo.addPartCapture.at, now)
	}
}

func TestRegisterParticipantUseCase_EventNotFound(t *testing.T) {
	repo := &fakeRepo{getByIDErr: repositories.ErrEventNotFound}
	uc := NewRegisterParticipantUseCase(repo, nil, fixedClock(time.Now()))
	err := uc.Execute(context.Background(), 101, 404)
	if !errors.Is(err, repositories.ErrEventNotFound) {
		t.Errorf("err = %v, want ErrEventNotFound", err)
	}
}

func TestRegisterParticipantUseCase_RejectsDraftStatus(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusDraft)
	repo := &fakeRepo{getByIDResult: existing}
	uc := NewRegisterParticipantUseCase(repo, nil, fixedClock(time.Now()))
	err := uc.Execute(context.Background(), 101, 99)
	if !errors.Is(err, entities.ErrEventNotOpenForRegistration) {
		t.Errorf("err = %v, want ErrEventNotOpenForRegistration", err)
	}
}

// ===== UnregisterParticipant =====

func TestUnregisterParticipantUseCase_HappyPath(t *testing.T) {
	existing := newSampleEvent(t, 42, entities.StatusPublished)
	repo := &fakeRepo{getByIDResult: existing}
	audit := &recordingAudit{}
	uc := NewUnregisterParticipantUseCase(repo, audit)
	if err := uc.Execute(context.Background(), 101, 99); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.removePartCalls != 1 {
		t.Errorf("removePartCalls = %d, want 1", repo.removePartCalls)
	}
}

func TestUnregisterParticipantUseCase_NotFound(t *testing.T) {
	repo := &fakeRepo{getByIDErr: repositories.ErrEventNotFound}
	uc := NewUnregisterParticipantUseCase(repo, nil)
	err := uc.Execute(context.Background(), 101, 404)
	if !errors.Is(err, repositories.ErrEventNotFound) {
		t.Errorf("err = %v, want ErrEventNotFound", err)
	}
}
