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

func TestScheduler_SelectNodes(t *testing.T) {
	tests := []struct {
		name          string
		request       GlobalSchedulingRequest
		mockSelector  *mockNodeSelector
		mockCapacity  *mockCapacityProvider
		expectCount   int
		expectError   bool
	}{
		{
			name: "select single node",
			request: GlobalSchedulingRequest{
				Job:         createTestJob("job-1", models.JobTypeBatch, 1),
				TargetCount: 1,
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{
					{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
					{NodeInfo: createTestNodeInfo("node-2", "us-east"), Rank: 8},
				},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{HealthyNodes: 2},
			},
			expectCount: 1,
		},
		{
			name: "select multiple nodes",
			request: GlobalSchedulingRequest{
				Job:         createTestJob("job-2", models.JobTypeBatch, 3),
				TargetCount: 3,
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{
					{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
					{NodeInfo: createTestNodeInfo("node-2", "us-west"), Rank: 9},
					{NodeInfo: createTestNodeInfo("node-3", "us-east"), Rank: 8},
					{NodeInfo: createTestNodeInfo("node-4", "eu-west"), Rank: 7},
				},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{HealthyNodes: 4},
			},
			expectCount: 3,
		},
		{
			name: "no matching nodes",
			request: GlobalSchedulingRequest{
				Job:         createTestJob("job-3", models.JobTypeBatch, 1),
				TargetCount: 1,
			},
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{HealthyNodes: 0},
			},
			expectCount: 0,
		},
		{
			name: "selector error",
			request: GlobalSchedulingRequest{
				Job:         createTestJob("job-4", models.JobTypeBatch, 1),
				TargetCount: 1,
			},
			mockSelector: &mockNodeSelector{
				err: assert.AnError,
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewScheduler(tt.mockSelector, tt.mockCapacity)

			selections, err := scheduler.SelectNodes(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, selections, tt.expectCount)
		})
	}
}

func TestScheduler_GetBestNodeForJob(t *testing.T) {
	tests := []struct {
		name         string
		job          *models.Job
		mockSelector *mockNodeSelector
		mockCapacity *mockCapacityProvider
		expectNode   string
		expectError  bool
	}{
		{
			name: "find best node",
			job:  createTestJob("job-1", models.JobTypeBatch, 1),
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{
					{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
					{NodeInfo: createTestNodeInfo("node-2", "us-east"), Rank: 8},
				},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{HealthyNodes: 2},
			},
			expectNode: "node-1", // Highest rank
		},
		{
			name: "no nodes available",
			job:  createTestJob("job-2", models.JobTypeBatch, 1),
			mockSelector: &mockNodeSelector{
				nodes: []orchestrator.NodeRank{},
			},
			mockCapacity: &mockCapacityProvider{
				capacity: &GlobalResources{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewScheduler(tt.mockSelector, tt.mockCapacity)

			selection, err := scheduler.GetBestNodeForJob(context.Background(), tt.job)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, selection)
			assert.Equal(t, tt.expectNode, selection.NodeID)
		})
	}
}

func TestScheduler_GetNodesByRegion(t *testing.T) {
	mockSelector := &mockNodeSelector{
		nodes: []orchestrator.NodeRank{
			{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
			{NodeInfo: createTestNodeInfo("node-2", "us-west"), Rank: 9},
			{NodeInfo: createTestNodeInfo("node-3", "us-east"), Rank: 8},
			{NodeInfo: createTestNodeInfo("node-4", "eu-west"), Rank: 7},
		},
	}
	mockCapacity := &mockCapacityProvider{
		capacity: &GlobalResources{HealthyNodes: 4},
	}

	scheduler := NewScheduler(mockSelector, mockCapacity)

	regions, err := scheduler.GetNodesByRegion(context.Background(), createTestJob("job-1", models.JobTypeBatch, 1))

	require.NoError(t, err)
	require.NotNil(t, regions)

	// Should have 3 regions
	assert.Len(t, regions, 3)

	// Verify region counts
	assert.Len(t, regions["us-west"], 2)
	assert.Len(t, regions["us-east"], 1)
	assert.Len(t, regions["eu-west"], 1)
}

func TestScheduler_ApplyPreferredRegions(t *testing.T) {
	scheduler := &Scheduler{}

	selections := []NodeSelection{
		{NodeID: "node-1", Region: "us-west", Rank: 10},
		{NodeID: "node-2", Region: "us-east", Rank: 8},
		{NodeID: "node-3", Region: "eu-west", Rank: 6},
	}

	result := scheduler.applyPreferredRegions(selections, []string{"us-east", "eu-west"})

	// us-east and eu-west should have boosted ranks
	assert.Equal(t, 10, result[0].Rank) // us-west unchanged
	assert.Equal(t, 108, result[1].Rank) // us-east boosted +100
	assert.Equal(t, 106, result[2].Rank) // eu-west boosted +100
}

func TestScheduler_ApplyLatencyConstraints(t *testing.T) {
	scheduler := &Scheduler{}

	selections := []NodeSelection{
		{NodeID: "node-1", Region: "local", Rank: 10, EstimatedLatency: 10 * time.Millisecond},
		{NodeID: "node-2", Region: "nearby", Rank: 8, EstimatedLatency: 50 * time.Millisecond},
		{NodeID: "node-3", Region: "far", Rank: 6, EstimatedLatency: 200 * time.Millisecond},
		{NodeID: "node-4", Region: "unknown", Rank: 4, EstimatedLatency: 0}, // No latency info
	}

	result := scheduler.applyLatencyConstraints(selections, 100*time.Millisecond)

	// Should include nodes with latency <= 100ms or no latency info
	assert.Len(t, result, 3)
	assert.Equal(t, "node-1", result[0].NodeID)
	assert.Equal(t, "node-2", result[1].NodeID)
	assert.Equal(t, "node-4", result[2].NodeID)
}

func TestScheduler_ApplyCostPreference(t *testing.T) {
	scheduler := &Scheduler{}

	selections := []NodeSelection{
		{NodeID: "node-1", Rank: 10, Cost: 5.0},
		{NodeID: "node-2", Rank: 8, Cost: 2.0},
		{NodeID: "node-3", Rank: 6, Cost: 1.0},
	}

	result := scheduler.applyCostPreference(selections)

	// Should be sorted by cost (lowest first)
	assert.Equal(t, "node-3", result[0].NodeID) // Cost 1.0
	assert.Equal(t, "node-2", result[1].NodeID) // Cost 2.0
	assert.Equal(t, "node-1", result[2].NodeID) // Cost 5.0

	// Ranks should be adjusted based on position
	assert.Greater(t, result[0].Rank, result[1].Rank)
	assert.Greater(t, result[1].Rank, result[2].Rank)
}

func TestScheduler_ApplyRegionSpread(t *testing.T) {
	scheduler := &Scheduler{}

	selections := []NodeSelection{
		{NodeID: "node-1", Region: "us-west", Rank: 10},
		{NodeID: "node-2", Region: "us-west", Rank: 9},
		{NodeID: "node-3", Region: "us-east", Rank: 8},
		{NodeID: "node-4", Region: "eu-west", Rank: 7},
	}

	// Request 2 regions spread
	result := scheduler.applyRegionSpread(selections, 2)

	// Should have 2 nodes from different regions
	assert.Len(t, result, 2)
	regions := make(map[string]bool)
	for _, sel := range result {
		regions[sel.Region] = true
	}
	assert.Len(t, regions, 2)

	// Request 3 regions spread
	result = scheduler.applyRegionSpread(selections, 3)
	assert.Len(t, result, 3)
}

func TestScheduler_ApplyExclusions(t *testing.T) {
	scheduler := &Scheduler{}

	selections := []NodeSelection{
		{NodeID: "node-1", Rank: 10},
		{NodeID: "node-2", Rank: 8},
		{NodeID: "node-3", Rank: 6},
		{NodeID: "node-4", Rank: 4},
	}

	result := scheduler.applyExclusions(selections, []string{"node-2", "node-4"})

	assert.Len(t, result, 2)
	assert.Equal(t, "node-1", result[0].NodeID)
	assert.Equal(t, "node-3", result[1].NodeID)
}

func TestScheduler_ExtractRegion(t *testing.T) {
	scheduler := &Scheduler{}

	tests := []struct {
		name          string
		info          models.NodeInfo
		expectedRegion string
	}{
		{
			name: "from region label",
			info: models.NodeInfo{
				NodeID: "node-1",
				Labels: map[string]string{"region": "us-west-1"},
			},
			expectedRegion: "us-west-1",
		},
		{
			name: "from kubernetes region label",
			info: models.NodeInfo{
				NodeID: "node-2",
				Labels: map[string]string{"topology.kubernetes.io/region": "eu-west-1"},
			},
			expectedRegion: "eu-west-1",
		},
		{
			name: "no region label",
			info: models.NodeInfo{
				NodeID: "node-3",
				Labels: map[string]string{"other": "value"},
			},
			expectedRegion: "default",
		},
		{
			name: "nil labels",
			info: models.NodeInfo{
				NodeID: "node-4",
			},
			expectedRegion: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := scheduler.extractRegion(tt.info)
			assert.Equal(t, tt.expectedRegion, region)
		})
	}
}

func TestRegionRanker_RankRegion(t *testing.T) {
	ranker := NewRegionRanker()

	// Set latencies and costs
	ranker.SetRegionLatency("us-west", 30*time.Millisecond)
	ranker.SetRegionLatency("eu-west", 150*time.Millisecond)
	ranker.SetRegionLatency("asia-east", 300*time.Millisecond)

	ranker.SetRegionCost("us-west", 1.5)
	ranker.SetRegionCost("eu-west", 1.2)
	ranker.SetRegionCost("asia-east", 0.8)

	// Rank regions
	usWestScore := ranker.RankRegion("us-west")
	euWestScore := ranker.RankRegion("eu-west")
	asiaEastScore := ranker.RankRegion("asia-east")

	// us-west: latency < 50ms (+20), cost 1.5 (no adjustment) = 70
	// eu-west: latency 150ms (no adjustment), cost 1.2 (no adjustment) = 50
	// asia-east: latency > 200ms (-10), cost < 1.0 (+10) = 50

	// us-west should score highest (low latency bonus)
	assert.Greater(t, usWestScore, euWestScore)
	assert.Greater(t, usWestScore, asiaEastScore)

	// eu-west should be at base score
	assert.Equal(t, 50, euWestScore)

	// asia-east should be base score (latency penalty offset by cost bonus)
	assert.Equal(t, 50, asiaEastScore)
}

func TestDefaultCostCalculator_CalculateCost(t *testing.T) {
	calc := &DefaultCostCalculator{}

	tests := []struct {
		name        string
		info        models.NodeInfo
		expectMin   float64
		expectMax   float64
	}{
		{
			name: "basic node",
			info: models.NodeInfo{
				ComputeNodeInfo: models.ComputeNodeInfo{
					AvailableCapacity: models.Resources{
						CPU:    4.0,
						Memory: 16 << 30,
					},
				},
			},
			expectMin: 1.0,
			expectMax: 2.0,
		},
		{
			name: "node with GPU",
			info: models.NodeInfo{
				ComputeNodeInfo: models.ComputeNodeInfo{
					AvailableCapacity: models.Resources{
						CPU:    8.0,
						Memory: 32 << 30,
						GPUs: []models.GPU{
							{Vendor: models.GPUVendorNvidia, Name: "RTX 4090"},
						},
					},
				},
			},
			expectMin: 2.0,
			expectMax: 5.0,
		},
		{
			name: "minimal node",
			info: models.NodeInfo{
				ComputeNodeInfo: models.ComputeNodeInfo{
					AvailableCapacity: models.Resources{
						CPU:    1.0,
						Memory: 1 << 30,
					},
				},
			},
			expectMin: 1.0,
			expectMax: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calc.CalculateCost(tt.info)
			assert.GreaterOrEqual(t, cost, tt.expectMin)
			assert.LessOrEqual(t, cost, tt.expectMax)
		})
	}
}

