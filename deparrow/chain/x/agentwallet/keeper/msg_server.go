// Package keeper implements the message server for the agentwallet module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// MsgCreateWallet represents a wallet creation message
type MsgCreateWallet struct {
	Address string `json:"address"`
}

// MsgCreateWalletResponse is the response for wallet creation
type MsgCreateWalletResponse struct {
	DID     string `json:"did"`
	Address string `json:"address"`
}

// MsgAddSpendingRule represents a spending rule addition message
type MsgAddSpendingRule struct {
	Address       string            `json:"address"`
	SpendingRule  types.SpendingRule `json:"spending_rule"`
}

// MsgAddSpendingRuleResponse is the response for rule addition
type MsgAddSpendingRuleResponse struct {
	Address string `json:"address"`
	Success bool   `json:"success"`
}

// MsgAddAutomationRule represents an automation rule addition message
type MsgAddAutomationRule struct {
	Address         string             `json:"address"`
	AutomationRule  types.AutomationRule `json:"automation_rule"`
}

// MsgAddAutomationRuleResponse is the response for automation rule addition
type MsgAddAutomationRuleResponse struct {
	Address string `json:"address"`
	Success bool   `json:"success"`
}

// MsgDeposit represents a deposit message
type MsgDeposit struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// MsgDepositResponse is the response for deposit
type MsgDepositResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// MsgWithdraw represents a withdrawal message
type MsgWithdraw struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// MsgWithdrawResponse is the response for withdrawal
type MsgWithdrawResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// MsgTransfer represents a transfer message
type MsgTransfer struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
}

// MsgTransferResponse is the response for transfer
type MsgTransferResponse struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
}

// CreateWallet handles wallet creation
func (k Keeper) CreateWallet(msg MsgCreateWallet, blockHeight int64) (*MsgCreateWalletResponse, error) {
	// Check if wallet already exists
	if _, found := k.GetWallet(msg.Address); found {
		return nil, types.ErrWalletAlreadyExists
	}

	// Generate DID
	didSuffix, err := k.IncrementDIDSuffix()
	if err != nil {
		return nil, fmt.Errorf("failed to generate DID: %w", err)
	}
	did := fmt.Sprintf("did:dpc:agent:%s", didSuffix)

	// Create wallet
	wallet := types.AgentWallet{
		DID:     did,
		Address: msg.Address,
		Balance: types.Coin{Denom: "dpc", Amount: "0"},
		SpendingRules: []types.SpendingRule{},
		AutomationRules: []types.AutomationRule{},
		EmergencyReserve: types.Coin{Denom: "dpc", Amount: "0"},
		CreatedAt: blockHeight,
	}

	if err := k.SetWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	log.Printf("[agentwallet] Wallet created for %s with DID %s", msg.Address, did)

	return &MsgCreateWalletResponse{
		DID:     did,
		Address: msg.Address,
	}, nil
}

// AddSpendingRule handles spending rule addition
func (k Keeper) AddSpendingRule(msg MsgAddSpendingRule) (*MsgAddSpendingRuleResponse, error) {
	wallet, found := k.GetWallet(msg.Address)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	// Check max rules
	params := k.GetParams()
	if len(wallet.SpendingRules) >= int(params.MaxRulesPerWallet) {
		return nil, types.ErrMaxRulesExceeded
	}

	wallet.SpendingRules = append(wallet.SpendingRules, msg.SpendingRule)

	if err := k.SetWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &MsgAddSpendingRuleResponse{
		Address: msg.Address,
		Success: true,
	}, nil
}

// AddAutomationRule handles automation rule addition
func (k Keeper) AddAutomationRule(msg MsgAddAutomationRule) (*MsgAddAutomationRuleResponse, error) {
	wallet, found := k.GetWallet(msg.Address)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	// Check max rules
	params := k.GetParams()
	if len(wallet.AutomationRules) >= int(params.MaxRulesPerWallet) {
		return nil, types.ErrMaxRulesExceeded
	}

	wallet.AutomationRules = append(wallet.AutomationRules, msg.AutomationRule)

	if err := k.SetWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &MsgAddAutomationRuleResponse{
		Address: msg.Address,
		Success: true,
	}, nil
}

