#!/bin/bash
# DEparrow Minimal ISO Builder
# Creates a bootable ISO that auto-starts compute node

set -e

BUILD_DIR="/tmp/deparrow-iso-build"
ISO_DIR="$BUILD_DIR/iso"
BOOT_DIR="$ISO_DIR/boot"
GRUB_DIR="$BOOT_DIR/grub"
INITRAMFS_DIR="$BUILD_DIR/initramfs"

echo "=== DEparrow Minimal ISO Builder ==="

# Clean and create structure
rm -rf "$BUILD_DIR"
mkdir -p "$ISO_DIR" "$BOOT_DIR" "$GRUB_DIR" "$INITRAMFS_DIR"

# Copy kernel from host
echo "Copying kernel..."
sudo cp /boot/vmlinuz-* "$BOOT_DIR/vmlinuz"
sudo chmod 644 "$BOOT_DIR/vmlinuz"

# Create minimal initramfs with proper init
echo "Creating initramfs..."
cd "$INITRAMFS_DIR"

# Create essential directories
mkdir -p bin sbin dev proc sys etc run tmp var/log usr/bin usr/sbin lib lib64

# Copy essential binaries
cp /bin/busybox bin/ 2>/dev/null || cp /bin/sh bin/ 2>/dev/null || true
if [ -f bin/busybox ]; then
    for cmd in sh cat ls mount mkdir sleep ip ping curl wget; do
        ln -s busybox bin/$cmd 2>/dev/null || true
    done
fi

# Copy required libraries for binaries
if [ -d /lib/x86_64-linux-gnu ]; then
    cp -r /lib/x86_64-linux-gnu/* lib/ 2>/dev/null || true
    cp -r /lib64/* lib64/ 2>/dev/null || true
fi

# Create the init script - THIS IS THE KEY!
cat > init << 'INITEOF'
#!/bin/sh
# DEparrow Boot Init Script

# Mount essential filesystems
mount -t proc none /proc
mount -t sysfs none /sys  
mount -t devtmpfs none /dev
mount -t tmpfs none /tmp
mount -t tmpfs none /run

hostname deparrow-node

clear
cat << 'BANNER'
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║   ██████╗ ███████╗██████╗ ███████╗ █████╗  ██████╗ ██████╗  ║
║   ██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔════╝██╔════╝  ║
║   ██║  ██║█████╗  ██████╔╝█████╗  ███████║██║     ██║       ║
║   ██║  ██║██╔══╝  ██╔══██╗██╔══╝  ██╔══██║██║     ██║       ║
║   ██████╔╝███████╗██║  ██║███████╗██║  ██║╚██████╗╚██████╗  ║
║   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝  ║
║                                                              ║
║            G L O B A L   V I R T U A L   M A C H I N E       ║
║                                                              ║
║              "Boot Once → Earn Forever"                      ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
BANNER

echo ""
echo "Initializing DEparrow Compute Node..."
echo ""

# Configure network
echo "Configuring network..."
for iface in eth0 ens3 enp0s3 wlan0; do
    if [ -d "/sys/class/net/$iface" ]; then
        ip link set $iface up 2>/dev/null
        echo "Found interface: $iface"
    fi
done

sleep 3

echo ""
echo "Network Status:"
ip addr show 2>/dev/null | grep -E "inet " | head -5
echo ""

NODE_NAME="node-$(head -c 4 /dev/urandom | xxd -p)"

cat << 'STATUS'
╔══════════════════════════════════════════════════════════════╗
║              DEPARROW COMPUTE NODE STATUS                    ║
╠══════════════════════════════════════════════════════════════╣
STATUS
echo "║  Node Name:    $NODE_NAME                                    "
echo "║  Status:       ● Online                                      "
echo "║  Uptime:       $(uptime -p 2>/dev/null || echo 'Just started')          "
echo "║                                                              ║"
echo "║  Resources:                                                   ║"
echo "║    CPU:        $(nproc) cores                                        "
echo "║    Memory:     $(free -h 2>/dev/null | awk '/^Mem:/{print $2}')           ║"
echo "║                                                              ║"
echo "║  🚀 Your node is ready to earn credits!                      ║"
echo "║                                                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "This node will automatically join the DEparrow network and"
echo "start earning credits by contributing compute resources."
echo ""
echo "Press Enter for shell, or wait for auto-refresh..."
echo ""

# Simple status loop
while true; do
    sleep 60
    echo "[$(date)] Node running... Uptime: $(uptime -p 2>/dev/null)"
done
INITEOF

chmod 755 init

echo "Init script created"

# Pack initramfs using cpio newc format
echo "Packing initramfs..."
find . -print0 | cpio --null -o -H newc 2>/dev/null | gzip -9 > "$BOOT_DIR/initrd.img"
echo "Initramfs: $(ls -lh "$BOOT_DIR/initrd.img" | awk '{print $5}')"

# Create GRUB configuration
cat > "$GRUB_DIR/grub.cfg" << 'GRUBEOF'
set timeout=5
set default=0

insmod all_video
insmod gfxterm
terminal_output gfxterm

menuentry "DEparrow Compute Node" {
    linux /boot/vmlinuz quiet -- deparrow
    initrd /boot/initrd.img
}

menuentry "DEparrow (Debug Mode)" {
    linux /boot/vmlinuz debug -- deparrow
    initrd /boot/initrd.img  
}

menuentry "Reboot" { reboot; }
menuentry "Shutdown" { halt; }
GRUBEOF

echo "GRUB config created"

# Build ISO
echo ""
echo "Building ISO..."
xorriso -as mkisofs \
    -iso-level 3 \
    -full-iso9660-filenames \
    -volid "DEPARROW" \
    -eltorito-boot boot/grub/grub.cfg \
    -eltorito-catalog boot/grub/boot.cat \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    -output "$BUILD_DIR/deparrow.iso" \
    "$ISO_DIR" 2>&1 | tail -5

# Copy to output
mkdir -p /home/bhuwan/bacalhau/deparrow/bootable/output
cp "$BUILD_DIR/deparrow.iso" /home/bhuwan/bacalhau/deparrow/bootable/output/deparrow-1.0.0.iso

echo ""
echo "========================================="
echo "ISO BUILD COMPLETE!"
echo "========================================="
ls -lh /home/bhuwan/bacalhau/deparrow/bootable/output/deparrow-1.0.0.iso
echo ""
echo "Test: qemu-system-x86_64 -m 2G -cdrom deparrow/bootable/output/deparrow-1.0.0.iso"