func TestScheduler_Integration(t *testing.T) {
	// Integration test combining multiple optimizations
	mockSelector := &mockNodeSelector{
		nodes: []orchestrator.NodeRank{
			{NodeInfo: createTestNodeInfo("node-1", "us-west"), Rank: 10},
			{NodeInfo: createTestNodeInfo("node-2", "us-west"), Rank: 9},
			{NodeInfo: createTestNodeInfo("node-3", "us-east"), Rank: 8},
			{NodeInfo: createTestNodeInfo("node-4", "eu-west"), Rank: 7},
			{NodeInfo: createTestNodeInfo("node-5", "asia-east"), Rank: 6},
		},
	}
	mockCapacity := &mockCapacityProvider{
		capacity: &GlobalResources{HealthyNodes: 5},
	}

	scheduler := NewScheduler(mockSelector, mockCapacity)

	// Request with multiple constraints
	req := GlobalSchedulingRequest{
		Job: createTestJob("integration-job", models.JobTypeBatch, 2),
		Scheduling: SchedulingOptions{
			PreferredRegions: []string{"us-west", "us-east"},
			PreferLowCost:    false,
			ExcludeNodeIDs:   []string{"node-4"},
		},
		TargetCount: 2,
	}

	selections, err := scheduler.SelectNodes(context.Background(), req)

	require.NoError(t, err)
	assert.Len(t, selections, 2)

	// node-4 should be excluded
	for _, sel := range selections {
		assert.NotEqual(t, "node-4", sel.NodeID)
	}
}
