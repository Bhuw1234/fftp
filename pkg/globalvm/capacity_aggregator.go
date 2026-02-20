// Package globalvm provides the Unified Global VM abstraction for DEparrow.
// It presents all compute nodes in the network as a single "infinite computer"
// where users submit jobs without needing to select specific nodes.
package globalvm

import (
	"context"
	"sync"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator/nodes"
	"github.com/rs/zerolog/log"
)

// GlobalResources represents the aggregated resources across all nodes.
type GlobalResources struct {
	// Total resources across all nodes
	TotalCPU     float64 `json:"TotalCPU"`
	TotalMemory  uint64  `json:"TotalMemory"`
	TotalDisk    uint64  `json:"TotalDisk"`
	TotalGPU     int     `json:"TotalGPU"`

	// Currently available resources
	AvailableCPU    float64 `json:"AvailableCPU"`
	AvailableMemory uint64  `json:"AvailableMemory"`
	AvailableDisk   uint64  `json:"AvailableDisk"`
	AvailableGPU    int     `json:"AvailableGPU"`

	// Node statistics
	TotalNodes   int `json:"TotalNodes"`
	HealthyNodes int `json:"HealthyNodes"`

	// GPU breakdown by vendor
	NVIDIAGPUs int `json:"NVIDIAGPUs,omitempty"`
	AMDGPUs    int `json:"AMDGPUs,omitempty"`
	IntelGPUs  int `json:"IntelGPUs,omitempty"`

	// Timestamp of this snapshot
	SnapshotTime time.Time `json:"SnapshotTime"`
}

// CapacitySnapshot captures the state of global capacity at a point in time.
type CapacitySnapshot struct {
	Resources   GlobalResources `json:"Resources"`
	NodeDetails []NodeCapacity  `json:"NodeDetails,omitempty"`
	Timestamp   time.Time       `json:"Timestamp"`
}

// NodeCapacity represents a single node's contribution to the global pool.
type NodeCapacity struct {
	NodeID       string          `json:"NodeID"`
	Resources    models.Resources `json:"Resources"`
	GPUs         []models.GPU    `json:"GPUs,omitempty"`
	Location     NodeLocation    `json:"Location"`
	LastSeen     time.Time       `json:"LastSeen"`
	IsHealthy    bool            `json:"IsHealthy"`
	CapabilityScore float64      `json:"CapabilityScore"`
}

// NodeLocation represents geographic location of a node.
type NodeLocation struct {
	Region    string `json:"Region,omitempty"`     // e.g., "us-west-1"
	Country   string `json:"Country,omitempty"`    // e.g., "US"
	City      string `json:"City,omitempty"`       // e.g., "San Francisco"
	Latitude  float64 `json:"Latitude,omitempty"`
	Longitude float64 `json:"Longitude,omitempty"`
}

// GlobalCapacityProvider provides a unified view of all compute resources.
type GlobalCapacityProvider interface {
	// GetGlobalCapacity returns the total available resources across all nodes.
	GetGlobalCapacity(ctx context.Context) (*GlobalResources, error)

	// GetAvailableCapacity returns currently available resources.
	GetAvailableCapacity(ctx context.Context) (*GlobalResources, error)

	// PredictCapacity models future capacity based on running jobs.
	PredictCapacity(ctx context.Context, horizon time.Duration) (*GlobalResources, error)

	// GetSnapshot returns a point-in-time snapshot of global capacity.
	GetSnapshot(ctx context.Context) (*CapacitySnapshot, error)

	// Subscribe returns a channel for capacity updates.
	Subscribe(ctx context.Context) (<-chan GlobalResources, error)
}

// CapacityAggregator aggregates capacity from multiple nodes into a global view.
type CapacityAggregator struct {
	nodeLookup   nodes.Lookup
	mu           sync.RWMutex
	lastSnapshot *CapacitySnapshot
	updateChan   chan GlobalResources
	snapshotInterval time.Duration
}

