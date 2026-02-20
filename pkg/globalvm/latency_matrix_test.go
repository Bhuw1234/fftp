//go:build unit

package globalvm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatencyMatrix_GetLatency(t *testing.T) {
	t.Run("returns zero for same region", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
		latency := matrix.GetLatency("us-east", "us-east")
		assert.Equal(t, time.Duration(0), latency)
	})

	t.Run("returns default for unknown region", func(t *testing.T) {
		config := DefaultLatencyMatrixConfig()
		config.DefaultLatency = 200 * time.Millisecond
		matrix := NewLatencyMatrix(config)
		latency := matrix.GetLatency("unknown-region", "another-unknown")
		assert.Equal(t, 200*time.Millisecond, latency)
	})

	t.Run("returns cached latency", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
		matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
		latency := matrix.GetLatency("us-east", "eu-west")
		assert.Equal(t, 85*time.Millisecond, latency)
	})

	t.Run("returns symmetric latency", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
		matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)

		// Check both directions
		latency1 := matrix.GetLatency("us-east", "eu-west")
		latency2 := matrix.GetLatency("eu-west", "us-east")
		assert.Equal(t, latency1, latency2)
	})
}

func TestLatencyMatrix_UpdateLatency(t *testing.T) {
	t.Run("updates latency for both directions", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

		matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

		// Check forward direction
		latency := matrix.GetLatency("us-east", "asia-east")
		assert.Equal(t, 200*time.Millisecond, latency)

		// Check reverse direction
		latency = matrix.GetLatency("asia-east", "us-east")
		assert.Equal(t, 200*time.Millisecond, latency)
	})

	t.Run("overwrites existing latency", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

		matrix.UpdateLatency("us-east", "eu-west", 100*time.Millisecond)
		matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)

		latency := matrix.GetLatency("us-east", "eu-west")
		assert.Equal(t, 85*time.Millisecond, latency)
	})
}

func TestLatencyMatrix_GetNearestNodes(t *testing.T) {
	t.Run("returns nodes sorted by latency", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

		// Set up latencies
		matrix.UpdateLatency("us-east", "us-west", 65*time.Millisecond)
		matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
		matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

		nodes := []NodeSelection{
			{NodeID: "node-asia", Region: "asia-east"},
			{NodeID: "node-eu", Region: "eu-west"},
			{NodeID: "node-us-west", Region: "us-west"},
		}

		nearest := matrix.GetNearestNodes("us-east", nodes)

		// Should be sorted by latency ascending
		require.Len(t, nearest, 3)
		assert.Equal(t, "node-us-west", nearest[0].NodeID)  // 65ms
		assert.Equal(t, "node-eu", nearest[1].NodeID)       // 85ms
		assert.Equal(t, "node-asia", nearest[2].NodeID)     // 200ms
	})

	t.Run("handles empty node list", func(t *testing.T) {
		matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
		nodes := []NodeSelection{}

		nearest := matrix.GetNearestNodes("us-east", nodes)
		assert.Empty(t, nearest)
	})

	t.Run("uses default latency for unknown regions", func(t *testing.T) {
		config := DefaultLatencyMatrixConfig()
		config.DefaultLatency = 300 * time.Millisecond
		matrix := NewLatencyMatrix(config)

		nodes := []NodeSelection{
			{NodeID: "node-known", Region: "us-west"},
			{NodeID: "node-unknown", Region: "unknown-region"},
		}

		matrix.UpdateLatency("us-east", "us-west", 65*time.Millisecond)

		nearest := matrix.GetNearestNodes("us-east", nodes)

		require.Len(t, nearest, 2)
		assert.Equal(t, "node-known", nearest[0].NodeID)    // 65ms
		assert.Equal(t, "node-unknown", nearest[1].NodeID)  // 300ms (default)
	})
}

func TestLatencyMatrix_ClearCache(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

	matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
	latency := matrix.GetLatency("us-east", "eu-west")
	assert.Equal(t, 85*time.Millisecond, latency)

	matrix.ClearCache()

	// After clear, should return default
	config := DefaultLatencyMatrixConfig()
	latency = matrix.GetLatency("us-east", "eu-west")
	assert.Equal(t, config.DefaultLatency, latency)
}

func TestLatencyMatrix_GetAllLatencies(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

	matrix.UpdateLatency("us-east", "us-west", 65*time.Millisecond)
	matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
	matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

	latencies := matrix.GetAllLatencies("us-east")

	assert.Len(t, latencies, 3)
	assert.Equal(t, 65*time.Millisecond, latencies["us-west"])
	assert.Equal(t, 85*time.Millisecond, latencies["eu-west"])
	assert.Equal(t, 200*time.Millisecond, latencies["asia-east"])
}

