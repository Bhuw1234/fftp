package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "proofofcompute"

	// StoreKey is the store key string for proofofcompute
	StoreKey = ModuleName

	// RouterKey is the message route for proofofcompute
	RouterKey = ModuleName

	// QuerierRoute is the querier route for proofofcompute
	QuerierRoute = ModuleName
)

// Job status constants
const (
	JobStatusPending   JobStatus = 0
	JobStatusRunning   JobStatus = 1
	JobStatusCompleted JobStatus = 2
	JobStatusFailed    JobStatus = 3
	JobStatusCancelled JobStatus = 4
)

// JobStatus represents the status of a job
type JobStatus int32

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusRunning:
		return "running"
	case JobStatusCompleted:
		return "completed"
	case JobStatusFailed:
		return "failed"
	case JobStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Job represents a compute job in the DEparrow network
type Job struct {
	ID           string
	Submitter    string    // Cosmos address
	ComputeNode  string    // Cosmos address
	Spec         []byte    // Job specification
	Result       []byte    // IPFS CID of results
	Stake        sdk.Coin
	Reward       sdk.Coin
	Status       JobStatus
	ComputeUnits uint64
	CreatedAt    int64
	CompletedAt  int64
	Complexity   uint32    // Complexity multiplier (1-5)
}

// ComputeProof represents proof of computation
type ComputeProof struct {
	JobID         string
	NodeID        string
	ComputeUnits  uint64
	ExecutionTime int64
	OutputHash    []byte
	Signature     []byte
}

// NewJob creates a new Job instance
func NewJob(id, submitter string) *Job {
	return &Job{
		ID:        id,
		Submitter: submitter,
		Status:    JobStatusPending,
		CreatedAt: 0, // Set by keeper
	}
}

// Validate performs basic validation of the job
func (j Job) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}
	if j.Submitter == "" {
		return fmt.Errorf("submitter cannot be empty")
	}
	if j.Stake.IsNegative() {
		return fmt.Errorf("stake cannot be negative")
	}
	return nil
}

// Params defines the parameters for the proofofcompute module.
type Params struct {
	MinComputeUnits      uint64
	RewardPerUnit        string   // DPC reward per compute unit (as string for precision)
	DifficultyAdjustment uint64   // Difficulty adjustment period (blocks)
	TargetBlockTime      uint32   // Target block time in seconds
	MaxSupply            string   // Maximum DPC supply (21B with 18 decimals)
	ComplexityMultiplier uint32   // Max complexity multiplier (default 5)
	MinStake             string   // Minimum stake required
}

// RewardParams holds parameters for reward calculation
type RewardParams struct {
	BaseRate           sdk.Dec // Base reward rate (0.001 DPC)
	ComplexityMultiplier uint32  // 1-5x multiplier
	ComputeUnits       uint64  // Actual compute units consumed
}

// DifficultyParams holds parameters for difficulty adjustment
type DifficultyParams struct {
	CurrentDifficulty  uint64  // Current difficulty level
	TargetJobsPerBlock float64 // Target jobs per block
	ActualJobsPerBlock float64 // Actual jobs per block
	AdjustmentFactor   float64 // Adjustment factor (0.25 for 25% max change)
}

// PendingReward represents a pending reward to be claimed
type PendingReward struct {
	NodeAddress string
	Amount      sdk.Coin
	JobIDs      []string // Jobs that contributed to this reward
}

// NewPendingReward creates a new pending reward
func NewPendingReward(nodeAddress string, amount sdk.Coin, jobIDs []string) PendingReward {
	return PendingReward{
		NodeAddress: nodeAddress,
		Amount:      amount,
		JobIDs:      jobIDs,
	}
}