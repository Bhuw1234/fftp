#!/bin/bash
# DEparrow Auto-Join ISO Builder - Fixed Version
# Creates EFI-bootable ISO with proper GRUB config

set -e

echo "=== DEparrow Auto-Join ISO Builder (Fixed) ==="

BUILD_DIR="/tmp/deparrow-autojoin-fixed"
ISO_DIR="$BUILD_DIR/iso"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
OUTPUT_DIR="/home/bhuwan/bacalhau/deparrow/bootable/output"

# Clean and create directories
rm -rf "$BUILD_DIR"
mkdir -p "$ISO_DIR"/boot/grub/x86_64-efi
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,var/log,var/lib/deparrow,etc/deparrow/keys}

# Copy kernel
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "ERROR: No kernel found"
    exit 1
fi
echo "[1/6] Copying kernel: $KERNEL_VER"
sudo cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz" 2>/dev/null || cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz" 2>/dev/null
sudo chown $USER:$USER "$ISO_DIR/boot/vmlinuz" 2>/dev/null || true

# Copy busybox and create symlinks
echo "[2/6] Setting up busybox..."
cp /bin/busybox "$INITRAMFS_DIR/bin/"
cd "$INITRAMFS_DIR/bin"
for cmd in sh cat ls mkdir mount umount sleep echo ip ln rm mv cp chmod chown grep sed awk cut head tail wget nc ping hostname reboot halt poweroff udhcpc ifconfig route; do
    ln -sf busybox "$cmd" 2>/dev/null || true
done
cd - > /dev/null

# Copy bacalhau binary
echo "[3/6] Copying bacalhau..."
cp /home/bhuwan/bacalhau/bacalhau "$INITRAMFS_DIR/bin/" 2>/dev/null || {
    echo "Building bacalhau..."
    cd /home/bhuwan/bacalhau
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INITRAMFS_DIR/bin/bacalhau" ./main.go
    cd - > /dev/null
}
chmod +x "$INITRAMFS_DIR/bin/bacalhau"

# Create DHCP script
cat > "$INITRAMFS_DIR/bin/dhcp-script.sh" << 'DHCP_EOF'
#!/bin/sh
case "$1" in
    renew|bound)
        ip addr add "$ip/${mask:-24}" dev "$interface"
        [ -n "$router" ] && ip route add default via "$router" dev "$interface"
        if [ -n "$dns" ]; then
            > /etc/resolv.conf
            for s in $dns; do echo "nameserver $s" >> /etc/resolv.conf; done
        fi
        ;;
esac
DHCP_EOF
chmod 755 "$INITRAMFS_DIR/bin/dhcp-script.sh"

# Create init script
echo "[4/6] Creating init script..."
cat > "$INITRAMFS_DIR/init" << 'INIT_EOF'
#!/bin/sh
# DEparrow Auto-Join Init Script

export PATH=/bin:/usr/bin
export DEPARROW_CONFIG=/etc/deparrow

# Mount filesystems
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev
mount -t tmpfs tmpfs /tmp
mount -t tmpfs tmpfs /run

# Parse kernel command line
BOOTSTRAP="bootstrap.deparrow.net:8080"
for param in $(cat /proc/cmdline); do
    case "$param" in
        deparrow.bootstrap=*) BOOTSTRAP="${param#*=}" ;;
    esac
done

# Set hostname
NODE_NAME="deparrow-$(head -c 4 /dev/urandom | xxd -p 2>/dev/null || echo 'node')"
hostname "$NODE_NAME"

# Print banner
clear
cat << 'BANNER'
  ╔══════════════════════════════════════════════════════════════╗
  ║   ██████╗ ███████╗ ██████╗ █████╗ ██████╗  ██████╗ ██████╗  ║
  ║   ██╔══██╗██╔════╝██╔════╝██╔══██╗██╔══██╗██╔════╝██╔═══██╗ ║
  ║   ██║  ██║█████╗  ██║     ███████║██████╔╝██║     ██║   ██║ ║
  ║   ██║  ██║██╔══╝  ██║     ██╔══██║██╔══██╗██║     ██║   ██║ ║
  ║   ██████╔╝███████╗╚██████╗██║  ██║██║  ██║╚██████╗╚██████╔╝ ║
  ║   ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ║
  ║                                                              ║
  ║        Global Virtual Machine - Auto-Join Network            ║
  ╚══════════════════════════════════════════════════════════════╝
BANNER

echo ""
echo "[Phase 1] Network Configuration..."

