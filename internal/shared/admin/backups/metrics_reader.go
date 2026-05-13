package backups

import (
	"context"
	"errors"
)

// BackupTypeMetrics is the parsed projection of one backup type's
// (postgres or minio) Prometheus-format counters and gauges emitted
// by the sidecar's metrics.sh script.
type BackupTypeMetrics struct {
	LastRunAt       int64
	LastSuccessAt   int64
	LastRunSuccess  bool
	DurationSeconds int64
	SizeBytes       int64
	AgeSeconds      int64
	TotalCount      int64
	SuccessCount    int64
	FailureCount    int64
}

// RemoteSyncMetrics covers the offsite-sync stream (no per-type
// dimension — sync is a single operation across all artifacts).
type RemoteSyncMetrics struct {
	LastRunAt       int64
	LastSuccessAt   int64
	LastRunSuccess  bool
	DurationSeconds int64
	TotalCount      int64
	SuccessCount    int64
	FailureCount    int64
}

// BackupMetrics is the top-level container returned by MetricsReader.
// Pointers are nil when the corresponding metrics block was not
// present in the .prom file — the sidecar omits a type entirely
// until its first run rather than emitting zero values.
type BackupMetrics struct {
	Postgres   *BackupTypeMetrics
	MinIO      *BackupTypeMetrics
	RemoteSync *RemoteSyncMetrics
}

// ErrInvalidMetricsPath signals an empty .prom path passed to the
// constructor. Behaves like ErrInvalidRoot for FileReader — fail fast
// at DI rather than per-request.
var ErrInvalidMetricsPath = errors.New("backups: invalid metrics path")

// MetricsReader parses the single backup_metrics.prom file produced
// by /backup/scripts/metrics.sh in its Prometheus textfile-collector
// format.
type MetricsReader struct {
	filePath string
}

// NewMetricsReader constructs a reader over the .prom file written
// by the sidecar (typically /var/backup_metrics/backup_metrics.prom
// inside the backend container).
func NewMetricsReader(filePath string) (*MetricsReader, error) {
	if filePath == "" {
		return nil, ErrInvalidMetricsPath
	}
	return &MetricsReader{filePath: filePath}, nil
}

// Read parses the .prom file and returns the typed projection. A
// missing file (sidecar has never run on a fresh volume) yields an
// empty BackupMetrics with all pointers nil — not an error.
//
// Pair-2 stub: returns the deferred-runtime sentinel until GREEN
// wires the line-based parser.
func (m *MetricsReader) Read(ctx context.Context) (*BackupMetrics, error) {
	return nil, errMetricsNotImplemented
}

var errMetricsNotImplemented = errors.New("backups: metrics reader not implemented (RED stub)")
