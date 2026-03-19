#!/bin/bash
# DEparrow ISO Builder v2 - Auto-Join Network ISO
# Creates a bootable ISO that auto-discovers and joins DEparrow network

set -e

echo "=== DEparrow ISO Builder v2 - Auto-Join Network ==="

BUILD_DIR="/tmp/deparrow-iso-v2"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
OUTPUT_DIR="/home/bhuwan/bacalhau/deparrow/bootable/output"
PROJECT_ROOT="/home/bhuwan/bacalhau"

# Default bootstrap endpoints (can be overridden at boot time)
DEFAULT_BOOTSTRAP="bootstrap.deparrow.net:8080"
FALLBACK_BOOTSTRAPS="bootstrap1.deparrow.net:8080,bootstrap2.deparrow.net:8080"

# Cleanup
rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log,var/lib/deparrow}
mkdir -p "$INITRAMFS_DIR"/etc/deparrow/keys
mkdir -p "$ISO_DIR/boot/grub"

# Get kernel version
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "ERROR: No kernel found in /boot"
    exit 1
fi
echo "Using kernel: $KERNEL_VER"

# Copy busybox static (provides all basic tools)
if [ -f /bin/busybox ]; then
    cp /bin/busybox "$INITRAMFS_DIR/bin/busybox"
elif [ -f /usr/bin/busybox ]; then
    cp /usr/bin/busybox "$INITRAMFS_DIR/bin/busybox"
else
    echo "ERROR: busybox not found"
    exit 1
fi
chmod +x "$INITRAMFS_DIR/bin/busybox"

# Create symlinks for busybox applets
cd "$INITRAMFS_DIR/bin"
for applet in sh cat ls mkdir mount umount sleep echo ip ln rm mv cp chmod chown grep sed awk cut head tail wget nc ping ifconfig route hostname reboot halt poweroff; do
    ln -sf busybox "$applet" 2>/dev/null || true
done
cd - > /dev/null

# Copy curl if available (needed for API calls)
if command -v curl &> /dev/null; then
    # Try to find static curl
    CURL_PATH=$(which curl)
    if ldd "$CURL_PATH" 2>/dev/null | grep -q "not a dynamic executable"; then
        cp "$CURL_PATH" "$INITRAMFS_DIR/bin/curl"
    else
        echo "Warning: curl is dynamically linked, using wget instead"
    fi
fi

# Build and copy bacalhau binary (static, no CGO)
echo "[*] Building bacalhau binary (static)..."
cd "$PROJECT_ROOT"
if [ -f "$PROJECT_ROOT/bacalhau" ]; then
    BACALHAU_BIN="$PROJECT_ROOT/bacalhau"
else
    # Build with CGO_ENABLED=0 for static binary
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BUILD_DIR/bacalhau" ./main.go
    BACALHAU_BIN="$BUILD_DIR/bacalhau"
fi
cp "$BACALHAU_BIN" "$INITRAMFS_DIR/bin/bacalhau"
chmod +x "$INITRAMFS_DIR/bin/bacalhau"
echo "    bacalhau: $(ls -lh "$INITRAMFS_DIR/bin/bacalhau" | awk '{print $5}')"

# Create init script with auto-join network functionality
cat > "$INITRAMFS_DIR/init" << 'INIT_EOF'
#!/bin/sh

# DEparrow Boot Init - Auto-Join Network
# Automatically discovers and joins DEparrow network on boot

export PATH=/bin:/usr/bin:/usr/sbin
export DEPARROW_CONFIG_DIR=/etc/deparrow
export DEPARROW_VAR_DIR=/var/lib/deparrow

# Configuration (can be overridden via kernel cmdline)
BOOTSTRAP_ENDPOINT="${DEPARROW_BOOTSTRAP:-bootstrap.deparrow.net:8080}"
NODE_NAME=""

# Logging helper
log() {
    echo "[$(date '+%H:%M:%S')] $1"
}

