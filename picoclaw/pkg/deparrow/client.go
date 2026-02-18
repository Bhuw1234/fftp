package deparrow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the HTTP client for the DEparrow Meta-OS API.
// It handles authentication, request formatting, and response parsing.
type Client struct {
	// Base URL for the Meta-OS API (e.g., "http://localhost:8080")
	baseURL string
	// JWT token for authentication
	jwtToken string
	// HTTP client with configurable timeout
	httpClient *http.Client
	// User ID extracted from JWT (set after authentication)
	userID string
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new DEparrow API client.
// The jwtToken is required for authenticated endpoints.
//
// Example:
//
//	client := deparrow.NewClient("http://localhost:8080", "my-jwt-token")
func NewClient(apiURL, jwtToken string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:   apiURL,
		jwtToken:  jwtToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SetUserID sets the user ID for the client.
// This is typically extracted from the JWT token.
func (c *Client) SetUserID(userID string) {
	c.userID = userID
}

// doRequest performs an HTTP request with authentication.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.jwtToken)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if jsonErr := json.Unmarshal(respBody, &apiErr); jsonErr == nil {
			apiErr.Code = resp.StatusCode
			return &apiErr
		}
		return &APIError{
			Code:    resp.StatusCode,
			Message: string(respBody),
		}
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Health checks the API health status.
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/health", nil, &result)
	return result, err
}

// SubmitJob submits a new compute job to the DEparrow network.
// The job spec must include at minimum an image to run.
//
// Example:
//
//	job, err := client.SubmitJob(ctx, &deparrow.JobSpec{
//	    Image:   "ubuntu:latest",
//	    Command: []string{"echo", "hello world"},
//	    Resources: &deparrow.ResourceSpec{
//	        CPU:    "500m",
//	        Memory: "256Mi",
//	    },
//	})
func (c *Client) SubmitJob(ctx context.Context, spec *JobSpec) (*Job, error) {
	// Calculate credit cost based on resources
	creditCost := calculateCreditCost(spec)

	req := map[string]interface{}{
		"spec":         spec,
		"credit_cost":  creditCost,
	}

	var result struct {
		Status          string    `json:"status"`
		JobID           string    `json:"job_id"`
		CreditDeducted  float64   `json:"credit_deducted"`
		RemainingBalance float64  `json:"remaining_balance"`
		Message         string    `json:"message"`
	}

	err := c.doRequest(ctx, http.MethodPost, "/api/v1/jobs/submit", req, &result)
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:          result.JobID,
		Status:      JobStatusPending,
		Spec:        spec,
		CreditCost:  result.CreditDeducted,
		SubmittedAt: time.Now(),
	}, nil
}

// GetJob retrieves the status of a job by ID.
func (c *Client) GetJob(ctx context.Context, jobID string) (*Job, error) {
	var result struct {
		JobID       string     `json:"job_id"`
		Status      JobStatus  `json:"status"`
		UserID      string     `json:"user_id"`
		CreditCost  float64    `json:"credit_cost"`
		SubmittedAt time.Time  `json:"submitted_at"`
		Results     *JobResults `json:"results,omitempty"`
	}

	err := c.doRequest(ctx, http.MethodGet, "/api/v1/jobs/"+url.PathEscape(jobID), nil, &result)
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:          result.JobID,
		UserID:      result.UserID,
		Status:      result.Status,
		CreditCost:  result.CreditCost,
		SubmittedAt: result.SubmittedAt,
		Results:     result.Results,
	}, nil
}

// ListJobs lists all jobs for the authenticated user.
func (c *Client) ListJobs(ctx context.Context) ([]Job, error) {
	var result struct {
		Jobs []Job `json:"jobs"`
	}

	err := c.doRequest(ctx, http.MethodGet, "/api/v1/jobs", nil, &result)
	return result.Jobs, err
}

// CancelJob cancels a running job and returns partial credit refund.
func (c *Client) CancelJob(ctx context.Context, jobID string) (refund float64, err error) {
	var result struct {
		Status           string  `json:"status"`
		JobID            string  `json:"job_id"`
		RefundAmount     float64 `json:"refund_amount"`
		RemainingBalance float64 `json:"remaining_balance"`
	}

	err = c.doRequest(ctx, http.MethodPost, "/api/v1/jobs/"+url.PathEscape(jobID)+"/cancel", nil, &result)
	return result.RefundAmount, err
}

// GetCredits retrieves the current credit balance for the authenticated user.
func (c *Client) GetCredits(ctx context.Context) (*CreditBalance, error) {
	var result struct {
		UserID       string    `json:"user_id"`
		Balance      float64   `json:"credit_balance"`
		LastActive   time.Time `json:"last_active"`
	}

	path := "/api/v1/credits/balance/" + c.userID
	if c.userID == "" {
		path = "/api/v1/credits"
	}

	err := c.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &CreditBalance{
		Balance:     result.Balance,
		LastUpdated: result.LastActive,
	}, nil
}

// CheckCredits verifies if the user has sufficient credits for an operation.
func (c *Client) CheckCredits(ctx context.Context, required float64) (bool, error) {
	req := map[string]interface{}{
		"required": required,
	}

	var result struct {
		HasSufficient bool    `json:"has_sufficient"`
		Required      float64 `json:"required"`
		Available     float64 `json:"available"`
		Difference    float64 `json:"difference"`
	}

	err := c.doRequest(ctx, http.MethodPost, "/api/v1/credits/check", req, &result)
	return result.HasSufficient, err
}

