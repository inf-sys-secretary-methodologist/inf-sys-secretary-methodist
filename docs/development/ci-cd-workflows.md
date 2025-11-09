# 🔄 CI/CD Workflows Configuration

Complete guide to Continuous Integration and Continuous Deployment workflows for the Information System project.

## 📊 Overview

The project uses **GitHub Actions** for automated testing, security scanning, documentation validation, and deployment processes through six specialized workflows:

| Workflow | Purpose | Trigger |
|----------|---------|---------|
| **Backend CI** | Go code quality, testing, and building | Push/PR to backend code |
| **Documentation CI** | Documentation quality and link validation | Push/PR to docs |
| **Security** | Vulnerability scanning and secret detection | Daily + Push/PR |
| **Database CI** | Database migrations and integration tests | Push/PR to migrations |
| **PR Validation** | PR title, branch naming, size checks | Pull requests |
| **Dependabot** | Automated dependency updates | Weekly |

## 🏗️ Backend CI Workflow

### File: `.github/workflows/backend-ci.yml`

Comprehensive Go backend testing and building pipeline.

#### Triggers
```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'internal/**'
      - 'cmd/**'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/backend-ci.yml'
      - '.golangci.yml'
  pull_request:
    paths: [same as above]
```

#### Jobs

##### 1. Lint (`lint`)
- **Tool**: golangci-lint v1.61.0
- **Linters**: 30+ enabled (errcheck, gosimple, govet, gosec, gocyclo, revive, etc.)
- **Configuration**: `.golangci.yml`
- **Timeout**: 5 minutes
- **Features**:
  - Caches results for faster subsequent runs
  - Checks test files
  - Reports issues inline in PR

##### 2. Test (`test`)
- **Go Version**: 1.21
- **Coverage**: 80% threshold
- **Features**:
  - Race detector enabled (`-race`)
  - Short mode (`-short`)
  - Coverage profile generated (`coverage.out`)
  - Detailed output with `-v`
- **Command**:
  ```bash
  go test -v -race -coverprofile=coverage.out -covermode=atomic -short ./...
  ```

##### 3. Build (`build`)
- **Strategy**: Matrix build for multiple OS
  - Linux (ubuntu-latest)
  - macOS (macos-latest)
- **Binary Output**: `bin/server`
- **Command**:
  ```bash
  go build -o bin/server cmd/server/main.go
  ```

##### 4. Format Check (`format-check`)
- **Tool**: `gofmt`
- **Purpose**: Ensures code formatting compliance
- **Command**:
  ```bash
  test -z $(gofmt -l .)
  ```

##### 5. Module Verification (`mod-verify`)
- **Purpose**: Validates `go.mod` and `go.sum` integrity
- **Command**:
  ```bash
  go mod verify
  ```

### Configuration: `.golangci.yml`

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor
    - bin

linters:
  enable:
    - errcheck       # Check error handling
    - gosimple       # Simplify code
    - govet          # Vet examines Go code
    - ineffassign    # Detect ineffectual assignments
    - staticcheck    # Advanced static analysis
    - unused         # Check for unused code
    - gosec          # Security checks
    - gocyclo        # Cyclomatic complexity
    - gofmt          # Format check
    - goimports      # Import management
    - revive         # Linting framework
    - misspell       # Spell checking
    - unconvert      # Unnecessary conversions
    - dupl           # Duplicate code
    - goconst        # Repeated strings
    - gocritic       # Comprehensive checks

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
  misspell:
    locale: US
  errcheck:
    check-blank: true

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

## 📚 Documentation CI Workflow

### File: `.github/workflows/docs.yml`

Documentation quality assurance and synchronization validation.

#### Triggers
```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'docs/**'
      - 'README.md'
      - '.env.example'
      - '.github/workflows/docs.yml'
  pull_request: [same paths]
```

#### Jobs

##### 1. Markdown Linting (`markdown-lint`)
- **Tool**: markdownlint-cli2
- **Config**: `.markdownlint.json`
- **Scope**: All `docs/**/*.md`, `README.md`

##### 2. Link Checking (`link-checker`)
- **Tool**: markdown-link-check
- **Config**: `.github/markdown-link-config.json`
- **Features**: Validates all links, configurable timeouts and retries

##### 3. Structure Validation (`structure-check`)
- **Purpose**: Ensures required documentation exists
- **Checks**:
  - `docs/README.md`
  - `docs/api/api-documentation.md`
  - `docs/architecture/modular-architecture.md`
  - `docs/development/development-guide.md`

##### 4. Spell Checking (`spell-check`)
- **Tool**: cspell
- **Vocabulary**: `.github/wordlist.txt`
- **Mode**: Non-blocking warnings

##### 5. Environment Config Sync (`env-config-sync`)
- **Purpose**: Validates `.env.example` matches `internal/shared/infrastructure/config/config.go`
- **Checks**:
  - All config fields documented in `.env.example`
  - No drift between config structs and environment variables
  - Proper naming conventions

