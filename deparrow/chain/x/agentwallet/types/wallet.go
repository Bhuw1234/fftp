package types

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
	"github.com/google/uuid"
)

// DIDRegex validates the DID format: did:deparrow:agent:<uuid>
var DIDRegex = regexp.MustCompile(`^did:deparrow:agent:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// AgentWalletExtended extends AgentWallet with additional metadata
type AgentWalletExtended struct {
	AgentWallet
	DailySpent        sdk.Coin `json:"daily_spent"`
	LastActivityTime  int64    `json:"last_activity_time"`
	TotalTransactions uint64   `json:"total_transactions"`
	IsPaused          bool     `json:"is_paused"`
	PausedReason      string   `json:"paused_reason,omitempty"`
}

// NewAgentWallet creates a new AgentWallet instance
func NewAgentWallet(did, address string) *AgentWallet {
	maxPerTx, _ := math.NewIntFromString("1000000000000000000")    // 1 DPC
	dailyBudget, _ := math.NewIntFromString("10000000000000000000") // 10 DPC
	emergencyReserve, _ := math.NewIntFromString("100000000000000000") // 0.1 DPC

	return &AgentWallet{
		DID:     did,
		Address: address,
		Balance: sdk.NewCoin("dpc", math.ZeroInt()),
		SpendingRules: []SpendingRule{
			{
				MaxPerTx:    sdk.NewCoin("dpc", maxPerTx),
				DailyBudget: sdk.NewCoin("dpc", dailyBudget),
				AllowedOps:  []string{OperationSubmitJob, OperationPayService, OperationBuyCompute},
				BlockedOps:  []string{OperationExternalTransfer},
			},
		},
		AutomationRules:  []AutomationRule{},
		EmergencyReserve: sdk.NewCoin("dpc", emergencyReserve),
		CreatedAt:        time.Now().Unix(),
	}
}

// NewAgentWalletExtended creates a new AgentWalletExtended instance
func NewAgentWalletExtended(did, address string) *AgentWalletExtended {
	return &AgentWalletExtended{
		AgentWallet:       *NewAgentWallet(did, address),
		DailySpent:        sdk.NewCoin("dpc", math.ZeroInt()),
		LastActivityTime:  time.Now().Unix(),
		TotalTransactions: 0,
		IsPaused:          false,
	}
}

// Validate performs basic validation of the wallet
func (w AgentWallet) Validate() error {
	if w.DID == "" {
		return fmt.Errorf("DID cannot be empty")
	}
	if !IsValidDID(w.DID) {
		return ErrInvalidDID
	}
	if w.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(w.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	if w.Balance.IsNegative() {
		return fmt.Errorf("balance cannot be negative")
	}
	if w.EmergencyReserve.IsNegative() {
		return fmt.Errorf("emergency reserve cannot be negative")
	}
	if w.CreatedAt <= 0 {
		return fmt.Errorf("invalid creation timestamp")
	}
	return nil
}

// Validate performs validation of extended wallet
func (w AgentWalletExtended) Validate() error {
	if err := w.AgentWallet.Validate(); err != nil {
		return err
	}
	if w.DailySpent.IsNegative() {
		return fmt.Errorf("daily spent cannot be negative")
	}
	return nil
}

// IsValidDID validates the DID format
func IsValidDID(did string) bool {
	return DIDRegex.MatchString(did)
}

// GenerateDID generates a new DID for an agent
func GenerateDID() string {
	return DIDPrefix + uuid.New().String()
}

// ParseDID extracts the UUID from a DID
func ParseDID(did string) (string, error) {
	if !IsValidDID(did) {
		return "", ErrInvalidDID
	}
	parts := strings.Split(did, ":")
	if len(parts) != 4 {
		return "", ErrInvalidDID
	}
	return parts[3], nil
}

// CanSpend checks if the wallet can spend the given amount
func (w AgentWallet) CanSpend(amount sdk.Coin, operation string) bool {
	// Check balance (excluding emergency reserve)
	availableBalance := w.GetAvailableBalance()
	if availableBalance.Amount.LT(amount.Amount) {
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
		if !rule.MaxPerTx.IsZero() && amount.Amount.GT(rule.MaxPerTx.Amount) {
			return false
		}
	}

	return true
}

// CanSpendWithDaily checks spending with daily budget
func (w AgentWalletExtended) CanSpendWithDaily(amount sdk.Coin, operation string) bool {
	// First check basic spending rules
	if !w.AgentWallet.CanSpend(amount, operation) {
		return false
	}

	// Check if wallet is paused
	if w.IsPaused {
		return false
	}

	// Check daily budget
	for _, rule := range w.SpendingRules {
		if !rule.DailyBudget.IsZero() {
			potentialDaily := w.DailySpent.Add(amount)
			if potentialDaily.Amount.GT(rule.DailyBudget.Amount) {
				return false
			}
		}
	}

	return true
}

// GetAvailableBalance returns balance minus emergency reserve
func (w AgentWallet) GetAvailableBalance() sdk.Coin {
	if w.Balance.Amount.LT(w.EmergencyReserve.Amount) {
		return sdk.NewCoin("dpc", math.ZeroInt())
	}
	return sdk.NewCoin("dpc", w.Balance.Amount.Sub(w.EmergencyReserve.Amount))
}

// IsOperationBlocked checks if an operation is blocked
func (w AgentWallet) IsOperationBlocked(operation string) bool {
	for _, rule := range w.SpendingRules {
		for _, blocked := range rule.BlockedOps {
			if blocked == operation {
				return true
			}
		}
	}
	return false
}

// IsOperationAllowed checks if an operation is allowed
func (w AgentWallet) IsOperationAllowed(operation string) bool {
	for _, rule := range w.SpendingRules {
		// Empty allowed list means all operations are allowed
		if len(rule.AllowedOps) == 0 {
			return true
		}
		for _, allowed := range rule.AllowedOps {
			if allowed == operation {
				return true
			}
		}
	}
	return false
}

// GetMaxPerTx returns the maximum per-transaction amount
func (w AgentWallet) GetMaxPerTx() sdk.Coin {
	for _, rule := range w.SpendingRules {
		if !rule.MaxPerTx.IsZero() {
			return rule.MaxPerTx
		}
	}
	return sdk.NewCoin("dpc", math.ZeroInt()) // No limit
}

// GetDailyBudget returns the daily budget
func (w AgentWallet) GetDailyBudget() sdk.Coin {
	for _, rule := range w.SpendingRules {
		if !rule.DailyBudget.IsZero() {
			return rule.DailyBudget
		}
	}
	return sdk.NewCoin("dpc", math.ZeroInt()) // No limit
}

// String returns a string representation of the wallet
func (w AgentWallet) String() string {
	return fmt.Sprintf("AgentWallet{DID: %s, Address: %s, Balance: %s, Reserve: %s}",
		w.DID, w.Address, w.Balance.String(), w.EmergencyReserve.String())
}
