package types

import (
	"encoding/json"
	"fmt"
)

// JobStatus represents the status of a compute job
type JobStatus int32

const (
	JobStatusUnspecified JobStatus = 0
	JobStatusPending     JobStatus = 1
	JobStatusRunning     JobStatus = 2
	JobStatusCompleted   JobStatus = 3
	JobStatusFailed      JobStatus = 4
	JobStatusCancelled   JobStatus = 5
)

// String returns the string representation of JobStatus
func (s JobStatus) String() string {
	switch s {
	case JobStatusUnspecified:
		return "unspecified"
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
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// Coin represents a token amount with denom
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// NewCoin creates a new coin
func NewCoin(denom, amount string) Coin {
	return Coin{Denom: denom, Amount: amount}
}

// Job represents a compute job in the DEparrow network
type Job struct {
	// ID is the unique job identifier
	ID string `json:"id"`
	// Submitter is the address of the job submitter
	Submitter string `json:"submitter"`
	// ComputeNode is the address of the node that executed the job
	ComputeNode string `json:"compute_node"`
	// Spec is the job specification (Bacalhau Job spec)
	Spec []byte `json:"spec"`
	// Result is the IPFS CID of results
	Result []byte `json:"result"`
	// Stake is the staked amount for the job
	Stake Coin `json:"stake"`
	// Reward is the reward amount for completion
	Reward Coin `json:"reward"`
	// Status is the current job status
	Status JobStatus `json:"status"`
	// ComputeUnits is the estimated compute units
	ComputeUnits uint64 `json:"compute_units"`
	// CreatedAt is the block height when job was created
	CreatedAt int64 `json:"created_at"`
	// CompletedAt is the block height when job was completed
	CompletedAt int64 `json:"completed_at"`
	// Complexity is the complexity multiplier (1-5)
	Complexity uint32 `json:"complexity"`
}

// NewJob creates a new job
func NewJob(id, submitter string, spec []byte, stake Coin, computeUnits uint64) Job {
	return Job{
		ID:           id,
		Submitter:    submitter,
		Spec:         spec,
		Stake:        stake,
		ComputeUnits: computeUnits,
		Status:       JobStatusPending,
		CreatedAt:    0, // Will be set during processing
		Complexity:   1, // Default complexity
	}
}

// Bytes serializes job to bytes
func (j Job) Bytes() []byte {
	bz, _ := json.Marshal(j)
	return bz
}

// ComputeProof represents proof of computation
type ComputeProof struct {
	// JobID is the ID of the job this proof is for
	JobID string `json:"job_id"`
	// NodeID is the node address
	NodeID string `json:"node_id"`
	// ComputeUnits is the actual compute units consumed
	ComputeUnits uint64 `json:"compute_units"`
	// ExecutionTime is the execution time in seconds
	ExecutionTime int64 `json:"execution_time"`
	// OutputHash is the deterministic hash of outputs
	OutputHash []byte `json:"output_hash"`
	// Signature is the node's signature
	Signature []byte `json:"signature"`
}

// Bytes serializes proof to bytes
func (p ComputeProof) Bytes() []byte {
	bz, _ := json.Marshal(p)
	return bz
}

// PendingReward tracks rewards pending claim by a node
type PendingReward struct {
	// NodeAddress is the address of the compute node
	NodeAddress string `json:"node_address"`
	// Amount is the total pending reward
	Amount string `json:"amount"`
	// JobsCompleted is the number of jobs completed
	JobsCompleted uint64 `json:"jobs_completed"`
}
