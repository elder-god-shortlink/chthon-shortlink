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

## � Table of Contents

1. [Features](#-features)
2. [Architecture](#️-architecture)
3. [Tech Stack](#-tech-stack)
4. [Services Overview](#-services)
5. [Quick Start](#-quick-start)
6. [API Documentation](#-api-documentation)
7. [Postman Collection](#-postman-collection)
8. [Cross-Platform Scripts](#-automation-scripts)
9. [Troubleshooting](#-troubleshooting)
10. [Development](#-development)
11. [Deployment](#-deployment)

## �🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Infrastructure Services**: PostgreSQL, Redis, MongoDB, Kafka (running on network)

### 1. Clone Repository

```bash
git clone https://github.com/chthon/shortlink.git
cd chthon-shortlink
```

### 2. Environment Setup

```bash
# Copy environment file
cp .env.example .env

# Edit configuration with your infrastructure details
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

### 4. Run Locally (Native Go Execution)

#### **Windows (PowerShell)**

```powershell
# Install dependencies
go mod download

# Build and run all services
.\run-all-services.ps1

# Stop all services
.\stop-all-services.ps1
```

#### **macOS/Linux (Bash)**

```bash
# Install dependencies
go mod download

# Make scripts executable (first time only)
chmod +x run-all-services.sh stop-all-services.sh

# Build and run all services
./run-all-services.sh

# Stop all services  
./stop-all-services.sh

# Stop with force (immediate kill)
./stop-all-services.sh --force
```

### 5. Manual Service Management

If you prefer to run services individually:

```bash
# Build all services
go build -o bin/api-gateway ./services/api-gateway/main.go
go build -o bin/shortlink-service ./services/shortlink/main.go
go build -o bin/redirect-service ./services/redirect/main.go
go build -o bin/analytics-service ./services/analytics/main.go
go build -o bin/user-management-service ./services/user-management/main.go

# Run services in separate terminals (order matters)
./bin/user-management-service     # Port 8084
./bin/shortlink-service          # Port 8081  
./bin/redirect-service           # Port 8082
./bin/analytics-service          # Port 8083
./bin/api-gateway               # Port 8080 (last)
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

#### **Automated Scripts**

The project includes cross-platform automation scripts for easy development:

**Windows (PowerShell)**
- `run-all-services.ps1` - Build and start all services with health monitoring
- `stop-all-services.ps1` - Gracefully stop all running services
- `create-database.ps1` - Database creation instructions

**macOS/Linux (Bash)**  
- `run-all-services.sh` - Build and start all services with health monitoring
- `stop-all-services.sh` - Gracefully stop all running services with options:
  - `--force` - Force kill all processes immediately
  - `--help` - Show usage information

#### **Script Features**

🔧 **Automated Building**: Compiles all services from source
🏥 **Health Monitoring**: Continuous service health checks
🔄 **Dependency Management**: Verifies infrastructure connectivity
📊 **Real-time Logging**: Service logs in dedicated log files
🛡️ **Graceful Shutdown**: Proper SIGTERM handling with fallback force kill
📍 **Port Management**: Automatic port conflict detection and resolution
🎯 **Service Discovery**: Automatic service registration and health reporting

#### **Manual Setup**
```bash
# Install dependencies
go mod download

# Build all services
make build         

# Start development environment
make dev-up        
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

## 🌐 Service URLs & Endpoints

Once all services are running, you can access them at:

### **API Gateway** (Port 8080)
- **Main API**: `http://localhost:8080`
- **Swagger UI**: `http://localhost:8080/docs/swagger/index.html`
- **Health Check**: `http://localhost:8080/health`
- **API Documentation**: `http://localhost:8080/docs/info`

### **Individual Services**
- **User Management**: `http://localhost:8084/health`
- **Shortlink Service**: `http://localhost:8081/health` 
- **Redirect Service**: `http://localhost:8082/health`
- **Analytics Service**: `http://localhost:8083/health`

### **API Endpoints**

#### **Authentication**
```bash
# Register new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -F "email=user@example.com" \
  -F "password=password123" \
  -F "role=user"

# Login user  
curl -X POST http://localhost:8080/api/v1/auth/login \
  -F "email=user@example.com" \
  -F "password=password123"

# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -F "refresh_token=YOUR_REFRESH_TOKEN"
```

#### **Shortlink Management**
```bash
# Create shortlink (requires authentication)
curl -X POST http://localhost:8080/api/v1/user/links \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com","title":"Test Link"}'

# Get user's shortlinks
curl -X GET http://localhost:8080/api/v1/user/links \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Update shortlink
curl -X PUT http://localhost:8080/api/v1/user/links/{id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title"}'

# Delete shortlink
curl -X DELETE http://localhost:8080/api/v1/user/links/{id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### **URL Redirection**
```bash
# Redirect via short code (public)
curl -L http://localhost:8080/api/v1/public/{shortCode}
```

#### **Analytics** 
```bash
# Get user analytics
curl -X GET http://localhost:8080/api/v1/user/analytics \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Admin analytics (admin role required)
curl -X GET http://localhost:8080/api/v1/admin/analytics \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### **Health Monitoring**
```bash
# Check all services health
curl http://localhost:8080/health

# Individual service health
curl http://localhost:8081/health  # Shortlink
curl http://localhost:8082/health  # Redirect  
curl http://localhost:8083/health  # Analytics
curl http://localhost:8084/health  # User Management
```

## 📊 API Documentation

### 🧪 Postman Collection
We provide a comprehensive Postman collection for easy API testing:

- **Collection File**: `Chthon-ShortLink-API.postman_collection.json`
- **Environment Files**: 
  - `Chthon-ShortLink-Local.postman_environment.json` (Local development)
  - `Chthon-ShortLink-Production.postman_environment.json` (Production)
- **Detailed Documentation**: See [POSTMAN.md](./POSTMAN.md) for setup and usage

**Features:**
- ✅ All 25+ API endpoints covered with automated tests
- ✅ Automated JWT token management
- ✅ Built-in response validation and error handling
- ✅ Environment variable management for flexible testing
- ✅ Complete workflow testing (Auth → CRUD → Analytics)

**Quick Start:**
```bash
# Import files into Postman
1. Import Chthon-ShortLink-API.postman_collection.json
2. Import environment file (Local or Production)  
3. Select environment and run "Health Check" folder
4. Start testing with "Authentication" → "Register User"
```

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

## � Troubleshooting

### **Common Issues**

#### **Service Won't Start**
```bash
# Check if ports are in use
# Windows
netstat -ano | findstr :8080

# macOS/Linux  
lsof -i :8080

# Kill process using port
# Windows
taskkill /F /PID <process_id>

# macOS/Linux
kill -9 <process_id>
```

#### **Infrastructure Connection Issues**
```bash
# Test database connections
# PostgreSQL
nc -zv 192.168.1.127 5432

# Redis
nc -zv 192.168.1.127 6379

# MongoDB  
nc -zv 192.168.1.127 27017

# Kafka
nc -zv 192.168.1.127 9092
```

#### **Environment Configuration**
```bash
# Check .env file exists
ls -la .env

# Validate environment variables
go run -tags debug services/api-gateway/main.go
```

#### **Service Logs**
```bash
# View service logs (after running scripts)
# Windows
type logs\api-gateway.log

# macOS/Linux
tail -f logs/api-gateway.log
tail -f logs/shortlink.log
tail -f logs/redirect.log
tail -f logs/analytics.log
tail -f logs/user-management.log
```

#### **Build Issues**
```bash
# Clean build cache
go clean -cache -modcache

# Re-download dependencies
rm go.sum
go mod tidy
go mod download

# Rebuild all services
rm -rf bin/
go build -o bin/api-gateway ./services/api-gateway/main.go
# ... repeat for other services
```

#### **Script Execution Issues**

**Windows PowerShell Execution Policy**
```powershell
# Check execution policy
Get-ExecutionPolicy

# Set execution policy (run as Administrator)  
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**macOS/Linux Permissions**
```bash
# Make scripts executable
chmod +x run-all-services.sh stop-all-services.sh

# Check script syntax
bash -n run-all-services.sh
```

### **Performance Issues**

#### **Slow Response Times**
```bash
# Check Redis connectivity
redis-cli -h 192.168.1.127 ping

# Monitor service resource usage
# Windows
wmic process where caption="api-gateway.exe" get processid,pagefile,workingsetsize

# macOS/Linux
ps aux | grep -E "(api-gateway|shortlink|redirect|analytics|user-management)"
```

#### **High CPU/Memory Usage**
```bash
# Profile Go services
go tool pprof http://localhost:8080/debug/pprof/profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Check for memory leaks
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

### **API Testing & Debugging**

#### **Swagger UI Issues**
- Ensure API Gateway is running on port 8080
- Check Swagger docs generation: `swag init -g main.go -o ./docs`
- Verify docs import in main.go: `_ "github.com/chthon/shortlink/services/api-gateway/docs"`

#### **Authentication Issues**
```bash
# Test JWT token generation
curl -X POST http://localhost:8080/api/v1/auth/login \
  -F "email=test@example.com" \
  -F "password=password123" \
  -v

# Verify token in requests
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -v
```

## �🐳 Production Deployment

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

## 🤖 Automation Scripts

### **Cross-Platform Service Management**

The project includes comprehensive automation scripts for both Windows and Unix-like systems:

#### **Windows PowerShell Scripts**
- **`run-all-services.ps1`**: Complete service orchestration with monitoring
- **`stop-all-services.ps1`**: Graceful service shutdown with cleanup

#### **macOS/Linux Bash Scripts**  
- **`run-all-services.sh`**: Full service lifecycle management
- **`stop-all-services.sh`**: Advanced shutdown with multiple options

### **Script Features & Capabilities**

#### 🏗️ **Automated Building**
- Compiles all services from source code
- Validates Go installation and dependencies
- Handles build errors with detailed reporting
- Creates optimized binaries for production

#### 🔍 **Infrastructure Validation**  
- Tests connectivity to PostgreSQL (port 5432)
- Verifies Redis accessibility (port 6379)
- Confirms MongoDB connection (port 27017)
- Validates Kafka broker availability (port 9092)

#### 🚀 **Service Orchestration**
- Starts services in correct dependency order
- Manages port allocation and conflict resolution  
- Implements health check validation
- Provides real-time startup progress reporting

#### 📊 **Health Monitoring**
- Continuous service health verification
- HTTP endpoint health check integration
- Automated failure detection and reporting
- Performance metrics collection

#### 📝 **Logging & Debugging**
- Dedicated log files for each service
- Structured logging with timestamps
- Process ID (PID) tracking and management
- Log rotation and cleanup automation

#### 🛡️ **Graceful Shutdown**
- SIGTERM signal handling for clean shutdowns
- Configurable timeout periods for graceful stops
- Force kill fallback for unresponsive processes
- Complete cleanup of temporary files and PIDs

### **Usage Examples**

#### **Quick Start (All Platforms)**
```bash
# Windows
.\run-all-services.ps1

# macOS/Linux
./run-all-services.sh
```

#### **Advanced Usage**
```bash
# Stop with immediate force kill (Linux/macOS)
./stop-all-services.sh --force

# View help information
./stop-all-services.sh --help

# Monitor logs during execution
tail -f logs/api-gateway.log     # Linux/macOS
Get-Content -Wait logs\api-gateway.log  # Windows
```

#### **Integration with Development Workflow**
```bash
# Development cycle automation
./stop-all-services.sh && \
go mod tidy && \
./run-all-services.sh
```

### **Script Architecture**

#### **Modular Design**
- Reusable function libraries
- Platform-specific adaptations  
- Configurable parameters and timeouts
- Extensible service definitions

#### **Error Handling**
- Comprehensive error catching and reporting
- Rollback mechanisms for failed starts
- Detailed diagnostics and troubleshooting hints
- Exit code management for CI/CD integration

#### **Cross-Platform Compatibility**
- Native PowerShell for Windows environments
- Bash scripting for Unix-like systems
- Consistent API across all platforms
- Platform-specific optimizations

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

## 🛠️ Cross-Platform Automation Scripts

### **Windows (PowerShell)**

- **`run-all-services.ps1`** - Start all services with health monitoring
- **`stop-all-services.ps1`** - Gracefully stop all services

```powershell
# Start all services
.\run-all-services.ps1

# Stop all services  
.\stop-all-services.ps1

# Force stop all services
.\stop-all-services.ps1 -Force
```

### **macOS/Linux (Bash)**

- **`run-all-services.sh`** - Cross-platform service orchestration
- **`stop-all-services.sh`** - Advanced shutdown management

```bash
# Make scripts executable (first time only)
chmod +x *.sh

# Start all services
./run-all-services.sh

# Stop all services gracefully
./stop-all-services.sh

# Force stop all services
./stop-all-services.sh --force
```

### **Script Features**

✅ **Infrastructure Validation** - Checks required tools and dependencies  
✅ **Port Conflict Detection** - Prevents conflicts with existing services  
✅ **Health Monitoring** - Automated health checks with retry logic  
✅ **Graceful Shutdown** - SIGTERM handling with cleanup timeouts  
✅ **Error Handling** - Comprehensive error reporting and recovery  
✅ **Cross-Platform Support** - Works on Windows, macOS, and Linux  

### **Script Output Example**

```bash
🚀 Starting Chthon ShortLink Services...
✅ Infrastructure validation passed
⚙️  Building services...
🔄 Starting API Gateway on port 8080...
🔄 Starting Shortlink Service on port 8081...
🔄 Starting Redirect Service on port 8082...
🔄 Starting Analytics Service on port 8083...
🔄 Starting User Management Service on port 8084...
✅ All services started successfully!
📊 API Gateway: http://localhost:8080
📚 Swagger UI: http://localhost:8080/docs/swagger/index.html
🧪 Postman Collection: Available in project root
```

## 🧪 Postman Collection

### **Comprehensive API Testing Collection**

- **File**: `Chthon-ShortLink-API.postman_collection.json`
- **Environments**: Local development and Production templates  
- **Coverage**: 25+ endpoints with automated testing
- **Documentation**: See [POSTMAN.md](./POSTMAN.md) for detailed guide

### **Key Features**
- ✅ Automated JWT token management
- ✅ Built-in response validation  
- ✅ Environment variable management
- ✅ Complete workflow testing
- ✅ Error handling scenarios
- ✅ Performance monitoring

### **Quick Import**
```bash
1. Import collection: Chthon-ShortLink-API.postman_collection.json
2. Import environment: Chthon-ShortLink-Local.postman_environment.json  
3. Select environment and run tests
```

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
