// Package types defines the types for the proofofcompute module
package types

// Module name and store key
const (
	ModuleName = "proofofcompute"
	StoreKey   = ModuleName
)

// KVStore key prefixes
var (
	// JobKey is the prefix for job entries
	JobKey = []byte{0x01}
	// ProofKey is the prefix for proof entries
	ProofKey = []byte{0x02}
	// PendingRewardKey is the prefix for pending rewards by node address
	PendingRewardKey = []byte{0x03}
	// JobCounterKey stores the next job ID
	JobCounterKey = []byte("next_job_id")
	// TotalSupplyKey stores total minted supply
	TotalSupplyKey = []byte("total_supply")
	// DifficultyKey stores current difficulty
	DifficultyKey = []byte("difficulty")
)

// KeyJob returns the key for a specific job
func KeyJob(jobID string) []byte {
	return append(JobKey, []byte(jobID)...)
}

// KeyProof returns the key for a proof by job ID
func KeyProof(jobID string) []byte {
	return append(ProofKey, []byte(jobID)...)
}

// KeyPendingReward returns the key for pending rewards by node address
func KeyPendingReward(nodeAddress string) []byte {
	return append(PendingRewardKey, []byte(nodeAddress)...)
}
