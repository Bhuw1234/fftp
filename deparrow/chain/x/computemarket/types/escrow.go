package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EscrowStatusExtended represents the status of an escrow (matching proto)
type EscrowStatusExtended int32

const (
	EscrowStatusUnspecified EscrowStatusExtended = 0
	EscrowStatusLocked      EscrowStatusExtended = 1
	EscrowStatusReleased    EscrowStatusExtended = 2
	EscrowStatusRefunded    EscrowStatusExtended = 3
	EscrowStatusDisputed    EscrowStatusExtended = 4
)

func (s EscrowStatusExtended) String() string {
	switch s {
	case EscrowStatusUnspecified:
		return "unspecified"
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

// Dispute represents a dispute over an escrow
type Dispute struct {
	ID           string          `json:"id"`
	EscrowID     string          `json:"escrow_id"`
	JobID        string          `json:"job_id"`
	Submitter    string          `json:"submitter"`
	Provider     string          `json:"provider"`
	Reason       string          `json:"reason"`
	Evidence     []byte          `json:"evidence"`
	Status       DisputeStatus   `json:"status"`
	Resolution   string          `json:"resolution"`
	Winner       string          `json:"winner"`
	OpenedAt     int64           `json:"opened_at"`
	ResolvedAt   int64           `json:"resolved_at"`
	SlashAmount  sdk.Coin        `json:"slash_amount"`
}

// DisputeStatus represents the status of a dispute
type DisputeStatus int32

const (
	DisputeStatusOpen     DisputeStatus = 0
	DisputeStatusResolved DisputeStatus = 1
	DisputeStatusExpired  DisputeStatus = 2
)

func (s DisputeStatus) String() string {
	switch s {
	case DisputeStatusOpen:
		return "open"
	case DisputeStatusResolved:
		return "resolved"
	case DisputeStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// NewDispute creates a new Dispute instance
func NewDispute(id, escrowID, jobID, submitter, provider, reason string, evidence []byte) *Dispute {
	return &Dispute{
		ID:        id,
		EscrowID:  escrowID,
		JobID:     jobID,
		Submitter: submitter,
		Provider:  provider,
		Reason:    reason,
		Evidence:  evidence,
		Status:    DisputeStatusOpen,
		OpenedAt:  0, // Set by keeper
	}
}

// Validate performs basic validation of the dispute
func (d Dispute) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("dispute ID cannot be empty")
	}
	if d.EscrowID == "" {
		return fmt.Errorf("escrow ID cannot be empty")
	}
	if d.Submitter == "" {
		return fmt.Errorf("submitter cannot be empty")
	}
	if d.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if d.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	return nil
}

// IsResolved returns true if the dispute is resolved
func (d Dispute) IsResolved() bool {
	return d.Status == DisputeStatusResolved || d.Status == DisputeStatusExpired
}

// EscrowExtended represents an escrow contract for job payment (extended)
type EscrowExtended struct {
	ID           string               `json:"id"`
	JobID        string               `json:"job_id"`
	Submitter    string               `json:"submitter"`
	Provider     string               `json:"provider"`
	Amount       sdk.Coin             `json:"amount"`
	Fee          sdk.Coin             `json:"fee"`        // Network fee
	Status       EscrowStatusExtended `json:"status"`
	Deadline     int64                `json:"deadline"`
	CreatedAt    int64                `json:"created_at"`
	ReleasedAt   int64                `json:"released_at"`
	RefundedAt   int64                `json:"refunded_at"`
	DisputeID    string               `json:"dispute_id"` // Set if disputed
}

// NewEscrowExtended creates a new extended Escrow instance
func NewEscrowExtended(id, jobID, submitter, provider string, amount, fee sdk.Coin, deadline int64) *EscrowExtended {
	return &EscrowExtended{
		ID:        id,
		JobID:     jobID,
		Submitter: submitter,
		Provider:  provider,
		Amount:    amount,
		Fee:       fee,
		Status:    EscrowStatusLocked,
		Deadline:  deadline,
		CreatedAt: 0, // Set by keeper
	}
}

// Validate performs basic validation of the escrow
func (e EscrowExtended) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("escrow ID cannot be empty")
	}
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
	if e.Deadline <= 0 {
		return fmt.Errorf("deadline must be positive")
	}
	return nil
}

// IsExpired returns true if the escrow has expired
func (e EscrowExtended) IsExpired(currentTime int64) bool {
	return currentTime > e.Deadline
}

// IsLocked returns true if the escrow is locked
func (e EscrowExtended) IsLocked() bool {
	return e.Status == EscrowStatusLocked
}

// IsReleased returns true if the escrow has been released
func (e EscrowExtended) IsReleased() bool {
	return e.Status == EscrowStatusReleased
}

// IsRefunded returns true if the escrow has been refunded
func (e EscrowExtended) IsRefunded() bool {
	return e.Status == EscrowStatusRefunded
}

// IsDisputed returns true if the escrow is disputed
func (e EscrowExtended) IsDisputed() bool {
	return e.Status == EscrowStatusDisputed
}

// TotalAmount returns the total amount (payment + fee)
func (e EscrowExtended) TotalAmount() sdk.Coins {
	return sdk.NewCoins(e.Amount).Add(e.Fee)
}

// JobMatch represents a match between a job and a provider
type JobMatch struct {
	JobID        string   `json:"job_id"`
	Provider     string   `json:"provider"`
	Score        uint32   `json:"score"`
	MatchedAt    int64    `json:"matched_at"`
	Requirements ProviderCapabilities `json:"requirements"`
}

// NewJobMatch creates a new JobMatch instance
func NewJobMatch(jobID, provider string, score uint32, requirements ProviderCapabilities) *JobMatch {
	return &JobMatch{
		JobID:        jobID,
		Provider:     provider,
		Score:        score,
		MatchedAt:    time.Now().Unix(),
		Requirements: requirements,
	}
}

// SlashingConfig defines configuration for slashing
type SlashingConfig struct {
	FailedJobSlashPercent    uint32   // Percentage to slash for failed jobs
	DisputeLoseSlashPercent  uint32   // Percentage to slash when losing dispute
	MinReputationThreshold   uint32   // Minimum reputation before slashing
	MaxSlashedAmount         math.Int // Maximum amount that can be slashed
}

// DefaultSlashingConfig returns default slashing configuration
func DefaultSlashingConfig() SlashingConfig {
	return SlashingConfig{
		FailedJobSlashPercent:    10,  // 10%
		DisputeLoseSlashPercent:  50,  // 50%
		MinReputationThreshold:   100, // Below 100 = high risk
		MaxSlashedAmount:         math.NewInt(10000), // Max 10,000 DPC per slash
	}
}
