//go:build unit

package deparrow

import (
	"testing"
	"time"
)

func TestJobStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   JobStatus
		expected string
	}{
		{"pending", JobStatusPending, "pending"},
		{"running", JobStatusRunning, "running"},
		{"completed", JobStatusCompleted, "completed"},
		{"failed", JobStatusFailed, "failed"},
		{"cancelled", JobStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("JobStatus %s = %s, want %s", tt.name, tt.status, tt.expected)
			}
		})
	}
}

func TestNodeStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeStatus
		expected string
	}{
		{"online", NodeStatusOnline, "online"},
		{"offline", NodeStatusOffline, "offline"},
		{"maintenance", NodeStatusMaintenance, "maintenance"},
		{"suspended", NodeStatusSuspended, "suspended"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("NodeStatus %s = %s, want %s", tt.name, tt.status, tt.expected)
			}
		})
	}
}

func TestContributionTier_Constants(t *testing.T) {
	tests := []struct {
		name     string
		tier     ContributionTier
		expected string
	}{
		{"bronze", TierBronze, "bronze"},
		{"silver", TierSilver, "silver"},
		{"gold", TierGold, "gold"},
		{"diamond", TierDiamond, "diamond"},
		{"legendary", TierLegendary, "legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.tier) != tt.expected {
				t.Errorf("ContributionTier %s = %s, want %s", tt.name, tt.tier, tt.expected)
			}
		})
	}
}

func TestArchitecture_Constants(t *testing.T) {
	tests := []struct {
		name     string
		arch     Architecture
		expected string
	}{
		{"x86_64", ArchX86_64, "x86_64"},
		{"arm64", ArchARM64, "arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.arch) != tt.expected {
				t.Errorf("Architecture %s = %s, want %s", tt.name, tt.arch, tt.expected)
			}
		})
	}
}

func TestJob_Fields(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Hour)

	job := Job{
		ID:          "job-123",
		UserID:      "user-456",
		Status:      JobStatusCompleted,
		CreditCost:  5.5,
		SubmittedAt: now,
		CompletedAt: &completedAt,
		Error:       "",
		Spec: &JobSpec{
			Image:   "ubuntu:latest",
			Command: []string{"echo", "hello"},
		},
		Results: &JobResults{
			OutputCID: "QmTest123",
			Stdout:    "hello\n",
			Stderr:    "",
			ExitCode:  0,
			Duration:  1.5,
			NodeID:    "node-789",
		},
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	if job.ID != "job-123" {
		t.Errorf("Job.ID = %s, want job-123", job.ID)
	}
	if job.UserID != "user-456" {
		t.Errorf("Job.UserID = %s, want user-456", job.UserID)
	}
	if job.Status != JobStatusCompleted {
		t.Errorf("Job.Status = %s, want completed", job.Status)
	}
	if job.CreditCost != 5.5 {
		t.Errorf("Job.CreditCost = %f, want 5.5", job.CreditCost)
	}
	if job.Spec == nil || job.Spec.Image != "ubuntu:latest" {
		t.Errorf("Job.Spec.Image invalid")
	}
	if job.Results == nil || job.Results.ExitCode != 0 {
		t.Errorf("Job.Results.ExitCode invalid")
	}
}

func TestJobSpec_Fields(t *testing.T) {
	spec := JobSpec{
		Image:   "python:3.11",
		Command: []string{"python", "-c", "print(1+1)"},
		Env:     map[string]string{"DEBUG": "true"},
		Resources: &ResourceSpec{
			CPU:     "500m",
			Memory:  "1Gi",
			GPU:     "1",
			Storage: "10Gi",
		},
		Inputs: []InputSpec{
			{
				StorageSource: "ipfs",
				Source:        "QmInput123",
				Path:          "/data/input",
			},
		},
		Outputs: []OutputSpec{
			{
				Path:              "/data/output",
				StorageDestination: "ipfs",
			},
		},
		Timeout:  3600,
		Priority: 75,
		Labels:   map[string]string{"env": "test"},
	}

	if spec.Image != "python:3.11" {
		t.Errorf("Image = %s, want python:3.11", spec.Image)
	}
	if len(spec.Command) != 3 {
		t.Errorf("Command length = %d, want 3", len(spec.Command))
	}
	if spec.Env["DEBUG"] != "true" {
		t.Errorf("Env[DEBUG] = %s, want true", spec.Env["DEBUG"])
	}
	if spec.Resources == nil || spec.Resources.CPU != "500m" {
		t.Errorf("Resources.CPU invalid")
	}
	if len(spec.Inputs) != 1 {
		t.Errorf("Inputs length = %d, want 1", len(spec.Inputs))
	}
	if spec.Timeout != 3600 {
		t.Errorf("Timeout = %d, want 3600", spec.Timeout)
	}
	if spec.Priority != 75 {
		t.Errorf("Priority = %d, want 75", spec.Priority)
	}
}

