# 📚 Индекс документации

> 👈 **Основная информация о проекте**: См. [../README.md](../README.md) для обзора проекта, возможностей и технологического стека.

## 📂 Структура документации

```
docs/
├── 🏗️ architecture/          # Проектирование системы и паттерны
│   ├── modular-architecture.md
│   ├── ddd-domain-modeling.md          # 🆕 DDD: Bounded Contexts, Aggregates, Events
│   ├── module-interaction-guide.md     # 🆕 Взаимодействие модулей, Event-Driven Architecture
│   ├── tech-stack-rationale.md         # 🆕 Обоснование выбора технологий
│   └── microservices-migration-guide.md
├── 💻 development/           # Руководства и практики разработки
│   ├── development-guide.md
│   ├── tdd-guide.md                    # 🆕 TDD: Red-Green-Refactor, примеры
│   ├── clean-code-patterns.md
│   ├── sprint-management.md
│   ├── pull-request-guide.md
│   ├── git-terminal-guide.md
│   ├── testing-strategy.md
│   └── ci-cd-workflows.md
├── 🚀 deployment/            # Инфраструктура и развертывание
│   ├── docker-setup.md
│   ├── infrastructure.md
│   ├── environment.md
│   └── production-deploy.md
├── 📊 api/                   # Документация API
│   ├── api-documentation.md
│   ├── authentication.md
│   ├── documents.md
│   ├── messaging.md
│   └── schedule.md
├── 🔗 integrations/          # Внешние интеграции
│   ├── composio-gmail.md
│   └── telegram-bot.md
├── 👥 users/                 # Управление пользователями
│   └── roles-and-permissions.md
├── 🔄 workflows/             # Бизнес-процессы
│   └── document-lifecycle.md
├── 🔒 security/              # Рекомендации по безопасности
│   └── security-guidelines.md
├── LOGGING_AND_PERFORMANCE.md # Логирование и оптимизация
├── project-overview.md       # Бизнес-требования
├── uptime-kuma.md            # Status page и мониторинг uptime
└── grafana-alerting.md       # Алерты с Telegram уведомлениями
```

## 🚀 Быстрая навигация

### Для новых разработчиков

1. [📖 Руководство по разработке](development/development-guide.md) - Полное руководство
2. [🧪 TDD Guide](development/tdd-guide.md) - **НОВОЕ**: Test-Driven Development процесс
3. [🏛️ DDD Domain Modeling](architecture/ddd-domain-modeling.md) - **НОВОЕ**: Domain-Driven Design
4. [🧩 Модульная архитектура](architecture/modular-architecture.md) - Проектирование системы
5. [🔀 Руководство по Pull Request](development/pull-request-guide.md) - Процесс участия

### Для архитекторов

1. [🏛️ DDD Domain Modeling](architecture/ddd-domain-modeling.md) - **НОВОЕ**: Bounded Contexts, Aggregates, Events
2. [🔄 Module Interaction Guide](architecture/module-interaction-guide.md) - **НОВОЕ**: Event-Driven Architecture, Saga Patterns
3. [⚙️ Tech Stack Rationale](architecture/tech-stack-rationale.md) - **НОВОЕ**: Почему Go, PostgreSQL, React, TypeScript
4. [🧩 Модульная архитектура](architecture/modular-architecture.md) - Паттерны модульного монолита
5. [🧩 Clean Code и паттерны](development/clean-code-patterns.md) - Принципы SOLID
6. [🚀 Миграция на микросервисы](architecture/microservices-migration-guide.md) - Стратегия миграции

### Для DevOps

1. [🐳 Настройка Docker](deployment/docker-setup.md) - Контейнеризация
2. [☁️ Инфраструктура](deployment/infrastructure.md) - Облачная архитектура
3. [🔄 CI/CD Workflows](development/ci-cd-workflows.md) - Автоматизация
4. [⚙️ Переменные окружения](deployment/environment.md) - Конфигурация

### Мониторинг стек

Проект включает полный стек мониторинга (запускается через `compose.monitoring.yml` или `--profile monitoring`):

