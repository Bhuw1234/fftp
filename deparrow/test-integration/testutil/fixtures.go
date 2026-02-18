// Package testutil provides testing utilities for DEparrow integration tests.
package testutil

import (
	"context"
	"fmt"
	"time"
)

// Fixtures provides test data fixtures for integration tests.
var Fixtures = &fixtures{}

type fixtures struct{}

// TestUser represents a test user fixture.
type TestUser struct {
	ID       string
	Email    string
	Password string
	Name     string
	Token    string
}

// TestNode represents a test node fixture.
type TestNode struct {
	ID        string
	PublicKey string
	Arch      string
	Status    string
	Resources NodeResources
	Labels    map[string]string
}

// NodeResources represents node resource fixture.
type NodeResources struct {
	CPU    int
	Memory string
	GPU    int
	Disk   string
}

// TestJob represents a test job fixture.
type TestJob struct {
	Name        string
	Image       string
	Command     []string
	Env         map[string]string
	CPU         string
	Memory      string
	GPU         string
	Timeout     int
	Priority    int
	CreditCost  float64
}

// DefaultUser returns a default test user fixture.
func (f *fixtures) DefaultUser() TestUser {
	return TestUser{
		ID:       "test-user-001",
		Email:    "test@example.com",
		Password: "testPassword123!",
		Name:     "Test User",
		Token:    "test-jwt-token-001",
	}
}

// PremiumUser returns a premium test user with more credits.
func (f *fixtures) PremiumUser() TestUser {
	return TestUser{
		ID:       "premium-user-001",
		Email:    "premium@example.com",
		Password: "premiumPassword123!",
		Name:     "Premium User",
		Token:    "premium-jwt-token-001",
	}
}

// AdminUser returns an admin test user.
func (f *fixtures) AdminUser() TestUser {
	return TestUser{
		ID:       "admin-user-001",
		Email:    "admin@example.com",
		Password: "adminPassword123!",
		Name:     "Admin User",
		Token:    "admin-jwt-token-001",
	}
}

