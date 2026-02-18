// Package testutil provides testing utilities for DEparrow integration tests.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockMetaOSServer provides a mock Meta-OS server for testing.
type MockMetaOSServer struct {
	Server      *httptest.Server
	URL         string
	JWTSecret   string
	mu          sync.RWMutex
	nodes       map[string]*MockNode
	jobs        map[string]*MockJob
	users       map[string]*MockUser
	credits     map[string]float64
	transactions []*MockTransaction
}

// MockNode represents a mock compute node.
type MockNode struct {
	ID            string            `json:"node_id"`
	PublicKey     string            `json:"public_key"`
	Arch          string            `json:"arch"`
	Status        string            `json:"status"`
	LastSeen      time.Time         `json:"last_seen"`
	Resources     *MockResources    `json:"resources"`
	CreditsEarned float64           `json:"credits_earned"`
	Labels        map[string]string `json:"labels"`
}

// MockResources represents node resources.
type MockResources struct {
	CPU    int    `json:"cpu_cores"`
	Memory string `json:"memory"`
	GPU    int    `json:"gpu_count"`
	Disk   string `json:"disk"`
}

// MockJob represents a mock compute job.
type MockJob struct {
	ID          string                 `json:"job_id"`
	UserID      string                 `json:"user_id"`
	NodeID      string                 `json:"node_id,omitempty"`
	Status      string                 `json:"status"`
	Spec        map[string]interface{} `json:"spec"`
	CreditCost  float64                `json:"credit_cost"`
	SubmittedAt time.Time              `json:"submitted_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// MockUser represents a mock user.
type MockUser struct {
	ID       string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"-"`
	Token    string `json:"token"`
}

// MockTransaction represents a credit transaction.
type MockTransaction struct {
	ID          string    `json:"transaction_id"`
	Type        string    `json:"type"`
	FromUser    string    `json:"from_user,omitempty"`
	ToUser      string    `json:"to_user,omitempty"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewMockMetaOSServer creates a new mock Meta-OS server.
func NewMockMetaOSServer() *MockMetaOSServer {
	mock := &MockMetaOSServer{
		JWTSecret:   "test-jwt-secret-key",
		nodes:       make(map[string]*MockNode),
		jobs:        make(map[string]*MockJob),
		users:       make(map[string]*MockUser),
		credits:     make(map[string]float64),
		transactions: make([]*MockTransaction, 0),
	}

	// Create test server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
	mock.URL = mock.Server.URL

	// Add default test user
	mock.AddTestUser("test-user", "test@example.com", "password123")
	mock.credits["test-user"] = 1000.0 // Give test user 1000 credits

	return mock
}

// Close closes the mock server.
func (m *MockMetaOSServer) Close() {
	m.Server.Close()
}

// AddTestUser adds a test user to the mock server.
func (m *MockMetaOSServer) AddTestUser(id, email, password string) *MockUser {
	m.mu.Lock()
	defer m.mu.Unlock()

	user := &MockUser{
		ID:       id,
		Email:    email,
		Name:     fmt.Sprintf("Test User %s", id),
		Password: password,
		Token:    fmt.Sprintf("test-jwt-token-%s", id),
	}
	m.users[id] = user
	m.credits[id] = 0
	return user
}

// AddTestNode adds a test node to the mock server.
func (m *MockMetaOSServer) AddTestNode(id string) *MockNode {
	m.mu.Lock()
	defer m.mu.Unlock()

	node := &MockNode{
		ID:        id,
		PublicKey: fmt.Sprintf("pubkey-%s", id),
		Arch:      "x86_64",
		Status:    "online",
		LastSeen:  time.Now(),
		Resources: &MockResources{
			CPU:    4,
			Memory: "8Gi",
			GPU:    0,
			Disk:   "100Gi",
		},
		CreditsEarned: 0,
		Labels:        make(map[string]string),
	}
	m.nodes[id] = node
	return node
}

// GetCredits returns the credit balance for a user.
func (m *MockMetaOSServer) GetCredits(userID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.credits[userID]
}

// SetCredits sets the credit balance for a user.
func (m *MockMetaOSServer) SetCredits(userID string, amount float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.credits[userID] = amount
}

// handleRequest handles incoming HTTP requests.
func (m *MockMetaOSServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Route requests
	switch {
	case r.URL.Path == "/api/v1/health":
		m.handleHealth(w, r)
	case r.URL.Path == "/api/v1/auth/login":
		m.handleLogin(w, r)
	case r.URL.Path == "/api/v1/auth/register":
		m.handleRegister(w, r)
	case r.URL.Path == "/api/v1/nodes/register":
		m.handleNodeRegister(w, r)
	case r.URL.Path == "/api/v1/nodes":
		m.handleListNodes(w, r)
	case r.URL.Path == "/api/v1/jobs/submit":
		m.handleJobSubmit(w, r)
	case r.URL.Path == "/api/v1/jobs":
		m.handleListJobs(w, r)
	case r.URL.Path == "/api/v1/credits/balance":
		m.handleCreditBalance(w, r)
	case r.URL.Path == "/api/v1/credits/transfer":
		m.handleCreditTransfer(w, r)
	case r.URL.Path == "/api/v1/agent/status":
		m.handleAgentStatus(w, r)
	case r.URL.Path == "/api/v1/agent/chat":
		m.handleAgentChat(w, r)
	case r.URL.Path == "/api/v1/providers":
		m.handleListProviders(w, r)
	case r.URL.Path == "/api/v1/network/contribution":
		m.handleNetworkContribution(w, r)
	case r.URL.Path == "/api/v1/network/leaderboard":
		m.handleLeaderboard(w, r)
	default:
		m.handleNotFound(w, r)
	}
}

func (m *MockMetaOSServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0-test",
		"services": map[string]string{
			"bootstrap": "healthy",
			"registry":  "healthy",
			"credits":   "healthy",
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == req.Email && user.Password == req.Password {
			response := map[string]interface{}{
				"token": user.Token,
				"user": map[string]interface{}{
					"id":       user.ID,
					"email":    user.Email,
					"name":     user.Name,
					"credits":  m.credits[user.ID],
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
}

func (m *MockMetaOSServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()
	user := m.AddTestUser(userID, req.Email, req.Password)
	user.Name = req.Name

	response := map[string]interface{}{
		"token": user.Token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.Name,
			"credits":  0,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleNodeRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID    string                 `json:"node_id"`
		PublicKey string                 `json:"public_key"`
		Resources map[string]interface{} `json:"resources"`
		Labels    map[string]string      `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create or update node
	node, exists := m.nodes[req.NodeID]
	if !exists {
		node = &MockNode{
			ID:        req.NodeID,
			PublicKey: req.PublicKey,
			Arch:      "x86_64",
			Status:    "online",
			LastSeen:  time.Now(),
			Resources: &MockResources{
				CPU:    4,
				Memory: "8Gi",
				GPU:    0,
				Disk:   "100Gi",
			},
			Labels:    make(map[string]string),
		}
		m.nodes[req.NodeID] = node
	}

	// Update resources
	if cpu, ok := req.Resources["cpu"].(float64); ok {
		node.Resources.CPU = int(cpu)
	}
	if mem, ok := req.Resources["memory"].(string); ok {
		node.Resources.Memory = mem
	}
	if disk, ok := req.Resources["disk"].(string); ok {
		node.Resources.Disk = disk
	}

	// Update labels
	for k, v := range req.Labels {
		node.Labels[k] = v
	}

	response := map[string]interface{}{
		"success":  true,
		"node_id":  node.ID,
		"status":   node.Status,
		"message":  "Node registered successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleListNodes(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]map[string]interface{}, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, map[string]interface{}{
			"node_id":        node.ID,
			"public_key":     node.PublicKey,
			"arch":           node.Arch,
			"status":         node.Status,
			"last_seen":      node.LastSeen,
			"resources":      node.Resources,
			"credits_earned": node.CreditsEarned,
			"labels":         node.Labels,
		})
	}

	response := map[string]interface{}{
		"nodes": nodes,
		"total": len(nodes),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleJobSubmit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID     string                 `json:"node_id"`
		Spec       map[string]interface{} `json:"spec"`
		Credits    float64                `json:"credits"`
		CreditCost float64                `json:"credit_cost"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Get user ID from auth header (simplified)
	userID := "test-user"

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check credits
	creditCost := req.CreditCost
	if creditCost == 0 {
		creditCost = 10.0 // Default cost
	}

	if m.credits[userID] < creditCost {
		http.Error(w, `{"error": "Insufficient credits"}`, http.StatusBadRequest)
		return
	}

	// Deduct credits
	m.credits[userID] -= creditCost

	// Create job
	jobID := fmt.Sprintf("job-%s", uuid.New().String()[:8])
	job := &MockJob{
		ID:          jobID,
		UserID:      userID,
		NodeID:      req.NodeID,
		Status:      "pending",
		Spec:        req.Spec,
		CreditCost:  creditCost,
		SubmittedAt: time.Now(),
	}
	m.jobs[jobID] = job

	// Add transaction
	m.transactions = append(m.transactions, &MockTransaction{
		ID:          fmt.Sprintf("txn-%s", uuid.New().String()[:8]),
		Type:        "spend",
		FromUser:    userID,
		Amount:      creditCost,
		Description: fmt.Sprintf("Job submission: %s", jobID),
		Timestamp:   time.Now(),
	})

	response := map[string]interface{}{
		"job_id":            jobID,
		"status":            "submitted",
		"credit_deducted":   creditCost,
		"remaining_balance": m.credits[userID],
		"message":           "Job submitted successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleListJobs(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]map[string]interface{}, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, map[string]interface{}{
			"job_id":       job.ID,
			"user_id":      job.UserID,
			"node_id":      job.NodeID,
			"status":       job.Status,
			"credit_cost":  job.CreditCost,
			"submitted_at": job.SubmittedAt,
		})
	}

	response := map[string]interface{}{
		"jobs": jobs,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleCreditBalance(w http.ResponseWriter, r *http.Request) {
	// Get user ID from auth header (simplified)
	userID := "test-user"

	m.mu.RLock()
	defer m.mu.RUnlock()

	response := map[string]interface{}{
		"user_id":       userID,
		"credit_balance": m.credits[userID],
		"last_active":    time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleCreditTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromUser string  `json:"from_user"`
		ToUser   string  `json:"to_user"`
		Amount   float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check sender has enough credits
	if m.credits[req.FromUser] < req.Amount {
		http.Error(w, `{"error": "Insufficient credits"}`, http.StatusBadRequest)
		return
	}

	// Ensure recipient exists
	if _, exists := m.credits[req.ToUser]; !exists {
		m.credits[req.ToUser] = 0
	}

	// Transfer credits
	m.credits[req.FromUser] -= req.Amount
	m.credits[req.ToUser] += req.Amount

	// Record transaction
	txnID := fmt.Sprintf("txn-%s", uuid.New().String()[:8])
	m.transactions = append(m.transactions, &MockTransaction{
		ID:        txnID,
		Type:      "transfer",
		FromUser:  req.FromUser,
		ToUser:    req.ToUser,
		Amount:    req.Amount,
		Timestamp: time.Now(),
	})

	response := map[string]interface{}{
		"success":        true,
		"transaction_id": txnID,
		"from_user":      req.FromUser,
		"to_user":        req.ToUser,
		"amount":         req.Amount,
		"message":        "Transfer successful",
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":              "agent-test-001",
		"name":            "Test Agent",
		"status":          "running",
		"uptime":          "1h30m",
		"credits_earned":  150.5,
		"credits_spent":   45.0,
		"tasks_completed": 12,
		"tools": []map[string]interface{}{
			{
				"name":        "job_submit",
				"description": "Submit compute jobs",
				"enabled":     true,
				"calls":       8,
			},
			{
				"name":        "credit_check",
				"description": "Check credit balance",
				"enabled":     true,
				"calls":       15,
			},
		},
		"resources": map[string]interface{}{
			"cpu_usage":    25.5,
			"memory_usage": 512.0,
			"disk_usage":   2.1,
		},
		"last_heartbeat": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleAgentChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Generate a mock response
	response := map[string]interface{}{
		"message": map[string]interface{}{
			"id":        fmt.Sprintf("msg-%s", uuid.New().String()[:8]),
			"role":      "assistant",
			"content":   fmt.Sprintf("I received your message: '%s'. How can I help you with the DEparrow network?", req.Message),
			"timestamp": time.Now(),
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleListProviders(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]map[string]interface{}, 0, len(m.nodes))
	for _, node := range m.nodes {
		if node.Status == "online" && node.Resources != nil {
			providerName := node.ID
			if len(node.ID) > 8 {
				providerName = node.ID[:8]
			}
			providers = append(providers, map[string]interface{}{
				"id":     node.ID,
				"name":   fmt.Sprintf("Provider %s", providerName),
				"status": node.Status,
				"location": map[string]interface{}{
					"region":  "us-west-2",
					"country": "USA",
					"city":    "San Francisco",
				},
				"resources": map[string]interface{}{
					"cpu_cores":     node.Resources.CPU,
					"cpu_available": node.Resources.CPU,
					"memory_total":  node.Resources.Memory,
					"memory_available": node.Resources.Memory,
					"gpu_count":     node.Resources.GPU,
					"gpu_available": node.Resources.GPU,
				},
				"pricing": map[string]interface{}{
					"cpu_per_hour":      1.0,
					"memory_per_gb_hour": 0.1,
					"gpu_per_hour":       5.0,
				},
				"stats": map[string]interface{}{
					"jobs_completed":       50,
					"success_rate":         0.98,
					"avg_response_time":    2.5,
					"uptime_30d":           0.99,
					"total_credits_earned": node.CreditsEarned,
				},
			})
		}
	}

	response := map[string]interface{}{
		"providers": providers,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleNetworkContribution(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalCPU := 0
	totalGPU := 0
	onlineNodes := 0

	for _, node := range m.nodes {
		if node.Status == "online" {
			onlineNodes++
			totalCPU += node.Resources.CPU
			totalGPU += node.Resources.GPU
		}
	}

	response := map[string]interface{}{
		"network": map[string]interface{}{
			"total_nodes":    len(m.nodes),
			"online_nodes":   onlineNodes,
			"total_cpu_cores": totalCPU,
			"total_gpu_count": totalGPU,
			"total_memory_gb": float64(len(m.nodes)) * 8.0,
			"live_gflops":     float64(totalCPU) * 50.0,
			"live_tflops":     float64(totalGPU) * 15.0,
		},
		"tiers": map[string]int{
			"bronze":   onlineNodes / 2,
			"silver":   onlineNodes / 4,
			"gold":     onlineNodes / 8,
			"diamond":  1,
		},
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	leaderboard := make([]map[string]interface{}, 0, len(m.nodes))
	rank := 1
	for _, node := range m.nodes {
		leaderboard = append(leaderboard, map[string]interface{}{
			"rank":             rank,
			"node_id":          node.ID,
			"tier":             "silver",
			"credits_earned":   node.CreditsEarned,
			"cpu_usage_hours":  node.CreditsEarned / 10.0,
			"gpu_usage_hours":  node.CreditsEarned / 50.0,
			"total_hours":      node.CreditsEarned / 10.0,
		})
		rank++
	}

	response := map[string]interface{}{
		"leaderboard": leaderboard,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *MockMetaOSServer) handleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error": "Not found"}`, http.StatusNotFound)
}

// WaitForHealthy waits for the server to be healthy.
func (m *MockMetaOSServer) WaitForHealthy(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := http.Get(m.URL + "/api/v1/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
