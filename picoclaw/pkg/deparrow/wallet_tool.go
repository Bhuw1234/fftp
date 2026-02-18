package deparrow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// WalletTool provides wallet operations for the DEparrow network.
// AI agents use wallets to manage their credits and track transactions.
type WalletTool struct {
	client *Client
}

// NewWalletTool creates a new wallet tool.
func NewWalletTool(client *Client) *WalletTool {
	return &WalletTool{client: client}
}

// Name returns the tool name.
func (t *WalletTool) Name() string {
	return "deparrow_wallet"
}

// Description returns the tool description.
func (t *WalletTool) Description() string {
	return `View your DEparrow wallet balance and transaction history.

Your wallet is your identity on the DEparrow network:
- Stores your credit balance
- Tracks all credit transactions
- Manages your identity for job submission

AI agents use wallets to:
- Pay for compute jobs
- Receive credits for contributions
- Transfer credits to other agents
`
}

// Parameters returns the JSON schema for tool parameters.
func (t *WalletTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"balance", "history", "info"},
				"description": "Action to perform: 'balance' to check balance, 'history' for transactions, 'info' for wallet details",
				"default":     "balance",
			},
		},
	}
}

// Execute runs the wallet tool.
func (t *WalletTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	action, _ := args["action"].(string)
	if action == "" {
		action = "balance"
	}

	switch action {
	case "balance":
		return t.getBalance(ctx)
	case "history":
		return t.getHistory(ctx)
	case "info":
		return t.getInfo(ctx)
	default:
		return tools.ErrorResult(fmt.Sprintf("Unknown action: %s", action))
	}
}

