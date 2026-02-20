//go:build unit

package globalvm

import (
	"context"
	"testing"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNodeSelector implements orchestrator.NodeSelector for testing
type mockNodeSelector struct {
	nodes    []orchestrator.NodeRank
	err      error
	allNodes []models.NodeInfo
}

func (m *mockNodeSelector) AllNodes(ctx context.Context) ([]models.NodeInfo, error) {
	return m.allNodes, nil
}

func (m *mockNodeSelector) MatchingNodes(ctx context.Context, job *models.Job) (matched, rejected []orchestrator.NodeRank, err error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.nodes, nil, nil
}

// mockCapacityProvider implements GlobalCapacityProvider for testing
type mockCapacityProvider struct {
	capacity *GlobalResources
	err      error
}

func (m *mockCapacityProvider) GetGlobalCapacity(ctx context.Context) (*GlobalResources, error) {
	return m.capacity, m.err
}

func (m *mockCapacityProvider) GetAvailableCapacity(ctx context.Context) (*GlobalResources, error) {
	return m.capacity, m.err
}

func (m *mockCapacityProvider) PredictCapacity(ctx context.Context, horizon time.Duration) (*GlobalResources, error) {
	return m.capacity, m.err
}

func (m *mockCapacityProvider) GetSnapshot(ctx context.Context) (*CapacitySnapshot, error) {
	return &CapacitySnapshot{Resources: *m.capacity}, m.err
}

func (m *mockCapacityProvider) Subscribe(ctx context.Context) (<-chan GlobalResources, error) {
	ch := make(chan GlobalResources, 1)
	return ch, nil
}

// mockJobSubmitter implements JobSubmitter for testing
type mockJobSubmitter struct {
	response *orchestrator.SubmitJobResponse
	err      error
}

func (m *mockJobSubmitter) SubmitJob(ctx context.Context, req orchestrator.SubmitJobRequest) (*orchestrator.SubmitJobResponse, error) {
	return m.response, m.err
}

// mockStatusProvider implements JobStatusProvider for testing
type mockStatusProvider struct {
	job        *models.Job
	executions []models.Execution
	jobErr     error
	execErr    error
}

func (m *mockStatusProvider) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	return m.job, m.jobErr
}

func (m *mockStatusProvider) GetExecutions(ctx context.Context, jobID string) ([]models.Execution, error) {
	return m.executions, m.execErr
}

