package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KVStore keys
var (
	// ProviderKey is the prefix for provider store
	ProviderKey = []byte{0x01}

	// ProviderByReputationKey is the prefix for providers indexed by reputation
	ProviderByReputationKey = []byte{0x02}

	// EscrowKey is the prefix for escrow store
	EscrowKey = []byte{0x03}

	// EscrowByJobKey is the prefix for escrows indexed by job ID
	EscrowByJobKey = []byte{0x04}

	// EscrowByProviderKey is the prefix for escrows indexed by provider
	EscrowByProviderKey = []byte{0x05}

	// DisputeKey is the prefix for dispute store
	DisputeKey = []byte{0x06}

	// DisputeByEscrowKey is the prefix for disputes indexed by escrow
	DisputeByEscrowKey = []byte{0x07}

	// ParamsKey is the prefix for params store
	ParamsKey = []byte{0x10}

	// ActiveProvidersKey stores the list of active providers
	ActiveProvidersKey = []byte{0x11}

	// ProviderCapabilitiesKey stores provider capabilities index
	ProviderCapabilitiesKey = []byte{0x12}

	// JobMatchKey stores job-provider matches
	JobMatchKey = []byte{0x13}
)

// uint32ToBigEndian converts uint32 to 4 bytes big endian
func uint32ToBigEndian(n uint32) []byte {
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, n)
	return bz
}

// GetProviderKey returns the store key for a provider
func GetProviderKey(address string) []byte {
	return append(ProviderKey, []byte(address)...)
}

// GetProviderByReputationKey returns the store key for providers by reputation
// Format: prefix | reputation_score (4 bytes big endian) | address
func GetProviderByReputationKey(reputation uint32, address string) []byte {
	key := append(ProviderByReputationKey, uint32ToBigEndian(reputation)...)
	return append(key, []byte(address)...)
}

// GetEscrowKey returns the store key for an escrow
func GetEscrowKey(escrowID string) []byte {
	return append(EscrowKey, []byte(escrowID)...)
}

// GetEscrowByJobKey returns the store key for escrow by job ID
func GetEscrowByJobKey(jobID string) []byte {
	return append(EscrowByJobKey, []byte(jobID)...)
}

// GetEscrowByProviderKey returns the store key for escrows by provider
func GetEscrowByProviderKey(provider sdk.AccAddress, escrowID string) []byte {
	return append(append(EscrowByProviderKey, provider...), []byte(escrowID)...)
}

// GetDisputeKey returns the store key for a dispute
func GetDisputeKey(disputeID string) []byte {
	return append(DisputeKey, []byte(disputeID)...)
}

// GetDisputeByEscrowKey returns the store key for dispute by escrow
func GetDisputeByEscrowKey(escrowID string) []byte {
	return append(DisputeByEscrowKey, []byte(escrowID)...)
}

// GetActiveProvidersKey returns the key for active providers list
func GetActiveProvidersKey() []byte {
	return ActiveProvidersKey
}

// GetProviderCapabilitiesKey returns the key for provider capabilities
func GetProviderCapabilitiesKey(provider string) []byte {
	return append(ProviderCapabilitiesKey, []byte(provider)...)
}

// GetJobMatchKey returns the key for a job match
func GetJobMatchKey(jobID string) []byte {
	return append(JobMatchKey, []byte(jobID)...)
}