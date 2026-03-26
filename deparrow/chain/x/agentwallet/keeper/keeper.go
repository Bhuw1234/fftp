package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// Keeper of the agentwallet module
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    bankkeeper.Keeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// is the governance module's address.
	authority string
}

// NewKeeper creates a new agentwallet Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	bankKeeper bankkeeper.Keeper,
	authority string,
) Keeper {
	// ensure authority is a valid AccAddress
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		storeKey:   key,
		cdc:        cdc,
		bankKeeper: bankKeeper,
		authority:  authority,
	}
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the total set of agentwallet parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the agentwallet parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetModuleAccountAddress returns the module account address
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return authtypes.NewModuleAddress(types.ModuleName)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) interface{} {
	return ctx.Logger()
}

// GetStoreKey returns the store key for the module
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// TransferCoins transfers coins from one account to another
func (k Keeper) TransferCoins(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return k.bankKeeper.SendCoins(ctx, from, to, amount)
}

// SendCoinsFromAccountToModule sends coins from an account to a module
func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, sender sdk.AccAddress, recipientModule string, amount sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, recipientModule, amount)
}

// SendCoinsFromModuleToAccount sends coins from a module to an account
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipient sdk.AccAddress, amount sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipient, amount)
}

// BurnCoins burns coins from a module
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amount sdk.Coins) error {
	return k.bankKeeper.BurnCoins(ctx, moduleName, amount)
}

// GetBalance returns the balance of an account
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.bankKeeper.GetBalance(ctx, addr, denom)
}

// SpendableCoins returns the spendable coins of an account
func (k Keeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bankKeeper.SpendableCoins(ctx, addr)
}

// GetAllWallets retrieves all wallets from the store
func (k Keeper) GetAllWallets(ctx sdk.Context) ([]types.AgentWalletExtended, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.WalletKey)
	defer iter.Close()

	var wallets []types.AgentWalletExtended
	for ; iter.Valid(); iter.Next() {
		var wallet types.AgentWalletExtended
		k.cdc.MustUnmarshal(iter.Value(), &wallet)
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}

// GetAllAgents retrieves all agents from the store
func (k Keeper) GetAllAgents(ctx sdk.Context) ([]types.AgentInfo, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.AgentRegistryKey)
	defer iter.Close()

	var agents []types.AgentInfo
	for ; iter.Valid(); iter.Next() {
		var agent types.AgentInfo
		k.cdc.MustUnmarshal(iter.Value(), &agent)
		agents = append(agents, agent)
	}
	return agents, nil
}

// GetTotalWallets returns the total number of wallets
func (k Keeper) GetTotalWallets(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("total_wallets"))
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTotalWallets sets the total number of wallets
func (k Keeper) SetTotalWallets(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("total_wallets"), sdk.Uint64ToBigEndian(count))
}

// IncrementTotalWallets increments the wallet count
func (k Keeper) IncrementTotalWallets(ctx sdk.Context) uint64 {
	count := k.GetTotalWallets(ctx)
	count++
	k.SetTotalWallets(ctx, count)
	return count
}

// GetTotalAgents returns the total number of agents
func (k Keeper) GetTotalAgents(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("total_agents"))
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTotalAgents sets the total number of agents
func (k Keeper) SetTotalAgents(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("total_agents"), sdk.Uint64ToBigEndian(count))
}

// IncrementTotalAgents increments the agent count
func (k Keeper) IncrementTotalAgents(ctx sdk.Context) uint64 {
	count := k.GetTotalAgents(ctx)
	count++
	k.SetTotalAgents(ctx, count)
	return count
}

// GetTotalTransactions returns the total number of transactions
func (k Keeper) GetTotalTransactions(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("total_transactions"))
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTotalTransactions sets the total number of transactions
func (k Keeper) SetTotalTransactions(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("total_transactions"), sdk.Uint64ToBigEndian(count))
}

// IncrementTotalTransactions increments the transaction count
func (k Keeper) IncrementTotalTransactions(ctx sdk.Context) uint64 {
	count := k.GetTotalTransactions(ctx)
	count++
	k.SetTotalTransactions(ctx, count)
	return count
}

// GetCurrentDay returns the current day (unix timestamp / 86400)
func (k Keeper) GetCurrentDay(ctx sdk.Context) int64 {
	return ctx.BlockTime().Unix() / 86400
}

// ResetDailySpending resets the daily spending counters for all wallets
func (k Keeper) ResetDailySpending(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, DailySpendingKey)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}

	// Update wallet daily spent fields
	wallets, _ := k.GetAllWallets(ctx)
	for _, wallet := range wallets {
		wallet.DailySpent = sdk.NewCoin("dpc", sdk.ZeroInt())
		_ = k.SetWallet(ctx, &wallet)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDailySpendingReset,
			sdk.NewAttribute("timestamp", ctx.BlockTime().String()),
		),
	)
}

// DailySpendingKey is exported for use in BeginBlocker
var DailySpendingKey = types.DailySpendingKey
