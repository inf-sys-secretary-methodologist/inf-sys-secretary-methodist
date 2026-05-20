# v0.156.0 — Documents Module Tier 1 Security Hotfix

Closes #266. Part of v1.0.0 batch 1 fix-cycle (`docs/plans/2026-05-20-v1.0.0-batch1-audit.md`).

## Scope

Tier 1 security + correctness blockers identified в documents module audit (mean 6.7/10, min 6/10, FIX-CYCLE). Split-mind module: workflow pack (v0.148.0+) exemplary, legacy code drags down.

1. Sentinel error mismatch (4 document-not-found + 3 version-not-found sites) — `errors.Is` returns false → workflow audit not_found branches dead → 500 instead of 404
2. Clickjacking via `?inline=true` — strips X-Frame-Options + rewrites CSP `frame-ancestors *`
3. Header injection via filename в Content-Disposition (no CRLF strip / RFC 5987 encoding)
4. Magic string `"admin"` — project enum = `"system_admin"`, SystemAdmin не bypasses access-control WHERE
5. Cross-module import + fire-and-forget `context.Background()` + Russian UI strings в `sharing_usecase`
6. UploadFile bypasses Register domain method — audit fields empty
7. Russian UI strings в usecases (broader sweep — tag_usecase + sharing_usecase + handler)

## ADRs locked upfront

### ADR-1: Sentinel error refactor — `ErrDocumentNotFound` + `ErrVersionNotFound`

Apply v0.155 ADR-8 pattern. Declare в `internal/modules/documents/domain/repositories/`:
- `var ErrDocumentNotFound = errors.New("document not found")`
- `var ErrVersionNotFound = errors.New("version not found")`

Replace `fmt.Errorf("document not found")` (4 sites in `document_repository_pg.go:129,178,209,601`) и `fmt.Errorf("version not found")` / `fmt.Errorf("no versions found")` (3 sites in `:499,534,628`) с direct return of sentinel. Handler `MapDomainError` does `errors.Is` discrimination → 404 instead of 500. Workflow audit branches re-enabled.

### ADR-2: Restrict `?inline=true` к whitelisted preview-friendly mime types

Replace blanket strip of X-Frame-Options + CSP rewrite с mime-type whitelist:
- `image/png`, `image/jpeg`, `image/gif`, `image/webp`, `image/svg+xml`
- `application/pdf`
- `text/plain`

For non-whitelisted MIME — ignore `?inline=true` (force attachment). For whitelisted — DO NOT strip X-Frame-Options globally; instead set `Content-Security-Policy: frame-ancestors 'self'` (same-origin only). Document trade-off в code comment.

Add table-driven test rejecting executable / script mime types (text/html, application/javascript).

### ADR-3: RFC 5987 filename encoding — header injection closed

Replace `c.Header("Content-Disposition", "...; filename=\""+fileInfo.FileName+"\"")` с `mime.FormatMediaType` или `filename*=UTF-8''<percent-encoded>` per RFC 5987 §3.2.

Use Go stdlib `mime.FormatMediaType("attachment", map[string]string{"filename*": fileInfo.FileName})` — automatically picks `filename*=UTF-8''...` для non-ASCII / unsafe chars + retains `filename="..."` ASCII fallback. CRLF stripped automatically (replaced с `_` by stdlib).

Add table-driven test для injection payloads: `evil.pdf\r\nX-Injected: yes`, `пример.docx`, `file"with"quotes.txt`.

### ADR-4: Replace magic `"admin"` с typed enum constant

Replace `filter.CurrentUserRole != "admin"` (2 sites: `document_repository_pg.go:255,932`) с `filter.CurrentUserRole != string(entities.RoleSystemAdmin)`. Also check if `CurrentUserRole` field type is plain `string` — if so, migrate к `entities.UserRole` strongly-typed.

Add integration test asserting system_admin role sees every row (admin-sees-all guarantee).

### ADR-5: NotificationSink narrow port для sharing_usecase

Apply v0.155 ADR-5 pattern. Define в `internal/modules/documents/application/usecases/`:
```go
type NotificationSink interface {
    NotifyDocumentShared(ctx context.Context, userID int64, docTitle string, link string) error
}
```

Refactor `SharingUseCase`:
- Replace concrete `notifUsecases.NotificationUseCase` field с `NotificationSink` interface
- Remove fire-and-forget goroutine с `context.Background()` — пусть adapter решит (синхронный call OR async с serverCtx). UseCase shouldn't know.
- Move Russian UI strings из usecase в adapter (interfaces/http/messages or notifications module surface).

Concrete adapter в `cmd/server/main.go` wires real `NotificationUseCase` + handles fire-and-forget с `serverCtx`-derived ctx. Russian strings live в adapter / message bundle.

