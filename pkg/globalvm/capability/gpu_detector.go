package capability

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/rs/zerolog/log"
)

// GPUDetector detects GPU devices on the system.
type GPUDetector interface {
	// DetectGPUs returns all detected GPU devices.
	DetectGPUs(ctx context.Context) ([]GPUCapability, error)

	// DetectNVIDIA detects NVIDIA GPUs specifically.
	DetectNVIDIA(ctx context.Context) ([]GPUCapability, error)

	// DetectAMD detects AMD GPUs specifically.
	DetectAMD(ctx context.Context) ([]GPUCapability, error)

	// DetectIntel detects Intel GPUs specifically.
	DetectIntel(ctx context.Context) ([]GPUCapability, error)
}

// DefaultGPUDetector is the default GPU detector implementation.
type DefaultGPUDetector struct {
	nvidiaDetector *nvidiaDetector
	amdDetector    *amdDetector
	intelDetector  *intelDetector
}

// NewDefaultGPUDetector creates a new default GPU detector.
func NewDefaultGPUDetector() *DefaultGPUDetector {
	return &DefaultGPUDetector{
		nvidiaDetector: &nvidiaDetector{},
		amdDetector:    &amdDetector{},
		intelDetector:  &intelDetector{},
	}
}

// DetectGPUs detects all GPUs from all vendors.
func (d *DefaultGPUDetector) DetectGPUs(ctx context.Context) ([]GPUCapability, error) {
	var allGPUs []GPUCapability

	// Detect NVIDIA GPUs
	nvidiaGPUs, err := d.DetectNVIDIA(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("NVIDIA detection failed")
	}
	allGPUs = append(allGPUs, nvidiaGPUs...)

	// Detect AMD GPUs
	amdGPUs, err := d.DetectAMD(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("AMD detection failed")
	}
	allGPUs = append(allGPUs, amdGPUs...)

	// Detect Intel GPUs
	intelGPUs, err := d.DetectIntel(ctx)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Intel detection failed")
	}
	allGPUs = append(allGPUs, intelGPUs...)

	// Also check sysfs for any missed GPUs
	sysfsGPUs := detectSysfsGPUs(ctx)
	allGPUs = mergeGPUs(allGPUs, sysfsGPUs)

	return allGPUs, nil
}

// DetectNVIDIA detects NVIDIA GPUs using nvidia-smi.
func (d *DefaultGPUDetector) DetectNVIDIA(ctx context.Context) ([]GPUCapability, error) {
	return d.nvidiaDetector.detect(ctx)
}

// DetectAMD detects AMD GPUs using rocm-smi and sysfs.
func (d *DefaultGPUDetector) DetectAMD(ctx context.Context) ([]GPUCapability, error) {
	return d.amdDetector.detect(ctx)
}

// DetectIntel detects Intel GPUs using xpu-smi and sysfs.
func (d *DefaultGPUDetector) DetectIntel(ctx context.Context) ([]GPUCapability, error) {
	return d.intelDetector.detect(ctx)
}

// nvidiaDetector detects NVIDIA GPUs.
type nvidiaDetector struct{}

func (n *nvidiaDetector) detect(ctx context.Context) ([]GPUCapability, error) {
	// Check if nvidia-smi is available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil, err
	}

	// Query GPU info
	cmd := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=index,name,memory.total,driver_version,compute_cap,temperature.gpu,utilization.gpu,pci.address",
		"--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return n.parseOutput(output)
}

