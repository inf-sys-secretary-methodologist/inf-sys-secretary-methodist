# System Components Overview

> Информационная система "Секретарь-методист" - обзор реализованных компонентов

---

## Backend (Go + Gin)

### Полностью реализованные модули

| Модуль | Entities | Use Cases | API Handlers | Тесты |
|--------|----------|-----------|--------------|-------|
| **Auth** | User, Session, Role, Permission | AuthUseCase | AuthHandler + Middleware | ✅ |
| **Documents** | Document, DocumentType | DocumentUseCase, CategoryUseCase, TagUseCase | DocumentHandler, CategoryHandler, TagHandler | ✅ |
| **Schedule** | Event | EventUseCase | EventHandler | ✅ |
| **Reporting** | Report, ReportType, ReportHistory, ReportAccess | ReportUseCase | ReportHandler | ✅ |
| **Tasks** | Task, Project, TaskComment, TaskAttachment, TaskHistory, TaskWatcher, TaskChecklist | TaskUseCase, ProjectUseCase | TaskHandler, ProjectHandler | ✅ |
| **Announcements** | Announcement | AnnouncementUseCase | AnnouncementHandler | - |
| **Dashboard** | - | DashboardUseCase | DashboardHandler | ✅ |
| **Notifications** | - | ComposioEmailService | EmailHandler | - |

### Модули в разработке (только структура)

| Модуль | Статус |
|--------|--------|
| `users` | Пустая структура |
| `files` | Пустая структура |
| `workflow` | Пустая структура |
| `integration` | Пустая структура |

---

## Shared Infrastructure

| Компонент | Файлы | Описание |
|-----------|-------|----------|
| **Database** | `database/connection.go`, `transaction.go`, `errors.go` | PostgreSQL connection, transactions, error handling |
| **Cache** | `cache/redis_cache.go` | Redis cache implementation |
| **Storage** | `storage/s3_client.go`, `file_validator.go` | MinIO S3 client + file validator |
| **Config** | `config/config.go` | Environment-based configuration |
| **Logging** | `logging/logger.go`, `security_logger.go` | Structured logging + security logger |
| **Metrics** | `metrics/prometheus.go` | Prometheus metrics |
| **Middleware** | `middleware/rate_limiting.go`, `correlation_id.go` | Rate limiting, Correlation ID |
| **Validation** | `validation/validator.go` | Custom validator |
| **Sanitization** | `sanitization/sanitizer.go` | Input sanitizer |
| **Composio** | `composio/client.go` | External integration client |
| **HTTP Response** | `http/response/` | Standardized API responses |

---

## Frontend (Next.js + React + TypeScript)

### Страницы (App Router)

| Страница | Путь | Описание |
|----------|------|----------|
| Home | `/` | Главная страница |
| Login | `/login` | Страница входа |
| Register | `/register` | Страница регистрации |
| Dashboard | `/dashboard` | Панель управления с KPI |
| Calendar | `/calendar` | Календарь событий |
| Documents | `/documents` | Управление документами |
| Students | `/students` | Список студентов |
| Users | `/users` | Управление пользователями |
| Profile | `/profile` | Профиль пользователя |
| Forbidden | `/forbidden` | Страница ошибки доступа |

### UI Компоненты (`/components/ui`)

```
├── button.tsx          # Кнопки
├── card.tsx            # Карточки
├── input.tsx           # Поля ввода
├── floating-input.tsx  # Floating label inputs
├── label.tsx           # Лейблы
├── select.tsx          # Select dropdown
├── dialog.tsx          # Модальные окна
├── popover.tsx         # Popovers
├── dropdown-menu.tsx   # Dropdown меню
├── avatar.tsx          # Аватары
├── badge.tsx           # Бейджи
├── table.tsx           # Таблицы
├── tabs.tsx            # Табы
├── separator.tsx       # Разделители
├── calendar.tsx        # Календарь (date picker)
├── number-ticker.tsx   # Анимированные числа
└── tubelight-navbar.tsx # Навигация
```

> **Storybook**: Настроен для всех UI компонентов (`.stories.tsx`)

### Feature-компоненты

#### Calendar (`/components/calendar`)

| Компонент | Описание |
|-----------|----------|
| `FullCalendar.tsx` | Основной компонент календаря |
| `MonthView.tsx` | Отображение месяца |
| `WeekView.tsx` | Отображение недели |
| `DayView.tsx` | Отображение дня |
| `CalendarHeader.tsx` | Заголовок с навигацией |
| `EventCard.tsx` | Карточка события |
| `EventModal.tsx` | Модальное окно события |

#### Dashboard (`/components/dashboard`)

