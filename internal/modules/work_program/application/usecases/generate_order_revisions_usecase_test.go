package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// fakeBulkRevisionGenerator is a deterministic RevisionDraftGenerator for
// the bulk-revision tests. It returns a fixed proposal, optionally erroring
// globally (err) or for a specific РПД title (errOnTitle) to exercise the
// best-effort path. It records the requests it received.
type fakeBulkRevisionGenerator struct {
	proposal   RevisionProposal
	err        error
	errOnTitle string
	requests   []RevisionDraftRequest
}

func (f *fakeBulkRevisionGenerator) GenerateRevision(
	_ context.Context, req RevisionDraftRequest,
) (RevisionProposal, error) {
	f.requests = append(f.requests, req)
	if f.err != nil {
		return RevisionProposal{}, f.err
	}
	if f.errOnTitle != "" && req.WorkProgramTitle == f.errOnTitle {
		return RevisionProposal{}, errors.New("generator failed for this РПД")
	}
	return f.proposal, nil
}

func okProposal() RevisionProposal {
	return RevisionProposal{
		ChangeType:    string(domain.RevisionChangeTypeHours),
		ChangeSummary: "Часы лекций приведены в соответствие приказу",
	}
}

func affectingOrder() *entities.MinobrnaukiOrder {
	return entities.ReconstituteMinobrnaukiOrder(entities.ReconstituteMinobrnaukiOrderInput{
		ID:          50,
		OrderNumber: "1234",
		Title:       "Об утверждении ФГОС ВО",
		PublishedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMajor,
		Summary:     "Изменены требования к часам по дисциплине",
		UploadedBy:  3,
		CreatedAt:   time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC),
	})
}

func TestGenerateOrderRevisionsUseCase_PanicsOnNilDeps(t *testing.T) {
	assert.Panics(t, func() {
		NewGenerateOrderRevisionsUseCase(nil, &fakeRevisionRepo{}, &fakeBulkRevisionGenerator{}, nil, nil)
	})
	assert.Panics(t, func() {
		NewGenerateOrderRevisionsUseCase(&fakeReadOrderRepo{}, nil, &fakeBulkRevisionGenerator{}, nil, nil)
	})
	assert.Panics(t, func() {
		NewGenerateOrderRevisionsUseCase(&fakeReadOrderRepo{}, &fakeRevisionRepo{}, nil, nil, nil)
	})
}

func TestGenerateOrderRevisionsUseCase_RoleGate(t *testing.T) {
	cases := []struct {
		role    string
		allowed bool
	}{
		{"methodist", true},
		{"academic_secretary", true},
		{"system_admin", true},
		{"teacher", false},
		{"student", false},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			orders := &fakeReadOrderRepo{order: affectingOrder(), affected: nil}
			targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{}}
			gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
			uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, &recordingAuditSink{})

			_, err := uc.Execute(context.Background(), 3, tc.role, 50)
			if tc.allowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, domain.ErrMinobrnaukiOrderScopeForbidden)
				assert.Zero(t, orders.getCalls, "forbidden caller must not reach the order load")
			}
		})
	}
}

func TestGenerateOrderRevisionsUseCase_RateLimited(t *testing.T) {
	orders := &fakeReadOrderRepo{order: affectingOrder()}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{}}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, &fakeRateLimiter{allowed: false}, nil)

	_, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.ErrorIs(t, err, domain.ErrGenerationRateLimited)
	assert.Zero(t, orders.getCalls, "rate-limited caller must not reach the order load")
}

func TestGenerateOrderRevisionsUseCase_OrderNotFound(t *testing.T) {
	orders := &fakeReadOrderRepo{getErr: repositories.ErrMinobrnaukiOrderNotFound}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{}}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, nil)

	_, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.ErrorIs(t, err, repositories.ErrMinobrnaukiOrderNotFound)
}

func TestGenerateOrderRevisionsUseCase_GeneratesDraftRevisions(t *testing.T) {
	const teacherID = int64(5)
	// Two affected РПД in revisable statuses, one in draft (skipped).
	approved := reconstituteWPWithStatus(t, 101, teacherID, domain.StatusApproved)
	needsRev := reconstituteWPWithStatus(t, 102, teacherID, domain.StatusNeedsRevision)
	draft := reconstituteWPWithStatus(t, 103, teacherID, domain.StatusDraft)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{101, 102, 103}}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{
		101: approved, 102: needsRev, 103: draft,
	}}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	audit := &recordingAuditSink{}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, allowingLimiter(), audit)

	res, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err)

	assert.Equal(t, 2, res.Generated, "two revisable РПД get a draft revision")
	assert.Equal(t, 1, res.Skipped, "the draft РПД has no approved edition to revise")
	assert.Equal(t, 0, res.Failures)

	// The generator ran once per revisable РПД, grounded on the order text.
	require.Len(t, gen.requests, 2)
	assert.Equal(t, "1234", gen.requests[0].OrderNumber)
	assert.Equal(t, "Изменены требования к часам по дисциплине", gen.requests[0].OrderSummary)

	// Each revisable РПД now carries exactly one draft revision authored by
	// the РПД author (teacher), NOT the triggering methodist.
	for _, wp := range []*entities.WorkProgram{approved, needsRev} {
		revs := wp.Revisions()
		require.Len(t, revs, 1)
		assert.Equal(t, domain.RevisionStatusDraft, revs[0].Status())
		assert.Equal(t, teacherID, revs[0].AuthorID(), "revision author is the РПД author, not the methodist")
		assert.Equal(t, domain.RevisionChangeTypeHours, revs[0].ChangeType())
	}
	assert.Empty(t, draft.Revisions(), "skipped draft РПД gets no revision")
	assert.Equal(t, []int64{101, 102}, targets.updateCalls, "only revisable РПД are persisted")
}

