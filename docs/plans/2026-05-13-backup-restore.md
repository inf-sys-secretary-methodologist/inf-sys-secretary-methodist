# Plan — Backup admin read-only UI (Phase 5 #2)

**Дата:** 2026-05-13
**Базовая версия:** v0.131.1 (`1e15954b`)
**Initiative:** Phase 5 #2 — admin observability поверх существующего `/backup/` sidecar

## Discovery / scope re-evaluation

**Initial assumption (rejected):** "Phase 5 #2 Backup/Restore admin UI" подразумевает full in-app pg_dump/pg_restore module с S3 upload, retention scheduler, async goroutine orchestration.

**Reality после recon:** репозиторий уже содержит mature backup/restore sidecar в `/backup/`:
- **Commits:** `3d5edf93` (encryption + notifications + metrics + CI/CD), `9e729270` (PostgreSQL + MinIO + offsite sync)
- **Capabilities:** pg_dump + MinIO tar backup (cron + one-shot режимы), age/GPG encryption, Prometheus textfile metrics, Telegram/Email/Webhook notifications, offsite S3 sync (Backblaze/Yandex/Selectel), retention auto-GC, restore scripts, CI/CD verification
- **Wiring:** `compose.yml` service `backup` под profile `backup`, named volumes `backup_data:/backups` и `backup_metrics:/var/lib/node_exporter/textfile_collector`

**CLAUDE.md Red Line:** "Не 'чинить' работающее" + "Не менять конфиги если значение странное — сначала понять". In-app re-implementation = duplicate ~300 LoC mature shell + ops drift. **Reject.**

**Re-scoped goal:** admin observability surface — system_admin видит state of backup operations (list files + Prometheus metrics + download), но НЕ управляет sidecar. Sidecar = single source of truth для backup lifecycle.

## Release split

| Release | Scope | TDD pairs |
|---------|-------|-----------|
| **v0.132.0** | Read-only admin UI: list files в `backup_data` volume + parse Prometheus textfile metrics + download files + frontend page + nav + i18n × 4 + compose.yml backend volume mount | 4 |
| v0.132.x (backlog) | Trigger one-shot backup от UI (sidecar API extension); restore UI (high blast — отдельный focused review) | TBD |

Этот документ покрывает **v0.132.0** только.

---

## ADR-1 — Read-only scope; sidecar остаётся source of truth

**Контекст:** альтернативы — full in-app pg_dump (rejected), trigger-capable UI (sidecar API extension), status-only dashboard tile.

**Решение:** v0.132.0 = read-only. UI отображает files + metrics + download. Никаких triggers / delete / restore / encryption-key handling. Sidecar управляется через `docker compose run --rm backup /scripts/backup-all.sh` (admin SSH-and-run) и через cron schedule.

**Why:**
- Sidecar uptime-proven, encryption + retention + notifications работают; не имеет смысла duplicate.
- Trigger via docker socket = security risk (exposes daemon control к app process). Trigger via sidecar API requires extending `/backup/entrypoint.sh` HTTP listener mode — отдельный risk, отдельный release.
- Restore = wipe DB; даже UI confirm dialog не уменьшает blast radius до приемлемого для дипломного demo.

**Trade-off:** admin не может trigger backup из UI — должен SSH. Backlog `v0.132.x` если требуется.

## ADR-2 — Volume mount strategy = read-only shared volumes на backend

`compose.yml` backend service получает 2 RO mount'а:

```yaml
backend:
  volumes:
    # existing mounts...
    - backup_data:/var/backups:ro
    - backup_metrics:/var/backup_metrics:ro
```

Env vars в config:
- `BACKUP_FILES_DIR` (default `/var/backups`)
- `BACKUP_METRICS_DIR` (default `/var/backup_metrics`)

**Why:** filesystem-level integration = simplest possible coupling; backend читает то же что sidecar пишет; no API surface, no network coordination, no auth secrets share.

**Trade-off:** только same-host deployment поддерживается. Multi-host backup аппроачи (NFS / S3 shared bucket) — backlog. Diplomas: not in scope.

## ADR-3 — Filename whitelist + path traversal hardening

Whitelist regex для listing (entries не matching — skipped tihme):
```go
postgresBackupRe = regexp.MustCompile(`^postgres_\d{8}_\d{6}\.sql\.gz(\.age|\.gpg)?$`)
minioBackupRe    = regexp.MustCompile(`^minio_\d{8}_\d{6}\.tar\.gz(\.age|\.gpg)?$`)
```

Download endpoint:
1. Parse `:name` URL param через `filepath.Base` (strip directory components).
2. Re-match через whitelist regex; reject 400 если no match.
3. `filepath.Join(BACKUP_FILES_DIR, "postgres", validatedName)` или `/minio/`.
4. `filepath.Clean` + `strings.HasPrefix(resolvedPath, BACKUP_FILES_DIR)` guard — defence-in-depth против любого symlink/escape.
5. `os.Open` → stream к `c.Writer` через `io.Copy`; `Content-Disposition: attachment; filename=...`.

