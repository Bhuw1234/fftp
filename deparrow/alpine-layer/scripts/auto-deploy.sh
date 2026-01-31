#!/bin/bash
# DEparrow Auto-Deploy Script
# Automatically deploys and configures a DEparrow Alpine Linux node

set -e

# Configuration
DEPARROW_VERSION="1.0.0"
IMAGE_NAME="deparrow/alpine-node"
CONFIG_DIR="/etc/deparrow"
DATA_DIR="/var/lib/deparrow"
LOG_DIR="/var/log/deparrow"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_system_requirements() {
    log_info "Checking system requirements..."
    
    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root"
        exit 1
    fi
    
    # Check architecture
    local arch=$(uname -m)
    if [[ "$arch" != "x86_64" && "$arch" != "aarch64" ]]; then
        log_warning "Unsupported architecture: $arch. Supported: x86_64, aarch64"
    fi
    
    # Check available memory
    local memory_gb=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$memory_gb" -lt 2 ]; then
        log_warning "Low memory: ${memory_gb}GB. Recommended: 4GB+"
    fi
    
    # Check available disk space
    local disk_gb=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$disk_gb" -lt 20 ]; then
        log_warning "Low disk space: ${disk_gb}GB. Recommended: 50GB+"
    fi
    
    log_success "System requirements check completed"
}

install_docker() {
    log_info "Installing Docker..."
    
    if command -v docker &> /dev/null; then
        log_info "Docker already installed"
        return 0
    fi
    
    # Install Docker on Alpine
    apk update
    apk add docker docker-cli-compose
    
    # Enable and start Docker service
    rc-update add docker default
    rc-service docker start
    
    log_success "Docker installed and started"
}

setup_deparrow_user() {
    log_info "Setting up DEparrow user..."
    
    # Create user and group
    if ! id "deparrow" &>/dev/null; then
        addgroup -S deparrow
        adduser -S deparrow -G deparrow -s /bin/bash
        echo "deparrow ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/deparrow
    fi
    
    # Add user to docker group
    adduser deparrow docker
    
    log_success "DEparrow user configured"
}

create_directories() {
    log_info "Creating directory structure..."
    
    mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    chown -R deparrow:deparrow "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
    
    log_success "Directory structure created"
}

generate_node_identity() {
    log_info "Generating node identity..."
    
    # Generate UUID for node ID
    if [ ! -f "$CONFIG_DIR/node-id" ]; then
        echo "$(cat /proc/sys/kernel/random/uuid)" > "$CONFIG_DIR/node-id"
        chmod 600 "$CONFIG_DIR/node-id"
        chown deparrow:deparrow "$CONFIG_DIR/node-id"
    fi
    
    # Generate SSH keys for secure communication
    if [ ! -f "$CONFIG_DIR/keys/private.pem" ]; then
        mkdir -p "$CONFIG_DIR/keys"
        openssl genrsa -out "$CONFIG_DIR/keys/private.pem" 2048
        openssl rsa -in "$CONFIG_DIR/keys/private.pem" -pubout -out "$CONFIG_DIR/keys/public.pem"
        chmod 600 "$CONFIG_DIR/keys/private.pem"
        chmod 644 "$CONFIG_DIR/keys/public.pem"
        chown -R deparrow:deparrow "$CONFIG_DIR/keys"
    fi
    
    NODE_ID=$(cat "$CONFIG_DIR/node-id")
    log_success "Node identity generated: $NODE_ID"
}

configure_environment() {
    log_info "Configuring environment variables..."
    
    # Create environment file
    cat > "$CONFIG_DIR/environment" << EOF
# DEparrow Node Configuration
NODE_ID=$NODE_ID
NODE_ARCH=$(uname -m)
NODE_CPU=$(nproc)
NODE_MEMORY=$(free -h | awk '/^Mem:/ {print $2}')
NODE_DISK=$(df -h / | awk 'NR==2 {print $2}')
NODE_GPU=0

# Bootstrap configuration
DEPARROW_BOOTSTRAP=${DEPARROW_BOOTSTRAP:-https://bootstrap.deparrow.net}
DEPARROW_ORCHESTRATOR_HOST=${DEPARROW_ORCHESTRATOR_HOST:-orchestrator.deparrow.net}

# API Configuration
DEPARROW_API_KEY=${DEPARROW_API_KEY:-}
DEPARROW_JWT_SECRET=${DEPARROW_JWT_SECRET:-deparrow-secret-key-change-me}

# Logging
LOG_LEVEL=info
EOF
    
    # Load environment
    source "$CONFIG_DIR/environment"
    
    log_success "Environment configured"
}

register_with_bootstrap() {
    if [ -n "$DEPARROW_API_KEY" ] && [ -n "$DEPARROW_BOOTSTRAP" ]; then
        log_info "Registering with DEparrow bootstrap..."
        
        # Read public key
        local public_key=$(base64 -w 0 "$CONFIG_DIR/keys/public.pem")
        
        # Register node
        local response=$(curl -s -X POST \
            -H "Authorization: Bearer $DEPARROW_API_KEY" \
            -H "Content-Type: application/json" \
            -d "{
                \"node_id\": \"$NODE_ID\",
                \"public_key\": \"$public_key\",
                \"resources\": {
                    \"cpu\": $NODE_CPU,
                    \"memory\": \"$NODE_MEMORY\",
                    \"disk\": \"$NODE_DISK\",
                    \"arch\": \"$NODE_ARCH\"
                }
            }" \
            "$DEPARROW_BOOTSTRAP/api/v1/nodes/register" 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            log_success "Node registered successfully with bootstrap"
            echo "$response" > "$CONFIG_DIR/registration.json"
        else
            log_warning "Failed to register with bootstrap (may retry later)"
        fi
    else
        log_info "Bootstrap API key not provided, skipping registration"
    fi
}

