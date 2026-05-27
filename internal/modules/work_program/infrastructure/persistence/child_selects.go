package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// Child SELECT helpers mirror migration 048 column order 1:1 (same
// drift-detection contract as child_inserts.go). Each helper returns
// the slice of hydrated child entities for the given root id; empty
// slice + nil error when no rows match.

func selectGoals(ctx context.Context, db DBTX, rootID int64) ([]*entities.Goal, error) {
	const query = `
		SELECT id, work_program_id, text, order_index, created_at
		FROM work_program_goals
		WHERE work_program_id = $1
		ORDER BY order_index ASC, id ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select goals: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.Goal
	for rows.Next() {
		var in entities.ReconstituteGoalInput
		if err := rows.Scan(&in.ID, &in.WorkProgramID, &in.Text, &in.OrderIndex, &in.CreatedAt); err != nil {
			return nil, fmt.Errorf("work_program: scan goal: %w", err)
		}
		out = append(out, entities.ReconstituteGoal(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate goals: %w", err)
	}
	return out, nil
}

func selectCompetences(ctx context.Context, db DBTX, rootID int64) ([]*entities.Competence, error) {
	const query = `
		SELECT id, work_program_id, code, type, description, created_at
		FROM work_program_competences
		WHERE work_program_id = $1
		ORDER BY code ASC, id ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select competences: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.Competence
	for rows.Next() {
		var in entities.ReconstituteCompetenceInput
		var typeStr string
		if err := rows.Scan(&in.ID, &in.WorkProgramID, &in.Code, &typeStr, &in.Description, &in.CreatedAt); err != nil {
			return nil, fmt.Errorf("work_program: scan competence: %w", err)
		}
		in.Type = domain.CompetenceType(typeStr)
		out = append(out, entities.ReconstituteCompetence(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate competences: %w", err)
	}
	return out, nil
}

func selectTopics(ctx context.Context, db DBTX, rootID int64) ([]*entities.Topic, error) {
	const query = `
		SELECT id, work_program_id, kind, title, hours, week_number, learning_outcomes, order_index
		FROM work_program_topics
		WHERE work_program_id = $1
		ORDER BY order_index ASC, id ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select topics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.Topic
	for rows.Next() {
		var in entities.ReconstituteTopicInput
		var kindStr string
		var week sql.NullInt32
		var outcomes sql.NullString
		if err := rows.Scan(&in.ID, &in.WorkProgramID, &kindStr, &in.Title, &in.Hours, &week, &outcomes, &in.OrderIndex); err != nil {
			return nil, fmt.Errorf("work_program: scan topic: %w", err)
		}
		in.Kind = domain.TopicKind(kindStr)
		if week.Valid {
			w := int(week.Int32)
			in.WeekNumber = &w
		}
		if outcomes.Valid {
			in.LearningOutcomes = outcomes.String
		}
		out = append(out, entities.ReconstituteTopic(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate topics: %w", err)
	}
	return out, nil
}

func selectAssessments(ctx context.Context, db DBTX, rootID int64) ([]*entities.AssessmentCriterion, error) {
	const query = `
		SELECT id, work_program_id, type, description, max_score, example_questions
		FROM work_program_assessment
		WHERE work_program_id = $1
		ORDER BY id ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select assessments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.AssessmentCriterion
	for rows.Next() {
		var in entities.ReconstituteAssessmentCriterionInput
		var typeStr string
		var questions pq.StringArray
		if err := rows.Scan(&in.ID, &in.WorkProgramID, &typeStr, &in.Description, &in.MaxScore, &questions); err != nil {
			return nil, fmt.Errorf("work_program: scan assessment: %w", err)
		}
		in.Type = domain.AssessmentType(typeStr)
		if len(questions) > 0 {
			in.ExampleQuestions = []string(questions)
		}
		out = append(out, entities.ReconstituteAssessmentCriterion(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate assessments: %w", err)
	}
	return out, nil
}

func selectReferences(ctx context.Context, db DBTX, rootID int64) ([]*entities.Reference, error) {
	const query = `
		SELECT id, work_program_id, kind, citation, year, isbn, url, order_index
		FROM work_program_references
		WHERE work_program_id = $1
		ORDER BY order_index ASC, id ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select references: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.Reference
	for rows.Next() {
		var in entities.ReconstituteReferenceInput
		var kindStr string
		var year sql.NullInt32
		var isbn, url sql.NullString
		if err := rows.Scan(&in.ID, &in.WorkProgramID, &kindStr, &in.Citation, &year, &isbn, &url, &in.OrderIndex); err != nil {
			return nil, fmt.Errorf("work_program: scan reference: %w", err)
		}
		in.Kind = domain.ReferenceKind(kindStr)
		if year.Valid {
			y := int(year.Int32)
			in.Year = &y
		}
		if isbn.Valid {
			in.ISBN = isbn.String
		}
		if url.Valid {
			in.URL = url.String
		}
		out = append(out, entities.ReconstituteReference(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate references: %w", err)
	}
	return out, nil
}

func selectRevisions(ctx context.Context, db DBTX, rootID int64) ([]*entities.Revision, error) {
	const query = `
		SELECT id, work_program_id, revision_number, change_type, change_summary,
			status, author_id, approver_id, approved_at, reject_reason,
			diff_payload, created_at, updated_at
		FROM work_program_revisions
		WHERE work_program_id = $1
		ORDER BY revision_number ASC`
	rows, err := db.QueryContext(ctx, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("work_program: select revisions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []*entities.Revision
	for rows.Next() {
		var in entities.ReconstituteRevisionInput
		var changeTypeStr, statusStr string
		var approverID sql.NullInt64
		var approvedAt sql.NullTime
		var rejectReason sql.NullString
		var diffPayload []byte
		if err := rows.Scan(
			&in.ID, &in.WorkProgramID, &in.RevisionNumber,
			&changeTypeStr, &in.ChangeSummary, &statusStr,
			&in.AuthorID, &approverID, &approvedAt, &rejectReason,
			&diffPayload, &in.CreatedAt, &in.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("work_program: scan revision: %w", err)
		}
		in.ChangeType = domain.RevisionChangeType(changeTypeStr)
		in.Status = domain.RevisionStatus(statusStr)
		if approverID.Valid {
			v := approverID.Int64
			in.ApproverID = &v
		}
		if approvedAt.Valid {
			t := approvedAt.Time
			in.ApprovedAt = &t
		}
		if rejectReason.Valid {
			in.RejectReason = rejectReason.String
		}
		in.DiffPayload = diffPayload
		out = append(out, entities.ReconstituteRevision(in))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("work_program: iterate revisions: %w", err)
	}
	return out, nil
}
