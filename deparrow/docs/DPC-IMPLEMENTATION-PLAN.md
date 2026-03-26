# DPC Token Phase 3 - Implementation Plan

**Created:** 2026-03-26
**Status:** Planning Complete, Ready for Implementation

---

## Overview

This document outlines the step-by-step implementation plan for DEparrow Coin (DPC) blockchain development.

---

## Requirements Summary

| Requirement | Version/Details |
|-------------|-----------------|
| **Go Version** | 1.21+ (compatible with Bacalhau Go 1.24.0) |
| **Cosmos SDK** | v0.50.x (latest stable) |
| **Ignite CLI** | v0.28+ for scaffolding |
| **CometBFT** | v0.38+ (consensus engine) |
| **Block Time** | 6 seconds default |
| **TPS** | ~10,000 |

---

## Custom Cosmos SDK Modules

| Module | Purpose |
|--------|---------|
| `x/proofofcompute` | Core PoC consensus, job verification, reward calculation |
| `x/computemarket` | Job submission, provider matching, payment escrow |
| `x/agentwallet` | AI agent wallets, autonomous transaction rules |

---

## Step-by-Step Implementation

### Step 1: Cosmos SDK Project Scaffold (1 week)

**Files to Create:**
```
deparrow/chain/
├── app/
│   ├── app.go                 # Main application wiring
│   ├── encoding.go            # Protobuf encoding config
│   ├── export.go              # State export/import
│   └── genesis.go             # Genesis state handlers
├── cmd/dpcd/
│   ├── main.go                # Daemon entry point
│   └── root.go                # CLI commands
├── x/                         # Custom modules (empty initially)
├── proto/
│   └── dpc/                   # Protobuf definitions
├── testutil/
├── Makefile
├── go.mod
└── README.md
```

**Commands:**
```bash
# Install Ignite CLI
curl https://get.ignite.com/cli! | bash

# Scaffold new chain
cd deparrow
ignite scaffold chain dpc --module github.com/deparrow/dpc
```

---

### Step 2: DPC Token Module (1 week)

**Files to Create:**
```
deparrow/chain/x/token/
├── keeper/
│   ├── keeper.go              # Token keeper
│   ├── mint.go                # Minting logic (PoC only)
│   └── query.go               # Balance queries
├── types/
│   ├── coin.go                # DPC coin definition
│   ├── errors.go              # Module errors
│   ├── events.go              # Event definitions
│   └── genesis.go             # Genesis state
├── module.go
└── handler.go
```

**Token Parameters:**
```go
const (
    Denom         = "dpc"
    Decimals      = 18
    MaxSupply     = 21_000_000_000 * 10^18
    InitialSupply = 1_000_000_000 * 10^18
)
```

---

### Step 3: Proof-of-Compute Module (4 weeks)

**Files to Create:**
```
deparrow/chain/x/proofofcompute/
├── keeper/
│   ├── keeper.go              # Main keeper
│   ├── job.go                 # Job state management
│   ├── reward.go              # Reward calculation
│   ├── verification.go        # Compute proof verification
│   ├── difficulty.go          # Difficulty adjustment
│   └── query.go               # gRPC queries
├── types/
│   ├── job.go                 # Job struct (maps from Bacalhau Job)
│   ├── compute_proof.go       # Proof structure
│   ├── params.go              # Module parameters
│   ├── errors.go
│   ├── events.go
│   ├── keys.go
│   ├── genesis.go
│   └── codec.go               # Protobuf codec
├── module.go
├── handler.go
└── abci.go                    # EndBlocker for reward distribution
```

**Key Types:**
```go
type Job struct {
    ID              string
    Submitter       string    // Cosmos address
    ComputeNode     string    // Cosmos address
    Spec            []byte    // Job specification
    Result          []byte    // IPFS CID of results
    Stake           sdk.Coin
    Reward          sdk.Coin
    Status          JobStatus
    ComputeUnits    uint64    // CPU-hours, GPU-hours
    CreatedAt       int64
    CompletedAt     int64
}

type ComputeProof struct {
    JobID           string
    NodeID          string
    ComputeUnits    uint64
    ExecutionTime   int64
    OutputHash      []byte    // Deterministic hash
    Signature       []byte    // Node signature
}
```

---

### Step 4: Compute Marketplace Module (3 weeks)

**Files to Create:**
```
deparrow/chain/x/computemarket/
├── keeper/
│   ├── keeper.go
│   ├── escrow.go              # Job stake escrow
│   ├── provider.go            # Provider registry
│   ├── matching.go            # Job-provider matching
│   └── dispute.go             # Dispute resolution
├── types/
│   ├── escrow.go              # Escrow contract types
│   ├── provider.go            # Provider registration
│   ├── matching.go            # Matching algorithms
│   ├── params.go
│   ├── errors.go
│   ├── events.go
│   └── genesis.go
├── module.go
└── handler.go
```

**Key Types:**
```go
type Provider struct {
    Address         string
    StakedAmount    sdk.Coin
    ReputationScore uint32    // 0-1000
    Capabilities    []byte    // CPU, GPU, memory specs
    CompletedJobs   uint64
    FailedJobs      uint64
    Active          bool
}

type Escrow struct {
    JobID           string
    Submitter       string
    Provider        string
    Amount          sdk.Coin
    Status          EscrowStatus
    Deadline        int64
    CreatedAt       int64
}
```

---

### Step 5: AI Agent Wallet Module (2 weeks)