## 🔒 Security Workflow

### File: `.github/workflows/security.yml`

Multi-layered security scanning pipeline.

#### Triggers
```yaml
on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
```

#### Jobs

##### 1. Trivy Vulnerability Scan (`trivy-scan`)
- **Target**: Filesystem vulnerability scanning
- **Format**: SARIF (uploaded to GitHub Security tab)
- **Severity**: Checks CRITICAL, HIGH, MEDIUM, LOW
- **Coverage**: Go dependencies, OS packages

##### 2. Go Security Check (`gosec`)
- **Tool**: gosec - Go security checker
- **Output**: SARIF format
- **Checks**:
  - SQL injection vulnerabilities
  - Hardcoded credentials
  - Weak crypto usage
  - File path traversal
  - And more...

##### 3. Secret Scanning (`trufflehog`)
- **Tool**: TruffleHog
- **Scope**: Git history and current code
- **Detects**: API keys, tokens, passwords, certificates

##### 4. CodeQL Analysis (`codeql`)
- **Language**: Go
- **Type**: SAST (Static Application Security Testing)
- **Queries**: Default security queries
- **Integration**: GitHub Advanced Security

##### 5. License Compliance (`license-check`)
- **Tool**: go-licenses
- **Purpose**: Ensure dependency license compatibility
- **Allowed**: MIT, Apache-2.0, BSD-*

## 🗄️ Database CI Workflow

### File: `.github/workflows/database-ci.yml`

Database migrations and integration testing.

#### Triggers
```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'internal/shared/infrastructure/database/**'
      - 'migrations/**'
      - 'docker-compose.yml'
  pull_request: [same paths]
```

#### Jobs

##### 1. Migration Tests (`migration-tests`)
- **Services**:
  - PostgreSQL 15 (port 5432)
  - Redis 7 (port 6379)
- **Tests**:
  - **Up Migration**: Apply all migrations
  - **Down Migration**: Rollback all migrations
  - **Idempotency**: Ensure migrations can run multiple times safely
  - **Data Integrity**: Verify constraints and indexes
- **Env Variables**: Configured for CI environment

##### 2. Integration Tests (`integration-tests`)
- **Purpose**: Test database layer with real PostgreSQL
- **Coverage**: Repository pattern, transactions, connection pooling
- **Dependencies**: Requires PostgreSQL and Redis services

##### 3. Docker Compose Validation (`docker-compose-check`)
- **Command**: `docker-compose config --quiet`
- **Purpose**: Validates `docker-compose.yml` syntax and configuration

## ✅ PR Validation Workflow

### File: `.github/workflows/pr-validation.yml`

Automated pull request quality checks.

#### Triggers
```yaml
on:
  pull_request:
    types: [opened, edited, synchronize, reopened]
```

#### Jobs

##### 1. Semantic PR Title (`pr-title-check`)
- **Tool**: semantic-pull-request action
- **Required Format**:
  ```
  <type>(<scope>): <description>

  Types: feat, fix, docs, style, refactor, test, chore
  Example: feat(auth): add JWT token validation
  ```

##### 2. Branch Naming (`branch-name-check`)
- **Pattern**: `^(feature|fix|refactor|docs|test|chore)/issue-[0-9]+-[a-z0-9-]+$`
- **Examples**:
  - ✅ `feature/issue-123-add-user-auth`
  - ✅ `fix/issue-456-resolve-memory-leak`
  - ❌ `my-feature-branch`

##### 3. PR Size Check (`pr-size-check`)
- **Thresholds**:
  - ⚠️ Warning: >500 lines changed
  - ❌ Failure: >1000 lines changed
- **Purpose**: Encourage smaller, reviewable PRs

##### 4. Issue Linking (`issue-link-check`)
- **Requirement**: PR must reference an issue
- **Patterns**: `#123`, `closes #123`, `fixes #123`, `issue-123`

## 🔄 Dependabot Configuration

### File: `.github/dependabot.yml`

Automated dependency update management.

#### Update Schedules

##### 1. Go Modules (Weekly)
```yaml
- package-ecosystem: "gomod"
  directory: "/"
  schedule:
    interval: "weekly"
    day: "monday"
    time: "02:00"
  open-pull-requests-limit: 10
```

##### 2. GitHub Actions (Weekly)
```yaml
- package-ecosystem: "github-actions"
  directory: "/"
  schedule:
    interval: "weekly"
    day: "monday"
```

##### 3. npm (Frontend - Weekly)
```yaml
- package-ecosystem: "npm"
  directory: "/frontend"
  schedule:
    interval: "weekly"
    day: "monday"
```

##### 4. Docker (Weekly)
```yaml
- package-ecosystem: "docker"
  directory: "/"
  schedule:
    interval: "weekly"
```

## 🔧 Local Development Integration

### Running CI Checks Locally

