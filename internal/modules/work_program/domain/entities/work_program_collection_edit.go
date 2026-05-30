package entities

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"

// Manual editing of РПД inner collections (slice 12): Remove / Update by id.
// Stub bodies — real implementation lands in the GREEN commit.

// RemoveGoal removes the Goal with the given id from the aggregate.
func (w *WorkProgram) RemoveGoal(id int64) error { return nil }

// UpdateGoal replaces the Goal's editable fields, re-validating via NewGoal.
func (w *WorkProgram) UpdateGoal(id int64, text string, orderIndex int) error { return nil }

// RemoveCompetence removes the Competence with the given id.
func (w *WorkProgram) RemoveCompetence(id int64) error { return nil }

// UpdateCompetence replaces the Competence's fields, re-validating via
// NewCompetence (code uniqueness excludes the competence itself).
func (w *WorkProgram) UpdateCompetence(id int64, code string, ctype domain.CompetenceType, description string) error {
	return nil
}

// RemoveTopic removes the Topic with the given id.
func (w *WorkProgram) RemoveTopic(id int64) error { return nil }

// UpdateTopic replaces the Topic's fields, re-validating via NewTopic.
func (w *WorkProgram) UpdateTopic(id int64, in NewTopicInput) error { return nil }

// RemoveAssessment removes the AssessmentCriterion with the given id.
func (w *WorkProgram) RemoveAssessment(id int64) error { return nil }

// UpdateAssessment replaces the AssessmentCriterion's fields, re-validating.
func (w *WorkProgram) UpdateAssessment(id int64, in NewAssessmentCriterionInput) error { return nil }

// RemoveReference removes the Reference with the given id.
func (w *WorkProgram) RemoveReference(id int64) error { return nil }

// UpdateReference replaces the Reference's fields, re-validating via NewReference.
func (w *WorkProgram) UpdateReference(id int64, in NewReferenceInput) error { return nil }
