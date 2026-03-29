package types

import "encoding/json"

// Coin represents a token amount with denom
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// SpendingRule defines rules for spending
type SpendingRule struct {
	MaxPerTx    Coin     `json:"max_per_tx"`
	DailyBudget Coin     `json:"daily_budget"`
	AllowedOps  []string `json:"allowed_ops"`
	BlockedOps  []string `json:"blocked_ops"`
}

// AutomationRule defines automation triggers
type AutomationRule struct {
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
	Amount  Coin   `json:"amount"`
	Enabled bool   `json:"enabled"`
}

// AgentWallet represents an AI agent's wallet
type AgentWallet struct {
	DID              string           `json:"did"`
	Address          string           `json:"address"`
	Balance          Coin             `json:"balance"`
	SpendingRules    []SpendingRule   `json:"spending_rules"`
	AutomationRules  []AutomationRule `json:"automation_rules"`
	EmergencyReserve Coin             `json:"emergency_reserve"`
	CreatedAt        int64            `json:"created_at"`
}

// GenesisState defines the agentwallet module's genesis state
type GenesisState struct {
	Params   Params        `json:"params"`
	Wallets  []AgentWallet `json:"wallets"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:  DefaultParams(),
		Wallets: []AgentWallet{},
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
