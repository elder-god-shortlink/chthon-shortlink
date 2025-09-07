# Chthon ShortLink Microservices

[![Build Status](https://github.com/chthon/shortlink/actions/workflows/ci.yml/badge.svg)](https://github.com/chthon/shortlink/actions)
[![Code Quality](https://github.com/chthon/shortlink/actions/workflows/quality.yml/badge.svg)](https://github.com/chthon/shortlink/actions)
[![SonarQube](https://github.com/chthon/shortlink/actions/workflows/sonarqube.yml/badge.svg)](https://github.com/chthon/shortlink/actions)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Coverage](https://codecov.io/gh/chthon/shortlink/branch/main/graph/badge.svg)](https://codecov.io/gh/chthon/shortlink)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=chthon-short-link&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=chthon-short-link)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=chthon-short-link&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=chthon-short-link)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=chthon-short-link&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=chthon-short-link)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/chthon/shortlink)](https://goreportcard.com/report/github.com/chthon/shortlink)

> **Enterprise-grade URL shortening microservices platform built with GoLang**

Ứng dụng rút gọn link microservices được xây dựng bằng GoLang với kiến trúc microservices hiện đại, hỗ trợ scale cao và performance tối ưu. Được thiết kế để xử lý hàng triệu request với độ trễ thấp và availability cao.

## ✨ Features

### 🔗 **Core URL Shortening**
- **Multiple Algorithms**: Random, Hash-based, Base64, Timestamp-based code generation
- **Custom Short Codes**: User-defined short codes với validation
- **Bulk Operations**: Mass URL shortening với async processing
- **URL Validation**: Comprehensive URL validation và normalization

### ⚡ **Performance & Scalability**
- **Redis Caching**: Sub-millisecond redirect response time
- **Database Sharding**: Horizontal scaling với PostgreSQL partitioning
- **CDN Integration**: Global content delivery network support
- **Load Balancing**: Automatic service discovery và health checks

### 📊 **Analytics & Monitoring**
- **Real-time Analytics**: Kafka + MongoDB cho tracking clicks realtime
- **Geographic Data**: IP-based country và city detection
- **Device Detection**: Browser, OS, device type identification
- **Custom Metrics**: Prometheus metrics với Grafana dashboards

### 🔐 **Security & Authentication**
- **JWT Authentication**: Secure token-based authentication
- **Role-based Access**: Admin, Premium, User role management  
- **Rate Limiting**: DDoS protection và abuse prevention
- **Input Sanitization**: XSS và injection attack protection

### 🛠️ **DevOps & Deployment**
- **Docker Containers**: Multi-stage builds with security scanning
- **Kubernetes Ready**: Helm charts với auto-scaling
- **CI/CD Pipeline**: GitHub Actions với automated testing
- **Infrastructure as Code**: Terraform configurations

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Web    │    │   Mobile App    │    │   Third Party   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │       API Gateway           │
                    │    (JWT, Rate Limiting)     │
                    └─────────────┬───────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
┌───────▼──────┐         ┌────────▼────────┐       ┌────────▼────────┐
│ Shortlink    │         │ Redirect        │       │ User Management │
│ Service      │         │ Service         │       │ Service         │
└──────────────┘         └─────────────────┘       └─────────────────┘
        │                         │
        │                 ┌───────▼────────┐
        │                 │ Analytics      │
        │                 │ Service        │
        │                 └────────────────┘
        │
┌───────▼──────┐         ┌─────────────────┐       ┌─────────────────┐
│ PostgreSQL   │         │ Redis Cache     │       │ MongoDB         │
└──────────────┘         └─────────────────┘       └─────────────────┘
        │
┌───────▼──────┐
│ Kafka        │
└──────────────┘
```

## Services

1. **API Gateway**: Xác thực, phân tích, routing requests
2. **Shortlink Service**: Tạo và quản lý shortlinks  
3. **Redirect Service**: Xử lý redirect và caching
4. **Analytics Service**: Thu thập và phân tích dữ liệu
5. **User Management Service**: Quản lý người dùng và phân quyền

## 🔧 Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin Web Framework
- **Databases**: 
  - PostgreSQL (Primary data storage)
  - Redis (Caching layer)
  - MongoDB (Analytics data)
- **Message Queue**: Apache Kafka
- **Authentication**: JWT with bcrypt password hashing
- **Monitoring**: Prometheus + Grafana
- **Containerization**: Docker + Docker Compose
- **Testing**: Go built-in testing + Testify
- **Quality Assurance**: SonarQube analysis, Go vet, golangci-lint
- **Dependency Management**: Dependabot automated updates

## 🛡️ Quality Assurance & Code Quality

### **Automated Testing**
- **Unit Tests**: Comprehensive test coverage với mocking
- **Integration Tests**: Database và service integration testing
- **Benchmark Tests**: Performance testing với Go benchmarks
- **End-to-End Tests**: Full API workflow testing

### **Code Quality Tools**
- **SonarQube**: Continuous code quality analysis
- **golangci-lint**: Advanced Go linting với 50+ linters
- **Go vet**: Static analysis tool từ Go toolchain
- **Security Scanning**: Gosec cho security vulnerability detection

### **CI/CD Pipeline**
- **GitHub Actions**: Automated testing trên mỗi commit
- **Quality Gates**: SonarQube quality gates enforcement
- **Dependency Updates**: Dependabot automated dependency management
- **Multi-platform Testing**: Linux, Windows, macOS compatibility

### **Monitoring & Observability**
- **Prometheus Metrics**: Custom metrics cho business KPIs
- **Distributed Tracing**: Request tracing across microservices  
- **Health Checks**: Automated health monitoring
- **Log Aggregation**: Structured logging với ELK stack

## 📦 Services

### 1. API Gateway (Port: 8080)
- **Purpose**: Single entry point, authentication, rate limiting
- **Features**: JWT middleware, CORS, request routing, load balancing
- **Dependencies**: All services

### 2. Shortlink Service (Port: 8081)  
- **Purpose**: Core URL shortening logic
- **Features**: Multiple generation algorithms, CRUD operations, validation
- **Database**: PostgreSQL
- **API**: REST endpoints for shortlink management

### 3. Redirect Service (Port: 8082)
- **Purpose**: Fast URL redirects with caching
- **Features**: Redis caching, click tracking, performance optimization
- **Database**: Redis (primary), PostgreSQL (fallback)
- **Performance**: < 5ms average response time

### 4. Analytics Service (Port: 8084)
- **Purpose**: Real-time analytics and reporting  
- **Features**: Kafka consumer, click tracking, statistics aggregation
- **Database**: MongoDB
- **Integration**: Receives events via Kafka

### 5. User Management Service (Port: 8085)
- **Purpose**: User authentication and authorization
- **Features**: Registration, login, RBAC, user profiles
- **Database**: PostgreSQL
- **Security**: bcrypt password hashing, JWT tokens

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)

### 1. Clone Repository

```bash
git clone https://github.com/chthon/shortlink.git
cd chthon-short-link
```

### 2. Environment Setup

```bash
# Copy environment file
cp .env.example .env

# Edit configuration
nano .env
```

### 3. Run with Docker Compose (Recommended)

```bash
# Start all services with infrastructure
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### 4. Run Locally for Development

```bash
# Install dependencies
go mod download

# Start infrastructure only
docker-compose up -d postgres redis mongodb kafka

# Build all services
make build

# Run each service in separate terminals
./bin/api-gateway.exe
./bin/shortlink-service.exe  
./bin/redirect-service.exe
./bin/analytics-service.exe
./bin/user-management-service.exe
```

## 🧪 Testing

### Run All Tests
```bash
# Unit tests
go test ./...

# Integration tests  
go test -tags=integration ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark tests
go test -bench=. -benchmem ./...
```

### Quality Checks
```bash
# Run all quality checks
make check

# Individual checks
make lint          # golangci-lint
make security      # gosec security scan
make test-coverage # Coverage với threshold check
```

## 👨‍💻 Development Workflow

### **Setup Development Environment**
```bash
# Quick setup script
chmod +x scripts/setup-dev.sh
./scripts/setup-dev.sh

# Manual setup
make deps          # Install dependencies
make build         # Build all services
make dev-up        # Start development environment
```

### **Code Quality Standards**
- **Pre-commit Hooks**: Automated formatting và linting
- **Test Coverage**: Minimum 80% coverage requirement
- **Documentation**: Comprehensive API documentation
- **Code Review**: Required approval từ maintainers

### **Branch Strategy**
- `main`: Production-ready code
- `develop`: Integration branch cho features
- `feature/*`: Feature development branches
- `hotfix/*`: Emergency production fixes

### **Dependency Management**
- **Dependabot**: Automated weekly dependency updates
- **Security Updates**: Immediate security patch application
- **Version Pinning**: Controlled dependency upgrades

### Manual API Testing
```bash
# Health check
curl http://localhost:8080/health

# Create user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Create shortlink
curl -X POST http://localhost:8080/api/v1/shortlinks \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com","title":"Test Link"}'
```

## 📊 API Documentation

### Service Endpoints

| Service | Port | Health Check | Swagger UI |
|---------|------|--------------|------------|
| API Gateway | 8080 | `/health` | `/swagger` |
| Shortlink | 8081 | `/health` | `/swagger` |
| Redirect | 8082 | `/health` | `/swagger` |
| Analytics | 8084 | `/health` | `/swagger` |
| User Management | 8085 | `/health` | `/swagger` |

### Key API Endpoints

#### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

#### Shortlinks  
- `POST /api/v1/shortlinks` - Create shortlink
- `GET /api/v1/shortlinks` - List user's shortlinks
- `GET /api/v1/shortlinks/{id}` - Get shortlink details
- `PUT /api/v1/shortlinks/{id}` - Update shortlink
- `DELETE /api/v1/shortlinks/{id}` - Delete shortlink

#### Redirects
- `GET /{code}` - Redirect to original URL
- `GET /api/v1/redirect/{code}/preview` - Preview shortlink

#### Analytics
- `GET /api/v1/analytics/{code}` - Get click statistics

## 🐳 Production Deployment

### **Docker Compose Production**
```bash
# Setup production environment
chmod +x scripts/setup-prod.sh
sudo ./scripts/setup-prod.sh

# Production deployment
docker-compose -f docker-compose.prod.yml up -d

# Scale services
docker-compose -f docker-compose.prod.yml up -d --scale shortlink-service=3

# Health check
./scripts/health-check.sh -d yourdomain.com
```

### **Kubernetes Deployment**
```bash
# Apply configurations
kubectl apply -f k8s/

# Check status
kubectl get pods -n shortlink

# Scale deployment
kubectl scale deployment shortlink-service --replicas=5
```

### **Production Checklist**
- [ ] Environment variables configured
- [ ] SSL certificates installed
- [ ] Database migrations applied
- [ ] Monitoring dashboards setup
- [ ] Backup scripts configured
- [ ] Log aggregation enabled

## 📈 Monitoring & Observability

### **Health Monitoring**
```bash
# Automated health checks
./scripts/health-check.sh

# Service status
systemctl status chthon-shortlink

# Performance metrics
curl http://localhost:9090/metrics
```

### **Dashboards**
- **Prometheus**: http://localhost:9090 - Metrics collection
- **Grafana**: http://localhost:3000 - Visualization dashboards
- **Jaeger**: http://localhost:16686 - Distributed tracing
- **ELK Stack**: http://localhost:5601 - Log analysis

### **Key Metrics**
- **Request Rate**: RPS across all services
- **Response Time**: P95, P99 latency metrics
- **Error Rate**: 4xx/5xx error percentages
- **Database Performance**: Query execution times
- **Cache Hit Rate**: Redis cache performance

## 🔧 Configuration Management

### **Environment Variables**
```bash
# Core Configuration
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
JWT_SECRET=your-secret-key

# Feature Flags
RATE_LIMIT_ENABLED=true
ANALYTICS_ENABLED=true
CACHE_ENABLED=true

# Monitoring
METRICS_ENABLED=true
LOG_LEVEL=info
```

### **Security Configuration**
- **CORS**: Configurable allowed origins
- **Rate Limiting**: Adjustable limits per user/IP
- **JWT**: Configurable expiration times
- **HTTPS**: TLS 1.3 support

## 🤝 Contributing

### **Development Setup**
1. Fork repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Make changes với test coverage
4. Run quality checks: `make check`
5. Commit changes: `git commit -m 'Add amazing feature'`
6. Push to branch: `git push origin feature/amazing-feature`
7. Create Pull Request

### **Code Standards**
- **Go**: Follow effective Go conventions
- **Testing**: Minimum 80% test coverage
- **Documentation**: Comment public APIs
- **Security**: Follow OWASP guidelines

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Community** - For excellent tooling và libraries
- **Docker** - For containerization platform
- **Kubernetes** - For orchestration capabilities
- **Open Source Contributors** - For amazing dependencies
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `JWT_EXPIRES_IN` | JWT expiration time | `24h` |

### Service Configuration

Each service can be configured via environment variables or config files in `/configs` directory.

## 📈 Monitoring & Observability

### Metrics
- **Prometheus**: Available at `http://localhost:9090`
- **Grafana**: Available at `http://localhost:3000` (admin/admin)

### Logs
```bash
# Service logs
docker-compose logs -f [service-name]

# Centralized logging
docker-compose logs -f | grep ERROR
```

### Health Checks
All services expose health check endpoints for monitoring and load balancer configuration.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Write tests for new features
- Follow Go coding standards
- Update documentation for API changes
- Ensure all CI checks pass

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/) for database ORM
- [Redis](https://redis.io/) for caching
- [Apache Kafka](https://kafka.apache.org/) for message streaming
- [MongoDB](https://www.mongodb.com/) for analytics storage
