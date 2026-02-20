package capability

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/rs/zerolog/log"
)

// EngineDetector detects available execution engines.
type EngineDetector interface {
	// DetectEngines returns all detected execution engines.
	DetectEngines(ctx context.Context) ([]EngineCapability, error)

	// DetectDocker detects Docker engine specifically.
	DetectDocker(ctx context.Context) (*EngineCapability, error)

	// DetectWasm detects WebAssembly engine specifically.
	DetectWasm(ctx context.Context) (*EngineCapability, error)

	// DetectNative detects native execution capability.
	DetectNative(ctx context.Context) (*EngineCapability, error)
}

// DefaultEngineDetector is the default engine detector implementation.
type DefaultEngineDetector struct{}

// NewDefaultEngineDetector creates a new default engine detector.
func NewDefaultEngineDetector() *DefaultEngineDetector {
	return &DefaultEngineDetector{}
}

// DetectEngines detects all available execution engines.
func (d *DefaultEngineDetector) DetectEngines(ctx context.Context) ([]EngineCapability, error) {
	var engines []EngineCapability

	// Detect Docker
	if docker, err := d.DetectDocker(ctx); err == nil && docker != nil {
		engines = append(engines, *docker)
	}

	// Detect WebAssembly
	if wasm, err := d.DetectWasm(ctx); err == nil && wasm != nil {
		engines = append(engines, *wasm)
	}

	// Detect Native
	if native, err := d.DetectNative(ctx); err == nil && native != nil {
		engines = append(engines, *native)
	}

	return engines, nil
}

// DetectDocker detects Docker engine availability and version.
func (d *DefaultEngineDetector) DetectDocker(ctx context.Context) (*EngineCapability, error) {
	engine := &EngineCapability{
		Type:        models.EngineDocker,
		Available:   false,
		Constraints: []string{},
		Features:    []string{},
	}

	// Check if docker command exists
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		log.Ctx(ctx).Debug().Msg("Docker not found in PATH")
		return engine, nil
	}

	// Get Docker version
	cmd := exec.CommandContext(ctx, dockerPath, "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to get Docker version")
		return engine, nil
	}

	engine.Version = strings.TrimSpace(string(output))
	engine.Available = true

	// Detect Docker features
	features := d.detectDockerFeatures(ctx, dockerPath)
	engine.Features = features

	// Check for constraints
	constraints := d.detectDockerConstraints(ctx, dockerPath)
	engine.Constraints = constraints

	return engine, nil
}

// detectDockerFeatures detects available Docker features.
func (d *DefaultEngineDetector) detectDockerFeatures(ctx context.Context, dockerPath string) []string {
	features := []string{"containers"}

	// Check for BuildKit
	cmd := exec.CommandContext(ctx, dockerPath, "buildx", "version")
	if err := cmd.Run(); err == nil {
		features = append(features, "buildx")
	}

	// Check for compose
	cmd = exec.CommandContext(ctx, dockerPath, "compose", "version")
	if err := cmd.Run(); err == nil {
		features = append(features, "compose")
	}

	// Check for GPU support
	cmd = exec.CommandContext(ctx, dockerPath, "info", "-f", "{{.Runtimes.nvidia}}")
	if err := cmd.Run(); err == nil {
		features = append(features, "gpu-nvidia")
	}

	// Check for available runtimes
	cmd = exec.CommandContext(ctx, dockerPath, "info", "-f", "{{range .Runtimes}}{{.Path}} {{end}}")
	if output, err := cmd.Output(); err == nil {
		runtimes := strings.TrimSpace(string(output))
		if runtimes != "" {
			features = append(features, "runtimes")
		}
	}

	return features
}

