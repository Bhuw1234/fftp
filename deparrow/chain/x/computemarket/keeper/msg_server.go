// Package keeper implements the message server for the computemarket module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// MsgRegisterProvider represents a provider registration message
type MsgRegisterProvider struct {
	Address      string                `json:"address"`
	Stake        string                `json:"stake"`
	Capabilities types.ProviderCapabilities `json:"capabilities"`
}

// MsgRegisterProviderResponse is the response for provider registration
type MsgRegisterProviderResponse struct {
	Address         string `json:"address"`
	ReputationScore uint32 `json:"reputation_score"`
}

// MsgUnregisterProvider represents a provider unregistration message
type MsgUnregisterProvider struct {
	Address string `json:"address"`
}

// MsgUnregisterProviderResponse is the response for provider unregistration
type MsgUnregisterProviderResponse struct{}

// MsgStakeProvider represents a stake increase message
type MsgStakeProvider struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// MsgStakeProviderResponse is the response for stake increase
type MsgStakeProviderResponse struct {
	Address     string `json:"address"`
	TotalStake  string `json:"total_stake"`
}

// MsgUnstakeProvider represents a stake decrease message
type MsgUnstakeProvider struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// MsgUnstakeProviderResponse is the response for stake decrease
type MsgUnstakeProviderResponse struct {
	Address         string `json:"address"`
	RemainingStake  string `json:"remaining_stake"`
}

// MsgCreateEscrow represents an escrow creation message
type MsgCreateEscrow struct {
	JobID     string `json:"job_id"`
	Submitter string `json:"submitter"`
	Provider  string `json:"provider"`
	Amount    string `json:"amount"`
	Duration  int64  `json:"duration"`
}

