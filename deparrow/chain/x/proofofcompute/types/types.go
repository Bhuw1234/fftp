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
	RewardPerUnit        sdk.Dec
	DifficultyAdjustment uint64
	TargetBlockTime      uint32
}

// DefaultParams returns default proofofcompute module parameters
func DefaultParams() Params {
	return Params{
		MinComputeUnits:      1,
		RewardPerUnit:        sdk.NewDec(1), // 1 DPC per compute unit
		DifficultyAdjustment: 1000,          // Adjust every 1000 blocks
		TargetBlockTime:      6,             // 6 seconds
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MinComputeUnits == 0 {
		return fmt.Errorf("min compute units must be positive")
	}
	if p.TargetBlockTime == 0 {
		return fmt.Errorf("target block time must be positive")
	}
	return nil
}