// detectDockerConstraints detects Docker limitations.
func (d *DefaultEngineDetector) detectDockerConstraints(ctx context.Context, dockerPath string) []string {
	var constraints []string

	// Check if running rootless
	cmd := exec.CommandContext(ctx, dockerPath, "context", "ls", "-f", "{{.Name}}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "rootless") {
		constraints = append(constraints, "rootless")
	}

	// Check if privileged mode is available
	if os.Getuid() != 0 {
		// Non-root user might have limited Docker access
		cmd = exec.CommandContext(ctx, dockerPath, "info", "-f", "{{.SecurityOptions}}")
		if out, err := cmd.Output(); err == nil {
			if !strings.Contains(string(out), "rootless") {
				// Running as non-root but not in rootless mode
				// Might need sudo
				constraints = append(constraints, "requires-permissions")
			}
		}
	}

	return constraints
}

// DetectWasm detects WebAssembly execution capability.
func (d *DefaultEngineDetector) DetectWasm(ctx context.Context) (*EngineCapability, error) {
	engine := &EngineCapability{
		Type:        models.EngineWasm,
		Available:   true, // Wasm is always available via wazero (pure Go)
		Features:    []string{"wazero", "no-runtime-dependency"},
		Constraints: []string{},
	}

	// Check for wasmtime binary (optional, faster in some cases)
	if _, err := exec.LookPath("wasmtime"); err == nil {
		engine.Features = append(engine.Features, "wasmtime")

		// Get wasmtime version
		cmd := exec.CommandContext(ctx, "wasmtime", "--version")
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				engine.Version = parts[1]
			}
		}
	}

	// Check for wasmer binary (optional)
	if _, err := exec.LookPath("wasmer"); err == nil {
		engine.Features = append(engine.Features, "wasmer")
	}

	return engine, nil
}

// DetectNative detects native execution capability.
func (d *DefaultEngineDetector) DetectNative(ctx context.Context) (*EngineCapability, error) {
	engine := &EngineCapability{
		Type:        "native",
		Available:   true,
		Features:    []string{"process-execution"},
		Constraints: []string{"os: " + os.Getenv("GOOS")},
	}

	// Detect shell capabilities
	shells := []string{"bash", "sh", "zsh"}
	for _, shell := range shells {
		if _, err := exec.LookPath(shell); err == nil {
			engine.Features = append(engine.Features, shell)
		}
	}

	// Detect common tools
	tools := []string{"python3", "python", "node", "perl", "ruby"}
	for _, tool := range tools {
		if path, err := exec.LookPath(tool); err == nil {
			engine.Features = append(engine.Features, tool)

			// Get version
			cmd := exec.CommandContext(ctx, path, "--version")
			if output, err := cmd.Output(); err == nil {
				// Parse first line for version
				lines := strings.Split(string(output), "\n")
				if len(lines) > 0 {
					// Just note availability, version parsing is tool-specific
				}
			}
		}
	}

	// Security constraints
	if os.Getuid() == 0 {
		engine.Constraints = append(engine.Constraints, "running-as-root")
	}

	return engine, nil
}

// detectHostname detects the system hostname.
func detectHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

// detectStorage detects storage capabilities.
func detectStorage(ctx context.Context) ([]StorageCapability, error) {
	var storage []StorageCapability

	// Detect local disk storage
	local := detectLocalStorage(ctx)
	storage = append(storage, local...)

	// Detect mounted network storage
	network := detectNetworkStorage(ctx)
	storage = append(storage, network...)

	return storage, nil
}