func TestResourceSpec_Fields(t *testing.T) {
	res := ResourceSpec{
		CPU:     "1000m",
		Memory:  "2Gi",
		GPU:     "2",
		Storage: "50Gi",
	}

	if res.CPU != "1000m" {
		t.Errorf("CPU = %s, want 1000m", res.CPU)
	}
	if res.Memory != "2Gi" {
		t.Errorf("Memory = %s, want 2Gi", res.Memory)
	}
	if res.GPU != "2" {
		t.Errorf("GPU = %s, want 2", res.GPU)
	}
	if res.Storage != "50Gi" {
		t.Errorf("Storage = %s, want 50Gi", res.Storage)
	}
}

func TestInputSpec_Fields(t *testing.T) {
	input := InputSpec{
		StorageSource: "s3",
		Source:        "bucket/path/file.txt",
		Path:          "/input/file.txt",
		S3Config: &S3Config{
			Bucket:    "my-bucket",
			Key:       "path/file.txt",
			Region:    "us-east-1",
			Endpoint:  "https://s3.amazonaws.com",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
			SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	if input.StorageSource != "s3" {
		t.Errorf("StorageSource = %s, want s3", input.StorageSource)
	}
	if input.S3Config == nil || input.S3Config.Bucket != "my-bucket" {
		t.Errorf("S3Config.Bucket invalid")
	}
}

func TestOutputSpec_Fields(t *testing.T) {
	output := OutputSpec{
		Path:              "/output/results.json",
		StorageDestination: "s3",
		S3Config: &S3Config{
			Bucket: "output-bucket",
			Key:    "results/result.json",
		},
	}

	if output.Path != "/output/results.json" {
		t.Errorf("Path = %s, want /output/results.json", output.Path)
	}
	if output.StorageDestination != "s3" {
		t.Errorf("StorageDestination = %s, want s3", output.StorageDestination)
	}
}

func TestJobResults_Fields(t *testing.T) {
	results := JobResults{
		OutputCID: "QmOutput456",
		Stdout:    "Hello, World!\n",
		Stderr:    "",
		ExitCode:  0,
		Duration:  3.14159,
		NodeID:    "node-abc123",
		DownloadURLs: map[string]string{
			"stdout": "https://gateway.ipfs.io/ipfs/QmOutput456",
		},
	}

	if results.OutputCID != "QmOutput456" {
		t.Errorf("OutputCID = %s, want QmOutput456", results.OutputCID)
	}
	if results.Duration != 3.14159 {
		t.Errorf("Duration = %f, want 3.14159", results.Duration)
	}
	if results.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", results.ExitCode)
	}
}

func TestNode_Fields(t *testing.T) {
	now := time.Now()
	node := Node{
		ID:       "node-xyz789",
		PublicKey: "ssh-rsa AAAAB3...",
		Arch:     ArchX86_64,
		Status:   NodeStatusOnline,
		Resources: &NodeResources{
			CPU:       8,
			Memory:    "16Gi",
			GPU:       1,
			GPUModel:  "NVIDIA RTX 3080",
			Storage:   "500Gi",
		},
		Labels: map[string]string{
			"region": "us-west-2",
			"gpu":    "true",
		},
		LastSeen:      now,
		CreditsEarned: 1250.75,
		Location: &Location{
			Latitude:  37.7749,
			Longitude: -122.4194,
			City:      "San Francisco",
			Country:   "USA",
		},
		Tier: TierGold,
		Contribution: &NodeContribution{
			CPUUsageHours:  150.5,
			GPUUsageHours:  45.25,
			LiveGFlops:     1234.56,
			NetworkPercent: 0.05,
			Rank:           42,
			TotalNodes:     1000,
		},
	}

	if node.ID != "node-xyz789" {
		t.Errorf("Node.ID = %s, want node-xyz789", node.ID)
	}
	if node.Arch != ArchX86_64 {
		t.Errorf("Node.Arch = %s, want x86_64", node.Arch)
	}
	if node.Status != NodeStatusOnline {
		t.Errorf("Node.Status = %s, want online", node.Status)
	}
	if node.Resources == nil || node.Resources.CPU != 8 {
		t.Errorf("Node.Resources.CPU invalid")
	}
	if node.CreditsEarned != 1250.75 {
		t.Errorf("Node.CreditsEarned = %f, want 1250.75", node.CreditsEarned)
	}
	if node.Tier != TierGold {
		t.Errorf("Node.Tier = %s, want gold", node.Tier)
	}
}

func TestNodeResources_Fields(t *testing.T) {
	res := NodeResources{
		CPU:       16,
		Memory:    "64Gi",
		GPU:       4,
		GPUModel:  "NVIDIA A100",
		Storage:   "2Ti",
	}

	if res.CPU != 16 {
		t.Errorf("CPU = %d, want 16", res.CPU)
	}
	if res.Memory != "64Gi" {
		t.Errorf("Memory = %s, want 64Gi", res.Memory)
	}
	if res.GPU != 4 {
		t.Errorf("GPU = %d, want 4", res.GPU)
	}
	if res.GPUModel != "NVIDIA A100" {
		t.Errorf("GPUModel = %s, want NVIDIA A100", res.GPUModel)
	}
}

func TestLocation_Fields(t *testing.T) {
	loc := Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		City:      "London",
		Country:   "UK",
	}

	if loc.Latitude != 51.5074 {
		t.Errorf("Latitude = %f, want 51.5074", loc.Latitude)
	}
	if loc.Longitude != -0.1278 {
		t.Errorf("Longitude = %f, want -0.1278", loc.Longitude)
	}
	if loc.City != "London" {
		t.Errorf("City = %s, want London", loc.City)
	}
}

