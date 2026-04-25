# Роли и пользовательские потоки

> **Версия проекта:** 0.102.0 (см. `VERSION` в корне)
> **Состояние на:** 25 апреля 2026 (после релиза Announcements Module + Backend Attachments, GitHub #202)
> **Источники:** код (`internal/modules/auth/domain/`, `frontend/src/lib/auth/`, `frontend/src/config/navigation.ts`), GitHub issues, `.taskmaster/`, `CHANGELOG.md`

## Содержание

1. [Роли в системе](#роли-в-системе)
2. [Что РАБОТАЕТ полностью](#что-работает-полностью)
3. [Что РАБОТАЕТ частично](#что-работает-частично)
4. [Что НЕ РАБОТАЕТ (заглушки)](#что-не-работает-заглушки)
5. [Сценарии по ролям](#сценарии-по-ролям)
6. [Открытые задачи](#открытые-задачи)

---

## Роли в системе

В коде определено **5 ролей** (`internal/modules/auth/domain/role.go`):

| Роль | Код | Назначение |
|------|-----|------------|
| **Системный администратор** | `system_admin` | Полное управление системой |
| **Методист** | `methodist` | Методическое обеспечение учебного процесса |
| **Академический секретарь** | `academic_secretary` | Административное сопровождение |
| **Преподаватель** | `teacher` | Реализация образовательного процесса |
| **Студент** | `student` | Участие в образовательном процессе (view-only) |

### Матрица доступа (PermissionMatrix)

| Ресурс | system_admin | methodist | academic_secretary | teacher | student |
|--------|:------------:|:---------:|:------------------:|:-------:|:-------:|
| **users** (CRUD) | full | read limited | read limited | read limited | own update |
| **curriculum** (учебные планы) | full | full | read | read+limited update | read limited |
| **schedule** (расписание) | full | read+limited | full | read | read |
| **assignments** (задания) | full | full+limited | read | full+own | own read+execute |
| **reports** (отчёты) | full | full | full | limited | denied |

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
| **integration** | 5557 | `/integration` (синк внешних сотрудников/студентов) | ✅ |
| **analytics** | 2430 | `/analytics` (риски студентов, тренды) | ✅ |
| **ai** | 5837 | `/ai` (RAG-чат с цитированием) | ✅ |

### Инфраструктура (всё работает)

- **OpenTelemetry tracing** — Tempo + OTEL Collector, корреляция логов с trace_id
- **n8n automation** — 3 workflow: уведомления документов, алерты пропусков, напоминания дедлайнов
- **Loki** — централизованное логирование через Grafana
- **Grafana Alerting** — 7 алертов в Telegram (CPU, RAM, диск, latency, errors, backup)
- **Web Push** — VAPID + Service Worker
- **Uptime Kuma** — status page (Caddy proxy не настроен)
- **Backup** — PostgreSQL + MinIO в S3 (offsite не подключён)
- **PWA** — Service Worker, offline page
- **i18n** — ru/en/fr/ar (RTL для арабского)

---

## Что РАБОТАЕТ частично

**Backend готов, но фронта нет** или есть, но неполный:

| Модуль | Backend | Frontend | Что отсутствует |
|--------|:-------:|:--------:|-----------------|
| **schedule** (расписание пар) | ⚠️ только events | ⚠️ только календарь events | Нет lesson-handlers/usecase, страницы расписания пар |
| **files** | ✅ 1933 LOC | ❌ | Нет файлового менеджера (только через документы и вложения) |

**Закрыто в недавних релизах:**
- **~~tasks~~** — GitHub [#200](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/200) в релизе **0.101.0**. Страница `/tasks` с фильтрами и Dialog для CRUD. Verdict: 9/9/9.
- **~~announcements~~** — GitHub [#202](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/issues/202) в релизе **0.102.0**. Страница `/announcements` со status-табами + Dialog с AttachmentList. Backend дополнен endpoint-ами для upload/remove attachments через S3. Verdict: 9/9/9/9.

**Покрытие тестами** (in-progress, цель 60%, сейчас ~50%):
- documents/usecases: 33%
- auth/usecases: 33.6%
- reporting/usecases: 38.6%
- notifications/usecases: 37.4%
- messaging/usecases: 59.2%

---

## Что НЕ РАБОТАЕТ (заглушки)

| Модуль | Состояние | GitHub |
|--------|-----------|:------:|
| **workflow** (согласование) | Полностью пустая папка, нет `.go` файлов. БД-схема и статусы готовы, движка нет | [#41](https://github.com/.../issues/41) |
| **Электронная подпись** | Не начато — УКЭП/УНЭП, КриптоПро | [#140](https://github.com/.../issues/140) |
| **Авто-расписание** | Не начато — CSP алгоритм | [#139](https://github.com/.../issues/139) |
| **Внешние календари** | Не начато — Google Calendar, Outlook, iCal | [#40](https://github.com/.../issues/40) |
| **Web Speech API** | Не начато — голосовой ввод/вывод в AI-чате | TM #23 |

---

## Сценарии по ролям

### 🔓 Гость (неавторизованный)

**Доступные страницы:** `/`, `/login`, `/register`, `/forgot-password`, `/reset-password`, `/forbidden`, `/offline`

**Что может:**
1. Зарегистрироваться (`/register`):
   - Поля: email, пароль, имя
   - **Выбор роли:** только `student` или `teacher` (whitelist для self-registration)
   - После регистрации — auto-login и редирект на `/dashboard`
2. Войти (`/login`) — JWT в httpOnly cookie + sessions в БД
3. Восстановить пароль (если email-сервис настроен)

**🔐 Защита от privilege escalation (фикс GitHub #199, 2026-04-25):**

Глубинная защита в 4 слоях:
- **Domain:** инвариант `RoleType.IsAllowedForSelfRegistration()` (только `student`, `teacher`)
- **Usecase:** `AuthUseCase.Register()` отвергает privileged-роли через `domain.ErrRoleNotAllowedForSelfRegistration`
- **Handler:** маппит ошибку в HTTP 403 Forbidden
- **Frontend:** `RegisterForm.tsx` показывает в `<select>` только 2 разрешённые роли

Привилегированные роли (`methodist`, `academic_secretary`, `system_admin`) создаются только администратором через user management.

---

### 👨‍🎓 Студент (`student`)

**Видит в меню:** Dashboard, Documents (просмотр), Calendar, Messages, AI Assistant, Profile, Settings

**Сценарий:**
1. **Регистрация** → выбирает роль "Студент" → авто-логин
2. **Dashboard** — виджеты: ближайшие события, последние объявления, непрочитанные сообщения, упоминания
3. **Documents** — только чтение публичных документов и тех, к которым выдан доступ. Создание/редактирование запрещено
4. **Calendar** — просмотр своих событий и расписания группы (когда будет страница расписания пар)
5. **Messages** — чаты с преподавателями и одногруппниками, WebSocket-уведомления
6. **AI Assistant** — RAG-поиск по доступным документам, чат с цитированием источников
7. **Profile** — редактирование своего профиля, привязка Telegram, настройки уведомлений (email/Telegram/WebPush)
8. **Notifications** — личные уведомления (новое сообщение, изменение расписания, дедлайн)

**Что НЕ может:**
- Создавать/редактировать документы
- Видеть отчёты (`/reports`)
- Видеть аналитику (`/analytics`)
- Управлять пользователями
- Настройки интеграции 1С

---

### 👨‍🏫 Преподаватель (`teacher`)

**Видит в меню:** Dashboard, Documents (full), Calendar, Messages, AI Assistant, Users (limited), Settings

**Сценарий:**
1. **Регистрация / создание администратором**
2. **Dashboard** — расширенные виджеты: задания на проверку, ближайшие пары, непрочитанные сообщения
3. **Documents** — создание/редактирование своих документов, шаблоны (read-only), маршруты согласования
4. **Calendar** — создание событий, назначение участников
5. **Users** — список студентов своих групп (read limited)
6. **Reports (limited)** — создание отчётов по своим группам, экспорт limited
7. **Messages** — групповые чаты со студентами
8. **AI Assistant** — расширенные права на RAG
9. **Profile / Settings** — стандартные

**Что НЕ может:**
- Видеть отчёты других преподавателей (только свои)
- Управлять учебными планами (curriculum)
- Создавать пользователей
- Настройки интеграции 1С

---

### 📋 Академический секретарь (`academic_secretary`)

**Видит в меню:** Dashboard, Documents, Analytics group (Reports + Analytics), Calendar, Messages, AI Assistant, Admin group (Users, Settings)

**Сценарий:**
1. **Создание администратором**
2. **Dashboard** — административные виджеты
3. **Documents** — full доступ, шаблоны (создание/редактирование)
4. **Schedule** (когда будет фронт) — **полное управление расписанием** (создание пар, замены, аудитории)
5. **Reports** — full create/read/export
6. **Analytics** — просмотр аналитики студентов (риски, посещаемость, успеваемость)
7. **Users** — read limited (просмотр)
8. **Calendar** — управление событиями
9. **Messages, AI, Profile** — стандартно

**Что НЕ может:**
- Управлять учебными планами (curriculum) — только читать
- Создавать пользователей
- Подписывать задания

---

### 📚 Методист (`methodist`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group, Calendar, Messages, AI Assistant, Admin group (Users + Integration + Settings)

**Сценарий:**
1. **Создание администратором**
2. **Dashboard** — методические виджеты
3. **Documents + Templates** — создание шаблонов документов, full CRUD
4. **Curriculum** (через документы и API) — **создание/редактирование учебных планов**, утверждение запрещено (только admin)
5. **Reports + Analytics** — full доступ, экспорт CSV/XLSX
6. **Schedule** — read full + limited update
7. **Users** — read limited
8. **Integration (1С)** — настройка синка внешних сотрудников/студентов из 1С
9. **AI Assistant** — расширенные права
10. **Calendar, Messages, Profile** — стандартно

**Что НЕ может:**
- Утверждать учебные планы (`ActionApprove` запрещён, только admin)
- Управлять расписанием (создавать пары — это секретарь)
- Подписывать электронной подписью (когда будет реализовано)

---

### 🛠 Системный администратор (`system_admin`)

**Видит в меню:** ВСЁ — Dashboard, Documents, Analytics, Calendar, Messages, AI Assistant, Users, Integration, Settings, `/admin/*`

**Сценарий:**
1. Создаётся при первом деплое или через миграцию
2. **Dashboard** — полная статистика
3. **Users** — full CRUD пользователей, ролей, профилей
4. **Documents** — full доступ ко всему
5. **Curriculum** — **утверждение учебных планов** (единственный с `ActionApprove`)
6. **Schedule, Reports, Analytics** — full
7. **Integration (1С)** — настройка
8. **Settings/Automation** — n8n workflows
9. **Settings/Appearance** — темы, brand
10. **Settings/Notifications** — глобальные настройки уведомлений
11. **Admin** (`/admin`) — admin-only роуты (если будут)

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
| **24** | **Tasks Module Frontend** | **✅ done 2026-04-25** (17 коммитов TDD, GH #200, релиз 0.101.0) |
| 25 | Schedule Lessons Frontend | **blocked** (нет backend для schedule_lessons — только events) |
| **26** | **Announcements Frontend + Backend Attachments** | **✅ done 2026-04-25** (24 коммита TDD full-stack, GH #202, релиз 0.102.0) |
| 27 | **Files Frontend** | pending medium |
| **28** | **🔐 SECURITY: Self-registration** | **✅ done 2026-04-25** (8 коммитов TDD, GH #199) |
| 23 | Web Speech API в AI-чате | pending medium |
| 2 | External calendars | pending medium |
| 5 | Auto schedule | pending low |
| 6 | Electronic signature | pending low |

---

## Краткая сводка

✅ **Готово к продакшну (12 модулей):** auth, users, documents, dashboard, notifications, messaging, reporting, integration, analytics, ai, **tasks (с 0.101.0)**, **announcements (с 0.102.0)**

⚠️ **Backend без UI:** schedule (пары — backend только частично), files

❌ **Не реализовано:** workflow (согласование), электронная подпись, авто-расписание, внешние календари, schedule lessons backend

🔐 **Безопасность:** privilege escalation при регистрации **закрыта** (GitHub #199, фикс 2026-04-25). Глубинная защита в 4 слоях: domain → usecase → handler → frontend.

📊 **Прогресс:** на 2026-04-25 закрыто 71+ GitHub issues, ~535+ коммитов. Code review compliance: все недавние релизы (0.100.1, 0.101.0, 0.102.0) прошли с оценкой ≥9/10 по TDD, DDD, CA.
