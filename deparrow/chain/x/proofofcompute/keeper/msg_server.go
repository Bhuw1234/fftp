// Package keeper implements the message server for the proofofcompute module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// MsgSubmitJob represents a job submission message
type MsgSubmitJob struct {
	Submitter    string `json:"submitter"`
	JobSpec      []byte `json:"job_spec"`
	Stake        string `json:"stake"`
	ComputeUnits uint64 `json:"compute_units"`
}

// MsgSubmitJobResponse is the response for job submission
type MsgSubmitJobResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// MsgSubmitProof represents a proof submission message
type MsgSubmitProof struct {
	JobID         string `json:"job_id"`
	NodeAddress   string `json:"node_address"`
	ComputeUnits  uint64 `json:"compute_units"`
	ExecutionTime int64  `json:"execution_time"`
	OutputHash    []byte `json:"output_hash"`
	Signature     []byte `json:"signature"`
}

// MsgSubmitProofResponse is the response for proof submission
type MsgSubmitProofResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Reward string `json:"reward"`
}

// MsgCancelJob represents a job cancellation message
type MsgCancelJob struct {
	JobID     string `json:"job_id"`
	Submitter string `json:"submitter"`
}

// MsgCancelJobResponse is the response for job cancellation
type MsgCancelJobResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// MsgClaimReward represents a reward claim message
type MsgClaimReward struct {
	NodeAddress string `json:"node_address"`
}

// MsgClaimRewardResponse is the response for reward claim
type MsgClaimRewardResponse struct {
	Amount string `json:"amount"`
}

// SubmitJob handles job submission with input validation
func (k Keeper) SubmitJob(msg MsgSubmitJob, blockHeight int64) (*MsgSubmitJobResponse, error) {
	// SECURITY: Input validation
	if msg.Submitter == "" {
		return nil, fmt.Errorf("submitter cannot be empty")
	}
	if !isValidAddress(msg.Submitter) {
		return nil, fmt.Errorf("invalid submitter address format")
	}
	if len(msg.JobSpec) == 0 {
		return nil, fmt.Errorf("job_spec cannot be empty")
	}
	if msg.ComputeUnits < 1 {
		return nil, fmt.Errorf("compute_units must be at least 1")
	}

	// Validate compute units
	params := k.GetParams()
	if msg.ComputeUnits < params.MinComputeUnits {
		return nil, types.ErrInvalidComputeUnits
	}

	// Get next job ID
	jobID, err := k.IncrementJobID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate job ID: %w", err)
	}

	// Create job
	job := types.Job{
		ID:           jobID,
		Submitter:    msg.Submitter,
		Spec:         msg.JobSpec,
		Stake:        types.Coin{Denom: "dpc", Amount: msg.Stake},
		ComputeUnits: msg.ComputeUnits,
		Status:       types.JobStatusPending,
		CreatedAt:    blockHeight,
		Complexity:   1, // Default complexity
	}

	// Store job
	if err := k.SetJob(job); err != nil {
		return nil, fmt.Errorf("failed to store job: %w", err)
	}

	log.Printf("[proofofcompute] Job %s submitted by %s", jobID, msg.Submitter)

	return &MsgSubmitJobResponse{
		JobID:  jobID,
		Status: types.JobStatusPending.String(),
	}, nil
}

// MaxComputeUnits is the maximum allowed compute units per job
const MaxComputeUnits = 1000000000 // 1 billion

// reentrancyGuard tracks addresses currently claiming rewards
var reentrancyGuard = make(map[string]bool)

