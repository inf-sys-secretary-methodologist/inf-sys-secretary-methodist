# 🐳 Docker Setup

## 📋 Обзор контейнеризации

Полная настройка Docker и Docker Compose для локальной разработки и развертывания микросервисной архитектуры.

## 🚀 Quick Start

### Запуск всей системы:
```bash
# Клонирование проекта
git clone https://github.com/your-org/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist

# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

---

## 🐳 Docker Configuration

### Main docker-compose.yml:
```yaml
version: '3.8'

services:
  # ===================
  # Infrastructure
  # ===================

  postgres:
    image: postgres:17-alpine
    container_name: inf-sys-postgres
    environment:
      POSTGRES_DB: inf_sys
      POSTGRES_USER: inf_sys_user
      POSTGRES_PASSWORD: ${DB_PASSWORD:-dev_password}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - inf-sys-network

  redis:
    image: redis:7-alpine
    container_name: inf-sys-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - inf-sys-network

  kafka:
    image: confluentinc/cp-kafka:latest
    container_name: inf-sys-kafka
    depends_on:
      - zookeeper
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"
    networks:
      - inf-sys-network

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    container_name: inf-sys-zookeeper
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    networks:
      - inf-sys-network

  # ===================
  # Backend Services
  # ===================

  auth-service:
    build:
      context: ./services/auth-service
      dockerfile: Dockerfile
    container_name: inf-sys-auth
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=auth_db
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    ports:
      - "8001:8080"
    depends_on:
      - postgres
      - redis
    networks:
      - inf-sys-network

  user-service:
    build:
      context: ./services/user-service
      dockerfile: Dockerfile
    container_name: inf-sys-users
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=users_db
    ports:
      - "8002:8080"
    depends_on:
      - postgres
    networks:
      - inf-sys-network

  document-service:
    build:
      context: ./services/document-service
      dockerfile: Dockerfile
    container_name: inf-sys-documents
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=documents_db
      - KAFKA_BROKERS=kafka:9092
    ports:
      - "8003:8080"
    depends_on:
      - postgres
      - kafka
    networks:
      - inf-sys-network

  workflow-service:
    build:
      context: ./services/workflow-service
      dockerfile: Dockerfile
    container_name: inf-sys-workflow
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=workflow_db
      - KAFKA_BROKERS=kafka:9092
    ports:
      - "8004:8080"
    depends_on:
      - postgres
      - kafka
    networks:
      - inf-sys-network

  # ===================
  # Frontend
  # ===================

  admin-dashboard:
    build:
      context: ./frontend/admin-dashboard
      dockerfile: Dockerfile
    container_name: inf-sys-admin-ui
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
      - NEXT_PUBLIC_ENV=development
    ports:
      - "3001:3000"
    networks:
      - inf-sys-network

  user-portal:
    build:
      context: ./frontend/user-portal
      dockerfile: Dockerfile
    container_name: inf-sys-user-ui
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8080
      - NEXT_PUBLIC_ENV=development
    ports:
      - "3002:3000"
    networks:
      - inf-sys-network

  # ===================
  # API Gateway
  # ===================

  api-gateway:
    image: nginx:alpine
    container_name: inf-sys-gateway
    ports:
      - "8080:80"
    volumes:
      - ./configs/nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - auth-service
      - user-service
      - document-service
      - workflow-service
    networks:
      - inf-sys-network

volumes:
  postgres_data:
  redis_data:

networks:
  inf-sys-network:
    driver: bridge
```

---

## 🔧 Service Dockerfiles

### Backend Service Dockerfile:
```dockerfile
# services/[service-name]/Dockerfile

# Build stage
FROM golang:1.25.3-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git ca-certificates tzdata

# Создание пользователя для безопасности
RUN adduser -D -g '' appuser

WORKDIR /app

# Копирование go mod файлов и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main cmd/server/main.go

# Runtime stage
FROM scratch

# Копирование CA сертификатов
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копирование timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копирование пользователя
COPY --from=builder /etc/passwd /etc/passwd

# Копирование приложения
COPY --from=builder /app/main /main
COPY --from=builder /app/configs /configs

# Использование непривилегированного пользователя
USER appuser

# Expose порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/main", "health"] || exit 1

# Запуск приложения
ENTRYPOINT ["/main"]
```

### Frontend Dockerfile:
```dockerfile
# frontend/[app-name]/Dockerfile

# Build stage
FROM node:25-alpine AS builder

WORKDIR /app

# Копирование package files
COPY package*.json ./
RUN npm ci --only=production

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN npm run build

# Runtime stage
FROM node:25-alpine AS runner

# Создание пользователя для безопасности
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

WORKDIR /app

