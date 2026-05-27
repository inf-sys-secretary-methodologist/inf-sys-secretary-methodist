package entities_test

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Reconstitute* are repository-side constructors that bypass invariant
// checks (DB CHECK constraints + the original NewX call already
// validated). The tests verify exact field preservation; semantic
// invariants are exercised by NewX tests in other files.

func TestReconstituteGoal_PreservesAllFields(t *testing.T) {
	ts := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	g := entities.ReconstituteGoal(entities.ReconstituteGoalInput{
		ID:            42,
		WorkProgramID: 7,
		Text:          "Освоить SQL",
		OrderIndex:    3,
		CreatedAt:     ts,
	})
	if g == nil {
		t.Fatal("expected non-nil Goal")
	}
	if g.ID() != 42 || g.WorkProgramID() != 7 || g.Text() != "Освоить SQL" ||
		g.OrderIndex() != 3 || !g.CreatedAt().Equal(ts) {
		t.Errorf("ReconstituteGoal field mismatch: id=%d wpid=%d text=%q order=%d createdAt=%v",
			g.ID(), g.WorkProgramID(), g.Text(), g.OrderIndex(), g.CreatedAt())
	}
}

func TestReconstituteCompetence_PreservesAllFields(t *testing.T) {
	ts := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	c := entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{
		ID:            10,
		WorkProgramID: 7,
		Code:          "ПК-3",
		Type:          domain.CompetenceTypePK,
		Description:   "Разработка СУБД",
		CreatedAt:     ts,
	})
	if c == nil {
		t.Fatal("expected non-nil Competence")
	}
	if c.ID() != 10 || c.WorkProgramID() != 7 || c.Code() != "ПК-3" ||
		c.Type() != domain.CompetenceTypePK || c.Description() != "Разработка СУБД" ||
		!c.CreatedAt().Equal(ts) {
		t.Errorf("ReconstituteCompetence field mismatch")
	}
}

func TestReconstituteTopic_PreservesAllFieldsIncludingNullableWeek(t *testing.T) {
	week := 5
	tp := entities.ReconstituteTopic(entities.ReconstituteTopicInput{
		ID:               11,
		WorkProgramID:    7,
		Kind:             domain.TopicKindLab,
		Title:            "Лабораторная по нормализации",
		Hours:            6,
		WeekNumber:       &week,
		LearningOutcomes: "Студент применяет 3НФ",
		OrderIndex:       2,
	})
	if tp == nil {
		t.Fatal("expected non-nil Topic")
	}
	if tp.ID() != 11 || tp.WorkProgramID() != 7 || tp.Kind() != domain.TopicKindLab ||
		tp.Title() != "Лабораторная по нормализации" || tp.Hours() != 6 ||
		tp.WeekNumber() == nil || *tp.WeekNumber() != 5 ||
		tp.LearningOutcomes() != "Студент применяет 3НФ" || tp.OrderIndex() != 2 {
		t.Errorf("ReconstituteTopic field mismatch")
	}
}

func TestReconstituteTopic_NilWeekPreserved(t *testing.T) {
	tp := entities.ReconstituteTopic(entities.ReconstituteTopicInput{
		ID:            12,
		WorkProgramID: 7,
		Kind:          domain.TopicKindSelfStudy,
		Title:         "СРС",
		Hours:         36,
	})
	if tp == nil {
		t.Fatal("expected non-nil Topic")
	}
	if tp.WeekNumber() != nil {
		t.Errorf("WeekNumber should remain nil, got *%d", *tp.WeekNumber())
	}
}

func TestReconstituteAssessmentCriterion_PreservesAllFields(t *testing.T) {
	a := entities.ReconstituteAssessmentCriterion(entities.ReconstituteAssessmentCriterionInput{
		ID:               20,
		WorkProgramID:    7,
		Type:             domain.AssessmentTypeFinal,
		Description:      "Экзамен",
		MaxScore:         100,
		ExampleQuestions: []string{"Q1", "Q2"},
	})
	if a == nil {
		t.Fatal("expected non-nil AssessmentCriterion")
	}
	if a.ID() != 20 || a.WorkProgramID() != 7 || a.Type() != domain.AssessmentTypeFinal ||
		a.Description() != "Экзамен" || a.MaxScore() != 100 ||
		len(a.ExampleQuestions()) != 2 || a.ExampleQuestions()[0] != "Q1" {
		t.Errorf("ReconstituteAssessmentCriterion field mismatch")
	}
}

func TestReconstituteReference_PreservesAllFields(t *testing.T) {
	year := 2024
	r := entities.ReconstituteReference(entities.ReconstituteReferenceInput{
		ID:            30,
		WorkProgramID: 7,
		Kind:          domain.ReferenceKindElectronic,
		Citation:      "Postgres docs",
		Year:          &year,
		ISBN:          "",
		URL:           "https://www.postgresql.org/docs/",
		OrderIndex:    1,
	})
	if r == nil {
		t.Fatal("expected non-nil Reference")
	}
	if r.ID() != 30 || r.WorkProgramID() != 7 || r.Kind() != domain.ReferenceKindElectronic ||
		r.Citation() != "Postgres docs" || r.Year() == nil || *r.Year() != 2024 ||
		r.URL() != "https://www.postgresql.org/docs/" || r.OrderIndex() != 1 {
		t.Errorf("ReconstituteReference field mismatch")
	}
}

func TestReconstituteRevision_PreservesAllFieldsApprovedState(t *testing.T) {
	approver := int64(99)
	approvedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 3, 30, 9, 0, 0, 0, time.UTC)
	updated := approvedAt
	payload := []byte(`{"before":{"hours":36},"after":{"hours":40}}`)

	rev := entities.ReconstituteRevision(entities.ReconstituteRevisionInput{
		ID:             50,
		WorkProgramID:  7,
		RevisionNumber: 2,
		ChangeType:     domain.RevisionChangeTypeHours,
		ChangeSummary:  "Перераспределение часов",
		Status:         domain.RevisionStatusApproved,
		AuthorID:       11,
		ApproverID:     &approver,
		ApprovedAt:     &approvedAt,
		RejectReason:   "",
		DiffPayload:    payload,
		CreatedAt:      created,
		UpdatedAt:      updated,
	})
	if rev == nil {
		t.Fatal("expected non-nil Revision")
	}
	if rev.ID() != 50 || rev.WorkProgramID() != 7 || rev.RevisionNumber() != 2 ||
		rev.ChangeType() != domain.RevisionChangeTypeHours ||
		rev.ChangeSummary() != "Перераспределение часов" ||
		rev.Status() != domain.RevisionStatusApproved || rev.AuthorID() != 11 ||
		rev.ApproverID() == nil || *rev.ApproverID() != 99 ||
		rev.ApprovedAt() == nil || !rev.ApprovedAt().Equal(approvedAt) ||
		string(rev.DiffPayload()) != string(payload) ||
		!rev.CreatedAt().Equal(created) || !rev.UpdatedAt().Equal(updated) {
		t.Errorf("ReconstituteRevision field mismatch: %+v", rev)
	}
}

func TestReconstituteRevision_NilApproverAndApprovedAtPreserved(t *testing.T) {
	rev := entities.ReconstituteRevision(entities.ReconstituteRevisionInput{
		ID:             51,
		WorkProgramID:  7,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeOther,
		ChangeSummary:  "Прочие правки",
		Status:         domain.RevisionStatusDraft,
		AuthorID:       11,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
	if rev == nil {
		t.Fatal("expected non-nil Revision")
	}
	if rev.ApproverID() != nil {
		t.Errorf("ApproverID should be nil, got *%d", *rev.ApproverID())
	}
	if rev.ApprovedAt() != nil {
		t.Errorf("ApprovedAt should be nil, got %v", *rev.ApprovedAt())
	}
}

func TestReconstituteWorkProgram_PreservesRootFieldsAndChildSlices(t *testing.T) {
	approver := int64(99)
	approvedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	updated := approvedAt

	goal := entities.ReconstituteGoal(entities.ReconstituteGoalInput{
		ID: 1, WorkProgramID: 7, Text: "цель", OrderIndex: 0, CreatedAt: created,
	})
	comp := entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{
		ID: 2, WorkProgramID: 7, Code: "ПК-1", Type: domain.CompetenceTypePK,
		Description: "д", CreatedAt: created,
	})
	tp := entities.ReconstituteTopic(entities.ReconstituteTopicInput{
		ID: 3, WorkProgramID: 7, Kind: domain.TopicKindLecture, Title: "Л", Hours: 4,
	})
	a := entities.ReconstituteAssessmentCriterion(entities.ReconstituteAssessmentCriterionInput{
		ID: 4, WorkProgramID: 7, Type: domain.AssessmentTypeFinal,
		Description: "Экз", MaxScore: 100,
	})
	ref := entities.ReconstituteReference(entities.ReconstituteReferenceInput{
		ID: 5, WorkProgramID: 7, Kind: domain.ReferenceKindMain, Citation: "Дейт",
	})
	rev := entities.ReconstituteRevision(entities.ReconstituteRevisionInput{
		ID: 6, WorkProgramID: 7, RevisionNumber: 1, ChangeType: domain.RevisionChangeTypeOther,
		ChangeSummary: "правки", Status: domain.RevisionStatusDraft, AuthorID: 11,
		CreatedAt: created, UpdatedAt: created,
	})

	wp := entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID:                 7,
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "СУБД",
		Status:             domain.StatusApproved,
		AuthorID:           11,
		ApproverID:         &approver,
		ApprovedAt:         &approvedAt,
		RejectReason:       "",
		Version:            3,
		CreatedAt:          created,
		UpdatedAt:          updated,
		Goals:              []*entities.Goal{goal},
		Competences:        []*entities.Competence{comp},
		Topics:             []*entities.Topic{tp},
		Assessments:        []*entities.AssessmentCriterion{a},
		References:         []*entities.Reference{ref},
		Revisions:          []*entities.Revision{rev},
	})

	if wp == nil {
		t.Fatal("expected non-nil WorkProgram")
	}
	if wp.ID() != 7 || wp.DisciplineID() != 42 || wp.SpecialtyCode() != "09.03.01" ||
		wp.ApplicableFromYear() != 2026 || wp.Title() != "Базы данных" ||
		wp.Annotation() != "СУБД" || wp.Status() != domain.StatusApproved ||
		wp.AuthorID() != 11 || wp.ApproverID() == nil || *wp.ApproverID() != 99 ||
		wp.ApprovedAt() == nil || !wp.ApprovedAt().Equal(approvedAt) ||
		wp.Version() != 3 || !wp.CreatedAt().Equal(created) || !wp.UpdatedAt().Equal(updated) {
		t.Errorf("ReconstituteWorkProgram root field mismatch")
	}
	if len(wp.Goals()) != 1 || len(wp.Competences()) != 1 || len(wp.Topics()) != 1 ||
		len(wp.Assessments()) != 1 || len(wp.References()) != 1 || len(wp.Revisions()) != 1 {
		t.Errorf("ReconstituteWorkProgram child slices not preserved: goals=%d comp=%d topics=%d ass=%d refs=%d revs=%d",
			len(wp.Goals()), len(wp.Competences()), len(wp.Topics()),
			len(wp.Assessments()), len(wp.References()), len(wp.Revisions()))
	}
}
