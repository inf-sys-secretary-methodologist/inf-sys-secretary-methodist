# 📚 Индекс документации

> 👈 **Основная информация о проекте**: См. [../README.md](../README.md) для обзора проекта, возможностей и технологического стека.

## 📂 Структура документации

```
docs/
├── 🏗️ architecture/          # Проектирование системы и паттерны
│   ├── modular-architecture.md
│   └── microservices-migration-guide.md
├── 💻 development/           # Руководства и практики разработки
│   ├── development-guide.md
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
│   └── documents.md
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
2. [🧩 Модульная архитектура](architecture/modular-architecture.md) - Проектирование системы
3. [🔀 Руководство по Pull Request](development/pull-request-guide.md) - Процесс участия

### Для архитекторов

1. [🧩 Модульная архитектура](architecture/modular-architecture.md) - Паттерны DDD
2. [🧩 Clean Code и паттерны](development/clean-code-patterns.md) - Принципы SOLID
3. [🚀 Миграция на микросервисы](architecture/microservices-migration-guide.md) - Стратегия миграции

### Для DevOps

1. [🐳 Настройка Docker](deployment/docker-setup.md) - Контейнеризация
2. [☁️ Инфраструктура](deployment/infrastructure.md) - Облачная архитектура
3. [🔄 CI/CD Workflows](development/ci-cd-workflows.md) - Автоматизация

### Оптимизация и мониторинг

1. [📊 Логирование и производительность](LOGGING_AND_PERFORMANCE.md) - Полное руководство по:
   - Security logging (логи безопасности)
   - Audit logging (audit trail)
   - Performance logging (мониторинг производительности)
   - Redis caching (прирост в 30x)
   - Correlation IDs для трейсинга
   - Метрики и best practices

---

**💡 Совет**: Все ссылки на документацию в главном README относительны к папке docs.
