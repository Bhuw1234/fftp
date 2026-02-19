//go:build unit

package deparrow

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// Test ToolsProvider
func TestNewToolsProvider(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	if provider == nil {
		t.Fatal("NewToolsProvider returned nil")
	}
	if provider.client == nil {
		t.Error("Provider client is nil")
	}
}

func TestNewToolsProviderFromConfig(t *testing.T) {
	provider := NewToolsProviderFromConfig("http://localhost:8080", "test-token", "user-123")

	if provider == nil {
		t.Fatal("NewToolsProviderFromConfig returned nil")
	}
	if provider.client == nil {
		t.Error("Provider client is nil")
	}
	if provider.client.userID != "user-123" {
		t.Errorf("userID = %s, want user-123", provider.client.userID)
	}
}

func TestToolsProvider_GetAllTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	tools := provider.GetAllTools()

	// Should have 14 tools
	if len(tools) != 14 {
		t.Errorf("GetAllTools() returned %d tools, want 14", len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name()] = true
	}

	expectedTools := []string{
		"deparrow_submit_job",
		"deparrow_job_status",
		"deparrow_list_jobs",
		"deparrow_cancel_job",
		"deparrow_credits",
		"deparrow_how_to_earn",
		"deparrow_network",
		"deparrow_leaderboard",
		"deparrow_nodes",
		"deparrow_contribution",
		"deparrow_orchestrators",
		"deparrow_wallet",
		"deparrow_transfer",
		"deparrow_health",
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("Missing tool: %s", name)
		}
	}
}

func TestToolsProvider_GetJobTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	tools := provider.GetJobTools()

	if len(tools) != 4 {
		t.Errorf("GetJobTools() returned %d tools, want 4", len(tools))
	}

	expectedNames := []string{
		"deparrow_submit_job",
		"deparrow_job_status",
		"deparrow_list_jobs",
		"deparrow_cancel_job",
	}

	for i, tool := range tools {
		if tool.Name() != expectedNames[i] {
			t.Errorf("Tool %d: name = %s, want %s", i, tool.Name(), expectedNames[i])
		}
	}
}

func TestToolsProvider_GetCreditTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	tools := provider.GetCreditTools()

	if len(tools) != 4 {
		t.Errorf("GetCreditTools() returned %d tools, want 4", len(tools))
	}

	expectedNames := []string{
		"deparrow_credits",
		"deparrow_how_to_earn",
		"deparrow_network",
		"deparrow_leaderboard",
	}

	for i, tool := range tools {
		if tool.Name() != expectedNames[i] {
			t.Errorf("Tool %d: name = %s, want %s", i, tool.Name(), expectedNames[i])
		}
	}
}

func TestToolsProvider_GetNodeTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	tools := provider.GetNodeTools()

	if len(tools) != 3 {
		t.Errorf("GetNodeTools() returned %d tools, want 3", len(tools))
	}

	expectedNames := []string{
		"deparrow_nodes",
		"deparrow_contribution",
		"deparrow_orchestrators",
	}

	for i, tool := range tools {
		if tool.Name() != expectedNames[i] {
			t.Errorf("Tool %d: name = %s, want %s", i, tool.Name(), expectedNames[i])
		}
	}
}

func TestToolsProvider_GetWalletTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	tools := provider.GetWalletTools()

	if len(tools) != 3 {
		t.Errorf("GetWalletTools() returned %d tools, want 3", len(tools))
	}

	expectedNames := []string{
		"deparrow_wallet",
		"deparrow_transfer",
		"deparrow_health",
	}

	for i, tool := range tools {
		if tool.Name() != expectedNames[i] {
			t.Errorf("Tool %d: name = %s, want %s", i, tool.Name(), expectedNames[i])
		}
	}
}

