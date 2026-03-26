package proofofcompute

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/proofofcompute/keeper"
	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// BeginBlocker is called at the beginning of each block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Reset job count for the new block
	k.ResetBlockJobCount(ctx)
}

// EndBlocker is called at the end of each block
// It handles:
// 1. Difficulty adjustment
// 2. Pending reward distribution
// 3. Expired job cleanup (if implemented)
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Adjust difficulty based on network activity
	k.AdjustDifficulty(ctx)

	// Process any pending rewards that weren't distributed immediately
	processPendingRewards(ctx, k)

	// Log block statistics
	logBlockStats(ctx, k)
}

// processPendingRewards processes pending rewards for all compute nodes
func processPendingRewards(ctx sdk.Context, k keeper.Keeper) {
	// In a production system, we would iterate through pending rewards
	// and distribute them. For now, rewards are distributed immediately
	// upon proof verification, so this is a no-op placeholder.

	// This could be used for:
	// - Batch reward distribution (more efficient than per-job)
	// - Retry failed distributions
	// - Handle deferred rewards for dispute resolution
}

// logBlockStats logs statistics for the current block
func logBlockStats(ctx sdk.Context, k keeper.Keeper) {
	blockHeight := ctx.BlockHeight()
	jobsInBlock := k.GetBlockJobCount(ctx)
	difficulty := k.GetCurrentDifficulty(ctx)
	totalSupply := k.GetTotalSupply(ctx)

	k.Logger(ctx).Debug(
		"block stats",
		"height", blockHeight,
		"jobs", jobsInBlock,
		"difficulty", difficulty,
		"total_supply", totalSupply.String(),
	)
}

// ABCI EndBlock response type for reward distribution events
type EndBlockResponse struct {
	Events []sdk.Event
}

// NewEndBlockResponse creates a new EndBlock response with events
func NewEndBlockResponse() *EndBlockResponse {
	return &EndBlockResponse{
		Events: make([]sdk.Event, 0),
	}
}

// AddEvent adds an event to the response
func (r *EndBlockResponse) AddEvent(event sdk.Event) {
	r.Events = append(r.Events, event)
}

// GetEvents returns all events
func (r *EndBlockResponse) GetEvents() []sdk.Event {
	return r.Events
}
