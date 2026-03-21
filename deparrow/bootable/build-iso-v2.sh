#!/bin/bash
# DEparrow ISO Builder v2 - Auto-Join Network ISO
# Creates a bootable ISO that auto-discovers and joins DEparrow network
# Supports both Ethernet (DHCP) and WiFi (WPA2-PSK)

set -e

echo "=== DEparrow ISO Builder v2 - Auto-Join Network with WiFi ==="

BUILD_DIR="/tmp/deparrow-iso-v2"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
OUTPUT_DIR="/home/bhuwan/bacalhau/deparrow/bootable/output"
PROJECT_ROOT="/home/bhuwan/bacalhau"

# Default bootstrap endpoints (can be overridden at boot time)
DEFAULT_BOOTSTRAP="bootstrap.deparrow.net:8080"
FALLBACK_BOOTSTRAPS="bootstrap1.deparrow.net:8080,bootstrap2.deparrow.net:8080"

# WiFi support flag
ENABLE_WIFI=${ENABLE_WIFI:-true}

# Cleanup
rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log,var/lib/deparrow}
mkdir -p "$INITRAMFS_DIR"/etc/deparrow/keys
mkdir -p "$INITRAMFS_DIR"/lib/firmware
mkdir -p "$INITRAMFS_DIR"/lib/x86_64-linux-gnu
mkdir -p "$INITRAMFS_DIR"/lib64
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

