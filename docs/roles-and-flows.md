# Роли и пользовательские потоки

> **Версия проекта:** 0.120.2 (см. `VERSION` в корне)
> **Состояние на:** 7 мая 2026 (после релизов password recovery, n8n absence-alert, Document.Update ownership, teacher analytics scope filter, Assignments bounded context + UI, Returned transition flow, Student resubmit + read endpoints + UI, Curriculum module backend [v0.116.0+v0.117.0], Curriculum frontend list page [v0.118.0], Curriculum detail+edit+submit dialogs [v0.119.0], Curriculum admin approve queue + Approve/Reject dialogs [v0.120.0] — **defence-ready минимум закрыт 100%, curriculum end-to-end UI clickable**)
> **Источники:** код (`internal/modules/auth/domain/`, `frontend/src/lib/auth/`, `frontend/src/config/navigation.ts`), GitHub issues, `.taskmaster/`, `CHANGELOG.md`

> **Изменения в 0.102.2:** Уточнена концепция личных vs глобальных настроек. Каждый пользователь любой роли (включая студента) может настраивать своё рабочее окружение: тему оформления и подключённые лично к нему каналы уведомлений (Telegram, email, WebPush). Глобальные настройки (SMTP-сервер, push-провайдер, brand системы, n8n workflows, интеграция с 1С) — остаются только у системного администратора.

> **Изменения в 0.106.0:** ResourceDocuments добавлен в PermissionMatrix как 5×4 таблица (5 ролей × 4 действия). До этого права на документы были разбросаны по отдельным проверкам в handler'ах; теперь — единая декларативная матрица в `auth/domain/permission.go`. Закрывает несоответствие между документированной матрицей в `roles-and-flows.md` и реально работающими проверками. Тесты pin парность ролей и действий, отказ для неизвестной роли — failure-closed.

> **Изменения в 0.107.0:** Logout endpoint + Redis token blacklist + JWTMiddlewareWithRevocation. До этого JWT жил до истечения срока даже после "выхода" — токен оставался валидным. Теперь `POST /api/auth/logout` помещает jti в Redis blacklist (TTL = remaining lifetime токена), `JWTMiddlewareWithRevocation` проверяет blacklist на каждом запросе. Закрывает security gap из аудита: «Logout не существует, токен невозможно отозвать». Без миграции — состояние blacklist живёт только в Redis, при перезапуске Redis остаются только активные сессии.

> **Изменения в 0.108.x:** v0.108.0 — полноценный flow восстановления пароля (request → verify → confirm) с anti-enumeration. v0.108.1 — алерты о пропусках студентов теперь отправляются и в n8n (раньше workflow JSON лежал, но не подключён). v0.108.2 — `Document.Update` теперь явно проверяет авторство; преподаватель не может редактировать чужие документы (доменное правило `Document.CanBeEditedBy` + 403). v0.108.3 — `/api/analytics/*` теперь применяет scope-фильтр «свои группы» для роли teacher: список групп выводится из `schedule_lessons + student_groups`, фильтр пушится в SQL (`WHERE group_name = ANY($N)`), запрос чужой группы или экспорт по чужим студентам возвращают 403/empty.
>
> **Изменения в 0.109.0:** Введён новый bounded context `assignments` (academic Tasks Context) — отдельно от существующего `tasks` модуля (project management / issue tracker). Миграция 029 добавила таблицы `assignments` и `submissions`. Endpoint `POST /api/assignments/:id/grades` (RequireNonStudent) принимает оценку: usecase `SaveGrade` загружает Assignment, проверяет авторство преподавателя через `Assignment.AuthorizeGrader` (только автор может grade — 403 при попытке grade чужого задания), валидирует Score VO против `MaxScore` (422 при выходе за границы), lazy-создаёт Submission если ещё нет, фиксирует переход pending→graded (повторный grade возвращает 409 ErrAlreadyGraded), пишет audit `assignment.graded`, отправляет уведомление студенту через `NotificationUseCase.Create` (best-effort: при ошибке отправки grade не откатывается, фиксируется отдельным audit `assignment.grade_notify_failed`).

