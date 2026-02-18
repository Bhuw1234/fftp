package deparrow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/tools"
)

// JobTool provides the ability to submit compute jobs to the DEparrow network.
// This is the primary way for AI agents to execute distributed compute tasks.
type JobTool struct {
	client *Client
}

// NewJobTool creates a new job submission tool.
func NewJobTool(client *Client) *JobTool {
	return &JobTool{client: client}
}

// Name returns the tool name.
func (t *JobTool) Name() string {
	return "deparrow_submit_job"
}

// Description returns the tool description.
func (t *JobTool) Description() string {
	return `Submit a compute job to the DEparrow network.

This tool allows you to run containerized workloads on the distributed
DEparrow compute network. Jobs can be Docker containers with custom commands,
resource requirements, and input/output specifications.

Use this when you need to:
- Execute compute-intensive tasks
- Run distributed data processing
- Perform machine learning inference
- Process large datasets in parallel

Example usage:
  image: "python:3.11-slim"
  command: "python -c 'print(2+2)'"
  resources:
    cpu: "500m"
    memory: "256Mi"
`
}

// Parameters returns the JSON schema for tool parameters.
func (t *JobTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"image": map[string]interface{}{
				"type":        "string",
				"description": "Docker image to run (e.g., 'ubuntu:latest', 'python:3.11-slim')",
			},
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Command to execute inside the container",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"items":       map[string]string{"type": "string"},
				"description": "Arguments to pass to the command",
			},
			"env": map[string]interface{}{
				"type":        "object",
				"additionalProperties": map[string]string{"type": "string"},
				"description": "Environment variables to set in the container",
			},
			"cpu": map[string]interface{}{
				"type":        "string",
				"description": "CPU requirement (e.g., '500m' for 0.5 cores)",
				"default":     "100m",
			},
			"memory": map[string]interface{}{
				"type":        "string",
				"description": "Memory requirement (e.g., '256Mi', '1Gi')",
				"default":     "128Mi",
			},
			"gpu": map[string]interface{}{
				"type":        "string",
				"description": "GPU requirement (e.g., '1' for one GPU)",
				"default":     "",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Job timeout in seconds (default: 600)",
				"default":     600,
			},
			"priority": map[string]interface{}{
				"type":        "integer",
				"description": "Job priority (0-100, higher = more priority)",
				"default":     50,
				"minimum":     0,
				"maximum":     100,
			},
			"inputs": map[string]interface{}{
				"type":        "array",
				"description": "Input data sources",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"storage_source": map[string]string{"type": "string"},
						"source":         map[string]string{"type": "string"},
						"path":           map[string]string{"type": "string"},
					},
				},
			},
			"wait": map[string]interface{}{
				"type":        "boolean",
				"description": "Wait for job completion and return results",
				"default":     false,
			},
		},
		"required": []string{"image"},
	}
}

