//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bacalhau-project/bacalhau/deparrow/test-integration/testutil"
)

// GUIE2ESuite tests GUI-related functionality through the API.
// Note: Full browser-based E2E tests would require Playwright/Cypress.
// This suite tests the backend APIs that the GUI uses.
type GUIE2ESuite struct {
	suite.Suite
	mockServer *testutil.MockMetaOSServer
	client     *testutil.HTTPClient
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetupSuite initializes the test suite.
func (s *GUIE2ESuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), testutil.LongTimeout)

	// Start mock server
	s.mockServer = testutil.NewMockMetaOSServer()
	s.client = testutil.NewHTTPClient(s.mockServer.URL, "")

	// Wait for server to be healthy
	err := s.mockServer.WaitForHealthy(s.ctx)
	require.NoError(s.T(), err, "Mock server should become healthy")
}

// TearDownSuite cleans up the test suite.
func (s *GUIE2ESuite) TearDownSuite() {
	if s.mockServer != nil {
		s.mockServer.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// TestLoginFlow tests the login flow used by the GUI.
func (s *GUIE2ESuite) TestLoginFlow() {
	s.T().Run("successful login returns token and user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Login request (simulating GUI login form)
		loginReq := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}

		resp, err := s.client.Post(ctx, "/api/v1/auth/login", loginReq)
		require.NoError(t, err, "Login should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Verify response structure expected by GUI
		assert.NotEmpty(t, result["token"], "Should return token for localStorage")
		assert.NotEmpty(t, result["user"], "Should return user object")

		user := result["user"].(map[string]interface{})
		assert.NotEmpty(t, user["id"], "User should have ID")
		assert.NotEmpty(t, user["email"], "User should have email")
		assert.NotEmpty(t, user["name"], "User should have name")
		assert.NotEmpty(t, user["credits"], "User should have credits for wallet display")
	})

	s.T().Run("invalid credentials shows error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		loginReq := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		resp, err := s.client.Post(ctx, "/api/v1/auth/login", loginReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode, "Should return 401")

		// GUI expects error message in response
		body, _ := testutil.ReadString(resp)
		assert.Contains(t, body, "error", "Should contain error message")
	})

	s.T().Run("registration flow", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		regReq := map[string]interface{}{
			"email":    "newuser@test.com",
			"password": "SecurePass123!",
			"name":     "New User",
		}

		resp, err := s.client.Post(ctx, "/api/v1/auth/register", regReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Same response structure as login
		assert.NotEmpty(t, result["token"], "Should return token")
		assert.NotEmpty(t, result["user"], "Should return user object")
	})
}

// TestDashboardData tests the data needed for the Dashboard page.
func (s *GUIE2ESuite) TestDashboardData() {
	s.T().Run("network stats for dashboard", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add some nodes for stats
		s.mockServer.AddTestNode("dashboard-node-1")
		s.mockServer.AddTestNode("dashboard-node-2")

		resp, err := s.client.Get(ctx, "/api/v1/network/contribution")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Dashboard needs these fields
		network := result["network"].(map[string]interface{})
		assert.Contains(t, network, "total_nodes", "Need total_nodes for dashboard")
		assert.Contains(t, network, "online_nodes", "Need online_nodes for dashboard")
		assert.Contains(t, network, "total_cpu_cores", "Need total_cpu_cores for dashboard")
		assert.Contains(t, network, "total_gpu_count", "Need total_gpu_count for dashboard")
		assert.Contains(t, network, "live_gflops", "Need live_gflops for contribution rings")
	})

	s.T().Run("leaderboard for dashboard widget", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/network/leaderboard")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "leaderboard", "Need leaderboard for dashboard")
	})
}

