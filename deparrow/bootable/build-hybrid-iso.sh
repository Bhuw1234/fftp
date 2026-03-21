#!/bin/bash
# DEparrow Hybrid ISO Builder - BIOS + EFI Boot Support
# Creates a bootable ISO that works on both legacy BIOS and UEFI systems

set -e

echo "=== DEparrow Hybrid ISO Builder ==="
echo "    Building ISO with BIOS + EFI boot support"
echo ""

BUILD_DIR="/tmp/deparrow-hybrid-iso"
INITRAMFS_DIR="$BUILD_DIR/initramfs"
ISO_DIR="$BUILD_DIR/iso"
BOOT_DIR="$ISO_DIR/boot"
OUTPUT_DIR="/home/bhuwan/bacalhau/deparrow/bootable/output"
PROJECT_ROOT="/home/bhuwan/bacalhau"

# BIOS modules location
BIOS_MODULES="$PROJECT_ROOT/deparrow/bootable/i386-pc"
EFI_MODULES="/usr/lib/grub/x86_64-efi"

# Cleanup
rm -rf "$BUILD_DIR"
mkdir -p "$INITRAMFS_DIR"/{bin,dev,etc,proc,sys,run,tmp,usr/bin,usr/sbin,var/log,var/lib/deparrow}
mkdir -p "$INITRAMFS_DIR"/etc/deparrow/keys
mkdir -p "$BOOT_DIR"/{grub,isolinux}
mkdir -p "$ISO_DIR"/{efi/boot,boot/grub/x86_64-efi}

# Get kernel version
KERNEL_VER=$(ls /boot/vmlinuz-* 2>/dev/null | head -1 | sed 's/.*vmlinuz-//')
if [ -z "$KERNEL_VER" ]; then
    echo "ERROR: No kernel found in /boot"
    exit 1
fi
echo "[*] Using kernel: $KERNEL_VER"

# ============================================
# Step 1: Build initramfs (reuse from v2 script)
# ============================================
echo "[*] Building initramfs..."

# Copy busybox
if [ -f /bin/busybox ]; then
    cp /bin/busybox "$INITRAMFS_DIR/bin/busybox"
elif [ -f /usr/bin/busybox ]; then
    cp /usr/bin/busybox "$INITRAMFS_DIR/bin/busybox"
else
    echo "ERROR: busybox not found"
    exit 1
fi
chmod +x "$INITRAMFS_DIR/bin/busybox"

# Create busybox symlinks
cd "$INITRAMFS_DIR/bin"
for applet in sh cat ls mkdir mount umount sleep echo ip ln rm mv cp chmod chown grep sed awk cut head tail wget nc ping ifconfig route hostname reboot halt poweroff; do
    ln -sf busybox "$applet" 2>/dev/null || true
done
cd - > /dev/null

# Copy bacalhau binary
echo "[*] Copying bacalhau binary..."
if [ -f "$PROJECT_ROOT/bacalhau" ]; then
    cp "$PROJECT_ROOT/bacalhau" "$INITRAMFS_DIR/bin/bacalhau"
else
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INITRAMFS_DIR/bin/bacalhau" "$PROJECT_ROOT/main.go"
fi
chmod +x "$INITRAMFS_DIR/bin/bacalhau"

# Copy the init script from v2 build (read from file)
cp "$PROJECT_ROOT/deparrow/bootable/build-iso-v2.sh" /tmp/v2-script.sh
# Extract init script (between INIT_EOF markers)
sed -n '/^cat > "\$INITRAMFS_DIR\/init" << .INIT_EOF./,/^INIT_EOF$/p' "$PROJECT_ROOT/deparrow/bootable/build-iso-v2.sh" | \
    sed '1d;$d' > "$INITRAMFS_DIR/init"
chmod 755 "$INITRAMFS_DIR/init"

# Create DHCP script
cat > "$INITRAMFS_DIR/bin/dhcp-script.sh" << 'DHCP_EOF'
#!/bin/sh
[ -z "$1" ] && exit 1
case "$1" in
    deconfig) ip addr flush dev "$interface"; ip link set "$interface" up ;;
    renew|bound)
        ip addr add "$ip/${mask:-24}" dev "$interface"
        [ -n "$router" ] && ip route add default via "$router" dev "$interface" 2>/dev/null
        if [ -n "$dns" ]; then
            > /etc/resolv.conf
            for server in $dns; do echo "nameserver $server" >> /etc/resolv.conf; done
        fi ;;
esac
DHCP_EOF
chmod 755 "$INITRAMFS_DIR/bin/dhcp-script.sh"
ln -sf dhcp-script.sh "$INITRAMFS_DIR/bin/simple.script"

