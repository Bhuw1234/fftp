//go:build unit

package deparrow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// Test CreditTool
func TestCreditTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)

	if tool.Name() != "deparrow_credits" {
		t.Errorf("Name() = %s, want deparrow_credits", tool.Name())
	}
}

func TestCreditTool_Description(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "credit") {
		t.Error("Description should mention credits")
	}
}

func TestCreditTool_Parameters(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)

	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil")
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not found")
	}

	if _, ok := props["action"]; !ok {
		t.Error("action parameter not found")
	}
	if _, ok := props["amount"]; !ok {
		t.Error("amount parameter not found")
	}
	if _, ok := props["to_user"]; !ok {
		t.Error("to_user parameter not found")
	}
}

func TestCreditTool_Execute_Balance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 150.5,
			"last_active":    "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "balance",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "150.5") {
		t.Errorf("Result should contain balance: %s", result.ForLLM)
	}
}

func TestCreditTool_Execute_DefaultAction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 100.0,
			"last_active":    "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	tool := NewCreditTool(client)
	ctx := context.Background()

	// No action specified, should default to balance
	args := map[string]interface{}{}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestCreditTool_Execute_Check(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_sufficient": true,
			"required":       50.0,
			"available":      150.0,
			"difference":     100.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "check",
		"amount": 50.0,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "sufficient") {
		t.Errorf("Result should mention sufficient: %s", result.ForLLM)
	}
}

func TestCreditTool_Execute_Check_MissingAmount(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "check",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing amount")
	}
}

func TestCreditTool_Execute_Transfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "check") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"has_sufficient": true,
				"required":       25.0,
				"available":      100.0,
			})
		} else if strings.Contains(r.URL.Path, "transfer") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action":  "transfer",
		"to_user": "user-456",
		"amount":  25.0,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestCreditTool_Execute_Transfer_MissingToUser(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "transfer",
		"amount": 25.0,
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing to_user")
	}
}

func TestCreditTool_Execute_Transfer_MissingAmount(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action":  "transfer",
		"to_user": "user-456",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing amount")
	}
}

func TestCreditTool_Execute_Transfer_NegativeAmount(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action":  "transfer",
		"to_user": "user-456",
		"amount":  -10.0,
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for negative amount")
	}
}

func TestCreditTool_Execute_UnknownAction(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "unknown",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for unknown action")
	}
}

func TestCreditTool_Execute_Check_IntAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_sufficient": true,
			"required":       50.0,
			"available":      100.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "check",
		"amount": 50, // int instead of float64
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestCreditTool_Execute_Transfer_IntAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "check") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"has_sufficient": true,
			})
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action":  "transfer",
		"to_user": "user-456",
		"amount":  25, // int instead of float64
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

// Test CreditEarnTool
func TestCreditEarnTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditEarnTool(client)

	if tool.Name() != "deparrow_how_to_earn" {
		t.Errorf("Name() = %s, want deparrow_how_to_earn", tool.Name())
	}
}

func TestCreditEarnTool_Execute(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewCreditEarnTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "Earn") {
		t.Errorf("Result should mention earning: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "CPU") {
		t.Errorf("Result should mention CPU: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "GPU") {
		t.Errorf("Result should mention GPU: %s", result.ForLLM)
	}
}

// Test NetworkStatsTool
func TestNetworkStatsTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNetworkStatsTool(client)

	if tool.Name() != "deparrow_network" {
		t.Errorf("Name() = %s, want deparrow_network", tool.Name())
	}
}

func TestNetworkStatsTool_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"network": map[string]interface{}{
				"total_nodes":     100,
				"online_nodes":    85,
				"total_cpu_cores": 500,
				"total_gpu_count": 20,
				"total_memory_gb": 1000.0,
				"live_gflops":     5000.0,
				"live_tflops":     5.0,
			},
			"tiers": map[string]int{
				"bronze": 50,
				"silver": 30,
				"gold":   5,
			},
			"timestamp": "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNetworkStatsTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "100") {
		t.Errorf("Result should contain total nodes: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "85") {
		t.Errorf("Result should contain online nodes: %s", result.ForLLM)
	}
}

func TestNetworkStatsTool_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIError{
			Code:    500,
			Message: "Internal error",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNetworkStatsTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if !result.IsError {
		t.Error("Expected error from API")
	}
}

// Test LeaderboardTool
func TestLeaderboardTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewLeaderboardTool(client)

	if tool.Name() != "deparrow_leaderboard" {
		t.Errorf("Name() = %s, want deparrow_leaderboard", tool.Name())
	}
}

