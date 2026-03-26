package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// OpenDispute opens a new dispute for an escrow
func (k Keeper) OpenDispute(ctx sdk.Context, escrowID, disputer, reason string, evidence []byte) (*types.Dispute, error) {
	// Use the DisputeEscrow method from escrow.go
	return k.DisputeEscrow(ctx, escrowID, disputer, reason, evidence)
}

// GetDispute retrieves a dispute by ID
func (k Keeper) GetDispute(ctx sdk.Context, disputeID string) (*types.Dispute, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDisputeKey(disputeID)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrDisputeNotFound
	}

	var dispute types.Dispute
	k.cdc.MustUnmarshal(bz, &dispute)
	return &dispute, nil
}

// GetDisputeByEscrow retrieves a dispute by escrow ID
func (k Keeper) GetDisputeByEscrow(ctx sdk.Context, escrowID string) (*types.Dispute, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDisputeByEscrowKey(escrowID)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrDisputeNotFound
	}

	disputeID := string(bz)
	return k.GetDispute(ctx, disputeID)
}

// SetDispute stores a dispute
func (k Keeper) SetDispute(ctx sdk.Context, dispute *types.Dispute) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(dispute)
	store.Set(types.GetDisputeKey(dispute.ID), bz)
	
	// Index by escrow ID
	store.Set(types.GetDisputeByEscrowKey(dispute.EscrowID), []byte(dispute.ID))
	
	return nil
}

// ResolveDispute resolves a dispute and distributes funds
func (k Keeper) ResolveDispute(ctx sdk.Context, disputeID, resolver, resolution, winner string) error {
	dispute, err := k.GetDispute(ctx, disputeID)
	if err != nil {
		return err
	}

	// Check dispute is not already resolved
	if dispute.IsResolved() {
		return types.ErrDisputeAlreadyResolved
	}

	// Validate resolution
	if resolution != types.ResolutionSubmitterWins && 
	   resolution != types.ResolutionProviderWins && 
	   resolution != types.ResolutionSplit {
		return types.ErrInvalidParams
	}

	// Validate winner is either submitter or provider
	if winner != dispute.Submitter && winner != dispute.Provider {
		return types.ErrInvalidAddress
	}

	// Get associated escrow
	escrow, err := k.GetEscrow(ctx, dispute.EscrowID)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Process resolution
	switch resolution {
	case types.ResolutionSubmitterWins:
		// Refund to submitter, slash provider
		if err := k.processSubmitterWin(ctx, escrow, dispute); err != nil {
			return err
		}

	case types.ResolutionProviderWins:
		// Release to provider
		if err := k.processProviderWin(ctx, escrow, dispute); err != nil {
			return err
		}

	case types.ResolutionSplit:
		// Split funds between both parties
		if err := k.processSplitResolution(ctx, escrow, dispute); err != nil {
			return err
		}
	}

	// Update dispute status
	dispute.Status = types.DisputeStatusResolved
	dispute.Resolution = resolution
	dispute.Winner = winner
	dispute.ResolvedAt = ctx.BlockTime().Unix()
	if err := k.SetDispute(ctx, dispute); err != nil {
		return err
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusReleased
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDisputeResolved,
			sdk.NewAttribute("dispute_id", disputeID),
			sdk.NewAttribute(types.AttributeKeyEscrowID, dispute.EscrowID),
			sdk.NewAttribute(types.AttributeKeyResolution, resolution),
			sdk.NewAttribute(types.AttributeKeyWinner, winner),
		),
	)

	return nil
}

// processSubmitterWin handles resolution where submitter wins
func (k Keeper) processSubmitterWin(ctx sdk.Context, escrow *types.EscrowExtended, dispute *types.Dispute) error {
	// Refund principal to submitter
	submitterAddr, err := sdk.AccAddressFromBech32(escrow.Submitter)
	if err != nil {
		return err
	}

	totalAmount := sdk.NewCoins(escrow.Amount).Add(escrow.Fee)
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, submitterAddr, totalAmount); err != nil {
		return fmt.Errorf("failed to refund submitter: %w", err)
	}

	// Slash provider
	slashAmount := escrow.Amount
	reason := dispute.Reason
	if reason == "" {
		reason = types.ReasonJobFailed
	}
	if err := k.SlashProvider(ctx, escrow.Provider, slashAmount, reason); err != nil {
		// Log slashing failure but don't fail the resolution
		k.Logger(ctx).Error("failed to slash provider", "provider", escrow.Provider, "error", err)
	}

	return nil
}

