#!/bin/bash
# DEparrow PicoClaw Agent Auto-Start Script
# Configures and starts PicoClaw in agent mode for autonomous AI operations

set -e

# Configuration
PICOCLAW_CONFIG_DIR="/etc/picoclaw"
PICOCLAW_CONFIG_FILE="${PICOCLAW_CONFIG_DIR}/config.json"
PICOCLAW_WORKSPACE="/var/lib/deparrow/workspace"
DEPARROW_CONFIG_DIR="/etc/deparrow"
DEPARROW_API_URL="${DEPARROW_API_URL:-http://localhost:8080}"

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if PicoClaw binary exists
    if ! command -v picoclaw &> /dev/null; then
        log_error "PicoClaw binary not found"
        exit 1
    fi
    
    # Check if DEparrow API is available
    if ! curl -sf "${DEPARROW_API_URL}/health" > /dev/null 2>&1; then
        log_warning "DEparrow API not responding at ${DEPARROW_API_URL}"
        log_info "Will continue with local configuration"
    fi
    
    log_success "Prerequisites check passed"
}

# Ensure directories exist
ensure_directories() {
    log_info "Creating directory structure..."
    
    mkdir -p "$PICOCLAW_CONFIG_DIR"
    mkdir -p "$PICOCLAW_WORKSPACE"
    mkdir -p "$$PICOCLAW_WORKSPACE/skills"
    mkdir -p "$PICOCLAW_WORKSPACE/memory"
    
    chown -R deparrow:deparrow "$PICOCLAW_CONFIG_DIR"
    chown -R deparrow:deparrow "$PICOCLAW_WORKSPACE"
    
    log_success "Directory structure created"
}

# Generate default configuration if not exists
generate_config() {
    if [ -f "$PICOCLAW_CONFIG_FILE" ]; then
        log_info "PicoClaw config already exists"
        return 0
    fi
    
    log_info "Generating PicoClaw configuration..."
    
    # Get node ID if available
    NODE_ID=""
    if [ -f "${DEPARROW_CONFIG_DIR}/node-id" ]; then
        NODE_ID=$(cat "${DEPARROW_CONFIG_DIR}/node-id")
    fi
    
    # Create configuration
    cat > "$PICOCLAW_CONFIG_FILE" << EOF
{
  "agents": {
    "defaults": {
      "workspace": "$PICOCLAW_WORKSPACE",
      "model": "${PICOCLAW_MODEL:-gpt-4o-mini}",
      "max_tokens": ${PICOCLAW_MAX_TOKENS:-4096},
      "temperature": ${PICOCLAW_TEMPERATURE:-0.7}
    },
    "enabled": true,
    "auto_start": true
  },
  "deparrow": {
    "enabled": true,
    "api_url": "$DEPARROW_API_URL",
    "node_id": "$NODE_ID",
    "auto_discover": true,
    "contribute_compute": true,
    "earn_credits": true
  },
  "gateway": {
    "host": "0.0.0.0",
    "port": ${PICOCLAW_PORT:-18790},
    "enable_cors": true
  },
  "llm": {
    "provider": "${PICOCLAW_LLM_PROVIDER:-openai}",
    "api_key": "${PICOCLAW_API_KEY:-}",
    "base_url": "${PICOCLAW_BASE_URL:-https://api.openai.com/v1}"
  },
  "memory": {
    "enabled": true,
    "path": "$PICOCLAW_WORKSPACE/memory",
    "max_entries": 1000,
    "ttl_hours": 168
  },
  "skills": {
    "enabled": true,
    "path": "$PICOCLAW_WORKSPACE/skills",
    "auto_load": true
  },
  "logging": {
    "level": "${PICOCLAW_LOG_LEVEL:-info}",
    "file": "/var/log/deparrow/picoclaw.log",
    "max_size_mb": 100,
    "max_backups": 5
  }
}
EOF
    
    chown deparrow:deparrow "$PICOCLAW_CONFIG_FILE"
    chmod 600 "$PICOCLAW_CONFIG_FILE"
    
    log_success "PicoClaw configuration generated"
}

