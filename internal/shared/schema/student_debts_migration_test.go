package schema_test

import (
	"strings"
	"testing"
)

// Migration 050 owns the student_debts bounded context tables. These
// parity tests guard against schema drift between the SQL and the Go
// domain: sqlmock unit tests mirror the Go side and would happily pass
// against a migration that declared the wrong column name or a CHECK
// whose enum values diverged from the domain constants (see
// feedback_sqlmock_masks_schema_drift). The literal expectations below
// mirror, byte-for-byte:
//   - internal/modules/student_debts/domain/entities/debt_status.go
//   - internal/modules/student_debts/domain/entities/resit_result.go
//   - internal/modules/student_debts/domain/entities/control_form.go
//   - the column projections in student_debt_repository_pg.go

// TestMigration050DeclaresStudentDebtsTable pins the student_debts root
// table: every persisted column, the natural-key UNIQUE, the best-effort
// FKs (nullable, non-cascading), and the semester range CHECK.
func TestMigration050DeclaresStudentDebtsTable(t *testing.T) {
	up := readMigration(t, "050_create_student_debts.up.sql")

	mustContainTable(t, up, "student_debts")

	cols := []string{
		"id", "student_full_name", "group_name", "discipline_name",
		"semester", "control_form", "student_user_id", "discipline_id",
		"source_ref", "source_hash", "status", "version",
		"created_at", "updated_at",
	}
	for _, c := range cols {
		mustContain(t, up, c, "student_debts column")
	}

	// Best-effort links: nullable, never cascade-delete the academic
	// record when a user/discipline row is removed.
	mustContain(t, up, "REFERENCES users(id) ON DELETE SET NULL", "student_user_id FK")
	mustContain(t, up, "REFERENCES curriculum_section_items(id) ON DELETE SET NULL", "discipline_id FK")

	// Natural key: one student × one discipline × one semester × group.
	mustContainAll(t, up, "student_debts natural-key UNIQUE",
		"UNIQUE", "group_name", "student_full_name", "discipline_name", "semester")

	// Semester ∈ [1,12] mirrors NewStudentDebt's invariant.
	mustContain(t, up, "semester BETWEEN 1 AND 12", "semester CHECK")

	// version optimistic-lock column defaults to 1 (mirrors Version: 1).
	mustContainAll(t, up, "version default", "version", "DEFAULT 1")
}

// TestMigration050StatusCheckMirrorsDomain pins the status CHECK enum to
// the DebtStatus constants byte-for-byte.
func TestMigration050StatusCheckMirrorsDomain(t *testing.T) {
	up := readMigration(t, "050_create_student_debts.up.sql")
	for _, v := range []string{"open", "resit_scheduled", "commission", "closed_passed", "closed_failed"} {
		mustContain(t, up, "'"+v+"'", "status CHECK value")
	}
}

// TestMigration050ControlFormCheckMirrorsDomain pins the control_form
// CHECK enum to the ControlForm constants byte-for-byte.
func TestMigration050ControlFormCheckMirrorsDomain(t *testing.T) {
	up := readMigration(t, "050_create_student_debts.up.sql")
	for _, v := range []string{"zachet", "exam", "course_project", "differential_zachet"} {
		mustContain(t, up, "'"+v+"'", "control_form CHECK value")
	}
}

// TestMigration050DeclaresResitAttemptsTable pins the child table:
// columns, the cascade FK to student_debts, the recorder FK, the
// (debt_id, attempt_no) UNIQUE and the result CHECK enum.
func TestMigration050DeclaresResitAttemptsTable(t *testing.T) {
	up := readMigration(t, "050_create_student_debts.up.sql")

	mustContainTable(t, up, "debt_resit_attempts")

	cols := []string{
		"id", "debt_id", "attempt_no", "scheduled_date", "examiner",
		"is_commission", "result", "grade", "recorded_by",
		"recorded_at", "created_at",
	}
	for _, c := range cols {
		mustContain(t, up, c, "debt_resit_attempts column")
	}

	// Attempts belong to the debt aggregate — deleting the debt deletes
	// its attempts.
	mustContain(t, up, "REFERENCES student_debts(id) ON DELETE CASCADE", "debt_id FK")
	mustContain(t, up, "REFERENCES users(id)", "recorded_by FK")

	mustContainAll(t, up, "attempt UNIQUE", "UNIQUE", "debt_id", "attempt_no")

	for _, v := range []string{"pending", "passed", "failed", "no_show"} {
		mustContain(t, up, "'"+v+"'", "result CHECK value")
	}
}

// TestMigration050MaintainsUpdatedAt pins the BEFORE UPDATE trigger that
// keeps student_debts.updated_at fresh — the domain leaves timestamps to
// the DB (NewStudentDebt sets neither created_at nor updated_at), so the
// trigger is the source of truth.
func TestMigration050MaintainsUpdatedAt(t *testing.T) {
	up := readMigration(t, "050_create_student_debts.up.sql")
	mustContain(t, up, "CREATE TRIGGER", "updated_at trigger")
	mustContain(t, up, "updated_at = NOW()", "trigger sets updated_at")
}

// TestMigration050DownDropsTables pins the rollback: both tables dropped.
func TestMigration050DownDropsTables(t *testing.T) {
	down := readMigration(t, "050_create_student_debts.down.sql")
	mustContain(t, down, "DROP TABLE", "down migration")
	mustContain(t, down, "debt_resit_attempts", "down drops child table")
	mustContain(t, down, "student_debts", "down drops root table")
}

// TestMigration051RealignsDisciplineFK pins migration 051, which realigns
// student_debts.discipline_id from curriculum_section_items(id) (migration
// 050) to disciplines(id) — the canonical discipline entity the rest of the
// system references (work_program migration 047, schedule_lessons migration
// 004). This is what lets teacher scoping share one id space:
// schedule_lessons.teacher_id → schedule_lessons.discipline_id (disciplines)
// = student_debts.discipline_id. discipline_id is best-effort/nullable and
// never populated yet, so the realignment needs no data migration.
func TestMigration051RealignsDisciplineFK(t *testing.T) {
	up := readMigration(t, "051_realign_student_debts_discipline_fk.up.sql")

	// Drop the old curriculum_section_items FK and add the disciplines one,
	// preserving the best-effort ON DELETE SET NULL semantics.
	mustContain(t, up, "DROP CONSTRAINT", "051 drops the old FK")
	mustContain(t, up, "REFERENCES disciplines(id) ON DELETE SET NULL", "051 adds disciplines FK")

	down := readMigration(t, "051_realign_student_debts_discipline_fk.down.sql")
	mustContain(t, down, "REFERENCES curriculum_section_items(id) ON DELETE SET NULL",
		"051 down restores the curriculum_section_items FK")
}

// --- assertion helpers -----------------------------------------------------

func mustContain(t *testing.T, body, needle, what string) {
	t.Helper()
	if !strings.Contains(body, needle) {
		t.Errorf("%s: migration must contain %q", what, needle)
	}
}

func mustContainAll(t *testing.T, body, what string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(body, n) {
			t.Errorf("%s: migration must contain %q", what, n)
		}
	}
}

func mustContainTable(t *testing.T, body, table string) {
	t.Helper()
	if !strings.Contains(body, "CREATE TABLE IF NOT EXISTS "+table) {
		t.Errorf("migration must CREATE TABLE IF NOT EXISTS %s", table)
	}
}