func TestHTTPLatencyProbe(t *testing.T) {
	t.Run("measures latency to server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // Simulate some delay
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		probe := NewHTTPLatencyProbe(5 * time.Second)
		latency, err := probe.Probe(context.Background(), server.URL)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, latency, 10*time.Millisecond)
	})

	t.Run("returns error for unreachable server", func(t *testing.T) {
		probe := NewHTTPLatencyProbe(1 * time.Second)
		_, err := probe.Probe(context.Background(), "http://localhost:99999")

		assert.Error(t, err)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second) // Long delay
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		probe := NewHTTPLatencyProbe(1 * time.Second)
		_, err := probe.Probe(ctx, server.URL)

		assert.Error(t, err)
	})
}

func TestEstimatedLatency(t *testing.T) {
	t.Run("returns zero for same region", func(t *testing.T) {
		latency := EstimatedLatency("us-east", "us-east")
		assert.Equal(t, 5*time.Millisecond, latency)
	})

	t.Run("returns estimated latency for known regions", func(t *testing.T) {
		latency := EstimatedLatency("us-east", "eu-west")
		assert.Equal(t, 85*time.Millisecond, latency)

		latency = EstimatedLatency("us-east", "asia-east")
		assert.Equal(t, 200*time.Millisecond, latency)
	})

	t.Run("returns default for unknown region pairs", func(t *testing.T) {
		latency := EstimatedLatency("unknown", "also-unknown")
		assert.Equal(t, 200*time.Millisecond, latency)
	})
}

func TestGeoRanker_RankNodes(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
	matrix.UpdateLatency("us-east", "us-west", 65*time.Millisecond)
	matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
	matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

	config := DefaultGeoRankerConfig()
	config.OriginRegion = "us-east"
	ranker := NewGeoRanker(config, matrix)

	nodes := []models.NodeInfo{
		createTestNode("node-us-east", "us-east"),
		createTestNode("node-us-west", "us-west"),
		createTestNode("node-eu-west", "eu-west"),
		createTestNode("node-asia-east", "asia-east"),
	}

	job := models.Job{
		ID: "test-job-1",
	}

	ranks, err := ranker.RankNodes(context.Background(), job, nodes)
	require.NoError(t, err)
	require.Len(t, ranks, 4)

	// Local region should have highest rank
	assert.Equal(t, "node-us-east", ranks[0].NodeInfo.ID())
	assert.GreaterOrEqual(t, ranks[0].Rank, orchestrator.RankPossible+config.LocalBoost)

	// Asia should have lowest rank (highest latency)
	assert.Equal(t, "node-asia-east", ranks[3].NodeInfo.ID())
}

func TestGeoRanker_PreferredRegions(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
	matrix.UpdateLatency("us-east", "eu-west", 85*time.Millisecond)
	matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

	config := DefaultGeoRankerConfig()
	config.OriginRegion = "us-east"
	ranker := NewGeoRanker(config, matrix)

	nodes := []models.NodeInfo{
		createTestNode("node-eu", "eu-west"),
		createTestNode("node-asia", "asia-east"),
	}

	// Job with preferred region
	job := models.Job{
		ID: "test-job-2",
		Labels: map[string]string{
			"preferred-regions": "asia-east",
		},
	}

	ranks, err := ranker.RankNodes(context.Background(), job, nodes)
	require.NoError(t, err)

	// Verify the ranks are correct
	// Asia: base(0) + preferred(30) - latency_penalty(4*5=20) = +10
	// EU: base(0) + preferred(0) - latency_penalty(1*5=5) = -5
	var asiaRank, euRank orchestrator.NodeRank
	for _, r := range ranks {
		if r.NodeInfo.ID() == "node-asia" {
			asiaRank = r
		} else if r.NodeInfo.ID() == "node-eu" {
			euRank = r
		}
	}

	assert.Equal(t, 10, asiaRank.Rank, "Asia node should have rank 10")
	assert.Equal(t, -5, euRank.Rank, "EU node should have rank -5")
	assert.Contains(t, asiaRank.Reason, "preferred region", "Asia node reason should mention preferred region")
	assert.Greater(t, asiaRank.Rank, euRank.Rank, "Asia node should rank higher than EU node")
}