# Копирование built приложения
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/api/health || exit 1

CMD ["node", "server.js"]
```

---

## 🔧 Development Environment

### docker-compose.dev.yml:
```yaml
# Расширение для разработки
version: '3.8'

services:
  auth-service:
    volumes:
      - ./services/auth-service:/app
    environment:
      - HOT_RELOAD=true
    command: ["air", "-c", ".air.toml"]

  user-service:
    volumes:
      - ./services/user-service:/app
    environment:
      - HOT_RELOAD=true
    command: ["air", "-c", ".air.toml"]

  # Дополнительные сервисы для разработки
  mailhog:
    image: mailhog/mailhog
    container_name: inf-sys-mailhog
    ports:
      - "1025:1025"  # SMTP
      - "8025:8025"  # Web UI
    networks:
      - inf-sys-network

  adminer:
    image: adminer:latest
    container_name: inf-sys-adminer
    ports:
      - "8080:8080"
    environment:
      ADMINER_DEFAULT_SERVER: postgres
    networks:
      - inf-sys-network
```

### Команды для разработки:
```bash
# Разработка с hot reload
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# Только инфраструктура (БД, Redis, Kafka)
docker-compose up -d postgres redis kafka zookeeper

# Остановка всех сервисов
docker-compose down

# Полная очистка (с удалением данных)
docker-compose down -v --remove-orphans
```

---

## 🗄️ Database Setup

### PostgreSQL Init Scripts:
```sql
-- migrations/001_init.sql
CREATE DATABASE auth_db;
CREATE DATABASE users_db;
CREATE DATABASE documents_db;
CREATE DATABASE workflow_db;
CREATE DATABASE schedule_db;
CREATE DATABASE reports_db;

-- Создание пользователей для каждого сервиса
CREATE USER auth_user WITH PASSWORD 'auth_password';
CREATE USER users_user WITH PASSWORD 'users_password';
CREATE USER docs_user WITH PASSWORD 'docs_password';
CREATE USER workflow_user WITH PASSWORD 'workflow_password';
CREATE USER schedule_user WITH PASSWORD 'schedule_password';
CREATE USER reports_user WITH PASSWORD 'reports_password';

-- Назначение прав
GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth_user;
GRANT ALL PRIVILEGES ON DATABASE users_db TO users_user;
GRANT ALL PRIVILEGES ON DATABASE documents_db TO docs_user;
GRANT ALL PRIVILEGES ON DATABASE workflow_db TO workflow_user;
GRANT ALL PRIVILEGES ON DATABASE schedule_db TO schedule_user;
GRANT ALL PRIVILEGES ON DATABASE reports_db TO reports_user;
```

### Database Environment Variables:
```bash
# .env.development
DB_HOST=localhost
DB_PORT=5432

# Service-specific databases
AUTH_DB_NAME=auth_db
AUTH_DB_USER=auth_user
AUTH_DB_PASSWORD=auth_password

USERS_DB_NAME=users_db
USERS_DB_USER=users_user
USERS_DB_PASSWORD=users_password

DOCS_DB_NAME=documents_db
DOCS_DB_USER=docs_user
DOCS_DB_PASSWORD=docs_password
```

---

## 🔥 Hot Reload Configuration

### Air для Go (Hot Reload):
```toml
# .air.toml для каждого сервиса
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main cmd/server/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

### Next.js Hot Reload:
```json
// package.json
{
  "scripts": {
    "dev": "next dev",
    "dev:docker": "next dev -H 0.0.0.0",
    "build": "next build",
    "start": "next start"
  }
}
```

---

## 🏭 Production Docker Setup

### Production docker-compose.yml:
```yaml
version: '3.8'

services:
  # Production-optimized configs
  auth-service:
    image: inf-sys/auth-service:${VERSION}
    restart: unless-stopped
    environment:
      - APP_ENV=production
      - DB_HOST=${DB_HOST}
      - DB_PASSWORD=${DB_PASSWORD}
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 128M

  postgres:
    image: postgres:17
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_prod_data:/var/lib/postgresql/data
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G

volumes:
  postgres_prod_data:
    external: true
```

---

## 🔒 Security Configuration

### Security Best Practices:

#### Non-root Users:
```dockerfile
# Создание непривилегированного пользователя
RUN adduser -D -g '' appuser
USER appuser
```

#### Read-only Root Filesystem:
```yaml
# docker-compose.yml
services:
  auth-service:
    read_only: true
    tmpfs:
      - /tmp:rw,size=100M
```

#### Resource Limits:
```yaml
services:
  auth-service:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
          pids: 100
```