log_error() {
    echo "[$(date '+%H:%M:%S')] ERROR: $1" >&2
}

# Mount essential filesystems
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
mkdir -p /dev/pts
mount -t devpts devpts /dev/pts

# Create device nodes if needed
[ -e /dev/console ] || mknod /dev/console c 5 1
[ -e /dev/null ] || mknod /dev/null c 1 3
[ -e /dev/tty ] || mknod /dev/tty c 5 0
[ -e /dev/urandom ] || mknod /dev/urandom c 1 9

# Mount tmpfs for runtime
mount -t tmpfs tmpfs /tmp
mount -t tmpfs tmpfs /run

# Parse kernel command line for configuration
parse_cmdline() {
    for param in $(cat /proc/cmdline); do
        case "$param" in
            deparrow.bootstrap=*)
                BOOTSTRAP_ENDPOINT="${param#*=}"
                ;;
            deparrow.name=*)
                NODE_NAME="${param#*=}"
                ;;
            deparrow.token=*)
                DEPARROW_TOKEN="${param#*=}"
                ;;
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

# Print banner
clear
echo ""
echo "  ╔══════════════════════════════════════════════════════════════╗"
echo "  ║                                                              ║"
echo "  ║   ██████╗ ███████╗ ██████╗ █████╗ ██████╗  ██████╗ ██████╗  ║"
echo "  ║   ██╔══██╗██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔═══██╗ ║"
echo "  ║   ██║  ██║█████╗  ██║     ███████║██████╔╝██║     ██║   ██║ ║"
echo "  ║   ██║  ██║██╔══╝  ██║     ██╔══██║██╔══██╗██║     ██║   ██║ ║"
echo "  ║   ██████╔╝███████╗╚██████╗██║  ██║██║  ██║╚██████╗╚██████╔╝ ║"
echo "  ║   ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ║"
echo "  ║                                                              ║"
echo "  ║        Global Virtual Machine - Auto-Join Network            ║"
echo "  ╚══════════════════════════════════════════════════════════════╝"
echo ""
log "Node Name: $NODE_NAME"
log "Bootstrap: $BOOTSTRAP_ENDPOINT"
echo ""

# ============================================
# PHASE 1: Network Configuration
# ============================================
log "[Phase 1] Configuring network..."

ip link set lo up

# Find and configure network interfaces
IFACES=""
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        log "  Found interface: $iface_name"
        ip link set "$iface_name" up 2>/dev/null
        IFACES="$IFACES $iface_name"
    fi
done

if [ -z "$IFACES" ]; then
    log_error "No network interfaces found!"
else
    # Start DHCP on all interfaces
    log "[Phase 1] Requesting DHCP leases..."
    for iface in $IFACES; do
        udhcpc -i "$iface" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null &
    done
    
    # Wait for network
    log "[Phase 1] Waiting for network connectivity..."
    WAIT_COUNT=0
    MAX_WAIT=30
    while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
        if ip route | grep -q "default"; then
            log "[Phase 1] Network configured successfully!"
            break
        fi
        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
        printf "."
    done
    echo ""
fi

# Show network status
echo ""
echo "  === Network Status ==="
ip addr show 2>/dev/null | grep -E "inet " | while read line; do
    echo "  $line"
done
echo "  Gateway: $(ip route | grep default | awk '{print $3}')"
echo ""

# ============================================
# PHASE 2: Node Identity Generation
# ============================================
log "[Phase 2] Generating node identity..."

mkdir -p "$DEPARROW_CONFIG_DIR/keys"

# Generate node ID if not exists
if [ ! -f "$DEPARROW_CONFIG_DIR/node-id" ]; then
    NODE_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || head -c 16 /dev/urandom | xxd -p)
    echo "$NODE_ID" > "$DEPARROW_CONFIG_DIR/node-id"
    chmod 600 "$DEPARROW_CONFIG_DIR/node-id"
