#!/bin/bash
#
# DPC Validator Setup Script
# 
# Usage: ./setup_validator.sh <moniker> <chain-id>
#
# Example: ./setup_validator.sh my-validator dpc-testnet-1
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DPCD_BIN="${DPCD_BIN:-./build/dpcd}"
DPC_HOME="${DPC_HOME:-$HOME/.dpc}"
KEYRING_BACKEND="${KEYRING_BACKEND:-file}"

# Parse arguments
MONIKER="${1:-my-validator}"
CHAIN_ID="${2:-dpc-testnet-1}"

echo -e "${GREEN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║           DPC VALIDATOR SETUP SCRIPT                       ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Moniker: $MONIKER"
echo "║  Chain ID: $CHAIN_ID"
echo "║  Home: $DPC_HOME"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if dpcd exists
if [ ! -f "$DPCD_BIN" ]; then
    echo -e "${RED}Error: dpcd binary not found at $DPCD_BIN${NC}"
    echo "Please build it first: go build -o build/dpcd ./cmd/dpcd"
    exit 1
fi

# Step 1: Initialize node
echo -e "${YELLOW}Step 1: Initializing node...${NC}"
$DPCD_BIN init "$MONIKER" --chain-id "$CHAIN_ID" --home "$DPC_HOME"
echo -e "${GREEN}✓ Node initialized${NC}"

# Step 2: Create validator key
echo -e "${YELLOW}Step 2: Creating validator key...${NC}"
echo "Enter a password for your validator key:"
$DPCD_BIN keys add validator --keyring-backend "$KEYRING_BACKEND" --home "$DPC_HOME"
VALIDATOR_ADDR=$($DPCD_BIN keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home "$DPC_HOME")
echo -e "${GREEN}✓ Validator key created${NC}"
echo -e "  Address: ${YELLOW}$VALIDATOR_ADDR${NC}"

# Step 3: Configure node
echo -e "${YELLOW}Step 3: Configuring node...${NC}"

# Update config.toml
CONFIG_FILE="$DPC_HOME/config/config.toml"
if [ -f "$CONFIG_FILE" ]; then
    # Set moniker
    sed -i "s/^moniker = .*/moniker = \"$MONIKER\"/" "$CONFIG_FILE"
    
    # Enable prometheus
    sed -i 's/^prometheus = false/prometheus = true/' "$CONFIG_FILE"
    
    # Configure peers (testnet)
    if [ "$CHAIN_ID" = "dpc-testnet-1" ]; then
        sed -i 's/^persistent_peers = ""/persistent_peers = "dpc@34.180.51.11:26656"/' "$CONFIG_FILE"
    fi
    
    echo -e "${GREEN}✓ Node configured${NC}"
else
    echo -e "${RED}Warning: config.toml not found${NC}"
fi

# Update app.toml
APP_FILE="$DPC_HOME/config/app.toml"
if [ -f "$APP_FILE" ]; then
    # Enable API
    sed -i 's/^enable = false/enable = true/' "$APP_FILE"
    
    # Enable CORS for API
    sed -i 's/^enabled-unsafe-cors = false/enabled-unsafe-cors = true/' "$APP_FILE"
    
    echo -e "${GREEN}✓ App configured${NC}"
fi

# Step 4: Create gentx (for genesis validators)
echo -e "${YELLOW}Step 4: Creating gentx (for genesis validators)...${NC}"
echo "Enter stake amount (default: 1000000000000000000000dpc = 1000 DPC):"
read -r STAKE_AMOUNT
STAKE_AMOUNT="${STAKE_AMOUNT:-1000000000000000000000dpc}"

echo "Enter commission rate (default: 0.05 = 5%):"
read -r COMMISSION_RATE
COMMISSION_RATE="${COMMISSION_RATE:-0.05}"

echo "Enter commission max rate (default: 0.20 = 20%):"
read -r COMMISSION_MAX_RATE
COMMISSION_MAX_RATE="${COMMISSION_MAX_RATE:-0.20}"

echo "Enter commission max change rate (default: 0.01 = 1%):"
read -r COMMISSION_MAX_CHANGE
COMMISSION_MAX_CHANGE="${COMMISSION_MAX_CHANGE:-0.01}"

# Add genesis account (for local test)
$DPCD_BIN add-genesis-account "$VALIDATOR_ADDR" "$STAKE_AMOUNT" --home "$DPC_HOME" 2>/dev/null || true

# Create gentx
$DPCD_BIN gentx validator "$STAKE_AMOUNT" \
    --moniker "$MONIKER" \
    --commission-rate "$COMMISSION_RATE" \
    --commission-max-rate "$COMMISSION_MAX_RATE" \
    --commission-max-change-rate "$COMMISSION_MAX_CHANGE" \
    --keyring-backend "$KEYRING_BACKEND" \
    --home "$DPC_HOME" \
    --chain-id "$CHAIN_ID"

echo -e "${GREEN}✓ Gentx created${NC}"

# Show gentx location
GENTX_FILE=$(ls -t "$DPC_HOME/config/gentx/"*.json 2>/dev/null | head -1)
if [ -n "$GENTX_FILE" ]; then
    echo -e "  Gentx file: ${YELLOW}$GENTX_FILE${NC}"
    echo -e "  Submit this file to join as genesis validator"
fi

# Step 5: Summary
echo -e "${GREEN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              SETUP COMPLETE                                ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Validator Address: $VALIDATOR_ADDR"
echo "║  Node Home: $DPC_HOME"
echo "║  Chain ID: $CHAIN_ID"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Next Steps:                                               ║"
echo "║  1. Submit your gentx to join genesis                      ║"
echo "║  2. Download final genesis.json                            ║"
echo "║  3. Start node: $DPCD_BIN start --home $DPC_HOME"
echo "║  4. Monitor: curl localhost:26657/status                   ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Create systemd service file
SYSTEMD_FILE="/tmp/dpcd.service"
cat > "$SYSTEMD_FILE" << EOF
[Unit]
Description=DPC Node
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=$(realpath "$DPCD_BIN") start --home "$DPC_HOME"
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

echo -e "${YELLOW}Systemd service file created: $SYSTEMD_FILE${NC}"
echo "Install with: sudo cp $SYSTEMD_FILE /etc/systemd/system/ && sudo systemctl enable dpcd"
