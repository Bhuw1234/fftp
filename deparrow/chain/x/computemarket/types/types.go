package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "computemarket"

	// StoreKey is the store key string for computemarket
	StoreKey = ModuleName

	// RouterKey is the message route for computemarket
	RouterKey = ModuleName

	// QuerierRoute is the querier route for computemarket
	QuerierRoute = ModuleName
)

// Provider represents a compute provider in the network
// This is the simplified version for backwards compatibility
// Use ProviderExtended for full functionality
type Provider struct {
	Address         string
	StakedAmount    sdk.Coin
	ReputationScore uint32 // 0-1000
	Capabilities    []byte // CPU, GPU, memory specs (serialized ProviderCapabilities)
	CompletedJobs   uint64
	FailedJobs      uint64
	Active          bool
}

// Escrow represents an escrow contract for job payment
// This is the simplified version for backwards compatibility
// Use EscrowExtended for full functionality
type Escrow struct {
	JobID     string
	Submitter string
	Provider  string
	Amount    sdk.Coin
	Status    EscrowStatusExtended
	Deadline  int64
	CreatedAt int64
}

// NewProvider creates a new Provider instance
func NewProvider(address string, stake sdk.Coin) *Provider {
	return &Provider{
		Address:         address,
		StakedAmount:    stake,
		ReputationScore: 500, // Start with neutral reputation
		Active:          true,
	}
}

// NewEscrow creates a new Escrow instance
func NewEscrow(jobID, submitter, provider string, amount sdk.Coin, deadline int64) *Escrow {
	return &Escrow{
		JobID:     jobID,
		Submitter: submitter,
		Provider:  provider,
		Amount:    amount,
		Status:    EscrowStatusLocked,
		Deadline:  deadline,
	}
}

// Validate performs basic validation of the provider
func (p Provider) Validate() error {
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

// Validate performs basic validation of the escrow
func (e Escrow) Validate() error {
	if e.JobID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}
	if e.Submitter == "" {
		return fmt.Errorf("submitter cannot be empty")
	}
	if e.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if e.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}
	return nil
}

// Params defines the parameters for the computemarket module.
type Params struct {
	MinStake            sdk.Coin // Minimum stake for providers
	DisputePeriod       uint32   // Dispute period in blocks
	MaxJobDuration      uint32   // Maximum job duration in seconds
	NetworkFee          math.LegacyDec // Network fee percentage (0.01 = 1%)
	MinReputation       uint32   // Minimum reputation to accept jobs
	SlashPercent        uint32   // Percentage to slash for failed jobs
	DisputeSlashPercent uint32   // Percentage to slash when losing dispute
}

// DefaultParams returns default computemarket module parameters
func DefaultParams() Params {
	return Params{
		MinStake:            sdk.NewCoin("dpc", math.NewInt(1000)), // 1000 DPC minimum stake
		DisputePeriod:       100,                                   // 100 blocks
		MaxJobDuration:      86400,                                 // 24 hours
		NetworkFee:          math.LegacyNewDecWithPrec(1, 2),       // 1% fee
		MinReputation:       100,                                   // Min 100 reputation
		SlashPercent:        10,                                    // 10% slash for failed jobs
		DisputeSlashPercent: 50,                                    // 50% slash for lost disputes
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MinStake.IsNegative() {
		return fmt.Errorf("min stake cannot be negative")
	}
	if p.DisputePeriod == 0 {
		return fmt.Errorf("dispute period must be positive")
	}
	if p.MaxJobDuration == 0 {
		return fmt.Errorf("max job duration must be positive")
	}
	if p.NetworkFee.IsNegative() || p.NetworkFee.GT(math.LegacyOneDec()) {
		return fmt.Errorf("network fee must be between 0 and 1")
	}
	if p.MinReputation > 1000 {
		return fmt.Errorf("min reputation must be between 0 and 1000")
	}
	if p.SlashPercent > 100 {
		return fmt.Errorf("slash percent must be between 0 and 100")
	}
	if p.DisputeSlashPercent > 100 {
		return fmt.Errorf("dispute slash percent must be between 0 and 100")
	}
	return nil
}

// GetMinStake returns the minimum stake as sdk.Int
func (p Params) GetMinStake() math.Int {
	return p.MinStake.Amount
}

// CalculateFee calculates the network fee for an amount
func (p Params) CalculateFee(amount math.Int) math.Int {
	return p.NetworkFee.MulInt(amount).TruncateInt()
}

// CalculateSlash calculates the slash amount for a stake
func (p Params) CalculateSlash(stake math.Int, isDispute bool) math.Int {
	percent := p.SlashPercent
	if isDispute {
		percent = p.DisputeSlashPercent
	}
	return stake.Mul(math.NewInt(int64(percent))).Quo(math.NewInt(100))
}