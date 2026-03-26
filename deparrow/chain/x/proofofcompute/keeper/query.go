package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// queryServerImpl implements the QueryServer interface
type queryServerImpl struct {
	k Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServerImpl{keeper}
}

var _ types.QueryServer = queryServerImpl{}

// Params queries the module parameters
func (q queryServerImpl) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Job queries a specific job by ID
func (q queryServerImpl) Job(ctx context.Context, req *types.QueryJobRequest) (*types.QueryJobResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	job, err := q.k.GetJob(sdkCtx, req.JobId)
	if err != nil {
		return nil, err
	}

	return &types.QueryJobResponse{
		Job: *job,
	}, nil
}

// Jobs queries all jobs
func (q queryServerImpl) Jobs(ctx context.Context, req *types.QueryJobsRequest) (*types.QueryJobsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get all jobs from store
	store := sdkCtx.KVStore(q.k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.JobKey)
	defer iter.Close()

	var jobs []types.Job
	for ; iter.Valid(); iter.Next() {
		var job types.Job
		q.k.cdc.MustUnmarshal(iter.Value(), &job)
		jobs = append(jobs, job)
	}

	return &types.QueryJobsResponse{
		Jobs: jobs,
	}, nil
}

// JobsBySubmitter queries jobs by submitter
func (q queryServerImpl) JobsBySubmitter(ctx context.Context, req *types.QueryJobsBySubmitterRequest) (*types.QueryJobsBySubmitterResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	jobs, err := q.k.GetJobsBySubmitter(sdkCtx, req.Submitter)
	if err != nil {
		return nil, err
	}

	return &types.QueryJobsBySubmitterResponse{
		Jobs: jobs,
	}, nil
}

// JobsByComputeNode queries jobs by compute node
func (q queryServerImpl) JobsByComputeNode(ctx context.Context, req *types.QueryJobsByComputeNodeRequest) (*types.QueryJobsByComputeNodeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	jobs, err := q.k.GetJobsByComputeNode(sdkCtx, req.NodeAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryJobsByComputeNodeResponse{
		Jobs: jobs,
	}, nil
}

// Proof queries a proof by job ID
func (q queryServerImpl) Proof(ctx context.Context, req *types.QueryProofRequest) (*types.QueryProofResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	proof, err := q.k.GetProof(sdkCtx, req.JobId)
	if err != nil {
		return nil, err
	}

	return &types.QueryProofResponse{
		Proof: *proof,
	}, nil
}

// Difficulty queries the current difficulty
func (q queryServerImpl) Difficulty(ctx context.Context, req *types.QueryDifficultyRequest) (*types.QueryDifficultyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	difficulty, jobsInPeriod, blocksToAdjustment := q.k.GetDifficultyStats(sdkCtx)

	return &types.QueryDifficultyResponse{
		CurrentDifficulty:  difficulty,
		JobsInPeriod:       jobsInPeriod,
		BlocksToAdjustment: blocksToAdjustment,
	}, nil
}

// RewardStats queries reward statistics
func (q queryServerImpl) RewardStats(ctx context.Context, req *types.QueryRewardStatsRequest) (*types.QueryRewardStatsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	totalDistributed, totalJobs, err := q.k.GetRewardStats(sdkCtx)
	if err != nil {
		return nil, err
	}

	params := q.k.GetParams(sdkCtx)
	maxSupply, _ := params.GetMaxSupply()

	return &types.QueryRewardStatsResponse{
		TotalDistributed: totalDistributed.String(),
		TotalJobs:        totalJobs,
		MaxSupply:        maxSupply.String(),
	}, nil
}

// PendingReward queries pending reward for a compute node
func (q queryServerImpl) PendingReward(ctx context.Context, req *types.QueryPendingRewardRequest) (*types.QueryPendingRewardResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount, err := q.k.GetPendingReward(sdkCtx, req.NodeAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryPendingRewardResponse{
		Amount: amount,
	}, nil
}
