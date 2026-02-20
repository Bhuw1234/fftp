package capability

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// benchmarkCPU runs a CPU performance benchmark.
func benchmarkCPU() int {
	// Simple CPU benchmark: count how many iterations we can do in 100ms
	iterations := int64(0)
	duration := 100 * time.Millisecond

	start := time.Now()
	done := make(chan bool)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localIter := int64(0)
			for {
				select {
				case <-done:
					atomic.AddInt64(&iterations, localIter)
					return
				default:
					// Simple compute operation
					localIter++
					_ = localIter * localIter
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	elapsed := time.Since(start)
	iterPerSec := float64(iterations) / elapsed.Seconds()

	// Normalize to 0-1000 scale
	// Typical range: 10M to 1B iterations per second
	score := int(iterPerSec / 1e6)
	if score > 1000 {
		score = 1000
	}
	if score < 0 {
		score = 0
	}

	return score
}

// benchmarkMemory runs a memory performance benchmark.
func benchmarkMemory() int {
	// Simple memory benchmark: measure allocation speed
	size := 1024 * 1024 // 1MB
	iterations := 100

	start := time.Now()
	for i := 0; i < iterations; i++ {
		buf := make([]byte, size)
		// Touch all pages
		for j := 0; j < size; j += 4096 {
			buf[j] = byte(j)
		}
		_ = buf
	}
	elapsed := time.Since(start)

	// Calculate MB/s
	mbPerSec := float64(iterations) / elapsed.Seconds()

	// Normalize to 0-1000 scale
	// Typical: 1-10 GB/s = 1000-10000 MB/s
	score := int(mbPerSec / 10)
	if score > 1000 {
		score = 1000
	}
	if score < 0 {
		score = 0
	}

	return score
}

// benchmarkDisk runs a disk I/O benchmark.
func benchmarkDisk() int {
	// Create a temp file for benchmarking
	tmpFile := "/tmp/capability_bench_tmp"
	defer os.Remove(tmpFile)

	size := 1024 * 1024 // 1MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Write benchmark
	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			return 0
		}
	}
	writeElapsed := time.Since(start)

	// Read benchmark
	start = time.Now()
	for i := 0; i < 10; i++ {
		if _, err := os.ReadFile(tmpFile); err != nil {
			return 0
		}
	}
	readElapsed := time.Since(start)

	// Calculate combined score
	// 10 MB in writeElapsed, 10 MB in readElapsed
	writeMBps := float64(10) / writeElapsed.Seconds()
	readMBps := float64(10) / readElapsed.Seconds()

	// Normalize to 0-1000 scale
	// Typical: 50-500 MB/s for SSD
	score := int((writeMBps + readMBps) / 2)
	if score > 1000 {
		score = 1000
	}
	if score < 0 {
		score = 0
	}

	return score
}

// benchmarkGPU runs a GPU performance benchmark.
func benchmarkGPU(gpu GPUCapability) int {
	// For now, estimate based on GPU memory and model
	// A real implementation would run CUDA/ROCm/OpenCL kernels

	score := 100 // Base score

	// Add points for memory
	if gpu.Memory >= 16384 { // 16GB+
		score += 200
	} else if gpu.Memory >= 8192 { // 8GB+
		score += 100
	} else if gpu.Memory >= 4096 { // 4GB+
		score += 50
	}

	// Add points for vendor (NVIDIA typically has better compute)
	switch gpu.Vendor {
	case "NVIDIA":
		score += 100
		// Check for compute capability
		if gpu.ComputeCapability >= "8.0" {
			score += 200
		} else if gpu.ComputeCapability >= "7.0" {
			score += 100
		}
	case "AMD/ATI":
		score += 80
	case "Intel":
		score += 50
	}

	// Add points for name patterns that indicate high-end cards
	name := gpu.Name
	if containsAny(name, []string{"A100", "H100", "MI250", "MI300", "RTX 40", "RTX 3090"}) {
		score += 300
	} else if containsAny(name, []string{"V100", "T4", "RTX 3080", "RTX 4080"}) {
		score += 200
	}

	if score > 1000 {
		score = 1000
	}

	return score
}

// benchmarkNetwork runs a network performance benchmark.
func benchmarkNetwork() int {
	// For now, return a default score
	// A real implementation would measure actual network throughput
	// by connecting to a known endpoint or running iperf

	// Check for high-bandwidth interfaces
	score := 200 // Base score

	// Check for 10GbE or faster
	if _, err := os.Stat("/sys/class/net"); err == nil {
		entries, _ := os.ReadDir("/sys/class/net")
		for _, entry := range entries {
			speedPath := "/sys/class/net/" + entry.Name() + "/speed"
			if data, err := os.ReadFile(speedPath); err == nil {
				speed := strings.TrimSpace(string(data))
				if speed >= "10000" {
					score += 500
				} else if speed >= "1000" {
					score += 200
				}
			}
		}
	}

	if score > 1000 {
		score = 1000
	}

	return score
}

// containsAny checks if s contains any of the substrings.
func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}