#!/bin/bash

# Chthon Short Link - Production Setup Script

set -e

echo "🏭 Setting up Chthon Short Link production environment..."

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

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "Please run this script as root or with sudo"
        exit 1
    fi
}

# Check if required tools are installed
check_requirements() {
    print_status "Checking requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        systemctl enable docker
        systemctl start docker
        print_success "Docker installed and started"
    else
        print_success "Docker found"
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Installing Docker Compose..."
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        print_success "Docker Compose installed"
    else
        print_success "Docker Compose found"
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        print_error "Git is not installed. Installing Git..."
        apt-get update && apt-get install -y git
        print_success "Git installed"
    else
        print_success "Git found"
    fi
}

# Setup firewall
setup_firewall() {
    print_status "Setting up firewall..."
    
    # Install ufw if not present
    if ! command -v ufw &> /dev/null; then
        apt-get install -y ufw
    fi
    
    # Reset firewall rules
    ufw --force reset
    
    # Default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH
    ufw allow ssh
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow application ports
    ufw allow 8080/tcp  # API Gateway
    
    # Enable firewall
    ufw --force enable
    
    print_success "Firewall configured"
}

# Setup directories
setup_directories() {
    print_status "Setting up directories..."
    
    # Create application directory
    APP_DIR="/opt/chthon-shortlink"
    mkdir -p $APP_DIR
    mkdir -p $APP_DIR/data
    mkdir -p $APP_DIR/logs
    mkdir -p $APP_DIR/backups
    mkdir -p $APP_DIR/ssl
    
    # Set permissions
    chmod 755 $APP_DIR
    chmod 755 $APP_DIR/data
    chmod 755 $APP_DIR/logs
    chmod 755 $APP_DIR/backups
    chmod 700 $APP_DIR/ssl
    
    print_success "Directories created"
}

# Setup environment file
setup_env() {
    print_status "Setting up production environment file..."
    
    APP_DIR="/opt/chthon-shortlink"
    
    if [ ! -f $APP_DIR/.env ]; then
        cat > $APP_DIR/.env << EOF
# Production Environment Configuration
# IMPORTANT: Update all passwords and secrets before running!

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=shortlink_user
DB_PASSWORD=CHANGE_THIS_PASSWORD_IN_PRODUCTION
DB_NAME=shortlink_prod
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=CHANGE_THIS_REDIS_PASSWORD
REDIS_DB=0
REDIS_POOL_SIZE=20

# MongoDB Configuration
MONGODB_URI=mongodb://mongo_user:CHANGE_THIS_MONGO_PASSWORD@mongodb:27017
MONGODB_DATABASE=shortlink_analytics_prod

# JWT Configuration
JWT_SECRET=CHANGE_THIS_TO_A_VERY_LONG_RANDOM_STRING_IN_PRODUCTION
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=7d

# Server Configuration
API_GATEWAY_PORT=8080
SHORTLINK_SERVICE_PORT=8081
REDIRECT_SERVICE_PORT=8082
ANALYTICS_SERVICE_PORT=8083
USER_MANAGEMENT_SERVICE_PORT=8084

# External Service URLs
SHORTLINK_SERVICE_URL=http://shortlink-service:8081
REDIRECT_SERVICE_URL=http://redirect-service:8082
ANALYTICS_SERVICE_URL=http://analytics-service:8083
USER_MANAGEMENT_SERVICE_URL=http://user-management-service:8084

# Kafka Configuration
KAFKA_BROKERS=kafka:9092
KAFKA_CONSUMER_GROUP=shortlink-prod-group
KAFKA_TOPIC_CLICKS=click-events-prod
KAFKA_TOPIC_ANALYTICS=analytics-events-prod

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m

# Security
CORS_ALLOWED_ORIGINS=https://yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization

# Logging
LOG_LEVEL=warn
LOG_FORMAT=json

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090

# Domain Configuration
BASE_DOMAIN=yourdomain.com
PROTOCOL=https

# Cache Configuration
CACHE_TTL=24h
CACHE_CLEANUP_INTERVAL=1h

# Worker Configuration
WORKER_POOL_SIZE=20
QUEUE_SIZE=10000

# Health Check
HEALTH_CHECK_INTERVAL=30s
HEALTH_CHECK_TIMEOUT=10s

# Production
DEV_MODE=false
DEBUG=false
EOF
        
        chmod 600 $APP_DIR/.env
        print_success "Production .env file created"
        print_warning "IMPORTANT: Update all passwords and secrets in $APP_DIR/.env before starting services!"
    else
        print_warning "Environment file already exists, skipping..."
    fi
}

