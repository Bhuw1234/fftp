// Package capability provides hardware and software capability detection for the Global VM.
// It detects GPUs, execution engines, storage, and network capabilities.
package capability

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/rs/zerolog/log"
)

// CapabilityDetector detects node capabilities including GPUs, engines, storage, and network.
type CapabilityDetector interface {
	// DetectAll detects all capabilities on the node.
	DetectAll(ctx context.Context) (*NodeCapabilities, error)

	// Benchmark runs performance benchmarks on detected capabilities.
	Benchmark(ctx context.Context) (*CapabilityBenchmarks, error)

	// Refresh redetects capabilities (useful for hot-plug hardware).
	Refresh(ctx context.Context) (*NodeCapabilities, error)
}

// NodeCapabilities represents all detected capabilities of a node.
type NodeCapabilities struct {
	// Engines are the available execution engines.
	Engines []EngineCapability `json:"Engines,omitempty"`

	// GPUs are the detected GPU devices.
	GPUs []GPUCapability `json:"GPUs,omitempty"`

	// Storage contains storage capabilities.
	Storage []StorageCapability `json:"Storage,omitempty"`

	// Network contains network capabilities.
	Network NetworkCapability `json:"Network"`

	// Benchmarks contains performance benchmarks (optional).
	Benchmarks *CapabilityBenchmarks `json:"Benchmarks,omitempty"`

	// DetectionTime is when these capabilities were detected.
	DetectionTime time.Time `json:"DetectionTime"`

	// Hostname is the detected hostname.
	Hostname string `json:"Hostname,omitempty"`

	// OS is the operating system.
	OS string `json:"OS,omitempty"`

	// Architecture is the CPU architecture.
	Architecture string `json:"Architecture,omitempty"`
}

// EngineCapability represents an execution engine capability.
type EngineCapability struct {
	// Type is the engine type (docker, wasm, native).
	Type string `json:"Type"`

	// Version is the engine version.
	Version string `json:"Version,omitempty"`

	// Available indicates if the engine is ready to use.
	Available bool `json:"Available"`

	// Constraints are any limitations or constraints.
	Constraints []string `json:"Constraints,omitempty"`

	// Features are supported features.
	Features []string `json:"Features,omitempty"`
}

// GPUCapability represents a GPU device capability.
type GPUCapability struct {
	// Index is the device index.
	Index uint64 `json:"Index"`

	// Name is the GPU model name.
	Name string `json:"Name"`

	// Vendor is the GPU vendor (NVIDIA, AMD, Intel).
	Vendor models.GPUVendor `json:"Vendor"`

	// Memory is total GPU memory in MiB.
	Memory uint64 `json:"Memory"`

	// PCIAddress is the PCI bus address.
	PCIAddress string `json:"PCIAddress,omitempty"`

	// Driver is the driver version.
	Driver string `json:"Driver,omitempty"`

	// CUDA is the CUDA version (NVIDIA only).
	CUDA string `json:"CUDA,omitempty"`

	// ROCm is the ROCm version (AMD only).
	ROCm string `json:"ROCm,omitempty"`

	// OpenCL is the OpenCL version if supported.
	OpenCL string `json:"OpenCL,omitempty"`

	// Vulkan is the Vulkan version if supported.
	Vulkan string `json:"Vulkan,omitempty"`

	// ComputeCapability is the compute capability (e.g., "8.6" for NVIDIA).
	ComputeCapability string `json:"ComputeCapability,omitempty"`

	// Available indicates if the GPU is available for use.
	Available bool `json:"Available"`

	// Temperature is the current temperature in Celsius.
	Temperature int `json:"Temperature,omitempty"`

	// Utilization is the current GPU utilization percentage.
	Utilization int `json:"Utilization,omitempty"`
}

// StorageCapability represents a storage capability.
type StorageCapability struct {
	// Type is the storage type (local, nfs, s3, etc).
	Type string `json:"Type"`

	// Path is the mount point or identifier.
	Path string `json:"Path,omitempty"`

	// Total is total capacity in bytes.
	Total uint64 `json:"Total"`

	// Available is available capacity in bytes.
	Available uint64 `json:"Available"`

	// ReadOnly indicates if the storage is read-only.
	ReadOnly bool `json:"ReadOnly,omitempty"`

	// Network indicates if this is network-attached storage.
	Network bool `json:"Network,omitempty"`

	// Latency is the estimated latency in microseconds.
	Latency int64 `json:"Latency,omitempty"`

	// Bandwidth is the estimated bandwidth in MB/s.
	Bandwidth int64 `json:"Bandwidth,omitempty"`
}

