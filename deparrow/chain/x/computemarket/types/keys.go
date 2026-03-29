// Package types defines the types for the computemarket module
package types

// Module name and store key
const (
	ModuleName = "computemarket"
	StoreKey   = ModuleName
)

// KVStore key prefixes
var (
	// ProviderKey is the prefix for provider entries
	ProviderKey = []byte{0x01}
	// EscrowKey is the prefix for escrow entries
	EscrowKey = []byte{0x02}
	// DisputeKey is the prefix for dispute entries
	DisputeKey = []byte{0x03}
	// JobMatchKey is the prefix for job match entries
	JobMatchKey = []byte{0x04}
	// EscrowCounterKey stores the next escrow ID
	EscrowCounterKey = []byte("next_escrow_id")
	// DisputeCounterKey stores the next dispute ID
	DisputeCounterKey = []byte("next_dispute_id")
	// TotalStakedKey stores total staked amount
	TotalStakedKey = []byte("total_staked")
	// TotalEscrowedKey stores total escrowed amount
	TotalEscrowedKey = []byte("total_escrowed")
)

// KeyProvider returns the key for a provider
func KeyProvider(address string) []byte {
	return append(ProviderKey, []byte(address)...)
}

// KeyEscrow returns the key for an escrow
func KeyEscrow(escrowID string) []byte {
	return append(EscrowKey, []byte(escrowID)...)
}

// KeyDispute returns the key for a dispute
func KeyDispute(disputeID string) []byte {
	return append(DisputeKey, []byte(disputeID)...)
}

// KeyJobMatch returns the key for a job match
func KeyJobMatch(jobID string) []byte {
	return append(JobMatchKey, []byte(jobID)...)
}