| Компонент | Описание |
|-----------|----------|
| `StatsCard.tsx` | Карточка статистики с KPI |
| `TrendChart.tsx` | График трендов |
| `ActivityFeed.tsx` | Лента активности |
| `ExportButton.tsx` | Кнопка экспорта |

#### Documents (`/components/documents`)

| Компонент | Описание |
|-----------|----------|
| `DocumentList.tsx` | Список документов |
| `DocumentFilters.tsx` | Фильтры документов |
| `DocumentUpload.tsx` | Загрузка документов |
| `DocumentPreview.tsx` | Превью документа |

#### Auth (`/components/auth`)

| Компонент | Описание |
|-----------|----------|
| `LoginForm.tsx` | Форма входа |
| `RegisterForm.tsx` | Форма регистрации |
| `withAuth.tsx` | HOC для защиты маршрутов |

#### Error (`/components/error`)

| Компонент | Описание |
|-----------|----------|
| `ErrorBoundary.tsx` | Error boundary для React |

### State Management

| Файл | Описание |
|------|----------|
| `stores/authStore.ts` | Zustand store для авторизации |

### Custom Hooks

| Hook | Описание |
|------|----------|
| `useAuth.ts` | Работа с авторизацией |
| `useCalendarEvents.ts` | Управление событиями календаря |
| `useDashboard.ts` | Данные дашборда |
| `useTheme.ts` | Управление темой |
| `useMediaQuery.ts` | Media queries |

### API Layer

| Файл | Описание |
|------|----------|
| `lib/api.ts` | Базовый API клиент |
| `lib/api/auth.ts` | API авторизации |
| `lib/api/dashboard.ts` | API дашборда |
| `lib/auth/jwt.ts` | JWT утилиты |
| `lib/auth/route-config.ts` | Конфигурация маршрутов |
| `lib/validations/auth.ts` | Валидация форм авторизации |

---

## Инфраструктура

### Docker & Containerization

| Файл | Описание |
|------|----------|
| `Dockerfile` | Docker образ приложения |
| `compose.yml` | Основной Docker Compose |
| `compose.monitoring.yml` | Мониторинг stack |
| `compose.override.yml.example` | Пример override конфига |

### Database Migrations

| # | Миграция | Описание |
|---|----------|----------|
| 001 | `create_users_table` | Таблица пользователей |
| 002 | `create_sessions_table` | Таблица сессий |
| 003 | `create_documents_schema` | Схема документов |
| 004 | `create_schedule_schema` | Схема расписания |
| 005 | `create_tasks_schema` | Схема задач |
| 006 | `create_reports_schema` | Схема отчетов |
| 007 | `create_events_schema` | Схема событий |
| 008 | `create_announcements_schema` | Схема объявлений |

### Monitoring Stack

| Компонент | Назначение |
|-----------|------------|
| **Prometheus** | Сбор метрик |
| **Grafana** | Визуализация и дашборды |
| **Loki** | Агрегация логов |
| **Promtail** | Сбор логов |

### Build System

- **NX Monorepo** - управление монорепозиторием
- **justfile** - команды для разработки

---

## Статус реализации

### По функционалу

| Функционал | Backend | Frontend | Интеграция |
|------------|:-------:|:--------:|:----------:|
| Аутентификация | ✅ | ✅ | ✅ |
| Управление документами | ✅ | ✅ | ⏳ |
| Календарь/Расписание | ✅ | ✅ | ⏳ |
| Задачи/Проекты | ✅ | ❌ | ❌ |
| Отчетность/Dashboard | ✅ | ✅ | ⏳ |
| Объявления | ✅ | ❌ | ❌ |
| Email уведомления | ✅ | ❌ | ⏳ |
| Студенты | ❌ | ✅ | ❌ |
| Пользователи | ❌ | ✅ | ❌ |

### Легенда

- ✅ Реализовано
- ⏳ В процессе / Частично
- ❌ Не реализовано

---

## Tech Stack

### Backend
- **Language**: Go 1.25
- **Framework**: Gin
- **Database**: PostgreSQL
- **Cache**: Redis
- **Storage**: MinIO (S3-compatible)
- **Auth**: JWT (access + refresh tokens)

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State**: Zustand
- **Forms**: React Hook Form + Zod
- **Components**: Radix UI primitives

### Infrastructure
- **Container**: Docker, Docker Compose
- **Metrics**: Prometheus
- **Logging**: Loki + Promtail
- **Visualization**: Grafana
- **Build**: NX Monorepo

---

## Прогресс

```
[██████████████░░░░░░░░] ~65%
```

**Основная архитектура и инфраструктура готовы.**

Требуется:
1. Интеграция frontend с backend API
2. Реализация модулей: users, files, workflow, integration
3. UI для задач и объявлений
4. E2E тестирование

---

*Последнее обновление: Декабрь 2025*
