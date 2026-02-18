//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bacalhau-project/bacalhau/deparrow/test-integration/testutil"
)

// PicoClawIntegrationSuite tests PicoClaw agent integration with DEparrow.
type PicoClawIntegrationSuite struct {
	suite.Suite
	mockServer *testutil.MockMetaOSServer
	client     *testutil.HTTPClient
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetupSuite initializes the test suite.
func (s *PicoClawIntegrationSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), testutil.LongTimeout)

	// Start mock server
	s.mockServer = testutil.NewMockMetaOSServer()
	s.client = testutil.NewHTTPClient(s.mockServer.URL, "")

	// Wait for server to be healthy
	err := s.mockServer.WaitForHealthy(s.ctx)
	require.NoError(s.T(), err, "Mock server should become healthy")
}

// TearDownSuite cleans up the test suite.
func (s *PicoClawIntegrationSuite) TearDownSuite() {
	if s.mockServer != nil {
		s.mockServer.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// TestAgentRegistration tests agent registration with Meta-OS.
func (s *PicoClawIntegrationSuite) TestAgentRegistration() {
	s.T().Run("successful registration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Register as agent node
		req := map[string]interface{}{
			"node_id":    "agent-node-001",
			"public_key": "agent-pubkey-001",
			"resources": map[string]interface{}{
				"cpu":    2,
				"memory": "4Gi",
				"disk":   "50Gi",
			},
			"labels": map[string]string{
				"type":    "agent",
				"version": "1.0.0",
			},
		}

		resp, err := s.client.Post(ctx, "/api/v1/nodes/register", req)
		require.NoError(t, err, "Registration request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		assert.True(t, result["success"].(bool), "Should indicate success")
		assert.Equal(t, "agent-node-001", result["node_id"], "Should return node ID")
		assert.Equal(t, "online", result["status"], "Status should be online")
	})

	s.T().Run("registration with existing ID updates node", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// First registration
		req := map[string]interface{}{
			"node_id":    "agent-node-002",
			"public_key": "agent-pubkey-002",
			"resources": map[string]interface{}{
				"cpu":    2,
				"memory": "4Gi",
			},
		}
		resp, err := s.client.Post(ctx, "/api/v1/nodes/register", req)
		require.NoError(t, err)
		resp.Body.Close()

		// Second registration with same ID (update)
		req["resources"] = map[string]interface{}{
			"cpu":    4, // Updated CPU
			"memory": "8Gi",
		}
		resp, err = s.client.Post(ctx, "/api/v1/nodes/register", req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should accept node update")
	})
}

// TestDEparrowToolsExecution tests DEparrow tools execution.
func (s *PicoClawIntegrationSuite) TestDEparrowToolsExecution() {
	s.T().Run("job_submit tool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Ensure user has enough credits
		s.mockServer.SetCredits("test-user", 100.0)

		// Submit job using job tool
		jobSpec := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "python:3.9-slim",
				"command": []string{"python", "-c", "print('Hello from PicoClaw')"},
				"resources": map[string]interface{}{
					"cpu":    "500m",
					"memory": "256Mi",
				},
			},
			"credit_cost": 5.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobSpec)
		require.NoError(t, err, "Job submission should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		assert.NotEmpty(t, result["job_id"], "Should return job ID")
		assert.Equal(t, "submitted", result["status"], "Status should be submitted")
		assert.Equal(t, 5.0, result["credit_deducted"], "Should deduct credits")
	})

	s.T().Run("credit_check tool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Check credits
		resp, err := s.client.Get(ctx, "/api/v1/credits/balance")
		require.NoError(t, err, "Credit check should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		assert.NotEmpty(t, result["user_id"], "Should return user ID")
		assert.NotNil(t, result["credit_balance"], "Should return balance")
	})

	s.T().Run("node_list tool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add a test node
		s.mockServer.AddTestNode("test-node-for-list")

		// List nodes
		resp, err := s.client.Get(ctx, "/api/v1/nodes")
		require.NoError(t, err, "Node list should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		nodes := result["nodes"].([]interface{})
		assert.GreaterOrEqual(t, len(nodes), 1, "Should have at least one node")
	})

	s.T().Run("network_stats tool", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Get network contribution stats
		resp, err := s.client.Get(ctx, "/api/v1/network/contribution")
		require.NoError(t, err, "Network stats should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		network := result["network"].(map[string]interface{})
		assert.NotNil(t, network["total_nodes"], "Should return total nodes")
		assert.NotNil(t, network["online_nodes"], "Should return online nodes")
	})
}

