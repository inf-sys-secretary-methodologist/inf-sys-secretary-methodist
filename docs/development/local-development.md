# 💻 Local Development Setup

Guide for setting up the development environment for the Information System project.

## 📋 Prerequisites

- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: Latest
- **PostgreSQL**: 15+ (or use Docker)
- **Git**: Latest

## 🚀 Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/your-org/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist
```

### 2. Install Development Tools
```bash
# Install markdown linting (for documentation)
npm install -g markdownlint-cli2

# Install link checker (for documentation)
npm install -g markdown-link-check

# Install spell checker (for documentation)
npm install -g cspell
```

## 🔍 Quality Assurance Tools

### Markdown Documentation Quality
```bash
# Check markdown formatting
markdownlint-cli2 --config .markdownlint.json "docs/**/*.md" README.md

# Check documentation links
find docs -name "*.md" -exec markdown-link-check {} --config .github/markdown-link-config.json \;
markdown-link-check README.md --config .github/markdown-link-config.json

# Spell check documentation
find docs -name "*.md" -exec cspell {} --no-progress \;
```

### Configuration Files
- **`.markdownlint.json`** - Markdown linting rules (relaxed for technical docs)
- **`.github/markdown-link-config.json`** - Link checking configuration with ignore patterns
- **`.github/wordlist.txt`** - Technical vocabulary for spell checking

## 🔄 CI/CD Workflows

### Main CI/CD Pipeline (`.github/workflows/ci.yml`)
- **Triggers**: Changes to `services/`, `frontend/`, Docker files
- **Jobs**: Backend tests, Frontend tests, Security scan, Build & Deploy
- **Smart Skipping**: Automatically skips jobs when code doesn't exist yet

### Documentation CI (`.github/workflows/docs.yml`)
- **Triggers**: Changes to `docs/`, `README.md`, workflow files
- **Jobs**: Markdown linting, Link checking, Structure validation, Spell checking
- **Quality Gates**: Ensures documentation quality and consistency

### Local Testing Before Push
```bash
# Test documentation quality locally
markdownlint-cli2 --config .markdownlint.json README.md
markdown-link-check README.md --config .github/markdown-link-config.json

# Check git status
git status
git add .
git commit -m "your commit message"
git push
```