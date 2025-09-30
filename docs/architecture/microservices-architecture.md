# 🏗️ Микросервисная архитектура

## 📋 Обзор архитектуры

Система построена на принципах микросервисной архитектуры для обеспечения масштабируемости, отказоустойчивости и независимости развития модулей.

## 🧩 Структура микросервисов

### Core Services (Основные сервисы)

#### 1. **auth-service** 🔐
**Назначение**: Аутентификация и авторизация
- OAuth 2.0 / OpenID Connect
- JWT токены
- Управление сессиями
- **Технологии**: Go, Redis, PostgreSQL
- **Порт**: 8001

#### 2. **user-service** 👥
**Назначение**: Управление пользователями
- CRUD операции с пользователями
- Профили и роли
- Интеграция с 1С
- **Технологии**: Go, PostgreSQL
- **Порт**: 8002

#### 3. **document-service** 📄
**Назначение**: Управление документами
- Хранение метаданных документов
- Версионирование
- Связи между документами
- **Технологии**: Go, PostgreSQL, MinIO/S3
- **Порт**: 8003

#### 4. **workflow-service** 🔄
**Назначение**: Workflow и бизнес-процессы
- Маршруты согласования
- Статусы документов
- Автоматизация процессов
- **Технологии**: Go, PostgreSQL, Redis
- **Порт**: 8004

### Business Services (Бизнес-сервисы)

#### 5. **schedule-service** 📅
**Назначение**: Управление расписанием
- Расписание занятий
- Календарь событий
- Планирование ресурсов
- **Технологии**: Go, PostgreSQL
- **Порт**: 8005

#### 6. **reporting-service** 📊
**Назначение**: Отчетность
- Генерация отчетов
- Аналитика
- Экспорт данных
- **Технологии**: Go, PostgreSQL, ClickHouse
- **Порт**: 8006

#### 7. **task-service** ✅
**Назначение**: Управление задачами
- Создание и назначение задач
- Отслеживание выполнения
- Напоминания
- **Технологии**: Go, PostgreSQL, Redis
- **Порт**: 8007

### Support Services (Поддерживающие сервисы)

#### 8. **notification-service** 📧
**Назначение**: Уведомления
- Email уведомления
- WebSocket для real-time
- Шаблоны сообщений
- **Технологии**: Go, Redis, SMTP
- **Порт**: 8008

#### 9. **file-service** 📁
**Назначение**: Файловое хранилище
- Загрузка/скачивание файлов
- Предварительный просмотр
- Конвертация форматов
- **Технологии**: Go, MinIO/S3, ImageMagick
- **Порт**: 8009

#### 10. **integration-service** 🔗
**Назначение**: Интеграции
- Синхронизация с 1С
- API для внешних систем
- Адаптеры протоколов
- **Технологии**: Go, PostgreSQL, Message Queue
- **Порт**: 8010

## 🌐 Frontend Applications

### 1. **admin-dashboard** 👨‍💼
**Назначение**: Административная панель
- **Технологии**: Next.js 14, MUI, TypeScript
- **Порт**: 3001

### 2. **user-portal** 👤
**Назначение**: Портал пользователей
- **Технологии**: Next.js 14, MUI, TypeScript
- **Порт**: 3002

## 🔧 Infrastructure Services

### Message Broker
- **Apache Kafka** - для асинхронной связи между сервисами
- **Redis Pub/Sub** - для real-time уведомлений

### Databases
- **PostgreSQL** - основная реляционная БД
- **Redis** - кэширование и сессии
- **ClickHouse** - аналитические данные (опционально)

### Service Discovery & Configuration
- **Consul** - service discovery
- **Vault** - управление секретами

### Monitoring & Logging
- **Prometheus + Grafana** - мониторинг
- **ELK Stack** - логирование и анализ

## 🔄 Межсервисное взаимодействие

### Синхронное взаимодействие
- **HTTP/REST** - для CRUD операций
- **gRPC** - для высокопроизводительных внутренних вызовов

### Асинхронное взаимодействие
- **Kafka** - для event-driven архитектуры
- **Redis Pub/Sub** - для real-time событий

## 📊 Схема развертывания

```
┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   API Gateway   │
│    (Nginx)      │────│   (Kong/Envoy)  │
└─────────────────┘    └─────────────────┘
           │                      │
    ┌──────┴──────────────────────┴──────┐
    │           Kubernetes Cluster        │
    │  ┌─────────┐ ┌─────────┐ ┌───────┐ │
    │  │Auth Svc │ │User Svc │ │Doc Svc│ │
    │  └─────────┘ └─────────┘ └───────┘ │
    │  ┌─────────┐ ┌─────────┐ ┌───────┐ │
    │  │Flow Svc │ │Sched Svc│ │Rep Svc│ │
    │  └─────────┘ └─────────┘ └───────┘ │
    └─────────────────────────────────────┘
           │              │
    ┌──────┴─────┐ ┌──────┴─────┐
    │PostgreSQL  │ │   Redis    │
    │  Cluster   │ │  Cluster   │
    └────────────┘ └────────────┘
```

