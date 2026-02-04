# Информационная система академического секретаря/методиста

![Backend CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/workflows/Backend%20CI/badge.svg)
![Documentation CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/workflows/Documentation%20CI/badge.svg)
![Security](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/workflows/Security%20Audit/badge.svg)
![Database CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/workflows/Database%20CI/badge.svg)

## 🎯 Основные возможности

* **Управление документами**: Полный жизненный цикл от создания до архивирования
* **Шаблоны документов**: Быстрое создание типовых документов с переменными
* **Автоматизация рабочих процессов**: Процессы согласования с маршрутизацией по ролям
* **Управление расписанием**: Академическое планирование и оптимизация ресурсов
* **Предиктивная аналитика**: Раннее выявление студентов в зоне риска
* **Отчетность и аналитика**: Комплексная бизнес-аналитика
* **Внутренние сообщения**: Real-time чаты, прямые и групповые сообщения через WebSocket
* **Мультиканальные уведомления**: Email (Composio Gmail) и Telegram Bot интеграция
* **Поддержка интеграций**: Бесшовное подключение к 1С и внешним системам
* **Мультиязычность (i18n)**: Поддержка русского, английского, французского и арабского языков
* **PWA**: Установка как приложение, офлайн-режим, push-уведомления
* **Темы оформления**: Светлая и тёмная тема с анимированными фонами
* **Доступность (a11y)**: WCAG 2.1 AA, клавиатурная навигация, ARIA-атрибуты
* **Резервное копирование**: Автоматический бэкап PostgreSQL и MinIO с offsite sync

## 🏗️ Архитектура

Построено на принципах **модульного монолита**:
- Domain-Driven Design (DDD)
- Паттерны Clean Architecture
- Event-driven коммуникация
- Структура, готовая к переходу на микросервисы

## 🚀 Быстрый старт

### Системные требования

- **Go** 1.25+
- **Node.js** 25+ (Current)
- **PostgreSQL** 17+
- **Redis** 7+
- **Docker** & Docker Compose

### Структура репозитория

Это **монорепозиторий**, содержащий:
- `backend/` - Backend на Go (модульный монолит)
- `frontend/` - Frontend на Next.js 15
- `docs/` - Документация проекта

### Быстрый запуск

