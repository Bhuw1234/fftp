package types

import "encoding/json"

// Params defines the parameters for the agentwallet module
type Params struct {
	// MaxRulesPerWallet is the maximum number of rules per wallet
	MaxRulesPerWallet uint32 `json:"max_rules_per_wallet"`
	// MinEmergencyReserve is the minimum emergency reserve
	MinEmergencyReserve string `json:"min_emergency_reserve"`
	// AllowExternalTransfers allows transfers to external addresses
	AllowExternalTransfers bool `json:"allow_external_transfers"`
}

// DefaultParams returns the default module parameters
func DefaultParams() Params {
	return Params{
		MaxRulesPerWallet:     10,
		MinEmergencyReserve:   "1000000000000000000", // 1 DPC
		AllowExternalTransfers: true,
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