// NetworkCapability represents network capabilities.
type NetworkCapability struct {
	// PublicIP is the detected public IP address.
	PublicIP string `json:"PublicIP,omitempty"`

	// PrivateIPs are detected private IP addresses.
	PrivateIPs []string `json:"PrivateIPs,omitempty"`

	// Bandwidth is the estimated bandwidth in Mbps.
	Bandwidth int64 `json:"Bandwidth,omitempty"`

	// Latency is the estimated network latency.
	Latency time.Duration `json:"Latency,omitempty"`

	// Region is the detected geographic region.
	Region string `json:"Region,omitempty"`

	// Zone is the detected availability zone.
	Zone string `json:"Zone,omitempty"`

	// ISP is the internet service provider.
	ISP string `json:"ISP,omitempty"`

	// NAT indicates if behind NAT.
	NAT bool `json:"NAT,omitempty"`

	// IPv6 indicates IPv6 support.
	IPv6 bool `json:"IPv6,omitempty"`
}

// CapabilityBenchmarks contains performance benchmark results.
type CapabilityBenchmarks struct {
	// CPUScore is a normalized CPU performance score (0-1000).
	CPUScore int `json:"CPUScore"`

	// MemoryScore is a normalized memory performance score (0-1000).
	MemoryScore int `json:"MemoryScore"`

	// DiskScore is a normalized disk I/O score (0-1000).
	DiskScore int `json:"DiskScore"`

	// GPUScores are per-GPU performance scores.
	GPUScores map[uint64]int `json:"GPUScores,omitempty"`

	// NetworkScore is a normalized network performance score (0-1000).
	NetworkScore int `json:"NetworkScore"`

	// BenchmarkTime is when benchmarks were run.
	BenchmarkTime time.Time `json:"BenchmarkTime"`

	// BenchmarkDuration is how long benchmarks took.
	BenchmarkDuration time.Duration `json:"BenchmarkDuration"`
}

// Detector implements the CapabilityDetector interface.
type Detector struct {
	gpuDetector    GPUDetector
	engineDetector EngineDetector

	mu           sync.RWMutex
	lastDetect   *NodeCapabilities
	cacheExpiry  time.Duration
}

// DetectorOption configures the detector.
type DetectorOption func(*Detector)

