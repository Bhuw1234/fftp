package globalvm

import (
	"context"
	"testing"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNodeLookup implements nodes.Lookup for testing
type mockNodeLookup struct {
	states []models.NodeState
	err    error
}

func (m *mockNodeLookup) List(ctx context.Context, filters ...nodes.NodeStateFilter) ([]models.NodeState, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Apply filters if any
	result := m.states
	for _, filter := range filters {
		filtered := make([]models.NodeState, 0)
		for _, state := range result {
			if filter(state) {
				filtered = append(filtered, state)
			}
		}
		result = filtered
	}
	return result, nil
}

func (m *mockNodeLookup) Get(ctx context.Context, nodeID string) (models.NodeState, error) {
	for _, s := range m.states {
		if s.Info.ID() == nodeID {
			return s, nil
		}
	}
	return models.NodeState{}, nodes.NewErrNodeNotFound(nodeID)
}

func (m *mockNodeLookup) GetByPrefix(ctx context.Context, prefix string) (models.NodeState, error) {
	for _, s := range m.states {
		if len(s.Info.ID()) >= len(prefix) && s.Info.ID()[:len(prefix)] == prefix {
			return s, nil
		}
	}
	return models.NodeState{}, nodes.NewErrNodeNotFound(prefix)
}

func TestCapacityAggregator_GetGlobalCapacity(t *testing.T) {
	tests := []struct {
		name         string
		states       []models.NodeState
		expected     GlobalResources
		expectError  bool
	}{
		{
			name: "empty cluster",
			states: []models.NodeState{},
			expected: GlobalResources{
				TotalNodes:   0,
				HealthyNodes: 0,
			},
		},
		{
			name: "single healthy node",
			states: []models.NodeState{
				createMockNodeState("node-1", true, 4.0, 16<<30, 100<<30, []models.GPU{
					{Vendor: models.GPUVendorNvidia, Name: "RTX 4090"},
				}),
			},
			expected: GlobalResources{
				TotalCPU:        4.0,
				TotalMemory:     16 << 30,
				TotalDisk:       100 << 30,
				TotalGPU:        1,
				AvailableCPU:    4.0,
				AvailableMemory: 16 << 30,
				AvailableDisk:   100 << 30,
				AvailableGPU:    1,
				TotalNodes:      1,
				HealthyNodes:    1,
				NVIDIAGPUs:      1,
			},
		},
		{
			name: "multiple nodes with mixed GPUs",
			states: []models.NodeState{
				createMockNodeState("node-1", true, 4.0, 16<<30, 100<<30, []models.GPU{
					{Vendor: models.GPUVendorNvidia, Name: "RTX 4090"},
				}),
				createMockNodeState("node-2", true, 8.0, 32<<30, 200<<30, []models.GPU{
					{Vendor: models.GPUVendorAMDATI, Name: "RX 7900 XTX"},
				}),
				createMockNodeState("node-3", false, 2.0, 8<<30, 50<<30, nil),
			},
			expected: GlobalResources{
				TotalCPU:        12.0,
				TotalMemory:     48 << 30,
				TotalDisk:       300 << 30,
				TotalGPU:        2,
				AvailableCPU:    12.0,
				AvailableMemory: 48 << 30,
				AvailableDisk:   300 << 30,
				AvailableGPU:    2,
				TotalNodes:      3,
				HealthyNodes:    2,
				NVIDIAGPUs:      1,
				AMDGPUs:        1,
			},
		},
		{
			name: "node without compute info",
			states: []models.NodeState{
				createMockNodeStateNoCompute("node-1", true),
			},
			expected: GlobalResources{
				TotalNodes:   1,
				HealthyNodes: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := &mockNodeLookup{states: tt.states}
			agg := NewCapacityAggregator(lookup)

			result, err := agg.GetGlobalCapacity(context.Background())
			
			if tt.expectError {
				require.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expected.TotalCPU, result.TotalCPU)
			assert.Equal(t, tt.expected.TotalMemory, result.TotalMemory)
			assert.Equal(t, tt.expected.TotalDisk, result.TotalDisk)
			assert.Equal(t, tt.expected.TotalGPU, result.TotalGPU)
			assert.Equal(t, tt.expected.TotalNodes, result.TotalNodes)
			assert.Equal(t, tt.expected.HealthyNodes, result.HealthyNodes)
			assert.Equal(t, tt.expected.NVIDIAGPUs, result.NVIDIAGPUs)
			assert.Equal(t, tt.expected.AMDGPUs, result.AMDGPUs)
		})
	}
}

func TestCapacityAggregator_GetSnapshot(t *testing.T) {
	lookup := &mockNodeLookup{
		states: []models.NodeState{
			createMockNodeState("node-1", true, 4.0, 16<<30, 100<<30, nil),
		},
	}
	agg := NewCapacityAggregator(lookup, WithSnapshotInterval(time.Minute))

	snapshot, err := agg.GetSnapshot(context.Background())
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, 1, snapshot.Resources.TotalNodes)
	assert.Equal(t, 1, snapshot.Resources.HealthyNodes)
	assert.Len(t, snapshot.NodeDetails, 1)
	assert.Equal(t, "node-1", snapshot.NodeDetails[0].NodeID)
}