# Setup production docker-compose
setup_docker_compose() {
    print_status "Setting up production docker-compose..."
    
    APP_DIR="/opt/chthon-shortlink"
    
    cat > $APP_DIR/docker-compose.prod.yml << 'EOF'
version: '3.8'

services:
  # API Gateway
  api-gateway:
    image: ghcr.io/chthon-short-link/api-gateway:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - ENV=production
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  # Shortlink Service
  shortlink-service:
    image: ghcr.io/chthon-short-link/shortlink-service:latest
    restart: unless-stopped
    environment:
      - ENV=production
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
      - kafka
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  # Redirect Service
  redirect-service:
    image: ghcr.io/chthon-short-link/redirect-service:latest
    restart: unless-stopped
    environment:
      - ENV=production
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
      - kafka
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  # Analytics Service
  analytics-service:
    image: ghcr.io/chthon-short-link/analytics-service:latest
    restart: unless-stopped
    environment:
      - ENV=production
    env_file:
      - .env
    depends_on:
      - mongodb
      - kafka
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8083/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  # User Management Service
  user-management-service:
    image: ghcr.io/chthon-short-link/user-management-service:latest
    restart: unless-stopped
    environment:
      - ENV=production
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8084/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: shortlink_prod
      POSTGRES_USER: shortlink_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U shortlink_user -d shortlink_prod"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Redis Cache
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "auth", "${REDIS_PASSWORD}", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  # MongoDB for Analytics
  mongodb:
    image: mongo:6
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: mongo_user
      MONGO_INITDB_ROOT_PASSWORD: ${MONGODB_PASSWORD}
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Apache Kafka
  kafka:
    image: confluentinc/cp-kafka:latest
    restart: unless-stopped
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    depends_on:
      - zookeeper
    volumes:
      - kafka_data:/var/lib/kafka/data

  # Zookeeper for Kafka
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    restart: unless-stopped
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    volumes:
      - zookeeper_data:/var/lib/zookeeper

  # Watchtower for automatic updates
  watchtower:
    image: containrrr/watchtower
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WATCHTOWER_CLEANUP=true
      - WATCHTOWER_POLL_INTERVAL=3600
      - WATCHTOWER_LABEL_ENABLE=true

volumes:
  postgres_data:
  redis_data:
  mongodb_data:
  kafka_data:
  zookeeper_data:
EOF
    
    print_success "Production docker-compose file created"
}

# Setup SSL certificates (Let's Encrypt)
setup_ssl() {
    print_status "Setting up SSL certificates..."
    
    # Install certbot
    if ! command -v certbot &> /dev/null; then
        apt-get update
        apt-get install -y certbot
    fi
    
    print_warning "SSL setup requires manual domain configuration"
    print_status "To setup SSL certificates:"
    print_status "1. Point your domain to this server's IP"
    print_status "2. Run: certbot certonly --standalone -d yourdomain.com"
    print_status "3. Setup automatic renewal: echo '0 12 * * * /usr/bin/certbot renew --quiet' | crontab -"
}

# Setup systemd service
setup_systemd() {
    print_status "Setting up systemd service..."
    
    cat > /etc/systemd/system/chthon-shortlink.service << EOF
[Unit]
Description=Chthon Short Link Services
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/chthon-shortlink
ExecStart=/usr/local/bin/docker-compose -f docker-compose.prod.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose.prod.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable chthon-shortlink.service
    
    print_success "Systemd service created and enabled"
}

