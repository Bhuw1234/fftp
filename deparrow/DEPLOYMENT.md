# DEparrow Platform Deployment Guide

This guide covers the deployment of all 4 layers of the DEparrow distributed compute platform.

## Architecture Overview

DEparrow consists of 4 layers:

1. **Bacalhau Execution Network** - Distributed compute orchestration
2. **Alpine Linux Base Layer** - Lightweight OS for compute nodes
3. **Meta-OS Control Plane** - Bootstrap server with credit system
4. **GUI Interface Layer** - Web-based user interface

## Prerequisites

### System Requirements
- Linux, macOS, or WSL2 on Windows
- Docker 20.10+
- Python 3.10+
- Node.js 18+
- 4GB RAM minimum, 8GB recommended
- 10GB free disk space

### Network Requirements
- Port 4222 (NATS messaging)
- Port 8080 (Meta-OS API)
- Port 3000 (GUI development)
- Port 80/443 (GUI production)

## Quick Start Deployment

### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd bacalhau/deparrow

# Run integration test
./test-integration.sh
```

### 2. Install Dependencies
```bash
# Python dependencies for Meta-OS
cd metaos-layer
pip install -r requirements.txt  # Create requirements.txt if needed

# Node.js dependencies for GUI
cd ../gui-layer
npm install
```

### 3. Configure Environment
```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your configuration
# Required variables:
# - DEPARROW_SECRET_KEY (for JWT tokens)
# - DEPARROW_DATABASE_URL
# - DEPARROW_BOOTSTRAP_ADDRESS
```

## Layer-by-Layer Deployment

### Layer 1: Bacalhau Execution Network

#### Option A: Using Provided Configurations
```bash
# Start orchestrator node
bacalhau serve --config bacalhau-layer/deparrow-orchestrator.yaml

# Start compute node (on another machine)
bacalhau serve --config bacalhau-layer/deparrow-compute.yaml
```

#### Option B: Custom Deployment
1. Update `bacalhau-layer/deparrow-orchestrator.yaml` with your settings
2. Update `bacalhau-layer/deparrow-compute.yaml` with bootstrap address
3. Deploy Bacalhau nodes using your preferred method (Docker, Kubernetes, etc.)

### Layer 2: Alpine Linux Base Layer

#### Build Docker Image
```bash
cd alpine-layer
./build.sh

# The script will generate:
# - Docker image: deparrow-node:latest
# - Kubernetes manifests
# - Docker Compose configuration
# - Systemd service files
```

#### Deployment Options

**Docker:**
```bash
docker run -d \
  --name deparrow-node \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  deparrow-node:latest
```

**Kubernetes:**
```bash
kubectl apply -f kubernetes/
```

**Docker Compose:**
```bash
docker-compose up -d
```

### Layer 3: Meta-OS Control Plane

#### Start Bootstrap Server
```bash
cd metaos-layer

# Development mode
python3 bootstrap-server.py

# Production mode (with gunicorn)
pip install gunicorn
gunicorn bootstrap-server:app --bind 0.0.0.0:8080 --worker-class aiohttp.GunicornWebWorker
```

#### Initialize Database
```bash
# The bootstrap server auto-creates SQLite database
# For PostgreSQL/MySQL, update database configuration
```

#### Verify API
```bash
curl http://localhost:8080/api/health
# Should return: {"status": "healthy", "version": "1.0.0"}
```

### Layer 4: GUI Interface Layer

#### Development
```bash
cd gui-layer
npm run dev
# Access at http://localhost:3000
```

#### Production Build
```bash
cd gui-layer
npm run build

# Serve with nginx
docker run -d \
  --name deparrow-gui \
  -p 80:80 \
  -v $(pwd)/dist:/usr/share/nginx/html \
  nginx:alpine
```

## Configuration Details

### Bacalhau Configuration

#### Orchestrator Node (`deparrow-orchestrator.yaml`)
Key settings:
```yaml
node:
  type: requester
  bootstrap_addresses:
    - /ip4/127.0.0.1/tcp/4222  # Meta-OS NATS server
  labels:
    platform: deparrow
    role: orchestrator
```

#### Compute Node (`deparrow-compute.yaml`)
Key settings:
```yaml
node:
  type: compute
  bootstrap_addresses:
    - /ip4/<metaos-ip>/tcp/4222  # Point to Meta-OS
  labels:
    platform: deparrow
    role: compute
```

### Meta-OS Configuration

Environment variables (`.env`):
```env
# Required
DEPARROW_SECRET_KEY=your-secret-key-here
DEPARROW_DATABASE_URL=sqlite:///deparrow.db
DEPARROW_BOOTSTRAP_ADDRESS=0.0.0.0:4222

# Optional
DEPARROW_API_PORT=8080
DEPARROW_DEBUG=true
DEPARROW_JWT_EXPIRE_HOURS=24
DEPARROW_INITIAL_CREDITS=1000
```

### GUI Configuration

Environment variables (`.env` in gui-layer):
```env
VITE_API_URL=http://localhost:8080/api
VITE_APP_NAME=DEparrow
VITE_APP_VERSION=1.0.0
```

## Scaling Deployment

### Horizontal Scaling

#### Multiple Compute Nodes
```bash
# Deploy multiple compute nodes with different labels
for i in {1..5}; do
  docker run -d \
    --name deparrow-node-$i \
    -e NODE_ID=node-$i \
    -e NODE_REGION=us-east-1 \
    deparrow-node:latest
