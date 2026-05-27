package persistence

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Child INSERT helpers. Mirror migration 048 column order 1:1 so a
// schema drift surfaces as a sqlmock arg mismatch (per
// feedback_sqlmock_masks_schema_drift). Each helper writes the
// nullable columns through the helpers.go wrappers.

func insertGoal(ctx context.Context, tx execQuerier, rootID int64, g *entities.Goal) error {
	const query = `
		INSERT INTO work_program_goals (
			work_program_id, text, order_index, created_at
		) VALUES ($1, $2, $3, $4)`
	if _, err := tx.ExecContext(ctx, query,
		rootID, g.Text(), g.OrderIndex(), g.CreatedAt(),
	); err != nil {
		return fmt.Errorf("work_program: save: insert goal: %w", err)
	}
	return nil
}

func insertCompetence(ctx context.Context, tx execQuerier, rootID int64, c *entities.Competence) error {
	const query = `
		INSERT INTO work_program_competences (
			work_program_id, code, type, description, created_at
		) VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, query,
		rootID, c.Code(), string(c.Type()), c.Description(), c.CreatedAt(),
	); err != nil {
		return fmt.Errorf("work_program: save: insert competence: %w", err)
	}
	return nil
}

func insertTopic(ctx context.Context, tx execQuerier, rootID int64, t *entities.Topic) error {
	const query = `
		INSERT INTO work_program_topics (
			work_program_id, kind, title, hours, week_number,
			learning_outcomes, order_index
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, query,
		rootID,
		string(t.Kind()),
		t.Title(),
		t.Hours(),
		nullableIntPtr(t.WeekNumber()),
		nullableString(t.LearningOutcomes()),
		t.OrderIndex(),
	); err != nil {
		return fmt.Errorf("work_program: save: insert topic: %w", err)
	}
	return nil
}

func insertAssessment(ctx context.Context, tx execQuerier, rootID int64, a *entities.AssessmentCriterion) error {
	const query = `
		INSERT INTO work_program_assessment (
			work_program_id, type, description, max_score, example_questions
		) VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.ExecContext(ctx, query,
		rootID,
		string(a.Type()),
		a.Description(),
		a.MaxScore(),
		pq.Array(a.ExampleQuestions()),
	); err != nil {
		return fmt.Errorf("work_program: save: insert assessment: %w", err)
	}
	return nil
}

func insertReference(ctx context.Context, tx execQuerier, rootID int64, r *entities.Reference) error {
	const query = `
		INSERT INTO work_program_references (
			work_program_id, kind, citation, year, isbn, url, order_index
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, query,
		rootID,
		string(r.Kind()),
		r.Citation(),
		nullableIntPtr(r.Year()),
		nullableString(r.ISBN()),
		nullableString(r.URL()),
		r.OrderIndex(),
	); err != nil {
		return fmt.Errorf("work_program: save: insert reference: %w", err)
	}
	return nil
}

func insertRevision(ctx context.Context, tx execQuerier, rootID int64, rv *entities.Revision) error {
	const query = `
		INSERT INTO work_program_revisions (
			work_program_id, revision_number, change_type, change_summary,
			status, author_id, approver_id, approved_at, reject_reason,
			diff_payload, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	var payload any
	if rv.DiffPayload() != nil {
		payload = rv.DiffPayload()
	}
	if _, err := tx.ExecContext(ctx, query,
		rootID,
		rv.RevisionNumber(),
		string(rv.ChangeType()),
		rv.ChangeSummary(),
		string(rv.Status()),
		rv.AuthorID(),
		nullableInt64Ptr(rv.ApproverID()),
		nullableTimePtr(rv.ApprovedAt()),
		nullableString(rv.RejectReason()),
		payload,
		rv.CreatedAt(),
		rv.UpdatedAt(),
	); err != nil {
		return fmt.Errorf("work_program: save: insert revision: %w", err)
	}
	return nil
}
