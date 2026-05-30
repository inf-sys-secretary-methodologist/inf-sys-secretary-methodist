package entities

import (
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// Manual editing of РПД inner collections (slice 12): Remove / Update by id.
// All mutations gate on canEditContent (draft / needs_revision) — checked
// before existence, so a frozen program reports ErrCannotEditFrozenStatus
// regardless of the id. A missing element reports ErrChildNotFound. Update
// re-validates through the NewX constructor (same invariants as Add) and
// preserves the element's identity (id / parent / createdAt).

// --- Goal ---

// RemoveGoal removes the Goal with the given id from the aggregate.
func (w *WorkProgram) RemoveGoal(id int64) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, g := range w.goals {
		if g.id == id {
			w.goals = append(w.goals[:i], w.goals[i+1:]...)
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: goal %d", domain.ErrChildNotFound, id)
}

// UpdateGoal replaces the Goal's editable fields, re-validating via NewGoal.
func (w *WorkProgram) UpdateGoal(id int64, text string, orderIndex int) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, g := range w.goals {
		if g.id == id {
			fresh, err := NewGoal(text, orderIndex)
			if err != nil {
				return err
			}
			fresh.id, fresh.workProgramID, fresh.createdAt = g.id, g.workProgramID, g.createdAt
			w.goals[i] = fresh
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: goal %d", domain.ErrChildNotFound, id)
}

// --- Competence ---

// RemoveCompetence removes the Competence with the given id.
func (w *WorkProgram) RemoveCompetence(id int64) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, c := range w.competences {
		if c.id == id {
			w.competences = append(w.competences[:i], w.competences[i+1:]...)
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: competence %d", domain.ErrChildNotFound, id)
}

// UpdateCompetence replaces the Competence's fields, re-validating via
// NewCompetence. Code uniqueness is enforced excluding the competence itself.
func (w *WorkProgram) UpdateCompetence(id int64, code string, ctype domain.CompetenceType, description string) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, c := range w.competences {
		if c.id == id {
			fresh, err := NewCompetence(code, ctype, description)
			if err != nil {
				return err
			}
			for j, other := range w.competences {
				if j != i && other.code == fresh.code {
					return fmt.Errorf("%w: code %q", domain.ErrDuplicateCompetenceCode, fresh.code)
				}
			}
			fresh.id, fresh.workProgramID, fresh.createdAt = c.id, c.workProgramID, c.createdAt
			w.competences[i] = fresh
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: competence %d", domain.ErrChildNotFound, id)
}

// --- Topic ---

// RemoveTopic removes the Topic with the given id.
func (w *WorkProgram) RemoveTopic(id int64) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, tp := range w.topics {
		if tp.id == id {
			w.topics = append(w.topics[:i], w.topics[i+1:]...)
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: topic %d", domain.ErrChildNotFound, id)
}

// UpdateTopic replaces the Topic's fields, re-validating via NewTopic.
func (w *WorkProgram) UpdateTopic(id int64, in NewTopicInput) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, tp := range w.topics {
		if tp.id == id {
			fresh, err := NewTopic(in)
			if err != nil {
				return err
			}
			fresh.id, fresh.workProgramID = tp.id, tp.workProgramID
			w.topics[i] = fresh
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: topic %d", domain.ErrChildNotFound, id)
}

// --- AssessmentCriterion ---

// RemoveAssessment removes the AssessmentCriterion with the given id.
func (w *WorkProgram) RemoveAssessment(id int64) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, a := range w.assessments {
		if a.id == id {
			w.assessments = append(w.assessments[:i], w.assessments[i+1:]...)
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: assessment %d", domain.ErrChildNotFound, id)
}

// UpdateAssessment replaces the AssessmentCriterion's fields, re-validating
// via NewAssessmentCriterion.
func (w *WorkProgram) UpdateAssessment(id int64, in NewAssessmentCriterionInput) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, a := range w.assessments {
		if a.id == id {
			fresh, err := NewAssessmentCriterion(in)
			if err != nil {
				return err
			}
			fresh.id, fresh.workProgramID = a.id, a.workProgramID
			w.assessments[i] = fresh
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: assessment %d", domain.ErrChildNotFound, id)
}

// --- Reference ---

// RemoveReference removes the Reference with the given id.
func (w *WorkProgram) RemoveReference(id int64) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, r := range w.references {
		if r.id == id {
			w.references = append(w.references[:i], w.references[i+1:]...)
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: reference %d", domain.ErrChildNotFound, id)
}

// UpdateReference replaces the Reference's fields, re-validating via NewReference.
func (w *WorkProgram) UpdateReference(id int64, in NewReferenceInput) error {
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for i, r := range w.references {
		if r.id == id {
			fresh, err := NewReference(in)
			if err != nil {
				return err
			}
			fresh.id, fresh.workProgramID = r.id, r.workProgramID
			w.references[i] = fresh
			w.updatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: reference %d", domain.ErrChildNotFound, id)
}