fi
NODE_ID=$(cat "$DEPARROW_CONFIG_DIR/node-id")
log "  Node ID: $NODE_ID"

# Generate RSA key pair for node authentication
if [ ! -f "$DEPARROW_CONFIG_DIR/keys/private.pem" ]; then
    # Use openssl if available, otherwise generate a simple key
    if command -v openssl >/dev/null 2>&1; then
        openssl genrsa -out "$DEPARROW_CONFIG_DIR/keys/private.pem" 2048 2>/dev/null
        openssl rsa -in "$DEPARROW_CONFIG_DIR/keys/private.pem" -pubout -out "$DEPARROW_CONFIG_DIR/keys/public.pem" 2>/dev/null
    else
        # Fallback: generate random key material
        head -c 32 /dev/urandom | base64 > "$DEPARROW_CONFIG_DIR/keys/private.pem"
        head -c 32 /dev/urandom | base64 > "$DEPARROW_CONFIG_DIR/keys/public.pem"
    fi
    chmod 600 "$DEPARROW_CONFIG_DIR/keys/private.pem"
    chmod 644 "$DEPARROW_CONFIG_DIR/keys/public.pem"
fi
PUBLIC_KEY=$(cat "$DEPARROW_CONFIG_DIR/keys/public.pem" | tr -d '\n')
log "  Public key generated"

# Detect system resources
CPU_CORES=$(nproc 2>/dev/null || echo 1)
MEMORY_KB=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}')
MEMORY_GB=$((MEMORY_KB / 1024 / 1024))
ARCH=$(uname -m)

log "  CPU: $CPU_CORES cores"
log "  Memory: ${MEMORY_GB}GB"
log "  Architecture: $ARCH"

# ============================================
# PHASE 3: Bootstrap Server Discovery
# ============================================
log "[Phase 3] Discovering bootstrap server..."

# Try multiple bootstrap endpoints
BOOTSTRAP_URL=""
HEALTH_ENDPOINTS="
    http://${BOOTSTRAP_ENDPOINT}/api/v1/health
    https://${BOOTSTRAP_ENDPOINT}/api/v1/health
"

for endpoint in $HEALTH_ENDPOINTS; do
    log "  Trying: $endpoint"
    if wget -q -O /dev/null --timeout=5 "$endpoint" 2>/dev/null; then
        BOOTSTRAP_URL=$(echo "$endpoint" | sed 's|/api/v1/health||')
        log "  Bootstrap server found: $BOOTSTRAP_URL"
        break
    fi
done

if [ -z "$BOOTSTRAP_URL" ]; then
    log "  Warning: No bootstrap server found, will retry in background"
fi

# ============================================
# PHASE 4: Node Registration
# ============================================
log "[Phase 4] Registering node with DEparrow network..."

REGISTERED=0
NODE_TOKEN=""
ORCHESTRATOR_HOST=""
ORCHESTRATOR_PORT=""

