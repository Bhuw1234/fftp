#!/bin/bash
#
# DEparrow OS ISO Builder
# Builds a fully branded DEparrow OS bootable ISO
#
# This script creates a production-ready bootable ISO that:
#   - Boots on both BIOS (legacy) and EFI (UEFI) systems
#   - Includes DEparrow branding throughout
#   - Uses the deparrow CLI wrapper for all commands
#   - Auto-joins the DEparrow network on boot
#   - Supports WiFi and Ethernet connectivity
#
# Usage: ./build-deparrow-iso.sh [--wifi] [--local]
#
# Options:
#   --wifi    Include WiFi firmware (larger ISO)
#   --local   Build for local testing (QEMU)
#   --help    Show this help message
#

set -e

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
VERSION="1.0.0"

# Build options
ENABLE_WIFI=${ENABLE_WIFI:-false}
LOCAL_MODE=${LOCAL_MODE:-false}

# Build directories
BUILD_DIR="/tmp/deparrow-os-build-$$"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
BOOT_DIR="$ISO_DIR/boot"

# Module directories
BIOS_MODULES="$SCRIPT_DIR/i386-pc"
EFI_MODULES="/usr/lib/grub/x86_64-efi"

# Default bootstrap server (GCP Production)
DEFAULT_BOOTSTRAP="34.180.51.11:8080"

# ============================================
# Parse Arguments
# ============================================
show_help() {
    echo "DEparrow OS ISO Builder v${VERSION}"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --wifi     Include WiFi firmware and drivers (larger ISO)"
    echo "  --local    Build for local QEMU testing (uses localhost)"
    echo "  --help     Show this help message"
    echo ""
    echo "Output:"
    echo "  $OUTPUT_DIR/deparrow-os-${VERSION}.iso"
    echo ""
    echo "Examples:"
    echo "  $0                    # Build minimal ISO (EFI + BIOS)"
    echo "  $0 --wifi             # Build with WiFi support"
    echo "  $0 --wifi --local     # Build for QEMU testing"
    echo ""
    exit 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --wifi)
            ENABLE_WIFI=true
            shift
            ;;
        --local)
            LOCAL_MODE=true
            shift
            ;;
        --help|-h)
            show_help
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# ============================================
# Banner
# ============================================
clear
echo ""
echo "================================================"
echo "       DEPARROW OS ISO BUILDER v${VERSION}"
echo "================================================"
echo ""
echo "  Global Virtual Machine for AI Agents"
echo ' "AI Agents Buy Compute to Run Themselves"'
echo ""
echo "================================================"
echo ""

# ============================================
# Preflight Checks
# ============================================
echo "[Preflight] Checking build requirements..."

# Check for bacalhau binary
BACALHAU_BIN="$PROJECT_ROOT/bacalhau"
if [ ! -x "$BACALHAU_BIN" ]; then
    echo "  Building bacalhau binary..."
    cd "$PROJECT_ROOT"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BACALHAU_BIN" ./main.go
    echo "    bacalhau built: $(ls -lh "$BACALHAU_BIN" | awk '{print $5}')"
else
    echo "  ✓ bacalhau binary found: $(ls -lh "$BACALHAU_BIN" | awk '{print $5}')"
fi

# Check for deparrow wrapper
DEPARROW_WRAPPER="$PROJECT_ROOT/bin/deparrow"
if [ ! -x "$DEPARROW_WRAPPER" ]; then
    echo "  ERROR: deparrow wrapper not found at $DEPARROW_WRAPPER"
    echo "  Please create the wrapper first."
    exit 1
fi
echo "  ✓ deparrow wrapper found"

# Alpine minirootfs will be downloaded if not present
ALPINE_ROOTFS="$SCRIPT_DIR/alpine-minirootfs-3.20.0-x86_64.tar.gz"
ALPINE_URL="https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz"

if [ ! -f "$ALPINE_ROOTFS" ]; then
    echo "  Downloading Alpine minirootfs..."
    wget -q -O "$ALPINE_ROOTFS" "$ALPINE_URL" || {
        echo "  ERROR: Failed to download Alpine minirootfs"
        exit 1
    }