done
```

#### Load Balancing
```nginx
# nginx configuration for GUI
upstream deparrow_api {
    server 127.0.0.1:8080;
    server 192.168.1.2:8080;
    server 192.168.1.3:8080;
}

server {
    listen 80;
    server_name deparrow.example.com;
    
    location /api {
        proxy_pass http://deparrow_api;
    }
    
    location / {
        root /var/www/deparrow-gui;
        try_files $uri $uri/ /index.html;
    }
}
```

### High Availability

#### Meta-OS Cluster
1. Deploy multiple Meta-OS instances
2. Use shared database (PostgreSQL/MySQL)
3. Configure load balancer
4. Enable session sharing

#### Database Backups
```bash
# SQLite backup
sqlite3 deparrow.db ".backup deparrow.backup.db"

# PostgreSQL backup
pg_dump deparrow > deparrow_backup.sql
```

## Monitoring and Maintenance

### Health Checks

#### API Health
```bash
curl http://localhost:8080/api/health
```

#### Node Health
```bash
# Check Bacalhau nodes
bacalhau node list

# Check Docker containers
docker ps | grep deparrow
```

#### Database Health
```bash
# SQLite
sqlite3 deparrow.db "SELECT COUNT(*) FROM users;"

# PostgreSQL
psql -d deparrow -c "SELECT COUNT(*) FROM users;"
```

### Logs

#### Meta-OS Logs
```bash
# Development
tail -f metaos-layer/logs/deparrow.log

# Docker
docker logs deparrow-metaos
```

#### GUI Logs
```bash
# Development
cd gui-layer && npm run dev 2>&1 | tee gui.log

# Production
docker logs deparrow-gui
```

#### Bacalhau Logs
```bash
bacalhau --log-level debug serve
```

### Backup and Recovery

#### Regular Backups
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups/deparrow_$DATE"

mkdir -p $BACKUP_DIR
cp deparrow.db $BACKUP_DIR/
docker exec deparrow-metaos pg_dumpall > $BACKUP_DIR/database.sql
tar -czf $BACKUP_DIR.tar.gz $BACKUP_DIR
```

#### Disaster Recovery
1. Stop all services
2. Restore database from backup
3. Update configuration if needed
4. Start services in order:
   - Meta-OS
   - Bacalhau orchestrator
   - Compute nodes
   - GUI

## Security Considerations

### Network Security
- Use firewalls to restrict access
- Enable TLS/SSL for all endpoints
- Use VPN for internal communication
- Implement rate limiting

### Authentication
- Use strong JWT secret keys
- Implement password policies
- Enable 2FA for admin accounts
- Regular token rotation

### Data Protection
- Encrypt sensitive data at rest
- Use HTTPS for all communications
- Regular security audits
- Access logging and monitoring

## Troubleshooting

### Common Issues

#### Nodes Not Joining
1. Check NATS connectivity: `nc -zv <metaos-ip> 4222`
2. Verify bootstrap addresses in config
3. Check firewall rules
4. Review Bacalhau logs

#### API Connection Errors
1. Verify Meta-OS is running: `curl http://localhost:8080/api/health`
2. Check CORS configuration
3. Verify JWT tokens are valid
4. Check database connection

#### GUI Not Loading
1. Check if GUI server is running
2. Verify API URL configuration
3. Check browser console for errors
4. Verify CORS headers

### Debug Mode

#### Enable Debug Logging
```bash
# Meta-OS
export DEPARROW_DEBUG=true
python3 bootstrap-server.py

# Bacalhau
bacalhau --log-level debug serve

# GUI
npm run dev -- --debug
```

#### Check System Resources
```bash
# CPU and memory
top -b -n 1 | grep -E "(deparrow|bacalhau|node)"

# Disk space
df -h

# Network connections
netstat -tulpn | grep -E "(4222|8080|3000)"
```

## Performance Optimization

### Database Optimization
- Use connection pooling
- Add appropriate indexes
- Regular vacuum (SQLite) or maintenance (PostgreSQL)
- Query optimization

### Caching
```python
# Meta-OS caching example
import aioredis

redis = await aioredis.create_redis_pool('redis://localhost')
await redis.set('key', 'value')
```

### Load Testing
```bash
# Using hey
hey -n 1000 -c 10 http://localhost:8080/api/health
```

## Upgrading

### Version Compatibility
Check version compatibility between layers before upgrading.

### Upgrade Procedure
1. Backup all data
2. Update configuration files
3. Upgrade services in order:
   - Meta-OS
   - Bacalhau nodes
   - GUI
4. Run integration tests
5. Monitor for issues

## Support

### Getting Help
- Check logs for error messages
- Review configuration files
- Test individual components
- Consult documentation

### Reporting Issues
Include:
1. DEparrow version
2. Deployment environment
3. Error logs
4. Steps to reproduce
5. Expected vs actual behavior

## License

DEparrow Platform - Distributed Compute Orchestration System
© 2024 DEparrow Team