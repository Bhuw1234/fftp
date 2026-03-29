// Package computemarket implements the Compute Market module for DPC blockchain
package computemarket

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/computemarket/keeper"
	"github.com/deparrow/dpc/x/computemarket/types"
)

// ModuleName is the name of this module
const ModuleName = types.ModuleName

// Module implements the computemarket module
type Module struct {
	keeper keeper.Keeper
}

// NewModule creates a new computemarket module
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

	// Initialize total staked
	if genesis.TotalStaked != "" {
		if err := m.keeper.SetTotalStaked(genesis.TotalStaked); err != nil {
			return fmt.Errorf("failed to set total staked: %w", err)
		}
	}

	// Initialize total escrowed
	if genesis.TotalEscrowed != "" {
		if err := m.keeper.SetTotalEscrowed(genesis.TotalEscrowed); err != nil {
			return fmt.Errorf("failed to set total escrowed: %w", err)
		}
	}

	// Initialize providers from genesis
	for _, provider := range genesis.Providers {
		if err := m.keeper.SetProvider(provider); err != nil {
			return fmt.Errorf("failed to set genesis provider: %w", err)
		}
	}

	// Initialize escrows from genesis
	for _, escrow := range genesis.Escrows {
		if err := m.keeper.SetEscrow(escrow); err != nil {
			return fmt.Errorf("failed to set genesis escrow: %w", err)
		}
	}

	// Initialize disputes from genesis
	for _, dispute := range genesis.Disputes {
		if err := m.keeper.SetDispute(dispute); err != nil {
			return fmt.Errorf("failed to set genesis dispute: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the module's genesis state
func (m *Module) ExportGenesis() []byte {
	genesis := types.GenesisState{
		Params:        m.keeper.GetParams(),
		Providers:     m.keeper.GetAllProviders(),
		Escrows:       m.keeper.GetAllEscrows(),
		Disputes:      m.keeper.GetAllDisputes(),
		TotalStaked:   m.keeper.GetTotalStaked(),
		TotalEscrowed: m.keeper.GetTotalEscrowed(),
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
	// No-op for now
}

// EndBlock performs end block logic
func (m *Module) EndBlock(height int64) {
	// Check for expired escrows and refund them
	// This could be implemented for automatic dispute resolution
}