// MultipleUsers returns multiple test users.
func (f *fixtures) MultipleUsers(count int) []TestUser {
	users := make([]TestUser, count)
	for i := 0; i < count; i++ {
		users[i] = TestUser{
			ID:       fmt.Sprintf("user-%03d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Password: fmt.Sprintf("password%d", i+1),
			Name:     fmt.Sprintf("User %d", i+1),
			Token:    fmt.Sprintf("token-%03d", i+1),
		}
	}
	return users
}

// DefaultNode returns a default test node fixture.
func (f *fixtures) DefaultNode() TestNode {
	return TestNode{
		ID:        "node-001",
		PublicKey: "ssh-rsa-AAAAB3NzaC1yc2EAAAADAQABAAAB...",
		Arch:      "x86_64",
		Status:    "online",
		Resources: NodeResources{
			CPU:    4,
			Memory: "8Gi",
			GPU:    0,
			Disk:   "100Gi",
		},
		Labels: map[string]string{
			"region":   "us-west-2",
			"provider": "deparrow",
		},
	}
}

// GPUNode returns a test node with GPU.
func (f *fixtures) GPUNode() TestNode {
	return TestNode{
		ID:        "gpu-node-001",
		PublicKey: "ssh-rsa-AAAAB3NzaC1yc2EAAAADAQABAAAB...",
		Arch:      "x86_64",
		Status:    "online",
		Resources: NodeResources{
			CPU:    8,
			Memory: "32Gi",
			GPU:    2,
			Disk:   "500Gi",
		},
		Labels: map[string]string{
			"region":     "us-west-2",
			"provider":   "deparrow",
			"gpu":        "true",
			"gpu_model":  "NVIDIA A100",
		},
	}
}

// ARMNode returns an ARM architecture test node.
func (f *fixtures) ARMNode() TestNode {
	return TestNode{
		ID:        "arm-node-001",
		PublicKey: "ssh-rsa-AAAAB3NzaC1yc2EAAAADAQABAAAB...",
		Arch:      "arm64",
		Status:    "online",
		Resources: NodeResources{
			CPU:    4,
			Memory: "4Gi",
			GPU:    0,
			Disk:   "50Gi",
		},
		Labels: map[string]string{
			"region":   "eu-west-1",
			"provider": "deparrow",
			"arch":     "arm64",
		},
	}
}

// OfflineNode returns an offline test node.
func (f *fixtures) OfflineNode() TestNode {
	return TestNode{
		ID:        "offline-node-001",
		PublicKey: "ssh-rsa-AAAAB3NzaC1yc2EAAAADAQABAAAB...",
		Arch:      "x86_64",
		Status:    "offline",
		Resources: NodeResources{
			CPU:    4,
			Memory: "8Gi",
			GPU:    0,
			Disk:   "100Gi",
		},
		Labels: map[string]string{
			"region":   "us-east-1",
			"provider": "deparrow",
		},
	}
}

// MultipleNodes returns multiple test nodes.
func (f *fixtures) MultipleNodes(count int, onlineCount int) []TestNode {
	nodes := make([]TestNode, count)
	for i := 0; i < count; i++ {
		status := "offline"
		if i < onlineCount {
			status = "online"
		}
		nodes[i] = TestNode{
			ID:        fmt.Sprintf("node-%03d", i+1),
			PublicKey: fmt.Sprintf("ssh-rsa-key-%d", i+1),
			Arch:      "x86_64",
			Status:    status,
			Resources: NodeResources{
				CPU:    4,
				Memory: "8Gi",
				GPU:    0,
				Disk:   "100Gi",
			},
			Labels: map[string]string{
				"region": fmt.Sprintf("region-%d", i%3),
			},
		}
	}
	return nodes
}

// DefaultJob returns a default test job fixture.
func (f *fixtures) DefaultJob() TestJob {
	return TestJob{
		Name:    "test-job",
		Image:   "ubuntu:latest",
		Command: []string{"echo", "Hello, DEparrow!"},
		Env: map[string]string{
			"TEST_ENV": "test-value",
		},
		CPU:        "500m",
		Memory:     "256Mi",
		GPU:        "",
		Timeout:    300,
		Priority:   50,
		CreditCost: 1.0,
	}
}

// GPUJob returns a job that requires GPU.
func (f *fixtures) GPUJob() TestJob {
	return TestJob{
		Name:    "gpu-job",
		Image:   "nvidia/cuda:11.0-base",
		Command: []string{"nvidia-smi"},
		Env: map[string]string{
			"NVIDIA_VISIBLE_DEVICES": "all",
		},
		CPU:        "2000m",
		Memory:     "4Gi",
		GPU:        "1",
		Timeout:    600,
		Priority:   75,
		CreditCost: 5.0,
	}
}

// HighPriorityJob returns a high priority job.
func (f *fixtures) HighPriorityJob() TestJob {
	return TestJob{
		Name:    "high-priority-job",
		Image:   "python:3.9-slim",
		Command: []string{"python", "-c", "print('high priority')"},
		Env: map[string]string{
			"PRIORITY": "high",
		},
		CPU:        "1000m",
		Memory:     "512Mi",
		GPU:        "",
		Timeout:    180,
		Priority:   100,
		CreditCost: 3.0,
	}
}

// BatchJob returns a batch processing job.
func (f *fixtures) BatchJob() TestJob {
	return TestJob{
		Name:    "batch-job",
		Image:   "python:3.9-slim",
		Command: []string{"python", "process_batch.py"},
		Env: map[string]string{
			"BATCH_SIZE": "1000",
		},
		CPU:        "4000m",
		Memory:     "8Gi",
		GPU:        "",
		Timeout:    3600,
		Priority:   30,
		CreditCost: 10.0,
	}
}

// MultipleJobs returns multiple test jobs.
func (f *fixtures) MultipleJobs(count int) []TestJob {
	jobs := make([]TestJob, count)
	for i := 0; i < count; i++ {
		jobs[i] = TestJob{
			Name:    fmt.Sprintf("job-%d", i+1),
			Image:   "ubuntu:latest",
			Command: []string{"echo", fmt.Sprintf("Job %d", i+1)},
			Env: map[string]string{
				"JOB_ID": fmt.Sprintf("%d", i+1),
			},
			CPU:        "500m",
			Memory:     "256Mi",
			GPU:        "",
			Timeout:    300,
			Priority:   50,
			CreditCost: float64(i%5) + 1.0,
		}
	}
	return jobs
}

// WasmJob returns a WebAssembly job fixture.
func (f *fixtures) WasmJob() TestJob {
	return TestJob{
		Name:       "wasm-job",
		Image:      "",
		Command:    []string{"_start"},
		Env:        map[string]string{},
		CPU:        "100m",
		Memory:     "64Mi",
		GPU:        "",
		Timeout:    60,
		Priority:   50,
		CreditCost: 0.5,
	}
}

// TestScenario represents a complete test scenario.
type TestScenario struct {
	Name        string
	Description string
	Users       []TestUser
	Nodes       []TestNode
	Jobs        []TestJob
	SetupFunc   func(ctx context.Context, mock *MockMetaOSServer) error
	VerifyFunc  func(ctx context.Context, mock *MockMetaOSServer) error
}

// NodeJoinScenario returns a node join test scenario.
func (f *fixtures) NodeJoinScenario() TestScenario {
	return TestScenario{
		Name:        "node_join",
		Description: "Test node joining the network",
		Users:       []TestUser{f.DefaultUser()},
		Nodes:       []TestNode{f.DefaultNode()},
		SetupFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			mock.AddTestUser("test-user-001", "test@example.com", "testPassword123!")
			return nil
		},
		VerifyFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			// Verify node is registered
			return nil
		},
	}
}

// JobSubmissionScenario returns a job submission test scenario.
func (f *fixtures) JobSubmissionScenario() TestScenario {
	return TestScenario{
		Name:        "job_submission",
		Description: "Test job submission workflow",
		Users:       []TestUser{f.DefaultUser()},
		Nodes:       []TestNode{f.DefaultNode()},
		Jobs:        []TestJob{f.DefaultJob()},
		SetupFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			mock.AddTestUser("test-user-001", "test@example.com", "testPassword123!")
			mock.AddTestNode("node-001")
			mock.SetCredits("test-user-001", 100.0)
			return nil
		},
		VerifyFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			// Verify job was submitted
			return nil
		},
	}
}

// CreditTransferScenario returns a credit transfer test scenario.
func (f *fixtures) CreditTransferScenario() TestScenario {
	return TestScenario{
		Name:        "credit_transfer",
		Description: "Test credit transfer between users",
		Users:       f.MultipleUsers(2),
		SetupFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			mock.AddTestUser("user-001", "user1@example.com", "password1")
			mock.AddTestUser("user-002", "user2@example.com", "password2")
			mock.SetCredits("user-001", 100.0)
			mock.SetCredits("user-002", 0)
			return nil
		},
		VerifyFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			// Verify transfer completed
			return nil
		},
	}
}

// AgentLifecycleScenario returns an agent lifecycle test scenario.
func (f *fixtures) AgentLifecycleScenario() TestScenario {
	return TestScenario{
		Name:        "agent_lifecycle",
		Description: "Test agent start, stop, and status",
		Users:       []TestUser{f.DefaultUser()},
		Nodes:       []TestNode{f.DefaultNode()},
		SetupFunc: func(ctx context.Context, mock *MockMetaOSServer) error {
			mock.AddTestUser("test-user-001", "test@example.com", "testPassword123!")
			return nil
		},
	}
}

// AllScenarios returns all test scenarios.
func (f *fixtures) AllScenarios() []TestScenario {
	return []TestScenario{
		f.NodeJoinScenario(),
		f.JobSubmissionScenario(),
		f.CreditTransferScenario(),
		f.AgentLifecycleScenario(),
	}
}

// Timeout constants for tests.
const (
	DefaultTimeout  = 30 * time.Second
	ShortTimeout    = 5 * time.Second
	LongTimeout     = 2 * time.Minute
	ConnectTimeout  = 10 * time.Second
	RequestTimeout  = 15 * time.Second
)

// Credit constants for tests.
const (
	MinCreditBalance   = 10.0
	DefaultCreditCost  = 1.0
	GPUJobCreditCost   = 5.0
	HighPriorityFactor = 1.5
)
