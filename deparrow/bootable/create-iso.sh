#!/bin/bash

# DEparrow Bootable ISO Creator
# Creates a bootable USB/ISO image for installing DEparrow on bare metal

set -e

echo "========================================="
echo "DEparrow Bootable ISO Creator"
echo "========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
ISO_NAME="deparrow-live"
ISO_VERSION="1.0.0"
WORK_DIR="/tmp/deparrow-iso"
ISO_DIR="$WORK_DIR/iso"
BOOT_DIR="$ISO_DIR/boot"
GRUB_DIR="$BOOT_DIR/grub"
LIVE_DIR="$ISO_DIR/live"
INSTALL_DIR="$ISO_DIR/install"
CONFIG_DIR="$ISO_DIR/config"

# Check prerequisites
check_prereqs() {
    echo -e "${BLUE}Checking prerequisites...${NC}"
    
    local missing=0
    for cmd in xorriso grub-mkrescue mkisofs curl docker; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}✗ $cmd not found${NC}"
            missing=1
        else
            echo -e "${GREEN}✓ $cmd found${NC}"
        fi
    done
    
    if [ $missing -eq 1 ]; then
        echo -e "${YELLOW}Install missing packages:${NC}"
        echo "  Ubuntu/Debian: sudo apt install xorriso grub-common grub-pc-bin curl docker.io"
        echo "  Fedora/RHEL: sudo dnf install xorriso grub2-tools curl docker"
        exit 1
    fi
}

# Clean previous builds
cleanup() {
    echo -e "${BLUE}Cleaning up...${NC}"
    rm -rf "$WORK_DIR"
    mkdir -p "$ISO_DIR" "$BOOT_DIR" "$GRUB_DIR" "$LIVE_DIR" "$INSTALL_DIR" "$CONFIG_DIR"
}

# Create directory structure
create_structure() {
    echo -e "${BLUE}Creating ISO directory structure...${NC}"
    
    # Create essential directories
    mkdir -p "$ISO_DIR/"{boot/grub,live,install,config,deparrow}
    
    # Copy DEparrow files
    echo "Copying DEparrow platform files..."
    cp -r ../bacalhau-layer "$ISO_DIR/deparrow/"
    cp -r ../alpine-layer "$ISO_DIR/deparrow/"
    cp -r ../metaos-layer "$ISO_DIR/deparrow/"
    cp -r ../gui-layer "$ISO_DIR/deparrow/"
    cp ../test-integration.sh "$ISO_DIR/deparrow/"
    cp ../DEPLOYMENT.md "$ISO_DIR/deparrow/"
    
    # Create version file
    echo "DEparrow Bootable ISO v$ISO_VERSION" > "$ISO_DIR/deparrow/VERSION"
    echo "Build date: $(date)" >> "$ISO_DIR/deparrow/VERSION"
}

# Create Alpine Linux live system
create_live_system() {
    echo -e "${BLUE}Creating Alpine Linux live system...${NC}"
    
    # Download Alpine Linux mini rootfs
    ALPINE_VERSION="3.20"
    ARCH="x86_64"
    ROOTFS_URL="https://dl-cdn.alpinelinux.org/alpine/v${ALPINE_VERSION}/releases/${ARCH}/alpine-minirootfs-${ALPINE_VERSION}.0-${ARCH}.tar.gz"
    
    echo "Downloading Alpine Linux rootfs..."
    curl -L -o "$WORK_DIR/alpine-rootfs.tar.gz" "$ROOTFS_URL"
    
    # Extract to live directory
    echo "Extracting rootfs..."
    tar -xzf "$WORK_DIR/alpine-rootfs.tar.gz" -C "$LIVE_DIR"
    
    # Create essential directories in live system
    mkdir -p "$LIVE_DIR/"{dev,proc,sys,run,tmp}
    
    # Create fstab
    cat > "$LIVE_DIR/etc/fstab" << EOF
# <file system> <mount point> <type> <options> <dump> <pass>
proc /proc proc defaults 0 0
sysfs /sys sysfs defaults 0 0
devpts /dev/pts devpts defaults 0 0
tmpfs /dev/shm tmpfs defaults 0 0
EOF
}

