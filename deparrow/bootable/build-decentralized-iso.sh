#!/bin/bash
#
# DEparrow Decentralized ISO Builder
# Builds a TRULY DECENTRALIZED bootable ISO
#
# This ISO:
#   - Connects to GCP DPC testnet (34.180.51.11:26657)
#   - Joins global compute mesh (34.180.51.11:4222)
#   - Earns DPC tokens for completed jobs
#   - No centralized authority - pure P2P mesh
#
# Usage: ./build-decentralized-iso.sh [--wifi]
#

set -e

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
VERSION="1.0.0-decentralized"

# Build options
ENABLE_WIFI=${ENABLE_WIFI:-false}


# Build directories
BUILD_DIR="/tmp/deparrow-decentralized-build-$$"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
BOOT_DIR="$ISO_DIR/boot"

# Module directories
BIOS_MODULES="$SCRIPT_DIR/i386-pc"
EFI_MODULES="/usr/lib/grub/x86_64-efi"

# GCP Testnet (Production)
GCP_TESTNET_RPC="34.180.51.11:26657"
GCP_BOOTSTRAP="34.180.51.11:8080"
GCP_ORCHESTRATOR="34.180.51.11:4222"

# ============================================
# Parse Arguments
# ============================================
while [[ $# -gt 0 ]]; do
    case $1 in
        --wifi) ENABLE_WIFI=true; shift ;;
        --help|-h)
            echo "DEparrow Decentralized ISO Builder v${VERSION}"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --wifi    Include WiFi firmware"
            echo "  --help    Show this help"
            echo ""
            echo "Output: $OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso"
            echo ""
            echo "Note: All nodes connect to production GCP network (34.180.51.11)"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# ============================================
# Banner
# ============================================
clear
cat << 'BANNER'

  ██████╗ ███████╗██████╗ ███╗   ███╗ █████╗ ██████╗ ██████╗ 
 ██╔════╝ ██╔════╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗██╔══██╗
 ██║  ███╗█████╗  ██████╔╝██╔████╔██║███████║██████╔╝██████╔╝
 ██║   ██║██╔══╝  ██╔══██╗██║╚██╔╝██║██╔══██║██╔═══╝ ██╔═══╝ 
 ╚██████╔╝███████╗██║  ██║██║ ╚═╝ ██║██║  ██║██║     ██║     
  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     

        🌐 DECENTRALIZED GLOBAL MESH BUILDER 🌐
   "AI Agents Buy Compute to Run Themselves"

BANNER

echo "  Build Type: Decentralized P2P Mesh"
echo "  DPC Testnet: $GCP_TESTNET_RPC"
echo "  Bootstrap: $GCP_BOOTSTRAP"
echo "  Orchestrator: $GCP_ORCHESTRATOR"
echo ""

# ============================================
# Preflight Checks
# ============================================
echo "[Preflight] Checking requirements..."

# Build bacalhau if needed
BACALHAU_BIN="$PROJECT_ROOT/bacalhau"
if [ ! -x "$BACALHAU_BIN" ]; then
    echo "  Building bacalhau..."
    cd "$PROJECT_ROOT"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BACALHAU_BIN" ./main.go
fi
echo "  ✓ bacalhau: $(ls -lh "$BACALHAU_BIN" | awk '{print $5}')"

# Check deparrow wrapper
DEPARROW_WRAPPER="$PROJECT_ROOT/bin/deparrow"
if [ ! -x "$DEPARROW_WRAPPER" ]; then
    echo "  ERROR: deparrow wrapper not found"
    exit 1
fi
echo "  ✓ deparrow wrapper"

# Check Alpine rootfs
ALPINE_ROOTFS="$SCRIPT_DIR/alpine-minirootfs-3.20.0-x86_64.tar.gz"
if [ ! -f "$ALPINE_ROOTFS" ]; then
    echo "  Downloading Alpine minirootfs..."
    wget -q -O "$ALPINE_ROOTFS" \
        "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz"
fi
echo "  ✓ Alpine rootfs: $(ls -lh "$ALPINE_ROOTFS" | awk '{print $5}')"

# Check kernel
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "  ERROR: No kernel found"
    exit 1
