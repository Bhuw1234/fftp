# DEparrow - Decentralized AI Operating System

<p align="center">
  <strong>🌐 The Operating System for Decentralized AI Compute</strong>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#features">Features</a> •
  <a href="#deployment">Deployment</a> •
  <a href="#clawdbot">Clawdbot CLI</a>
</p>

---

## What is DEparrow?

DEparrow is a **decentralized AI operating system** that turns distributed compute resources into a unified AI compute network. Built on top of [Bacalhau](https://bacalhau.org), DEparrow enables:

- **Anyone to contribute compute** and earn credits
- **Anyone to use distributed AI** by spending credits
- **Natural language interaction** via Clawdbot terminal

## Quick Start

### Option 1: Development Mode

```bash
cd deparrow
./start.sh dev
```

This starts:
- 🌐 Meta-OS API at http://localhost:8080
- 🎨 GUI at http://localhost:5173

### Option 2: Production (Docker Compose)

```bash
cd deparrow
./start.sh prod
```

This starts the full stack:
- 🌐 Meta-OS API at http://localhost:8080
- 🎨 GUI at http://localhost:3000
- 📊 Prometheus at http://localhost:9090
- 📈 Grafana at http://localhost:3001

### Option 3: Kubernetes

```bash
kubectl apply -k deparrow/k8s/base
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      User Interfaces                         │
├────────────────┬────────────────┬────────────────────────────┤
│   🦞 Clawdbot  │   🌐 Web GUI   │   🐍 Python SDK            │
│    Terminal    │   Dashboard    │      API                   │
├────────────────┴────────────────┴────────────────────────────┤
│                    Meta-OS Control Plane                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │Bootstrap │ │ Credit   │ │   Job    │ │   JWT    │        │
│  │ Server   │ │ System   │ │Admission │ │   Auth   │        │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
├─────────────────────────────────────────────────────────────┤
│                    Alpine Linux Nodes                        │
│            Auto-join • Health Check • Resource Report        │
├─────────────────────────────────────────────────────────────┤
│                    Bacalhau Compute Network                  │
│          Docker • WebAssembly • NATS • libp2p • IPFS        │
└─────────────────────────────────────────────────────────────┘
```

## Features

### 🌐 Decentralized Compute
- No single point of failure
- Global node network
- Automatic job distribution

### 💰 Credit Economy
- Earn credits by contributing compute
- Spend credits to run AI jobs
- Fair market-based pricing

### 🦞 Clawdbot Terminal
- Natural language interaction
- "Train my model on the network"
- "Check my credit balance"
- Multi-channel: Terminal, WhatsApp, Telegram, Slack

### 🔐 Security
- JWT authentication
- Sandboxed execution (Docker/WASM)
- Job admission control

## Clawdbot CLI

DEparrow is built into Clawdbot for easy terminal access:

```bash
# Install Clawdbot
npm install -g clawdbot

# Check network status
clawdbot deparrow status

# Check your credits
clawdbot deparrow credits

# Submit a job
clawdbot deparrow submit -t docker -i python:3.11 -c "python train.py"

# List your jobs
clawdbot deparrow jobs

# Or use natural language
clawdbot agent --message "Train my model on DEparrow"
```

## Directory Structure

```
deparrow/
├── alpine-layer/        # Node OS (Dockerfile, build scripts)
├── bacalhau-layer/      # Bacalhau configurations
├── bootable/            # ISO/bootable image creation
├── gui-layer/           # React/Vite web dashboard
├── metaos-layer/        # Python control plane (Flask API)
├── k8s/                 # Kubernetes manifests
├── config/              # Prometheus, Grafana configs
├── test-integration/    # E2E tests
├── docker-compose.prod.yml
├── start.sh             # Quick start script
└── README.md
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/health` | GET | Health check |
| `/api/v1/auth/login` | POST | Get JWT token |
| `/api/v1/nodes` | GET | List nodes |
| `/api/v1/credits` | GET | Get credit balance |
| `/api/v1/jobs` | GET | List jobs |
| `/api/v1/jobs` | POST | Submit job |
| `/api/v1/jobs/:id` | GET | Get job details |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DEPARROW_SECRET_KEY` | (required) | JWT signing key |
| `DEPARROW_API_URL` | `http://localhost:8080` | Meta-OS API URL |
| `DATABASE_URL` | - | PostgreSQL connection |
| `REDIS_URL` | - | Redis connection |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `./test-integration.sh`
5. Submit a pull request

## License

Apache 2.0 - See [LICENSE](../LICENSE)

---

<p align="center">
  Built with ❤️ for decentralized AI
</p>