# Create installer scripts
create_installer() {
    echo -e "${BLUE}Creating installer scripts...${NC}"
    
    # Main installer script
    cat > "$INSTALL_DIR/install.sh" << 'EOF'
#!/bin/bash

# DEparrow Bare Metal Installer
# Run from live boot environment

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
TARGET_DISK=""
INSTALL_TYPE="node"  # node, orchestrator, or full
DEPARROW_VERSION="1.0.0"
INSTALL_DIR="/mnt/install"
DEPARROW_SRC="/cdrom/deparrow"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --disk)
            TARGET_DISK="$2"
            shift 2
            ;;
        --type)
            INSTALL_TYPE="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

show_help() {
    cat << HELP
DEparrow Bare Metal Installer

Usage: install.sh [options]

Options:
  --disk DEVICE    Target disk device (e.g., /dev/sda)
  --type TYPE      Installation type: node, orchestrator, or full
  --help           Show this help message

Examples:
  install.sh --disk /dev/sda --type node
  install.sh --disk /dev/nvme0n1 --type orchestrator
HELP
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}Please run as root${NC}"
        exit 1
    fi
}

# Detect disks
detect_disks() {
    echo -e "${YELLOW}Available disks:${NC}"
    lsblk -d -o NAME,SIZE,TYPE,MODEL | grep -v loop
    echo ""
}

# Partition disk
partition_disk() {
    local disk=$1
    
    echo -e "${YELLOW}Partitioning $disk...${NC}"
    
    # Clear existing partition table
    sgdisk -Z "$disk"
    
    # Create GPT partition table
    sgdisk -o "$disk"
    
    # Create EFI partition (512MB)
    sgdisk -n 1:0:+512M -t 1:ef00 -c 1:"EFI System" "$disk"
    
    # Create boot partition (1GB)
    sgdisk -n 2:0:+1G -t 2:8300 -c 2:"Linux boot" "$disk"
    
    # Create root partition (rest of disk)
    sgdisk -n 3:0:0 -t 3:8300 -c 3:"Linux root" "$disk"
    
    # Inform kernel of partition changes
    partprobe "$disk"
    
    # Format partitions
    mkfs.vfat -F32 "${disk}1"
    mkfs.ext4 "${disk}2"
    mkfs.ext4 "${disk}3"
}

# Mount partitions
mount_partitions() {
    local disk=$1
    
    echo -e "${YELLOW}Mounting partitions...${NC}"
    
    mkdir -p "$INSTALL_DIR"
    mount "${disk}3" "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/boot"
    mount "${disk}2" "$INSTALL_DIR/boot"
    mkdir -p "$INSTALL_DIR/boot/efi"
    mount "${disk}1" "$INSTALL_DIR/boot/efi"
}