# ============================================
# WiFi Support
# ============================================
if [ "$ENABLE_WIFI" = "true" ]; then
    echo "[*] Adding WiFi support..."
    
    # Helper function to copy binary with dependencies
    copy_with_deps() {
        local bin="$1"
        local dest_dir="$2"
        
        if [ ! -f "$bin" ]; then
            echo "    Warning: $bin not found"
            return 1
        fi
        
        # Copy the binary
        cp "$bin" "$dest_dir/"
        
        # Copy dependencies
        ldd "$bin" 2>/dev/null | while read line; do
            lib=$(echo "$line" | grep -oE '/[^ ]+' | head -1)
            if [ -n "$lib" ] && [ -f "$lib" ]; then
                lib_name=$(basename "$lib")
                lib_dir=$(dirname "$lib")
                dest_lib_dir="$INITRAMFS_DIR$lib_dir"
                mkdir -p "$dest_lib_dir"
                cp -n "$lib" "$dest_lib_dir/" 2>/dev/null || true
            fi
        done
        
        # Copy the dynamic linker
        if [ -f /lib64/ld-linux-x86-64.so.2 ]; then
            cp -n /lib64/ld-linux-x86-64.so.2 "$INITRAMFS_DIR/lib64/" 2>/dev/null || true
        fi
    }
    
    # Copy wpa_supplicant
    if [ -f /usr/sbin/wpa_supplicant ]; then
        echo "    Adding wpa_supplicant..."
        copy_with_deps /usr/sbin/wpa_supplicant "$INITRAMFS_DIR/usr/sbin"
        cp /usr/sbin/wpa_supplicant "$INITRAMFS_DIR/bin/" 2>/dev/null || true
    fi
    
    # Copy wpa_passphrase (for generating PSK)
    if [ -f /usr/bin/wpa_passphrase ]; then
        echo "    Adding wpa_passphrase..."
        copy_with_deps /usr/bin/wpa_passphrase "$INITRAMFS_DIR/usr/bin"
        cp /usr/bin/wpa_passphrase "$INITRAMFS_DIR/bin/" 2>/dev/null || true
    fi
    
    # Copy wpa_cli
    if [ -f /usr/sbin/wpa_cli ]; then
        echo "    Adding wpa_cli..."
        copy_with_deps /usr/sbin/wpa_cli "$INITRAMFS_DIR/usr/sbin"
    fi
    
    # Copy iwconfig (wireless-tools) - simpler dependency chain
    if [ -f /usr/sbin/iwconfig ]; then
        echo "    Adding iwconfig..."
        copy_with_deps /usr/sbin/iwconfig "$INITRAMFS_DIR/usr/sbin"
        cp /usr/sbin/iwconfig "$INITRAMFS_DIR/bin/" 2>/dev/null || true
        
        # Also copy iwlist for scanning
        if [ -f /usr/sbin/iwlist ]; then
            copy_with_deps /usr/sbin/iwlist "$INITRAMFS_DIR/usr/sbin"
            cp /usr/sbin/iwlist "$INITRAMFS_DIR/bin/" 2>/dev/null || true
        fi
    fi
    
    # Copy iw tool if available (more modern)
    if command -v iw &> /dev/null; then
        echo "    Adding iw tool..."
        copy_with_deps "$(which iw)" "$INITRAMFS_DIR/usr/sbin"
    fi
    
    # Copy essential shared libraries
    echo "    Copying essential libraries..."
    for lib in \
        /lib/x86_64-linux-gnu/libnl-3.so.200 \
        /lib/x86_64-linux-gnu/libnl-genl-3.so.200 \
        /lib/x86_64-linux-gnu/libnl-route-3.so.200 \
        /lib/x86_64-linux-gnu/libssl.so.3 \
        /lib/x86_64-linux-gnu/libcrypto.so.3 \
        /lib/x86_64-linux-gnu/libdbus-1.so.3 \
        /lib/x86_64-linux-gnu/libiw.so.30 \
        /lib/x86_64-linux-gnu/libm.so.6 \
        /lib/x86_64-linux-gnu/libc.so.6 \
        /lib/x86_64-linux-gnu/libpthread.so.0 \
        /lib/x86_64-linux-gnu/libdl.so.2 \
        /lib/x86_64-linux-gnu/librt.so.1 \
    ; do
        if [ -f "$lib" ]; then
            lib_name=$(basename "$lib")
            cp -n "$lib" "$INITRAMFS_DIR/lib/x86_64-linux-gnu/" 2>/dev/null || true
        fi
    done
    
    # Copy additional libraries that wpa_supplicant needs
    for lib in \
        /lib/x86_64-linux-gnu/libgcrypt.so.20 \
        /lib/x86_64-linux-gnu/libgpg-error.so.0 \
        /lib/x86_64-linux-gnu/libsystemd.so.0 \
        /lib/x86_64-linux-gnu/libcap.so.2 \
        /lib/x86_64-linux-gnu/liblz4.so.1 \
        /lib/x86_64-linux-gnu/liblzma.so.5 \
        /lib/x86_64-linux-gnu/libzstd.so.1 \
        /lib/x86_64-linux-gnu/libpcsclite.so.1 \
    ; do
        if [ -f "$lib" ]; then
            cp -n "$lib" "$INITRAMFS_DIR/lib/x86_64-linux-gnu/" 2>/dev/null || true
        fi
    done
    
    # Copy WiFi firmware - Essential chipsets only (optimized for size)
    echo "    Copying WiFi firmware (optimized selection)..."
    FIRMWARE_SIZE=0
    
    # Intel iwlwifi - Common chips only (NOT all 183 files!)
    # Include only: 7260, 7265, 8265, 9000, AX200/201/210
    mkdir -p "$INITRAMFS_DIR/lib/firmware"
    
    # Common Intel WiFi chips (most popular)
    INTEL_FW_PATTERNS=(
        "iwlwifi-7260-*"           # Intel 7260 (common laptop)
        "iwlwifi-7265-*"           # Intel 7265
        "iwlwifi-7265D-*"          # Intel 7265D
        "iwlwifi-8265-*"           # Intel 8265 (common laptop)
        "iwlwifi-9000-pu-b0-jf-b0-*"  # Intel 9260/9560
        "iwlwifi-Qu-*"             # Intel AX200/AX201
        "iwlwifi-cc-a0-*"          # Intel AX200
        "iwlwifi-ty-a0-*"          # Intel AX210
        "iwlwifi-3160-*"           # Intel 3160
        "iwlwifi-3168-*"           # Intel 3168 (cheap adapters)
    )
    
    for pattern in "${INTEL_FW_PATTERNS[@]}"; do
        for fw in /lib/firmware/$pattern.ucode.zst; do
            if [ -f "$fw" ]; then
                # Only copy the latest version (highest number)
                latest=$(ls /lib/firmware/$pattern.ucode.zst 2>/dev/null | sort -V | tail -1)
                if [ "$fw" = "$latest" ]; then
                    cp "$fw" "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
                fi
            fi
        done
    done
    
    # Copy pnvm files for Intel AX chips
    for pnvm in /lib/firmware/iwlwifi-*.pnvm.zst; do
        if [ -f "$pnvm" ]; then
            case "$pnvm" in
                *bz-b0*|*ty-a0*|*cc-a0*|*Qu-*)
                    cp "$pnvm" "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
                    ;;
            esac
        fi
    done
    
    # Realtek rtlwifi (older chips - compact)
    if [ -d /lib/firmware/rtlwifi ]; then
        mkdir -p "$INITRAMFS_DIR/lib/firmware/rtlwifi"
        cp -r /lib/firmware/rtlwifi/* "$INITRAMFS_DIR/lib/firmware/rtlwifi/" 2>/dev/null || true
    fi
    
    # Realtek rtw88/rtw89 (newer chips) - only essential files
    for dir in rtw88 rtw89; do
        if [ -d "/lib/firmware/$dir" ]; then
            mkdir -p "$INITRAMFS_DIR/lib/firmware/$dir"
            # Only copy firmware files, not debug data
            cp /lib/firmware/$dir/*.bin* "$INITRAMFS_DIR/lib/firmware/$dir/" 2>/dev/null || true
        fi
    done
    
    # Atheros ath9k_htc (USB adapters - small)
    if [ -d /lib/firmware/ath9k_htc ]; then
        mkdir -p "$INITRAMFS_DIR/lib/firmware/ath9k_htc"
        cp -r /lib/firmware/ath9k_htc/* "$INITRAMFS_DIR/lib/firmware/ath9k_htc/" 2>/dev/null || true
    fi
    
    # Atheros ath10k - only QCA6174 (very common in laptops)
    if [ -d /lib/firmware/ath10k ]; then
        mkdir -p "$INITRAMFS_DIR/lib/firmware/ath10k/QCA6174"
        if [ -d "/lib/firmware/ath10k/QCA6174" ]; then
            cp -r /lib/firmware/ath10k/QCA6174/* "$INITRAMFS_DIR/lib/firmware/ath10k/QCA6174/" 2>/dev/null || true
        fi
        # Also include QCA988X for older adapters
        mkdir -p "$INITRAMFS_DIR/lib/firmware/ath10k/QCA988X"
        if [ -d "/lib/firmware/ath10k/QCA988X" ]; then
            cp -r /lib/firmware/ath10k/QCA988X/* "$INITRAMFS_DIR/lib/firmware/ath10k/QCA988X/" 2>/dev/null || true
        fi
    fi
    
    # MediaTek mt7601u (common cheap USB adapters)
    for fw in /lib/firmware/mt7601u.bin*; do
        if [ -f "$fw" ]; then
            cp "$fw" "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
        fi
    done
    
    # Regulatory database (required for WiFi)
    if [ -f /lib/firmware/regulatory.db ]; then
        cp /lib/firmware/regulatory.db "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
    fi
    if [ -f /lib/firmware/regulatory.db.p7s ]; then
        cp /lib/firmware/regulatory.db.p7s "$INITRAMFS_DIR/lib/firmware/" 2>/dev/null || true
    fi
    
    # Create wpa_supplicant directory
    mkdir -p "$INITRAMFS_DIR/var/run/wpa_supplicant"
    
    # Calculate total firmware size
    TOTAL_FW_SIZE=$(du -sh "$INITRAMFS_DIR/lib/firmware" 2>/dev/null | awk '{print $1}')
    echo "    WiFi firmware size: $TOTAL_FW_SIZE"
fi

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
            wifi.ssid=*)
                WIFI_SSID="${param#*=}"
                ;;
            wifi.password=*)
                WIFI_PASSWORD="${param#*=}"
                ;;
            wifi.psk=*)
                WIFI_PSK="${param#*=}"
                ;;
            wifi.priority=*)
                WIFI_PRIORITY="${param#*=}"
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
# PHASE 1: Network Configuration (Ethernet + WiFi)
# ============================================
log "[Phase 1] Configuring network..."

ip link set lo up

# Find and classify network interfaces
ETH_IFACES=""
WIFI_IFACES=""

for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        log "  Found interface: $iface_name"
        ip link set "$iface_name" up 2>/dev/null
        
        # Detect if interface is wireless
        if [ -d "/sys/class/net/$iface_name/wireless" ] || \
           iwconfig "$iface_name" 2>&1 | grep -q "IEEE 802.11"; then
            WIFI_IFACES="$WIFI_IFACES $iface_name"
            log "    (WiFi interface detected)"
        else
            ETH_IFACES="$ETH_IFACES $iface_name"
        fi
    fi
done

NETWORK_CONNECTED=false

# Function to connect to WiFi
connect_wifi() {
    local iface="$1"
    local ssid="$2"
    local password="$3"
    local psk="$4"
    
    log "  Attempting WiFi connection on $iface..."
    log "    SSID: $ssid"
    
    # Kill any existing wpa_supplicant
    killall wpa_supplicant 2>/dev/null || true
    sleep 1
    
    # Generate PSK if password provided
    if [ -n "$password" ] && [ -z "$psk" ]; then
        psk=$(wpa_passphrase "$ssid" "$password" 2>/dev/null | grep "psk=" | head -1 | sed 's/.*psk=//')
    fi
    
    # Create wpa_supplicant config
    cat > /tmp/wpa_supplicant.conf << WPAEOF
ctrl_interface=/var/run/wpa_supplicant
ctrl_interface_group=0
update_config=1

network={
    ssid="$ssid"
    psk="$psk"
    key_mgmt=WPA-PSK
    proto=WPA2
    pairwise=CCMP TKIP
    group=CCMP TKIP
    scan_ssid=1
}
WPAEOF
    
    # Start wpa_supplicant
    mkdir -p /var/run/wpa_supplicant
    wpa_supplicant -B -i "$iface" -c /tmp/wpa_supplicant.conf -D nl80211,wext 2>/dev/null
    
    if [ $? -ne 0 ]; then
        log_error "Failed to start wpa_supplicant"
        return 1
    fi
    
    # Wait for connection
    log "    Waiting for WiFi association..."
    local wait=0
    while [ $wait -lt 30 ]; do
        if iwconfig "$iface" 2>/dev/null | grep -q "ESSID:\"$ssid\""; then
            log "    WiFi associated successfully!"
            
            # Get DHCP lease
            udhcpc -i "$iface" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null
            sleep 2
            
            if ip route | grep -q "default.*$iface"; then
                log "    WiFi connected: $ssid"
                return 0
            fi
            break
        fi
        sleep 1
        wait=$((wait + 1))
    done
    
    log_error "WiFi connection timeout"
    return 1
}

# Try Ethernet first (preferred)
if [ -n "$ETH_IFACES" ]; then
    log "[Phase 1] Trying Ethernet DHCP..."
    for iface in $ETH_IFACES; do
        udhcpc -i "$iface" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null &
    done
    
    # Wait for Ethernet connection
    log "[Phase 1] Waiting for Ethernet connectivity..."
    WAIT_COUNT=0
    MAX_WAIT=15
    while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
        if ip route | grep -q "default"; then
            NETWORK_CONNECTED=true
            log "[Phase 1] Ethernet connected successfully!"
            break
        fi
        sleep 1
        WAIT_COUNT=$((WAIT_COUNT + 1))
        printf "."
    done
    echo ""
fi

# Try WiFi if Ethernet failed and WiFi credentials provided
if [ "$NETWORK_CONNECTED" = "false" ] && [ -n "$WIFI_IFACES" ]; then
    if [ -n "$WIFI_SSID" ]; then
        log "[Phase 1] Trying WiFi connection..."
        log "  WiFi SSID configured: $WIFI_SSID"
        
        for iface in $WIFI_IFACES; do
            if connect_wifi "$iface" "$WIFI_SSID" "$WIFI_PASSWORD" "$WIFI_PSK"; then
                NETWORK_CONNECTED=true
                break
            fi
        done
    else
        log "[Phase 1] WiFi interfaces detected but no credentials provided"
        log "  To use WiFi, add to kernel cmdline:"
        log "    wifi.ssid=YourNetwork wifi.password=YourPassword"
        
        # Try to scan and list available networks
        if command -v iwlist >/dev/null 2>&1; then
            log "  Available WiFi networks (scan results):"
            for iface in $WIFI_IFACES; do
                iwlist "$iface" scan 2>/dev/null | grep "ESSID:" | head -5 | while read line; do
                    log "    $line"
                done
            done
        fi
    fi
fi

# Final network check
if [ "$NETWORK_CONNECTED" = "false" ]; then
    log_error "No network connection established!"
    log "  Troubleshooting:"
    log "  1. Connect Ethernet cable and reboot"
    log "  2. Add WiFi credentials to boot parameters:"
    log "     wifi.ssid=NetworkName wifi.password=Password"
else
    log "[Phase 1] Network configured successfully!"
fi

# Show network status
echo ""
echo "  === Network Status ==="
ip addr show 2>/dev/null | grep -E "inet " | while read line; do
    echo "  $line"
done
echo "  Gateway: $(ip route | grep default | awk '{print $3}')"
if [ -n "$WIFI_SSID" ]; then
    for iface in $WIFI_IFACES; do
        connected_ssid=$(iwconfig "$iface" 2>/dev/null | grep ESSID | sed 's/.*ESSID:"\(.*\)".*/\1/')
        if [ -n "$connected_ssid" ] && [ "$connected_ssid" != "off/any" ]; then
            echo "  WiFi: $connected_ssid"
        fi
    done