| Сервис | Порт | Назначение |
|--------|------|------------|
| Prometheus | 9090 | Сбор метрик |
| Grafana | 3001 | Визуализация и алерты (admin/admin) |
| Loki | 3100 | Агрегация логов |
| Promtail | - | Сбор логов из контейнеров |
| Uptime Kuma | 3002 | Status page и мониторинг uptime |

**Запуск с мониторингом:**
```bash
# Полный стек мониторинга
docker compose -f compose.yml -f compose.monitoring.yml --profile monitoring up -d
```

**Документация:**
- [Uptime Kuma](uptime-kuma.md) - Настройка status page
- [Grafana Alerting](grafana-alerting.md) - Алерты с Telegram уведомлениями

**Настроенные алерты Grafana:**
| Алерт | Условие | Severity |
|-------|---------|----------|
| High CPU Usage | CPU > 80% (5 мин) | warning |
| High Memory Usage | RAM > 85% (5 мин) | warning |
| High Disk Usage | Disk > 85% (5 мин) | warning |
| High API Error Rate | 5xx > 1% (5 мин) | critical |
| High API Latency | p95 > 2s (5 мин) | warning |
| Backup Failed | success = 0 | critical |
| Backup Stale | age > 24h | warning |

**Health endpoints:**
- `/health` - Полный health check (DB + Redis)
- `/live` - Kubernetes liveness probe
- `/ready` - Kubernetes readiness probe
- `/metrics` - Prometheus метрики

### Интеграции

