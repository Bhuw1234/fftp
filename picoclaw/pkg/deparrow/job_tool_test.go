//go:build unit

package deparrow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sipeed/picoclaw/pkg/tools"
)

func TestJobTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobTool(client)

	if tool.Name() != "deparrow_submit_job" {
		t.Errorf("Name() = %s, want deparrow_submit_job", tool.Name())
	}
}

func TestJobTool_Description(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobTool(client)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !contains(desc, "Submit a compute job") {
		t.Error("Description should mention submitting jobs")
	}
}

func TestJobTool_Parameters(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobTool(client)

	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil")
	}

	// Check required parameters
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not found or wrong type")
	}

	if _, ok := props["image"]; !ok {
		t.Error("image parameter not found")
	}
	if _, ok := props["command"]; !ok {
		t.Error("command parameter not found")
	}
	if _, ok := props["cpu"]; !ok {
		t.Error("cpu parameter not found")
	}
	if _, ok := props["memory"]; !ok {
		t.Error("memory parameter not found")
	}

	// Check required array
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("required not found or wrong type")
	}
	if len(required) != 1 || required[0] != "image" {
		t.Errorf("required = %v, want [image]", required)
	}
}

func TestJobTool_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-test-123",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image":   "ubuntu:latest",
		"command": "echo hello",
		"cpu":     "500m",
		"memory":  "256Mi",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "job-test-123") {
		t.Errorf("Result should contain job ID: %s", result.ForLLM)
	}
}

func TestJobTool_Execute_MissingImage(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"command": "echo hello",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing image")
	}
	if !contains(result.ForLLM, "image parameter is required") {
		t.Errorf("Error message: %s", result.ForLLM)
	}
}

func TestJobTool_Execute_EmptyImage(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image": "",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for empty image")
	}
}

func TestJobTool_Execute_WithResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-resource-test",
			"credit_deducted":   2.5,
			"remaining_balance": 97.5,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image":   "python:3.11",
		"command": "python -c 'print(1+1)'",
		"cpu":     "1000m",
		"memory":  "1Gi",
		"gpu":     "1",
		"timeout": 3600,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestJobTool_Execute_WithEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-env-test",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image":   "alpine:latest",
		"command": "env",
		"env": map[string]interface{}{
			"DEBUG":  "true",
			"ENV":    "test",
			"API_KEY": "secret123",
		},
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestJobTool_Execute_WithInputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-input-test",
			"credit_deducted":   1.5,
			"remaining_balance": 98.5,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image": "ubuntu:latest",
		"inputs": []interface{}{
			map[string]interface{}{
				"storage_source": "ipfs",
				"source":         "QmInput123",
				"path":           "/data/input",
			},
		},
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

func TestJobTool_Execute_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{
			Code:    400,
			Message: "Invalid image name",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image": "invalid image!",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error from API")
	}
}

func TestJobTool_Execute_WithPriority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-priority-test",
			"credit_deducted":   1.5,
			"remaining_balance": 98.5,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image":    "ubuntu:latest",
		"priority": 80,
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

// Test JobStatusTool
func TestJobStatusTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobStatusTool(client)

	if tool.Name() != "deparrow_job_status" {
		t.Errorf("Name() = %s, want deparrow_job_status", tool.Name())
	}
}

func TestJobStatusTool_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job_id":       "job-status-test",
			"status":       "completed",
			"user_id":      "user-123",
			"credit_cost":  2.5,
			"submitted_at": "2024-01-01T00:00:00Z",
			"results": map[string]interface{}{
				"output_cid":       "QmTest",
				"stdout":           "Hello!",
				"exit_code":        0,
				"duration_seconds": 5.5,
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobStatusTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"job_id": "job-status-test",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "job-status-test") {
		t.Errorf("Result should contain job ID: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "completed") {
		t.Errorf("Result should contain status: %s", result.ForLLM)
	}
}

func TestJobStatusTool_Execute_MissingJobID(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobStatusTool(client)
	ctx := context.Background()

	args := map[string]interface{}{}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing job_id")
	}
}

func TestJobStatusTool_Execute_JobNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{
			Code:    404,
			Message: "Job not found",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobStatusTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"job_id": "nonexistent",
	}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for job not found")
	}
}

// Test JobListTool
func TestJobListTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobListTool(client)

	if tool.Name() != "deparrow_list_jobs" {
		t.Errorf("Name() = %s, want deparrow_list_jobs", tool.Name())
	}
}

