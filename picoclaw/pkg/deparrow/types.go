// Package deparrow provides DEparrow network integration for PicoClaw agents.
// It enables AI agents to interact with the DEparrow compute marketplace,
// submit jobs, check credits, and manage their wallet.
package deparrow

import "time"

// JobStatus represents the current state of a compute job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// NodeStatus represents the current state of a compute node.
type NodeStatus string

const (
	NodeStatusOnline     NodeStatus = "online"
	NodeStatusOffline    NodeStatus = "offline"
	NodeStatusMaintenance NodeStatus = "maintenance"
	NodeStatusSuspended  NodeStatus = "suspended"
)

// ContributionTier represents a node's contribution level.
type ContributionTier string

const (
	TierBronze   ContributionTier = "bronze"
	TierSilver   ContributionTier = "silver"
	TierGold     ContributionTier = "gold"
	TierDiamond  ContributionTier = "diamond"
	TierLegendary ContributionTier = "legendary"
)

// Architecture represents CPU architecture.
type Architecture string

const (
	ArchX86_64 Architecture = "x86_64"
	ArchARM64  Architecture = "arm64"
)

// Job represents a compute job on the DEparrow network.
type Job struct {
	ID           string                 `json:"job_id"`
	UserID       string                 `json:"user_id"`
	Status       JobStatus              `json:"status"`
	Spec         *JobSpec               `json:"spec,omitempty"`
	Results      *JobResults            `json:"results,omitempty"`
	CreditCost   float64                `json:"credit_cost"`
	SubmittedAt  time.Time              `json:"submitted_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// JobSpec defines the specification for a compute job.
type JobSpec struct {
	// Docker image to run
	Image string `json:"image"`
	// Command to execute inside the container
	Command []string `json:"command,omitempty"`
	// Environment variables
	Env map[string]string `json:"env,omitempty"`
	// Resource requirements
	Resources *ResourceSpec `json:"resources,omitempty"`
	// Input data sources
	Inputs []InputSpec `json:"inputs,omitempty"`
	// Output specifications
	Outputs []OutputSpec `json:"outputs,omitempty"`
	// Timeout in seconds (0 = default)
	Timeout int `json:"timeout,omitempty"`
	// Priority level (0-100, higher = more priority)
	Priority int `json:"priority,omitempty"`
	// Labels for job categorization
	Labels map[string]string `json:"labels,omitempty"`
}

// ResourceSpec defines resource requirements for a job.
type ResourceSpec struct {
	CPU     string `json:"cpu,omitempty"`     // e.g., "500m" for 0.5 cores
	Memory  string `json:"memory,omitempty"`  // e.g., "1Gi"
	GPU     string `json:"gpu,omitempty"`     // e.g., "1" for 1 GPU
	Storage string `json:"storage,omitempty"` // e.g., "10Gi"
}

// InputSpec defines an input data source.
type InputSpec struct {
	// Storage source type: "ipfs", "s3", "url", "inline"
	StorageSource string `json:"storage_source"`
	// Source URL/IPFS CID/S3 path
	Source string `json:"source"`
	// Path to mount in the container
	Path string `json:"path"`
	// Optional: S3 configuration
	S3Config *S3Config `json:"s3_config,omitempty"`
}

// OutputSpec defines an output specification.
type OutputSpec struct {
	// Path inside container to capture
	Path string `json:"path"`
	// Storage destination: "ipfs", "s3"
	StorageDestination string `json:"storage_destination"`
	// Optional: S3 configuration for output
	S3Config *S3Config `json:"s3_config,omitempty"`
}

// S3Config holds S3 storage configuration.
type S3Config struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Region    string `json:"region,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

// JobResults contains the results of a completed job.
type JobResults struct {
	// Output CID (IPFS)
	OutputCID string `json:"output_cid,omitempty"`
	// Stdout from the job
	Stdout string `json:"stdout,omitempty"`
	// Stderr from the job
	Stderr string `json:"stderr,omitempty"`
	// Exit code
	ExitCode int `json:"exit_code"`
	// Duration in seconds
	Duration float64 `json:"duration_seconds"`
	// Node that executed the job
	NodeID string `json:"node_id,omitempty"`
	// Download URLs for outputs
	DownloadURLs map[string]string `json:"download_urls,omitempty"`
}

// Node represents a compute node on the DEparrow network.
type Node struct {
	ID             string                 `json:"node_id"`
	PublicKey      string                 `json:"public_key,omitempty"`
	Arch           Architecture           `json:"arch"`
	Status         NodeStatus             `json:"status"`
	Resources      *NodeResources         `json:"resources,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
	LastSeen       time.Time              `json:"last_seen"`
	CreditsEarned  float64                `json:"credits_earned"`
	Location       *Location              `json:"location,omitempty"`
	Tier           ContributionTier       `json:"tier"`
	Contribution   *NodeContribution      `json:"contribution,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NodeResources describes a node's available resources.
type NodeResources struct {
	CPU     int    `json:"cpu_cores"`
	Memory  string `json:"memory"`
	GPU     int    `json:"gpu_count"`
	GPUModel string `json:"gpu_model,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// Location represents geographical location.
type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	City      string  `json:"city,omitempty"`
	Country   string  `json:"country,omitempty"`
}

// NodeContribution contains node contribution statistics.
type NodeContribution struct {
	CPUUsageHours   float64 `json:"cpu_usage_hours"`
	GPUUsageHours   float64 `json:"gpu_usage_hours"`
	LiveGFlops      float64 `json:"live_gflops"`
	NetworkPercent  float64 `json:"network_percent"`
	Rank            int     `json:"rank"`
	TotalNodes      int     `json:"total_nodes"`
}

// CreditBalance represents a user's credit information.
type CreditBalance struct {
	Balance     float64 `json:"balance"`
	Earned      float64 `json:"total_earned"`
	Spent       float64 `json:"total_spent"`
	LastUpdated time.Time `json:"last_updated"`
	// Minimum balance required for job submission
	MinBalance float64 `json:"min_balance,omitempty"`
}

// Wallet represents a DEparrow wallet with transaction history.
type Wallet struct {
	Address     string        `json:"address"`
	Balance     float64       `json:"balance"`
	Transactions []Transaction `json:"transactions,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Transaction represents a credit transaction.
type Transaction struct {
	ID          string    `json:"transaction_id"`
	Type        string    `json:"type"` // "earn", "spend", "transfer"
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	// For transfers
	FromUser string `json:"from_user,omitempty"`
	ToUser   string `json:"to_user,omitempty"`
}

// NetworkStats represents overall network statistics.
type NetworkStats struct {
	TotalNodes     int            `json:"total_nodes"`
	OnlineNodes    int            `json:"online_nodes"`
	TotalCPU       int            `json:"total_cpu_cores"`
	TotalGPU       int            `json:"total_gpu_count"`
	TotalMemory    float64        `json:"total_memory_gb"`
	LiveGFlops     float64        `json:"live_gflops"`
	LiveTFlops     float64        `json:"live_tflops"`
	TierDistribution map[string]int `json:"tiers"`
	Timestamp      time.Time      `json:"timestamp"`
}

// LeaderboardEntry represents a node's leaderboard position.
type LeaderboardEntry struct {
	Rank            int              `json:"rank"`
	NodeID          string           `json:"node_id"`
	Tier            ContributionTier `json:"tier"`
	CreditsEarned   float64          `json:"credits_earned"`
	CPUHours        float64          `json:"cpu_usage_hours"`
	GPUHours        float64          `json:"gpu_usage_hours"`
	TotalHours      float64          `json:"total_hours"`
	Location        *Location        `json:"location,omitempty"`
}

// APIError represents an error response from the API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}
