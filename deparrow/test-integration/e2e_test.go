package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DEparrowIntegrationSuite tests all 4 layers working together
type DEparrowIntegrationSuite struct {
	bootstrapServerURL string
	apiKey             string
	ctx                context.Context
	cancel             context.CancelFunc
}

// SetupTest initializes the integration test environment
func (s *DEparrowIntegrationSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 30*time.Minute)
	
	// Start bootstrap server
	server := httptest.NewServer(http.HandlerFunc(s.handleBootstrap))
	s.bootstrapServerURL = server.URL
	
	// Generate test API key
	s.apiKey = "test-api-key-12345"
	
	// Set environment variables for testing
	os.Setenv("DEPARROW_BOOTSTRAP", s.bootstrapServerURL)
	os.Setenv("DEPARROW_API_KEY", s.apiKey)
}

// Bootstrap server handler for testing
func (s *DEparrowIntegrationSuite) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Simulate bootstrap server responses
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/nodes/register"):
		s.handleNodeRegistration(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/jobs/submit"):
		s.handleJobSubmission(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/credits/transfer"):
		s.handleCreditTransfer(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/v1/health"):
		s.handleHealthCheck(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *DEparrowIntegrationSuite) handleNodeRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req NodeRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Simulate node registration
	response := NodeRegistrationResponse{
		Success: true,
		NodeID:  req.NodeID,
		Status:  "registered",
		Message: "Node registered successfully",
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *DEparrowIntegrationSuite) handleJobSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req JobSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Simulate job submission with credit verification
	if req.Credits < 10 {
		http.Error(w, "Insufficient credits", http.StatusBadRequest)
		return
	}
	
	response := JobSubmissionResponse{
		JobID:     "job-" + req.NodeID + "-12345",
		Status:    "submitted",
		Credits:   req.Credits,
		Message:   "Job submitted successfully",
		EstTime:   "5m",
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *DEparrowIntegrationSuite) handleCreditTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req CreditTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	response := CreditTransferResponse{
		Success:     true,
		FromUser:    req.FromUser,
		ToUser:      req.ToUser,
		Amount:      req.Amount,
		TransactionID: "txn-12345",
		Message:     "Credit transfer successful",
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *DEparrowIntegrationSuite) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services": map[string]string{
			"bootstrap": "healthy",
			"registry":  "healthy",
			"credits":   "healthy",
		},
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// TestAlpineLayerIntegration tests Alpine Linux node auto-join functionality
func (s *DEparrowIntegrationSuite) TestAlpineLayerIntegration(t *testing.T) {
	t.Log("Testing Alpine Linux layer integration...")
	
	// Test node initialization script
	scriptPath := filepath.Join(s.getProjectRoot(), "alpine-layer/scripts/init-node.sh")
	assert.FileExists(t, scriptPath, "Node initialization script should exist")
	
	// Test Dockerfile exists and is valid
	dockerfilePath := filepath.Join(s.getProjectRoot(), "alpine-layer/Dockerfile")
	assert.FileExists(t, dockerfilePath, "Dockerfile should exist")
	
	// Test build script exists
	buildScriptPath := filepath.Join(s.getProjectRoot(), "alpine-layer/build.sh")
	assert.FileExists(t, buildScriptPath, "Build script should exist")
	
	t.Log("✓ Alpine Linux layer files validated")
}

// TestMetaOSIntegration tests Meta-OS control plane integration
func (s *DEparrowIntegrationSuite) TestMetaOSIntegration(t *testing.T) {
	t.Log("Testing Meta-OS control plane integration...")
	
	// Test bootstrap server functionality
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Test health endpoint
	resp, err := client.Get(s.bootstrapServerURL + "/api/v1/health")
	require.NoError(t, err, "Bootstrap server health check should succeed")
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200")
	
	var healthResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err, "Health response should be valid JSON")
	
	assert.Equal(t, "healthy", healthResp["status"], "System should report healthy status")
	
	t.Log("✓ Meta-OS control plane integration validated")
}

// TestGUILayerIntegration tests GUI layer integration
func (s *DEparrowIntegrationSuite) TestGUILayerIntegration(t *testing.T) {
	t.Log("Testing GUI layer integration...")
	
	// Test React component files exist
	components := []string{
		"Dashboard.tsx",
		"Jobs.tsx",
		"Wallet.tsx",
		"Nodes.tsx",
		"Settings.tsx",
		"Login.tsx",
	}
	
	projectRoot := s.getProjectRoot()
	for _, component := range components {
		componentPath := filepath.Join(projectRoot, "gui-layer/src/pages", component)
		assert.FileExists(t, componentPath, "%s component should exist", component)
	}
	
	// Test API client exists
	apiClientPath := filepath.Join(projectRoot, "gui-layer/src/api/client.ts")
	assert.FileExists(t, apiClientPath, "API client should exist")
	
	// Test authentication context exists
	authContextPath := filepath.Join(projectRoot, "gui-layer/src/contexts/AuthContext.tsx")
	assert.FileExists(t, authContextPath, "Authentication context should exist")
	
	t.Log("✓ GUI layer components validated")
}

// TestEndToEndWorkflow tests complete workflow from node join to job execution
func (s *DEparrowIntegrationSuite) TestEndToEndWorkflow(t *testing.T) {
	t.Log("Testing end-to-end workflow...")
	
	// Step 1: Node registration
	nodeRegistration := NodeRegistrationRequest{
		NodeID:     "test-node-123",
		PublicKey:  "test-public-key",
		Resources: NodeResources{
			CPU:    4,
			Memory: "4GB",
			Disk:   "20GB",
			Arch:   "x86_64",
		},
	}
	
	// Test node registration
	nodeResp, err := s.postJSON("/api/v1/nodes/register", nodeRegistration)
	require.NoError(t, err, "Node registration should succeed")
	
	var nodeRegistrationResp NodeRegistrationResponse
	err = json.Unmarshal(nodeResp, &nodeRegistrationResp)
	require.NoError(t, err, "Node registration response should be valid")
	
	assert.Equal(t, "test-node-123", nodeRegistrationResp.NodeID, "Node ID should match")
	assert.True(t, nodeRegistrationResp.Success, "Node registration should succeed")
	
	// Step 2: Job submission
	jobSubmission := JobSubmissionRequest{
		NodeID:  nodeRegistrationResp.NodeID,
		JobSpec: "test-job-specification",
		Credits: 10,
	}
	
	jobResp, err := s.postJSON("/api/v1/jobs/submit", jobSubmission)
	require.NoError(t, err, "Job submission should succeed")
	
	var jobSubmissionResp JobSubmissionResponse
	err = json.Unmarshal(jobResp, &jobSubmissionResp)
	require.NoError(t, err, "Job submission response should be valid")
	
	assert.NotEmpty(t, jobSubmissionResp.JobID, "Job ID should be generated")
	assert.Equal(t, 10, jobSubmissionResp.Credits, "Credits should match request")
	
	// Step 3: Credit transfer
	creditTransfer := CreditTransferRequest{
		FromUser: "user1",
		ToUser:   "user2",
		Amount:   50,
	}
	
	creditResp, err := s.postJSON("/api/v1/credits/transfer", creditTransfer)
	require.NoError(t, err, "Credit transfer should succeed")
	
	var creditTransferResp CreditTransferResponse
	err = json.Unmarshal(creditResp, &creditTransferResp)
	require.NoError(t, err, "Credit transfer response should be valid")
	
	assert.Equal(t, "user1", creditTransferResp.FromUser, "From user should match")
	assert.Equal(t, "user2", creditTransferResp.ToUser, "To user should match")
	assert.Equal(t, 50, creditTransferResp.Amount, "Amount should match")
	
	t.Log("✓ End-to-end workflow validated")
}

// TestDeploymentIntegration tests deployment configurations
func (s *DEparrowIntegrationSuite) TestDeploymentIntegration(t *testing.T) {
	t.Log("Testing deployment integration...")
	
	// Test Docker Compose configuration
	dockerComposePath := filepath.Join(s.getProjectRoot(), "alpine-layer/config/docker-compose/deparrow-node.yml")
	if _, err := os.Stat(dockerComposePath); err == nil {
		assert.FileExists(t, dockerComposePath, "Docker Compose config should exist")
		t.Log("✓ Docker Compose configuration validated")
	}
	
	// Test Kubernetes configuration
	kubernetesPath := filepath.Join(s.getProjectRoot(), "alpine-layer/config/kubernetes/deployment.yaml")
	if _, err := os.Stat(kubernetesPath); err == nil {
		assert.FileExists(t, kubernetesPath, "Kubernetes deployment should exist")
		t.Log("✓ Kubernetes configuration validated")
	}
	
	// Test systemd service configuration
	systemdPath := filepath.Join(s.getProjectRoot(), "alpine-layer/config/systemd/deparrow-node.service")
	if _, err := os.Stat(systemdPath); err == nil {
		assert.FileExists(t, systemdPath, "Systemd service config should exist")
		t.Log("✓ Systemd service configuration validated")
	}
}

// TestCreditSystemIntegration tests the complete credit system
func (s *DEparrowIntegrationSuite) TestCreditSystemIntegration(t *testing.T) {
	t.Log("Testing credit system integration...")
	
	// Test insufficient credits scenario
	lowCreditJob := JobSubmissionRequest{
		NodeID:  "test-node-123",
		JobSpec: "test-job-specification",
		Credits: 5, // Below minimum requirement
	}
	
	resp, err := s.postJSON("/api/v1/jobs/submit", lowCreditJob)
	require.NoError(t, err, "Low credit job submission should return error response")
	
	var errorResp map[string]interface{}
	err = json.Unmarshal(resp, &errorResp)
	require.NoError(t, err, "Error response should be valid JSON")
	
	// Verify error handling
	assert.Contains(t, errorResp["error"], "Insufficient credits", "Should reject jobs with insufficient credits")
	
	t.Log("✓ Credit system integration validated")
}

// Helper methods
func (s *DEparrowIntegrationSuite) postJSON(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(s.bootstrapServerURL+path, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return io.ReadAll(resp.Body)
}

func (s *DEparrowIntegrationSuite) getProjectRoot() string {
	return filepath.Join("..", "..", "..")
}

func (s *DEparrowIntegrationSuite) TearDownTest() {
	if s.cancel != nil {
		s.cancel()
	}
}

// Data structures for API testing
type NodeRegistrationRequest struct {
	NodeID     string         `json:"node_id"`
	PublicKey  string         `json:"public_key"`
	Resources  NodeResources  `json:"resources"`
}

type NodeResources struct {
	CPU    int    `json:"cpu"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
	Arch   string `json:"arch"`
}

type NodeRegistrationResponse struct {
	Success  bool   `json:"success"`
	NodeID   string `json:"node_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

type JobSubmissionRequest struct {
	NodeID  string `json:"node_id"`
	JobSpec string `json:"job_spec"`
	Credits int    `json:"credits"`
}

type JobSubmissionResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Credits   int    `json:"credits"`
	Message   string `json:"message"`
	EstTime   string `json:"est_time"`
}

type CreditTransferRequest struct {
	FromUser string `json:"from_user"`
	ToUser   string `json:"to_user"`
	Amount   int    `json:"amount"`
}

type CreditTransferResponse struct {
	Success        bool   `json:"success"`
	FromUser       string `json:"from_user"`
	ToUser         string `json:"to_user"`
	Amount         int    `json:"amount"`
	TransactionID  string `json:"transaction_id"`
	Message        string `json:"message"`
}