package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// GetAutomationRules retrieves all automation rules for a wallet
func (k Keeper) GetAutomationRules(ctx sdk.Context, address string) ([]types.AutomationRuleExtended, error) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := append(types.AutomationRuleKey, []byte(address)...)
	iter := store.Iterator(keyPrefix, append(keyPrefix, 0xFF))
	defer iter.Close()

	var rules []types.AutomationRuleExtended
	for ; iter.Valid(); iter.Next() {
		var rule types.AutomationRuleExtended
		if err := json.Unmarshal(iter.Value(), &rule); err == nil {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

// GetAutomationRule retrieves a specific automation rule
func (k Keeper) GetAutomationRule(ctx sdk.Context, address string, ruleIndex uint32) (*types.AutomationRuleExtended, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var rule types.AutomationRuleExtended
	if err := json.Unmarshal(bz, &rule); err != nil {
		return nil, false
	}
	return &rule, true
}

// AddAutomationRule adds a new automation rule to a wallet
func (k Keeper) AddAutomationRule(ctx sdk.Context, address string, rule types.AutomationRule) (uint32, error) {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return 0, types.ErrWalletNotFound
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return 0, types.ErrInvalidAutomationRule
	}

	// Check rule limit
	params := k.GetParams(ctx)
	currentRules, _ := k.GetAutomationRules(ctx, address)
	if uint32(len(currentRules)) >= params.MaxRulesPerWallet {
		return 0, types.ErrRuleLimitExceeded
	}

	// Determine rule index
	ruleIndex := uint32(len(currentRules))

	// Create extended rule
	extendedRule := types.AutomationRuleExtended{
		AutomationRule: rule,
		ID:             ruleIndex,
		CreatedAt:      ctx.BlockTime().Unix(),
		UpdatedAt:      ctx.BlockTime().Unix(),
		TriggerCount:   0,
	}

	// Store rule using JSON marshal
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	bz, err := json.Marshal(&extendedRule)
	if err != nil {
		return 0, err
	}
	store.Set(key, bz)

	// Update wallet with new rule in memory
	wallet.AutomationRules = append(wallet.AutomationRules, rule)
	if err := k.SetWallet(ctx, wallet); err != nil {
		return 0, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationRuleAdded,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
		),
	)

	return ruleIndex, nil
}

// UpdateAutomationRule updates an existing automation rule
func (k Keeper) UpdateAutomationRule(ctx sdk.Context, address string, ruleIndex uint32, rule types.AutomationRule) error {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Validate rule
	if err := rule.Validate(); err != nil {
		return types.ErrInvalidAutomationRule
	}

	// Check if rule exists
	existingRule, exists := k.GetAutomationRule(ctx, address, ruleIndex)
	if !exists {
		return types.ErrInvalidAutomationRule
	}

	// Update extended rule
	existingRule.AutomationRule = rule
	existingRule.UpdatedAt = ctx.BlockTime().Unix()

	// Store updated rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	bz, err := json.Marshal(existingRule)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Update wallet rules in memory
	if int(ruleIndex) < len(wallet.AutomationRules) {
		wallet.AutomationRules[ruleIndex] = rule
		if err := k.SetWallet(ctx, wallet); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationRuleUpdated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
		),
	)

	return nil
}

// RemoveAutomationRule removes an automation rule from a wallet
func (k Keeper) RemoveAutomationRule(ctx sdk.Context, address string, ruleIndex uint32) error {
	// Verify wallet exists
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return types.ErrWalletNotFound
	}

	// Check if rule exists
	if _, exists := k.GetAutomationRule(ctx, address, ruleIndex); !exists {
		return types.ErrInvalidAutomationRule
	}

	// Delete rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	store.Delete(key)

	// Update wallet rules in memory
	if int(ruleIndex) < len(wallet.AutomationRules) {
		wallet.AutomationRules = append(
			wallet.AutomationRules[:ruleIndex],
			wallet.AutomationRules[ruleIndex+1:]...,
		)
		if err := k.SetWallet(ctx, wallet); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationRuleRemoved,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRuleIndex, string(rune(ruleIndex))),
		),
	)

	return nil
}

// EnableAutomationRule enables an automation rule
func (k Keeper) EnableAutomationRule(ctx sdk.Context, address string, ruleIndex uint32) error {
	existingRule, exists := k.GetAutomationRule(ctx, address, ruleIndex)
	if !exists {
		return types.ErrInvalidAutomationRule
	}

	existingRule.Enabled = true
	existingRule.UpdatedAt = ctx.BlockTime().Unix()

	// Store updated rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	bz, err := json.Marshal(existingRule)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

// DisableAutomationRule disables an automation rule
func (k Keeper) DisableAutomationRule(ctx sdk.Context, address string, ruleIndex uint32) error {
	existingRule, exists := k.GetAutomationRule(ctx, address, ruleIndex)
	if !exists {
		return types.ErrInvalidAutomationRule
	}

	existingRule.Enabled = false
	existingRule.UpdatedAt = ctx.BlockTime().Unix()

	// Store updated rule
	store := ctx.KVStore(k.storeKey)
	key := types.GetAutomationRuleKey(address, ruleIndex)
	bz, err := json.Marshal(existingRule)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

// CheckAndTriggerAutomations checks all automation rules and triggers appropriate actions
// This should be called during EndBlocker
func (k Keeper) CheckAndTriggerAutomations(ctx sdk.Context) error {
	wallets, err := k.GetAllWallets(ctx)
	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		rules, err := k.GetAutomationRules(ctx, wallet.Address)
		if err != nil {
			continue
		}

		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}

			if rule.ShouldTrigger(&wallet) {
				if err := k.executeAutomation(ctx, &wallet, &rule); err != nil {
					// Log error but continue with other rules
					continue
				}

				// Update trigger count and last trigger time
				rule.TriggerCount++
				rule.LastTrigger = ctx.BlockTime().Unix()

				// Store updated rule
				store := ctx.KVStore(k.storeKey)
				key := types.GetAutomationRuleKey(wallet.Address, rule.ID)
				bz, _ := json.Marshal(&rule)
				store.Set(key, bz)
			}
		}
	}

	return nil
}

// executeAutomation executes an automation action
func (k Keeper) executeAutomation(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	switch rule.Action {
	case types.ActionBuyCompute:
		return k.executeBuyCompute(ctx, wallet, rule)

	case types.ActionTransferToReserve:
		return k.executeTransferToReserve(ctx, wallet, rule)

	case types.ActionNotify:
		return k.executeNotify(ctx, wallet, rule)

	case types.ActionPauseOperations:
		return k.executePauseOperations(ctx, wallet, rule)

	case types.ActionResumeOperations:
		return k.executeResumeOperations(ctx, wallet, rule)

	case types.ActionRequestFunding:
		return k.executeRequestFunding(ctx, wallet, rule)

	default:
		return types.ErrInvalidAction
	}
}

// executeBuyCompute executes the buy_compute automation action
func (k Keeper) executeBuyCompute(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	amount := rule.Amount
	if amount.IsZero() {
		// Use default amount based on trigger
		amount = sdk.NewCoin("dpc", math.NewInt(100000000000000000)) // 0.1 DPC
	}

	// Check if we have enough balance
	availableBalance := wallet.GetAvailableBalance()
	if availableBalance.Amount.LT(amount.Amount) {
		return types.ErrInsufficientFunds
	}

	// Emit event for external processing
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

// executeTransferToReserve executes the transfer_to_reserve automation action
func (k Keeper) executeTransferToReserve(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	amount := rule.Amount
	if amount.IsZero() {
		amount = sdk.NewCoin("dpc", math.NewInt(50000000000000000)) // 0.05 DPC default
	}

	// Check if we have enough balance
	if wallet.Balance.Amount.LT(amount.Amount) {
		return types.ErrInsufficientFunds
	}

	// Transfer from balance to emergency reserve
	newReserve := wallet.EmergencyReserve.Add(amount)
	if newReserve.Amount.GT(wallet.Balance.Amount) {
		return types.ErrInsufficientReserve
	}

	wallet.EmergencyReserve = newReserve
	wallet.Balance = wallet.Balance.Sub(amount)

	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyReserve, newReserve.String()),
		),
	)

	return nil
}

// executeNotify executes the notify automation action
func (k Keeper) executeNotify(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	// Emit event for external processing
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyReason, "notification_triggered"),
		),
	)

	return nil
}

// executePauseOperations executes the pause_operations automation action
func (k Keeper) executePauseOperations(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	wallet.IsPaused = true
	wallet.PausedReason = "Automation triggered: " + rule.Trigger

	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyReason, "operations_paused"),
		),
	)

	return nil
}

// executeResumeOperations executes the resume_operations automation action
func (k Keeper) executeResumeOperations(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	wallet.IsPaused = false
	wallet.PausedReason = ""

	if err := k.SetWallet(ctx, wallet); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyReason, "operations_resumed"),
		),
	)

	return nil
}

// executeRequestFunding executes the request_funding automation action
func (k Keeper) executeRequestFunding(ctx sdk.Context, wallet *types.AgentWalletExtended, rule *types.AutomationRuleExtended) error {
	// Emit event for external processing
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutomationTriggered,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, wallet.Address),
			sdk.NewAttribute(types.AttributeKeyTrigger, rule.Trigger),
			sdk.NewAttribute(types.AttributeKeyAction, rule.Action),
			sdk.NewAttribute(types.AttributeKeyAmount, rule.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyReason, "funding_requested"),
		),
	)

	return nil
}

// GetPendingAutomations retrieves all pending automations
func (k Keeper) GetPendingAutomations(ctx sdk.Context, address string) ([]types.PendingAutomation, error) {
	store := ctx.KVStore(k.storeKey)

	var automations []types.PendingAutomation

	if address != "" {
		// Get for specific address
		keyPrefix := append(types.PendingAutomationKey, []byte(address)...)
		iter := store.Iterator(keyPrefix, append(keyPrefix, 0xFF))
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			var automation types.PendingAutomation
			if err := json.Unmarshal(iter.Value(), &automation); err == nil {
				automations = append(automations, automation)
			}
		}
	} else {
		// Get all pending automations
		iter := store.Iterator(types.PendingAutomationKey, append(types.PendingAutomationKey, 0xFF))
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			var automation types.PendingAutomation
			if err := json.Unmarshal(iter.Value(), &automation); err == nil {
				automations = append(automations, automation)
			}
		}
	}

	return automations, nil
}

// ScheduleAutomation schedules a future automation
func (k Keeper) ScheduleAutomation(ctx sdk.Context, did, address, trigger, action string, amount sdk.Coin, scheduledTime int64) error {
	automation := types.PendingAutomation{
		DID:       did,
		Address:   address,
		Trigger:   trigger,
		Action:    action,
		Amount:    amount.String(),
		Scheduled: scheduledTime,
	}

	store := ctx.KVStore(k.storeKey)
	key := append(types.GetPendingAutomationKey(address), []byte(scheduledTime)...)
	bz, err := json.Marshal(&automation)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

// ProcessScheduledAutomations processes all scheduled automations that are due
func (k Keeper) ProcessScheduledAutomations(ctx sdk.Context) error {
	automations, err := k.GetPendingAutomations(ctx, "")
	if err != nil {
		return err
	}

	currentTime := ctx.BlockTime().Unix()

	for _, automation := range automations {
		if automation.Scheduled <= currentTime {
			// Get wallet
			wallet, exists := k.GetWallet(ctx, automation.Address)
			if !exists {
				continue
			}

			// Parse amount
			amount, err := sdk.ParseCoinNormalized(automation.Amount)
			if err != nil {
				continue
			}

			// Create a temporary rule to execute
			rule := &types.AutomationRuleExtended{
				AutomationRule: types.AutomationRule{
					Trigger: automation.Trigger,
					Action:  automation.Action,
					Amount:  amount,
					Enabled: true,
				},
			}

			// Execute the automation
			_ = k.executeAutomation(ctx, wallet, rule)

			// Remove the pending automation
			store := ctx.KVStore(k.storeKey)
			key := append(types.GetPendingAutomationKey(automation.Address), []byte(automation.Scheduled)...)
			store.Delete(key)
		}
	}

	return nil
}