func TestLeaderboardTool_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard": []map[string]interface{}{
				{
					"rank":            1,
					"node_id":         "node-champion-abc123", // Longer ID to avoid slice bounds error
					"tier":            "legendary",
					"credits_earned":  10000.0,
					"cpu_usage_hours": 500.0,
					"gpu_usage_hours": 200.0,
				},
				{
					"rank":            2,
					"node_id":         "node-runner-up-xyz789", // Longer ID
					"tier":            "diamond",
					"credits_earned":  5000.0,
					"cpu_usage_hours": 300.0,
					"gpu_usage_hours": 100.0,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewLeaderboardTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"limit": 10,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	// Check that result contains leaderboard content
	if !strings.Contains(result.ForLLM, "Leaderboard") {
		t.Errorf("Result should contain Leaderboard: %s", result.ForLLM)
	}
}

func TestLeaderboardTool_Execute_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewLeaderboardTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "No entries") {
		t.Errorf("Result should indicate no entries: %s", result.ForLLM)
	}
}

func TestLeaderboardTool_Execute_LimitParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewLeaderboardTool(client)
	ctx := context.Background()

	// Test with float64 (JSON default)
	args := map[string]interface{}{
		"limit": float64(5),
	}

	result := tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with float64 limit returned error: %s", result.ForLLM)
	}

	// Test with int
	args = map[string]interface{}{
		"limit": 5,
	}

	result = tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with int limit returned error: %s", result.ForLLM)
	}
}

// Test tool interface compliance
func TestCreditTools_InterfaceCompliance(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")

	var _ tools.Tool = NewCreditTool(client)
	var _ tools.Tool = NewCreditEarnTool(client)
	var _ tools.Tool = NewNetworkStatsTool(client)
	var _ tools.Tool = NewLeaderboardTool(client)
}

// Test CreditTool with low balance
func TestCreditTool_Execute_Balance_Low(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 5.0, // Low balance
			"last_active":    "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "balance",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "low") {
		t.Errorf("Result should mention low balance: %s", result.ForLLM)
	}
}

// Test CreditTool with good balance
func TestCreditTool_Execute_Balance_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 150.0, // Good balance
			"last_active":    "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "balance",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "plenty") {
		t.Errorf("Result should mention plenty of credits: %s", result.ForLLM)
	}
}

// Test CreditTool with insufficient balance
func TestCreditTool_Execute_Check_Insufficient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_sufficient": false,
			"required":       100.0,
			"available":      50.0,
			"difference":     -50.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "check",
		"amount": 100.0,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "Insufficient") {
		t.Errorf("Result should mention insufficient: %s", result.ForLLM)
	}
}

// Test CreditTool with earned/spent balance
func TestCreditTool_Execute_Balance_WithHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 100.0,
			"last_active":    "2024-01-01T00:00:00Z",
			// Note: The API response structure determines what's shown
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	tool := NewCreditTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"action": "balance",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	// Check that balance is shown
	if !strings.Contains(result.ForLLM, "100") {
		t.Errorf("Result should contain balance: %s", result.ForLLM)
	}
}

// Test tier icons in leaderboard
func TestLeaderboardTool_Execute_TierIcons(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard": []map[string]interface{}{
				{"rank": 1, "node_id": "node-bronze-tier-01", "tier": "bronze", "credits_earned": 100.0, "cpu_usage_hours": 10.0, "gpu_usage_hours": 0.0},
				{"rank": 2, "node_id": "node-silver-tier-02", "tier": "silver", "credits_earned": 200.0, "cpu_usage_hours": 20.0, "gpu_usage_hours": 5.0},
				{"rank": 3, "node_id": "node-gold-tier-00003", "tier": "gold", "credits_earned": 500.0, "cpu_usage_hours": 50.0, "gpu_usage_hours": 10.0},
				{"rank": 4, "node_id": "node-diamond-tier-04", "tier": "diamond", "credits_earned": 1000.0, "cpu_usage_hours": 100.0, "gpu_usage_hours": 20.0},
				{"rank": 5, "node_id": "node-legendary-ti05", "tier": "legendary", "credits_earned": 5000.0, "cpu_usage_hours": 500.0, "gpu_usage_hours": 100.0},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewLeaderboardTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	// All tiers should be represented with icons
	if !strings.Contains(result.ForLLM, "Leaderboard") {
		t.Errorf("Result should contain Leaderboard: %s", result.ForLLM)
	}
}
