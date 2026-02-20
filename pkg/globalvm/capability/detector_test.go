//go:build unit

package capability

import (
	"context"
	"testing"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	require.NotNil(t, detector)
	assert.NotNil(t, detector.gpuDetector)
	assert.NotNil(t, detector.engineDetector)
}

func TestDetectorWithOptions(t *testing.T) {
	customCache := 60 * time.Second
	detector := NewDetector(
		WithCacheExpiry(customCache),
	)

	require.NotNil(t, detector)
	assert.Equal(t, customCache, detector.cacheExpiry)
}

func TestDetectAll(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	caps, err := detector.DetectAll(ctx)
	require.NoError(t, err)
	require.NotNil(t, caps)

	// Basic sanity checks
	assert.NotZero(t, caps.DetectionTime)
	assert.NotEmpty(t, caps.OS)
	assert.NotEmpty(t, caps.Architecture)
}

func TestDetectAllCaching(t *testing.T) {
	detector := NewDetector(WithCacheExpiry(5 * time.Minute))
	ctx := context.Background()

	// First detection
	caps1, err := detector.DetectAll(ctx)
	require.NoError(t, err)

	// Second detection should return cached result
	caps2, err := detector.DetectAll(ctx)
	require.NoError(t, err)

	// Should be the same pointer (cached)
	assert.Same(t, caps1, caps2)
}

func TestRefresh(t *testing.T) {
	detector := NewDetector(WithCacheExpiry(5 * time.Minute))
	ctx := context.Background()

	// First detection
	caps1, err := detector.DetectAll(ctx)
	require.NoError(t, err)

	// Refresh should clear cache and detect again
	caps2, err := detector.Refresh(ctx)
	require.NoError(t, err)

	// Should be different objects (not cached)
	assert.NotSame(t, caps1, caps2)
}

func TestNodeCapabilitiesHasEngine(t *testing.T) {
	caps := &NodeCapabilities{
		Engines: []EngineCapability{
			{Type: models.EngineDocker, Available: true},
			{Type: models.EngineWasm, Available: true},
			{Type: "native", Available: false},
		},
	}

	assert.True(t, caps.HasEngine(models.EngineDocker))
	assert.True(t, caps.HasEngine(models.EngineWasm))
	assert.False(t, caps.HasEngine("native"))
	assert.False(t, caps.HasEngine("nonexistent"))
}

func TestNodeCapabilitiesGetEngine(t *testing.T) {
	caps := &NodeCapabilities{
		Engines: []EngineCapability{
			{Type: models.EngineDocker, Version: "24.0.0", Available: true},
		},
	}

	engine := caps.GetEngine(models.EngineDocker)
	require.NotNil(t, engine)
	assert.Equal(t, "24.0.0", engine.Version)

	engine = caps.GetEngine("nonexistent")
	assert.Nil(t, engine)
}

func TestNodeCapabilitiesHasGPUVendor(t *testing.T) {
	caps := &NodeCapabilities{
		GPUs: []GPUCapability{
			{Vendor: models.GPUVendorNvidia, Available: true},
			{Vendor: models.GPUVendorAMDATI, Available: false},
		},
	}

	assert.True(t, caps.HasGPUVendor(models.GPUVendorNvidia))
	assert.False(t, caps.HasGPUVendor(models.GPUVendorAMDATI))
	assert.False(t, caps.HasGPUVendor(models.GPUVendorIntel))
}

func TestNodeCapabilitiesTotalGPUMemory(t *testing.T) {
	caps := &NodeCapabilities{
		GPUs: []GPUCapability{
			{Memory: 8192, Available: true},
			{Memory: 16384, Available: true},
			{Memory: 4096, Available: false}, // Not available
		},
	}

	total := caps.TotalGPUMemory()
	assert.Equal(t, uint64(24576), total) // 8192 + 16384
}

