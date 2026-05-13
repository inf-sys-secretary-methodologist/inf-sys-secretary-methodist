package backups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/backups"
)

func init() { gin.SetMode(gin.TestMode) }

// stubAuthMiddleware mirrors the production JWTMiddleware contract.
// Pinned per memory feedback_handler_context_key_must_match_middleware
// — must write the same gin-context keys ("user_id", "role") the
// handler reads.
func stubAuthMiddleware(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func requireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "system_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, map[string]any{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

type fakeFileLister struct {
	files []backups.BackupFile
	err   error
}

func (f *fakeFileLister) List(_ context.Context) ([]backups.BackupFile, error) {
	return f.files, f.err
}

type fakeMetricsScraper struct {
	metrics *backups.BackupMetrics
	err     error
}

func (f *fakeMetricsScraper) Read(_ context.Context) (*backups.BackupMetrics, error) {
	return f.metrics, f.err
}

// spyAuditSink records every emission so tests can assert that
// `backup.downloaded` fires with the right resource + fields.
type spyAuditSink struct {
	mu     sync.Mutex
	events []auditCall
}

type auditCall struct {
	action   string
	resource string
	fields   map[string]any
}

func (s *spyAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, auditCall{action: action, resource: resource, fields: fields})
}

func (s *spyAuditSink) snapshot() []auditCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]auditCall, len(s.events))
	copy(out, s.events)
	return out
}

func newTestRouter(t *testing.T, uc *backups.AdminBackupUseCase, userID int64, role string) http.Handler {
	t.Helper()
	r := gin.New()
	h := backups.NewAdminBackupHandler(uc)

	adminGroup := r.Group("/api/admin")
	adminGroup.Use(stubAuthMiddleware(userID, role))
	adminGroup.Use(requireSystemAdmin())
	adminGroup.GET("/backups", h.List)
	adminGroup.GET("/backups/:type/:name/download", h.Download)
	return r
}

func doGET(t *testing.T, router http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAdminBackup_List_RoleGate(t *testing.T) {
	cases := []struct {
		name     string
		role     string
		wantCode int
	}{
		{"system_admin allowed", "system_admin", http.StatusOK},
		{"methodist forbidden", "methodist", http.StatusForbidden},
		{"academic_secretary forbidden", "academic_secretary", http.StatusForbidden},
		{"teacher forbidden", "teacher", http.StatusForbidden},
		{"student forbidden", "student", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := backups.NewAdminBackupUseCase(
				&fakeFileLister{files: []backups.BackupFile{}},
				&fakeMetricsScraper{metrics: &backups.BackupMetrics{}},
				"/tmp",
			)
			router := newTestRouter(t, uc, 1, tc.role)
			rec := doGET(t, router, "/api/admin/backups")
			require.Equal(t, tc.wantCode, rec.Code)
		})
	}
}

func TestAdminBackup_List_HappyPath(t *testing.T) {
	files := []backups.BackupFile{
		{
			Name: "postgres_20250121_020000.sql.gz.age", Type: backups.BackupTypePostgres,
			Size: 100, ModifiedAt: 1705708800, Encryption: backups.EncryptionAge,
		},
	}
	metrics := &backups.BackupMetrics{
		Postgres: &backups.BackupTypeMetrics{
			LastRunAt: 1705708800, LastRunSuccess: true, SizeBytes: 100, AgeSeconds: 3600,
		},
	}

	uc := backups.NewAdminBackupUseCase(
		&fakeFileLister{files: files},
		&fakeMetricsScraper{metrics: metrics},
		"/tmp",
	)
	router := newTestRouter(t, uc, 1, "system_admin")
	rec := doGET(t, router, "/api/admin/backups")

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data backups.CombinedResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Files, 1)
	require.Equal(t, "postgres_20250121_020000.sql.gz.age", body.Data.Files[0].Name)
	require.Equal(t, "postgres", body.Data.Files[0].Type)
	require.Equal(t, "age", body.Data.Files[0].Encryption)
	require.NotNil(t, body.Data.Metrics)
	require.NotNil(t, body.Data.Metrics.Postgres)
	require.True(t, body.Data.Metrics.Postgres.LastRunSuccess)
	require.Nil(t, body.Data.Metrics.MinIO)
}

// TestAdminBackup_Download covers the full validation matrix:
// happy path with audit emission, unknown type, filename grammar
// rejection, path-traversal attempt, and missing-file 404.
func TestAdminBackup_Download(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "postgres"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "postgres", "postgres_20250121_020000.sql.gz"),
		[]byte("dump-contents"),
		0o644,
	))

	cases := []struct {
		name        string
		urlPath     string
		wantCode    int
		wantBody    string
		wantAudit   bool
		auditFields map[string]any
	}{
		{
			name:      "happy path streams file and emits audit",
			urlPath:   "/api/admin/backups/postgres/postgres_20250121_020000.sql.gz/download",
			wantCode:  http.StatusOK,
			wantBody:  "dump-contents",
			wantAudit: true,
			auditFields: map[string]any{
				"filename":        "postgres_20250121_020000.sql.gz",
				"file_size_bytes": int64(13),
				"backup_type":     "postgres",
			},
		},
		{
			name:     "unknown type 400",
			urlPath:  "/api/admin/backups/unknown/postgres_20250121_020000.sql.gz/download",
			wantCode: http.StatusBadRequest,
		},
		{
			// `etcpasswd` does not match the sidecar filename grammar
			// so the whitelist regex rejects it before any filesystem
			// access — the canonical defense against path-style
			// traversal at the use-case layer.
			name:     "filename does not match grammar 400",
			urlPath:  "/api/admin/backups/postgres/etcpasswd/download",
			wantCode: http.StatusBadRequest,
		},
		{
			// Whitelist accepts but no file exists — race with the
			// sidecar's retention GC or a manual rm. Maps to 404.
			name:     "valid grammar but file missing 404",
			urlPath:  "/api/admin/backups/postgres/postgres_19990101_000000.sql.gz/download",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			audit := &spyAuditSink{}
			uc := backups.NewAdminBackupUseCase(
				&fakeFileLister{files: []backups.BackupFile{}},
				&fakeMetricsScraper{metrics: &backups.BackupMetrics{}},
				root,
			).WithAuditSink(audit)
			router := newTestRouter(t, uc, 42, "system_admin")

			rec := doGET(t, router, tc.urlPath)
			require.Equalf(t, tc.wantCode, rec.Code, "body=%s", rec.Body.String())

			if tc.wantBody != "" {
				require.Equal(t, tc.wantBody, rec.Body.String())
				require.Contains(t, rec.Header().Get("Content-Disposition"), "attachment")
			}

			events := audit.snapshot()
			if tc.wantAudit {
				require.Len(t, events, 1, "expected one audit event")
				require.Equal(t, "backup.downloaded", events[0].action)
				require.Equal(t, "backup_admin", events[0].resource)
				for k, want := range tc.auditFields {
					require.Equalf(t, want, events[0].fields[k], "field %s", k)
				}
				require.Equal(t, int64(42), events[0].fields["actor_user_id"])
			} else {
				require.Empty(t, events, "no audit on rejected/error path")
			}
		})
	}
}
