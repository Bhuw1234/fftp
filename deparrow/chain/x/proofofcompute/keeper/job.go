package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// SubmitJob submits a new compute job to the network
func (k Keeper) SubmitJob(ctx sdk.Context, msg *types.MsgSubmitJob) (*types.Job, error) {
	params := k.GetParams(ctx)

	// Validate minimum compute units
	if msg.ComputeUnits < params.MinComputeUnits {
		return nil, sdkerrors.Wrapf(
			types.ErrInsufficientCompute,
			"minimum compute units is %d, got %d",
			params.MinComputeUnits, msg.ComputeUnits,
		)
	}

	// Validate minimum stake
	minStake, err := params.GetMinStake()
	if err != nil {
		return nil, err
	}

	if msg.Stake.Amount.LT(minStake) {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidStake,
			"minimum stake is %s, got %s",
			minStake.String(), msg.Stake.Amount.String(),
		)
	}

	// Generate job ID (combines timestamp + submitter + counter)
	jobID := k.generateJobID(ctx, msg.Submitter)

	// Calculate complexity multiplier (1-5) based on job spec
	complexity := k.calculateComplexity(msg.JobSpec, msg.ComputeUnits)

	// Create the job
	job := &types.Job{
		ID:           jobID,
		Submitter:    msg.Submitter,
		Spec:         msg.JobSpec,
		Stake:        msg.Stake,
		Status:       types.JobStatusPending,
		ComputeUnits: msg.ComputeUnits,
		CreatedAt:    ctx.BlockTime().Unix(),
		Complexity:   complexity,
	}

	// Validate job
	if err := job.Validate(); err != nil {
		return nil, err
	}

	// Lock stake from submitter
	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidSubmitter, err.Error())
	}

	stakeCoins := sdk.NewCoins(msg.Stake)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, submitterAddr, types.ModuleName, stakeCoins,
	); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidStake, "failed to lock stake")
	}

	// Store the job
	if err := k.SetJob(ctx, job); err != nil {
		return nil, err
	}

	// Index by submitter
	if err := k.SetJobBySubmitter(ctx, job); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobSubmitted,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeySubmitter, job.Submitter),
			sdk.NewAttribute(types.AttributeKeyStake, job.Stake.String()),
			sdk.NewAttribute(types.AttributeKeyComputeUnits, fmt.Sprintf("%d", job.ComputeUnits)),
		),
	)

	k.Logger(ctx).Info(
		"job submitted",
		"job_id", job.ID,
		"submitter", job.Submitter,
		"compute_units", job.ComputeUnits,
		"stake", job.Stake.String(),
	)

	return job, nil
}

// StartJob marks a job as running on a compute node
func (k Keeper) StartJob(ctx sdk.Context, jobID, computeNode string) error {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != types.JobStatusPending {
		return sdkerrors.Wrapf(
			types.ErrJobNotRunning,
			"job status is %s, expected pending",
			job.Status.String(),
		)
	}

	// Validate compute node address
	_, err = sdk.AccAddressFromBech32(computeNode)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidComputeNode, err.Error())
	}

	job.ComputeNode = computeNode
	job.Status = types.JobStatusRunning

	if err := k.SetJob(ctx, job); err != nil {
		return err
	}

	// Index by compute node
	if err := k.SetJobByComputeNode(ctx, job); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobStarted,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeyComputeNode, computeNode),
		),
	)

	return nil
}

// CompleteJob marks a job as completed and stores the result
func (k Keeper) CompleteJob(ctx sdk.Context, jobID string, result []byte, computeUnits uint64) error {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != types.JobStatusRunning {
		return sdkerrors.Wrapf(
			types.ErrJobNotRunning,
			"job status is %s, expected running",
			job.Status.String(),
		)
	}

	job.Result = result
	job.ComputeUnits = computeUnits
	job.Status = types.JobStatusCompleted
	job.CompletedAt = ctx.BlockTime().Unix()

	if err := k.SetJob(ctx, job); err != nil {
		return err
	}

	// Increment block job count for difficulty adjustment
	k.IncrementBlockJobCount(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobCompleted,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeyComputeNode, job.ComputeNode),
			sdk.NewAttribute(types.AttributeKeyComputeUnits, fmt.Sprintf("%d", computeUnits)),
			sdk.NewAttribute(types.AttributeKeyResult, string(result)),
		),
	)

	return nil
}

// FailJob marks a job as failed
func (k Keeper) FailJob(ctx sdk.Context, jobID string, reason string) error {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status == types.JobStatusCompleted || job.Status == types.JobStatusFailed {
		return sdkerrors.Wrap(types.ErrJobAlreadyCompleted, "job already finished")
	}

	job.Status = types.JobStatusFailed
	job.CompletedAt = ctx.BlockTime().Unix()

	if err := k.SetJob(ctx, job); err != nil {
		return err
	}

	// Refund stake to submitter
	submitterAddr, err := sdk.AccAddressFromBech32(job.Submitter)
	if err != nil {
		return err
	}

	stakeCoins := sdk.NewCoins(job.Stake)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, submitterAddr, stakeCoins,
	); err != nil {
		k.Logger(ctx).Error("failed to refund stake", "job_id", jobID, "error", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobFailed,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeyStatus, reason),
		),
	)

	return nil
}

// CancelJob cancels a pending job
func (k Keeper) CancelJob(ctx sdk.Context, jobID, submitter string) error {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	// Verify submitter is the job owner
	if job.Submitter != submitter {
		return sdkerrors.Wrap(types.ErrUnauthorized, "only submitter can cancel")
	}

	if job.Status != types.JobStatusPending {
		return sdkerrors.Wrapf(
			types.ErrJobAlreadyCompleted,
			"job status is %s, can only cancel pending jobs",
			job.Status.String(),
		)
	}

	job.Status = types.JobStatusCancelled
	job.CompletedAt = ctx.BlockTime().Unix()

	if err := k.SetJob(ctx, job); err != nil {
		return err
	}

	// Refund stake to submitter
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return err
	}

	stakeCoins := sdk.NewCoins(job.Stake)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, submitterAddr, stakeCoins,
	); err != nil {
		k.Logger(ctx).Error("failed to refund stake on cancel", "job_id", jobID, "error", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobCancelled,
			sdk.NewAttribute(types.AttributeKeyJobID, job.ID),
			sdk.NewAttribute(types.AttributeKeySubmitter, submitter),
		),
	)

	return nil
}

// GetJob retrieves a job by ID
func (k Keeper) GetJob(ctx sdk.Context, jobID string) (*types.Job, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetJobKey(jobID)

	bz := store.Get(key)
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrJobNotFound, jobID)
	}

	var job types.Job
	k.cdc.MustUnmarshal(bz, &job)
	return &job, nil
}

// SetJob stores a job
func (k Keeper) SetJob(ctx sdk.Context, job *types.Job) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetJobKey(job.ID)
	bz := k.cdc.MustMarshal(job)
	store.Set(key, bz)
	return nil
}

// SetJobBySubmitter indexes a job by submitter
func (k Keeper) SetJobBySubmitter(ctx sdk.Context, job *types.Job) error {
	store := ctx.KVStore(k.storeKey)
	submitterAddr, err := sdk.AccAddressFromBech32(job.Submitter)
	if err != nil {
		return err
	}
	key := types.GetJobBySubmitterKey(submitterAddr, job.ID)
	bz := k.cdc.MustMarshal(job)
	store.Set(key, bz)
	return nil
}

// SetJobByComputeNode indexes a job by compute node
func (k Keeper) SetJobByComputeNode(ctx sdk.Context, job *types.Job) error {
	store := ctx.KVStore(k.storeKey)
	nodeAddr, err := sdk.AccAddressFromBech32(job.ComputeNode)
	if err != nil {
		return err
	}
	key := types.GetJobByComputeNodeKey(nodeAddr, job.ID)
	bz := k.cdc.MustMarshal(job)
	store.Set(key, bz)
	return nil
}

// GetJobsBySubmitter retrieves all jobs for a submitter
func (k Keeper) GetJobsBySubmitter(ctx sdk.Context, submitter string) ([]types.Job, error) {
	store := ctx.KVStore(k.storeKey)
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return nil, err
	}

	var jobs []types.Job
	prefix := append(types.JobBySubmitterKey, submitterAddr...)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var job types.Job
		k.cdc.MustUnmarshal(iter.Value(), &job)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetJobsByComputeNode retrieves all jobs for a compute node
func (k Keeper) GetJobsByComputeNode(ctx sdk.Context, nodeAddr string) ([]types.Job, error) {
	store := ctx.KVStore(k.storeKey)
	nodeAddress, err := sdk.AccAddressFromBech32(nodeAddr)
	if err != nil {
		return nil, err
	}

	var jobs []types.Job
	prefix := append(types.JobByComputeNodeKey, nodeAddress...)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var job types.Job
		k.cdc.MustUnmarshal(iter.Value(), &job)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// generateJobID generates a unique job ID
func (k Keeper) generateJobID(ctx sdk.Context, submitter string) string {
	// Use timestamp + submitter + block height for uniqueness
	timestamp := time.Now().UnixNano()
	blockHeight := ctx.BlockHeight()

	// Create a unique ID
	idBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(idBytes[:8], uint64(timestamp))
	binary.BigEndian.PutUint64(idBytes[8:], uint64(blockHeight))

	// Combine with submitter for additional uniqueness
	shortAddr := submitter
	if len(submitter) > 8 {
		shortAddr = submitter[:8]
	}
	return fmt.Sprintf("job-%d-%d-%s", timestamp, blockHeight, shortAddr)
}

// calculateComplexity calculates the complexity multiplier (1-5) based on job spec
func (k Keeper) calculateComplexity(jobSpec []byte, computeUnits uint64) uint32 {
	params := k.GetParams(ctx)
	maxMultiplier := params.ComplexityMultiplier

	// Base complexity on compute units
	// More compute units = higher complexity
	if computeUnits > 10000 {
		return minUint32(maxMultiplier, 5)
	} else if computeUnits > 5000 {
		return minUint32(maxMultiplier, 4)
	} else if computeUnits > 1000 {
		return minUint32(maxMultiplier, 3)
	} else if computeUnits > 100 {
		return minUint32(maxMultiplier, 2)
	}
	return 1
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}