func TestNodeCapabilitiesCapabilityScore(t *testing.T) {
	// Test with engines only
	caps := &NodeCapabilities{
		Engines: []EngineCapability{
			{Type: models.EngineDocker, Available: true},
			{Type: models.EngineWasm, Available: true},
		},
	}
	score := caps.CapabilityScore()
	assert.GreaterOrEqual(t, score, 100) // At least 50 per engine

	// Test with GPUs
	caps = &NodeCapabilities{
		Engines: []EngineCapability{
			{Type: models.EngineDocker, Available: true},
		},
		GPUs: []GPUCapability{
			{Memory: 16384, Available: true}, // 16GB GPU
		},
	}
	score = caps.CapabilityScore()
	assert.GreaterOrEqual(t, score, 200) // Engine + GPU + memory bonus

	// Test with benchmarks
	caps = &NodeCapabilities{
		Engines: []EngineCapability{
			{Type: models.EngineDocker, Available: true},
		},
		Benchmarks: &CapabilityBenchmarks{
			CPUScore:    500,
			MemoryScore: 500,
			DiskScore:   500,
		},
	}
	score = caps.CapabilityScore()
	assert.GreaterOrEqual(t, score, 200) // Engine + benchmark contributions
}

func TestToModelsGPUs(t *testing.T) {
	caps := []GPUCapability{
		{
			Index:      0,
			Name:       "Tesla T4",
			Vendor:     models.GPUVendorNvidia,
			Memory:     16384,
			PCIAddress: "0000:00:1e.0",
		},
		{
			Index:      1,
			Name:       "RTX 3080",
			Vendor:     models.GPUVendorNvidia,
			Memory:     10240,
			PCIAddress: "0000:01:00.0",
		},
	}

	gpus := ToModelsGPUs(caps)
	require.Len(t, gpus, 2)

	assert.Equal(t, uint64(0), gpus[0].Index)
	assert.Equal(t, "Tesla T4", gpus[0].Name)
	assert.Equal(t, models.GPUVendorNvidia, gpus[0].Vendor)
	assert.Equal(t, uint64(16384), gpus[0].Memory)
	assert.Equal(t, "0000:00:1e.0", gpus[0].PCIAddress)
}

func TestBenchmark(t *testing.T) {
	detector := NewDetector()
	ctx := context.Background()

	benchmarks, err := detector.Benchmark(ctx)
	require.NoError(t, err)
	require.NotNil(t, benchmarks)

	// Basic sanity checks
	assert.NotZero(t, benchmarks.BenchmarkTime)
	assert.Positive(t, benchmarks.BenchmarkDuration)
	assert.GreaterOrEqual(t, benchmarks.CPUScore, 0)
	assert.LessOrEqual(t, benchmarks.CPUScore, 1000)
	assert.GreaterOrEqual(t, benchmarks.MemoryScore, 0)
	assert.LessOrEqual(t, benchmarks.MemoryScore, 1000)
	assert.GreaterOrEqual(t, benchmarks.DiskScore, 0)
	assert.LessOrEqual(t, benchmarks.DiskScore, 1000)
	assert.GreaterOrEqual(t, benchmarks.NetworkScore, 0)
	assert.LessOrEqual(t, benchmarks.NetworkScore, 1000)
}

func TestBenchmarkCPU(t *testing.T) {
	score := benchmarkCPU()
	assert.GreaterOrEqual(t, score, 0)
	assert.LessOrEqual(t, score, 1000)
}

func TestBenchmarkMemory(t *testing.T) {
	score := benchmarkMemory()
	assert.GreaterOrEqual(t, score, 0)
	assert.LessOrEqual(t, score, 1000)
}

func TestBenchmarkDisk(t *testing.T) {
	score := benchmarkDisk()
	assert.GreaterOrEqual(t, score, 0)
	assert.LessOrEqual(t, score, 1000)
}

