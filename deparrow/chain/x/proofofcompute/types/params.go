package types

import (
	"encoding/json"
)

// Params defines the parameters for the proofofcompute module
type Params struct {
	// MinComputeUnits is the minimum compute units required for reward
	MinComputeUnits uint64 `json:"min_compute_units"`
	// RewardPerUnit is the base DPC reward per compute unit (in base units, 18 decimals)
	RewardPerUnit string `json:"reward_per_unit"`
	// MaxSupply is the maximum DPC supply (21B with 18 decimals)
	MaxSupply string `json:"max_supply"`
	// ComplexityMultiplier is the max complexity multiplier (default 5)
	ComplexityMultiplier uint32 `json:"complexity_multiplier"`
	// MinStake is the minimum stake required for job submission (in base units)
	MinStake string `json:"min_stake"`
	// DifficultyAdjustment is the period (in blocks) for difficulty adjustment
	DifficultyAdjustment uint64 `json:"difficulty_adjustment"`
}

// DefaultParams returns the default module parameters
func DefaultParams() Params {
	return Params{
		MinComputeUnits:       1,
		RewardPerUnit:         "1000000000000000", // 0.001 DPC (18 decimals)
		MaxSupply:             "21000000000000000000000000000", // 21B DPC
		ComplexityMultiplier:  5,
		MinStake:              "1000000000000000000", // 1 DPC minimum
		DifficultyAdjustment:  100, // Adjust every 100 blocks
	}
}

// Validate validates the params
func (p Params) Validate() error {
	return nil
}

// Bytes serializes params to bytes
func (p Params) Bytes() []byte {
	bz, _ := json.Marshal(p)
	return bz
}
