package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// Keeper of the computemarket module
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    bankkeeper.Keeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// is the governance module's address.
	authority string
}

// NewKeeper creates a new computemarket Keeper instance
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

// GetParams returns the total set of computemarket parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the computemarket parameters to the param space.
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
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetStoreKey returns the store key for the module
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetAllProviders retrieves all providers from the store
func (k Keeper) GetAllProviders(ctx sdk.Context) ([]types.ProviderExtended, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ProviderKey)
	defer iter.Close()

	var providers []types.ProviderExtended
	for ; iter.Valid(); iter.Next() {
		var provider types.ProviderExtended
		k.cdc.MustUnmarshal(iter.Value(), &provider)
		providers = append(providers, provider)
	}
	return providers, nil
}

// GetAllEscrows retrieves all escrows from the store
func (k Keeper) GetAllEscrows(ctx sdk.Context) ([]types.EscrowExtended, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.EscrowKey)
	defer iter.Close()

	var escrows []types.EscrowExtended
	for ; iter.Valid(); iter.Next() {
		var escrow types.EscrowExtended
		k.cdc.MustUnmarshal(iter.Value(), &escrow)
		escrows = append(escrows, escrow)
	}
	return escrows, nil
}

// GetAllDisputes retrieves all disputes from the store
func (k Keeper) GetAllDisputes(ctx sdk.Context) ([]types.Dispute, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DisputeKey)
	defer iter.Close()

	var disputes []types.Dispute
	for ; iter.Valid(); iter.Next() {
		var dispute types.Dispute
		k.cdc.MustUnmarshal(iter.Value(), &dispute)
		disputes = append(disputes, dispute)
	}
	return disputes, nil
}

// GetAllJobMatches retrieves all job matches from the store
func (k Keeper) GetAllJobMatches(ctx sdk.Context) ([]types.JobMatch, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.JobMatchKey)
	defer iter.Close()

	var matches []types.JobMatch
	for ; iter.Valid(); iter.Next() {
		var match types.JobMatch
		k.cdc.MustUnmarshal(iter.Value(), &match)
		matches = append(matches, match)
	}
	return matches, nil
}

// GetTotalStaked returns the total DPC staked by providers
func (k Keeper) GetTotalStaked(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("total_staked"))
	if bz == nil {
		return math.ZeroInt()
	}
	var total math.Int
	k.cdc.MustUnmarshal(bz, &total)
	return total
}

// SetTotalStaked sets the total DPC staked
func (k Keeper) SetTotalStaked(ctx sdk.Context, total math.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&total)
	store.Set([]byte("total_staked"), bz)
}

// GetTotalEscrowed returns the total DPC locked in escrows
func (k Keeper) GetTotalEscrowed(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("total_escrowed"))
	if bz == nil {
		return math.ZeroInt()
	}
	var total math.Int
	k.cdc.MustUnmarshal(bz, &total)
	return total
}

// SetTotalEscrowed sets the total DPC escrowed
func (k Keeper) SetTotalEscrowed(ctx sdk.Context, total math.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&total)
	store.Set([]byte("total_escrowed"), bz)
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