// TransferCredits transfers credits to another user.
func (c *Client) TransferCredits(ctx context.Context, toUserID string, amount float64) error {
	req := map[string]interface{}{
		"to_user_id": toUserID,
		"amount":     amount,
	}

	return c.doRequest(ctx, http.MethodPost, "/api/v1/credits/transfer", req, nil)
}

// ListNodes retrieves all registered compute nodes.
func (c *Client) ListNodes(ctx context.Context) ([]Node, error) {
	var result struct {
		Nodes []struct {
			NodeID        string            `json:"node_id"`
			Arch          Architecture      `json:"arch"`
			Status        NodeStatus        `json:"status"`
			LastSeen      time.Time         `json:"last_seen"`
			Resources     *NodeResources    `json:"resources"`
			CreditsEarned float64           `json:"credits_earned"`
			Labels        map[string]string `json:"labels"`
		} `json:"nodes"`
		Total  int `json:"total"`
		Online int `json:"online"`
	}

	err := c.doRequest(ctx, http.MethodGet, "/api/v1/nodes", nil, &result)
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, len(result.Nodes))
	for i, n := range result.Nodes {
		nodes[i] = Node{
			ID:            n.NodeID,
			Arch:          n.Arch,
			Status:        n.Status,
			LastSeen:      n.LastSeen,
			Resources:     n.Resources,
			CreditsEarned: n.CreditsEarned,
			Labels:        n.Labels,
		}
	}

	return nodes, nil
}

// GetNode retrieves details for a specific node.
func (c *Client) GetNode(ctx context.Context, nodeID string) (*Node, error) {
	var result Node
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/nodes/"+url.PathEscape(nodeID), nil, &result)
	return &result, err
}

// GetNodeContribution retrieves contribution statistics for a specific node.
func (c *Client) GetNodeContribution(ctx context.Context, nodeID string) (*NodeContribution, error) {
	var result struct {
		NodeID       string              `json:"node_id"`
		Contribution *NodeContribution   `json:"contribution"`
		Ranking      struct {
			Rank       int              `json:"rank"`
			TotalNodes int              `json:"total_nodes"`
			Tier       ContributionTier `json:"tier"`
		} `json:"ranking"`
	}

	err := c.doRequest(ctx, http.MethodGet, "/api/v1/nodes/"+url.PathEscape(nodeID)+"/contribution", nil, &result)
	if err != nil {
		return nil, err
	}

	result.Contribution.Rank = result.Ranking.Rank
	result.Contribution.TotalNodes = result.Ranking.TotalNodes
	return result.Contribution, nil
}

// GetWallet retrieves the wallet information for the authenticated user.
func (c *Client) GetWallet(ctx context.Context) (*Wallet, error) {
	// The wallet endpoint returns credit balance
	credits, err := c.GetCredits(ctx)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Address: c.userID,
		Balance: credits.Balance,
	}, nil
}

// GetNetworkStats retrieves overall network statistics.
func (c *Client) GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	var result struct {
		Network struct {
			TotalNodes    int            `json:"total_nodes"`
			OnlineNodes   int            `json:"online_nodes"`
			TotalCPU      int            `json:"total_cpu_cores"`
			TotalGPU      int            `json:"total_gpu_count"`
			TotalMemory   float64        `json:"total_memory_gb"`
			LiveGFlops    float64        `json:"live_gflops"`
			LiveTFlops    float64        `json:"live_tflops"`
		} `json:"network"`
		Tiers     map[string]int `json:"tiers"`
		Timestamp time.Time      `json:"timestamp"`
	}

	err := c.doRequest(ctx, http.MethodGet, "/api/v1/network/contribution", nil, &result)
	if err != nil {
		return nil, err
	}

	return &NetworkStats{
		TotalNodes:      result.Network.TotalNodes,
		OnlineNodes:     result.Network.OnlineNodes,
		TotalCPU:        result.Network.TotalCPU,
		TotalGPU:        result.Network.TotalGPU,
		TotalMemory:     result.Network.TotalMemory,
		LiveGFlops:      result.Network.LiveGFlops,
		LiveTFlops:      result.Network.LiveTFlops,
		TierDistribution: result.Tiers,
		Timestamp:       result.Timestamp,
	}, nil
}

// GetLeaderboard retrieves the contribution leaderboard.
func (c *Client) GetLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	path := "/api/v1/network/leaderboard"
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}

	var result struct {
		Leaderboard []LeaderboardEntry `json:"leaderboard"`
	}

	err := c.doRequest(ctx, http.MethodGet, path, nil, &result)
	return result.Leaderboard, err
}

// GetMetrics retrieves system metrics.
func (c *Client) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/metrics", nil, &result)
	return result, err
}

// calculateCreditCost estimates the credit cost for a job based on resources.
func calculateCreditCost(spec *JobSpec) float64 {
	baseCost := 1.0 // Base cost per job

	if spec.Resources == nil {
		return baseCost
	}

	// Add cost based on resources
	// This is a simplified calculation; the actual implementation
	// should match the server-side pricing model

	// Memory cost: +0.1 per GB
	if spec.Resources.Memory != "" {
		// Parse memory string (simplified)
		baseCost += 0.1
	}

	// GPU cost: +2.0 per GPU
	if spec.Resources.GPU != "" && spec.Resources.GPU != "0" {
		baseCost += 2.0
	}

	// Timeout adjustment
	if spec.Timeout > 0 {
		baseCost *= float64(spec.Timeout) / 3600.0 // Normalize to hours
	}

	// Priority adjustment
	if spec.Priority > 50 {
		baseCost *= 1.5 // 50% extra for high priority
	}

	return baseCost
}