### Secret Management:
```yaml
# Использование Docker secrets в production
secrets:
  db_password:
    file: ./secrets/db_password.txt
  jwt_secret:
    file: ./secrets/jwt_secret.txt

services:
  auth-service:
    secrets:
      - db_password
      - jwt_secret
    environment:
      - DB_PASSWORD_FILE=/run/secrets/db_password
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
```

---

## 📊 Monitoring в Docker

### Prometheus + Grafana:
```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: inf-sys-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    networks:
      - inf-sys-network

  grafana:
    image: grafana/grafana:latest
    container_name: inf-sys-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    networks:
      - inf-sys-network

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: inf-sys-jaeger
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - inf-sys-network
```

### Health Checks:
```yaml
# Для каждого сервиса
services:
  auth-service:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

---

## 🚀 Deployment Commands

### Development:
```bash
# Полный старт для разработки
make dev-start

# Только инфраструктура
make infra-start

# Пересборка сервиса
make rebuild SERVICE=auth-service

# Просмотр логов сервиса
make logs SERVICE=auth-service

# Подключение к базе данных
make db-connect
```

### Makefile для автоматизации:
```makefile
# Makefile
.PHONY: dev-start infra-start rebuild logs db-connect

dev-start:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

infra-start:
	docker-compose up -d postgres redis kafka zookeeper

rebuild:
	docker-compose build $(SERVICE)
	docker-compose up -d $(SERVICE)

logs:
	docker-compose logs -f $(SERVICE)

db-connect:
	docker-compose exec postgres psql -U inf_sys_user -d inf_sys

stop:
	docker-compose down

clean:
	docker-compose down -v --remove-orphans
	docker system prune -af

test:
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Production commands
prod-deploy:
	docker-compose -f docker-compose.prod.yml up -d

prod-update:
	docker-compose -f docker-compose.prod.yml pull
	docker-compose -f docker-compose.prod.yml up -d --remove-orphans
```

---

## 🧪 Testing с Docker

### docker-compose.test.yml:
```yaml
version: '3.8'

services:
  test-postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_DB: test_db
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_password
    tmpfs:
      - /var/lib/postgresql/data

  auth-service-test:
    build:
      context: ./services/auth-service
      dockerfile: Dockerfile.test
    environment:
      - APP_ENV=test
      - DB_HOST=test-postgres
    depends_on:
      - test-postgres
    command: ["go", "test", "./...", "-v", "-race", "-coverprofile=coverage.out"]

  integration-tests:
    build:
      context: ./tests
      dockerfile: Dockerfile.integration
    depends_on:
      - auth-service
      - user-service
      - document-service
    environment:
      - API_BASE_URL=http://api-gateway:80
    command: ["go", "test", "./integration/...", "-v"]
```

### Test Dockerfile:
```dockerfile
# Dockerfile.test
FROM golang:1.25.3-alpine

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Установка testing tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install github.com/securecodewarrior/nancy@latest

# Запуск тестов
CMD ["make", "test"]
```

---

## ⚡ Docker Performance Tuning

### Resource Optimization:

#### Memory Limits:
```yaml
# Оптимизированные лимиты памяти
services:
  small-service:
    deploy:
      resources:
        limits:
          memory: 256M
        reservations:
          memory: 128M

  medium-service:
    deploy:
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

#### CPU Optimization:
```yaml
# CPU limits для предотвращения noisy neighbor
services:
  cpu-intensive-service:
    deploy:
      resources:
        limits:
          cpus: '1.0'
        reservations:
          cpus: '0.25'
```

### Caching Strategies:
```dockerfile
# Layer caching optimization
COPY go.mod go.sum ./
RUN go mod download  # Этот слой кэшируется

COPY . .             # Этот слой пересобирается при изменении кода
RUN go build ...
```

---

## 🔧 Troubleshooting

### Common Issues:

#### Port Conflicts:
```bash
# Проверка занятых портов
lsof -i :8080

# Остановка конфликтующих процессов
docker-compose down
pkill -f "8080"
```

#### Memory Issues:
```bash
# Мониторинг использования памяти
docker stats

# Очистка неиспользуемых ресурсов
docker system prune -af
docker volume prune -f
```

#### Database Connection Issues:
```bash
# Проверка подключения к БД
docker-compose exec postgres psql -U inf_sys_user -d inf_sys -c "SELECT 1;"

# Проверка логов БД
docker-compose logs postgres
```

### Debug Commands:
```bash
# Подключение к контейнеру
docker-compose exec auth-service sh

# Просмотр переменных окружения
docker-compose exec auth-service env

# Проверка сетевой связности
docker-compose exec auth-service ping postgres
```

Конфигурация Docker обеспечивает простое разработку и масштабирование микросервисной архитектуры!
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

