package entities_test

import (
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Slice 12 — manual editing of РПД inner collections (Remove / Update by id).
// Mirrors the AddX gate (canEditContent: only draft / needs_revision); a
// missing child surfaces ErrChildNotFound; Update reuses the NewX validation
// and preserves the child's identity (id / parent / createdAt).

// --- local builders for reconstituted (id-bearing) children ---

func recGoal(id int64, text string, order int) *entities.Goal {
	return entities.ReconstituteGoal(entities.ReconstituteGoalInput{ID: id, WorkProgramID: 1, Text: text, OrderIndex: order})
}

func draftWith2Goals(t *testing.T) *entities.WorkProgram {
	t.Helper()
	wp := newDraft(t)
	if err := wp.AddGoal(recGoal(1, "Цель A", 0)); err != nil {
		t.Fatalf("seed goal 1: %v", err)
	}
	if err := wp.AddGoal(recGoal(2, "Цель B", 1)); err != nil {
		t.Fatalf("seed goal 2: %v", err)
	}
	return wp
}

// --- RemoveGoal ---

func TestWorkProgram_RemoveGoal_RemovesByID(t *testing.T) {
	wp := draftWith2Goals(t)
	if err := wp.RemoveGoal(1); err != nil {
		t.Fatalf("RemoveGoal: %v", err)
	}
	goals := wp.Goals()
	if len(goals) != 1 || goals[0].ID() != 2 {
		t.Errorf("after RemoveGoal(1): got %d goals, want [id=2]", len(goals))
	}
}

func TestWorkProgram_RemoveGoal_NotFound(t *testing.T) {
	wp := draftWith2Goals(t)
	if err := wp.RemoveGoal(99); !errors.Is(err, domain.ErrChildNotFound) {
		t.Errorf("RemoveGoal(99): got %v, want ErrChildNotFound", err)
	}
}

// The status gate is checked before existence, so removing from a frozen
// program fails with ErrCannotEditFrozenStatus regardless of the id.
func TestWorkProgram_RemoveX_OnFrozenStatus(t *testing.T) {
	cases := []struct {
		name   string
		remove func(*entities.WorkProgram) error
	}{
		{"goal", func(w *entities.WorkProgram) error { return w.RemoveGoal(1) }},
		{"competence", func(w *entities.WorkProgram) error { return w.RemoveCompetence(1) }},
		{"topic", func(w *entities.WorkProgram) error { return w.RemoveTopic(1) }},
		{"assessment", func(w *entities.WorkProgram) error { return w.RemoveAssessment(1) }},
		{"reference", func(w *entities.WorkProgram) error { return w.RemoveReference(1) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := approved(t) // frozen
			if err := tc.remove(wp); !errors.Is(err, domain.ErrCannotEditFrozenStatus) {
				t.Errorf("Remove%s on approved: got %v, want ErrCannotEditFrozenStatus", tc.name, err)
			}
		})
	}
}

// --- RemoveX happy paths for the other four collections ---

func TestWorkProgram_RemoveOtherCollections_ByID(t *testing.T) {
	t.Run("competence", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddCompetence(entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{ID: 1, Code: "ПК-1", Type: domain.CompetenceTypePK, Description: "d1"}))
		_ = wp.AddCompetence(entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{ID: 2, Code: "ПК-2", Type: domain.CompetenceTypePK, Description: "d2"}))
		if err := wp.RemoveCompetence(1); err != nil {
			t.Fatalf("RemoveCompetence: %v", err)
		}
		if c := wp.Competences(); len(c) != 1 || c[0].ID() != 2 {
			t.Errorf("got %d competences, want [id=2]", len(c))
		}
	})
	t.Run("topic", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddTopic(entities.ReconstituteTopic(entities.ReconstituteTopicInput{ID: 1, Kind: domain.TopicKindLecture, Title: "T1", Hours: 2, OrderIndex: 0}))
		_ = wp.AddTopic(entities.ReconstituteTopic(entities.ReconstituteTopicInput{ID: 2, Kind: domain.TopicKindLecture, Title: "T2", Hours: 2, OrderIndex: 1}))
		if err := wp.RemoveTopic(2); err != nil {
			t.Fatalf("RemoveTopic: %v", err)
		}
		if tp := wp.Topics(); len(tp) != 1 || tp[0].ID() != 1 {
			t.Errorf("got %d topics, want [id=1]", len(tp))
		}
	})
	t.Run("assessment", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddAssessment(entities.ReconstituteAssessmentCriterion(entities.ReconstituteAssessmentCriterionInput{ID: 1, Type: domain.AssessmentTypeCurrent, Description: "A1", MaxScore: 5}))
		if err := wp.RemoveAssessment(1); err != nil {
			t.Fatalf("RemoveAssessment: %v", err)
		}
		if len(wp.Assessments()) != 0 {
			t.Errorf("got %d assessments, want 0", len(wp.Assessments()))
		}
	})
	t.Run("reference", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddReference(entities.ReconstituteReference(entities.ReconstituteReferenceInput{ID: 1, Kind: domain.ReferenceKindMain, Citation: "R1", OrderIndex: 0}))
		if err := wp.RemoveReference(1); err != nil {
			t.Fatalf("RemoveReference: %v", err)
		}
		if len(wp.References()) != 0 {
			t.Errorf("got %d references, want 0", len(wp.References()))
		}
	})
}

// --- UpdateGoal ---

func TestWorkProgram_UpdateGoal_RevalidatesAndPreservesIdentity(t *testing.T) {
	wp := draftWith2Goals(t)
	before := wp.Goals()[0] // id=1
	createdAt := before.CreatedAt()

	if err := wp.UpdateGoal(1, "Обновлённая цель", 3); err != nil {
		t.Fatalf("UpdateGoal: %v", err)
	}
	var g *entities.Goal
	for _, x := range wp.Goals() {
		if x.ID() == 1 {
			g = x
		}
	}
	if g == nil {
		t.Fatal("goal id=1 missing after update")
	}
	if g.Text() != "Обновлённая цель" || g.OrderIndex() != 3 {
		t.Errorf("update not applied: text=%q order=%d", g.Text(), g.OrderIndex())
	}
	if g.ID() != 1 || !g.CreatedAt().Equal(createdAt) {
		t.Errorf("identity not preserved: id=%d createdAt=%v want id=1 createdAt=%v", g.ID(), g.CreatedAt(), createdAt)
	}
}

func TestWorkProgram_UpdateGoal_InvalidValueRejected(t *testing.T) {
	wp := draftWith2Goals(t)
	if err := wp.UpdateGoal(1, "   ", 0); !errors.Is(err, domain.ErrInvalidWorkProgram) {
		t.Errorf("UpdateGoal empty text: got %v, want ErrInvalidWorkProgram", err)
	}
	// the original value is untouched on a rejected update
	if wp.Goals()[0].Text() != "Цель A" {
		t.Errorf("rejected update mutated the goal: %q", wp.Goals()[0].Text())
	}
}

func TestWorkProgram_UpdateGoal_NotFound(t *testing.T) {
	wp := draftWith2Goals(t)
	if err := wp.UpdateGoal(99, "x", 0); !errors.Is(err, domain.ErrChildNotFound) {
		t.Errorf("UpdateGoal(99): got %v, want ErrChildNotFound", err)
	}
}

func TestWorkProgram_UpdateGoal_OnFrozenStatus(t *testing.T) {
	wp := approved(t)
	if err := wp.UpdateGoal(1, "x", 0); !errors.Is(err, domain.ErrCannotEditFrozenStatus) {
		t.Errorf("UpdateGoal on approved: got %v, want ErrCannotEditFrozenStatus", err)
	}
}

// --- UpdateCompetence: code uniqueness excludes self ---

func TestWorkProgram_UpdateCompetence_DuplicateCodeRejected(t *testing.T) {
	wp := newDraft(t)
	_ = wp.AddCompetence(entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{ID: 1, Code: "ПК-1", Type: domain.CompetenceTypePK, Description: "d1"}))
	_ = wp.AddCompetence(entities.ReconstituteCompetence(entities.ReconstituteCompetenceInput{ID: 2, Code: "ПК-2", Type: domain.CompetenceTypePK, Description: "d2"}))

	// Updating competence 2 to competence 1's code collides.
	if err := wp.UpdateCompetence(2, "ПК-1", domain.CompetenceTypePK, "d2"); !errors.Is(err, domain.ErrDuplicateCompetenceCode) {
		t.Errorf("UpdateCompetence to existing code: got %v, want ErrDuplicateCompetenceCode", err)
	}
	// Updating a competence while keeping its OWN code is allowed.
	if err := wp.UpdateCompetence(2, "ПК-2", domain.CompetenceTypeUK, "новое описание"); err != nil {
		t.Errorf("UpdateCompetence keeping own code: %v", err)
	}
}

// --- UpdateX happy paths for topic / assessment / reference ---

func TestWorkProgram_UpdateOtherCollections(t *testing.T) {
	t.Run("topic", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddTopic(entities.ReconstituteTopic(entities.ReconstituteTopicInput{ID: 1, Kind: domain.TopicKindLecture, Title: "Старая", Hours: 2, OrderIndex: 0}))
		if err := wp.UpdateTopic(1, entities.NewTopicInput{Kind: domain.TopicKindPractice, Title: "Новая тема", Hours: 4, OrderIndex: 0}); err != nil {
			t.Fatalf("UpdateTopic: %v", err)
		}
		got := wp.Topics()[0]
		if got.ID() != 1 || got.Title() != "Новая тема" || got.Hours() != 4 {
			t.Errorf("UpdateTopic not applied: id=%d title=%q hours=%d", got.ID(), got.Title(), got.Hours())
		}
	})
	t.Run("assessment", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddAssessment(entities.ReconstituteAssessmentCriterion(entities.ReconstituteAssessmentCriterionInput{ID: 1, Type: domain.AssessmentTypeCurrent, Description: "Старое", MaxScore: 5}))
		if err := wp.UpdateAssessment(1, entities.NewAssessmentCriterionInput{Type: domain.AssessmentTypeFinal, Description: "Экзамен", MaxScore: 40}); err != nil {
			t.Fatalf("UpdateAssessment: %v", err)
		}
		got := wp.Assessments()[0]
		if got.ID() != 1 || got.Description() != "Экзамен" || got.MaxScore() != 40 {
			t.Errorf("UpdateAssessment not applied: id=%d desc=%q score=%d", got.ID(), got.Description(), got.MaxScore())
		}
	})
	t.Run("reference", func(t *testing.T) {
		wp := newDraft(t)
		_ = wp.AddReference(entities.ReconstituteReference(entities.ReconstituteReferenceInput{ID: 1, Kind: domain.ReferenceKindMain, Citation: "Старая книга", OrderIndex: 0}))
		if err := wp.UpdateReference(1, entities.NewReferenceInput{Kind: domain.ReferenceKindAdditional, Citation: "Новая книга", OrderIndex: 0}); err != nil {
			t.Fatalf("UpdateReference: %v", err)
		}
		got := wp.References()[0]
		if got.ID() != 1 || got.Citation() != "Новая книга" {
			t.Errorf("UpdateReference not applied: id=%d citation=%q", got.ID(), got.Citation())
		}
	})
}