func TestToolsProvider_RegisterAll(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	provider.RegisterAll(registry)

	// Verify all 14 tools are registered
	if registry.Count() != 14 {
		t.Errorf("Registry count = %d, want 14", registry.Count())
	}

	// Verify each tool is accessible
	expectedTools := []string{
		"deparrow_submit_job",
		"deparrow_job_status",
		"deparrow_list_jobs",
		"deparrow_cancel_job",
		"deparrow_credits",
		"deparrow_how_to_earn",
		"deparrow_network",
		"deparrow_leaderboard",
		"deparrow_nodes",
		"deparrow_contribution",
		"deparrow_orchestrators",
		"deparrow_wallet",
		"deparrow_transfer",
		"deparrow_health",
	}

	for _, name := range expectedTools {
		if _, ok := registry.Get(name); !ok {
			t.Errorf("Tool %s not found in registry", name)
		}
	}
}

func TestToolsProvider_RegisterJobs(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	provider.RegisterJobs(registry)

	if registry.Count() != 4 {
		t.Errorf("Registry count = %d, want 4", registry.Count())
	}
}

func TestToolsProvider_RegisterCredits(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	provider.RegisterCredits(registry)

	if registry.Count() != 4 {
		t.Errorf("Registry count = %d, want 4", registry.Count())
	}
}

func TestToolsProvider_RegisterNodes(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	provider.RegisterNodes(registry)

	if registry.Count() != 3 {
		t.Errorf("Registry count = %d, want 3", registry.Count())
	}
}

func TestToolsProvider_RegisterWallet(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	provider.RegisterWallet(registry)

	if registry.Count() != 3 {
		t.Errorf("Registry count = %d, want 3", registry.Count())
	}
}

// Test ToolNames function
func TestToolNames(t *testing.T) {
	names := ToolNames()

	if len(names) != 14 {
		t.Errorf("ToolNames() returned %d names, want 14", len(names))
	}

	// Verify all expected names are present
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	expectedNames := []string{
		"deparrow_submit_job",
		"deparrow_job_status",
		"deparrow_list_jobs",
		"deparrow_cancel_job",
		"deparrow_credits",
		"deparrow_how_to_earn",
		"deparrow_network",
		"deparrow_leaderboard",
		"deparrow_nodes",
		"deparrow_contribution",
		"deparrow_orchestrators",
		"deparrow_wallet",
		"deparrow_transfer",
		"deparrow_health",
	}

	for _, name := range expectedNames {
		if !nameSet[name] {
			t.Errorf("Missing tool name: %s", name)
		}
	}
}

// Test ToolDescriptions function
func TestToolDescriptions(t *testing.T) {
	descs := ToolDescriptions()

	if len(descs) != 14 {
		t.Errorf("ToolDescriptions() returned %d descriptions, want 14", len(descs))
	}

	// Verify each description is non-empty
	for name, desc := range descs {
		if desc == "" {
			t.Errorf("Empty description for tool: %s", name)
		}
	}

	// Verify some specific descriptions
	if !containsStr(descs["deparrow_submit_job"], "Submit") {
		t.Error("deparrow_submit_job description should mention Submit")
	}
	if !containsStr(descs["deparrow_credits"], "credit") {
		t.Error("deparrow_credits description should mention credit")
	}
	if !containsStr(descs["deparrow_wallet"], "wallet") {
		t.Error("deparrow_wallet description should mention wallet")
	}
}

// Test tool categories
func TestToolCategories(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)

	// Job tools
	jobTools := provider.GetJobTools()
	for _, tool := range jobTools {
		name := tool.Name()
		if !containsStr(name, "job") {
			t.Errorf("Job tool %s should contain 'job' in name", name)
		}
	}

	// Credit tools
	creditTools := provider.GetCreditTools()
	for _, tool := range creditTools {
		name := tool.Name()
		valid := containsStr(name, "credit") || containsStr(name, "earn") || 
		         containsStr(name, "network") || containsStr(name, "leaderboard")
		if !valid {
			t.Errorf("Credit tool %s has unexpected name", name)
		}
	}

	// Node tools
	nodeTools := provider.GetNodeTools()
	for _, tool := range nodeTools {
		name := tool.Name()
		valid := containsStr(name, "node") || containsStr(name, "contribution") || 
		         containsStr(name, "orchestrator")
		if !valid {
			t.Errorf("Node tool %s has unexpected name", name)
		}
	}

	// Wallet tools
	walletTools := provider.GetWalletTools()
	for _, tool := range walletTools {
		name := tool.Name()
		valid := containsStr(name, "wallet") || containsStr(name, "transfer") || 
		         containsStr(name, "health")
		if !valid {
			t.Errorf("Wallet tool %s has unexpected name", name)
		}
	}
}

