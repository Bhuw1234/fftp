package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// RegisterProvider registers a new compute provider
func (k Keeper) RegisterProvider(ctx sdk.Context, address string, stake sdk.Coin, capabilities types.ProviderCapabilities) (*types.ProviderExtended, error) {
	// Check if provider already exists
	_, err := k.GetProvider(ctx, address)
	if err == nil {
		return nil, types.ErrProviderAlreadyExists
	}

	// Validate stake meets minimum
	params := k.GetParams(ctx)
	if stake.IsLT(params.MinStake) {
		return nil, types.ErrInsufficientStake
	}

	// Get provider address
	providerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Check provider has sufficient balance
	balance := k.GetBalance(ctx, providerAddr, stake.Denom)
	if balance.IsLT(stake) {
		return nil, types.ErrInsufficientFunds
	}

	// Transfer stake to module account
	if err := k.SendCoinsFromAccountToModule(ctx, providerAddr, types.ModuleName, sdk.NewCoins(stake)); err != nil {
		return nil, fmt.Errorf("failed to lock stake: %w", err)
	}

	// Create provider
	provider := types.NewProviderExtended(address, stake, capabilities)
	provider.RegisteredAt = ctx.BlockTime().Unix()
	provider.LastActiveAt = provider.RegisteredAt

	// Store provider
	if err := k.SetProvider(ctx, provider); err != nil {
		return nil, err
	}

	// Add to active providers
	if err := k.AddActiveProvider(ctx, address); err != nil {
		return nil, err
	}

	// Update total staked
	totalStaked := k.GetTotalStaked(ctx)
	k.SetTotalStaked(ctx, totalStaked.Add(stake.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderRegistered,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
			sdk.NewAttribute(types.AttributeKeyStake, stake.String()),
			sdk.NewAttribute(types.AttributeKeyCapabilities, string(capabilities.ToBytes())),
		),
	)

	return provider, nil
}

// GetProvider retrieves a provider by address
func (k Keeper) GetProvider(ctx sdk.Context, address string) (*types.ProviderExtended, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetProviderKey(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrProviderNotFound
	}

	var provider types.ProviderExtended
	k.cdc.MustUnmarshal(bz, &provider)
	return &provider, nil
}

// SetProvider stores a provider
func (k Keeper) SetProvider(ctx sdk.Context, provider *types.ProviderExtended) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(provider)
	store.Set(types.GetProviderKey(provider.Address), bz)

	// Also index by reputation for efficient matching
	reputationKey := types.GetProviderByReputationKey(provider.ReputationScore, provider.Address)
	store.Set(reputationKey, []byte(provider.Address))

	return nil
}

// UnregisterProvider unregisters a compute provider
func (k Keeper) UnregisterProvider(ctx sdk.Context, address string) error {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return err
	}

	// Check for pending escrows
	escrows, err := k.GetEscrowsByProvider(ctx, address)
	if err != nil {
		return err
	}
	for _, escrow := range escrows {
		if escrow.IsLocked() {
			return fmt.Errorf("cannot unregister provider with active escrows")
		}
	}

	// Get provider address
	providerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Return staked amount
	stake := sdk.NewCoins(provider.StakedAmount)
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, stake); err != nil {
		return fmt.Errorf("failed to return stake: %w", err)
	}

	// Update total staked
	totalStaked := k.GetTotalStaked(ctx)
	k.SetTotalStaked(ctx, totalStaked.Sub(provider.StakedAmount.Amount))

	// Remove from active providers
	if err := k.RemoveActiveProvider(ctx, address); err != nil {
		return err
	}

	// Delete provider
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetProviderKey(address))
	store.Delete(types.GetProviderByReputationKey(provider.ReputationScore, address))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderUnregistered,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
		),
	)

	return nil
}

// StakeProvider adds more stake to a provider
func (k Keeper) StakeProvider(ctx sdk.Context, address string, amount sdk.Coin) (*types.ProviderExtended, error) {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return nil, err
	}

	// Get provider address
	providerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Check provider has sufficient balance
	balance := k.GetBalance(ctx, providerAddr, amount.Denom)
	if balance.IsLT(amount) {
		return nil, types.ErrInsufficientFunds
	}

	// Transfer stake to module account
	if err := k.SendCoinsFromAccountToModule(ctx, providerAddr, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return nil, fmt.Errorf("failed to lock stake: %w", err)
	}

	// Update provider stake
	oldStake := provider.StakedAmount
	provider.StakedAmount = provider.StakedAmount.Add(amount)

	// Remove old reputation index
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetProviderByReputationKey(provider.ReputationScore, address))

	// Boost reputation for additional stake (small bonus)
	if provider.ReputationScore < 900 {
		provider.ReputationScore = min(provider.ReputationScore+5, 900)
	}

	// Store updated provider
	if err := k.SetProvider(ctx, provider); err != nil {
		return nil, err
	}

	// Update total staked
	totalStaked := k.GetTotalStaked(ctx)
	k.SetTotalStaked(ctx, totalStaked.Add(amount.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderStaked,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
			sdk.NewAttribute(types.AttributeKeyStake, amount.String()),
			sdk.NewAttribute("total_stake", provider.StakedAmount.String()),
		),
	)

	return provider, nil
}

