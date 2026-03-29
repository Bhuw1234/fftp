// Package types defines the types for the agentwallet module
package types

// Module name and store key
const (
	ModuleName = "agentwallet"
	StoreKey   = ModuleName
)

// KVStore key prefixes
var (
	// WalletKey is the prefix for wallet entries
	WalletKey = []byte{0x01}
	// DIDKey is the prefix for DID to address mapping
	DIDKey = []byte{0x02}
	// DIDCounterKey stores the next DID suffix
	DIDCounterKey = []byte("next_did_id")
)

// KeyWallet returns the key for a wallet by address
func KeyWallet(address string) []byte {
	return append(WalletKey, []byte(address)...)
}

// KeyDID returns the key for a DID mapping
func KeyDID(did string) []byte {
	return append(DIDKey, []byte(did)...)
}
