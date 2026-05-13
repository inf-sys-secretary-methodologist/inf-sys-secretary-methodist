package backups_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/backups"
)

func TestNewMetricsReader_RejectsEmptyPath(t *testing.T) {
	t.Parallel()

	_, err := backups.NewMetricsReader("")
	if !errors.Is(err, backups.ErrInvalidMetricsPath) {
		t.Fatalf("expected ErrInvalidMetricsPath, got %v", err)
	}
}

func TestMetricsReader_Read_FullFixture(t *testing.T) {
	t.Parallel()

	r, err := backups.NewMetricsReader(filepath.Join("testdata", "backup_metrics_full.prom"))
	if err != nil {
		t.Fatalf("NewMetricsReader: %v", err)
	}
	m, err := r.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if m.Postgres == nil {
		t.Fatal("expected Postgres metrics populated")
	}
	if got, want := m.Postgres.LastRunAt, int64(1705708800); got != want {
		t.Errorf("Postgres.LastRunAt = %d, want %d", got, want)
	}
	if got, want := m.Postgres.LastSuccessAt, int64(1705708800); got != want {
		t.Errorf("Postgres.LastSuccessAt = %d, want %d", got, want)
	}
	if !m.Postgres.LastRunSuccess {
		t.Errorf("Postgres.LastRunSuccess = false, want true")
	}
	if got, want := m.Postgres.DurationSeconds, int64(120); got != want {
		t.Errorf("Postgres.DurationSeconds = %d, want %d", got, want)
	}
	if got, want := m.Postgres.SizeBytes, int64(1048576); got != want {
		t.Errorf("Postgres.SizeBytes = %d, want %d", got, want)
	}
	if got, want := m.Postgres.AgeSeconds, int64(3600); got != want {
		t.Errorf("Postgres.AgeSeconds = %d, want %d", got, want)
	}
	if got, want := m.Postgres.TotalCount, int64(100); got != want {
		t.Errorf("Postgres.TotalCount = %d, want %d", got, want)
	}
	if got, want := m.Postgres.SuccessCount, int64(99); got != want {
		t.Errorf("Postgres.SuccessCount = %d, want %d", got, want)
	}
	if got, want := m.Postgres.FailureCount, int64(1); got != want {
		t.Errorf("Postgres.FailureCount = %d, want %d", got, want)
	}

	if m.MinIO == nil {
		t.Fatal("expected MinIO metrics populated")
	}
	if got, want := m.MinIO.LastRunAt, int64(1705712400); got != want {
		t.Errorf("MinIO.LastRunAt = %d, want %d", got, want)
	}
	if m.MinIO.LastRunSuccess {
		t.Errorf("MinIO.LastRunSuccess = true, want false (fixture has =0)")
	}
	if got, want := m.MinIO.DurationSeconds, int64(45); got != want {
		t.Errorf("MinIO.DurationSeconds = %d, want %d", got, want)
	}
	if got, want := m.MinIO.SizeBytes, int64(524288); got != want {
		t.Errorf("MinIO.SizeBytes = %d, want %d", got, want)
	}
	if got, want := m.MinIO.FailureCount, int64(2); got != want {
		t.Errorf("MinIO.FailureCount = %d, want %d", got, want)
	}

	if m.RemoteSync == nil {
		t.Fatal("expected RemoteSync metrics populated")
	}
	if got, want := m.RemoteSync.LastRunAt, int64(1705710000); got != want {
		t.Errorf("RemoteSync.LastRunAt = %d, want %d", got, want)
	}
	if !m.RemoteSync.LastRunSuccess {
		t.Errorf("RemoteSync.LastRunSuccess = false, want true")
	}
	if got, want := m.RemoteSync.DurationSeconds, int64(30); got != want {
		t.Errorf("RemoteSync.DurationSeconds = %d, want %d", got, want)
	}
	if got, want := m.RemoteSync.TotalCount, int64(25); got != want {
		t.Errorf("RemoteSync.TotalCount = %d, want %d", got, want)
	}
}

func TestMetricsReader_Read_PartialFixturePostgresOnly(t *testing.T) {
	t.Parallel()

	r, err := backups.NewMetricsReader(filepath.Join("testdata", "backup_metrics_postgres_only.prom"))
	if err != nil {
		t.Fatalf("NewMetricsReader: %v", err)
	}
	m, err := r.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if m.Postgres == nil {
		t.Fatal("expected Postgres metrics populated")
	}
	if got, want := m.Postgres.LastRunAt, int64(1705708800); got != want {
		t.Errorf("Postgres.LastRunAt = %d, want %d", got, want)
	}
	if got, want := m.Postgres.SizeBytes, int64(2097152); got != want {
		t.Errorf("Postgres.SizeBytes = %d, want %d", got, want)
	}
	if !m.Postgres.LastRunSuccess {
		t.Errorf("Postgres.LastRunSuccess = false, want true")
	}

	if m.MinIO != nil {
		t.Errorf("expected MinIO nil (fixture omits minio block), got %+v", m.MinIO)
	}
	if m.RemoteSync != nil {
		t.Errorf("expected RemoteSync nil (fixture omits remote_sync block), got %+v", m.RemoteSync)
	}
}

func TestMetricsReader_Read_MissingFileYieldsEmpty(t *testing.T) {
	t.Parallel()

	r, err := backups.NewMetricsReader(filepath.Join(t.TempDir(), "does_not_exist.prom"))
	if err != nil {
		t.Fatalf("NewMetricsReader: %v", err)
	}
	m, err := r.Read(context.Background())
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil BackupMetrics container")
	}
	if m.Postgres != nil || m.MinIO != nil || m.RemoteSync != nil {
		t.Errorf("expected all nil for missing file, got %+v", m)
	}
}
