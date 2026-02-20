//go:build unit

package globalvm

import (
	"context"
	"fmt"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator"
	"github.com/rs/zerolog/log"
)

// GlobalJobRequest represents a job submission request to the Global VM.
// It encapsulates both the job specification and scheduling preferences.
type GlobalJobRequest struct {
	// Job is the job specification to execute.
	Job *models.Job `json:"Job"`

	// Scheduling contains options for how the job should be distributed.
	Scheduling SchedulingOptions `json:"Scheduling"`

	// ClientID identifies the client submitting the job.
	ClientID string `json:"ClientID,omitempty"`

	// PriorityBoost allows increasing job priority for urgent workloads.
	PriorityBoost int `json:"PriorityBoost,omitempty"`
}

// SchedulingOptions controls how jobs are distributed across the Global VM.
type SchedulingOptions struct {
	// SpreadAcrossRegions indicates how many different regions the job should span.
	// Set to 0 for automatic (single best region).
	// Set to 1 for single region placement.
	// Set to N > 1 for multi-region distribution.
	SpreadAcrossRegions int `json:"SpreadAcrossRegions,omitempty"`

	// MaxLatency specifies the maximum acceptable network latency to nodes.
	// Nodes with higher latency will be deprioritized or excluded.
	MaxLatency time.Duration `json:"MaxLatency,omitempty"`

	// PreferredRegions is a list of regions to prefer for job placement.
	// Nodes in these regions will be ranked higher.
	PreferredRegions []string `json:"PreferredRegions,omitempty"`

	// PreferLowCost when true, prioritizes nodes with lower cost.
	PreferLowCost bool `json:"PreferLowCost,omitempty"`

	// RequireGPUVendor specifies required GPU vendors (e.g., "nvidia", "amd").
	RequireGPUVendor []string `json:"RequireGPUVendor,omitempty"`

	// MinMemoryGB specifies minimum memory per node in GB.
	MinMemoryGB uint64 `json:"MinMemoryGB,omitempty"`

	// MinCPU specifies minimum CPU cores per node.
	MinCPU float64 `json:"MinCPU,omitempty"`

	// ExcludeNodeIDs is a list of node IDs to exclude from placement.
	ExcludeNodeIDs []string `json:"ExcludeNodeIDs,omitempty"`

	// Exclusive when true, requests dedicated nodes without other workloads.
	Exclusive bool `json:"Exclusive,omitempty"`
}

// GlobalJobResponse is returned after a successful job submission.
type GlobalJobResponse struct {
	// JobID is the unique identifier assigned to the submitted job.
	JobID string `json:"JobID"`

	// EvaluationID tracks the scheduling evaluation.
	EvaluationID string `json:"EvaluationID,omitempty"`

	// AllocatedNodes lists nodes selected for initial execution.
	AllocatedNodes []NodeSelection `json:"AllocatedNodes,omitempty"`

	// Warnings contains non-fatal issues encountered during submission.
	Warnings []string `json:"Warnings,omitempty"`

	// EstimatedCost is the projected credit cost for the job.
	EstimatedCost float64 `json:"EstimatedCost,omitempty"`

	// QueuePosition indicates the job's position if queued (0 if running).
	QueuePosition int `json:"QueuePosition,omitempty"`
}

// GlobalJobStatus represents the current state of a job in the Global VM.
type GlobalJobStatus struct {
	// JobID is the unique identifier of the job.
	JobID string `json:"JobID"`

	// State is the current state of the job.
	State models.JobStateType `json:"State"`

	// Executions tracks all executions of this job.
	Executions []GlobalExecutionStatus `json:"Executions,omitempty"`

	// TotalNodes is the number of nodes currently running this job.
	TotalNodes int `json:"TotalNodes"`

	// CompletedNodes is the number of nodes that completed successfully.
	CompletedNodes int `json:"CompletedNodes"`

	// FailedNodes is the number of nodes that failed.
	FailedNodes int `json:"FailedNodes"`

	// Regions is the set of regions where the job is running.
	Regions []string `json:"Regions,omitempty"`

	// CreateTime is when the job was submitted.
	CreateTime time.Time `json:"CreateTime"`

	// TotalRuntime is the total execution time across all nodes.
	TotalRuntime time.Duration `json:"TotalRuntime"`

	// CreditsUsed is the total credits consumed so far.
	CreditsUsed float64 `json:"CreditsUsed,omitempty"`
}

