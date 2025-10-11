# Backend - Information System of Academic Secretary/Methodologist

Модульный монолит на Go с использованием Domain-Driven Design (DDD) и Clean Architecture.

## 🏗️ Архитектура

Проект построен на принципах модульного монолита с четкими границами между модулями:

### Структура проекта

```
.
├── cmd/
│   └── server/              # Application entry point
│       └── main.go          # Server initialization and startup
├── internal/
│   ├── shared/              # Shared kernel (cross-cutting concerns)
│   │   ├── domain/
│   │   │   ├── common/      # Base entities, value objects
│   │   │   └── events/      # Domain events and EventBus
│   │   ├── infrastructure/
│   │   │   ├── config/      # Configuration management
│   │   │   ├── database/    # Database connections, transactions
│   │   │   └── logging/     # Structured logging
│   │   └── application/
│   │       └── middleware/  # HTTP middleware (logging, auth, etc.)
│   └── modules/             # Business modules (DDD Bounded Contexts)
│       └── auth/            # Authentication & Authorization module
│           ├── domain/
│           │   ├── entities/      # Domain entities (User, Role)
│           │   ├── repositories/  # Repository interfaces
│           │   └── services/      # Domain services
│           ├── application/
│           │   ├── usecases/      # Use cases (Login, Register)
│           │   └── dto/           # Data Transfer Objects
│           └── infrastructure/
│               ├── persistence/   # Repository implementations
│               └── http/          # HTTP handlers
├── migrations/              # Database migrations
├── .env.example            # Environment variables template
├── go.mod                  # Go module definition
└── go.sum                  # Go dependencies checksums
```

## 🏗️ Architecture

### Modular Monolith

The backend follows a **modular monolith** architecture, designed for future migration to microservices:

- **10 Planned Modules**: auth, users, documents, workflow, schedule, reporting, tasks, notifications, files, integration
- **Shared Kernel**: Common domain logic, infrastructure, and utilities
- **Event-Driven Communication**: Modules communicate via domain events through EventBus
- **Clean Boundaries**: Each module is self-contained with clear interfaces

### Clean Architecture Layers

Each module follows Clean Architecture with 4 layers:

1. **Domain Layer** (`domain/`)
   - Entities (business objects)
   - Value Objects (immutable domain concepts)
   - Repository Interfaces (data access contracts)
   - Domain Services (business logic that doesn't fit entities)
   - Domain Events (business events)

2. **Application Layer** (`application/`)
   - Use Cases (application business rules)
   - DTOs (data transfer objects)
   - Application Services (orchestration)

3. **Infrastructure Layer** (`infrastructure/`)
   - Repository Implementations (PostgreSQL)
   - External Service Integrations
   - Framework-specific code

4. **Interface Layer** (`interfaces/` or `infrastructure/http/`)
   - HTTP Handlers
   - Request/Response mapping
   - API routing

### Key Patterns

- **Repository Pattern**: Abstract data access
- **Unit of Work**: Manage database transactions
- **CQRS**: Separate read and write operations
- **Event Sourcing Ready**: Domain events infrastructure
- **Dependency Injection**: Via constructor injection

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Make (optional)

### Installation

1. **Clone and navigate to repository**:
   ```bash
   git clone https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist.git
   cd inf-sys-secretary-methodist
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Setup database**:
   ```bash
   # Using Docker
   docker-compose up -d postgres redis

   # Or install PostgreSQL and Redis locally
   # Then run migrations (when available)
   ```

5. **Run the server**:
   ```bash
   go run cmd/server/main.go
   ```

   Server will start on `http://localhost:8080`

### Development Commands

```bash
# Run server with hot reload (requires air)
air

# Run tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage in browser
go tool cover -html=coverage.out

# Run linter
golangci-lint run

# Format code
gofmt -w .

# Build binary
go build -o bin/server cmd/server/main.go

# Run binary
./bin/server
```

## 🔧 Configuration

Configuration is managed via environment variables. See `.env.example` for all available options.

### Key Configuration Sections

#### Server Configuration
```env
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=120s
```

#### Database Configuration
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=inf_sys_db
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

#### Redis Configuration
```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

#### Logging Configuration
```env
LOG_LEVEL=info
LOG_FORMAT=json
```

#### Application Configuration
```env
APP_ENV=development
APP_VERSION=0.1.0
```

## 🧪 Testing

### Test Structure

```
internal/
└── modules/
    └── auth/
        ├── domain/
        │   └── entities/
        │       ├── user.go
        │       └── user_test.go      # Unit tests
        ├── application/
        │   └── usecases/
        │       ├── login_user.go
        │       └── login_user_test.go
        └── infrastructure/
            └── persistence/
                ├── user_repository.go
                └── user_repository_integration_test.go
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/modules/auth/...

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Integration tests (require database)
go test -tags=integration ./...

# Verbose output
go test -v ./...
```

### Test Coverage Goals

- **Domain Layer**: 90%+
- **Application Layer**: 85%+
- **Infrastructure Layer**: 70%+
- **Overall**: 80%+

## 🔒 Security

### Authentication & Authorization

- **JWT Tokens**: Stateless authentication
- **Role-Based Access Control (RBAC)**: 5 roles (admin, secretary, methodist, teacher, student)
- **Password Hashing**: bcrypt with salt
- **Token Refresh**: Secure refresh token mechanism

### Security Best Practices

- Input validation on all endpoints
- SQL injection prevention (parameterized queries)
- XSS protection
- CSRF protection
- Rate limiting
- Security headers (CORS, CSP, etc.)

## 📊 Database

### PostgreSQL Schema

Managed via migrations (planned):
- User authentication and authorization
- Document management
- Workflow states
- Schedule data
- Audit logs

### Migrations

```bash
# Apply migrations (planned)
migrate -path migrations -database "postgresql://user:pass@localhost:5432/inf_sys_db?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgresql://user:pass@localhost:5432/inf_sys_db?sslmode=disable" down 1
```

## 🔄 CI/CD

The backend has comprehensive CI/CD pipelines via GitHub Actions:

### Backend CI (`.github/workflows/backend-ci.yml`)
- **Linting**: golangci-lint with 30+ linters
- **Testing**: Unit and integration tests with 80% coverage threshold
- **Building**: Multi-OS builds (Linux, macOS)
- **Format Check**: gofmt compliance
- **Module Verification**: go.mod/go.sum integrity

### Security (`.github/workflows/security.yml`)
- **Trivy**: Vulnerability scanning
- **gosec**: Go security checker
- **TruffleHog**: Secret detection
- **CodeQL**: SAST analysis

### Database CI (`.github/workflows/database-ci.yml`)
- Migration testing (up/down/idempotency)
- Integration tests with PostgreSQL and Redis
- Docker Compose validation

See [CI/CD Workflows Documentation](docs/development/ci-cd-workflows.md) for details.

## 🎯 Code Quality Standards

### Linting Rules

Configured in `.golangci.yml`:
- errcheck: Check error handling
- gosimple: Simplify code
- govet: Standard Go checks
- gosec: Security issues
- gocyclo: Cyclomatic complexity (max 15)
- revive: Linting framework
- 20+ additional linters

### Code Style

- Follow [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- Use `gofmt` for formatting
- Maximum function complexity: 15
- Meaningful variable names
- Comprehensive error handling
- Document exported functions and types

### Commit Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

Types: feat, fix, refactor, docs, test, chore
Scope: module name (auth, documents, workflow, etc.)

Examples:
- feat(auth): add JWT token validation
- fix(documents): resolve file upload issue
- refactor(workflow): simplify state machine
- docs(api): update authentication endpoints
```

## 📚 API Documentation

### Health Check

```bash
GET /health
Response: {"status": "ok"}
```

### Authentication Endpoints (Planned)

```bash
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh
GET  /api/auth/me
```

Full API documentation will be available at `/api/docs` (Swagger UI - planned).

## 🚀 Deployment

### Docker

```bash
# Build image
docker build -t inf-sys-backend:latest .

# Run container
docker run -p 8080:8080 --env-file .env inf-sys-backend:latest
```

### Docker Compose

```bash
# Start all services (backend, PostgreSQL, Redis)
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down
```

### Binary Deployment

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/server cmd/server/main.go

# Copy binary and .env to server
scp bin/server user@server:/opt/inf-sys/
scp .env user@server:/opt/inf-sys/

# Run on server
cd /opt/inf-sys && ./server
```

## 📖 Additional Documentation

- [Project Overview](docs/project-overview.md)
- [Modular Architecture](docs/architecture/modular-architecture.md)
- [Microservices Migration Guide](docs/architecture/microservices-migration-guide.md)
- [Clean Code Patterns](docs/development/clean-code-patterns.md)
- [Development Guide](docs/development/development-guide.md)
- [CI/CD Workflows](docs/development/ci-cd-workflows.md)
- [Pull Request Guide](docs/development/pull-request-guide.md)

## 🤝 Contributing

1. Read [Development Guide](docs/development/development-guide.md)
2. Check [Pull Request Guide](docs/development/pull-request-guide.md)
3. Create feature branch: `feature/issue-N-description`
4. Write tests for new features
5. Ensure all CI checks pass
6. Submit PR with conventional commit format

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Tech Stack**: Go 1.21 • PostgreSQL 15 • Redis 7 • net/http • DDD • Clean Architecture