if [ -n "$BOOTSTRAP_URL" ]; then
    # Prepare registration payload
    cat > /tmp/register.json << REGEOF
{
    "node_id": "$NODE_ID",
    "public_key": "$PUBLIC_KEY",
    "arch": "$ARCH",
    "resources": {
        "cpu": $CPU_CORES,
        "memory": "${MEMORY_GB}GB"
    },
    "labels": {
        "hostname": "$NODE_NAME",
        "auto_join": "true"
    }
}
REGEOF

    # Register with bootstrap server
    REG_RESPONSE=$(wget -q -O - --timeout=10 \
        --header="Content-Type: application/json" \
        --post-file=/tmp/register.json \
        "${BOOTSTRAP_URL}/api/v1/nodes/register" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$REG_RESPONSE" ]; then
        log "  Registration successful!"
        
        # Parse response (simple JSON parsing without jq)
        NODE_TOKEN=$(echo "$REG_RESPONSE" | sed -n 's/.*"token"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        ORCHESTRATOR_HOST=$(echo "$REG_RESPONSE" | sed -n 's/.*"host"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)
        ORCHESTRATOR_PORT=$(echo "$REG_RESPONSE" | sed -n 's/.*"port"[[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p' | head -1)
        
        if [ -z "$ORCHESTRATOR_PORT" ]; then
            ORCHESTRATOR_PORT="4222"
        fi
        
        # Save token for later use
        if [ -n "$NODE_TOKEN" ]; then
            echo "$NODE_TOKEN" > "$DEPARROW_CONFIG_DIR/node-token"
            chmod 600 "$DEPARROW_CONFIG_DIR/node-token"
        fi
        
        REGISTERED=1
        log "  Token received: ${NODE_TOKEN:0:20}..."
        log "  Orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
    else
        log "  Registration failed, will use default configuration"
    fi
fi

# ============================================
# PHASE 5: Bacalhau Configuration
# ============================================
log "[Phase 5] Generating bacalhau configuration..."

mkdir -p "$DEPARROW_CONFIG_DIR"

# Generate bacalhau config
cat > "$DEPARROW_CONFIG_DIR/bacalhau.yaml" << BACEOF
# DEparrow Compute Node Configuration
# Auto-generated on boot

node:
  type: compute
  labels:
    deparrow: "true"
    auto_join: "true"
    node_id: "$NODE_ID"
    hostname: "$NODE_NAME"

compute:
  capacity:
    totalResourceLimits:
      cpu: "$CPU_CORES"
      memory: "${MEMORY_GB}GB"
  
  execution:
    engines:
      - docker
      - wasm
    timeout: 1h

$(if [ -n "$ORCHESTRATOR_HOST" ]; then
    echo "orchestrators:"
    echo "  - \"${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}\""
fi)

identity:
  nodeId: "$NODE_ID"
BACEOF

log "  Configuration written to $DEPARROW_CONFIG_DIR/bacalhau.yaml"

# ============================================
# PHASE 6: Start Bacalhau Compute Node
# ============================================
log "[Phase 6] Starting bacalhau compute node..."

# Start bacalhau in background
if [ -n "$ORCHESTRATOR_HOST" ]; then
    log "  Connecting to orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
    bacalhau serve --compute \
        --config "$DEPARROW_CONFIG_DIR/bacalhau.yaml" \
        --orchestrator "${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}" \
        > /var/log/bacalhau.log 2>&1 &
else
    log "  Starting in standalone mode (no orchestrator)"
    bacalhau serve --compute \
        --config "$DEPARROW_CONFIG_DIR/bacalhau.yaml" \
        > /var/log/bacalhau.log 2>&1 &
fi

BACALHAU_PID=$!
sleep 2

# Check if bacalhau started
if kill -0 $BACALHAU_PID 2>/dev/null; then
    log "  Bacalhau started successfully (PID: $BACALHAU_PID)"
else
    log_error "Bacalhau failed to start. Check /var/log/bacalhau.log"
fi

# ============================================
# FINAL: Display Status
# ============================================
echo ""
echo "  ╔══════════════════════════════════════════════════════════════╗"
echo "  ║              DEPARROW COMPUTE NODE ONLINE                     ║"
echo "  ╠══════════════════════════════════════════════════════════════╣"
echo "  ║  Node ID:     $NODE_ID                    ║"
echo "  ║  Hostname:    $NODE_NAME                                   ║"
echo "  ║  CPU:         $CPU_CORES cores                                    ║"
echo "  ║  Memory:      ${MEMORY_GB}GB                                         ║"
echo "  ║  Registered:  $([ $REGISTERED -eq 1 ] && echo 'Yes' || echo 'Pending')                                         ║"
$(if [ -n "$ORCHESTRATOR_HOST" ]; then
    echo "  ║  Orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT                        ║"
fi)
echo "  ║                                                              ║"
echo "  ║  🚀 This node is now earning credits!                        ║"
echo "  ╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "  Commands:"
echo "    status    - Show node status"
echo "    logs      - View bacalhau logs"
echo "    credits   - Check earned credits"
echo "    help      - Show all commands"
echo ""

# Save environment for scripts
cat > /tmp/deparrow.env << ENVEOF
export NODE_ID="$NODE_ID"
export NODE_NAME="$NODE_NAME"
export BOOTSTRAP_URL="$BOOTSTRAP_URL"
export ORCHESTRATOR_HOST="$ORCHESTRATOR_HOST"
export ORCHESTRATOR_PORT="$ORCHESTRATOR_PORT"
export BACALHAU_PID="$BACALHAU_PID"
export REGISTERED="$REGISTERED"
ENVEOF

# ============================================
# Background Services
# ============================================

# Start heartbeat daemon (sends periodic heartbeats to bootstrap)
start_heartbeat_daemon() {
    (
        while true; do
            if [ -n "$BOOTSTRAP_URL" ] && [ -n "$NODE_TOKEN" ]; then
                wget -q -O /dev/null --timeout=5 \
                    --header="Authorization: Bearer $NODE_TOKEN" \
                    "${BOOTSTRAP_URL}/api/v1/nodes/${NODE_ID}/heartbeat" 2>/dev/null
            fi
            sleep 60
        done
    ) &
    log "Heartbeat daemon started"
}

# Start credit reporter (logs credit earnings)
start_credit_reporter() {
    (
        while true; do
            if [ -n "$BOOTSTRAP_URL" ] && [ -n "$NODE_TOKEN" ]; then
                CREDITS=$(wget -q -O - --timeout=5 \
                    --header="Authorization: Bearer $NODE_TOKEN" \
                    "${BOOTSTRAP_URL}/api/v1/credits/balance/${NODE_ID}" 2>/dev/null | \
                    sed -n 's/.*"balance"[[[:space:]]*:[[:space:]]*\([0-9.]*\).*/\1/p')
                if [ -n "$CREDITS" ]; then
                    log "Credits earned: $CREDITS DPC"
                fi
            fi
            sleep 300  # Check every 5 minutes
        done
    ) &
    log "Credit reporter started"
}

if [ $REGISTERED -eq 1 ]; then
    start_heartbeat_daemon
    start_credit_reporter
fi

# ============================================
# Interactive Shell
# ============================================
exec /bin/sh
INIT_EOF

chmod 755 "$INITRAMFS_DIR/init"

# Create DHCP script for udhcpc
cat > "$INITRAMFS_DIR/bin/dhcp-script.sh" << 'DHCP_EOF'
#!/bin/sh
# udhcpc script for DEparrow

[ -z "$1" ] && echo "Error: should be called from udhcpc" && exit 1

case "$1" in
    deconfig)
        ip addr flush dev "$interface"
        ip link set "$interface" up
        ;;
    renew|bound)
        # Configure IP address
        ip addr add "$ip/${mask:-24}" dev "$interface"
        
        # Add default route
        if [ -n "$router" ]; then
            ip route add default via "$router" dev "$interface" 2>/dev/null
        fi
        
        # Configure DNS
        if [ -n "$dns" ]; then
            > /etc/resolv.conf
            for server in $dns; do
                echo "nameserver $server" >> /etc/resolv.conf
            done
        fi
        
        # Log configuration
        echo "[DHCP] $interface: $ip/$mask via $router" > /dev/console 2>&1
        ;;
    nak)
        echo "[DHCP] NAK received for $interface" > /dev/console 2>&1
        ;;
