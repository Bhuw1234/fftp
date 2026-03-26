package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
type AgentWallet struct {
	DID              string
	Address          string
	Balance          sdk.Coin
	SpendingRules    []SpendingRule
	AutomationRules  []AutomationRule
	EmergencyReserve sdk.Coin
	CreatedAt        int64
}

// SpendingRule defines rules for spending
type SpendingRule struct {
	MaxPerTx    sdk.Coin
	DailyBudget sdk.Coin
	AllowedOps  []string // ["submit_job", "pay_service"]
	BlockedOps  []string // ["external_transfer"]
}

// AutomationRule defines automation triggers
type AutomationRule struct {
	Trigger string        // "balance_below", "job_completed", etc.
	Action  string        // "buy_compute", "transfer_to_reserve"
	Amount  sdk.Coin      // Amount for the action
	Enabled bool          // Whether the rule is active
}

// NewAgentWallet creates a new AgentWallet instance
func NewAgentWallet(did, address string) *AgentWallet {
	return &AgentWallet{
		DID:     did,
		Address: address,
		Balance: sdk.NewCoin("dpc", sdk.ZeroInt()),
		SpendingRules: []SpendingRule{
			{
				MaxPerTx:    sdk.NewCoin("dpc", sdk.NewInt(1000)),
				DailyBudget: sdk.NewCoin("dpc", sdk.NewInt(10000)),
				AllowedOps:  []string{"submit_job", "pay_service"},
				BlockedOps:  []string{},
			},
		},
		AutomationRules:  []AutomationRule{},
		EmergencyReserve: sdk.NewCoin("dpc", sdk.NewInt(100)),
	}
}

// Validate performs basic validation of the wallet
func (w AgentWallet) Validate() error {
	if w.DID == "" {
		return fmt.Errorf("DID cannot be empty")
	}
	if w.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if w.Balance.IsNegative() {
		return fmt.Errorf("balance cannot be negative")
	}
	if w.EmergencyReserve.IsNegative() {
		return fmt.Errorf("emergency reserve cannot be negative")
	}
	return nil
}

// CanSpend checks if the wallet can spend the given amount
func (w AgentWallet) CanSpend(amount sdk.Coin, operation string) bool {
	// Check balance
	if w.Balance.IsLT(amount) {
		return false
	}

	// Check spending rules
	for _, rule := range w.SpendingRules {
		// Check if operation is blocked
		for _, blocked := range rule.BlockedOps {
			if blocked == operation {
				return false
			}
		}

		// Check if operation is allowed
		allowed := len(rule.AllowedOps) == 0 // Empty means all allowed
		for _, op := range rule.AllowedOps {
			if op == operation {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}

		// Check max per transaction
		if !rule.MaxPerTx.IsZero() && amount.IsGT(rule.MaxPerTx) {
			return false
		}
	}

	return true
}

// Params defines the parameters for the agentwallet module.
type Params struct {
	MaxRulesPerWallet    uint32
	MinEmergencyReserve  sdk.Coin
	AllowExternalTransfers bool
}

// DefaultParams returns default agentwallet module parameters
func DefaultParams() Params {
	return Params{
		MaxRulesPerWallet:     10,
		MinEmergencyReserve:   sdk.NewCoin("dpc", sdk.NewInt(100)),
		AllowExternalTransfers: false,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MinEmergencyReserve.IsNegative() {
		return fmt.Errorf("min emergency reserve cannot be negative")
	}
	return nil
}
