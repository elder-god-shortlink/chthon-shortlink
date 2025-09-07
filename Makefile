# Chthon ShortLink Microservices Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
API_GATEWAY_BINARY=bin/api-gateway
SHORTLINK_BINARY=bin/shortlink-service
REDIRECT_BINARY=bin/redirect-service
ANALYTICS_BINARY=bin/analytics-service
USER_MANAGEMENT_BINARY=bin/user-management-service

# Docker compose
DOCKER_COMPOSE=docker-compose

.PHONY: all build clean test deps up down logs help

# Default target
all: deps build

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build all services"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run all tests"
	@echo "  deps       - Download dependencies"
	@echo "  up         - Start all services with Docker Compose"
	@echo "  down       - Stop all services"
	@echo "  logs       - Show logs from all services"
	@echo "  build-api-gateway    - Build API Gateway service"
	@echo "  build-shortlink      - Build Shortlink service"
	@echo "  build-redirect       - Build Redirect service"
	@echo "  build-analytics      - Build Analytics service"
	@echo "  build-user-mgmt      - Build User Management service"
	@echo "  run-api-gateway      - Run API Gateway service locally"
	@echo "  run-shortlink        - Run Shortlink service locally"
	@echo "  run-redirect         - Run Redirect service locally"
	@echo "  dev                  - Start development environment"

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build all services
build: build-api-gateway build-shortlink build-redirect build-analytics build-user-mgmt

# Create bin directory
bin:
	mkdir -p bin

# Build individual services
build-api-gateway: bin
	$(GOBUILD) -o $(API_GATEWAY_BINARY) ./services/api-gateway

build-shortlink: bin
	$(GOBUILD) -o $(SHORTLINK_BINARY) ./services/shortlink

build-redirect: bin
	$(GOBUILD) -o $(REDIRECT_BINARY) ./services/redirect

build-analytics: bin
	$(GOBUILD) -o $(ANALYTICS_BINARY) ./services/analytics

build-user-mgmt: bin
	$(GOBUILD) -o $(USER_MANAGEMENT_BINARY) ./services/user-management

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Docker Compose commands
up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f

# Development environment (infrastructure only)
dev:
	$(DOCKER_COMPOSE) up -d postgres redis mongodb kafka zookeeper prometheus grafana

# Run services locally (for development)
run-api-gateway: build-api-gateway
	./$(API_GATEWAY_BINARY)

run-shortlink: build-shortlink
	./$(SHORTLINK_BINARY)

run-redirect: build-redirect
	./$(REDIRECT_BINARY)

run-analytics: build-analytics
	./$(ANALYTICS_BINARY)

run-user-mgmt: build-user-mgmt
	./$(USER_MANAGEMENT_BINARY)

# Database migration
migrate-up:
	@echo "Running database migrations..."
	docker exec shortlink_postgres psql -U shortlink_user -d shortlink -f /docker-entrypoint-initdb.d/init-db.sql

migrate-down:
	@echo "Rolling back database migrations..."
	# Add rollback scripts here

# Generate documentation
docs:
	@echo "Generating API documentation..."
	# Add swagger generation here

# Linting
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Security scan
security:
	gosec ./...

# Build for production (with optimizations)
build-prod: clean
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="-w -s" -o $(API_GATEWAY_BINARY) ./services/api-gateway
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="-w -s" -o $(SHORTLINK_BINARY) ./services/shortlink
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="-w -s" -o $(REDIRECT_BINARY) ./services/redirect
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="-w -s" -o $(ANALYTICS_BINARY) ./services/analytics
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="-w -s" -o $(USER_MANAGEMENT_BINARY) ./services/user-management

# Docker build
docker-build:
	docker build -t chthon/api-gateway -f services/api-gateway/Dockerfile .
	docker build -t chthon/shortlink-service -f services/shortlink/Dockerfile .
	docker build -t chthon/redirect-service -f services/redirect/Dockerfile .
	docker build -t chthon/analytics-service -f services/analytics/Dockerfile .
	docker build -t chthon/user-management-service -f services/user-management/Dockerfile .

# Install development tools
install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec
	$(GOGET) -u github.com/swaggo/swag/cmd/swag

# Quick setup for new developers
setup: install-tools deps build dev
	@echo "🚀 Development environment is ready!"
	@echo "✅ Dependencies installed"
	@echo "✅ Services built"
	@echo "✅ Infrastructure started"
	@echo ""
	@echo "Next steps:"
	@echo "1. Copy .env.example to .env and configure"
	@echo "2. Run 'make run-api-gateway' in one terminal"
	@echo "3. Run other services as needed"
	@echo ""
	@echo "Available endpoints:"
	@echo "- API Gateway: http://localhost:8080"
	@echo "- PostgreSQL: localhost:5432"
	@echo "- Redis: localhost:6379"
	@echo "- MongoDB: localhost:27017"
	@echo "- Kafka: localhost:9092"
	@echo "- Prometheus: http://localhost:9090"
	@echo "- Grafana: http://localhost:3000"

# Health check all services
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health || echo "❌ API Gateway down"
	@curl -s http://localhost:8082/health || echo "❌ Shortlink Service down"
	@curl -s http://localhost:8083/health || echo "❌ Redirect Service down"
	@curl -s http://localhost:8084/health || echo "❌ Analytics Service down"
	@curl -s http://localhost:8085/health || echo "❌ User Management Service down"