// Test that tools implement the Tool interface
func TestTools_InterfaceCompliance(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")

	// Test all tools implement tools.Tool
	var _ tools.Tool = NewJobTool(client)
	var _ tools.Tool = NewJobStatusTool(client)
	var _ tools.Tool = NewJobListTool(client)
	var _ tools.Tool = NewJobCancelTool(client)
	var _ tools.Tool = NewCreditTool(client)
	var _ tools.Tool = NewCreditEarnTool(client)
	var _ tools.Tool = NewNetworkStatsTool(client)
	var _ tools.Tool = NewLeaderboardTool(client)
	var _ tools.Tool = NewNodeTool(client)
	var _ tools.Tool = NewNodeContributionTool(client)
	var _ tools.Tool = NewOrchestratorTool(client)
	var _ tools.Tool = NewWalletTool(client)
	var _ tools.Tool = NewTransferTool(client)
	var _ tools.Tool = NewHealthTool(client)
}

// Test tool registration is idempotent
func TestToolsProvider_RegisterAll_Idempotent(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	registry := tools.NewToolRegistry()

	// Register twice
	provider.RegisterAll(registry)
	initialCount := registry.Count()

	provider.RegisterAll(registry)
	finalCount := registry.Count()

	// Count should be the same (tools are overwritten, not duplicated)
	if finalCount != initialCount {
		t.Errorf("RegisterAll should be idempotent: initial=%d, final=%d", initialCount, finalCount)
	}
}

// Test tool descriptions match actual tools
func TestToolDescriptions_MatchTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	descs := ToolDescriptions()

	allTools := provider.GetAllTools()
	for _, tool := range allTools {
		desc, ok := descs[tool.Name()]
		if !ok {
			t.Errorf("Tool %s not in ToolDescriptions", tool.Name())
			continue
		}
		// The description from ToolDescriptions should be a subset/summary
		// of the full description from the tool
		if desc == "" {
			t.Errorf("Empty description for tool %s", tool.Name())
		}
	}
}

// Test tool names match tools
func TestToolNames_MatchTools(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	provider := NewToolsProvider(client)
	names := ToolNames()

	allTools := provider.GetAllTools()
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	for _, tool := range allTools {
		if !nameSet[tool.Name()] {
			t.Errorf("Tool %s not in ToolNames", tool.Name())
		}
	}
}

// Test provider with different configurations
func TestToolsProvider_DifferentConfigs(t *testing.T) {
	tests := []struct {
		name    string
		apiURL  string
		token   string
		userID  string
	}{
		{
			name:   "localhost",
			apiURL: "http://localhost:8080",
			token:  "local-token",
			userID: "user-local",
		},
		{
			name:   "production",
			apiURL: "https://api.deparrow.io",
			token:  "prod-token",
			userID: "user-prod",
		},
		{
			name:   "staging",
			apiURL: "https://staging.deparrow.io",
			token:  "stage-token",
			userID: "user-stage",
		},
		{
			name:   "empty_token",
			apiURL: "http://localhost:8080",
			token:  "",
			userID: "user-notoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewToolsProviderFromConfig(tt.apiURL, tt.token, tt.userID)

			if provider == nil {
				t.Fatal("Provider is nil")
			}

			tools := provider.GetAllTools()
			if len(tools) != 14 {
				t.Errorf("GetAllTools returned %d tools, want 14", len(tools))
			}
		})
	}
}

// Helper function
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
