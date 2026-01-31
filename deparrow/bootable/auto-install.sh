#!/bin/bash

# DEparrow Auto-Install Script
# For quick installation on existing Linux systems

set -e

echo "========================================="
echo "DEparrow Auto-Installer"
echo "========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
INSTALL_TYPE="node"
BOOTSTRAP_ADDRESS=""
NODE_REGION="us-east-1"
NODE_NAME="deparrow-$(hostname)"
INSTALL_DIR="/opt/deparrow"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            INSTALL_TYPE="$2"
            shift 2
            ;;
        --bootstrap)
            BOOTSTRAP_ADDRESS="$2"
            shift 2
            ;;
        --region)
            NODE_REGION="$2"
            shift 2
            ;;
        --name)
            NODE_NAME="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

show_help() {
    cat << HELP
DEparrow Auto-Installer

Installs DEparrow on existing Linux systems.

Usage: auto-install.sh [options]

Options:
  --type TYPE        Installation type: node, orchestrator, or full (default: node)
  --bootstrap ADDR   Bootstrap server address (e.g., 192.168.1.100:4222)
  --region REGION    Node region (default: us-east-1)
  --name NAME        Node name (default: hostname)
  --help             Show this help

Examples:
  # Install compute node
  sudo ./auto-install.sh --type node --bootstrap 192.168.1.100:4222
  
  # Install orchestrator
  sudo ./auto-install.sh --type orchestrator
  
  # Install full platform
  sudo ./auto-install.sh --type full
HELP
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run as root: sudo $0${NC}"
        exit 1
    fi
}

# Detect Linux distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    else
        echo "unknown"
    fi
}

# Install dependencies based on distro
install_dependencies() {
    local distro=$1
    
    echo -e "${YELLOW}Installing dependencies for $distro...${NC}"
    
    case $distro in
        ubuntu|debian)
            apt update
            apt install -y \
                docker.io \
                docker-compose \
                curl \
                python3 \
                python3-pip \
                nodejs \
                npm \
                git \
                jq
            ;;
        fedora|centos|rhel)
            dnf install -y \
                docker \
                docker-compose \
                curl \
                python3 \
                python3-pip \
                nodejs \
                npm \
                git \
                jq
            ;;
        alpine)
            apk add \
                docker \
                docker-compose \
                curl \
                python3 \
                py3-pip \
                nodejs \
                npm \
                git \
                jq
            ;;
        *)
            echo -e "${RED}Unsupported distribution: $distro${NC}"
            echo "Please install manually: Docker, Python3, Node.js, curl, git"
            exit 1
            ;;
    esac
    
    # Start and enable Docker
    systemctl enable docker
    systemctl start docker
    
    # Install Python packages
    pip3 install aiohttp pyjwt
    
    echo -e "${GREEN}Dependencies installed${NC}"
}

# Clone or copy DEparrow
setup_deparrow() {
    echo -e "${YELLOW}Setting up DEparrow...${NC}"
    
    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    
    # Check if we're in deparrow directory
    if [ -f "../bacalhau-layer/deparrow-orchestrator.yaml" ]; then
        echo "Copying DEparrow from parent directory..."
        cp -r ../bacalhau-layer "$INSTALL_DIR/"
        cp -r ../alpine-layer "$INSTALL_DIR/"
        cp -r ../metaos-layer "$INSTALL_DIR/"
        cp -r ../gui-layer "$INSTALL_DIR/"
        cp ../test-integration.sh "$INSTALL_DIR/"
        cp ../DEPLOYMENT.md "$INSTALL_DIR/"
    else
        echo "Downloading DEparrow..."
        # In production, this would clone from git
        echo -e "${YELLOW}Please place DEparrow files in $INSTALL_DIR${NC}"
        mkdir -p "$INSTALL_DIR/bacalhau-layer"
        mkdir -p "$INSTALL_DIR/alpine-layer"
        mkdir -p "$INSTALL_DIR/metaos-layer"
        mkdir -p "$INSTALL_DIR/gui-layer"
    fi
    
    # Create configuration
    cat > "$INSTALL_DIR/config.env" << CONFIG
# DEparrow Configuration
INSTALL_TYPE=$INSTALL_TYPE
NODE_NAME=$NODE_NAME
NODE_REGION=$NODE_REGION
BOOTSTRAP_ADDRESS=$BOOTSTRAP_ADDRESS
CONFIG
    
    echo -e "${GREEN}DEparrow setup complete${NC}"
}

