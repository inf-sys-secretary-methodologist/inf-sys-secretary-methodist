package backups_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/backups"
)

func TestNewFileReader_RejectsEmptyRoot(t *testing.T) {
	t.Parallel()

	_, err := backups.NewFileReader("")
	if !errors.Is(err, backups.ErrInvalidRoot) {
		t.Fatalf("expected ErrInvalidRoot, got %v", err)
	}
}

// TestFileReader_List_FiltersAndClassifies pins the contract: only
// filenames matching the sidecar grammar are returned, both backup
// types are surfaced, and the .age / .gpg suffix becomes a typed
// EncryptionScheme on the DTO.
func TestFileReader_List_FiltersAndClassifies(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "postgres"), 0o755); err != nil {
		t.Fatalf("mkdir postgres: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "minio"), 0o755); err != nil {
		t.Fatalf("mkdir minio: %v", err)
	}

	// Use distinct mod-times so the newest-first sort order can be
	// asserted deterministically.
	postgresOldest := mustTouch(t, root, "postgres/postgres_20250120_020000.sql.gz", time.Date(2025, 1, 20, 2, 0, 0, 0, time.UTC))
	postgresEncrypted := mustTouch(t, root, "postgres/postgres_20250121_020000.sql.gz.age", time.Date(2025, 1, 21, 2, 0, 0, 0, time.UTC))
	minioGPG := mustTouch(t, root, "minio/minio_20250122_020000.tar.gz.gpg", time.Date(2025, 1, 22, 2, 0, 0, 0, time.UTC))

	// Decoys — must be filtered out.
	mustTouch(t, root, "postgres/README.txt", time.Now())
	mustTouch(t, root, "postgres/postgres_invalid_name.sql.gz", time.Now())
	mustTouch(t, root, "minio/random.tar", time.Now())
	mustTouch(t, root, "minio/.hidden", time.Now())

	r, err := backups.NewFileReader(root)
	if err != nil {
		t.Fatalf("NewFileReader: %v", err)
	}
	files, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if got, want := len(files), 3; got != want {
		t.Fatalf("List returned %d files, want %d (files=%+v)", got, want, files)
	}

	// Newest first: minioGPG, postgresEncrypted, postgresOldest.
	want := []backups.BackupFile{
		{
			Name:       "minio_20250122_020000.tar.gz.gpg",
			Type:       backups.BackupTypeMinIO,
			Size:       int64(len("seed-content")),
			ModifiedAt: minioGPG.Unix(),
			Encryption: backups.EncryptionGPG,
		},
		{
			Name:       "postgres_20250121_020000.sql.gz.age",
			Type:       backups.BackupTypePostgres,
			Size:       int64(len("seed-content")),
			ModifiedAt: postgresEncrypted.Unix(),
			Encryption: backups.EncryptionAge,
		},
		{
			Name:       "postgres_20250120_020000.sql.gz",
			Type:       backups.BackupTypePostgres,
			Size:       int64(len("seed-content")),
			ModifiedAt: postgresOldest.Unix(),
			Encryption: backups.EncryptionNone,
		},
	}

	for i, f := range files {
		if f.Name != want[i].Name {
			t.Errorf("file[%d].Name = %q, want %q", i, f.Name, want[i].Name)
		}
		if f.Type != want[i].Type {
			t.Errorf("file[%d].Type = %q, want %q", i, f.Type, want[i].Type)
		}
		if f.Encryption != want[i].Encryption {
			t.Errorf("file[%d].Encryption = %q, want %q", i, f.Encryption, want[i].Encryption)
		}
		if f.Size != want[i].Size {
			t.Errorf("file[%d].Size = %d, want %d", i, f.Size, want[i].Size)
		}
		if f.ModifiedAt != want[i].ModifiedAt {
			t.Errorf("file[%d].ModifiedAt = %d, want %d", i, f.ModifiedAt, want[i].ModifiedAt)
		}
	}
}

func TestFileReader_List_EmptyTreeReturnsEmpty(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// No subdirs at all — sidecar volume may not yet have run.

	r, err := backups.NewFileReader(root)
	if err != nil {
		t.Fatalf("NewFileReader: %v", err)
	}
	files, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty slice, got %d files", len(files))
	}
}

func mustTouch(t *testing.T, root, rel string, mtime time.Time) time.Time {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.WriteFile(p, []byte("seed-content"), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", rel, err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatalf("Chtimes %s: %v", rel, err)
	}
	return mtime
}
