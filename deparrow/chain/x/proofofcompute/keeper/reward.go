package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// CalculateReward calculates the DPC reward for a completed job
// Formula: DPC_reward = Base_Rate × Job_Complexity × Compute_Units
// Base_Rate = 0.001 DPC
// Complexity multiplier: 1x-5x
func (k Keeper) CalculateReward(ctx sdk.Context, computeUnits uint64, complexity uint32) (sdk.Coin, error) {
	params := k.GetParams(ctx)

	// Get base rate (0.001 DPC)
	baseRate, err := params.GetRewardPerUnit()
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to get reward per unit: %w", err)
	}

	// Calculate reward: Base_Rate × Complexity × Compute_Units
	// Convert complexity to Dec for multiplication
	complexityDec := sdk.NewDec(int64(complexity))
	computeUnitsDec := sdk.NewDec(int64(computeUnits))

	// Reward in DPC (smallest unit: 1e-18 DPC)
	rewardDec := baseRate.Mul(complexityDec).Mul(computeUnitsDec)

	// Convert to coin (DPC has 18 decimals, so multiply by 1e18)
	rewardAmount := rewardDec.MulInt(sdk.NewInt(1e18)).TruncateInt()

	// Ensure minimum reward
	if rewardAmount.LT(sdk.OneInt()) {
		rewardAmount = sdk.OneInt()
	}

	return sdk.NewCoin("dpc", rewardAmount), nil
}

// DistributeReward distributes reward to a compute node for completed work
func (k Keeper) DistributeReward(ctx sdk.Context, job *types.Job) error {
	// Calculate reward
	reward, err := k.CalculateReward(ctx, job.ComputeUnits, job.Complexity)
	if err != nil {
		return fmt.Errorf("failed to calculate reward: %w", err)
	}

	// Check max supply constraint
	params := k.GetParams(ctx)
	maxSupply, err := params.GetMaxSupply()
	if err != nil {
		return fmt.Errorf("failed to get max supply: %w", err)
	}

	currentSupply := k.GetTotalSupply(ctx)
	newSupply := currentSupply.Add(reward.Amount)

	if newSupply.GT(maxSupply) {
		k.Logger(ctx).Warn(
			"reward would exceed max supply, adjusting",
			"requested", reward.Amount.String(),
			"current_supply", currentSupply.String(),
			"max_supply", maxSupply.String(),
		)
		// Adjust reward to not exceed max supply
		remaining := maxSupply.Sub(currentSupply)
		if remaining.IsZero() || remaining.IsNegative() {
			return types.ErrMaxSupplyExceeded
		}
		reward = sdk.NewCoin("dpc", remaining)
	}

	// Mint and send reward to compute node
	if err := k.MintCoins(ctx, job.ComputeNode, sdk.NewCoins(reward)); err != nil {
		return fmt.Errorf("failed to mint reward: %w", err)
	}

	// Update job with reward amount
	job.Reward = reward
	if err := k.SetJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job reward: %w", err)
	}

	// Emit reward distribution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardDistributed,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeyComputeNode, job.ComputeNode),
			sdk.NewAttribute(types.AttributeKeyReward, reward.String()),
			sdk.NewAttribute(types.AttributeKeyTotalSupply, k.GetTotalSupply(ctx).String()),
		),
	)

	k.Logger(ctx).Info(
		"reward distributed",
		"job_id", job.ID,
		"compute_node", job.ComputeNode,
		"reward", reward.String(),
		"compute_units", job.ComputeUnits,
		"complexity", job.Complexity,
	)

	return nil
}

// GetPendingReward gets the pending reward for a compute node
func (k Keeper) GetPendingReward(ctx sdk.Context, nodeAddress string) (sdk.Coin, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.PendingRewardsKey, []byte(nodeAddress)...)

	bz := store.Get(key)
	if bz == nil {
		return sdk.NewCoin("dpc", sdk.ZeroInt()), nil
	}

	var reward sdk.Coin
	k.cdc.MustUnmarshal(bz, &reward)
	return reward, nil
}

// SetPendingReward sets the pending reward for a compute node
func (k Keeper) SetPendingReward(ctx sdk.Context, nodeAddress string, reward sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.PendingRewardsKey, []byte(nodeAddress)...)
	bz := k.cdc.MustMarshal(&reward)
	store.Set(key, bz)
}

// AddPendingReward adds to the pending reward for a compute node
func (k Keeper) AddPendingReward(ctx sdk.Context, nodeAddress string, amount sdk.Coin) error {
	current, err := k.GetPendingReward(ctx, nodeAddress)
	if err != nil {
		return err
	}

	newReward := current.Add(amount)
	k.SetPendingReward(ctx, nodeAddress, newReward)
	return nil
}

// ClaimReward claims pending rewards for a compute node
func (k Keeper) ClaimReward(ctx sdk.Context, nodeAddress string) (sdk.Coin, error) {
	pending, err := k.GetPendingReward(ctx, nodeAddress)
	if err != nil {
		return sdk.Coin{}, err
	}

	if pending.IsZero() {
		return sdk.NewCoin("dpc", sdk.ZeroInt()), nil
	}

	// Mint and send the pending reward
	if err := k.MintCoins(ctx, nodeAddress, sdk.NewCoins(pending)); err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to mint pending reward: %w", err)
	}

	// Clear pending reward
	k.SetPendingReward(ctx, nodeAddress, sdk.NewCoin("dpc", sdk.ZeroInt()))

	k.Logger(ctx).Info(
		"pending reward claimed",
		"node", nodeAddress,
		"amount", pending.String(),
	)

	return pending, nil
}

// CalculateEstimatedReward estimates reward for given parameters (read-only)
func (k Keeper) CalculateEstimatedReward(ctx sdk.Context, computeUnits uint64, complexity uint32) (sdk.Coin, error) {
	return k.CalculateReward(ctx, computeUnits, complexity)
}

// GetRewardStats returns reward statistics
func (k Keeper) GetRewardStats(ctx sdk.Context) (totalDistributed sdk.Int, totalJobs uint64, err error) {
	totalDistributed = k.GetTotalSupply(ctx)

	// Count completed jobs (this is a simplified approach)
	// In production, you'd want to track this separately
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.JobKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var job types.Job
		k.cdc.MustUnmarshal(iter.Value(), &job)
		if job.Status == types.JobStatusCompleted {
			totalJobs++
		}
	}

	return totalDistributed, totalJobs, nil
}
