//go:build unit

// Package globalvm provides global scheduling capabilities for the distributed compute network.
// This file implements geographic-aware node ranking.
package globalvm

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator"
	"github.com/rs/zerolog/log"
)

// GeoRankerConfig configures the geographic ranker.
type GeoRankerConfig struct {
	// OriginRegion is the region from which jobs originate (for latency calculation).
	OriginRegion string

	// PreferLocal gives preference to nodes in the same region.
	PreferLocal bool

	// LocalBoost is the rank boost for local region nodes.
	LocalBoost int

	// MaxLatency is the maximum acceptable latency for a node.
	MaxLatency time.Duration

	// LatencyPenalty is the rank penalty per 50ms of latency.
	LatencyPenalty int

	// ExcludeHighLatency excludes nodes exceeding MaxLatency.
	ExcludeHighLatency bool

	// PreferContinent gives preference to nodes on the same continent.
	PreferContinent bool

	// ContinentBoost is the rank boost for same-continent nodes.
	ContinentBoost int
}

// DefaultGeoRankerConfig returns the default configuration.
func DefaultGeoRankerConfig() GeoRankerConfig {
	return GeoRankerConfig{
		PreferLocal:        true,
		LocalBoost:         50,
		MaxLatency:         500 * time.Millisecond,
		LatencyPenalty:     5,
		ExcludeHighLatency: true,
		PreferContinent:    true,
		ContinentBoost:     20,
	}
}

// GeoRanker implements the NodeRanker interface with geographic awareness.
// It ranks nodes based on their proximity and latency characteristics.
type GeoRanker struct {
	config       GeoRankerConfig
	latencyMatrix LatencyMatrix
	locationDetector LocationDetector
}

// NewGeoRanker creates a new geographic ranker.
func NewGeoRanker(config GeoRankerConfig, latencyMatrix LatencyMatrix) *GeoRanker {
	return &GeoRanker{
		config:        config,
		latencyMatrix: latencyMatrix,
		locationDetector: NewLocationDetector(DefaultLocationDetectorConfig()),
	}
}

// NewGeoRankerWithDetector creates a geographic ranker with a custom location detector.
func NewGeoRankerWithDetector(config GeoRankerConfig, latencyMatrix LatencyMatrix, detector LocationDetector) *GeoRanker {
	return &GeoRanker{
		config:           config,
		latencyMatrix:    latencyMatrix,
		locationDetector: detector,
	}
}

// RankNodes ranks nodes based on geographic proximity and latency.
func (r *GeoRanker) RankNodes(ctx context.Context, job models.Job, nodes []models.NodeInfo) ([]orchestrator.NodeRank, error) {
	ranks := make([]orchestrator.NodeRank, len(nodes))

	// Determine origin region from job labels or config
	originRegion := r.config.OriginRegion
	if job.Labels != nil {
		if region, ok := job.Labels["region"]; ok {
			originRegion = region
		}
		if region, ok := job.Labels["origin-region"]; ok {
			originRegion = region
		}
	}

	// Check for scheduling constraints
	preferredRegions := r.getPreferredRegions(job)
	excludedRegions := r.getExcludedRegions(job)
	maxLatency := r.getMaxLatency(job)

	for i, node := range nodes {
		rank, reason := r.rankNode(ctx, node, originRegion, preferredRegions, excludedRegions, maxLatency)
		ranks[i] = orchestrator.NodeRank{
			NodeInfo:  node,
			Rank:      rank,
			Reason:    reason,
			Retryable: true,
		}
		log.Ctx(ctx).Trace().Object("Rank", ranks[i]).Msg("Geo-ranked node")
	}

	return ranks, nil
}

