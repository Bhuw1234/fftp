package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
)

// QueryClient is an alias for the generated query client
// In production, this would be generated from the proto files
type QueryClient interface {
	Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error)
	Job(ctx context.Context, in *QueryJobRequest, opts ...interface{}) (*QueryJobResponse, error)
	Jobs(ctx context.Context, in *QueryJobsRequest, opts ...interface{}) (*QueryJobsResponse, error)
	JobsBySubmitter(ctx context.Context, in *QueryJobsBySubmitterRequest, opts ...interface{}) (*QueryJobsBySubmitterResponse, error)
	JobsByComputeNode(ctx context.Context, in *QueryJobsByComputeNodeRequest, opts ...interface{}) (*QueryJobsByComputeNodeResponse, error)
	Proof(ctx context.Context, in *QueryProofRequest, opts ...interface{}) (*QueryProofResponse, error)
	Difficulty(ctx context.Context, in *QueryDifficultyRequest, opts ...interface{}) (*QueryDifficultyResponse, error)
	RewardStats(ctx context.Context, in *QueryRewardStatsRequest, opts ...interface{}) (*QueryRewardStatsResponse, error)
	PendingReward(ctx context.Context, in *QueryPendingRewardRequest, opts ...interface{}) (*QueryPendingRewardResponse, error)
}

// NewQueryClient creates a new QueryClient from client context
func NewQueryClient(clientCtx client.Context) QueryClient {
	return &queryClient{clientCtx: clientCtx}
}

// queryClient implements QueryClient interface
type queryClient struct {
	clientCtx client.Context
}

func (q *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error) {
	// Placeholder - would use gRPC in production
	return &QueryParamsResponse{Params: DefaultParams()}, nil
}

func (q *queryClient) Job(ctx context.Context, in *QueryJobRequest, opts ...interface{}) (*QueryJobResponse, error) {
	return &QueryJobResponse{}, nil
}

func (q *queryClient) Jobs(ctx context.Context, in *QueryJobsRequest, opts ...interface{}) (*QueryJobsResponse, error) {
	return &QueryJobsResponse{}, nil
}

func (q *queryClient) JobsBySubmitter(ctx context.Context, in *QueryJobsBySubmitterRequest, opts ...interface{}) (*QueryJobsBySubmitterResponse, error) {
	return &QueryJobsBySubmitterResponse{}, nil
}

func (q *queryClient) JobsByComputeNode(ctx context.Context, in *QueryJobsByComputeNodeRequest, opts ...interface{}) (*QueryJobsByComputeNodeResponse, error) {
	return &QueryJobsByComputeNodeResponse{}, nil
}

func (q *queryClient) Proof(ctx context.Context, in *QueryProofRequest, opts ...interface{}) (*QueryProofResponse, error) {
	return &QueryProofResponse{}, nil
}

func (q *queryClient) Difficulty(ctx context.Context, in *QueryDifficultyRequest, opts ...interface{}) (*QueryDifficultyResponse, error) {
	return &QueryDifficultyResponse{}, nil
}

func (q *queryClient) RewardStats(ctx context.Context, in *QueryRewardStatsRequest, opts ...interface{}) (*QueryRewardStatsResponse, error) {
	return &QueryRewardStatsResponse{}, nil
}

func (q *queryClient) PendingReward(ctx context.Context, in *QueryPendingRewardRequest, opts ...interface{}) (*QueryPendingRewardResponse, error) {
	return &QueryPendingRewardResponse{}, nil
}

// RegisterQueryHandlerClient registers the query handler for gRPC gateway
// This is a placeholder - in production, this would be generated from proto
func RegisterQueryHandlerClient(ctx context.Context, mux interface{}, client QueryClient) error {
	// Placeholder for gRPC gateway registration
	return nil
}