func (n *nvidiaDetector) parseOutput(output []byte) ([]GPUCapability, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	gpus := make([]GPUCapability, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 4 {
			continue
		}

		gpu := GPUCapability{
			Vendor:    models.GPUVendorNvidia,
			Available: true,
		}

		// Parse index
		if idx, err := strconv.ParseUint(strings.TrimSpace(fields[0]), 10, 64); err == nil {
			gpu.Index = idx
		}

		// Parse name
		gpu.Name = strings.TrimSpace(fields[1])

		// Parse memory (in MiB)
		if mem, err := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64); err == nil {
			gpu.Memory = mem
		}

		// Parse driver version
		gpu.Driver = strings.TrimSpace(fields[3])

		// Parse compute capability if available
		if len(fields) > 4 {
			gpu.ComputeCapability = strings.TrimSpace(fields[4])
		}

		// Parse temperature if available
		if len(fields) > 5 {
			if temp, err := strconv.Atoi(strings.TrimSpace(fields[5])); err == nil {
				gpu.Temperature = temp
			}
		}

		// Parse utilization if available
		if len(fields) > 6 {
			if util, err := strconv.Atoi(strings.TrimSpace(fields[6])); err == nil {
				gpu.Utilization = util
			}
		}

		// Parse PCI address if available
		if len(fields) > 7 {
			gpu.PCIAddress = strings.ToLower(strings.TrimSpace(fields[7]))
		}

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// amdDetector detects AMD GPUs.
type amdDetector struct{}

func (a *amdDetector) detect(ctx context.Context) ([]GPUCapability, error) {
	var gpus []GPUCapability

	// Try rocm-smi first
	if rocmGPUs, err := a.detectROCm(ctx); err == nil {
		gpus = append(gpus, rocmGPUs...)
	}

	// Also check sysfs for AMD GPUs
	sysfsGPUs := a.detectSysfsAMD(ctx)
	gpus = mergeGPUs(gpus, sysfsGPUs)

	return gpus, nil
}

func (a *amdDetector) detectROCm(ctx context.Context) ([]GPUCapability, error) {
	// Check if rocm-smi is available
	if _, err := exec.LookPath("rocm-smi"); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "rocm-smi",
		"--showproductname", "--showbus", "--showmeminfo", "vram", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return a.parseROCmOutput(output)
}

func (a *amdDetector) parseROCmOutput(output []byte) ([]GPUCapability, error) {
	// Parse JSON output from rocm-smi
	// Example: {"card0": {"Card series": "Instinct MI210", "VRAM Total Memory (B)": "68702699520", ...}}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	gpus := make([]GPUCapability, 0)

	for _, line := range lines {
		if strings.Contains(line, `"card`) {
			gpu := GPUCapability{
				Vendor:    models.GPUVendorAMDATI,
				Available: true,
			}

			// Extract card index
			if idx := strings.Index(line, `"card`); idx >= 0 {
				if num, err := strconv.ParseUint(line[idx+5:idx+6], 10, 64); err == nil {
					gpu.Index = num
				}
			}

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

func (a *amdDetector) detectSysfsAMD(ctx context.Context) []GPUCapability {
	var gpus []GPUCapability

	// Check /sys/class/drm for AMD GPU devices
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return gpus
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}

		// Check if this is an AMD GPU
		devicePath := "/sys/class/drm/" + name + "/device/vendor"
		vendorBytes, err := os.ReadFile(devicePath)
		if err != nil {
			continue
		}

		vendor := strings.TrimSpace(string(vendorBytes))
		// AMD vendor ID is 0x1002
		if vendor != "0x1002" {
			continue
		}

		gpu := GPUCapability{
			Vendor:    models.GPUVendorAMDATI,
			Available: true,
		}

		// Extract index from card name
		if idx, err := strconv.ParseUint(name[4:], 10, 64); err == nil {
			gpu.Index = idx
		}

		// Try to get GPU name
		namePath := "/sys/class/drm/" + name + "/device/uevent"
		if uevent, err := os.ReadFile(namePath); err == nil {
			lines := strings.Split(string(uevent), "\n")
			for _, l := range lines {
				if strings.HasPrefix(l, "PCI_SLOT_NAME=") {
					gpu.PCIAddress = strings.ToLower(strings.TrimPrefix(l, "PCI_SLOT_NAME="))
				}
			}
		}

		// Try to get VRAM size
		vramPath := "/sys/class/drm/" + name + "/device/mem_info_vram_total"
		if vram, err := os.ReadFile(vramPath); err == nil {
			if memBytes, err := strconv.ParseUint(strings.TrimSpace(string(vram)), 10, 64); err == nil {
				gpu.Memory = memBytes / (1024 * 1024) // Convert to MiB
			}
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// intelDetector detects Intel GPUs.
type intelDetector struct{}

func (i *intelDetector) detect(ctx context.Context) ([]GPUCapability, error) {
	var gpus []GPUCapability

	// Try xpu-smi first
	if xpuGPUs, err := i.detectXPU(ctx); err == nil {
		gpus = append(gpus, xpuGPUs...)
	}

	// Also check sysfs for Intel GPUs
	sysfsGPUs := i.detectSysfsIntel(ctx)
	gpus = mergeGPUs(gpus, sysfsGPUs)

	return gpus, nil
}

func (i *intelDetector) detectXPU(ctx context.Context) ([]GPUCapability, error) {
	// Check if xpu-smi is available
	if _, err := exec.LookPath("xpu-smi"); err != nil {
		return nil, err
	}

	// Get device list
	cmd := exec.CommandContext(ctx, "xpu-smi", "discovery", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return i.parseXPUOutput(output)
}

func (i *intelDetector) parseXPUOutput(output []byte) ([]GPUCapability, error) {
	// Parse JSON output from xpu-smi
	// Example: {"device_list": [{"device_id": 0, "device_name": "Intel Data Center GPU", ...}]}
	var gpus []GPUCapability

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"device_id"`) {
			gpu := GPUCapability{
				Vendor:    models.GPUVendorIntel,
				Available: true,
			}

			// Extract device ID
			if idx := strings.Index(line, `"device_id"`); idx >= 0 {
				// Parse the JSON value
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					val := strings.TrimSpace(parts[1])
					val = strings.TrimSuffix(val, ",")
					if id, err := strconv.ParseUint(val, 10, 64); err == nil {
						gpu.Index = id
					}
				}
			}

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

func (i *intelDetector) detectSysfsIntel(ctx context.Context) []GPUCapability {
	var gpus []GPUCapability

	// Check /sys/class/drm for Intel GPU devices
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return gpus
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}

		// Check if this is an Intel GPU
		devicePath := "/sys/class/drm/" + name + "/device/vendor"
		vendorBytes, err := os.ReadFile(devicePath)
		if err != nil {
			continue
		}

		vendor := strings.TrimSpace(string(vendorBytes))
		// Intel vendor ID is 0x8086
		if vendor != "0x8086" {
			continue
		}

		gpu := GPUCapability{
			Vendor:    models.GPUVendorIntel,
			Available: true,
		}

		// Extract index from card name
		if idx, err := strconv.ParseUint(name[4:], 10, 64); err == nil {
			gpu.Index = idx
		}

		// Get GPU name from driver
		driverPath := "/sys/class/drm/" + name + "/device/driver"
		if link, err := os.Readlink(driverPath); err == nil {
			gpu.Name = strings.TrimPrefix(filepath.Base(link), "card")
			if gpu.Name == "" {
				gpu.Name = "Intel GPU"
			}
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// detectSysfsGPUs detects GPUs via sysfs as a fallback.
func detectSysfsGPUs(ctx context.Context) []GPUCapability {
	var gpus []GPUCapability

	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return gpus
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}

		// Check vendor
		devicePath := "/sys/class/drm/" + name + "/device/vendor"
		vendorBytes, err := os.ReadFile(devicePath)
		if err != nil {
			continue
		}

		vendor := strings.TrimSpace(string(vendorBytes))
		var gpuVendor models.GPUVendor
		switch vendor {
		case "0x10de":
			gpuVendor = models.GPUVendorNvidia
		case "0x1002":
			gpuVendor = models.GPUVendorAMDATI
		case "0x8086":
			gpuVendor = models.GPUVendorIntel
		default:
			continue
		}

		gpu := GPUCapability{
			Vendor:    gpuVendor,
			Available: true,
		}

		// Extract index
		if idx, err := strconv.ParseUint(name[4:], 10, 64); err == nil {
			gpu.Index = idx
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// mergeGPUs merges GPU lists, avoiding duplicates by index.
func mergeGPUs(existing, newGPUs []GPUCapability) []GPUCapability {
	existingIndices := make(map[uint64]bool)
	for _, gpu := range existing {
		existingIndices[gpu.Index] = true
	}

	for _, gpu := range newGPUs {
		if !existingIndices[gpu.Index] {
			existing = append(existing, gpu)
		}
	}

	return existing
}

// detectOpenCL detects OpenCL support.
func detectOpenCL(ctx context.Context) (string, error) {
	// Check for OpenCL platforms
	if _, err := os.Stat("/etc/OpenCL/vendors"); err == nil {
		entries, _ := os.ReadDir("/etc/OpenCL/vendors")
		if len(entries) > 0 {
			return "1.2", nil // Assume 1.2 if vendors exist
		}
	}

	// Check clinfo command
	if _, err := exec.LookPath("clinfo"); err == nil {
		cmd := exec.CommandContext(ctx, "clinfo", "-l")
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			return "1.2", nil
		}
	}

	return "", nil
}

// detectVulkan detects Vulkan support.
func detectVulkan(ctx context.Context) (string, error) {
	if _, err := exec.LookPath("vulkaninfo"); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "vulkaninfo", "--summary")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Vulkan Instance Version") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "1.0", nil
}
