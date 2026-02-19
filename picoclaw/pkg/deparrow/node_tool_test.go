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

// Test NodeTool
func TestNodeTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNodeTool(client)

	if tool.Name() != "deparrow_nodes" {
		t.Errorf("Name() = %s, want deparrow_nodes", tool.Name())
	}
}

func TestNodeTool_Description(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNodeTool(client)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "node") {
		t.Error("Description should mention nodes")
	}
}

func TestNodeTool_Parameters(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNodeTool(client)

	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil")
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not found")
	}

	if _, ok := props["node_id"]; !ok {
		t.Error("node_id parameter not found")
	}
	if _, ok := props["status"]; !ok {
		t.Error("status parameter not found")
	}
	if _, ok := props["contribution"]; !ok {
		t.Error("contribution parameter not found")
	}
}

func TestNodeTool_Execute_ListNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes" {
			t.Errorf("Path = %s, want /api/v1/nodes", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"node_id":        "node-01-abc123-xyz", // Longer ID to avoid slice bounds error
					"arch":           "x86_64",
					"status":         "online",
					"last_seen":      "2024-01-01T00:00:00Z",
					"credits_earned": 100.0,
					"resources": map[string]interface{}{
						"cpu_cores": 8,
						"memory":    "16Gi",
						"gpu_count": 1,
					},
				},
				{
					"node_id":        "node-02-def456-uvw",
					"arch":           "arm64",
					"status":         "online",
					"last_seen":      "2024-01-01T00:00:00Z",
					"credits_earned": 200.0,
				},
			},
			"total":  2,
			"online": 2,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	// Check that result contains node information
	if !strings.Contains(result.ForLLM, "Compute Nodes") {
		t.Errorf("Result should mention Compute Nodes: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_FilterByStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"node_id":        "node-online-test01",
					"arch":           "x86_64",
					"status":         "online",
					"last_seen":      "2024-01-01T00:00:00Z",
					"credits_earned": 100.0,
				},
				{
					"node_id":        "node-offline-test02",
					"arch":           "x86_64",
					"status":         "offline",
					"last_seen":      "2024-01-01T00:00:00Z",
					"credits_earned": 50.0,
				},
			},
			"total":  2,
			"online": 1,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	// Filter by online status
	args := map[string]interface{}{
		"status": "online",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_GetSpecificNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "contribution") {
			// Return contribution data
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "node-specific-abc123",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 100.0,
					"gpu_usage_hours": 50.0,
					"live_gflops":     500.0,
					"network_percent": 2.5,
				},
				"ranking": map[string]interface{}{
					"rank":        10,
					"total_nodes": 100,
					"tier":        "gold",
				},
			})
		} else {
			// Return node data
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id":        "node-specific-abc123",
				"arch":           "x86_64",
				"status":         "online",
				"last_seen":      "2024-01-01T00:00:00Z",
				"credits_earned": 500.0,
				"tier":           "gold",
				"resources": map[string]interface{}{
					"cpu_cores": 16,
					"memory":    "64Gi",
					"gpu_count": 2,
					"gpu_model": "NVIDIA A100",
				},
				"labels": map[string]string{
					"region": "us-west-2",
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "node-specific-abc123",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "node-specific") {
		t.Errorf("Result should contain node ID: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_NodeNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{
			Code:    404,
			Message: "Node not found",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "nonexistent",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for node not found")
	}
}

func TestNodeTool_Execute_WithContribution(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "contribution") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "node-with-contrib-01",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 100.0,
					"gpu_usage_hours": 50.0,
					"live_gflops":     500.0,
					"network_percent": 1.5,
				},
				"ranking": map[string]interface{}{
					"rank":        10,
					"total_nodes": 100,
					"tier":        "silver",
				},
			})
		} else if r.URL.Path == "/api/v1/nodes" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"node_id":        "node-with-contrib-01",
						"arch":           "x86_64",
						"status":         "online",
						"last_seen":      "2024-01-01T00:00:00Z",
						"credits_earned": 500.0,
					},
				},
				"total":  1,
				"online": 1,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id":        "node-with-contrib-01",
				"arch":           "x86_64",
				"status":         "online",
				"last_seen":      "2024-01-01T00:00:00Z",
				"credits_earned": 500.0,
				"tier":           "silver",
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"contribution": true,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes":  []interface{}{},
			"total":  0,
			"online": 0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"status": "online",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "No nodes found") {
		t.Errorf("Result should indicate no nodes: %s", result.ForLLM)
	}
}

// Test NodeContributionTool
func TestNodeContributionTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNodeContributionTool(client)

	if tool.Name() != "deparrow_contribution" {
		t.Errorf("Name() = %s, want deparrow_contribution", tool.Name())
	}
}

func TestNodeContributionTool_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "contribution") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "node-contrib-test-01",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 200.0,
					"gpu_usage_hours": 100.0,
					"live_gflops":     1000.0,
					"network_percent": 2.5,
				},
				"ranking": map[string]interface{}{
					"rank":        5,
					"total_nodes": 200,
					"tier":        "gold",
				},
			})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id":        "node-contrib-test-01",
				"arch":           "x86_64",
				"status":         "online",
				"credits_earned": 1000.0,
				"tier":           "gold",
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeContributionTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "node-contrib-test-01",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "200") {
		t.Errorf("Result should contain CPU hours: %s", result.ForLLM)
	}
}

func TestNodeContributionTool_Execute_MissingNodeID(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewNodeContributionTool(client)
	ctx := context.Background()

	args := map[string]interface{}{}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing node_id")
	}
}