# Setup monitoring
setup_monitoring() {
    print_status "Setting up basic monitoring..."
    
    # Create monitoring script
    cat > /opt/chthon-shortlink/monitor.sh << 'EOF'
#!/bin/bash
# Basic monitoring script for Chthon Short Link

LOGFILE="/opt/chthon-shortlink/logs/monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Check if services are running
if docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml ps | grep -q "Up"; then
    echo "[$DATE] Services are running" >> $LOGFILE
else
    echo "[$DATE] ERROR: Some services are down" >> $LOGFILE
    # Restart services
    docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml up -d
fi

# Check disk space
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "[$DATE] WARNING: Disk usage is ${DISK_USAGE}%" >> $LOGFILE
fi

# Check memory usage
MEM_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
if [ $MEM_USAGE -gt 80 ]; then
    echo "[$DATE] WARNING: Memory usage is ${MEM_USAGE}%" >> $LOGFILE
fi
EOF
    
    chmod +x /opt/chthon-shortlink/monitor.sh
    
    # Add to crontab
    (crontab -l 2>/dev/null; echo "*/5 * * * * /opt/chthon-shortlink/monitor.sh") | crontab -
    
    print_success "Basic monitoring setup completed"
}

# Setup backup script
setup_backup() {
    print_status "Setting up backup script..."
    
    cat > /opt/chthon-shortlink/backup.sh << 'EOF'
#!/bin/bash
# Backup script for Chthon Short Link

BACKUP_DIR="/opt/chthon-shortlink/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Database backup
docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml exec -T postgres pg_dump -U shortlink_user shortlink_prod > $BACKUP_DIR/postgres_$DATE.sql

# Redis backup
docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml exec -T redis redis-cli --rdb /data/dump_$DATE.rdb

# MongoDB backup
docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml exec -T mongodb mongodump --out /data/backup_$DATE

# Compress backups
tar -czf $BACKUP_DIR/backup_$DATE.tar.gz $BACKUP_DIR/*_$DATE.*
rm -f $BACKUP_DIR/*_$DATE.sql $BACKUP_DIR/*_$DATE.rdb
rm -rf $BACKUP_DIR/backup_$DATE

# Remove old backups (keep last 7 days)
find $BACKUP_DIR -name "backup_*.tar.gz" -mtime +7 -delete

echo "Backup completed: backup_$DATE.tar.gz"
EOF
    
    chmod +x /opt/chthon-shortlink/backup.sh
    
    # Add daily backup to crontab
    (crontab -l 2>/dev/null; echo "0 2 * * * /opt/chthon-shortlink/backup.sh") | crontab -
    
    print_success "Backup script setup completed"
}

# Main setup function
main() {
    echo "============================================"
    echo "🏭 Chthon Short Link Production Setup"
    echo "============================================"
    echo
    
    check_root
    echo
    
    check_requirements
    echo
    
    setup_firewall
    echo
    
    setup_directories
    echo
    
    setup_env
    echo
    
    setup_docker_compose
    echo
    
    setup_ssl
    echo
    
    setup_systemd
    echo
    
    setup_monitoring
    echo
    
    setup_backup
    echo
    
    print_success "🎉 Production environment setup completed!"
    echo
    echo "Next steps:"
    echo "1. Update passwords and secrets in /opt/chthon-shortlink/.env"
    echo "2. Setup your domain and SSL certificates"
    echo "3. Start services: systemctl start chthon-shortlink"
    echo "4. Check status: systemctl status chthon-shortlink"
    echo
    echo "Important files:"
    echo "  Configuration: /opt/chthon-shortlink/.env"
    echo "  Docker Compose: /opt/chthon-shortlink/docker-compose.prod.yml"
    echo "  Logs: /opt/chthon-shortlink/logs/"
    echo "  Backups: /opt/chthon-shortlink/backups/"
    echo
    echo "Management commands:"
    echo "  systemctl start chthon-shortlink    - Start services"
    echo "  systemctl stop chthon-shortlink     - Stop services"
    echo "  systemctl restart chthon-shortlink  - Restart services"
    echo "  docker-compose -f /opt/chthon-shortlink/docker-compose.prod.yml logs -f  - View logs"
    echo
    echo "🔒 Don't forget to update all passwords and secrets before starting!"
}

# Run main function
main "$@"