> **Изменения в 0.110.0:** Полный read-side flow для assignments + frontend grading UI. Три новых backend endpoint'а: `GET /api/assignments` (список с фильтрами `subject` / `group_name`, пагинация, scope-aware: преподаватель видит только свои, methodist / academic_secretary / system_admin — все), `GET /api/assignments/:id` (single), `GET /api/assignments/:id/submissions` (список submissions с JOIN на `users.name` для рендера). Все три за `RequireNonStudent` + защита-в-глубину на уровне handler'а: `callerScopeFromContext` whitelist'ит четыре non-student роли явно (любая неизвестная роль → 401, не "fall-through к unrestricted"). Read-side caller-scope правило вынесено в domain как `Assignment.AuthorizeAccess(unrestricted, userID)` — единый источник истины для grade/get/list submissions. Use case `Assignment.NewSubmissionScore(value)` забрал cross-aggregate validation Score↔MaxScore из SaveGrade — domain leak закрыт; `Score.max` (мёртвое поле) удалён. Frontend: `/assignments` (list page с фильтрами) и `/assignments/[id]/submissions` (детали + inline grade form per submission row), валидация баллов на клиенте через `validateGrade` (NEGATIVE / OVER_MAX / NOT_INTEGER / NOT_A_NUMBER), TZ-стабильный парсер дат `parseLocalDate` (CLAUDE.md правило #9), статус-фильтр (pending / graded / returned), error mapping 409/422/403 → локализованные toast'ы. i18n × 4 (ru/en/fr/ar). Студенты redirect'ятся на `/forbidden` на клиенте + RequireNonStudent на сервере.

> **Изменения в 0.111.0:** Returned transition flow — учитель / методист / академический секретарь / системный администратор теперь может вернуть работу студента на доработку через `POST /api/assignments/:id/returns` с body `{student_id, reason}`. Domain метод `Submission.Return(reason, returnedBy, now)` принимает переход `pending → returned` ИЛИ `graded → returned` (повторный возврат отклоняется через `ErrAlreadyReturned` → 409); инварианты в domain: причина обязательна, ≤4096 символов, returnedBy > 0 (`ErrInvalidReturn` → 422). При переходе из `graded` все grade-поля очищаются на entity и в БД (исторический grade сохраняется в audit как `previous_grade` / `previous_feedback`). Миграция 030 добавила колонки `submissions.{return_reason, returned_by, returned_at}` + двусторонний CHECK constraint (`status='returned'` ↔ return triple non-NULL AND grade triple NULL — defense-in-depth поверх domain). Use case `ReturnSubmissionUseCase` симметричен `SaveGradeUseCase`: load → AuthorizeGrader → load-or-fresh submission → Return → Save → notify (best-effort) → audit. Три audit события: `assignment.returned` (с captured previous_grade когда был), `assignment.return_denied` (попытка не-автором), `assignment.return_notify_failed` (SMTP fail не откатывает return). Извлечён narrow port `usecases.AuditSink` — `SaveGradeUseCase` и `ReturnSubmissionUseCase` теперь зависят от интерфейса (DIP), а `*logging.AuditLogger` структурно его удовлетворяет (zero blast radius на DI). Frontend: `ReturnDialog` (Radix modal с textarea для причины, character counter 0/4096, error mapping 409/422/403/default → локализованные toast'ы), кнопка `Вернуть на доработку` в `SubmissionRow` для статусов `pending|graded` (скрыта при `returned`), отображение `return_reason` + `returned_at` (sky-styled блок, дата через `parseLocalDate`). `GradeForm` теперь disabled при любом не-pending статусе (re-grade returned работы возможен только после resubmit'а студента — flow вынесен в v0.112.0). i18n × 4: 16 новых ключей (`assignments.returnDialog.*`, `assignments.returnButton`, `submissionRow.returned*`, `gradeForm.returnedReadOnly`). Out of scope для 0.111.0: student-side resubmit (`returned → pending`) — отдельный minor v0.112.0 когда появится student UI.