func TestGeoRanker_ExcludedRegions(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())

	config := DefaultGeoRankerConfig()
	config.OriginRegion = "us-east"
	ranker := NewGeoRanker(config, matrix)

	nodes := []models.NodeInfo{
		createTestNode("node-us-west", "us-west"),
		createTestNode("node-eu-west", "eu-west"),
	}

	// Job with excluded region
	job := models.Job{
		ID: "test-job-3",
		Labels: map[string]string{
			"exclude-regions": "eu-west",
		},
	}

	ranks, err := ranker.RankNodes(context.Background(), job, nodes)
	require.NoError(t, err)

	// EU node should be unsuitable
	for _, rank := range ranks {
		if rank.NodeInfo.ID() == "node-eu-west" {
			assert.Equal(t, orchestrator.RankUnsuitable, rank.Rank)
		} else {
			assert.GreaterOrEqual(t, rank.Rank, orchestrator.RankPossible)
		}
	}
}

func TestGeoRanker_MaxLatencyConstraint(t *testing.T) {
	matrix := NewLatencyMatrix(DefaultLatencyMatrixConfig())
	matrix.UpdateLatency("us-east", "us-west", 65*time.Millisecond)
	matrix.UpdateLatency("us-east", "asia-east", 200*time.Millisecond)

	config := DefaultGeoRankerConfig()
	config.OriginRegion = "us-east"
	config.ExcludeHighLatency = true
	config.MaxLatency = 100 * time.Millisecond
	ranker := NewGeoRanker(config, matrix)

	nodes := []models.NodeInfo{
		createTestNode("node-us-west", "us-west"),
		createTestNode("node-asia-east", "asia-east"),
	}

	job := models.Job{
		ID: "test-job-4",
	}

	ranks, err := ranker.RankNodes(context.Background(), job, nodes)
	require.NoError(t, err)

	// Asia node should be excluded due to high latency
	for _, rank := range ranks {
		if rank.NodeInfo.ID() == "node-asia-east" {
			assert.Equal(t, orchestrator.RankUnsuitable, rank.Rank)
		}
	}
}

func TestLocationDetector_GetRegion(t *testing.T) {
	detector := NewLocationDetector(DefaultLocationDetectorConfig())

	t.Run("returns existing region", func(t *testing.T) {
		loc := &Location{Region: "us-west"}
		region := detector.GetRegion(loc)
		assert.Equal(t, "us-west", region)
	})

	t.Run("returns cloud region if set", func(t *testing.T) {
		loc := &Location{CloudRegion: "eu-central"}
		region := detector.GetRegion(loc)
		assert.Equal(t, "eu-central", region)
	})

	t.Run("maps country to region", func(t *testing.T) {
		loc := &Location{Country: "JP"}
		region := detector.GetRegion(loc)
		assert.Equal(t, "asia-east", region)
	})

	t.Run("returns default for unknown", func(t *testing.T) {
		loc := &Location{}
		region := detector.GetRegion(loc)
		assert.Equal(t, "default", region)
	})
}

func TestGetLocationFromNodeInfo(t *testing.T) {
	t.Run("extracts region from labels", func(t *testing.T) {
		node := models.NodeInfo{
			NodeID: "test-node",
			Labels: map[string]string{
				"region": "us-east",
				"zone":   "us-east-1a",
			},
		}

		loc := GetLocationFromNodeInfo(node)
		require.NotNil(t, loc)
		assert.Equal(t, "us-east", loc.Region)
		assert.Equal(t, "us-east-1a", loc.Zone)
	})

	t.Run("extracts from kubernetes labels", func(t *testing.T) {
		node := models.NodeInfo{
			NodeID: "test-node",
			Labels: map[string]string{
				"topology.kubernetes.io/region": "eu-west",
				"topology.kubernetes.io/zone":   "eu-west-1a",
			},
		}

		loc := GetLocationFromNodeInfo(node)
		require.NotNil(t, loc)
		assert.Equal(t, "eu-west", loc.Region)
		assert.Equal(t, "eu-west-1a", loc.Zone)
	})

	t.Run("returns nil for no labels", func(t *testing.T) {
		node := models.NodeInfo{
			NodeID: "test-node",
		}

		loc := GetLocationFromNodeInfo(node)
		assert.Nil(t, loc)
	})
}

func TestRegionToContinent(t *testing.T) {
	tests := []struct {
		region    string
		continent string
	}{
		{"us-east", "north-america"},
		{"us-west", "north-america"},
		{"eu-west", "europe"},
		{"eu-central", "europe"},
		{"asia-east", "asia"},
		{"asia-south", "asia"},
		{"south-america", "south-america"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			continent := RegionToContinent(tt.region)
			assert.Equal(t, tt.continent, continent)
		})
	}
}

// Helper function to create test nodes
func createTestNode(id, region string) models.NodeInfo {
	return models.NodeInfo{
		NodeID: id,
		Labels: map[string]string{
			"region": region,
		},
		ComputeNodeInfo: models.ComputeNodeInfo{
			AvailableCapacity: models.Resources{
				CPU: 4,
				Memory: 8 * 1024 * 1024 * 1024, // 8GB
			},
		},
	}
}