// GlobalExecutionStatus represents the status of a single execution.
type GlobalExecutionStatus struct {
	// ExecutionID is the unique identifier for this execution.
	ExecutionID string `json:"ExecutionID"`

	// NodeID is the node running this execution.
	NodeID string `json:"NodeID"`

	// NodeRegion is the region of the node.
	NodeRegion string `json:"NodeRegion,omitempty"`

	// State is the current execution state.
	State models.ExecutionStateType `json:"State"`

	// StartTime is when this execution started.
	StartTime time.Time `json:"StartTime,omitempty"`

	// EndTime is when this execution ended (if completed).
	EndTime time.Time `json:"EndTime,omitempty"`

	// ExitCode is the process exit code (if completed).
	ExitCode int `json:"ExitCode,omitempty"`

	// Error contains any execution error.
	Error string `json:"Error,omitempty"`
}

// GlobalVMEndpoint defines the API for submitting and managing jobs
// on the Global VM. It presents the entire network as a single computer.
type GlobalVMEndpoint interface {
	// SubmitJob submits a job to the Global VM for execution.
	// The scheduler automatically selects appropriate nodes based on
	// job requirements and scheduling options.
	SubmitJob(ctx context.Context, req GlobalJobRequest) (*GlobalJobResponse, error)

	// GetJobStatus retrieves the current status of a job.
	GetJobStatus(ctx context.Context, jobID string) (*GlobalJobStatus, error)

	// ScaleJob adjusts the number of nodes running a job.
	// This is useful for batch jobs that need more or fewer workers.
	ScaleJob(ctx context.Context, jobID string, targetCount int) error

	// CancelJob stops a running job across all nodes.
	CancelJob(ctx context.Context, jobID string, reason string) error

	// GetJobLogs retrieves logs from a job execution.
	GetJobLogs(ctx context.Context, jobID string, opts LogOptions) (*LogResponse, error)
}

// LogOptions controls log retrieval.
type LogOptions struct {
	// ExecutionID specifies a particular execution (empty for all).
	ExecutionID string `json:"ExecutionID,omitempty"`

	// Tail returns only the last N lines.
	Tail int `json:"Tail,omitempty"`

	// Follow streams logs in real-time.
	Follow bool `json:"Follow,omitempty"`
}

// LogResponse contains job logs.
type LogResponse struct {
	// Logs contains the log output.
	Logs string `json:"Logs"`

	// Complete indicates if the execution has finished.
	Complete bool `json:"Complete"`
}

// Endpoint implements the GlobalVMEndpoint interface.
// It coordinates job submission between the Global VM abstraction
// and the underlying orchestrator components.
type Endpoint struct {
	scheduler          GlobalScheduler
	capacityProvider   GlobalCapacityProvider
	jobSubmitter       JobSubmitter
	statusProvider     JobStatusProvider
	nodeSelector       orchestrator.NodeSelector
}

// JobSubmitter is an interface for submitting jobs to the orchestrator.
type JobSubmitter interface {
	SubmitJob(ctx context.Context, req orchestrator.SubmitJobRequest) (*orchestrator.SubmitJobResponse, error)
}

// JobStatusProvider is an interface for getting job status.
type JobStatusProvider interface {
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	GetExecutions(ctx context.Context, jobID string) ([]models.Execution, error)
}

// EndpointOption configures the endpoint.
type EndpointOption func(*Endpoint)

