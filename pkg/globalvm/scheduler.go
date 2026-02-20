//go:build unit

package globalvm

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator/nodes"
	"github.com/rs/zerolog/log"
)

// GlobalSchedulingRequest contains all information needed for global scheduling.
type GlobalSchedulingRequest struct {
	// Job is the job to schedule.
	Job *models.Job `json:"Job"`

	// Scheduling contains the scheduling preferences.
	Scheduling SchedulingOptions `json:"Scheduling"`

	// TargetCount is the desired number of nodes.
	TargetCount int `json:"TargetCount"`

	// AvailableCapacity is the current global capacity snapshot.
	AvailableCapacity *GlobalResources `json:"AvailableCapacity,omitempty"`

	// ExistingExecutions are nodes already running this job (for scaling).
	ExistingExecutions []string `json:"ExistingExecutions,omitempty"`
}

// NodeSelection represents a selected node for job execution.
type NodeSelection struct {
	// NodeID is the identifier of the selected node.
	NodeID string `json:"NodeID"`

	// Rank is the node's ranking score.
	Rank int `json:"Rank"`

	// Reason explains why this node was selected.
	Reason string `json:"Reason,omitempty"`

	// Region is the node's geographic region.
	Region string `json:"Region,omitempty"`

	// Resources is the available resources on this node.
	Resources models.Resources `json:"Resources,omitempty"`

	// EstimatedLatency is the estimated network latency to this node.
	EstimatedLatency time.Duration `json:"EstimatedLatency,omitempty"`

	// Cost is the relative cost of using this node.
	Cost float64 `json:"Cost,omitempty"`
}

// GlobalScheduler provides intelligent scheduling across the global compute network.
// It wraps the existing node selector and adds global optimization capabilities.
type GlobalScheduler interface {
	// SelectNodes selects the best nodes for a job based on global scheduling rules.
	SelectNodes(ctx context.Context, req GlobalSchedulingRequest) ([]NodeSelection, error)

	// GetBestNodeForJob returns a single best node for a job.
	// This is useful for single-node jobs or picking a leader.
	GetBestNodeForJob(ctx context.Context, job *models.Job) (*NodeSelection, error)

	// GetNodesByRegion returns nodes grouped by region for multi-region placement.
	GetNodesByRegion(ctx context.Context, job *models.Job) (map[string][]NodeSelection, error)
}

// Scheduler implements the GlobalScheduler interface.
// It integrates with the existing orchestrator selection components
// and adds global optimization capabilities.
type Scheduler struct {
	nodeSelector       orchestrator.NodeSelector
	nodeRanker         orchestrator.NodeRanker
	capacityProvider   GlobalCapacityProvider
	nodeLookup         nodes.Lookup
	regionRanker       *RegionRanker
	costCalculator     CostCalculator
}

// SchedulerOption configures the scheduler.
type SchedulerOption func(*Scheduler)

