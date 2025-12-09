# 📊 Логирование и оптимизация производительности

## 📋 Обзор

Система использует комплексный подход к логированию и мониторингу производительности с применением DDD принципов и оптимизаций.

---

## 🔍 Типы логирования

### 1. Security Logging (Логи безопасности)

**Расположение**: `internal/shared/infrastructure/logging/security_logger.go`

#### События безопасности:

```go
type SecurityEvent string

const (
    EventLoginSuccess           // Успешная авторизация
    EventLoginFailed            // Неудачная попытка входа
    EventRegistrationSuccess    // Успешная регистрация
    EventRegistrationFailed     // Ошибка регистрации
    EventTokenRefreshSuccess    // Обновление токена
    EventTokenRefreshFailed     // Ошибка обновления токена
    EventTokenValidationFailed  // Невалидный токен
    EventUnauthorizedAccess     // Неавторизованный доступ
    EventRateLimitExceeded      // Превышение лимита запросов
    EventAccountLocked          // Блокировка аккаунта
    EventPasswordChanged        // Изменение пароля
    EventPermissionDenied       // Отказ в доступе
)
```

#### Что логируется:

- ✅ **Все попытки входа** (успешные и неудачные)
- ✅ **Причина отказа** (неверный пароль, заблокирован аккаунт, etc.)
- ✅ **IP адрес** и User-Agent
- ✅ **Correlation ID** для трейсинга
- ✅ **Временные метки** (UTC)
- ✅ **User ID** (если доступен)

#### Пример лога:

