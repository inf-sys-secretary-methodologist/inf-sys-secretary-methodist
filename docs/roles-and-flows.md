# Роли и пользовательские потоки

> **Версия проекта:** 0.161.0 (см. `VERSION` в корне)
> **Состояние на:** 23 мая 2026 — **v0.161.0 files Tier 1 security hotfix shipped** (AuthorizeFileAccess IDOR closure / Version endpoints gated / Validator wired full path / Cleanup admin-only gate). Batch 1 audit + curriculum role swap closed; batch 2 hotfixes **3/5 closed** (auth ✅ v0.159.0, users ✅ v0.160.0, files ✅ v0.161.0 — messaging/announcements v0.162.0..v0.163.0 pending); v0.161.1 polish patch tracked for ADR-4 helper extraction + 12 Tier 2 items; v0.160.1 polish patch also tracked; Phase 6 #196 strict > 90% coverage ✅; 5-phase Documents workflow pack #227 closed end-to-end; #41 Workflow automation closed; Phase 5 admin observability closed; B-feature triad closed (curriculum + assignments + B4 annual report); MFA полностью end-to-end UI.
> **Источники:** код (`internal/modules/auth/domain/`, `frontend/src/lib/auth/`, `frontend/src/config/navigation.ts`), GitHub issues, `.taskmaster/`, `CHANGELOG.md`, история релизов в GitHub Releases.

