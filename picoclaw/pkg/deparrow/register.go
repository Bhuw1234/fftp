package deparrow

import (
	"github.com/sipeed/picoclaw/pkg/tools"
)

// ToolsProvider creates and provides all DEparrow tools.
// Use this to easily register all DEparrow tools with an agent.
type ToolsProvider struct {
	client *Client
}

// NewToolsProvider creates a new tools provider with the given client.
func NewToolsProvider(client *Client) *ToolsProvider {
	return &ToolsProvider{client: client}
}

// NewToolsProviderFromConfig creates a tools provider from configuration.
func NewToolsProviderFromConfig(apiURL, jwtToken string, userID string) *ToolsProvider {
	client := NewClient(apiURL, jwtToken)
	client.SetUserID(userID)
	return NewToolsProvider(client)
}

// GetAllTools returns all DEparrow tools.
// This is the recommended way to get all tools for registration.
func (p *ToolsProvider) GetAllTools() []tools.Tool {
	return []tools.Tool{
		// Job management
		NewJobTool(p.client),
		NewJobStatusTool(p.client),
		NewJobListTool(p.client),
		NewJobCancelTool(p.client),

		// Credit management
		NewCreditTool(p.client),
		NewCreditEarnTool(p.client),
		NewNetworkStatsTool(p.client),
		NewLeaderboardTool(p.client),

		// Node management
		NewNodeTool(p.client),
		NewNodeContributionTool(p.client),
		NewOrchestratorTool(p.client),

		// Wallet management
		NewWalletTool(p.client),
		NewTransferTool(p.client),
		NewHealthTool(p.client),
	}
}

// GetJobTools returns tools for job management.
func (p *ToolsProvider) GetJobTools() []tools.Tool {
	return []tools.Tool{
		NewJobTool(p.client),
		NewJobStatusTool(p.client),
		NewJobListTool(p.client),
		NewJobCancelTool(p.client),
	}
}

// GetCreditTools returns tools for credit management.
func (p *ToolsProvider) GetCreditTools() []tools.Tool {
	return []tools.Tool{
		NewCreditTool(p.client),
		NewCreditEarnTool(p.client),
		NewNetworkStatsTool(p.client),
		NewLeaderboardTool(p.client),
	}
}

// GetNodeTools returns tools for node management.
func (p *ToolsProvider) GetNodeTools() []tools.Tool {
	return []tools.Tool{
		NewNodeTool(p.client),
		NewNodeContributionTool(p.client),
		NewOrchestratorTool(p.client),
	}
}

// GetWalletTools returns tools for wallet management.
func (p *ToolsProvider) GetWalletTools() []tools.Tool {
	return []tools.Tool{
		NewWalletTool(p.client),
		NewTransferTool(p.client),
		NewHealthTool(p.client),
	}
}

// RegisterAll registers all DEparrow tools with the provided registry.
func (p *ToolsProvider) RegisterAll(registry *tools.ToolRegistry) {
	for _, tool := range p.GetAllTools() {
		registry.Register(tool)
	}
}

// RegisterJobs registers job management tools.
func (p *ToolsProvider) RegisterJobs(registry *tools.ToolRegistry) {
	for _, tool := range p.GetJobTools() {
		registry.Register(tool)
	}
}

// RegisterCredits registers credit management tools.
func (p *ToolsProvider) RegisterCredits(registry *tools.ToolRegistry) {
	for _, tool := range p.GetCreditTools() {
		registry.Register(tool)
	}
}

// RegisterNodes registers node management tools.
func (p *ToolsProvider) RegisterNodes(registry *tools.ToolRegistry) {
	for _, tool := range p.GetNodeTools() {
		registry.Register(tool)
	}
}

// RegisterWallet registers wallet management tools.
func (p *ToolsProvider) RegisterWallet(registry *tools.ToolRegistry) {
	for _, tool := range p.GetWalletTools() {
		registry.Register(tool)
	}
}

// ToolNames returns the names of all available DEparrow tools.
func ToolNames() []string {
	return []string{
		// Job management
		"deparrow_submit_job",
		"deparrow_job_status",
		"deparrow_list_jobs",
		"deparrow_cancel_job",

		// Credit management
		"deparrow_credits",
		"deparrow_how_to_earn",
		"deparrow_network",
		"deparrow_leaderboard",

		// Node management
		"deparrow_nodes",
		"deparrow_contribution",
		"deparrow_orchestrators",

		// Wallet management
		"deparrow_wallet",
		"deparrow_transfer",
		"deparrow_health",
	}
}

// ToolDescriptions returns a map of tool names to descriptions.
func ToolDescriptions() map[string]string {
	return map[string]string{
		// Job management
		"deparrow_submit_job":   "Submit a compute job to the DEparrow network",
		"deparrow_job_status":   "Check the status of a submitted DEparrow job",
		"deparrow_list_jobs":    "List all jobs submitted by the authenticated user",
		"deparrow_cancel_job":   "Cancel a running job and receive partial credit refund",

		// Credit management
		"deparrow_credits":      "Check your DEparrow credit balance and transaction history",
		"deparrow_how_to_earn":  "Learn how to earn DEparrow credits by contributing compute resources",
		"deparrow_network":      "View DEparrow network statistics including total nodes and compute power",
		"deparrow_leaderboard":  "View the DEparrow contribution leaderboard showing top nodes",

		// Node management
		"deparrow_nodes":         "List and inspect compute nodes on the DEparrow network",
		"deparrow_contribution":  "View detailed contribution statistics for a specific node",
		"deparrow_orchestrators": "List orchestrator nodes in the DEparrow network",

		// Wallet management
		"deparrow_wallet":  "View your DEparrow wallet balance and transaction history",
		"deparrow_transfer": "Transfer credits to another DEparrow user",
		"deparrow_health":   "Check the health of your DEparrow connection and the network",
	}
}
