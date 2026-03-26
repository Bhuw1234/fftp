package types

import (
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

// KVStore keys
var (
	// JobKey is the prefix for job store
	JobKey = []byte{0x01}

	// JobBySubmitterKey is the prefix for jobs indexed by submitter
	JobBySubmitterKey = []byte{0x02}

	// JobByComputeNodeKey is the prefix for jobs indexed by compute node
	JobByComputeNodeKey = []byte{0x03}

	// ProofKey is the prefix for proof store
	ProofKey = []byte{0x04}

	// ParamsKey is the prefix for params store
	ParamsKey = []byte{0x05}

	// DifficultyKey is the key for current difficulty
	DifficultyKey = []byte{0x06}

	// TotalSupplyKey tracks total DPC minted
	TotalSupplyKey = []byte{0x07}

	// BlockJobCountKey tracks jobs completed per block for difficulty
	BlockJobCountKey = []byte{0x08}

	// PendingRewardsKey stores pending rewards to distribute
	PendingRewardsKey = []byte{0x09}
)

// GetJobKey returns the store key for a job
func GetJobKey(jobID string) []byte {
	return append(JobKey, []byte(jobID)...)
}

// GetJobBySubmitterKey returns the store key for jobs by submitter
func GetJobBySubmitterKey(submitter sdk.AccAddress, jobID string) []byte {
	return append(append(JobBySubmitterKey, submitter...), []byte(jobID)...)
}

// GetJobByComputeNodeKey returns the store key for jobs by compute node
func GetJobByComputeNodeKey(nodeAddr sdk.AccAddress, jobID string) []byte {
	return append(append(JobByComputeNodeKey, nodeAddr...), []byte(jobID)...)
}

// GetProofKey returns the store key for a proof
func GetProofKey(jobID string) []byte {
	return append(ProofKey, []byte(jobID)...)
}