fi
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
        # Show WiFi status if available
        for iface in /sys/class/net/wlan*; do
            if [ -d "$iface" ]; then
                ifname=$(basename "$iface")
                ssid=$(iwconfig "$ifname" 2>/dev/null | grep ESSID | sed 's/.*ESSID:"\(.*\)".*/\1/')
                if [ -n "$ssid" ] && [ "$ssid" != "off/any" ]; then
                    echo "WiFi: Connected to $ssid ($ifname)"
                else
                    echo "WiFi: $ifname (not connected)"
                fi
            fi
        done
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
    wifi-scan)
        echo "Scanning for WiFi networks..."
        for iface in /sys/class/net/wlan*; do
            if [ -d "$iface" ]; then
                ifname=$(basename "$iface")
                echo "=== $ifname ==="
                iwlist "$ifname" scan 2>/dev/null | grep -E "ESSID:|Quality:|Encryption:" | head -20
            fi
        done
        ;;
    wifi-connect)
        if [ -z "$2" ] || [ -z "$3" ]; then
            echo "Usage: deparrow wifi-connect <ssid> <password>"
            exit 1
        fi
        ssid="$2"
        password="$3"
        echo "Connecting to WiFi: $ssid"
        
        # Find WiFi interface
        WIFI_IFACE=""
        for iface in /sys/class/net/wlan*; do
            if [ -d "$iface" ]; then
                WIFI_IFACE=$(basename "$iface")
                break
            fi
        done
        
        if [ -z "$WIFI_IFACE" ]; then
            echo "No WiFi interface found"
            exit 1
        fi
        
        # Kill existing wpa_supplicant
        killall wpa_supplicant 2>/dev/null || true
        sleep 1
        
        # Generate config
        psk=$(wpa_passphrase "$ssid" "$password" 2>/dev/null | grep "psk=" | head -1 | sed 's/.*psk=//')
        cat > /tmp/wpa_supplicant.conf << WPAEOF
