package backups

import (
	"bufio"
	"context"
	"errors"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
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

// metricLineRe captures: <name>{<labels>} <value>. The label block
// can be empty (`backup_remote_sync_...{}`); we tolerate but do not
// require it. The value is a positive integer or float (sidecar emits
// integers but Prometheus contract allows floats).
var metricLineRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)(?:\{([^}]*)\})?\s+(-?\d+(?:\.\d+)?)$`)

// Read parses the .prom file and returns the typed projection. A
// missing file (sidecar has never run on a fresh volume) yields an
// empty BackupMetrics with all pointers nil — not an error.
func (m *MetricsReader) Read(ctx context.Context) (*BackupMetrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &BackupMetrics{}

	f, err := os.Open(m.filePath) // #nosec G304 — path comes from validated config, not user input
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return result, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	// Allow up to 1 MiB per line — Prometheus lines are small but
	// defensively guard against an unexpectedly large textfile.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := metricLineRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		name := match[1]
		labels := parseLabels(match[2])
		value, err := strconv.ParseFloat(match[3], 64)
		if err != nil {
			continue
		}

		applyMetric(result, name, labels, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func parseLabels(raw string) map[string]string {
	out := map[string]string{}
	if raw == "" {
		return out
	}
	for kv := range strings.SplitSeq(raw, ",") {
		key, val, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"`)
		out[key] = val
	}
	return out
}

func applyMetric(result *BackupMetrics, name string, labels map[string]string, value float64) {
	if strings.HasPrefix(name, "backup_remote_sync_") {
		if result.RemoteSync == nil {
			result.RemoteSync = &RemoteSyncMetrics{}
		}
		applyRemoteSyncField(result.RemoteSync, name, value)
		return
	}
	if !strings.HasPrefix(name, "backup_") {
		return
	}
	switch labels["type"] {
	case "postgres":
		if result.Postgres == nil {
			result.Postgres = &BackupTypeMetrics{}
		}
		applyTypeField(result.Postgres, name, value)
	case "minio":
		if result.MinIO == nil {
			result.MinIO = &BackupTypeMetrics{}
		}
		applyTypeField(result.MinIO, name, value)
	}
}

func applyTypeField(m *BackupTypeMetrics, name string, value float64) {
	switch name {
	case "backup_last_run_timestamp_seconds":
		m.LastRunAt = int64(value)
	case "backup_last_success_timestamp_seconds":
		m.LastSuccessAt = int64(value)
	case "backup_last_run_success":
		m.LastRunSuccess = value != 0
	case "backup_duration_seconds":
		m.DurationSeconds = int64(value)
	case "backup_size_bytes":
		m.SizeBytes = int64(value)
	case "backup_age_seconds":
		m.AgeSeconds = int64(value)
	case "backup_total_count":
		m.TotalCount = int64(value)
	case "backup_success_count":
		m.SuccessCount = int64(value)
	case "backup_failure_count":
		m.FailureCount = int64(value)
	}
}

func applyRemoteSyncField(m *RemoteSyncMetrics, name string, value float64) {
	switch name {
	case "backup_remote_sync_last_run_timestamp_seconds":
		m.LastRunAt = int64(value)
	case "backup_remote_sync_last_success_timestamp_seconds":
		m.LastSuccessAt = int64(value)
	case "backup_remote_sync_last_run_success":
		m.LastRunSuccess = value != 0
	case "backup_remote_sync_duration_seconds":
		m.DurationSeconds = int64(value)
	case "backup_remote_sync_total_count":
		m.TotalCount = int64(value)
	case "backup_remote_sync_success_count":
		m.SuccessCount = int64(value)
	case "backup_remote_sync_failure_count":
		m.FailureCount = int64(value)
	}
}
