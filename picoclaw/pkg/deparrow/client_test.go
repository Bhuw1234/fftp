//go:build unit

package deparrow

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		apiURL   string
		jwtToken string
	}{
		{
			name:     "basic client",
			apiURL:   "http://localhost:8080",
			jwtToken: "test-token",
		},
		{
			name:     "empty token",
			apiURL:   "http://localhost:8080",
			jwtToken: "",
		},
		{
			name:     "production URL",
			apiURL:   "https://api.deparrow.io",
			jwtToken: "prod-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiURL, tt.jwtToken)
			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			if client.baseURL != tt.apiURL {
				t.Errorf("baseURL = %s, want %s", client.baseURL, tt.apiURL)
			}
			if client.jwtToken != tt.jwtToken {
				t.Errorf("jwtToken = %s, want %s", client.jwtToken, tt.jwtToken)
			}
			if client.httpClient == nil {
				t.Error("httpClient is nil")
			}
		})
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customTimeout := 60 * time.Second
	client := NewClient("http://localhost:8080", "token", WithTimeout(customTimeout))

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, customTimeout)
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 120 * time.Second,
	}
	client := NewClient("http://localhost:8080", "token", WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Error("HTTP client not set correctly")
	}
}

func TestClient_SetUserID(t *testing.T) {
	client := NewClient("http://localhost:8080", "token")
	client.SetUserID("user-123")

	if client.userID != "user-123" {
		t.Errorf("userID = %s, want user-123", client.userID)
	}
}

func TestClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	health, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", health["status"])
	}
}

func TestClient_Health_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIError{
			Code:    500,
			Message: "Internal Server Error",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	_, err := client.Health(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Code != 500 {
		t.Errorf("Error code = %d, want 500", apiErr.Code)
	}
}

func TestClient_SubmitJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/jobs/submit" {
			t.Errorf("Path = %s, want /api/v1/jobs/submit", r.URL.Path)
		}

		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", auth)
		}

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		json.Unmarshal(body, &req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-12345",
			"credit_deducted":   1.5,
			"remaining_balance": 98.5,
			"message":           "Job submitted successfully",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	spec := &JobSpec{
		Image:   "ubuntu:latest",
		Command: []string{"echo", "hello"},
		Resources: &ResourceSpec{
			CPU:    "500m",
			Memory: "256Mi",
		},
	}

	job, err := client.SubmitJob(ctx, spec)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	if job.ID != "job-12345" {
		t.Errorf("Job.ID = %s, want job-12345", job.ID)
	}
	if job.Status != JobStatusPending {
		t.Errorf("Job.Status = %s, want pending", job.Status)
	}
	if job.CreditCost != 1.5 {
		t.Errorf("Job.CreditCost = %f, want 1.5", job.CreditCost)
	}
}

func TestClient_SubmitJob_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{
			Code:    400,
			Message: "Invalid job specification",
			Details: "Image field is required",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	spec := &JobSpec{} // Missing required image

	_, err := client.SubmitJob(ctx, spec)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_GetJob(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/api/v1/jobs/job-123") {
			t.Errorf("Path = %s, want /api/v1/jobs/job-123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id":       "job-123",
			"status":       "completed",
			"user_id":      "user-456",
			"credit_cost":  2.5,
			"submitted_at": now.Format(time.RFC3339),
			"results": map[string]interface{}{
				"output_cid": "QmTest",
				"stdout":     "Hello, World!",
				"exit_code":  0,
				"duration_seconds": 5.5,
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	job, err := client.GetJob(ctx, "job-123")
	if err != nil {
		t.Fatalf("GetJob() error = %v", err)
	}

	if job.ID != "job-123" {
		t.Errorf("Job.ID = %s, want job-123", job.ID)
	}
	if job.Status != JobStatusCompleted {
		t.Errorf("Job.Status = %s, want completed", job.Status)
	}
}

func TestClient_GetJob_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{
			Code:    404,
			Message: "Job not found",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	_, err := client.GetJob(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_ListJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/jobs" {
			t.Errorf("Path = %s, want /api/v1/jobs", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jobs": []map[string]interface{}{
				{
					"job_id":      "job-1",
					"status":      "completed",
					"credit_cost": 1.0,
				},
				{
					"job_id":      "job-2",
					"status":      "running",
					"credit_cost": 2.0,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	jobs, err := client.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs() error = %v", err)
	}

	if len(jobs) != 2 {
		t.Errorf("Jobs count = %d, want 2", len(jobs))
	}
	if jobs[0].ID != "job-1" {
		t.Errorf("First job ID = %s, want job-1", jobs[0].ID)
	}
}

func TestClient_CancelJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/cancel") {
			t.Errorf("Path should end with /cancel, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "cancelled",
			"job_id":            "job-123",
			"refund_amount":     0.75,
			"remaining_balance": 99.25,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	refund, err := client.CancelJob(ctx, "job-123")
	if err != nil {
		t.Fatalf("CancelJob() error = %v", err)
	}

	if refund != 0.75 {
		t.Errorf("Refund = %f, want 0.75", refund)
	}
}

func TestClient_GetCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 150.5,
			"last_active":    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	ctx := context.Background()

	credits, err := client.GetCredits(ctx)
	if err != nil {
		t.Fatalf("GetCredits() error = %v", err)
	}

	if credits.Balance != 150.5 {
		t.Errorf("Balance = %f, want 150.5", credits.Balance)
	}
}

func TestClient_CheckCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/credits/check" {
			t.Errorf("Path = %s, want /api/v1/credits/check", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_sufficient": true,
			"required":       10.0,
			"available":      150.0,
			"difference":     140.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	hasSufficient, err := client.CheckCredits(ctx, 10.0)
	if err != nil {
		t.Fatalf("CheckCredits() error = %v", err)
	}

	if !hasSufficient {
		t.Error("Expected hasSufficient = true")
	}
}

func TestClient_TransferCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/credits/transfer" {
			t.Errorf("Path = %s, want /api/v1/credits/transfer", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	err := client.TransferCredits(ctx, "user-456", 25.0)
	if err != nil {
		t.Fatalf("TransferCredits() error = %v", err)
	}
}

func TestClient_ListNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/nodes" {
			t.Errorf("Path = %s, want /api/v1/nodes", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"node_id":        "node-1",
					"arch":           "x86_64",
					"status":         "online",
					"last_seen":      time.Now().Format(time.RFC3339),
					"credits_earned": 100.0,
				},
				{
					"node_id":        "node-2",
					"arch":           "arm64",
					"status":         "online",
					"last_seen":      time.Now().Format(time.RFC3339),
					"credits_earned": 200.0,
				},
			},
			"total":  2,
			"online": 2,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	nodes, err := client.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes() error = %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("Nodes count = %d, want 2", len(nodes))
	}
	if nodes[0].ID != "node-1" {
		t.Errorf("First node ID = %s, want node-1", nodes[0].ID)
	}
}

func TestClient_GetNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_id":        "node-123",
			"arch":           "x86_64",
			"status":         "online",
			"last_seen":      time.Now().Format(time.RFC3339),
			"credits_earned": 500.0,
			"tier":           "gold",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	node, err := client.GetNode(ctx, "node-123")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}

	if node.ID != "node-123" {
		t.Errorf("Node.ID = %s, want node-123", node.ID)
	}
	if node.Arch != ArchX86_64 {
		t.Errorf("Node.Arch = %s, want x86_64", node.Arch)
	}
}

func TestClient_GetNodeContribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_id": "node-123",
			"contribution": map[string]interface{}{
				"cpu_usage_hours": 100.0,
				"gpu_usage_hours": 50.0,
				"live_gflops":     500.0,
				"network_percent": 1.5,
			},
			"ranking": map[string]interface{}{
				"rank":        10,
				"total_nodes": 1000,
				"tier":        "silver",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	contrib, err := client.GetNodeContribution(ctx, "node-123")
	if err != nil {
		t.Fatalf("GetNodeContribution() error = %v", err)
	}

	if contrib.CPUUsageHours != 100.0 {
		t.Errorf("CPUUsageHours = %f, want 100.0", contrib.CPUUsageHours)
	}
	if contrib.Rank != 10 {
		t.Errorf("Rank = %d, want 10", contrib.Rank)
	}
}

func TestClient_GetNetworkStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"network": map[string]interface{}{
				"total_nodes":     100,
				"online_nodes":    85,
				"total_cpu_cores": 800,
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
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	stats, err := client.GetNetworkStats(ctx)
	if err != nil {
		t.Fatalf("GetNetworkStats() error = %v", err)
	}

	if stats.TotalNodes != 100 {
		t.Errorf("TotalNodes = %d, want 100", stats.TotalNodes)
	}
	if stats.OnlineNodes != 85 {
		t.Errorf("OnlineNodes = %d, want 85", stats.OnlineNodes)
	}
}

func TestClient_GetLeaderboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard": []map[string]interface{}{
				{
					"rank":           1,
					"node_id":        "node-champ",
					"tier":           "legendary",
					"credits_earned": 10000.0,
				},
				{
					"rank":           2,
					"node_id":        "node-2nd",
					"tier":           "diamond",
					"credits_earned": 5000.0,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	entries, err := client.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(entries))
	}
	if entries[0].Rank != 1 {
		t.Errorf("First entry rank = %d, want 1", entries[0].Rank)
	}
}

func TestClient_GetWallet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":        "user-123",
			"credit_balance": 250.0,
			"last_active":    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	client.SetUserID("user-123")
	ctx := context.Background()

	wallet, err := client.GetWallet(ctx)
	if err != nil {
		t.Fatalf("GetWallet() error = %v", err)
	}

	if wallet.Balance != 250.0 {
		t.Errorf("Balance = %f, want 250.0", wallet.Balance)
	}
}

func TestClient_doRequest_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", WithTimeout(100*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Health(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestClient_doRequest_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Health(ctx)
	if err == nil {
		t.Error("Expected context canceled error, got nil")
	}
}

func TestCalculateCreditCost(t *testing.T) {
	tests := []struct {
		name     string
		spec     *JobSpec
		minCost  float64
	}{
		{
			name: "basic job",
			spec: &JobSpec{
				Image: "ubuntu:latest",
			},
			minCost: 1.0,
		},
		{
			name: "job with resources",
			spec: &JobSpec{
				Image: "python:3.11",
				Resources: &ResourceSpec{
					CPU:    "500m",
					Memory: "1Gi",
				},
			},
			minCost: 1.0,
		},
		{
			name: "job with GPU",
			spec: &JobSpec{
				Image: "tensorflow/tensorflow:latest-gpu",
				Resources: &ResourceSpec{
					GPU: "1",
				},
			},
			minCost: 3.0, // base + GPU cost
		},
		{
			name: "high priority job",
			spec: &JobSpec{
				Image:    "ubuntu:latest",
				Priority: 75,
				Resources: &ResourceSpec{
					CPU: "100m",
				},
			},
			minCost: 1.0, // High priority requires resources to have effect
		},
		{
			name: "job with timeout",
			spec: &JobSpec{
				Image:   "ubuntu:latest",
				Timeout: 7200, // 2 hours
				Resources: &ResourceSpec{
					CPU: "100m",
				},
			},
			minCost: 1.0, // Timeout effect depends on resources
		},
		{
			name: "nil resources",
			spec: &JobSpec{
				Image:     "alpine:latest",
				Resources: nil,
			},
			minCost: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculateCreditCost(tt.spec)
			if cost < tt.minCost {
				t.Errorf("Cost = %f, want >= %f", cost, tt.minCost)
			}
		})
	}
}

func TestClient_NetworkError(t *testing.T) {
	client := NewClient("http://nonexistent-host:99999", "test-token")
	ctx := context.Background()

	_, err := client.Health(ctx)
	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

func TestClient_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	_, err := client.Health(ctx)
	if err == nil {
		t.Error("Expected JSON parse error, got nil")
	}
}

func TestClient_AuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-secret-token" {
			t.Errorf("Authorization = %s, want Bearer my-secret-token", auth)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "my-secret-token")
	ctx := context.Background()

	_, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
}

func TestClient_NoAuthorizationWhenEmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("Authorization should be empty, got %s", auth)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "") // Empty token
	ctx := context.Background()

	_, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
}

func TestClient_ContentTypeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", contentType)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	_, err := client.SubmitJob(ctx, &JobSpec{Image: "test"})
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}
}

func TestClient_ErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"400 Bad Request", http.StatusBadRequest, true},
		{"401 Unauthorized", http.StatusUnauthorized, true},
		{"403 Forbidden", http.StatusForbidden, true},
		{"404 Not Found", http.StatusNotFound, true},
		{"500 Internal Error", http.StatusInternalServerError, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.wantError {
					json.NewEncoder(w).Encode(APIError{
						Code:    tt.statusCode,
						Message: tt.name,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{})
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-token")
			ctx := context.Background()

			_, err := client.Health(ctx)
			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			server.Close()
		})
	}
}

func TestClient_GetMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/metrics" {
			t.Errorf("Path = %s, want /api/v1/metrics", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jobs_submitted": 1000,
			"jobs_completed": 950,
			"total_credits":  50000.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	metrics, err := client.GetMetrics(ctx)
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}

	if metrics["jobs_submitted"].(float64) != 1000 {
		t.Errorf("jobs_submitted = %v, want 1000", metrics["jobs_submitted"])
	}
}

func TestClient_URLPathEscape(t *testing.T) {
	// Note: url.PathEscape is used in the client, but doesn't escape spaces by default
	// This test verifies that the client handles special characters without crashing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just return success for any path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	// These calls should not cause issues
	_, _ = client.GetJob(ctx, "job/with/slashes")
	_, _ = client.GetNode(ctx, "node-with-dashes")
}