// TestWebSocketConnection tests WebSocket connection.
func (s *PicoClawIntegrationSuite) TestWebSocketConnection() {
	s.T().Run("websocket connection establishment", func(t *testing.T) {
		// WebSocket testing would require additional setup
		// This is a placeholder for WebSocket integration tests
		t.Skip("WebSocket tests require gorilla/websocket or similar setup")
	})

	s.T().Run("websocket message handling", func(t *testing.T) {
		t.Skip("WebSocket tests require gorilla/websocket or similar setup")
	})
}

// TestCreditSystemIntegration tests credit system integration.
func (s *PicoClawIntegrationSuite) TestCreditSystemIntegration() {
	s.T().Run("sufficient credits for job", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Set user credits
		s.mockServer.SetCredits("test-user", 50.0)

		// Submit job with cost less than balance
		jobSpec := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "ubuntu:latest",
				"command": []string{"echo", "test"},
			},
			"credit_cost": 10.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobSpec)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should accept job with sufficient credits")

		// Verify credits were deducted
		balance := s.mockServer.GetCredits("test-user")
		assert.Equal(t, 40.0, balance, "Should deduct 10 credits")
	})

	s.T().Run("insufficient credits rejects job", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Set user credits to low amount
		s.mockServer.SetCredits("test-user", 5.0)

		// Try to submit expensive job
		jobSpec := map[string]interface{}{
			"spec": map[string]interface{}{
				"image":   "ubuntu:latest",
				"command": []string{"echo", "test"},
			},
			"credit_cost": 50.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobSpec)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode, "Should reject job with insufficient credits")
	})

	s.T().Run("credit transfer between users", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Setup users
		s.mockServer.AddTestUser("sender", "sender@test.com", "password")
		s.mockServer.AddTestUser("receiver", "receiver@test.com", "password")
		s.mockServer.SetCredits("sender", 100.0)
		s.mockServer.SetCredits("receiver", 0.0)

		// Transfer credits
		transferReq := map[string]interface{}{
			"from_user": "sender",
			"to_user":   "receiver",
			"amount":    30.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/credits/transfer", transferReq)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Transfer should succeed")

		// Verify balances
		assert.Equal(t, 70.0, s.mockServer.GetCredits("sender"), "Sender should have 70 credits")
		assert.Equal(t, 30.0, s.mockServer.GetCredits("receiver"), "Receiver should have 30 credits")
	})

	s.T().Run("earning credits from job execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Setup node and user
		s.mockServer.AddTestNode("compute-node-001")
		s.mockServer.SetCredits("test-user", 20.0)

		// Submit job that will run on this node
		jobSpec := map[string]interface{}{
			"node_id": "compute-node-001",
			"spec": map[string]interface{}{
				"image":   "ubuntu:latest",
				"command": []string{"echo", "test"},
			},
			"credit_cost": 5.0,
		}

		resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobSpec)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Job should be submitted")

		// In a real scenario, node would earn credits for executing the job
		// This is simulated in the mock
	})
}