fi
echo "  ✓ kernel: $KERNEL_VER"

# Check tools
for tool in grub-mkrescue xorriso; do
    if ! command -v $tool &> /dev/null; then
        echo "  ERROR: $tool not found"
        exit 1
    fi
    echo "  ✓ $tool"
done

# Check modules
[ -d "$BIOS_MODULES" ] && echo "  ✓ BIOS modules: $(ls $BIOS_MODULES/*.mod 2>/dev/null | wc -l)"
[ -d "$EFI_MODULES" ] && echo "  ✓ EFI modules"

echo ""

# ============================================
# Prepare Build Directory
# ============================================
echo "[Setup] Preparing build directory..."

rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log,var/lib/deparrow}
mkdir -p "$INITRAMFS_DIR"/etc/deparrow/keys
mkdir -p "$INITRAMFS_DIR"/lib/firmware
mkdir -p "$BOOT_DIR"/grub
mkdir -p "$ISO_DIR"/{efi/boot,boot/grub/x86_64-efi,boot/grub/i386-pc}

echo "  Build dir: $BUILD_DIR"
echo ""

# ============================================
# Extract Alpine Minirootfs
# ============================================
echo "[Build] Extracting Alpine base..."

tar -xzf "$ALPINE_ROOTFS" -C "$INITRAMFS_DIR"
echo "  ✓ Alpine rootfs extracted"

# Create directories
mkdir -p "$INITRAMFS_DIR/var/lib/deparrow"
mkdir -p "$INITRAMFS_DIR/etc/deparrow/keys"

# Copy binaries
cp "$BACALHAU_BIN" "$INITRAMFS_DIR/bin/bacalhau"
chmod +x "$INITRAMFS_DIR/bin/bacalhau"

# Copy deparrow wrapper
cp "$DEPARROW_WRAPPER" "$INITRAMFS_DIR/bin/deparrow"
chmod +x "$INITRAMFS_DIR/bin/deparrow"

# Patch wrapper to use correct path in ISO
sed -i 's|BACALHAU_BIN=.*|BACALHAU_BIN="/bin/bacalhau.real"|g' "$INITRAMFS_DIR/bin/deparrow"

# Create symlink
mv "$INITRAMFS_DIR/bin/bacalhau" "$INITRAMFS_DIR/bin/bacalhau.real"
ln -sf deparrow "$INITRAMFS_DIR/bin/bacalhau"

echo "  ✓ bacalhau + deparrow wrapper"

# ============================================
# WiFi Support (Optional)
# ============================================
if [ "$ENABLE_WIFI" = "true" ]; then
    echo "[Build] Adding WiFi support..."
    
    # Copy wpa_supplicant
    for tool in /usr/sbin/wpa_supplicant /usr/bin/wpa_passphrase /usr/sbin/iwlist; do
        [ -f "$tool" ] && cp "$tool" "$INITRAMFS_DIR/bin/"
    done
    
    # Copy libraries
    for lib in /lib/x86_64-linux-gnu/lib*.so.* /lib64/ld-linux-*.so.*; do
        [ -f "$lib" ] && cp "$lib" "$INITRAMFS_DIR/lib/x86_64-linux-gnu/" 2>/dev/null || true
    done
    
    # Copy firmware
    for dir in iwlwifi rtlwifi ath9k; do
        [ -d "/lib/firmware/$dir" ] && cp -r "/lib/firmware/$dir" "$INITRAMFS_DIR/lib/firmware/"
    done
    
    echo "  ✓ WiFi support added"
fi

# ============================================
# Copy Decentralized Init Script
# ============================================
echo "[Build] Installing decentralized init script..."

cp "$SCRIPT_DIR/decentralized-init.sh" "$INITRAMFS_DIR/init"
chmod 755 "$INITRAMFS_DIR/init"

echo "  ✓ Decentralized init installed (PRODUCTION mode - connects to GCP network)"

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

# Basic config
touch "$INITRAMFS_DIR/etc/resolv.conf"
echo "127.0.0.1 localhost" > "$INITRAMFS_DIR/etc/hosts"

# ============================================
# Create Helper Commands
# ============================================
cat > "$INITRAMFS_DIR/bin/deparrow" << 'DEPCMD'
#!/bin/sh
# DEparrow CLI wrapper for ISO

