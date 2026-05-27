package entities_test

import (
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// --- Content-mutation methods: happy paths (draft) ---

func TestWorkProgram_AddGoal_OnDraft_Appends(t *testing.T) {
	wp := newDraft(t)
	g := mustGoal(t, "Освоить SQL", 0)
	if err := wp.AddGoal(g); err != nil {
		t.Fatalf("AddGoal: %v", err)
	}
	if len(wp.Goals()) != 1 || wp.Goals()[0] != g {
		t.Errorf("Goals: got %v, want [g]", wp.Goals())
	}
}

func TestWorkProgram_AddCompetence_OnDraft_Appends(t *testing.T) {
	wp := newDraft(t)
	c := mustCompetence(t, "ПК-3", domain.CompetenceTypePK, "Разработка СУБД")
	if err := wp.AddCompetence(c); err != nil {
		t.Fatalf("AddCompetence: %v", err)
	}
	if len(wp.Competences()) != 1 {
		t.Errorf("Competences len: got %d, want 1", len(wp.Competences()))
	}
}

func TestWorkProgram_AddTopic_OnDraft_Appends(t *testing.T) {
	wp := newDraft(t)
	tp := mustTopic(t, domain.TopicKindLecture, "Введение в реляционную модель", 4)
	if err := wp.AddTopic(tp); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}
	if len(wp.Topics()) != 1 {
		t.Errorf("Topics len: got %d, want 1", len(wp.Topics()))
	}
}

func TestWorkProgram_AddAssessment_OnDraft_Appends(t *testing.T) {
	wp := newDraft(t)
	a := mustAssessment(t, domain.AssessmentTypeCurrent, "Опрос", 5)
	if err := wp.AddAssessment(a); err != nil {
		t.Fatalf("AddAssessment: %v", err)
	}
	if len(wp.Assessments()) != 1 {
		t.Errorf("Assessments len: got %d, want 1", len(wp.Assessments()))
	}
}

func TestWorkProgram_AddReference_OnDraft_Appends(t *testing.T) {
	wp := newDraft(t)
	r := mustReference(t, domain.ReferenceKindMain, "Дейт К. Дж. Введение в системы баз данных")
	if err := wp.AddReference(r); err != nil {
		t.Fatalf("AddReference: %v", err)
	}
	if len(wp.References()) != 1 {
		t.Errorf("References len: got %d, want 1", len(wp.References()))
	}
}

// --- Content mutations also allowed in needs_revision ---

func TestWorkProgram_AddGoal_OnNeedsRevision_Appends(t *testing.T) {
	wp := approvedThenNeedsRevision(t)
	g := mustGoal(t, "Дополнительная цель", 0)
	if err := wp.AddGoal(g); err != nil {
		t.Fatalf("AddGoal on needs_revision: %v", err)
	}
	if len(wp.Goals()) != 1 {
		t.Errorf("Goals len: got %d, want 1", len(wp.Goals()))
	}
}

// --- Content mutations forbidden in frozen statuses ---