// Deposit handles deposit
func (k Keeper) Deposit(msg MsgDeposit) (*MsgDepositResponse, error) {
	wallet, found := k.GetWallet(msg.Address)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	currentBalance := parseUint64(wallet.Balance.Amount)
	depositAmount := parseUint64(msg.Amount)
	newBalance := currentBalance + depositAmount

	wallet.Balance.Amount = fmt.Sprintf("%d", newBalance)

	if err := k.SetWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &MsgDepositResponse{
		Address: msg.Address,
		Balance: wallet.Balance.Amount,
	}, nil
}

// Withdraw handles withdrawal
func (k Keeper) Withdraw(msg MsgWithdraw) (*MsgWithdrawResponse, error) {
	wallet, found := k.GetWallet(msg.Address)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	currentBalance := parseUint64(wallet.Balance.Amount)
	withdrawAmount := parseUint64(msg.Amount)

	if currentBalance < withdrawAmount {
		return nil, types.ErrInsufficientBalance
	}

	wallet.Balance.Amount = fmt.Sprintf("%d", currentBalance-withdrawAmount)

	if err := k.SetWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return &MsgWithdrawResponse{
		Address: msg.Address,
		Balance: wallet.Balance.Amount,
	}, nil
}

// Transfer handles transfer between wallets
func (k Keeper) Transfer(msg MsgTransfer) (*MsgTransferResponse, error) {
	// Get source wallet
	fromWallet, found := k.GetWallet(msg.FromAddress)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	// Check external transfers
	params := k.GetParams()
	if !params.AllowExternalTransfers {
		// Check if destination wallet exists
		if _, found := k.GetWallet(msg.ToAddress); !found {
			return nil, types.ErrOperationNotAllowed
		}
	}

	// Check balance
	transferAmount := parseUint64(msg.Amount)
	currentBalance := parseUint64(fromWallet.Balance.Amount)

	if currentBalance < transferAmount {
		return nil, types.ErrInsufficientBalance
	}

	// Check spending rules
	for _, rule := range fromWallet.SpendingRules {
		// Check if transfer is blocked
		for _, blocked := range rule.BlockedOps {
			if blocked == "external_transfer" {
				return nil, types.ErrOperationNotAllowed
			}
		}
		// Check max per tx
		maxPerTx := parseUint64(rule.MaxPerTx.Amount)
		if maxPerTx > 0 && transferAmount > maxPerTx {
			return nil, types.ErrSpendingRuleViolation
		}
	}

	// Debit from source
	fromWallet.Balance.Amount = fmt.Sprintf("%d", currentBalance-transferAmount)
	if err := k.SetWallet(fromWallet); err != nil {
		return nil, fmt.Errorf("failed to update source wallet: %w", err)
	}

	// Credit to destination (if exists)
	toWallet, found := k.GetWallet(msg.ToAddress)
	if found {
		toBalance := parseUint64(toWallet.Balance.Amount)
		toWallet.Balance.Amount = fmt.Sprintf("%d", toBalance+transferAmount)
		if err := k.SetWallet(toWallet); err != nil {
			return nil, fmt.Errorf("failed to update destination wallet: %w", err)
		}
	}

	log.Printf("[agentwallet] Transfer %s DPC from %s to %s", msg.Amount, msg.FromAddress, msg.ToAddress)

	return &MsgTransferResponse{
		FromAddress: msg.FromAddress,
		ToAddress:   msg.ToAddress,
		Amount:      msg.Amount,
	}, nil
}

// ProcessTransaction processes a transaction based on type
func (k Keeper) ProcessTransaction(txType string, txData json.RawMessage, blockHeight int64) (interface{}, error) {
	switch txType {
	case "create_wallet":
		var msg MsgCreateWallet
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse create_wallet: %w", err)
		}
		return k.CreateWallet(msg, blockHeight)

	case "add_spending_rule":
		var msg MsgAddSpendingRule
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse add_spending_rule: %w", err)
		}
		return k.AddSpendingRule(msg)

	case "add_automation_rule":
		var msg MsgAddAutomationRule
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse add_automation_rule: %w", err)
		}
		return k.AddAutomationRule(msg)

	case "deposit":
		var msg MsgDeposit
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse deposit: %w", err)
		}
		return k.Deposit(msg)

	case "withdraw":
		var msg MsgWithdraw
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse withdraw: %w", err)
		}
		return k.Withdraw(msg)

	case "transfer":
		var msg MsgTransfer
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse transfer: %w", err)
		}
		return k.Transfer(msg)

	default:
		return nil, fmt.Errorf("unknown transaction type: %s", txType)
	}
}
