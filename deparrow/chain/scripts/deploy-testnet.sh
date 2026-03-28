#!/bin/bash
#
# DPC Testnet Deployment Script for GCP VM
# ========================================
# This script deploys the DPC blockchain testnet to a GCP VM.
#
# Usage:
#   ./deploy-testnet.sh [command]
#
# Commands:
#   copy      - Copy dpcd binary to GCP VM
#   init      - Initialize DPC node on GCP VM
#   start     - Start DPC chain on GCP VM
#   stop      - Stop DPC chain on GCP VM
#   status    - Check DPC chain status
#   all       - Run all steps (copy, init, start)
#   logs      - View DPC chain logs
#

set -e

# ========================================
# Configuration
# ========================================
GCP_VM_NAME="deparrow-node"
GCP_ZONE="asia-south1-b"
GCP_IP="34.180.51.11"
DPCD_BINARY="build/dpcd-full"
REMOTE_PATH="/home/bhuwan/dpc"
CHAIN_ID="dpc-testnet-1"
MONIKER="dpc-validator-1"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# ========================================
# Commands
# ========================================

cmd_copy() {
    log_info "Copying DPC binary to GCP VM..."
    
    # Check if binary exists
    if [ ! -f "$DPCD_BINARY" ]; then
        log_error "Binary not found: $DPCD_BINARY"
        log_info "Run 'make build' first or use Docker:"
        log_info "  docker run --rm -v \$(pwd):/app -w /app golang:1.21 go build -o build/dpcd-full ./cmd/dpcd"
        exit 1
    fi
    
    # Create remote directory
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="mkdir -p $REMOTE_PATH"
    
    # Copy binary using gcloud scp
    gcloud compute scp $DPCD_BINARY $GCP_VM_NAME:$REMOTE_PATH/dpcd --zone=$GCP_ZONE
    
    # Make executable
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="chmod +x $REMOTE_PATH/dpcd"
    
    log_success "Binary copied to GCP VM: $REMOTE_PATH/dpcd"
    log_info "Binary size: $(du -h $DPCD_BINARY | cut -f1)"
}

cmd_init() {
    log_info "Initializing DPC node on GCP VM..."
    
    # Initialize the node
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        cd $REMOTE_PATH
        
        # Check if already initialized
        if [ -d ~/.dpc ]; then
            echo 'Node already initialized. Remove ~/.dpc to reinitialize.'
            exit 0
        fi
        
        # Initialize
        ./dpcd init $MONIKER --chain-id $CHAIN_ID
        
        echo ''
        echo '✓ Node initialized successfully!'
        echo 'Chain ID: $CHAIN_ID'
        echo 'Moniker: $MONIKER'
    "
    
    log_success "DPC node initialized"
    
    # Show genesis info
    log_info "Genesis configuration:"
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        cat ~/.dpc/config/genesis.json | python3 -c '
import json, sys
data = json.load(sys.stdin)
print(f\"  Chain ID: {data['chain_id']}\")
print(f\"  Genesis Time: {data['genesis_time']}\")
stake = data[\"app_state\"][\"staking\"][\"params\"]
print(f\"  Max Validators: {stake['max_validators']}\")
print(f\"  Bond Denom: {stake['bond_denom']}\")
mint = data[\"app_state\"][\"mint\"][\"params\"]
print(f\"  Inflation: {mint['inflation_rate_change']}\")
'
    " 2>/dev/null || true
}

cmd_keys() {
    log_info "Creating validator key on GCP VM..."
    
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        cd $REMOTE_PATH
        ./dpcd keys add validator --keyring-backend test
    "
    
    log_success "Validator key created"
}

cmd_start() {
    log_info "Starting DPC chain on GCP VM..."
    
    # Create systemd service file
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        sudo tee /etc/systemd/system/dpcd.service > /dev/null << 'EOF'
[Unit]
Description=DPC Blockchain Node
After=network.target

[Service]
Type=simple
User=bhuwan
WorkingDirectory=$REMOTE_PATH
ExecStart=$REMOTE_PATH/dpcd start
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

        sudo systemctl daemon-reload
        sudo systemctl enable dpcd
        sudo systemctl start dpcd
        
        echo ''
        echo '✓ DPC chain started as systemd service'
    "
    
    log_success "DPC chain started"
    log_info "Check status with: ./deploy-testnet.sh status"
}

cmd_stop() {
    log_info "Stopping DPC chain on GCP VM..."
    
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        sudo systemctl stop dpcd 2>/dev/null || true
        sudo systemctl disable dpcd 2>/dev/null || true
    "
    
    log_success "DPC chain stopped"
}

cmd_status() {
    log_info "Checking DPC chain status..."
    
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        echo '=== System Status ==='
        sudo systemctl status dpcd --no-pager 2>/dev/null || echo 'Service not running'
        
        echo ''
        echo '=== Network Ports ==='
        ss -tlnp | grep -E '26656|26657|1317|9090' || echo 'No DPC ports listening'
        
        echo ''
        echo '=== DPC Node Info ==='
        if [ -f ~/.dpc/config/genesis.json ]; then
            echo 'Node initialized: YES'
            echo \"Chain ID: \$(cat ~/.dpc/config/genesis.json | grep chain_id | cut -d'\"' -f4)\"
        else
            echo 'Node initialized: NO'
        fi
    "
}

cmd_logs() {
    log_info "Viewing DPC chain logs (Ctrl+C to exit)..."
    
    gcloud compute ssh $GCP_VM_NAME --zone=$GCP_ZONE --command="
        sudo journalctl -u dpcd -f --no-pager
    "
}

cmd_all() {
    log_info "Running full deployment pipeline..."
    echo ""
    
    cmd_copy
    echo ""
    
    cmd_init
    echo ""
    
    cmd_keys
    echo ""
    
    cmd_start
    echo ""
    
    log_success "DPC Testnet deployment complete!"
    echo ""
    echo "Endpoints:"
    echo "  - P2P:  $GCP_IP:26656"
    echo "  - RPC:  $GCP_IP:26657"
    echo "  - REST: $GCP_IP:1317"
    echo "  - gRPC: $GCP_IP:9090"
    echo ""
    echo "Useful commands:"
    echo "  ./deploy-testnet.sh status  - Check node status"
    echo "  ./deploy-testnet.sh logs    - View live logs"
    echo "  ./deploy-testnet.sh stop    - Stop the chain"
}

cmd_help() {
    echo "DPC Testnet Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  copy      Copy dpcd binary to GCP VM"
    echo "  init      Initialize DPC node on GCP VM"
    echo "  keys      Create validator key"
    echo "  start     Start DPC chain as systemd service"
    echo "  stop      Stop DPC chain"
    echo "  status    Check DPC chain status"
    echo "  logs      View DPC chain logs (live)"
    echo "  all       Run full deployment (copy, init, keys, start)"
    echo "  help      Show this help"
    echo ""
    echo "Configuration:"
    echo "  GCP VM:     $GCP_VM_NAME ($GCP_ZONE)"
    echo "  IP:         $GCP_IP"
    echo "  Chain ID:   $CHAIN_ID"
    echo "  Binary:     $DPCD_BINARY"
}

# ========================================
# Main
# ========================================

case "${1:-help}" in
    copy)   cmd_copy ;;
    init)   cmd_init ;;
    keys)   cmd_keys ;;
    start)  cmd_start ;;
    stop)   cmd_stop ;;
    status) cmd_status ;;
    logs)   cmd_logs ;;
    all)    cmd_all ;;
    help|*) cmd_help ;;
esac
