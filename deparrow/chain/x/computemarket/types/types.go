package types

import (
	fmt "fmt"

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

// EscrowStatus represents the status of an escrow
type EscrowStatus int32

const (
	EscrowStatusLocked    EscrowStatus = 0
	EscrowStatusReleased  EscrowStatus = 1
	EscrowStatusRefunded  EscrowStatus = 2
	EscrowStatusDisputed  EscrowStatus = 3
)

func (s EscrowStatus) String() string {
	switch s {
	case EscrowStatusLocked:
		return "locked"
	case EscrowStatusReleased:
		return "released"
	case EscrowStatusRefunded:
		return "refunded"
	case EscrowStatusDisputed:
		return "disputed"
	default:
		return "unknown"
	}
}

// Provider represents a compute provider in the network
type Provider struct {
	Address         string
	StakedAmount    sdk.Coin
	ReputationScore uint32 // 0-1000
	Capabilities    []byte // CPU, GPU, memory specs
	CompletedJobs   uint64
	FailedJobs      uint64
	Active          bool
}

// Escrow represents an escrow contract for job payment
type Escrow struct {
	JobID     string
	Submitter string
	Provider  string
	Amount    sdk.Coin
	Status    EscrowStatus
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
	MinStake       sdk.Coin
	DisputePeriod  uint32 // in blocks
	MaxJobDuration uint32 // in seconds
}

// DefaultParams returns default computemarket module parameters
func DefaultParams() Params {
	return Params{
		MinStake:       sdk.NewCoin("dpc", sdk.NewInt(1000)), // 1000 DPC minimum stake
		DisputePeriod:  100,                                  // 100 blocks
		MaxJobDuration: 86400,                                // 24 hours
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
	return nil
}