### ADR-6: UploadFile calls Register domain method properly

Replace direct mutation `doc.Status = entities.DocumentStatusRegistered` (`document_usecase.go:306-308`) одним из вариантов:
- (a) Remove auto-register entirely — let separate `Register` endpoint handle status transition (preferred if no UX regression)
- (b) Call `doc.Register(actorID, registrationNumber, registrationDate)` через domain method если auto-register нужен для UX — pass real inputs

Investigate UX dependency first. Add test asserting either:
- (a) UploadFile leaves status Draft (or whatever upload-time default)
- (b) When auto-register chosen, RegisteredBy/RegistrationNumber/RegistrationDate fields are populated

### ADR-7: Russian UI strings → domain errors / i18n bundle

Promote к `var ErrTagNotFound`, `var ErrDocumentLinkExists` etc. в `application/usecases/errors.go` (new file). Handler `MapDomainError` already covers similar pattern — extend для tag/document-link sentinels. UI Russian strings live в `interfaces/http/messages` per CLAUDE.md gate.

Scope: tag_usecase.go (5 distinct messages), sharing_usecase.go (lift to adapter per ADR-5), document_handler.go:231 (handler-level — extract к messages package).

Table-driven test для each sentinel ensures handler maps к right HTTP status + i18n key.

## TDD pairs (rough plan; refine in execution)

| # | RED | GREEN |
|---|-----|-------|
| 1 | TestDocumentRepoPG_GetByID_NotFound_ReturnsSentinel (errors.Is) | Declare ErrDocumentNotFound в domain/repositories; replace fmt.Errorf 4 sites; same for ErrVersionNotFound × 3 sites |
| 2 | TestDownloadFile_InlineRestrictsToImageMime (table-driven 6 mime types) | Add isInlineSafeMime whitelist; ignore inline=true для non-whitelisted; CSP frame-ancestors 'self' |
| 3 | TestDownloadFile_FilenameRFC5987Encoding (table-driven: ASCII / Cyrillic / CRLF injection / quotes) | Replace string concat с mime.FormatMediaType |
| 4 | TestDocumentRepoPG_SystemAdminSeesAllRows | Replace magic "admin" с entities.RoleSystemAdmin constant в 2 sites |
| 5 | TestSharingUseCase_NotifiesViaSinkAdapter | Define NotificationSink port; refactor SharingUseCase; adapter в main.go |
| 6 | TestDocumentUseCase_UploadFile_DoesNotForgeRegisteredStatus | Remove direct mutation OR call doc.Register с proper inputs; populate audit fields |
| 7 | TestTagUseCase_NotFoundReturnsSentinel (table-driven 5 sentinels) | Lift Russian strings к domain sentinels + handler MapDomainError extension |

## Acceptance criteria

- [ ] All 4 document-not-found + 3 version-not-found sites return sentinel; errors.Is works in workflow audit
- [ ] `?inline=true` mime-type whitelisted; CSP frame-ancestors не wildcarded
- [ ] Content-Disposition uses RFC 5987 encoding; CRLF injection rejected
- [ ] SystemAdmin bypasses access-control WHERE (integration test green)
- [ ] sharing_usecase imports zero `notifications/` packages; ports + adapter live
- [ ] UploadFile audit fields populated OR Draft status preserved (per ADR-6 choice)
- [ ] tag_usecase / sharing_usecase / handler Russian UI strings replaced с sentinels + i18n keys
- [ ] Reviewer verdict ≥9/min ≥8 через `superpowers:code-reviewer`
- [ ] CI green; 8 version files bumped → 0.156.0
- [ ] CHANGELOG entry; plan ADR doc referenced в release commit

## Out of scope (deferred к later patches)

- Documents Tier 2/3 items (legacy usecase narrow-port refactor для template / version / sharing беyond ADR-5 scope)
- `DocumentUseCase` S3Client DIP refactor (roadmap-deferred)
- Russian comments в Go docstrings (not security-critical)
- workflow pack (already exemplary, v0.148.0+)

## Carry-forward references

- v0.154.0 plan precedent: `docs/plans/2026-05-20-v0154-reporting-security.md`
- v0.155.0 plan precedent: `docs/plans/2026-05-20-v0155-ai-security.md`
- v0.155 ADR-8 sentinel pattern: `memory/feedback_split_release_for_retroactive_review.md`
- v0.155 ADR-5 narrow-port pattern: `memory/feedback_narrow_port_full_extraction.md`
- Reviewer single-pass SHIP preconditions: `memory/feedback_single_pass_reviewer_ship.md`
- Reviewer fix-cycle absorb pattern: `memory/feedback_tier2_absorb_same_release.md`
- B4 Annual sub-module narrow-port pattern (gold standard): `reports/annual/application/usecases/annual_report_usecase.go:15-32`