> **Изменения в 0.112.0:** Student-side resubmit flow (`returned → pending`) — замыкает академический re-grading loop, открытый в 0.111.0. Студент теперь может пересдать возвращённую работу через `POST /api/assignments/:id/resubmit` (тело пустое: `:id` из path + студент = JWT subject = единственно возможный owner). Domain метод `Submission.Resubmit(now)` принимает только `status='returned'` (иные статусы → `ErrNotReturned` → 409); очищает return triple (`return_reason / returned_by / returned_at`) и flip'ает status в `pending`. Grade triple НЕ ресуррект'ится — Return уже его обнулил, Resubmit оставляет нулям. Доменный invariant ownership вынесен в `Submission.AuthorizeResubmitter(actorID)` + sentinel `ErrSubmissionOwnerOnly` (отдельный от `ErrAssignmentScopeForbidden`: teacher-side rule на Assignment vs student-side rule на Submission — конфлятить нельзя). Use case `ResubmitSubmissionUseCase`: load assignment (для teacherID) → load submission по `(assignmentID, studentID)` → AuthorizeResubmitter → capture previous_return_reason → Resubmit → Save → notify teacher (best-effort) → audit. Три audit события: `assignment.resubmitted` (с captured `previous_return_reason` — forensic mirror к `previous_grade` на return стороне), `assignment.resubmit_denied` (попытка не-owner — security forensic trail), `assignment.resubmit_notify_failed` (SMTP fail — observability, не откатывает resubmit). Routing: новая sibling-группа `studentAssignmentsGroup` под `protectedGroup.Group("/assignments")` с `RequireRole("student")` middleware (assignmentsGroup за `RequireNonStudent` — exact inverse, так что endpoint ушёл в отдельную группу). Защита-в-глубину: `studentIDFromContext` whitelist'ит ТОЛЬКО роль student на handler-уровне. Failure-closed DI: `NewResubmitHandler` panics при nil usecase. Никакой новой миграции — схема 030 уже допускает `(status='pending', returnTriple=null)`. `logAudit` дублированный код выделен в package-private helper `emitAudit` (триггер «extract on N=3» от reviewer'а v0.111.0). Out of scope для 0.112.0: **frontend resubmit UI** — student dashboard в проекте ещё не существует (все assignments-страницы выгоняют студентов через `if user.role === 'student' router.replace('/forbidden')`), кнопка «Пересдать» отложена до отдельной student-dashboard initiative. Endpoint лежит готовым seam'ом для будущего UI.

> **Изменения в 0.120.0:** Curriculum admin approve queue + Approve/Reject dialogs — **defence-ready минимум закрыт 100%**. `/admin/curriculum/approve` admin-only страница (single-role allowlist; non-admin redirected к /forbidden) с pending_approval queue, status pill, Approve + Reject action buttons per row. ApproveCurriculumDialog (Radix confirmation, mirror к SubmitCurriculumDialog — no form, empty body POST). RejectCurriculumDialog (Radix form mirror к ReturnDialog — reason textarea с client validation trim non-empty + ≤4096 chars + character counter; variant='destructive' confirm button). Хелперы `approveCurriculum(id)` + `rejectCurriculum(id, body)` POST wrappers. RejectCurriculumRequest type re-introduced alongside its consumer. Navigation entry `curriculumApprove` в adminGroup с ClipboardCheck icon, single-role allowlist mirror к backend RequireRole(SystemAdmin) gate. i18n × 4 parity (+37 keys → 117 total per locale). 46 new module tests (16 hook + 7 ApproveDialog + 12 RejectDialog + 17 page + 5 nav); frontend total 184 suites / 2629 tests passing. Reviewer SHIP mean 9.33/10 single-pass + fix-cycle (added nav filter test для non-admin + spinner shell test + conditional dialog render + inline comment). Curriculum end-to-end UI: methodist creates draft → edits → submits → admin approves OR rejects (с reason) → если rejected methodist revises + resubmits. Все 7 backend endpoints (v0.116.0+v0.117.0) имеют UI consumers (v0.118.0+v0.119.0+v0.120.0). Out of scope: create dialog (v0.121.0), pagination UI / status pill badge (v0.121.0), bulk approve, notification UI.

> **Изменения в 0.119.0:** Curriculum detail page + edit dialog + Submit dialog — закрывает methodist self-edit cycle. `/curriculum/[id]` страница со status pill, status-aware action buttons (Edit + Submit visible только status='draft'; pending+approved+archived render read-only metadata + status hint в colored panel). `EditCurriculumDialog` (Radix modal с 5-field form: title/code/specialty/year/description; client validation mirrors domain invariants verbatim — trim non-empty / year ∈ [2000, 2100] / description ≤ 4096; error mapping sentinel-first 409→codeExists / 422→notEditable / 403→forbidden; dialog stays open on error; useEffect resets form state on reopen). `SubmitCurriculumDialog` (Radix confirmation modal, mirror к ResubmitDialog — state transitions consistently use dialogs across codebase). Хелперы `updateCurriculum(id, body)` + `submitCurriculum(id)` POST wrappers. Shared `status.ts` module экспортирует `STATUS_STYLES` (color palette × 4 lifecycle states + lucide icons) + `statusKey()` mapper — single source of truth, used by CurriculumCard + detail page. `UpdateCurriculumRequest` type re-introduced alongside its consumer dialog (CLAUDE.md "никаких на будущее"). i18n × 4: 80 keys/locale (+58 vs v0.118.0). 53 new module tests; frontend total 181 suites / 2583 tests passing. Reviewer SHIP mean 9.43/10 post-fix-cycle (single-pass был 8.67/10 с axis 7 cohesion 6/10 due to duplication; fix-cycle extracted shared module + wrapped Submit в dialog + добавил 2 missing tests + обновил plan ADR-5).