# Configure Bacalhau
configure_bacalhau() {
    echo -e "${YELLOW}Configuring Bacalhau...${NC}"
    
    # Download Bacalhau if not present
    if ! command -v bacalhau &> /dev/null; then
        echo "Downloading Bacalhau..."
        curl -sL https://get.bacalhau.org/install.sh | bash
    fi
    
    # Update configuration files
    if [ -n "$BOOTSTRAP_ADDRESS" ]; then
        # Extract IP from bootstrap address
        BOOTSTRAP_IP=$(echo "$BOOTSTRAP_ADDRESS" | cut -d: -f1)
        
        # Update compute node config
        if [ -f "$INSTALL_DIR/bacalhau-layer/deparrow-compute.yaml" ]; then
            sed -i "s/bootstrap_addresses:.*/bootstrap_addresses: [\"/ip4/$BOOTSTRAP_IP/tcp/4222\"]/" \
                "$INSTALL_DIR/bacalhau-layer/deparrow-compute.yaml"
            
            # Add node labels
            cat >> "$INSTALL_DIR/bacalhau-layer/deparrow-compute.yaml" << LABELS
  labels:
    platform: deparrow
    role: compute
    region: $NODE_REGION
    name: $NODE_NAME
LABELS
        fi
    fi
    
    echo -e "${GREEN}Bacalhau configured${NC}"
}