// TestJobsPage tests the Jobs page functionality.
func (s *GUIE2ESuite) TestJobsPage() {
	s.T().Run("list jobs for jobs table", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Setup user with credits and submit a job
		s.mockServer.SetCredits("test-user", 100.0)

		jobReq := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "python:3.9",
				"command": []string{"python", "-c", "print('test')"},
			},
			"credit_cost": 5.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobReq)
		require.NoError(t, err)
		resp.Body.Close()

		// Now list jobs
		resp, err = s.client.Get(ctx, "/api/v1/jobs")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "jobs", "Need jobs array for table")
		jobs := result["jobs"].([]interface{})
		if len(jobs) > 0 {
			job := jobs[0].(map[string]interface{})
			// GUI needs these fields for the jobs table
			assert.Contains(t, job, "job_id", "Need job_id")
			assert.Contains(t, job, "status", "Need status for badge")
			assert.Contains(t, job, "credit_cost", "Need credit_cost")
			assert.Contains(t, job, "submitted_at", "Need submitted_at for timestamp")
		}
	})

	s.T().Run("submit job from GUI", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		s.mockServer.SetCredits("test-user", 50.0)

		// Job spec as submitted from GUI form
		jobReq := map[string]interface{}{
			"name": "Test Job from GUI",
			"spec": map[string]interface{}{
				"engine":  "docker",
				"image":   "ubuntu:latest",
				"command": []string{"echo", "Hello from GUI"},
				"resources": map[string]interface{}{
					"cpu":    "500m",
					"memory": "256Mi",
				},
			},
			"priority":   50,
			"credit_cost": 10.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Job submission should succeed")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.NotEmpty(t, result["job_id"], "Should return job ID for navigation")
	})

	s.T().Run("job estimation before submit", func(t *testing.T) {
		// GUI would call this to show estimated cost before submission
		// This is a placeholder for the estimate endpoint
		t.Skip("Estimate endpoint not yet implemented in mock server")
	})
}

// TestWalletPage tests the Wallet page functionality.
func (s *GUIE2ESuite) TestWalletPage() {
	s.T().Run("credit balance display", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		s.mockServer.SetCredits("test-user", 500.0)

		resp, err := s.client.Get(ctx, "/api/v1/credits/balance")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Wallet needs these fields
		assert.Contains(t, result, "user_id", "Need user_id")
		assert.Contains(t, result, "credit_balance", "Need credit_balance")
		assert.Equal(t, 500.0, result["credit_balance"], "Balance should match")
	})

	s.T().Run("transfer credits from wallet", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Setup
		s.mockServer.AddTestUser("wallet-test-user", "wallet@test.com", "pass")
		s.mockServer.SetCredits("test-user", 100.0)
		s.mockServer.SetCredits("wallet-test-user", 0)

		transferReq := map[string]interface{}{
			"from_user": "test-user",
			"to_user":   "wallet-test-user",
			"amount":    25.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/credits/transfer", transferReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Transfer should succeed")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.True(t, result["success"].(bool), "Should indicate success")
		assert.NotEmpty(t, result["transaction_id"], "Should return transaction ID")
	})
}

// TestNodesPage tests the Nodes page functionality.
func (s *GUIE2ESuite) TestNodesPage() {
	s.T().Run("list nodes for nodes table", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add nodes
		for i := 0; i < 3; i++ {
			s.mockServer.AddTestNode(fmt.Sprintf("nodes-page-%d", i))
		}

		resp, err := s.client.Get(ctx, "/api/v1/nodes")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "nodes", "Need nodes array")
		assert.Contains(t, result, "total", "Need total count")

		nodes := result["nodes"].([]interface{})
		if len(nodes) > 0 {
			node := nodes[0].(map[string]interface{})
			// Nodes table needs these fields
			assert.Contains(t, node, "node_id", "Need node_id")
			assert.Contains(t, node, "status", "Need status for badge")
			assert.Contains(t, node, "arch", "Need arch")
			assert.Contains(t, node, "last_seen", "Need last_seen")
		}
	})

	s.T().Run("node stats summary", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/network/contribution")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Nodes page summary cards need these
		network := result["network"].(map[string]interface{})
		assert.Contains(t, network, "total_nodes", "Need total_nodes")
		assert.Contains(t, network, "online_nodes", "Need online_nodes")
	})
}

