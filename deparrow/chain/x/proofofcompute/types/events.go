package types

// Event types for the proofofcompute module
const (
	EventTypeJobSubmitted    = "job_submitted"
	EventTypeJobStarted      = "job_started"
	EventTypeJobCompleted    = "job_completed"
	EventTypeJobFailed       = "job_failed"
	EventTypeJobCancelled    = "job_cancelled"
	EventTypeProofSubmitted  = "proof_submitted"
	EventTypeProofVerified   = "proof_verified"
	EventTypeRewardDistributed = "reward_distributed"
	EventTypeDifficultyAdjusted = "difficulty_adjusted"

	AttributeKeyJobID        = "job_id"
	AttributeKeySubmitter    = "submitter"
	AttributeKeyComputeNode  = "compute_node"
	AttributeKeyStatus       = "status"
	AttributeKeyComputeUnits = "compute_units"
	AttributeKeyReward       = "reward"
	AttributeKeyStake        = "stake"
	AttributeKeyResult       = "result"
	AttributeKeyProofHash    = "proof_hash"
	AttributeKeyDifficulty   = "difficulty"
	AttributeKeyExecutionTime = "execution_time"
	AttributeKeyComplexity   = "complexity_multiplier"
	AttributeKeyTotalSupply  = "total_supply"
)

// NewJobSubmittedEvent creates a new job submitted event
func NewJobSubmittedEvent(jobID, submitter string, stake string) string {
	return "" // Event attributes are handled by sdk.NewEvent
}

// NewJobCompletedEvent creates a new job completed event
func NewJobCompletedEvent(jobID, computeNode string, computeUnits uint64, reward string) string {
	return ""
}

// NewRewardDistributedEvent creates a new reward distributed event
func NewRewardDistributedEvent(computeNode string, reward string, jobID string) string {
	return ""
}
