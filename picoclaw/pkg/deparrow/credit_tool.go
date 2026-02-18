package deparrow

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// CreditTool provides the ability to check credit balance and history.
// AI agents use credits to submit jobs and earn credits by contributing compute.
type CreditTool struct {
	client *Client
}

// NewCreditTool creates a new credit tool.
func NewCreditTool(client *Client) *CreditTool {
	return &CreditTool{client: client}
}

// Name returns the tool name.
func (t *CreditTool) Name() string {
	return "deparrow_credits"
}

// Description returns the tool description.
func (t *CreditTool) Description() string {
	return `Check your DEparrow credit balance and transaction history.

Credits are the currency of the DEparrow network:
- Earn credits by contributing compute resources
- Spend credits to run compute jobs
- Transfer credits to other users

Use this tool to:
- Check your current balance
- See if you have enough credits for a job
- View your earning/spending history
`
}

// Parameters returns the JSON schema for tool parameters.
func (t *CreditTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"balance", "check", "transfer"},
				"description": "Action to perform: 'balance' to check balance, 'check' to verify sufficient credits, 'transfer' to send credits",
				"default":     "balance",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"description": "Amount of credits (for 'check' or 'transfer' actions)",
			},
			"to_user": map[string]interface{}{
				"type":        "string",
				"description": "User ID to transfer credits to (for 'transfer' action)",
			},
		},
	}
}

// Execute runs the credit tool.
func (t *CreditTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	action, _ := args["action"].(string)
	if action == "" {
		action = "balance"
	}

	switch action {
	case "balance":
		return t.getBalance(ctx)
	case "check":
		return t.checkCredits(ctx, args)
	case "transfer":
		return t.transferCredits(ctx, args)
	default:
		return tools.ErrorResult(fmt.Sprintf("Unknown action: %s", action))
	}
}

// getBalance retrieves the current credit balance.
func (t *CreditTool) getBalance(ctx context.Context) *tools.ToolResult {
	balance, err := t.client.GetCredits(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get credit balance: %v", err))
	}

	var result strings.Builder
	result.WriteString("💰 DEparrow Credit Balance\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("  Current Balance: %.2f credits\n", balance.Balance))

	if balance.Earned > 0 {
		result.WriteString(fmt.Sprintf("  Total Earned:    %.2f credits\n", balance.Earned))
	}
	if balance.Spent > 0 {
		result.WriteString(fmt.Sprintf("  Total Spent:     %.2f credits\n", balance.Spent))
	}

	// Provide guidance on usage
	result.WriteString("\n💡 Tips:\n")
	if balance.Balance < 10 {
		result.WriteString("  • Balance is low. Contribute compute to earn more credits.\n")
	} else if balance.Balance >= 100 {
		result.WriteString("  • You have plenty of credits for compute jobs!\n")
	} else {
		result.WriteString("  • You have enough credits for several jobs.\n")
	}

	return tools.UserResult(result.String())
}

// checkCredits verifies if the user has sufficient credits.
func (t *CreditTool) checkCredits(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	amount, ok := args["amount"].(float64)
	if !ok {
		// Try int as well
		if amountInt, ok := args["amount"].(int); ok {
			amount = float64(amountInt)
		} else {
			return tools.ErrorResult("amount parameter is required for 'check' action")
		}
	}

	hasSufficient, err := t.client.CheckCredits(ctx, amount)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to check credits: %v", err))
	}

	var result strings.Builder
	if hasSufficient {
		result.WriteString(fmt.Sprintf("✅ You have sufficient credits for this operation (%.2f required).\n", amount))
	} else {
		result.WriteString(fmt.Sprintf("❌ Insufficient credits. %.2f required.\n", amount))
		result.WriteString("\nContribute compute resources to earn more credits.")
	}

	return tools.UserResult(result.String())
}

// transferCredits transfers credits to another user.
func (t *CreditTool) transferCredits(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	toUser, ok := args["to_user"].(string)
	if !ok || toUser == "" {
		return tools.ErrorResult("to_user parameter is required for 'transfer' action")
	}

	amount, ok := args["amount"].(float64)
	if !ok {
		if amountInt, ok := args["amount"].(int); ok {
			amount = float64(amountInt)
		} else {
			return tools.ErrorResult("amount parameter is required for 'transfer' action")
		}
	}

	if amount <= 0 {
		return tools.ErrorResult("amount must be positive")
	}

	err := t.client.TransferCredits(ctx, toUser, amount)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to transfer credits: %v", err))
	}

	return tools.UserResult(fmt.Sprintf(
		"✅ Successfully transferred %.2f credits to user %s",
		amount, toUser,
	))
}

// CreditEarnTool provides guidance on earning credits.
type CreditEarnTool struct {
	client *Client
}

// NewCreditEarnTool creates a new credit earning guide tool.
func NewCreditEarnTool(client *Client) *CreditEarnTool {
	return &CreditEarnTool{client: client}
}

// Name returns the tool name.
func (t *CreditEarnTool) Name() string {
	return "deparrow_how_to_earn"
}

// Description returns the tool description.
func (t *CreditEarnTool) Description() string {
	return "Learn how to earn DEparrow credits by contributing compute resources."
}

