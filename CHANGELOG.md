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
