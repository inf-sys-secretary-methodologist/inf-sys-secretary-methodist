package schema_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// migrationsDir resolves to repo-root/migrations regardless of where
// `go test` is launched from (CI runs ./... from root; IDEs run from
// the package dir).
func migrationsDir(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	t.Fatalf("migrations/ directory not found above %s", cwd)
	return ""
}

func readMigration(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(migrationsDir(t), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(body)
}

// TestMigration004UsesAmEColumnName guards against the spelling drift
// where migration 004 originally declared schedule_lessons with the
// British-English form (double-l) while every Go consumer queries the
// American-English form (single-l). The drift caused
// GET /api/schedule/lessons/timetable to return 500 against a freshly
// migrated DB. Tests pinned via sqlmock missed it because they mirrored
// the Go form rather than the actual DB schema.
func TestMigration004UsesAmEColumnName(t *testing.T) {
	body := readMigration(t, "004_create_schedule_schema.up.sql")
	const brEForm = "is_cancel" + "led"
	const aMEForm = "is_canceled"
	if strings.Contains(body, brEForm) {
		t.Errorf("migration 004 still contains %q; Go code uses %q", brEForm, aMEForm)
	}
	if !strings.Contains(body, aMEForm) {
		t.Errorf("migration 004 must declare schedule_lessons.%s column", aMEForm)
	}
}

// TestMigration005DoesNotDuplicateTaskReminders guards the canonical
// ownership of the task_reminders table. Migration 005 originally
// created an outdated schema (remind_at, no reminder_type) which
// silently shadowed migration 038's canonical schema via
// CREATE TABLE IF NOT EXISTS — backend scheduler crashed because
// reminder_type / minutes_before columns were absent from the DB.
// Migration 038 owns the table; 005 must not declare it.
func TestMigration005DoesNotDuplicateTaskReminders(t *testing.T) {
	body := readMigration(t, "005_create_tasks_schema.up.sql")
	rx := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(IF\s+NOT\s+EXISTS\s+)?task_reminders\b`)
	if rx.MatchString(body) {
		t.Errorf("migration 005 must not CREATE task_reminders; canonical owner is migration 038")
	}
}

// TestScheduleRepoUsesUsersNameColumn guards against a second SQL drift
// in the schedule module: lesson_repository_pg.go's GetByID and
// GetTimetable used to project u.first_name || ' ' || u.last_name from
// the users join, but the users table in migration 001 declares a
// single `name` column. Any timetable / lesson-by-id query crashed
// with column "u.first_name" does not exist against a fresh DB.
func TestScheduleRepoUsesUsersNameColumn(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoPath := ""
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "internal/modules/schedule/infrastructure/persistence/lesson_repository_pg.go")
		if _, err := os.Stat(candidate); err == nil {
			repoPath = candidate
			break
		}
	}
	if repoPath == "" {
		t.Fatal("lesson_repository_pg.go not found")
	}
	body, err := os.ReadFile(repoPath)
	if err != nil {
		t.Fatalf("read repo: %v", err)
	}
	src := string(body)
	if strings.Contains(src, "u.first_name") || strings.Contains(src, "u.last_name") {
		t.Errorf("lesson_repository_pg.go references u.first_name/u.last_name; users table declares only `name`")
	}
}