ctrl_interface=/var/run/wpa_supplicant
network={
    ssid="$ssid"
    psk="$psk"
    key_mgmt=WPA-PSK
}
WPAEOF
        
        # Start wpa_supplicant
        wpa_supplicant -B -i "$WIFI_IFACE" -c /tmp/wpa_supplicant.conf -D nl80211,wext
        
        # Wait for connection
        sleep 5
        
        # Get DHCP
        udhcpc -i "$WIFI_IFACE" -s /bin/dhcp-script.sh -T 3 -t 5 -n
        
        echo "WiFi connection attempt completed"
        ip addr show "$WIFI_IFACE"
        ;;
    wifi-status)
        echo "=== WiFi Status ==="
        for iface in /sys/class/net/wlan*; do
            if [ -d "$iface" ]; then
                ifname=$(basename "$iface")
                echo "--- $ifname ---"
                iwconfig "$ifname" 2>/dev/null
                echo ""
            fi
        done
        ;;
    help|*)
        echo "DEparrow Commands:"
        echo "  status       - Show node status"
        echo "  logs         - View bacalhau logs"
        echo "  credits      - Check earned credits"
        echo "  register     - Manually register with network"
        echo "  restart      - Restart bacalhau compute node"
        echo "  wifi-scan    - Scan for WiFi networks"
        echo "  wifi-connect - Connect to WiFi (ssid password)"
        echo "  wifi-status  - Show WiFi interface status"
        echo "  help         - Show this help"
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
FIRMWARE_SIZE=$(du -sh lib/firmware 2>/dev/null | awk '{print $1}')
echo "    Components:"
echo "      busybox: $BUSYBOX_SIZE"
echo "      bacalhau: $BACALHAU_SIZE"
if [ -n "$FIRMWARE_SIZE" ] && [ "$FIRMWARE_SIZE" != "0" ]; then
    echo "      wifi firmware: $FIRMWARE_SIZE"
