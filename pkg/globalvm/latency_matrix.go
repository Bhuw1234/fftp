//go:build unit

// Package globalvm provides global scheduling capabilities for the distributed compute network.
// This file implements latency tracking and geographic-aware node selection.
package globalvm

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/rs/zerolog/log"
)

// LatencyMatrix provides latency tracking between regions/nodes.
// It maintains a matrix of latencies and supports nearest node selection.
type LatencyMatrix interface {
	// GetLatency returns the latency between two regions or nodes.
	GetLatency(from, to string) time.Duration

	// UpdateLatency updates the latency between two regions.
	UpdateLatency(from, to string, latency time.Duration)

	// GetNearestNodes returns nodes sorted by proximity to a region.
	GetNearestNodes(region string, nodes []NodeSelection) []NodeSelection

	// ProbeLatency actively probes latency to a target.
	ProbeLatency(ctx context.Context, target string) (time.Duration, error)

	// GetAllLatencies returns all known latencies from a region.
	GetAllLatencies(from string) map[string]time.Duration

	// ClearCache clears the latency cache.
	ClearCache()
}

// LatencyProbe defines the interface for latency probing.
type LatencyProbe interface {
	// Probe measures the latency to a target endpoint.
	Probe(ctx context.Context, target string) (time.Duration, error)

	// ProbeWithPort measures latency to a specific port on a target.
	ProbeWithPort(ctx context.Context, target string, port int) (time.Duration, error)
}

// LatencyMatrixConfig configures the latency matrix.
type LatencyMatrixConfig struct {
	// ProbeTimeout is the timeout for individual probes.
	ProbeTimeout time.Duration

	// CacheExpiry is how long to cache latency results.
	CacheExpiry time.Duration

	// MaxProbes is the maximum concurrent probes.
	MaxProbes int

	// DefaultLatency is used when latency is unknown.
	DefaultLatency time.Duration
}

// DefaultLatencyMatrixConfig returns the default configuration.
func DefaultLatencyMatrixConfig() LatencyMatrixConfig {
	return LatencyMatrixConfig{
		ProbeTimeout:   5 * time.Second,
		CacheExpiry:    5 * time.Minute,
		MaxProbes:      10,
		DefaultLatency: 200 * time.Millisecond,
	}
}

// matrixEntry represents a cached latency entry.
type matrixEntry struct {
	latency    time.Duration
	measuredAt time.Time
	source     string // "probe", "reported", "estimated"
}

// latencyMatrix implements the LatencyMatrix interface.
type latencyMatrix struct {
	config LatencyMatrixConfig
	probe  LatencyProbe

	mu     sync.RWMutex
	matrix map[string]map[string]*matrixEntry // from -> to -> entry
}

// NewLatencyMatrix creates a new latency matrix.
func NewLatencyMatrix(config LatencyMatrixConfig) LatencyMatrix {
	return &latencyMatrix{
		config: config,
		probe:  NewHTTPLatencyProbe(config.ProbeTimeout),
		matrix: make(map[string]map[string]*matrixEntry),
	}
}

// NewLatencyMatrixWithProbe creates a latency matrix with a custom probe.
func NewLatencyMatrixWithProbe(config LatencyMatrixConfig, probe LatencyProbe) LatencyMatrix {
	return &latencyMatrix{
		config: config,
		probe:  probe,
		matrix: make(map[string]map[string]*matrixEntry),
	}
}

// GetLatency returns the latency between two regions.
func (m *latencyMatrix) GetLatency(from, to string) time.Duration {
	if from == to {
		return 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if fromMap, ok := m.matrix[from]; ok {
		if entry, ok := fromMap[to]; ok {
			// Check if entry is still valid
			if time.Since(entry.measuredAt) < m.config.CacheExpiry {
				return entry.latency
			}
		}
	}

	// Return default if unknown
	return m.config.DefaultLatency
}

// UpdateLatency updates the latency between two regions.
func (m *latencyMatrix) UpdateLatency(from, to string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.matrix[from] == nil {
		m.matrix[from] = make(map[string]*matrixEntry)
	}

	m.matrix[from][to] = &matrixEntry{
		latency:    latency,
		measuredAt: time.Now(),
		source:     "reported",
	}

	// Also update the reverse direction (symmetric assumption)
	if m.matrix[to] == nil {
		m.matrix[to] = make(map[string]*matrixEntry)
	}
	m.matrix[to][from] = &matrixEntry{
		latency:    latency,
		measuredAt: time.Now(),
		source:     "reported",
	}
}

// GetNearestNodes returns nodes sorted by proximity to a region.
func (m *latencyMatrix) GetNearestNodes(region string, nodes []NodeSelection) []NodeSelection {
	if len(nodes) == 0 {
		return nodes
	}

	// Create a copy to avoid modifying the original
	result := make([]NodeSelection, len(nodes))
	copy(result, nodes)

	// Get all latencies from the region
	latencies := m.GetAllLatencies(region)

	// Sort by latency (ascending)
	for i := range result {
		if lat, ok := latencies[result[i].Region]; ok {
			result[i].EstimatedLatency = lat
		} else {
			result[i].EstimatedLatency = m.config.DefaultLatency
		}
	}

	// Simple insertion sort by latency
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].EstimatedLatency < result[j-1].EstimatedLatency; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	return result
}