func TestWorkProgram_AddX_OnFrozenStatus_ReturnsErrCannotEditFrozenStatus(t *testing.T) {
	adders := []struct {
		name string
		add  func(t *testing.T, wp *entities.WorkProgram) error
	}{
		{
			name: "AddGoal",
			add: func(t *testing.T, wp *entities.WorkProgram) error {
				t.Helper()
				return wp.AddGoal(mustGoal(t, "цель", 0))
			},
		},
		{
			name: "AddCompetence",
			add: func(t *testing.T, wp *entities.WorkProgram) error {
				t.Helper()
				return wp.AddCompetence(mustCompetence(t, "ПК-1", domain.CompetenceTypePK, "д"))
			},
		},
		{
			name: "AddTopic",
			add: func(t *testing.T, wp *entities.WorkProgram) error {
				t.Helper()
				return wp.AddTopic(mustTopic(t, domain.TopicKindLecture, "тема", 2))
			},
		},
		{
			name: "AddAssessment",
			add: func(t *testing.T, wp *entities.WorkProgram) error {
				t.Helper()
				return wp.AddAssessment(mustAssessment(t, domain.AssessmentTypeCurrent, "д", 5))
			},
		},
		{
			name: "AddReference",
			add: func(t *testing.T, wp *entities.WorkProgram) error {
				t.Helper()
				return wp.AddReference(mustReference(t, domain.ReferenceKindMain, "Дейт"))
			},
		},
	}

	frozen := []struct {
		name  string
		setup func(t *testing.T, wp *entities.WorkProgram)
	}{
		{
			name: "pending_approval",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
			},
		},
		{
			name: "approved",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := wp.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
		{
			name: "archived",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Archive(); err != nil {
					t.Fatalf("setup Archive: %v", err)
				}
			},
		},
	}

	for _, a := range adders {
		for _, f := range frozen {
			t.Run(a.name+"/"+f.name, func(t *testing.T) {
				wp := newDraft(t)
				f.setup(t, wp)
				err := a.add(t, wp)
				if !errors.Is(err, domain.ErrCannotEditFrozenStatus) {
					t.Errorf("expected ErrCannotEditFrozenStatus, got %v", err)
				}
			})
		}
	}
}

// --- Nil-input rejection ---

func TestWorkProgram_AddX_NilInput_ReturnsInvariantError(t *testing.T) {
	cases := []struct {
		name string
		add  func(wp *entities.WorkProgram) error
	}{
		{"AddGoal", func(wp *entities.WorkProgram) error { return wp.AddGoal(nil) }},
		{"AddCompetence", func(wp *entities.WorkProgram) error { return wp.AddCompetence(nil) }},
		{"AddTopic", func(wp *entities.WorkProgram) error { return wp.AddTopic(nil) }},
		{"AddAssessment", func(wp *entities.WorkProgram) error { return wp.AddAssessment(nil) }},
		{"AddReference", func(wp *entities.WorkProgram) error { return wp.AddReference(nil) }},
		{"AddRevision", func(wp *entities.WorkProgram) error { return wp.AddRevision(nil) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wp := newDraft(t)
			err := c.add(wp)
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("expected ErrInvalidWorkProgram on nil input, got %v", err)
			}
		})
	}
}

// --- Competence code uniqueness ---

func TestWorkProgram_AddCompetence_DuplicateCode_Rejected(t *testing.T) {
	wp := newDraft(t)
	first := mustCompetence(t, "ПК-3", domain.CompetenceTypePK, "Первая")
	second := mustCompetence(t, "ПК-3", domain.CompetenceTypePK, "Вторая с тем же кодом")
	if err := wp.AddCompetence(first); err != nil {
		t.Fatalf("first AddCompetence: %v", err)
	}
	err := wp.AddCompetence(second)
	if !errors.Is(err, domain.ErrDuplicateCompetenceCode) {
		t.Errorf("expected ErrDuplicateCompetenceCode, got %v", err)
	}
	if len(wp.Competences()) != 1 {
		t.Errorf("Competences should not contain duplicate; len=%d", len(wp.Competences()))
	}
}

// --- Revision lifecycle: status gate ---

func TestWorkProgram_AddRevision_OnApproved_AppendsWithNumber1(t *testing.T) {
	wp := approved(t)
	rev := mustRevision(t, 1)
	if err := wp.AddRevision(rev); err != nil {
		t.Fatalf("AddRevision: %v", err)
	}
	if len(wp.Revisions()) != 1 {
		t.Errorf("Revisions len: got %d, want 1", len(wp.Revisions()))
	}
}

func TestWorkProgram_AddRevision_OnNeedsRevision_Appends(t *testing.T) {
	wp := approvedThenNeedsRevision(t)
	rev := mustRevision(t, 1)
	if err := wp.AddRevision(rev); err != nil {
		t.Fatalf("AddRevision on needs_revision: %v", err)
	}
	if len(wp.Revisions()) != 1 {
		t.Errorf("Revisions len: got %d, want 1", len(wp.Revisions()))
	}
}

