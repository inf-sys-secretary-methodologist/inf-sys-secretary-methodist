# Роли и пользовательские потоки

> **Версия проекта:** 0.108.3 (см. `VERSION` в корне)
> **Состояние на:** 4 мая 2026 (после релизов password recovery, n8n absence-alert, Document.Update ownership, teacher analytics scope filter)
> **Источники:** код (`internal/modules/auth/domain/`, `frontend/src/lib/auth/`, `frontend/src/config/navigation.ts`), GitHub issues, `.taskmaster/`, `CHANGELOG.md`

> **Изменения в 0.102.2:** Уточнена концепция личных vs глобальных настроек. Каждый пользователь любой роли (включая студента) может настраивать своё рабочее окружение: тему оформления и подключённые лично к нему каналы уведомлений (Telegram, email, WebPush). Глобальные настройки (SMTP-сервер, push-провайдер, brand системы, n8n workflows, интеграция с 1С) — остаются только у системного администратора.

> **Изменения в 0.108.x:** v0.108.0 — полноценный flow восстановления пароля (request → verify → confirm) с anti-enumeration. v0.108.1 — алерты о пропусках студентов теперь отправляются и в n8n (раньше workflow JSON лежал, но не подключён). v0.108.2 — `Document.Update` теперь явно проверяет авторство; преподаватель не может редактировать чужие документы (доменное правило `Document.CanBeEditedBy` + 403). v0.108.3 — `/api/analytics/*` теперь применяет scope-фильтр «свои группы» для роли teacher: список групп выводится из `schedule_lessons + student_groups`, фильтр пушится в SQL (`WHERE group_name = ANY($N)`), запрос чужой группы или экспорт по чужим студентам возвращают 403/empty.

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

**Видит в меню:** Dashboard, Documents (просмотр), Schedule (просмотр), Calendar, Messages, AI Assistant, Profile

1. Регистрация → авто-логин
2. **Dashboard** — виджеты: ближайшие события, объявления, непрочитанные сообщения
3. **Documents** — только чтение публичных и доступных документов
4. **Schedule** — **просмотр расписания** своей группы (сетка по дням, фильтр по группе/преподавателю)
5. **Calendar** — свои события, расписание группы
6. **Messages** — WebSocket-чаты
7. **AI Assistant** — RAG с цитированием
8. **Tasks** — просмотр заданий (own read+execute)
9. **Announcements** — просмотр объявлений
10. *(Личные настройки — стандартно для всех ролей)*

**Что НЕ может:** создавать/редактировать расписание, создавать документы, отчёты (`denied`), аналитика, управление пользователями, любые системные настройки.

---

### 👨‍🏫 Преподаватель (`teacher`)

**Видит в меню:** Dashboard, Documents (full), Schedule (просмотр), Calendar, Messages, AI Assistant, Users (limited), Profile

1. Регистрация / создание администратором
2. **Dashboard** — виджеты: задания на проверку, ближайшие пары
3. **Documents** — создание/редактирование своих, шаблоны (read-only), маршруты согласования
4. **Schedule** — **просмотр расписания** своих пар (сетка по дням, фильтр по группе/аудитории)
5. **Calendar** — создание событий, назначение участников
6. **Users** — список студентов своих групп (read limited)
7. **Reports (limited)** — по своим группам, экспорт limited
8. **Messages** — групповые чаты со студентами
9. **AI Assistant** — расширенные права на RAG
10. *(Личные настройки — стандартно)*

**Что НЕ может:** создавать/редактировать расписание, видеть отчёты других преподавателей, управлять curriculum, создавать пользователей, любые системные настройки.

---

### 📋 Академический секретарь (`academic_secretary`)

**Видит в меню:** Dashboard, Documents, Analytics group (Reports + Analytics), Calendar, Messages, AI Assistant, Admin group (Users — read limited), Profile

1. Создание администратором
2. **Dashboard** — административные виджеты
3. **Documents** — full, шаблоны (создание/редактирование)
4. **Schedule** — **полное управление расписанием** (создание пар, замены, аудитории)
5. **Reports** — full create/read/export
6. **Analytics** — просмотр аналитики студентов (риски, посещаемость, успеваемость)
7. **Users** — read limited
8. **Calendar** — управление событиями
9. **Messages, AI** — стандартно
10. *(Личные настройки — стандартно)*

**Что НЕ может:** управлять curriculum (только читать), создавать пользователей, подписывать задания, любые системные настройки.

---

### 📚 Методист (`methodist`)

**Видит в меню:** Dashboard, Documents (full + Templates), Analytics group, Calendar, Messages, AI Assistant, Users (read limited), Profile

1. Создание администратором
2. **Dashboard** — методические виджеты
3. **Documents + Templates** — full CRUD, создание шаблонов документов
4. **Curriculum** — **создание/редактирование учебных планов**, отправка на утверждение администратору. Утверждение `ActionApprove` запрещено
5. **Reports + Analytics** — full доступ, экспорт CSV/XLSX
6. **Schedule** — read full + limited update
7. **Users** — read limited
8. **AI Assistant** — расширенные права
9. **Calendar, Messages** — стандартно
10. *(Личные настройки — стандартно)*

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

**Видит в меню:** ВСЁ — Dashboard, Documents, Analytics, Calendar, Messages, AI Assistant, Users, Integration, Settings, `/admin/*`

1. Создаётся при первом деплое или через миграцию
2. **Dashboard** — полная статистика
3. **Users** — full CRUD пользователей и ролей. Единственный, кто создаёт привилегированные роли
4. **Documents** — full доступ ко всему
5. **Curriculum** — **утверждение учебных планов** (единственный с `ActionApprove`), может отклонять с замечаниями
6. **Schedule, Reports, Analytics** — full
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
