package types

import (
	"encoding/json"
	"fmt"
)

// GenesisState defines the computemarket module's genesis state
type GenesisState struct {
	Params        Params             `json:"params"`
	Providers     []ProviderExtended `json:"providers"`
	Escrows       []EscrowExtended   `json:"escrows"`
	Disputes      []Dispute          `json:"disputes"`
	JobMatches    []JobMatch         `json:"job_matches"`
	TotalStaked   string             `json:"total_staked"`
	TotalEscrowed string             `json:"total_escrowed"`
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		Providers:     []ProviderExtended{},
		Escrows:       []EscrowExtended{},
		Disputes:      []Dispute{},
		JobMatches:    []JobMatch{},
		TotalStaked:   "0",
		TotalEscrowed: "0",
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate all providers
	providerSet := make(map[string]bool)
	for _, provider := range gs.Providers {
		if err := provider.Validate(); err != nil {
			return fmt.Errorf("invalid provider %s: %w", provider.Address, err)
		}
		if providerSet[provider.Address] {
			return fmt.Errorf("duplicate provider address: %s", provider.Address)
		}
		providerSet[provider.Address] = true
	}

	// Validate all escrows
	escrowSet := make(map[string]bool)
	for _, escrow := range gs.Escrows {
		if err := escrow.Validate(); err != nil {
			return fmt.Errorf("invalid escrow %s: %w", escrow.ID, err)
		}
		if escrowSet[escrow.ID] {
			return fmt.Errorf("duplicate escrow ID: %s", escrow.ID)
		}
		escrowSet[escrow.ID] = true
	}

	// Validate all disputes
	disputeSet := make(map[string]bool)
	for _, dispute := range gs.Disputes {
		if err := dispute.Validate(); err != nil {
			return fmt.Errorf("invalid dispute %s: %w", dispute.ID, err)
		}
		if disputeSet[dispute.ID] {
			return fmt.Errorf("duplicate dispute ID: %s", dispute.ID)
		}
		disputeSet[dispute.ID] = true
	}

	// Validate total staked
	if gs.TotalStaked != "" {
		_, err := ValidateStakeAmount(gs.TotalStaked)
		if err != nil {
			return fmt.Errorf("invalid total staked: %w", err)
		}
	}

	// Validate total escrowed
	if gs.TotalEscrowed != "" {
		_, err := ValidateStakeAmount(gs.TotalEscrowed)
		if err != nil {
			return fmt.Errorf("invalid total escrowed: %w", err)
		}
	}

	return nil
}

// ValidateStakeAmount validates a stake amount string
func ValidateStakeAmount(amount string) (int64, error) {
	if amount == "" {
		return 0, nil
	}
	var stake int64
	_, err := fmt.Sscanf(amount, "%d", &stake)
	if err != nil {
		return 0, fmt.Errorf("invalid stake amount: %w", err)
	}
	if stake < 0 {
		return 0, fmt.Errorf("stake amount cannot be negative")
	}
	return stake, nil
}

// String implements stringer interface
func (gs GenesisState) String() string {
	bz, _ := json.Marshal(gs)
	return string(bz)
}