// Execute runs the job submission tool.
func (t *JobTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	// Extract image (required)
	image, ok := args["image"].(string)
	if !ok || image == "" {
		return tools.ErrorResult("image parameter is required")
	}

	// Build job spec
	spec := &JobSpec{
		Image: image,
		Resources: &ResourceSpec{
			CPU:    "100m",
			Memory: "128Mi",
		},
		Timeout: 600,
		Labels:  make(map[string]string),
	}

	// Parse command
	if cmd, ok := args["command"].(string); ok && cmd != "" {
		// Split command into parts for proper execution
		spec.Command = strings.Fields(cmd)
	}

	// Parse args if provided separately
	if argsList, ok := args["args"].([]interface{}); ok {
		for _, arg := range argsList {
			if argStr, ok := arg.(string); ok {
				spec.Command = append(spec.Command, argStr)
			}
		}
	}

	// Parse environment variables
	if env, ok := args["env"].(map[string]interface{}); ok {
		spec.Env = make(map[string]string)
		for k, v := range env {
			if vStr, ok := v.(string); ok {
				spec.Env[k] = vStr
			}
		}
	}

	// Parse resource requirements
	if cpu, ok := args["cpu"].(string); ok && cpu != "" {
		spec.Resources.CPU = cpu
	}
	if mem, ok := args["memory"].(string); ok && mem != "" {
		spec.Resources.Memory = mem
	}
	if gpu, ok := args["gpu"].(string); ok && gpu != "" {
		spec.Resources.GPU = gpu
	}

	// Parse timeout
	if timeout, ok := args["timeout"].(float64); ok {
		spec.Timeout = int(timeout)
	} else if timeout, ok := args["timeout"].(int); ok {
		spec.Timeout = timeout
	}

	// Parse priority
	if priority, ok := args["priority"].(float64); ok {
		spec.Priority = int(priority)
	} else if priority, ok := args["priority"].(int); ok {
		spec.Priority = priority
	}

	// Parse inputs
	if inputs, ok := args["inputs"].([]interface{}); ok {
		for _, input := range inputs {
			if inputMap, ok := input.(map[string]interface{}); ok {
				inputSpec := InputSpec{}
				if src, ok := inputMap["storage_source"].(string); ok {
					inputSpec.StorageSource = src
				}
				if src, ok := inputMap["source"].(string); ok {
					inputSpec.Source = src
				}
				if path, ok := inputMap["path"].(string); ok {
					inputSpec.Path = path
				}
				spec.Inputs = append(spec.Inputs, inputSpec)
			}
		}
	}

	// Submit job
	job, err := t.client.SubmitJob(ctx, spec)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to submit job: %v", err))
	}

	// Check if we should wait for completion
	if wait, _ := args["wait"].(bool); wait {
		return t.waitForJob(ctx, job.ID)
	}

	// Return immediate response with job ID
	result := fmt.Sprintf(
		"Job submitted successfully!\n\nJob ID: %s\nStatus: %s\nCredit Cost: %.2f\n\n"+
			"Use 'deparrow_job_status' with job_id='%s' to check progress.",
		job.ID, job.Status, job.CreditCost, job.ID,
	)

	return tools.UserResult(result)
}

// waitForJob polls for job completion and returns the results.
func (t *JobTool) waitForJob(ctx context.Context, jobID string) *tools.ToolResult {
	for {
		job, err := t.client.GetJob(ctx, jobID)
		if err != nil {
			return tools.ErrorResult(fmt.Sprintf("Failed to get job status: %v", err))
		}

		switch job.Status {
		case JobStatusCompleted:
			result := fmt.Sprintf(
				"Job completed successfully!\n\nJob ID: %s\nDuration: %.1fs\n\n",
				job.ID, job.Results.Duration,
			)
			if job.Results.Stdout != "" {
				result += "Output:\n" + job.Results.Stdout
			}
			if job.Results.OutputCID != "" {
				result += fmt.Sprintf("\n\nOutput CID: %s", job.Results.OutputCID)
			}
			return tools.UserResult(result)

		case JobStatusFailed:
			errMsg := job.Error
			if job.Results != nil && job.Results.Stderr != "" {
				errMsg = job.Results.Stderr
			}
			return tools.ErrorResult(fmt.Sprintf("Job failed: %s", errMsg))

		case JobStatusCancelled:
			return tools.ErrorResult("Job was cancelled")

		case JobStatusPending, JobStatusRunning:
			// Continue waiting
			select {
			case <-ctx.Done():
				return tools.ErrorResult("Job wait cancelled by context")
			default:
				// Wait a bit before polling again
				// In a real implementation, use proper sleep
				continue
			}
		}
	}
}

// JobStatusTool provides the ability to check job status.
type JobStatusTool struct {
	client *Client
}

// NewJobStatusTool creates a new job status tool.
func NewJobStatusTool(client *Client) *JobStatusTool {
	return &JobStatusTool{client: client}
}

// Name returns the tool name.
func (t *JobStatusTool) Name() string {
	return "deparrow_job_status"
}

// Description returns the tool description.
func (t *JobStatusTool) Description() string {
	return "Check the status of a submitted DEparrow job."
}

// Parameters returns the JSON schema for tool parameters.
func (t *JobStatusTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"job_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the job to check",
			},
		},
		"required": []string{"job_id"},
	}
}