> **Изменения в 0.118.0:** Curriculum frontend list page — `/curriculum` страница для methodist / system_admin / academic_secretary / teacher с filters (status select со 5 опциями, year, specialty), CurriculumCard grid (1-2-3 col responsive) + status pill (color-coded slate=draft / amber=pending / emerald=approved / zinc=archived с lucide иконками), empty state, error block, page-shell guard ДО data-loading branch. Hooks `useCurricula(filter?, opts?)` + `useCurriculum(id, opts?)` поверх SWR с `FetchOpts.enabled` flag (default true) — false short-circuits SWR key к null для skip 401 round-trip когда student briefly authenticated до /forbidden redirect (SEC pattern из v0.114.0). Three-condition fetch gate: `!isLoading && isAuthenticated && user?.role !== 'student'` (стрictly stricter mirror к /my-assignments). Navigation entry `BookMarked` icon в educationGroup, role whitelist mirror к backend `RequireNonStudent` gate. i18n × 4: 22 keys per locale (21 curriculum.* + nav.curriculum) — title / description / loadFailed / countLabel / filters.{status, year, specialty, statusOptions.{all, draft, pending, approved, archived}} / empty / card.{openAria, status.{draft, pending, approved, archived}}; `pending_approval` wire format маппится к UI-shorter `pending` key. Curriculum types layer (Curriculum, CurriculumStatus enum, CurriculumListResponse, CurriculumListFilter) — без forward-looking request DTOs (CLAUDE.md "никаких на будущее"; Create/Update/Reject types лендятся alongside их UI consumers в v0.119.0/v0.120.0). 50 new tests (8 hook + 9 card + 12 page + 21 nav). Frontend total 178 suites / 2530 tests passing (+3 suites / +30 tests vs v0.115.0 baseline). Reviewer SHIP **mean 9.43/10** single-pass + fix-cycle (drop 3 future request types + add countLabel-hidden coverage). Out of scope: detail page `/curriculum/[id]` (v0.119.0), edit dialog + submit button (v0.119.0), admin approve page `/admin/curriculum/approve` (v0.120.0), pagination UI + status pill в navigation badge (v0.121.0).

> **Изменения в 0.117.0:** Curriculum approve workflow — три lifecycle transitions поверх v0.116.0 CRUD замыкают author→approve loop на backend. `SubmitForApproval` (methodist или admin → pending_approval), `Approve` (admin-only → approved + записывает approved_by/at), `Reject` (admin-only → draft, причина в audit per ADR-3, не на entity). Три новых endpoint'а: `POST /api/curriculum/:id/submit` (под curriculumGroup за RequireNonStudent, handler whitelist methodist+admin), `POST /api/curriculum/:id/approve` и `POST /api/curriculum/:id/reject` (под новым adminCurriculumGroup sibling за `RequireRole(SystemAdmin)`, mirror v0.112.0 sibling-route pattern). `Reject` body `{reason}` — handler enforces non-empty trim (400). Audit symmetry: 6 новых событий (`submitted`/`submit_denied`/`approved`/`approve_denied`/`rejected`/`reject_denied`) + transport-skip-audit invariant. `emitAudit` + `denialFields` helpers extracted (N=5 trigger), v0.116.0 callers migrated в same release. Без миграции — migration 031 уже provisioned status/approved_by/approved_at nullable. Reviewer SHIP **mean 10.0/10 every axis** single-pass (TDD 10 / DDD 10 / CA 10 / Security 10 / Tests 10 / Cohesion 10) — второй 10/10 single-pass за curriculum line. Author→Approve loop теперь полностью замкнут на backend: demo flow методист создаёт draft → submits → admin approves OR rejects → (если rejected) методист revises + re-submits. UI consume семи endpoint'ов в v0.118.0+. Out of scope: Discipline child entity (post-defence), Archive transition (v0.122.0+), permanent rejection_reason column (audit-only сейчас), frontend (v0.118.0–v0.121.0), workflow approval #41 заглушка.

