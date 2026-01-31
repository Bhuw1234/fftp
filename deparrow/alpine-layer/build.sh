#!/bin/bash
set -e

echo "=== Building DEparrow Alpine Linux Node Image ==="

# Configuration
IMAGE_NAME="deparrow/alpine-node"
VERSION="1.0.0"
PLATFORMS="linux/amd64,linux/arm64"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Error: Docker is not installed${NC}"
        exit 1
    fi
    
    # Check Docker buildx
    if ! docker buildx version &> /dev/null; then
        echo -e "${YELLOW}Warning: Docker Buildx not available, installing...${NC}"
        docker buildx install
    fi
    
    # Create builder instance if needed
    if ! docker buildx ls | grep -q "deparrow-builder"; then
        echo "Creating Docker Buildx builder..."
        docker buildx create --name deparrow-builder --use
    fi
    
    echo -e "${GREEN}✓ Prerequisites satisfied${NC}"
}

# Build multi-architecture image
build_image() {
    echo "Building multi-architecture image..."
    
    # Create build context
    mkdir -p build-context
    cp Dockerfile build-context/
    cp -r ../config build-context/
    cp -r ../scripts build-context/
    
    # Build for multiple platforms
    docker buildx build \
        --platform ${PLATFORMS} \
        -t ${IMAGE_NAME}:${VERSION} \
        -t ${IMAGE_NAME}:latest \
        --push \
        build-context/
    
    # Cleanup
    rm -rf build-context
    
    echo -e "${GREEN}✓ Image built and pushed: ${IMAGE_NAME}:${VERSION}${NC}"
    echo -e "${GREEN}✓ Platforms: ${PLATFORMS}${NC}"
}

# Generate deployment manifests
generate_manifests() {
    echo "Generating deployment manifests..."
    
    # Kubernetes deployment
    cat > ../config/kubernetes/deployment.yaml << EOF
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
        env:
        - name: NODE_ARCH
          value: "\$(NODE_ARCH)"
        - name: DEPARROW_API_KEY
          valueFrom:
            secretKeyRef:
              name: deparrow-secrets
              key: api-key
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
  type: ClusterIP
EOF
    
    # Docker Compose configuration
    cat > ../config/docker-compose/deparrow-node.yml << EOF
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
    environment:
      - NODE_ARCH=amd64
      - NODE_CPU=4
      - NODE_MEMORY=4GB
      - NODE_DISK=20GB
      - DEPARROW_API_KEY=\${DEPARROW_API_KEY}
      - DEPARROW_BOOTSTRAP=https://bootstrap.deparrow.net
      - DEPARROW_ORCHESTRATOR_HOST=orchestrator.deparrow.net
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
    cat > ../config/systemd/deparrow-node.service << EOF
[Unit]
Description=DEparrow Compute Node
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=deparrow
Group=deparrow
Environment="NODE_ARCH=\$(uname -m)"
Environment="DEPARROW_API_KEY=\${DEPARROW_API_KEY}"
Environment="DEPARROW_BOOTSTRAP=https://bootstrap.deparrow.net"
Environment="DEPARROW_ORCHESTRATOR_HOST=orchestrator.deparrow.net"
ExecStartPre=/usr/bin/docker pull ${IMAGE_NAME}:${VERSION}
ExecStart=/usr/bin/docker run --rm \
  --name deparrow-node \
  --privileged \
  -p 4222:4222 \
  -p 9090:9090 \
  -e NODE_ARCH \
  -e DEPARROW_API_KEY \
  -e DEPARROW_BOOTSTRAP \
  -e DEPARROW_ORCHESTRATOR_HOST \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /var/lib/deparrow:/var/lib/deparrow \
  -v /etc/deparrow:/etc/deparrow \
  ${IMAGE_NAME}:${VERSION}
ExecStop=/usr/bin/docker stop deparrow-node
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    
    echo -e "${GREEN}✓ Deployment manifests generated${NC}"
}

# Main execution
main() {
    echo "DEparrow Alpine Node Builder"
    echo "============================"
    
    check_prerequisites
    build_image
    generate_manifests
    
    echo ""
    echo -e "${GREEN}=== Build Complete ==="
    echo "Image: ${IMAGE_NAME}:${VERSION}"
    echo "Platforms: ${PLATFORMS}"
    echo "Manifests generated in:"
    echo "  - ../config/kubernetes/deployment.yaml"
    echo "  - ../config/docker-compose/deparrow-node.yml"
    echo "  - ../config/systemd/deparrow-node.service${NC}"
}

# Run main function
main "$@"