package backups

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AdminBackupHandler exposes the two admin endpoints
// (`GET /api/admin/backups`, `GET /api/admin/backups/:type/:name/download`).
// Mounted under the admin route group with RequireRole(system_admin)
// — handler-level role guard is intentionally absent because the
// route-level middleware is the canonical gate. Integration tests
// pin the middleware-handler pair (per memory
// `feedback_handler_context_key_must_match_middleware`).
type AdminBackupHandler struct {
	uc *AdminBackupUseCase
}

// NewAdminBackupHandler wires the handler against the use case.
// Panics on a nil use case so misconfigured DI fails at construction.
func NewAdminBackupHandler(uc *AdminBackupUseCase) *AdminBackupHandler {
	if uc == nil {
		panic("backups: nil AdminBackupUseCase")
	}
	return &AdminBackupHandler{uc: uc}
}

// CombinedResponse is the JSON projection returned by GET /api/admin/backups.
type CombinedResponse struct {
	Files   []BackupFileResponse `json:"files"`
	Metrics *MetricsResponse     `json:"metrics"`
}

// BackupFileResponse is one row in the file listing JSON.
type BackupFileResponse struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
	Encryption string `json:"encryption"`
}

// MetricsResponse mirrors the BackupMetrics container with nullable
// per-stream sub-objects so the frontend can render "no data yet"
// without confusing zero values with real measurements.
type MetricsResponse struct {
	Postgres   *TypeMetricsResponse `json:"postgres"`
	MinIO      *TypeMetricsResponse `json:"minio"`
	RemoteSync *RemoteSyncResponse  `json:"remote_sync"`
}

// TypeMetricsResponse is the per-type metrics row (postgres / minio).
type TypeMetricsResponse struct {
	LastRunAt       int64 `json:"last_run_at"`
	LastSuccessAt   int64 `json:"last_success_at"`
	LastRunSuccess  bool  `json:"last_run_success"`
	DurationSeconds int64 `json:"duration_seconds"`
	SizeBytes       int64 `json:"size_bytes"`
	AgeSeconds      int64 `json:"age_seconds"`
	TotalCount      int64 `json:"total_count"`
	SuccessCount    int64 `json:"success_count"`
	FailureCount    int64 `json:"failure_count"`
}

// RemoteSyncResponse is the offsite-sync metrics row.
type RemoteSyncResponse struct {
	LastRunAt       int64 `json:"last_run_at"`
	LastSuccessAt   int64 `json:"last_success_at"`
	LastRunSuccess  bool  `json:"last_run_success"`
	DurationSeconds int64 `json:"duration_seconds"`
	TotalCount      int64 `json:"total_count"`
	SuccessCount    int64 `json:"success_count"`
	FailureCount    int64 `json:"failure_count"`
}

// List handles GET /api/admin/backups.
//
// @Summary List backup files and metrics (admin only)
// @Tags admin
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/admin/backups [get]
func (h *AdminBackupHandler) List(c *gin.Context) {
	snap, err := h.uc.ListWithMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to list backups"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projectSnapshot(snap)})
}

// Download handles GET /api/admin/backups/:type/:name/download.
//
// @Summary Download a backup file (admin only)
// @Tags admin
// @Produce octet-stream
// @Param type path string true "Backup type (postgres | minio)"
// @Param name path string true "Backup filename"
// @Success 200 {file} binary
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/admin/backups/{type}/{name}/download [get]
func (h *AdminBackupHandler) Download(c *gin.Context) {
	backupType := BackupType(c.Param("type"))
	name := c.Param("name")
	actorID := readActorID(c)

	res, err := h.uc.Download(c.Request.Context(), actorID, backupType, name)
	switch {
	case errors.Is(err, ErrInvalidBackupName):
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid backup name or type"))
		return
	case errors.Is(err, ErrBackupNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("backup"))
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to download backup"))
		return
	}
	defer func() { _ = res.Reader.Close() }()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, res.Filename))
	c.DataFromReader(http.StatusOK, res.Size, res.ContentType, res.Reader, nil)
}

// readActorID lifts the authenticated user_id off the gin context.
// Returns 0 when missing — audit emit treats 0 as the cron-triggered
// sentinel (mirrors v0.131.1 integration emission pattern); admin
// downloads always have an authenticated user so 0 here means a
// production middleware misconfig that the route gate would have
// already rejected.
func readActorID(c *gin.Context) int64 {
	v, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch id := v.(type) {
	case int64:
		return id
	case int:
		return int64(id)
	}
	return 0
}

func projectSnapshot(snap CombinedSnapshot) CombinedResponse {
	files := make([]BackupFileResponse, 0, len(snap.Files))
	for _, f := range snap.Files {
		files = append(files, BackupFileResponse{
			Name:       f.Name,
			Type:       string(f.Type),
			Size:       f.Size,
			ModifiedAt: f.ModifiedAt,
			Encryption: string(f.Encryption),
		})
	}
	return CombinedResponse{
		Files:   files,
		Metrics: projectMetrics(snap.Metrics),
	}
}

func projectMetrics(m *BackupMetrics) *MetricsResponse {
	if m == nil {
		return nil
	}
	return &MetricsResponse{
		Postgres:   projectTypeMetrics(m.Postgres),
		MinIO:      projectTypeMetrics(m.MinIO),
		RemoteSync: projectRemoteSync(m.RemoteSync),
	}
}

func projectTypeMetrics(t *BackupTypeMetrics) *TypeMetricsResponse {
	if t == nil {
		return nil
	}
	return &TypeMetricsResponse{
		LastRunAt:       t.LastRunAt,
		LastSuccessAt:   t.LastSuccessAt,
		LastRunSuccess:  t.LastRunSuccess,
		DurationSeconds: t.DurationSeconds,
		SizeBytes:       t.SizeBytes,
		AgeSeconds:      t.AgeSeconds,
		TotalCount:      t.TotalCount,
		SuccessCount:    t.SuccessCount,
		FailureCount:    t.FailureCount,
	}
}

func projectRemoteSync(r *RemoteSyncMetrics) *RemoteSyncResponse {
	if r == nil {
		return nil
	}
	return &RemoteSyncResponse{
		LastRunAt:       r.LastRunAt,
		LastSuccessAt:   r.LastSuccessAt,
		LastRunSuccess:  r.LastRunSuccess,
		DurationSeconds: r.DurationSeconds,
		TotalCount:      r.TotalCount,
		SuccessCount:    r.SuccessCount,
		FailureCount:    r.FailureCount,
	}
}