> **Изменения в 0.116.0:** Curriculum module backend (basic CRUD) — закрывает последний 🔴 defence-critical gap из PermissionMatrix. Новый bounded context `internal/modules/curriculum/` с full Clean Architecture stack: `Curriculum` aggregate root (учебный план: title / code unique / specialty / year ∈ [2000,2100] / description / status / created_by / approved_by/at) с инвариантами + `CurriculumStatus` typed enum (draft/pending_approval/approved/archived) + `AuthorizeEdit` (status-gate ПЕРЕД ownership; admin override на ownership но не на status freeze) + `UpdateBasics` атомарный content-edit. Persistence — `CurriculumRepositoryPG` поверх `database/sql` с filterClause (status / year / specialty / created_by) + ON unique-violation → `ErrCurriculumCodeExists` (409). 4 use cases (Create / Get / List / Update) с failure-closed nil-repo panic, audit symmetry для каждого denial reason (`invalid` / `code_conflict` / `not_found` / `forbidden` / `not_editable`), transport errors propagate БЕЗ audit (audit log records policy decisions, not infrastructure outages). HTTP layer — 4 endpoints под `/api/curriculum` за `RequireNonStudent`: `POST` (write whitelist methodist+admin), `GET /:id`, `GET ?status=...&year=...&specialty=...&created_by=...&limit=&offset=`, `PUT /:id` (admin override через `isAdmin` flag). Strict boundary parsing — `parsePositiveID` отвергает fractional/zero/negative; `parseListInput` отвергает unknown status literal / out-of-range year / non-positive created_by ДО DB round-trip. Migration 031 — `curricula` table с 7 CHECK constraints mirror domain + `chk_curricula_approved_consistency` (defence-in-depth: status=approved implies approved_by/at populated, ловит direct SQL bypass). Frontend pages `/curriculum`, `/curriculum/:id`, `/admin/curriculum/approve` — deferred к v0.118.0–v0.120.0. Approve workflow (admin-only `ActionApprove`) + `Discipline` child entity — deferred к v0.117.0. Reviewer SHIP **mean 9.4/10** every axis (TDD 10 / DDD 10 / CA 9 / Security 9 / Tests 9 / Migration 10 / Cohesion 9), три полировки после ревью.

> **Изменения в 0.115.0:** Student Resubmit UI — **закрывает academic re-grading loop end-to-end в UI**. На detail page `/my-assignments/[id]` для status='returned' появляется кнопка «Пересдать работу» (visible ONLY когда status='returned' — pending/graded прячут кнопку, чтобы не приглашать гарантированный 409 NOT_RETURNED). Клик открывает `ResubmitDialog` (Radix modal — title + description + confirm/cancel, без textarea: backend resubmit endpoint v0.112.0 принимает empty body). Confirm → `resubmitSubmission(assignmentId)` POST хелпер → `toast.success` + `mutate()` SWR refresh → status pill flips на pending без manual reload + close dialog. Error mapping sentinel-first: 409 NOT_RETURNED → toast «Эта работа уже не в статусе Возвращено», 403 forbidden → toast «Можно пересдавать только свои работы» (defended даже когда unreachable через HTTP — handler hardwires student_id = JWT subject), generic → fallback toast. Dialog stays open on error — student может retry без re-opening (mirrors ReturnDialog UX). Hook surface: `resubmitSubmission(id)` thin POST wrapper рядом с `useMyAssignments` / `useMyAssignment`. i18n × 4: новый `myAssignments.resubmitButton` + `myAssignments.resubmitDialog.*` namespace (10 keys × 4 locales, parity verified). Удалён v0.114.0 `myAssignments.detail.resubmitHint` (button live, hint больше не нужен). Тесты: 11 новых (2 hook helper + 8 ResubmitDialog + 1 page button-visibility); frontend total 175 suites / 2500 tests green (+1 suite / +11 vs v0.114.0). После v0.115.0 demo-flow на защите: teacher grade → returned → student resubmit → re-grade — полностью прокликиваемый в UI, не curl.

