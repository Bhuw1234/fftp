package types

// Event types for the computemarket module
const (
	EventTypeProviderRegistered = "provider_registered"
	EventTypeProviderUnregistered = "provider_unregistered"
	EventTypeProviderStaked     = "provider_staked"
	EventTypeProviderUnstaked   = "provider_unstaked"
	EventTypeProviderSlashed    = "provider_slashed"
	EventTypeReputationUpdated  = "reputation_updated"
	EventTypeEscrowCreated      = "escrow_created"
	EventTypeEscrowLocked       = "escrow_locked"
	EventTypeEscrowReleased     = "escrow_released"
	EventTypeEscrowRefunded     = "escrow_refunded"
	EventTypeEscrowDisputed     = "escrow_disputed"
	EventTypeDisputeOpened      = "dispute_opened"
	EventTypeDisputeResolved    = "dispute_resolved"
	EventTypeJobMatched         = "job_matched"
	EventTypeMatchingFailed     = "matching_failed"

	AttributeKeyProvider        = "provider"
	AttributeKeySubmitter       = "submitter"
	AttributeKeyEscrowID        = "escrow_id"
	AttributeKeyJobID           = "job_id"
	AttributeKeyAmount          = "amount"
	AttributeKeyStake           = "stake"
	AttributeKeyReputation      = "reputation"
	AttributeKeyStatus          = "status"
	AttributeKeyReason          = "reason"
	AttributeKeySlashedAmount   = "slashed_amount"
	AttributeKeyNewReputation   = "new_reputation"
	AttributeKeyOldReputation   = "old_reputation"
	AttributeKeyDeadline        = "deadline"
	AttributeKeyResolution      = "resolution"
	AttributeKeyWinner          = "winner"
	AttributeKeyCapabilities    = "capabilities"
	AttributeKeyMatchScore      = "match_score"
)

// Dispute resolution types
const (
	ResolutionSubmitterWins = "submitter_wins"
	ResolutionProviderWins  = "provider_wins"
	ResolutionSplit         = "split"
)

// Dispute reasons
const (
	ReasonJobFailed      = "job_failed"
	ReasonOutputInvalid  = "output_invalid"
	ReasonTimeout        = "timeout"
	ReasonMalicious      = "malicious"
	ReasonOther          = "other"
)