// UnstakeProvider removes stake from a provider
func (k Keeper) UnstakeProvider(ctx sdk.Context, address string, amount sdk.Coin) (*types.ProviderExtended, error) {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)

	// Check unstake amount is valid
	if amount.IsGT(provider.StakedAmount) {
		return nil, types.ErrInsufficientStake
	}

	// Check remaining stake meets minimum
	remainingStake := provider.StakedAmount.Sub(amount)
	if remainingStake.IsLT(params.MinStake) && remainingStake.IsPositive() {
		return nil, fmt.Errorf("remaining stake would be below minimum: %s < %s", remainingStake.String(), params.MinStake.String())
	}

	// Get provider address
	providerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Return unstaked amount
	if err := k.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, sdk.NewCoins(amount)); err != nil {
		return nil, fmt.Errorf("failed to return stake: %w", err)
	}

	// Remove old reputation index
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetProviderByReputationKey(provider.ReputationScore, address))

	// Update provider stake
	provider.StakedAmount = remainingStake

	// If stake is zero, mark as inactive
	if remainingStake.IsZero() {
		provider.Status = types.ProviderStatusInactive
		if err := k.RemoveActiveProvider(ctx, address); err != nil {
			return nil, err
		}
	}

	// Store updated provider
	if err := k.SetProvider(ctx, provider); err != nil {
		return nil, err
	}

	// Update total staked
	totalStaked := k.GetTotalStaked(ctx)
	k.SetTotalStaked(ctx, totalStaked.Sub(amount.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderUnstaked,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
			sdk.NewAttribute(types.AttributeKeyStake, amount.String()),
			sdk.NewAttribute("remaining_stake", remainingStake.String()),
		),
	)

	return provider, nil
}

// SlashProvider slashes a provider's stake
func (k Keeper) SlashProvider(ctx sdk.Context, address string, slashAmount sdk.Coin, reason string) error {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return err
	}

	// Calculate slash amount
	params := k.GetParams(ctx)
	maxSlash := params.CalculateSlash(provider.StakedAmount.Amount, true) // true = dispute slash
	if slashAmount.Amount.GT(maxSlash) {
		slashAmount = sdk.NewCoin(slashAmount.Denom, maxSlash)
	}

	// Don't slash more than staked
	if slashAmount.IsGT(provider.StakedAmount) {
		slashAmount = provider.StakedAmount
	}

	// Burn slashed amount
	if err := k.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(slashAmount)); err != nil {
		return fmt.Errorf("failed to slash: %w", err)
	}

	// Remove old reputation index
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetProviderByReputationKey(provider.ReputationScore, address))

	// Update provider
	provider.StakedAmount = provider.StakedAmount.Sub(slashAmount)
	provider.SlashedCount++
	provider.FailedJobs++
	
	// Reduce reputation
	reputationPenalty := uint32(50) // Base penalty
	if reason == types.ReasonMalicious {
		reputationPenalty = 200 // Higher penalty for malicious behavior
	}
	if provider.ReputationScore > reputationPenalty {
		provider.ReputationScore -= reputationPenalty
	} else {
		provider.ReputationScore = 0
	}

	// If stake below minimum or reputation too low, mark as slashed
	if provider.StakedAmount.IsLT(params.MinStake) || provider.ReputationScore < params.MinReputation {
		provider.Status = types.ProviderStatusSlashed
		_ = k.RemoveActiveProvider(ctx, address)
	}

	// Store updated provider
	if err := k.SetProvider(ctx, provider); err != nil {
		return err
	}

	// Update total staked
	totalStaked := k.GetTotalStaked(ctx)
	k.SetTotalStaked(ctx, totalStaked.Sub(slashAmount.Amount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProviderSlashed,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
			sdk.NewAttribute(types.AttributeKeySlashedAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyNewReputation, fmt.Sprintf("%d", provider.ReputationScore)),
		),
	)

	return nil
}

// UpdateReputation updates a provider's reputation score
func (k Keeper) UpdateReputation(ctx sdk.Context, address string, delta int32) error {
	provider, err := k.GetProvider(ctx, address)
	if err != nil {
		return err
	}

	// Remove old reputation index
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetProviderByReputationKey(provider.ReputationScore, address))

	oldReputation := provider.ReputationScore

	// Update reputation (capped at 0-1000)
	newReputation := int32(provider.ReputationScore) + delta
	if newReputation < 0 {
		newReputation = 0
	}
	if newReputation > 1000 {
		newReputation = 1000
	}
	provider.ReputationScore = uint32(newReputation)

	// Store updated provider
	if err := k.SetProvider(ctx, provider); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReputationUpdated,
			sdk.NewAttribute(types.AttributeKeyProvider, address),
			sdk.NewAttribute(types.AttributeKeyOldReputation, fmt.Sprintf("%d", oldReputation)),
			sdk.NewAttribute(types.AttributeKeyNewReputation, fmt.Sprintf("%d", provider.ReputationScore)),
		),
	)

	return nil
}

// AddActiveProvider adds a provider to the active providers list
func (k Keeper) AddActiveProvider(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetActiveProvidersKey(), []byte(address)...)
	store.Set(key, []byte{1})
	return nil
}

// RemoveActiveProvider removes a provider from the active providers list
func (k Keeper) RemoveActiveProvider(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := append(types.GetActiveProvidersKey(), []byte(address)...)
	store.Delete(key)
	return nil
}

// GetActiveProviders returns all active providers
func (k Keeper) GetActiveProviders(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetActiveProvidersKey()
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	var providers []string
	for ; iter.Valid(); iter.Next() {
		// Extract address from key (skip prefix)
		address := string(iter.Key()[len(prefix):])
		providers = append(providers, address)
	}
	return providers
}
