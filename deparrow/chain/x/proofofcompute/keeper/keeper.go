package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// Keeper of the proofofcompute module
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    bankkeeper.Keeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// is the governance module's address.
	authority string
}

// NewKeeper creates a new proofofcompute Keeper instance
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

// GetParams returns the total set of proofofcompute parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the proofofcompute parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetTotalSupply returns the total DPC minted through PoC rewards
func (k Keeper) GetTotalSupply(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TotalSupplyKey)
	if bz == nil {
		return sdk.ZeroInt()
	}
	var supply sdk.Int
	k.cdc.MustUnmarshal(bz, &supply)
	return supply
}

// SetTotalSupply sets the total DPC minted
func (k Keeper) SetTotalSupply(ctx sdk.Context, supply sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(supply)
	store.Set(types.TotalSupplyKey, bz)
}

// GetBlockJobCount returns the number of jobs completed in the current block
func (k Keeper) GetBlockJobCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BlockJobCountKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetBlockJobCount sets the number of jobs completed in the current block
func (k Keeper) SetBlockJobCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.BlockJobCountKey, sdk.Uint64ToBigEndian(count))
}

// IncrementBlockJobCount increments the job count for the current block
func (k Keeper) IncrementBlockJobCount(ctx sdk.Context) uint64 {
	count := k.GetBlockJobCount(ctx)
	count++
	k.SetBlockJobCount(ctx, count)
	return count
}

// ResetBlockJobCount resets the job count at the beginning of each block
func (k Keeper) ResetBlockJobCount(ctx sdk.Context) {
	k.SetBlockJobCount(ctx, 0)
}

// MintCoins mints new DPC coins and sends them to the recipient
func (k Keeper) MintCoins(ctx sdk.Context, recipient string, amount sdk.Coins) error {
	// Check max supply constraint
	params := k.GetParams(ctx)
	maxSupply, err := params.GetMaxSupply()
	if err != nil {
		return fmt.Errorf("failed to get max supply: %w", err)
	}

	currentSupply := k.GetTotalSupply(ctx)
	newSupply := currentSupply.Add(amount.AmountOf("dpc"))

	if newSupply.GT(maxSupply) {
		return types.ErrMaxSupplyExceeded
	}

	// Mint coins to module account
	coins := sdk.NewCoins(sdk.NewCoin("dpc", amount.AmountOf("dpc")))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return fmt.Errorf("failed to mint coins: %w", err)
	}

	// Send to recipient
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, recipientAddr, coins,
	); err != nil {
		return fmt.Errorf("failed to send coins: %w", err)
	}

	// Update total supply
	k.SetTotalSupply(ctx, newSupply)

	return nil
}

// GetModuleAccountAddress returns the module account address
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return authtypes.NewModuleAddress(types.ModuleName)
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) sdk.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetStoreKey returns the store key for the module
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetAllJobs retrieves all jobs from the store
func (k Keeper) GetAllJobs(ctx sdk.Context) ([]types.Job, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.JobKey)
	defer iter.Close()

	var jobs []types.Job
	for ; iter.Valid(); iter.Next() {
		var job types.Job
		k.cdc.MustUnmarshal(iter.Value(), &job)
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// GetAllProofs retrieves all proofs from the store
func (k Keeper) GetAllProofs(ctx sdk.Context) ([]types.ComputeProof, error) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ProofKey)
	defer iter.Close()

	var proofs []types.ComputeProof
	for ; iter.Valid(); iter.Next() {
		var proof types.ComputeProof
		k.cdc.MustUnmarshal(iter.Value(), &proof)
		proofs = append(proofs, proof)
	}
	return proofs, nil
}