func TestWorkProgram_AddRevision_OnDraft_Rejected(t *testing.T) {
	wp := newDraft(t)
	rev := mustRevision(t, 1)
	err := wp.AddRevision(rev)
	if !errors.Is(err, domain.ErrRevisionNotPermitted) {
		t.Errorf("expected ErrRevisionNotPermitted, got %v", err)
	}
}

func TestWorkProgram_AddRevision_OnPendingApproval_Rejected(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	err := wp.AddRevision(mustRevision(t, 1))
	if !errors.Is(err, domain.ErrRevisionNotPermitted) {
		t.Errorf("expected ErrRevisionNotPermitted, got %v", err)
	}
}

func TestWorkProgram_AddRevision_OnArchived_Rejected(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Archive(); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	err := wp.AddRevision(mustRevision(t, 1))
	if !errors.Is(err, domain.ErrRevisionNotPermitted) {
		t.Errorf("expected ErrRevisionNotPermitted, got %v", err)
	}
}

// --- Revision monotonic numbering ---

func TestWorkProgram_NextRevisionNumber_EmptyReturns1(t *testing.T) {
	wp := approved(t)
	if got := wp.NextRevisionNumber(); got != 1 {
		t.Errorf("NextRevisionNumber on empty: got %d, want 1", got)
	}
}

func TestWorkProgram_NextRevisionNumber_AfterAddIncrements(t *testing.T) {
	wp := approved(t)
	if err := wp.AddRevision(mustRevision(t, 1)); err != nil {
		t.Fatalf("AddRevision 1: %v", err)
	}
	if got := wp.NextRevisionNumber(); got != 2 {
		t.Errorf("NextRevisionNumber after 1 revision: got %d, want 2", got)
	}
	if err := wp.AddRevision(mustRevision(t, 2)); err != nil {
		t.Fatalf("AddRevision 2: %v", err)
	}
	if got := wp.NextRevisionNumber(); got != 3 {
		t.Errorf("NextRevisionNumber after 2 revisions: got %d, want 3", got)
	}
}

func TestWorkProgram_AddRevision_WrongNumber_Rejected(t *testing.T) {
	cases := []struct {
		name      string
		preCount  int // how many revisions to add first (with correct numbers)
		badNumber int
	}{
		{"first revision must be 1, got 2", 0, 2},
		{"first revision must be 1, got 5", 0, 5},
		{"second revision must be 2, got 1", 1, 1}, // duplicate
		{"second revision must be 2, got 3", 1, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := approved(t)
			for i := 1; i <= tc.preCount; i++ {
				if err := wp.AddRevision(mustRevision(t, i)); err != nil {
					t.Fatalf("setup AddRevision %d: %v", i, err)
				}
			}
			err := wp.AddRevision(mustRevision(t, tc.badNumber))
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("expected ErrInvalidWorkProgram, got %v", err)
			}
		})
	}
}

// --- Accessors return defensive copies ---

func TestWorkProgram_Accessors_ReturnDefensiveCopies(t *testing.T) {
	wp := newDraft(t)
	if err := wp.AddGoal(mustGoal(t, "цель", 0)); err != nil {
		t.Fatalf("AddGoal: %v", err)
	}
	got := wp.Goals()
	got[0] = nil // mutate the returned slice

	if len(wp.Goals()) != 1 || wp.Goals()[0] == nil {
		t.Error("Goals() must return defensive copy; internal mutation observed")
	}
}

// --- HoursTotal aggregation ---

func TestWorkProgram_HoursTotal_EmptyTopics_AllKindsZero(t *testing.T) {
	wp := newDraft(t)
	got := wp.HoursTotal()

	want := map[domain.TopicKind]int{
		domain.TopicKindLecture:   0,
		domain.TopicKindPractice:  0,
		domain.TopicKindLab:       0,
		domain.TopicKindSelfStudy: 0,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("HoursTotal[%s]: got %d, want %d", k, got[k], v)
		}
	}
	if len(got) != 4 {
		t.Errorf("HoursTotal should always carry all 4 kinds, got %d keys", len(got))
	}
}

