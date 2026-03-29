package types

import "encoding/json"

// Params defines the parameters for the computemarket module
type Params struct {
	// MinStake is the minimum stake required for providers
	MinStake string `json:"min_stake"`
	// DisputePeriod is the dispute period in blocks
	DisputePeriod uint32 `json:"dispute_period"`
	// MaxJobDuration is the maximum job duration in seconds
	MaxJobDuration uint32 `json:"max_job_duration"`
	// NetworkFee is the network fee percentage
	NetworkFee string `json:"network_fee"`
	// MinReputation is the minimum reputation to accept jobs
	MinReputation uint32 `json:"min_reputation"`
	// SlashPercent is the percentage to slash for failed jobs
	SlashPercent uint32 `json:"slash_percent"`
	// DisputeSlashPercent is the percentage to slash when losing dispute
	DisputeSlashPercent uint32 `json:"dispute_slash_percent"`
}

// DefaultParams returns the default module parameters
func DefaultParams() Params {
	return Params{
		MinStake:            "100000000000000000000", // 100 DPC
		DisputePeriod:       100,                     // 100 blocks
		MaxJobDuration:      86400,                   // 24 hours
		NetworkFee:          "0.01",                  // 1%
		MinReputation:       100,                     // Min reputation score
		SlashPercent:        10,                      // 10% slash for failed jobs
		DisputeSlashPercent: 50,                      // 50% slash for disputes
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