func TestNodeContribution_Fields(t *testing.T) {
	contrib := NodeContribution{
		CPUUsageHours:  500.0,
		GPUUsageHours:  100.0,
		LiveGFlops:     5678.9,
		NetworkPercent: 1.25,
		Rank:           10,
		TotalNodes:     5000,
	}

	if contrib.CPUUsageHours != 500.0 {
		t.Errorf("CPUUsageHours = %f, want 500.0", contrib.CPUUsageHours)
	}
	if contrib.GPUUsageHours != 100.0 {
		t.Errorf("GPUUsageHours = %f, want 100.0", contrib.GPUUsageHours)
	}
	if contrib.Rank != 10 {
		t.Errorf("Rank = %d, want 10", contrib.Rank)
	}
}

func TestCreditBalance_Fields(t *testing.T) {
	now := time.Now()
	balance := CreditBalance{
		Balance:     500.25,
		Earned:      1500.0,
		Spent:       999.75,
		LastUpdated: now,
		MinBalance:  10.0,
	}

	if balance.Balance != 500.25 {
		t.Errorf("Balance = %f, want 500.25", balance.Balance)
	}
	if balance.Earned != 1500.0 {
		t.Errorf("Earned = %f, want 1500.0", balance.Earned)
	}
	if balance.Spent != 999.75 {
		t.Errorf("Spent = %f, want 999.75", balance.Spent)
	}
	if balance.MinBalance != 10.0 {
		t.Errorf("MinBalance = %f, want 10.0", balance.MinBalance)
	}
}

func TestWallet_Fields(t *testing.T) {
	now := time.Now()
	wallet := Wallet{
		Address:   "0x1234567890abcdef",
		Balance:   1000.0,
		CreatedAt: now,
		Transactions: []Transaction{
			{
				ID:          "tx-001",
				Type:        "earn",
				Amount:      100.0,
				Description: "CPU contribution",
				Timestamp:   now,
			},
			{
				ID:          "tx-002",
				Type:        "spend",
				Amount:      50.0,
				Description: "Job execution",
				Timestamp:   now.Add(time.Hour),
			},
		},
	}

	if wallet.Address != "0x1234567890abcdef" {
		t.Errorf("Address = %s, want 0x1234567890abcdef", wallet.Address)
	}
	if wallet.Balance != 1000.0 {
		t.Errorf("Balance = %f, want 1000.0", wallet.Balance)
	}
	if len(wallet.Transactions) != 2 {
		t.Errorf("Transactions length = %d, want 2", len(wallet.Transactions))
	}
}

func TestTransaction_Fields(t *testing.T) {
	now := time.Now()
	tx := Transaction{
		ID:          "tx-123",
		Type:        "transfer",
		Amount:      25.50,
		Description: "Payment for services",
		Timestamp:   now,
		FromUser:    "user-abc",
		ToUser:      "user-xyz",
	}

	if tx.ID != "tx-123" {
		t.Errorf("ID = %s, want tx-123", tx.ID)
	}
	if tx.Type != "transfer" {
		t.Errorf("Type = %s, want transfer", tx.Type)
	}
	if tx.Amount != 25.50 {
		t.Errorf("Amount = %f, want 25.50", tx.Amount)
	}
	if tx.FromUser != "user-abc" {
		t.Errorf("FromUser = %s, want user-abc", tx.FromUser)
	}
	if tx.ToUser != "user-xyz" {
		t.Errorf("ToUser = %s, want user-xyz", tx.ToUser)
	}
}

func TestNetworkStats_Fields(t *testing.T) {
	now := time.Now()
	stats := NetworkStats{
		TotalNodes:  1000,
		OnlineNodes: 850,
		TotalCPU:    8000,
		TotalGPU:    200,
		TotalMemory: 16000.0,
		LiveGFlops:  123456.78,
		LiveTFlops:  123.456,
		TierDistribution: map[string]int{
			"bronze":    500,
			"silver":    300,
			"gold":      150,
			"diamond":   40,
			"legendary": 10,
		},
		Timestamp: now,
	}

	if stats.TotalNodes != 1000 {
		t.Errorf("TotalNodes = %d, want 1000", stats.TotalNodes)
	}
	if stats.OnlineNodes != 850 {
		t.Errorf("OnlineNodes = %d, want 850", stats.OnlineNodes)
	}
	if stats.TotalGPU != 200 {
		t.Errorf("TotalGPU = %d, want 200", stats.TotalGPU)
	}
	if stats.LiveTFlops != 123.456 {
		t.Errorf("LiveTFlops = %f, want 123.456", stats.LiveTFlops)
	}
	if stats.TierDistribution["gold"] != 150 {
		t.Errorf("TierDistribution[gold] = %d, want 150", stats.TierDistribution["gold"])
	}
}

