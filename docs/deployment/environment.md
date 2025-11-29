# Environment Configuration

## Обзор

Управление конфигурацией для различных окружений с использованием переменных среды и Docker Compose.

## Переменные окружения

### Server

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `SERVER_PORT` | Порт HTTP сервера | `8080` | Нет |
| `SERVER_READ_TIMEOUT` | Таймаут чтения запроса | `10s` | Нет |
| `SERVER_WRITE_TIMEOUT` | Таймаут записи ответа | `10s` | Нет |
| `SERVER_IDLE_TIMEOUT` | Таймаут простоя соединения | `120s` | Нет |
| `ENVIRONMENT` | Окружение (development/staging/production) | `development` | Нет |
| `VERSION` | Версия приложения | `0.1.0` | Нет |

### Database (PostgreSQL)

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `DB_HOST` | Хост PostgreSQL | - | **Да** |
| `DB_PORT` | Порт PostgreSQL | `5432` | Нет |
| `DB_USER` | Пользователь БД | - | **Да** |
| `DB_PASSWORD` | Пароль БД | - | **Да** |
| `DB_NAME` | Имя базы данных | - | **Да** |
| `DB_SSL_MODE` | SSL режим | `disable` | Нет |
| `DB_MAX_OPEN_CONNS` | Макс. открытых соединений | `25` | Нет |
| `DB_MAX_IDLE_CONNS` | Макс. idle соединений | `5` | Нет |
| `DB_CONN_MAX_LIFETIME` | Макс. время жизни соединения | `5m` | Нет |

### Redis

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `REDIS_HOST` | Хост Redis | - | **Да** |
| `REDIS_PORT` | Порт Redis | `6379` | Нет |
| `REDIS_PASSWORD` | Пароль Redis | ` ` (пустой) | Нет |
| `REDIS_DB` | Номер базы Redis | `0` | Нет |
| `REDIS_POOL_SIZE` | Размер пула соединений | `10` | Нет |

### Authentication (JWT)

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `JWT_ACCESS_SECRET` | Секрет для access токенов | - | **Да** |
| `JWT_REFRESH_SECRET` | Секрет для refresh токенов | - | **Да** |
| `JWT_ACCESS_TTL` | Время жизни access токена | `15m` | Нет |
| `JWT_REFRESH_TTL` | Время жизни refresh токена | `168h` (7 дней) | Нет |

### CORS

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `CORS_ALLOWED_ORIGINS` | Разрешённые origins | `http://localhost:3000` | Нет |
| `CORS_ALLOWED_METHODS` | Разрешённые методы | `GET,POST,PUT,DELETE,OPTIONS` | Нет |
| `CORS_ALLOWED_HEADERS` | Разрешённые заголовки | `Content-Type,Authorization` | Нет |

### Logging

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `LOG_LEVEL` | Уровень логирования | `info` | Нет |
| `LOG_FORMAT` | Формат логов (json/text) | `json` | Нет |

### S3/MinIO Storage

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `S3_ENDPOINT` | Endpoint S3/MinIO | - | Для документов |
| `S3_ACCESS_KEY_ID` | Access Key | - | Для документов |
| `S3_SECRET_ACCESS_KEY` | Secret Key | - | Для документов |
| `S3_BUCKET_NAME` | Имя bucket | - | Для документов |
| `S3_REGION` | Регион | `us-east-1` | Нет |
| `S3_USE_SSL` | Использовать SSL | `true` | Нет |

### Composio (Gmail Integration)

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `COMPOSIO_API_KEY` | API ключ Composio | - | Для email |
| `COMPOSIO_ENTITY_ID` | Entity ID Composio | - | Для email |
| `COMPOSIO_MCP_CONFIG_ID` | MCP Config ID | - | Нет |

### Frontend (Next.js)

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `NEXT_PUBLIC_API_URL` | URL бэкенда | `http://localhost:8080` | Нет |
| `NODE_ENV` | Окружение Node.js | `development` | Нет |

### Monitoring

| Переменная | Описание | Default | Required |
|------------|----------|---------|----------|
| `BACKEND_PORT` | Порт backend на хосте | `8080` | Нет |
| `FRONTEND_PORT` | Порт frontend на хосте | `3000` | Нет |
| `PROMETHEUS_PORT` | Порт Prometheus на хосте | `9090` | Нет |
| `GRAFANA_PORT` | Порт Grafana на хосте | `3001` | Нет |
| `LOKI_PORT` | Порт Loki на хосте | `3100` | Нет |
| `GRAFANA_ADMIN_USER` | Админ Grafana | `admin` | Нет |
| `GRAFANA_ADMIN_PASSWORD` | Пароль админа Grafana | `admin` | Нет |

---

## Примеры конфигурации

### Development (.env)

