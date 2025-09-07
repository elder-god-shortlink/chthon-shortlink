#!/bin/bash

# Chthon Short Link - Development Setup Script

set -e

echo "🚀 Setting up Chthon Short Link development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_requirements() {
    print_status "Checking requirements..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION found"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    print_success "Docker found"
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose."
        exit 1
    fi
    print_success "Docker Compose found"
    
    # Check Make
    if ! command -v make &> /dev/null; then
        print_warning "Make is not installed. You can still use go commands directly."
    else
        print_success "Make found"
    fi
}

# Setup environment file
setup_env() {
    print_status "Setting up environment file..."
    
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "Created .env file from .env.example"
            print_warning "Please review and update .env file with your configuration"
        else
            print_error ".env.example file not found"
            exit 1
        fi
    else
        print_warning ".env file already exists, skipping..."
    fi
}

# Download Go dependencies
setup_dependencies() {
    print_status "Downloading Go dependencies..."
    go mod download
    go mod tidy
    print_success "Dependencies downloaded"
}

# Generate any required files
generate_files() {
    print_status "Generating required files..."
    
    # Generate mocks if mockgen is available
    if command -v mockgen &> /dev/null; then
        go generate ./...
        print_success "Generated mocks"
    else
        print_warning "mockgen not found, skipping mock generation"
        print_status "Install mockgen: go install github.com/golang/mock/mockgen@latest"
    fi
}

# Setup database
setup_database() {
    print_status "Setting up database..."
    
    # Start only database services
    docker-compose up -d postgres redis mongodb
    
    print_status "Waiting for databases to be ready..."
    sleep 10
    
    # Check if databases are running
    if docker-compose ps postgres | grep -q "Up"; then
        print_success "PostgreSQL is running"
    else
        print_error "Failed to start PostgreSQL"
        exit 1
    fi
    
    if docker-compose ps redis | grep -q "Up"; then
        print_success "Redis is running"
    else
        print_error "Failed to start Redis"
        exit 1
    fi
    
    if docker-compose ps mongodb | grep -q "Up"; then
        print_success "MongoDB is running"
    else
        print_error "Failed to start MongoDB"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_status "Running tests..."
    go test ./... -v
    print_success "All tests passed"
}

# Build services
build_services() {
    print_status "Building services..."
    make build
    print_success "All services built successfully"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    mkdir -p bin
    mkdir -p logs
    mkdir -p data
    print_success "Directories created"
}

# Main setup function
main() {
    echo "============================================"
    echo "🔗 Chthon Short Link Development Setup"
    echo "============================================"
    echo
    
    check_requirements
    echo
    
    create_directories
    echo
    
    setup_env
    echo
    
    setup_dependencies
    echo
    
    generate_files
    echo
    
    setup_database
    echo
    
    run_tests
    echo
    
    build_services
    echo
    
    print_success "🎉 Development environment setup completed!"
    echo
    echo "Next steps:"
    echo "1. Review and update .env file if needed"
    echo "2. Start all services: make up"
    echo "3. View logs: make logs"
    echo "4. Access API Gateway at: http://localhost:8080"
    echo
    echo "Available commands:"
    echo "  make help     - Show all available commands"
    echo "  make dev      - Start development environment"
    echo "  make test     - Run tests"
    echo "  make build    - Build all services"
    echo
    echo "Happy coding! 🚀"
}

# Run main function
main "$@"
