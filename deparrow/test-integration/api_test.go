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

// APICompatibilitySuite tests all Meta-OS API endpoints.
type APICompatibilitySuite struct {
	suite.Suite
	mockServer *testutil.MockMetaOSServer
	client     *testutil.HTTPClient
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetupSuite initializes the test suite.
func (s *APICompatibilitySuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), testutil.LongTimeout)

	// Start mock server
	s.mockServer = testutil.NewMockMetaOSServer()
	s.client = testutil.NewHTTPClient(s.mockServer.URL, "")

	// Wait for server to be healthy
	err := s.mockServer.WaitForHealthy(s.ctx)
	require.NoError(s.T(), err, "Mock server should become healthy")
}

// TearDownSuite cleans up the test suite.
func (s *APICompatibilitySuite) TearDownSuite() {
	if s.mockServer != nil {
		s.mockServer.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// TestHealthEndpoint tests the health endpoint.
func (s *APICompatibilitySuite) TestHealthEndpoint() {
	s.T().Run("GET /api/v1/health returns healthy status", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/health")
		require.NoError(t, err, "Health request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse JSON response")

		assert.Equal(t, "healthy", result["status"], "Status should be healthy")
		assert.NotEmpty(t, result["timestamp"], "Should have timestamp")
		assert.NotEmpty(t, result["version"], "Should have version")
	})

	s.T().Run("health response includes services", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		services, ok := result["services"].(map[string]interface{})
		require.True(t, ok, "Should have services object")
		assert.NotEmpty(t, services["bootstrap"], "Should have bootstrap service status")
		assert.NotEmpty(t, services["registry"], "Should have registry service status")
		assert.NotEmpty(t, services["credits"], "Should have credits service status")
	})
}

// TestAuthEndpoints tests authentication endpoints.
func (s *APICompatibilitySuite) TestAuthEndpoints() {
	s.T().Run("POST /api/v1/auth/register", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		regReq := map[string]interface{}{
			"email":    "api-test@example.com",
			"password": "TestPassword123",
			"name":     "API Test User",
		}

		resp, err := s.client.Post(ctx, "/api/v1/auth/register", regReq)
		require.NoError(t, err, "Registration should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		testutil.AssertJSONContains(t, testutil.MustMarshal(result), "token", "user")
	})

	s.T().Run("POST /api/v1/auth/login - valid credentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

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

		assert.NotEmpty(t, result["token"], "Should return token")
	})

	s.T().Run("POST /api/v1/auth/login - invalid credentials", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		loginReq := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		resp, err := s.client.Post(ctx, "/api/v1/auth/login", loginReq)
		require.NoError(t, err, "Request should complete")
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode, "Should return 401 Unauthorized")
	})
}

// TestNodeEndpoints tests node management endpoints.
func (s *APICompatibilitySuite) TestNodeEndpoints() {
	s.T().Run("POST /api/v1/nodes/register", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		regReq := map[string]interface{}{
			"node_id":    "api-test-node-001",
			"public_key": "test-public-key",
			"resources": map[string]interface{}{
				"cpu":    4,
				"memory": "8Gi",
				"disk":   "100Gi",
			},
		}

		resp, err := s.client.Post(ctx, "/api/v1/nodes/register", regReq)
		require.NoError(t, err, "Node registration should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.True(t, result["success"].(bool), "Should indicate success")
		assert.Equal(t, "api-test-node-001", result["node_id"], "Should return node ID")
	})

	s.T().Run("GET /api/v1/nodes", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/nodes")
		require.NoError(t, err, "Node list should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.NotEmpty(t, result["nodes"], "Should return nodes array")
		assert.NotEmpty(t, result["total"], "Should return total count")
	})

	s.T().Run("node response structure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add a test node
		s.mockServer.AddTestNode("structure-test-node")

		resp, err := s.client.Get(ctx, "/api/v1/nodes")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		nodes := result["nodes"].([]interface{})
		require.GreaterOrEqual(t, len(nodes), 1, "Should have at least one node")

		node := nodes[0].(map[string]interface{})
		requiredFields := []string{"node_id", "arch", "status", "last_seen"}
		for _, field := range requiredFields {
			assert.Contains(t, node, field, "Node should have %s field", field)
		}
	})
}

// TestJobEndpoints tests job management endpoints.
func (s *APICompatibilitySuite) TestJobEndpoints() {
	s.T().Run("POST /api/v1/jobs/submit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Ensure sufficient credits
		s.mockServer.SetCredits("test-user", 100.0)

		jobReq := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "ubuntu:latest",
				"command": []string{"echo", "API test"},
			},
			"credit_cost": 10.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobReq)
		require.NoError(t, err, "Job submission should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		testutil.AssertJSONContains(t, testutil.MustMarshal(result), "job_id", "status", "credit_deducted")
	})

	s.T().Run("GET /api/v1/jobs", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/jobs")
		require.NoError(t, err, "Job list should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "jobs", "Should return jobs array")
	})

	s.T().Run("job submission with insufficient credits", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Set low balance
		s.mockServer.SetCredits("test-user", 5.0)

		jobReq := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "ubuntu:latest",
				"command": []string{"echo", "test"},
			},
			"credit_cost": 50.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode, "Should return 400 Bad Request")
	})
}

