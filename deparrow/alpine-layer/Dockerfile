# DEparrow Alpine Linux Base Image
# Lightweight node OS with auto-join capability

FROM alpine:3.21

# Set environment variables
ENV NODE_ARCH=x86_64
ENV NODE_CPU=4
ENV NODE_MEMORY=4GB
ENV NODE_DISK=20GB
ENV NODE_GPU=0
ENV DEPARROW_BOOTSTRAP=https://bootstrap.deparrow.net
ENV DEPARROW_ORCHESTRATOR_HOST=orchestrator.deparrow.net

# Install system dependencies
RUN apk add --no-cache \
    bash \
    curl \
    wget \
    ca-certificates \
    openssl \
    openssh-client \
    docker \
    docker-cli \
    docker-compose \
    openrc \
    sudo \
    tzdata \
    nano \
    htop \
    iftop \
    iotop \
    sysstat \
    logrotate \
    cronie \
    rsync \
    git \
    jq \
    yq \
    python3 \
    py3-pip \
    nodejs \
    npm

# Install Docker (if not already installed)
RUN if ! command -v docker &> /dev/null; then \
    apk add --no-cache docker; \
    fi

# Create DEparrow user and group
RUN addgroup -S deparrow && \
    adduser -S deparrow -G deparrow -s /bin/bash && \
    adduser deparrow docker && \
    echo "deparrow ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers.d/deparrow

# Create directory structure
RUN mkdir -p /etc/deparrow /var/lib/deparrow /var/log/deparrow /opt/deparrow
RUN chown -R deparrow:deparrow /etc/deparrow /var/lib/deparrow /var/log/deparrow /opt/deparrow

# Install Bacalhau binary
WORKDIR /opt/deparrow
RUN wget -q https://github.com/bacalhau-project/bacalhau/releases/download/v1.0.0/bacalhau_linux_amd64 -O bacalhau && \
    chmod +x bacalhau && \
    ln -s /opt/deparrow/bacalhau /usr/local/bin/bacalhau

# Copy DEparrow configuration files
COPY --chown=deparrow:deparrow config/ /etc/deparrow/
COPY --chown=deparrow:deparrow scripts/ /opt/deparrow/scripts/

# Create OpenRC service for Bacalhau
RUN cat > /etc/init.d/bacalhau << 'EOF'
#!/sbin/openrc-run

name="bacalhau"
description="DEparrow Bacalhau Compute Node"
command="/usr/local/bin/bacalhau"
command_args="serve --compute --config /etc/deparrow/bacalhau-layer/deparrow-compute.yaml"
command_user="deparrow"
command_background=true
pidfile="/var/run/${RC_SVCNAME}.pid"

depend() {
    need docker
    need network
    after firewall
}

start_pre() {
    checkpath -f -m 0644 -o deparrow:deparrow /var/log/deparrow/bacalhau.log
}

start() {
    ebegin "Starting ${name}"
    start-stop-daemon --start \
        --exec ${command} \
        --user ${command_user} \
        --pidfile ${pidfile} \
        --make-pidfile \
        --background \
        --stdout /var/log/deparrow/bacalhau.log \
        --stderr /var/log/deparrow/bacalhau.error.log \
        -- ${command_args}
    eend $?
}

stop() {
    ebegin "Stopping ${name}"
    start-stop-daemon --stop \
        --pidfile ${pidfile} \
        --retry 30
    eend $?
}
EOF

RUN chmod +x /etc/init.d/bacalhau

# Create DEparrow node initialization script
RUN cat > /opt/deparrow/scripts/init-node.sh << 'EOF'
#!/bin/bash
set -e

echo "Initializing DEparrow node..."

# Generate node identity if not exists
if [ ! -f /etc/deparrow/node-id ]; then
    echo "Generating new node identity..."
    NODE_ID=$(cat /proc/sys/kernel/random/uuid)
    echo "$NODE_ID" > /etc/deparrow/node-id
    chmod 600 /etc/deparrow/node-id
fi

# Load node identity
NODE_ID=$(cat /etc/deparrow/node-id)
echo "Node ID: $NODE_ID"

# Generate keys if not exists
if [ ! -f /etc/deparrow/keys/private.pem ]; then
    echo "Generating node keys..."
    mkdir -p /etc/deparrow/keys
    openssl genrsa -out /etc/deparrow/keys/private.pem 2048
    openssl rsa -in /etc/deparrow/keys/private.pem -pubout -out /etc/deparrow/keys/public.pem
    chmod 600 /etc/deparrow/keys/private.pem
    chmod 644 /etc/deparrow/keys/public.pem