// detectLocalStorage detects local disk storage.
func detectLocalStorage(ctx context.Context) []StorageCapability {
	var storage []StorageCapability

	// Read /proc/mounts for mounted filesystems
	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Msg("Failed to read /proc/mounts")
		return storage
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// Skip pseudo filesystems
		skipFS := map[string]bool{
			"sysfs": true, "proc": true, "devtmpfs": true, "devpts": true,
			"tmpfs": true, "securityfs": true, "cgroup": true, "cgroup2": true,
			"pstore": true, "debugfs": true, "tracefs": true, "configfs": true,
			"fusectl": true, "mqueue": true, "hugetlbfs": true, "autofs": true,
			"binfmt_misc": true, "overlay": true, // Skip overlay for now
		}

		if skipFS[fsType] {
			continue
		}

		// Skip special mount points
		if strings.HasPrefix(mountPoint, "/sys") ||
			strings.HasPrefix(mountPoint, "/proc") ||
			strings.HasPrefix(mountPoint, "/dev") {
			continue
		}

		cap := StorageCapability{
			Type:    "local",
			Path:    mountPoint,
			Network: false,
		}

		// Detect if it's network storage based on device name
		if strings.HasPrefix(device, "//") ||
			strings.HasPrefix(device, "nfs") ||
			strings.Contains(device, ":/") {
			cap.Type = "network"
			cap.Network = true
		}

		// Try to get disk usage
		if usage := getDiskUsage(mountPoint); usage != nil {
			cap.Total = usage.Total
			cap.Available = usage.Available
		}

		// Check if read-only
		if len(fields) >= 4 && strings.Contains(fields[3], "ro") {
			cap.ReadOnly = true
		}

		storage = append(storage, cap)
	}

	return storage
}

// detectNetworkStorage detects network-attached storage.
func detectNetworkStorage(ctx context.Context) []StorageCapability {
	var storage []StorageCapability

	// Check for NFS mounts
	if _, err := exec.LookPath("showmount"); err == nil {
		cmd := exec.CommandContext(ctx, "showmount", "-e", "localhost")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "/") {
					storage = append(storage, StorageCapability{
						Type:    "nfs",
						Path:    strings.Fields(line)[0],
						Network: true,
					})
				}
			}
		}
	}

	return storage
}

// diskUsage represents disk usage information.
type diskUsage struct {
	Total     uint64
	Available uint64
}

// getDiskUsage gets disk usage for a path using statfs.
func getDiskUsage(path string) *diskUsage {
	var stat syscallStatfs
	if err := statfs(path, &stat); err != nil {
		return nil
	}

	return &diskUsage{
		Total:     stat.Blocks * uint64(stat.Bsize),
		Available: stat.Bavail * uint64(stat.Bsize),
	}
}

// detectNetwork detects network capabilities.
func detectNetwork(ctx context.Context) (NetworkCapability, error) {
	cap := NetworkCapability{}

	// Detect private IPs
	privateIPs := detectPrivateIPs(ctx)
	cap.PrivateIPs = privateIPs

	// Check for IPv6 support
	for _, ip := range privateIPs {
		if strings.Contains(ip, ":") {
			cap.IPv6 = true
			break
		}
	}

	// Try to detect public IP (best effort)
	if publicIP, err := detectPublicIP(ctx); err == nil {
		cap.PublicIP = publicIP
	}

	// Try to detect region/zone from cloud metadata
	region, zone := detectCloudMetadata(ctx)
	cap.Region = region
	cap.Zone = zone

	// Check for NAT
	cap.NAT = detectNAT(cap)

	return cap, nil
}

// detectPrivateIPs detects local IP addresses.
func detectPrivateIPs(ctx context.Context) []string {
	var ips []string

	// Use ip command if available
	if _, err := exec.LookPath("ip"); err == nil {
		cmd := exec.CommandContext(ctx, "ip", "-brief", "addr", "show")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				for i, f := range fields {
					if f == "inet" && i+1 < len(fields) {
						// Extract IP without CIDR
						ip := strings.Split(fields[i+1], "/")[0]
						if !strings.HasPrefix(ip, "127.") {
							ips = append(ips, ip)
						}
					} else if f == "inet6" && i+1 < len(fields) {
						ip := strings.Split(fields[i+1], "/")[0]
						if !strings.HasPrefix(ip, "::1") && !strings.HasPrefix(ip, "fe80:") {
							ips = append(ips, ip)
						}
					}
				}
			}
		}
	}

	// Fallback: read from /proc/net/if_inet6 for IPv6
	if _, err := os.Stat("/proc/net/if_inet6"); err == nil {
		data, err := os.ReadFile("/proc/net/if_inet6")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 6 && fields[5] != "lo" {
					// Convert hex IPv6 to standard format
					rawIP := fields[0]
					if len(rawIP) == 32 {
						// Format IPv6 address
						ip := formatIPv6(rawIP)
						if ip != "" && !strings.HasPrefix(ip, "fe80:") {
							ips = append(ips, ip)
						}
					}
				}
			}
		}
	}

	return ips
}