fi
echo "  ✓ Alpine minirootfs ready ($(ls -lh "$ALPINE_ROOTFS" | awk '{print $5}'))"

# Check for kernel
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "  ERROR: No kernel found in /boot"
    echo "  Install with: sudo apt-get install linux-image-generic"
    exit 1
fi
echo "  ✓ kernel found: $KERNEL_VER"

# Check for GRUB tools
if ! command -v grub-mkrescue &> /dev/null; then
    echo "  ERROR: grub-mkrescue not found"
    echo "  Install with: sudo apt-get install grub-common grub-efi-amd64-bin"
    exit 1
fi
echo "  ✓ grub-mkrescue found"

# Check for xorriso
if ! command -v xorriso &> /dev/null; then
    echo "  ERROR: xorriso not found"
    echo "  Install with: sudo apt-get install xorriso"
    exit 1
fi
echo "  ✓ xorriso found"

# Check BIOS modules
if [ -d "$BIOS_MODULES" ]; then
    echo "  ✓ BIOS modules found: $(ls "$BIOS_MODULES"/*.mod 2>/dev/null | wc -l) modules"
else
    echo "  ⚠ BIOS modules not found (ISO will be EFI-only)"
    echo "    Install with: sudo apt-get install grub-pc-bin"
fi

# Check EFI modules
if [ -d "$EFI_MODULES" ]; then
    echo "  ✓ EFI modules found"
else
    echo "  ⚠ EFI modules not found (ISO will be BIOS-only)"
    echo "    Install with: sudo apt-get install grub-efi-amd64-bin"
fi

# Check for banner file
BANNER_FILE="$SCRIPT_DIR/deparrow-banner.txt"
if [ ! -f "$BANNER_FILE" ]; then
    echo "  ⚠ Banner file not found, will use embedded banner"
else
    echo "  ✓ DEparrow banner found"
fi

echo ""

# ============================================
# Prepare Build Directory
# ============================================
echo "[Setup] Preparing build directory..."

rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log,var/lib/deparrow}
mkdir -p "$INITRAMFS_DIR"/etc/deparrow/keys
mkdir -p "$INITRAMFS_DIR"/lib/firmware
mkdir -p "$INITRAMFS_DIR"/lib/x86_64-linux-gnu
mkdir -p "$INITRAMFS_DIR"/lib64
mkdir -p "$BOOT_DIR"/grub
mkdir -p "$ISO_DIR"/{efi/boot,boot/grub/x86_64-efi,boot/grub/i386-pc}

echo "  Build directory: $BUILD_DIR"
echo ""

# ============================================
# Extract Alpine Minirootfs
# ============================================
echo "[Build] Extracting Alpine minirootfs..."

# Extract Alpine minirootfs as base (provides all busybox applets + Alpine configs)
tar -xzf "$ALPINE_ROOTFS" -C "$INITRAMFS_DIR"
echo "  ✓ Alpine minirootfs extracted (all busybox applets included)"

# Create additional directories needed by DEparrow
mkdir -p "$INITRAMFS_DIR/var/lib/deparrow"
mkdir -p "$INITRAMFS_DIR/etc/deparrow/keys"

# Copy bacalhau binary (will be wrapped by deparrow)
cp "$BACALHAU_BIN" "$INITRAMFS_DIR/bin/bacalhau"
chmod +x "$INITRAMFS_DIR/bin/bacalhau"
echo "  ✓ bacalhau binary: $(ls -lh "$INITRAMFS_DIR/bin/bacalhau" | awk '{print $5}')"

# Copy deparrow wrapper
cp "$DEPARROW_WRAPPER" "$INITRAMFS_DIR/bin/deparrow"
chmod +x "$INITRAMFS_DIR/bin/deparrow"
echo "  ✓ deparrow wrapper"

