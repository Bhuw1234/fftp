package app

import (
	"encoding/json"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file dump as well as the list of validators with their power.
func (app *DPCApp) ExportAppStateAndValidators(
	forZeroHeight bool,
	jailAllowedAddrs []string,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start
	ctx := app.NewContextLegacy(true, tmproto.Header{Height: app.LastBlockHeight()})

	// We export at least the modules requested.
	if len(modulesToExport) == 0 {
		modulesToExport = []string{
			"auth", "bank", "staking", "distribution", "slashing", "gov", "mint",
		}
	}

	// Export state to json.
	genState := app.mm.ExportGenesisForModules(ctx, app.appCodec, modulesToExport)

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	// Export validators.
	validators, err := app.exportValidators(ctx)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          app.LastBlockHeight(),
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}

// exportValidators returns the validators for the genesis state.
func (app *DPCApp) exportValidators(ctx sdk.Context) ([]stakingtypes.Validator, error) {
	return app.StakingKeeper.GetAllValidators(ctx)
}

// prepare for zero height genesis reset.
func (app *DPCApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) error {
	// Apply any special logic for zero-height genesis state reset here.
	// This is typically used for testing or chain resets.
	
	// Just a stub for now - implement when needed.
	return nil
}

// Release releasing the db.
func (app *DPCApp) Release() {
	// Close the underlying database connection
	if app.BaseApp != nil {
		app.BaseApp.Close()
	}
}
