package types

import (
	"encoding/json"
	"fmt"
)

// GenesisState defines the agentwallet module's genesis state
type GenesisState struct {
	Params           Params                      `json:"params"`
	Wallets          []AgentWalletExtended       `json:"wallets"`
	Agents           []AgentInfo                 `json:"agents"`
	PendingAutomations []PendingAutomation       `json:"pending_automations"`
	TotalWallets     uint64                      `json:"total_wallets"`
	TotalAgents      uint64                      `json:"total_agents"`
}

// AgentInfo stores agent registration information
type AgentInfo struct {
	DID         string `json:"did"`
	Address     string `json:"address"`
	AgentType   string `json:"agent_type"`
	Metadata    string `json:"metadata"`
	RegisteredAt int64 `json:"registered_at"`
	IsActive    bool   `json:"is_active"`
}

// PendingAutomation stores pending automation triggers
type PendingAutomation struct {
	DID       string `json:"did"`
	Address   string `json:"address"`
	Trigger   string `json:"trigger"`
	Action    string `json:"action"`
	Amount    string `json:"amount"`
	Scheduled int64  `json:"scheduled"`
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		Wallets:            []AgentWalletExtended{},
		Agents:             []AgentInfo{},
		PendingAutomations: []PendingAutomation{},
		TotalWallets:       0,
		TotalAgents:        0,
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate all wallets
	walletSet := make(map[string]bool)
	addressSet := make(map[string]bool)
	for _, wallet := range gs.Wallets {
		if err := wallet.Validate(); err != nil {
			return fmt.Errorf("invalid wallet %s: %w", wallet.DID, err)
		}
		if walletSet[wallet.DID] {
			return fmt.Errorf("duplicate wallet DID: %s", wallet.DID)
		}
		if addressSet[wallet.Address] {
			return fmt.Errorf("duplicate wallet address: %s", wallet.Address)
		}
		walletSet[wallet.DID] = true
		addressSet[wallet.Address] = true
	}

	// Validate all agents
	agentSet := make(map[string]bool)
	for _, agent := range gs.Agents {
		if err := agent.Validate(); err != nil {
			return fmt.Errorf("invalid agent %s: %w", agent.DID, err)
		}
		if agentSet[agent.DID] {
			return fmt.Errorf("duplicate agent DID: %s", agent.DID)
		}
		agentSet[agent.DID] = true

		// Verify agent has a corresponding wallet
		if !walletSet[agent.DID] {
			return fmt.Errorf("agent %s has no corresponding wallet", agent.DID)
		}
	}

	// Validate pending automations
	for _, automation := range gs.PendingAutomations {
		if err := automation.Validate(); err != nil {
			return fmt.Errorf("invalid pending automation: %w", err)
		}
	}

	// Validate counts
	if uint64(len(gs.Wallets)) != gs.TotalWallets {
		return fmt.Errorf("wallet count mismatch: expected %d, got %d", gs.TotalWallets, len(gs.Wallets))
	}
	if uint64(len(gs.Agents)) != gs.TotalAgents {
		return fmt.Errorf("agent count mismatch: expected %d, got %d", gs.TotalAgents, len(gs.Agents))
	}

	return nil
}

// Validate validates agent info
func (a AgentInfo) Validate() error {
	if !IsValidDID(a.DID) {
		return ErrInvalidDID
	}
	if a.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if a.AgentType == "" {
		return fmt.Errorf("agent type cannot be empty")
	}
	if a.RegisteredAt <= 0 {
		return fmt.Errorf("invalid registration timestamp")
	}
	return nil
}

// Validate validates pending automation
func (a PendingAutomation) Validate() error {
	if !IsValidDID(a.DID) {
		return ErrInvalidDID
	}
	if a.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if !isValidTrigger(a.Trigger) {
		return ErrInvalidTrigger
	}
	if !isValidAction(a.Action) {
		return ErrInvalidAction
	}
	if a.Scheduled <= 0 {
		return fmt.Errorf("invalid scheduled timestamp")
	}
	return nil
}

// String implements stringer interface
func (gs GenesisState) String() string {
	bz, _ := json.Marshal(gs)
	return string(bz)
}

// String implements stringer interface
func (a AgentInfo) String() string {
	return fmt.Sprintf("AgentInfo{DID: %s, Type: %s, Active: %v}", a.DID, a.AgentType, a.IsActive)
}

// String implements stringer interface
func (a PendingAutomation) String() string {
	return fmt.Sprintf("PendingAutomation{DID: %s, Trigger: %s, Action: %s}", a.DID, a.Trigger, a.Action)
}