case "$1" in
    status)
        . /tmp/deparrow.env 2>/dev/null
        echo "═══════════════════════════════════════"
        echo "  DEPARROW NODE STATUS"
        echo "═══════════════════════════════════════"
        echo "  Node ID:    $NODE_ID"
        echo "  Wallet:     $WALLET_ADDRESS"
        echo "  IP:         $MY_IP"
        echo "  DPC RPC:    $DPC_RPC"
        echo "  Mesh:       $ORCHESTRATOR_HOST:$ORCHESTRATOR_PORT"
        echo "═══════════════════════════════════════"
        ;;
    balance)
        . /tmp/deparrow.env 2>/dev/null
        if [ -n "$WALLET_ADDRESS" ]; then
            echo "Checking DPC balance for $WALLET_ADDRESS..."
            wget -q -O - "${DPC_RPC}/abci_query?path=\"/accounts/${WALLET_ADDRESS}\"" 2>/dev/null || \
                echo "Balance query not yet available"
        fi
        ;;
    logs)
        echo "=== Last 50 lines of compute log ==="
        tail -50 /var/log/bacalhau.log
        ;;
    jobs)
        . /tmp/deparrow.env 2>/dev/null
        echo "Completed jobs (DPC earning history):"
        [ -f /var/lib/deparrow/jobs.log ] && tail -20 /var/lib/deparrow/jobs.log || \
            echo "No jobs completed yet"
        ;;
    connect)
        . /tmp/deparrow.env 2>/dev/null
        echo "Testing connections..."
        echo "  DPC Testnet: $(wget -q -O /dev/null --timeout=3 $DPC_RPC/health && echo '✓' || echo '✗')"
        echo "  Bootstrap:   $(wget -q -O /dev/null --timeout=3 http://${BOOTSTRAP_URL#http*://}/api/v1/health && echo '✓' || echo '✗')"
        echo "  Orchestrator: $(nc -z $ORCHESTRATOR_HOST $ORCHESTRATOR_PORT 2>/dev/null && echo '✓' || echo '✗')"
        ;;
    help|*)
        echo "DEparrow Decentralized Node Commands:"
        echo "  status   - Show node status"
        echo "  balance  - Check DPC balance"
        echo "  logs     - View compute logs"
        echo "  jobs     - List completed jobs"
        echo "  connect  - Test network connectivity"
        ;;
esac
DEPCMD
chmod 755 "$INITRAMFS_DIR/bin/deparrow"

echo "  ✓ Helper commands installed"

# ============================================
# Build Initramfs
# ============================================
echo "[Build] Creating initramfs..."

cd "$INITRAMFS_DIR"
find . | cpio -H newc -o 2>/dev/null | gzip -9 > "$BOOT_DIR/initrd.img"
cd - > /dev/null

echo "  ✓ initrd.img: $(ls -lh $BOOT_DIR/initrd.img | awk '{print $5}')"

# Copy kernel
cp "/boot/vmlinuz-$KERNEL_VER" "$BOOT_DIR/vmlinuz"
echo "  ✓ vmlinuz: $(ls -lh $BOOT_DIR/vmlinuz | awk '{print $5}')"

# ============================================
# Create GRUB Configuration
# ============================================
echo "[Build] Creating GRUB config..."

# Production GCP endpoints (always)
GRUB_BOOTSTRAP="34.180.51.11:8080"
GRUB_ORCHESTRATOR="34.180.51.11:4222"
GRUB_DPC="34.180.51.11:26657"

cat > "$ISO_DIR/boot/grub/grub.cfg" << GRUBEOF
set default=0
set timeout=5
set gfxpayload=keep

insmod all_video
insmod gfxterm
terminal_output gfxterm
insmod font
loadfont unicode

echo ""
echo "***************************************************************"
echo "*        🌐 DEPARROW DECENTRALIZED GLOBAL MESH 🌐              *"
echo "*   \"AI Agents Buy Compute to Run Themselves\"                *"
echo "*                                                             *"
echo "*   DPC Testnet: ${GRUB_DPC}                        *"
echo "*   Mesh:        ${GRUB_ORCHESTRATOR}                            *"
echo "***************************************************************"
echo ""