// Parameters returns the JSON schema for tool parameters.
func (t *CreditEarnTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute runs the credit earning guide tool.
func (t *CreditEarnTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	var result strings.Builder
	result.WriteString("💡 How to Earn DEparrow Credits\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString("Earn credits by contributing compute resources:\n\n")
	result.WriteString("🖥️  CPU Contribution:\n")
	result.WriteString("   • +10 credits/hour per CPU core\n")
	result.WriteString("   • Runs in background when idle\n\n")
	result.WriteString("🎮 GPU Contribution:\n")
	result.WriteString("   • +50 credits/hour per GPU\n")
	result.WriteString("   • Higher demand, more rewards\n\n")
	result.WriteString("📅 Long-term Rewards:\n")
	result.WriteString("   • +100 credits/day for 24/7 uptime\n")
	result.WriteString("   • +500 credits for referring new nodes\n\n")
	result.WriteString("🏆 Tier Bonuses:\n")
	result.WriteString("   • Bronze:    100 total hours\n")
	result.WriteString("   • Silver:    1,000 total hours\n")
	result.WriteString("   • Gold:      5,000 total hours\n")
	result.WriteString("   • Diamond:   10,000 total hours\n")
	result.WriteString("   • Legendary: 25,000 total hours\n\n")
	result.WriteString("To start earning, ensure your node is registered and online.\n")
	result.WriteString("Use 'deparrow_nodes' to check your contribution status.")

	return tools.UserResult(result.String())
}

// NetworkStatsTool provides network-wide statistics.
type NetworkStatsTool struct {
	client *Client
}

// NewNetworkStatsTool creates a new network stats tool.
func NewNetworkStatsTool(client *Client) *NetworkStatsTool {
	return &NetworkStatsTool{client: client}
}

// Name returns the tool name.
func (t *NetworkStatsTool) Name() string {
	return "deparrow_network"
}

// Description returns the tool description.
func (t *NetworkStatsTool) Description() string {
	return "View DEparrow network statistics including total nodes, compute power, and tier distribution."
}

// Parameters returns the JSON schema for tool parameters.
func (t *NetworkStatsTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute runs the network stats tool.
func (t *NetworkStatsTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	stats, err := t.client.GetNetworkStats(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get network stats: %v", err))
	}

	var result strings.Builder
	result.WriteString("🌐 DEparrow Network Statistics\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	result.WriteString(fmt.Sprintf("  Total Nodes:  %d\n", stats.TotalNodes))
	result.WriteString(fmt.Sprintf("  Online:       %d\n", stats.OnlineNodes))
	result.WriteString(fmt.Sprintf("  CPU Cores:    %d\n", stats.TotalCPU))
	result.WriteString(fmt.Sprintf("  GPUs:         %d\n", stats.TotalGPU))
	result.WriteString(fmt.Sprintf("  Memory:       %.1f GB\n", stats.TotalMemory))
	result.WriteString("\n⚡ Live Compute Power:\n")
	result.WriteString(fmt.Sprintf("  %.2f GFLOPS (%.2f TFLOPS)\n", stats.LiveGFlops, stats.LiveTFlops))
	result.WriteString("\n🏆 Tier Distribution:\n")
	for tier, count := range stats.TierDistribution {
		if count > 0 {
			result.WriteString(fmt.Sprintf("  %-10s: %d nodes\n", tier, count))
		}
	}

	return tools.UserResult(result.String())
}

// LeaderboardTool shows the top contributing nodes.
type LeaderboardTool struct {
	client *Client
}

// NewLeaderboardTool creates a new leaderboard tool.
func NewLeaderboardTool(client *Client) *LeaderboardTool {
	return &LeaderboardTool{client: client}
}

// Name returns the tool name.
func (t *LeaderboardTool) Name() string {
	return "deparrow_leaderboard"
}

// Description returns the tool description.
func (t *LeaderboardTool) Description() string {
	return "View the DEparrow contribution leaderboard showing top nodes."
}

// Parameters returns the JSON schema for tool parameters.
func (t *LeaderboardTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Number of entries to show (default: 10)",
				"default":     10,
			},
		},
	}
}

// Execute runs the leaderboard tool.
func (t *LeaderboardTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := args["limit"].(int); ok {
		limit = l
	}

	entries, err := t.client.GetLeaderboard(ctx, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get leaderboard: %v", err))
	}

	if len(entries) == 0 {
		return tools.UserResult("No entries on the leaderboard yet. Be the first to contribute!")
	}

	var result strings.Builder
	result.WriteString("🏆 DEparrow Contribution Leaderboard\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	tierIcons := map[ContributionTier]string{
		TierBronze:    "🥉",
		TierSilver:    "🥈",
		TierGold:      "🥇",
		TierDiamond:   "💎",
		TierLegendary: "🔥",
	}

	for _, entry := range entries {
		icon := tierIcons[entry.Tier]
		result.WriteString(fmt.Sprintf("%2d. %s %s\n", entry.Rank, icon, entry.NodeID[:12]+"..."))
		result.WriteString(fmt.Sprintf("    %.0f credits | %.1f CPUh + %.1f GPUh\n",
			entry.CreditsEarned, entry.CPUHours, entry.GPUHours))
	}

	result.WriteString("\nUse 'deparrow_nodes' to see your own contribution.")

	return tools.UserResult(result.String())
}

// Ensure tools implement the Tool interface
var _ tools.Tool = (*CreditTool)(nil)
var _ tools.Tool = (*CreditEarnTool)(nil)
var _ tools.Tool = (*NetworkStatsTool)(nil)
var _ tools.Tool = (*LeaderboardTool)(nil)
