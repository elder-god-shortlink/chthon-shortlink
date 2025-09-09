#!/bin/bash

# Chthon ShortLink - Run All Services Script (macOS/Linux)
# This script starts all microservices in the correct order with health checks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICES_DIR="$(dirname "$0")"
BIN_DIR="$SERVICES_DIR/bin"
LOG_DIR="$SERVICES_DIR/logs"
MAX_RETRIES=30
RETRY_INTERVAL=2

# Create logs directory if it doesn't exist
mkdir -p "$LOG_DIR"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] $message${NC}"
}

# Function to check if a service is healthy
check_service_health() {
    local service_name=$1
    local health_url=$2
    local retries=0
    
    print_status $YELLOW "Checking health of $service_name..."
    
    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -s -f "$health_url" > /dev/null 2>&1; then
            print_status $GREEN "$service_name is healthy ✓"
            return 0
        fi
        
        retries=$((retries + 1))
        if [ $retries -lt $MAX_RETRIES ]; then
            print_status $YELLOW "$service_name not ready, retrying in ${RETRY_INTERVAL}s... ($retries/$MAX_RETRIES)"
            sleep $RETRY_INTERVAL
        fi
    done
    
    print_status $RED "$service_name health check failed after $MAX_RETRIES attempts ✗"
    return 1
}

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        print_status $RED "Port $port is already in use"
        return 1
    fi
    return 0
}

# Function to wait for port to be available
wait_for_port() {
    local port=$1
    local service_name=$2
    local retries=0
    
    while [ $retries -lt $MAX_RETRIES ]; do
        if lsof -i :$port > /dev/null 2>&1; then
            print_status $GREEN "$service_name is listening on port $port ✓"
            return 0
        fi
        
        retries=$((retries + 1))
        if [ $retries -lt $MAX_RETRIES ]; then
            sleep 1
        fi
    done
    
    print_status $RED "$service_name failed to start on port $port ✗"
    return 1
}

# Function to build services
build_services() {
    print_status $BLUE "Building all services..."
    
    cd "$SERVICES_DIR"
    
    # Build each service
    print_status $YELLOW "Building API Gateway..."
    go build -o bin/api-gateway ./services/api-gateway/main.go
    
    print_status $YELLOW "Building Shortlink Service..."
    go build -o bin/shortlink-service ./services/shortlink/main.go
    
    print_status $YELLOW "Building Redirect Service..."
    go build -o bin/redirect-service ./services/redirect/main.go
    
    print_status $YELLOW "Building Analytics Service..."
    go build -o bin/analytics-service ./services/analytics/main.go
    
    print_status $YELLOW "Building User Management Service..."
    go build -o bin/user-management-service ./services/user-management/main.go
    
    print_status $GREEN "All services built successfully ✓"
}

# Function to check infrastructure connectivity
check_infrastructure() {
    print_status $BLUE "Checking infrastructure connectivity..."
    
    # Check PostgreSQL
    if ! nc -z 192.168.1.127 5432 2>/dev/null; then
        print_status $RED "PostgreSQL is not accessible at 192.168.1.127:5432 ✗"
        return 1
    fi
    print_status $GREEN "PostgreSQL connection OK ✓"
    
    # Check Redis
    if ! nc -z 192.168.1.127 6379 2>/dev/null; then
        print_status $RED "Redis is not accessible at 192.168.1.127:6379 ✗"
        return 1
    fi
    print_status $GREEN "Redis connection OK ✓"
    
    # Check MongoDB
    if ! nc -z 192.168.1.127 27017 2>/dev/null; then
        print_status $RED "MongoDB is not accessible at 192.168.1.127:27017 ✗"
        return 1
    fi
    print_status $GREEN "MongoDB connection OK ✓"
    
    # Check Kafka
    if ! nc -z 192.168.1.127 9092 2>/dev/null; then
        print_status $RED "Kafka is not accessible at 192.168.1.127:9092 ✗"
        return 1
    fi
    print_status $GREEN "Kafka connection OK ✓"
    
    return 0
}

# Function to start a service
start_service() {
    local service_name=$1
    local binary_path=$2
    local port=$3
    local health_endpoint=$4
    
    print_status $BLUE "Starting $service_name..."
    
    # Check if port is available
    if ! check_port $port; then
        print_status $YELLOW "Killing existing process on port $port..."
        local pid=$(lsof -ti :$port 2>/dev/null || true)
        if [ ! -z "$pid" ]; then
            kill -9 $pid 2>/dev/null || true
            sleep 2
        fi
    fi
    
    # Start the service in background
    cd "$SERVICES_DIR"
    nohup "$binary_path" > "$LOG_DIR/$service_name.log" 2>&1 &
    local service_pid=$!
    
    # Save PID for later cleanup
    echo $service_pid > "$LOG_DIR/$service_name.pid"
    
    # Wait for service to start
    if ! wait_for_port $port "$service_name"; then
        print_status $RED "Failed to start $service_name"
        return 1
    fi
    
    # Check health if endpoint provided
    if [ ! -z "$health_endpoint" ]; then
        if ! check_service_health "$service_name" "$health_endpoint"; then
            print_status $RED "Failed to start $service_name (health check failed)"
            return 1
        fi
    fi
    
    print_status $GREEN "$service_name started successfully (PID: $service_pid) ✓"
    return 0
}

