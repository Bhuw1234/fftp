#!/bin/bash
set -e

echo "=== Building DEparrow Alpine Linux Node Image ==="

# Configuration
IMAGE_NAME="deparrow/alpine-node"
VERSION="1.0.0"
PLATFORMS="linux/amd64,linux/arm64"

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ALPINE_LAYER_DIR="$(dirname "$SCRIPT_DIR")"
DEPARROW_DIR="$(dirname "$ALPINE_LAYER_DIR")"
PROJECT_ROOT="$(dirname "$DEPARROW_DIR")"
PICOCLAW_DIR="${PROJECT_ROOT}/picoclaw"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Check Docker buildx
    if ! docker buildx version &> /dev/null; then
        log_warning "Docker Buildx not available, installing..."
        docker buildx install
    fi
    
    # Create builder instance if needed
    if ! docker buildx ls | grep -q "deparrow-builder"; then
        log_info "Creating Docker Buildx builder..."
        docker buildx create --name deparrow-builder --use
    fi
    
    # Check PicoClaw source
    if [ ! -d "$PICOCLAW_DIR" ]; then
        log_error "PicoClaw source not found at: $PICOCLAW_DIR"
        exit 1
    fi
    
    if [ ! -f "$PICOCLAW_DIR/go.mod" ]; then
        log_error "PicoClaw go.mod not found"
        exit 1
    fi
    
    log_success "Prerequisites satisfied"
}

# Build PicoClaw for target architectures
build_picoclaw() {
    log_info "Building PicoClaw binaries..."
    
    local build_dir="${ALPINE_LAYER_DIR}/build"
    mkdir -p "$build_dir"
    
    # Build for linux/amd64
    log_info "Building PicoClaw for linux/amd64..."
    cd "$PICOCLAW_DIR"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "${build_dir}/picoclaw-linux-amd64" ./cmd/picoclaw
    
    # Build for linux/arm64
    log_info "Building PicoClaw for linux/arm64..."
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o "${build_dir}/picoclaw-linux-arm64" ./cmd/picoclaw
    
    log_success "PicoClaw binaries built"
    
    cd "$ALPINE_LAYER_DIR"
}

# Build multi-architecture image
build_image() {
    log_info "Building multi-architecture Docker image..."
    
    # Create build context
    local build_context="${ALPINE_LAYER_DIR}/build-context"
    mkdir -p "$build_context"
    
    # Copy Dockerfile
    cp "${ALPINE_LAYER_DIR}/Dockerfile" "$build_context/"
    
    # Copy config
    cp -r "${ALPINE_LAYER_DIR}/config" "$build_context/"
    
    # Copy scripts
    cp -r "${ALPINE_LAYER_DIR}/scripts" "$build_context/"
    
    # Copy PicoClaw source for in-image build
    log_info "Copying PicoClaw source..."
    cp -r "$PICOCLAW_DIR" "$build_context/picoclaw"
    
    # Build for multiple platforms
    docker buildx build \
        --platform ${PLATFORMS} \
        -t ${IMAGE_NAME}:${VERSION} \
        -t ${IMAGE_NAME}:latest \
        --build-context picoclaw-source="$PICOCLAW_DIR" \
        --push \
        "$build_context"
    
    # Cleanup
    rm -rf "$build_context"
    
    log_success "Image built and pushed: ${IMAGE_NAME}:${VERSION}"
    log_success "Platforms: ${PLATFORMS}"
}

# Build single architecture image (for local testing)
build_local() {
    local arch="${1:-amd64}"
    
    log_info "Building local image for ${arch}..."
    
    # Create build context
    local build_context="${ALPINE_LAYER_DIR}/build-context"
    mkdir -p "$build_context"
    
    # Copy files
    cp "${ALPINE_LAYER_DIR}/Dockerfile" "$build_context/"
    cp -r "${ALPINE_LAYER_DIR}/config" "$build_context/"
    cp -r "${ALPINE_LAYER_DIR}/scripts" "$build_context/"
    cp -r "$PICOCLAW_DIR" "$build_context/picoclaw"
    
    # Build for single platform
    docker build \
        --platform linux/${arch} \
        -t ${IMAGE_NAME}:${VERSION}-${arch} \
        -t ${IMAGE_NAME}:local \
        --build-arg TARGETARCH=${arch} \
        "$build_context"
    
    # Cleanup
    rm -rf "$build_context"
    
    log_success "Local image built: ${IMAGE_NAME}:${VERSION}-${arch}"
}