// SubmitProof handles proof submission with security checks
func (k Keeper) SubmitProof(msg MsgSubmitProof, blockHeight int64) (*MsgSubmitProofResponse, error) {
	// SECURITY: Input validation
	if msg.JobID == "" {
		return nil, fmt.Errorf("job_id cannot be empty")
	}
	if msg.NodeAddress == "" {
		return nil, fmt.Errorf("node_address cannot be empty")
	}
	if !isValidAddress(msg.NodeAddress) {
		return nil, fmt.Errorf("invalid node_address format")
	}
	if msg.ComputeUnits == 0 {
		return nil, types.ErrInvalidComputeUnits
	}
	// SECURITY: Maximum compute units check
	if msg.ComputeUnits > MaxComputeUnits {
		return nil, fmt.Errorf("compute_units exceeds maximum (%d)", MaxComputeUnits)
	}
	if msg.ExecutionTime < 0 {
		return nil, fmt.Errorf("execution_time cannot be negative")
	}

	// SECURITY: Signature verification (placeholder - implement with actual crypto)
	// TODO: Implement proper Ed25519 signature verification
	// if !k.VerifySignature(msg.NodeAddress, msg.JobID, msg.OutputHash, msg.Signature) {
	//     return nil, fmt.Errorf("invalid signature")
	// }
	// For now, require signature to be non-empty
	if len(msg.Signature) == 0 {
		return nil, fmt.Errorf("signature is required")
	}

	// Get job
	job, found := k.GetJob(msg.JobID)
	if !found {
		return nil, types.ErrJobNotFound
	}

	// Validate job state
	if job.Status == types.JobStatusCompleted {
		return nil, types.ErrJobAlreadyCompleted
	}
	if job.Status != types.JobStatusPending && job.Status != types.JobStatusRunning {
		return nil, types.ErrJobNotRunning
	}

	// Calculate reward
	// Formula: DPC = 0.001 × Complexity × ComputeUnits
	reward := k.CalculateReward(msg.ComputeUnits, job.Complexity)

	// Check max supply
	params := k.GetParams()
	currentSupply := k.GetTotalSupply()
	rewardUint := parseUint64(reward)
	maxSupply := parseUint64(params.MaxSupply)
	currentSupplyUint := parseUint64(currentSupply)

	// SECURITY: Overflow check for supply
	if currentSupplyUint > maxSupply-rewardUint {
		// Cap reward to remaining supply
		remaining := maxSupply - currentSupplyUint
		reward = fmt.Sprintf("%d", remaining)
	}

	// Create proof
	proof := types.ComputeProof{
		JobID:         msg.JobID,
		NodeID:        msg.NodeAddress,
		ComputeUnits:  msg.ComputeUnits,
		ExecutionTime: msg.ExecutionTime,
		OutputHash:    msg.OutputHash,
		Signature:     msg.Signature,
	}

	// Store proof
	if err := k.SetProof(proof); err != nil {
		return nil, fmt.Errorf("failed to store proof: %w", err)
	}

	// Update job
	job.ComputeNode = msg.NodeAddress
	job.Result = msg.OutputHash
	job.Status = types.JobStatusCompleted
	job.CompletedAt = blockHeight
	job.Reward = types.Coin{Denom: "dpc", Amount: reward}

	// Store updated job
	if err := k.SetJob(job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	// Add pending reward for the compute node
	if err := k.AddPendingReward(msg.NodeAddress, reward); err != nil {
		return nil, fmt.Errorf("failed to add pending reward: %w", err)
	}

	// Update total supply
	if err := k.AddToTotalSupply(reward); err != nil {
		return nil, fmt.Errorf("failed to update total supply: %w", err)
	}

	log.Printf("[proofofcompute] Proof submitted for job %s, reward: %s DPC", msg.JobID, formatDPC(reward))

	return &MsgSubmitProofResponse{
		JobID:  msg.JobID,
		Status: types.JobStatusCompleted.String(),
		Reward: reward,
	}, nil
}

// CancelJob handles job cancellation
func (k Keeper) CancelJob(msg MsgCancelJob) (*MsgCancelJobResponse, error) {
	// Get job
	job, found := k.GetJob(msg.JobID)
	if !found {
		return nil, types.ErrJobNotFound
	}

	// Validate submitter
	if job.Submitter != msg.Submitter {
		return nil, types.ErrInvalidSubmitter
	}

	// Check if job can be cancelled
	if job.Status == types.JobStatusCompleted {
		return nil, types.ErrJobAlreadyCompleted
	}

	// Update job status
	job.Status = types.JobStatusCancelled

	// Store updated job
	if err := k.SetJob(job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	log.Printf("[proofofcompute] Job %s cancelled by %s", msg.JobID, msg.Submitter)

	return &MsgCancelJobResponse{
		JobID:  msg.JobID,
		Status: types.JobStatusCancelled.String(),
	}, nil
}

// ClaimReward handles reward claiming with reentrancy protection
func (k Keeper) ClaimReward(msg MsgClaimReward) (*MsgClaimRewardResponse, error) {
	// SECURITY: Input validation
	if msg.NodeAddress == "" {
		return nil, fmt.Errorf("node_address cannot be empty")
	}
	if !isValidAddress(msg.NodeAddress) {
		return nil, fmt.Errorf("invalid node_address format")
	}

	// SECURITY: Reentrancy guard
	if reentrancyGuard[msg.NodeAddress] {
		return nil, fmt.Errorf("claim already in progress for this address")
	}
	reentrancyGuard[msg.NodeAddress] = true
	defer func() { delete(reentrancyGuard, msg.NodeAddress) }()

	// Get pending reward
	reward, found := k.GetPendingReward(msg.NodeAddress)
	if !found || reward.Amount == "0" {
		return nil, types.ErrNoPendingReward
	}

	// Clear pending reward
	amount, err := k.ClearPendingReward(msg.NodeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to claim reward: %w", err)
	}

	log.Printf("[proofofcompute] Reward claimed by %s: %s DPC", msg.NodeAddress, formatDPC(amount))

	return &MsgClaimRewardResponse{
		Amount: amount,
	}, nil
}

// ProcessTransaction processes a transaction based on type
func (k Keeper) ProcessTransaction(txType string, txData json.RawMessage, blockHeight int64) (interface{}, error) {
	switch txType {
	case "submit_job":
		var msg MsgSubmitJob
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse submit_job: %w", err)
		}
		return k.SubmitJob(msg, blockHeight)

	case "submit_proof":
		var msg MsgSubmitProof
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse submit_proof: %w", err)
		}
		return k.SubmitProof(msg, blockHeight)

	case "cancel_job":
		var msg MsgCancelJob
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse cancel_job: %w", err)
		}
		return k.CancelJob(msg)

	case "claim_reward":
		var msg MsgClaimReward
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse claim_reward: %w", err)
		}
		return k.ClaimReward(msg)

	default:
		return nil, fmt.Errorf("unknown transaction type: %s", txType)
	}
}

// formatDPC formats a base unit amount to DPC display format
func formatDPC(amount string) string {
	// For display purposes, convert base units to DPC
	// 1 DPC = 10^18 base units
	val := parseUint64(amount)
	if val == 0 {
		return "0"
	}
	dpc := float64(val) / 1e18
	return fmt.Sprintf("%.6f", dpc)
}

// isValidAddress validates a DPC address format (bech32 with dpc prefix)
func isValidAddress(address string) bool {
	// Basic bech32 validation: starts with dpc1 and has reasonable length
	if len(address) < 10 || len(address) > 100 {
		return false
	}
	if !strings.HasPrefix(address, "dpc1") {
		return false
	}
	// Check that remaining characters are valid bech32
	for _, c := range address[4:] {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// parseUint64Safe safely parses a string to uint64
func parseUint64Safe(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
