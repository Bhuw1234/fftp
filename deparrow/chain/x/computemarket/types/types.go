package types

import "encoding/json"

// ProviderStatus represents the status of a provider
type ProviderStatus int32

const (
	ProviderStatusUnspecified ProviderStatus = 0
	ProviderStatusActive      ProviderStatus = 1
	ProviderStatusInactive    ProviderStatus = 2
	ProviderStatusSlashed     ProviderStatus = 3
)

// EscrowStatus represents the status of an escrow
type EscrowStatus int32

const (
	EscrowStatusUnspecified EscrowStatus = 0
	EscrowStatusLocked      EscrowStatus = 1
	EscrowStatusReleased    EscrowStatus = 2
	EscrowStatusRefunded    EscrowStatus = 3
	EscrowStatusDisputed    EscrowStatus = 4
)

// DisputeStatus represents the status of a dispute
type DisputeStatus int32

const (
	DisputeStatusOpen     DisputeStatus = 0
	DisputeStatusResolved DisputeStatus = 1
	DisputeStatusExpired  DisputeStatus = 2
)

// Coin represents a token amount with denom
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// ProviderCapabilities represents a provider's compute capabilities
type ProviderCapabilities struct {
	CPU     uint64   `json:"cpu"`
	Memory  uint64   `json:"memory"`
	GPU     uint64   `json:"gpu"`
	Storage uint64   `json:"storage"`
	Regions []string `json:"regions"`
	Tags    []string `json:"tags"`
}

// Provider represents a compute provider in the network
type Provider struct {
	Address         string              `json:"address"`
	StakedAmount    Coin                `json:"staked_amount"`
	ReputationScore uint32              `json:"reputation_score"`
	Capabilities    ProviderCapabilities `json:"capabilities"`
	CompletedJobs   uint64              `json:"completed_jobs"`
	FailedJobs      uint64              `json:"failed_jobs"`
	SlashedCount    uint64              `json:"slashed_count"`
	Status          ProviderStatus      `json:"status"`
	RegisteredAt    int64               `json:"registered_at"`
	LastActiveAt    int64               `json:"last_active_at"`
}

// Escrow represents an escrow contract for job payment
type Escrow struct {
	ID          string       `json:"id"`
	JobID       string       `json:"job_id"`
	Submitter   string       `json:"submitter"`
	Provider    string       `json:"provider"`
	Amount      Coin         `json:"amount"`
	Fee         Coin         `json:"fee"`
	Status      EscrowStatus `json:"status"`
	Deadline    int64        `json:"deadline"`
	CreatedAt   int64        `json:"created_at"`
	ReleasedAt  int64        `json:"released_at"`
	RefundedAt  int64        `json:"refunded_at"`
	DisputeID   string       `json:"dispute_id"`
}

// Dispute represents a dispute over an escrow
type Dispute struct {
	ID           string         `json:"id"`
	EscrowID     string         `json:"escrow_id"`
	JobID        string         `json:"job_id"`
	Submitter    string         `json:"submitter"`
	Provider     string         `json:"provider"`
	Reason       string         `json:"reason"`
	Evidence     []byte         `json:"evidence"`
	Status       DisputeStatus  `json:"status"`
	Resolution   string         `json:"resolution"`
	Winner       string         `json:"winner"`
	OpenedAt     int64          `json:"opened_at"`
	ResolvedAt   int64          `json:"resolved_at"`
	SlashAmount  Coin           `json:"slash_amount"`
}

// JobMatch represents a match between a job and a provider
type JobMatch struct {
	JobID        string               `json:"job_id"`
	Provider     string               `json:"provider"`
	Score        uint32               `json:"score"`
	MatchedAt    int64                `json:"matched_at"`
	Requirements ProviderCapabilities `json:"requirements"`
}

// GenesisState defines the computemarket module's genesis state
type GenesisState struct {
	Params        Params      `json:"params"`
	Providers     []Provider  `json:"providers"`
	Escrows       []Escrow    `json:"escrows"`
	Disputes      []Dispute   `json:"disputes"`
	JobMatches    []JobMatch  `json:"job_matches"`
	TotalStaked   string      `json:"total_staked"`
	TotalEscrowed string      `json:"total_escrowed"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:        DefaultParams(),
		TotalStaked:   "0",
		TotalEscrowed: "0",
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}

// Bytes serializes the genesis state
func (gs GenesisState) Bytes() []byte {
	bz, _ := json.Marshal(gs)
	return bz
}