func TestBenchmarkGPU(t *testing.T) {
	tests := []struct {
		name string
		gpu  GPUCapability
	}{
		{
			name: "NVIDIA Tesla T4",
			gpu: GPUCapability{
				Vendor:            models.GPUVendorNvidia,
				Name:              "Tesla T4",
				Memory:            16384,
				ComputeCapability: "7.5",
				Available:         true,
			},
		},
		{
			name: "AMD MI210",
			gpu: GPUCapability{
				Vendor:    models.GPUVendorAMDATI,
				Name:      "Instinct MI210",
				Memory:    65536,
				Available: true,
			},
		},
		{
			name: "Intel Arc",
			gpu: GPUCapability{
				Vendor:    models.GPUVendorIntel,
				Name:      "Intel Arc A770",
				Memory:    16384,
				Available: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := benchmarkGPU(tt.gpu)
			assert.GreaterOrEqual(t, score, 0)
			assert.LessOrEqual(t, score, 1000)
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name string
		s    string
		subs []string
		want bool
	}{
		{"A100 match", "A100-SXM4-40GB", []string{"A100", "H100"}, true},
		{"Tesla T4 no match", "Tesla T4", []string{"A100", "H100"}, false},
		{"RTX 4090 match", "RTX 4090", []string{"RTX 40", "RTX 50"}, true},
		{"RTX 3080 match", "RTX 3080", []string{"RTX 30"}, true}, // "RTX 3080" contains "RTX 30"
		{"empty string", "", []string{"A100"}, false},
		{"empty subs", "A100", []string{}, false},
		{"V100 match", "Tesla V100", []string{"V100", "A100"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.s, tt.subs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEngineCapability(t *testing.T) {
	cap := EngineCapability{
		Type:        models.EngineDocker,
		Version:     "24.0.0",
		Available:   true,
		Constraints: []string{"rootless"},
		Features:    []string{"gpu-nvidia", "buildx"},
	}

	assert.Equal(t, models.EngineDocker, cap.Type)
	assert.True(t, cap.Available)
	assert.Len(t, cap.Constraints, 1)
	assert.Len(t, cap.Features, 2)
}

func TestGPUCapability(t *testing.T) {
	cap := GPUCapability{
		Index:             0,
		Name:              "Tesla T4",
		Vendor:            models.GPUVendorNvidia,
		Memory:            16384,
		PCIAddress:        "0000:00:1e.0",
		Driver:            "535.86.05",
		CUDA:              "12.2",
		ComputeCapability: "7.5",
		Available:         true,
		Temperature:       45,
		Utilization:       10,
	}

	assert.Equal(t, uint64(0), cap.Index)
	assert.Equal(t, "Tesla T4", cap.Name)
	assert.True(t, cap.Available)
}

func TestStorageCapability(t *testing.T) {
	cap := StorageCapability{
		Type:       "local",
		Path:       "/data",
		Total:      1 << 40, // 1TB
		Available:  800 << 30, // 800GB
		ReadOnly:   false,
		Network:    false,
	}

	assert.Equal(t, "local", cap.Type)
	assert.Equal(t, "/data", cap.Path)
	assert.False(t, cap.ReadOnly)
	assert.False(t, cap.Network)
}

func TestNetworkCapability(t *testing.T) {
	cap := NetworkCapability{
		PublicIP:   "203.0.113.50",
		PrivateIPs: []string{"10.0.0.5", "172.17.0.1"},
		Region:     "us-west-2",
		Zone:       "us-west-2a",
		IPv6:       true,
		NAT:        true,
	}

	assert.Equal(t, "203.0.113.50", cap.PublicIP)
	assert.Len(t, cap.PrivateIPs, 2)
	assert.True(t, cap.IPv6)
	assert.True(t, cap.NAT)
}

func TestCapabilityBenchmarks(t *testing.T) {
	bench := CapabilityBenchmarks{
		CPUScore:    500,
		MemoryScore: 600,
		DiskScore:   400,
		GPUScores:   map[uint64]int{0: 700, 1: 750},
		NetworkScore: 300,
	}

	assert.Equal(t, 500, bench.CPUScore)
	assert.Equal(t, 600, bench.MemoryScore)
	assert.Len(t, bench.GPUScores, 2)
}