# Create symlink: bacalhau -> deparrow (for compatibility)
# This ensures any script calling bacalhau uses the deparrow wrapper
mv "$INITRAMFS_DIR/bin/bacalhau" "$INITRAMFS_DIR/bin/bacalhau.real"
ln -sf deparrow "$INITRAMFS_DIR/bin/bacalhau"
echo "  ✓ Created symlink: bacalhau -> deparrow"

# ============================================
# WiFi Support (Optional)
# ============================================
if [ "$ENABLE_WIFI" = "true" ]; then
    echo "[Build] Adding WiFi support..."

    # Helper function to copy binary with dependencies
    copy_with_deps() {
        local bin="$1"
        local dest_dir="$2"
        
        if [ ! -f "$bin" ]; then
            return 1
        fi
        
        cp "$bin" "$dest_dir/"
        
        # Copy dependencies
        ldd "$bin" 2>/dev/null | while read line; do
            lib=$(echo "$line" | grep -oE '/[^ ]+' | head -1)
            if [ -n "$lib" ] && [ -f "$lib" ]; then
                lib_dir=$(dirname "$lib")
                dest_lib_dir="$INITRAMFS_DIR$lib_dir"
                mkdir -p "$dest_lib_dir"
                cp -n "$lib" "$dest_lib_dir/" 2>/dev/null || true
            fi
        done
    }
    
    # Copy wpa_supplicant and tools
    for tool in /usr/sbin/wpa_supplicant /usr/bin/wpa_passphrase /usr/sbin/wpa_cli /usr/sbin/iwconfig /usr/sbin/iwlist; do
        if [ -f "$tool" ]; then
            copy_with_deps "$tool" "$INITRAMFS_DIR/bin"
        fi
    done
    
    # Copy essential libraries
    for lib in /lib/x86_64-linux-gnu/libnl-3.so.200 /lib/x86_64-linux-gnu/libnl-genl-3.so.200 /lib/x86_64-linux-gnu/libssl.so.3 /lib/x86_64-linux-gnu/libcrypto.so.3 /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libpthread.so.0 /lib64/ld-linux-x86-64.so.2; do
        if [ -f "$lib" ]; then
            lib_dir=$(dirname "$lib")
            mkdir -p "$INITRAMFS_DIR$lib_dir"
            cp -n "$lib" "$INITRAMFS_DIR$lib_dir/" 2>/dev/null || true
        fi
    done
    
    # Copy WiFi firmware (optimized selection)
    echo "  Copying WiFi firmware..."
    
    # Intel WiFi (common chips)
    for pattern in iwlwifi-7260 iwlwifi-7265 iwlwifi-8265 iwlwifi-9000 iwlwifi-Qu iwlwifi-3168; do
        for fw in /lib/firmware/${pattern}*.ucode.zst /lib/firmware/${pattern}*.ucode; do
            if [ -f "$fw" ]; then
                cp "$fw" "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
            fi
        done
    done
    
    # Realtek and Atheros
    for dir in rtlwifi rtw88 rtw89 ath9k_htc; do
        if [ -d "/lib/firmware/$dir" ]; then
            mkdir -p "$INITRAMFS_DIR/lib/firmware/$dir"
            cp -r /lib/firmware/$dir/* "$INITRAMFS_DIR/lib/firmware/$dir/" 2>/dev/null || true
        fi
    done
    
    # Regulatory database
    cp /lib/firmware/regulatory.db* "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
    
    FW_SIZE=$(du -sh "$INITRAMFS_DIR/lib/firmware" 2>/dev/null | awk '{print $1}')
    echo "  ✓ WiFi firmware: $FW_SIZE"
    
    mkdir -p "$INITRAMFS_DIR/var/run/wpa_supplicant"
fi

# ============================================
# Create Init Script
# ============================================
echo "[Build] Creating init script..."

# Set bootstrap server based on mode
if [ "$LOCAL_MODE" = "true" ]; then
    BOOTSTRAP_DEFAULT="localhost:8080"
else
    BOOTSTRAP_DEFAULT="$DEFAULT_BOOTSTRAP"
fi

cat > "$INITRAMFS_DIR/init" << 'INITEOF'
#!/bin/sh
#
# DEparrow OS Init - Auto-Join Network
# Automatically discovers and joins DEparrow network on boot
#

export PATH=/bin:/sbin:/usr/bin:/usr/sbin
export DEPARROW_CONFIG_DIR=/etc/deparrow
export DEPARROW_VAR_DIR=/var/lib/deparrow

# Configuration
BOOTSTRAP_ENDPOINT="${DEPARROW_BOOTSTRAP:-BOOTSTRAP_DEFAULT}"
NODE_NAME=""

# ============================================
# Mount Essential Filesystems
# ============================================
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
mkdir -p /dev/pts
mount -t devpts devpts /dev/pts

# Create device nodes
[ -e /dev/console ] || mknod /dev/console c 5 1
[ -e /dev/null ] || mknod /dev/null c 1 3
[ -e /dev/tty ] || mknod /dev/tty c 5 0
[ -e /dev/urandom ] || mknod /dev/urandom c 1 9

mount -t tmpfs tmpfs /tmp
mount -t tmpfs tmpfs /run

# ============================================
# Parse Kernel Command Line
# ============================================
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
            wifi.ssid=*)
                WIFI_SSID="${param#*=}"
                ;;
            wifi.password=*)
                WIFI_PASSWORD="${param#*=}"
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

# ============================================
# Display DEparrow Banner
# ============================================
clear
cat << 'BANNER'

  ██████╗ ███████╗██████╗ ███╗   ███╗ █████╗ ██████╗ ██████╗ 
 ██╔════╝ ██╔════╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗██╔══██╗
 ██║  ███╗█████╗  ██████╔╝██╔████╔██║███████║██████╔╝██████╔╝
 ██║   ██║██╔══╝  ██╔══██╗██║╚██╔╝██║██╔══██║██╔═══╝ ██╔═══╝ 
 ╚██████╔╝███████╗██║  ██║██║ ╚═╝ ██║██║  ██║██║     ██║     
  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     

        Global Virtual Machine for AI Agents
   "AI Agents Buy Compute to Run Themselves"
   
   Version: 1.0.0

BANNER

echo ""
echo "  ══════════════════════════════════════════════════════════"
echo ""

log() {
    echo "[$(date '+%H:%M:%S')] $1"
}

log "Node Name: $NODE_NAME"
log "Bootstrap: $BOOTSTRAP_ENDPOINT"
echo ""

# ============================================
# PHASE 1: Network Configuration
# ============================================
log "[Phase 1] Configuring network..."

ip link set lo up

# Find network interfaces
ETH_IFACES=""
WIFI_IFACES=""

for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        log "  Found interface: $iface_name"
        ip link set "$iface_name" up 2>/dev/null
        
        if [ -d "/sys/class/net/$iface_name/wireless" ]; then
            WIFI_IFACES="$WIFI_IFACES $iface_name"
        else
            ETH_IFACES="$ETH_IFACES $iface_name"
        fi
    fi
done

NETWORK_CONNECTED=false

# Try Ethernet first
if [ -n "$ETH_IFACES" ]; then
    log "  Trying Ethernet DHCP..."
    for iface in $ETH_IFACES; do
        udhcpc -i "$iface" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null &
    done
    
    WAIT_COUNT=0
    while [ $WAIT_COUNT -lt 15 ]; do
        if ip route | grep -q "default"; then
            NETWORK_CONNECTED=true
            log "  Ethernet connected!"
            break
        fi
        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
    done
fi

# Try WiFi if available
if [ "$NETWORK_CONNECTED" = "false" ] && [ -n "$WIFI_IFACES" ] && [ -n "$WIFI_SSID" ]; then
    log "  Trying WiFi: $WIFI_SSID"
    
    for iface in $WIFI_IFACES; do
        # Generate PSK and connect
        psk=$(wpa_passphrase "$WIFI_SSID" "$WIFI_PASSWORD" 2>/dev/null | grep "psk=" | head -1 | sed 's/.*psk=//')
        
        cat > /tmp/wpa_supplicant.conf << WPAEOF
ctrl_interface=/var/run/wpa_supplicant
network={
    ssid="$WIFI_SSID"
    psk="$psk"
    key_mgmt=WPA-PSK
}
WPAEOF
        
        mkdir -p /var/run/wpa_supplicant
        wpa_supplicant -B -i "$iface" -c /tmp/wpa_supplicant.conf -D nl80211,wext 2>/dev/null
        
        sleep 5
        udhcpc -i "$iface" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null
        
        if ip route | grep -q "default.*$iface"; then
            NETWORK_CONNECTED=true
            log "  WiFi connected: $WIFI_SSID"
            break
        fi
    done
fi

if [ "$NETWORK_CONNECTED" = "true" ]; then
    log "[Phase 1] Network configured successfully!"
else
    log "[Phase 1] No network connection"
fi

# ============================================
# PHASE 2: Node Identity
# ============================================
log "[Phase 2] Generating node identity..."

mkdir -p "$DEPARROW_CONFIG_DIR/keys"

if [ ! -f "$DEPARROW_CONFIG_DIR/node-id" ]; then
    NODE_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || head -c 16 /dev/urandom | xxd -p)
    echo "$NODE_ID" > "$DEPARROW_CONFIG_DIR/node-id"
fi
NODE_ID=$(cat "$DEPARROW_CONFIG_DIR/node-id")

CPU_CORES=$(nproc 2>/dev/null || echo 1)
MEMORY_KB=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}')
MEMORY_GB=$((MEMORY_KB / 1024 / 1024))

log "  Node ID: $NODE_ID"
log "  CPU: $CPU_CORES cores, Memory: ${MEMORY_GB}GB"

# ============================================
# PHASE 3: Bootstrap Discovery
# ============================================
log "[Phase 3] Discovering bootstrap server..."

BOOTSTRAP_URL=""
for proto in http https; do
    endpoint="${proto}://${BOOTSTRAP_ENDPOINT}/api/v1/health"
    if wget -q -O /dev/null --timeout=5 "$endpoint" 2>/dev/null; then
        BOOTSTRAP_URL="${proto}://${BOOTSTRAP_ENDPOINT}"
        log "  Bootstrap found: $BOOTSTRAP_URL"
        break
    fi
done

# ============================================
# PHASE 4: Node Registration
# ============================================
log "[Phase 4] Registering node..."

REGISTERED=0
ORCHESTRATOR_HOST=""
ORCHESTRATOR_PORT="4222"

if [ -n "$BOOTSTRAP_URL" ]; then
    REG_RESPONSE=$(wget -q -O - --timeout=10 \
        --header="Content-Type: application/json" \
        --post-data="{\"node_id\":\"$NODE_ID\",\"arch\":\"$(uname -m)\",\"resources\":{\"cpu\":$CPU_CORES,\"memory\":\"${MEMORY_GB}GB\"}}" \
        "${BOOTSTRAP_URL}/api/v1/nodes/register" 2>/dev/null)
    
    if [ -n "$REG_RESPONSE" ]; then
        REGISTERED=1
        NODE_TOKEN=$(echo "$REG_RESPONSE" | sed -n 's/.*"token"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        ORCHESTRATOR_HOST=$(echo "$REG_RESPONSE" | sed -n 's/.*"host"[[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
        ORCHESTRATOR_PORT=$(echo "$REG_RESPONSE" | sed -n 's/.*"port"[[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p')
        
        [ -n "$NODE_TOKEN" ] && echo "$NODE_TOKEN" > "$DEPARROW_CONFIG_DIR/node-token"
        log "  Registration successful!"
    fi
fi

# ============================================
# PHASE 5: Start Compute Node
# ============================================
log "[Phase 5] Starting DEparrow compute node..."

# Generate bacalhau config
cat > "$DEPARROW_CONFIG_DIR/bacalhau.yaml" << BACEOF
node:
  type: compute
  labels:
    deparrow: "true"
    node_id: "$NODE_ID"
compute:
  capacity:
    totalResourceLimits:
      cpu: "$CPU_CORES"
      memory: "${MEMORY_GB}GB"
BACEOF

# Start bacalhau using deparrow wrapper
if [ -n "$ORCHESTRATOR_HOST" ]; then
    deparrow serve --compute --orchestrator "${ORCHESTRATOR_HOST}:${ORCHESTRATOR_PORT}" > /var/log/bacalhau.log 2>&1 &
else
    deparrow serve --compute > /var/log/bacalhau.log 2>&1 &
fi

BACALHAU_PID=$!
sleep 2

if kill -0 $BACALHAU_PID 2>/dev/null; then
    log "  Compute node started (PID: $BACALHAU_PID)"
else
    log "  ERROR: Compute node failed to start"
fi

# ============================================
# Final Status
# ============================================
echo ""
echo "  ═════════════════════════════════════════════════════════════"
echo "  │           DEPARROW COMPUTE NODE READY                       │"
echo "  ═════════════════════════════════════════════════════════════"
echo ""
echo "  Node ID:     $NODE_ID"
echo "  Hostname:    $NODE_NAME"
echo "  CPU:         $CPU_CORES cores"
echo "  Memory:      ${MEMORY_GB}GB"
echo "  Registered:  $([ $REGISTERED -eq 1 ] && echo 'Yes' || echo 'Pending')"
echo ""
echo "  [*] This node is now earning DPC credits!"
echo ""
echo "  Commands: deparrow status, deparrow logs, deparrow credits"
echo ""

# Save environment
cat > /tmp/deparrow.env << ENVEOF
export NODE_ID="$NODE_ID"
export NODE_NAME="$NODE_NAME"
export BOOTSTRAP_URL="$BOOTSTRAP_URL"
export ORCHESTRATOR_HOST="$ORCHESTRATOR_HOST"
export ORCHESTRATOR_PORT="$ORCHESTRATOR_PORT"
export BACALHAU_PID="$BACALHAU_PID"
ENVEOF

exec /bin/sh
INITEOF

# Replace placeholder
sed -i "s|BOOTSTRAP_DEFAULT|$BOOTSTRAP_DEFAULT|g" "$INITRAMFS_DIR/init"
chmod 755 "$INITRAMFS_DIR/init"
echo "  ✓ init script created"

# ============================================
# Create DHCP Script
# ============================================
cat > "$INITRAMFS_DIR/bin/dhcp-script.sh" << 'DHCPEOF'
#!/bin/sh
[ -z "$1" ] && exit 1
case "$1" in
    deconfig)
        ip addr flush dev "$interface"
        ip link set "$interface" up
        ;;
    renew|bound)
        ip addr add "$ip/${mask:-24}" dev "$interface"
        [ -n "$router" ] && ip route add default via "$router" dev "$interface" 2>/dev/null
        if [ -n "$dns" ]; then
            > /etc/resolv.conf
            for server in $dns; do echo "nameserver $server" >> /etc/resolv.conf; done
        fi
        ;;
esac
DHCPEOF
chmod 755 "$INITRAMFS_DIR/bin/dhcp-script.sh"

# Create basic config files
touch "$INITRAMFS_DIR/etc/resolv.conf"
echo "127.0.0.1 localhost" > "$INITRAMFS_DIR/etc/hosts"

# ============================================
# Create In-RAMFS deparrow Helper
# ============================================
cat > "$INITRAMFS_DIR/bin/deparrow-commands" << 'CMDEOF'
#!/bin/sh
case "$1" in
    status)
        . /tmp/deparrow.env 2>/dev/null
        echo "=== DEparrow Node Status ==="
        echo "Node ID: $NODE_ID"
        echo "Hostname: $NODE_NAME"
        echo "Orchestrator: $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
        ;;
    logs)
        tail -50 /var/log/bacalhau.log
        ;;
    credits)
        . /tmp/deparrow.env 2>/dev/null
        [ -n "$BOOTSTRAP_URL" ] && wget -q -O - "${BOOTSTRAP_URL}/api/v1/credits/balance/${NODE_ID}"
        ;;
    help|*)
        echo "DEparrow Commands:"
        echo "  status  - Show node status"
        echo "  logs    - View compute logs"
        echo "  credits - Check DPC balance"
        ;;
esac
CMDEOF
chmod 755 "$INITRAMFS_DIR/bin/deparrow-commands"

echo "  ✓ DHCP script and helper commands"

# ============================================
# Build Initramfs
# ============================================
echo "[Build] Creating initramfs..."

cd "$INITRAMFS_DIR"
find . | cpio -H newc -o 2>/dev/null | gzip -9 > "$BOOT_DIR/initrd.img"
cd - > /dev/null

INITRD_SIZE=$(ls -lh "$BOOT_DIR/initrd.img" | awk '{print $5}')
echo "  ✓ initrd.img: $INITRD_SIZE"

# Copy kernel
cp "/boot/vmlinuz-$KERNEL_VER" "$BOOT_DIR/vmlinuz"
chmod 644 "$BOOT_DIR/vmlinuz"
KERNEL_SIZE=$(ls -lh "$BOOT_DIR/vmlinuz" | awk '{print $5}')
echo "  ✓ vmlinuz: $KERNEL_SIZE"

# ============================================
# Create GRUB Configuration
# ============================================
echo "[Build] Creating GRUB configuration..."

# Set bootstrap URL BEFORE heredoc (QEMU uses 10.0.2.2 for host)
if [ "$LOCAL_MODE" = "true" ]; then
    GRUB_BOOTSTRAP="10.0.2.2:8080"
else
    GRUB_BOOTSTRAP="34.180.51.11:8080"
fi

cat > "$ISO_DIR/boot/grub/grub.cfg" << GRUBEOF
set default=0
set timeout=5
set gfxpayload=keep

# Load graphics modules
if [ "\${grub_platform}" = "pc" ]; then
    insmod vbe
    insmod vga
fi
insmod all_video
insmod gfxterm
terminal_output gfxterm
insmod font
loadfont unicode

# DEparrow OS Header
echo ""
echo "*********************************************************"
echo "*           DEPARROW OS - Global Virtual Machine        *"
echo "*       \"AI Agents Buy Compute to Run Themselves\"       *"
echo "*********************************************************"
echo ""

menuentry "DEparrow OS - Auto-Join Network" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=${GRUB_BOOTSTRAP}
    initrd /boot/initrd.img
}

menuentry "DEparrow OS - Debug Mode" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 debug loglevel=7 deparrow.bootstrap=${GRUB_BOOTSTRAP}
    initrd /boot/initrd.img
}

menuentry "DEparrow OS - Standalone (No Network)" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=none
    initrd /boot/initrd.img
}

submenu "WiFi Configuration >>>" {
    menuentry "DEparrow OS - WiFi Mode (Interactive)" {
        echo "Enter WiFi SSID:"
        read wifi_ssid
        echo "Enter WiFi Password:"
        read wifi_password
        linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=${GRUB_BOOTSTRAP} wifi.ssid=\$wifi_ssid wifi.password=\$wifi_password
        initrd /boot/initrd.img
    }
}

menuentry "Reboot System" {
    reboot
}

menuentry "Shutdown System" {
    halt
}
GRUBEOF

echo "  ✓ grub.cfg created"

# ============================================
# Build EFI Boot Image
# ============================================
echo "[Build] Creating EFI boot image..."

if [ -d "$EFI_MODULES" ]; then
    # Create EFI boot directory
    mkdir -p "$ISO_DIR/efi/boot"
    
    # Build EFI GRUB image
    grub-mkimage -O x86_64-efi \
        -o "$ISO_DIR/efi/boot/bootx64.efi" \
        -p "/boot/grub" \
        -d "$EFI_MODULES" \
        part_gpt part_msdos iso9660 linux normal echo ls cat \
        all_video gfxterm font search gzio serial reboot halt \
        2>/dev/null || echo "  Note: EFI image may need grub-efi-amd64-bin package"
    
    if [ -f "$ISO_DIR/efi/boot/bootx64.efi" ]; then
        EFI_SIZE=$(ls -lh "$ISO_DIR/efi/boot/bootx64.efi" | awk '{print $5}')
        echo "  ✓ EFI bootloader: $EFI_SIZE"
    fi
    
    # Copy EFI modules for loading at runtime
    mkdir -p "$ISO_DIR/boot/grub/x86_64-efi"
    cp "$EFI_MODULES"/*.mod "$ISO_DIR/boot/grub/x86_64-efi/" 2>/dev/null || true
else
    echo "  ⚠ EFI modules not found, skipping EFI boot"
fi

# ============================================
# Build BIOS Boot Image
# ============================================
echo "[Build] Creating BIOS boot image..."

if [ -d "$BIOS_MODULES" ]; then
    mkdir -p "$ISO_DIR/boot/grub/i386-pc"
    
    # Copy BIOS modules
    cp "$BIOS_MODULES"/*.mod "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.lst "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.img "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    
    BIOS_MOD_COUNT=$(ls "$ISO_DIR/boot/grub/i386-pc/"*.mod 2>/dev/null | wc -l)
    echo "  ✓ BIOS modules: $BIOS_MOD_COUNT modules"
else
    echo "  ⚠ BIOS modules not found, skipping BIOS boot"
fi

# ============================================
# Build ISO
# ============================================
echo "[Build] Creating ISO image..."

mkdir -p "$OUTPUT_DIR"

# Use grub-mkrescue for hybrid boot
grub-mkrescue -o "$OUTPUT_DIR/deparrow-os-${VERSION}.iso" "$ISO_DIR" 2>&1 | tail -5

if [ ! -f "$OUTPUT_DIR/deparrow-os-${VERSION}.iso" ]; then
    echo "  ERROR: ISO creation failed"
    exit 1
fi

ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-os-${VERSION}.iso" | awk '{print $5}')

# ============================================
# Verify ISO
# ============================================
echo "[Verify] Checking ISO boot capabilities..."

xorriso -indev "$OUTPUT_DIR/deparrow-os-${VERSION}.iso" 2>&1 | grep -E "Boot record|El-Torito|platform" | head -5

# Check boot capabilities
BIOS_OK="✗"
EFI_OK="✗"

if xorriso -indev "$OUTPUT_DIR/deparrow-os-${VERSION}.iso" 2>&1 | grep -q "i386-pc"; then
    BIOS_OK="✓"
fi

if [ -f "$ISO_DIR/efi/boot/bootx64.efi" ]; then
    EFI_OK="✓"
fi

# ============================================
# Cleanup
# ============================================
rm -rf "$BUILD_DIR"

# ============================================
# Summary
# ============================================
echo ""
echo "================================================"
echo "  DEPARROW OS ISO BUILD COMPLETE!"
echo "================================================"
echo ""
echo "  Output: $OUTPUT_DIR/deparrow-os-${VERSION}.iso"
echo "  Size:   $ISO_SIZE"
echo ""
echo "  Boot Support:"
echo "    BIOS (Legacy): $BIOS_OK"
echo "    EFI (UEFI):    $EFI_OK"
echo ""
echo "  Features:"
echo "    ✓ DEparrow branding throughout"
echo "    ✓ deparrow CLI wrapper included"
echo "    ✓ Auto-join network on boot"
echo "    ✓ Compute node auto-start"
echo "    ✓ DPC credit earning enabled"
if [ "$ENABLE_WIFI" = "true" ]; then
    echo "    ✓ WiFi support included"
fi
if [ "$LOCAL_MODE" = "true" ]; then
    echo "    ✓ Local testing mode (localhost:8080)"
fi
echo ""
echo "  Test with QEMU:"
echo ""
echo "    # BIOS boot:"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-os-${VERSION}.iso -nographic"
echo ""
echo "    # EFI boot:"
echo "    qemu-system-x86_64 -m 2G \\"
echo "      -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "      -cdrom $OUTPUT_DIR/deparrow-os-${VERSION}.iso"
echo ""
echo "  Burn to USB:"
echo "    sudo dd if=$OUTPUT_DIR/deparrow-os-${VERSION}.iso of=/dev/sdX bs=4M status=progress && sync"
echo ""
