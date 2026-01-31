# DEparrow Bootable Installation Guide

## Overview

DEparrow can be installed on bare metal servers, VMs, or existing Linux systems using:
1. **Bootable USB Drive** - For fresh installations
2. **Auto-Install Script** - For existing Linux systems
3. **Docker Containers** - For containerized deployment

## Method 1: Bootable USB Installation (Recommended for New Servers)

### What You Need
- USB drive (8GB minimum, 16GB recommended)
- Server/PC to install on
- Internet connection (for initial setup)

### Step 1: Create Bootable USB

#### Option A: Using Provided Script (Linux/macOS)
```bash
# 1. Download or copy DEparrow files
cd deparrow/bootable

# 2. Create ISO image
chmod +x create-iso.sh
sudo ./create-iso.sh

# 3. Write ISO to USB
sudo ./write-to-usb.sh
```

#### Option B: Manual (Any OS)
1. Download `deparrow-live-1.0.0.iso`
2. Use tool to write to USB:
   - **Windows**: Rufus, Etcher
   - **macOS**: Balena Etcher
   - **Linux**: dd command

### Step 2: Boot from USB

1. Insert USB drive into target machine
2. Enter BIOS/UEFI (usually F2, F12, DEL, ESC during boot)
3. Change boot order to USB first
4. Save and reboot

### Step 3: Install DEparrow

#### Installation Menu
When booted, you'll see:
```
=========================================
        DEparrow Bootable Installer
=========================================

1. Install DEparrow Compute Node
2. Install DEparrow Orchestrator  
3. Install Full DEparrow Platform
4. Run Live System (No Install)
5. Disk Utilities
6. Network Configuration
7. Exit to Shell
8. Reboot
9. Shutdown
```

#### Installation Types

**1. Compute Node** - Joins DEparrow network, earns credits
- Best for: Users with spare compute resources
- Requirements: 2+ CPU cores, 4GB+ RAM, 20GB+ storage
- Earns: Credits for providing compute power

**2. Orchestrator** - Manages network, runs GUI
- Best for: Platform administrators
- Requirements: 4+ CPU cores, 8GB+ RAM, 50GB+ storage
- Controls: Job scheduling, credit system, user management

**3. Full Platform** - Both orchestrator and node
- Best for: All-in-one testing or small deployments
- Requirements: 6+ CPU cores, 12GB+ RAM, 100GB+ storage

### Step 4: Configure Installation

#### For Compute Node:
```
Select option: 1
Enter target disk: /dev/sda
Enter bootstrap address: 192.168.1.100:4222
Enter node name: my-compute-node
Enter region: us-east-1
```

#### For Orchestrator:
```
Select option: 2  
Enter target disk: /dev/sda
Enter admin email: admin@example.com
Enter admin password: ********
```

### Step 5: Complete Installation

1. Installation takes 5-15 minutes
2. System will reboot automatically
3. Remove USB drive when prompted
4. DEparrow starts automatically on boot

## Method 2: Auto-Install on Existing Linux

### Quick Install (One Command)
```bash
# For compute node
curl -sL https://get.deparrow.org/install.sh | sudo bash -s -- --type node --bootstrap 192.168.1.100:4222

# For orchestrator  
curl -sL https://get.deparrow.org/install.sh | sudo bash -s -- --type orchestrator

# For full platform
curl -sL https://get.deparrow.org/install.sh | sudo bash -s -- --type full
```

### Manual Install
```bash
# 1. Download DEparrow
git clone https://github.com/your-org/deparrow.git
cd deparrow/bootable

# 2. Run auto-installer
chmod +x auto-install.sh
sudo ./auto-install.sh --type node --bootstrap 192.168.1.100:4222
```

## Method 3: Docker Container Installation

### Quick Start
```bash
# Run compute node
docker run -d \
  --name deparrow-node \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deparrow/node:latest

# Run orchestrator
docker run -d \
  --name deparrow-orchestrator \
  -p 3000:3000 \
  -p 8080:8080 \
  -p 4222:4222 \
  deparrow/orchestrator:latest
```

## Post-Installation Setup

### For Compute Node Users

#### 1. Check Node Status
```bash
sudo systemctl status deparrow
/opt/deparrow/scripts/status.sh
```

#### 2. View Earnings
1. Access orchestrator GUI: `http://<orchestrator-ip>:3000`
2. Login with your node credentials
3. Check "Wallet" for earned credits

#### 3. Monitor Resources
```bash
# View resource usage
docker stats deparrow-node

# View logs
docker logs -f deparrow-node
journalctl -u deparrow -f
```

### For Orchestrator Administrators

#### 1. Initial Setup
1. Access GUI: `http://<server-ip>:3000`
2. Login with admin credentials
3. Configure platform settings
4. Set credit pricing

#### 2. Add Compute Nodes
1. Share bootstrap address with node operators
2. Nodes auto-join when they boot
3. Monitor in "Nodes" section