esac
DHCP_EOF
chmod 755 "$INITRAMFS_DIR/bin/dhcp-script.sh"

# Also create simple.script as symlink for compatibility
ln -sf dhcp-script.sh "$INITRAMFS_DIR/bin/simple.script"

# Create basic resolv.conf and hosts
touch "$INITRAMFS_DIR/etc/resolv.conf"
echo "127.0.0.1 localhost" > "$INITRAMFS_DIR/etc/hosts"
echo "::1 localhost" >> "$INITRAMFS_DIR/etc/hosts"

# Create simple command helper script
cat > "$INITRAMFS_DIR/bin/deparrow" << 'DEPEOF'
#!/bin/sh
# DEparrow CLI helper

case "$1" in
    status)
        . /tmp/deparrow.env 2>/dev/null
        echo "=== DEparrow Node Status ==="
        echo "Node ID: $NODE_ID"
        echo "Hostname: $NODE_NAME"
        echo "Registered: $REGISTERED"
        echo "Orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
        if [ -n "$BACALHAU_PID" ] && kill -0 "$BACALHAU_PID" 2>/dev/null; then
            echo "Bacalhau: Running (PID $BACALHAU_PID)"
        else
            echo "Bacalhau: Not running"
        fi
        ;;
    logs)
        tail -100 /var/log/bacalhau.log
        ;;
    credits)
        . /tmp/deparrow.env 2>/dev/null
        if [ -n "$BOOTSTRAP_URL" ] && [ -n "$NODE_TOKEN" ]; then
            wget -q -O - --timeout=5 \
                --header="Authorization: Bearer $NODE_TOKEN" \
                "${BOOTSTRAP_URL}/api/v1/credits/balance/${NODE_ID}" 2>/dev/null
        else
            echo "Not registered with network"
        fi
        ;;
    register)
        . /tmp/deparrow.env 2>/dev/null
        echo "Attempting manual registration..."
        wget -q -O - --timeout=10 \
            --header="Content-Type: application/json" \
            --post-data "{\"node_id\":\"$NODE_ID\",\"public_key\":\"$(cat /etc/deparrow/keys/public.pem | tr -d '\n')\"}" \
            "${BOOTSTRAP_URL}/api/v1/nodes/register" 2>/dev/null
        ;;
    restart)
        . /tmp/deparrow.env 2>/dev/null
        if [ -n "$BACALHAU_PID" ]; then
            kill "$BACALHAU_PID" 2>/dev/null
            sleep 2
        fi
        bacalhau serve --compute --config /etc/deparrow/bacalhau.yaml > /var/log/bacalhau.log 2>&1 &
        echo "Bacalhau restarted"
        ;;
    help|*)
        echo "DEparrow Commands:"
        echo "  status    - Show node status"
        echo "  logs      - View bacalhau logs"
        echo "  credits   - Check earned credits"
        echo "  register  - Manually register with network"
        echo "  restart   - Restart bacalhau compute node"
        echo "  help      - Show this help"
        ;;
