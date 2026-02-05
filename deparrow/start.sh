#!/bin/bash

# DEparrow Quick Start Script
# Usage: ./start.sh [dev|prod]

set -e

MODE=${1:-dev}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Starting DEparrow - Decentralized AI Operating System"
echo "========================================================="
echo ""

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Generate secret key if not set
if [ -z "$DEPARROW_SECRET_KEY" ]; then
    export DEPARROW_SECRET_KEY=$(openssl rand -hex 32)
    echo "🔑 Generated random secret key"
fi

case $MODE in
    dev)
        echo "📦 Starting in DEVELOPMENT mode..."
        echo ""
        
        # Start Meta-OS
        echo "🌐 Starting Meta-OS Control Plane..."
        cd "$SCRIPT_DIR/metaos-layer"
        if [ ! -d "venv" ]; then
            python3 -m venv venv
            source venv/bin/activate
            pip install flask flask-cors flask-jwt-extended
        else
            source venv/bin/activate
        fi
        python3 bootstrap-server.py &
        METAOS_PID=$!
        echo "   ✅ Meta-OS started (PID: $METAOS_PID)"
        
        # Start GUI
        echo "🎨 Starting GUI..."
        cd "$SCRIPT_DIR/gui-layer"
        if [ ! -d "node_modules" ]; then
            npm install
        fi
        npm run dev &
        GUI_PID=$!
        echo "   ✅ GUI started (PID: $GUI_PID)"
        
        echo ""
        echo "========================================================="
        echo "✅ DEparrow is running!"
        echo ""
        echo "   🌐 Meta-OS API:  http://localhost:8080"
        echo "   🎨 GUI:          http://localhost:5173"
        echo ""
        echo "Press Ctrl+C to stop"
        echo "========================================================="
        
        # Wait for processes
        trap "kill $METAOS_PID $GUI_PID 2>/dev/null" EXIT
        wait
        ;;
        
    prod)
        echo "📦 Starting in PRODUCTION mode..."
        echo ""
        
        cd "$SCRIPT_DIR"
        docker compose -f docker-compose.prod.yml up -d
        
        echo ""
        echo "========================================================="
        echo "✅ DEparrow is running in production!"
        echo ""
        echo "   🌐 Meta-OS API:  http://localhost:8080"
        echo "   🎨 GUI:          http://localhost:3000"
        echo "   📊 Prometheus:   http://localhost:9090"
        echo "   📈 Grafana:      http://localhost:3001 (admin/admin)"
        echo ""
        echo "To stop: docker compose -f docker-compose.prod.yml down"
        echo "========================================================="
        ;;
        
    *)
        echo "Usage: $0 [dev|prod]"
        echo ""
        echo "  dev  - Start in development mode (local Python + npm)"
        echo "  prod - Start in production mode (Docker Compose)"
        exit 1
        ;;
esac
