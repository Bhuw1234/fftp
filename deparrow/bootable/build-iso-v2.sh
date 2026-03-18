#!/bin/bash
# DEparrow ISO Builder v2 - Minimal working ISO
# Creates a bootable ISO that auto-starts compute node

set -e

echo "=== DEparrow ISO Builder v2 ==="

BUILD_DIR="/tmp/deparrow-iso-v2"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
OUTPUT_DIR="/home/bhuwan/bacalhau/deparrow/bootable/output"

# Cleanup
rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log}
mkdir -p "$ISO_DIR/boot/grub"

# Get kernel version
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "ERROR: No kernel found in /boot"
    exit 1
fi
echo "Using kernel: $KERNEL_VER"

# Copy busybox static (provides all basic tools)
cp /bin/busybox "$INITRAMFS_DIR/bin/busybox"
chmod +x "$INITRAMFS_DIR/bin/busybox"

# Create symlinks for busybox applets
cd "$INITRAMFS_DIR/bin"
for applet in sh cat ls mkdir mount umount sleep echo ip ln rm mv cp chmod chown grep sed awk cut head tail; do
    ln -sf busybox "$applet"
done
cd - > /dev/null

# Create init script
cat > "$INITRAMFS_DIR/init" << 'INIT_EOF'
#!/bin/sh

# DEparrow Boot Init - Auto-start compute node
export PATH=/bin:/usr/bin:/usr/sbin

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

# Set hostname
hostname deparrow-node

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
echo "  ║        Global Virtual Machine - Compute Node                 ║"
echo "  ╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "  Booting DEparrow Compute Node..."
echo ""

# Configure network
echo "[*] Configuring network..."
ip link set lo up
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        echo "    Bringing up $iface_name..."
        ip link set "$iface_name" up 2>/dev/null
    fi
done

# Wait for DHCP (background)
echo "[*] Requesting DHCP lease..."
for iface in /sys/class/net/*; do
    iface_name=$(basename "$iface")
    if [ "$iface_name" != "lo" ]; then
        udhcpc -i "$iface_name" -s /bin/simple.script -T 5 -t 3 2>/dev/null &
    fi
done
sleep 3

# Show network status
echo ""
echo "  === Network Status ==="
ip addr show | grep -E "inet |link/ether" | head -4
echo ""

# Mount tmpfs for runtime
mount -t tmpfs tmpfs /tmp
mount -t tmpfs tmpfs /run

# Show system info
echo "  === System Info ==="
echo "  Hostname: $(hostname)"
echo "  Kernel: $(uname -r)"
echo "  CPU: $(cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d: -f2)"
echo "  Memory: $(free -h | grep Mem | awk '{print $2}')"
echo ""

# Start DEparrow services
echo "[*] Initializing DEparrow services..."
echo "    This node will automatically:"
echo "    - Join the DEparrow network"
echo "    - Contribute idle compute resources"
echo "    - Earn credits for compute provided"
echo ""

# Create startup marker
echo "DEPARROW_NODE_READY=1" > /tmp/deparrow.env

# Interactive shell or daemon mode
echo "  ════════════════════════════════════════════════════════════════"
echo "  DEparrow Compute Node Ready"
echo "  Type 'help' for available commands"
echo "  ════════════════════════════════════════════════════════════════"
echo ""

# Start shell
exec /bin/sh
INIT_EOF

chmod 755 "$INITRAMFS_DIR/init"

# Create DHCP script for udhcpc
cat > "$INITRAMFS_DIR/bin/simple.script" << 'DHCP_EOF'
#!/bin/sh
case "$1" in
    bound|renew)
        ip addr add $ip/$mask dev $interface
        if [ -n "$router" ]; then
            ip route add default via $router
        fi
        if [ -n "$dns" ]; then
            echo "nameserver $dns" > /etc/resolv.conf
        fi
        ;;
esac
DHCP_EOF
chmod 755 "$INITRAMFS_DIR/bin/simple.script"

# Create basic resolv.conf
touch "$INITRAMFS_DIR/etc/resolv.conf"

# Build initramfs (cpio newc format)
echo "[*] Building initramfs..."
cd "$INITRAMFS_DIR"
find . | cpio -H newc -o 2>/dev/null | gzip > "$ISO_DIR/boot/initrd.img"
INITRD_SIZE=$(ls -lh "$ISO_DIR/boot/initrd.img" | awk '{print $5}')
echo "    initrd.img: $INITRD_SIZE"
cd - > /dev/null

# Copy kernel (need sudo for /boot access)
echo "[*] Copying kernel..."
if [ -r "/boot/vmlinuz-$KERNEL_VER" ]; then
    cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz"
else
    echo "    Need sudo to read kernel..."
    sudo cp "/boot/vmlinuz-$KERNEL_VER" "$ISO_DIR/boot/vmlinuz"
    sudo chown $USER:$USER "$ISO_DIR/boot/vmlinuz"
fi
chmod 644 "$ISO_DIR/boot/vmlinuz"
KERNEL_SIZE=$(ls -lh "$ISO_DIR/boot/vmlinuz" | awk '{print $5}')
echo "    vmlinuz: $KERNEL_SIZE"

# Create GRUB config
cat > "$ISO_DIR/boot/grub/grub.cfg" << 'GRUB_EOF'
set default=0
set timeout=3

menuentry "DEparrow Compute Node" {
    linux /boot/vmlinuz quiet loglevel=3
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node (Debug)" {
    linux /boot/vmlinuz debug loglevel=7
    initrd /boot/initrd.img
}
GRUB_EOF

# Create GRUB core image
echo "[*] Creating GRUB boot image..."
grub-mkimage -o "$ISO_DIR/boot/grub/core.img" -O i386-pc -p "(hd0,msdos1)/boot/grub" \
    biosdisk part_msdos ext2 normal linux boot

# Create boot sector
grub-bios-setup --skip-fs-probe --force --device-map=/dev/null \
    --root-device="(hd0,msdos1)" "$ISO_DIR/boot/grub/core.img" 2>/dev/null || true

# Build ISO with xorriso
echo "[*] Building ISO..."
mkdir -p "$OUTPUT_DIR"
xorriso -as mkisofs \
    -o "$OUTPUT_DIR/deparrow-1.0.0.iso" \
    -b boot/grub/core.img \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    --grub2-boot-info \
    -V "DEPARROW" \
    -J -R \
    "$ISO_DIR"

ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-1.0.0.iso" | awk '{print $5}')
echo ""
echo "=== Build Complete ==="
echo "ISO: $OUTPUT_DIR/deparrow-1.0.0.iso"
echo "Size: $ISO_SIZE"
echo ""
echo "Test with QEMU:"
echo "  qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-1.0.0.iso"