// NewCapacityAggregator creates a new capacity aggregator.
func NewCapacityAggregator(nodeLookup nodes.Lookup, opts ...AggregatorOption) *CapacityAggregator {
	a := &CapacityAggregator{
		nodeLookup:       nodeLookup,
		updateChan:       make(chan GlobalResources, 100),
		snapshotInterval: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// AggregatorOption configures the capacity aggregator.
type AggregatorOption func(*CapacityAggregator)

// WithSnapshotInterval sets the interval for capacity snapshots.
func WithSnapshotInterval(d time.Duration) AggregatorOption {
	return func(a *CapacityAggregator) {
		a.snapshotInterval = d
	}
}

// GetGlobalCapacity returns the total available resources across all nodes.
func (a *CapacityAggregator) GetGlobalCapacity(ctx context.Context) (*GlobalResources, error) {
	snapshot, err := a.computeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &snapshot.Resources, nil
}

// GetAvailableCapacity returns currently available resources.
func (a *CapacityAggregator) GetAvailableCapacity(ctx context.Context) (*GlobalResources, error) {
	return a.GetGlobalCapacity(ctx)
}

// PredictCapacity models future capacity based on running jobs.
func (a *CapacityAggregator) PredictCapacity(ctx context.Context, horizon time.Duration) (*GlobalResources, error) {
	// Get current capacity
	current, err := a.GetGlobalCapacity(ctx)
	if err != nil {
		return nil, err
	}

	// For now, return current capacity as prediction
	// TODO: Implement predictive modeling based on job queue and historical patterns
	predicted := *current
	predicted.SnapshotTime = time.Now().Add(horizon)
	
	return &predicted, nil
}

// GetSnapshot returns a point-in-time snapshot of global capacity.
func (a *CapacityAggregator) GetSnapshot(ctx context.Context) (*CapacitySnapshot, error) {
	a.mu.RLock()
	if a.lastSnapshot != nil && time.Since(a.lastSnapshot.Timestamp) < a.snapshotInterval {
		defer a.mu.RUnlock()
		return a.lastSnapshot, nil
	}
	a.mu.RUnlock()

	// Compute new snapshot
	snapshot, err := a.computeSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	a.lastSnapshot = snapshot
	a.mu.Unlock()

	return snapshot, nil
}

// Subscribe returns a channel for capacity updates.
func (a *CapacityAggregator) Subscribe(ctx context.Context) (<-chan GlobalResources, error) {
	ch := make(chan GlobalResources, 10)
	
	go func() {
		ticker := time.NewTicker(a.snapshotInterval)
		defer close(ch)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				resources, err := a.GetGlobalCapacity(ctx)
				if err != nil {
					log.Warn().Err(err).Msg("failed to get global capacity")
					continue
				}
				select {
				case ch <- *resources:
				default:
					// Channel full, skip update
				}
			}
		}
	}()

	return ch, nil
}

// computeSnapshot computes a fresh snapshot from all nodes.
func (a *CapacityAggregator) computeSnapshot(ctx context.Context) (*CapacitySnapshot, error) {
	snapshot := &CapacitySnapshot{
		Timestamp: time.Now(),
	}

	// Get all nodes from the lookup
	nodeStates, err := a.nodeLookup.List(ctx)
	if err != nil {
		return nil, err
	}

	snapshot.Resources.TotalNodes = len(nodeStates)

	for _, nodeState := range nodeStates {
		capacity := a.nodeStateToCapacity(nodeState)
		snapshot.NodeDetails = append(snapshot.NodeDetails, capacity)
		
		if capacity.IsHealthy {
			snapshot.Resources.HealthyNodes++
			
			// Aggregate total resources
			snapshot.Resources.TotalCPU += capacity.Resources.CPU
			snapshot.Resources.TotalMemory += capacity.Resources.Memory
			snapshot.Resources.TotalDisk += capacity.Resources.Disk
			snapshot.Resources.TotalGPU += len(capacity.GPUs)

			// Count GPUs by vendor
			for _, gpu := range capacity.GPUs {
				switch gpu.Vendor {
				case models.GPUVendorNvidia:
					snapshot.Resources.NVIDIAGPUs++
				case models.GPUVendorAMDATI:
					snapshot.Resources.AMDGPUs++
				case models.GPUVendorIntel:
					snapshot.Resources.IntelGPUs++
				}
			}

			// Aggregate available resources
			// Note: This assumes nodeInfo contains available capacity
			// In production, we'd get this from ComputeNodeInfo.AvailableCapacity
			snapshot.Resources.AvailableCPU += capacity.Resources.CPU
			snapshot.Resources.AvailableMemory += capacity.Resources.Memory
			snapshot.Resources.AvailableDisk += capacity.Resources.Disk
			snapshot.Resources.AvailableGPU += len(capacity.GPUs)
		}
	}

	snapshot.Resources.SnapshotTime = snapshot.Timestamp
	return snapshot, nil
}