# Create startup scripts
create_startup_scripts() {
    echo -e "${YELLOW}Creating startup scripts...${NC}"
    
    mkdir -p "$INSTALL_DIR/scripts"
    
    # Node startup script
    cat > "$INSTALL_DIR/scripts/start-node.sh" << 'NODE_SCRIPT'
#!/bin/bash
# DEparrow Compute Node Startup

set -e

# Load configuration
source /opt/deparrow/config.env

echo "Starting DEparrow compute node: $NODE_NAME"

# Build Alpine node image
cd /opt/deparrow/alpine-layer
chmod +x build.sh
./build.sh

# Run DEparrow node container
docker run -d \
  --name deparrow-node \
  --network host \
  --restart always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /opt/deparrow:/deparrow \
  -e NODE_NAME="$NODE_NAME" \
  -e NODE_REGION="$NODE_REGION" \
  -e BOOTSTRAP_ADDRESS="$BOOTSTRAP_ADDRESS" \
  deparrow-node:latest

echo "DEparrow compute node started"
NODE_SCRIPT
    
    # Orchestrator startup script
    cat > "$INSTALL_DIR/scripts/start-orchestrator.sh" << 'ORCH_SCRIPT'
#!/bin/bash
# DEparrow Orchestrator Startup

set -e

echo "Starting DEparrow orchestrator..."

# Start Meta-OS bootstrap server
cd /opt/deparrow/metaos-layer
python3 bootstrap-server.py > /var/log/deparrow-metaos.log 2>&1 &
METAOS_PID=$!
echo $METAOS_PID > /var/run/deparrow-metaos.pid

# Wait for bootstrap to start
sleep 5

# Start Bacalhau orchestrator
bacalhau serve --config /opt/deparrow/bacalhau-layer/deparrow-orchestrator.yaml > /var/log/deparrow-bacalhau.log 2>&1 &
BACALHAU_PID=$!
echo $BACALHAU_PID > /var/run/deparrow-bacalhau.pid

# Build and serve GUI
cd /opt/deparrow/gui-layer
npm install
npm run build
npx serve -s dist -l 3000 > /var/log/deparrow-gui.log 2>&1 &
GUI_PID=$!
echo $GUI_PID > /var/run/deparrow-gui.pid

echo "DEparrow orchestrator started"
echo "  • Meta-OS API: http://localhost:8080"
echo "  • GUI: http://localhost:3000"
echo "  • NATS: localhost:4222"
ORCH_SCRIPT
    
    # Full platform startup script
    cat > "$INSTALL_DIR/scripts/start-full.sh" << 'FULL_SCRIPT'
#!/bin/bash
# DEparrow Full Platform Startup

set -e

echo "Starting full DEparrow platform..."

# Start orchestrator components
/opt/deparrow/scripts/start-orchestrator.sh

# Wait for orchestrator to initialize
sleep 10

# Start compute node
/opt/deparrow/scripts/start-node.sh

echo "Full DEparrow platform started"
FULL_SCRIPT
    
    # Stop script
    cat > "$INSTALL_DIR/scripts/stop.sh" << 'STOP_SCRIPT'
#!/bin/bash
# Stop DEparrow

echo "Stopping DEparrow..."

# Stop containers
docker stop deparrow-node 2>/dev/null || true
docker rm deparrow-node 2>/dev/null || true

# Stop processes
for pid_file in /var/run/deparrow-*.pid; do
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        kill $pid 2>/dev/null || true
        rm -f "$pid_file"
    fi
done

echo "DEparrow stopped"
STOP_SCRIPT
    
    # Status script
    cat > "$INSTALL_DIR/scripts/status.sh" << 'STATUS_SCRIPT'
#!/bin/bash
# DEparrow Status Check

echo "DEparrow Status"
echo "==============="

# Check Docker containers
echo ""
echo "Docker Containers:"
docker ps --filter "name=deparrow" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Check processes
echo ""
echo "Processes:"
for pid_file in /var/run/deparrow-*.pid; do
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null; then
            echo "  ✓ $(basename $pid_file .pid) running (PID: $pid)"
        else
            echo "  ✗ $(basename $pid_file .pid) not running"
        fi
    fi
done

# Check services
echo ""
echo "Services:"
curl -s http://localhost:8080/api/health 2>/dev/null && echo "  ✓ Meta-OS API healthy" || echo "  ✗ Meta-OS API down"
curl -s http://localhost:3000 2>/dev/null && echo "  ✓ GUI accessible" || echo "  ✗ GUI down"
nc -z localhost 4222 2>/dev/null && echo "  ✓ NATS running" || echo "  ✗ NATS down"

# Check Bacalhau
echo ""
echo "Bacalhau:"
if command -v bacalhau &> /dev/null; then
    bacalhau node list 2>/dev/null | grep -q "ID" && echo "  ✓ Bacalhau nodes available" || echo "  ✗ No Bacalhau nodes"
else
    echo "  ✗ Bacalhau not installed"
fi
STATUS_SCRIPT
    
    # Make scripts executable
    chmod +x "$INSTALL_DIR/scripts/"*.sh
    
    echo -e "${GREEN}Startup scripts created${NC}"
}

# Create systemd service
create_systemd_service() {
    echo -e "${YELLOW}Creating systemd service...${NC}"
    
    cat > /etc/systemd/system/deparrow.service << SERVICE
[Unit]
Description=DEparrow Distributed Compute Platform
After=docker.service network.target
Requires=docker.service
Wants=network.target

[Service]
Type=forking
User=root
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/config.env
ExecStart=$INSTALL_DIR/scripts/start-$INSTALL_TYPE.sh
ExecStop=$INSTALL_DIR/scripts/stop.sh
ExecReload=$INSTALL_DIR/scripts/stop.sh && $INSTALL_DIR/scripts/start-$INSTALL_TYPE.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICE
    
    # Enable and start service
    systemctl daemon-reload
    systemctl enable deparrow.service
    
    echo -e "${GREEN}Systemd service created${NC}"
}

