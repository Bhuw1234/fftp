package agentwallet

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/deparrow/dpc/x/agentwallet/keeper"
	"github.com/deparrow/dpc/x/agentwallet/types"
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the agentwallet module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the agentwallet module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the agentwallet module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the agentwallet
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the agentwallet module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the agentwallet module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here
	// In production, this would use the generated gRPC gateway code
	// _ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetQueryCmd returns the cli query commands for the agentwallet module
func (AppModuleBasic) GetQueryCmd() string {
	return types.QuerierRoute
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the agentwallet module.
type AppModule struct {
	AppModuleBasic

	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		bankKeeper:     bankKeeper,
	}
}

// Name returns the agentwallet module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// RegisterInvariants registers the agentwallet module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants here
}

// InitGenesis performs the agentwallet module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genesisState)

	InitGenesis(ctx, am.keeper, genesisState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the agentwallet module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// BeginBlock executes all ABCI BeginBlock logic respective to the agentwallet module.
func (am AppModule) BeginBlock(ctx sdk.Context) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the agentwallet module.
// It returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// ----------------------------------------------------------------------------
// Genesis
// ----------------------------------------------------------------------------

// InitGenesis initializes the module's genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genesisState types.GenesisState) {
	// Set params
	if err := k.SetParams(ctx, genesisState.Params); err != nil {
		panic(err)
	}

	// Set total wallets
	k.SetTotalWallets(ctx, genesisState.TotalWallets)

	// Set total agents
	k.SetTotalAgents(ctx, genesisState.TotalAgents)

	// Initialize wallets from genesis
	for _, wallet := range genesisState.Wallets {
		if err := k.SetWallet(ctx, &wallet); err != nil {
			panic(fmt.Errorf("failed to set wallet %s: %w", wallet.DID, err))
		}
	}

	// Initialize agents from genesis
	for _, agent := range genesisState.Agents {
		// Agent will be linked to wallet during wallet creation
		// Here we just register the agent
		_, _ = k.RegisterAgent(ctx, agent.DID, agent.Address, agent.AgentType, agent.Metadata)
	}

	// Initialize pending automations from genesis
	for _, automation := range genesisState.PendingAutomations {
		amount, err := sdk.ParseCoinNormalized(automation.Amount)
		if err != nil {
			panic(fmt.Errorf("invalid automation amount: %w", err))
		}
		if err := k.ScheduleAutomation(ctx, automation.DID, automation.Address, automation.Trigger, automation.Action, amount, automation.Scheduled); err != nil {
			panic(fmt.Errorf("failed to schedule automation: %w", err))
		}
	}
}

// ExportGenesis exports the module's genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesisState := types.DefaultGenesis()

	// Get params
	genesisState.Params = k.GetParams(ctx)

	// Get total wallets
	genesisState.TotalWallets = k.GetTotalWallets(ctx)

	// Get total agents
	genesisState.TotalAgents = k.GetTotalAgents(ctx)

	// Get all wallets
	wallets, _ := k.GetAllWallets(ctx)
	genesisState.Wallets = wallets

	// Get all agents
	agents, _ := k.GetAllAgents(ctx)
	genesisState.Agents = agents

	// Get pending automations
	automations, _ := k.GetPendingAutomations(ctx, "")
	genesisState.PendingAutomations = automations

	return genesisState
}

// ----------------------------------------------------------------------------
// BeginBlocker / EndBlocker
// ----------------------------------------------------------------------------

// BeginBlocker handles block begin logic
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Reset daily spending if it's a new day
	// This is a simplified check - in production you'd track the last reset day
	currentDay := k.GetCurrentDay(ctx)
	store := ctx.KVStore(k.GetStoreKey())
	lastResetKey := []byte("last_daily_reset")

	lastResetBz := store.Get(lastResetKey)
	var lastResetDay int64 = 0
	if lastResetBz != nil {
		// Simple unmarshal
		for i, b := range lastResetBz {
			lastResetDay |= int64(b) << (8 * i)
		}
	}

	if currentDay > lastResetDay {
		// It's a new day, reset daily spending
		k.ResetDailySpending(ctx)

		// Update last reset day
		resetBz := make([]byte, 8)
		for i := 0; i < 8; i++ {
			resetBz[i] = byte(currentDay >> (8 * i))
		}
		store.Set(lastResetKey, resetBz)
	}

	// Process scheduled automations
	_ = k.ProcessScheduledAutomations(ctx)
}

// EndBlocker handles block end logic
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Check and trigger automations
	_ = k.CheckAndTriggerAutomations(ctx)
}