// MsgCreateEscrowResponse is the response for escrow creation
type MsgCreateEscrowResponse struct {
	EscrowID string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgReleaseEscrow represents an escrow release message
type MsgReleaseEscrow struct {
	EscrowID  string `json:"escrow_id"`
	Submitter string `json:"submitter"`
}

// MsgReleaseEscrowResponse is the response for escrow release
type MsgReleaseEscrowResponse struct {
	EscrowID string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgRefundEscrow represents an escrow refund message
type MsgRefundEscrow struct {
	EscrowID string `json:"escrow_id"`
	Caller   string `json:"caller"`
}

// MsgRefundEscrowResponse is the response for escrow refund
type MsgRefundEscrowResponse struct {
	EscrowID string `json:"escrow_id"`
	Status   string `json:"status"`
}

// MsgOpenDispute represents a dispute opening message
type MsgOpenDispute struct {
	EscrowID string `json:"escrow_id"`
	Disputer string `json:"disputer"`
	Reason   string `json:"reason"`
	Evidence []byte `json:"evidence"`
}

// MsgOpenDisputeResponse is the response for dispute opening
type MsgOpenDisputeResponse struct {
	DisputeID string `json:"dispute_id"`
	Status    string `json:"status"`
}

// MsgResolveDispute represents a dispute resolution message
type MsgResolveDispute struct {
	DisputeID  string `json:"dispute_id"`
	Resolver   string `json:"resolver"`
	Resolution string `json:"resolution"`
	Winner     string `json:"winner"`
}

// MsgResolveDisputeResponse is the response for dispute resolution
type MsgResolveDisputeResponse struct {
	DisputeID  string `json:"dispute_id"`
	Resolution string `json:"resolution"`
}

// RegisterProvider handles provider registration
func (k Keeper) RegisterProvider(msg MsgRegisterProvider, blockHeight int64) (*MsgRegisterProviderResponse, error) {
	// Check if provider already exists
	if _, found := k.GetProvider(msg.Address); found {
		return nil, types.ErrProviderAlreadyRegistered
	}

	// Validate stake
	params := k.GetParams()
	if parseUint64(msg.Stake) < parseUint64(params.MinStake) {
		return nil, types.ErrInsufficientStake
	}

	// Create provider
	provider := types.Provider{
		Address:         msg.Address,
		StakedAmount:    types.Coin{Denom: "dpc", Amount: msg.Stake},
		ReputationScore: 500, // Start with neutral reputation
		Capabilities:    msg.Capabilities,
		CompletedJobs:   0,
		FailedJobs:      0,
		SlashedCount:    0,
		Status:          types.ProviderStatusActive,
		RegisteredAt:    blockHeight,
		LastActiveAt:    blockHeight,
	}

	// Store provider
	if err := k.SetProvider(provider); err != nil {
		return nil, fmt.Errorf("failed to store provider: %w", err)
	}

	// Update total staked
	currentStaked := parseUint64(k.GetTotalStaked())
	newStaked := currentStaked + parseUint64(msg.Stake)
	k.SetTotalStaked(fmt.Sprintf("%d", newStaked))

	log.Printf("[computemarket] Provider %s registered with stake %s DPC", msg.Address, msg.Stake)

	return &MsgRegisterProviderResponse{
		Address:         msg.Address,
		ReputationScore: provider.ReputationScore,
	}, nil
}

// UnregisterProvider handles provider unregistration
func (k Keeper) UnregisterProvider(msg MsgUnregisterProvider) (*MsgUnregisterProviderResponse, error) {
	provider, found := k.GetProvider(msg.Address)
	if !found {
		return nil, types.ErrProviderNotFound
	}

	// Set provider as inactive
	provider.Status = types.ProviderStatusInactive

	if err := k.SetProvider(provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	// Update total staked
	currentStaked := parseUint64(k.GetTotalStaked())
	stakedAmount := parseUint64(provider.StakedAmount.Amount)
	if currentStaked >= stakedAmount {
		k.SetTotalStaked(fmt.Sprintf("%d", currentStaked-stakedAmount))
	}

	log.Printf("[computemarket] Provider %s unregistered", msg.Address)

	return &MsgUnregisterProviderResponse{}, nil
}

// StakeProvider handles stake increase
func (k Keeper) StakeProvider(msg MsgStakeProvider) (*MsgStakeProviderResponse, error) {
	provider, found := k.GetProvider(msg.Address)
	if !found {
		return nil, types.ErrProviderNotFound
	}

	// Increase stake
	currentStake := parseUint64(provider.StakedAmount.Amount)
	addAmount := parseUint64(msg.Amount)
	provider.StakedAmount.Amount = fmt.Sprintf("%d", currentStake+addAmount)

	if err := k.SetProvider(provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	// Update total staked
	totalStaked := parseUint64(k.GetTotalStaked())
	k.SetTotalStaked(fmt.Sprintf("%d", totalStaked+addAmount))

	return &MsgStakeProviderResponse{
		Address:    msg.Address,
		TotalStake: provider.StakedAmount.Amount,
	}, nil
}

// UnstakeProvider handles stake decrease
func (k Keeper) UnstakeProvider(msg MsgUnstakeProvider) (*MsgUnstakeProviderResponse, error) {
	provider, found := k.GetProvider(msg.Address)
	if !found {
		return nil, types.ErrProviderNotFound
	}

	params := k.GetParams()
	currentStake := parseUint64(provider.StakedAmount.Amount)
	unstakeAmount := parseUint64(msg.Amount)

	// Ensure minimum stake remains
	if currentStake-unstakeAmount < parseUint64(params.MinStake) {
		return nil, types.ErrInsufficientStake
	}

	provider.StakedAmount.Amount = fmt.Sprintf("%d", currentStake-unstakeAmount)

	if err := k.SetProvider(provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	// Update total staked
	totalStaked := parseUint64(k.GetTotalStaked())
	k.SetTotalStaked(fmt.Sprintf("%d", totalStaked-unstakeAmount))

	return &MsgUnstakeProviderResponse{
		Address:        msg.Address,
		RemainingStake: provider.StakedAmount.Amount,
	}, nil
}

// CreateEscrow handles escrow creation
func (k Keeper) CreateEscrow(msg MsgCreateEscrow, blockHeight int64) (*MsgCreateEscrowResponse, error) {
	// Check provider exists and is active
	provider, found := k.GetProvider(msg.Provider)
	if !found {
		return nil, types.ErrProviderNotFound
	}
	if provider.Status != types.ProviderStatusActive {
		return nil, types.ErrProviderNotActive
	}

	// Get escrow ID
	escrowID, err := k.IncrementEscrowID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate escrow ID: %w", err)
	}

	// Calculate fee (1%)
	params := k.GetParams()
	amount := parseUint64(msg.Amount)
	fee := amount / 100 // 1% fee

	// Create escrow
	escrow := types.Escrow{
		ID:        escrowID,
		JobID:     msg.JobID,
		Submitter: msg.Submitter,
		Provider:  msg.Provider,
		Amount:    types.Coin{Denom: "dpc", Amount: fmt.Sprintf("%d", amount)},
		Fee:       types.Coin{Denom: "dpc", Amount: fmt.Sprintf("%d", fee)},
		Status:    types.EscrowStatusLocked,
		Deadline:  blockHeight + int64(params.DisputePeriod),
		CreatedAt: blockHeight,
	}

	if err := k.SetEscrow(escrow); err != nil {
		return nil, fmt.Errorf("failed to store escrow: %w", err)
	}

	// Update total escrowed
	totalEscrowed := parseUint64(k.GetTotalEscrowed())
	k.SetTotalEscrowed(fmt.Sprintf("%d", totalEscrowed+amount))

	log.Printf("[computemarket] Escrow %s created for job %s", escrowID, msg.JobID)

	return &MsgCreateEscrowResponse{
		EscrowID: escrowID,
		Status:   "locked",
	}, nil
}

// ReleaseEscrow handles escrow release
func (k Keeper) ReleaseEscrow(msg MsgReleaseEscrow, blockHeight int64) (*MsgReleaseEscrowResponse, error) {
	escrow, found := k.GetEscrow(msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound
	}

	// Validate submitter
	if escrow.Submitter != msg.Submitter {
		return nil, types.ErrUnauthorized
	}

	// Check escrow status
	if escrow.Status != types.EscrowStatusLocked {
		return nil, types.ErrEscrowAlreadyReleased
	}

	// Release escrow
	escrow.Status = types.EscrowStatusReleased
	escrow.ReleasedAt = blockHeight

	if err := k.SetEscrow(escrow); err != nil {
		return nil, fmt.Errorf("failed to update escrow: %w", err)
	}

	// Update provider stats
	provider, _ := k.GetProvider(escrow.Provider)
	provider.CompletedJobs++
	provider.LastActiveAt = blockHeight
	k.SetProvider(provider)

	// Update total escrowed
	amount := parseUint64(escrow.Amount.Amount)
	totalEscrowed := parseUint64(k.GetTotalEscrowed())
	if totalEscrowed >= amount {
		k.SetTotalEscrowed(fmt.Sprintf("%d", totalEscrowed-amount))
	}

	log.Printf("[computemarket] Escrow %s released to provider %s", msg.EscrowID, escrow.Provider)

	return &MsgReleaseEscrowResponse{
		EscrowID: msg.EscrowID,
		Status:   "released",
	}, nil
}

// RefundEscrow handles escrow refund
func (k Keeper) RefundEscrow(msg MsgRefundEscrow, blockHeight int64) (*MsgRefundEscrowResponse, error) {
	escrow, found := k.GetEscrow(msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound
	}

	// Check escrow status
	if escrow.Status != types.EscrowStatusLocked {
		return nil, types.ErrEscrowNotLocked
	}

	// Refund escrow
	escrow.Status = types.EscrowStatusRefunded
	escrow.RefundedAt = blockHeight

	if err := k.SetEscrow(escrow); err != nil {
		return nil, fmt.Errorf("failed to update escrow: %w", err)
	}

	// Update provider stats
	provider, _ := k.GetProvider(escrow.Provider)
	provider.FailedJobs++
	k.SetProvider(provider)

	// Update total escrowed
	amount := parseUint64(escrow.Amount.Amount)
	totalEscrowed := parseUint64(k.GetTotalEscrowed())
	if totalEscrowed >= amount {
		k.SetTotalEscrowed(fmt.Sprintf("%d", totalEscrowed-amount))
	}

	log.Printf("[computemarket] Escrow %s refunded to submitter", msg.EscrowID)

	return &MsgRefundEscrowResponse{
		EscrowID: msg.EscrowID,
		Status:   "refunded",
	}, nil
}

// OpenDispute handles dispute opening
func (k Keeper) OpenDispute(msg MsgOpenDispute, blockHeight int64) (*MsgOpenDisputeResponse, error) {
	escrow, found := k.GetEscrow(msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound
	}

	// Get dispute ID
	disputeID, err := k.IncrementDisputeID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate dispute ID: %w", err)
	}

	// Create dispute
	dispute := types.Dispute{
		ID:         disputeID,
		EscrowID:   msg.EscrowID,
		JobID:      escrow.JobID,
		Submitter:  escrow.Submitter,
		Provider:   escrow.Provider,
		Reason:     msg.Reason,
		Evidence:   msg.Evidence,
		Status:     types.DisputeStatusOpen,
		OpenedAt:   blockHeight,
	}

	if err := k.SetDispute(dispute); err != nil {
		return nil, fmt.Errorf("failed to store dispute: %w", err)
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusDisputed
	escrow.DisputeID = disputeID
	k.SetEscrow(escrow)

	log.Printf("[computemarket] Dispute %s opened for escrow %s", disputeID, msg.EscrowID)

	return &MsgOpenDisputeResponse{
		DisputeID: disputeID,
		Status:    "open",
	}, nil
}

// ResolveDispute handles dispute resolution
func (k Keeper) ResolveDispute(msg MsgResolveDispute, blockHeight int64) (*MsgResolveDisputeResponse, error) {
	dispute, found := k.GetDispute(msg.DisputeID)
	if !found {
		return nil, types.ErrDisputeNotFound
	}

	// Check dispute status
	if dispute.Status != types.DisputeStatusOpen {
		return nil, types.ErrDisputeAlreadyResolved
	}

	// Resolve dispute
	dispute.Status = types.DisputeStatusResolved
	dispute.Resolution = msg.Resolution
	dispute.Winner = msg.Winner
	dispute.ResolvedAt = blockHeight

	if err := k.SetDispute(dispute); err != nil {
		return nil, fmt.Errorf("failed to update dispute: %w", err)
	}

	// Update escrow
	escrow, _ := k.GetEscrow(dispute.EscrowID)
	if msg.Winner == "submitter" {
		escrow.Status = types.EscrowStatusRefunded
		escrow.RefundedAt = blockHeight
	} else {
		escrow.Status = types.EscrowStatusReleased
		escrow.ReleasedAt = blockHeight
	}
	k.SetEscrow(escrow)

	log.Printf("[computemarket] Dispute %s resolved, winner: %s", msg.DisputeID, msg.Winner)

	return &MsgResolveDisputeResponse{
		DisputeID:  msg.DisputeID,
		Resolution: msg.Resolution,
	}, nil
}

// ProcessTransaction processes a transaction based on type
func (k Keeper) ProcessTransaction(txType string, txData json.RawMessage, blockHeight int64) (interface{}, error) {
	switch txType {
	case "register_provider":
		var msg MsgRegisterProvider
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse register_provider: %w", err)
		}
		return k.RegisterProvider(msg, blockHeight)

	case "unregister_provider":
		var msg MsgUnregisterProvider
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse unregister_provider: %w", err)
		}
		return k.UnregisterProvider(msg)

	case "stake_provider":
		var msg MsgStakeProvider
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse stake_provider: %w", err)
		}
		return k.StakeProvider(msg)

	case "unstake_provider":
		var msg MsgUnstakeProvider
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse unstake_provider: %w", err)
		}
		return k.UnstakeProvider(msg)

	case "create_escrow":
		var msg MsgCreateEscrow
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse create_escrow: %w", err)
		}
		return k.CreateEscrow(msg, blockHeight)

	case "release_escrow":
		var msg MsgReleaseEscrow
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse release_escrow: %w", err)
		}
		return k.ReleaseEscrow(msg, blockHeight)

	case "refund_escrow":
		var msg MsgRefundEscrow
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse refund_escrow: %w", err)
		}
		return k.RefundEscrow(msg, blockHeight)

	case "open_dispute":
		var msg MsgOpenDispute
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse open_dispute: %w", err)
		}
		return k.OpenDispute(msg, blockHeight)

	case "resolve_dispute":
		var msg MsgResolveDispute
		if err := json.Unmarshal(txData, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse resolve_dispute: %w", err)
		}
		return k.ResolveDispute(msg, blockHeight)

	default:
		return nil, fmt.Errorf("unknown transaction type: %s", txType)
	}
}