# Register agent with DEparrow network
register_agent() {
    log_info "Registering PicoClaw agent with DEparrow network..."
    
    if [ -z "$DEPARROW_API_KEY" ]; then
        log_warning "No DEparrow API key set, skipping registration"
        return 0
    fi
    
    # Get node information
    NODE_ID=$(cat "${DEPARROW_CONFIG_DIR}/node-id" 2>/dev/null || echo "unknown")
    NODE_ARCH=$(uname -m)
    NODE_CPU=$(nproc)
    NODE_MEMORY=$(free -h | awk '/^Mem:/ {print $2}')
    
    # Register agent
    local response=$(curl -sf -X POST \
        -H "Authorization: Bearer $DEPARROW_API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\n            \"node_id\": \"$NODE_ID\",\n            \"agent_type\": \"picoclaw\",\n            \"capabilities\": [\"compute\", \"ai\", \"automation\"],\n            \"resources\": {\n                \"cpu\": $NODE_CPU,\n                \"memory\": \"$NODE_MEMORY\",\n                \"arch\": \"$NODE_ARCH\"\n            }\n        }" \
        "${DEPARROW_API_URL}/api/v1/agents/register" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        log_success "Agent registered with DEparrow network"
        echo "$response" > "${PICOCLAW_CONFIG_DIR}/registration.json"
    else
        log_warning "Failed to register agent (will retry later)"
    fi
}

# Initialize PicoClaw workspace
init_workspace() {
    log_info "Initializing PicoClaw workspace..."
    
    # Run PicoClaw onboard if workspace not initialized
    if [ ! -f "${PICOCLAW_WORKSPACE}/.initialized" ]; then
        cd /var/lib/deparrow
        su - deparrow -c "picoclaw onboard --workspace $PICOCLAW_WORKSPACE" || true
        touch "${PICOCLAW_WORKSPACE}/.initialized"
        chown deparrow:deparrow "${PICOCLAW_WORKSPACE}/.initialized"
        log_success "PicoClaw workspace initialized"
    else
        log_info "Workspace already initialized"
    fi
}

# Start PicoClaw agent
start_agent() {
    log_info "Starting PicoClaw agent..."
    
    # Export environment variables for DEparrow integration
    export DEPARROW_API_URL
    export DEPARROW_NODE_ID=$(cat "${DEPARROW_CONFIG_DIR}/node-id" 2>/dev/null || echo "")
    
    # Start PicoClaw in gateway mode with agent capabilities
    cd /var/lib/deparrow
    
    # Check if running via OpenRC or directly
    if command -v rc-service &> /dev/null && rc-service picoclaw status &> /dev/null; then
        log_info "Starting via OpenRC..."
        rc-service picoclaw start
    else
        log_info "Starting PicoClaw directly..."
        su - deparrow -c "picoclaw gateway --config $PICOCLAW_CONFIG_FILE" &
        echo $! > /var/run/picoclaw.pid
    fi
    
    # Wait for gateway to start
    local max_retries=30
    local retry=0
    while [ $retry -lt $max_retries ]; do
        if curl -sf "http://localhost:18790/health" > /dev/null 2>&1; then
            log_success "PicoClaw gateway is healthy"
            return 0
        fi
        sleep 1
        retry=$((retry + 1))
    done
    
    log_warning "PicoClaw gateway health check timed out"
    return 1
}

# Enable DEparrow agent loop
enable_agent_loop() {
    log_info "Configuring DEparrow agent loop..."
    
    # Create agent loop configuration
    cat > "${PICOCLAW_CONFIG_DIR}/agent-loop.json" << EOF
{
  "enabled": true,
  "check_interval_seconds": 60,
  "tasks": {
    "credit_monitoring": {
      "enabled": true,
      "interval_seconds": 300,
      "action": "check_credits"
    },
    "job_discovery": {
      "enabled": true,
      "interval_seconds": 60,
      "action": "discover_jobs"
    },
    "compute_contribution": {
      "enabled": true,
      "interval_seconds": 30,
      "action": "contribute_compute"
    },
    "self_sustain": {
      "enabled": true,
      "interval_seconds": 600,
      "action": "ensure_self_running"
    }
  },
  "policies": {
    "min_credits": 100,
    "max_jobs_concurrent": 3,
    "priority": "normal"
  }
}
EOF
    
    chown deparrow:deparrow "${PICOCLAW_CONFIG_DIR}/agent-loop.json"
    
    log_success "Agent loop configured"
}

# Show status
show_status() {
    echo ""
    log_success "=== PicoClaw Agent Status ==="
    echo ""
    echo "Configuration:"
    echo "  Config file: $PICOCLAW_CONFIG_FILE"
    echo "  Workspace: $PICOCLAW_WORKSPACE"
    echo "  API URL: $DEPARROW_API_URL"
    echo ""
    echo "Gateway:"
    echo "  URL: http://localhost:18790"
    echo "  Health: http://localhost:18790/health"
    echo ""
    echo "Logs:"
    echo "  File: /var/log/deparrow/picoclaw.log"
    echo "  View: tail -f /var/log/deparrow/picoclaw.log"
    echo ""
}

# Main function
main() {
    echo "=== DEparrow PicoClaw Agent Startup ==="
    echo ""
    
    check_prerequisites
    ensure_directories
    generate_config
    init_workspace
    enable_agent_loop
    register_agent
    start_agent
    
    show_status
}

# Run main function
main "$@"