// formatIPv6 converts a hex IPv6 address to standard format.
func formatIPv6(hex string) string {
	if len(hex) != 32 {
		return ""
	}

	var parts []string
	for i := 0; i < 32; i += 4 {
		part := hex[i : i+4]
		// Remove leading zeros
		part = strings.TrimLeft(part, "0")
		if part == "" {
			part = "0"
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, ":")
}

// detectPublicIP attempts to detect the public IP address.
func detectPublicIP(ctx context.Context) (string, error) {
	// Try multiple services for reliability
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, service := range services {
		if _, err := exec.LookPath("curl"); err != nil {
			continue
		}

		cmd := exec.CommandContext(ctx, "curl", "-s", "-m", "2", service)
		output, err := cmd.Output()
		if err == nil {
			ip := strings.TrimSpace(string(output))
			if ip != "" {
				return ip, nil
			}
		}
	}

	return "", nil
}

// detectCloudMetadata detects cloud provider region and zone.
func detectCloudMetadata(ctx context.Context) (region, zone string) {
	// Try AWS metadata
	if _, err := exec.LookPath("curl"); err == nil {
		// AWS IMDSv2
		cmd := exec.CommandContext(ctx, "curl", "-s", "-m", "1",
			"http://169.254.169.254/latest/meta-data/placement/region")
		if output, err := cmd.Output(); err == nil {
			region = strings.TrimSpace(string(output))
			if region != "" {
				cmd = exec.CommandContext(ctx, "curl", "-s", "-m", "1",
					"http://169.254.169.254/latest/meta-data/placement/availability-zone")
				if output, err := cmd.Output(); err == nil {
					zone = strings.TrimSpace(string(output))
				}
				return
			}
		}

		// GCP metadata
		cmd = exec.CommandContext(ctx, "curl", "-s", "-m", "1",
			"-H", "Metadata-Flavor: Google",
			"http://metadata.google.internal/computeMetadata/v1/instance/zone")
		if output, err := cmd.Output(); err == nil {
			zone = strings.TrimSpace(string(output))
			// Extract region from zone (e.g., "projects/123/zones/us-central1-a" -> "us-central1")
			if parts := strings.Split(zone, "/"); len(parts) > 0 {
				zonePart := parts[len(parts)-1]
				if idx := strings.LastIndex(zonePart, "-"); idx > 0 {
					region = zonePart[:idx]
					zone = zonePart
				}
			}
			return
		}

		// Azure metadata
		cmd = exec.CommandContext(ctx, "curl", "-s", "-m", "1",
			"-H", "Metadata: true",
			"http://169.254.169.254/metadata/instance/compute/location?api-version=2021-02-01&format=text")
		if output, err := cmd.Output(); err == nil {
			region = strings.TrimSpace(string(output))
			return
		}
	}

	return "", ""
}

// detectNAT detects if the node is behind NAT.
func detectNAT(cap NetworkCapability) bool {
	// If we have a public IP and private IPs, check if they match
	if cap.PublicIP != "" && len(cap.PrivateIPs) > 0 {
		// If public IP is not in private IPs, we're behind NAT
		for _, privateIP := range cap.PrivateIPs {
			if privateIP == cap.PublicIP {
				return false
			}
		}
		return true
	}

	// If no public IP but have private IPs, likely behind NAT
	if cap.PublicIP == "" && len(cap.PrivateIPs) > 0 {
		return true
	}

	return false
}
