package persistence

import (
	"context"
	"fmt"
)

// deleteAllChildren removes every child row belonging to the given
// root id inside the open tx. Used by Update before the reinsert pass
// — the simplest correct re-sync algorithm for an aggregate root with
// many child tables (Vaughn Vernon "delete + reinsert" pattern).
//
// Order does not matter — every child table is independent (each
// child only references the root). Migration 048 has no cross-child
// FKs.
func deleteAllChildren(ctx context.Context, tx execQuerier, rootID int64) error {
	tables := []string{
		"work_program_goals",
		"work_program_competences",
		"work_program_topics",
		"work_program_assessment",
		"work_program_references",
		"work_program_revisions",
	}
	for _, table := range tables {
		query := `DELETE FROM ` + table + ` WHERE work_program_id = $1`
		if _, err := tx.ExecContext(ctx, query, rootID); err != nil {
			return fmt.Errorf("work_program: update: delete children %s: %w", table, err)
		}
	}
	return nil
}