func TestCapacityAggregator_Subscribe(t *testing.T) {
	lookup := &mockNodeLookup{
		states: []models.NodeState{
			createMockNodeState("node-1", true, 4.0, 16<<30, 100<<30, nil),
		},
	}
	agg := NewCapacityAggregator(lookup, WithSnapshotInterval(100*time.Millisecond))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ch, err := agg.Subscribe(ctx)
	require.NoError(t, err)

	// Should receive at least one update
	select {
	case update := <-ch:
		assert.Equal(t, 1, update.TotalNodes)
	case <-ctx.Done():
		t.Fatal("timed out waiting for capacity update")
	}
}

func TestGlobalResources_Summary(t *testing.T) {
	tests := []struct {
		name     string
		input    GlobalResources
		expected GlobalVMSummary
	}{
		{
			name: "small cluster",
			input: GlobalResources{
				TotalCPU:        16.0,
				TotalMemory:     64 << 30,
				TotalGPU:        4,
				AvailableCPU:    8.0,
				AvailableMemory: 32 << 30,
				AvailableGPU:    2,
				TotalNodes:      4,
				HealthyNodes:    4,
				NVIDIAGPUs:      2,
				AMDGPUs:        2,
			},
			expected: GlobalVMSummary{
				TotalCPU:     "Many cores",
				TotalMemory:  "Terabytes",
				TotalGPU:     4,
				AvailableCPU: "Many cores",
				AvailableMemory: "Terabytes",
				AvailableGPU: 2,
				TotalNodes:   4,
				HealthyNodes: 4,
				GPUBreakdown: "Mixed: NVIDIA, AMD",
			},
		},
		{
			name: "large cluster",
			input: GlobalResources{
				TotalCPU:        1000.0,
				TotalMemory:     2 << 50,
				TotalGPU:        100,
				AvailableCPU:    800.0,
				AvailableMemory: 1 << 50,
				AvailableGPU:    80,
				TotalNodes:      100,
				HealthyNodes:    95,
				NVIDIAGPUs:      50,
				AMDGPUs:        30,
				IntelGPUs:     20,
			},
			expected: GlobalVMSummary{
				TotalCPU:     "Infinite",
				TotalMemory:  "Petabytes",
				TotalGPU:     100,
				AvailableCPU: "Infinite",
				AvailableMemory: "Petabytes",
				AvailableGPU: 80,
				TotalNodes:   100,
				HealthyNodes: 95,
				GPUBreakdown: "Mixed: NVIDIA, AMD, Intel",
			},
		},
		{
			name: "no GPUs",
			input: GlobalResources{
				TotalNodes:   10,
				HealthyNodes: 10,
			},
			expected: GlobalVMSummary{
				TotalCPU:     "Many cores",
				TotalMemory:  "Gigabytes",
				TotalNodes:   10,
				HealthyNodes: 10,
				GPUBreakdown: "No GPUs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Summary()
			assert.Equal(t, tt.expected.TotalCPU, result.TotalCPU)
			assert.Equal(t, tt.expected.TotalMemory, result.TotalMemory)
			assert.Equal(t, tt.expected.TotalGPU, result.TotalGPU)
			assert.Equal(t, tt.expected.GPUBreakdown, result.GPUBreakdown)
		})
	}
}

// Helper functions

func createMockNodeState(id string, connected bool, cpu float64, memory, disk uint64, gpus []models.GPU) models.NodeState {
	connectionStatus := models.NodeStates.DISCONNECTED
	if connected {
		connectionStatus = models.NodeStates.CONNECTED
	}
	return models.NodeState{
		Info: models.NodeInfo{
			NodeID:   id,
			NodeType: models.NodeTypeCompute,
			ComputeNodeInfo: models.ComputeNodeInfo{
				AvailableCapacity: models.Resources{
					CPU:    cpu,
					Memory: memory,
					Disk:   disk,
				},
				MaxCapacity: models.Resources{
					CPU:    cpu,
					Memory: memory,
					Disk:   disk,
					GPUs:   gpus,
				},
			},
		},
		ConnectionState: models.ConnectionState{
			Status:        connectionStatus,
			LastHeartbeat: time.Now(),
		},
	}
}

func createMockNodeStateNoCompute(id string, connected bool) models.NodeState {
	connectionStatus := models.NodeStates.DISCONNECTED
	if connected {
		connectionStatus = models.NodeStates.CONNECTED
	}
	return models.NodeState{
		Info: models.NodeInfo{
			NodeID:   id,
			NodeType: models.NodeTypeCompute,
		},
		ConnectionState: models.ConnectionState{
			Status:        connectionStatus,
			LastHeartbeat: time.Now(),
		},
	}
}