1. [📧 Composio Gmail](integrations/composio-gmail.md) - Интеграция с Gmail для email уведомлений
2. [📱 Telegram Bot](integrations/telegram-bot.md) - Интеграция с Telegram для push-уведомлений
3. [🏢 1C Integration](integrations/README.md#1c-integration) - Синхронизация данных с 1С

### Frontend Features

1. **🌍 Мультиязычность (i18n)**
   - Поддержка 4 языков: русский, английский, французский, арабский
   - RTL поддержка для арабского языка
   - Переключатель языка в настройках
   - Локализованные даты, числа, валюты

2. **📱 PWA (Progressive Web App)**
   - Установка как нативное приложение
   - Офлайн-режим через Service Worker
   - Push-уведомления
   - manifest.json для мобильных устройств

3. **🎨 Настройки внешнего вида** (`/settings/appearance`)
   - Выбор темы: светлая / тёмная / системная
   - Анимированные шейдерные фоны (GrainGradient, Warp, MeshGradient)
   - Настройка скорости и интенсивности анимации
   - Режим уменьшения движения для доступности
   - Сохранение в localStorage через Zustand

4. **♿ Доступность (a11y)**
   - WCAG 2.1 AA соответствие
   - ARIA-атрибуты для скринридеров
   - Клавиатурная навигация
   - Фокус-менеджмент

5. **🔔 Настройки уведомлений** (`/settings/notifications`)
   - Привязка Telegram аккаунта
   - Настройка типов уведомлений

6. **💬 Сообщения** (`/messages`)
   - Прямые сообщения и групповые чаты
   - Real-time обмен через WebSocket
   - Прикрепление файлов и изображений
   - AI-подсказки (запланировать встречу, создать задачу, и т.д.)
   - Ответы, редактирование, удаление сообщений
   - Индикаторы набора и статус прочтения

7. **💾 Резервное копирование** (DevOps)
   - Автоматический бэкап PostgreSQL и MinIO
   - Cron расписание (по умолчанию 2:00 ежедневно)
   - Offsite sync на внешний S3
   - Документация: [backup/README.md](../backup/README.md)

### Оптимизация и мониторинг

1. [📊 Логирование и производительность](LOGGING_AND_PERFORMANCE.md) - Полное руководство по:
   - Security logging (логи безопасности)
   - Audit logging (audit trail)
   - Performance logging (мониторинг производительности)
   - Redis caching (прирост в 30x)
   - Correlation IDs для трейсинга
   - Метрики и best practices

---

## 📖 Новые подробные руководства

### 🏛️ [DDD Domain Modeling Guide](architecture/ddd-domain-modeling.md)
**12,000+ строк** | Дата актуальности: 2025-12-09

Полное руководство по Domain-Driven Design:
- **Почему DDD?** - Обоснование выбора с trade-offs
- **10 Bounded Contexts** - Auth, Documents, Workflow, Schedule, Tasks, Reporting, Notifications, Files, Integration, Users
- **Ubiquitous Language** - Словарь из 50+ терминов
- **Aggregates и Entities** - С примерами кода и инвариантами
- **Domain Events Catalog** - 30+ событий с примерами
- **Anti-Corruption Layer** - Для интеграции с 1С
- **Event Storming** - Процесс моделирования домена

### 🧪 [TDD Process Guide](development/tdd-guide.md)
**7,000+ строк** | Дата актуальности: 2025-12-09

Полное руководство по Test-Driven Development:
- **Почему TDD?** - Обоснование выбора с trade-offs
- **Red-Green-Refactor цикл** - С детальными примерами
- **TDD для DDD** - Тестирование Aggregates, Domain Services, Events
- **Outside-in TDD** - От E2E тестов к Unit тестам
- **Test Doubles стратегия** - Mock vs Stub vs Fake vs Spy
- **Best Practices** - Naming, AAA pattern, Table-driven tests
- **Frontend TDD** - Тестирование React компонентов

### ⚙️ [Tech Stack Rationale](architecture/tech-stack-rationale.md)
**1,800+ строк** | Дата актуальности: 2025-12-09

Детальное обоснование выбора технологий:
- **Backend**: Go 1.25+ (vs Python/Java/Rust) - с benchmarks
- **Database**: PostgreSQL 17+ (vs MySQL/MongoDB) - с примерами
- **Cache**: Redis 7+ (vs Memcached) - с use cases
- **Frontend**: Next.js 15 + React 19 (vs CRA/Vue/Angular)
- **TypeScript 5.7** (vs JavaScript/Flow)
- **Tailwind CSS v4** (vs CSS Modules/styled-components)
- **Testing Stack**: Jest, Playwright, testify, gomock
- **Infrastructure**: Docker, Kafka, Prometheus, Grafana

### 🔄 [Module Interaction Guide](architecture/module-interaction-guide.md)
**2,500+ строк** | Дата актуальности: 2025-12-09

Взаимодействие модулей в модульном монолите:
- **Event-Driven Architecture** - Асинхронная коммуникация через события
- **Domain Events Catalog** - Полный каталог всех 30+ событий
- **Saga Patterns** - Orchestration vs Choreography с примерами
- **Eventual Consistency** - Паттерны для согласованности данных
- **Resilience Patterns** - Circuit Breaker, Retry, Bulkhead
- **Anti-Corruption Layer** - Интеграция с внешней системой 1С
- **Migration to Microservices** - Готовность к выделению модулей

---

## ❓ Часто задаваемые вопросы

### DDD vs TDD - что выбрать?

**Ответ:** Это НЕ взаимоисключающие подходы! Мы используем **оба вместе**:

- **DDD (Domain-Driven Design)** - это подход к **архитектуре** (ЧТО строить)
  - Определяет структуру кода (Entities, Aggregates, Value Objects)
  - Разделяет систему на модули (Bounded Contexts)
  - Моделирует бизнес-логику (Ubiquitous Language)

- **TDD (Test-Driven Development)** - это **процесс разработки** (КАК строить)
  - Определяет порядок написания кода (сначала тест, потом код)
  - Гарантирует качество через автоматические тесты
  - Позволяет безопасно рефакторить

**Вместе:** DDD определяет архитектуру, TDD обеспечивает качество реализации.

См. подробное объяснение в [TDD Guide](development/tdd-guide.md#ddd-vs-tdd)

---

**💡 Совет**: Все ссылки на документацию в главном README относительны к папке docs.

---

**📅 Актуальность документа**
**Последнее обновление**: 2026-01-30
**Версия проекта**: 0.3.0
**Статус**: Актуальный

