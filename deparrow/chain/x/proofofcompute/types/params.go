package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default proofofcompute module parameters
func DefaultParams() Params {
	return Params{
		MinComputeUnits:      1,
		RewardPerUnit:        "0.001",  // 0.001 DPC per compute unit (base rate)
		DifficultyAdjustment: 1000,     // Adjust every 1000 blocks
		TargetBlockTime:      6,        // 6 seconds (CometBFT default)
		MaxSupply:            "21000000000000000000000000000", // 21B DPC with 18 decimals
		ComplexityMultiplier: 5,        // Max 5x complexity multiplier
		MinStake:             "1000000000000000000", // 1 DPC minimum stake
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MinComputeUnits == 0 {
		return fmt.Errorf("min compute units must be positive")
	}
	if p.TargetBlockTime == 0 {
		return fmt.Errorf("target block time must be positive")
	}
	if p.DifficultyAdjustment == 0 {
		return fmt.Errorf("difficulty adjustment period must be positive")
	}
	if p.ComplexityMultiplier == 0 {
		return fmt.Errorf("complexity multiplier must be positive")
	}

	// Validate reward per unit is a valid decimal
	_, err := sdk.NewDecFromStr(p.RewardPerUnit)
	if err != nil {
		return fmt.Errorf("invalid reward per unit: %w", err)
	}

	// Validate max supply is a valid integer
	_, err = sdk.NewIntFromString(p.MaxSupply)
	if err != nil {
		return fmt.Errorf("invalid max supply: %w", err)
	}

	// Validate min stake is a valid integer
	_, err = sdk.NewIntFromString(p.MinStake)
	if err != nil {
		return fmt.Errorf("invalid min stake: %w", err)
	}

	return nil
}

// GetRewardPerUnit returns the reward per unit as sdk.Dec
func (p Params) GetRewardPerUnit() (sdk.Dec, error) {
	return sdk.NewDecFromStr(p.RewardPerUnit)
}

// GetMaxSupply returns the max supply as sdk.Int
func (p Params) GetMaxSupply() (sdk.Int, error) {
	return sdk.NewIntFromString(p.MaxSupply)
}

// GetMinStake returns the min stake as sdk.Int
func (p Params) GetMinStake() (sdk.Int, error) {
	return sdk.NewIntFromString(p.MinStake)
}

// String implements stringer interface
func (p Params) String() string {
	bz, _ := json.Marshal(p)
	return string(bz)
}
