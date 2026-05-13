// Backup admin frontend types matching backend DTOs at
// internal/shared/admin/backups/handler.go (CombinedResponse +
// nested projection structs). The /admin/backups surface is read-
// only — file listing + Prometheus textfile metrics from the
// /backup sidecar.
//
// Bounded context: admin observability (only `system_admin` reads).

// BackupType discriminates the two artifact families produced by the
// sidecar — full PostgreSQL dump versus MinIO bucket tarball.
export type BackupType = 'postgres' | 'minio'

// EncryptionScheme reflects the optional encryption applied by the
// sidecar. Empty string ('') = unencrypted; downstream UI surfaces
// a "Зашифровано (age/GPG)" badge when non-empty.
export type EncryptionScheme = '' | 'age' | 'gpg'

// BackupFile is one row in the file listing. `modified_at` is a
// Unix timestamp in seconds — formatted client-side via the same
// helper used elsewhere in the codebase so locale formatting stays
// consistent.
export interface BackupFile {
  name: string
  type: BackupType
  size: number
  modified_at: number
  encryption: EncryptionScheme
}

// TypeMetrics is the per-type (postgres / minio) Prometheus
// textfile projection. Zero values are legitimate when the sidecar
// has yet to run once on that type — the page surfaces an "ok / no
// data" state by checking `last_run_at === 0`.
export interface TypeMetrics {
  last_run_at: number
  last_success_at: number
  last_run_success: boolean
  duration_seconds: number
  size_bytes: number
  age_seconds: number
  total_count: number
  success_count: number
  failure_count: number
}

// RemoteSyncMetrics is the offsite-sync stream (no per-type
// dimension). Currently surfaced as a banner on the page; absent in
// the response when remote sync is disabled.
export interface RemoteSyncMetrics {
  last_run_at: number
  last_success_at: number
  last_run_success: boolean
  duration_seconds: number
  total_count: number
  success_count: number
  failure_count: number
}

// BackupMetricsResponse is the metrics container. Each nullable
// sub-object distinguishes "the sidecar has yet to write this
// stream" from "this stream has zero counts".
export interface BackupMetricsResponse {
  postgres: TypeMetrics | null
  minio: TypeMetrics | null
  remote_sync: RemoteSyncMetrics | null
}

// BackupListResponse is the wire shape under `data` of the envelope
// returned by GET /api/admin/backups. The hook lifts it directly so
// consumers do not need to traverse the envelope past the hook.
export interface BackupListResponse {
  files: BackupFile[]
  metrics: BackupMetricsResponse | null
}