# Install base system
install_base() {
    echo -e "${YELLOW}Installing Alpine Linux base system...${NC}"
    
    # Setup apk repositories
    cat > "$INSTALL_DIR/etc/apk/repositories" << REPOS
https://dl-cdn.alpinelinux.org/alpine/v3.20/main
https://dl-cdn.alpinelinux.org/alpine/v3.20/community
REPOS
    
    # Copy rootfs from live system
    cp -a /cdrom/live/* "$INSTALL_DIR/"
    
    # Install essential packages
    chroot "$INSTALL_DIR" apk update
    chroot "$INSTALL_DIR" apk add \
        alpine-base \
        linux-lts \
        grub grub-efi \
        e2fsprogs \
        dosfstools \
        parted \
        curl \
        docker \
        openrc \
        openssh \
        sudo \
        bash \
        python3 \
        nodejs \
        npm
}

# Install DEparrow
install_deparrow() {
    echo -e "${YELLOW}Installing DEparrow platform...${NC}"
    
    # Copy DEparrow files
    mkdir -p "$INSTALL_DIR/opt/deparrow"
    cp -r "$DEPARROW_SRC"/* "$INSTALL_DIR/opt/deparrow/"
    
    # Create installation type file
    echo "$INSTALL_TYPE" > "$INSTALL_DIR/opt/deparrow/INSTALL_TYPE"
    
    # Create systemd service
    cat > "$INSTALL_DIR/etc/systemd/system/deparrow.service" << SERVICE
[Unit]
Description=DEparrow Compute Node
After=docker.service network.target
Wants=docker.service network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/deparrow
ExecStart=/opt/deparrow/scripts/start-$INSTALL_TYPE.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICE
    
    # Create startup scripts
    mkdir -p "$INSTALL_DIR/opt/deparrow/scripts"
    
    # Node startup script
    cat > "$INSTALL_DIR/opt/deparrow/scripts/start-node.sh" << 'NODE_SCRIPT'
#!/bin/bash
# DEparrow Compute Node Startup Script

# Wait for network
sleep 5

# Start Docker
systemctl start docker

# Load DEparrow Alpine image
cd /opt/deparrow/alpine-layer
./build.sh

# Run DEparrow node
docker run -d \
  --name deparrow-node \
  --network host \
  --restart always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /opt/deparrow:/deparrow \
  deparrow-node:latest

echo "DEparrow compute node started"
NODE_SCRIPT
    
    # Orchestrator startup script
    cat > "$INSTALL_DIR/opt/deparrow/scripts/start-orchestrator.sh" << 'ORCH_SCRIPT'
#!/bin/bash
# DEparrow Orchestrator Startup Script

# Wait for network
sleep 5

# Start Docker
systemctl start docker

# Start Meta-OS bootstrap
cd /opt/deparrow/metaos-layer
python3 bootstrap-server.py &

# Start Bacalhau orchestrator
bacalhau serve --config /opt/deparrow/bacalhau-layer/deparrow-orchestrator.yaml &

# Start GUI (optional)
cd /opt/deparrow/gui-layer
npm run build
npx serve -s dist -l 3000 &

echo "DEparrow orchestrator started"
ORCH_SCRIPT
    
    # Full platform startup script
    cat > "$INSTALL_DIR/opt/deparrow/scripts/start-full.sh" << 'FULL_SCRIPT'
#!/bin/bash
# DEparrow Full Platform Startup Script

# Start all components
/opt/deparrow/scripts/start-orchestrator.sh
sleep 10
/opt/deparrow/scripts/start-node.sh

echo "DEparrow full platform started"
FULL_SCRIPT
    
    chmod +x "$INSTALL_DIR/opt/deparrow/scripts/"*.sh
}

# Configure system
configure_system() {
    echo -e "${YELLOW}Configuring system...${NC}"
    
    # Set hostname
    echo "deparrow-$INSTALL_TYPE" > "$INSTALL_DIR/etc/hostname"
    
    # Configure network
    cat > "$INSTALL_DIR/etc/network/interfaces" << NETWORK
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp
NETWORK
    
    # Create deparrow user
    chroot "$INSTALL_DIR" adduser -D -s /bin/bash deparrow
    chroot "$INSTALL_DIR" addgroup deparrow docker
    chroot "$INSTALL_DIR" addgroup deparrow sudo
    echo "deparrow ALL=(ALL) NOPASSWD:ALL" > "$INSTALL_DIR/etc/sudoers.d/deparrow"
    
    # Enable services
    chroot "$INSTALL_DIR" rc-update add docker default
    chroot "$INSTALL_DIR" rc-update add sshd default
    chroot "$INSTALL_DIR" rc-update add deparrow default
}

# Install bootloader
install_bootloader() {
    echo -e "${YELLOW}Installing bootloader...${NC}"
    
    # Install GRUB
    chroot "$INSTALL_DIR" grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=DEparrow
    chroot "$INSTALL_DIR" grub-mkconfig -o /boot/grub/grub.cfg
    
    # Create GRUB configuration
    cat > "$INSTALL_DIR/boot/grub/grub.cfg" << GRUB
set timeout=5
set default=0

menuentry "DEparrow Compute Platform" {
    linux /boot/vmlinuz-lts quiet root=/dev/sda3
    initrd /boot/initramfs-lts
}

menuentry "DEparrow Installer" {
    linux /boot/vmlinuz-lts quiet root=/cdrom/live
    initrd /boot/initramfs-lts
}
GRUB
}

# Main installation function
main_install() {
    check_root
    
    if [ -z "$TARGET_DISK" ]; then
        echo -e "${RED}No target disk specified${NC}"
        detect_disks
        exit 1
    fi
    
    if [ ! -b "$TARGET_DISK" ]; then
        echo -e "${RED}$TARGET_DISK is not a block device${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Starting DEparrow installation...${NC}"
    echo "Target disk: $TARGET_DISK"
    echo "Install type: $INSTALL_TYPE"
    echo ""
    
    read -p "This will ERASE ALL DATA on $TARGET_DISK. Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Installation cancelled"
        exit 0
    fi
    
    # Installation steps
    partition_disk "$TARGET_DISK"
    mount_partitions "$TARGET_DISK"
    install_base
    install_deparrow
    configure_system
    install_bootloader
    
    # Cleanup
    umount -R "$INSTALL_DIR"
    
    echo -e "${GREEN}Installation complete!${NC}"
    echo "Reboot to start DEparrow $INSTALL_TYPE"
}

# Run installation
main_install
EOF
    
    chmod +x "$INSTALL_DIR/install.sh"
    
    # Create auto-install configuration
    cat > "$CONFIG_DIR/autoinstall.yaml" << EOF
# DEparrow Auto-Install Configuration
version: 1
identity:
  hostname: deparrow-node
  username: deparrow
  password: deparrow123

storage:
  layout:
    name: direct
    match:
      size: largest
  swap:
    size: 0

network:
  version: 2
  ethernets:
    eth0:
      dhcp4: true

deparrow:
  install_type: node
  bootstrap_address: 192.168.1.100:4222
  region: us-east-1
EOF
    
    # Create simple menu script
    cat > "$ISO_DIR/autorun.sh" << 'EOF'
#!/bin/bash

# DEparrow Boot Menu

while true; do
    clear
    echo "========================================="
    echo "        DEparrow Bootable Installer"
    echo "========================================="
    echo ""
    echo "1. Install DEparrow Compute Node"
    echo "2. Install DEparrow Orchestrator"
    echo "3. Install Full DEparrow Platform"
    echo "4. Run Live System (No Install)"
    echo "5. Disk Utilities"
    echo "6. Network Configuration"
    echo "7. Exit to Shell"
    echo "8. Reboot"
    echo "9. Shutdown"
    echo ""
    read -p "Select option: " choice
    
    case $choice in
        1)
            ./install.sh --type node
            ;;
        2)
            ./install.sh --type orchestrator
            ;;
        3)
            ./install.sh --type full
            ;;
        4)
            echo "Starting live system..."
            /bin/bash
            ;;
        5)
            fdisk -l
            read -p "Press Enter to continue..."
            ;;
        6)
            ip addr show
            read -p "Press Enter to continue..."
            ;;
        7)
            /bin/bash
            ;;
        8)
            reboot
            ;;
        9)
            poweroff
            ;;
        *)
            echo "Invalid option"
            ;;
    esac
done
EOF
    
    chmod +x "$ISO_DIR/autorun.sh"
}

# Create GRUB configuration
create_grub_config() {
    echo -e "${BLUE}Creating GRUB configuration...${NC}"
    
    cat > "$GRUB_DIR/grub.cfg" << GRUB
set timeout=10
set default=0

menuentry "DEparrow Live & Install" {
    linux /boot/vmlinuz-lts quiet root=/cdrom/live
    initrd /boot/initramfs-lts
}

menuentry "DEparrow Installer (Text Mode)" {
    linux /boot/vmlinuz-lts quiet root=/cdrom/live console=tty0 console=ttyS0,115200
    initrd /boot/initramfs-lts
}

menuentry "Memory Test" {
    linux16 /boot/memtest86+.bin
}

menuentry "Reboot" {
    reboot
}

menuentry "Shutdown" {
    halt
}
GRUB
}

# Build ISO
build_iso() {
    echo -e "${BLUE}Building ISO image...${NC}"
    
    # Copy kernel and initramfs
    cp "$LIVE_DIR/boot/vmlinuz-lts" "$BOOT_DIR/"
    cp "$LIVE_DIR/boot/initramfs-lts" "$BOOT_DIR/"
    
    # Create ISO
    xorriso -as mkisofs \
        -iso-level 3 \
        -full-iso9660-filenames \
        -volid "DEparrow-$ISO_VERSION" \
        -eltorito-boot boot/grub/grub.cfg \
        -eltorito-catalog boot/grub/boot.cat \
        -no-emul-boot -boot-load-size 4 -boot-info-table \
        -isohybrid-mbr /usr/lib/grub/i386-pc/boot_hybrid.img \
        -output "$WORK_DIR/$ISO_NAME-$ISO_VERSION.iso" \
        "$ISO_DIR"
    
    echo -e "${GREEN}ISO created: $WORK_DIR/$ISO_NAME-$ISO_VERSION.iso${NC}"
    ls -lh "$WORK_DIR/$ISO_NAME-$ISO_VERSION.iso"
}

# Create USB writing script
create_usb_script() {
    echo -e "${BLUE}Creating USB writing script...${NC}"
    
    cat > "$WORK_DIR/write-to-usb.sh" << 'USB_SCRIPT'
#!/bin/bash

# DEparrow USB Writer
# Writes ISO to USB drive for bootable installation

set -e

ISO_FILE="deparrow-live-1.0.0.iso"

# Check root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root: sudo $0"
    exit 1
fi

# List available disks
echo "Available disks:"
lsblk -d -o NAME,SIZE,TYPE,MODEL | grep -v loop
echo ""

# Get target USB device
read -p "Enter USB device (e.g., /dev/sdb): " USB_DEVICE

if [ -z "$USB_DEVICE" ]; then
    echo "No device specified"
    exit 1
fi

if [ ! -b "$USB_DEVICE" ]; then
    echo "$USB_DEVICE is not a block device"
    exit 1
fi

# Confirm
echo ""
echo "WARNING: This will ERASE ALL DATA on $USB_DEVICE"
read -p "Are you sure? (type YES to continue): " CONFIRM

if [ "$CONFIRM" != "YES" ]; then
    echo "Operation cancelled"
    exit 0
fi

# Write ISO to USB
echo "Writing ISO to $USB_DEVICE..."
dd if="$ISO_FILE" of="$USB_DEVICE" bs=4M status=progress oflag=sync

echo "Done! USB drive is ready for booting."
echo "Boot from USB to install DEparrow."
USB_SCRIPT
    
    chmod +x "$WORK_DIR/write-to-usb.sh"
}

# Main execution
main() {
    echo -e "${BLUE}Starting DEparrow bootable ISO creation...${NC}"
    
    check_prereqs
    cleanup
    create_structure
    create_live_system
    create_installer
    create_grub_config
    build_iso
    create_usb_script
    
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}Bootable ISO creation complete!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    echo "Files created:"
    echo "  • ISO: $WORK_DIR/$ISO_NAME-$ISO_VERSION.iso"
    echo "  • USB writer: $WORK_DIR/write-to-usb.sh"
    echo ""
    echo "To create bootable USB:"
    echo "  1. sudo $WORK_DIR/write-to-usb.sh"
    echo "  2. Select your USB device"
    echo "  3. Boot from USB to install"
    echo ""
    echo "Installation types:"
    echo "  • Compute Node: Joins DEparrow network, earns credits"
    echo "  • Orchestrator: Manages jobs and nodes"
    echo "  • Full Platform: Both orchestrator and node"
}

# Run main function
main
