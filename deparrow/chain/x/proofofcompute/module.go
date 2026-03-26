package proofofcompute

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

	"github.com/deparrow/dpc/x/proofofcompute/keeper"
	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the proofofcompute module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the proofofcompute module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the proofofcompute module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the proofofcompute
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the proofofcompute module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the proofofcompute module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes here
	// In production, this would use the generated gRPC gateway code
	// _ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetQueryCmd returns the cli query commands for the proofofcompute module
func (AppModuleBasic) GetQueryCmd() string {
	return types.QuerierRoute
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the proofofcompute module.
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

// Name returns the proofofcompute module's name.
func (am AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// RegisterInvariants registers the proofofcompute module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants here
}

// InitGenesis performs the proofofcompute module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genesisState)

	InitGenesis(ctx, am.keeper, genesisState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the proofofcompute module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// BeginBlock executes all ABCI BeginBlock logic respective to the proofofcompute module.
func (am AppModule) BeginBlock(ctx sdk.Context) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the proofofcompute module.
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

	// Set total supply
	if genesisState.TotalSupply != "" {
		supply, err := sdk.NewIntFromString(genesisState.TotalSupply)
		if err != nil {
			panic(fmt.Errorf("invalid total supply in genesis: %w", err))
		}
		k.SetTotalSupply(ctx, supply)
	}

	// Set current difficulty
	k.SetCurrentDifficulty(ctx, genesisState.CurrentDifficulty)

	// Initialize jobs from genesis
	for _, job := range genesisState.Jobs {
		if err := k.SetJob(ctx, &job); err != nil {
			panic(fmt.Errorf("failed to set job %s: %w", job.ID, err))
		}
	}

	// Initialize proofs from genesis
	for _, proof := range genesisState.Proofs {
		if err := k.SetProof(ctx, &proof); err != nil {
			panic(fmt.Errorf("failed to set proof for job %s: %w", proof.JobID, err))
		}
	}
}

// ExportGenesis exports the module's genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesisState := types.DefaultGenesis()

	// Get params
	genesisState.Params = k.GetParams(ctx)

	// Get total supply
	genesisState.TotalSupply = k.GetTotalSupply(ctx).String()

	// Get current difficulty
	genesisState.CurrentDifficulty = k.GetCurrentDifficulty(ctx)

	// Get all jobs using the keeper's storeKey accessor
	jobs, _ := k.GetAllJobs(ctx)
	genesisState.Jobs = jobs

	// Get all proofs
	proofs, _ := k.GetAllProofs(ctx)
	genesisState.Proofs = proofs

	return genesisState
}