#### 3. Manage Users
1. Create user accounts
2. Set credit limits
3. Monitor job submissions

## Usage Examples

### Example 1: Home User with Spare PC
```
User: Home enthusiast with old gaming PC
Goal: Earn credits by providing compute

Steps:
1. Create bootable USB on main computer
2. Boot old PC from USB
3. Install "Compute Node"
4. Enter friend's orchestrator address
5. PC joins network, starts earning
6. Check earnings via web interface
```

### Example 2: Small Business
```
Business: Web hosting company
Goal: Monetize unused servers

Steps:
1. Install "Orchestrator" on main server
2. Install "Compute Nodes" on 5 spare servers
3. Set credit pricing ($0.10 per CPU-hour)
4. Share platform with customers
5. Customers submit jobs, business earns
```

### Example 3: Research Lab
```
Lab: University research group  
Goal: Distributed data processing

Steps:
1. Install "Full Platform" on lab server
2. Students install "Compute Nodes" on laptops
3. Submit research jobs via GUI
4. Process data across all lab machines
5. No external cloud costs
```

## Troubleshooting

### Common Issues

#### USB Won't Boot
- Check BIOS/UEFI settings
- Enable "Legacy Boot" or "CSM"
- Try different USB port
- Recreate USB with different tool

#### Node Won't Join Network
```bash
# Check network connectivity
ping <orchestrator-ip>

# Check NATS port
nc -zv <orchestrator-ip> 4222

# Check logs
docker logs deparrow-node
journalctl -u deparrow -f
```

#### GUI Not Accessible
```bash
# Check if services are running
systemctl status deparrow

# Check ports
netstat -tulpn | grep -E "(3000|8080)"

# Check firewall
sudo ufw allow 3000/tcp
sudo ufw allow 8080/tcp
```

#### Low Earnings
- Check node has sufficient resources
- Verify node is properly registered
- Check orchestrator has job demand
- Ensure network connectivity

### Recovery

#### Reset Node
```bash
# Stop DEparrow
sudo systemctl stop deparrow

# Remove configuration
sudo rm -rf /opt/deparrow/config.env

# Restart
sudo systemctl start deparrow
```

#### Reinstall
```bash
# Complete reinstall
sudo systemctl stop deparrow
sudo rm -rf /opt/deparrow
sudo ./auto-install.sh --type node --bootstrap <new-address>
```

## Security Best Practices

### For Node Operators
1. Use strong passwords
2. Enable firewall
3. Regular system updates
4. Monitor resource usage
5. Use separate network if possible

### For Orchestrator Admins
1. Change default admin password
2. Enable HTTPS for GUI/API
3. Regular backups
4. Monitor for suspicious activity
5. Set credit limits per user

## Advanced Configuration

### Custom Resource Limits
Edit `/opt/deparrow/config.env`:
```bash
# Limit CPU usage
MAX_CPU_PERCENT=50

# Limit memory
MAX_MEMORY_GB=4

# Limit storage
MAX_STORAGE_GB=20
```

### Multiple Network Interfaces
```bash
# Specify interface
NETWORK_INTERFACE=eth1

# Static IP
NODE_IP=192.168.2.100
BOOTSTRAP_ADDRESS=192.168.2.1:4222
```

### Proxy Support
```bash
# Set proxy for node
HTTP_PROXY=http://proxy.example.com:8080
HTTPS_PROXY=http://proxy.example.com:8080
```

## Support

### Getting Help
1. Check logs: `journalctl -u deparrow -f`
2. Community forum: forum.deparrow.org
3. Documentation: docs.deparrow.org

### Reporting Issues
Include:
- DEparrow version
- Installation type
- Error messages
- System specifications
- Steps to reproduce

## FAQ

### Q: How much can I earn?
A: Depends on:
- Your hardware (CPU, GPU, RAM)
- Network demand
- Credit pricing set by orchestrator
- Typical: $0.05-$0.50 per CPU-hour

### Q: Is my data safe?
A: Yes:
- Jobs run in isolated containers
- No access to host filesystem
- Encrypted communications
- You control what runs

### Q: Can I use this for mining?
A: No:
- DEparrow is for general compute
- Mining is specifically blocked
- Focus on legitimate workloads

### Q: What internet speed needed?
A: Minimum:
- Download: 10 Mbps
- Upload: 5 Mbps
- Latency: < 100ms to orchestrator

### Q: Can I run on Raspberry Pi?
A: Yes:
- ARM64 supported
- Use Alpine Linux version
- Lower resource requirements

## Updates

### Check for Updates
```bash
cd /opt/deparrow
git pull origin main
sudo systemctl restart deparrow
```

### Automatic Updates (Optional)
```bash
# Create cron job
echo "0 2 * * * cd /opt/deparrow && git pull && systemctl restart deparrow" | sudo crontab -
```

---

**Ready to start?** Choose your installation method above and join the DEparrow network!