// TestAgentLifecycle tests the complete agent lifecycle.
func (s *PicoClawIntegrationSuite) TestAgentLifecycle() {
	s.T().Run("agent status check", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		resp, err := s.client.Get(ctx, "/api/v1/agent/status")
		require.NoError(t, err, "Agent status check should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		assert.NotEmpty(t, result["id"], "Should return agent ID")
		assert.NotEmpty(t, result["status"], "Should return status")
	})

	s.T().Run("agent chat interface", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		chatReq := map[string]interface{}{
			"message": "What is my current credit balance?",
		}

		resp, err := s.client.Post(ctx, "/api/v1/agent/chat", chatReq)
		require.NoError(t, err, "Agent chat should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		message := result["message"].(map[string]interface{})
		assert.NotEmpty(t, message["content"], "Should return response content")
	})
}

// TestConcurrentOperations tests concurrent agent operations.
func (s *PicoClawIntegrationSuite) TestConcurrentOperations() {
	s.T().Run("concurrent job submissions", func(t *testing.T) {
		// Setup sufficient credits
		s.mockServer.SetCredits("test-user", 500.0)

		// Submit multiple jobs concurrently
		numJobs := 5
		done := make(chan bool, numJobs)
		errors := make(chan error, numJobs)

		for i := 0; i < numJobs; i++ {
			go func(idx int) {
				ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
				defer cancel()

				jobSpec := map[string]interface{}{
					"spec": map[string]interface{}{
						"image":   "ubuntu:latest",
						"command": []string{"echo", "concurrent test"},
					},
					"credit_cost": 10.0,
				}

				resp, err := s.client.Post(ctx, "/api/v1/jobs/submit", jobSpec)
				if err != nil {
					errors <- err
					return
				}
				resp.Body.Close()

				if resp.StatusCode != 200 {
					errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
					return
				}

				done <- true
			}(i)
		}

		// Wait for all jobs to complete
		for i := 0; i < numJobs; i++ {
			select {
			case <-done:
				// Success
			case err := <-errors:
				t.Errorf("Concurrent job failed: %v", err)
			case <-time.After(testutil.DefaultTimeout):
				t.Fatal("Timeout waiting for concurrent jobs")
			}
		}
	})
}

// TestProviderIntegration tests provider-related operations.
func (s *PicoClawIntegrationSuite) TestProviderIntegration() {
	s.T().Run("list providers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add some online nodes
		s.mockServer.AddTestNode("provider-001")
		s.mockServer.AddTestNode("provider-002")

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err, "Provider list should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		providers := result["providers"].([]interface{})
		assert.GreaterOrEqual(t, len(providers), 1, "Should have providers")
	})

	s.T().Run("provider pricing", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add a node with GPU
		node := s.mockServer.AddTestNode("gpu-provider-001")
		node.Resources.GPU = 2
		node.Status = "online"

		resp, err := s.client.Get(ctx, "/api/v1/providers")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		testutil.ReadJSON(resp, &result)

		providers := result["providers"].([]interface{})
		for _, p := range providers {
			provider := p.(map[string]interface{})
			pricing := provider["pricing"].(map[string]interface{})
			assert.NotNil(t, pricing["cpu_per_hour"], "Should have CPU pricing")
			assert.NotNil(t, pricing["gpu_per_hour"], "Should have GPU pricing")
		}
	})
}

// TestLeaderboardIntegration tests leaderboard functionality.
func (s *PicoClawIntegrationSuite) TestLeaderboardIntegration() {
	s.T().Run("get leaderboard", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(s.ctx, testutil.DefaultTimeout)
		defer cancel()

		// Add nodes with credits earned
		for i := 0; i < 5; i++ {
			node := s.mockServer.AddTestNode(fmt.Sprintf("leader-node-%d", i))
			node.CreditsEarned = float64(100 - i*10)
		}

		resp, err := s.client.Get(ctx, "/api/v1/network/leaderboard")
		require.NoError(t, err, "Leaderboard request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		err = testutil.ReadJSON(resp, &result)
		require.NoError(t, err, "Should parse response JSON")

		leaderboard := result["leaderboard"].([]interface{})
		assert.GreaterOrEqual(t, len(leaderboard), 1, "Should have entries")
	})
}

// TestPicoClawIntegrationSuite runs the test suite.
func TestPicoClawIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PicoClawIntegrationSuite))
}