// getBalance retrieves and displays wallet balance.
func (t *WalletTool) getBalance(ctx context.Context) *tools.ToolResult {
	wallet, err := t.client.GetWallet(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get wallet: %v", err))
	}

	var result strings.Builder
	result.WriteString("👛 DEparrow Wallet\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("  Address:  %s\n", wallet.Address[:16]+"...\n"))
	result.WriteString(fmt.Sprintf("  Balance:  %.2f credits 💰\n", wallet.Balance))
	result.WriteString(fmt.Sprintf("  Created:  %s\n", wallet.CreatedAt.Format("2006-01-02")))

	// Calculate spending power
	result.WriteString("\n💡 Spending Power:\n")
	avgJobCost := 5.0 // Average job cost
	jobsCanRun := int(wallet.Balance / avgJobCost)
	result.WriteString(fmt.Sprintf("  Can run ~%d standard jobs (%.0f credits each)\n", jobsCanRun, avgJobCost))

	return tools.UserResult(result.String())
}

// getHistory displays transaction history.
func (t *WalletTool) getHistory(ctx context.Context) *tools.ToolResult {
	wallet, err := t.client.GetWallet(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get wallet: %v", err))
	}

	var result strings.Builder
	result.WriteString("📜 Transaction History\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if len(wallet.Transactions) == 0 {
		result.WriteString("No transactions yet.\n\n")
		result.WriteString("💡 Transactions will appear here when you:\n")
		result.WriteString("  • Submit compute jobs (spend)\n")
		result.WriteString("  • Contribute compute resources (earn)\n")
		result.WriteString("  • Receive transfers from other users\n")
		return tools.UserResult(result.String())
	}

	// Display last 10 transactions
	displayCount := len(wallet.Transactions)
	if displayCount > 10 {
		displayCount = 10
	}

	runningBalance := wallet.Balance
	for i := len(wallet.Transactions) - 1; i >= 0 && i >= len(wallet.Transactions)-displayCount; i-- {
		tx := wallet.Transactions[i]

		var icon string
		var amountStr string
		switch tx.Type {
		case "earn":
			icon = "📈"
			amountStr = fmt.Sprintf("+%.2f", tx.Amount)
			runningBalance -= tx.Amount // Go back in time
		case "spend":
			icon = "📉"
			amountStr = fmt.Sprintf("-%.2f", tx.Amount)
			runningBalance += tx.Amount
		case "transfer":
			if tx.FromUser != "" {
				icon = "📤"
				amountStr = fmt.Sprintf("-%.2f", tx.Amount)
				runningBalance += tx.Amount
			} else {
				icon = "📥"
				amountStr = fmt.Sprintf("+%.2f", tx.Amount)
				runningBalance -= tx.Amount
			}
		default:
			icon = "💳"
			amountStr = fmt.Sprintf("%.2f", tx.Amount)
		}

		result.WriteString(fmt.Sprintf("%s %s %s credits\n",
			tx.Timestamp.Format("2006-01-02 15:04"),
			icon,
			amountStr,
		))
		result.WriteString(fmt.Sprintf("   %s\n\n", tx.Description))
	}

	result.WriteString(fmt.Sprintf("\nCurrent Balance: %.2f credits\n", wallet.Balance))

	return tools.UserResult(result.String())
}

// getInfo displays wallet information.
func (t *WalletTool) getInfo(ctx context.Context) *tools.ToolResult {
	wallet, err := t.client.GetWallet(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get wallet: %v", err))
	}

	var result strings.Builder
	result.WriteString("📋 Wallet Information\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("Wallet Address:\n  %s\n\n", wallet.Address))
	result.WriteString(fmt.Sprintf("Current Balance:\n  %.2f credits\n\n", wallet.Balance))
	result.WriteString(fmt.Sprintf("Created:\n  %s\n\n", wallet.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Total Transactions:\n  %d\n\n", len(wallet.Transactions)))

	result.WriteString("💡 Usage Tips:\n")
	result.WriteString("  • Your wallet address is your identity on DEparrow\n")
	result.WriteString("  • Keep your wallet funded by contributing compute\n")
	result.WriteString("  • Use 'deparrow_credits' for detailed balance info\n")
	result.WriteString("  • Use 'deparrow_how_to_earn' to learn about earning")

	return tools.UserResult(result.String())
}

// TransferTool provides credit transfer functionality.
type TransferTool struct {
	client *Client
}

// NewTransferTool creates a new transfer tool.
func NewTransferTool(client *Client) *TransferTool {
	return &TransferTool{client: client}
}

// Name returns the tool name.
func (t *TransferTool) Name() string {
	return "deparrow_transfer"
}

// Description returns the tool description.
func (t *TransferTool) Description() string {
	return "Transfer credits to another DEparrow user."
}

// Parameters returns the JSON schema for tool parameters.
func (t *TransferTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"to_user_id": map[string]interface{}{
				"type":        "string",
				"description": "The user ID to transfer credits to",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"description": "Amount of credits to transfer",
				"minimum":     0.01,
			},
			"memo": map[string]interface{}{
				"type":        "string",
				"description": "Optional memo for the transfer",
			},
		},
		"required": []string{"to_user_id", "amount"},
	}
}

// Execute runs the transfer tool.
func (t *TransferTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	toUserID, ok := args["to_user_id"].(string)
	if !ok || toUserID == "" {
		return tools.ErrorResult("to_user_id is required")
	}

	amount, ok := args["amount"].(float64)
	if !ok {
		// Try int as well
		if amountInt, ok := args["amount"].(int); ok {
			amount = float64(amountInt)
		} else {
			return tools.ErrorResult("amount is required and must be a number")
		}
	}

	if amount <= 0 {
		return tools.ErrorResult("amount must be positive")
	}

	// Check if we have enough credits first
	hasSufficient, err := t.client.CheckCredits(ctx, amount)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to check balance: %v", err))
	}

	if !hasSufficient {
		return tools.ErrorResult(fmt.Sprintf("Insufficient balance for transfer of %.2f credits", amount))
	}

	// Perform transfer
	err = t.client.TransferCredits(ctx, toUserID, amount)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Transfer failed: %v", err))
	}

	var result strings.Builder
	result.WriteString("✅ Transfer Successful\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("  To:      %s\n", toUserID[:16]+"..."))
	result.WriteString(fmt.Sprintf("  Amount:  %.2f credits\n", amount))

	if memo, ok := args["memo"].(string); ok && memo != "" {
		result.WriteString(fmt.Sprintf("  Memo:    %s\n", memo))
	}

	result.WriteString(fmt.Sprintf("\n  Time:    %s\n", time.Now().Format("2006-01-02 15:04:05")))

	return tools.UserResult(result.String())
}

// HealthTool provides health check for the DEparrow connection.
type HealthTool struct {
	client *Client
}

// NewHealthTool creates a new health tool.
func NewHealthTool(client *Client) *HealthTool {
	return &HealthTool{client: client}
}

// Name returns the tool name.
func (t *HealthTool) Name() string {
	return "deparrow_health"
}

// Description returns the tool description.
func (t *HealthTool) Description() string {
	return "Check the health of your DEparrow connection and the network."
}

// Parameters returns the JSON schema for tool parameters.
func (t *HealthTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute runs the health tool.
func (t *HealthTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	health, err := t.client.Health(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("DEparrow health check failed: %v", err))
	}

	var result strings.Builder
	result.WriteString("🏥 DEparrow Health Check\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Parse health response
	if status, ok := health["status"].(string); ok {
		if status == "healthy" {
			result.WriteString("  Status:    ✅ Healthy\n")
		} else {
			result.WriteString(fmt.Sprintf("  Status:    ⚠️  %s\n", status))
		}
	}

	if version, ok := health["version"].(string); ok {
		result.WriteString(fmt.Sprintf("  Version:   %s\n", version))
	}

	if timestamp, ok := health["timestamp"].(string); ok {
		result.WriteString(fmt.Sprintf("  Time:      %s\n", timestamp))
	}

	// Show component counts
	if components, ok := health["components"].(map[string]interface{}); ok {
		result.WriteString("\n📊 Components:\n")
		for name, count := range components {
			result.WriteString(fmt.Sprintf("  %-15s: %v\n", name, count))
		}
	}

	result.WriteString("\n✅ Your connection to DEparrow is working!")

	return tools.UserResult(result.String())
}

// Ensure tools implement the Tool interface
var _ tools.Tool = (*WalletTool)(nil)
var _ tools.Tool = (*TransferTool)(nil)
var _ tools.Tool = (*HealthTool)(nil)