esac
DEPEOF
chmod 755 "$INITRAMFS_DIR/bin/deparrow"

# Build initramfs (cpio newc format)
echo "[*] Building initramfs..."
cd "$INITRAMFS_DIR"

# Calculate sizes before compression
BACALHAU_SIZE=$(ls -lh bin/bacalhau 2>/dev/null | awk '{print $5}')
BUSYBOX_SIZE=$(ls -lh bin/busybox 2>/dev/null | awk '{print $5}')
echo "    Components:"
echo "      busybox: $BUSYBOX_SIZE"
echo "      bacalhau: $BACALHAU_SIZE"

find . | cpio -H newc -o 2>/dev/null | gzip -9 > "$ISO_DIR/boot/initrd.img"
INITRD_SIZE=$(ls -lh "$ISO_DIR/boot/initrd.img" | awk '{print $5}')
echo "    initrd.img: $INITRD_SIZE (compressed)"
cd - > /dev/null

# Copy kernel - try multiple sources
echo "[*] Copying kernel..."

# Option 1: Try reading from /boot directly
if [ -r "/boot/vmlinuz-$KERNEL_VER" ]; then
    cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz"
    echo "    Copied from /boot/vmlinuz-$KERNEL_VER"
# Option 2: Extract from existing Alpine ISO if available
elif [ -f "$OUTPUT_DIR/deparrow-alpine.iso" ]; then
    echo "    Extracting kernel from Alpine ISO..."
    EXTRACT_DIR="/tmp/kernel-extract-$"
    mkdir -p "$EXTRACT_DIR"
    xorriso -osirrox on -indev "$OUTPUT_DIR/deparrow-alpine.iso" -extract / "$EXTRACT_DIR" 2>/dev/null
    if [ -f "$EXTRACT_DIR/boot/vmlinuz" ]; then
        cp "$EXTRACT_DIR/boot/vmlinuz" "$ISO_DIR/boot/vmlinuz"
        echo "    Extracted from Alpine ISO"
    fi
    rm -rf "$EXTRACT_DIR"