# Generate deployment manifests
generate_manifests() {
    log_info "Generating deployment manifests..."
    
    local config_dir="${DEPARROW_DIR}/config"
    mkdir -p "${config_dir}/kubernetes"
    mkdir -p "${config_dir}/docker-compose"
    mkdir -p "${config_dir}/systemd"
    
    # Kubernetes deployment
    cat > "${config_dir}/kubernetes/deployment.yaml" << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deparrow-node
  namespace: deparrow
  labels:
    app: deparrow
    component: node
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deparrow
      component: node
  template:
    metadata:
      labels:
        app: deparrow
        component: node
    spec:
      containers:
      - name: deparrow-node
        image: ${IMAGE_NAME}:${VERSION}
        imagePullPolicy: Always
        ports:
        - containerPort: 4222
          name: nats
        - containerPort: 9090
          name: metrics
        - containerPort: 18790
          name: picoclaw
        env:
        - name: NODE_ARCH
          value: "\$(NODE_ARCH)"
        - name: DEPARROW_API_KEY
          valueFrom:
            secretKeyRef:
              name: deparrow-secrets
              key: api-key
        - name: PICOCLAW_API_KEY
          valueFrom:
            secretKeyRef:
              name: deparrow-secrets
              key: picoclaw-api-key
        - name: DEPARROW_BOOTSTRAP
          value: "https://bootstrap.deparrow.net"
        - name: DEPARROW_ORCHESTRATOR_HOST
          value: "orchestrator.deparrow.net"
        resources:
          requests:
            cpu: "100m"
            memory: "256Mi"
          limits:
            cpu: "2"
            memory: "4Gi"
        securityContext:
          privileged: true
          capabilities:
            add:
            - SYS_ADMIN
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock
        - name: deparrow-data
          mountPath: /var/lib/deparrow
        - name: deparrow-config
          mountPath: /etc/deparrow
      volumes:
      - name: docker-socket
        hostPath:
          path: /var/run/docker.sock
      - name: deparrow-data
        hostPath:
          path: /var/lib/deparrow
          type: DirectoryOrCreate
      - name: deparrow-config
        secret:
          secretName: deparrow-config
      nodeSelector:
        kubernetes.io/arch: "\$(NODE_ARCH)"
---
apiVersion: v1
kind: Service
metadata:
  name: deparrow-node
  namespace: deparrow
spec:
  selector:
    app: deparrow
    component: node
  ports:
  - port: 4222
    targetPort: nats
    name: nats
  - port: 9090
    targetPort: metrics
    name: metrics
  - port: 18790
    targetPort: picoclaw
    name: picoclaw
  type: ClusterIP
EOF
    
    # Docker Compose configuration
    cat > "${config_dir}/docker-compose/deparrow-node.yml" << EOF
version: '3.8'

services:
  deparrow-node:
    image: ${IMAGE_NAME}:${VERSION}
    container_name: deparrow-node
    restart: unless-stopped
    privileged: true
    ports:
      - "4222:4222"
      - "9090:9090"
      - "18790:18790"
    environment:
      - NODE_ARCH=amd64
      - NODE_CPU=4
      - NODE_MEMORY=4GB
      - NODE_DISK=20GB
      - DEPARROW_API_KEY=\${DEPARROW_API_KEY}
      - DEPARROW_BOOTSTRAP=https://bootstrap.deparrow.net
      - DEPARROW_ORCHESTRATOR_HOST=orchestrator.deparrow.net
      - PICOCLAW_API_KEY=\${PICOCLAW_API_KEY}
      - PICOCLAW_MODEL=gpt-4o-mini
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - deparrow-data:/var/lib/deparrow
      - deparrow-config:/etc/deparrow
    networks:
      - deparrow-net

volumes:
  deparrow-data:
  deparrow-config:

networks:
  deparrow-net:
    driver: bridge
EOF
    
    # Systemd service file
    cat > "${config_dir}/systemd/deparrow-node.service" << EOF
[Unit]
Description=DEparrow Compute Node with PicoClaw Agent
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=deparrow
Group=deparrow
Environment="NODE_ARCH=\$(uname -m)"
Environment="DEPARROW_API_KEY=\${DEPARROW_API_KEY}"
Environment="PICOCLAW_API_KEY=\${PICOCLAW_API_KEY}"
Environment="DEPARROW_BOOTSTRAP=https://bootstrap.deparrow.net"
Environment="DEPARROW_ORCHESTRATOR_HOST=orchestrator.deparrow.net"
ExecStartPre=/usr/bin/docker pull ${IMAGE_NAME}:${VERSION}
ExecStart=/usr/bin/docker run --rm \\
  --name deparrow-node \\
  --privileged \\
  -p 4222:4222 \\
  -p 9090:9090 \\
  -p 18790:18790 \\
  -e NODE_ARCH \\
  -e DEPARROW_API_KEY \\
  -e PICOCLAW_API_KEY \\
  -e DEPARROW_BOOTSTRAP \\
  -e DEPARROW_ORCHESTRATOR_HOST \\
  -v /var/run/docker.sock:/var/run/docker.sock \\
  -v /var/lib/deparrow:/var/lib/deparrow \\
  -v /etc/deparrow:/etc/deparrow \\
  ${IMAGE_NAME}:${VERSION}
ExecStop=/usr/bin/docker stop deparrow-node
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    
    log_success "Deployment manifests generated"
}