> **Изменения в 0.114.0:** Student My Assignments — frontend pages поверх v0.113.0 backend endpoints. `/my-assignments` (list page с status-фильтром tabs: all / pending / graded / returned, grid из `StudentAssignmentCard` с status pills и color-coded secondary line — grade fraction для graded, return-reason snippet для returned), `/my-assignments/[id]` (detail page с assignment metadata header + status-aware panel: pending — amber «Ожидает проверки», graded — emerald с {value}/{max} + feedback block, returned — sky с return_reason + дата + hint что Resubmit button появится в v0.115.0). Обе pages с auth guard: non-student → /forbidden client-side mirror к существующему `/assignments` redirect (backend GET /api/assignments/my за `RequireRole("student")` уже отдаёт 401, client-guard skip useless round-trip). Hooks: `useMyAssignments(status?)`, `useMyAssignment(id)` — SWR conventions match существующих `useAssignments` / `useAssignment` (dedupingInterval=SHORT, revalidateOnFocus=false, null-id short-circuit для detail). New navigation entry `myAssignments` под академической группой только для `UserRole.STUDENT` — sits параллельно с existing `assignments` (teacher/admin), не replaces. i18n × 4: новый top-level `myAssignments.*` namespace + `nav.myAssignments` (ru/en/fr/ar parity verified). Тесты: 25 новых (6 hook + 6 card + 6 list page + 7 detail page); frontend total 174 suites / 2486 tests green. Out of scope для 0.114.0: Resubmit dialog + button — v0.115.0 закроет academic loop end-to-end в UI.

> **Изменения в 0.113.0:** Student-facing read endpoints — backend extension для будущего student dashboard. Два новых endpoint'а в существующей `studentAssignmentsGroup` (за `RequireRole("student")`): `GET /api/assignments/my` (список своих работ — denormalised JOIN `submissions × assignments` где `student_id = JWT.user_id`, опциональный `?status=pending|graded|returned` фильтр) и `GET /api/assignments/:id/my` (детальный view конкретной работы — assignment metadata + submission state в одном round-trip). Read scope = "submissions where I am owner": JWT не несёт group_id, расширять JWT = blast radius (re-issue всех токенов), поэтому group-scoped variant отложен до отдельной initiative с group enrichment. Trade-off: студент видит assignment только после первой grade-попытки (lazy submission creation в SaveGrade) — known UX gap, замкнёт eager-seeding initiative. Domain: `Submission.AuthorizeReader(actorID)` (defence-in-depth mirror к `AuthorizeResubmitter` — same predicate, separate verb для clarity на read use case). View `views.StudentAssignmentView` (denormalised assignment + submission columns — single SQL round-trip). Repo: `SubmissionRepository.ListByStudent(studentID, status?)` — JOIN с assignments, `ORDER BY COALESCE(a.due_date, a.created_at) DESC, a.id DESC` (assignments без due_date сортируются по created_at — стабильный fallback вместо bottom of list). Use cases: `ListMyAssignmentsUseCase` (narrow port `MyAssignmentsRepository`, failure-closed validation на non-positive student id), `GetMyAssignmentDetailUseCase` (load assignment → load submission по pair → AuthorizeReader → buildStudentAssignmentView). Handler: `MyAssignmentsHandler` с `List` + `Detail`, ports defined locally, sentinel-first error mapping (404 / 403 / 500), failure-closed DI panics. Тесты: 4 sqlmock + 5 use case + 8 handler — table-driven где ≥3 cases. Out of scope: **frontend pages** (отдельный minor v0.114.0).

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
| **Утверждение учебных планов** | `ActionApprove` — единственная роль с этой привилегией | `/admin/curriculum/approve` |
| **Backup, логи, метрики, алерты** | Эксплуатация системы | `/admin/infra/*` |

**Принцип**: всё, что является системной настройкой и влияет на работу системы для всех пользователей или на её взаимодействие с внешним миром — это исключительно admin.

### Матрица доступа (PermissionMatrix)

| Ресурс | system_admin | methodist | academic_secretary | teacher | student |
|--------|:------------:|:---------:|:------------------:|:-------:|:-------:|
| **users** (CRUD) | full | read limited | read limited | read limited | own update |
| **curriculum** (учебные планы) | full + approve | full | read | read+limited update | read limited |
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
| **auth** | 2046 | `/login`, `/register`, `/forgot-password` | ✅ |
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
| **curriculum** (учебные планы) | ~2800 | `/curriculum` (список с фильтрами), `/curriculum/[id]` (детали + edit + submit), `/admin/curriculum/approve` (admin queue с Approve/Reject) | ✅ |

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
| **schedule** (расписание пар) | ✅ events + lessons | ✅ `/schedule` timetable grid + `/calendar` events | Полноценное расписание: CRUD пар, замены, справочники |
| **files** | ✅ 1933 LOC | ❌ | Нет файлового менеджера (только через документы и вложения) |

