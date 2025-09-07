#!/bin/bash

# Chthon Short Link - Health Check Script
# This script checks the health of all services and reports any issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_GATEWAY_URL="http://localhost:8080"
SHORTLINK_SERVICE_URL="http://localhost:8081"
REDIRECT_SERVICE_URL="http://localhost:8082"
ANALYTICS_SERVICE_URL="http://localhost:8083"
USER_MANAGEMENT_SERVICE_URL="http://localhost:8084"

TIMEOUT=10
RETRY_COUNT=3

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Function to check HTTP endpoint
check_http_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="${3:-200}"
    
    print_status "Checking $name..."
    
    for i in $(seq 1 $RETRY_COUNT); do
        if response=$(curl -s -w "%{http_code}" --max-time $TIMEOUT "$url/health" 2>/dev/null); then
            http_code="${response: -3}"
            if [ "$http_code" = "$expected_status" ]; then
                print_success "$name is healthy (HTTP $http_code)"
                return 0
            else
                print_warning "$name returned HTTP $http_code (attempt $i/$RETRY_COUNT)"
            fi
        else
            print_warning "$name is not responding (attempt $i/$RETRY_COUNT)"
        fi
        
        if [ $i -lt $RETRY_COUNT ]; then
            sleep 2
        fi
    done
    
    print_error "$name is unhealthy or not responding"
    return 1
}

# Function to check Docker container
check_docker_container() {
    local container_name="$1"
    
    if docker ps --filter "name=$container_name" --filter "status=running" | grep -q "$container_name"; then
        print_success "Container $container_name is running"
        return 0
    else
        print_error "Container $container_name is not running"
        return 1
    fi
}

# Function to check database connectivity
check_database() {
    print_status "Checking database connectivity..."
    
    if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
        print_success "PostgreSQL is responding"
    else
        print_error "PostgreSQL is not responding"
        return 1
    fi
    
    if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
        print_success "Redis is responding"
    else
        print_error "Redis is not responding"
        return 1
    fi
    
    if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        print_success "MongoDB is responding"
    else
        print_error "MongoDB is not responding"
        return 1
    fi
}

# Function to check system resources
check_system_resources() {
    print_status "Checking system resources..."
    
    # Check disk space
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ "$DISK_USAGE" -lt 80 ]; then
        print_success "Disk usage: ${DISK_USAGE}%"
    elif [ "$DISK_USAGE" -lt 90 ]; then
        print_warning "Disk usage: ${DISK_USAGE}% (Warning: above 80%)"
    else
        print_error "Disk usage: ${DISK_USAGE}% (Critical: above 90%)"
    fi
    
    # Check memory usage
    MEM_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
    if [ "$MEM_USAGE" -lt 80 ]; then
        print_success "Memory usage: ${MEM_USAGE}%"
    elif [ "$MEM_USAGE" -lt 90 ]; then
        print_warning "Memory usage: ${MEM_USAGE}% (Warning: above 80%)"
    else
        print_error "Memory usage: ${MEM_USAGE}% (Critical: above 90%)"
    fi
    
    # Check load average
    LOAD_AVG=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    CPU_CORES=$(nproc)
    if (( $(echo "$LOAD_AVG < $CPU_CORES" | bc -l) )); then
        print_success "Load average: $LOAD_AVG (${CPU_CORES} cores)"
    else
        print_warning "Load average: $LOAD_AVG (${CPU_CORES} cores) - High load"
    fi
}

# Function to check logs for errors
check_logs_for_errors() {
    print_status "Checking recent logs for errors..."
    
    # Check for recent errors in docker-compose logs
    ERROR_COUNT=$(docker-compose logs --since="5m" 2>/dev/null | grep -i "error\|fatal\|panic" | wc -l)
    
    if [ "$ERROR_COUNT" -eq 0 ]; then
        print_success "No recent errors found in logs"
    elif [ "$ERROR_COUNT" -lt 5 ]; then
        print_warning "Found $ERROR_COUNT recent errors in logs"
    else
        print_error "Found $ERROR_COUNT recent errors in logs (check logs: docker-compose logs)"
    fi
}

