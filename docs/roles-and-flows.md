# Роли и пользовательские потоки

> **Версия проекта:** 0.153.12 (см. `VERSION` в корне)
> **Состояние на:** 19 мая 2026 — **Phase 6 #196 CLOSED** (backend coverage 90.2%, strict > 90.0% ✅), **5-phase Documents workflow pack #227 CLOSED** end-to-end, **#41 Workflow automation CLOSED**, **Phase 5 admin observability CLOSED** (audit logs + backup + sentry + integrations + composio + branding), **B-feature triad CLOSED** (curriculum + assignments + B4 annual report), **MFA полностью end-to-end UI**.
> **Источники:** код (`internal/modules/auth/domain/`, `frontend/src/lib/auth/`, `frontend/src/config/navigation.ts`), GitHub issues, `.taskmaster/`, `CHANGELOG.md`, история релизов в GitHub Releases.

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
