#!/bin/bash

# Chthon ShortLink - Stop All Services Script (macOS/Linux)
# This script stops all running microservices gracefully

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICES_DIR="$(dirname "$0")"
LOG_DIR="$SERVICES_DIR/logs"
TIMEOUT=10

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}[$(date '+%Y-%m-%d %H:%M:%S')] $message${NC}"
}

# Function to stop a service by port
stop_service_by_port() {
    local port=$1
    local service_name=$2
    
    print_status $YELLOW "Stopping $service_name (port $port)..."
    
    # Find process using the port
    local pid=$(lsof -ti :$port 2>/dev/null || true)
    
    if [ -z "$pid" ]; then
        print_status $GREEN "$service_name is not running ✓"
        return 0
    fi
    
    print_status $YELLOW "Found $service_name process (PID: $pid), stopping..."
    
    # Try graceful shutdown first
    if kill -TERM $pid 2>/dev/null; then
        print_status $YELLOW "Sent SIGTERM to $service_name, waiting for graceful shutdown..."
        
        # Wait for graceful shutdown
        local count=0
        while [ $count -lt $TIMEOUT ]; do
            if ! kill -0 $pid 2>/dev/null; then
                print_status $GREEN "$service_name stopped gracefully ✓"
                return 0
            fi
            sleep 1
            count=$((count + 1))
        done
        
        # Force kill if still running
        print_status $YELLOW "$service_name didn't stop gracefully, force killing..."
        if kill -9 $pid 2>/dev/null; then
            print_status $GREEN "$service_name force killed ✓"
        else
            print_status $RED "Failed to kill $service_name ✗"
            return 1
        fi
    else
        print_status $RED "Failed to send SIGTERM to $service_name ✗"
        return 1
    fi
    
    return 0
}

# Function to stop service by PID file
stop_service_by_pid() {
    local service_name=$1
    local pid_file="$LOG_DIR/$service_name.pid"
    
    if [ ! -f "$pid_file" ]; then
        return 1
    fi
    
    local pid=$(cat "$pid_file" 2>/dev/null || true)
    
    if [ -z "$pid" ]; then
        rm -f "$pid_file"
        return 1
    fi
    
    if ! kill -0 $pid 2>/dev/null; then
        print_status $YELLOW "$service_name PID file exists but process is not running"
        rm -f "$pid_file"
        return 1
    fi
    
    print_status $YELLOW "Stopping $service_name (PID: $pid from file)..."
    
    # Try graceful shutdown first
    if kill -TERM $pid 2>/dev/null; then
        # Wait for graceful shutdown
        local count=0
        while [ $count -lt $TIMEOUT ]; do
            if ! kill -0 $pid 2>/dev/null; then
                print_status $GREEN "$service_name stopped gracefully ✓"
                rm -f "$pid_file"
                return 0
            fi
            sleep 1
            count=$((count + 1))
        done
        
        # Force kill if still running
        print_status $YELLOW "$service_name didn't stop gracefully, force killing..."
        if kill -9 $pid 2>/dev/null; then
            print_status $GREEN "$service_name force killed ✓"
            rm -f "$pid_file"
        else
            print_status $RED "Failed to kill $service_name ✗"
            return 1
        fi
    else
        print_status $RED "Failed to send SIGTERM to $service_name ✗"
        rm -f "$pid_file"
        return 1
    fi
    
    return 0
}

# Function to cleanup log and pid files
cleanup_files() {
    print_status $BLUE "Cleaning up log and PID files..."
    
    if [ -d "$LOG_DIR" ]; then
        # Remove PID files
        rm -f "$LOG_DIR"/*.pid 2>/dev/null || true
        
        # Remove cleanup marker
        rm -f "$LOG_DIR/cleanup_needed" 2>/dev/null || true
        
        print_status $GREEN "Cleanup completed ✓"
    fi
}

# Function to check if any services are still running
check_remaining_services() {
    local remaining=false
    
    for port in 8080 8081 8082 8083 8084; do
        if lsof -i :$port > /dev/null 2>&1; then
            local pid=$(lsof -ti :$port 2>/dev/null || true)
            print_status $RED "Service still running on port $port (PID: $pid) ✗"
            remaining=true
        fi
    done
    
    return $remaining
}

# Main execution
main() {
    print_status $BLUE "============================================"
    print_status $BLUE "  Chthon ShortLink - Stopping All Services"
    print_status $BLUE "============================================"
    
    # Check if lsof is available
    if ! command -v lsof &> /dev/null; then
        print_status $RED "lsof is not installed (required for process management) ✗"
        exit 1
    fi
    
    # Create logs directory if it doesn't exist
    mkdir -p "$LOG_DIR"
    
    # Stop services in reverse order (API Gateway first)
    print_status $BLUE "Stopping services in reverse order..."
    
    # 1. API Gateway (port 8080)
    if ! stop_service_by_pid "api-gateway" && ! stop_service_by_port 8080 "API Gateway"; then
        print_status $YELLOW "API Gateway may not have been running"
    fi
    
    # 2. Analytics Service (port 8083)
    if ! stop_service_by_pid "analytics" && ! stop_service_by_port 8083 "Analytics Service"; then
        print_status $YELLOW "Analytics Service may not have been running"
    fi
    
    # 3. Redirect Service (port 8082)
    if ! stop_service_by_pid "redirect" && ! stop_service_by_port 8082 "Redirect Service"; then
        print_status $YELLOW "Redirect Service may not have been running"
    fi
    
    # 4. Shortlink Service (port 8081)
    if ! stop_service_by_pid "shortlink" && ! stop_service_by_port 8081 "Shortlink Service"; then
        print_status $YELLOW "Shortlink Service may not have been running"
    fi
    
    # 5. User Management Service (port 8084)
    if ! stop_service_by_pid "user-management" && ! stop_service_by_port 8084 "User Management Service"; then
        print_status $YELLOW "User Management Service may not have been running"
    fi
    
    # Wait a moment for all processes to fully terminate
    print_status $YELLOW "Waiting for all processes to terminate..."
    sleep 2
    
    # Check if any services are still running
    if check_remaining_services; then
        print_status $RED "Some services are still running. Attempting force cleanup..."
        
        # Force kill any remaining processes
        for port in 8080 8081 8082 8083 8084; do
            local pid=$(lsof -ti :$port 2>/dev/null || true)
            if [ ! -z "$pid" ]; then
                print_status $YELLOW "Force killing process on port $port (PID: $pid)..."
                kill -9 $pid 2>/dev/null || true
            fi
        done
        
        sleep 1
        
        # Final check
        if check_remaining_services; then
            print_status $RED "Failed to stop all services ✗"
            exit 1
        fi
    fi
    
    # Cleanup files
    cleanup_files
    
    print_status $GREEN "============================================"
    print_status $GREEN "  All services stopped successfully! 🛑"
    print_status $GREEN "============================================"
    
    return 0
}

# Handle script arguments
case "${1:-}" in
    --force|-f)
        print_status $YELLOW "Force mode enabled - will kill all processes immediately"
        TIMEOUT=0
        ;;
    --help|-h)
        echo "Usage: $0 [--force|-f] [--help|-h]"
        echo ""
        echo "Options:"
        echo "  --force, -f    Force kill all processes immediately (no graceful shutdown)"
        echo "  --help, -h     Show this help message"
        echo ""
        echo "This script stops all Chthon ShortLink microservices."
        exit 0
        ;;
esac

# Run main function
main
