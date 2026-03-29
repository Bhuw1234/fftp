// Package agentwallet implements the AI Agent Wallet module for DPC blockchain
package agentwallet

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/agentwallet/keeper"
	"github.com/deparrow/dpc/x/agentwallet/types"
)

// ModuleName is the name of this module
const ModuleName = types.ModuleName

// Module implements the agentwallet module
type Module struct {
	keeper keeper.Keeper
}

// NewModule creates a new agentwallet module
func NewModule(db *badger.DB) *Module {
	return &Module{
		keeper: keeper.NewKeeper(db),
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return ModuleName
}

// Keeper returns the module's keeper
func (m *Module) Keeper() keeper.Keeper {
	return m.keeper
}

// DefaultGenesis returns the default genesis state
func (m *Module) DefaultGenesis() []byte {
	genesis := types.DefaultGenesisState()
	bz, _ := json.Marshal(genesis)
	return bz
}

// InitGenesis initializes the module from genesis state
func (m *Module) InitGenesis(bz []byte) error {
	var genesis types.GenesisState
	if err := json.Unmarshal(bz, &genesis); err != nil {
		// If genesis is empty or invalid, use defaults
		genesis = types.DefaultGenesisState()
	}

	// Initialize wallets from genesis
	for _, wallet := range genesis.Wallets {
		if err := m.keeper.SetWallet(wallet); err != nil {
			return fmt.Errorf("failed to set genesis wallet: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the module's genesis state
func (m *Module) ExportGenesis() []byte {
	genesis := types.GenesisState{
		Params:  m.keeper.GetParams(),
		Wallets: m.keeper.GetAllWallets(),
	}
	bz, _ := json.Marshal(genesis)
	return bz
}

// ProcessTransaction processes a transaction for this module
func (m *Module) ProcessTransaction(txType string, txData json.RawMessage, blockHeight int64) (interface{}, error) {
	return m.keeper.ProcessTransaction(txType, txData, blockHeight)
}

// HandleQuery handles a query for this module
func (m *Module) HandleQuery(path string, data []byte) ([]byte, error) {
	return m.keeper.HandleQuery(path, data)
}

// BeginBlock performs begin block logic
func (m *Module) BeginBlock(height int64) {
	// Check automation rules and execute them
	// This could be implemented for automated agent behaviors
}

// EndBlock performs end block logic
func (m *Module) EndBlock(height int64) {
	// No-op for now
}
