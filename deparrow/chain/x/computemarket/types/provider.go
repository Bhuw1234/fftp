package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProviderStatus represents the status of a provider
type ProviderStatus int32

const (
	ProviderStatusActive   ProviderStatus = 0
	ProviderStatusInactive ProviderStatus = 1
	ProviderStatusSlashed  ProviderStatus = 2
)

func (s ProviderStatus) String() string {
	switch s {
	case ProviderStatusActive:
		return "active"
	case ProviderStatusInactive:
		return "inactive"
	case ProviderStatusSlashed:
		return "slashed"
	default:
		return "unknown"
	}
}

// ProviderCapabilities represents a provider's compute capabilities
type ProviderCapabilities struct {
	CPU     uint64 `json:"cpu"`      // CPU cores
	Memory  uint64 `json:"memory"`   // Memory in MB
	GPU     uint64 `json:"gpu"`      // GPU count
	Storage uint64 `json:"storage"`  // Storage in GB
	Regions []string `json:"regions"` // Supported regions
	Tags    []string `json:"tags"`    // Capability tags (e.g., "cuda", "rocm")
}

// NewProviderCapabilities creates a new ProviderCapabilities instance
func NewProviderCapabilities(cpu, memory, gpu, storage uint64, regions, tags []string) ProviderCapabilities {
	return ProviderCapabilities{
		CPU:     cpu,
		Memory:  memory,
		GPU:     gpu,
		Storage: storage,
		Regions: regions,
		Tags:    tags,
	}
}

// ToBytes serializes capabilities to bytes
func (c ProviderCapabilities) ToBytes() []byte {
	bz, _ := json.Marshal(c)
	return bz
}

// ProviderCapabilitiesFromBytes deserializes capabilities from bytes
func ProviderCapabilitiesFromBytes(bz []byte) (ProviderCapabilities, error) {
	var caps ProviderCapabilities
	err := json.Unmarshal(bz, &caps)
	return caps, err
}

// Matches checks if capabilities match the required specifications
func (c ProviderCapabilities) Matches(required ProviderCapabilities) bool {
	if c.CPU < required.CPU {
		return false
	}
	if c.Memory < required.Memory {
		return false
	}
	if c.GPU < required.GPU {
		return false
	}
	if c.Storage < required.Storage {
		return false
	}
	// Check if provider supports required regions
	if len(required.Regions) > 0 {
		regionSet := make(map[string]bool)
		for _, r := range c.Regions {
			regionSet[r] = true
		}
		for _, req := range required.Regions {
			if !regionSet[req] {
				return false
			}
		}
	}
	// Check if provider has required tags
	if len(required.Tags) > 0 {
		tagSet := make(map[string]bool)
		for _, t := range c.Tags {
			tagSet[t] = true
		}
		for _, req := range required.Tags {
			if !tagSet[req] {
				return false
			}
		}
	}
	return true
}

// Provider represents a compute provider in the network (extended from types.go)
type ProviderExtended struct {
	Address         string              `json:"address"`
	StakedAmount    sdk.Coin            `json:"staked_amount"`
	ReputationScore uint32              `json:"reputation_score"`  // 0-1000
	Capabilities    ProviderCapabilities `json:"capabilities"`
	CompletedJobs   uint64              `json:"completed_jobs"`
	FailedJobs      uint64              `json:"failed_jobs"`
	SlashedCount    uint64              `json:"slashed_count"`
	Status          ProviderStatus      `json:"status"`
	RegisteredAt    int64               `json:"registered_at"`
	LastActiveAt    int64               `json:"last_active_at"`
}

// NewProviderExtended creates a new extended Provider instance
func NewProviderExtended(address string, stake sdk.Coin, capabilities ProviderCapabilities) *ProviderExtended {
	return &ProviderExtended{
		Address:         address,
		StakedAmount:    stake,
		ReputationScore: 500, // Start with neutral reputation
		Capabilities:    capabilities,
		Status:          ProviderStatusActive,
		RegisteredAt:    0, // Set by keeper
		LastActiveAt:    0, // Set by keeper
	}
}

// Validate performs validation of the extended provider
func (p ProviderExtended) Validate() error {
	if p.Address == "" {
		return fmt.Errorf("provider address cannot be empty")
	}
	if p.StakedAmount.IsNegative() {
		return fmt.Errorf("staked amount cannot be negative")
	}
	if p.ReputationScore > 1000 {
		return fmt.Errorf("reputation score must be between 0 and 1000")
	}
	return nil
}

// IsActive returns true if the provider is active
func (p ProviderExtended) IsActive() bool {
	return p.Status == ProviderStatusActive
}

// IsSlashed returns true if the provider has been slashed
func (p ProviderExtended) IsSlashed() bool {
	return p.Status == ProviderStatusSlashed
}

// SuccessRate calculates the provider's success rate
func (p ProviderExtended) SuccessRate() float64 {
	total := p.CompletedJobs + p.FailedJobs
	if total == 0 {
		return 1.0 // New provider has 100% success rate
	}
	return float64(p.CompletedJobs) / float64(total)
}

// EffectiveReputation calculates effective reputation considering success rate
func (p ProviderExtended) EffectiveReputation() uint32 {
	successRate := p.SuccessRate()
	// Penalize low success rate
	if successRate < 0.5 {
		return uint32(float64(p.ReputationScore) * successRate * 2)
	}
	return p.ReputationScore
}

// MatchScore calculates a matching score for job assignment (0-1000)
func (p ProviderExtended) MatchScore(required ProviderCapabilities) uint32 {
	if !p.Capabilities.Matches(required) {
		return 0
	}
	
	// Base score from reputation
	score := float64(p.EffectiveReputation())
	
	// Bonus for high success rate
	successRate := p.SuccessRate()
	score += successRate * 100
	
	// Bonus for extra capacity (more capable providers are preferred)
	if p.Capabilities.CPU > required.CPU {
		score += 20
	}
	if p.Capabilities.Memory > required.Memory {
		score += 20
	}
	if p.Capabilities.GPU > required.GPU && required.GPU > 0 {
		score += 50
	}
	
	// Cap at 1000
	if score > 1000 {
		score = 1000
	}
	
	return uint32(score)
}