func TestWorkProgram_HoursTotal_SumsWithinKind(t *testing.T) {
	wp := newDraft(t)
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindLecture, "Лекция 1", 4)); err != nil {
		t.Fatalf("AddTopic 1: %v", err)
	}
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindLecture, "Лекция 2", 6)); err != nil {
		t.Fatalf("AddTopic 2: %v", err)
	}
	got := wp.HoursTotal()
	if got[domain.TopicKindLecture] != 10 {
		t.Errorf("HoursTotal[lecture]: got %d, want 10", got[domain.TopicKindLecture])
	}
}

func TestWorkProgram_HoursTotal_SumsIndependentlyPerKind(t *testing.T) {
	wp := newDraft(t)
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindLecture, "Л", 4)); err != nil {
		t.Fatalf("AddTopic lecture: %v", err)
	}
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindPractice, "П", 2)); err != nil {
		t.Fatalf("AddTopic practice: %v", err)
	}
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindLab, "Лаб", 6)); err != nil {
		t.Fatalf("AddTopic lab: %v", err)
	}
	if err := wp.AddTopic(mustTopic(t, domain.TopicKindSelfStudy, "СРС", 36)); err != nil {
		t.Fatalf("AddTopic self_study: %v", err)
	}
	got := wp.HoursTotal()
	want := map[domain.TopicKind]int{
		domain.TopicKindLecture:   4,
		domain.TopicKindPractice:  2,
		domain.TopicKindLab:       6,
		domain.TopicKindSelfStudy: 36,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("HoursTotal[%s]: got %d, want %d", k, got[k], v)
		}
	}
}

// --- Test helpers ---

func approved(t *testing.T) *entities.WorkProgram {
	t.Helper()
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("approved: Submit: %v", err)
	}
	if err := wp.Approve(99); err != nil {
		t.Fatalf("approved: Approve: %v", err)
	}
	return wp
}

func approvedThenNeedsRevision(t *testing.T) *entities.WorkProgram {
	t.Helper()
	wp := approved(t)
	if err := wp.MarkNeedsRevision(); err != nil {
		t.Fatalf("approvedThenNeedsRevision: MarkNeedsRevision: %v", err)
	}
	return wp
}

func mustGoal(t *testing.T, text string, order int) *entities.Goal {
	t.Helper()
	g, err := entities.NewGoal(text, order)
	if err != nil {
		t.Fatalf("mustGoal: %v", err)
	}
	return g
}

func mustCompetence(t *testing.T, code string, ct domain.CompetenceType, desc string) *entities.Competence {
	t.Helper()
	c, err := entities.NewCompetence(code, ct, desc)
	if err != nil {
		t.Fatalf("mustCompetence: %v", err)
	}
	return c
}

func mustTopic(t *testing.T, kind domain.TopicKind, title string, hours int) *entities.Topic {
	t.Helper()
	tp, err := entities.NewTopic(entities.NewTopicInput{
		Kind:  kind,
		Title: title,
		Hours: hours,
	})
	if err != nil {
		t.Fatalf("mustTopic: %v", err)
	}
	return tp
}

func mustAssessment(t *testing.T, at domain.AssessmentType, desc string, score int) *entities.AssessmentCriterion {
	t.Helper()
	a, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type:        at,
		Description: desc,
		MaxScore:    score,
	})
	if err != nil {
		t.Fatalf("mustAssessment: %v", err)
	}
	return a
}

func mustReference(t *testing.T, kind domain.ReferenceKind, citation string) *entities.Reference {
	t.Helper()
	r, err := entities.NewReference(entities.NewReferenceInput{
		Kind:     kind,
		Citation: citation,
	})
	if err != nil {
		t.Fatalf("mustReference: %v", err)
	}
	return r
}

func mustRevision(t *testing.T, number int) *entities.Revision {
	t.Helper()
	r, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: number,
		ChangeType:     domain.RevisionChangeTypeOther,
		ChangeSummary:  "Прочие правки",
		AuthorID:       7,
	})
	if err != nil {
		t.Fatalf("mustRevision: %v", err)
	}
	return r
}