# Function to cleanup on exit
cleanup() {
    print_status $YELLOW "Cleaning up..."
    if [ -f "$LOG_DIR/cleanup_needed" ]; then
        rm -f "$LOG_DIR/cleanup_needed"
    fi
}

trap cleanup EXIT

# Main execution
main() {
    print_status $BLUE "============================================"
    print_status $BLUE "  Chthon ShortLink - Starting All Services"
    print_status $BLUE "============================================"
    
    # Create cleanup marker
    touch "$LOG_DIR/cleanup_needed"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status $RED "Go is not installed or not in PATH ✗"
        exit 1
    fi
    
    # Check if required tools are available
    if ! command -v lsof &> /dev/null; then
        print_status $RED "lsof is not installed (required for port checking) ✗"
        exit 1
    fi
    
    if ! command -v nc &> /dev/null; then
        print_status $RED "nc (netcat) is not installed (required for connectivity checks) ✗"
        exit 1
    fi
    
    # Build services
    if ! build_services; then
        print_status $RED "Failed to build services ✗"
        exit 1
    fi
    
    # Check infrastructure
    if ! check_infrastructure; then
        print_status $RED "Infrastructure connectivity check failed ✗"
        print_status $YELLOW "Please ensure PostgreSQL, Redis, MongoDB, and Kafka are running on 192.168.1.127"
        exit 1
    fi
    
    # Start services in correct order
    print_status $BLUE "Starting services..."
    
    # 1. User Management Service (port 8084)
    if ! start_service "user-management" "$BIN_DIR/user-management-service" 8084 "http://localhost:8084/health"; then
        exit 1
    fi
    
    # 2. Shortlink Service (port 8081)
    if ! start_service "shortlink" "$BIN_DIR/shortlink-service" 8081 "http://localhost:8081/health"; then
        exit 1
    fi
    
    # 3. Redirect Service (port 8082)
    if ! start_service "redirect" "$BIN_DIR/redirect-service" 8082 "http://localhost:8082/health"; then
        exit 1
    fi
    
    # 4. Analytics Service (port 8083)
    if ! start_service "analytics" "$BIN_DIR/analytics-service" 8083 "http://localhost:8083/health"; then
        exit 1
    fi
    
    # 5. API Gateway (port 8080) - started last
    if ! start_service "api-gateway" "$BIN_DIR/api-gateway" 8080 "http://localhost:8080/health"; then
        exit 1
    fi
    
    # Final health check for API Gateway
    print_status $BLUE "Performing final system health check..."
    if check_service_health "API Gateway" "http://localhost:8080/health"; then
        print_status $GREEN "============================================"
        print_status $GREEN "  All services started successfully! 🚀"
        print_status $GREEN "============================================"
        print_status $BLUE "Service URLs:"
        print_status $BLUE "  • API Gateway:     http://localhost:8080"
        print_status $BLUE "  • Swagger Docs:    http://localhost:8080/docs/swagger/index.html"
        print_status $BLUE "  • User Management: http://localhost:8084"
        print_status $BLUE "  • Shortlink:       http://localhost:8081"
        print_status $BLUE "  • Redirect:        http://localhost:8082"
        print_status $BLUE "  • Analytics:       http://localhost:8083"
        print_status $YELLOW ""
        print_status $YELLOW "Logs are available in: $LOG_DIR/"
        print_status $YELLOW "To stop all services, run: ./stop-all-services.sh"
        print_status $YELLOW ""
    else
        print_status $RED "System health check failed ✗"
        exit 1
    fi
    
    # Keep running and monitor services
    print_status $BLUE "Monitoring services... Press Ctrl+C to stop all services"
    
    while true; do
        sleep 30
        
        # Check if all services are still running
        all_healthy=true
        
        for service in "api-gateway:8080" "user-management:8084" "shortlink:8081" "redirect:8082" "analytics:8083"; do
            service_name=$(echo $service | cut -d: -f1)
            port=$(echo $service | cut -d: -f2)
            
            if ! lsof -i :$port > /dev/null 2>&1; then
                print_status $RED "$service_name (port $port) is not running ✗"
                all_healthy=false
            fi
        done
        
        if [ "$all_healthy" = true ]; then
            print_status $GREEN "All services are healthy ✓"
        else
            print_status $RED "Some services are down! Check logs in $LOG_DIR/"
        fi
    done
}

# Handle Ctrl+C
handle_interrupt() {
    print_status $YELLOW ""
    print_status $YELLOW "Interrupt received. Stopping all services..."
    
    # Stop all services
    if [ -f "./stop-all-services.sh" ]; then
        ./stop-all-services.sh
    else
        print_status $YELLOW "Stopping services manually..."
        for port in 8080 8081 8082 8083 8084; do
            local pid=$(lsof -ti :$port 2>/dev/null || true)
            if [ ! -z "$pid" ]; then
                print_status $YELLOW "Stopping service on port $port (PID: $pid)..."
                kill -TERM $pid 2>/dev/null || kill -9 $pid 2>/dev/null || true
            fi
        done
    fi
    
    print_status $GREEN "All services stopped."
    exit 0
}

trap handle_interrupt INT TERM

# Run main function
main
