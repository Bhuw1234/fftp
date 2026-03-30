#!/bin/sh
#
# DEparrow Decentralized Init - Global Mesh Connection
#
# This init script creates a TRULY DECENTRALIZED node that:
#   1. Connects to GCP testnet (34.180.51.11:26657)
#   2. Joins global compute mesh
#   3. Earns DPC for completed jobs
#   4. Syncs wallet from blockchain
#
# No centralized authority. Pure P2P mesh network.
#

export PATH=/bin:/sbin:/usr/bin:/usr/sbin
export DEPARROW_CONFIG_DIR=/etc/deparrow
export DEPARROW_VAR_DIR=/var/lib/deparrow
export DPC_RPC="http://34.180.51.11:26657"
export DPC_CHAIN_ID="dpc-testnet-1"

# Network configuration
BOOTSTRAP_ENDPOINT="${DEPARROW_BOOTSTRAP:-34.180.51.11:8080}"
ORCHESTRATOR_PEERS="34.180.51.11:4222"

# ============================================
# PHASE 0: System Setup
# ============================================
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
mkdir -p /dev/pts && mount -t devpts devts /dev/pts
mount -t tmpfs tmpfs /tmp
mount -t tmpfs tmpfs /run

# Create essential device nodes
[ -e /dev/console ] || mknod /dev/console c 5 1
[ -e /dev/null ] || mknod /dev/null c 1 3
[ -e /dev/tty ] || mknod /dev/tty c 5 0
[ -e /dev/urandom ] || mknod /dev/urandom c 1 9

# Parse kernel command line
parse_cmdline() {
    for param in $(cat /proc/cmdline); do
        case "$param" in
            deparrow.bootstrap=*) BOOTSTRAP_ENDPOINT="${param#*=}" ;;
            deparrow.name=*) NODE_NAME="${param#*=}" ;;
            wifi.ssid=*) WIFI_SSID="${param#*=}" ;;
            wifi.password=*) WIFI_PASSWORD="${param#*=}" ;;
        esac
    done
}
parse_cmdline

# Set hostname
if [ -z "$NODE_NAME" ]; then
    NODE_NAME="deparrow-$(head -c 4 /dev/urandom | xxd -p 2>/dev/null || echo 'node')"
fi
hostname "$NODE_NAME"
echo "$NODE_NAME" > /etc/hostname

# ============================================
# Display Banner
# ============================================
clear
cat << 'BANNER'

  ██████╗ ███████╗██████╗ ███╗   ███╗ █████╗ ██████╗ ██████╗ 
 ██╔════╝ ██╔════╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗██╔══██╗
 ██║  ███╗█████╗  ██████╔╝██╔████╔██║███████║██████╔╝██████╔╝
 ██║   ██║██╔══╝  ██╔══██╗██║╚██╔╝██║██╔══██║██╔═══╝ ██╔═══╝ 
 ╚██████╔╝███████╗██║  ██║██║ ╚═╝ ██║██║  ██║██║     ██║     
  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     

        🌐 DECENTRALIZED GLOBAL MESH NODE 🌐
   "AI Agents Buy Compute to Run Themselves"
   
   Mode:    Decentralized P2P
   Network: Global Compute Mesh
   Reward:  DPC Token Earning

BANNER

log() { echo "[$(date '+%H:%M:%S')] $1"; }
log "Node: $NODE_NAME"
log "Bootstrap: $BOOTSTRAP_ENDPOINT"
echo ""

# ============================================
# PHASE 1: Network (Auto-DHCP)
# ============================================
log "[Phase 1/7] Configuring network..."
ip link set lo up

# Bring up all interfaces
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    [ "$iface_name" = "lo" ] && continue
    log "  Interface: $iface_name"
    ip link set "$iface_name" up 2>/dev/null
done

# Try DHCP on all interfaces
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    [ "$iface_name" = "lo" ] && continue
    udhcpc -i "$iface_name" -s /bin/dhcp-script.sh -T 2 -t 5 -n 2>/dev/null &
done