**Backend:**
```bash
cd backend
cp .env.example .env
go mod download
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

**Docker (Full Stack):**
```bash
docker-compose up -d
```

## 📚 Документация

### 🏃‍♂️ Быстрый старт для разработчиков

- [📖 Руководство по разработке](docs/development/development-guide.md) - **Начните отсюда!** Полное руководство
- [🧩 Clean Code и паттерны](docs/development/clean-code-patterns.md) - Практики уровня Senior и SOLID
- [🔀 Руководство по Pull Request](docs/development/pull-request-guide.md) - Процесс и шаблоны PR

### 🏗️ Архитектура и дизайн

- [🎯 Обзор проекта](docs/project-overview.md) - Цели, пользователи и требования
- [🧩 Модульная архитектура](docs/architecture/modular-architecture.md) - Модульный дизайн на основе DDD
- [🚀 Руководство по миграции на микросервисы](docs/architecture/microservices-migration-guide.md) - Стратегия миграции
- [📄 Жизненный цикл документов](docs/workflows/document-lifecycle.md) - Бизнес-процессы

### 📋 Процесс разработки

- [🔧 Руководство по Git в терминале](docs/development/git-terminal-guide.md) - Полный справочник команд Git
- [🔄 Управление спринтами](docs/development/sprint-management.md) - Agile workflow с GitHub Projects
- [🧪 Стратегия тестирования](docs/development/testing-strategy.md) - Подход к тестированию
- [🔄 CI/CD Workflows](docs/development/ci-cd-workflows.md) - Пайплайн автоматизации

### 🚀 Развертывание и эксплуатация

- [🐳 Настройка Docker](docs/deployment/docker-setup.md) - Руководство по контейнеризации
- [☁️ Инфраструктура](docs/deployment/infrastructure.md) - Облачная архитектура и масштабирование
- [🔧 Конфигурация окружения](docs/deployment/environment.md) - Переменные окружения
- [☁️ Развертывание в production](docs/deployment/production-deploy.md) - Настройка production
- [🛡️ Руководство по безопасности](docs/security/security-guidelines.md) - Фреймворк безопасности
- [💾 Резервное копирование](backup/README.md) - Backup PostgreSQL, MinIO и offsite sync

### 📊 API и интеграции

- [📖 Справочник REST API](docs/api/api-documentation.md) - Полная документация API
- [🔐 Аутентификация](docs/api/authentication.md) - Endpoints аутентификации
- [📄 API управления документами](docs/api/documents.md) - Endpoints документов
- [💬 Messaging API](docs/api/messaging.md) - API внутренних сообщений и чатов
- [📧 Composio Gmail Integration](docs/integrations/composio-gmail.md) - Email уведомления через Composio
- [📱 Telegram Integration](docs/integrations/telegram-bot.md) - Уведомления через Telegram бота
- [👥 Роли и права доступа](docs/users/roles-and-permissions.md) - Роли пользователей и контроль доступа

## 🏗️ Технологический стек

### Backend (Модульный монолит)

- **Язык**: Go 1.25+
- **Архитектура**: DDD + Clean Architecture
- **Фреймворк**: Gin + gRPC (готов к микросервисам)
- **База данных**: PostgreSQL 17+ (основная), Redis (кеш/сессии)
- **Messaging**: In-memory EventBus (готов к Kafka)
- **Аутентификация**: OAuth 2.0 + JWT
- **Уведомления**: Email (Composio Gmail API) + Telegram Bot (polling/webhook)
- **Паттерны**: Repository, CQRS, Event Sourcing, Unit of Work
- **Логирование**: Структурированное JSON логирование с correlation IDs
- **Кеширование**: Redis с Decorator pattern (прирост производительности в 30x)

### Frontend

- **Фреймворк**: Next.js 15 + TypeScript 5.7
- **UI библиотека**: Material-UI 6
- **Управление состоянием**: Zustand 5
- **Загрузка данных**: SWR 2.3 + axios 1.7
- **Тестирование**: Jest 29 + React Testing Library 16 + Playwright 1.49
- **Интернационализация**: next-intl (ru, en, fr, ar с RTL)
- **PWA**: Service Worker, офлайн-режим, manifest.json
- **Темы**: next-themes с анимированными shader-фонами
- **Доступность**: ARIA-атрибуты, клавиатурная навигация, фокус-менеджмент

### Инфраструктура

- **Оркестрация**: Kubernetes (GKE/AKS)
- **Мониторинг**: Prometheus + Grafana + Loki + Uptime Kuma
- **Distributed Tracing**: OpenTelemetry SDK + OTEL Collector + Grafana Tempo
- **Алерты**: Grafana Alerting → Telegram (CPU, RAM, Disk, API errors, Backup)
- **Логирование**: Loki + Promtail (агрегация логов из контейнеров)
- **CI/CD**: GitHub Actions (Backend CI, Security, Database, PR Validation, Dependabot)
- **Load Balancer**: Nginx + CloudFlare

## 🔒 Безопасность и производительность

### Реализованные меры безопасности

- ✅ **JWT с полными security claims** (iat, nbf, jti, aud, iss)
- ✅ **Bcrypt cost 14** - защита от брутфорса
- ✅ **Timing attack prevention** - dummy hash comparison
- ✅ **Rate limiting** (5 req/15 min) - защита от DDoS
- ✅ **Security headers** - 6 заголовков безопасности
- ✅ **Полное логирование** всех security events

### Оптимизация производительности

| Метрика | Без кеша | С Redis кешем | Улучшение |
|---------|----------|---------------|-----------|
| GetUserByID | 15-20ms | 0.5-1ms | **30x быстрее** |
| GetUserByEmail | 15-20ms | 0.5-1ms | **30x быстрее** |
| Login (полный flow) | 350-400ms | 200-250ms | **40% быстрее** |
| Пропускная способность | ~200-300 req/sec | ~800-1200 req/sec | **4x больше** |

**Подробности**: См. [docs/LOGGING_AND_PERFORMANCE.md](docs/LOGGING_AND_PERFORMANCE.md)

## 🤝 Участие в разработке

1. Прочитайте [Руководство по разработке](docs/development/development-guide.md)
2. Ознакомьтесь с [Руководством по управлению спринтами](docs/development/sprint-management.md)
3. Следуйте [Процессу Pull Request](docs/development/pull-request-guide.md)

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](LICENSE) для деталей.

---

**Стек технологий (обновлено 2026-02-04)**:
- Backend: Go 1.25 • PostgreSQL 17 • Redis 7 • Gin • DDD • Clean Architecture
- Frontend: Next.js 15 • React 19 • TypeScript 5.7 • MUI 6 • Zustand 5
- DevOps: Docker • Kubernetes • GitHub Actions • Prometheus • Grafana

*"Повышение эффективности организации через автоматизацию и методическую поддержку."*
