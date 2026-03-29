package types

import (
	"encoding/json"
)

// GenesisState defines the proofofcompute module's genesis state
type GenesisState struct {
	Params            Params         `json:"params"`
	Jobs              []Job          `json:"jobs"`
	Proofs            []ComputeProof `json:"proofs"`
	TotalSupply       string         `json:"total_supply"`
	CurrentDifficulty uint64         `json:"current_difficulty"`
}

// NewGenesisState creates a new genesis state
func NewGenesisState(params Params, jobs []Job, proofs []ComputeProof, totalSupply string, difficulty uint64) GenesisState {
	return GenesisState{
		Params:            params,
		Jobs:              jobs,
		Proofs:            proofs,
		TotalSupply:       totalSupply,
		CurrentDifficulty: difficulty,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:            DefaultParams(),
		Jobs:              []Job{},
		Proofs:            []ComputeProof{},
		TotalSupply:       "1000000000000000000000000000", // 1B initial
		CurrentDifficulty: 1,
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}

// Bytes serializes the genesis state
func (gs GenesisState) Bytes() []byte {
	bz, _ := json.Marshal(gs)
	return bz
}