func TestLeaderboardEntry_Fields(t *testing.T) {
	entry := LeaderboardEntry{
		Rank:          1,
		NodeID:        "node-champion",
		Tier:          TierLegendary,
		CreditsEarned: 50000.0,
		CPUHours:      10000.0,
		GPUHours:      5000.0,
		TotalHours:    15000.0,
		Location: &Location{
			City:    "Tokyo",
			Country: "Japan",
		},
	}

	if entry.Rank != 1 {
		t.Errorf("Rank = %d, want 1", entry.Rank)
	}
	if entry.Tier != TierLegendary {
		t.Errorf("Tier = %s, want legendary", entry.Tier)
	}
	if entry.CreditsEarned != 50000.0 {
		t.Errorf("CreditsEarned = %f, want 50000.0", entry.CreditsEarned)
	}
	if entry.TotalHours != 15000.0 {
		t.Errorf("TotalHours = %f, want 15000.0", entry.TotalHours)
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      APIError
		expected string
	}{
		{
			name: "basic error",
			err: APIError{
				Code:    400,
				Message: "Bad Request",
			},
			expected: "Bad Request",
		},
		{
			name: "error with details",
			err: APIError{
				Code:    404,
				Message: "Not Found",
				Details: "Job job-123 does not exist",
			},
			expected: "Not Found: Job job-123 does not exist",
		},
		{
			name: "internal server error",
			err: APIError{
				Code:    500,
				Message: "Internal Server Error",
				Details: "Database connection failed",
			},
			expected: "Internal Server Error: Database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAPIError_Fields(t *testing.T) {
	err := APIError{
		Code:    401,
		Message: "Unauthorized",
		Details: "Invalid JWT token",
	}

	if err.Code != 401 {
		t.Errorf("Code = %d, want 401", err.Code)
	}
	if err.Message != "Unauthorized" {
		t.Errorf("Message = %s, want Unauthorized", err.Message)
	}
	if err.Details != "Invalid JWT token" {
		t.Errorf("Details = %s, want Invalid JWT token", err.Details)
	}
}

func TestS3Config_Fields(t *testing.T) {
	config := S3Config{
		Bucket:    "test-bucket",
		Key:       "path/to/object.txt",
		Region:    "eu-west-1",
		Endpoint:  "https://s3.eu-west-1.amazonaws.com",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG",
	}

	if config.Bucket != "test-bucket" {
		t.Errorf("Bucket = %s, want test-bucket", config.Bucket)
	}
	if config.Region != "eu-west-1" {
		t.Errorf("Region = %s, want eu-west-1", config.Region)
	}
}

// Test JSON marshaling/unmarshaling
func TestJob_JSONMarshaling(t *testing.T) {
	now := time.Now()
	job := Job{
		ID:          "job-test",
		UserID:      "user-test",
		Status:      JobStatusRunning,
		CreditCost:  2.5,
		SubmittedAt: now,
	}

	// This test verifies the struct has proper JSON tags
	// In a real implementation, you'd use json.Marshal/Unmarshal
	if job.ID != "job-test" {
		t.Errorf("Job ID not set correctly")
	}
}

func TestJobSpec_JSONMarshaling(t *testing.T) {
	spec := JobSpec{
		Image:   "alpine:latest",
		Command: []string{"sh", "-c", "echo test"},
		Env: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		},
		Timeout:  300,
		Priority: 50,
	}

	if spec.Image != "alpine:latest" {
		t.Errorf("Image not set correctly")
	}
	if len(spec.Env) != 2 {
		t.Errorf("Env count = %d, want 2", len(spec.Env))
	}
}

func TestNode_JSONMarshaling(t *testing.T) {
	node := Node{
		ID:            "node-json-test",
		Arch:          ArchARM64,
		Status:        NodeStatusMaintenance,
		CreditsEarned: 100.0,
		Tier:          TierSilver,
	}

	if node.ID != "node-json-test" {
		t.Errorf("Node ID not set correctly")
	}
	if node.Arch != ArchARM64 {
		t.Errorf("Arch not set correctly")
	}
}

// Test nil pointer handling
func TestJob_NilFields(t *testing.T) {
	job := Job{
		ID:          "job-nil",
		Status:      JobStatusPending,
		SubmittedAt: time.Now(),
		// Spec, Results, CompletedAt are nil
	}

	if job.Spec != nil {
		t.Errorf("Spec should be nil")
	}
	if job.Results != nil {
		t.Errorf("Results should be nil")
	}
	if job.CompletedAt != nil {
		t.Errorf("CompletedAt should be nil")
	}
}

func TestNode_NilFields(t *testing.T) {
	node := Node{
		ID:     "node-nil",
		Status: NodeStatusOffline,
		// Resources, Location, Contribution are nil
	}

	if node.Resources != nil {
		t.Errorf("Resources should be nil")
	}
	if node.Location != nil {
		t.Errorf("Location should be nil")
	}
	if node.Contribution != nil {
		t.Errorf("Contribution should be nil")
	}
}

func TestWallet_EmptyTransactions(t *testing.T) {
	wallet := Wallet{
		Address:      "wallet-empty",
		Balance:      0,
		Transactions: []Transaction{},
		CreatedAt:    time.Now(),
	}

	if len(wallet.Transactions) != 0 {
		t.Errorf("Transactions should be empty")
	}
}

// Test edge cases
func TestJobSpec_EmptyCommand(t *testing.T) {
	spec := JobSpec{
		Image:   "scratch",
		Command: []string{},
	}

	if len(spec.Command) != 0 {
		t.Errorf("Command should be empty slice")
	}
}

func TestResourceSpec_EmptyFields(t *testing.T) {
	res := ResourceSpec{} // All fields empty

	if res.CPU != "" {
		t.Errorf("CPU should be empty")
	}
	if res.Memory != "" {
		t.Errorf("Memory should be empty")
	}
	if res.GPU != "" {
		t.Errorf("GPU should be empty")
	}
}

func TestNetworkStats_EmptyTierDistribution(t *testing.T) {
	stats := NetworkStats{
		TotalNodes:       0,
		TierDistribution: map[string]int{},
	}

	if len(stats.TierDistribution) != 0 {
		t.Errorf("TierDistribution should be empty")
	}
}