**Why:** path traversal = #1 security risk любого filesystem-serve handler; whitelist + canonical-prefix double-check — canonical pattern (CWE-22 mitigation).

## ADR-4 — Metrics source = Prometheus textfile parser

`/var/lib/node_exporter/textfile_collector/*.prom` files имеют формат:
```
# TYPE backup_last_run_timestamp_seconds gauge
backup_last_run_timestamp_seconds{server_id="production",type="postgres"} 1705708800
backup_duration_seconds{server_id="production",type="postgres"} 120
backup_size_bytes{server_id="production",type="postgres"} 1048576
backup_last_run_success{server_id="production",type="postgres"} 1
```

Backend parser: line-by-line scan, skip `# TYPE`/`# HELP`, regex `^backup_(\w+)\{[^}]*type="(\w+)"[^}]*\}\s+(\d+(\.\d+)?)$` → map metric_name+type → value. Aggregate per type (postgres + minio) в `BackupMetricsDTO {Postgres, MinIO BackupTypeMetrics}` с полями `LastRunAt`, `LastSuccessAt`, `LastRunSuccess`, `DurationSeconds`, `SizeBytes`, `AgeSeconds`, `TotalCount`, `SuccessCount`, `FailureCount`.

**Why:** альтернатива — call Prometheus HTTP API. Но:
- Prometheus сам должен scrape backup textfile через node_exporter; в diploma compose Prometheus может быть undeployed.
- Textfile format = simple subset, parsing ~50 LoC; зависимости от Prometheus uptime устраняем.
- Same-host volume mount уже available (ADR-2).

**Trade-off:** parser fragile к format changes от backup scripts (`metrics.sh`). Migration risk: пишем integration test с real `.prom` fixture от existing sidecar.

## ADR-5 — Download strategy = filesystem stream, encrypted-suffix warning

Backend `GET /api/admin/backups/:type/:name/download` (type ∈ {postgres, minio}):
- Stream через `c.DataFromReader(http.StatusOK, fileSize, contentType, io.Reader, headers)` — gin built-in.
- `Content-Type`: `application/gzip` (или `application/octet-stream` для encrypted `.age`/`.gpg`).
- `Content-Disposition: attachment; filename="<name>"` (validated name).

Frontend: button "Скачать" → simple `<a href={downloadUrl} download>` с Authorization header injection через blob fetch + URL.createObjectURL (existing pattern в codebase для documents download).

Encrypted suffix `.age` / `.gpg` → UI badge "Зашифровано (age)" / "Зашифровано (GPG)" + tooltip "Для расшифровки требуется приватный ключ".

**Why:** filesystem stream — straightforward; нет S3 presigned URL потому что backups в local volume не в S3 (offsite sync — separate сценарий ADR-2 trade-off).

## ADR-6 — Audit emissions (2 actions)

`AuditSink` narrow port (reuse pattern):

| Action | Resource | Fields |
|--------|----------|--------|
| `backup.viewed` | `backup` | (none — page load, possibly omit если noisy) |
| `backup.downloaded` | `backup` | `filename`, `file_size_bytes`, `backup_type` |

**Note:** `backup.viewed` — debatable; некоторые audit philosophies omit "view" events чтобы не drown signal. Решение: **omit для v0.132.0**. Только download emits — это actual data exfiltration vector worth recording.

## ADR-7 — Module placement = `internal/shared/admin/backups/`

Mirror v0.131.0 auditlog placement (`internal/shared/admin/auditlog/`). No domain invariants (no `BackupFile` entity со state transitions — это plain read DTO). Files:
- `internal/shared/admin/backups/file_reader.go` + `file_reader_test.go`
- `internal/shared/admin/backups/metrics_reader.go` + `metrics_reader_test.go`
- `internal/shared/admin/backups/usecase.go` (thin orchestrator: read files + metrics, combine response)
- `internal/shared/admin/backups/handler.go` + `handler_test.go`
- `internal/shared/admin/backups/dto.go` (BackupFileDTO, BackupMetricsDTO, BackupTypeMetrics)

**Why:** plain DTO-shaped read API без mutating ops; full module (`internal/modules/backups/`) — over-engineering для two endpoints, mirror auditlog precedent.

## ADR-8 — Frontend page shape = mirror `/admin/audit-logs`

`frontend/src/app/admin/backups/page.tsx`:
- Role guard через `useAuthCheck` + `useEffect` redirect на `/forbidden` для non-`system_admin` (mirror audit-logs lines 74–89).
- `AppLayout` wrapper.
- Header `<h1>` с `HardDrive` lucide icon.
- **Metrics tile** (grid 2×3): postgres + minio cards с last_run / age / success_rate / duration / total.
- **Files table** (Radix Table): name, type badge, size (formatFileSize), modified_at, encrypted badge, download button.
- Sort by modified_at DESC client-side.
- Empty state когда `files.length === 0`.
- Loading state (`Loader2` spinner) пока SWR fetching.

`useBackups` SWR hook: single GET `/api/admin/backups` returning `{files, metrics}` combined response. Refresh interval — статичный (no in-flight polling нужен — sidecar cron jobs независимы).

Navigation entry (`frontend/src/config/navigation.ts`):
```ts
{
  nameKey: 'backups',
  url: '/admin/backups',
  icon: HardDrive,
  roles: [UserRole.SYSTEM_ADMIN],
}
```

i18n × 4 namespace `adminBackups`:
- `title`, `description`, `loadFailed`, `empty.{title,description}`
- `metrics.{postgres,minio,lastRun,lastSuccess,age,duration,sizeBytes,totalCount,successCount,failureCount,never,running,ok,failed}`
- `columns.{name,type,size,modifiedAt,encrypted,actions}`
- `actions.{download,viewing}`
- `encryption.{none,age,gpg,tooltip}`
- `types.{postgres,minio}`

JSON-load parity test (mirror v0.131.0 pattern): load real ru/en/fr/ar JSON files, assert structural parity.

---

## TDD pair plan (4 pairs)

1. **BackupFileReader** (`internal/shared/admin/backups/file_reader.go`) — filesystem listing с filename whitelist + path traversal guard. Table-driven test ≥5 filenames (regular postgres, encrypted age, encrypted gpg, invalid pattern, dot-file).
2. **BackupMetricsReader** (`internal/shared/admin/backups/metrics_reader.go`) — Prometheus textfile parser. Fixture-based: 3 .prom test files под `testdata/`.
3. **AdminBackupHandler + main.go wiring + compose.yml mount** — 2 routes (`GET /admin/backups`, `GET /admin/backups/:type/:name/download`). Integration test через real `gin.Engine` (mirror v0.131.0 handler_test). compose.yml backend volumes extension. `BACKUP_FILES_DIR`/`BACKUP_METRICS_DIR` config.
4. **Frontend** (`useBackups` hook + types + `/admin/backups` page + DownloadButton + nav + i18n × 4 + parity test) — mirror /admin/audit-logs shape.

---

## Carry-forward backlog (v0.132.x patches / future)

- **Trigger one-shot backup** — extend `/backup/entrypoint.sh` HTTP listener mode (новый `BACKUP_MODE=api`), backend POST'ит trigger через internal docker network. Reentrancy semantics + auth secret needed.
- **Restore UI** — separate plan doc; very high blast radius; SSH-and-run остаётся canonical path.
- **PII redaction в audit `filename` field** — filename содержит date stamp, нечувствительно; defer.
- **Multi-replica backup file access** — backend deployment scales, but backups только на single host. Backlog.
- **Audit `backup.viewed`** — если post-launch traffic suggests low noise, enable.

---

## Definition of Done v0.132.0

- 4 TDD pairs (8 work commits) + reviewer Tier 2 absorb (если требуется) + release commit
- Reviewer round verdict SHIP mean ≥9 / min ≥8 (TDD/DDD/CA/Security/Testing/i18n)
- Backend: golangci-lint 0 / gosec 0 / 108+ packages green (added 1 shared subpackage)
- Frontend: tests pass для new useBackups hook + page + i18n parity (4 locales)
- Integration test через real `gin.Engine` pинит middleware-handler contract (v0.131.0 pattern)
- audit_logs emits для `backup.downloaded` verified
- `compose.yml` extension с backup volumes mounted RO на backend (verified `docker compose config` parses)
- docs/roles-and-flows.md обновлён с admin backup observability surface
- CHANGELOG.md entry
- 8-file version sync (VERSION + frontend/VERSION + main.go @version + versionString + frontend/package.json + lock + 3 swagger)
- handoff + chronicles + MEMORY.md entries

---

## Out of scope (explicit)

- ❌ Trigger backup от UI (sidecar API extension — backlog)
- ❌ Delete backup file from UI (sidecar retention scheduler handles это; manual delete = SSH)
- ❌ Restore от UI (high blast radius — отдельный release когда дойдём)
- ❌ Backup encryption key management UI (admin SSH'ит и работает с age/GPG keys offline)
- ❌ Cron schedule editing UI (BACKUP_SCHEDULE env var, compose redeploy для change)
- ❌ Notifications config UI (`NOTIFY_TELEGRAM_*` env vars; admin-владелец редактирует compose env)
- ❌ Trigger remote S3 sync (sidecar handles это автоматически когда `REMOTE_SYNC_ENABLED=true`)