> **Изменения с 0.160.0 по 0.161.0 (batch 2 audit hotfix #3 — files module Tier 1 security)**:
>
> - **v0.161.0 files Tier 1 security hotfix #290** — 4 of 5 ADRs shipped (4 TIER 0 + 1 Tier 1 admin gate); ADR-4 deferred к v0.161.1. Delivered как **3-PR split** per `feedback_phase_split_pr_size_gate` (PR Size 1000-line hard gate, 1490 LOC total): PR-A #292 (632 LOC, foundation + ADR-1) + PR-B #293 (461 LOC, ADR-2/3/5 + misspell) + PR-C #294 (455 LOC, FIX-CYCLE + release). 2 reviewer rounds absorbed в PR-C: round 1 REJECT (5 IDOR endpoints missed) → round 2 FIX-CYCLE (T0 validator overshoot + T1 cross-uploader visibility).
>   - **ADR-1 (TIER 0) IDOR closure** — `internal/modules/files/domain/authz.go` introduces `AuthorizeFileAccess(actor, file, action)` rule (sentinel `ErrFileAccessDenied`); wired into `GetFile` + `Download` + `Attach` + `Delete` use-cases + 5 group-list endpoints (`ListFiles` / `ListByDocument` / `ListByTask` / `ListByAnnouncement` / `GetVersions`). Previously any authenticated user could fetch ANY file by ID regardless of ownership / association. Action constants `FileAction*` define stable audit keys.
>   - **ADR-2 (TIER 0) Version endpoints gated** — `CreateVersion` / `DeleteVersion` / `DownloadVersion` (all 3 previously open) now route through `AuthorizeFileAccess`; version comment sanitized via shared input-validation guard.
>   - **ADR-3 (TIER 0) Validator wired full path** — `internal/shared/infrastructure/storage/file_validator.go` formerly defined but unused (dead code per audit). Wired в Upload pipeline: MIME-type whitelist + size cap + extension match; octet-stream loophole (any-binary bypass) closed.
>   - **ADR-5 (TIER 1) Admin-gate cleanup endpoint** — `POST /api/files/cleanup` (orphan reaper) was open to authenticated callers; now wrapped через `RequireRole(SystemAdmin)`.
>   - **Reviewer Tier 0/1 absorbs** (round 1 REJECT): test middleware mirror к production gin context contract; Upload validator consistency; dead `PermissionError` branch removed.
>   - **Reviewer Tier 0 regression fix** (round 2 FIX-CYCLE): validator overshoot rejected 7+ legitimate MIMEs (narrowed whitelist back); file-level filter dropped (blocked legitimate cross-uploader visibility — replaced by group-level ownership check); err mapping corrected.
>   - **Carry-forward к v0.161.1**: ADR-4 (IsInlineSafeMime + BuildContentDisposition helper extraction к shared pkg) + 12 Tier 2 items (DIP relocation / ValidationError struct → sentinel / MaxBytesReader / per-user quota / UI strings extraction / concrete `*storage.S3Client` → narrow port / typed slices / pre-commit hook bash 3.2 portability / emitAccessDenied shared helper / gateRead/gateWrite decorator / FileActionDeleteVersion semantic split / sniffBufSize const).
>   - **CI flakes recorded**: auth `TestRefreshToken_RotatesAndDetectsReuse/concurrent_refresh` mock SET-NX timing flake (confirmed via rerun on PR-A and PR-C); CodeQL `Analyze (go)` GHA checkout token race (`event=dynamic`, не retriable через API). См. `docs/plans/2026-05-23-v0161-files-security.md`.
> - **Batch 2 hotfix line progress**: **3/5 closed** (auth ✅ + users ✅ + files ✅). Next: v0.162.0 messaging (WebSocket eavesdrop), v0.163.0 announcements (REJECT verdict — student-can-publish).

> **Изменения с 0.159.0 по 0.160.0 (batch 2 audit hotfix #2 — users module Tier 1 security)**:
>
> - **v0.160.0 users Tier 1 security hotfix #283** — 4 ADR (3 TIER 0 + 1 Tier 1) RED→GREEN TDD pairs + 5 reviewer absorb commits across 3 review rounds (FIX-CYCLE 6.5/6 → FIX-CYCLE 6.75/5 → SHIP 8.5/7.5).
>   - **ADR-1 Profile takeover** — `PUT /api/users/:id/profile` accepted ANY caller; audit row wrote `user_id=target` without `actor_user_id`. Domain free function `AuthorizeProfileEdit(actor, target, role)` — actor==target OR system_admin override; `ErrProfileEditForbidden` sentinel → 403; usecase signature gains actorID + actorRole; audit row now records `actor_user_id` + `target_user_id`. Handler reads `user_id` + `role` из gin context (mirror v0.126 wrong-key-bug contract).
>   - **ADR-2 Departments/Positions role gate** — v0.133.0 admin-gate split applied только к `/users`; `/departments` + `/positions` POST/PUT/DELETE открыты для студента. `RegisterDepartmentRoutes` + `RegisterPositionRoutes` в `internal/modules/users/interfaces/http/routes/` split write subgroup behind `RequireRole(system_admin)` (read endpoints permissive — cross-module resolvers + frontend dropdowns зависят на open read surface). `main.go` switched от inline registration к the new registrars.
>   - **ADR-3 Avatar arbitrary write bypass** — `UpdateProfile` accepted any string в Avatar field, persisted как MinIO storage key, signed как presigned URL pointing к HR records/exam reports. Domain free function `ValidateAvatarKey(key, targetID)` prefix-check against `avatars/{targetID}_` (mirror format avatar Upload handler emits); empty key allowed (clear avatar legitimate). `ErrInvalidAvatarKey` → 400.
>   - **ADR-4 DeleteUser + UpdateUserStatus guards** — actor could DELETE `/api/users/<own_id>` (bricks session) or remove the only system_admin (locks org out of recovery); same applied к Status block/deactivate. `AuthorizeUserDelete(actor, target, role, headcount)` two guards: `ErrCannotDeleteSelf` (actor==target unconditional), `ErrLastAdminProtected` (target=admin && headcount<=1). New `UserAccountRepository.CountByRole` method, conditional lookup only when target is admin (no perf hit on common path). `UpdateUserStatus` reuses the same guards when status==inactive/blocked (active bypasses — non-destructive). Both sentinels → 409.
>   - **Reviewer Tier 0 absorbs** (round 2): (a) all 4 new sentinels mapped через `error_mapper.go` (pre-fix returned 500 instead of documented codes); (b) UpdateUserStatus parity with DeleteUser; (c) handler-level integration tests pinning sentinel HTTP status mapping.
>   - **Reviewer Tier 1 absorbs**: denial-audit emission на rejection paths (`update_denied` / `delete_denied` / `status_change_denied` actions with stable reason codes); `NewCachedUserRepository` wrapper-time `CountByRole` check (boot-time invariant failure vs deferred request-time crash); ADR-2 plan doc clarified к system_admin only.
>   - **Reviewer Tier 2 DEFERRED к v0.160.1** (per `feedback_tier2_absorb_same_release` ≤4-item cap, mirror v0.155→.1 / v0.157→.1 precedent): DIP relocation × 3 (user/department/position repos `domain/repositories/` → `application/usecases/`); AuditSink narrow port (concrete `*logging.AuditLogger` → narrow interface); cross-module DI adapter for notifications; `validate:` → `binding:` tag rename sweep; 2 fire-and-forget `context.Background()` goroutines → lifecycleCtx; audit consistency for `UpdateUserRole` + `BulkUpdateDepartment/Position`; direct unit tests для AuthorizeProfileEdit/AuthorizeUserDelete/ValidateAvatarKey free functions.
>   - **Test coverage**: 4 ADR × (RED + GREEN) = 8 TDD pairs; `TestUserUseCase_DeleteUser_Guards` table-driven 5 cases; `TestUserUseCase_UpdateUserStatus_Guards` table-driven 7 cases; 4 handler-level integration tests pinning sentinel HTTP statuses; `TestRegisterOrgRoutes_*` 24+5+5 sub-test matrix for admin gate; `TestNewCachedUserRepository_PanicsOnMissingCountByRole` для invariant.
>   - **Operational impact**: new HTTP error codes surfaced to frontend — `PROFILE_EDIT_FORBIDDEN` (403), `INVALID_AVATAR_KEY` (400), `CANNOT_DELETE_SELF` (409), `LAST_ADMIN_PROTECTED` (409). См. `docs/plans/2026-05-23-v0160-users-security.md`.
> - **Batch 2 hotfix line progress**: **2/5 closed** (auth ✅ + users ✅). Next: v0.161.0 files (4 TIER 0 — worst module of batch — IDOR + dead validator), v0.162.0 messaging (WebSocket eavesdrop), v0.163.0 announcements (REJECT verdict — student-can-publish).

> **Изменения с 0.158.3 по 0.159.0 (batch 2 audit hotfix #1 — auth module Tier 1 security)**:
>
> - **v0.159.0 auth Tier 1 security hotfix #279** — 6 RED→GREEN TDD ADR pairs (8 RED + 8 GREEN коммита) + 1 wiring + 1 FIX-CYCLE absorption commit. Reviewer round-1 FIX-CYCLE 7.4/6 → round-2 SHIP **9.17/9** после T0+T1 absorption. Closes 7 Tier 1 findings из v1.0.0 batch 2 audit:
>   - **ADR-1 revocation bypass** — `JWTMiddlewareWithRevocation` rewritten: revocation check выполняется BEFORE `c.Next()`. Side-effect spy test pins что handler NOT executed для revoked JTI. Раньше `c.Next()` runs handler → status code overwritten к 401 но side-effects committed = logout effectively no-op для всех write endpoints.
>   - **ADR-2 refresh rotation + RFC 6749 §10.4 cascade** — `RevokeIfAbsent` SET NX atomic claim BEFORE generateTokens (eliminates concurrent re-use window) + `RevokeAllForUser` per-user epoch Lua script cascade при reuse detection + `IsRevokedForUser` check на refresh path. 8-goroutine concurrent test passes под `-race`.
>   - **ADR-3a per-account brute-force lockout** — `LoginAttemptTracker` interface (4 methods) + `RedisLoginAttemptTracker` impl (Lua INCR+EXPIRE atomic). `LOGIN_LOCKOUT_THRESHOLD` / `LOGIN_LOCKOUT_WINDOW` env-tunable defaults. Wired через chainable setter `WithLoginAttemptTracking`.
>   - **ADR-3b trusted-proxy CIDR allowlist** — `getRealIPWithTrustedProxies` honors `X-Forwarded-For` ТОЛЬКО когда `RemoteAddr` в CIDR allowlist. Secure-by-default (empty `TRUSTED_PROXY_CIDRS` = ignore XFF). `RateLimiter.WithTrustedProxies` setter. Closes XFF spoofing → IP rate-limit bypass.
>   - **ADR-4a/b MFA secret AES-256-GCM at rest** — `internal/shared/infrastructure/crypto/secretbox.go` (`EncryptString`/`DecryptString` nonce-prepended base64; `ParseKEKHex` 64-hex = 32-byte KEK). `UserRepositoryPG.WithMFASecretKEK` lazy migration: load → if `mfa_secret_encrypted=false` then ciphertext == raw secret (compat); on next save → encrypt + flip flag. Migration 045 ADD COLUMN. Раньше TOTP secret хранился plaintext = DB dump = bypass всего MFA.
>   - **ADR-5 password reset token sha256** — `hashResetToken(rawToken)` SHA-256 + base64 url-safe; applied симметрично на `Store` / `LookupUser` / `Delete`. Raw token никогда не попадает в Redis (DB-dump / Redis-RDB exposure closed).
>   - **ADR-6 GETDEL atomic consume** — `LookupUserAndConsume` использует Redis `GETDEL` atomic command. `ConfirmReset` rewired: consume token BEFORE bcrypt+`Save` → если Save fails (FK / unique / panic), token уже deleted = no replay window. Раньше `Lookup` + `Save` + `Delete` 3-step pattern давал race window для concurrent reset.
>   - **Tier 2 absorbed в release**: dead `Session` entity + repo + tests (`domain/entities/session.go`, `infrastructure/persistence/session_repository*.go`) полностью deleted — ADR-2 использует `RevokedTokenRepository` без consumer; dead in-memory `RateLimiter`/`RateLimitMiddleware` (+ tests) removed (production = Redis-backed shared limiter); `domain/services/auth_service.go` deleted ENTIRELY (JWTService/AuthorizationService/Scope/TokenPair все dead per grep — T1-3); `RegisterInput.Role` `binding:"omitempty,oneof=student teacher"` — privileged-role attempts fail at Gin boundary 400, не usecase 403; `entities/user.go` sentinels `fmt.Errorf` → `errors.New` (DDD gate); `RegisterFailure` errors теперь audit-emit (T1-1); split `WithRefreshRotation` setter от `WithMFAVerification` (T1-4 explicit wiring intent); stale `SessionRepository` doc comments (`logout_usecase.go`, main.go:1626/1768) rewritten.
>   - **main.go wiring**: `ParseTrustedProxyCIDRs(os.Getenv("TRUSTED_PROXY_CIDRS"))` на both public + auth limiters; `crypto.ParseKEKHex(os.Getenv("MFA_SECRET_ENC_KEY"))` + `baseUserRepo.WithMFASecretKEK(kek)` с ENABLED/DISABLED startup log; `LOGIN_LOCKOUT_THRESHOLD`/`LOGIN_LOCKOUT_WINDOW` env → `NewRedisLoginAttemptTracker` + `WithLoginAttemptTracking`; both `WithMFAVerification` AND `WithRefreshRotation` invoked explicitly когда Redis available.
>   - **Remaining T2 (follow-up, NOT blocking SHIP)**: reviewer round-2 T2-NEW — `IsRevokedForUser` consulted ТОЛЬКО на refresh path, не в `JWTMiddlewareWithRevocation` для access tokens. После cascade revoke currently-live access tokens (15min lifetime) остаются valid до natural expiry. Defense-in-depth follow-up, deferred.
>   - 16 RED/GREEN commits + 1 absorb commit, `.env.example` sync, migration 045 up/down. См. `docs/plans/2026-05-21-v0159-auth-security.md`.
> - **Batch 2 hotfix line progress**: 1/5 закрыт (auth). Next: v0.160.0 users (3 TIER 0 profile takeover) → v0.161.0 files (4 TIER 0 IDOR + dead validator) → v0.162.0 messaging (WebSocket eavesdrop) → v0.163.0 announcements (REJECT verdict — student-can-publish).

> **Изменения с 0.157.1 по 0.158.0 (curriculum role swap — business-logic correctness fix per diploma role matrix)**:
>
> - **v0.158.0 curriculum role swap** — academic secretary (Волкова) теперь curriculum AUTHOR end-to-end (план + sections + discipline items + submit); methodist — APPROVER (approve/reject pending_approval); system_admin сохраняет emergency override на обеих сторонах. Предыдущая wiring (methodist author, system_admin approver) была некорректна по project spec.
>   - Backend whitelist swap: `canWrite(role)` методист→academic_secretary; `canApprove(role)` system_admin→methodist || system_admin (handler-level); route group `RequireRole(SystemAdmin)` → `RequireRole(Methodist, SystemAdmin)` (defense in depth).
>   - Frontend `PERMISSION_MATRIX`: METHODIST.CURRICULUM FULL→LIMITED (read + approve only); ACADEMIC_SECRETARY.CURRICULUM LIMITED→FULL (full author cycle); `can()` APPROVE branch теперь methodist || system_admin; `CURRICULUM_WRITE_ROLES = [SYSTEM_ADMIN, ACADEMIC_SECRETARY]`.
>   - Test maintenance: handler tests rewrite (`Approve_HappyPath_Admin` → `_AuthorizedRoles` table-driven covering methodist + admin); `Approve_RejectsNonAdminRoles` → `_RejectsNonApproverRoles` (drops methodist from rejection list); bulk sed `"methodist"` → `"academic_secretary"` в всех handler tests где role = author; frontend `permissions.test.ts` + `permission-matrix-integration.test.ts` + `page.test.tsx` rewrites.
>   - Zero DB schema changes — `created_by` сохраняет historical user_id; UI отображает author by user id, не by role.
>   - См. `docs/plans/2026-05-20-v0158-curriculum-role-swap.md`.

> **Изменения с 0.157.0 по 0.157.1 (curriculum DIP polish — closes ADR-1 carry-forward от #269)**:
>
> - **v0.157.1 curriculum DIP cleanup chore release** — mechanical refactor per CLAUDE.md gate ("Repository interfaces — в пакете-потребителе (`usecase/`), НЕ в `domain/`"). Закрывает ADR-1 carry-forward от v0.157.0 hotfix #269 (deferred к polish patch по precedent v0.155 ADR-5 → v0.155.1).
>   - **5 wide repository ports relocated**: `CurriculumRepository`, `SectionRepository`, `DisciplineItemRepository`, `BulkDisciplineItemsUnitOfWork`, `BulkDisciplineItemsTx` — из `internal/modules/curriculum/domain/repositories/` к single consolidated `internal/modules/curriculum/application/usecases/repository_interfaces.go`. Zero behavior change.
>   - **Sentinels + query DTOs остаются в `domain/repositories/`** (intentional): sentinels `ErrCurriculumNotFound`/`ErrCurriculumCodeExists`/`ErrCurriculumVersionConflict`/`ErrSection*`/`ErrDisciplineItem*`/`ErrBulkTxFinished` — domain values; query DTOs `CurriculumListFilter`/`CurriculumListResult`/`CurriculumYearSpecialtyAgg`/`DisciplineItemHoursAgg` — потребляются cross-module `reports/annual` (producer-module DTO ownership pattern).
>   - **13 qualified reference sites updated** (plan doc undercount 9 → refined к 13 в exec): 4 в `infrastructure/persistence/bulk_unit_of_work_pg.go` (Begin + Items/Sections/Curricula accessor returns), 1 в `application/usecases/bulk_edit_discipline_items_usecase.go` (narrow-port Begin return), 8 в `bulk_edit_discipline_items_usecase_test.go` (collapsed к unqualified, same-package).
>   - **Tier 2 absorbs в release commit** (per `feedback_tier2_absorb_same_release`): (a) plan doc count fix; (b) docstring slip в `repository_interfaces.go` header; (c) **4 compile-time `var _ usecases.<Port> = (*XxxPG)(nil)` assertions** в `*_pg.go` files — signature drift catches at infra compile site, не at DI wiring.
>   - Reviewer single-pass SHIP mean 8.78/10 (post-absorb ~9.0+). См. `docs/plans/2026-05-20-v0157-1-dip-cleanup.md`.
> - **Batch 1 v1.0.0 audit line completely closed** (4 модуля Tier 1 + ADR-1 polish). Next milestone: batches 2-4 reviews для 13 remaining modules.

> **Изменения с 0.156.0 по 0.157.0 (batch 1 audit hotfix #4 — curriculum module, partial)**:
>
> - **v0.157.0 curriculum Tier 1 security hotfix #269** — 1 RED→GREEN TDD pair + 4 reviewer Tier 1 absorbs + 3 Tier 2 absorbs. Closes 1 of 2 Tier 1 audit findings (ADR-2 lost-update race); ADR-1 (repository interface relocation, pure DIP cleanup) explicitly deferred к v0.157.1 polish patch (same precedent as v0.155 ADR-5 → v0.155.1).
>   - **ADR-2: curricula optimistic-lock** — `curricula` table previously had no `version` column (migration 031); `CurriculumRepositoryPG.Update` used bare `WHERE id = ?` so methodist A + B concurrent edits could silently lose first writer's change. Section + DisciplineItem aggregates already used optimistic lock (migrations 034/035 + v0.128.0+); aggregate root needed к catch up.
>   - migration 044 ADD COLUMN version INT NOT NULL DEFAULT 0 + idempotent CHECK constraint (DO/EXCEPTION block для partial-apply safety per reviewer Tier 1 absorb).
>   - Curriculum entity: `version` field + `Version()` accessor + `BumpCurriculumVersion` exported helper + `ReconstituteCurriculum` signature gains trailing `version int` param (6 sites updated).
>   - CurriculumRepositoryPG: `SELECT` columns + `Update WHERE id = ? AND version = ?` clause; `disambiguateAbsentUpdate` follow-up SELECT 1 distinguishes `ErrCurriculumVersionConflict` (row exists, stale) от `ErrCurriculumNotFound` (row vanished). New sentinel в `domain/repositories`.
>   - **Critical reviewer absorbs (Tier 1)**: (a) `update_curriculum_usecase.go` now `errors.Is` branches на VersionConflict + emits audit `curriculum.update_denied` reason=version_conflict (was missing); (b) `curriculum_handler.mapCurriculumError` adds VERSION_CONFLICT case → HTTP 409 (was falling через к 500 — defeated entire optimistic-lock goal); (c) handler 409 mapping test added; (d) sqlmock WithArgs pinned для mutation-resistance.
>   - Reviewer round-1 FIX-CYCLE 7.17/5 → round-2 SHIP expected после fix-cycle absorb (4 Tier 1 + 3 Tier 2 closed). См. `docs/plans/2026-05-20-v0157-curriculum-security.md`.
>
> **Изменения с 0.155.0 по 0.156.0 (batch 1 audit hotfix #3 — documents module)**:
>
> - **v0.156.0 documents Tier 1 security hotfix #266** — 6 RED→GREEN TDD pairs + 1 config-fix + 1 reviewer Tier 2 absorb + release. Closes 7 Tier 1 audit findings: (1) **sentinel error mismatch** — `ErrDocumentNotFound` + `ErrVersionNotFound` declared в `domain/repositories`; 7 repo sites routed (was `fmt.Errorf("document not found")` → 500 fallthrough в workflow audit); (2) **clickjacking via `?inline=true`** — `IsInlineSafeMime` whitelist (image/png|jpeg|gif|webp|svg+xml, application/pdf, text/plain) + CSP `frame-ancestors 'self'` + `X-Frame-Options SAMEORIGIN` (was wildcard `*` для ANY auth download); (3) **header injection** — `BuildContentDisposition` via `mime.FormatMediaType` (RFC 2231 `filename*=utf-8''<percent>`) + control-byte strip (CRLF / NUL / DEL); (4) **magic `"admin"` string** — `IsAdminRole` helper recognizes both legacy `"admin"` + canonical `"system_admin"` case-insensitive; admin bypass restored (was production JWT carrying `"system_admin"` caught by access-control WHERE); (5) **cross-module import + fire-and-forget** — `ShareNotifier` narrow port в `application/usecases`; chainable `WithShareNotifier(n)` setter; adapter `documentsShareNotifier` в main.go DI seam (mirror `assignmentsGradeNotifier` pattern); Russian UI strings + concrete cross-module dep + `context.Background()` goroutine all closed одним рефактором; (6) **UploadFile auto-register bypass removed** — 3-line direct mutation `doc.Status = Registered` skipped `Document.Register` invariants (number ≥3 chars / status==Approved / audit triplet atomic); workflow now owns transitions; (7) **Russian strings → tag sentinels** — `ErrTagNotFound` + `ErrTagAlreadyExists` chain через shared `domainErrors.ErrNotFound`/`ErrAlreadyExists`; 11 fmt.Errorf sites replaced; `response.MapDomainError` now fires correct 404/409. Tier 2 absorb: notifier failure now emits audit event `document_share_notification_failed` so dropped notifications observable в `/admin/audit-logs`. Reviewer single-pass SHIP 8.67/8. См. `docs/plans/2026-05-20-v0156-documents-security.md`.
>
> **Изменения с 0.154.0 по 0.155.0 (batch 1 audit hotfix #2 — ai module)**:
>
> - **v0.155.0 ai Tier 1 security hotfix #263** — 1 RED→GREEN TDD pair + 7 config/wiring fixes + 1 reviewer fix-cycle absorb commit + release. Closes 7 of 8 Tier 1 audit findings: (1) **`CreateConversation` broken shortcut** fixed — actually persists через `conversationRepo.Create` (was returning first existing conv as "newly created"); (2) **`validate:` → `binding:` tag rename** на 5 DTOs incl. previously untagged `CreateConversationRequest.Title`/`Model` (Gin reads binding, not validate — Max=10000 на Content was ignored = token-cost DoS surface); (3) **per-user rate limit** — new `RateLimitByUserMiddleware()` keys на `rate_limit:user:%d` mounted на `aiAPIGroup` (was IP-only с NAT shared bucket + X-Forwarded-For trust); (4) **schedulers wire lifecycle ctx + `Stop()`** — `factScheduler` + `indexingScheduler` registered со shutdown, `fact_scheduler` sleep wraps `select` on `ctx.Done()` + 2 tick-cancel pin tests; (5) **`defaultOpenAIChatModel = "gpt-4o-mini"`** (was Gemini hardcode in OpenAI provider); (6) **`getAuthedUserID(c)` helper** rolled out to all 8 chat-flow handlers (silent-zero closed); (7) **`ErrConversationNotFound` + `ErrConversationAccessDenied` sentinels** в `ai/domain/repositories` + `respondChatError` maps к 404/403 без `err.Error()` leak — 6 existing tests migrated к `errors.Is`. ADR-5 cross-module imports refactor deferred к v0.155.1 polish. Reviewer round-1 FIX-CYCLE 7.17/6 → round-2 SHIP expected после fix-cycle absorb (4 Tier 1 + 4 Tier 2 closed). См. `docs/plans/2026-05-20-v0155-ai-security.md`.
>
> **Изменения с 0.153.13 по 0.154.0 (v1.0.0 path opens, batch 1 audit)**:
>
> - **v0.154.0 reporting Tier 1 security hotfix #260** — 5 RED→GREEN TDD pairs + 1 config-fix + plan doc + release commit. Closes 4 Tier 1 audit findings: (1) **SQL injection** через user `field.Alias` JSON closed by **3-layer defense-in-depth** (domain `SelectedField.Validate` + persistence `SetFieldsFromJSON` + query builder `Execute` re-validation; PG identifier whitelist regex `^[A-Za-z_][A-Za-z0-9_]{0,62}$`); (2) **privilege escalation** on `/api/custom-reports` group closed via `RequireNonStudent()` gate; (3) **shipped fake report generation** — `simulateGeneration` removed, `Generate` returns `ErrGenerationNotImplemented` mapped to HTTP 501 + audit log emission; (4) `context.Background()` goroutine leak partially closed. 21 table-driven sub-tests + 4 struct-literal bypass tests. Reviewer SHIP 8.83/8 single-pass. См. `docs/plans/2026-05-20-v1.0.0-batch1-audit.md` + `docs/plans/2026-05-20-v0154-reporting-security.md`.
>
> **Изменения с 0.152.0 (Documents workflow Phase 5) по 0.153.11 (Phase 6 strict-90)** — 13 релизов в v0.153.x sprint + Phase 5 closure. Самое важное:
>
> - **Documents workflow #227 5-phase pack CLOSED end-to-end** (v0.148.0 → v0.152.1): полный lifecycle draft → review → registered → routing → execution → executed → archived/resubmitted. 11 transition endpoints за `RequireRole(AcademicSecretary, SystemAdmin)`, миграции 040-043, 5 nullable audit fields per phase, i18n × 4 для каждой dialog.
> - **Phase 5 admin observability CLOSED** end-to-end:
>   - #1 **Audit logs** (v0.130.0+v0.131.0+v0.131.1) — persistence + read API + `/admin/audit-logs` UI + coverage gaps for messaging/integration sinks.
>   - #2 **Backup observability** (v0.132.0) — read-only `/admin/backup` UI поверх sidecar (list files + Prometheus metrics + secured download).
>   - #3 **Sentry + `/admin/users`** (v0.133.0) — TIER 0 security gap closure (RequireRole split) + admin observability bundle.
>   - #4 **VAPID + n8n + Branding** (v0.134.0+v0.136.0+v0.137.0) — config view + DB-backed singleton brand settings + login page integration via BrandedHeader.
>   - #5 **Composio** (v0.135.0) — read-only config view; SetReminder cross-module Composio scheduling deferred.
> - **B4 Annual methodist report** (v0.129.0+v0.129.1) — cross-module DOCX orchestrator + pure-stdlib synth. **B-feature triad замкнут** (curriculum + assignments + documents → annual report consumer).
> - **Phase 6 #196 backend coverage strict > 90.0%** (v0.153.0 → v0.153.11, 12 releases) — **85.0% → 90.2%** (+5.2pp). Strict gate ✅. Patterns codified: embedded narrow interface fakes, real `*UseCase` + fake repo для DIP-blocked handlers, httptest.Server для odata client tests, sqlmock для persistence batch, setter pattern для optional UC deps.
> - **Security baseline**: SECURITY.md + 7 GitHub Security toggles + CodeQL default setup (v0.128.8+v0.128.10). **Dependabot alerts 44 → 0**, **CodeQL 34 → 0**, **secret-scanning 2 → 0**. All transitive npm vulns закрыты через `overrides` в package.json (ws + tootallnate sweeps).
> - **CI/CD cleanup**: ci.yml 262 → 96 LoC (v0.128.5), pre-commit hook (v0.127.0, AmE/BrE + golangci + prettier + eslint), CodeQL/secret-scanning resolution flow codified.
> - **B1a Section aggregate initiative CLOSED 5/5** (v0.128.0-v0.128.4) — bulk-edit РПД с UnitOfWork RepeatableRead + frontend useReducer table view UI.
> - **MFA TOTP enrollment + login flow** (v0.124.0+v0.125.0+v0.125.2+v0.125.3) — RFC 6238 self-implemented (supply-chain block на `pquerna/otp`), intermediate JWT (5min) + verify-login endpoint + frontend MFAVerifyLoginStep + status-aware error mapping. Полностью end-to-end UI для system_admin.
> - **Templates filter teacher-own** (v0.126.0+v0.126.3) — DocumentType.MethodistOnly + CanAccessByRole + migration 033. Полный end-to-end (backend + UI toggle).

---
## Содержание

1. [Роли в системе](#роли-в-системе)
2. [Личные vs глобальные настройки](#личные-vs-глобальные-настройки)
3. [Что РАБОТАЕТ полностью](#что-работает-полностью)
4. [Что РАБОТАЕТ частично](#что-работает-частично)
5. [Что НЕ РАБОТАЕТ (заглушки)](#что-не-работает-заглушки)
6. [Сценарии по ролям](#сценарии-по-ролям)
7. [Открытые задачи](#открытые-задачи)

---

## Роли в системе

В коде определено **5 ролей** (`internal/modules/auth/domain/role.go`):

| Роль | Код | Назначение |
|------|-----|------------|
| **Системный администратор** | `system_admin` | Полное управление системой, **все системные настройки и интеграции** |
| **Методист** | `methodist` | Методическое обеспечение учебного процесса (без системных настроек) |
| **Академический секретарь** | `academic_secretary` | Административное сопровождение |
| **Преподаватель** | `teacher` | Реализация образовательного процесса |
| **Студент** | `student` | Участие в образовательном процессе (view-only) |

---

## Личные vs глобальные настройки

С версии 0.102.2 в системе чётко разграничены два уровня настроек.

### Личные настройки — доступны ВСЕМ ролям

Любой авторизованный пользователь, независимо от роли, может настраивать своё рабочее окружение. Эти настройки применяются только к данному пользователю и не влияют на других:

| Настройка | Что делает | Где |
|-----------|------------|-----|
| **Выбор темы оформления** | Переключение между светлой и тёмной темой | `Profile` → Appearance |
| **Подключение каналов уведомлений** | Привязка Telegram, выбор email/WebPush, тестовое уведомление | `Profile` → Notifications |
| **Привязка Telegram** | Получение auth-токена бота, верификация | `Profile` → Notifications → Telegram |
| **Тестовое уведомление** | Проверка доставки по выбранному каналу | при настройке канала |
| **Переключение языка UI (i18n)** | ru/en/fr/ar (RTL для арабского) | `Profile` → Language |
| **Редактирование своего профиля** | имя, контакты, фото | `/profile` |

Это базовая функция, такая же тривиальная как «выйти из системы» — никаких особых прав не требует.

### Глобальные настройки — только `system_admin`

| Настройка | Что делает | Где |
|-----------|------------|-----|
| **Глобальная тема и brand** | Корпоративная цветовая схема, логотип, fav icon — применяется ко ВСЕМ | `/admin/settings/appearance` |
| **Глобальные настройки уведомлений** | SMTP-сервер, push VAPID-ключи, токен Telegram-бота | `/admin/settings/notifications` |
| **Управление n8n workflows** | 3 workflow: уведомления документов, алерты пропусков, напоминания дедлайнов | `/admin/settings/automation` |
| **Интеграция с 1С:Университет** | Настройка соединения, маппинг, синхронизация сотрудников/студентов | `/admin/integration` |
| **Управление пользователями (CRUD)** | Создание/редактирование/удаление, назначение ролей | `/admin/users` |
| **Утверждение учебных планов** | `ActionApprove` — admin override; основной approver — методист (v0.158.0+) | `/admin/curriculum/approve` |
| **Backup, логи, метрики, алерты** | Эксплуатация системы | `/admin/infra/*` |

**Принцип**: всё, что является системной настройкой и влияет на работу системы для всех пользователей или на её взаимодействие с внешним миром — это исключительно admin.

### Матрица доступа (PermissionMatrix)

| Ресурс | system_admin | methodist | academic_secretary | teacher | student |
|--------|:------------:|:---------:|:------------------:|:-------:|:-------:|
| **users** (CRUD) | full | read limited | read limited | read limited | own update |
| **curriculum** (учебные планы, v0.158.0+) | full + approve override | read + approve/reject | **full author cycle** (create / edit / submit) | read+limited update | read limited |
| **schedule** (расписание) | full | read+limited | full | read | read |
| **assignments** (задания) | full | full+limited | read | full+own | own read+execute |
| **reports** (отчёты) | full | full | full | limited | denied |
| **integration** (1С) | **full** | denied | denied | denied | denied |
| **system_settings** (глобальные) | **full** | denied | denied | denied | denied |
| **personal_settings** (свои) | **own** | **own** | **own** | **own** | **own** |

Уровни: `denied < limited < own < full`.

---

## Что РАБОТАЕТ полностью

Backend + Frontend + API + проверено в use-flow.

| Модуль | Backend (LOC) | Frontend | API |
|--------|:-------------:|:---------|:---:|
| **auth** | 2046 | `/login` (+ MFA verify step для admin'а), `/register`, `/forgot-password`, `/admin/settings/security` (MFA enrollment, system_admin only) | ✅ |
| **users** | 2757 | `/users`, `/profile`, `/users/[id]` | ✅ |
| **documents** | 8392 | `/documents`, `/documents/templates`, `/documents/shared` | ✅ |
| **dashboard** | 859 | `/dashboard` (агрегатор виджетов) | ✅ |
| **notifications** | 5666 | `/notifications`, Telegram, Slack, WebPush, Email | ✅ |
| **messaging** | 3521 | `/messages`, `/messages/[id]` (WebSocket) | ✅ |
| **reporting** | 6628 | `/reports`, `/reports/builder` | ✅ |
| **integration** | 5557 | `/integration` (синк 1С) — **только admin** | ✅ |
| **analytics** | 2430 | `/analytics` (риски студентов, тренды) | ✅ |
| **ai** | 5837 | `/ai` (RAG-чат с цитированием) | ✅ |
| **assignments** (академические задания) | ~3500 | `/assignments` (список + grading), `/assignments/[id]/submissions`, `/my-assignments`, `/my-assignments/[id]` (студенту) | ✅ |
| **curriculum** (учебные планы) | ~2800 | `/curriculum` (список с фильтрами), `/curriculum/[id]` (детали + edit + submit), `/admin/curriculum/approve` (admin queue с Approve/Reject), bulk-edit РПД РПД с UnitOfWork | ✅ |
| **schedule** (расписание + lessons) | ~5800 | `/schedule` timetable grid + `/calendar` events + week-type табы | ✅ |
| **tasks** (project management) | ~1900 | `/tasks` + Telegram/email/push reminders | ✅ |
| **announcements** | ~2900 | `/announcements` (CRUD + attachments) | ✅ |
| **reports/annual** (B4) | ~700 | `/reports/annual` (year selector + DOCX download — methodist+admin) | ✅ |
| **workflow** (документооборот) | в `documents/` | dialogs: Submit/Approve/Reject/Register/StartRouting/SignVisa/AssignExecutor/MarkExecuted/Archive/Resubmit — все 9 transitions с i18n × 4 | ✅ |
| **audit logs** | ~1100 | `/admin/audit-logs` (фильтры + пагинация) | ✅ |
| **branding admin** | ~600 | `/admin/branding` (DB-backed singleton + login page BrandedHeader) | ✅ |
| **MFA TOTP** | в `auth/` | `/admin/settings/security` enrollment + login flow MFA verify step (system_admin) | ✅ |
| **files** | ~1933 | прикрепляются к документам/задачам/объявлениям через UI; standalone file manager не реализован | partial |

### Инфраструктура (всё работает, эксплуатацию ведёт админ)

- **OpenTelemetry tracing** — Tempo + OTEL Collector, корреляция логов с trace_id
- **n8n automation** — 3 workflow: уведомления документов, алерты пропусков, напоминания дедлайнов (управляет admin)
- **Loki** — централизованное логирование через Grafana
- **Grafana Alerting** — 7 алертов в Telegram (CPU, RAM, диск, latency, errors, backup)
- **Web Push** — VAPID + Service Worker
- **Uptime Kuma** — status page (Caddy proxy не настроен)
- **Backup** — PostgreSQL + MinIO в S3 (offsite не подключён)
- **PWA** — Service Worker, offline page
- **i18n** — ru/en/fr/ar (RTL для арабского)

---

## Что РАБОТАЕТ частично

| Модуль | Backend | Frontend | Что отсутствует |
|--------|:-------:|:--------:|-----------------|
| **files** | ✅ 1933 LOC | ⚠️ только через attachments | Standalone file manager UI (defer post-defense) |
| **SetReminder cross-module** | Phase 5 #5 partial (Composio view shipped) | — | Composio TELEGRAM_SEND_MESSAGE для task deadline reminders (deferred) |

**Закрыто в недавних релизах:**
- **~~workflow (согласование)~~** — реализовано как часть documents модуля (по DDD), 9 transitions полностью end-to-end, **GH #41 closed 2026-05-19**
- **~~schedule полный CRUD~~** — расписание пар (events + lessons + замены + справочники) полностью отгружено, UI на `/schedule` + week-type табы
- **~~tasks~~** — GH [#200](https://github.com/.../issues/200) в **0.101.0**
- **~~announcements~~** — GH [#202](https://github.com/.../issues/202) в **0.102.0**
- **~~admin-permissions-rebalance~~** — внутреннее изменение в **0.102.1**: интеграция 1С → admin
- **~~personal-settings-clarification~~** — **0.102.2**: личные настройки доступны всем ролям

---

## Что НЕ РАБОТАЕТ (заглушки)

| Модуль | Состояние | GitHub |
|--------|-----------|:------:|
| **Электронная подпись** | Не начато — УКЭП/УНЭП, КриптоПро | [#140](https://github.com/.../issues/140) |
| **Авто-расписание** | Не начато — CSP алгоритм | [#139](https://github.com/.../issues/139) |
| **Внешние календари** | Не начато — Google Calendar, Outlook, iCal | [#40](https://github.com/.../issues/40) |
| **Web Speech API** | Не начато — голосовой ввод/вывод в AI-чате | TM #23 |

> **Закрыто в 2026-05:** ~~workflow (согласование документов) [#41]~~ — реализовано как поведение агрегата Document (по DDD): 9 transition endpoints в `documents/interfaces/http/handlers/workflow_handler.go` (submit / approve / reject / register / start-routing / sign-visa / assign-executor / mark-executed / archive / resubmit) за `RequireRole(AcademicSecretary, SystemAdmin)`. Полный lifecycle draft → review → registered → routing → execution → executed → archived/resubmitted с frontend dialogs для каждого transition + i18n × 4. См. 5-phase pack #227 (v0.148.0 → v0.152.1).

---

## Сценарии по ролям

> **Личные настройки опускаются в каждом сценарии** — они одинаковы у всех ролей и описаны выше в [Личные vs глобальные настройки](#личные-vs-глобальные-настройки).

### 🔓 Гость (неавторизованный)

**Доступные страницы:** `/`, `/login`, `/register`, `/forgot-password`, `/reset-password`, `/forbidden`, `/offline`

1. Зарегистрироваться (`/register`):
   - Поля: email, пароль, имя
   - **Выбор роли:** только `student` или `teacher` (whitelist для self-registration)
   - После регистрации — auto-login и редирект на `/dashboard`
2. Войти (`/login`) — JWT в httpOnly cookie + sessions в БД
3. Восстановить пароль (если email-сервис настроен)

**🔐 Защита от privilege escalation (фикс GH #199, 2026-04-25):** глубинная защита в 4 слоях (Domain / Usecase / Handler / Frontend).

---

### 👨‍🎓 Студент (`student`)

**Видит в меню:** Dashboard, Documents (просмотр), Schedule (просмотр), Calendar, **My Assignments** (свои работы), Tasks, Announcements, Messages, AI Assistant, Profile

1. Регистрация → авто-логин
2. **Dashboard** — виджеты: ближайшие события, объявления, непрочитанные сообщения
3. **Documents** — только чтение публичных и доступных документов
4. **Schedule** — **просмотр расписания** своей группы (сетка по дням, фильтр по группе/преподавателю)
5. **Calendar** — свои события, расписание группы
6. **Messages** — WebSocket-чаты
7. **AI Assistant** — RAG с цитированием
8. **Tasks** — просмотр заданий (own read+execute)
9. **Announcements** — просмотр объявлений
10. **Мои работы (полный flow с 0.114.0):** `/my-assignments` (список своих submission'ов с status-фильтром: all / pending / graded / returned) + `/my-assignments/[id]` (детальный view с status-aware panel — оценка/feedback или причина возврата). Backend (0.113.0): GET `/api/assignments/my` + GET `/api/assignments/:id/my` за `RequireRole("student")`.
11. **Resubmit на assignments (полный flow с 0.115.0):** на detail page `/my-assignments/[id]` для status='returned' доступна кнопка «Пересдать работу» → `ResubmitDialog` (confirm/cancel) → `POST /api/assignments/:id/resubmit` → status flips на pending. Backend (0.112.0) ownership invariant `Submission.AuthorizeResubmitter` отклоняет попытку пересдать чужую работу с 403 + audit `assignment.resubmit_denied`.
12. *(Личные настройки — стандартно для всех ролей)*

**Что НЕ может:** создавать/редактировать расписание, создавать документы, отчёты (`denied`), аналитика, управление пользователями, любые системные настройки. **Не может пересдавать чужие работы:** ownership проверяется на entity-уровне через `Submission.AuthorizeResubmitter` — попытка пересдать работу другого студента отклоняется с 403 + audit `assignment.resubmit_denied`.

---

### 👨‍🏫 Преподаватель (`teacher`)

**Видит в меню:** Dashboard, Documents (full), Schedule (просмотр), Calendar, Tasks, **Assignments** (grading), **Curriculum** (read), Announcements, Messages, AI Assistant, Users (limited), Profile

1. Регистрация / создание администратором
2. **Dashboard** — виджеты: задания на проверку, ближайшие пары
3. **Documents** — создание/редактирование своих, шаблоны (read-only), маршруты согласования
4. **Schedule** — **просмотр расписания** своих пар (сетка по дням, фильтр по группе/аудитории)
5. **Calendar** — создание событий, назначение участников
6. **Users** — список студентов своих групп (read limited)
7. **Reports (limited)** — по своим группам, экспорт limited
8. **Assignments (полный grading flow с 0.110.0–0.115.0):** `/assignments` (список своих заданий с фильтрами subject/group_name) + `/assignments/[id]/submissions` (inline grade form per submission row, status-фильтр pending/graded/returned). Может вернуть работу через `ReturnDialog` (с textarea причины ≤ 4096 символов) — `Submission.Return` очищает grade triple, статус submission → returned, audit `assignment.returned` сохраняет previous_grade. Студент пересдаёт, учитель re-grade'ит.
9. **Curriculum (read+limited update с 0.118.0+):** `/curriculum` (список с фильтрами по статусу/году/специальности) + `/curriculum/[id]` (детали с status pill). Read-only для учителя; редактирование закрыто `AuthorizeEdit` гейтом (только academic_secretary или admin — v0.158.0+).
10. **Messages** — групповые чаты со студентами
11. **AI Assistant** — расширенные права на RAG
12. *(Личные настройки — стандартно)*

**Что НЕ может:** создавать/редактировать расписание, видеть отчёты других преподавателей, **редактировать curriculum** (только read), grade'ить чужие assignments (`Assignment.AuthorizeGrader` — только автор), создавать пользователей, любые системные настройки.

---

### 📋 Академический секретарь (`academic_secretary`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group (Reports + Analytics), Schedule, Calendar, Tasks, **Assignments** (read), **Curriculum** (full author cycle), Announcements, Messages, AI Assistant, Admin group (Users — read limited), Profile

1. Создание администратором
2. **Dashboard** — административные виджеты
3. **Documents + Templates** — full CRUD, шаблоны (создание/редактирование)
4. **Schedule** — **полное управление расписанием** (создание пар, замены, аудитории)
5. **Reports** — full create/read/export
6. **Analytics** — просмотр аналитики студентов (риски, посещаемость, успеваемость)
7. **Users** — read limited
8. **Calendar** — управление событиями
9. **Assignments (read с 0.110.0):** `/assignments` (список всех заданий, не только своих — caller scope unrestricted) + `/assignments/[id]/submissions` (просмотр работ студентов). Может вернуть работу через `ReturnDialog` (`AuthorizeGrader` принимает 4 non-student роли в read-only сценарии; grading закрыт за teacher's ownership).
10. **Curriculum (полный author cycle с v0.158.0):**
    - `/curriculum` — список всех учебных планов с фильтрами status/year/specialty + кнопка **Создать** (academic_secretary + admin) + цветной status pill
    - `/curriculum/[id]` — детали с status-aware панелью: для status='draft' доступны **Редактировать** + **Отправить на утверждение**; для pending/approved/archived — read-only с подсказкой почему
    - **CreateCurriculumDialog** / **EditCurriculumDialog** (Radix modal с 5 полями: title / code / specialty / year ∈ [2000, 2100] / description ≤ 4096) — client-side валидация зеркальная domain invariants, error mapping 409→codeExists / 422→notEditable / 403→forbidden
    - **Sections + DisciplineItems (РПД)** — полный CRUD + bulk-edit таблица с UnitOfWork RepeatableRead транзакцией
    - **SubmitCurriculumDialog** — confirmation modal для перехода draft → pending_approval. После confirm учебный план уходит к методисту на утверждение; редактирование блокируется до решения
    - **Утверждение запрещено** — `ActionApprove` принадлежит методисту. Если методист отклоняет с reason — учебный план возвращается в draft, секретарь видит причину в audit log + UI feedback, правит и отправляет повторно
11. **Messages, AI** — стандартно
12. *(Личные настройки — стандартно)*

**Что НЕ может:** утверждать собственные учебные планы (`ActionApprove` → methodist + admin), создавать пользователей, подписывать задания, любые системные настройки.

---

### 📚 Методист (`methodist`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group, Schedule, Calendar, Tasks, **Assignments** (read), **Curriculum** (read + approve), Announcements, Messages, AI Assistant, Users (read limited), Profile

1. Создание администратором
2. **Dashboard** — методические виджеты
3. **Documents + Templates** — full CRUD, создание шаблонов документов
4. **Curriculum (approver role с v0.158.0):**
   - `/curriculum` — список всех учебных планов с фильтрами status/year/specialty + цветной status pill (черновик / на утверждении / утверждён / архив). Read-only — методист не создаёт планы, это работа академического секретаря
   - `/curriculum/[id]` — детали без edit/submit кнопок (это author's surface)
   - `/admin/curriculum/approve` — **очередь pending_approval** с **ApproveCurriculumDialog** + **RejectCurriculumDialog** (reason mandatory). После approve учебный план переходит в approved; после reject — возвращается в draft с reason в audit log, секретарь видит причину и правит
   - Bulk-edit РПД доступен только для read (academic_secretary редактирует)
5. **Reports + Analytics** — full доступ, экспорт CSV/XLSX
6. **Schedule** — read full + limited update
7. **Assignments (read с 0.110.0):** просмотр всех заданий и работ студентов — caller scope unrestricted для методиста
8. **Users** — read limited
9. **AI Assistant** — расширенные права
10. **Calendar, Messages** — стандартно
11. *(Личные настройки — стандартно)*

**Что НЕ может:**
- Создавать или редактировать учебные планы (`canWrite` → academic_secretary + admin)
- Отправлять учебные планы на утверждение (только автор-секретарь делает submit; методист принимает решение)
- Управлять расписанием (создавать пары — это секретарь)
- Подписывать ЭП (#140)
- Создавать пользователей
- **Настраивать интеграцию с 1С** — только admin
- **Запускать синхронизацию с 1С**
- **Менять глобальные настройки уведомлений** (SMTP, push, Telegram-бот)
- **Управлять n8n workflows**
- **Менять глобальный brand системы**

> ⚠️ **0.102.1:** ранее методист имел доступ к `/integration` для настройки 1С — передано админу. Методист по-прежнему получает уведомления о результатах синхронизации и видит данные внешних сотрудников/студентов в своих use-flow.

---

### 🛠 Системный администратор (`system_admin`)

**Видит в меню:** ВСЁ — Dashboard, Documents, Analytics, Schedule, Calendar, Tasks, **Assignments**, **Curriculum**, Announcements, Messages, AI Assistant, Users, Integration, **Curriculum approval** (`/admin/curriculum/approve`), Admin Settings, `/admin/*`

1. Создаётся при первом деплое или через миграцию
2. **Dashboard** — полная статистика
3. **Users** — full CRUD пользователей и ролей. Единственный, кто создаёт привилегированные роли
4. **Documents** — full доступ ко всему
5. **Curriculum (полный approve workflow с 0.116.0–0.120.0):**
   - `/curriculum` — список всех учебных планов (тот же view что у методиста + фильтры по статусу)
   - `/curriculum/[id]` — детали с full edit override (`isAdmin` flag → `AuthorizeEdit` пропускает ownership-чек на draft'е любого методиста)
   - **`/admin/curriculum/approve`** — admin-only очередь учебных планов в статусе pending_approval. Single-role allowlist (non-admin redirected → /forbidden). Каждая строка показывает curriculum metadata + status pill + кнопки Approve / Reject:
     - **ApproveCurriculumDialog** — confirmation modal → status=pending_approval → approved + записывает approved_by/at + audit `curriculum.approved`
     - **RejectCurriculumDialog** — Radix form с обязательной textarea причины (≤ 4096 символов, character counter, destructive variant) → status → draft + audit `curriculum.rejected` с reason. Reason **audit-only** (не stored на entity per ADR-3) — методист видит её в audit log и UI feedback
   - **Уникальная привилегия `ActionApprove`** — единственная роль которая может утверждать или отклонять учебные планы
6. **Schedule, Reports, Analytics, Assignments** — full
7. **Integration (1С)** — **полное управление**: настройка соединения, маппинг полей, синхронизация, расписание автосинка (cron), частичный синк, откат при ошибках
8. **Settings/Automation** — управление n8n workflows, запуск тестов вручную
9. **Settings/Appearance** — **глобальная** тема и brand системы (применяется ко всем)
10. **Settings/Notifications** — **глобальные** настройки SMTP, push, Telegram-бота
11. **Admin** (`/admin`) — admin-only роуты
12. **Infrastructure** — backup, логи Loki, алерты Grafana, метрики OTEL, восстановление из backup
13. *(Личные настройки — администратор тоже их использует, как любой другой пользователь — выбирает свою тему, привязывает Telegram. Это не привилегия)*

**Уникальные привилегии (нет ни у кого больше):**
- `ActionApprove` на curriculum
- Создание privileged-ролей
- Все глобальные настройки и интеграции
- Управление n8n
- Управление backup и infrastructure

---

## Открытые задачи

### Из GitHub (open issues) — на 2026-05-19

| # | Заголовок | Приоритет |
|---|-----------|-----------|
| [#40](https://github.com/.../issues/40) | Внешние календари (Google, Outlook) | — |
| [#80](https://github.com/.../issues/80) | Анализ рынка | medium |
| [#139](https://github.com/.../issues/139) | Авто-расписание (CSP) | low |
| [#140](https://github.com/.../issues/140) | Электронная подпись | low |

**Закрытые ключевые issues:**
- ~~[#196](https://github.com/.../issues/196)~~ Backend Test Coverage strict > 90% — **closed 2026-05-19** (90.2% achieved через 12 release v0.153.0 → v0.153.11)
- ~~[#41](https://github.com/.../issues/41)~~ Workflow automation — **closed 2026-05-19** (shipped через documents/workflow_handler в 5-phase pack #227 v0.148.0 → v0.152.1)
- ~~[#227](https://github.com/.../issues/227)~~ Documents workflow 5-phase pack — closed v0.152.1
- ~~[#226](https://github.com/.../issues/226)~~ WebPush sendPushReminder — closed v0.147.0

### Из Taskmaster (актуальное состояние)

| Задача | Статус |
|--------|--------|
| Workflow automation | ✅ done (см. #41 closure) |
| Backend Test Coverage strict > 90% | ✅ done (90.2%) |
| Files Frontend | pending medium — defer post-defense |
| Web Speech API в AI-чате | pending medium |
| External calendars | pending medium (#40) |
| Auto schedule | pending low (#139) |
| Electronic signature | pending low (#140) |

---

## Краткая сводка

✅ **Готово к продакшну (16+ модулей):** auth (+ MFA TOTP end-to-end), users, documents (+ workflow 5-phase), dashboard, notifications, messaging, reporting (+ B4 annual DOCX), integration *(admin-only)*, analytics, ai, **tasks** (+ deadline reminders Telegram/email/push), **announcements**, **schedule** *(timetable + lessons + замены)*, **assignments** (academic grading + return + resubmit loop), **curriculum** (CRUD + approve workflow + bulk-edit РПД), **audit logs** (admin observability), **branding admin** (DB-backed singleton).

⚠️ **Backend без полноценного UI:** files (standalone file manager — defer post-defense)

❌ **Не реализовано:** электронная подпись (#140), авто-расписание (#139), внешние календари (#40), Web Speech API в AI-чате

🔐 **Безопасность:**
- Privilege escalation при регистрации закрыта (GH #199), 4-layer defense-in-depth
- **MFA TOTP** для system_admin (enrollment + login flow verify step)
- **Audit logs** persistence + read API + UI для админа
- **Security baseline 2026-05:** SECURITY.md + 7 GitHub Security toggles + CodeQL default scan + secret-scanning + dependabot all clean (44+34+2 alerts → 0)

🛠 **Административное разделение:** все системные настройки и интеграции — только `system_admin` (с 0.102.1). 5 admin observability dashboards (audit logs / backup / sentry / integrations / composio / branding).

⚙️ **Личные настройки:** тема и подключение каналов уведомлений доступны **всем ролям** как стандартная функция профиля.

📅 **Расписание пар:** полноценный модуль schedule_lessons — CRUD пар, замены, справочники. Доступ: секретарь/admin — full, остальные — просмотр.

📋 **Документооборот (workflow):** полный lifecycle draft → review → registered → routing (sign-visa) → execution (assign-executor + mark-executed) → archived/resubmitted, все 11 transition endpoints за `RequireRole(AcademicSecretary, SystemAdmin)`, frontend dialogs + i18n × 4 для каждого перехода. **#41 закрыт 2026-05-19**.

📊 **Прогресс (на 2026-05-19):**
- **103 релиза cumulatively** (70 minor + 32 patch + 1 micro)
- **#196 Phase 6 backend coverage CLOSED** strict > 90.0% (90.2%, +5.2pp over 12 releases v0.153.0 → v0.153.11)
- **#227 Documents workflow 5-phase pack CLOSED** end-to-end
- **#41 Workflow automation CLOSED** (реализовано в documents модуле)
- **B-feature triad CLOSED** (curriculum + assignments + B4 annual report)
- **Phase 5 admin observability CLOSED** 5/5 (audit logs / backup / sentry+users / VAPID+n8n+branding / Composio)
- **B1a Section aggregate CLOSED 5/5** (bulk-edit РПД end-to-end)
- Code review compliance: недавние релизы single-pass SHIP mean ≥9/10 по TDD/DDD/CA/Security/Tests
