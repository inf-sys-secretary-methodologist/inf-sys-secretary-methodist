# Backend - Information System of Academic Secretary/Methodologist

Модульный монолит на Go с использованием Domain-Driven Design (DDD) и Clean Architecture.

## 🏗️ Архитектура

Проект построен на принципах модульного монолита с четкими границами между модулями:

### Структура проекта

```
.
├── cmd/
│   └── server/           # Точка входа приложения
│       └── main.go
├── internal/
│   ├── modules/          # Бизнес-модули
│   │   ├── auth/         # Аутентификация и авторизация
│   │   ├── users/        # Управление пользователями
│   │   ├── documents/    # Управление документами
│   │   ├── workflow/     # Согласование и маршрутизация
│   │   ├── schedule/     # Расписание
│   │   ├── reporting/    # Отчетность
│   │   ├── tasks/        # Задачи
│   │   ├── notifications/# Уведомления
│   │   ├── files/        # Файловое хранилище
│   │   └── integration/  # Интеграции (1C)
│   ├── shared/           # Shared Kernel
│   │   ├── domain/       # Общие доменные компоненты
│   │   ├── infrastructure/# Общая инфраструктура
│   │   └── application/  # Общий application слой
│   └── gateway/          # API Gateway
└── pkg/                  # Публичные библиотеки
```

### Структура модуля

Каждый модуль следует Clean Architecture:

```
module_name/
├── domain/               # Доменный слой
│   ├── entities/        # Бизнес-сущности
│   ├── repositories/    # Интерфейсы репозиториев
│   ├── services/        # Доменные сервисы
│   └── events/          # Доменные события
├── application/         # Слой приложения
│   ├── usecases/       # Use cases (бизнес-логика)
│   ├── commands/       # Команды (CQRS)
│   └── queries/        # Запросы (CQRS)
├── infrastructure/     # Инфраструктурный слой
│   ├── persistence/    # Реализация репозиториев
│   ├── external/       # Внешние сервисы
│   └── messaging/      # Messaging/Events
└── interfaces/         # Слой интерфейсов
    ├── http/          # HTTP handlers
    ├── grpc/          # gRPC handlers (опционально)
    └── events/        # Event handlers
```

## 🚀 Быстрый старт

### Требования

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist
```

1. Скопируйте `.env.example` в `.env`:
```bash
cp .env.example .env
```

1. Установите зависимости:
```bash
go mod download
```

1. Запустите приложение:
```bash
go run cmd/server/main.go
```

## 🧩 Модули

### Core Modules (Основные)

#### 1. **Auth Module** 🔐
- OAuth 2.0 / OpenID Connect
- JWT токены
- Управление сессиями и ролями

#### 2. **Users Module** 👥
- Профили пользователей
- Синхронизация с 1С
- Управление подразделениями

#### 3. **Documents Module** 📄
- Создание и редактирование документов
- Версионирование
- Поиск по документам

#### 4. **Workflow Module** 🔄
- Маршрутизация документов
- Согласование и одобрение
- Workflow engine

### Business Modules (Бизнес)

#### 5. **Schedule Module** 📅
- Составление расписания
- Управление аудиториями
- Оптимизация

#### 6. **Reporting Module** 📊
- Генерация отчетов
- Аналитика
- Экспорт данных

#### 7. **Tasks Module** ✅
- Управление задачами
- Напоминания
- Трекинг выполнения

### Supporting Modules (Поддерживающие)

#### 8. **Notifications Module** 📧
- Email уведомления
- Push уведомления
- Шаблоны уведомлений

#### 9. **Files Module** 📁
- Хранение файлов
- Конвертация документов
- Превью

#### 10. **Integration Module** 🔗
- Синхронизация с 1С
- Внешние API
- Data mapping

## 🛠️ Разработка

### Код-стайл

Проект следует стандартным Go conventions:

```bash
# Форматирование кода
go fmt ./...

# Проверка кода
golangci-lint run
```

### Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Запуск специфичных тестов
go test ./internal/modules/auth/...
```

### Миграции БД

```bash
# TODO: Добавить инструкции по миграциям
```

## 📚 Документация

- [Модульная архитектура](docs/architecture/modular-architecture.md)
- [Миграция к микросервисам](docs/architecture/microservices-migration-guide.md)
- [Clean Code паттерны](docs/development/clean-code-patterns.md)
- [API документация](docs/api/api-documentation.md)

## 🤝 Contributing

См. [Development Guide](docs/development/development-guide.md) и [Pull Request Guide](docs/development/pull-request-guide.md)

## 📝 License

MIT License - см. [LICENSE](LICENSE) файл