// processProviderWin handles resolution where provider wins
func (k Keeper) processProviderWin(ctx sdk.Context, escrow *types.EscrowExtended, dispute *types.Dispute) error {
	// Release to provider
	providerAddr, err := sdk.AccAddressFromBech32(escrow.Provider)
	if err != nil {
		return err
	}

	// Release amount to provider
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, sdk.NewCoins(escrow.Amount)); err != nil {
		return fmt.Errorf("failed to pay provider: %w", err)
	}

	// Burn fee
	if err := k.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(escrow.Fee)); err != nil {
		return fmt.Errorf("failed to burn fee: %w", err)
	}

	// Boost provider reputation for winning dispute
	if err := k.UpdateReputation(ctx, escrow.Provider, 20); err != nil {
		k.Logger(ctx).Error("failed to boost reputation", "provider", escrow.Provider, "error", err)
	}

	// Penalize submitter (reduce their ability to open frivolous disputes)
	// This could be tracked separately

	return nil
}

// processSplitResolution handles split resolution
func (k Keeper) processSplitResolution(ctx sdk.Context, escrow *types.EscrowExtended, dispute *types.Dispute) error {
	// Calculate split (50/50 by default, could be configurable)
	amount := escrow.Amount.Amount
	halfAmount := amount.Quo(sdk.NewInt(2))
	
	// Refund half to submitter
	submitterAddr, err := sdk.AccAddressFromBech32(escrow.Submitter)
	if err != nil {
		return err
	}
	submitterAmount := sdk.NewCoin(escrow.Amount.Denom, halfAmount)
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, submitterAddr, sdk.NewCoins(submitterAmount)); err != nil {
		return fmt.Errorf("failed to refund submitter: %w", err)
	}

	// Pay half to provider
	providerAddr, err := sdk.AccAddressFromBech32(escrow.Provider)
	if err != nil {
		return err
	}
	providerAmount := sdk.NewCoin(escrow.Amount.Denom, amount.Sub(halfAmount))
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, sdk.NewCoins(providerAmount)); err != nil {
		return fmt.Errorf("failed to pay provider: %w", err)
	}

	// Burn fee
	if err := k.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(escrow.Fee)); err != nil {
		return fmt.Errorf("failed to burn fee: %w", err)
	}

	// Small reputation penalty to both parties
	_ = k.UpdateReputation(ctx, escrow.Provider, -10)

	return nil
}

// ExpireDispute marks a dispute as expired if dispute period has passed
func (k Keeper) ExpireDispute(ctx sdk.Context, disputeID string) error {
	dispute, err := k.GetDispute(ctx, disputeID)
	if err != nil {
		return err
	}

	if dispute.IsResolved() {
		return nil // Already resolved
	}

	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime().Unix()
	disputeExpiryBlocks := int64(params.DisputePeriod)

	// Check if dispute has expired
	if dispute.OpenedAt+disputeExpiryBlocks > currentTime {
		return nil // Not expired yet
	}

	// Mark as expired
	dispute.Status = types.DisputeStatusExpired
	if err := k.SetDispute(ctx, dispute); err != nil {
		return err
	}

	// Auto-resolve in favor of provider (job completed, no valid dispute)
	escrow, err := k.GetEscrow(ctx, dispute.EscrowID)
	if err != nil {
		return err
	}

	// Release to provider
	return k.processProviderWin(ctx, escrow, dispute)
}

// GetOpenDisputes retrieves all open disputes
func (k Keeper) GetOpenDisputes(ctx sdk.Context) ([]types.Dispute, error) {
	disputes, err := k.GetAllDisputes(ctx)
	if err != nil {
		return nil, err
	}

	var openDisputes []types.Dispute
	for _, dispute := range disputes {
		if dispute.Status == types.DisputeStatusOpen {
			openDisputes = append(openDisputes, dispute)
		}
	}
	return openDisputes, nil
}

// GetDisputesByProvider retrieves all disputes for a provider
func (k Keeper) GetDisputesByProvider(ctx sdk.Context, provider string) ([]types.Dispute, error) {
	disputes, err := k.GetAllDisputes(ctx)
	if err != nil {
		return nil, err
	}

	var providerDisputes []types.Dispute
	for _, dispute := range disputes {
		if dispute.Provider == provider {
			providerDisputes = append(providerDisputes, dispute)
		}
	}
	return providerDisputes, nil
}

// GetDisputesBySubmitter retrieves all disputes by a submitter
func (k Keeper) GetDisputesBySubmitter(ctx sdk.Context, submitter string) ([]types.Dispute, error) {
	disputes, err := k.GetAllDisputes(ctx)
	if err != nil {
		return nil, err
	}

	var submitterDisputes []types.Dispute
	for _, dispute := range disputes {
		if dispute.Submitter == submitter {
			submitterDisputes = append(submitterDisputes, dispute)
		}
	}
	return submitterDisputes, nil
}