menuentry "DEparrow Decentralized - Join Global Mesh" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet \
        deparrow.bootstrap=${GRUB_BOOTSTRAP} \
        deparrow.dpc_rpc=${GRUB_DPC}
    initrd /boot/initrd.img
}

menuentry "DEparrow Decentralized - Debug Mode" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 debug loglevel=7 \
        deparrow.bootstrap=${GRUB_BOOTSTRAP} \
        deparrow.dpc_rpc=${GRUB_DPC}
    initrd /boot/initrd.img
}

menuentry "DEparrow Decentralized - Standalone (Offline)" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet \
        deparrow.bootstrap=none \
        deparrow.dpc_rpc=none
    initrd /boot/initrd.img
}

menuentry "Reboot" { reboot }
menuentry "Shutdown" { halt }
GRUBEOF

echo "  ✓ GRUB config created"

# ============================================
# Build EFI Boot
# ============================================
echo "[Build] Creating EFI bootloader..."

if [ -d "$EFI_MODULES" ]; then
    mkdir -p "$ISO_DIR/efi/boot"
    grub-mkimage -O x86_64-efi \
        -o "$ISO_DIR/efi/boot/bootx64.efi" \
        -p "/boot/grub" \
        -d "$EFI_MODULES" \
        part_gpt part_msdos iso9660 linux normal echo ls cat \
        all_video gfxterm font search gzio serial reboot halt \
        2>/dev/null || true
    
    [ -f "$ISO_DIR/efi/boot/bootx64.efi" ] && echo "  ✓ EFI bootloader"
    
    mkdir -p "$ISO_DIR/boot/grub/x86_64-efi"
    cp "$EFI_MODULES"/*.mod "$ISO_DIR/boot/grub/x86_64-efi/" 2>/dev/null || true
fi

# ============================================
# Build BIOS Boot
# ============================================
echo "[Build] Creating BIOS bootloader..."

if [ -d "$BIOS_MODULES" ]; then
    mkdir -p "$ISO_DIR/boot/grub/i386-pc"
    cp "$BIOS_MODULES"/*.mod "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.lst "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.img "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    echo "  ✓ BIOS modules: $(ls $ISO_DIR/boot/grub/i386-pc/*.mod 2>/dev/null | wc -l)"
fi

# ============================================
# Build ISO
# ============================================
echo "[Build] Creating ISO image..."

mkdir -p "$OUTPUT_DIR"
grub-mkrescue -o "$OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso" "$ISO_DIR" 2>&1 | tail -5

if [ ! -f "$OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso" ]; then
    echo "  ERROR: ISO creation failed"
    exit 1
fi

ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso" | awk '{print $5}')

# ============================================
# Cleanup
# ============================================
rm -rf "$BUILD_DIR"

# ============================================
# Summary
# ============================================
echo ""
echo "══════════════════════════════════════════════════════════════════"
echo "        🌐 DECENTRALIZED ISO BUILD COMPLETE! 🌐"
echo "══════════════════════════════════════════════════════════════════"
echo ""
echo "  Output: $OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso"
echo "  Size:   $ISO_SIZE"
echo ""
echo "  Decentralized Features:"
echo "    ✓ Connects to DPC testnet (34.180.51.11:26657)"
echo "    ✓ Joins global compute mesh (34.180.51.11:4222)"
echo "    ✓ Earns DPC tokens for completed jobs"
echo "    ✓ No centralized authority - pure P2P"
echo "    ✓ AI Agent autonomous operation ready"
echo ""
echo "  Test with QEMU:"
echo ""
echo "    # BIOS boot:"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso -nographic"
echo ""
echo "    # EFI boot:"
echo "    qemu-system-x86_64 -m 2G \\"
echo "      -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "      -cdrom $OUTPUT_DIR/deparrow-decentralized-${VERSION}.iso"
echo ""
echo "  Boot the ISO and:"
echo "    - Node will auto-register with global mesh"
echo "    - Wallet will be generated automatically"
echo "    - Start earning DPC for compute contributions"
echo ""
echo "══════════════════════════════════════════════════════════════════"