// TestCreditEndpoints tests credit management endpoints.
func (s *APICompatibilitySuite) TestCreditEndpoints() {
	s.T().Run("GET /api/v1/credits/balance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		s.mockServer.SetCredits("test-user", 250.0)

		resp, err := s.client.Get(ctx, "/api/v1/credits/balance")
		require.NoError(t, err, "Credit balance should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Equal(t, "test-user", result["user_id"], "Should return user ID")
		assert.Equal(t, 250.0, result["credit_balance"], "Should return correct balance")
	})

	s.T().Run("POST /api/v1/credits/transfer", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Setup users
		s.mockServer.AddTestUser("sender-user", "sender@test.com", "pass")
		s.mockServer.AddTestUser("receiver-user", "receiver@test.com", "pass")
		s.mockServer.SetCredits("sender-user", 100.0)
		s.mockServer.SetCredits("receiver-user", 0.0)

		transferReq := map[string]interface{}{
			"from_user": "sender-user",
			"to_user":   "receiver-user",
			"amount":    50.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/credits/transfer", transferReq)
		require.NoError(t, err, "Transfer should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.True(t, result["success"].(bool), "Should indicate success")
		assert.NotEmpty(t, result["transaction_id"], "Should return transaction ID")
	})

	s.T().Run("transfer with insufficient balance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		transferReq := map[string]interface{}{
			"from_user": "sender-user",
			"to_user":   "receiver-user",
			"amount":    1000.0, // More than available
		}

		resp, err := s.client.Post(ctx, "/api/v1/credits/transfer", transferReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode, "Should return 400 Bad Request")
	})
}

// TestAgentEndpoints tests agent-related endpoints.
func (s *APICompatibilitySuite) TestAgentEndpoints() {
	s.T().Run("GET /api/v1/agent/status", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/agent/status")
		require.NoError(t, err, "Agent status should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		requiredFields := []string{"id", "name", "status", "uptime", "credits_earned"}
		for _, field := range requiredFields {
			assert.Contains(t, result, field, "Should have %s field", field)
		}
	})

	s.T().Run("POST /api/v1/agent/chat", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		chatReq := map[string]interface{}{
			"message": "What can you do?",
		}

		resp, err := s.client.Post(ctx, "/api/v1/agent/chat", chatReq)
		require.NoError(t, err, "Agent chat should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "message", "Should return message")
		message := result["message"].(map[string]interface{})
		assert.Contains(t, message, "content", "Message should have content")
	})

	s.T().Run("agent tools structure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/agent/status")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		tools, ok := result["tools"].([]interface{})
		require.True(t, ok, "Should have tools array")
		require.GreaterOrEqual(t, len(tools), 1, "Should have at least one tool")

		tool := tools[0].(map[string]interface{})
		toolFields := []string{"name", "description", "enabled", "calls"}
		for _, field := range toolFields {
			assert.Contains(t, tool, field, "Tool should have %s field", field)
		}
	})
}

// TestProviderEndpoints tests provider endpoints.
func (s *APICompatibilitySuite) TestProviderEndpoints() {
	s.T().Run("GET /api/v1/providers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add online node
		node := s.mockServer.AddTestNode("provider-api-test")
		node.Status = "online"

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err, "Provider list should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "providers", "Should return providers")
	})

	s.T().Run("provider response structure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		node := s.mockServer.AddTestNode("provider-structure-test")
		node.Status = "online"

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		providers := result["providers"].([]interface{})
		if len(providers) > 0 {
			provider := providers[0].(map[string]interface{})
			requiredFields := []string{"id", "name", "status", "location", "resources", "pricing", "stats"}
			for _, field := range requiredFields {
				assert.Contains(t, provider, field, "Provider should have %s field", field)
			}
		}
	})
}

// TestNetworkEndpoints tests network statistics endpoints.
func (s *APICompatibilitySuite) TestNetworkEndpoints() {
	s.T().Run("GET /api/v1/network/contribution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/network/contribution")
		require.NoError(t, err, "Network contribution should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "network", "Should return network stats")
		assert.Contains(t, result, "tiers", "Should return tier distribution")
		assert.Contains(t, result, "timestamp", "Should return timestamp")

		network := result["network"].(map[string]interface{})
		networkFields := []string{"total_nodes", "online_nodes", "total_cpu_cores"}
		for _, field := range networkFields {
			assert.Contains(t, network, field, "Network should have %s field", field)
		}
	})

	s.T().Run("GET /api/v1/network/leaderboard", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add nodes for leaderboard
		for i := 0; i < 3; i++ {
			s.mockServer.AddTestNode(fmt.Sprintf("leader-node-%d", i))
		}

		resp, err := s.client.Get(ctx, "/api/v1/network/leaderboard")
		require.NoError(t, err, "Leaderboard should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		assert.Contains(t, result, "leaderboard", "Should return leaderboard")
	})
}

// TestErrorHandling tests API error handling.
func (s *APICompatibilitySuite) TestErrorHandling() {
	s.T().Run("invalid JSON returns error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Send invalid JSON
		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", "invalid json")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode, "Should return 400 Bad Request")
	})

	s.T().Run("non-existent endpoint returns 404", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode, "Should return 404 Not Found")
	})

	s.T().Run("error response structure", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := testutil.ReadString(resp)
		assert.Contains(t, body, "error", "Error response should contain error message")
	})
}

// TestResponseContentType tests response content types.
func (s *APICompatibilitySuite) TestResponseContentType() {
	s.T().Run("JSON responses have correct content type", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		endpoints := []string{
			"/api/v1/health",
			"/api/v1/nodes",
			"/api/v1/jobs",
			"/api/v1/credits/balance",
		}

		for _, endpoint := range endpoints {
			resp, err := s.client.Get(ctx, endpoint)
			require.NoError(t, err, "Request to %s should succeed", endpoint)
			resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			assert.Contains(t, contentType, "application/json",
				"%s should return JSON content type", endpoint)
		}
	})
}

// TestAPICompatibilitySuite runs the test suite.
func TestAPICompatibilitySuite(t *testing.T) {
	suite.Run(t, new(APICompatibilitySuite))
}
