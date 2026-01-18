# Технологический стек проекта

## Inf-sys-secretary-methodist

**Информационная система для секретаря-методиста** — система управления образовательным учреждением, включающая документооборот, отчётность, расписание, уведомления и внутренний мессенджер.

**Версия проекта:** 0.3.0
**Последнее обновление:** Январь 2026

---

## Оглавление

1. [Общая архитектура](#общая-архитектура)
2. [Backend (Go)](#backend-go)
3. [Frontend (TypeScript/React)](#frontend-typescriptreact)
4. [База данных PostgreSQL](#база-данных-postgresql)
5. [Кеширование (Redis)](#кеширование-redis)
6. [Файловое хранилище (MinIO)](#файловое-хранилище-minio)
7. [Инфраструктура и DevOps](#инфраструктура-и-devops)
8. [Мониторинг и логирование](#мониторинг-и-логирование)
9. [Внешние интеграции](#внешние-интеграции)
10. [Тестирование](#тестирование)
11. [Структура базы данных](#структура-базы-данных)

---

## Общая архитектура

Проект построен на принципах **Clean Architecture / Hexagonal Architecture** с разделением на модули:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend                                 │
│                   Next.js 15 + React 19                         │
│                      (TypeScript)                                │
└─────────────────────────┬───────────────────────────────────────┘
                          │ REST API / WebSocket
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Backend                                  │
│                      Go 1.25 + Gin                              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Modules                               │   │
│  │  auth │ documents │ notifications │ messaging │ users   │   │
│  │  schedule │ files │ integration │ reports                │   │
│  └─────────────────────────────────────────────────────────┘   │
└──────┬──────────────────┬──────────────────┬────────────────────┘
       │                  │                  │
       ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  PostgreSQL  │  │    Redis     │  │    MinIO     │
│      17      │  │      7       │  │     S3       │
└──────────────┘  └──────────────┘  └──────────────┘
```

**Структура модулей:**
```
internal/modules/<module>/
├── domain/           # Сущности, типы, интерфейсы репозиториев
├── application/      # Use cases, DTO, бизнес-логика
├── infrastructure/   # Реализация репозиториев (PostgreSQL, Redis)
└── interfaces/       # HTTP-хендлеры, middleware
```

---

## Backend (Go)

### Основные технологии

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Язык программирования** | Go | 1.25 | Основной язык бэкенда |
| **HTTP-фреймворк** | Gin | 1.11.0 | Роутинг, middleware, контекст |
| **Валидация** | go-playground/validator | v10.30.0 | Валидация структур через теги |
| **JWT-аутентификация** | golang-jwt | v5.3.0 | Access + Refresh токены |
| **UUID** | google/uuid | 1.6.0 | Генерация уникальных идентификаторов |
| **WebSocket** | gorilla/websocket | 1.5.3 | Real-time коммуникация (чаты) |
| **Криптография** | golang.org/x/crypto | 0.46.0 | Bcrypt хеширование паролей |

### Работа с данными

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **PostgreSQL драйвер** | lib/pq | 1.10.9 | Подключение к БД |
| **Redis клиент** | go-redis | v9.17.2 | Кеширование, сессии |
| **MinIO SDK** | minio-go | v7.0.97 | S3-совместимое хранилище |

### Фоновые задачи и метрики

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Планировщик** | gocron | v2.19.0 | Cron-задачи (напоминания, синхронизация) |
| **Cron** | robfig/cron | v3.0.1 | Парсинг cron-выражений |
| **Prometheus** | client_golang | 1.23.2 | Экспорт метрик |

### Генерация документов

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **PDF** | gofpdf | 1.16.2 | Генерация PDF-отчётов |
| **Excel** | excelize | v2.10.0 | Экспорт в XLSX |

### Тестирование

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Assertions** | testify | 1.11.1 | Unit/Integration тесты |
| **Mock Redis** | miniredis | v2.35.0 | Мок Redis для тестов |
| **Mocks** | uber/mock | 0.5.0 | Генерация моков |

---

## Frontend (TypeScript/React)

### Основные технологии

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Фреймворк** | Next.js | 15.1.0 | SSR, App Router, оптимизация |
| **UI-библиотека** | React | 19.0.0 | Компонентный UI |
| **Язык** | TypeScript | 5.7.2 | Строгая типизация |
| **Стили** | Tailwind CSS | 4.1.17 | Utility-first CSS |
| **PostCSS** | @tailwindcss/postcss | 4.1.17 | Обработка стилей |

### UI-компоненты (Radix UI)

| Компонент | Версия | Назначение |
|-----------|--------|------------|
| react-dialog | 1.1.15 | Модальные окна |
| react-dropdown-menu | 2.1.16 | Выпадающие меню |
| react-select | 2.2.6 | Селекты |
| react-tabs | 1.1.13 | Табы |
| react-avatar | 1.1.11 | Аватары |
| react-popover | 1.1.15 | Поповеры |
| react-switch | 1.2.6 | Переключатели |
| react-slider | 1.3.6 | Слайдеры |
| react-alert-dialog | 1.1.15 | Диалоги подтверждения |
| react-scroll-area | 1.2.10 | Кастомный скролл |

### Состояние и данные

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **State Manager** | Zustand | 5.0.2 | Глобальное состояние |
| **Data Fetching** | SWR | 2.3.0 | Кеширование запросов, ревалидация |
| **HTTP-клиент** | Axios | 1.7.9 | API-запросы |
| **Cookies** | js-cookie | 3.0.5 | Работа с cookies |

### Формы и валидация

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Формы** | React Hook Form | 7.66.0 | Управление формами |
| **Валидация** | Zod | 4.1.12 | Схемы валидации |
| **Resolvers** | @hookform/resolvers | 5.2.2 | Интеграция Zod + RHF |

### Анимации и визуализация

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Анимации** | Framer Motion | 12.23.24 | Плавные переходы |
| **Motion** | motion | 12.23.25 | Дополнительные анимации |
| **Графики** | Recharts | 3.5.1 | Визуализация данных |
| **Шейдеры** | @paper-design/shaders-react | 0.0.68 | Анимированные фоны |

### Утилиты

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Даты** | date-fns | 4.1.0 | Форматирование дат |
| **Иконки** | Lucide React | 0.553.0 | SVG-иконки |
| **Toasts** | Sonner | 2.0.7 | Уведомления |
| **Темы** | next-themes | 0.4.6 | Тёмная/светлая тема |
| **i18n** | next-intl | 4.6.1 | Мультиязычность |
| **Diff** | diff | 8.0.2 | Сравнение текстов |
| **Crop** | react-easy-crop | 5.5.6 | Обрезка изображений |
| **Calendar** | react-day-picker | 9.11.3 | Выбор дат |
| **Class Utils** | clsx, tailwind-merge, cva | разные | Условные классы |

### Поддерживаемые языки (i18n)

- Русский (RU) — основной
- English (EN)
- Français (FR)
- العربية (AR)

---

## База данных PostgreSQL

### Общие характеристики

| Параметр | Значение |
|----------|----------|
| **Версия** | PostgreSQL 17 Alpine |
| **Кодировка** | UTF-8 |
| **Коллация** | ru_RU.UTF-8 |
| **Количество миграций** | 18 |
| **Количество таблиц** | ~35 |

### Конфигурация подключения

| Параметр | Значение по умолчанию |
|----------|----------------------|
| `DB_MAX_OPEN_CONNS` | 25 |
| `DB_MAX_IDLE_CONNS` | 5 |
| `DB_CONN_MAX_LIFETIME` | 5m |
| `DB_SSL_MODE` | disable (dev) / require (prod) |

### Особенности реализации

| Функция | Описание |
|---------|----------|
| **Full-text Search** | `tsvector` + GIN-индексы с русской морфологией |
| **JSONB** | Гибкие метаданные, настройки уведомлений |
| **Enum Types** | `notification_type`, `notification_priority` |
| **Triggers** | Автообновление `updated_at`, версионирование документов |
| **Soft Delete** | `deleted_at` для документов и сообщений |
| **Partial Indexes** | Оптимизация запросов по условиям |
| **Constraints** | CHECK, UNIQUE, FOREIGN KEY с CASCADE/RESTRICT |

### Роли пользователей

| Роль | Описание |
|------|----------|
| `system_admin` | Системный администратор |
| `methodist` | Методист |
| `academic_secretary` | Учёный секретарь |
| `teacher` | Преподаватель |
| `student` | Студент |

---

## Кеширование (Redis)

### Характеристики

| Параметр | Значение |
|----------|----------|
| **Версия** | Redis 7 Alpine |
| **Pool Size** | 10 (настраивается) |
| **Аутентификация** | Опциональный пароль |

### Использование

| Назначение | Описание |
|------------|----------|
| **Сессии** | Хранение refresh-токенов |
| **Кеш** | Часто запрашиваемые данные |
| **Rate Limiting** | Ограничение запросов |
| **Pub/Sub** | Real-time уведомления |

---

## Файловое хранилище (MinIO)

### Характеристики

| Параметр | Значение |
|----------|----------|
| **Протокол** | S3-совместимый API |
| **Порт API** | 9000 |
| **Порт Console** | 9001 |
| **Bucket по умолчанию** | `documents` |
| **Макс. размер файла** | 100 MB (настраивается) |

### Хранимые данные

- Документы (PDF, DOCX, XLSX)
- Вложения сообщений
- Аватары пользователей
- Резервные копии

---

## Инфраструктура и DevOps

### Docker Compose

| Сервис | Образ | Порт | Описание |
|--------|-------|------|----------|
| `postgres` | postgres:17-alpine | 5432 (internal) | База данных |
| `redis` | redis:7-alpine | 6379 (internal) | Кеш |
| `minio` | minio/minio:latest | 9000, 9001 | Файлы |
| `backend` | Custom | 8080 | Go API |
| `frontend` | Custom | 3000 | Next.js |
| `backup` | Custom | — | Бэкап-сервис |

### Система резервного копирования

| Функция | Описание |
|---------|----------|
| **PostgreSQL backup** | pg_dump с gzip-сжатием |
| **MinIO backup** | mc mirror в tar.gz |
| **Расписание** | Cron (по умолчанию 2:00 ежедневно) |
| **Ретенция** | 7 дней (настраивается) |
| **Remote Sync** | S3-совместимое offsite хранилище |
| **Шифрование** | Age / GPG (опционально) |
| **Уведомления** | Telegram, Webhook, Email |

### Поддерживаемые offsite-хранилища

- AWS S3
- Backblaze B2
- Yandex Object Storage
- Selectel Cloud Storage

### NX Monorepo

Проект управляется через **NX** для:
- Кеширования сборок
- Параллельного выполнения задач
- Управления зависимостями между проектами

---

## Мониторинг и логирование

### Prometheus Metrics

| Метрика | Описание |
|---------|----------|
| `http_requests_total` | Общее количество HTTP-запросов |
| `http_request_duration_seconds` | Время выполнения запросов |
| `db_connections_active` | Активные подключения к БД |
| `backup_last_success_timestamp` | Время последнего успешного бэкапа |

### Логирование

| Параметр | Значение |
|----------|----------|
| **Формат** | JSON (production) / Text (development) |
| **Уровни** | debug, info, warn, error |
| **Labels** | service=backend, logging=promtail |

### Опциональный стек мониторинга

| Компонент | Назначение |
|-----------|------------|
| Prometheus | Сбор метрик |
| Grafana | Дашборды |
| Promtail | Сбор логов |
| Loki | Хранение логов |

---

## Внешние интеграции

### 1С:Предприятие

| Параметр | Описание |
|----------|----------|
| **Протокол** | REST / OData |
| **Синхронизация** | Сотрудники, студенты |
| **Расписание** | Каждые 6 часов (cron) |
| **Batch Size** | 100 записей |
| **Конфликты** | Автоматическое разрешение с логированием |

### Telegram Bot

| Функция | Описание |
|---------|----------|
| **Режимы** | Polling / Webhook |
| **Привязка аккаунта** | Верификация по коду |
| **Уведомления** | Напоминания, события, документы |

### Email (Composio)

| Функция | Описание |
|---------|----------|
| **Провайдер** | Gmail API через Composio |
| **Типы писем** | Уведомления, напоминания, дайджесты |

---

## Тестирование

### Backend (Go)

| Тип | Инструмент | Описание |
|-----|------------|----------|
| Unit | testify | Изолированные тесты |
| Integration | testify + miniredis | Тесты с моками |
| Coverage | go test -cover | Покрытие кода |

### Frontend (TypeScript)

| Тип | Инструмент | Описание |
|-----|------------|----------|
| Unit | Jest | Тесты компонентов |
| Component | Testing Library | React-тестирование |
| E2E | Playwright | Сквозные тесты |

### Команды

```bash
# Backend
go test ./...
go test -cover ./...

# Frontend
npm run test           # Unit-тесты
npm run test:coverage  # С покрытием
npm run test:e2e       # E2E Playwright
```

---

## Структура базы данных

### Модули и таблицы

#### Модуль Auth (Аутентификация)

| Таблица | Описание |
|---------|----------|
| `users` | Пользователи системы |
| `sessions` | JWT-сессии (refresh tokens) |

#### Модуль Documents (Документооборот)

| Таблица | Описание |
|---------|----------|
| `documents` | Основная таблица документов |
| `document_types` | Типы документов (приказ, служебная записка и т.д.) |
| `document_categories` | Категории (иерархические) |
| `document_versions` | История версий |
| `document_routes` | Маршруты согласования |
| `document_permissions` | Права доступа |
| `document_relations` | Связи между документами |
| `document_tags` | Теги |
| `document_tag_relations` | Связь документ-тег |
| `document_history` | Аудит действий |
| `document_public_links` | Публичные ссылки |

#### Модуль Schedule (Расписание)

| Таблица | Описание |
|---------|----------|
| `schedule_events` | События расписания |
| `event_participants` | Участники событий |
| `event_reminders` | Напоминания |

#### Модуль Tasks (Задачи)

| Таблица | Описание |
|---------|----------|
| `tasks` | Задачи |
| `task_assignments` | Назначения на задачи |

#### Модуль Reports (Отчётность)

| Таблица | Описание |
|---------|----------|
| `reports` | Отчёты |
| `custom_reports` | Пользовательские отчёты (конструктор) |

#### Модуль Users (Пользователи)

| Таблица | Описание |
|---------|----------|
| `user_profiles` | Расширенные профили |
| `user_avatars` | Аватары |

#### Модуль Files (Файлы)

| Таблица | Описание |
|---------|----------|
| `file_metadata` | Метаданные файлов MinIO |

#### Модуль Notifications (Уведомления)

| Таблица | Описание |
|---------|----------|
| `notifications` | In-app уведомления |
| `notification_preferences` | Настройки уведомлений |
| `notification_delivery_log` | Лог доставки |
| `user_telegram_connections` | Привязки Telegram |
| `user_slack_connections` | Привязки Slack |
| `telegram_verification_codes` | Коды верификации |

#### Модуль Integration (Интеграции)

| Таблица | Описание |
|---------|----------|
| `integration_sync_logs` | Логи синхронизации с 1С |
| `integration_conflicts` | Конфликты синхронизации |

#### Модуль Messaging (Мессенджер)

| Таблица | Описание |
|---------|----------|
| `conversations` | Чаты (личные и групповые) |
| `conversation_participants` | Участники чатов |
| `messages` | Сообщения |
| `message_attachments` | Вложения сообщений |

### ER-диаграмма (ключевые связи)

```
                                users
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
     documents              notifications           conversations
          │                       │                       │
    ┌─────┴─────┐                 │                       │
    │           │                 │                       ▼
    ▼           ▼                 │            conversation_participants
document_   document_             │                       │
versions    routes                │                       ▼
    │           │                 │                   messages
    │           │                 │                       │
    ▼           ▼                 ▼                       ▼
document_   document_    notification_           message_attachments
history    permissions   delivery_log                     │
                                                          ▼
                                                    file_metadata
                                                          │
                                                          ▼
                                                     MinIO S3
```

---

## Переменные окружения

### Обязательные (Production)

```env
# Database
POSTGRES_PASSWORD=<secure_password>
POSTGRES_DB=inf_sys_db
POSTGRES_USER=postgres

# JWT (ОБЯЗАТЕЛЬНО сгенерировать уникальные!)
JWT_ACCESS_SECRET=<random_32_chars>
JWT_REFRESH_SECRET=<random_32_chars>

# MinIO
MINIO_ROOT_USER=<username>
MINIO_ROOT_PASSWORD=<secure_password>
```

### Опциональные

```env
# JWT TTL
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Redis
REDIS_PASSWORD=<password>
REDIS_POOL_SIZE=10

# S3/MinIO
S3_BUCKET_NAME=documents
S3_MAX_FILE_SIZE=104857600

# Telegram Bot
TELEGRAM_BOT_TOKEN=<bot_token>
TELEGRAM_BOT_USERNAME=<bot_username>

# 1C Integration
INTEGRATION_1C_ENABLED=false
INTEGRATION_1C_BASE_URL=<url>

# Monitoring
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## Быстрый старт

### Локальная разработка

```bash
# Клонирование
git clone <repo_url>
cd inf-sys-secretary-methodist

# Копирование конфигурации
cp compose.override.yml.example compose.override.yml
cp .env.example .env

# Редактирование .env (установить пароли)
vim .env

# Запуск инфраструктуры
docker compose up -d postgres redis minio

# Запуск бэкенда
go run cmd/server/main.go

# Запуск фронтенда
cd frontend && npm install && npm run dev
```

### Production-деплой

```bash
# Сборка и запуск всех сервисов
docker compose up -d

# С бэкап-сервисом
docker compose --profile backup up -d
```

---

## Контакты и ссылки

- **Репозиторий:** [GitHub](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist)
- **Документация:** `/docs`
- **Миграции:** `/migrations`

---

*Документ сгенерирован автоматически на основе анализа кодовой базы.*
