# 🔧 Environment Configuration

## 📋 Обзор конфигурации окружений

Управление конфигурацией для различных окружений с использованием переменных среды, секретов и файлов конфигурации.

## 🌍 Окружения

### Development (Разработка)
```bash
# .env.development
APP_ENV=development
LOG_LEVEL=debug
DEBUG=true

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=inf_sys_dev
DB_USER=dev_user
DB_PASSWORD=dev_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Kafka
KAFKA_BROKERS=localhost:9092

# Authentication
JWT_SECRET=your-dev-jwt-secret-key
JWT_EXPIRES_IN=15m
REFRESH_TOKEN_EXPIRES_IN=7d

# OAuth
GOOGLE_CLIENT_ID=dev-google-client-id
GOOGLE_CLIENT_SECRET=dev-google-client-secret
AZURE_CLIENT_ID=dev-azure-client-id
AZURE_CLIENT_SECRET=dev-azure-client-secret

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_ENV=development
```

### Staging (Тестирование)
```bash
# .env.staging
APP_ENV=staging
LOG_LEVEL=info
DEBUG=false

# Database
DB_HOST=staging-db.example.com
DB_PORT=5432
DB_NAME=inf_sys_staging
DB_USER=staging_user
DB_PASSWORD=${DB_PASSWORD}  # From secrets

# Redis
REDIS_HOST=staging-redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}

# Kafka
KAFKA_BROKERS=staging-kafka.example.com:9092

# Authentication
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRES_IN=15m
REFRESH_TOKEN_EXPIRES_IN=7d

# OAuth
GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
AZURE_CLIENT_ID=${AZURE_CLIENT_ID}
AZURE_CLIENT_SECRET=${AZURE_CLIENT_SECRET}

# Frontend
NEXT_PUBLIC_API_URL=https://staging-api.inf-sys.example.com
NEXT_PUBLIC_ENV=staging
```

### Production (Продакшн)
```bash
# .env.production
APP_ENV=production
LOG_LEVEL=warn
DEBUG=false

# Database
DB_HOST=${DB_HOST}
DB_PORT=5432
DB_NAME=${DB_NAME}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}

# Redis
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}

# Kafka
KAFKA_BROKERS=${KAFKA_BROKERS}

# Authentication
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRES_IN=15m
REFRESH_TOKEN_EXPIRES_IN=7d

# OAuth
GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
AZURE_CLIENT_ID=${AZURE_CLIENT_ID}
AZURE_CLIENT_SECRET=${AZURE_CLIENT_SECRET}

# Frontend
NEXT_PUBLIC_API_URL=https://api.inf-sys.example.com
NEXT_PUBLIC_ENV=production

# Monitoring
PROMETHEUS_URL=${PROMETHEUS_URL}
GRAFANA_URL=${GRAFANA_URL}
JAEGER_URL=${JAEGER_URL}
```

---

## 🔐 Secrets Management

### HashiCorp Vault Configuration

#### Vault Setup:
```bash
# Install Vault
wget https://releases.hashicorp.com/vault/1.15.0/vault_1.15.0_linux_amd64.zip
unzip vault_1.15.0_linux_amd64.zip
sudo mv vault /usr/local/bin/

# Initialize Vault
vault server -dev
export VAULT_ADDR='http://127.0.0.1:8200'
vault auth -method=userpass
```

#### Secrets Structure:
```bash
# Authentication secrets
vault kv put secret/inf-sys/auth \
  jwt_secret="your-production-jwt-secret" \
  google_client_secret="google-oauth-secret" \
  azure_client_secret="azure-oauth-secret"

# Database secrets
vault kv put secret/inf-sys/database \
  postgres_password="secure-db-password" \
  redis_password="secure-redis-password"

# External services
vault kv put secret/inf-sys/integrations \
  one_c_api_key="1c-integration-key" \
  email_smtp_password="smtp-password"
```

### Kubernetes Secrets

#### Secret Creation:
```yaml
# k8s-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: inf-sys-database
  namespace: inf-sys
type: Opaque
data:
  postgres-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>

---
apiVersion: v1
kind: Secret
metadata:
  name: inf-sys-auth
  namespace: inf-sys
type: Opaque
data:
  jwt-secret: <base64-encoded-jwt-secret>
  google-client-secret: <base64-encoded-google-secret>
  azure-client-secret: <base64-encoded-azure-secret>
```

#### Secret Usage in Pods:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  template:
    spec:
      containers:
      - name: auth-service
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: inf-sys-database
              key: postgres-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: inf-sys-auth
              key: jwt-secret
```

---

## 📄 Configuration Files

### Service Configuration Templates

#### Auth Service Config:
```yaml
# configs/auth-service.yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DB_NAME}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  pool_size: 20
  max_idle: 5
  max_lifetime: 1h

redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}
  db: 0
  pool_size: 10