create_deployment_config() {
    log_info "Creating deployment configuration..."
    
    # Create systemd service file
    cat > /etc/systemd/system/deparrow-node.service << EOF
[Unit]
Description=DEparrow Compute Node
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=deparrow
Group=deparrow
EnvironmentFile=$CONFIG_DIR/environment
ExecStartPre=/usr/bin/docker pull $IMAGE_NAME:$DEPARROW_VERSION
ExecStart=/usr/bin/docker run --rm \\
  --name=deparrow-node \\
  --privileged \\
  --network=host \\
  -p 4222:4222 \\
  -p 9090:9090 \\
  -e NODE_ID \\
  -e NODE_ARCH \\
  -e NODE_CPU \\
  -e NODE_MEMORY \\
  -e NODE_DISK \\
  -e NODE_GPU \\
  -e DEPARROW_BOOTSTRAP \\
  -e DEPARROW_ORCHESTRATOR_HOST \\
  -e DEPARROW_API_KEY \\
  -e DEPARROW_JWT_SECRET \\
  -v /var/run/docker.sock:/var/run/docker.sock \\
  -v $DATA_DIR:/var/lib/deparrow \\
  -v $CONFIG_DIR:/etc/deparrow \\
  -v $LOG_DIR:/var/log/deparrow \\
  $IMAGE_NAME:$DEPARROW_VERSION
ExecStop=/usr/bin/docker stop deparrow-node || true
Restart=always
RestartSec=10
KillMode=mixed
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
EOF
    
    # Create cron job for monitoring
    cat > /etc/cron.d/deparrow-monitor << EOF
# DEparrow Node Monitoring
*/5 * * * * root /opt/deparrow/scripts/monitor-node.sh >> $LOG_DIR/monitor.log 2>&1
EOF
    
    log_success "Deployment configuration created"
}

start_services() {
    log_info "Starting services..."
    
    # Reload systemd
    systemctl daemon-reload
    
    # Enable and start the service
    systemctl enable deparrow-node
    systemctl start deparrow-node
    
    log_success "DEparrow node service started"
}

verify_installation() {
    log_info "Verifying installation..."
    
    # Check if Docker is running
    if ! docker ps > /dev/null 2>&1; then
        log_error "Docker is not running"
        return 1
    fi
    
    # Check if DEparrow container is running
    if ! docker ps --filter "name=deparrow-node" --format "table {{.Names}}\t{{.Status}}" | grep -q "Up"; then
        log_error "DEparrow container is not running"
        return 1
    fi
    
    # Check if Bacalhau is responding
    local max_retries=30
    local retry=0
    while [ $retry -lt $max_retries ]; do
        if curl -sf http://localhost:1234/api/v1/health > /dev/null 2>&1; then
            break
        fi
        sleep 2
        retry=$((retry + 1))
    done
    
    if [ $retry -eq $max_retries ]; then
        log_warning "Bacalhau API not responding yet (may still be starting)"
    else
        log_success "Installation verified successfully"
    fi
    
    return 0
}

show_summary() {
    log_success "=== DEparrow Node Deployment Complete ==="
    echo
    echo "Node Information:"
    echo "  Node ID: $NODE_ID"
    echo "  Architecture: $(uname -m)"
    echo "  CPU Cores: $(nproc)"
    echo "  Memory: $(free -h | awk '/^Mem:/ {print $2}')"
    echo "  Disk: $(df -h / | awk 'NR==2 {print $2}')"
    echo
    echo "Service Management:"
    echo "  Start:   systemctl start deparrow-node"
    echo "  Stop:    systemctl stop deparrow-node"
    echo "  Restart: systemctl restart deparrow-node"
    echo "  Status:  systemctl status deparrow-node"
    echo
    echo "Logs:"
    echo "  Service: journalctl -u deparrow-node -f"
    echo "  Container: docker logs -f deparrow-node"
    echo "  Application: tail -f $LOG_DIR/bacalhau.log"
    echo
    echo "Configuration:"
    echo "  Config: $CONFIG_DIR"
    echo "  Data: $DATA_DIR"
    echo "  Logs: $LOG_DIR"
    echo
}

main() {
    echo "=== DEparrow Alpine Linux Node Auto-Deploy ==="
    echo "This script will set up a complete DEparrow compute node"
    echo
    
    check_system_requirements
    install_docker
    setup_deparrow_user
    create_directories
    generate_node_identity
    configure_environment
    register_with_bootstrap
    create_deployment_config
    start_services
    
    if verify_installation; then
        show_summary
    else
        log_warning "Installation completed with warnings"
        echo "Check logs for details: journalctl -u deparrow-node"
    fi
}

# Run main function
main "$@"