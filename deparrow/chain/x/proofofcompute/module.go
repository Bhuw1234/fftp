// Package proofofcompute implements the Proof-of-Compute module for DPC blockchain
package proofofcompute

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/proofofcompute/keeper"
	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// ModuleName is the name of this module
const ModuleName = types.ModuleName

// Module implements the proofofcompute module
type Module struct {
	keeper keeper.Keeper
}

// NewModule creates a new proofofcompute module
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

	// Initialize params
	// Params are stored in keeper

	// Initialize total supply
	if genesis.TotalSupply != "" {
		if err := m.keeper.SetTotalSupply(genesis.TotalSupply); err != nil {
			return fmt.Errorf("failed to set total supply: %w", err)
		}
	}

	// Initialize difficulty
	if genesis.CurrentDifficulty > 0 {
		if err := m.keeper.SetCurrentDifficulty(genesis.CurrentDifficulty); err != nil {
			return fmt.Errorf("failed to set difficulty: %w", err)
		}
	}

	// Initialize jobs from genesis
	for _, job := range genesis.Jobs {
		if err := m.keeper.SetJob(job); err != nil {
			return fmt.Errorf("failed to set genesis job: %w", err)
		}
	}

	// Initialize proofs from genesis
	for _, proof := range genesis.Proofs {
		if err := m.keeper.SetProof(proof); err != nil {
			return fmt.Errorf("failed to set genesis proof: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the module's genesis state
func (m *Module) ExportGenesis() []byte {
	genesis := types.GenesisState{
		Params:       m.keeper.GetParams(),
		Jobs:         m.keeper.GetAllJobs(),
		Proofs:       []types.ComputeProof{}, // TODO: implement GetAllProofs
		TotalSupply:  m.keeper.GetTotalSupply(),
		CurrentDifficulty: m.keeper.GetCurrentDifficulty(),
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

// EndBlock performs end block logic (difficulty adjustment)
func (m *Module) EndBlock(height int64) {
	params := m.keeper.GetParams()

	// Adjust difficulty periodically
	if height > 0 && uint64(height)%params.DifficultyAdjustment == 0 {
		stats := m.keeper.GetStats()
		jobs := stats["total_jobs"].(int)
		completed := stats["completed_jobs"].(int)

		currentDiff := m.keeper.GetCurrentDifficulty()

		// If completion rate > 50%, increase difficulty
		if jobs > 0 && completed > jobs/2 {
			m.keeper.SetCurrentDifficulty(currentDiff + 1)
		} else if currentDiff > 1 {
			// If completion rate < 50%, decrease difficulty
			m.keeper.SetCurrentDifficulty(currentDiff - 1)
		}
	}
}