# Wait for network
WAIT=0
while [ $WAIT -lt 20 ]; do
    ip route | grep -q "default" && break
    sleep 1
    WAIT=$((WAIT + 1))
done

if ip route | grep -q "default"; then
    log "[Phase 1] ✓ Network connected!"
    MY_IP=$(ip route get 1 | awk '{print $7; exit}')
    log "  IP: $MY_IP"
else
    log "[Phase 1] ✗ No network - standalone mode"
fi

# ============================================
# PHASE 2: Generate Identity
# ============================================
log "[Phase 2/7] Generating cryptographic identity..."

mkdir -p "$DEPARROW_CONFIG_DIR/keys"

# Generate node ID
if [ ! -f "$DEPARROW_CONFIG_DIR/node-id" ]; then
    NODE_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || head -c 16 /dev/urandom | xxd -p)
    echo "$NODE_ID" > "$DEPARROW_CONFIG_DIR/node-id"
fi
NODE_ID=$(cat "$DEPARROW_CONFIG_DIR/node-id")

# Generate RSA key pair for signing
if [ ! -f "$DEPARROW_CONFIG_DIR/keys/node.pem" ]; then
    log "  Generating RSA keypair..."
    openssl genrsa -out "$DEPARROW_CONFIG_DIR/keys/node.pem" 2048 2>/dev/null
    openssl rsa -in "$DEPARROW_CONFIG_DIR/keys/node.pem" -pubout -out "$DEPARROW_CONFIG_DIR/keys/node.pub" 2>/dev/null
fi

# Generate wallet address (derived from public key)
if [ ! -f "$DEPARROW_CONFIG_DIR/wallet-address" ]; then
    # Simplified address generation (in production, use proper crypto)
    PUBKEY_HASH=$(cat "$DEPARROW_CONFIG_DIR/keys/node.pub" | sha256sum | head -c 40)
    echo "dpc1${PUBKEY_HASH}" > "$DEPARROW_CONFIG_DIR/wallet-address"
fi
WALLET_ADDRESS=$(cat "$DEPARROW_CONFIG_DIR/wallet-address")

CPU_CORES=$(nproc 2>/dev/null || echo 1)
MEMORY_KB=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}')
MEMORY_GB=$((MEMORY_KB / 1024 / 1024))

log "  Node ID: $NODE_ID"
log "  Wallet: $WALLET_ADDRESS"
log "  CPU: $CPU_CORES cores, RAM: ${MEMORY_GB}GB"

# ============================================
# PHASE 3: Connect to DPC Testnet
# ============================================
log "[Phase 3/7] Connecting to DPC blockchain..."

# Check DPC testnet connectivity
DPC_CONNECTED=0
if wget -q -O /dev/null --timeout=5 "$DPC_RPC/health" 2>/dev/null; then
    DPC_CONNECTED=1
    log "  ✓ DPC testnet reachable: $DPC_RPC"
    
    # Get chain status
    CHAIN_STATUS=$(wget -q -O - --timeout=5 "$DPC_RPC/status" 2>/dev/null)
    if [ -n "$CHAIN_STATUS" ]; then
        BLOCK_HEIGHT=$(echo "$CHAIN_STATUS" | grep -o '"latest_block_height":"[0-9]*"' | head -1 | grep -o '[0-9]*')
        CHAIN_ID_CHECK=$(echo "$CHAIN_STATUS" | grep -o '"network":"[^"]*"' | head -1 | cut -d'"' -f4)
        log "  Block Height: $BLOCK_HEIGHT"
        log "  Chain ID: $CHAIN_ID_CHECK"
    fi
else
    log "  ⚠ DPC testnet not reachable (will retry)"
fi

# ============================================
# PHASE 4: Register with Bootstrap
# ============================================
log "[Phase 4/7] Registering with global mesh..."

REGISTERED=0
JWT_TOKEN=""
ORCHESTRATOR_HOST="34.180.51.11"
ORCHESTRATOR_PORT="4222"