func TestNodeContributionTool_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{
			Code:    404,
			Message: "Node not found",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeContributionTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "nonexistent",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error from API")
	}
}

// Test OrchestratorTool
func TestOrchestratorTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewOrchestratorTool(client)

	if tool.Name() != "deparrow_orchestrators" {
		t.Errorf("Name() = %s, want deparrow_orchestrators", tool.Name())
	}
}

func TestOrchestratorTool_Execute(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewOrchestratorTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "Orchestrator") {
		t.Errorf("Result should mention orchestrators: %s", result.ForLLM)
	}
}

// Test getNextTier helper
func TestGetNextTier(t *testing.T) {
	tests := []struct {
		currentTier  ContributionTier
		totalHours   float64
		wantNext     ContributionTier
		wantHours    float64
	}{
		{TierBronze, 50, TierSilver, 50},
		{TierBronze, 100, TierGold, 900},
		{TierSilver, 500, TierGold, 500},
		{TierSilver, 1000, TierDiamond, 4000},
		{TierGold, 3000, TierDiamond, 2000},
		{TierGold, 5000, TierLegendary, 5000},
		{TierDiamond, 8000, TierLegendary, 2000},
		{TierDiamond, 10000, "", 0}, // Max tier
		{TierLegendary, 20000, "", 0}, // Already max
	}

	for _, tt := range tests {
		t.Run(string(tt.currentTier), func(t *testing.T) {
			nextTier, hoursNeeded := getNextTier(tt.currentTier, tt.totalHours)
			if nextTier != tt.wantNext {
				t.Errorf("nextTier = %s, want %s", nextTier, tt.wantNext)
			}
			if hoursNeeded != tt.wantHours {
				t.Errorf("hoursNeeded = %f, want %f", hoursNeeded, tt.wantHours)
			}
		})
	}
}

// Test tool interface compliance
func TestNodeTools_InterfaceCompliance(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")

	var _ tools.Tool = NewNodeTool(client)
	var _ tools.Tool = NewNodeContributionTool(client)
	var _ tools.Tool = NewOrchestratorTool(client)
}

// Test status filtering
func TestNodeTool_Execute_StatusFilter_All(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"node_id": "node-online-test001", "arch": "x86_64", "status": "online", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 100.0},
				{"node_id": "node-offline-test02", "arch": "x86_64", "status": "offline", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 50.0},
				{"node_id": "node-maint-test0003", "arch": "x86_64", "status": "maintenance", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 25.0},
			},
			"total":  3,
			"online": 1,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"status": "all",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_StatusFilter_Offline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"node_id": "node-online-test001", "arch": "x86_64", "status": "online", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 100.0},
				{"node_id": "node-offline-test02", "arch": "x86_64", "status": "offline", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 50.0},
			},
			"total":  2,
			"online": 1,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"status": "offline",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestNodeTool_Execute_StatusFilter_Maintenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"node_id": "node-online-test001", "arch": "x86_64", "status": "online", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 100.0},
				{"node_id": "node-maint-test0003", "arch": "x86_64", "status": "maintenance", "last_seen": "2024-01-01T00:00:00Z", "credits_earned": 50.0},
			},
			"total":  2,
			"online": 1,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"status": "maintenance",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

// Test node with GPU resources
func TestNodeTool_Execute_WithGPU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "contribution") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "gpu-node-h100-abc123",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 500.0,
					"gpu_usage_hours": 200.0,
					"live_gflops":     2000.0,
					"network_percent": 5.0,
				},
				"ranking": map[string]interface{}{
					"rank":        5,
					"total_nodes": 100,
					"tier":        "diamond",
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id":        "gpu-node-h100-abc123",
				"arch":           "x86_64",
				"status":         "online",
				"last_seen":      "2024-01-01T00:00:00Z",
				"credits_earned": 2000.0,
				"tier":           "diamond",
				"resources": map[string]interface{}{
					"cpu_cores": 32,
					"memory":    "128Gi",
					"gpu_count": 8,
					"gpu_model": "NVIDIA H100",
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "gpu-node-h100-abc123",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "GPU") {
		t.Errorf("Result should mention GPU: %s", result.ForLLM)
	}
}

// Test contribution progress
func TestNodeContributionTool_Execute_Progress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "contribution") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "node-progress-test01",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 500.0,
					"gpu_usage_hours": 200.0,
					"live_gflops":     800.0,
					"network_percent": 3.0,
				},
				"ranking": map[string]interface{}{
					"rank":        20,
					"total_nodes": 500,
					"tier":        "silver",
				},
			})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id":        "node-progress-test01",
				"tier":           "silver",
				"credits_earned": 500.0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeContributionTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "node-progress-test01",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	// Should show progress to next tier
}

// Test node contribution with node error (fallback)
func TestNodeContributionTool_Execute_NodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "contribution") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_id": "node-test-errortest01",
				"contribution": map[string]interface{}{
					"cpu_usage_hours": 100.0,
					"gpu_usage_hours": 50.0,
					"live_gflops":     500.0,
					"network_percent": 1.0,
				},
				"ranking": map[string]interface{}{
					"rank":        50,
					"total_nodes": 100,
					"tier":        "bronze",
				},
			})
		} else {
			// Node endpoint fails
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewNodeContributionTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"node_id": "node-test-errortest01",
	}

	result := tool.Execute(ctx, args)

	// Should still succeed with contribution data
	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}
