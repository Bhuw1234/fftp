package types

import (
	"encoding/json"
	"fmt"
)

// GenesisState defines the proofofcompute module's genesis state
type GenesisState struct {
	Params            Params         `json:"params"`
	Jobs              []Job          `json:"jobs"`
	Proofs            []ComputeProof `json:"proofs"`
	TotalSupply       string         `json:"total_supply"`
	CurrentDifficulty uint64         `json:"current_difficulty"`
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:            DefaultParams(),
		Jobs:              []Job{},
		Proofs:            []ComputeProof{},
		TotalSupply:       "0",
		CurrentDifficulty: 1, // Start with difficulty 1
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate all jobs
	for _, job := range gs.Jobs {
		if err := job.Validate(); err != nil {
			return err
		}
	}

	// Validate total supply is a valid integer
	_, err := ValidateTotalSupply(gs.TotalSupply)
	if err != nil {
		return err
	}

	return nil
}

// ValidateTotalSupply validates the total supply string
func ValidateTotalSupply(totalSupply string) (int64, error) {
	if totalSupply == "" {
		return 0, nil
	}
	var supply int64
	_, err := fmt.Sscanf(totalSupply, "%d", &supply)
	if err != nil {
		return 0, fmt.Errorf("invalid total supply: %w", err)
	}
	if supply < 0 {
		return 0, fmt.Errorf("total supply cannot be negative")
	}
	return supply, nil
}

// String implements stringer interface
func (gs GenesisState) String() string {
	bz, _ := json.Marshal(gs)
	return string(bz)
}