// rankNode ranks a single node based on geographic criteria.
func (r *GeoRanker) rankNode(
	ctx context.Context,
	node models.NodeInfo,
	originRegion string,
	preferredRegions, excludedRegions map[string]bool,
	maxLatency time.Duration,
) (int, string) {
	// Check if node is in excluded region
	nodeRegion := r.getNodeRegion(node)
	if excludedRegions[nodeRegion] {
		return orchestrator.RankUnsuitable, fmt.Sprintf("region %s is excluded", nodeRegion)
	}

	// Start with base rank
	rank := orchestrator.RankPossible
	reasons := make([]string, 0)

	// Check preferred regions
	if len(preferredRegions) > 0 {
		if preferredRegions[nodeRegion] {
			rank += 30
			reasons = append(reasons, "preferred region")
		}
	}

	// Calculate latency
	latency := r.latencyMatrix.GetLatency(originRegion, nodeRegion)

	// Check max latency constraint
	if maxLatency > 0 && latency > maxLatency {
		if r.config.ExcludeHighLatency {
			return orchestrator.RankUnsuitable, fmt.Sprintf("latency %v exceeds max %v", latency, maxLatency)
		}
	}

	// Apply latency penalty
	if latency > 0 {
		// Penalty for every 50ms of latency
		latencyPenalty := int(latency/(50*time.Millisecond)) * r.config.LatencyPenalty
		rank -= latencyPenalty
	}

	// Boost for local region
	if r.config.PreferLocal && nodeRegion == originRegion {
		rank += r.config.LocalBoost
		reasons = append(reasons, "local region")
	}

	// Boost for same continent
	if r.config.PreferContinent {
		originContinent := RegionToContinent(originRegion)
		nodeContinent := RegionToContinent(nodeRegion)
		if originContinent == nodeContinent && originContinent != "unknown" {
			rank += r.config.ContinentBoost
			reasons = append(reasons, "same continent")
		}
	}

	// Build reason string
	reason := fmt.Sprintf("region=%s latency=%v", nodeRegion, latency)
	if len(reasons) > 0 {
		reason += " (" + joinReasons(reasons) + ")"
	}

	return rank, reason
}

// getNodeRegion extracts the region from a node.
func (r *GeoRanker) getNodeRegion(node models.NodeInfo) string {
	// Try to get location from node labels
	if loc := GetLocationFromNodeInfo(node); loc != nil {
		return loc.Region
	}

	// Default region
	return "default"
}

// getPreferredRegions extracts preferred regions from job constraints.
func (r *GeoRanker) getPreferredRegions(job models.Job) map[string]bool {
	regions := make(map[string]bool)

	// Check job constraints
	for _, constraint := range job.Constraints {
		if constraint.Key == "region" || constraint.Key == "preferred-region" {
			for _, val := range constraint.Values {
				regions[val] = true
			}
		}
	}

	// Check job labels
	if job.Labels != nil {
		if regionsStr, ok := job.Labels["preferred-regions"]; ok {
			for _, region := range splitRegions(regionsStr) {
				regions[region] = true
			}
		}
	}

	return regions
}

// getExcludedRegions extracts excluded regions from job constraints.
func (r *GeoRanker) getExcludedRegions(job models.Job) map[string]bool {
	regions := make(map[string]bool)

	// Check job constraints
	for _, constraint := range job.Constraints {
		if constraint.Key == "exclude-region" || constraint.Key == "excluded-region" {
			for _, val := range constraint.Values {
				regions[val] = true
			}
		}
	}

	// Check job labels
	if job.Labels != nil {
		if regionsStr, ok := job.Labels["exclude-regions"]; ok {
			for _, region := range splitRegions(regionsStr) {
				regions[region] = true
			}
		}
	}

	return regions
}

// getMaxLatency extracts max latency from job constraints.
func (r *GeoRanker) getMaxLatency(job models.Job) time.Duration {
	// Check job constraints
	for _, constraint := range job.Constraints {
		if constraint.Key == "max-latency" {
			if len(constraint.Values) > 0 {
				if lat, err := time.ParseDuration(constraint.Values[0]); err == nil {
					return lat
				}
			}
		}
	}

	// Check job labels
	if job.Labels != nil {
		if latStr, ok := job.Labels["max-latency"]; ok {
			if lat, err := time.ParseDuration(latStr); err == nil {
				return lat
			}
		}
	}

	return r.config.MaxLatency
}

// GeoAwareNodeSelector wraps a node selector with geographic awareness.
type GeoAwareNodeSelector struct {
	selector     orchestrator.NodeSelector
	geoRanker    *GeoRanker
	latencyMatrix LatencyMatrix
}

// NewGeoAwareNodeSelector creates a geo-aware node selector.
func NewGeoAwareNodeSelector(
	selector orchestrator.NodeSelector,
	latencyMatrix LatencyMatrix,
	config GeoRankerConfig,
) *GeoAwareNodeSelector {
	return &GeoAwareNodeSelector{
		selector:      selector,
		geoRanker:     NewGeoRanker(config, latencyMatrix),
		latencyMatrix: latencyMatrix,
	}
}