func TestEndpoint_SubmitJob(t *testing.T) {
	tests := []struct {
		name           string
		request        GlobalJobRequest
		mockSelector   *mockNodeSelector
		mockCapacity   *mockCapacityProvider
		mockSubmitter  *mockJobSubmitter
		expectError    bool
		expectJobID    bool
		expectedNodes  int
	}{
		{
			name: "successful submission",
			request: GlobalJobRequest{
				Job: createTestJob("test-job", models.JobTypeBatch, 1),
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{
					{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
					{NodeInfo: createTestNodeInfo("node-2", "us-east"), Rank: 8},
				},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{
					AvailableCPU:    100.0,
					AvailableMemory: 1024 << 30,
					AvailableGPU:    10,
					HealthyNodes:    5,
				},
			},
			mockSubmitter: &mockJobSubmitter{
				response: &orchestrator.SubmitJobResponse{
					JobID:        "test-job-id",
					EvaluationID: "eval-123",
				},
			},
			expectJobID:   true,
			expectedNodes: 1,
		},
		{
			name: "no suitable nodes",
			request: GlobalJobRequest{
				Job: createTestJob("test-job", models.JobTypeBatch, 1),
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{
					AvailableCPU:    100.0,
					AvailableMemory: 1024 << 30,
					HealthyNodes:    5,
				},
			},
			expectJobID: true,
		},
		{
			name: "capacity check failure",
			request: GlobalJobRequest{
				Job: createTestJob("test-job", models.JobTypeBatch, 1),
			},
			mockCapacity: &mockCapacityProvider{
				err: assert.AnError,
			},
			expectError: true,
		},
		{
			name: "multi-node job",
			request: GlobalJobRequest{
				Job: createTestJob("test-job", models.JobTypeBatch, 3),
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{
					{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
					{NodeInfo: createTestNodeInfo("node-2", "us-west"), Rank: 9},
					{NodeInfo: createTestNodeInfo("node-3", "us-east"), Rank: 8},
					{NodeInfo: createTestNodeInfo("node-4", "us-east"), Rank: 7},
				},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{
					AvailableCPU:    100.0,
					AvailableMemory: 1024 << 30,
					HealthyNodes:    5,
				},
			},
			expectJobID:   true,
			expectedNodes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create scheduler with mock selector
			scheduler := NewScheduler(tt.mockSelector, tt.mockCapacity)

			// Create endpoint
			endpoint := NewEndpoint(scheduler, tt.mockCapacity)
			if tt.mockSubmitter != nil {
				endpoint = NewEndpoint(scheduler, tt.mockCapacity,
					WithJobSubmitter(tt.mockSubmitter))
			}

			// Submit job
			response, err := endpoint.SubmitJob(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)

			if tt.expectJobID {
				assert.NotEmpty(t, response.JobID)
			}

			if tt.expectedNodes > 0 {
				assert.Len(t, response.AllocatedNodes, tt.expectedNodes)
			}
		})
	}
}

func TestEndpoint_GetJobStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		jobID         string
		mockStatus    *mockStatusProvider
		expectError   bool
		expectedState models.JobStateType
	}{
		{
			name:  "running job",
			jobID: "job-123",
			mockStatus: &mockStatusProvider{
				job: &models.Job{
					ID:        "job-123",
					State:     models.State[models.JobStateType]{StateType: models.JobStateTypeRunning},
					CreateTime: now.UnixNano(),
				},
				executions: []models.Execution{
					{
						ID:         "exec-1",
						JobID:      "job-123",
						NodeID:     "node-1",
						CreateTime: now.UnixNano(),
						ComputeState: models.State[models.ExecutionStateType]{
							StateType: models.ExecutionStateRunning,
						},
					},
				},
			},
			expectedState: models.JobStateTypeRunning,
		},
		{
			name:  "completed job",
			jobID: "job-456",
			mockStatus: &mockStatusProvider{
				job: &models.Job{
					ID:        "job-456",
					State:     models.State[models.JobStateType]{StateType: models.JobStateTypeCompleted},
					CreateTime: now.UnixNano(),
				},
				executions: []models.Execution{
					{
						ID:         "exec-1",
						JobID:      "job-456",
						NodeID:     "node-1",
						CreateTime: now.UnixNano(),
						ModifyTime: now.Add(time.Hour).UnixNano(),
						ComputeState: models.State[models.ExecutionStateType]{
							StateType: models.ExecutionStateCompleted,
						},
					},
				},
			},
			expectedState: models.JobStateTypeCompleted,
		},
		{
			name:  "job not found",
			jobID: "nonexistent",
			mockStatus: &mockStatusProvider{
				jobErr: assert.AnError,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := NewEndpoint(nil, nil, WithStatusProvider(tt.mockStatus))

			status, err := endpoint.GetJobStatus(context.Background(), tt.jobID)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, status)
			assert.Equal(t, tt.jobID, status.JobID)
			assert.Equal(t, tt.expectedState, status.State)
		})
	}
}