#### Backend Checks
```bash
# Linting
golangci-lint run

# Testing with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Build
go build -o bin/server cmd/server/main.go

# Format check
gofmt -l .

# Module verification
go mod verify
```

#### Frontend Checks
```bash
cd frontend

# Install dependencies
npm install

# Linting
npm run lint

# Type checking
npm run type-check

# Unit tests
npm run test:unit

# E2E tests
npm run test:e2e
```

#### Documentation Checks
```bash
# Markdown linting
markdownlint-cli2 "docs/**/*.md" README.md

# Link checking
find docs -name "*.md" -exec markdown-link-check {} \;

# Spell checking
cspell "docs/**/*.md"
```

#### Security Checks
```bash
# Trivy scan
trivy fs .

# gosec
gosec ./...

# Check for secrets
trufflehog filesystem .
```

### Installing Tools

#### Backend Tools
```bash
# golangci-lint
brew install golangci-lint  # macOS
# or
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

#### Documentation Tools
```bash
npm install -g markdownlint-cli2 markdown-link-check cspell
```

#### Security Tools
```bash
# Trivy
brew install aquasecurity/trivy/trivy  # macOS

# TruffleHog
brew install trufflesecurity/trufflehog/trufflehog  # macOS
```

## 📈 Workflow Status

### GitHub Actions Badges

Add these to your README.md:

```markdown
![Backend CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/backend-ci.yml/badge.svg)
![Documentation CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/docs.yml/badge.svg)
![Security](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/security.yml/badge.svg)
![Database CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/database-ci.yml/badge.svg)
```

## 🚨 Troubleshooting

### Common Issues

#### 1. golangci-lint Timeout
- **Cause**: Large codebase or slow linters
- **Solution**: Increase timeout in `.golangci.yml`:
  ```yaml
  run:
    timeout: 10m
  ```

#### 2. Test Coverage Below Threshold
- **Cause**: Insufficient test coverage
- **Solution**: Add tests or adjust threshold temporarily

#### 3. Migration Test Failures
- **Cause**: PostgreSQL connection issues or migration conflicts
- **Solution**: Check migration order and dependencies

#### 4. PR Size Check Failure
- **Cause**: PR too large (>1000 lines)
- **Solution**: Split into smaller PRs

#### 5. Branch Naming Violation
- **Cause**: Branch doesn't follow naming convention
- **Solution**: Rename branch:
  ```bash
  git branch -m feature/issue-123-description
  ```

### Best Practices

1. **🔍 Test Locally First**: Run CI checks before pushing
2. **📝 Small PRs**: Keep changes focused and reviewable
3. **🏷️ Semantic Commits**: Use conventional commit format
4. **🔧 Monitor Actions**: Check GitHub Actions tab regularly
5. **🔒 Security First**: Address security findings immediately
6. **📊 Coverage**: Maintain or improve test coverage with each PR

## 🔄 CI/CD Pipeline Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Push/PR                        │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
┌───────────────┐       ┌──────────────┐
│  Backend CI   │       │  Docs CI     │
│  - Lint       │       │  - Markdown  │
│  - Test       │       │  - Links     │
│  - Build      │       │  - Spell     │
│  - Format     │       │  - Env Sync  │
└───────┬───────┘       └──────────────┘
        │
        ▼
┌───────────────────────────────────────┐
│          Security Workflow             │
│  - Trivy (vulnerabilities)            │
│  - gosec (Go security)                │
│  - TruffleHog (secrets)               │
│  - CodeQL (SAST)                      │
└───────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────┐
│        Database CI Workflow            │
│  - Migration tests                    │
│  - Integration tests                  │
│  - Docker Compose validation          │
└───────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────┐
│        PR Validation Workflow          │
│  - Semantic PR title                  │
│  - Branch naming                      │
│  - PR size check                      │
│  - Issue linking                      │
└───────────────────────────────────────┘
```

## 🎯 Quality Gates

All PRs must pass:

✅ **Backend CI**

- All linters pass
- Tests pass with ≥80% coverage
- Builds successfully on Linux & macOS
- Code properly formatted
- Go modules verified

✅ **Documentation CI**

- Markdown properly formatted
- All links valid
- Required docs exist
- Config files in sync

✅ **Security**

- No critical/high vulnerabilities
- No security issues found by gosec
- No secrets in code
- No unsafe dependencies

✅ **PR Validation**

- Semantic PR title
- Valid branch name
- Reasonable size (<1000 lines)
- Linked to issue

## 🔄 Continuous Improvement

The CI/CD setup is designed to:
- **Scale**: Easy to add new modules and workflows
- **Adapt**: Intelligent path-based triggering
- **Maintain Quality**: Enforce code and documentation standards
- **Security First**: Multi-layered security scanning
- **Fast Feedback**: Parallel job execution
- **Developer Friendly**: Local testing support

As the project evolves from modular monolith to microservices, the workflows will adapt while maintaining quality gates and security standards.

---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