// NewDetector creates a new capability detector.
func NewDetector(opts ...DetectorOption) *Detector {
	d := &Detector{
		gpuDetector:    NewDefaultGPUDetector(),
		engineDetector: NewDefaultEngineDetector(),
		cacheExpiry:    30 * time.Second,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// WithGPUDetector sets a custom GPU detector.
func WithGPUDetector(detector GPUDetector) DetectorOption {
	return func(d *Detector) {
		d.gpuDetector = detector
	}
}

// WithEngineDetector sets a custom engine detector.
func WithEngineDetector(detector EngineDetector) DetectorOption {
	return func(d *Detector) {
		d.engineDetector = detector
	}
}

// WithCacheExpiry sets the cache expiry duration.
func WithCacheExpiry(dur time.Duration) DetectorOption {
	return func(d *Detector) {
		d.cacheExpiry = dur
	}
}

// DetectAll detects all capabilities on the node.
func (d *Detector) DetectAll(ctx context.Context) (*NodeCapabilities, error) {
	// Check cache
	d.mu.RLock()
	if d.lastDetect != nil && time.Since(d.lastDetect.DetectionTime) < d.cacheExpiry {
		defer d.mu.RUnlock()
		return d.lastDetect, nil
	}
	d.mu.RUnlock()

	log.Ctx(ctx).Debug().Msg("Detecting node capabilities")

	caps := &NodeCapabilities{
		DetectionTime: time.Now(),
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
	}

	// Detect hostname
	hostname, err := detectHostname()
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to detect hostname")
	} else {
		caps.Hostname = hostname
	}

	// Detect engines
	engines, err := d.engineDetector.DetectEngines(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to detect engines")
	} else {
		caps.Engines = engines
	}

	// Detect GPUs
	gpus, err := d.gpuDetector.DetectGPUs(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to detect GPUs")
	} else {
		caps.GPUs = gpus
	}

	// Detect storage
	storage, err := detectStorage(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to detect storage")
	} else {
		caps.Storage = storage
	}

	// Detect network
	network, err := detectNetwork(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to detect network")
	} else {
		caps.Network = network
	}

	// Update cache
	d.mu.Lock()
	d.lastDetect = caps
	d.mu.Unlock()

	return caps, nil
}

// Benchmark runs performance benchmarks on detected capabilities.
func (d *Detector) Benchmark(ctx context.Context) (*CapabilityBenchmarks, error) {
	log.Ctx(ctx).Info().Msg("Running capability benchmarks")

	start := time.Now()
	benchmarks := &CapabilityBenchmarks{
		GPUScores:     make(map[uint64]int),
		BenchmarkTime: start,
	}

	// Run CPU benchmark
	benchmarks.CPUScore = benchmarkCPU()

	// Run memory benchmark
	benchmarks.MemoryScore = benchmarkMemory()

	// Run disk benchmark
	benchmarks.DiskScore = benchmarkDisk()

	// Run GPU benchmarks
	gpus, err := d.gpuDetector.DetectGPUs(ctx)
	if err == nil {
		for _, gpu := range gpus {
			benchmarks.GPUScores[gpu.Index] = benchmarkGPU(gpu)
		}
	}

	// Run network benchmark
	benchmarks.NetworkScore = benchmarkNetwork()

	benchmarks.BenchmarkDuration = time.Since(start)

	return benchmarks, nil
}

// Refresh redetects capabilities (useful for hot-plug hardware).
func (d *Detector) Refresh(ctx context.Context) (*NodeCapabilities, error) {
	d.mu.Lock()
	d.lastDetect = nil
	d.mu.Unlock()

	return d.DetectAll(ctx)
}

// ToModelsGPUs converts GPUCapability slice to models.GPU slice.
func ToModelsGPUs(caps []GPUCapability) []models.GPU {
	gpus := make([]models.GPU, len(caps))
	for i, cap := range caps {
		gpus[i] = models.GPU{
			Index:      cap.Index,
			Name:       cap.Name,
			Vendor:     cap.Vendor,
			Memory:     cap.Memory,
			PCIAddress: cap.PCIAddress,
		}
	}
	return gpus
}

// HasGPUVendor checks if any GPU of the specified vendor is available.
func (c *NodeCapabilities) HasGPUVendor(vendor models.GPUVendor) bool {
	for _, gpu := range c.GPUs {
		if gpu.Vendor == vendor && gpu.Available {
			return true
		}
	}
	return false
}

// HasEngine checks if a specific engine type is available.
func (c *NodeCapabilities) HasEngine(engineType string) bool {
	for _, engine := range c.Engines {
		if engine.Type == engineType && engine.Available {
			return true
		}
	}
	return false
}

// GetEngine returns a specific engine capability.
func (c *NodeCapabilities) GetEngine(engineType string) *EngineCapability {
	for i := range c.Engines {
		if c.Engines[i].Type == engineType {
			return &c.Engines[i]
		}
	}
	return nil
}

// TotalGPUMemory returns the total GPU memory across all GPUs.
func (c *NodeCapabilities) TotalGPUMemory() uint64 {
	var total uint64
	for _, gpu := range c.GPUs {
		if gpu.Available {
			total += gpu.Memory
		}
	}
	return total
}

// CapabilityScore returns an overall capability score (0-1000).
func (c *NodeCapabilities) CapabilityScore() int {
	score := 0

	// Base score for having engines
	for _, engine := range c.Engines {
		if engine.Available {
			score += 50
		}
	}

	// Score for GPUs
	for _, gpu := range c.GPUs {
		if gpu.Available {
			score += 100
			// Bonus for memory
			if gpu.Memory >= 16384 { // 16GB+
				score += 50
			}
		}
	}

	// Add benchmark scores if available
	if c.Benchmarks != nil {
		score += c.Benchmarks.CPUScore / 10
		score += c.Benchmarks.MemoryScore / 10
		score += c.Benchmarks.DiskScore / 10
	}

	// Cap at 1000
	if score > 1000 {
		score = 1000
	}

	return score
}