jwt:
  secret: ${JWT_SECRET}
  expires_in: ${JWT_EXPIRES_IN}
  refresh_expires_in: ${REFRESH_TOKEN_EXPIRES_IN}

oauth:
  google:
    client_id: ${GOOGLE_CLIENT_ID}
    client_secret: ${GOOGLE_CLIENT_SECRET}
    redirect_url: ${GOOGLE_REDIRECT_URL}
  azure:
    client_id: ${AZURE_CLIENT_ID}
    client_secret: ${AZURE_CLIENT_SECRET}
    tenant_id: ${AZURE_TENANT_ID}

logging:
  level: ${LOG_LEVEL}
  format: json
  output: stdout
```

#### Document Service Config:
```yaml
# configs/document-service.yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DOCS_DB_NAME}
  user: ${DOCS_DB_USER}
  password: ${DOCS_DB_PASSWORD}

kafka:
  brokers: ${KAFKA_BROKERS}
  topics:
    document_events: "document-events"
    workflow_events: "workflow-events"

storage:
  type: ${STORAGE_TYPE}  # local, s3, gcs
  local:
    path: "/app/storage"
  s3:
    bucket: ${S3_BUCKET}
    region: ${S3_REGION}
    access_key: ${S3_ACCESS_KEY}
    secret_key: ${S3_SECRET_KEY}

search:
  elasticsearch:
    url: ${ELASTICSEARCH_URL}
    index: "documents"
```

---

## 🐳 Docker Environment Configuration

### Development docker-compose.override.yml:
```yaml
version: '3.8'

services:
  auth-service:
    environment:
      - APP_ENV=development
      - LOG_LEVEL=debug
      - DEBUG=true
      - DB_PASSWORD=dev_password
    volumes:
      - ./configs/development:/app/configs

  postgres:
    environment:
      - POSTGRES_PASSWORD=dev_password
    ports:
      - "5432:5432"  # Expose for local development
```

### Production docker-compose.yml:
```yaml
version: '3.8'

services:
  auth-service:
    environment:
      - APP_ENV=production
      - LOG_LEVEL=warn
      - DEBUG=false
    env_file:
      - .env.production
    secrets:
      - db_password
      - jwt_secret
    configs:
      - source: auth_config
        target: /app/configs/config.yaml

secrets:
  db_password:
    external: true
  jwt_secret:
    external: true

configs:
  auth_config:
    file: ./configs/production/auth-service.yaml
```

---

## ☁️ Kubernetes ConfigMaps

### Service Configuration:
```yaml
# configmap-auth-service.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: inf-sys
data:
  config.yaml: |
    server:
      port: 8080
      host: "0.0.0.0"

    database:
      host: postgres-service
      port: 5432
      name: auth_db
      pool_size: 20

    logging:
      level: info
      format: json

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-environment
  namespace: inf-sys
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
  KAFKA_BROKERS: "kafka-service:9092"
```

### ConfigMap Usage:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  template:
    spec:
      containers:
      - name: auth-service
        envFrom:
        - configMapRef:
            name: app-environment
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
      volumes:
      - name: config-volume
        configMap:
          name: auth-service-config
```

---

## 🔄 Environment Variables Loading

### Go Configuration Loading:
```go
package config

import (
    "os"
    "time"
    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    // Server
    Port     string `envconfig:"PORT" default:"8080"`
    Host     string `envconfig:"HOST" default:"0.0.0.0"`
    AppEnv   string `envconfig:"APP_ENV" default:"development"`
    Debug    bool   `envconfig:"DEBUG" default:"false"`
    LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

    // Database
    DBHost     string `envconfig:"DB_HOST" required:"true"`
    DBPort     string `envconfig:"DB_PORT" default:"5432"`
    DBName     string `envconfig:"DB_NAME" required:"true"`
    DBUser     string `envconfig:"DB_USER" required:"true"`
    DBPassword string `envconfig:"DB_PASSWORD" required:"true"`

    // Redis
    RedisHost     string `envconfig:"REDIS_HOST" required:"true"`
    RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
    RedisPassword string `envconfig:"REDIS_PASSWORD"`

    // JWT
    JWTSecret         string        `envconfig:"JWT_SECRET" required:"true"`
    JWTExpiresIn      time.Duration `envconfig:"JWT_EXPIRES_IN" default:"15m"`
    RefreshExpiresIn  time.Duration `envconfig:"REFRESH_TOKEN_EXPIRES_IN" default:"168h"`

    // OAuth
    GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
    GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
    AzureClientID      string `envconfig:"AZURE_CLIENT_ID"`
    AzureClientSecret  string `envconfig:"AZURE_CLIENT_SECRET"`
}

func Load() (*Config, error) {
    // Load .env file if exists
    if _, err := os.Stat(".env"); err == nil {
        if err := godotenv.Load(); err != nil {
            return nil, err
        }
    }

    // Load environment-specific .env file
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }

    envFile := fmt.Sprintf(".env.%s", env)
    if _, err := os.Stat(envFile); err == nil {
        if err := godotenv.Load(envFile); err != nil {
            return nil, err
        }
    }

    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### Next.js Environment Configuration:
```javascript
// next.config.js
const { PHASE_DEVELOPMENT_SERVER } = require('next/constants')

