# 🔄 CI/CD Workflows Configuration

Complete guide to Continuous Integration and Continuous Deployment workflows for the Information System project.

## 📊 Overview

The project uses GitHub Actions for automated testing, documentation validation, and deployment processes through two main workflows:

- **Main CI/CD Pipeline** (`.github/workflows/ci.yml`) - Code testing and deployment
- **Documentation CI** (`.github/workflows/docs.yml`) - Documentation quality assurance

## 🚀 Main CI/CD Pipeline

### File: `.github/workflows/ci.yml`

#### Triggers
```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'services/**'
      - 'frontend/**'
      - '.github/workflows/ci.yml'
      - 'docker-compose.yml'
      - 'Dockerfile*'
  pull_request: # Same paths
```

#### Jobs

##### 1. Backend Testing (`backend-test`)
- **Matrix Strategy**: Tests all 10 microservices in parallel
- **Services**: `auth`, `user`, `document`, `workflow`, `schedule`, `reporting`, `task`, `notification`, `file`, `integration`
- **Smart Execution**: Automatically skips if service directory doesn't exist
- **Steps**:
  1. Check service existence
  2. Setup Go 1.21
  3. Cache Go modules
  4. Install dependencies (`go mod download`)
  5. Run tests with coverage (`go test -v -race -coverprofile=coverage.out ./...`)
  6. Upload coverage to Codecov

##### 2. Frontend Testing (`frontend-test`)
- **Smart Execution**: Skips if `frontend/` directory doesn't exist
- **Steps**:
  1. Check frontend existence
  2. Setup Node.js 18
  3. Install dependencies (`npm ci`)
  4. Run linting (`npm run lint`)
  5. Run type checking (`npm run type-check`)
  6. Run unit tests (`npm run test:unit`)
  7. Run E2E tests (`npm run test:e2e`)

##### 3. Security Scanning (`security-scan`)
- **Tool**: Trivy vulnerability scanner
- **Smart Execution**: Only runs if code directories exist
- **Output**: SARIF format uploaded to GitHub Security tab

##### 4. Build and Push (`build-and-push`)
- **Condition**: Only on `main` branch after successful tests
- **Matrix Strategy**: Builds Docker images for all components
- **Registry**: GitHub Container Registry (ghcr.io)
- **Smart Execution**: Skips missing components

##### 5. Deployment Jobs
- **Staging**: Deploys on `develop` branch
- **Production**: Deploys on `main` branch

### Key Features
- **🧠 Smart Skipping**: Jobs automatically skip when corresponding code doesn't exist
- **⚡ Parallel Execution**: Multiple services tested simultaneously
- **📊 Coverage Reporting**: Automatic code coverage collection
- **🔒 Security First**: Vulnerability scanning on every push
- **🐳 Container Ready**: Automated Docker image building

## 📚 Documentation CI

### File: `.github/workflows/docs.yml`

#### Triggers
```yaml
on:
  push:
    branches: [ main, develop ]
    paths:
      - 'docs/**'
      - 'README.md'
      - '.github/workflows/docs.yml'
  pull_request: # Same paths
```

#### Jobs

##### 1. Markdown Linting (`markdown-lint`)
- **Tool**: markdownlint-cli2
- **Config**: `.markdownlint.json` (relaxed rules for technical docs)
- **Scope**: All `docs/**/*.md` and `README.md`

##### 2. Link Checking (`link-checker`)
- **Tool**: markdown-link-check CLI
- **Config**: `.github/markdown-link-config.json`
- **Features**:
  - Checks all documentation links
  - Ignores example/placeholder URLs
  - Configurable timeout and retry logic

##### 3. Structure Validation (`structure-check`)
- **Purpose**: Ensures required documentation files exist
- **Checks**:
  - Required directories: `docs/api`, `docs/architecture`, etc.
  - Required files: `docs/README.md`, API docs, etc.

##### 4. Spell Checking (`spell-check`)
- **Tool**: cspell
- **Vocabulary**: `.github/wordlist.txt` (5877+ technical terms)
- **Mode**: Non-blocking (warns but doesn't fail CI)

## ⚙️ Configuration Files

### 1. `.markdownlint.json`
Markdown linting configuration with relaxed rules for technical documentation:

```json
{
  "line-length": false,
  "no-trailing-punctuation": false,
  "no-emphasis-as-heading": false,
  // ... other relaxed rules for technical docs
}
```

### 2. `.github/markdown-link-config.json`
Link checking configuration:

```json
{
  "ignorePatterns": [
    {"pattern": "^http://localhost"},
    {"pattern": "^https://api.inf-sys.example.com"},
    // ... other ignore patterns
  ],
  "timeout": "20s",
  "retryOn429": true,
  "retryCount": 3
}
```

### 3. `.github/wordlist.txt`
Technical vocabulary for spell checking (5877+ terms):
- Microservices terminology
- Cloud and infrastructure terms
- Programming languages and frameworks
- API and security terms

## 🔧 Local Development Integration

### Pre-commit Testing
```bash
# Test documentation quality
markdownlint-cli2 --config .markdownlint.json "docs/**/*.md" README.md

# Check links
find docs -name "*.md" -exec markdown-link-check {} --config .github/markdown-link-config.json \;

# Basic spell check
find docs -name "*.md" -exec cspell {} --no-progress \;
```

### Installing Tools
```bash
npm install -g markdownlint-cli2 markdown-link-check cspell
```

## 📈 Workflow Status

### GitHub Actions Badges
```markdown
[![CI/CD Pipeline](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/ci.yml/badge.svg)](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/ci.yml)
[![Documentation CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/docs.yml/badge.svg)](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/docs.yml)
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Backend Tests Failing
- **Cause**: Service directory doesn't exist yet
- **Solution**: CI automatically skips non-existent services

#### 2. Frontend Tests Failing
- **Cause**: Frontend directory doesn't exist yet
- **Solution**: CI automatically skips if `frontend/` missing

#### 3. Documentation Links Failing
- **Cause**: External URLs or placeholder links
- **Solution**: Add patterns to `.github/markdown-link-config.json` ignore list

#### 4. Spell Check Failures
- **Cause**: Technical terms not in wordlist
- **Solution**: Add terms to `.github/wordlist.txt`

### Best Practices

1. **🔍 Test Locally First**: Run quality checks before pushing
2. **📝 Update Documentation**: Keep CI configs documented when changed
3. **🏷️ Use Descriptive Commits**: Help CI skip unnecessary runs
4. **🔧 Monitor Actions**: Check GitHub Actions tab for failures

## 🔄 Continuous Improvement

The CI/CD setup is designed to:
- **Scale**: Easily add new microservices to the matrix
- **Adapt**: Skip jobs intelligently as the project grows
- **Maintain Quality**: Enforce documentation and code standards
- **Deploy Safely**: Separate staging and production environments

As the project evolves, the workflows will automatically adapt to new services and components while maintaining quality gates and deployment safety.