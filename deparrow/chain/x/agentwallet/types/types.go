package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
)

const (
	// ModuleName is the name of the module
	ModuleName = "agentwallet"

	// StoreKey is the store key string for agentwallet
	StoreKey = ModuleName

	// RouterKey is the message route for agentwallet
	RouterKey = ModuleName

	// QuerierRoute is the querier route for agentwallet
	QuerierRoute = ModuleName

	// DIDPrefix is the DID prefix for DEparrow agents
	DIDPrefix = "did:deparrow:agent:"
)

// AgentWallet represents an AI agent's wallet
// Note: Detailed implementation in wallet.go
type AgentWallet struct {
	DID              string          `json:"did"`
	Address          string          `json:"address"`
	Balance          sdk.Coin        `json:"balance"`
	SpendingRules    []SpendingRule  `json:"spending_rules"`
	AutomationRules  []AutomationRule `json:"automation_rules"`
	EmergencyReserve sdk.Coin        `json:"emergency_reserve"`
	CreatedAt        int64           `json:"created_at"`
}

// SpendingRule defines rules for spending
// Note: Detailed implementation in rules.go
type SpendingRule struct {
	MaxPerTx    sdk.Coin  `json:"max_per_tx"`
	DailyBudget sdk.Coin  `json:"daily_budget"`
	AllowedOps  []string  `json:"allowed_ops"`
	BlockedOps  []string  `json:"blocked_ops"`
}

// AutomationRule defines automation triggers
// Note: Detailed implementation in rules.go
type AutomationRule struct {
	Trigger string    `json:"trigger"`
	Action  string    `json:"action"`
	Amount  sdk.Coin  `json:"amount"`
	Enabled bool      `json:"enabled"`
}

// Params defines the parameters for the agentwallet module.
type Params struct {
	MaxRulesPerWallet      uint32  `json:"max_rules_per_wallet"`
	MinEmergencyReserve    string  `json:"min_emergency_reserve"`
	AllowExternalTransfers bool    `json:"allow_external_transfers"`
	MaxDailyBudget         string  `json:"max_daily_budget"`
	AutonomousTxEnabled    bool    `json:"autonomous_tx_enabled"`
}

// DefaultParams returns default agentwallet module parameters
func DefaultParams() Params {
	return Params{
		MaxRulesPerWallet:      10,
		MinEmergencyReserve:    "100000000000000000", // 0.1 DPC (18 decimals)
		AllowExternalTransfers: false,
		MaxDailyBudget:         "1000000000000000000000", // 1000 DPC
		AutonomousTxEnabled:    true,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxRulesPerWallet == 0 {
		return fmt.Errorf("max rules per wallet must be positive")
	}

	// Validate min emergency reserve
	if p.MinEmergencyReserve != "" {
		_, ok := math.NewIntFromString(p.MinEmergencyReserve)
		if !ok {
			return fmt.Errorf("invalid min emergency reserve: %s", p.MinEmergencyReserve)
		}
	}

	// Validate max daily budget
	if p.MaxDailyBudget != "" {
		_, ok := math.NewIntFromString(p.MaxDailyBudget)
		if !ok {
			return fmt.Errorf("invalid max daily budget: %s", p.MaxDailyBudget)
		}
	}

	return nil
}

// GetMinEmergencyReserve returns the min emergency reserve as math.Int
func (p Params) GetMinEmergencyReserve() math.Int {
	val, ok := math.NewIntFromString(p.MinEmergencyReserve)
	if !ok {
		return math.ZeroInt()
	}
	return val
}

// GetMaxDailyBudget returns the max daily budget as math.Int
func (p Params) GetMaxDailyBudget() math.Int {
	val, ok := math.NewIntFromString(p.MaxDailyBudget)
	if !ok {
		return math.ZeroInt()
	}
	return val
}
