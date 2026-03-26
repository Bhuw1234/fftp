package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// CreateEscrow creates a new escrow for a job
func (k Keeper) CreateEscrow(ctx sdk.Context, jobID, submitter, provider string, amount sdk.Coin, duration int64) (*types.EscrowExtended, error) {
	params := k.GetParams(ctx)
	
	// Validate provider exists and is active
	providerObj, err := k.GetProvider(ctx, provider)
	if err != nil {
		return nil, types.ErrProviderNotFound
	}
	if !providerObj.IsActive() {
		return nil, types.ErrProviderInactive
	}

	// Check provider reputation
	if providerObj.ReputationScore < params.MinReputation {
		return nil, types.ErrReputationTooLow
	}

	// Validate duration
	if duration > int64(params.MaxJobDuration) {
		return nil, fmt.Errorf("duration exceeds maximum allowed: %d > %d", duration, params.MaxJobDuration)
	}

	// Calculate fee
	feeAmount := params.CalculateFee(amount.Amount)
	fee := sdk.NewCoin(amount.Denom, feeAmount)
	totalAmount := sdk.NewCoins(amount).Add(fee)

	// Check submitter has sufficient funds
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	balance := k.GetBalance(ctx, submitterAddr, amount.Denom)
	if balance.IsLT(totalAmount[0]) {
		return nil, types.ErrInsufficientFunds
	}

	// Generate escrow ID
	escrowID := generateEscrowID(jobID, submitter, provider)
	
	// Calculate deadline
	currentTime := ctx.BlockTime().Unix()
	deadline := currentTime + duration

	// Create escrow
	escrow := &types.EscrowExtended{
		ID:        escrowID,
		JobID:     jobID,
		Submitter: submitter,
		Provider:  provider,
		Amount:    amount,
		Fee:       fee,
		Status:    types.EscrowStatusLocked,
		Deadline:  deadline,
		CreatedAt: currentTime,
	}

	// Store escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return nil, err
	}

	// Transfer funds to module account
	if err := k.SendCoinsFromAccountToModule(ctx, submitterAddr, types.ModuleName, totalAmount); err != nil {
		return nil, fmt.Errorf("failed to lock escrow funds: %w", err)
	}

	// Update total escrowed
	totalEscrowed := k.GetTotalEscrowed(ctx)
	k.SetTotalEscrowed(ctx, totalEscrowed.Add(amount.Amount))

	// Create job match
	match := types.NewJobMatch(jobID, provider, providerObj.MatchScore(providerObj.Capabilities), providerObj.Capabilities)
	if err := k.SetJobMatch(ctx, match); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowCreated,
			sdk.NewAttribute(types.AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute(types.AttributeKeyJobID, jobID),
			sdk.NewAttribute(types.AttributeKeySubmitter, submitter),
			sdk.NewAttribute(types.AttributeKeyProvider, provider),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDeadline, fmt.Sprintf("%d", deadline)),
		),
	)

	return escrow, nil
}

// GetEscrow retrieves an escrow by ID
func (k Keeper) GetEscrow(ctx sdk.Context, escrowID string) (*types.EscrowExtended, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEscrowKey(escrowID)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrEscrowNotFound
	}

	var escrow types.EscrowExtended
	k.cdc.MustUnmarshal(bz, &escrow)
	return &escrow, nil
}

// GetEscrowByJob retrieves an escrow by job ID
func (k Keeper) GetEscrowByJob(ctx sdk.Context, jobID string) (*types.EscrowExtended, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEscrowByJobKey(jobID)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrEscrowNotFound
	}

	escrowID := string(bz)
	return k.GetEscrow(ctx, escrowID)
}

// SetEscrow stores an escrow
func (k Keeper) SetEscrow(ctx sdk.Context, escrow *types.EscrowExtended) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(escrow)
	store.Set(types.GetEscrowKey(escrow.ID), bz)
	
	// Index by job ID
	store.Set(types.GetEscrowByJobKey(escrow.JobID), []byte(escrow.ID))
	
	return nil
}