func TestJobListTool_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jobs": []map[string]interface{}{
				{
					"job_id":       "job-1",
					"status":       "completed",
					"credit_cost":  1.0,
					"submitted_at": "2024-01-01T00:00:00Z",
					"spec": map[string]interface{}{
						"image": "ubuntu:latest",
					},
				},
				{
					"job_id":       "job-2",
					"status":       "running",
					"credit_cost":  2.0,
					"submitted_at": "2024-01-01T01:00:00Z",
					"spec": map[string]interface{}{
						"image": "python:3.11",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobListTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "job-1") {
		t.Errorf("Result should contain job-1: %s", result.ForLLM)
	}
}

func TestJobListTool_Execute_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jobs": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobListTool(client)
	ctx := context.Background()

	result := tool.Execute(ctx, nil)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "No jobs found") {
		t.Errorf("Result should indicate no jobs: %s", result.ForLLM)
	}
}

// Test JobCancelTool
func TestJobCancelTool_Name(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobCancelTool(client)

	if tool.Name() != "deparrow_cancel_job" {
		t.Errorf("Name() = %s, want deparrow_cancel_job", tool.Name())
	}
}

func TestJobCancelTool_Execute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "cancelled",
			"job_id":            "job-cancel-test",
			"refund_amount":     0.75,
			"remaining_balance": 99.25,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobCancelTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"job_id": "job-cancel-test",
	}

	result := tool.Execute(ctx, args)

	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "cancelled") {
		t.Errorf("Result should mention cancellation: %s", result.ForLLM)
	}
	if !contains(result.ForLLM, "0.75") {
		t.Errorf("Result should show refund amount: %s", result.ForLLM)
	}
}

func TestJobCancelTool_Execute_MissingJobID(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")
	tool := NewJobCancelTool(client)
	ctx := context.Background()

	args := map[string]interface{}{}

	result := tool.Execute(ctx, args)

	if !result.IsError {
		t.Error("Expected error for missing job_id")
	}
}

// Test tool interface compliance
func TestJobTools_InterfaceCompliance(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-token")

	var _ tools.Tool = NewJobTool(client)
	var _ tools.Tool = NewJobStatusTool(client)
	var _ tools.Tool = NewJobListTool(client)
	var _ tools.Tool = NewJobCancelTool(client)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test timeout parameter parsing
func TestJobTool_Execute_TimeoutParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-timeout-test",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	// Test with float64 (JSON default)
	args := map[string]interface{}{
		"image":   "ubuntu:latest",
		"timeout": float64(300),
	}

	result := tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with float64 timeout returned error: %s", result.ForLLM)
	}

	// Test with int
	args = map[string]interface{}{
		"image":   "ubuntu:latest",
		"timeout": 600,
	}

	result = tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with int timeout returned error: %s", result.ForLLM)
	}
}

// Test priority parameter parsing
func TestJobTool_Execute_PriorityParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-priority-test",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	// Test with float64
	args := map[string]interface{}{
		"image":    "ubuntu:latest",
		"priority": float64(50),
	}

	result := tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with float64 priority returned error: %s", result.ForLLM)
	}

	// Test with int
	args = map[string]interface{}{
		"image":    "ubuntu:latest",
		"priority": 75,
	}

	result = tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() with int priority returned error: %s", result.ForLLM)
	}
}

// Test command parsing
func TestJobTool_Execute_CommandParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-cmd-test",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	// Test command string splitting
	args := map[string]interface{}{
		"image":   "ubuntu:latest",
		"command": "echo hello world",
	}

	result := tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

// Test args parameter
func TestJobTool_Execute_WithArgs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "submitted",
			"job_id":            "job-args-test",
			"credit_deducted":   1.0,
			"remaining_balance": 99.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tool := NewJobTool(client)
	ctx := context.Background()

	args := map[string]interface{}{
		"image":   "ubuntu:latest",
		"command": "python",
		"args":    []interface{}{"-c", "print(1+1)"},
	}

	result := tool.Execute(ctx, args)
	if result.IsError {
		t.Errorf("Execute() returned error: %s", result.ForLLM)
	}
}

// Test marshalJobInfo helper
func TestMarshalJobInfo(t *testing.T) {
	job := &Job{
		ID:     "test-job",
		Status: JobStatusRunning,
	}

	output := marshalJobInfo(job)
	if output == "" {
		t.Error("marshalJobInfo returned empty string")
	}
	if !contains(output, "test-job") {
		t.Errorf("Output should contain job ID: %s", output)
	}
}