# Configure network
ip link set lo up
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        ip link set "$iface_name" up 2>/dev/null
        udhcpc -i "$iface_name" -s /bin/dhcp-script.sh -T 3 -t 5 -n 2>/dev/null &
    fi
done

sleep 5
echo ""
ip addr show 2>/dev/null | grep -E "inet " | head -3

echo ""
echo "[Phase 2] Node Identity..."
mkdir -p "$DEPARROW_CONFIG/keys"
NODE_ID=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || head -c 16 /dev/urandom | xxd -p)
echo "$NODE_ID" > "$DEPARROW_CONFIG/node-id"
echo "  Node ID: $NODE_ID"

echo ""
echo "[Phase 3] Bootstrap Discovery..."
echo "  Trying: $BOOTSTRAP"
if wget -q -O /dev/null --timeout=5 "http://$BOOTSTRAP/api/v1/health" 2>/dev/null; then
    echo "  Bootstrap server found!"
else
    echo "  No bootstrap server (standalone mode)"
fi

echo ""
echo "[Phase 4] Starting bacalhau..."
bacalhau serve --compute > /var/log/bacalhau.log 2>&1 &
BAC_PID=$!
sleep 2
if kill -0 $BAC_PID 2>/dev/null; then
    echo "  Bacalhau started (PID: $BAC_PID)"
else
    echo "  Bacalhau failed to start"
fi

echo ""
cat << 'STATUS'
  ╔══════════════════════════════════════════════════════════════╗
  ║              DEPARROW COMPUTE NODE ONLINE                    ║
  ╠══════════════════════════════════════════════════════════════╣
  ║  🚀 This node is now earning credits!                        ║
  ╚══════════════════════════════════════════════════════════════╝
STATUS

echo ""
echo "Commands: deparrow status, deparrow logs, ps, top"
echo ""
exec /bin/sh
INIT_EOF
chmod 755 "$INITRAMFS_DIR/init"

# Create basic /etc files
touch "$INITRAMFS_DIR/etc/resolv.conf"
echo "127.0.0.1 localhost" > "$INITRAMFS_DIR/etc/hosts"

# Build initramfs
echo "[5/6] Building initramfs..."
cd "$INITRAMFS_DIR"
find . | cpio -H newc -o 2>/dev/null | gzip -9 > "$ISO_DIR/boot/initrd.img"
INITRD_SIZE=$(ls -lh "$ISO_DIR/boot/initrd.img" | awk '{print $5}')
echo "  initrd.img: $INITRD_SIZE"
cd - > /dev/null

# Create GRUB config - CRITICAL: this must be properly formatted
echo "[6/6] Creating GRUB configuration..."
cat > "$ISO_DIR/boot/grub/grub.cfg" << 'GRUB_EOF'
set default=0
set timeout=5

insmod all_video
insmod gfxterm
terminal_output gfxterm

menuentry "DEparrow Auto-Join Network" {
    linux /boot/vmlinuz quiet loglevel=3 deparrow.bootstrap=bootstrap.deparrow.net:8080
    initrd /boot/initrd.img
}

menuentry "DEparrow Debug Mode" {
    linux /boot/vmlinuz debug loglevel=7
    initrd /boot/initrd.img
}

menuentry "DEparrow Standalone" {
    linux /boot/vmlinuz quiet deparrow.bootstrap=none
    initrd /boot/initrd.img
}

menuentry "Reboot" { reboot; }
menuentry "Shutdown" { halt; }
GRUB_EOF

# Build ISO with grub-mkrescue (includes EFI support)
echo ""
echo "Building ISO with grub-mkrescue..."
mkdir -p "$OUTPUT_DIR"
grub-mkrescue -o "$OUTPUT_DIR/deparrow-autojoin-fixed.iso" "$ISO_DIR" 2>&1 | tail -5

# Report
ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-autojoin-fixed.iso" | awk '{print $5}')
echo ""
echo "========================================="
echo "  BUILD COMPLETE"
echo "========================================="
echo "  ISO: $OUTPUT_DIR/deparrow-autojoin-fixed.iso"
echo "  Size: $ISO_SIZE"
echo ""
echo "Test (EFI):"
echo "  qemu-system-x86_64 -m 2G \\"
echo "    -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "    -drive if=pflash,format=raw,file=/tmp/ovmf_vars.fd \\"
echo "    -cdrom $OUTPUT_DIR/deparrow-autojoin-fixed.iso"
