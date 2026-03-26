package computemarket

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

	"github.com/deparrow/dpc/x/computemarket/keeper"
	"github.com/deparrow/dpc/x/computemarket/types"
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the computemarket module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the computemarket module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the computemarket module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the computemarket
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the computemarket module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the computemarket module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here
	// In production, this would use the generated gRPC gateway code
	// _ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetQueryCmd returns the cli query commands for the computemarket module
func (AppModuleBasic) GetQueryCmd() string {
	return types.QuerierRoute
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the computemarket module.
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

// Name returns the computemarket module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	// types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// RegisterInvariants registers the computemarket module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants here
}

// InitGenesis performs the computemarket module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genesisState)

	InitGenesis(ctx, am.keeper, genesisState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the computemarket module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// BeginBlock executes all ABCI BeginBlock logic respective to the computemarket module.
func (am AppModule) BeginBlock(ctx sdk.Context) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the computemarket module.
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

	// Set total staked
	if genesisState.TotalStaked != "" {
		staked, err := types.ValidateStakeAmount(genesisState.TotalStaked)
		if err != nil {
			panic(fmt.Errorf("invalid total staked in genesis: %w", err))
		}
		k.SetTotalStaked(ctx, sdk.NewInt(staked))
	}

	// Set total escrowed
	if genesisState.TotalEscrowed != "" {
		escrowed, err := types.ValidateStakeAmount(genesisState.TotalEscrowed)
		if err != nil {
			panic(fmt.Errorf("invalid total escrowed in genesis: %w", err))
		}
		k.SetTotalEscrowed(ctx, sdk.NewInt(escrowed))
	}

	// Initialize providers from genesis
	for _, provider := range genesisState.Providers {
		if err := k.SetProvider(ctx, &provider); err != nil {
			panic(fmt.Errorf("failed to set provider %s: %w", provider.Address, err))
		}
		if provider.IsActive() {
			_ = k.AddActiveProvider(ctx, provider.Address)
		}
	}

	// Initialize escrows from genesis
	for _, escrow := range genesisState.Escrows {
		if err := k.SetEscrow(ctx, &escrow); err != nil {
			panic(fmt.Errorf("failed to set escrow %s: %w", escrow.ID, err))
		}
	}

	// Initialize disputes from genesis
	for _, dispute := range genesisState.Disputes {
		if err := k.SetDispute(ctx, &dispute); err != nil {
			panic(fmt.Errorf("failed to set dispute %s: %w", dispute.ID, err))
		}
	}

	// Initialize job matches from genesis
	for _, match := range genesisState.JobMatches {
		if err := k.SetJobMatch(ctx, &match); err != nil {
			panic(fmt.Errorf("failed to set job match %s: %w", match.JobID, err))
		}
	}
}

// ExportGenesis exports the module's genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesisState := types.DefaultGenesis()

	// Get params
	genesisState.Params = k.GetParams(ctx)

	// Get total staked
	genesisState.TotalStaked = k.GetTotalStaked(ctx).String()

	// Get total escrowed
	genesisState.TotalEscrowed = k.GetTotalEscrowed(ctx).String()

	// Get all providers
	providers, _ := k.GetAllProviders(ctx)
	genesisState.Providers = providers

	// Get all escrows
	escrows, _ := k.GetAllEscrows(ctx)
	genesisState.Escrows = escrows

	// Get all disputes
	disputes, _ := k.GetAllDisputes(ctx)
	genesisState.Disputes = disputes

	// Get all job matches
	matches, _ := k.GetAllJobMatches(ctx)
	genesisState.JobMatches = matches

	return genesisState
}

// ----------------------------------------------------------------------------
// BeginBlocker / EndBlocker
// ----------------------------------------------------------------------------

// BeginBlocker handles block begin logic
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Process expired escrows
	processExpiredEscrows(ctx, k)
}

// EndBlocker handles block end logic
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// Process expired disputes
	processExpiredDisputes(ctx, k)
}

// processExpiredEscrows handles expired escrows
func processExpiredEscrows(ctx sdk.Context, k keeper.Keeper) {
	escrows, err := k.GetAllEscrows(ctx)
	if err != nil {
		return
	}

	currentTime := ctx.BlockTime().Unix()

	for _, escrow := range escrows {
		if escrow.IsLocked() && escrow.IsExpired(currentTime) {
			// Auto-refund expired escrows
			_ = k.RefundEscrow(ctx, escrow.ID, escrow.Submitter)
		}
	}
}

// processExpiredDisputes handles expired disputes
func processExpiredDisputes(ctx sdk.Context, k keeper.Keeper) {
	disputes, err := k.GetAllDisputes(ctx)
	if err != nil {
		return
	}

	for _, dispute := range disputes {
		if dispute.Status == types.DisputeStatusOpen {
			_ = k.ExpireDispute(ctx, dispute.ID)
		}
	}
}