## 🔒 Безопасность

### Аутентификация
- OAuth 2.0 с внешними провайдерами
- JWT токены с коротким TTL
- Refresh token rotation

### Авторизация
- RBAC (Role-Based Access Control)
- Attribute-Based Access Control для документов
- API Gateway для centralized authorization

### Сетевая безопасность
- mTLS между сервисами
- Network policies в Kubernetes
- WAF на уровне Load Balancer

## 📈 Масштабирование

### Горизонтальное масштабирование
- Stateless сервисы
- Auto-scaling на основе метрик (до 1000 одновременных пользователей)
- Load balancing между репликами
- CDN для статических ресурсов
- Redis кэширование для улучшения производительности

### Вертикальное масштабирование
- Resource limits и requests в Kubernetes
- Мониторинг использования ресурсов
- Оптимизация производительности БД

## 🔗 Схемы интеграции

### 📊 Интеграция с 1С
```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   1С: Предприятие   │◄──►│ integration-svc  │◄──►│   user-service  │
│                     │    │                  │    │                 │
│ • Сотрудники        │    │ • REST API       │    │ • Синхронизация │
│ • Студенты          │    │ • Планировщик    │    │ • Маппинг ролей │
│ • Организация       │    │ • Валидация      │    │ • Уведомления   │
│ • Финансы           │    │ • Логирование    │    │ • Конфликты     │
└─────────────────────┘    └──────────────────┘    └─────────────────┘
```

**Протоколы интеграции:**
- **REST API** - основной канал обмена данными
- **WebService (SOAP)** - для legacy систем 1С
- **File Exchange** - CSV/XML файлы через SFTP
- **Real-time sync** - через Kafka events

### 📧 Email Integration
```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   SMTP Server       │◄──►│ notification-svc │◄──►│  All Services   │
│                     │    │                  │    │                 │
│ • Корп. почта       │    │ • Шаблоны        │    │ • Workflow      │
│ • Gmail/Outlook     │    │ • Очереди        │    │ • Alerts        │
│ • Mail.ru           │    │ • Retry logic    │    │ • Reports       │
└─────────────────────┘    └──────────────────┘    └─────────────────┘
```

### 🔐 Электронная подпись
```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Crypto Pro CSP     │◄──►│ signature-svc    │◄──►│ document-svc    │
│                     │    │                  │    │                 │
│ • УЦ Сертификаты    │    │ • PKCS#7         │    │ • PDF signing   │
│ • Ключевые носители │    │ • Валидация      │    │ • XAdES         │
│ • ФЗ-63 соответствие│    │ • Timestamps     │    │ • Архив подписей│
└─────────────────────┘    └──────────────────┘    └─────────────────┘
```

### 💾 Файловые хранилища
```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Object Storage    │◄──►│   file-service   │◄──►│  Frontend Apps  │
│                     │    │                  │    │                 │
│ • MinIO/S3          │    │ • Versioning     │    │ • Upload UI     │
│ • Yandex Cloud      │    │ • Compression    │    │ • Preview       │
│ • Local FS          │    │ • Virus scan     │    │ • Download      │
│ • Network drives    │    │ • Thumbnails     │    │ • Search        │
└─────────────────────┘    └──────────────────┘    └─────────────────┘
```

### 📱 External APIs
```
┌─────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   External APIs     │◄──►│ integration-svc  │◄──►│  Core Services  │
│                     │    │                  │    │                 │
│ • Госуслуги ЕСИА    │    │ • API Gateway    │    │ • Auth service  │
│ • ФИС ФРДО          │    │ • Rate limiting  │    │ • User service  │
│ • Telegram Bot API  │    │ • Circuit break  │    │ • Notification  │
│ • SMS providers     │    │ • Monitoring     │    │ • Reporting     │
└─────────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 CI/CD Pipeline

1. **Code Commit** → Git webhook
2. **Build** → Docker images
3. **Test** → Unit, Integration, E2E
4. **Security Scan** → Vulnerability check
5. **Deploy** → Staging environment
6. **Approval** → Manual/Automated
7. **Production Deploy** → Blue-Green deployment

## 🔄 Data Flow Architecture

### Основной поток данных
```
User Request → API Gateway → Auth Service → Business Service → Database
     ↓              ↓            ↓              ↓              ↓
Response ← Load Balancer ← Cache Layer ← Service Mesh ← Storage Layer
```

### Event-Driven Architecture
```
Service A → Kafka Topic → Service B → Event Store → Analytics
    ↓           ↓           ↓           ↓            ↓
 Events → Message Queue → Consumers → Aggregation → Reports
```