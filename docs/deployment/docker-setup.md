# Docker Setup

## Обзор контейнеризации

Полная настройка Docker и Docker Compose для локальной разработки и развертывания модульной монолитной архитектуры.

### Архитектура проекта

Проект построен по принципу **модульного монолита**:
- **backend** - единый Go бэкенд с модулями (auth, documents, schedule, notifications)
- **frontend** - Next.js веб-интерфейс
- **postgres** - PostgreSQL 17 для хранения данных
- **redis** - Redis 7 для кэширования и сессий

Дополнительно доступен **мониторинг стек** (Prometheus, Grafana, Loki, Promtail).

---

## Quick Start

### Запуск системы:
```bash
# Клонирование проекта
git clone https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist

# Создание файла с секретами
cp compose.override.example.yml compose.override.yml
# Отредактируйте compose.override.yml и установите POSTGRES_PASSWORD

# Создание .env файла
cp .env.example .env
# Установите JWT_ACCESS_SECRET и JWT_REFRESH_SECRET

# Запуск всех сервисов
docker compose up -d

# Проверка статуса
docker compose ps

# Просмотр логов
docker compose logs -f
```

### Запуск с мониторингом:
```bash
docker compose -f compose.yml -f compose.monitoring.yml up -d
```

---

## Docker Compose конфигурация

### Основной compose.yml:
```yaml
services:
  postgres:
    image: postgres:17-alpine
    container_name: postgres-dev
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-inf_sys_db}
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?Please set POSTGRES_PASSWORD in compose.override.yml}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-postgres} -d ${POSTGRES_DB:-inf_sys_db}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: redis-dev
    command: >
      sh -c '
        if [ -n "$$REDIS_PASSWORD" ]; then
          redis-server --requirepass "$$REDIS_PASSWORD"
        else
          redis-server
        fi
      '
    environment:
      REDIS_PASSWORD: ${REDIS_PASSWORD:-}
    volumes:
      - redis_data:/data
    healthcheck:
      test:
        - CMD
        - sh
        - -c
        - |
          if [ -n "$$REDIS_PASSWORD" ]; then
            redis-cli -a "$$REDIS_PASSWORD" ping
          else
            redis-cli ping
          fi
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: backend-dev
    ports:
      - "${BACKEND_PORT:-8080}:8080"
    labels:
      - "logging=promtail"
      - "service=backend"
    environment:
      # Server
      SERVER_PORT: 8080
      SERVER_READ_TIMEOUT: ${SERVER_READ_TIMEOUT:-10s}
      SERVER_WRITE_TIMEOUT: ${SERVER_WRITE_TIMEOUT:-10s}
      SERVER_IDLE_TIMEOUT: ${SERVER_IDLE_TIMEOUT:-120s}

      # Database
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${POSTGRES_USER:-postgres}
      DB_PASSWORD: ${POSTGRES_PASSWORD:?Please set POSTGRES_PASSWORD}
      DB_NAME: ${POSTGRES_DB:-inf_sys_db}
      DB_SSL_MODE: ${DB_SSL_MODE:-disable}
      DB_MAX_OPEN_CONNS: ${DB_MAX_OPEN_CONNS:-25}
      DB_MAX_IDLE_CONNS: ${DB_MAX_IDLE_CONNS:-5}
      DB_CONN_MAX_LIFETIME: ${DB_CONN_MAX_LIFETIME:-5m}

      # Redis
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD:-}
      REDIS_DB: ${REDIS_DB:-0}
      REDIS_POOL_SIZE: ${REDIS_POOL_SIZE:-10}

      # Authentication
      JWT_ACCESS_SECRET: ${JWT_ACCESS_SECRET:?Please set JWT_ACCESS_SECRET in .env}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET:?Please set JWT_REFRESH_SECRET in .env}
      JWT_ACCESS_TTL: ${JWT_ACCESS_TTL:-15m}
      JWT_REFRESH_TTL: ${JWT_REFRESH_TTL:-168h}

      # Composio (Gmail integration)
      COMPOSIO_API_KEY: ${COMPOSIO_API_KEY:-}
      COMPOSIO_ENTITY_ID: ${COMPOSIO_ENTITY_ID:-}
      COMPOSIO_MCP_CONFIG_ID: ${COMPOSIO_MCP_CONFIG_ID:-}

      # CORS
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000}
      CORS_ALLOWED_METHODS: ${CORS_ALLOWED_METHODS:-GET,POST,PUT,DELETE,OPTIONS}
      CORS_ALLOWED_HEADERS: ${CORS_ALLOWED_HEADERS:-Content-Type,Authorization}

      # Logging
      LOG_LEVEL: ${LOG_LEVEL:-info}
      LOG_FORMAT: ${LOG_FORMAT:-json}

      # Application
      ENVIRONMENT: ${ENVIRONMENT:-development}
      VERSION: ${VERSION:-0.1.0}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: frontend-dev
    ports:
      - "${FRONTEND_PORT:-3000}:3000"
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080}
      NODE_ENV: ${NODE_ENV:-development}
    depends_on:
      - backend

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
```

---

## Мониторинг стек