// TestSettingsPage tests the Settings page functionality.
func (s *GUIE2ESuite) TestSettingsPage() {
	s.T().Run("system health for settings", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Settings needs health status
		assert.Contains(t, result, "status", "Need status")
		assert.Contains(t, result, "version", "Need version")
		assert.Contains(t, result, "services", "Need services status")
	})
}

// TestAgentConsolePage tests the Agent Console page functionality.
func (s *GUIE2ESuite) TestAgentConsolePage() {
	s.T().Run("agent status display", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/agent/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		// Agent console needs these fields
		assert.Contains(t, result, "id", "Need agent ID")
		assert.Contains(t, result, "status", "Need status for indicator")
		assert.Contains(t, result, "uptime", "Need uptime")
		assert.Contains(t, result, "tools", "Need tools list")
	})

	s.T().Run("agent chat interface", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Test chat messages
		testMessages := []string{
			"What is my credit balance?",
			"Submit a simple job",
			"List available nodes",
		}

		for _, msg := range testMessages {
			chatReq := map[string]interface{}{
				"message": msg,
			}

			resp, err := s.client.Post(ctx, "/api/v1/agent/chat", chatReq)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, 200, resp.StatusCode, "Chat should succeed")

			var result map[string]interface{}
			testutil.ReadJSON(resp, &result)

			assert.Contains(t, result, "message", "Should return message")
			message := result["message"].(map[string]interface{})
			assert.Contains(t, message, "content", "Message should have content")
			assert.Contains(t, message, "timestamp", "Message should have timestamp")
		}
	})

	s.T().Run("agent tools management", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/agent/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		tools := result["tools"].([]interface{})
		if len(tools) > 0 {
			tool := tools[0].(map[string]interface{})
			// GUI needs these for tool cards
			assert.Contains(t, tool, "name", "Tool needs name")
			assert.Contains(t, tool, "description", "Tool needs description")
			assert.Contains(t, tool, "enabled", "Tool needs enabled toggle")
			assert.Contains(t, tool, "calls", "Tool needs call count")
		}
	})
}

// TestProvidersPage tests the Providers page functionality.
func (s *GUIE2ESuite) TestProvidersPage() {
	s.T().Run("list providers for marketplace", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add online providers
		for i := 0; i < 3; i++ {
			node := s.mockServer.AddTestNode(fmt.Sprintf("provider-%d", i))
			node.Status = "online"
		}

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "providers", "Need providers array")

		providers := result["providers"].([]interface{})
		if len(providers) > 0 {
			provider := providers[0].(map[string]interface{})
			// Provider cards need these
			assert.Contains(t, provider, "id", "Need provider ID")
			assert.Contains(t, provider, "name", "Need provider name")
			assert.Contains(t, provider, "status", "Need status")
			assert.Contains(t, provider, "location", "Need location")
			assert.Contains(t, provider, "pricing", "Need pricing")
			assert.Contains(t, provider, "stats", "Need stats")
		}
	})

	s.T().Run("provider pricing display", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		providers := result["providers"].([]interface{})
		if len(providers) > 0 {
			provider := providers[0].(map[string]interface{})
			pricing := provider["pricing"].(map[string]interface{})
			// Pricing breakdown needs these
			assert.Contains(t, pricing, "cpu_per_hour", "Need CPU pricing")
			assert.Contains(t, pricing, "memory_per_gb_hour", "Need memory pricing")
			assert.Contains(t, pricing, "gpu_per_hour", "Need GPU pricing")
		}
	})
}

// TestWebSocketConnection tests WebSocket connectivity for real-time updates.
func (s *GUIE2ESuite) TestWebSocketConnection() {
	s.T().Run("websocket endpoint available", func(t *testing.T) {
		// WebSocket testing would require gorilla/websocket
		// This is a placeholder for full browser-based E2E tests
		t.Skip("WebSocket tests require browser environment or gorilla/websocket")
	})
}

// TestGUIE2ESuite runs the test suite.
func TestGUIE2ESuite(t *testing.T) {
	suite.Run(t, new(GUIE2ESuite))
}