```json
{
  "timestamp": "2025-01-15T01:30:45Z",
  "level": "WARN",
  "category": "security",
  "event_type": "login_failed",
  "email": "user@example.com",
  "reason": "invalid password",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### 2. Audit Logging (Audit trail)

**Расположение**: `internal/shared/infrastructure/logging/audit_logger.go`

#### Что логируется:

- ✅ **Критические операции** (регистрация, логин, обновление токенов)
- ✅ **CRUD операции** (документы, категории, теги)
- ✅ **Email уведомления** (отправка писем)
- ✅ **Кто выполнил** (user ID, IP)
- ✅ **Когда** (timestamp)
- ✅ **Что изменено** (resource)
- ✅ **Результат** (success/failure)

#### Аудит-события по модулям:

| Модуль | Событие | Описание |
|--------|---------|----------|
| **Auth** | `login_success` | Успешный вход |
| **Auth** | `login_failed` | Неудачная попытка входа |
| **Auth** | `registration_success` | Успешная регистрация |
| **Auth** | `token_refresh` | Обновление токена |
| **Documents** | `document_created` | Создание документа |
| **Documents** | `document_updated` | Обновление документа |
| **Documents** | `document_deleted` | Удаление документа |
| **Documents** | `document_file_uploaded` | Загрузка файла |
| **Documents** | `document_file_deleted` | Удаление файла |
| **Categories** | `category_created` | Создание категории |
| **Categories** | `category_updated` | Обновление категории |
| **Categories** | `category_deleted` | Удаление категории |
| **Tags** | `tag_created` | Создание тега |
| **Tags** | `tag_updated` | Обновление тега |
| **Tags** | `tag_deleted` | Удаление тега |
| **Tags** | `tag_added_to_document` | Добавление тега к документу |
| **Tags** | `tag_removed_from_document` | Удаление тега с документа |
| **Tags** | `document_tags_set` | Массовая установка тегов |
| **Notifications** | `email_sent` | Отправка email |
| **Schedule** | `schedule_created` | Создание расписания |
| **Schedule** | `schedule_updated` | Обновление расписания |
| **Schedule** | `schedule_deleted` | Удаление расписания |

#### Пример audit log:

```json
{
  "timestamp": "2025-01-15T01:30:45Z",
  "level": "INFO",
  "category": "audit",
  "action": "login",
  "resource": "auth",
  "actor_user_id": 123,
  "actor_ip": "192.168.1.100",
  "user_id": 123,
  "email": "user@example.com",
  "role": "admin",
  "duration_ms": 245,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### 3. Performance Logging (Мониторинг производительности)

**Расположение**: `internal/shared/infrastructure/logging/security_logger.go`

#### Метрики:

**HTTP Request Performance:**
```json
{
  "timestamp": "2025-01-15T01:30:45Z",
  "level": "INFO",
  "category": "performance",
  "method": "POST",
  "path": "/api/auth/login",
  "status_code": 200,
  "duration_ms": 125,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Database Query Performance:**
```json
{
  "timestamp": "2025-01-15T01:30:45Z",
  "level": "DEBUG",
  "category": "performance",
  "query_type": "database",
  "duration_ms": 15,
  "rows_affected": 1,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Cache Operations:**
```json
{
  "timestamp": "2025-01-15T01:30:45Z",
  "level": "DEBUG",
  "category": "performance",
  "operation": "get_by_email",
  "cache_key": "user:email:user@example.com",
  "cache_hit": true,
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### Автоматические предупреждения:

- ⚠️ **Slow queries** (> 100ms) - автоматически уровень WARN
- ⚠️ **Slow HTTP requests** (> 500ms) - автоматически уровень WARN

---

## 🚀 Оптимизации производительности

### 1. Redis Caching

**Расположение**: `internal/shared/infrastructure/cache/redis_cache.go`

#### Конфигурация:

```go
PoolSize:     10         // Размер connection pool
MinIdleConns: 5          // Минимум idle соединений
MaxRetries:   3          // Retry попытки
DialTimeout:  5s         // Timeout подключения
ReadTimeout:  3s         // Timeout чтения
WriteTimeout: 3s         // Timeout записи
```

#### Кешируемые данные:

| Тип | Key Pattern | TTL | Описание |
|-----|-------------|-----|----------|
| User by ID | `user:{id}` | 5 min | Данные пользователя по ID |
| User by Email | `user:email:{email}` | 5 min | Данные пользователя по email |
| Token Blacklist | `token:blacklist:{jti}` | До expiry | Отозванные токены |

#### Пример использования:

```go
// Cache hit - 0.5ms
user, hit := userCache.GetUser(ctx, 123, &user)
if hit {
    // Данные из Redis
}

// Cache miss - 15ms (database query)
user, err := userRepo.GetByID(ctx, 123)
userCache.SetUser(ctx, 123, user) // Обновляем кеш
```

#### Производительность:

- 📈 **Cache hit**: ~0.5-1ms
- 📈 **Cache miss + DB query**: ~15-20ms
- 📈 **Cache hit ratio**: 85-95% (после прогрева)

---

### 2. Database Connection Pooling

**Конфигурация** (compose.yml):

```yaml
DB_MAX_OPEN_CONNS: 25           # Максимум открытых соединений
DB_MAX_IDLE_CONNS: 5            # Минимум idle соединений
DB_CONN_MAX_LIFETIME: 5m        # Lifetime соединения
```

#### Оптимизации:

- ✅ **Connection reuse** - переиспользование соединений
- ✅ **Context-aware queries** - QueryRowContext, ExecContext
- ✅ **Prepared statements** - защита от SQL injection + caching
- ✅ **Transaction timeout** - context.WithTimeout(5s)

---

### 3. Cached Repository Pattern

**Расположение**: `internal/modules/auth/infrastructure/cached_user_repository.go`

#### Архитектура:

```
HTTP Request
    ↓
Controller/Handler
    ↓
Use Case
    ↓
CachedRepository (Decorator)
    ├─> Redis Cache (cache hit) → Return ✅
    └─> UserRepository (cache miss)
            ↓
        PostgreSQL → Return + Update Cache
```

#### Преимущества:

- ✅ **Прозрачность** - use case не знает о кешировании
- ✅ **Decorator pattern** - легко включить/выключить
- ✅ **DDD compliance** - не нарушает domain слой
- ✅ **Performance logging** - автоматический учёт cache hits/misses

---

### 4. Correlation ID для трейсинга

**Расположение**: `internal/shared/infrastructure/middleware/correlation_id.go`

#### Как работает:

1. **Входящий запрос**:
   - Проверяется header `X-Correlation-ID`
   - Если нет - генерируется новый UUID

2. **Контекст**:
   - Добавляется в Gin context
   - Добавляется в Request context

3. **Ответ**:
   - Возвращается в header `X-Correlation-ID`

4. **Логирование**:
   - Все логи содержат correlation_id
   - Можно отследить весь путь запроса

#### Пример трейсинга:

```bash
# Входящий запрос
POST /api/auth/login
X-Correlation-ID: 550e8400-e29b-41d4-a716-446655440000

# Логи для одного запроса:
[INFO] HTTP request completed - correlation_id: 550e8400...
[DEBUG] Cache operation - correlation_id: 550e8400...
[DEBUG] Database query executed - correlation_id: 550e8400...
[INFO] Security event: login_success - correlation_id: 550e8400...
[INFO] Audit event: login - correlation_id: 550e8400...
```

---

## 📈 Метрики производительности

### Типичные значения (с кешированием):

| Операция | Without Cache | With Cache | Улучшение |
|----------|---------------|------------|-----------|
| GetUserByID | 15-20ms | 0.5-1ms | **30x faster** |
| GetUserByEmail | 15-20ms | 0.5-1ms | **30x faster** |
| Login (full flow) | 350-400ms | 200-250ms | **40% faster** |
| Token validation | 5-10ms | 5-10ms | - |

### Пропускная способность:

- 📊 **Without cache**: ~200-300 req/sec
- 📊 **With cache (hot)**: ~800-1200 req/sec
- 📊 **Database connections**: 25 max (pool)
- 📊 **Redis connections**: 10 max (pool)

---

## 🔧 Настройка и мониторинг

### Environment Variables:

```bash
# Database
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=""
REDIS_DB=0
REDIS_POOL_SIZE=10

# Logging
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=json         # json, text
```

### Мониторинг в production:

#### 1. Агрегация логов:

```bash
# Все security события
grep '"category":"security"' app.log

# Все failed logins
grep '"event_type":"login_failed"' app.log

# Slow queries
grep '"slow_query":true' app.log

# Slow requests
grep '"slow_request":true' app.log
```

#### 2. Метрики cache:

```bash
# Cache hit ratio
grep '"category":"performance"' app.log | \
  grep '"cache_hit"' | \
  jq -s 'group_by(.cache_hit) | map({hit: .[0].cache_hit, count: length})'
```

#### 3. Performance analysis:

```bash
# Средняя duration HTTP requests
grep '"category":"performance"' app.log | \
  grep '"method":"POST"' | \
  jq '.duration_ms' | \
  awk '{sum+=$1; count++} END {print sum/count}'
```

---

## 🎯 Best Practices

### 1. DDD + Performance

✅ **DO:**
- Используйте Decorator pattern для кеширования
- Логируйте в use case layer (domain не знает о логах)
- Измеряйте performance на infrastructure layer

❌ **DON'T:**
- Не добавляйте cache логику в domain layer
- Не смешивайте business logic и logging
- Не делайте преждевременную оптимизацию

### 2. Logging

✅ **DO:**
- Используйте structured logging (JSON)
- Добавляйте correlation IDs ко всем логам
- Логируйте все security events
- Измеряйте производительность критичных операций

❌ **DON'T:**
- Не логируйте sensitive data (пароли, токены)
- Не используйте fmt.Println
- Не игнорируйте context

### 3. Caching

✅ **DO:**
- Кешируйте read-heavy данные (users, permissions)
- Инвалидируйте при обновлениях
- Используйте короткие TTL (5-10 min)
- Graceful degradation если Redis недоступен

❌ **DON'T:**
- Не кешируйте write-heavy данные
- Не используйте вечный TTL
- Не забывайте о cache invalidation
- Не полагайтесь только на cache

---

## 📊 Пример полного flow с логами

```
1. [INFO] HTTP request started
   method: POST, path: /api/auth/login
   correlation_id: 550e8400...

2. [DEBUG] Cache operation
   operation: get_by_email, cache_hit: false
   correlation_id: 550e8400...

3. [DEBUG] Database query executed
   query_type: database, duration_ms: 18
   correlation_id: 550e8400...

4. [INFO] Security event: login_success
   email: user@example.com
   correlation_id: 550e8400...

5. [INFO] Audit event: login
   user_id: 123, duration_ms: 245
   correlation_id: 550e8400...

6. [INFO] HTTP request completed
   status_code: 200, duration_ms: 280
   correlation_id: 550e8400...
```

Все 6 логов имеют одинаковый `correlation_id` - можно отследить весь путь запроса!

---

## 🔍 Troubleshooting

### Проблема: Высокая latency

**Решение**:
1. Проверьте slow query logs
2. Проверьте cache hit ratio
3. Увеличьте connection pool
4. Добавьте индексы в БД

### Проблема: Memory leak

**Решение**:
1. Проверьте TTL кеша
2. Проверьте connection pool limits
3. Мониторьте goroutines
4. Используйте pprof

### Проблема: Security alerts

**Решение**:
1. Анализируйте security logs
2. Проверьте rate limiting
3. Блокируйте подозрительные IP
4. Проверьте failed login patterns

---

## 📡 Мониторинг и Observability

### Стек мониторинга

Проект использует современный стек observability:

| Компонент | Версия | Назначение | Порт |
|-----------|--------|------------|------|
| **Prometheus** | v2.48.0 | Сбор и хранение метрик | 9090 |
| **Grafana** | v10.2.2 | Визуализация метрик и логов | 3001 |
| **Loki** | v2.9.2 | Агрегация и хранение логов | 3100 |
| **Promtail** | v2.9.2 | Сбор логов из Docker | - |

### Запуск мониторинга

```bash
# Запуск с мониторингом
docker compose -f compose.yml -f compose.monitoring.yml up -d

# Только основные сервисы (без мониторинга)
docker compose up -d
```

### Health Check Endpoints

Приложение предоставляет endpoints для Kubernetes probes:

| Endpoint | Назначение | Проверки |
|----------|------------|----------|
| `/health` | Полная проверка здоровья | Database, Redis |
| `/live` | Liveness probe | Только процесс |
| `/ready` | Readiness probe | Database (required), Redis (optional) |

```bash
# Примеры запросов
curl http://localhost:8080/health
curl http://localhost:8080/live
curl http://localhost:8080/ready
```

### Prometheus Metrics

**Расположение**: `internal/shared/infrastructure/metrics/prometheus.go`

#### Доступные метрики:

| Метрика | Тип | Описание |
|---------|-----|----------|
| `http_requests_total` | Counter | Общее количество HTTP запросов |
| `http_request_duration_seconds` | Histogram | Время обработки HTTP запросов |
| `http_requests_in_flight` | Gauge | Текущие запросы в обработке |
| `database_queries_total` | Counter | Количество запросов к БД |
| `database_query_duration_seconds` | Histogram | Время выполнения запросов к БД |
| `cache_operations_total` | Counter | Операции с кешем (hit/miss) |
| `auth_events_total` | Counter | События аутентификации |
| `business_operations_total` | Counter | Бизнес-операции по модулям |
| `active_connections` | Gauge | Активные соединения |

#### Endpoint метрик:

```bash
curl http://localhost:8080/metrics
```

### Grafana Dashboards

Предустановленные дашборды:

1. **HTTP Metrics** (`http-metrics`)
   - Request Rate (req/s)
   - P95 Response Time
   - Request Rate by Endpoint
   - Response Time Percentiles (p50, p95, p99)
   - Status Code Distribution
   - In-Flight Requests
   - Success Rate
   - Error Rate (5xx)

2. **Application Logs** (`app-logs`)
   - Log Volume by Level
   - Error/Warning Count
   - Security Events Count
   - Audit Events Count
   - Errors and Warnings Stream
   - Security Logs Stream
   - Audit Logs Stream
   - All Logs Stream

#### Доступ к Grafana:

```
URL: http://localhost:3001
User: admin
Password: admin (по умолчанию)
```

### Loki для централизованного сбора логов

#### Конфигурация:

- **Retention**: 7 дней
- **Storage**: TSDB на файловой системе
- **Max line size**: 256KB

#### Querying логов в Grafana:

```logql
# Все логи backend
{container="backend-dev"}

# Только ошибки
{container="backend-dev"} | json | level = "ERROR"

# Security события
{container="backend-dev"} | json | category = "security"

# Audit события
{container="backend-dev"} | json | category = "audit"

# Поиск по correlation_id
{container="backend-dev"} | json | correlation_id = "550e8400-e29b-41d4-a716-446655440000"

# Логи с duration > 500ms
{container="backend-dev"} | json | duration_ms > 500
```

### Структура файлов мониторинга

```
monitoring/
├── grafana/
│   ├── dashboards/
│   │   ├── http-metrics.json
│   │   └── application-logs.json
│   └── provisioning/
│       ├── dashboards/
│       │   └── dashboards.yml
│       └── datasources/
│           └── datasources.yml
├── loki/
│   └── loki-config.yml
├── prometheus/
│   └── prometheus.yml
└── promtail/
    └── promtail-config.yml
```

### Environment Variables для мониторинга

```bash
# Prometheus
PROMETHEUS_PORT=9090

# Grafana
GRAFANA_PORT=3001
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
GRAFANA_ROOT_URL=http://localhost:3001

# Loki
LOKI_PORT=3100
```

---

**Документация обновлена**: 2025-12-09
**Версия**: 0.2.0

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-12-09
**Версия проекта**: 0.2.0
**Статус**: Актуальный