```bash
# .env

# Server
ENVIRONMENT=development
VERSION=0.1.0

# Database
POSTGRES_DB=inf_sys_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password

# Redis (optional password)
REDIS_PASSWORD=

# JWT (обязательно!)
JWT_ACCESS_SECRET=your_32_char_access_secret_key_here
JWT_REFRESH_SECRET=your_32_char_refresh_secret_key_here

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Composio (optional)
# COMPOSIO_API_KEY=your_composio_api_key
# COMPOSIO_ENTITY_ID=your_entity_id

# S3/MinIO (optional, for documents)
# S3_ENDPOINT=localhost:9000
# S3_ACCESS_KEY_ID=minioadmin
# S3_SECRET_ACCESS_KEY=minioadmin
# S3_BUCKET_NAME=documents
```

### compose.override.yml

```yaml
# compose.override.yml (НЕ коммитить!)
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: your_secure_password
    ports:
      - "5432:5432"

  redis:
    environment:
      REDIS_PASSWORD: optional_redis_password
    ports:
      - "6379:6379"
```

---

## Генерация секретов

### JWT секреты:
```bash
# Генерация безопасных секретов
openssl rand -base64 32

# Или через /dev/urandom
head -c 32 /dev/urandom | base64
```

### Пример для .env:
```bash
JWT_ACCESS_SECRET=$(openssl rand -base64 32)
JWT_REFRESH_SECRET=$(openssl rand -base64 32)
```

---

## Docker Compose конфигурация

### Основные сервисы (compose.yml):

```yaml
services:
  backend:
    environment:
      # Server
      SERVER_PORT: 8080

      # Database
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${POSTGRES_USER:-postgres}
      DB_PASSWORD: ${POSTGRES_PASSWORD}
      DB_NAME: ${POSTGRES_DB:-inf_sys_db}

      # Redis
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD:-}

      # JWT
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}

      # Composio
      COMPOSIO_API_KEY: ${COMPOSIO_API_KEY:-}
      COMPOSIO_ENTITY_ID: ${COMPOSIO_ENTITY_ID:-}

      # CORS
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000}

      # Logging
      LOG_LEVEL: ${LOG_LEVEL:-info}
      LOG_FORMAT: ${LOG_FORMAT:-json}
```

### Мониторинг (compose.monitoring.yml):

```yaml
services:
  grafana:
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER:-admin}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
    ports:
      - "${GRAFANA_PORT:-3001}:3000"

  prometheus:
    ports:
      - "${PROMETHEUS_PORT:-9090}:9090"

  loki:
    ports:
      - "${LOKI_PORT:-3100}:3100"
```

---

## Валидация конфигурации

### Обязательные переменные:

При запуске backend требуются:
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`

### Опциональные модули:

- **Documents**: Требуются `S3_*` переменные
- **Email**: Требуются `COMPOSIO_*` переменные

### Проверка перед запуском:
```bash
# Проверка обязательных переменных
if [ -z "$JWT_ACCESS_SECRET" ]; then
    echo "ERROR: JWT_ACCESS_SECRET not set"
    exit 1
fi
```

---

## Безопасность

### Рекомендации:

1. **Никогда не коммитьте** файлы с секретами:
   - `.env`
   - `compose.override.yml`
   - Любые файлы с паролями

2. **Добавьте в .gitignore**:
   ```
   .env
   .env.*
   compose.override.yml
   !.env.example
   ```

3. **Используйте сильные секреты**:
   - JWT секреты: минимум 32 символа
   - Пароли БД: минимум 16 символов

4. **В production**:
   - Используйте Docker secrets
   - Не открывайте порты БД наружу
   - Установите `LOG_LEVEL=warn` или `info`

### Docker secrets (production):
```yaml
secrets:
  db_password:
    file: ./secrets/db_password.txt
  jwt_access_secret:
    file: ./secrets/jwt_access_secret.txt
  jwt_refresh_secret:
    file: ./secrets/jwt_refresh_secret.txt

services:
  backend:
    secrets:
      - db_password
      - jwt_access_secret
      - jwt_refresh_secret
    environment:
      DB_PASSWORD_FILE: /run/secrets/db_password
```

---

## Окружения

### Development
```bash
ENVIRONMENT=development
LOG_LEVEL=debug
DEBUG=true
```

### Staging
```bash
ENVIRONMENT=staging
LOG_LEVEL=info
DEBUG=false
```

### Production
```bash
ENVIRONMENT=production
LOG_LEVEL=warn
DEBUG=false
```

---

## Quick Start

```bash
# 1. Создать файл с переменными
cp .env.example .env

# 2. Установить обязательные переменные
nano .env

# 3. Создать compose.override.yml
cat > compose.override.yml << EOF
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: your_password
EOF

# 4. Запустить
docker compose up -d

# 5. Проверить
curl http://localhost:8080/health
```

---

**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный
