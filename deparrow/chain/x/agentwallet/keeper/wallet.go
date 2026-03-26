package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// GetWallet retrieves a wallet by address
func (k Keeper) GetWallet(ctx sdk.Context, address string) (*types.AgentWalletExtended, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWalletKey(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var wallet types.AgentWalletExtended
	k.cdc.MustUnmarshal(bz, &wallet)
	return &wallet, true
}

// GetWalletByDID retrieves a wallet by DID
func (k Keeper) GetWalletByDID(ctx sdk.Context, did string) (*types.AgentWalletExtended, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWalletByDIDKey(did)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var wallet types.AgentWalletExtended
	k.cdc.MustUnmarshal(bz, &wallet)
	return &wallet, true
}

// SetWallet stores a wallet
func (k Keeper) SetWallet(ctx sdk.Context, wallet *types.AgentWalletExtended) error {
	if err := wallet.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	// Store by address
	key := types.GetWalletKey(wallet.Address)
	bz := k.cdc.MustMarshal(wallet)
	store.Set(key, bz)

	// Store by DID for reverse lookup
	didKey := types.GetWalletByDIDKey(wallet.DID)
	store.Set(didKey, bz)

	return nil
}

// CreateWallet creates a new agent wallet
func (k Keeper) CreateWallet(ctx sdk.Context, did, address string, initialFunds sdk.Coin) (*types.AgentWalletExtended, error) {
	// Check if wallet already exists
	if _, exists := k.GetWallet(ctx, address); exists {
		return nil, types.ErrWalletAlreadyExists
	}

	// Check if DID is already used
	if _, exists := k.GetWalletByDID(ctx, did); exists {
		return nil, types.ErrWalletAlreadyExists
	}

	// Validate DID
	if !types.IsValidDID(did) {
		return nil, types.ErrInvalidDID
	}

	// Validate address
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Create wallet with default rules
	wallet := types.NewAgentWalletExtended(did, address)
	wallet.Balance = initialFunds

	// Set defaults
	params := k.GetParams(ctx)
	minReserve, _ := params.GetMinEmergencyReserve()
	wallet.EmergencyReserve = sdk.NewCoin("dpc", minReserve)
	wallet.CreatedAt = ctx.BlockTime().Unix()
	wallet.LastActivityTime = ctx.BlockTime().Unix()

	// Store wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Increment total wallets
	k.IncrementTotalWallets(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWalletCreated,
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyBalance, initialFunds.String()),
			sdk.NewAttribute(types.AttributeKeyTimestamp, ctx.BlockTime().String()),
		),
	)

	return wallet, nil
}

// DeleteWallet deletes a wallet and returns the remaining balance
func (k Keeper) DeleteWallet(ctx sdk.Context, address string) (sdk.Coin, error) {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return sdk.Coin{}, types.ErrWalletNotFound
	}

	// Calculate withdrawable amount (balance minus reserve)
	withdrawable := wallet.GetAvailableBalance()

	store := ctx.KVStore(k.storeKey)

	// Delete wallet entries
	store.Delete(types.GetWalletKey(address))
	store.Delete(types.GetWalletByDIDKey(wallet.DID))

	// Delete related data
	k.deleteWalletSpendingRules(ctx, address)
	k.deleteWalletAutomationRules(ctx, address)
	k.deleteWalletDailySpending(ctx, address)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWalletDeleted,
			sdk.NewAttribute(types.AttributeKeyDID, wallet.DID),
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyAmount, withdrawable.String()),
		),
	)

	return withdrawable, nil
}

// Deposit deposits funds into a wallet
func (k Keeper) Deposit(ctx sdk.Context, address string, amount sdk.Coin) (*types.AgentWalletExtended, error) {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return nil, types.ErrWalletNotFound
	}

	// Update balance
	wallet.Balance = wallet.Balance.Add(amount)
	wallet.LastActivityTime = ctx.BlockTime().Unix()
	wallet.TotalTransactions++

	// Store updated wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Increment global transaction count
	k.IncrementTotalTransactions(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFundsDeposited,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyBalance, wallet.Balance.String()),
		),
	)

	return wallet, nil
}

// Withdraw withdraws funds from a wallet
func (k Keeper) Withdraw(ctx sdk.Context, address string, amount sdk.Coin) (*types.AgentWalletExtended, error) {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return nil, types.ErrWalletNotFound
	}

	// Check if withdrawal is allowed (must leave emergency reserve)
	availableBalance := wallet.GetAvailableBalance()
	if availableBalance.IsLT(amount) {
		return nil, types.ErrInsufficientFunds
	}

	// Check spending rules
	if !wallet.CanSpendWithDaily(amount, types.OperationWithdraw) {
		return nil, types.ErrSpendingRuleViolation
	}

	// Update balance
	wallet.Balance = wallet.Balance.Sub(amount)
	wallet.LastActivityTime = ctx.BlockTime().Unix()
	wallet.TotalTransactions++

	// Update daily spending
	k.updateDailySpending(ctx, address, amount)

	// Store updated wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Increment global transaction count
	k.IncrementTotalTransactions(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFundsWithdrawn,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyBalance, wallet.Balance.String()),
		),
	)

	return wallet, nil
}

// Transfer transfers funds between wallets
func (k Keeper) Transfer(ctx sdk.Context, senderAddr, recipientAddr string, amount sdk.Coin, operation string) (*types.AgentWalletExtended, *types.AgentWalletExtended, error) {
	// Get sender wallet
	sender, exists := k.GetWallet(ctx, senderAddr)
	if !exists {
		return nil, nil, types.ErrWalletNotFound
	}

	// Get recipient wallet
	recipient, exists := k.GetWallet(ctx, recipientAddr)
	if !exists {
		return nil, nil, types.ErrWalletNotFound
	}

	// Check if transfer is allowed
	availableBalance := sender.GetAvailableBalance()
	if availableBalance.IsLT(amount) {
		return nil, nil, types.ErrInsufficientFunds
	}

	// Check spending rules
	if !sender.CanSpendWithDaily(amount, operation) {
		return nil, nil, types.ErrSpendingRuleViolation
	}

	// Check if operation is blocked
	if sender.IsOperationBlocked(operation) {
		return nil, nil, types.ErrOperationBlocked
	}

	// Check external transfer permission
	if operation == types.OperationExternalTransfer {
		params := k.GetParams(ctx)
		if !params.AllowExternalTransfers {
			return nil, nil, types.ErrExternalTransferBlocked
		}
	}

	// Update balances
	sender.Balance = sender.Balance.Sub(amount)
	sender.LastActivityTime = ctx.BlockTime().Unix()
	sender.TotalTransactions++

	recipient.Balance = recipient.Balance.Add(amount)
	recipient.LastActivityTime = ctx.BlockTime().Unix()
	recipient.TotalTransactions++

	// Update daily spending for sender
	k.updateDailySpending(ctx, senderAddr, amount)

	// Store updated wallets
	if err := k.SetWallet(ctx, sender); err != nil {
		return nil, nil, err
	}
	if err := k.SetWallet(ctx, recipient); err != nil {
		return nil, nil, err
	}

	// Increment global transaction count
	k.IncrementTotalTransactions(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFundsTransferred,
			sdk.NewAttribute(types.AttributeKeySender, senderAddr),
			sdk.NewAttribute(types.AttributeKeyRecipient, recipientAddr),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyOperation, operation),
		),
	)

	return sender, recipient, nil
}

// AutonomousSpend performs an autonomous spend operation for AI agents
func (k Keeper) AutonomousSpend(ctx sdk.Context, did, address, recipient string, amount sdk.Coin, operation string, signature []byte) (*types.AgentWalletExtended, string, error) {
	// Verify autonomous signature
	if !k.verifyAutonomousSignature(ctx, did, address, recipient, amount, operation, signature) {
		return nil, "", types.ErrInvalidSignature
	}

	// Get wallet
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return nil, "", types.ErrWalletNotFound
	}

	// Check if autonomous transactions are enabled
	params := k.GetParams(ctx)
	if !params.AutonomousTxEnabled {
		return nil, "", types.ErrUnauthorized
	}

	// Perform the transfer
	sender, _, err := k.Transfer(ctx, address, recipient, amount, operation)
	if err != nil {
		return nil, "", err
	}

	// Generate tx hash (simplified)
	txHash := generateTxHash(ctx, did, amount, operation)

	// Emit autonomous transaction event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAutonomousTxExecuted,
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyOperation, operation),
			sdk.NewAttribute(types.AttributeKeyTxHash, txHash),
		),
	)

	return sender, txHash, nil
}

// RegisterAgent registers a new AI agent
func (k Keeper) RegisterAgent(ctx sdk.Context, did, address, agentType, metadata string) (*types.AgentInfo, error) {
	// Check if agent already exists
	if _, exists := k.GetAgent(ctx, did); exists {
		return nil, types.ErrAgentAlreadyRegistered
	}

	// Ensure wallet exists for this agent
	if _, exists := k.GetWallet(ctx, address); !exists {
		return nil, types.ErrWalletNotFound
	}

	agent := &types.AgentInfo{
		DID:          did,
		Address:      address,
		AgentType:    agentType,
		Metadata:     metadata,
		RegisteredAt: ctx.BlockTime().Unix(),
		IsActive:     true,
	}

	// Store agent
	store := ctx.KVStore(k.storeKey)
	key := types.GetAgentRegistryKey(did)
	bz := k.cdc.MustMarshal(agent)
	store.Set(key, bz)

	// Increment total agents
	k.IncrementTotalAgents(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAgentRegistered,
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyAgentType, agentType),
		),
	)

	return agent, nil
}

// UnregisterAgent unregisters an AI agent
func (k Keeper) UnregisterAgent(ctx sdk.Context, did string) error {
	agent, exists := k.GetAgent(ctx, did)
	if !exists {
		return types.ErrAgentNotRegistered
	}

	agent.IsActive = false

	// Update agent
	store := ctx.KVStore(k.storeKey)
	key := types.GetAgentRegistryKey(did)
	bz := k.cdc.MustMarshal(agent)
	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAgentUnregistered,
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyAddress, agent.Address),
		),
	)

	return nil
}

// GetAgent retrieves an agent by DID
func (k Keeper) GetAgent(ctx sdk.Context, did string) (*types.AgentInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAgentRegistryKey(did)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var agent types.AgentInfo
	k.cdc.MustUnmarshal(bz, &agent)
	return &agent, true
}

// GetAgentByAddress retrieves an agent by address
func (k Keeper) GetAgentByAddress(ctx sdk.Context, address string) (*types.AgentInfo, bool) {
	agents, _ := k.GetAllAgents(ctx)
	for _, agent := range agents {
		if agent.Address == address {
			return &agent, true
		}
	}
	return nil, false
}

// SetEmergencyReserve sets the emergency reserve for a wallet
func (k Keeper) SetEmergencyReserve(ctx sdk.Context, address string, amount sdk.Coin) (*types.AgentWalletExtended, error) {
	wallet, exists := k.GetWallet(ctx, address)
	if !exists {
		return nil, types.ErrWalletNotFound
	}

	// Ensure reserve doesn't exceed balance
	if amount.IsGT(wallet.Balance) {
		return nil, types.ErrInsufficientReserve
	}

	// Update reserve
	wallet.EmergencyReserve = amount
	wallet.LastActivityTime = time.Now().Unix()

	// Store updated wallet
	if err := k.SetWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEmergencyReserveUpdated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyReserve, amount.String()),
		),
	)

	return wallet, nil
}

// Helper functions

func (k Keeper) deleteWalletSpendingRules(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := append(types.SpendingRuleKey, []byte(address)...)
	iter := sdk.KVStorePrefixIterator(store, keyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (k Keeper) deleteWalletAutomationRules(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := append(types.AutomationRuleKey, []byte(address)...)
	iter := sdk.KVStorePrefixIterator(store, keyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (k Keeper) deleteWalletDailySpending(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	keyPrefix := append(types.DailySpendingKey, []byte(address)...)
	iter := sdk.KVStorePrefixIterator(store, keyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

func (k Keeper) updateDailySpending(ctx sdk.Context, address string, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	day := k.GetCurrentDay(ctx)
	key := types.GetDailySpendingKey(address, day)

	// Get current daily spent
	var dailySpent sdk.Coin
	bz := store.Get(key)
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &dailySpent)
	} else {
		dailySpent = sdk.NewCoin("dpc", sdk.ZeroInt())
	}

	// Update daily spent
	dailySpent = dailySpent.Add(amount)
	bz = k.cdc.MustMarshal(&dailySpent)
	store.Set(key, bz)
}

func (k Keeper) verifyAutonomousSignature(ctx sdk.Context, did, address, recipient string, amount sdk.Coin, operation string, signature []byte) bool {
	// In production, this would verify a cryptographic signature
	// For now, we accept any non-empty signature from a registered agent
	agent, exists := k.GetAgent(ctx, did)
	if !exists || !agent.IsActive {
		return false
	}

	// Verify address matches
	if agent.Address != address {
		return false
	}

	// Accept any signature for autonomous agents (placeholder)
	return len(signature) > 0
}

func generateTxHash(ctx sdk.Context, did string, amount sdk.Coin, operation string) string {
	// Simplified tx hash generation
	// In production, this would be the actual transaction hash
	return sdk.HexBytesToHex([]byte(did + amount.String() + operation + ctx.BlockTime().String()))[:16]
}