func TestGenerateOrderRevisionsUseCase_BestEffort_PersistFailureDoesNotAbort(t *testing.T) {
	const teacherID = int64(5)
	wpA := reconstituteWPWithStatus(t, 201, teacherID, domain.StatusApproved)
	wpB := reconstituteWPWithStatus(t, 202, teacherID, domain.StatusApproved)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{201, 202}}
	targets := &fakeRevisionRepo{
		programs:  map[int64]*entities.WorkProgram{201: wpA, 202: wpB},
		updateErr: map[int64]error{201: errors.New("db down for 201")},
	}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, nil)

	res, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err, "best-effort run never returns a hard error")
	assert.Equal(t, 1, res.Generated, "the second РПД still persists")
	assert.Equal(t, 1, res.Failures, "the persist failure is counted, not fatal")
}

func TestGenerateOrderRevisionsUseCase_BestEffort_LoadFailureDoesNotAbort(t *testing.T) {
	const teacherID = int64(5)
	wpB := reconstituteWPWithStatus(t, 302, teacherID, domain.StatusApproved)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{301, 302}}
	targets := &fakeRevisionRepo{
		programs: map[int64]*entities.WorkProgram{302: wpB},
		getErr:   map[int64]error{301: errors.New("load failed for 301")},
	}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, nil)

	res, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err)
	assert.Equal(t, 1, res.Generated)
	assert.Equal(t, 1, res.Failures)
}

func TestGenerateOrderRevisionsUseCase_BestEffort_GeneratorErrorCounted(t *testing.T) {
	const teacherID = int64(5)
	wpA := reconstituteWPWithStatus(t, 401, teacherID, domain.StatusApproved)
	wpB := reconstituteWPWithStatus(t, 402, teacherID, domain.StatusApproved)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{401, 402}}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{401: wpA, 402: wpB}}
	gen := &fakeBulkRevisionGenerator{err: errors.New("LLM unavailable")}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, nil)

	res, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err)
	assert.Equal(t, 0, res.Generated)
	assert.Equal(t, 2, res.Failures, "every generator error is counted, none persisted")
	assert.Empty(t, targets.updateCalls)
}

func TestGenerateOrderRevisionsUseCase_BestEffort_InvalidProposalCounted(t *testing.T) {
	const teacherID = int64(5)
	wpA := reconstituteWPWithStatus(t, 501, teacherID, domain.StatusApproved)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{501}}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{501: wpA}}
	// A malformed change_type fails NewRevision — must be counted, not panic.
	gen := &fakeBulkRevisionGenerator{proposal: RevisionProposal{
		ChangeType:    "not_a_real_change_type",
		ChangeSummary: "x",
	}}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, nil)

	res, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err)
	assert.Equal(t, 0, res.Generated)
	assert.Equal(t, 1, res.Failures)
	assert.Empty(t, wpA.Revisions(), "a rejected proposal adds no revision")
}

func TestGenerateOrderRevisionsUseCase_EmitsSummaryAudit(t *testing.T) {
	const teacherID = int64(5)
	wpA := reconstituteWPWithStatus(t, 601, teacherID, domain.StatusApproved)
	draft := reconstituteWPWithStatus(t, 602, teacherID, domain.StatusDraft)

	orders := &fakeReadOrderRepo{order: affectingOrder(), affected: []int64{601, 602}}
	targets := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{601: wpA, 602: draft}}
	gen := &fakeBulkRevisionGenerator{proposal: okProposal()}
	audit := &recordingAuditSink{}
	uc := NewGenerateOrderRevisionsUseCase(orders, targets, gen, nil, audit)

	_, err := uc.Execute(context.Background(), 3, "methodist", 50)
	require.NoError(t, err)

	require.NotEmpty(t, audit.events)
	last := audit.events[len(audit.events)-1]
	assert.Equal(t, "minobrnauki_order.revisions_generated", last.Action)
	assert.Equal(t, 1, last.Fields["generated"])
	assert.Equal(t, 1, last.Fields["skipped"])
	assert.Equal(t, 0, last.Fields["failures"])
}