### compose.monitoring.yml:
```yaml
services:
  prometheus:
    image: prom/prometheus:v2.48.0
    container_name: prometheus
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
      - '--web.enable-lifecycle'
    ports:
      - "${PROMETHEUS_PORT:-9090}:9090"
    networks:
      - monitoring
      - default
    restart: unless-stopped

  grafana:
    image: grafana/grafana:10.2.2
    container_name: grafana
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER:-admin}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_SERVER_ROOT_URL=${GRAFANA_ROOT_URL:-http://localhost:3001}
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana_data:/var/lib/grafana
    ports:
      - "${GRAFANA_PORT:-3001}:3000"
    networks:
      - monitoring
      - default
    depends_on:
      - prometheus
      - loki
    restart: unless-stopped

  loki:
    image: grafana/loki:2.9.2
    container_name: loki
    volumes:
      - ./monitoring/loki/loki-config.yml:/etc/loki/local-config.yaml:ro
      - loki_data:/loki
    command: -config.file=/etc/loki/local-config.yaml
    ports:
      - "${LOKI_PORT:-3100}:3100"
    networks:
      - monitoring
      - default
    restart: unless-stopped

  promtail:
    image: grafana/promtail:2.9.2
    container_name: promtail
    volumes:
      - ./monitoring/promtail/promtail-config.yml:/etc/promtail/config.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - monitoring
      - default
    depends_on:
      - loki
    restart: unless-stopped

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
  loki_data:
    driver: local
```

### Запуск с мониторингом:
```bash
docker compose -f compose.yml -f compose.monitoring.yml up -d
```

### Доступные дашборды:
- **Grafana**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Loki**: http://localhost:3100

---

## Dockerfiles

### Backend Dockerfile:
```dockerfile
# Stage 1: Build
FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Копируем зависимости (кэшируется)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копируем код
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Stage 2: Runtime (scratch)
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Frontend Dockerfile:
```dockerfile
# Stage 1: Dependencies
FROM node:24-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --production=false --ignore-scripts

# Stage 2: Builder
FROM node:24-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# Stage 3: Runner
FROM node:24-alpine AS runner
WORKDIR /app

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

USER nextjs

ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://0.0.0.0:3000/', (r) => {process.exit(r.statusCode === 200 ? 0 : 1)})" || exit 1

CMD ["node", "server.js"]
```

---

## compose.override.yml

Для локальной разработки создайте `compose.override.yml`:

```yaml
# compose.override.yml (НЕ коммитить!)
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: your_secure_password_here
    ports:
      - "5432:5432"  # Раскомментировать для доступа с хоста

  redis:
    environment:
      REDIS_PASSWORD: optional_redis_password
    ports:
      - "6379:6379"  # Раскомментировать для доступа с хоста
```

---

## Health Endpoints

Backend предоставляет endpoints для мониторинга:

| Endpoint | Назначение |
|----------|------------|
| `/health` | Полный health check (DB + Redis) |
| `/live` | Kubernetes liveness probe |
| `/ready` | Kubernetes readiness probe |
| `/metrics` | Prometheus метрики |

```bash
# Проверка health
curl http://localhost:8080/health

# Проверка liveness
curl http://localhost:8080/live

# Проверка readiness
curl http://localhost:8080/ready

# Просмотр метрик
curl http://localhost:8080/metrics
```

---

## Команды управления

### Основные команды:
```bash
# Запуск всех сервисов
docker compose up -d

# Запуск с мониторингом
docker compose -f compose.yml -f compose.monitoring.yml up -d

# Остановка
docker compose down

# Просмотр логов
docker compose logs -f backend
docker compose logs -f frontend

# Пересборка backend
docker compose build backend
docker compose up -d backend

# Полная очистка (включая volumes)
docker compose down -v --remove-orphans
```

### Работа с базой данных:
```bash
# Подключение к PostgreSQL
docker compose exec postgres psql -U postgres -d inf_sys_db

# Подключение к Redis
docker compose exec redis redis-cli
```

---

## Troubleshooting

### Проверка занятых портов:
```bash
lsof -i :8080
lsof -i :3000
lsof -i :5432
```

### Просмотр использования ресурсов:
```bash
docker stats
```

### Очистка Docker ресурсов:
```bash
docker system prune -af
docker volume prune -f
```

### Проверка подключения к БД:
```bash
docker compose exec postgres psql -U postgres -d inf_sys_db -c "SELECT 1;"
```

### Проверка логов сервиса:
```bash
docker compose logs --tail=100 backend
```

---

## Безопасность

### Рекомендации:
1. **Никогда не коммитьте** `compose.override.yml` и `.env` с секретами
2. **Используйте сильные пароли** для PostgreSQL и JWT секретов
3. **Не открывайте порты БД** в production (убрать ports mapping)
4. **Используйте непривилегированных пользователей** в контейнерах (уже настроено)

### Docker secrets (production):
```yaml
secrets:
  db_password:
    file: ./secrets/db_password.txt
  jwt_secret:
    file: ./secrets/jwt_secret.txt

services:
  backend:
    secrets:
      - db_password
      - jwt_secret
```

---

**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный
