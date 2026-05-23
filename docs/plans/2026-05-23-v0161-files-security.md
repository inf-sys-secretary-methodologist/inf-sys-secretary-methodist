# Plan — v0.161.0 files Tier 1 security hotfix

**Date**: 2026-05-23
**Module**: `internal/modules/files`
**Audit verdict**: REJECT mean 4.4 / min 2 — see `docs/plans/2026-05-20-v1.0.0-batch2-audit.md` §files (worst module of batch 2)
**Batch 2 progress**: 3/5 (after this release; auth ✅ v0.159.0 + users ✅ v0.160.0 already shipped)
**Issue**: [#290](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/290)
**Branch**: `hotfix/issue-290-v0161-files-security`

## Scope — 5 ADRs (4 TIER 0 + 1 Tier 1 elevated)

### ADR-1 — Cross-user file enumeration + IDOR (TIER 0)

**Threat**: `GetFile`/`GetFileWithDownloadURL`/`DownloadFile`/`AttachFile` принимают только `id int64`. NO `actorID` arg, no ownership check. Sequential BIGSERIAL → enumerate + download every file ever uploaded (HR records, exam reports, diploma drafts). 1-hour MinIO presigned URL is anonymous external sharing surface.

**Files**: `file_usecase.go:117-206` (GetFile/GetFileWithDownloadURL/DownloadFile/AttachFile), `file_handler.go` (routes), `file_metadata.go` (entity — `UploadedBy` exists)

**Fix**:
- New domain free function `AuthorizeFileAccess(actorID int64, actorRole string, file *FileMetadata, action FileAction) error` in `files/domain/`. Read actions: actor==uploader OR system_admin. Write actions (Attach/DeleteVersion/CreateVersion): actor==uploader only (no admin override — uploaders are the only legit mutators; admin can read but not impersonate).
- `FileAction` typed enum: `FileActionRead`, `FileActionAttach`, `FileActionCreateVersion`, `FileActionDelete`.
- Usecase signatures: `GetFile(ctx, id, actorID, actorRole)`, `DownloadFile(ctx, id, actorID, actorRole)`, `GetFileWithDownloadURL(ctx, id, urlExpiration, actorID, actorRole)`, `AttachFile(ctx, input)` — input.UserID & input.UserRole already exist в `AttachFileInput`, just enforce.
- Handler reads `actorID := c.Get("user_id")`, `role := c.Get("role")`, passes through.
- Sentinel `ErrFileAccessDenied` (var Err* — `*PermissionError` struct → migrate to var sentinel; closes Tier 1 DDD-9 absorb partially).
- Handler maps → 403 с stable error code `file_access_denied`.
- Audit denial: emit `file_access_denied` с `actor_user_id`, `target_file_id`, `action`, `reason` (mirror v0.160.0 ADR-4 pattern).

**RED test** (table-driven):
- (read, actor=uploader) → allow
- (read, actor=other, role=student) → ErrFileAccessDenied + audit denial
- (read, actor=other, role=system_admin) → allow
- (attach, actor=uploader) → allow
- (attach, actor=other, role=system_admin) → ErrFileAccessDenied (no admin override on write)
- (delete, actor=uploader) → allow
- (delete, actor=other) → ErrFileAccessDenied

### ADR-2 — CreateVersion ownership hijack (TIER 0)

**Threat**: `version_usecase.go:48-108` — CreateVersion looks up file but only checks `IsTemporary`. Anyone pushes new version в anyone's file. UI shows latest = attacker-controlled (defacement + supply-chain).

**Files**: `version_usecase.go:48-108`, `file_handler.go` (CreateVersion route)

**Fix**:
- CreateVersion calls `AuthorizeFileAccess(actorID, actorRole, file, FileActionCreateVersion)` BEFORE проверки `IsTemporary` (fail-fast on auth).
- `CreateVersionInput.UserID` already exists; add `UserRole string` для polymorphic admin override consistency (но admin не permitted на write per ADR-1 — argument kept for symmetry, will always fail-deny на non-uploader).
- Sentinel: same `ErrFileAccessDenied` (reused).
- Handler maps → 403.
- Audit denial: same pattern (`file_version_create_denied`).
- Bonus: sanitize `Comment` через `SanitizeString` (closes Tier 1 #8 stored XSS).

**RED test**:
- (uploader creates version) → ok
- (other student creates version) → ErrFileAccessDenied + audit denial
- (system_admin creates version в чужом файле) → ErrFileAccessDenied (no admin write override)
- (comment с `<script>`) → sanitized в version row

### ADR-3 — DEAD `FileValidator.ValidateFile` wire-in (TIER 0)

**Threat**: `file_validator.go:99-164` (MIME whitelist + magic bytes sniff + size cap + extension whitelist) wired в `main.go:778-803`, но usecase calls only `ValidateFileName` (cosmetic sanitization). Upload `evil.exe` с `Content-Type: application/octet-stream` succeeds. Magic-byte detection / MIME whitelist / size cap entirely BYPASSED.

**Files**: `file_usecase.go:55-115` (UploadFile), `storage.go` (interface), `file_validator.go` (shared)

**Fix**:
- Narrow port в `files/application/usecases/storage.go`: `type FileValidator interface { ValidateFile(ctx, header *multipart.FileHeader, reader io.Reader) (*ValidationResult, error); ValidateFileName(string) (string, error) }` — both methods, не только ValidateFileName.
- UploadFile signature accepts `*multipart.FileHeader` for header validation (size + MIME from Content-Type header) OR struct DTO `ValidateInput { Name string; Size int64; DeclaredMIME string }` if multipart-coupling is too tight. Choose **DTO option** to keep usecase pure (Clean Architecture).
- Call sequence: `ValidateFileName` → `ValidateFile` (full pipeline, reading sniff bytes без consuming reader через `bufio.Reader.Peek` or tee).
- `evil.exe` test: declared MIME `application/octet-stream` + first bytes 4D 5A (PE) → reject with `ErrFileTypeForbidden`.

**RED test** (table-driven, reader-based):
- Valid PDF (magic %PDF-) declared `application/pdf` → ok
- evil.exe (4D 5A magic, declared octet-stream) → ErrFileTypeForbidden
- 100MB file (size > 50MB cap) → ErrFileSizeExceeded
- HTML disguised as image (declared image/png, magic <html) → ErrMimeMismatch

### ADR-4 — Extract `IsInlineSafeMime` + `BuildContentDisposition` к shared pkg (Tier 1 elevated)

**Threat**: No clickjacking protection — raw MinIO presigned URL with original `Content-Type`. v0.156.0 documents hardening (`IsInlineSafeMime` + `BuildContentDisposition`) не применяется к files module. HTML/SVG uploaded as `text/html` opens inline → XSS / framejacking surface.

**Current location**: `internal/modules/documents/interfaces/http/handlers/inline_mime.go` + `content_disposition.go`.

**Files**: extract to `internal/shared/infrastructure/http/headers/` — `inline_mime.go` + `content_disposition.go` + `rfc2231.go` (filename encoder); update documents callers (import path change); files DownloadResponse adds `ContentDisposition` field, handler sets header.

**Fix**:
- New shared pkg `internal/shared/infrastructure/http/headers/` (or similar) — pure functions, zero deps beyond stdlib.
- Documents handler imports change.
- Files Download (`file_handler.go`) returns presigned URL **plus** safe Content-Disposition header (`attachment; filename*=UTF-8''<encoded>` for non-inline MIME, `inline; filename*=...` for whitelisted inline like image/pdf).
- Files Version Download same treatment.

**RED test**:
- `IsInlineSafeMime("image/png")` → true
- `IsInlineSafeMime("text/html")` → false (force download)
- `IsInlineSafeMime("application/pdf")` → true (с separate sandbox header) — match v0.156.0 behavior
- `BuildContentDisposition("файл.pdf", "application/pdf")` → `inline; filename*=UTF-8''%D1%84%D0%B0%D0%B9%D0%BB.pdf` exact match

### ADR-5 — `/api/files/cleanup` admin gate (Tier 1)

**Threat**: `main.go:2703` route registers `POST /api/files/cleanup` WITHOUT `RequireRole(system_admin)`. Handler comment "Доступ должен быть ограничен администраторам" — security-by-comment.

**Files**: `main.go:2703`, `file_handler.go:368-384` (handler), `routes.go` (if exists)

**Fix**:
- Apply `RequireRole(SystemAdmin)` middleware к route — mirror v0.133.0 `usersAdminMW` pattern + v0.160.0 ADR-2 (departments/positions).
- Handler comment removed (no longer authoritative).

**RED test**: student JWT POST `/api/files/cleanup` → 403 at middleware boundary; admin JWT → 200.

## Tier 2 deferred к v0.161.1 (per `feedback_tier2_absorb_same_release` ≤4 cap; v0.161.0 absorbs only minimal-cost cleanups)

1. DIP relocation — `files/domain/repositories/` → `files/application/usecases/` (mirror v0.157.1, v0.160.1)
2. `*ValidationError` → `var ErrFileValidation` (other ValidationError struct types в module)
3. `MaxBytesReader` wrapper на upload route (DoS mitigation)
4. Per-user quota for uploads
5. UI strings extraction (Russian errors in usecase → handler/messages package)
6. Concrete `*storage.S3Client` → narrow port в constructor signature
7. `Files`/`Users interface{}` untyped slices в DTOs → typed slices

(7 items mirror v0.155→.1, v0.157→.1, v0.160→.1 split precedent — likely ship v0.161.1 within 1-2 sessions of v0.161.0 SHIP.)

## Reviewer trajectory expectation

- Round 1: FIX-CYCLE 5.x/4 (worst module — 4 TIER 0, expect Tests/DDD axis dips)
- Round 2: FIX-CYCLE 7.x/6 (absorbs landed, may surface Tier 2 escalations or test-coverage gaps for round-1 absorbs)
- Round 3: SHIP ≥8/8 (mirror v0.160.0 trajectory)

**Tests-axis discipline (v0.160.0 lesson)**: every FIX-CYCLE absorb behavior change gets RED→GREEN pair within absorb, не "patching gap of original RED".

## Release ritual (mechanical, ~30-60min after SHIP)

1. `bash _tools/bump_version.sh 0.161.0` (8 files)
2. Explicit stage version-bump files (per `feedback_explicit_stage_for_release`)
3. Release commit `fix(files): v0.161.0 — Tier 1 security hotfix (closes #XXX)` (NOT `release(files)` — Validate PR Metadata)
4. Push + PR + likely admin-merge при >1000 LOC (precedent v0.158/v0.159/v0.160)
5. **STICKY**: `git tag -a v0.161.0` + `git push origin v0.161.0` + `gh release create v0.161.0` IMMEDIATELY post-merge
6. Docs PR `docs/roles-and-flows.md` refresh — banner 0.160.0 → 0.161.0 + batch 2 progress 3/5
7. Memory artifacts: `project_v0161_0_files_security.md` (auto-memory) + MEMORY.md index + chronicles entry + handoff + `memory/context/workspace.md`