fi

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

# Set up BIOS boot support (i386-pc modules)
# Check if BIOS modules are available (either system-wide or local)
BIOS_MODULES_DIR=""
if [ -d "/usr/lib/grub/i386-pc" ]; then
    BIOS_MODULES_DIR="/usr/lib/grub/i386-pc"
elif [ -d "$PROJECT_ROOT/deparrow/bootable/i386-pc" ]; then
    BIOS_MODULES_DIR="$PROJECT_ROOT/deparrow/bootable/i386-pc"
    echo "[*] Using local i386-pc modules for BIOS boot support"
else
    echo "[!] Warning: No i386-pc BIOS modules found. ISO will be EFI-only."
fi

# Copy BIOS modules to ISO directory for hybrid boot
if [ -n "$BIOS_MODULES_DIR" ]; then
    mkdir -p "$ISO_DIR/boot/grub/i386-pc"
    cp "$BIOS_MODULES_DIR"/*.mod "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES_DIR"/*.img "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES_DIR"/*.lst "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    echo "    Copied $(ls "$ISO_DIR/boot/grub/i386-pc/"*.mod 2>/dev/null | wc -l) BIOS modules"
fi

# Create GRUB config with kernel parameters for auto-join
# IMPORTANT: grub.cfg must be at boot/grub/grub.cfg for grub-mkrescue to embed it
cat > "$ISO_DIR/boot/grub/grub.cfg" << 'GRUB_EOF'
set default=0
set timeout=5

insmod all_video
insmod gfxterm
terminal_output gfxterm

menuentry "DEparrow Compute Node - Auto-Join Network (Ethernet)" {
    linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=bootstrap.deparrow.net:8080
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node - WiFi Mode" {
    echo "Enter WiFi SSID:"
    read wifi_ssid
    echo "Enter WiFi Password:"
    read wifi_password
    linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=bootstrap.deparrow.net:8080 wifi.ssid=$wifi_ssid wifi.password=$wifi_password
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

submenu "Advanced Options >>>" {
    menuentry "WiFi with Custom Bootstrap" {
        echo "Enter WiFi SSID:"
        read wifi_ssid
        echo "Enter WiFi Password:"
        read wifi_password
        echo "Enter Bootstrap Server (host:port):"
        read bootstrap_server
        linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=$bootstrap_server wifi.ssid=$wifi_ssid wifi.password=$wifi_password
        initrd /boot/initrd.img
    }
    
    menuentry "Ethernet with Custom Bootstrap" {
        echo "Enter Bootstrap Server (host:port):"
        read bootstrap_server
        linux /boot/vmlinuz console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=$bootstrap_server
        initrd /boot/initrd.img
    }
    
    menuentry "Test Mode (No Auto-Join)" {
        linux /boot/vmlinuz console=ttyS0,115200n8 debug loglevel=7 deparrow.bootstrap=none
        initrd /boot/initrd.img
    }
}

menuentry "Reboot" {
    reboot
}

menuentry "Shutdown" {
    halt
}
GRUB_EOF

echo "[*] GRUB config created at $ISO_DIR/boot/grub/grub.cfg"

# Build ISO using grub-mkrescue with hybrid boot support
# This creates an ISO that boots on both EFI and BIOS systems
echo "[*] Building ISO with grub-mkrescue..."
mkdir -p "$OUTPUT_DIR"

# Check for EFI modules
EFI_MODULES_DIR=""
if [ -d "/usr/lib/grub/x86_64-efi" ]; then
    EFI_MODULES_DIR="/usr/lib/grub/x86_64-efi"
    echo "    EFI modules found: $EFI_MODULES_DIR"
fi

# Create EFI boot image if modules are available
if [ -n "$EFI_MODULES_DIR" ]; then
    echo "    Creating EFI boot image..."
    mkdir -p "$ISO_DIR/EFI/BOOT"
    
    # Create EFI GRUB core image (use only essential modules that exist)
    if grub-mkimage -O x86_64-efi -o "$ISO_DIR/EFI/BOOT/BOOTX64.EFI" \
        -p "/boot/grub" \
        -d "$EFI_MODULES_DIR" \
        part_gpt part_msdos linux normal echo ls cat help \
        all_video gfxterm font \
        search search_label search_fs_uuid search_fs_file \
        fat ext2 ntfs hfsplus \
        gzio \
        serial; then
        echo "    EFI boot image created: $(ls -lh "$ISO_DIR/EFI/BOOT/BOOTX64.EFI" | awk '{print $5}')"
    else
        echo "    Warning: EFI boot image creation failed"
        rm -f "$ISO_DIR/EFI/BOOT/BOOTX64.EFI"
    fi
fi

# Set GRUB platform for hybrid boot
# grub-mkrescue will use available platforms automatically
if [ -d "$BIOS_MODULES_DIR" ] && [ "$BIOS_MODULES_DIR" != "/usr/lib/grub/i386-pc" ]; then
    # Use local modules - need to set environment for grub-mkrescue
    export GRUB_PREFIX="$ISO_DIR/boot/grub"
    
    # Build BIOS core image first
    if [ -f "$BIOS_MODULES_DIR/cdboot.img" ] && [ -f "$BIOS_MODULES_DIR/boot.img" ]; then
        echo "    Creating hybrid boot image with BIOS support..."
        
        # Build the ISO with xorriso directly for more control over hybrid boot
        # This method creates both El Torito (CD boot) and MBR (USB/disk boot)
        
        # First, create a BIOS bootable core image
        GRUB_CORE="$BUILD_DIR/core.img"
        grub-mkimage -O i386-pc -o "$GRUB_CORE" \
            -p "/boot/grub" \
            -d "$BIOS_MODULES_DIR" \
            biosdisk part_msdos part_gcd iso9660 linux normal echo ls cat help \
            all_video gfxterm font vbe vga video_fb video_cirrus video_bochs \
            search search_label search_fs_uuid search_fs_file \
            fat ext2 ntfs hfsplus \
            gzio \
            serial \
            2>/dev/null || echo "    Note: BIOS core image creation failed"
    fi
fi

# Build the ISO
# grub-mkrescue automatically handles both platforms if modules are available
grub-mkrescue -o "$OUTPUT_DIR/deparrow-autojoin.iso" "$ISO_DIR" 2>&1 | tail -10

ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-autojoin.iso" 2>/dev/null | awk '{print $5}')

# Verify boot capabilities
echo "[*] Verifying ISO boot capabilities..."
xorriso -indev "$OUTPUT_DIR/deparrow-autojoin.iso" 2>&1 | grep -E "Boot record|platform" | head -5

# Check for BIOS boot (MBR boot code)
if xorriso -indev "$OUTPUT_DIR/deparrow-autojoin.iso" 2>&1 | grep -q "i386-pc"; then
    BIOS_SUPPORT="✓"
elif [ -d "$BIOS_MODULES_DIR" ]; then
    # BIOS modules exist but may not be embedded properly
    # Check if MBR boot code is present
    if dd if="$OUTPUT_DIR/deparrow-autojoin.iso" bs=1 count=512 2>/dev/null | strings | grep -q "GRUB"; then
        BIOS_SUPPORT="✓"
    else
        BIOS_SUPPORT="⚠ (modules present, may need manual setup)"
    fi
else
    BIOS_SUPPORT="✗"
fi

# Check for EFI boot
if [ -f "$ISO_DIR/EFI/BOOT/BOOTX64.EFI" ] && [ -s "$ISO_DIR/EFI/BOOT/BOOTX64.EFI" ]; then
    EFI_SUPPORT="✓"
else
    EFI_SUPPORT="✗"
fi

echo ""
echo "========================================="
echo "  DEPARROW AUTO-JOIN ISO BUILD COMPLETE"
echo "========================================="
echo ""
echo "  ISO: $OUTPUT_DIR/deparrow-autojoin.iso"
echo "  Size: $ISO_SIZE"
echo ""
echo "  Boot Support:"
echo "    BIOS (Legacy): $BIOS_SUPPORT"
echo "    EFI (UEFI):    $EFI_SUPPORT"
echo ""
echo "  Network Support:"
echo "    ✓ Ethernet (DHCP auto-config)"
echo "    ✓ WiFi (WPA2-PSK, via kernel cmdline)"
echo ""
echo "  Features:"
echo "    ✓ Auto-discover bootstrap server"
echo "    ✓ Auto-register node identity"
echo "    ✓ Auto-connect to orchestrator"
echo "    ✓ Bacalhau compute node auto-start"
echo "    ✓ Credit earning enabled"
echo ""
echo "  WiFi Configuration (kernel cmdline):"
echo "    wifi.ssid=YourNetworkName"
echo "    wifi.password=YourPassword"
echo "    wifi.psk=PreSharedKey (optional, instead of password)"
echo ""
echo "  Test with QEMU:"
echo ""
echo "    # BIOS boot test (no EFI firmware needed):"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-autojoin.iso -nographic"
echo ""
echo "    # EFI boot test:"
echo "    qemu-system-x86_64 -m 2G \\"
echo "      -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "      -drive if=pflash,format=raw,file=/tmp/ovmf_vars.fd \\"
echo "      -cdrom $OUTPUT_DIR/deparrow-autojoin.iso"
echo ""
echo "    # QEMU with simulated WiFi (needs wireless adapter passthrough):"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-autojoin.iso \\"
echo "      -nographic -append \"wifi.ssid=TestNet wifi.password=TestPass\""
echo ""
echo "  Burn to USB (bootable on both BIOS and EFI systems):"
echo "    sudo dd if=$OUTPUT_DIR/deparrow-autojoin.iso of=/dev/sdX bs=4M status=progress && sync"
echo ""