// ProbeLatency actively probes latency to a target.
func (m *latencyMatrix) ProbeLatency(ctx context.Context, target string) (time.Duration, error) {
	latency, err := m.probe.Probe(ctx, target)
	if err != nil {
		return m.config.DefaultLatency, err
	}

	return latency, nil
}

// GetAllLatencies returns all known latencies from a region.
func (m *latencyMatrix) GetAllLatencies(from string) map[string]time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]time.Duration)

	if fromMap, ok := m.matrix[from]; ok {
		for to, entry := range fromMap {
			if time.Since(entry.measuredAt) < m.config.CacheExpiry {
				result[to] = entry.latency
			}
		}
	}

	return result
}

// ClearCache clears the latency cache.
func (m *latencyMatrix) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matrix = make(map[string]map[string]*matrixEntry)
}

// httpLatencyProbe implements LatencyProbe using HTTP.
type httpLatencyProbe struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPLatencyProbe creates a new HTTP-based latency probe.
func NewHTTPLatencyProbe(timeout time.Duration) LatencyProbe {
	return &httpLatencyProbe{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DisableKeepAlives:  true,
				DisableCompression: true,
			},
		},
		timeout: timeout,
	}
}

// Probe measures latency to a target using HTTP HEAD request.
func (p *httpLatencyProbe) Probe(ctx context.Context, target string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	log.Ctx(ctx).Trace().
		Str("target", target).
		Dur("latency", latency).
		Int("status", resp.StatusCode).
		Msg("Latency probe completed")

	return latency, nil
}

// ProbeWithPort measures latency to a specific port.
func (p *httpLatencyProbe) ProbeWithPort(ctx context.Context, target string, port int) (time.Duration, error) {
	url := target
	if port > 0 {
		// Simple URL construction - assumes target is a host
		url = "http://" + target
		if port != 80 {
			url = "http://" + target + ":" + string(rune(port))
		}
	}
	return p.Probe(ctx, url)
}

// EstimatedLatency estimates latency between regions based on geographic distance.
// This provides a fallback when actual measurements aren't available.
func EstimatedLatency(fromRegion, toRegion string) time.Duration {
	// Predefined inter-region latencies (approximate values in milliseconds)
	// Based on typical public cloud inter-region latencies
	regionLatencies := map[string]map[string]time.Duration{
		"us-east": {
			"us-west":     65 * time.Millisecond,
			"eu-west":     85 * time.Millisecond,
			"eu-central":  95 * time.Millisecond,
			"asia-east":   200 * time.Millisecond,
			"asia-south":  210 * time.Millisecond,
			"south-america": 130 * time.Millisecond,
		},
		"us-west": {
			"us-east":     65 * time.Millisecond,
			"eu-west":     140 * time.Millisecond,
			"asia-east":   140 * time.Millisecond,
			"asia-south":  220 * time.Millisecond,
		},
		"eu-west": {
			"us-east":     85 * time.Millisecond,
			"us-west":     140 * time.Millisecond,
			"eu-central":  20 * time.Millisecond,
			"asia-east":   200 * time.Millisecond,
		},
		"eu-central": {
			"us-east":     95 * time.Millisecond,
			"eu-west":     20 * time.Millisecond,
			"asia-east":   180 * time.Millisecond,
		},
		"asia-east": {
			"us-east":     200 * time.Millisecond,
			"us-west":     140 * time.Millisecond,
			"eu-west":     200 * time.Millisecond,
			"asia-south":  60 * time.Millisecond,
		},
		"asia-south": {
			"us-east":     210 * time.Millisecond,
			"us-west":     220 * time.Millisecond,
			"asia-east":   60 * time.Millisecond,
		},
	}

	if fromRegion == toRegion {
		return 5 * time.Millisecond // Same region
	}

	if fromMap, ok := regionLatencies[fromRegion]; ok {
		if lat, ok := fromMap[toRegion]; ok {
			return lat
		}
	}

	// Default latency for unknown region pairs
	return 200 * time.Millisecond
}

// LatencyAwareNodeInfo wraps NodeInfo with latency information.
type LatencyAwareNodeInfo struct {
	models.NodeInfo
	Latency time.Duration
	Region  string
}

// ByLatency implements sort.Interface for LatencyAwareNodeInfo.
type ByLatency []LatencyAwareNodeInfo

func (a ByLatency) Len() int           { return len(a) }
func (a ByLatency) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLatency) Less(i, j int) bool { return a[i].Latency < a[j].Latency }
