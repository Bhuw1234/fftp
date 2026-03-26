package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// SubmitJob submits a new compute job to the network
	SubmitJob(context.Context, *MsgSubmitJob) (*MsgSubmitJobResponse, error)
	// SubmitProof submits a compute proof for verification
	SubmitProof(context.Context, *MsgSubmitProof) (*MsgSubmitProofResponse, error)
	// CancelJob cancels a pending job
	CancelJob(context.Context, *MsgCancelJob) (*MsgCancelJobResponse, error)
	// ClaimReward claims pending rewards
	ClaimReward(context.Context, *MsgClaimReward) (*MsgClaimRewardResponse, error)
}

// MsgSubmitJobResponse is the response type for the SubmitJob RPC method.
type MsgSubmitJobResponse struct {
	JobId  string `json:"job_id"`
	Status string `json:"status"`
}

// MsgSubmitProofResponse is the response type for the SubmitProof RPC method.
type MsgSubmitProofResponse struct {
	JobId  string `json:"job_id"`
	Status string `json:"status"`
	Reward string `json:"reward"`
}

// MsgCancelJobResponse is the response type for the CancelJob RPC method.
type MsgCancelJobResponse struct {
	JobId  string `json:"job_id"`
	Status string `json:"status"`
}

// MsgClaimRewardResponse is the response type for the ClaimReward RPC method.
type MsgClaimRewardResponse struct {
	Amount string `json:"amount"`
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Params queries the parameters of the module
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// Job queries a specific job by ID
	Job(context.Context, *QueryJobRequest) (*QueryJobResponse, error)
	// Jobs queries all jobs
	Jobs(context.Context, *QueryJobsRequest) (*QueryJobsResponse, error)
	// JobsBySubmitter queries jobs by submitter
	JobsBySubmitter(context.Context, *QueryJobsBySubmitterRequest) (*QueryJobsBySubmitterResponse, error)
	// JobsByComputeNode queries jobs by compute node
	JobsByComputeNode(context.Context, *QueryJobsByComputeNodeRequest) (*QueryJobsByComputeNodeResponse, error)
	// Proof queries a proof by job ID
	Proof(context.Context, *QueryProofRequest) (*QueryProofResponse, error)
	// Difficulty queries the current difficulty
	Difficulty(context.Context, *QueryDifficultyRequest) (*QueryDifficultyResponse, error)
	// RewardStats queries reward statistics
	RewardStats(context.Context, *QueryRewardStatsRequest) (*QueryRewardStatsResponse, error)
	// PendingReward queries pending reward for a compute node
	PendingReward(context.Context, *QueryPendingRewardRequest) (*QueryPendingRewardResponse, error)
}

// Query params request/response types

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryJobRequest struct {
	JobId string `json:"job_id"`
}

type QueryJobResponse struct {
	Job Job `json:"job"`
}

type QueryJobsRequest struct {
	// Pagination support could be added here
}

type QueryJobsResponse struct {
	Jobs []Job `json:"jobs"`
}

type QueryJobsBySubmitterRequest struct {
	Submitter string `json:"submitter"`
}

type QueryJobsBySubmitterResponse struct {
	Jobs []Job `json:"jobs"`
}

type QueryJobsByComputeNodeRequest struct {
	NodeAddress string `json:"node_address"`
}

type QueryJobsByComputeNodeResponse struct {
	Jobs []Job `json:"jobs"`
}

type QueryProofRequest struct {
	JobId string `json:"job_id"`
}

type QueryProofResponse struct {
	Proof ComputeProof `json:"proof"`
}

type QueryDifficultyRequest struct{}

type QueryDifficultyResponse struct {
	CurrentDifficulty  uint64 `json:"current_difficulty"`
	JobsInPeriod       uint64 `json:"jobs_in_period"`
	BlocksToAdjustment int64  `json:"blocks_to_adjustment"`
}

type QueryRewardStatsRequest struct{}

type QueryRewardStatsResponse struct {
	TotalDistributed string `json:"total_distributed"`
	TotalJobs        uint64 `json:"total_jobs"`
	MaxSupply        string `json:"max_supply"`
}

type QueryPendingRewardRequest struct {
	NodeAddress string `json:"node_address"`
}

type QueryPendingRewardResponse struct {
	Amount sdk.Coin `json:"amount"`
}

// Unimplemented MsgServer
type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) SubmitJob(context.Context, *MsgSubmitJob) (*MsgSubmitJobResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) SubmitProof(context.Context, *MsgSubmitProof) (*MsgSubmitProofResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) CancelJob(context.Context, *MsgCancelJob) (*MsgCancelJobResponse, error) {
	return nil, nil
}
func (UnimplementedMsgServer) ClaimReward(context.Context, *MsgClaimReward) (*MsgClaimRewardResponse, error) {
	return nil, nil
}

// Unimplemented QueryServer
type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Job(context.Context, *QueryJobRequest) (*QueryJobResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Jobs(context.Context, *QueryJobsRequest) (*QueryJobsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) JobsBySubmitter(context.Context, *QueryJobsBySubmitterRequest) (*QueryJobsBySubmitterResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) JobsByComputeNode(context.Context, *QueryJobsByComputeNodeRequest) (*QueryJobsByComputeNodeResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Proof(context.Context, *QueryProofRequest) (*QueryProofResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) Difficulty(context.Context, *QueryDifficultyRequest) (*QueryDifficultyResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) RewardStats(context.Context, *QueryRewardStatsRequest) (*QueryRewardStatsResponse, error) {
	return nil, nil
}
func (UnimplementedQueryServer) PendingReward(context.Context, *QueryPendingRewardRequest) (*QueryPendingRewardResponse, error) {
	return nil, nil
}
