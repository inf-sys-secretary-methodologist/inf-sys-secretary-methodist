# Information System of the Academic Secretary/Methodologist

![CI/CD Pipeline](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/ci.yml/badge.svg)
![Documentation CI](https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/actions/workflows/docs.yml/badge.svg)

## 🎯 Main Features
* **Document Management**: Complete lifecycle from creation to archiving
* **Workflow Automation**: Approval processes with role-based routing
* **Schedule Management**: Academic planning and resource optimization
* **Reporting & Analytics**: Comprehensive business intelligence
* **Integration Support**: Seamless connection with 1C and external systems

## 🏗️ Architecture
Built on **modular monolith** principles with:
- Domain-Driven Design (DDD)
- Clean Architecture patterns
- Event-driven communication
- Microservices-ready structure

## 🚀 Getting Started

### Prerequisites
- **Go** 1.21+
- **Node.js** 18+
- **PostgreSQL** 15+
- **Redis** 7+
- **Docker** & Docker Compose

### Quick Start
```bash
# Clone repository
git clone https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist.git
cd inf-sys-secretary-methodist

# Start with Docker
docker-compose up -d

# Or manual setup
make setup
make dev
```

## 📚 Documentation

### 🏃‍♂️ Quick Start for Developers
- [📖 Development Guide](docs/development/development-guide.md) - **Start here!** Complete handbook
- [🧩 Clean Code & Patterns](docs/development/clean-code-patterns.md) - Senior-level practices & SOLID
- [🔀 Pull Request Guide](docs/development/pull-request-guide.md) - PR process and templates

### 🏗️ Architecture & Design
- [🎯 Project Overview](docs/project-overview.md) - Goals, users, and requirements
- [🧩 Modular Architecture](docs/architecture/modular-architecture.md) - DDD-based modular design
- [🚀 Microservices Migration Guide](docs/architecture/microservices-migration-guide.md) - Migration strategy
- [📄 Document Lifecycle & Workflows](docs/workflows/document-lifecycle.md) - Business processes

### 📋 Development Process
- [🔧 Git Terminal Guide](docs/development/git-terminal-guide.md) - Complete Git commands reference
- [🔄 Sprint Management](docs/development/sprint-management.md) - Agile workflow with GitHub Projects
- [🧪 Testing Strategy](docs/development/testing-strategy.md) - Testing approach
- [🔄 CI/CD Workflows](docs/development/ci-cd-workflows.md) - Automation pipeline

### 🚀 Deployment & Operations
- [🐳 Docker Setup](docs/deployment/docker-setup.md) - Containerization guide
- [☁️ Infrastructure](docs/deployment/infrastructure.md) - Cloud architecture and scaling
- [🔧 Environment Configuration](docs/deployment/environment.md) - Environment variables
- [☁️ Production Deployment](docs/deployment/production-deploy.md) - Production setup
- [🛡️ Security Guidelines](docs/security/security-guidelines.md) - Security framework

### 📊 API & Integration
- [📖 REST API Reference](docs/api/api-documentation.md) - Complete API docs
- [🔐 Authentication](docs/api/authentication.md) - Auth endpoints
- [📄 Document Management API](docs/api/documents.md) - Document endpoints
- [👥 Roles and Permissions](docs/users/roles-and-permissions.md) - User roles and access control

## 🏗️ Technology Stack

### Backend (Modular Monolith)
- **Language**: Go 1.21+
- **Architecture**: DDD + Clean Architecture
- **Framework**: Gin + gRPC (ready for microservices)
- **Database**: PostgreSQL 15+ (Primary), Redis (Cache/Sessions)
- **Messaging**: Apache Kafka (Event-driven)
- **Authentication**: OAuth 2.0 + JWT
- **Patterns**: Repository, CQRS, Event Sourcing

### Frontend
- **Framework**: Next.js 14 + TypeScript
- **UI Library**: MUI (Material-UI) + Tailwind CSS
- **State Management**: Zustand
- **Testing**: Jest + React Testing Library + Playwright

### Infrastructure
- **Orchestration**: Kubernetes (GKE/AKS)
- **Monitoring**: Prometheus + Grafana + Jaeger
- **Logging**: ELK Stack (Elasticsearch + Logstash + Kibana)
- **CI/CD**: GitLab CI/CD
- **Load Balancer**: Nginx + CloudFlare

## 🤝 Contributing

1. Read our [Development Guidelines](docs/development/)
2. Check the [Sprint Management Guide](docs/development/sprint-management.md)
3. Follow our [Pull Request Process](docs/development/pull-request-guide.md)

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
*"Enhancing organizational efficiency through automation and methodological support."*