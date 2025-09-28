# Information System of the Academic Secretary/Methodologist

## Main features
* Document management
* Scheduling  
* Reporting
* Communication tools to support efficient organizational operations

## Purpose
The project is created to:
* Enhance productivity
* Ensure consistency in process execution
* Facilitate information exchange within teams or departments

## Getting Started
*Instructions for installation and usage here.*

## 📚 Documentation

### 📋 Project Overview
- [🎯 Project Overview](docs/project-overview.md) - Goals, users, and requirements
- [🏗️ Microservices Architecture](docs/architecture/microservices-architecture.md) - System design and services
- [👥 Roles and Permissions](docs/users/roles-and-permissions.md) - User roles and access control

### 🔄 Business Logic
- [📄 Document Lifecycle & Workflows](docs/workflows/document-lifecycle.md) - Business processes and automation

### 🚀 Development
- [💻 Local Development Setup](docs/development/local-development.md) - Environment setup
- [⚙️ Coding Standards](docs/development/coding-standards.md) - Go & Next.js best practices
- [🧪 Testing Strategy](docs/development/testing-strategy.md) - Comprehensive testing approach
- [🔄 Sprint Management Guide](docs/development/sprint-management.md) - GitHub Projects workflow
- [🔀 Pull Request Guidelines](docs/development/pull-request-guide.md) - PR standards and process

### 🚀 Deployment & Infrastructure
- [☁️ Infrastructure](docs/deployment/infrastructure.md) - Cloud architecture and scaling
- [🐳 Docker Setup](docs/deployment/docker-setup.md) - Containerization
- [🔧 Environment Configuration](docs/deployment/environment.md) - Environment variables
- [☁️ Production Deployment](docs/deployment/production-deploy.md) - Production setup

### 🔒 Security
- [🛡️ Security Guidelines](docs/security/security-guidelines.md) - Comprehensive security framework

### 📊 API Documentation
- [📖 REST API Reference](docs/api/api-documentation.md) - Complete API docs
- [🔐 Authentication](docs/api/authentication.md) - Auth endpoints
- [📄 Document Management API](docs/api/documents.md) - Document endpoints

## 🏗️ Technology Stack

### Backend (Микросервисы)
- **Language**: Go 1.21+
- **Framework**: Gin + gRPC
- **Database**: PostgreSQL 15+ (Primary), Redis (Cache/Sessions)
- **Message Queue**: Apache Kafka
- **Authentication**: OAuth 2.0 + JWT
- **Service Discovery**: Consul
- **Secrets Management**: HashiCorp Vault

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