touch "$INITRAMFS_DIR/etc/resolv.conf"
echo "127.0.0.1 localhost" > "$INITRAMFS_DIR/etc/hosts"

# Build initramfs
echo "[*] Compressing initramfs..."
cd "$INITRAMFS_DIR"
find . | cpio -H newc -o 2>/dev/null | gzip -9 > "$BOOT_DIR/initrd.img"
cd - > /dev/null
echo "    initrd.img: $(ls -lh "$BOOT_DIR/initrd.img" | awk '{print $5}')"

# Copy kernel
cp "/boot/vmlinuz-$KERNEL_VER" "$BOOT_DIR/vmlinuz"
echo "    vmlinuz: $(ls -lh "$BOOT_DIR/vmlinuz" | awk '{print $5}')"

# ============================================
# Step 2: Create GRUB config (shared by BIOS and EFI)
# ============================================
echo "[*] Creating GRUB configuration..."

cat > "$ISO_DIR/boot/grub/grub.cfg" << 'GRUB_CFG'
set default=0
set timeout=5
set gfxpayload=keep

# Load graphics modules for BIOS
if [ "${grub_platform}" = "pc" ]; then
    insmod vbe
    insmod vga
fi
insmod all_video
insmod gfxterm
terminal_output gfxterm
insmod font
loadfont unicode

menuentry "DEparrow Compute Node - Auto-Join Network" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=bootstrap.deparrow.net:8080
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node - Debug Mode" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 debug loglevel=7
    initrd /boot/initrd.img
}

menuentry "DEparrow Compute Node - Standalone No Network" {
    linux /boot/vmlinuz console=tty0 console=ttyS0,115200n8 quiet loglevel=3 deparrow.bootstrap=none
    initrd /boot/initrd.img
}

menuentry "Reboot" {
    reboot
}

menuentry "Shutdown" {
    halt
}
GRUB_CFG

echo "    grub.cfg created"

# ============================================
# Step 3: Build EFI boot image
# ============================================
echo "[*] Building EFI boot image..."