// NewEndpoint creates a new Global VM endpoint.
func NewEndpoint(
	scheduler GlobalScheduler,
	capacityProvider GlobalCapacityProvider,
	opts ...EndpointOption,
) *Endpoint {
	e := &Endpoint{
		scheduler:        scheduler,
		capacityProvider: capacityProvider,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// WithJobSubmitter sets the job submitter.
func WithJobSubmitter(submitter JobSubmitter) EndpointOption {
	return func(e *Endpoint) {
		e.jobSubmitter = submitter
	}
}

// WithStatusProvider sets the status provider.
func WithStatusProvider(provider JobStatusProvider) EndpointOption {
	return func(e *Endpoint) {
		e.statusProvider = provider
	}
}

// WithNodeSelector sets the node selector.
func WithNodeSelector(selector orchestrator.NodeSelector) EndpointOption {
	return func(e *Endpoint) {
		e.nodeSelector = selector
	}
}

// SubmitJob submits a job to the Global VM.
func (e *Endpoint) SubmitJob(ctx context.Context, req GlobalJobRequest) (*GlobalJobResponse, error) {
	log.Ctx(ctx).Info().
		Str("jobName", req.Job.Name).
		Str("clientID", req.ClientID).
		Msg("Submitting job to Global VM")

	// Validate the job
	if err := req.Job.Validate(); err != nil {
		return nil, fmt.Errorf("job validation failed: %w", err)
	}

	// Check global capacity
	capacity, err := e.capacityProvider.GetAvailableCapacity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check capacity: %w", err)
	}

	// Validate capacity requirements
	if err := e.validateCapacity(ctx, req.Job, capacity); err != nil {
		return nil, fmt.Errorf("insufficient capacity: %w", err)
	}

	// Select nodes for the job
	schedulingReq := GlobalSchedulingRequest{
		Job:               req.Job,
		Scheduling:        req.Scheduling,
		TargetCount:       req.Job.Count,
		AvailableCapacity: capacity,
	}

	selections, err := e.scheduler.SelectNodes(ctx, schedulingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to select nodes: %w", err)
	}

	// If no nodes selected, queue the job
	if len(selections) == 0 {
		return &GlobalJobResponse{
			JobID:          req.Job.ID,
			Warnings:       []string{"No suitable nodes available, job queued"},
			QueuePosition:  1, // Would be calculated from actual queue
		}, nil
	}

	// Estimate cost
	estimatedCost := e.estimateCost(req.Job, selections)

	// Submit to orchestrator if submitter is configured
	var evalID string
	if e.jobSubmitter != nil {
		submitReq := orchestrator.SubmitJobRequest{
			Job:              req.Job,
			ClientInstanceID: req.ClientID,
		}
		resp, err := e.jobSubmitter.SubmitJob(ctx, submitReq)
		if err != nil {
			return nil, fmt.Errorf("failed to submit job: %w", err)
		}
		evalID = resp.EvaluationID
	}

	return &GlobalJobResponse{
		JobID:          req.Job.ID,
		EvaluationID:   evalID,
		AllocatedNodes: selections,
		EstimatedCost:  estimatedCost,
	}, nil
}

// GetJobStatus retrieves the current status of a job.
func (e *Endpoint) GetJobStatus(ctx context.Context, jobID string) (*GlobalJobStatus, error) {
	if e.statusProvider == nil {
		return nil, fmt.Errorf("status provider not configured")
	}

	job, err := e.statusProvider.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	executions, err := e.statusProvider.GetExecutions(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	status := &GlobalJobStatus{
		JobID:       jobID,
		State:       job.State.StateType,
		TotalNodes:  len(executions),
		CreateTime:  job.GetCreateTime(),
	}

	// Process executions
	regionSet := make(map[string]bool)
	for _, exec := range executions {
		execStatus := GlobalExecutionStatus{
			ExecutionID: exec.ID,
			NodeID:      exec.NodeID,
			State:       exec.ComputeState.StateType,
			StartTime:   time.Unix(0, exec.CreateTime),
		}

		if exec.IsTerminalState() {
			status.CompletedNodes++
			if exec.ComputeState.StateType == models.ExecutionStateFailed {
				status.FailedNodes++
			}
			execStatus.EndTime = time.Unix(0, exec.ModifyTime)
		}

		if exec.RunOutput != nil {
			execStatus.ExitCode = exec.RunOutput.ExitCode
			execStatus.Error = exec.RunOutput.ErrorMsg
		}

		status.Executions = append(status.Executions, execStatus)
	}

	// Unique regions
	status.Regions = make([]string, 0, len(regionSet))
	for r := range regionSet {
		status.Regions = append(status.Regions, r)
	}

	return status, nil
}

// ScaleJob adjusts the number of nodes running a job.
func (e *Endpoint) ScaleJob(ctx context.Context, jobID string, targetCount int) error {
	if targetCount < 0 {
		return fmt.Errorf("target count must be non-negative")
	}

	// Get current job status
	status, err := e.GetJobStatus(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job status: %w", err)
	}

	if status.State.IsTerminal() {
		return fmt.Errorf("cannot scale job in terminal state %s", status.State)
	}

	// TODO: Implement actual scaling via job update
	log.Ctx(ctx).Info().
		Str("jobID", jobID).
		Int("current", status.TotalNodes).
		Int("target", targetCount).
		Msg("Scaling job")

	return nil
}

// CancelJob stops a running job.
func (e *Endpoint) CancelJob(ctx context.Context, jobID string, reason string) error {
	log.Ctx(ctx).Info().
		Str("jobID", jobID).
		Str("reason", reason).
		Msg("Canceling job")

	// TODO: Implement via orchestrator StopJob
	return nil
}

// GetJobLogs retrieves logs from a job execution.
func (e *Endpoint) GetJobLogs(ctx context.Context, jobID string, opts LogOptions) (*LogResponse, error) {
	// TODO: Implement log retrieval
	return &LogResponse{
		Logs:     "Log retrieval not yet implemented",
		Complete: false,
	}, nil
}

// validateCapacity checks if the job can fit in available capacity.
func (e *Endpoint) validateCapacity(ctx context.Context, job *models.Job, capacity *GlobalResources) error {
	task := job.Task()
	if task == nil {
		return nil
	}

	resources := task.ResourcesConfig
	if resources == nil {
		return nil
	}

	// Parse CPU requirement
	if resources.CPU != "" {
		// Simple validation - in production would parse the CPU string
		if capacity.AvailableCPU < 0.1 {
			return fmt.Errorf("insufficient CPU capacity")
		}
	}

	// Parse memory requirement
	if resources.Memory != "" {
		// Simple validation
		if capacity.AvailableMemory < 1<<20 { // Less than 1MB
			return fmt.Errorf("insufficient memory capacity")
		}
	}

	// Check GPU availability if required
	if len(resources.GPU) > 0 && capacity.AvailableGPU == 0 {
		return fmt.Errorf("no GPU capacity available")
	}

	return nil
}

// estimateCost estimates the credit cost for running a job.
func (e *Endpoint) estimateCost(job *models.Job, selections []NodeSelection) float64 {
	// Base cost calculation
	// In production, this would consider:
	// - Node region/cost tier
	// - Resource usage
	// - Job duration estimate
	// - Priority

	baseCost := 1.0 // 1 credit base

	// Add cost per node
	nodeCost := float64(len(selections)) * 0.5

	// Add GPU premium
	task := job.Task()
	if task != nil && task.ResourcesConfig != nil && len(task.ResourcesConfig.GPU) > 0 {
		nodeCost *= 2.0 // GPU jobs cost 2x
	}

	return baseCost + nodeCost
}