**Files to Create:**
```
deparrow/chain/x/agentwallet/
├── keeper/
│   ├── keeper.go
│   ├── wallet.go              # Wallet CRUD
│   ├── rules.go               # Spending rules
│   ├── automation.go          # Automation triggers
│   └── query.go
├── types/
│   ├── wallet.go              # Agent wallet struct
│   ├── rules.go               # Spending rules
│   ├── automation.go          # Automation config
│   ├── params.go
│   ├── errors.go
│   ├── events.go
│   └── genesis.go
├── module.go
└── handler.go
```

**Key Types:**
```go
type AgentWallet struct {
    DID             string    // did:deparrow:agent:<uuid>
    Address         string    // Cosmos address
    Balance         sdk.Coin
    SpendingRules   []SpendingRule
    AutomationRules []AutomationRule
    EmergencyReserve sdk.Coin
    CreatedAt       int64
}

type SpendingRule struct {
    MaxPerTx        sdk.Coin
    DailyBudget     sdk.Coin
    AllowedOps      []string  // ["submit_job", "pay_service"]
    BlockedOps      []string  // ["external_transfer"]
}
```

---

### Step 6: Bacalhau Integration Layer (2 weeks)

**Files to Create:**
```
deparrow/chain/integration/
├── bacalhau/
│   ├── client.go              # DPC chain client for Bacalhau
│   ├── job_adapter.go         # Convert Bacalhau Job to DPC Job
│   ├── proof_generator.go     # Generate compute proofs
│   ├── reward_reporter.go     # Report completed jobs
│   └── config.go
├── bridge/
│   ├── event_listener.go      # Listen to Bacalhau events
│   └── tx_submitter.go        # Submit to DPC chain
└── metrics/
    └── integration_metrics.go
```

**Files to Modify:**
| File | Changes |
|------|---------|
| `pkg/compute/executor.go` | Add proof generation after job completion |
| `pkg/jobstore/types.go` | Add `ComputeUnits` field to Job struct |
| `pkg/orchestrator/scheduler_provider.go` | Report compute units to DPC |
| `picoclaw/pkg/deparrow/types.go` | Add DPC wallet types |

---

### Step 7: gRPC/REST API & SDK (2 weeks)

**Files to Create:**
```
deparrow/chain/api/
├── grpc/
│   ├── query.go               # gRPC query service
│   └── tx.go                  # gRPC tx service
├── rest/
│   └── swagger.yaml           # OpenAPI spec
deparrow/sdk/
├── go/
│   ├── client.go              # Go SDK client
│   ├── wallet.go              # Wallet operations
│   ├── job.go                 # Job operations
│   └── query.go               # Query helpers
├── python/
│   ├── client.py
│   └── wallet.py
└── typescript/
    ├── client.ts
    └── wallet.ts
```

---

### Step 8: Testnet Deployment (1 week)

**Files to Create:**
```
deparrow/chain/testnet/
├── genesis.json               # Genesis configuration
├── gentx/                     # Initial validator transactions
├── config/
│   ├── config.toml
│   └── app.toml
├── scripts/
│   ├── init.sh                # Initialize testnet
│   ├── add-validator.sh
│   └── start.sh
└── monitoring/
    ├── prometheus.yml
    └── grafana-dashboard.json
```

---

### Step 9: Security Audit Preparation (4 weeks)

**Files to Create:**
```
deparrow/chain/audit/
├── SPECIFICATION.md           # Module specifications
├── THREAT_MODEL.md            # Threat model document
├── TEST_PLAN.md               # Audit test plan
└── findings/                  # Placeholder for audit findings
```

---

## Integration Points with Bacalhau

| Bacalhau Component | DPC Integration |
|-------------------|-----------------|
| `pkg/jobstore/types.go` | Job completion triggers DPC reward |
| `pkg/orchestrator/interfaces.go` | Scheduler reports compute units |
| `pkg/compute/executor.go` | Generate compute proofs |
| `picoclaw/pkg/deparrow/types.go` | Extend with DPC wallet types |

---

## Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **PoC manipulation** | Medium | High | Deterministic verification, spot-checking, staking |
| **Integration complexity** | High | Medium | Phased rollout, extensive integration tests |
| **Security vulnerabilities** | Medium | Critical | External audit, bug bounty program |
| **Performance degradation** | Low | Medium | Load testing, optimize hot paths |
| **Regulatory uncertainty** | Low | High | Legal review, utility-only design |

---

## Success Criteria

1. **Technical Metrics:**
   - DPC chain produces blocks every 6 seconds
   - 1,000+ TPS throughput
   - Job submission to reward in < 30 seconds
   - Zero critical security findings in audit

2. **Integration Metrics:**
   - 100% of completed Bacalhau jobs generate DPC proofs
   - Proof verification success rate > 99.9%
   - Reward distribution accuracy 100%

3. **Operational Metrics:**
   - Testnet uptime > 99.9%
   - 4+ active validators
   - Monitoring dashboards operational

---

## Estimated Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| Step 1: Scaffold | 1 week | Working chain skeleton |
| Step 2: Token Module | 1 week | DPC token functional |
| Step 3: PoC Module | 4 weeks | Core consensus working |
| Step 4: Marketplace Module | 3 weeks | Escrow and matching |
| Step 5: Agent Wallet | 2 weeks | Autonomous wallets |
| Step 6: Integration | 2 weeks | Bacalhau connected |
| Step 7: APIs & SDKs | 2 weeks | Developer tools |
| Step 8: Testnet | 1 week | Public testnet |
| Step 9: Audit Prep | 4 weeks | Audit-ready code |

**Total Estimated Time: 16-20 weeks (4-5 months)**

---

## Next Actions

1. Install Ignite CLI: `curl https://get.ignite.com/cli! | bash`
2. Create `deparrow/chain/` directory structure
3. Scaffold Cosmos SDK chain
4. Begin Step 1 implementation

---

*Document created: 2026-03-26*