// SelectNodes selects nodes with geographic optimization.
func (s *GeoAwareNodeSelector) SelectNodes(
	ctx context.Context,
	job *models.Job,
	preferredRegions []string,
	maxLatency time.Duration,
	minNodes int,
) ([]orchestrator.NodeRank, error) {
	// Get all matching nodes from the underlying selector
	matched, rejected, err := s.selector.MatchingNodes(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching nodes: %w", err)
	}

	// Log rejected nodes
	if len(rejected) > 0 {
		log.Ctx(ctx).Debug().
			Int("rejected", len(rejected)).
			Msg("Nodes rejected by selector")
	}

	// Apply geographic ranking
	geoRanks, err := s.geoRanker.RankNodes(ctx, *job, extractNodeInfos(matched))
	if err != nil {
		return nil, fmt.Errorf("failed to geo-rank nodes: %w", err)
	}

	// Combine original ranks with geo ranks
	combinedRanks := s.combineRanks(matched, geoRanks)

	// Apply preferred regions filter
	if len(preferredRegions) > 0 {
		combinedRanks = s.applyPreferredRegions(combinedRanks, preferredRegions)
	}

	// Apply latency filter
	if maxLatency > 0 {
		combinedRanks = s.applyLatencyFilter(combinedRanks, maxLatency)
	}

	// Ensure minimum nodes
	if len(combinedRanks) < minNodes {
		log.Ctx(ctx).Warn().
			Int("available", len(combinedRanks)).
			Int("required", minNodes).
			Msg("Not enough nodes after geographic filtering")
	}

	return combinedRanks, nil
}

// combineRanks combines base ranks with geographic ranks.
func (s *GeoAwareNodeSelector) combineRanks(baseRanks, geoRanks []orchestrator.NodeRank) []orchestrator.NodeRank {
	// Create map of node ID to geo rank
	geoRankMap := make(map[string]orchestrator.NodeRank)
	for _, gr := range geoRanks {
		geoRankMap[gr.NodeInfo.ID()] = gr
	}

	// Combine ranks
	result := make([]orchestrator.NodeRank, len(baseRanks))
	for i, base := range baseRanks {
		result[i] = base

		if geo, ok := geoRankMap[base.NodeInfo.ID()]; ok {
			// Add geo rank to base rank
			result[i].Rank += geo.Rank

			// If geo rank marked as unsuitable, propagate
			if geo.Rank < orchestrator.RankPossible {
				result[i].Rank = geo.Rank
				result[i].Reason = geo.Reason
			} else {
				// Combine reasons
				result[i].Reason = base.Reason + "; " + geo.Reason
			}
		}
	}

	// Sort by rank descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Rank > result[j].Rank
	})

	return result
}

// applyPreferredRegions boosts nodes in preferred regions.
func (s *GeoAwareNodeSelector) applyPreferredRegions(ranks []orchestrator.NodeRank, preferred []string) []orchestrator.NodeRank {
	prefSet := make(map[string]bool)
	for _, r := range preferred {
		prefSet[r] = true
	}

	for i := range ranks {
		nodeRegion := s.geoRanker.getNodeRegion(ranks[i].NodeInfo)
		if prefSet[nodeRegion] {
			ranks[i].Rank += 30
			ranks[i].Reason += "; preferred region"
		}
	}

	return ranks
}

// applyLatencyFilter filters out nodes with high latency.
func (s *GeoAwareNodeSelector) applyLatencyFilter(ranks []orchestrator.NodeRank, maxLatency time.Duration) []orchestrator.NodeRank {
	filtered := make([]orchestrator.NodeRank, 0, len(ranks))

	for _, rank := range ranks {
		nodeRegion := s.geoRanker.getNodeRegion(rank.NodeInfo)
		latency := s.latencyMatrix.GetLatency(s.geoRanker.config.OriginRegion, nodeRegion)

		if latency <= maxLatency {
			filtered = append(filtered, rank)
		}
	}

	return filtered
}

// extractNodeInfos extracts NodeInfo from NodeRank slice.
func extractNodeInfos(ranks []orchestrator.NodeRank) []models.NodeInfo {
	infos := make([]models.NodeInfo, len(ranks))
	for i, r := range ranks {
		infos[i] = r.NodeInfo
	}
	return infos
}

// joinReasons joins reason strings.
func joinReasons(reasons []string) string {
	result := ""
	for i, r := range reasons {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}

// splitRegions splits a comma-separated region string.
func splitRegions(s string) []string {
	var regions []string
	for _, r := range splitByComma(s) {
		r = trimSpace(r)
		if r != "" {
			regions = append(regions, r)
		}
	}
	return regions
}

// splitByComma splits a string by comma.
func splitByComma(s string) []string {
	var result []string
	var current string
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// trimSpace trims whitespace from a string.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && isWhitespace(s[start]) {
		start++
	}
	for end > start && isWhitespace(s[end-1]) {
		end--
	}
	return s[start:end]
}

// isWhitespace checks if a byte is whitespace.
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