// NewScheduler creates a new global scheduler.
func NewScheduler(
	nodeSelector orchestrator.NodeSelector,
	capacityProvider GlobalCapacityProvider,
	opts ...SchedulerOption,
) *Scheduler {
	s := &Scheduler{
		nodeSelector:     nodeSelector,
		capacityProvider: capacityProvider,
		regionRanker:     NewRegionRanker(),
		costCalculator:   &DefaultCostCalculator{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithNodeRanker sets a custom node ranker.
func WithNodeRanker(ranker orchestrator.NodeRanker) SchedulerOption {
	return func(s *Scheduler) {
		s.nodeRanker = ranker
	}
}

// WithNodeLookup sets the node lookup.
func WithNodeLookup(lookup nodes.Lookup) SchedulerOption {
	return func(s *Scheduler) {
		s.nodeLookup = lookup
	}
}

// WithRegionRanker sets a custom region ranker.
func WithRegionRanker(ranker *RegionRanker) SchedulerOption {
	return func(s *Scheduler) {
		s.regionRanker = ranker
	}
}

// WithCostCalculator sets a custom cost calculator.
func WithCostCalculator(calc CostCalculator) SchedulerOption {
	return func(s *Scheduler) {
		s.costCalculator = calc
	}
}

// SelectNodes selects the best nodes for a job based on global scheduling rules.
func (s *Scheduler) SelectNodes(ctx context.Context, req GlobalSchedulingRequest) ([]NodeSelection, error) {
	log.Ctx(ctx).Debug().
		Str("jobID", req.Job.ID).
		Int("targetCount", req.TargetCount).
		Msg("Selecting nodes for global scheduling")

	// Get ranked nodes from the existing selector
	matched, rejected, err := s.nodeSelector.MatchingNodes(ctx, req.Job)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching nodes: %w", err)
	}

	// Log rejected nodes for debugging
	if len(rejected) > 0 {
		log.Ctx(ctx).Debug().
			Int("rejected", len(rejected)).
			Msg("Nodes rejected by selector")
	}

	// Convert to selections
	selections := s.convertToSelections(ctx, matched)

	// Apply global scheduling optimizations
	selections = s.applyGlobalOptimizations(ctx, req, selections)

	// Limit to target count
	if req.TargetCount > 0 && len(selections) > req.TargetCount {
		selections = selections[:req.TargetCount]
	}

	return selections, nil
}

// GetBestNodeForJob returns a single best node for a job.
func (s *Scheduler) GetBestNodeForJob(ctx context.Context, job *models.Job) (*NodeSelection, error) {
	req := GlobalSchedulingRequest{
		Job:         job,
		TargetCount: 1,
	}

	selections, err := s.SelectNodes(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(selections) == 0 {
		return nil, fmt.Errorf("no suitable nodes available for job %s", job.ID)
	}

	return &selections[0], nil
}

// GetNodesByRegion returns nodes grouped by region for multi-region placement.
func (s *Scheduler) GetNodesByRegion(ctx context.Context, job *models.Job) (map[string][]NodeSelection, error) {
	req := GlobalSchedulingRequest{
		Job:         job,
		TargetCount: 0, // No limit
	}

	selections, err := s.SelectNodes(ctx, req)
	if err != nil {
		return nil, err
	}

	// Group by region
	regions := make(map[string][]NodeSelection)
	for _, sel := range selections {
		region := sel.Region
		if region == "" {
			region = "unknown"
		}
		regions[region] = append(regions[region], sel)
	}

	return regions, nil
}

// convertToSelections converts orchestrator node ranks to global node selections.
func (s *Scheduler) convertToSelections(ctx context.Context, ranks []orchestrator.NodeRank) []NodeSelection {
	selections := make([]NodeSelection, 0, len(ranks))

	for _, rank := range ranks {
		selection := NodeSelection{
			NodeID:     rank.NodeInfo.ID(),
			Rank:       rank.Rank,
			Reason:     rank.Reason,
			Region:     s.extractRegion(rank.NodeInfo),
			Resources:  rank.NodeInfo.ComputeNodeInfo.AvailableCapacity,
		}

		// Calculate cost
		selection.Cost = s.costCalculator.CalculateCost(rank.NodeInfo)

		selections = append(selections, selection)
	}

	return selections
}

// applyGlobalOptimizations applies global scheduling optimizations.
func (s *Scheduler) applyGlobalOptimizations(ctx context.Context, req GlobalSchedulingRequest, selections []NodeSelection) []NodeSelection {
	if len(selections) == 0 {
		return selections
	}

	// Apply preferred regions filter
	if len(req.Scheduling.PreferredRegions) > 0 {
		selections = s.applyPreferredRegions(selections, req.Scheduling.PreferredRegions)
	}

	// Apply latency constraints
	if req.Scheduling.MaxLatency > 0 {
		selections = s.applyLatencyConstraints(selections, req.Scheduling.MaxLatency)
	}

	// Apply cost preference
	if req.Scheduling.PreferLowCost {
		selections = s.applyCostPreference(selections)
	}

	// Apply multi-region spread
	if req.Scheduling.SpreadAcrossRegions > 1 {
		selections = s.applyRegionSpread(selections, req.Scheduling.SpreadAcrossRegions)
	}

	// Apply exclusions
	if len(req.Scheduling.ExcludeNodeIDs) > 0 {
		selections = s.applyExclusions(selections, req.Scheduling.ExcludeNodeIDs)
	}

	// Sort by final rank
	sort.Slice(selections, func(i, j int) bool {
		return selections[i].Rank > selections[j].Rank
	})

	return selections
}

// applyPreferredRegions boosts ranking for preferred regions.
func (s *Scheduler) applyPreferredRegions(selections []NodeSelection, preferred []string) []NodeSelection {
	preferredSet := make(map[string]bool)
	for _, r := range preferred {
		preferredSet[r] = true
	}

	for i := range selections {
		if preferredSet[selections[i].Region] {
			selections[i].Rank += 100 // Boost preferred regions
			selections[i].Reason = "preferred region: " + selections[i].Region
		}
	}

	return selections
}

// applyLatencyConstraints filters nodes by latency.
func (s *Scheduler) applyLatencyConstraints(selections []NodeSelection, maxLatency time.Duration) []NodeSelection {
	filtered := make([]NodeSelection, 0, len(selections))

	for _, sel := range selections {
		// If no latency info, include the node
		if sel.EstimatedLatency == 0 {
			filtered = append(filtered, sel)
			continue
		}

		if sel.EstimatedLatency <= maxLatency {
			filtered = append(filtered, sel)
		}
	}

	return filtered
}

// applyCostPreference sorts by cost for budget-conscious jobs.
func (s *Scheduler) applyCostPreference(selections []NodeSelection) []NodeSelection {
	// Sort by cost ascending
	sort.Slice(selections, func(i, j int) bool {
		return selections[i].Cost < selections[j].Cost
	})

	// Give higher rank to lower cost nodes
	for i := range selections {
		selections[i].Rank += (len(selections) - i) * 10
	}

	return selections
}

// applyRegionSpread ensures distribution across regions.
func (s *Scheduler) applyRegionSpread(selections []NodeSelection, targetRegions int) []NodeSelection {
	// Group by region
	regions := make(map[string][]NodeSelection)
	for _, sel := range selections {
		regions[sel.Region] = append(regions[sel.Region], sel)
	}

	// If we have enough regions, take one from each
	if len(regions) >= targetRegions {
		result := make([]NodeSelection, 0, targetRegions)
		for _, regionNodes := range regions {
			if len(result) >= targetRegions {
				break
			}
			if len(regionNodes) > 0 {
				result = append(result, regionNodes[0])
			}
		}
		return result
	}

	// Otherwise return selections as-is
	return selections
}

// applyExclusions removes specific nodes from selection.
func (s *Scheduler) applyExclusions(selections []NodeSelection, excludeIDs []string) []NodeSelection {
	excludeSet := make(map[string]bool)
	for _, id := range excludeIDs {
		excludeSet[id] = true
	}

	filtered := make([]NodeSelection, 0, len(selections))
	for _, sel := range selections {
		if !excludeSet[sel.NodeID] {
			filtered = append(filtered, sel)
		}
	}

	return filtered
}

// extractRegion extracts region information from node info.
func (s *Scheduler) extractRegion(info models.NodeInfo) string {
	// Try to get region from node labels
	if info.Labels != nil {
		if region, ok := info.Labels["region"]; ok {
			return region
		}
		if region, ok := info.Labels["topology.kubernetes.io/region"]; ok {
			return region
		}
	}

	// Default region
	return "default"
}

// RegionRanker ranks regions based on various criteria.
type RegionRanker struct {
	regionLatencies map[string]time.Duration
	regionCosts     map[string]float64
}

// NewRegionRanker creates a new region ranker.
func NewRegionRanker() *RegionRanker {
	return &RegionRanker{
		regionLatencies: make(map[string]time.Duration),
		regionCosts:     make(map[string]float64),
	}
}

// SetRegionLatency sets the estimated latency to a region.
func (r *RegionRanker) SetRegionLatency(region string, latency time.Duration) {
	r.regionLatencies[region] = latency
}

// SetRegionCost sets the relative cost for a region.
func (r *RegionRanker) SetRegionCost(region string, cost float64) {
	r.regionCosts[region] = cost
}

// RankRegion returns a score for a region (higher is better).
func (r *RegionRanker) RankRegion(region string) int {
	score := 50 // Base score

	// Adjust for latency (lower is better)
	if latency, ok := r.regionLatencies[region]; ok {
		if latency < 50*time.Millisecond {
			score += 20
		} else if latency < 100*time.Millisecond {
			score += 10
		} else if latency > 200*time.Millisecond {
			score -= 10
		}
	}

	// Adjust for cost (lower is better)
	if cost, ok := r.regionCosts[region]; ok {
		if cost < 1.0 {
			score += 10
		} else if cost > 2.0 {
			score -= 10
		}
	}

	return score
}

// CostCalculator calculates the relative cost of using a node.
type CostCalculator interface {
	CalculateCost(info models.NodeInfo) float64
}

// DefaultCostCalculator provides a default cost calculation.
type DefaultCostCalculator struct{}

// CalculateCost calculates cost based on node resources.
func (d *DefaultCostCalculator) CalculateCost(info models.NodeInfo) float64 {
	cost := 1.0 // Base cost

	resources := info.ComputeNodeInfo.AvailableCapacity

	// Add cost for CPU
	cost += float64(resources.CPU) * 0.1

	// Add cost for memory (per GB)
	cost += float64(resources.Memory>>30) * 0.01

	// Add premium for GPUs
	cost += float64(len(resources.GPUs)) * 0.5

	return cost
}
