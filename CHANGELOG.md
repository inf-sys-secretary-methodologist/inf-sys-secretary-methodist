# Changelog

Все значимые изменения проекта документируются в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/),
проект следует [Semantic Versioning 2.0.0](https://semver.org/lang/ru/).

## Правила версионирования

- **MAJOR (X.y.z)** — несовместимые изменения публичного API (например, breaking changes в REST endpoints, удаление полей из ответов)
- **MINOR (x.Y.z)** — новая обратно-совместимая функциональность (новые модули, новые endpoints, новые поля)
- **PATCH (x.y.Z)** — обратно-совместимые исправления багов и патчи безопасности

При закрытии каждой задачи или применении патча — версия обновляется. Источник истины: файл `VERSION` в корне проекта; `cmd/server/main.go @version` и `frontend/package.json` синхронизируются с ним.

---

## [0.113.0] — 2026-05-06

### Added — Student-facing read endpoints (backend)

- `GET /api/assignments/my` — список своих submission'ов с опциональным фильтром `?status=pending|graded|returned`. JOIN `submissions × assignments` по `student_id = JWT.user_id`, `ORDER BY COALESCE(due_date, created_at) DESC, id DESC` (стабильный fallback вместо bottom-of-list для assignments без срока).
- `GET /api/assignments/:id/my` — детальный view конкретной работы (assignment metadata + submission state в одном round-trip).
- Оба endpoint'а в существующей `studentAssignmentsGroup` за `RequireRole("student")` + handler whitelist `studentIDFromContext` (defence-in-depth — sibling-pattern reuse от v0.112.0).
- Domain: `Submission.AuthorizeReader(actorID)` — read-side mirror к `AuthorizeResubmitter`, тот же sentinel `ErrSubmissionOwnerOnly`. Defence-in-depth: invariant defends даже когда HTTP-layer уже keys submission по `(assignmentID, JWT subject)`.
- View: `views.StudentAssignmentView` — denormalised assignment + submission columns (single SQL round-trip). Лежит в `domain/views/`, public-field DTO без поведения.
- Repo: `SubmissionRepository.ListByStudent(studentID, status?)` — JOIN с assignments, status passthrough через empty-string sentinel.
- Use cases: `ListMyAssignmentsUseCase` (narrow port `MyAssignmentsRepository`, failure-closed validation на non-positive student id), `GetMyAssignmentDetailUseCase` (load assignment → load submission → AuthorizeReader → buildStudentAssignmentView). Failure-closed DI: nil deps panic на construction.
- Handler: `MyAssignmentsHandler` с `List` + `Detail`, ports defined locally, sentinel-first error mapping (404 / 403 / 500), failure-closed DI, role exact-match.
- Тесты: 4 sqlmock case + 6 use case case (включая прямой exercise AuthorizeReader через `forceReturn` test-knob — invariant breaks test if dropped) + 8 handler case. Reviewer SHIP mean 9.0/10 после fix-cycle (gofmt + AuthorizeReader test gap + doc/SQL drift).
- ADR-1: scope = "I have a submission" (не group-scoped). JWT не несёт `group_id`, расширение = blast radius. Trade-off: студент видит assignment только после первой grade-попытки teacher'a (lazy submission creation); eager-seeding отложен.
- Out of scope: frontend pages — v0.114.0.
- Sync: 8 files version bump (VERSION × 2, main.go @version, package.json × 2, swagger × 3). `docs/roles-and-flows.md` 0.113.0 banner + новый bullet "Мои работы" в student section.

## [0.112.0] — 2026-05-04

### Added — Assignments student resubmit flow (backend-only)

Closes the academic re-grading loop opened in v0.111.0: the student can now flip a returned submission back to pending so they can supply revisions for a fresh grading cycle. State machine `pending → graded → returned → pending` is fully closed at the domain / use case / endpoint / audit layers.

**Backend-only release** — student-facing frontend pages do not exist yet (every assignments page redirects students to `/forbidden` via `if user.role === 'student' router.replace(...)`). The endpoint lands as a ready seam for a future student-dashboard initiative.

- **Domain**:
  - `Submission.Resubmit(now)` — invariant: `status='returned'` (else `ErrNotReturned` → 409); clears the return triple (`returnReason / returnedBy / returnedAt`); flips status to `pending`; advances `updatedAt`. The grade triple is left alone — `Return` already nilled it, and resurrecting a stale grade would let the prior teacher verdict bleed into the next attempt.
  - `Submission.AuthorizeResubmitter(actorID)` — invariant: `actorID > 0 AND actorID == studentID`. Returns `ErrSubmissionOwnerOnly` otherwise (handlers map to 403). Defensive `actorID > 0` rejects missing JWT context even if a student record were ever stored with id 0.
  - `var ErrNotReturned` and `var ErrSubmissionOwnerOnly` sentinels next to the existing `ErrAlreadyReturned` / `ErrInvalidReturn` / `ErrAlreadyGraded` / `ErrAssignmentScopeForbidden`. `ErrSubmissionOwnerOnly` is intentionally distinct from `ErrAssignmentScopeForbidden` — the former is the student-side rule on the Submission aggregate; the latter is the teacher-side rule on the Assignment aggregate. Conflating would mislead any future reader tracing a 403 back to its invariant.
- **Use case** `ResubmitSubmissionUseCase` — load assignment (for the teacher id used by the notifier) → load submission by `(assignmentID, studentID)` → `AuthorizeResubmitter` → capture `previous_return_reason` BEFORE clearing → `Resubmit` → `Save` → notify teacher (best-effort) → audit. Three audit events:
  - `assignment.resubmitted` — success path, includes `previous_return_reason` (forensic mirror to v0.111.0's `previous_grade` capture on the return side).
  - `assignment.resubmit_denied` — refused attempt (foreign student); records both `actor_user_id` and `student_id` so on-call sees who tried to impersonate whom.
  - `assignment.resubmit_notify_failed` — SMTP outage observability; does NOT abort the resubmit (best-effort semantics, fifth release in a row).
- **Endpoint** `POST /api/assignments/:id/resubmit` — student-only.
  - Path: `:id` = assignmentID. studentID is derived from JWT context (= actorID); no body. Eliminates the entire class of body-vs-JWT mismatch attacks.
  - Routing: new `studentAssignmentsGroup := protectedGroup.Group("/assignments")` with `RequireRole("student")` middleware. Sibling group rather than under the existing `assignmentsGroup` because that one is gated by `RequireNonStudent` — exact inverse of what resubmit needs.
  - Handler `studentIDFromContext` whitelist (only `role == "student"`) — defence in depth on top of the route middleware. Removing either alone still rejects every non-student request.
  - Failure-closed DI: `NewResubmitHandler` panics on nil usecase.
  - Error mapping: `errors.Is` sentinel-first BEFORE `MapDomainError` fallback — `ErrAssignmentNotFound` 404, `ErrSubmissionNotFound` 404, `ErrSubmissionOwnerOnly` 403, `ErrNotReturned` 409, generic 500.
- **No new migration** — schema 030 already permits `(status='pending', returnTriple=null, gradeTriple=null)`. Resubmit is a state flip plus three nulls; the bidirectional CHECK constraint passes via the `status<>returned` branch. YAGNI on `resubmit_count` / `resubmitted_at` columns without a consumer (audit log captures the count via `assignment.resubmitted` events).
- **No frontend, no i18n** — UI deferred to whenever the student dashboard initiative ships. The single inline Russian notification text ("Студент пересдал работу" + body) lives in the `assignmentsResubmitNotifier` adapter at `cmd/server/main.go`, consistent with the grade / return notifier convention.
- **Refactor `emitAudit` extracted (T0 prereq)** — `SaveGradeUseCase.logAudit` and `ReturnSubmissionUseCase.logAudit` were identical method-pairs. The v0.111.0 reviewer flagged the duplication as "extract on N=3"; the third call site (ResubmitSubmissionUseCase) is exactly that trigger. Helper sits next to the `AuditSink` port in `audit_sink.go` and uses `maps.Copy` instead of the for-range pattern the original methods carried — closes the recurring `mapsloop` lint hint at the single place that now owns the merge.

### Tests

- 4 RED→GREEN domain pairs: `Submission.Resubmit` happy path / `Resubmit` invariant matrix / `AuthorizeResubmitter` happy + denial / use case happy path / handler happy path.
- 4 honest `test: backfill` commits — labelled accurately, not as RED→GREEN: repo `Save` round-trip post-Resubmit (no impl change), use case denial path (authz wired in T7 alongside happy path), use case error matrix table-driven (5 cases), handler error matrix (8+8 cases plus nil-usecase + payload echo).
- All Go packages green; frontend tests untouched (no UI changes).

### Code review

`superpowers:code-reviewer` verdict: **SHIP**.

| Axis | Score |
|------|-------|
| TDD | 9/10 |
| DDD | 10/10 |
| Clean Architecture | 9/10 |
| Security | 10/10 |
| Tests | 9/10 |
| Migration | N/A (justified — schema 030 sufficient) |
| i18n | N/A (justified — backend-only) |

Reviewer should-fix N1 (rename `teacherIDFromContext` → `userIDFromContext`) and N2 (replace redundant role-mismatch case with case-mismatch verification) addressed in post-review polish commits before tagging. N3 (borderline 2-case table-driven) deemed cosmetic, deferred. Defense-in-depth comment added at the `AuthorizeResubmitter` call site so a future maintainer doing dead-code elimination understands the unreachable-via-HTTP branch is intentional.

### Architecture decisions worth noting

- **Sibling route group pattern**: when an existing route group's middleware is the *exact inverse* of what a new endpoint needs (`RequireNonStudent` vs `RequireRole("student")`), don't try to special-case the existing group — register a new sibling group at the same prefix. The two `protectedGroup.Group("/assignments")` registrations live in the same file paragraphs apart and document the inverse-middleware reasoning inline.
- **Defense-in-depth on student endpoint**: route middleware (`RequireRole("student")`) + handler whitelist (`studentIDFromContext`) + entity invariant (`AuthorizeResubmitter`). Each layer can stand alone; together they ensure a misconfigured route, a removed middleware, or a future caller bypassing the handler still cannot resubmit another student's work.
- **`AuthorizeResubmitter` is unreachable through HTTP** at HEAD because the handler enforces `studentID == actorID`. The check is preserved as defence in depth for future callers (CLI, agent runners, alternate routes) and documented at the use-case call site so it is not deleted as dead code.
- **`emitAudit` N=3 timing**: the helper was deliberately NOT extracted at N=2 (after grade + return). The cost of a premature port — wrong shape, wrong abstraction — outweighs the duplication of two methods. By N=3 the right shape is obvious from the data, and the extraction is mechanical.

---

## [0.111.0] — 2026-05-04

### Added — Assignments returned-transition flow

Closes academic re-grading loop, оставленный недоделанным после v0.110.0. Teacher / methodist / academic_secretary / system_admin теперь может вернуть submission на доработку с reason — статус flip'ается `pending|graded → returned`, prior grade clears на entity и в DB.

- **Domain**:
  - `Submission.Return(reason, returnedBy, now)` — invariants: status≠returned (else `ErrAlreadyReturned`); reason trimmed-non-empty; reason ≤4096 chars; `returnedBy > 0` (`ErrInvalidReturn` wraps все три). Clears prior grade на entity (`gradeValue / feedback / gradedBy / gradedAt = nil`).
  - `Assignment.AuthorizeGrader(actorID)` reused для return permission **без нового метода**. Семантика «может ли user mutate submissions on this assignment» symmetric для grade и return. YAGNI на split — если когда-нибудь permissions diverge, тогда и расщеплять.
- **Use case** `ReturnSubmissionUseCase` — load → `AuthorizeGrader` → load-or-create Submission → capture prior grade (для audit) → `Return` → persist → notify (best-effort) → audit `assignment.returned` (с `previous_grade` / `previous_feedback` если был prior grade). Denied attempts emit `assignment.return_denied`. Notifier failure logs `assignment.return_notify_failed` без abort'а transition'а.
- **Narrow port `AuditSink`** (DIP) — `usecases.AuditSink` interface извлечён в `audit_sink.go`. SaveGrade + ReturnSubmission теперь зависят от interface, не от concrete `*logging.AuditLogger`. Production wiring не трогался — `*logging.AuditLogger` структурно satisfies interface (`map[string]any` ≡ `map[string]interface{}` в Go ≥1.18). Tests fakeable через `recordingAuditSink` с defensive map copy.
- **Endpoint** `POST /api/assignments/:id/returns` за `RequireNonStudent` + handler role whitelist (defense-in-depth). Failure-closed DI.
- **Migration 030** — columns `return_reason TEXT` / `returned_by BIGINT REFERENCES users(id) ON DELETE SET NULL` / `returned_at TIMESTAMPTZ` + bidirectional CHECK constraint `chk_submissions_returned_consistency`:
  - При `status='returned'`: returnTriple non-null AND gradeTriple null.
  - При `status<>'returned'`: anything goes.
  - Defense-in-depth: catches direct SQL bypass + Reconstitute path. Symmetric с `chk_submissions_graded_consistency` (existing).
  - Length cap `chk_submissions_return_reason_length` (4096) mirrors entity invariant.
- **Frontend**:
  - `ReturnDialog` component — reason textarea, confirm button, validation на frontend (mirrors entity invariants).
  - `SubmissionRow` integration — кнопка «Вернуть на доработку» открывает dialog.
  - Returned submissions render metadata block (reason + `returned_at` через `parseLocalDate`) в `bg-sky-50 dark:bg-sky-950/30` (cohesive со `STATUS_STYLES.returned` — `RotateCcw` icon, `text-sky-700`).
  - `GradeForm` `isAlreadyGraded` predicate widened: `status === 'graded'` → `status !== 'pending'`. Returned submission показывает empty disabled форму до student-side resubmit'а.
- **i18n × 4** (ru/en/fr/ar) — 16 новых ключей (dialog labels, error messages, status badge, action button). Parity verified.

### Tests

- 13 RED→GREEN пар (domain entity + invariant validation, migration round-trip via repo Save, repo List, use case happy path, audit content + AuditSink port extraction, handler happy path, frontend API, ReturnDialog, SubmissionRow integration, return metadata render, GradeForm gating).
- 5 backfill `test:` commits (invariant validation, clear-prior-grade assertion, table-driven authz, notifier-failure semantics, handler whitelist matrix).
- Frontend: 170 suites / 2461 tests green.
- Backend: все packages green.

### Code review

`superpowers:code-reviewer` T20: **APPROVED** mean **9.5/10**, must-fix none. DDD / Security / i18n / Migration все 10/10. Single review pass — fix-cycle не понадобился.

### Out of scope

- **Student-side resubmit flow** (`returned → pending`) — отдельный bounded surface (student-facing), отложен к v0.112.0.
- **`logAudit` helper duplication N=2** между SaveGrade и ReturnSubmission — reviewer flagged как «extract on N=3»; deferred к v0.112.0 когда `ResubmitSubmissionUseCase` станет третьим callsite.

### Architecture

- **Bidirectional CHECK constraint** pattern — для всех future state-transition aggregates: domain enforces invariant on writes, DB CHECK ловит non-domain paths.
- **Best-effort notifier** semantics закрепился (4-й релиз подряд после v0.108.0 / v0.108.1 / v0.109.0). State (return) — system of record, notification — побочный эффект; failure НЕ откатывает.
- **Subagent-driven dispatch + structured plan = natural TDD discipline.** 25 tasks × ~2 subagents (impl + review) = ~50 dispatches без потери контекста. Vs v0.109.0 где TDD borderline в Cycle 5 — здесь zero violations.

---

## [0.110.0] — 2026-05-04

### Added — Assignments read-side + grading UI

Парная фича к v0.109.0. Backend write-side (SaveGrade) написан и покрыт тестами в v0.109.0; v0.110.0 добавляет read-side endpoints, frontend pages и закрывает 2 should-fix из reviewer'а v0.109.0.

- **3 GET endpoints**:
  - `GET /api/assignments` — список assignments с pagination.
  - `GET /api/assignments/:id` — single assignment.
  - `GET /api/assignments/:id/submissions` — per-student submission rows.
- Все три endpoints за `RequireNonStudent` middleware **+ handler-level role whitelist** (defense-in-depth поверх middleware).
- **Domain reorg**:
  - `Assignment.AuthorizeAccess(userID, role, unrestricted)` — централизация read-side authz в aggregate. Teacher с `unrestricted=false` видит только свои assignments; methodist / academic_secretary / system_admin (`unrestricted=true`) — все.
  - `Assignment.NewSubmissionScore(value)` — cross-aggregate validation `Score↔MaxScore` перенесён из use case в domain (DDD compliance: invariant в правильном слое).
  - `Score.max` dropped — dead data after `NewSubmissionScore` thread-through.
- **Frontend**:
  - `/assignments` — list page с filter и pagination.
  - `/assignments/[id]` — detail + grading list view.
  - `GradeForm` component — score input с live `validateGrade` (frontend-side invariant mirror).
  - `parseLocalDate` utility — TIMESTAMPTZ из API парсится как local midnight, не `new Date(iso)` (TZ-shift bug в negative-UTC zones).
  - SWR hooks `useAssignments` / `useAssignment` / `useSubmissions`.
  - Per-resource Assignment / Submission TypeScript types.
  - Navigation entry «Задания» (hidden от students через config).
- **i18n × 4** (ru/en/fr/ar) — assignments page strings, GradeForm labels, статусы submission'ов. Parity verified через `python3 -c "import json; json.load(...)"`.

### Architecture — pagination pattern

Two-query pagination фиксирован: separate `COUNT(*) FROM ... WHERE ...` + `SELECT ... FROM ... WHERE ... LIMIT $X OFFSET $Y` с **тем же** WHERE-предикатом. Window function `COUNT(*) OVER ()` отвергнут — он шёл бы по всему dataset'у в каждой row, и при teacher scope filter (v0.108.3) дал бы wrong totals. Pattern закрепился для всех list endpoints.

### Versioning correction

Pre-versioned как `v0.109.1` (patch) и скорректирован в начале сессии до `v0.110.0` (minor). Reasoning: 3 новых endpoints **+ frontend pages = новая обратно-совместимая функциональность**, SemVer minor по правилам проекта.

### Tests

- Backend: 4 use case tests (List/Get/ListSubmissions) + sqlmock pg tests + 7 handler tests (whitelist role matrix, 403 mapping, pagination) + `Assignment.AuthorizeAccess` table-driven.
- Frontend: SWR hook tests + GradeForm validation tests + parseLocalDate tests + assignments-page integration tests.
- 2 backfill commits для test-quality gaps (covered missed branches), 1 backfill для frontend component coverage.

### Code review

Reviewer ≥9 каждая ось. Final SHIP.

---

## [0.109.0] — 2026-05-04

### Added — Assignments bounded context (academic Tasks Context)

AUDIT_REPORT.md row #8 («SaveGrade в tasks») первоначально звучал как «допилить grading в существующем `internal/modules/tasks/`». При вчитывании в код — категориальная ошибка: `tasks` это **project-management module** (`AssigneeID` / `Watchers` / `Checklists` / `Progress 0..100` / status workflow `new → assigned → in_progress → review → completed → cancelled`), не academic homework. Прилеплять `MaxScore` / `SubjectID` / grade-on-task размыло бы агрегат.

Создан **параллельный bounded context** `internal/modules/assignments/` за миграцией 029 (tables `assignments` + `submissions`). Существующий `tasks` модуль не тронут. AUDIT_REPORT row #8 переформулирован (`✅ via assignments, не tasks`).

- **Domain entities**:
  - `Score` value object — invariants (`max > 0`, `value ≥ 0`, `value ≤ max`).
  - `Submission` entity — `pending → graded` transition, `ErrAlreadyGraded` sentinel (re-grading требует явного `returned` перехода — landed в v0.111.0).
  - `Assignment` aggregate — trimmed canonical title/description, `ErrAssignmentScopeForbidden` (только автор может grade).
- **Use case** `SaveGradeUseCase` — load → `AuthorizeGrader` → Score validation → lazy-create or load Submission → Grade transition → persist (upsert) → notify (best-effort) → audit. Best-effort notification semantics: SMTP failure НЕ откатывает grade (audit `assignment.grade_notify_failed` отдельным event'ом).
- **Endpoint** `POST /api/assignments/:id/grades` за `RequireNonStudent` middleware. Failure-closed DI (panic on nil usecase). 403 mapping через `errors.Is` ДО `MapDomainError`.
- **Adapter pattern в DI seam** — `assignmentsGradeNotifier` объявлен в `cmd/server/main.go` (не в `assignments/infrastructure/`), реализует narrow port `usecases.SaveGradeNotifier`, делегирует на `notificationUseCase.Create`. Domain/usecase пакеты assignments — без cross-module Go imports.
- **Migration 029** — tables `assignments` (id, title, description, due_date, max_score, subject_id, teacher_id) и `submissions` (id, assignment_id, student_id, grade_value, feedback, graded_by, graded_at, status) с FK на `users(id)` и CHECK constraints (status enum, status='graded' ⇒ grade triple non-null + feedback length 4096).
- **Backend-only релиз** — read-side endpoints + frontend UI идут в v0.110.0.

### Tests

- 6 TDD циклов RED→GREEN: Score VO / Submission entity / Assignment aggregate / SaveGrade usecase / HTTP handler / trim-fix post-review.
- 1 честно label'ed `test:backfill` для pg-repos (infra-thin, ON CONFLICT upsert).
- Authz table-driven (3+ кейса), error matrix table-driven.

### Code review

Первоначальный verdict — **BLOCK** (DDD=7, Test=8, Security=8, SQL=8, Cohesion=7). 5 must-fix исправлены в том же релизе:

1. Trim-bug в `NewAssignment` (validated trimmed, stored raw) → TDD-cycle 6.
2. `submissions.student_id` без FK → добавлен в `029.up.sql` + `feedback` length CHECK 4096.
3. ON CONFLICT path в pg-repo не покрыт sqlmock → добавлен.
4. «existing pending → graded» path в usecase uncovered → покрыт.
5. Stale swagger → `SaveGradeRequest` exported, `@tag.name assignments` declared, `swag init` re-run.

Plus 2 should-fix включены: permission-denial audit (`assignment.grade_denied`) для security-relevant denied attempts, ранее silent.

Final ≥9 каждая ось.

### TDD honesty

Cycle 5 (handler) — *чуть не нарушил гейт*. Под давлением контекста начал writing full impl сразу. Поймал себя, откатил handler в stub, переписал tests-first → RED → GREEN. Lesson: **под давлением контекста гейт соблюдать. Senior contraction = откатить и сделать правильно даже если уже написал.**

### Known limitations

- `swag` не генерирует `/assignments/*` paths в `swagger.json` несмотря на `@Tags assignments` annotations и export'нутый request type. Tasks endpoints **тоже отсутствуют** в spec (64 paths total, без `/tasks/`, без `/assignments/`). Не регрессия v0.109.0 — existing legacy. Полное покрытие `swag` — отдельная инициатива.
- `Score.max` сейчас dead data (либо drop, либо thread-through) → выполнено в v0.110.0.
- `Assignment.NewSubmissionScore(value)` — переместить cross-aggregate validation Score↔MaxScore из usecase в domain → выполнено в v0.110.0.

---

## [0.108.3] — 2026-05-04

### Security — Teacher analytics scope filter «свои группы» (correct pagination)

До этого релиза учитель, открыв `/api/analytics/*`, видел статистику и risk score **по всем группам** учебного заведения — security gap из аудита: teacher должен видеть только группы, которым он реально преподаёт. Кроме того, наивный фильтр в Go-коде после выборки сломал бы pagination (`COUNT(*)` шёл бы по всему dataset'у, а возвращалось бы только подмножество).

После релиза:

- **Domain** — `TeacherScope` value object в `analytics/domain` (private `map[string]struct{}` whitelist групп, фабрика `NewTeacherScope`, методы `AllowsGroup` / `AllowsGroupPtr` / `FilterGroupNames` / `AllowedGroupNames`). Sentinel `ErrAnalyticsScopeForbidden` (`var ErrXxx = errors.New(...)`).
- **Specific endpoints** (`GetGroupSummary` / `GetStudentRisk` / `GetStudentRiskHistory`) проверяют scope перед I/O — non-allowed group возвращает `ErrAnalyticsScopeForbidden`, маппится в HTTP 403 через `errors.Is` ДО `MapDomainError` (иначе вылезло бы 500).
- **List endpoints** — фильтр **на уровне SQL** (`WHERE group_name = ANY($N)`), один и тот же предикат разделяет `COUNT` и data query → корректный pagination. Empty whitelist превращается в `'{}'::text[]` → 0 rows (deny-all).
- **Repository** — `TeacherScopeRepository` port + pg implementation (`SELECT DISTINCT sg.name FROM schedule_lessons sl JOIN student_groups sg ON sg.id = sl.group_id WHERE sl.teacher_id = $1`). Cross-module read **на уровне SQL** — Go импорт schedule из analytics запрещён архитектурой.
- **Handler** — `buildScope`: для `role=teacher` собирает non-nil scope из repo, для остальных ролей nil (=full access). Missing role/user_id из gin.Context → explicit 500 (defense-in-depth: handler-level invariant, не полагаться на upstream middleware).
- **Failure-closed DI** — `NewAnalyticsHandler` panics при nil scopeRepo с non-nil usecase. Production wiring должен fail loudly, если кто-то забыл подключить scope repo.

### Tests

- `TeacherScope` VO — table-driven: 4 кейса (allows-listed / denies-non-listed / handles-empty / handles-nil-scope).
- 9 usecase tests на scope checks (включая denied-attempt path с `AssertNotCalled` на mutation).
- 4 sqlmock-теста на repository (round-trip, deny-all path).
- 7 handler tests — scope assembly + 403 mapping + missing-role 500 + role-other-than-teacher → unrestricted.

### Code review

`superpowers:code-reviewer`: ≥9 каждая ось после fix-цикла (M2 missing-role guard, M3 panic-on-nil-DI, N3 / N2 minor). Verdict: **SHIP**.

### TDD honesty

Cycle 3a (`feat(analytics): plumb TeacherScope through repository layer`, `b70cf32d`) был написан как combined feat+test commit: signature change в repo interface мандатно требовал compile fixes везде (handlers / scheduler / mood_usecase / pg test). RED-only commit с new tests **без** signature update сломал бы Go build. Засчитан как backfill (test-after), не как RED→GREEN cycle. На будущее: для signature changes использовать паттерн «новый метод alongside старого → миграция callers → удаление старого» — RED становится возможен.

---

## [0.108.2] — 2026-05-03

### Fixed — Document.Update teacher ownership check

До этого релиза `usecase.Update` принимал `userID`, но **не сверял** его с `Document.AuthorID`. Любой не-student мог редактировать чужие документы — критическая дыра из аудита.

После релиза:

- **Domain rule** `Document.CanBeEditedBy(userID, role) error` — единый источник истины:
  - `methodist` / `academic_secretary` / `system_admin` → редактируют любой;
  - `teacher` → только свои (`userID == AuthorID`);
  - `student` / неизвестная роль → deny (defense-in-depth, поверх `RequireNonStudent` middleware из v0.105.3).
- **Sentinel** `entities.ErrDocumentEditDenied` (`var ErrXxx = errors.New(...)`) — handler маппит в HTTP `403` через `errors.Is` ДО generic `MapDomainError` (иначе бы вылезло 500).
- **Order matters**: проверка стоит между `GetByID` и любой мутацией / `AddHistory` — denied call не оставляет audit-history breadcrumb.

### Added

- `entities.UserRole` enum расширен константами `RoleAcademicSecretary` (`"academic_secretary"`) и `RoleSystemAdmin` (`"system_admin"`) — соответствуют wire-значениям `auth.RoleType`. Cross-module импорт `auth/domain` запрещён архитектурой, отсюда параллельный enum (комментарий в `permission.go` фиксирует duality).
- `usecase.Update` сигнатура: `(ctx, id, input, userID, role entities.UserRole)`. Breaking change на use-case границе; единственные callers — handler этого модуля и тесты (обновлены в том же commit).

### Tests

- `Document.CanBeEditedBy` — table-driven 7 ролей + sentinel-via-`errors.Is` тест.
- `TestDocumentUseCase_Update_OwnershipEnforcement` — 4 кейса: methodist/teacher own/teacher other/student. Deny path pin'ится через `AssertNotCalled` на `Update` И `AddHistory`.

### Code review

`superpowers:code-reviewer`: TDD=9, DDD=10, CA=9, Security=10, Tests=9, i18n=N/A — verdict **SHIP** (каждая ось ≥9).

---

## [0.108.1] — 2026-05-03

### Fixed — n8n absence-alert connection

До этого релиза `n8nClient` инициализировался в `cmd/server/main.go`, но сразу же выбрасывался в `_` — ни один production-путь не использовал webhook'и. Сегодняшний `RiskRecalcScheduler` создавал уведомления внутри системы, но не сигнализировал внешним workflow'ам (`workflows/absence-alert.json` мог реагировать только своим Schedule trigger'ом раз в час).

После релиза:

- `RiskRecalcScheduler.recalculate()` для каждого студента с `risk score ≥ 70` дополнительно вызывает `n8nClient.TriggerAsync(n8n.PathRiskAlertDetected, ...)`. Внешние workflow'ы (Telegram broadcast, curator email digests) реагируют в течение секунд вместо часов.
- Async, fire-and-forget — batch loop scheduler'а не блокируется на webhook latency. `TriggerAsync` поглощает транспортные ошибки в Warn log; flaky n8n не останавливает daily recalc.

### Added

- `n8n/event_handler.go` константы путей и event-типов для единого источника истины:
  - `PathDocumentCreated` / `PathDocumentUpdated` / `PathRiskAlertDetected`
  - `EventTypeDocumentCreated` / `EventTypeDocumentUpdated` / `EventTypeRiskAlertDetected`
- Запись `risk_alert.detected` → `risk-alert-detected` в `pathMap` — любое будущее `RiskAlertDetected` доменное событие, опубликованное на EventBus, маршрутизируется в существующий absence-alert workflow без дополнительных кодовых изменений.
- Audit log событие `risk_alert_dispatched_to_n8n` пишется до каждого fan-out — есть аудит-след вне зависимости от исхода webhook'а.

### Tests

- `event_handler_test.go` — два теста: table-driven по всем трём известным event-типам (через реальный `httptest.Server` + реальный `Client`, а не моки), плюс кейс на молчаливый drop неизвестных событий. Pin'ит и `occurred_at` в RFC3339 формате — drift в форматтере моментально ловится.

### PII / data classification note

Payload содержит `student_name` + `risk_score` уровня риска уходящие на operator-controlled n8n. Curator notifications уже несут те же поля — новых data-classification дыр нет, но любое будущее добавление поля (email, phone, grades) требует review против `docs/roles-and-flows.md`.

### Code review

`superpowers:code-reviewer`: TDD=9, DDD=9, CA=8, Security=7, Tests=8, i18n=N/A — verdict **SHIP**. После фиксов post-review (event_type constants, audit log, PII note, occurred_at assertion) подняты CA, Security, Tests до целевых ≥9.

### Conflict resolution (housekeeping)

Также в этом релизе очищены остатки незавершённого `git stash pop` в 5 файлах (`.env.example`, `compose.yml`, `frontend/src/app/page.tsx`, `internal/shared/infrastructure/config/config.go`, `internal/modules/auth/interfaces/http/handlers/auth_handler_test.go`) — каждый конфликт разрешён в пользу текущего main HEAD. `stash@{0}` ('WIP: environment config updates and login button') оставлен в стеке для последующего ручного review/drop.

---

## [0.108.0] — 2026-05-03

### Added — Password recovery flow (request → verify → confirm)

- `POST /api/auth/password-reset/request` — принимает `{email}`, всегда возвращает `204` независимо от того, существует ли email (anti-enumeration honored end-to-end на уровне usecase И handler).
- `GET /api/auth/password-reset/verify/:token` — read-only проверка токена; `204` если действителен, `410 Gone` если истёк/неизвестен. Frontend вызывает перед рендером формы, чтобы не показывать поле пароля для мёртвой ссылки.
- `POST /api/auth/password-reset/confirm` — принимает `{token, password}`, маппит ошибки usecase в стабильные HTTP-коды: `400` для слабого пароля (можно исправить), `410 Gone` для невалидного токена (нужен новый).
- Новый interface `repositories.PasswordResetTokenRepository` (`Store` / `LookupUser` / `Delete`). Доменный sentinel `ErrPasswordResetTokenNotFound` для `errors.Is`.
- Redis-реализация `persistence.RedisPasswordResetTokenRepository` использует `SET pwreset:<token> <userID> EX <ttl>` + `GET` / `DEL`. Empty token / non-positive TTL отвергаются на границе репо.
- `usecases.PasswordResetUseCase` — 3 публичных метода: `RequestReset(email)`, `VerifyToken(token)`, `ConfirmReset(token, newPassword)`. Доменные ошибки `ErrInvalidResetToken` (collapses unknown-token + vanished-user в один shape, чтобы каллер не мог пробить таблицу users) и `ErrWeakResetPassword`.
- `entities.User.UpdatePassword(hash)` — domain method, атомарно меняет пароль и `UpdatedAt`. Усекает риск, что usecase забудет обновить timestamp при ротации пароля.
- `entities.ReconstituteUser(...)` — DDD-фабрика для восстановления загруженного из БД user'а (id + status + timestamps). Тесты v0.108.0 используют её вместо raw `&entities.User{...}`.
- Frontend: страницы `/forgot-password` и `/reset-password` (Next 15 App Router, под `(auth)` route group). Анти-енумерация на UI: успех показывается одинаковым текстом независимо от ответа API. Form для нового пароля переходит обратно в "ссылка истекла" если confirm вернул 410 (другая вкладка спалила токен).
- API-методы `authApi.requestPasswordReset`, `verifyPasswordResetToken`, `confirmPasswordReset` в `frontend/src/lib/api/auth.ts`.

### Security

- Токен — 256 бит энтропии (`crypto/rand` + `base64.RawURLEncoding`, ~43 url-safe chars). TTL = 1 час.
- Single-use enforcement: `ConfirmReset` удаляет токен после успешного сохранения; failure to delete fatal'но (re-usable token defeats the bound).
- Order matters: weak-password проверяется ДО любого I/O. Утечка токена не даёт спалить аккаунт на 1-символьный пароль.
- Backend min-length floor (8) намеренно слабее frontend composition policy (upper/lower/digit/special) — backend не может доверять frontend; frontend накладывает UX-проверки сверху.
- Bcrypt cost 14 (как в Register).
- Anti-enumeration shape pin'ится на трёх уровнях: usecase test, handler test, frontend manual UX. Unknown email и blocked user возвращают одинаковую успешную "204"-shape без I/O. Storage fault для существующих пользователей всё ещё может теоретически отличаться от 204 через 500 — известный trade-off, задокументирован в коде.
- Endpoints за публичным rate-limiter'ом (`publicRateLimiter` через `authGroup`) — атакующий не может спамом запроса спалить таблицу через side-effects.

### Frontend

- i18n × 4: `forgotPasswordPage` (9 ключей) и `resetPasswordPage` (16 ключей) добавлены в `ru/en/fr/ar.json` с реальными переводами (не fallback на английский). Parity verified через скрипт.
- `authPages.forgotPasswordMeta` / `resetPasswordMeta` для страничных метаданных.
- Schemas `createPasswordRecoverySchema` / `createPasswordResetSchema` уже существовали — недоставала только UI-обвязка.

### Tests

- Backend: `PasswordResetUseCase` — 2 table-driven теста (3 + 4 кейса) + 2 теста VerifyToken. Каждый кейс pin'ит конкретное поведение (anti-enumeration shape, single-use, validation order, opaque error для vanished-user).
- Backend: `RedisPasswordResetTokenRepository` — 5 кейсов с `miniredis` (round-trip, TTL expiry, single-use Delete, defensive input rejection table-driven).
- Backend: `PasswordResetHandler` — 3 table-driven (4 + 2 + 5 кейсов) + 1 end-to-end bcrypt-verify через реальный usecase + fake repos.
- Frontend: `ResetPasswordForm.test.tsx` — 4 кейса (token-missing, verify-resolves, verify-rejects, confirm-410-flip back to expired).

### Code review

Прогон через `superpowers:code-reviewer` агента: TDD=9, DDD=9, CA=9, Security=9, Tests=9, i18n=10. Verdict: **SHIP** (каждая ось ≥9). Первоначальные 8/8 на DDD и Tests подняты после рефакторинга `&entities.User{}` → `ReconstituteUser` и консолидации тестов в table-driven.

---

## [0.107.0] — 2026-05-03

### Added — Logout endpoint + Redis-based token blacklist

- `POST /api/auth/logout` принимает Bearer access token, добавляет его JTI в Redis blacklist с TTL = `exp − now`. Endpoint защищён обычным `JWTMiddleware` (revocation check намеренно не применяется к самому logout, чтобы он оставался идемпотентным).
- Новый interface `repositories.RevokedTokenRepository` (`Revoke(jti, ttl)` / `IsRevoked(jti)`).
- Redis-имплементация `persistence.RedisRevokedTokenRepository` использует `SET jwt:revoked:<jti> 1 EX <ttl>` + `EXISTS`.
- Новый middleware `JWTMiddlewareWithRevocation(authUseCase, revokedRepo)` — после стандартной валидации сверяется со списком отозванных. Применён к `protectedGroup` и `aiAPIGroup`. При ошибке хранилища fail-closes 401 (лучше принудительный re-login, чем риск пропустить отозванный токен).
- Если Redis недоступен (`redisCache == nil`), revocation отключается gracefully — middleware ведёт себя как обычный `JWTMiddleware`.

### Tests

- `LogoutUseCase` — 4 теста (happy path, invalid signature, expired exp, missing JTI).
- `JWTMiddlewareWithRevocation` — 3 теста (revoked → 401, active → 200, nil repo → bypass).
- `LogoutHandler` — 3 теста (204 / 401 missing auth / 401 invalid token).

### Known limitations

- Refresh tokens не отзываются на сервере: клиент должен их выкинуть. Полная серверная инвалидация refresh — отдельная задача (`SessionRepository` уже существует, но не подключён к `AuthUseCase`).
- `/api/integration/*` остаётся под обычным `JWTMiddleware`: admin-guard и так блокирует не-админов; admin'у достаточно подождать TTL access token (15 мин).

## [0.106.0] — 2026-05-03

### Added — `ResourceDocuments` resource type in PermissionMatrix

- Новая константа `domain.ResourceDocuments` (`"documents"`) рядом с уже существующими `ResourceUsers`, `ResourceCurriculum`, `ResourceSchedule`, `ResourceAssignments`, `ResourceReports`.
- Записи в `PermissionMatrix` для всех 5 ролей с явными access levels:
  - `system_admin`, `methodist`, `academic_secretary` — Full CRUD;
  - `teacher` — Full create + Limited read (ACL) + Own update + Own delete;
  - `student` — Denied create/update/delete + Limited read (ACL).
- **До фикса:** documents был центральной сущностью, но в матрице отсутствовал — код опирался на ad-hoc проверки в `sharing_usecase`. Аудит зафиксировал это как критическую неполноту авторизации (AUDIT_REPORT, academic_secretary section).

### Tests

- `permission_documents_test.go` — `TestPermissionMatrix_DocumentsResourceForAllRoles` (5 ролей × 4 действия) и `TestPermissionMatrix_DocumentsAccessLevels` (20 sub-tests, фиксирует точные access levels — защита от регрессии).

## [0.105.3] — 2026-05-03

### Security — block student from documents.create / reports / analytics

- Новая helper-функция `authMiddleware.RequireNonStudent()` — whitelist четырёх не-студенческих ролей через существующий `RequireRole`.
- Применена в `cmd/server/main.go` на трёх точках:
  - `POST /api/documents` (read paths остаются открытыми — студенты видят shared через ACL)
  - `/api/reports/*` — вся группа закрыта от студентов
  - `/api/analytics/*` — вся группа закрыта от студентов
- **До фикса:** student имел AccessDenied в Permission Matrix, но handler-уровень не проверял — студенты могли вызывать endpoints напрямую (AUDIT_REPORT item #1).

### Tests

- 6 sub-tests на `RequireNonStudent`: blocks student / allows 4 other roles / blocks missing role.

## [0.105.2] — 2026-05-03

### Security — admin guard on /api/integration/*

- `Module.RegisterRoutes(router, requireAdmin)` теперь принимает `gin.HandlerFunc` admin-middleware и применяет его ко всей группе `/integration` (sync, employees, students, conflicts).
- В `cmd/server/main.go` передаётся `authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin))` — только роль `system_admin` может запускать 1С sync, читать sync logs, просматривать external маппинги и резолвить конфликты.
- **До фикса:** любой авторизованный пользователь (включая student) мог триггерить sync — критическая дыра из AUDIT_REPORT item #3.

### Fixed — admin role string mismatch on /api/admin/*

- `cmd/server/main.go:2023` использовал `RequireRole("admin")`, но БД хранит роль как `"system_admin"` (см. migration 001 CHECK constraint). Старый guard молча не пропускал никого. Заменено на `string(authDomain.RoleSystemAdmin)` — `/api/admin/*` стал реально доступен админам.

## [0.105.0] — 2026-04-28

### Added — Schedule Lessons module (GitHub [#201](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/201))

**Backend:**
- Domain: Lesson, Classroom, ScheduleChange, StudentGroup, Discipline, Semester, LessonType entities
- Domain types: DayOfWeek, WeekType, ChangeType enums with validators
- Repository interfaces + PostgreSQL implementations (lessons, classrooms, references, changes)
- LessonUseCase with CRUD, timetable query, schedule changes, reference data
- LessonHandler with 17 HTTP endpoints + permission guards (admin/secretary only for writes)
- Routes: `/api/schedule/lessons/*`, `/api/schedule/changes`, `/api/classrooms`, `/api/student-groups`, `/api/disciplines`, `/api/semesters`, `/api/lesson-types`

**Frontend:**
- TimetableGrid component (6 columns Mon-Sat × 5 time slots)
- LessonCard component (colored by lesson type, discipline/teacher/classroom info)
- ScheduleFilters (semester, group, classroom selects)
- `/schedule` page with week-type tabs, role-based edit controls, loading/empty states
- SWR hooks for all schedule endpoints
- i18n ×4 (ru/en/fr/ar) — 46 new keys per locale

**Security:**
- Permission guards: only `system_admin` and `academic_secretary` can create/update/delete lessons
- 22 handler tests covering permission matrix (401/403 paths)

## [0.104.0] — 2026-04-26

### Added — Resource-based permission matrix (GitHub [#206](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/206))

- Resource-based permission system: `can(role, resource, action)` replacing primitive `canEdit()`
- 8 resources × 5 actions × 5 roles = full PermissionMatrix matching `docs/roles-and-flows.md`
- `Resource`, `Action`, `AccessLevel` enums for type-safe permission checks
- `getAccessLevel(role, resource)` for granular access level inspection
- `APPROVE` action special-cased to `system_admin` + `curriculum` only
- Users page migrated from manual `isAdmin` check to `can(role, USERS, CREATE)`
- Legacy functions (`canEdit`, `isAdmin`, `isViewOnly`) marked `@deprecated`, kept for backward compat
- 92 new permission tests (73 unit + 19 integration)

## [0.103.0] — 2026-04-26

### Added — Admin Settings + Schedule placeholder + permission matrix fix (GitHub [#205](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/205), [#208](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/208), [#209](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/209))

**Admin Settings (GH #205):**
- 3 новые страницы `/admin/settings/{appearance,notifications,automation}` с tab-навигацией
- Appearance: загрузка логотипа, favicon, выбор основного цвета
- Notifications: настройки SMTP, Push (VAPID), Telegram
- Automation: статусы n8n workflows (документы, посещаемость, дедлайны)
- Navigation entry `adminSettings` (system_admin only), иконка Shield
- 4 локализации (ru/en/fr/ar) parity verified

**Schedule placeholder (GH #208):**
- Страница `/schedule` с заглушкой «coming soon»
- Route-config: доступ всем 5 ролям
- Navigation entry с иконкой GraduationCap
- 4 локализации (ru/en/fr/ar) parity verified

**Permission matrix alignment (GH #205):**
- `/integration` — только system_admin (убран methodist)
- `/tasks` — все 5 ролей (было 3)
- `/reports` — добавлен teacher
- `/schedule` — все 5 ролей (новый маршрут)
- Navigation config синхронизирован с route-config

**Documents uploader fix (GH #209):**
- DocumentUpload: заменён `Button onClick → .click()` на нативный `<label htmlFor>` (Safari fix, аналогично FileUploader из 0.102.1)

## [0.102.2] — 2026-04-26

### Changed — Docs: roles-and-flows update + personal vs global settings clarification

- Обновлён `docs/roles-and-flows.md`: добавлен раздел «Личные vs глобальные настройки», матрица доступа (PermissionMatrix), уточнены сценарии по ролям, актуализированы открытые задачи

## [0.102.1] — 2026-04-26

### Fixed — Files Frontend bug fix (GitHub [#203](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/203))

- **FileUploader**: заменён программный `inputRef.click()` на нативный `<label htmlFor>` — Safari блокировал программный клик на скрытом file input, диалог выбора файлов не открывался при нажатии на дропзону

## [0.102.0] — 2026-04-25

### Added — Announcements Module Frontend + Backend Attachments (GitHub [#202](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/202))

**Frontend:**
- Страница `/announcements` с табами по статусу (all/draft/published/archived) и Dialog для CRUD + AttachmentList
- Компоненты `AnnouncementCard`, `AnnouncementFilters`, `AnnouncementForm`, `AttachmentList`
- SWR hooks: `useAnnouncements`, `useAnnouncement` + CRUD + status actions (publish/unpublish/archive) + attachment mutations
- TypeScript types для всех enum (status × 3, priority × 4, target_audience × 5)
- Navigation entry (видим всем 5 ролям, EDIT-actions только не-student)
- Route-config rule для `/announcements`
- 4 локализации (ru/en/fr/ar) parity verified

**Backend (новое для attachments):**
- `AttachmentStorage` interface (consumer-side в usecase, DIP pattern из files module)
- `AnnouncementUseCase.AddAttachment/RemoveAttachment` с rollback на ошибке persist
- HTTP handlers `UploadAttachment/DeleteAttachment` (multipart parsing)
- Routes `POST /api/announcements/:id/attachments`, `DELETE /api/announcements/:id/attachments/:attachmentID`
- DI: `SetAttachmentStorage(s3Client)` в main.go
- Sentinel errors: `ErrStorageNotConfigured`, `ErrAttachmentNotFound`
- Storage keying scheme isolated в `attachmentStorageKey()` helper

### Fixed
- Timezone bug в `<input type="date">` submit (AnnouncementForm + TaskForm) — теперь parsing как local midnight, не UTC midnight
- AttachmentList использует `useId()` для защиты от коллизий htmlFor→id при multiple lists
- Magic `limit: 100` извлечён в `ANNOUNCEMENTS_PAGE_SIZE` константу

**Code review verdict:** TDD 9 / DDD 9 / Clean Architecture 9 / Code Quality 9.

**Тесты:** Frontend 148 suites / **2265 tests** + Backend 92 packages — все green.

## [0.101.0] — 2026-04-25

### Added — Tasks Module Frontend (GitHub [#200](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/200))

- Страница `/tasks` — управление задачами с фильтрами и Dialog для create/edit
- Компоненты `TaskCard`, `TaskFilters`, `TaskForm` — стиль mirrors `calendar/EventCard`
- SWR hooks `useTasks`, `useTask` + CRUD мутации (`createTask`, `updateTask`, `deleteTask`)
- TypeScript типы: `Task`, `TaskStatus`, `TaskPriority`, `CreateTaskInput`, `UpdateTaskInput`, `TaskFilterParams`
- Navigation entry для `system_admin`, `methodist`, `academic_secretary`
- 4 локализации (ru/en/fr/ar) с parity verification
- Locale-aware date formatting через `useLocale() + localeMap` pattern
- Table-driven тесты для 7 статусов и 4 приоритетов

### Fixed
- Pre-existing pre-existing assertions для TEACHER/ACADEMIC_SECRETARY ролей в `navigation.test.ts` (admin group flatten после settingsPage)

**Code review verdict:** TDD 9 / Frontend Architecture 9 / Code Quality 9.

**Тесты:** 142 frontend suites / 2214 tests + 92 backend packages — all green.

## [0.100.1] — 2026-04-25

### Security
- Закрыта уязвимость privilege escalation при self-registration ([#199](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/199)). Глубинная защита в 4 слоях:
  - Domain: новый инвариант `RoleType.IsAllowedForSelfRegistration()` и `var ErrRoleNotAllowedForSelfRegistration`
  - Usecase: `AuthUseCase.Register` отвергает privileged-роли до side effects
  - Handler: маппинг auth-domain ошибки в HTTP 403 Forbidden
  - Frontend: `RegisterForm` показывает в `<select>` только `student` и `teacher`

### Added
- Файл `VERSION` в корне проекта — единый источник истины версии
- `CHANGELOG.md` — журнал изменений по Keep a Changelog
- Пакет `internal/modules/auth/interfaces/http/messages` — централизованные UI-строки auth-handler'а
- Документация `docs/roles-and-flows.md` — карта ролей и сценариев для академического руководителя
- Документация `docs/er-diagram-chen.drawio` — ER-диаграмма 40 ключевых таблиц в нотации Чена

### Changed
- `cmd/server/main.go @version` синхронизирован с VERSION (0.3.0 → 0.100.1)
- `frontend/package.json` version синхронизирован (0.1.0 → 0.100.1)
- `docs/swagger/{docs.go,swagger.json,swagger.yaml}` обновлены до 0.100.1

## [0.100.0] — до 2026-04-25

Исторический baseline. Закрыто **68 GitHub issues**, **493 коммита** (162 `feat:`, 74 `fix:`, 16 `refactor:`). Версия восстановлена пост-фактум из истории git и трекера.

Под semver pre-1.0 каждый закрытый issue представляет MINOR bump; 68 issues + накопленные мелкие изменения = ~100 minor bumps.

Ключевые завершённые фичи (см. git log для полной истории):
- Auth, users, documents, dashboard, notifications, messaging, reporting, integration, analytics, ai modules — production-ready
- OpenTelemetry distributed tracing (Tempo + OTEL Collector)
- n8n automation (3 workflow + Go webhook client)
- Predictive analytics (risk scoring студентов)
- Grafana Loki / Tempo / Alerting (7 алертов в Telegram)
- Web Push notifications (VAPID + Service Worker)
- Backend test coverage 17.6% → ~50%
- API documentation (Swagger / OpenAPI)
- Performance testing (k6: smoke/load/stress/spike/soak)
- Security hardening (SSH-keys, headers, autoupdates)
- Backup automation (PostgreSQL + MinIO + offsite S3)
- Uptime Kuma status page
- PWA (Service Worker, offline page)
- i18n (ru/en/fr/ar с RTL для арабского)
