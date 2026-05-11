package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	assignmentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	assignmentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	curriculumEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	curriculumRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	documentEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	documentRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
)

// --- fakes --------------------------------------------------------------

// callTracker pins the order in which the use case fans out across its
// dependencies. Reorder regressions (e.g. audit emitted BEFORE render
// completes) would otherwise slip past per-fake recording assertions.
type callTracker struct {
	calls []string
}

func (c *callTracker) record(name string) {
	if c == nil {
		return
	}
	c.calls = append(c.calls, name)
}

type fakeCurriculumAggRepo struct {
	tracker *callTracker
	gotYear int
	result  []curriculumRepos.CurriculumYearSpecialtyAgg
	err     error
}

func (f *fakeCurriculumAggRepo) AggregateByYearSpecialty(_ context.Context, year int) ([]curriculumRepos.CurriculumYearSpecialtyAgg, error) {
	f.tracker.record("curriculum")
	f.gotYear = year
	return f.result, f.err
}

type fakeAssignmentAggRepo struct {
	tracker        *callTracker
	gotFrom, gotTo time.Time
	result         []assignmentRepos.AssignmentGradeDistributionAgg
	err            error
}

func (f *fakeAssignmentAggRepo) AggregateGradeDistribution(_ context.Context, from, to time.Time) ([]assignmentRepos.AssignmentGradeDistributionAgg, error) {
	f.tracker.record("assignment")
	f.gotFrom, f.gotTo = from, to
	return f.result, f.err
}

type fakeItemAggRepo struct {
	tracker *callTracker
	gotYear int
	result  []curriculumRepos.DisciplineItemHoursAgg
	err     error
}

func (f *fakeItemAggRepo) AggregateHoursByYear(_ context.Context, year int) ([]curriculumRepos.DisciplineItemHoursAgg, error) {
	f.tracker.record("item")
	f.gotYear = year
	return f.result, f.err
}

type fakeDocumentAggRepo struct {
	tracker        *callTracker
	gotFrom, gotTo time.Time
	result         []documentRepos.DocumentActivityByTypeAgg
	err            error
}

func (f *fakeDocumentAggRepo) AggregateActivityByType(_ context.Context, from, to time.Time) ([]documentRepos.DocumentActivityByTypeAgg, error) {
	f.tracker.record("document")
	f.gotFrom, f.gotTo = from, to
	return f.result, f.err
}

type fakeRenderer struct {
	tracker      *callTracker
	gotYear      int
	gotCurricula []curriculumRepos.CurriculumYearSpecialtyAgg
	gotGrades    []assignmentRepos.AssignmentGradeDistributionAgg
	gotHours     []curriculumRepos.DisciplineItemHoursAgg
	gotActivity  []documentRepos.DocumentActivityByTypeAgg
	result       []byte
	err          error
}

func (f *fakeRenderer) RenderAnnualReport(
	year int,
	curricula []curriculumRepos.CurriculumYearSpecialtyAgg,
	grades []assignmentRepos.AssignmentGradeDistributionAgg,
	hours []curriculumRepos.DisciplineItemHoursAgg,
	activity []documentRepos.DocumentActivityByTypeAgg,
) ([]byte, error) {
	f.tracker.record("render")
	f.gotYear = year
	f.gotCurricula = curricula
	f.gotGrades = grades
	f.gotHours = hours
	f.gotActivity = activity
	return f.result, f.err
}

type auditCall struct {
	action   string
	resource string
	fields   map[string]any
}

type fakeAuditSink struct {
	tracker *callTracker
	calls   []auditCall
}

func (f *fakeAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	if f.tracker != nil {
		f.tracker.record("audit")
	}
	f.calls = append(f.calls, auditCall{action: action, resource: resource, fields: fields})
}

// --- harness ------------------------------------------------------------

type harness struct {
	cur     *fakeCurriculumAggRepo
	assign  *fakeAssignmentAggRepo
	item    *fakeItemAggRepo
	doc     *fakeDocumentAggRepo
	render  *fakeRenderer
	audit   *fakeAuditSink
	tracker *callTracker
	uc      *usecases.AnnualReportUseCase
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	tracker := &callTracker{}
	h := &harness{
		tracker: tracker,
		cur:     &fakeCurriculumAggRepo{tracker: tracker},
		assign:  &fakeAssignmentAggRepo{tracker: tracker},
		item:    &fakeItemAggRepo{tracker: tracker},
		doc:     &fakeDocumentAggRepo{tracker: tracker},
		render:  &fakeRenderer{tracker: tracker, result: []byte{0x50, 0x4B, 0x03, 0x04, 'D', 'O', 'C', 'X'}},
		audit:   &fakeAuditSink{tracker: tracker},
	}
	h.uc = usecases.NewAnnualReportUseCase(h.cur, h.assign, h.item, h.doc, h.render, h.audit)
	return h
}