// ReleaseEscrow releases escrow funds to the provider
func (k Keeper) ReleaseEscrow(ctx sdk.Context, escrowID, caller string) error {
	escrow, err := k.GetEscrow(ctx, escrowID)
	if err != nil {
		return err
	}

	// Validate caller is the submitter
	if escrow.Submitter != caller {
		return types.ErrUnauthorizedEscrow
	}

	// Check escrow status
	if !escrow.IsLocked() {
		return types.ErrInvalidEscrowStatus
	}

	// Get provider address
	providerAddr, err := sdk.AccAddressFromBech32(escrow.Provider)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Transfer funds to provider
	amount := sdk.NewCoins(escrow.Amount)
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, amount); err != nil {
		return fmt.Errorf("failed to release escrow: %w", err)
	}

	// Handle fee (burn or send to treasury)
	feeCoins := sdk.NewCoins(escrow.Fee)
	if err := k.BurnCoins(ctx, types.ModuleName, feeCoins); err != nil {
		return fmt.Errorf("failed to burn fee: %w", err)
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusReleased
	escrow.ReleasedAt = ctx.BlockTime().Unix()
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Update provider stats
	provider, err := k.GetProvider(ctx, escrow.Provider)
	if err == nil {
		provider.CompletedJobs++
		// Increase reputation for successful job (up to 1000)
		if provider.ReputationScore < 1000 {
			provider.ReputationScore = min(provider.ReputationScore+10, 1000)
		}
		provider.LastActiveAt = ctx.BlockTime().Unix()
		_ = k.SetProvider(ctx, provider)
	}

	// Update total escrowed
	totalEscrowed := k.GetTotalEscrowed(ctx)
	k.SetTotalEscrowed(ctx, totalEscrowed.Sub(escrow.Amount.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowReleased,
			sdk.NewAttribute(types.AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute(types.AttributeKeyJobID, escrow.JobID),
			sdk.NewAttribute(types.AttributeKeyProvider, escrow.Provider),
			sdk.NewAttribute(types.AttributeKeyAmount, escrow.Amount.String()),
		),
	)

	return nil
}

// RefundEscrow refunds escrow funds to the submitter
func (k Keeper) RefundEscrow(ctx sdk.Context, escrowID, caller string) error {
	escrow, err := k.GetEscrow(ctx, escrowID)
	if err != nil {
		return err
	}

	// Validate caller is submitter or authorized
	if escrow.Submitter != caller {
		return types.ErrUnauthorizedEscrow
	}

	// Check escrow status
	if !escrow.IsLocked() {
		return types.ErrInvalidEscrowStatus
	}

	// Check if escrow has expired or caller is submitter
	currentTime := ctx.BlockTime().Unix()
	if !escrow.IsExpired(currentTime) && escrow.Submitter != caller {
		return types.ErrEscrowLocked
	}

	// Get submitter address
	submitterAddr, err := sdk.AccAddressFromBech32(escrow.Submitter)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Refund principal + fee
	totalAmount := sdk.NewCoins(escrow.Amount).Add(escrow.Fee)
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, submitterAddr, totalAmount); err != nil {
		return fmt.Errorf("failed to refund escrow: %w", err)
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusRefunded
	escrow.RefundedAt = currentTime
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Update total escrowed
	totalEscrowed := k.GetTotalEscrowed(ctx)
	k.SetTotalEscrowed(ctx, totalEscrowed.Sub(escrow.Amount.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowRefunded,
			sdk.NewAttribute(types.AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute(types.AttributeKeyJobID, escrow.JobID),
			sdk.NewAttribute(types.AttributeKeySubmitter, escrow.Submitter),
			sdk.NewAttribute(types.AttributeKeyAmount, totalAmount.String()),
		),
	)

	return nil
}

// DisputeEscrow marks an escrow as disputed
func (k Keeper) DisputeEscrow(ctx sdk.Context, escrowID, disputer, reason string, evidence []byte) (*types.Dispute, error) {
	escrow, err := k.GetEscrow(ctx, escrowID)
	if err != nil {
		return nil, err
	}

	// Validate disputer is either submitter or provider
	if escrow.Submitter != disputer && escrow.Provider != disputer {
		return nil, types.ErrUnauthorizedEscrow
	}

	// Check escrow status
	if !escrow.IsLocked() {
		return nil, types.ErrInvalidEscrowStatus
	}

	// Check dispute period hasn't expired
	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime().Unix()
	if escrow.IsExpired(currentTime - int64(params.DisputePeriod)) {
		return nil, types.ErrDisputePeriodExpired
	}

	// Check if already disputed
	if escrow.IsDisputed() {
		return nil, types.ErrEscrowDisputed
	}

	// Create dispute
	disputeID := generateDisputeID(escrowID)
	dispute := types.NewDispute(disputeID, escrowID, escrow.JobID, escrow.Submitter, escrow.Provider, reason, evidence)
	dispute.OpenedAt = currentTime

	// Store dispute
	if err := k.SetDispute(ctx, dispute); err != nil {
		return nil, err
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusDisputed
	escrow.DisputeID = disputeID
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowDisputed,
			sdk.NewAttribute(types.AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute(types.AttributeKeyJobID, escrow.JobID),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDisputeOpened,
			sdk.NewAttribute(types.AttributeKeyEscrowID, escrowID),
			sdk.NewAttribute("dispute_id", disputeID),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	return dispute, nil
}

// GetEscrowsByProvider retrieves all escrows for a provider
func (k Keeper) GetEscrowsByProvider(ctx sdk.Context, provider string) ([]types.EscrowExtended, error) {
	store := ctx.KVStore(k.storeKey)
	providerAddr := sdk.MustAccAddressFromBech32(provider)
	prefix := types.GetEscrowByProviderKey(providerAddr, "")
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var escrows []types.EscrowExtended
	for ; iter.Valid(); iter.Next() {
		escrowID := string(iter.Value())
		escrow, err := k.GetEscrow(ctx, escrowID)
		if err != nil {
			continue
		}
		escrows = append(escrows, *escrow)
	}
	return escrows, nil
}

// Helper functions

func generateEscrowID(jobID, submitter, provider string) string {
	data := fmt.Sprintf("%s:%s:%s:%d", jobID, submitter, provider, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func generateDisputeID(escrowID string) string {
	data := fmt.Sprintf("%s:%d", escrowID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