// Execute runs the job status tool.
func (t *JobStatusTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return tools.ErrorResult("job_id parameter is required")
	}

	job, err := t.client.GetJob(ctx, jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to get job: %v", err))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Job ID: %s\n", job.ID))
	result.WriteString(fmt.Sprintf("Status: %s\n", job.Status))
	result.WriteString(fmt.Sprintf("Credit Cost: %.2f\n", job.CreditCost))
	result.WriteString(fmt.Sprintf("Submitted: %s\n", job.SubmittedAt.Format("2006-01-02 15:04:05")))

	if job.Results != nil {
		result.WriteString(fmt.Sprintf("\nDuration: %.1f seconds\n", job.Results.Duration))
		result.WriteString(fmt.Sprintf("Exit Code: %d\n", job.Results.ExitCode))
		if job.Results.OutputCID != "" {
			result.WriteString(fmt.Sprintf("Output CID: %s\n", job.Results.OutputCID))
		}
		if job.Results.Stdout != "" {
			result.WriteString(fmt.Sprintf("\nOutput:\n%s", job.Results.Stdout))
		}
	}

	return tools.UserResult(result.String())
}

// JobListTool provides the ability to list jobs.
type JobListTool struct {
	client *Client
}

// NewJobListTool creates a new job list tool.
func NewJobListTool(client *Client) *JobListTool {
	return &JobListTool{client: client}
}

// Name returns the tool name.
func (t *JobListTool) Name() string {
	return "deparrow_list_jobs"
}

// Description returns the tool description.
func (t *JobListTool) Description() string {
	return "List all jobs submitted by the authenticated user."
}

// Parameters returns the JSON schema for tool parameters.
func (t *JobListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute runs the job list tool.
func (t *JobListTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	jobs, err := t.client.ListJobs(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to list jobs: %v", err))
	}

	if len(jobs) == 0 {
		return tools.UserResult("No jobs found.")
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d jobs:\n\n", len(jobs)))

	for i, job := range jobs {
		result.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, job.ID, job.Status))
		if job.Spec != nil && job.Spec.Image != "" {
			result.WriteString(fmt.Sprintf("   Image: %s\n", job.Spec.Image))
		}
		result.WriteString(fmt.Sprintf("   Credits: %.2f\n", job.CreditCost))
		result.WriteString(fmt.Sprintf("   Submitted: %s\n\n", job.SubmittedAt.Format("2006-01-02 15:04:05")))
	}

	return tools.UserResult(result.String())
}

// JobCancelTool provides the ability to cancel a running job.
type JobCancelTool struct {
	client *Client
}

// NewJobCancelTool creates a new job cancel tool.
func NewJobCancelTool(client *Client) *JobCancelTool {
	return &JobCancelTool{client: client}
}

// Name returns the tool name.
func (t *JobCancelTool) Name() string {
	return "deparrow_cancel_job"
}

// Description returns the tool description.
func (t *JobCancelTool) Description() string {
	return "Cancel a running job and receive partial credit refund."
}

// Parameters returns the JSON schema for tool parameters.
func (t *JobCancelTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"job_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the job to cancel",
			},
		},
		"required": []string{"job_id"},
	}
}

// Execute runs the job cancel tool.
func (t *JobCancelTool) Execute(ctx context.Context, args map[string]interface{}) *tools.ToolResult {
	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return tools.ErrorResult("job_id parameter is required")
	}

	refund, err := t.client.CancelJob(ctx, jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("Failed to cancel job: %v", err))
	}

	return tools.UserResult(fmt.Sprintf(
		"Job %s cancelled successfully.\nCredit refund: %.2f",
		jobID, refund,
	))
}

// Ensure tools implement the Tool interface
var _ tools.Tool = (*JobTool)(nil)
var _ tools.Tool = (*JobStatusTool)(nil)
var _ tools.Tool = (*JobListTool)(nil)
var _ tools.Tool = (*JobCancelTool)(nil)

// Helper function to marshal job info
func marshalJobInfo(job *Job) string {
	data, _ := json.MarshalIndent(job, "", "  ")
	return string(data)
}