# Function to perform functional tests
perform_functional_tests() {
    print_status "Performing basic functional tests..."
    
    # Test API Gateway
    if response=$(curl -s "$API_GATEWAY_URL/api/v1/health" 2>/dev/null); then
        print_success "API Gateway health endpoint is accessible"
    else
        print_error "API Gateway health endpoint is not accessible"
        return 1
    fi
    
    # Test short link creation (if API is available)
    if curl -s -f "$API_GATEWAY_URL/api/v1/links" >/dev/null 2>&1; then
        print_success "Short link API endpoint is accessible"
    else
        print_warning "Short link API endpoint is not accessible (may require authentication)"
    fi
}

# Function to check SSL certificate expiration
check_ssl_certificate() {
    local domain="$1"
    
    if [ -n "$domain" ]; then
        print_status "Checking SSL certificate for $domain..."
        
        if expiry_date=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2); then
            expiry_timestamp=$(date -d "$expiry_date" +%s)
            current_timestamp=$(date +%s)
            days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
            
            if [ "$days_until_expiry" -gt 30 ]; then
                print_success "SSL certificate expires in $days_until_expiry days"
            elif [ "$days_until_expiry" -gt 7 ]; then
                print_warning "SSL certificate expires in $days_until_expiry days"
            else
                print_error "SSL certificate expires in $days_until_expiry days (Critical: renew soon!)"
            fi
        else
            print_warning "Could not check SSL certificate for $domain"
        fi
    fi
}

# Main health check function
main() {
    local overall_health=0
    
    print_header "🏥 Chthon Short Link Health Check"
    echo
    
    # Check Docker containers
    print_header "🐳 Docker Containers"
    check_docker_container "api-gateway" || overall_health=1
    check_docker_container "shortlink-service" || overall_health=1
    check_docker_container "redirect-service" || overall_health=1
    check_docker_container "analytics-service" || overall_health=1
    check_docker_container "user-management-service" || overall_health=1
    echo
    
    # Check databases
    print_header "🗄️ Databases"
    check_database || overall_health=1
    echo
    
    # Check HTTP endpoints
    print_header "🌐 HTTP Endpoints"
    check_http_endpoint "API Gateway" "$API_GATEWAY_URL" || overall_health=1
    check_http_endpoint "Shortlink Service" "$SHORTLINK_SERVICE_URL" || overall_health=1
    check_http_endpoint "Redirect Service" "$REDIRECT_SERVICE_URL" || overall_health=1
    check_http_endpoint "Analytics Service" "$ANALYTICS_SERVICE_URL" || overall_health=1
    check_http_endpoint "User Management Service" "$USER_MANAGEMENT_SERVICE_URL" || overall_health=1
    echo
    
    # Check system resources
    print_header "💻 System Resources"
    check_system_resources
    echo
    
    # Check logs
    print_header "📋 Log Analysis"
    check_logs_for_errors
    echo
    
    # Perform functional tests
    print_header "🧪 Functional Tests"
    perform_functional_tests || overall_health=1
    echo
    
    # Check SSL certificate if domain is provided
    if [ -n "$1" ]; then
        print_header "🔒 SSL Certificate"
        check_ssl_certificate "$1"
        echo
    fi
    
    # Summary
    print_header "📊 Health Check Summary"
    if [ $overall_health -eq 0 ]; then
        print_success "All systems are healthy! 🎉"
        echo
        echo "✅ All services are running properly"
        echo "✅ All health endpoints are responding"
        echo "✅ All databases are accessible"
        echo "✅ System resources are within normal limits"
    else
        print_error "Some issues were detected! ⚠️"
        echo
        echo "❌ One or more services have issues"
        echo "💡 Check the detailed output above for specific problems"
        echo "📖 Run 'docker-compose logs' to see detailed logs"
        echo "🔧 Run 'docker-compose ps' to see container status"
    fi
    
    echo
    print_header "📝 Quick Commands"
    echo "View logs:           docker-compose logs -f"
    echo "Restart services:    docker-compose restart"
    echo "Check status:        docker-compose ps"
    echo "Monitor resources:   docker stats"
    
    exit $overall_health
}

# Parse command line arguments
DOMAIN=""
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--domain)
            DOMAIN="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -d, --domain DOMAIN    Check SSL certificate for DOMAIN"
            echo "  -h, --help            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    Basic health check"
            echo "  $0 -d example.com     Health check with SSL certificate check"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main "$DOMAIN"