**Закрыто в недавних релизах:**
- **~~tasks~~** — GH [#200](https://github.com/.../issues/200) в **0.101.0**
- **~~announcements~~** — GH [#202](https://github.com/.../issues/202) в **0.102.0**
- **~~admin-permissions-rebalance~~** — внутреннее изменение в **0.102.1**: интеграция 1С → admin
- **~~personal-settings-clarification~~** — **0.102.2**: личные настройки доступны всем ролям

---

## Что НЕ РАБОТАЕТ (заглушки)

| Модуль | Состояние | GitHub |
|--------|-----------|:------:|
| **workflow** (согласование) | Полностью пустая папка, нет `.go` файлов | [#41](https://github.com/.../issues/41) |
| **Электронная подпись** | Не начато — УКЭП/УНЭП, КриптоПро | [#140](https://github.com/.../issues/140) |
| **Авто-расписание** | Не начато — CSP алгоритм | [#139](https://github.com/.../issues/139) |
| **Внешние календари** | Не начато — Google Calendar, Outlook, iCal | [#40](https://github.com/.../issues/40) |
| **Web Speech API** | Не начато — голосовой ввод/вывод в AI-чате | TM #23 |

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
9. **Curriculum (read+limited update с 0.118.0+):** `/curriculum` (список с фильтрами по статусу/году/специальности) + `/curriculum/[id]` (детали с status pill). Read-only для учителя; редактирование закрыто `AuthorizeEdit` гейтом (только методист или admin).
10. **Messages** — групповые чаты со студентами
11. **AI Assistant** — расширенные права на RAG
12. *(Личные настройки — стандартно)*

**Что НЕ может:** создавать/редактировать расписание, видеть отчёты других преподавателей, **редактировать curriculum** (только read), grade'ить чужие assignments (`Assignment.AuthorizeGrader` — только автор), создавать пользователей, любые системные настройки.

---

### 📋 Академический секретарь (`academic_secretary`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group (Reports + Analytics), Schedule, Calendar, Tasks, **Assignments** (read), **Curriculum** (read), Announcements, Messages, AI Assistant, Admin group (Users — read limited), Profile

1. Создание администратором
2. **Dashboard** — административные виджеты
3. **Documents + Templates** — full CRUD, шаблоны (создание/редактирование)
4. **Schedule** — **полное управление расписанием** (создание пар, замены, аудитории)
5. **Reports** — full create/read/export
6. **Analytics** — просмотр аналитики студентов (риски, посещаемость, успеваемость)
7. **Users** — read limited
8. **Calendar** — управление событиями
9. **Assignments (read с 0.110.0):** `/assignments` (список всех заданий, не только своих — caller scope unrestricted) + `/assignments/[id]/submissions` (просмотр работ студентов). Может вернуть работу через `ReturnDialog` (`AuthorizeGrader` принимает 4 non-student роли в read-only сценарии; grading закрыт за teacher's ownership).
10. **Curriculum (read с 0.118.0+):** `/curriculum` (список с фильтрами) + `/curriculum/[id]` (детали с status pill). Read-only — `canWrite` whitelist'ит только methodist + admin.
11. **Messages, AI** — стандартно
12. *(Личные настройки — стандартно)*

**Что НЕ может:** редактировать curriculum (только read), создавать/обновлять учебные планы (только методист или admin), создавать пользователей, подписывать задания, утверждать учебные планы (admin-only), любые системные настройки.

---

### 📚 Методист (`methodist`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group, Schedule, Calendar, Tasks, **Assignments** (read), **Curriculum** (full без approve), Announcements, Messages, AI Assistant, Users (read limited), Profile

1. Создание администратором
2. **Dashboard** — методические виджеты
3. **Documents + Templates** — full CRUD, создание шаблонов документов
4. **Curriculum (полный self-edit cycle с 0.118.0–0.119.0):**
   - `/curriculum` — список всех учебных планов с фильтрами status/year/specialty + цветной status pill (черновик / на утверждении / утверждён / архив)
   - `/curriculum/[id]` — детали с status-aware панелью: для status='draft' доступны кнопки **Редактировать** + **Отправить на утверждение**; для pending/approved/archived — read-only с подсказкой почему
   - **EditCurriculumDialog** (Radix modal с 5 полями: title / code / specialty / year ∈ [2000, 2100] / description ≤ 4096) — client-side валидация зеркальная к domain invariants, error mapping 409→codeExists / 422→notEditable / 403→forbidden, dialog stays open on error для retry
   - **SubmitCurriculumDialog** — confirmation modal для перехода draft → pending_approval. После confirm учебный план уходит на утверждение администратору; редактирование блокируется до решения
   - **Утверждение запрещено** (`ActionApprove` → admin-only). Если admin отклоняет с reason — учебный план возвращается в draft, методист видит причину в audit log + UI feedback, правит и отправляет повторно
5. **Reports + Analytics** — full доступ, экспорт CSV/XLSX
6. **Schedule** — read full + limited update
7. **Assignments (read с 0.110.0):** просмотр всех заданий и работ студентов — caller scope unrestricted для методиста
8. **Users** — read limited
9. **AI Assistant** — расширенные права
10. **Calendar, Messages** — стандартно
11. *(Личные настройки — стандартно)*

**Что НЕ может:**
- Утверждать учебные планы (`ActionApprove` → только admin)
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

### Из GitHub (open issues)

| # | Заголовок | Приоритет |
|---|-----------|-----------|
| [#196](https://github.com/.../issues/196) | Backend Test Coverage до 90% | high |
| [#41](https://github.com/.../issues/41) | Workflow automation (согласование документов) | (workflow label) |
| [#40](https://github.com/.../issues/40) | Внешние календари (Google, Outlook) | — |
| [#80](https://github.com/.../issues/80) | Анализ рынка | medium |
| [#139](https://github.com/.../issues/139) | Авто-расписание (CSP) | low |
| [#140](https://github.com/.../issues/140) | Электронная подпись | low |

### Из Taskmaster (in-progress / pending)

| ID | Задача | Статус |
|----|--------|--------|
| 1 | Workflow automation | in-progress (15%) |
| 8 | Backend Test Coverage 60% | in-progress (~50%) |
| **24** | **Tasks Module Frontend** | ✅ done 2026-04-25 (релиз 0.101.0) |
| 25 | Schedule Lessons Frontend | **blocked** (нет backend для schedule_lessons) |
| **26** | **Announcements Frontend + Backend Attachments** | ✅ done 2026-04-25 (релиз 0.102.0) |
| 27 | **Files Frontend** | pending medium |
| **28** | **🔐 SECURITY: Self-registration** | ✅ done 2026-04-25 (GH #199) |
| **29** | **Admin permissions rebalance (1С → admin)** | ✅ done 2026-04-26 (релиз 0.102.1) |
| **30** | **Personal settings clarification** | ✅ done 2026-04-26 (релиз 0.102.2) |
| 23 | Web Speech API в AI-чате | pending medium |
| 2 | External calendars | pending medium |
| 5 | Auto schedule | pending low |
| 6 | Electronic signature | pending low |

---

## Краткая сводка

✅ **Готово к продакшну (13 модулей):** auth, users, documents, dashboard, notifications, messaging, reporting, integration *(admin-only)*, analytics, ai, **tasks**, **announcements**, **schedule** *(расписание пар + события)*

⚠️ **Backend без UI:** files

❌ **Не реализовано:** workflow (согласование), электронная подпись, авто-расписание, внешние календари

🔐 **Безопасность:** privilege escalation при регистрации **закрыта** (GH #199). Глубинная защита в 4 слоях.

🛠 **Административное разделение (0.102.1):** все системные настройки и интеграции — только `system_admin`.

⚙️ **Личные настройки (0.102.2):** тема и подключение каналов уведомлений доступны **всем ролям** как стандартная функция профиля. Глобальные параметры (SMTP, brand, 1С, n8n) остаются у admin.

📅 **Расписание пар (0.105.1):** полноценный модуль schedule_lessons — CRUD пар, замены, справочники (аудитории, группы, дисциплины, типы занятий, семестры). Сетка расписания `/schedule` с фильтрами и week-type табами. Доступ: секретарь/admin — полное управление, остальные — просмотр.

📊 **Прогресс:** на 2026-04-26 закрыто 73+ GH issues, ~545+ коммитов. Code review compliance: все недавние релизы (0.100.1, 0.101.0, 0.102.0, 0.102.1, 0.102.2) с оценкой ≥9/10 по TDD, DDD, CA.