// nodeStateToCapacity converts a NodeState to NodeCapacity.
func (a *CapacityAggregator) nodeStateToCapacity(nodeState models.NodeState) NodeCapacity {
	capacity := NodeCapacity{
		NodeID:    nodeState.Info.ID(),
		IsHealthy: nodeState.IsConnected(),
		LastSeen:  nodeState.ConnectionState.LastHeartbeat,
	}

	// Extract compute node info if available
	computeInfo := nodeState.Info.ComputeNodeInfo
	if computeInfo.AvailableCapacity.CPU > 0 || computeInfo.AvailableCapacity.Memory > 0 {
		capacity.Resources = computeInfo.AvailableCapacity
	} else {
		capacity.Resources = computeInfo.MaxCapacity
	}
	// Extract GPUs from the resources
	capacity.GPUs = computeInfo.MaxCapacity.GPUs

	return capacity
}

// GlobalVMSummary returns a human-readable summary of the Global VM.
type GlobalVMSummary struct {
	TotalCPU        string `json:"TotalCPU"`
	TotalMemory     string `json:"TotalMemory"`
	TotalGPU        int    `json:"TotalGPU"`
	AvailableCPU    string `json:"AvailableCPU"`
	AvailableMemory string `json:"AvailableMemory"`
	AvailableGPU    int    `json:"AvailableGPU"`
	TotalNodes      int    `json:"TotalNodes"`
	HealthyNodes    int    `json:"HealthyNodes"`
	GPUBreakdown    string `json:"GPUBreakdown"`
}

// Summary returns a human-readable summary.
func (g *GlobalResources) Summary() GlobalVMSummary {
	return GlobalVMSummary{
		TotalCPU:        formatCPU(g.TotalCPU),
		TotalMemory:     formatBytes(g.TotalMemory),
		TotalGPU:        g.TotalGPU,
		AvailableCPU:    formatCPU(g.AvailableCPU),
		AvailableMemory: formatBytes(g.AvailableMemory),
		AvailableGPU:    g.AvailableGPU,
		TotalNodes:      g.TotalNodes,
		HealthyNodes:    g.HealthyNodes,
		GPUBreakdown:    formatGPUBreakdown(g.NVIDIAGPUs, g.AMDGPUs, g.IntelGPUs),
	}
}

func formatCPU(cpu float64) string {
	if cpu >= 1000 {
		return "Infinite"
	}
	if cpu >= 100 {
		return "100+ cores"
	}
	return "Many cores"
}

func formatBytes(bytes uint64) string {
	if bytes >= 1<<60 {
		return "Infinite"
	}
	if bytes >= 1<<40 {
		return "Petabytes"
	}
	if bytes >= 1<<30 {
		return "Terabytes"
	}
	return "Gigabytes"
}

func formatGPUBreakdown(nvidia, amd, intel int) string {
	parts := []string{}
	if nvidia > 0 {
		parts = append(parts, "NVIDIA")
	}
	if amd > 0 {
		parts = append(parts, "AMD")
	}
	if intel > 0 {
		parts = append(parts, "Intel")
	}
	if len(parts) == 0 {
		return "No GPUs"
	}
	return "Mixed: " + joinStrings(parts, ", ")
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