# Setup GUI (for orchestrator/full)
setup_gui() {
    if [ "$INSTALL_TYPE" = "node" ]; then
        return
    fi
    
    echo -e "${YELLOW}Setting up GUI...${NC}"
    
    cd "$INSTALL_DIR/gui-layer"
    
    # Install dependencies
    npm install
    
    # Create production build
    npm run build
    
    # Create nginx config for production
    cat > /etc/nginx/sites-available/deparrow << NGINX
server {
    listen 80;
    server_name _;
    
    location / {
        root $INSTALL_DIR/gui-layer/dist;
        try_files \$uri \$uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
NGINX
    
    ln -sf /etc/nginx/sites-available/deparrow /etc/nginx/sites-enabled/
    systemctl restart nginx
    
    echo -e "${GREEN}GUI setup complete${NC}"
}

# Post-install instructions
post_install() {
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}DEparrow Installation Complete!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    
    case $INSTALL_TYPE in
        node)
            echo "Compute Node Installation:"
            echo "  • Node will auto-join DEparrow network"
            echo "  • Will earn credits for compute work"
            echo ""
            echo "To start:"
            echo "  systemctl start deparrow"
            echo ""
            echo "To check status:"
            echo "  $INSTALL_DIR/scripts/status.sh"
            ;;
        orchestrator)
            echo "Orchestrator Installation:"
            echo "  • Manages DEparrow network"
            echo "  • Runs Meta-OS bootstrap server"
            echo "  • Hosts web GUI"
            echo ""
            echo "Access points:"
            echo "  • GUI: http://$(hostname -I | awk '{print $1}'):3000"
            echo "  • API: http://$(hostname -I | awk '{print $1}'):8080"
            echo "  • NATS: $(hostname -I | awk '{print $1}'):4222"
            echo ""
            echo "To start:"
            echo "  systemctl start deparrow"
            ;;
        full)
            echo "Full Platform Installation:"
            echo "  • Includes orchestrator AND compute node"
            echo "  • Complete DEparrow platform on single machine"
            echo ""
            echo "Access points:"
            echo "  • GUI: http://$(hostname -I | awk '{print $1}'):3000"
            echo "  • API: http://$(hostname -I | awk '{print $1}'):8080"
            echo ""
            echo "To start:"
            echo "  systemctl start deparrow"
            ;;
    esac
    
    echo ""
    echo "Installation directory: $INSTALL_DIR"
    echo "Configuration: $INSTALL_DIR/config.env"
    echo "Logs: /var/log/deparrow-*.log"
    echo ""
    echo "Next steps:"
    echo "1. Start DEparrow: systemctl start deparrow"
    echo "2. Check status: $INSTALL_DIR/scripts/status.sh"
    echo "3. Access GUI (orchestrator/full only)"
    echo "4. Add more compute nodes to scale"
}

# Main installation
main() {
    check_root
    
    echo -e "${BLUE}Starting DEparrow auto-install...${NC}"
    echo "Install type: $INSTALL_TYPE"
    echo "Node name: $NODE_NAME"
    echo "Region: $NODE_REGION"
    if [ -n "$BOOTSTRAP_ADDRESS" ]; then
        echo "Bootstrap: $BOOTSTRAP_ADDRESS"
    fi
    echo ""
    
    # Detect distribution
    DISTRO=$(detect_distro)
    echo "Detected distribution: $DISTRO"
    
    # Install dependencies
    install_dependencies "$DISTRO"
    
    # Setup DEparrow
    setup_deparrow
    
    # Configure Bacalhau
    configure_bacalhau
    
    # Create startup scripts
    create_startup_scripts
    
    # Create systemd service
    create_systemd_service
    
    # Setup GUI if needed
    setup_gui
    
    # Post-install
    post_install
}

# Run installation
main
