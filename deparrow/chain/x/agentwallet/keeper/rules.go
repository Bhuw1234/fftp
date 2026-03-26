package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// GetSpendingRules retrieves all spending rules for a wallet
func (k Keeper) GetSpendingRules(ctx sdk.Context, address string) ([]types.SpendingRuleExtended, error) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := append(types.SpendingRuleKey, []byte(address)...)
	iter := store.Iterator(keyPrefix, append(keyPrefix, 0xFF))
	defer iter.Close()

	var rules []types.SpendingRuleExtended
	for ; iter.Valid(); iter.Next() {
		var rule types.SpendingRuleExtended
		if err := json.Unmarshal(iter.Value(), &rule); err == nil {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

// GetSpendingRule retrieves a specific spending rule
func (k Keeper) GetSpendingRule(ctx sdk.Context, address string, ruleIndex uint32) (*types.SpendingRuleExtended, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetSpendingRuleKey(address, ruleIndex)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var rule types.SpendingRuleExtended
	if err := json.Unmarshal(bz, &rule); err != nil {
		return nil, false
	}
	return &rule, true
}

// AddSpendingRule adds a new spending rule to a wallet
func (k Keeper) AddSpendingRule(ctx sdk.Context, address string, rule types.SpendingRule) (uint32, error) {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return 0, types.ErrWalletNotFound
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return 0, types.ErrInvalidRule
	}

	// Check rule limit
	params := k.GetParams(ctx)
	currentRules, _ := k.GetSpendingRules(ctx, address)
	if uint32(len(currentRules)) >= params.MaxRulesPerWallet {
		return 0, types.ErrRuleLimitExceeded
	}

	// Determine rule index
	ruleIndex := uint32(len(currentRules))

	// Create extended rule
	extendedRule := types.SpendingRuleExtended{
		SpendingRule: rule,
		ID:           ruleIndex,
		CreatedAt:    ctx.BlockTime().Unix(),
		UpdatedAt:    ctx.BlockTime().Unix(),
	}

	// Store rule using JSON marshal
	store := ctx.KVStore(k.storeKey)
	key := types.GetSpendingRuleKey(address, ruleIndex)
	bz, err := json.Marshal(&extendedRule)
	if err != nil {
		return 0, err
	}
	store.Set(key, bz)

	// Update wallet with new rule in memory
	wallet.SpendingRules = append(wallet.SpendingRules, rule)
	if err := k.SetWallet(ctx, wallet); err != nil {
		return 0, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSpendingRuleAdded,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
			sdk.NewAttribute(types.AttributeKeyMaxPerTx, rule.MaxPerTx.String()),
			sdk.NewAttribute(types.AttributeKeyDailyBudget, rule.DailyBudget.String()),
		),
	)

	return ruleIndex, nil
}

// UpdateSpendingRule updates an existing spending rule
func (k Keeper) UpdateSpendingRule(ctx sdk.Context, address string, ruleIndex uint32, rule types.SpendingRule) error {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return types.ErrInvalidRule
	}

	// Check if rule exists
	existingRule, exists := k.GetSpendingRule(ctx, address, ruleIndex)
	if !exists {
		return types.ErrInvalidRule
	}

	// Update extended rule
	existingRule.SpendingRule = rule
	existingRule.UpdatedAt = ctx.BlockTime().Unix()

	// Store updated rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetSpendingRuleKey(address, ruleIndex)
	bz, err := json.Marshal(existingRule)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Update wallet rules in memory
	if int(ruleIndex) < len(wallet.SpendingRules) {
		wallet.SpendingRules[ruleIndex] = rule
		if err := k.SetWallet(ctx, wallet); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSpendingRuleUpdated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
		),
	)

	return nil
}

// RemoveSpendingRule removes a spending rule from a wallet
func (k Keeper) RemoveSpendingRule(ctx sdk.Context, address string, ruleIndex uint32) error {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Check if rule exists
	if _, exists := k.GetSpendingRule(ctx, address, ruleIndex); !exists {
		return types.ErrInvalidRule
	}

	// Delete rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetSpendingRuleKey(address, ruleIndex)
	store.Delete(key)

	// Update wallet rules in memory (remove the rule)
	if int(ruleIndex) < len(wallet.SpendingRules) {
		wallet.SpendingRules = append(
			wallet.SpendingRules[:ruleIndex],
			wallet.SpendingRules[ruleIndex+1:]...,
		)
		if err := k.SetWallet(ctx, wallet); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSpendingRuleRemoved,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
		),
	)

	return nil
}

// GetDailySpending retrieves the daily spending for a wallet
func (k Keeper) GetDailySpending(ctx sdk.Context, address string) (sdk.Coin, error) {
	store := ctx.KVStore(k.storeKey)
	day := k.GetCurrentDay(ctx)
	key := types.GetDailySpendingKey(address, day)

	bz := store.Get(key)
	if bz == nil {
		return sdk.NewCoin("dpc", sdk.ZeroInt()), nil
	}

	var dailySpent sdk.Coin
	if err := json.Unmarshal(bz, &dailySpent); err != nil {
		return sdk.NewCoin("dpc", sdk.ZeroInt()), nil
	}
	return dailySpent, nil
}

// BlockOperation blocks an operation for a wallet
func (k Keeper) BlockOperation(ctx sdk.Context, address string, operation string) error {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Add to blocked operations in the first rule
	if len(wallet.SpendingRules) > 0 {
		// Check if already blocked
		for _, blocked := range wallet.SpendingRules[0].BlockedOps {
			if blocked == operation {
				return nil // Already blocked
			}
		}
		wallet.SpendingRules[0].BlockedOps = append(wallet.SpendingRules[0].BlockedOps, operation)
	}

	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOperationBlocked,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyOperation, operation),
		),
	)

	return nil
}

// AllowOperation adds an operation to the allowed list
func (k Keeper) AllowOperation(ctx sdk.Context, address string, operation string) error {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Add to allowed operations in the first rule
	if len(wallet.SpendingRules) > 0 {
		// Check if already allowed
		for _, allowed := range wallet.SpendingRules[0].AllowedOps {
			if allowed == operation {
				return nil // Already allowed
			}
		}
		wallet.SpendingRules[0].AllowedOps = append(wallet.SpendingRules[0].AllowedOps, operation)
	}

	return k.SetWallet(ctx, wallet)
}

// DisallowOperation removes an operation from the allowed list
func (k Keeper) DisallowOperation(ctx sdk.Context, address string, operation string) error {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Remove from allowed operations in the first rule
	if len(wallet.SpendingRules) > 0 {
		allowedOps := wallet.SpendingRules[0].AllowedOps
		for i, allowed := range allowedOps {
			if allowed == operation {
				wallet.SpendingRules[0].AllowedOps = append(
					allowedOps[:i],
					allowedOps[i+1:]...,
				)
				break
			}
		}
	}

	return k.SetWallet(ctx, wallet)
}

// SetDailyBudget sets the daily budget for a wallet
func (k Keeper) SetDailyBudget(ctx sdk.Context, address string, budget sdk.Coin) error {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Update in the first rule
	if len(wallet.SpendingRules) > 0 {
		wallet.SpendingRules[0].DailyBudget = budget
	}

	return k.SetWallet(ctx, wallet)
}

// SetMaxPerTx sets the max per transaction limit for a wallet
func (k Keeper) SetMaxPerTx(ctx sdk.Context, address string, maxPerTx sdk.Coin) error {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Update in the first rule
	if len(wallet.SpendingRules) > 0 {
		wallet.SpendingRules[0].MaxPerTx = maxPerTx
	}

	return k.SetWallet(ctx, wallet)
}