# Option 3: Try with sudo
else
    echo "    Need sudo to read kernel..."
    sudo cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz" 2>/dev/null && \
        sudo chown $USER:$USER "$ISO_DIR/boot/vmlinuz" || {
        echo "ERROR: Cannot access kernel. Try: sudo chmod 644 /boot/vmlinuz-*"
        exit 1
    }
fi

if [ ! -f "$ISO_DIR/boot/vmlinuz" ]; then
    echo "ERROR: Failed to copy kernel"
    exit 1
fi
chmod 644 "$ISO_DIR/boot/vmlinuz"
KERNEL_SIZE=$(ls -lh "$ISO_DIR/boot/vmlinuz" | awk '{print $5}')
echo "    vmlinuz: $KERNEL_SIZE"

# Create GRUB config with kernel parameters for auto-join
# IMPORTANT: grub.cfg must be at boot/grub/grub.cfg for grub-mkrescue to embed it
cat > "$ISO_DIR/boot/grub/grub.cfg" << 'GRUB_EOF'
set default=0
set timeout=5

insmod all_video
insmod gfxterm
terminal_output gfxterm

menuentry "DEparrow Compute Node - Auto-Join Network" {
    linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=bootstrap.deparrow.net:8080
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node - Debug Mode" {
    linux /boot/vmlinuz console=ttyS0,115200n8 debug loglevel=7
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node - Standalone (No Network)" {
    linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=none
    initrd /boot/initrd.img
}

menuentry "Reboot" {
    reboot
}

menuentry "Shutdown" {
    halt
}
GRUB_EOF

echo "[*] GRUB config created at $ISO_DIR/boot/grub/grub.cfg"

# Build ISO using grub-mkrescue (supports both EFI and BIOS boot)
# This is the correct method - it automatically embeds grub.cfg into the boot image
echo "[*] Building ISO with grub-mkrescue (EFI + BIOS support)..."
mkdir -p "$OUTPUT_DIR"
grub-mkrescue -o "$OUTPUT_DIR/deparrow-autojoin.iso" "$ISO_DIR" 2>&1 | tail -5

ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-autojoin.iso" 2>/dev/null | awk '{print $5}')

echo ""
echo "========================================="
echo "  DEPARROW AUTO-JOIN ISO BUILD COMPLETE"
echo "========================================="
echo ""
echo "  ISO: $OUTPUT_DIR/deparrow-autojoin.iso"
echo "  Size: $ISO_SIZE"
echo ""
echo "  Features:"
echo "    ✓ Auto-discover bootstrap server"
echo "    ✓ Auto-register node identity"
echo "    ✓ Auto-connect to orchestrator"
echo "    ✓ Bacalhau compute node auto-start"
echo "    ✓ Credit earning enabled"
echo ""
echo "  Test with QEMU:"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-autojoin.iso"
echo ""
echo "  For EFI boot:"
echo "    qemu-system-x86_64 -m 2G \\"
echo "      -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "      -drive if=pflash,format=raw,file=/tmp/ovmf_vars.fd \\"
echo "      -cdrom $OUTPUT_DIR/deparrow-autojoin.iso"
echo ""

