package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
)

// SpendingRuleExtended extends SpendingRule with additional metadata
type SpendingRuleExtended struct {
	SpendingRule
	ID          uint32 `json:"id"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// AutomationRuleExtended extends AutomationRule with additional metadata
type AutomationRuleExtended struct {
	AutomationRule
	ID           uint32 `json:"id"`
	Description  string `json:"description,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	LastTrigger  int64  `json:"last_trigger,omitempty"`
	TriggerCount uint64 `json:"trigger_count"`
}

// NewSpendingRule creates a new spending rule with default values
func NewSpendingRule(maxPerTx, dailyBudget sdk.Coin, allowedOps, blockedOps []string) SpendingRule {
	return SpendingRule{
		MaxPerTx:    maxPerTx,
		DailyBudget: dailyBudget,
		AllowedOps:  allowedOps,
		BlockedOps:  blockedOps,
	}
}

// DefaultSpendingRule returns a default spending rule for new wallets
func DefaultSpendingRule() SpendingRule {
	maxPerTx, _ := math.NewIntFromString("1000000000000000000")    // 1 DPC
	dailyBudget, _ := math.NewIntFromString("10000000000000000000") // 10 DPC
	return SpendingRule{
		MaxPerTx:    sdk.NewCoin("dpc", maxPerTx),
		DailyBudget: sdk.NewCoin("dpc", dailyBudget),
		AllowedOps:  []string{OperationSubmitJob, OperationPayService, OperationBuyCompute},
		BlockedOps:  []string{OperationExternalTransfer},
	}
}

// Validate validates a spending rule
func (r SpendingRule) Validate() error {
	if r.MaxPerTx.IsNegative() {
		return fmt.Errorf("max per tx cannot be negative")
	}
	if r.DailyBudget.IsNegative() {
		return fmt.Errorf("daily budget cannot be negative")
	}

	// Validate allowed operations
	for _, op := range r.AllowedOps {
		if !isValidOperation(op) {
			return fmt.Errorf("invalid allowed operation: %s", op)
		}
	}

	// Validate blocked operations
	for _, op := range r.BlockedOps {
		if !isValidOperation(op) {
			return fmt.Errorf("invalid blocked operation: %s", op)
		}
	}

	// Check for conflicts (operation in both allowed and blocked)
	for _, allowed := range r.AllowedOps {
		for _, blocked := range r.BlockedOps {
			if allowed == blocked {
				return fmt.Errorf("operation %s cannot be both allowed and blocked", allowed)
			}
		}
	}

	return nil
}

// NewAutomationRule creates a new automation rule
func NewAutomationRule(trigger, action string, amount sdk.Coin, enabled bool) AutomationRule {
	return AutomationRule{
		Trigger: trigger,
		Action:  action,
		Amount:  amount,
		Enabled: enabled,
	}
}

// DefaultAutomationRules returns default automation rules for new wallets
func DefaultAutomationRules() []AutomationRule {
	amount1, _ := math.NewIntFromString("500000000000000000")  // 0.5 DPC
	amount2, _ := math.NewIntFromString("100000000000000000")  // 0.1 DPC
	return []AutomationRule{
		{
			Trigger: TriggerBalanceBelow,
			Action:  ActionTransferToReserve,
			Amount:  sdk.NewCoin("dpc", amount1),
			Enabled: true,
		},
		{
			Trigger: TriggerEmergencyLow,
			Action:  ActionRequestFunding,
			Amount:  sdk.NewCoin("dpc", amount2),
			Enabled: true,
		},
	}
}

// Validate validates an automation rule
func (r AutomationRule) Validate() error {
	if !isValidTrigger(r.Trigger) {
		return ErrInvalidTrigger
	}
	if !isValidAction(r.Action) {
		return ErrInvalidAction
	}
	if r.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}
	return nil
}

// isValidOperation checks if an operation is valid
func isValidOperation(op string) bool {
	validOps := map[string]bool{
		OperationSubmitJob:       true,
		OperationPayService:      true,
		OperationBuyCompute:      true,
		OperationTransfer:        true,
		OperationExternalTransfer: true,
		OperationWithdraw:        true,
		OperationDeposit:         true,
	}
	return validOps[op]
}

// isValidTrigger checks if a trigger is valid
func isValidTrigger(trigger string) bool {
	validTriggers := map[string]bool{
		TriggerBalanceBelow: true,
		TriggerBalanceAbove: true,
		TriggerJobCompleted: true,
		TriggerIdleTimeout:  true,
		TriggerScheduled:    true,
		TriggerDailyReset:   true,
		TriggerEmergencyLow: true,
	}
	return validTriggers[trigger]
}

// isValidAction checks if an action is valid
func isValidAction(action string) bool {
	validActions := map[string]bool{
		ActionBuyCompute:       true,
		ActionTransferToReserve: true,
		ActionNotify:           true,
		ActionPauseOperations:  true,
		ActionResumeOperations: true,
		ActionRequestFunding:   true,
	}
	return validActions[action]
}

// ShouldTrigger checks if the automation rule should trigger based on wallet state
func (r AutomationRule) ShouldTrigger(wallet *AgentWalletExtended) bool {
	if !r.Enabled {
		return false
	}

	availableBalance := wallet.GetAvailableBalance()

	switch r.Trigger {
	case TriggerBalanceBelow:
		return availableBalance.Amount.LT(r.Amount.Amount)
	case TriggerBalanceAbove:
		return availableBalance.Amount.GTE(r.Amount.Amount)
	case TriggerEmergencyLow:
		return availableBalance.Amount.LT(wallet.EmergencyReserve.Amount)
	case TriggerIdleTimeout:
		// Check if wallet has been idle for too long
		// This would need actual implementation with time tracking
		return false
	default:
		return false
	}
}

// String returns a string representation of the spending rule
func (r SpendingRule) String() string {
	return fmt.Sprintf("SpendingRule{MaxPerTx: %s, DailyBudget: %s, Allowed: %v, Blocked: %v}",
		r.MaxPerTx.String(), r.DailyBudget.String(), r.AllowedOps, r.BlockedOps)
}

// String returns a string representation of the automation rule
func (r AutomationRule) String() string {
	return fmt.Sprintf("AutomationRule{Trigger: %s, Action: %s, Amount: %s, Enabled: %v}",
		r.Trigger, r.Action, r.Amount.String(), r.Enabled)
}

// GetMinEmergencyReserve returns the min emergency reserve as math.Int
func GetMinEmergencyReserve(p Params) (math.Int, error) {
	val, ok := math.NewIntFromString(p.MinEmergencyReserve)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid min emergency reserve")
	}
	return val, nil
}