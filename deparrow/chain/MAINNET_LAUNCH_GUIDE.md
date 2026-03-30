# DPC Mainnet Launch Guide

This document provides step-by-step instructions for launching the DPC mainnet.

---

## Prerequisites

- Go 1.24+
- DPC binary (dpcd) compiled from source
- Minimum 4 CPU cores, 8GB RAM, 500GB SSD
- Static public IP address
- Open ports: 26656 (P2P), 26657 (RPC), 1317 (REST API), 9090 (gRPC)

---

## Phase 1: Genesis File Preparation

### 1.1 Create Genesis Template

```json
{
  "genesis_time": "2026-04-15T00:00:00Z",
  "chain_id": "dpc-mainnet-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    },
    "version": {}
  },
  "app_hash": "",
  "app_state": {
    "auth": {
      "params": {
        "max_memo_characters": "256",
        "tx_sig_limit": "7",
        "tx_size_cost_per_byte": "10",
        "sig_verify_cost_ed25519": "590",
        "sig_verify_cost_secp256k1": "1000"
      },
      "accounts": []
    },
    "bank": {
      "params": {
        "send_enabled": [],
        "default_send_enabled": true
      },
      "balances": [],
      "supply": [],
      "denom_metadata": []
    },
    "staking": {
      "params": {
        "unbonding_time": "1814400s",
        "max_validators": 100,
        "max_entries": 7,
        "historical_entries": 10000,
        "bond_denom": "dpc"
      },
      "last_total_power": "0",
      "last_validator_powers": [],
      "validators": [],
      "delegations": [],
      "unbonding_delegations": [],
      "redelegations": [],
      "exported": false
    },
    "distribution": {
      "params": {
        "community_tax": "0.020000000000000000",
        "base_proposer_reward": "0.010000000000000000",
        "bonus_proposer_reward": "0.040000000000000000",
        "withdraw_addr_enabled": true
      },
      "fee_pool": {
        "community_pool": []
      },
      "delegator_withdraw_infos": [],
      "previous_proposer": "",
      "outstanding_rewards": [],
      "outstanding_validators": [],
      "all_validators": [],
      "delegator_starting_infos": [],
      "validator_accumulated_commissions": [],
      "validator_historical_rewards": [],
      "validator_current_rewards": [],
      "delegator_withdraw_address": []
    },
    "proofofcompute": {
      "params": {
        "min_compute_units": "1",
        "reward_per_unit": "1000000000000000",
        "max_supply": "21000000000000000000000000000",
        "complexity_multiplier": 5,
        "min_stake": "1000000000000000000",
        "difficulty_adjustment": 100
      },
      "jobs": [],
      "proofs": [],
      "pending_rewards": [],
      "total_supply": "1000000000000000000000000000",
      "difficulty": 1
    },
    "computemarket": {
      "params": {
        "min_provider_stake": "1000000000000000000000",
        "escrow_period": "86400s",
        "max_dispute_period": "604800s",
        "provider_slash_percent": "10"
      },
      "providers": [],
      "escrows": [],
      "reputations": []
    },
    "agentwallet": {
      "params": {
        "max_automation_triggers": 10,
        "max_spending_rules": 20
      },
      "identities": [],
      "spending_rules": [],
      "automation_triggers": []
    }
  }
}
```

### 1.2 Token Distribution

| Allocation | Amount | Percentage | Recipient |
|------------|--------|------------|-----------|
| Compute Mining Rewards | 14,700,000,000 DPC | 70% | Mined via PoC |
| Development Fund | 3,150,000,000 DPC | 15% | Treasury multisig |
| Ecosystem Fund | 2,100,000,000 DPC | 10% | DAO treasury |
| Initial Sale | 1,050,000,000 DPC | 5% | Early supporters |

### 1.3 Add Genesis Accounts

```bash
# Add genesis account
dpcd add-genesis-account <address> 1000000000000000000000dpc

# Add development fund
dpcd add-genesis-account dpc1devfund... 3150000000000000000000000000dpc

# Add ecosystem fund
dpcd add-genesis-account dpc1ecofund... 2100000000000000000000000000dpc
```

---

## Phase 2: Validator Setup

### 2.1 Initialize Node

```bash
# Initialize node
dpcd init <moniker> --chain-id dpc-mainnet-1

# Download genesis
wget https://genesis.deparrow.io/dpc-mainnet-1/genesis.json -O ~/.dpc/config/genesis.json

# Verify genesis
dpcd validate-genesis
```

### 2.2 Create Validator Key

```bash
# Generate new key
dpcd keys add validator --keyring-backend file

# Show address
dpcd keys show validator -a --keyring-backend file
```

### 2.3 Create Validator (Genesis)

```bash
# Create gentx
dpcd gentx validator 1000000000000000000000dpc \
  --moniker "<your-moniker>" \
  --commission-rate "0.05" \
  --commission-max-rate "0.20" \
  --commission-max-change-rate "0.01" \
  --min-self-delegation "1" \
  --keyring-backend file

# Submit gentx to https://github.com/deparrow/networks
```

### 2.4 Validator Configuration

Edit `~/.dpc/config/config.toml`:

```toml
# Peer connections
persistent_peers = "id1@peer1.deparrow.io:26656,id2@peer2.deparrow.io:26656"

# Seed nodes
seeds = "seed1.deparrow.io:26656,seed2.deparrow.io:26656"

# Performance
max_open_connections = 200
max_num_inbound_peers = 50
max_num_outbound_peers = 50

# Security
addr_book_strict = true
```

---

## Phase 3: Node Security

### 3.1 Firewall Setup

```bash
# Allow required ports
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 26656/tcp   # P2P
sudo ufw allow 26657/tcp   # RPC (optional, for sentry nodes)
sudo ufw enable
```

### 3.2 Key Management

```bash
# Backup validator key
cp ~/.dpc/config/priv_validator_key.json /secure/backup/location/

# Use HSM for production (recommended)
```

### 3.3 Sentry Node Architecture

```
Internet → Sentry Nodes (Public) → Validator (Private)
```

- Run validator behind sentry nodes
- Don't expose validator RPC to public
- Use VPN or private network between sentry and validator

---

## Phase 4: Launch Sequence

### 4.1 Pre-Launch Checklist

- [ ] Genesis file validated
- [ ] All gentx collected
- [ ] Persistent peers configured
- [ ] Firewall configured
- [ ] Monitoring set up
- [ ] Backup procedures in place

### 4.2 Genesis Launch

```bash
# Collect gentx
dpcd collect-gentxs

# Validate final genesis
dpcd validate-genesis

# Start node
dpcd start

# Monitor logs
journalctl -u dpcd -f
```

### 4.3 Post-Launch Monitoring

```bash
# Check node status
dpcd status

# Check peers
dpcd query tendermint-validator-set

# Check block height
curl http://localhost:26657/status | jq .result.sync_info.latest_block_height
```

---

## Phase 5: Governance

### 5.1 Initial Governance Parameters

| Parameter | Value |
|-----------|-------|
| Min Deposit | 1000 DPC |
| Voting Period | 14 days |
| Quorum | 33.4% |
| Threshold | 50% |
| Veto Threshold | 33.4% |

### 5.2 Create Proposal

```bash
dpcd tx gov submit-proposal \
  --title "Proposal Title" \
  --description "Description" \
  --type "Text" \
  --deposit 1000000000000000000000dpc \
  --from validator \
  --keyring-backend file
```

---

## Resources

- **Genesis File:** https://genesis.deparrow.io/dpc-mainnet-1/genesis.json
- **Peer List:** https://peers.deparrow.io/dpc-mainnet-1
- **Block Explorer:** https://explorer.deparrow.io
- **Documentation:** https://docs.deparrow.io

---

## Support

- Discord: https://discord.gg/deparrow
- Telegram: https://t.me/deparrow
- GitHub Issues: https://github.com/Bhuw1234/fftp/issues
