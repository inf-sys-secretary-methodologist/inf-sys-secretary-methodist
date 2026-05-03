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