# Build ISO image
build_iso() {
    log_info "Building bootable ISO image..."
    
    local iso_dir="${ALPINE_LAYER_DIR}/iso-build"
    mkdir -p "$iso_dir"
    
    # This would use Alpine's mkimage or similar tools
    # For now, we create a Docker-based ISO builder
    
    cat > "${iso_dir}/build-iso.sh" << 'EOFSH'
#!/bin/bash
set -e

# Build bootable ISO using Alpine Linux tools
# This requires root or proper permissions

ARCH="${1:-x86_64}"
OUTPUT_DIR="${2:-/tmp/deparrow-iso}"

mkdir -p "$OUTPUT_DIR"

# Create Alpine rootfs
apk add --root "$OUTPUT_DIR/rootfs" --initdb \
    alpine-base linux-virt openrc docker docker-cli-compose \
    bash curl wget jq yq python3 nodejs npm

# Install DEparrow binaries
cp /opt/deparrow/bacalhau "$OUTPUT_DIR/rootfs/usr/local/bin/"
cp /usr/local/bin/picoclaw "$OUTPUT_DIR/rootfs/usr/local/bin/"
ln -s picoclaw "$OUTPUT_DIR/rootfs/usr/local/bin/deparrow-agent"

# Create init scripts
cp /etc/init.d/bacalhau "$OUTPUT_DIR/rootfs/etc/init.d/"
cp /etc/init.d/picoclaw "$OUTPUT_DIR/rootfs/etc/init.d/"

# Configure auto-start
chroot "$OUTPUT_DIR/rootfs" rc-update add bacalhau default
chroot "$OUTPUT_DIR/rootfs" rc-update add picoclaw default

echo "ISO build complete at $OUTPUT_DIR"
EOFSH
    
    chmod +x "${iso_dir}/build-iso.sh"
    
    log_warning "ISO building requires additional setup. Script created at ${iso_dir}/build-iso.sh"
    log_info "Run with: sudo ${iso_dir}/build-iso.sh [arch] [output_dir]"
}

# Clean build artifacts
clean() {
    log_info "Cleaning build artifacts..."
    
    rm -rf "${ALPINE_LAYER_DIR}/build"
    rm -rf "${ALPINE_LAYER_DIR}/build-context"
    
    log_success "Clean complete"
}

# Show usage
show_usage() {
    echo "DEparrow Alpine Node Builder"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  all          Build everything (default)"
    echo "  check        Check prerequisites only"
    echo "  picoclaw     Build PicoClaw binaries only"
    echo "  image        Build Docker image only"
    echo "  local [arch] Build local image for testing (default: amd64)"
    echo "  manifests    Generate deployment manifests"
    echo "  iso          Build bootable ISO image"
    echo "  clean        Clean build artifacts"
    echo ""
    echo "Examples:"
    echo "  $0 all              # Build everything"
    echo "  $0 local arm64      # Build local ARM64 image"
    echo "  $0 clean            # Clean artifacts"
}

# Main execution
main() {
    local command="${1:-all}"
    
    echo "DEparrow Alpine Node Builder"
    echo "============================"
    echo ""
    
    case "$command" in
        all)
            check_prerequisites
            build_picoclaw
            build_image
            generate_manifests
            ;;
        check)
            check_prerequisites
            ;;
        picoclaw)
            check_prerequisites
            build_picoclaw
            ;;
        image)
            check_prerequisites
            build_image
            ;;
        local)
            check_prerequisites
            build_local "${2:-amd64}"
            ;;
        manifests)
            generate_manifests
            ;;
        iso)
            build_iso
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            log_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
    
    echo ""
    log_success "=== Build Complete ==="
    echo "Image: ${IMAGE_NAME}:${VERSION}"
    echo "Platforms: ${PLATFORMS}"
    echo ""
    echo "Components:"
    echo "  - Bacalhau compute node"
    echo "  - PicoClaw AI agent"
    echo "  - Docker runtime"
    echo ""
    echo "Ports:"
    echo "  - 4222: NATS messaging"
    echo "  - 9090: Metrics endpoint"
    echo "  - 18790: PicoClaw gateway"
}

# Run main function
main "$@"
