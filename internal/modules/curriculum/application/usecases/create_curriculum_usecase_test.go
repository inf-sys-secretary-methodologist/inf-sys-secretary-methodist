package usecases

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// fakeCreateRepo is a minimal CurriculumRepository test double covering
// only Save (the use case under test makes no other calls).
type fakeCreateRepo struct {
	saveCalls  int
	saved      *entities.Curriculum
	saveErr    error
	idAssigned int64
}

func (f *fakeCreateRepo) Save(ctx context.Context, c *entities.Curriculum) error {
	f.saveCalls++
	f.saved = c
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.idAssigned > 0 {
		c.ID = f.idAssigned
	}
	return nil
}

// recordingAuditSink captures audit calls without touching real logging.
type recordingAuditSink struct {
	events []auditCall
}

type auditCall struct {
	Action   string
	Resource string
	Fields   map[string]any
}

func (r *recordingAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	// Defensive copy — production code may mutate the map after dispatch
	// (it doesn't, but pinning the snapshot here protects future maintainers).
	cp := make(map[string]any, len(fields))
	maps.Copy(cp, fields)
	r.events = append(r.events, auditCall{Action: action, Resource: resource, Fields: cp})
}

func TestNewCreateCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewCreateCurriculumUseCase(nil, ...) did not panic")
		}
	}()
	NewCreateCurriculumUseCase(nil, &recordingAuditSink{}, time.Now)
}

func TestCreateCurriculumUseCase_HappyPath(t *testing.T) {
	repo := &fakeCreateRepo{idAssigned: 42}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	uc := NewCreateCurriculumUseCase(repo, audit, func() time.Time { return frozenNow })
	c, err := uc.Execute(context.Background(), 7, CreateCurriculumInput{
		Title:       "ИВТ-2026",
		Code:        "09.03.04-2026",
		Specialty:   "Информатика",
		Year:        2026,
		Description: "desc",
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, int64(42), c.ID)
	assert.Equal(t, int64(7), c.CreatedBy())
	assert.Equal(t, entities.StatusDraft, c.Status())
	assert.Equal(t, frozenNow, c.CreatedAt())

	require.Equal(t, 1, repo.saveCalls)
	require.NotNil(t, repo.saved)
	assert.Equal(t, "ИВТ-2026", repo.saved.Title())
	assert.Equal(t, "09.03.04-2026", repo.saved.Code())

	require.Len(t, audit.events, 1, "one audit event expected")
	ev := audit.events[0]
	assert.Equal(t, "curriculum.created", ev.Action)
	assert.Equal(t, "curriculum", ev.Resource)
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(42), ev.Fields["curriculum_id"])
	assert.Equal(t, "09.03.04-2026", ev.Fields["code"])
	assert.Equal(t, 2026, ev.Fields["year"])
}

func TestCreateCurriculumUseCase_InvalidInputAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeCreateRepo{}
	audit := &recordingAuditSink{}
	uc := NewCreateCurriculumUseCase(repo, audit, time.Now)

	// Empty title violates the entity invariant.
	_, err := uc.Execute(context.Background(), 7, CreateCurriculumInput{
		Title:     "",
		Code:      "09.03.04-2026",
		Specialty: "Информатика",
		Year:      2026,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidCurriculum),
		"expected ErrInvalidCurriculum, got %v", err)

	// Repo never called — input never reached storage.
	assert.Zero(t, repo.saveCalls,
		"repo.Save must not be called on invariant failure")

	// Denial audit: forensic trail captures actor + reason without
	// leaking PII (the original validation message is kept opaque to
	// the caller; the audit field 'reason' is enough for forensics).
	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.create_denied", ev.Action)
	assert.Equal(t, "curriculum", ev.Resource)
	assert.Equal(t, int64(7), ev.Fields["actor_user_id"])
	assert.Equal(t, "invalid", ev.Fields["reason"])
}

func TestCreateCurriculumUseCase_CodeConflictAuditsDenialAndReturnsSentinel(t *testing.T) {
	repo := &fakeCreateRepo{saveErr: repositories.ErrCurriculumCodeExists}
	audit := &recordingAuditSink{}
	uc := NewCreateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 7, CreateCurriculumInput{
		Title:     "T",
		Code:      "DUP-2026",
		Specialty: "S",
		Year:      2026,
	})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumCodeExists),
		"expected ErrCurriculumCodeExists, got %v", err)
	assert.Equal(t, 1, repo.saveCalls, "repo.Save should be attempted exactly once")

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "curriculum.create_denied", ev.Action)
	assert.Equal(t, "code_conflict", ev.Fields["reason"])
	assert.Equal(t, "DUP-2026", ev.Fields["code"])
}

func TestCreateCurriculumUseCase_OtherSaveErrorPropagatesWithoutSuccessAudit(t *testing.T) {
	repo := &fakeCreateRepo{saveErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}
	uc := NewCreateCurriculumUseCase(repo, audit, time.Now)

	_, err := uc.Execute(context.Background(), 7, CreateCurriculumInput{
		Title:     "T",
		Code:      "OK-2026",
		Specialty: "S",
		Year:      2026,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")

	// On a non-domain transport failure we do NOT emit the success audit
	// (the row is not persisted) and we do NOT emit the create_denied
	// audit either — that sentinel is reserved for *domain* denials so
	// the audit log doesn't conflate infrastructure outages with policy
	// rejections. Operators read transport failures from the logger
	// stack trace instead.
	assert.Empty(t, audit.events,
		"transport errors must not produce a created/denied audit event")
}

func TestCreateCurriculumUseCase_NilSinkIsTolerated(t *testing.T) {
	repo := &fakeCreateRepo{idAssigned: 1}
	uc := NewCreateCurriculumUseCase(repo, nil, time.Now)

	c, err := uc.Execute(context.Background(), 7, CreateCurriculumInput{
		Title:     "T",
		Code:      "OK-2026",
		Specialty: "S",
		Year:      2026,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), c.ID)
}