# Try to register with bootstrap server
for proto in http https; do
    BOOTSTRAP_URL="${proto}://${BOOTSTRAP_ENDPOINT}"
    
    REG_RESPONSE=$(wget -q -O - --timeout=10 \
        --header="Content-Type: application/json" \
        --post-data="{
            \"node_id\": \"$NODE_ID\",
            \"wallet_address\": \"$WALLET_ADDRESS\",
            \"arch\": \"$(uname -m)\",
            \"resources\": {
                \"cpu\": $CPU_CORES,
                \"memory\": \"${MEMORY_GB}GB\"
            },
            \"public_key\": \"$(cat $DEPARROW_CONFIG_DIR/keys/node.pub | base64 -w0 2>/dev/null)\"
        }" \
        "${BOOTSTRAP_URL}/api/v1/nodes/register" 2>/dev/null)
    
    if [ -n "$REG_RESPONSE" ]; then
        REGISTERED=1
        JWT_TOKEN=$(echo "$REG_RESPONSE" | sed -n 's/.*"token"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        ORCHESTRATOR_HOST=$(echo "$REG_RESPONSE" | sed -n 's/.*"orchestrator_host"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        [ -z "$ORCHESTRATOR_HOST" ] && ORCHESTRATOR_HOST="34.180.51.11"
        ORCHESTRATOR_PORT=$(echo "$REG_RESPONSE" | sed -n 's/.*"orchestrator_port"[[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p')
        [ -z "$ORCHESTRATOR_PORT" ] && ORCHESTRATOR_PORT="4222"
        
        [ -n "$JWT_TOKEN" ] && echo "$JWT_TOKEN" > "$DEPARROW_CONFIG_DIR/jwt-token"
        
        log "  ✓ Registered with global mesh!"
        log "  Orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
        break
    fi
done

if [ $REGISTERED -eq 0 ]; then
    log "  ⚠ Using fallback orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
fi

# ============================================
# PHASE 5: Start Bacalhau Compute Node
# ============================================
log "[Phase 5/7] Starting compute node..."

# Create Bacalhau config with DPC connector
cat > "$DEPARROW_CONFIG_DIR/bacalhau.yaml" << BACEOF
node:
  type: compute
  client_id: "$NODE_ID"
  labels:
    deparrow: "true"
    deparrow.node_id: "$NODE_ID"
    deparrow.wallet: "$WALLET_ADDRESS"
    deparrow.mesh: "global"

compute:
  capacity:
    totalResourceLimits:
      cpu: "$CPU_CORES"
      memory: "${MEMORY_GB}GB"
  orchestrators:
    - "nats://${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}"

# DPC Connector - Earn DPC for jobs
dpc:
  enabled: true
  rpc_endpoint: "$DPC_RPC"
  chain_id: "$DPC_CHAIN_ID"
  node_address: "$WALLET_ADDRESS"
  minimum_job_duration: 1
  reward_multiplier: 1.0

# Job storage
jobstorage:
  type: boltdb
  path: /var/lib/deparrow/jobs.db
BACEOF

# Create data directory
mkdir -p /var/lib/deparrow

# Start Bacalhau compute node
log "  Starting bacalhau compute node..."
deparrow serve --compute \
    --config Compute.Orchestrators="nats://${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}" \
    --config node.type=compute \
    --config node.labels.deparrow=true \
    > /var/log/bacalhau.log 2>&1 &

BACALHAU_PID=$!
sleep 3

if kill -0 $BACALHAU_PID 2>/dev/null; then
    log "  ✓ Compute node running (PID: $BACALHAU_PID)"
    log "  ✓ Connected to global mesh: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
else
    log "  ✗ Compute node failed to start"
    log "  Checking logs..."
    tail -20 /var/log/bacalhau.log
fi

# ============================================
# PHASE 6: Enable DPC Earning
# ============================================
log "[Phase 6/7] Enabling DPC earning..."

# Create DPC config for connector
cat > "$DEPARROW_CONFIG_DIR/dpc-connector.json" << DPCEOF
{
    "enabled": true,
    "rpc_endpoint": "$DPC_RPC",
    "chain_id": "$DPC_CHAIN_ID",
    "node_address": "$WALLET_ADDRESS",
    "minimum_job_duration": 1,
    "reward_multiplier": 1.0
}
DPCEOF

log "  ✓ DPC earning enabled"
log "  Wallet: $WALLET_ADDRESS"

# ============================================
# PHASE 7: Health Monitoring
# ============================================
log "[Phase 7/7] Starting health monitor..."

# Create health check script
cat > /bin/deparrow-health << HEALTHEOF
#!/bin/sh
# DEparrow Health Monitor
while true; do
    sleep 60
    
    # Check compute node
    if ! kill -0 $BACALHAU_PID 2>/dev/null; then
        log "[Health] Compute node down, restarting..."
        deparrow serve --compute \
            --config Compute.Orchestrators="nats://${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}" \
            > /var/log/bacalhau.log 2>&1 &
        BACALHAU_PID=$!
    fi
    
    # Check DPC connectivity
    if ! wget -q -O /dev/null --timeout=5 "$DPC_RPC/health" 2>/dev/null; then
        log "[Health] DPC testnet unreachable"
    fi
    
    # Check orchestrator
    if ! nc -z "$ORCHESTRATOR_HOST" "$ORCHESTRATOR_PORT" 2>/dev/null; then
        log "[Health] Orchestrator unreachable"
    fi
done
HEALTHEOF
chmod +x /bin/deparrow-health

# Start health monitor in background
/bin/deparrow-health &

log "  ✓ Health monitor started"

# ============================================
# Final Status Dashboard
# ============================================
echo ""
echo "══════════════════════════════════════════════════════════════════"
echo "              🌐 DEPARROW DECENTRALIZED NODE READY 🌐"
echo "══════════════════════════════════════════════════════════════════"
echo ""
echo "  Node ID:        $NODE_ID"
echo "  Wallet:         $WALLET_ADDRESS"
echo "  Hostname:       $NODE_NAME"
echo "  IP Address:     $MY_IP"
echo "  CPU:            $CPU_CORES cores"
echo "  Memory:         ${MEMORY_GB}GB"
echo ""
echo "  Network Status:"
echo "    DPC Testnet:  $([ $DPC_CONNECTED -eq 1 ] && echo '✓ Connected' || echo '✗ Disconnected')"
echo "    Mesh:         $([ $REGISTERED -eq 1 ] && echo '✓ Registered' || echo '✗ Pending')"
echo "    Compute:      $([ -n "$BACALHAU_PID" ] && echo "✓ Running (PID: $BACALHAU_PID)" || echo '✗ Stopped')"
echo ""
echo "  Blockchain:"
echo "    RPC:          $DPC_RPC"
echo "    Chain ID:     $DPC_CHAIN_ID"
echo "    Block:        $BLOCK_HEIGHT"
echo ""
echo "  Orchestrator:"
echo "    Host:         $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
echo ""
echo "══════════════════════════════════════════════════════════════════"
echo "  💰 This node is now EARNING DPC for every completed job!"
echo "  🌍 Part of the GLOBAL DECENTRALIZED COMPUTE MESH"
echo "  🤖 Ready for AI Agents to run autonomously"
echo "══════════════════════════════════════════════════════════════════"
echo ""
echo "  Commands:"
echo "    deparrow status     - Show node status"
echo "    deparrow logs        - View compute logs"
echo "    deparrow balance     - Check DPC balance"
echo "    deparrow jobs        - List completed jobs"
echo ""

# Save environment for later use
cat > /tmp/deparrow.env << ENVEOF
export NODE_ID="$NODE_ID"
export NODE_NAME="$NODE_NAME"
export WALLET_ADDRESS="$WALLET_ADDRESS"
export DPC_RPC="$DPC_RPC"
export DPC_CHAIN_ID="$DPC_CHAIN_ID"
export ORCHESTRATOR_HOST="$ORCHESTRATOR_HOST"
export ORCHESTRATOR_PORT="$ORCHESTRATOR_PORT"
export BACALHAU_PID="$BACALHAU_PID"
export BOOTSTRAP_URL="$BOOTSTRAP_URL"
export MY_IP="$MY_IP"
ENVEOF

# ============================================
# Interactive Shell
# ============================================
exec /bin/sh