fi

# Set environment variables
export NODE_ID
export NODE_PUBLIC_KEY=$(cat /etc/deparrow/keys/public.pem | base64 -w 0)
export NODE_PRIVATE_KEY=$(cat /etc/deparrow/keys/private.pem | base64 -w 0)

# Detect system resources
if [ -z "$NODE_CPU" ]; then
    NODE_CPU=$(nproc)
    export NODE_CPU
fi

if [ -z "$NODE_MEMORY" ]; then
    NODE_MEMORY=$(free -h | awk '/^Mem:/ {print $2}')
    export NODE_MEMORY
fi

if [ -z "$NODE_DISK" ]; then
    NODE_DISK=$(df -h / | awk 'NR==2 {print $2}')
    export NODE_DISK
fi

# Detect architecture
if [ -z "$NODE_ARCH" ]; then
    NODE_ARCH=$(uname -m)
    export NODE_ARCH
fi

echo "System detected:"
echo "  CPU: $NODE_CPU cores"
echo "  Memory: $NODE_MEMORY"
echo "  Disk: $NODE_DISK"
echo "  Architecture: $NODE_ARCH"

# Register with DEparrow bootstrap if API key is set
if [ -n "$DEPARROW_API_KEY" ]; then
    echo "Registering with DEparrow bootstrap..."
    curl -X POST \
        -H "Authorization: Bearer $DEPARROW_API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"node_id\":\"$NODE_ID\",\"public_key\":\"$NODE_PUBLIC_KEY\",\"resources\":{\"cpu\":$NODE_CPU,\"memory\":\"$NODE_MEMORY\",\"disk\":\"$NODE_DISK\",\"arch\":\"$NODE_ARCH\"}}" \
        "$DEPARROW_BOOTSTRAP/api/v1/nodes/register"
fi

echo "DEparrow node initialization complete!"
EOF

RUN chmod +x /opt/deparrow/scripts/init-node.sh

# Create health check script
RUN cat > /opt/deparrow/scripts/health-check.sh << 'EOF'
#!/bin/bash

# Check if Bacalhau is running
if ! pgrep -f "bacalhau serve" > /dev/null; then
    echo "Bacalhau process not running"
    exit 1
fi

# Check Docker
if ! docker ps > /dev/null 2>&1; then
    echo "Docker not responding"
    exit 1
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    echo "Disk usage too high: $DISK_USAGE%"
    exit 1
fi

# Check memory
MEM_USAGE=$(free | awk '/^Mem:/ {print int($3/$2 * 100)}')
if [ "$MEM_USAGE" -gt 90 ]; then
    echo "Memory usage too high: $MEM_USAGE%"
    exit 1
fi

echo "Node healthy"
exit 0
EOF

RUN chmod +x /opt/deparrow/scripts/health-check.sh

# Set up log rotation
RUN cat > /etc/logrotate.d/deparrow << 'EOF'
/var/log/deparrow/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 deparrow deparrow
    sharedscripts
    postrotate
        /etc/init.d/bacalhau reload > /dev/null 2>&1 || true
    endscript
}
EOF

# Create entrypoint script
RUN cat > /opt/deparrow/entrypoint.sh << 'EOF'
#!/bin/bash
set -e

echo "=== DEparrow Alpine Node ==="
echo "Starting at $(date)"

# Run initialization
/opt/deparrow/scripts/init-node.sh

# Start Docker
echo "Starting Docker..."
rc-service docker start

# Start Bacalhau
echo "Starting Bacalhau compute node..."
rc-service bacalhau start

# Start health check cron
echo "Starting health monitoring..."
crond -f -l 8 &

# Keep container running
echo "DEparrow node is running..."
tail -f /var/log/deparrow/bacalhau.log
EOF

RUN chmod +x /opt/deparrow/entrypoint.sh

# Set working directory and entrypoint
WORKDIR /opt/deparrow
ENTRYPOINT ["/opt/deparrow/entrypoint.sh"]

# Expose necessary ports
EXPOSE 4222 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD /opt/deparrow/scripts/health-check.sh

# Labels
LABEL org.deparrow.version="1.0.0"
LABEL org.deparrow.arch="multiarch"
LABEL org.deparrow.layer="alpine-base"
LABEL org.deparrow.description="DEparrow Alpine Linux Compute Node"