module.exports = (phase, { defaultConfig }) => {
  const isDev = phase === PHASE_DEVELOPMENT_SERVER

  return {
    env: {
      CUSTOM_KEY: process.env.CUSTOM_KEY,
    },
    publicRuntimeConfig: {
      apiUrl: process.env.NEXT_PUBLIC_API_URL,
      environment: process.env.NEXT_PUBLIC_ENV,
    },
    serverRuntimeConfig: {
      apiSecret: process.env.API_SECRET,
    },
  }
}
```

---

## 🔍 Environment Validation

### Environment Validation Script:
```bash
#!/bin/bash
# scripts/validate-env.sh

echo "🔍 Validating environment configuration..."

# Required variables check
required_vars=(
    "APP_ENV"
    "DB_HOST"
    "DB_NAME"
    "DB_USER"
    "DB_PASSWORD"
    "REDIS_HOST"
    "JWT_SECRET"
)

missing_vars=()

for var in "${required_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        missing_vars+=("$var")
    fi
done

if [[ ${#missing_vars[@]} -gt 0 ]]; then
    echo "❌ Missing required environment variables:"
    printf '   %s\n' "${missing_vars[@]}"
    exit 1
fi

# Validate JWT secret strength
if [[ ${#JWT_SECRET} -lt 32 ]]; then
    echo "❌ JWT_SECRET must be at least 32 characters"
    exit 1
fi

# Validate database connection
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; then
    echo "❌ Cannot connect to database"
    exit 1
fi

# Validate Redis connection
if ! redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null; then
    echo "❌ Cannot connect to Redis"
    exit 1
fi

echo "✅ Environment validation passed"
```

### Go Environment Validation:
```go
func (cfg *Config) Validate() error {
    if cfg.AppEnv == "production" {
        if len(cfg.JWTSecret) < 32 {
            return errors.New("JWT secret must be at least 32 characters in production")
        }

        if cfg.Debug {
            return errors.New("debug mode must be disabled in production")
        }

        if cfg.LogLevel == "debug" {
            return errors.New("log level should not be debug in production")
        }
    }

    return nil
}
```

---

## 📊 Environment Monitoring

### Health Check Endpoints:
```go
func (h *HealthHandler) Check(c *gin.Context) {
    status := map[string]interface{}{
        "status": "healthy",
        "environment": os.Getenv("APP_ENV"),
        "version": os.Getenv("APP_VERSION"),
        "timestamp": time.Now().UTC(),
        "checks": map[string]interface{}{
            "database": h.checkDatabase(),
            "redis": h.checkRedis(),
            "kafka": h.checkKafka(),
        },
    }

    c.JSON(http.StatusOK, status)
}
```

### Environment Info Endpoint:
```go
func (h *InfoHandler) GetInfo(c *gin.Context) {
    info := map[string]interface{}{
        "environment": os.Getenv("APP_ENV"),
        "version": os.Getenv("APP_VERSION"),
        "build_time": os.Getenv("BUILD_TIME"),
        "commit_hash": os.Getenv("COMMIT_HASH"),
        "go_version": runtime.Version(),
    }

    c.JSON(http.StatusOK, info)
}
```

---

## 🚀 Deployment Scripts

### Environment Setup Script:
```bash
#!/bin/bash
# scripts/setup-environment.sh

set -e

ENVIRONMENT=${1:-development}

echo "🚀 Setting up $ENVIRONMENT environment..."

# Create necessary directories
mkdir -p logs storage temp

# Copy environment file
if [[ -f ".env.$ENVIRONMENT" ]]; then
    cp ".env.$ENVIRONMENT" ".env"
    echo "✅ Environment file copied"
else
    echo "❌ Environment file .env.$ENVIRONMENT not found"
    exit 1
fi

# Validate environment
if [[ -x "scripts/validate-env.sh" ]]; then
    ./scripts/validate-env.sh
fi

# Run migrations
if [[ "$ENVIRONMENT" != "production" ]]; then
    echo "🔄 Running database migrations..."
    make migrate-up
fi

# Start services
if [[ "$ENVIRONMENT" == "development" ]]; then
    echo "🔄 Starting development services..."
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
else
    echo "🔄 Starting services..."
    docker-compose up -d
fi

echo "✅ Environment setup complete"
```

Правильная конфигурация окружений обеспечивает безопасность и надежность системы!
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