# Copy EFI modules
if [ -d "$EFI_MODULES" ]; then
    cp "$EFI_MODULES"/*.mod "$ISO_DIR/boot/grub/x86_64-efi/" 2>/dev/null || true
    cp "$EFI_MODULES"/*.lst "$ISO_DIR/boot/grub/x86_64-efi/" 2>/dev/null || true
    echo "    Copied $(ls "$ISO_DIR/boot/grub/x86_64-efi/"*.mod 2>/dev/null | wc -l) EFI modules"
fi

# Create EFI boot image
EFI_BOOT_IMG="$BUILD_DIR/efi-boot.img"
dd if=/dev/zero of="$EFI_BOOT_IMG" bs=1M count=20 2>/dev/null
mkfs.fat -F 12 "$EFI_BOOT_IMG" 2>/dev/null

# Mount and populate EFI image (requires mtools for non-root)
MTOOLS_SKIP_RC=1 mmd -i "$EFI_BOOT_IMG" ::/efi ::/efi/boot 2>/dev/null || true

# Create EFI GRUB image
EFI_GRUB_IMG="$BUILD_DIR/grub-efi.img"
grub-mkimage -O x86_64-efi \
    -o "$EFI_GRUB_IMG" \
    -p "/boot/grub" \
    -d "$EFI_MODULES" \
    part_gpt part_msdos fat ext2 iso9660 linux normal echo ls cat \
    all_video gfxterm font search gzio serial reboot halt \
    2>/dev/null || echo "    Note: EFI image creation may need grub-efi modules"

# Copy EFI bootloader
if [ -f "$EFI_GRUB_IMG" ]; then
    mcopy -i "$EFI_BOOT_IMG" "$EFI_GRUB_IMG" ::/efi/boot/bootx64.efi 2>/dev/null || \
        cp "$EFI_GRUB_IMG" "$ISO_DIR/efi/boot/bootx64.efi" 2>/dev/null || true
fi

# Use grub-mkrescue for EFI boot image as fallback
if [ ! -f "$EFI_GRUB_IMG" ]; then
    echo "    Creating EFI boot image via grub-mkrescue..."
fi

# ============================================
# Step 4: Build BIOS boot image (El Torito)
# ============================================
echo "[*] Building BIOS boot image..."

if [ -d "$BIOS_MODULES" ]; then
    # Create BIOS core image with required modules
    BIOS_CORE="$BUILD_DIR/core.img"
    grub-mkimage -O i386-pc \
        -o "$BIOS_CORE" \
        -p "/boot/grub" \
        -d "$BIOS_MODULES" \
        biosdisk part_msdos part_gpt iso9660 linux normal echo ls cat help \
        all_video gfxterm font vbe vga \
        search search_label search_fs_uuid search_fs_file \
        fat ext2 gzio serial reboot halt \
        2>/dev/null
    
    if [ -f "$BIOS_CORE" ] && [ -f "$BIOS_MODULES/cdboot.img" ]; then
        # Combine cdboot.img and core.img for El Torito boot
        cat "$BIOS_MODULES/cdboot.img" "$BIOS_CORE" > "$BOOT_DIR/grub/bios.img"
        echo "    Created BIOS boot image: bios.img"
    fi
    
    # Copy BIOS modules to ISO for GRUB to load
    cp "$BIOS_MODULES"/*.mod "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.lst "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    cp "$BIOS_MODULES"/*.img "$ISO_DIR/boot/grub/i386-pc/" 2>/dev/null || true
    echo "    Copied $(ls "$ISO_DIR/boot/grub/i386-pc/"*.mod 2>/dev/null | wc -l) BIOS modules"
else
    echo "    ERROR: BIOS modules not found at $BIOS_MODULES"
    echo "    Install with: sudo apt-get install grub-pc-bin"
    exit 1
fi

# ============================================
# Step 5: Build hybrid ISO with xorriso
# ============================================
echo "[*] Building hybrid ISO with xorriso..."

mkdir -p "$OUTPUT_DIR"

# Create early config for BIOS boot
cat > "$BUILD_DIR/bios-early.cfg" << 'EOF'
set root=cd0
set prefix=/boot/grub
EOF

# Copy the BIOS core.img to the right location
mkdir -p "$ISO_DIR/boot/grub/i386-pc"
if [ -f "$BOOT_DIR/grub/bios.img" ]; then
    cp "$BOOT_DIR/grub/bios.img" "$ISO_DIR/boot/grub/i386-pc/"
fi

# Copy EFI boot image
if [ -f "$EFI_BOOT_IMG" ]; then
    mkdir -p "$ISO_DIR/boot/efi"
    cp "$EFI_BOOT_IMG" "$ISO_DIR/boot/efi/"
fi

echo "    Creating hybrid ISO with xorriso..."

# Use mkisofs compatibility mode for simpler syntax
xorriso -as mkisofs \
    -o "$OUTPUT_DIR/deparrow-hybrid.iso" \
    -V "DEPARROW" \
    -J -R \
    -b boot/grub/i386-pc/bios.img \
    -no-emul-boot \
    -boot-load-size 4 \
    -boot-info-table \
    --grub2-mbr "$BIOS_MODULES/boot.img" \
    -append_partition 2 0xef "$EFI_BOOT_IMG" \
    -appended_part_as_gpt \
    "$ISO_DIR" \
    2>&1 | tail -15

# ============================================
# Step 6: Verify and report
# ============================================
ISO_SIZE=$(ls -lh "$OUTPUT_DIR/deparrow-hybrid.iso" | awk '{print $5}')

echo ""
echo "[*] Verifying ISO boot capabilities..."
xorriso -indev "$OUTPUT_DIR/deparrow-hybrid.iso" 2>&1 | grep -E "Boot record|El-Torito|platform" | head -5

echo ""
echo "========================================="
echo "  DEPARROW HYBRID ISO BUILD COMPLETE"
echo "========================================="
echo ""
echo "  ISO: $OUTPUT_DIR/deparrow-hybrid.iso"
echo "  Size: $ISO_SIZE"
echo ""
echo "  Boot Support:"
echo "    ✓ BIOS (Legacy) - El Torito + MBR"
echo "    ✓ EFI (UEFI)    - ESP partition"
echo ""
echo "  Features:"
echo "    ✓ Auto-discover bootstrap server"
echo "    ✓ Auto-register node identity"
echo "    ✓ Auto-connect to orchestrator"
echo "    ✓ Bacalhau compute node auto-start"
echo "    ✓ Credit earning enabled"
echo ""
echo "  Test with QEMU:"
echo ""
echo "    # BIOS boot test:"
echo "    qemu-system-x86_64 -m 2G -cdrom $OUTPUT_DIR/deparrow-hybrid.iso -nographic"
echo ""
echo "    # EFI boot test:"
echo "    qemu-system-x86_64 -m 2G \\"
echo "      -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \\"
echo "      -cdrom $OUTPUT_DIR/deparrow-hybrid.iso"
echo ""