// --- tests --------------------------------------------------------------

func TestAnnualReportUseCase_Generate_HappyPath(t *testing.T) {
	h := newHarness(t)
	h.cur.result = []curriculumRepos.CurriculumYearSpecialtyAgg{
		{Specialty: "Информатика и вычислительная техника", Status: curriculumEntities.StatusApproved, Count: 3},
	}
	h.assign.result = []assignmentRepos.AssignmentGradeDistributionAgg{
		{Subject: "Алгоритмы", Status: assignmentEntities.StatusGraded, Count: 12},
	}
	h.item.result = []curriculumRepos.DisciplineItemHoursAgg{
		{CurriculumID: 1, CurriculumTitle: "ИВТ-2026", Lectures: 64, Practice: 32, Lab: 16, SelfStudy: 88},
	}
	h.doc.result = []documentRepos.DocumentActivityByTypeAgg{
		{TypeName: "Приказ", Status: documentEntities.DocumentStatusApproved, Count: 5},
	}

	bytes, err := h.uc.Generate(context.Background(), usecases.GenerateAnnualReportInput{Year: 2026, ActorID: 42})
	require.NoError(t, err)
	require.Equal(t, []byte{0x50, 0x4B, 0x03, 0x04, 'D', 'O', 'C', 'X'}, bytes)

	require.Equal(t, 2026, h.cur.gotYear)
	require.Equal(t, 2026, h.item.gotYear)
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), h.assign.gotFrom)
	require.Equal(t, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), h.assign.gotTo)
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), h.doc.gotFrom)
	require.Equal(t, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), h.doc.gotTo)

	require.Equal(t, 2026, h.render.gotYear)
	require.Len(t, h.render.gotCurricula, 1)
	require.Len(t, h.render.gotGrades, 1)
	require.Len(t, h.render.gotHours, 1)
	require.Len(t, h.render.gotActivity, 1)

	require.Len(t, h.audit.calls, 1)
	require.Equal(t, "report.annual_generated", h.audit.calls[0].action)
	require.Equal(t, "report", h.audit.calls[0].resource)
	require.Equal(t, 2026, h.audit.calls[0].fields["year"])
	require.Equal(t, int64(42), h.audit.calls[0].fields["actor_user_id"])

	// Order matters: aggregates fan out before render; audit fires only
	// after a successful render (forensic trail tracks completed
	// generations, не failed attempts).
	require.Equal(t, []string{"curriculum", "assignment", "item", "document", "render", "audit"}, h.tracker.calls)
}

func TestAnnualReportUseCase_Generate_RepoErrorsPropagate(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*harness)
	}{
		{
			name:  "curriculum repo error",
			setup: func(h *harness) { h.cur.err = errors.New("boom curriculum") },
		},
		{
			name:  "assignment repo error",
			setup: func(h *harness) { h.assign.err = errors.New("boom assignment") },
		},
		{
			name:  "discipline_item repo error",
			setup: func(h *harness) { h.item.err = errors.New("boom item") },
		},
		{
			name:  "document repo error",
			setup: func(h *harness) { h.doc.err = errors.New("boom doc") },
		},
		{
			name:  "renderer error",
			setup: func(h *harness) { h.render.err = errors.New("boom render") },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			tc.setup(h)

			bytes, err := h.uc.Generate(context.Background(), usecases.GenerateAnnualReportInput{Year: 2026, ActorID: 42})
			require.Error(t, err)
			require.Nil(t, bytes)
			require.Empty(t, h.audit.calls, "audit must NOT fire when pipeline fails before render completes")
		})
	}
}

func TestAnnualReportUseCase_Generate_NilAuditSink_DoesNotPanic(t *testing.T) {
	cur := &fakeCurriculumAggRepo{}
	assign := &fakeAssignmentAggRepo{}
	item := &fakeItemAggRepo{}
	doc := &fakeDocumentAggRepo{}
	render := &fakeRenderer{result: []byte{0x50, 0x4B}}
	uc := usecases.NewAnnualReportUseCase(cur, assign, item, doc, render, nil)

	bytes, err := uc.Generate(context.Background(), usecases.GenerateAnnualReportInput{Year: 2026, ActorID: 42})
	require.NoError(t, err)
	require.Equal(t, []byte{0x50, 0x4B}, bytes)
}