func TestEndpoint_ScaleJob(t *testing.T) {
	tests := []struct {
		name         string
		jobID        string
		targetCount  int
		mockStatus   *mockStatusProvider
		expectError  bool
	}{
		{
			name:        "scale up",
			jobID:       "job-123",
			targetCount: 5,
			mockStatus: &mockStatusProvider{
				job: &models.Job{
					ID:    "job-123",
					State: models.State[models.JobStateType]{StateType: models.JobStateTypeRunning},
				},
				executions: []models.Execution{
					{ID: "exec-1", NodeID: "node-1"},
				},
			},
		},
		{
			name:        "scale down",
			jobID:       "job-456",
			targetCount: 2,
			mockStatus: &mockStatusProvider{
				job: &models.Job{
					ID:    "job-456",
					State: models.State[models.JobStateType]{StateType: models.JobStateTypeRunning},
				},
				executions: []models.Execution{
					{ID: "exec-1", NodeID: "node-1"},
					{ID: "exec-2", NodeID: "node-2"},
					{ID: "exec-3", NodeID: "node-3"},
				},
			},
		},
		{
			name:        "negative count",
			jobID:       "job-789",
			targetCount: -1,
			expectError: true,
		},
		{
			name:        "terminal job",
			jobID:       "job-completed",
			targetCount: 5,
			mockStatus: &mockStatusProvider{
				job: &models.Job{
					ID:    "job-completed",
					State: models.State[models.JobStateType]{StateType: models.JobStateTypeCompleted},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := NewEndpoint(nil, nil, WithStatusProvider(tt.mockStatus))

			err := endpoint.ScaleJob(context.Background(), tt.jobID, tt.targetCount)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestEndpoint_ValidateCapacity(t *testing.T) {
	endpoint := &Endpoint{}

	tests := []struct {
		name      string
		job       *models.Job
		capacity  *GlobalResources
		expectErr bool
	}{
		{
			name: "sufficient capacity",
			job:  createTestJobWithResources("job-1", 4.0, 16<<30),
			capacity: &GlobalResources{
				AvailableCPU:    100.0,
				AvailableMemory: 1024 << 30,
				AvailableGPU:    10,
			},
		},
		{
			name: "insufficient CPU",
			job:  createTestJobWithResources("job-2", 4.0, 16<<30),
			capacity: &GlobalResources{
				AvailableCPU:    0.05,
				AvailableMemory: 1024 << 30,
			},
			expectErr: true,
		},
		{
			name: "no GPU when required",
			job:  createTestJobWithGPU("job-3", "nvidia"),
			capacity: &GlobalResources{
				AvailableCPU:    100.0,
				AvailableMemory: 1024 << 30,
				AvailableGPU:    0,
			},
			expectErr: true,
		},
		{
			name: "job without resources",
			job:  createTestJob("job-4", models.JobTypeBatch, 1),
			capacity: &GlobalResources{
				AvailableCPU:    100.0,
				AvailableMemory: 1024 << 30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := endpoint.validateCapacity(context.Background(), tt.job, tt.capacity)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEndpoint_EstimateCost(t *testing.T) {
	endpoint := &Endpoint{}

	tests := []struct {
		name           string
		job            *models.Job
		selections     []NodeSelection
		expectMinCost  float64
	}{
		{
			name:          "single node batch job",
			job:           createTestJob("job-1", models.JobTypeBatch, 1),
			selections:    []NodeSelection{{NodeID: "node-1", Rank: 10}},
			expectMinCost: 1.0,
		},
		{
			name:          "multi-node job",
			job:           createTestJob("job-2", models.JobTypeBatch, 3),
			selections:    []NodeSelection{{NodeID: "node-1"}, {NodeID: "node-2"}, {NodeID: "node-3"}},
			expectMinCost: 1.0,
		},
		{
			name:          "GPU job",
			job:           createTestJobWithGPU("job-3", "nvidia"),
			selections:    []NodeSelection{{NodeID: "gpu-node-1"}},
			expectMinCost: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := endpoint.estimateCost(tt.job, tt.selections)
			assert.GreaterOrEqual(t, cost, tt.expectMinCost)
		})
	}
}

// Helper functions

func createTestJob(id string, jobType string, count int) *models.Job {
	job := &models.Job{
		ID:        id,
		Name:      id,
		Type:      jobType,
		Count:     count,
		Namespace: "default",
		Tasks: []*models.Task{
			{
				Name: "main",
				Engine: &models.SpecConfig{
					Type: "docker",
					Params: map[string]interface{}{
						"Image": "ubuntu:latest",
						"Entrypoint": []string{"echo", "hello"},
					},
				},
				Publisher: &models.SpecConfig{},
			},
		},
	}
	job.Normalize()
	return job
}

func createTestJobWithResources(id string, cpu float64, memory uint64) *models.Job {
	job := createTestJob(id, models.JobTypeBatch, 1)
	job.Tasks[0].ResourcesConfig = &models.ResourcesConfig{
		CPU:    "4",
		Memory: "16GiB",
	}
	return job
}

func createTestJobWithGPU(id string, vendor string) *models.Job {
	job := createTestJob(id, models.JobTypeBatch, 1)
	job.Tasks[0].ResourcesConfig = &models.ResourcesConfig{
		GPU: "1",
	}
	return job
}

func createTestNodeInfo(id string, region string) models.NodeInfo {
	return models.NodeInfo{
		NodeID:   id,
		NodeType: models.NodeTypeCompute,
		Labels: map[string]string{
			"region": region,
		},
		ComputeNodeInfo: models.ComputeNodeInfo{
			AvailableCapacity: models.Resources{
				CPU:    4.0,
				Memory: 16 << 30,
				Disk:   100 << 30,
			},
		},
	}
}
