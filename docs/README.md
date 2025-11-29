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
│   └── schedule.md
├── 🔗 integrations/          # Внешние интеграции
│   └── composio-gmail.md
├── 👥 users/                 # Управление пользователями
│   └── roles-and-permissions.md
├── 🔄 workflows/             # Бизнес-процессы
│   └── document-lifecycle.md
├── 🔒 security/              # Рекомендации по безопасности
│   └── security-guidelines.md
├── LOGGING_AND_PERFORMANCE.md # Логирование и оптимизация
└── project-overview.md       # Бизнес-требования
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

Проект включает полный стек мониторинга (запускается через `compose.monitoring.yml`):

| Сервис | Порт | Назначение |
|--------|------|------------|
| Prometheus | 9090 | Сбор метрик |
| Grafana | 3001 | Визуализация (admin/admin) |
| Loki | 3100 | Агрегация логов |
| Promtail | - | Сбор логов из контейнеров |

**Запуск с мониторингом:**
```bash
docker compose -f compose.yml -f compose.monitoring.yml up -d
```

**Health endpoints:**
- `/health` - Полный health check (DB + Redis)
- `/live` - Kubernetes liveness probe
- `/ready` - Kubernetes readiness probe
- `/metrics` - Prometheus метрики

### Интеграции

1. [📧 Composio Gmail](integrations/composio-gmail.md) - Интеграция с Gmail для email уведомлений

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
**12,000+ строк** | Дата актуальности: 2025-11-29

Полное руководство по Domain-Driven Design:
- **Почему DDD?** - Обоснование выбора с trade-offs
- **10 Bounded Contexts** - Auth, Documents, Workflow, Schedule, Tasks, Reporting, Notifications, Files, Integration, Users
- **Ubiquitous Language** - Словарь из 50+ терминов
- **Aggregates и Entities** - С примерами кода и инвариантами
- **Domain Events Catalog** - 30+ событий с примерами
- **Anti-Corruption Layer** - Для интеграции с 1С
- **Event Storming** - Процесс моделирования домена

### 🧪 [TDD Process Guide](development/tdd-guide.md)
**7,000+ строк** | Дата актуальности: 2025-11-29

Полное руководство по Test-Driven Development:
- **Почему TDD?** - Обоснование выбора с trade-offs
- **Red-Green-Refactor цикл** - С детальными примерами
- **TDD для DDD** - Тестирование Aggregates, Domain Services, Events
- **Outside-in TDD** - От E2E тестов к Unit тестам
- **Test Doubles стратегия** - Mock vs Stub vs Fake vs Spy
- **Best Practices** - Naming, AAA pattern, Table-driven tests
- **Frontend TDD** - Тестирование React компонентов

### ⚙️ [Tech Stack Rationale](architecture/tech-stack-rationale.md)
**1,800+ строк** | Дата актуальности: 2025-11-29

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
**2,500+ строк** | Дата актуальности: 2025-11-29

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
**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный

