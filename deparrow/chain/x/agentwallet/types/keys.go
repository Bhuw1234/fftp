package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KVStore keys
var (
	// WalletKey is the prefix for wallet store
	WalletKey = []byte{0x01}

	// WalletByDIDKey is the prefix for wallets indexed by DID
	WalletByDIDKey = []byte{0x02}

	// WalletByAddressKey is the prefix for wallets indexed by address
	WalletByAddressKey = []byte{0x03}

	// SpendingRuleKey is the prefix for spending rules
	SpendingRuleKey = []byte{0x04}

	// AutomationRuleKey is the prefix for automation rules
	AutomationRuleKey = []byte{0x05}

	// DailySpendingKey tracks daily spending per wallet
	DailySpendingKey = []byte{0x06}

	// TransactionHistoryKey stores transaction history
	TransactionHistoryKey = []byte{0x07}

	// ParamsKey is the prefix for params store
	ParamsKey = []byte{0x10}

	// PendingAutomationKey stores pending automation triggers
	PendingAutomationKey = []byte{0x11}

	// EmergencyReserveKey stores emergency reserves
	EmergencyReserveKey = []byte{0x12}

	// AgentRegistryKey stores agent registry
	AgentRegistryKey = []byte{0x13}
)

// uint64ToBigEndian converts uint64 to 8 bytes big endian
func uint64ToBigEndian(n uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, n)
	return bz
}

// int64ToBigEndian converts int64 to 8 bytes big endian
func int64ToBigEndian(n int64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(n))
	return bz
}

// GetWalletKey returns the store key for a wallet by address
func GetWalletKey(address string) []byte {
	return append(WalletKey, []byte(address)...)
}

// GetWalletByDIDKey returns the store key for wallets by DID
func GetWalletByDIDKey(did string) []byte {
	return append(WalletByDIDKey, []byte(did)...)
}

// GetWalletByAddressKey returns the store key for wallet by address
func GetWalletByAddressKey(address sdk.AccAddress) []byte {
	return append(WalletByAddressKey, address...)
}

// GetSpendingRuleKey returns the store key for spending rules
func GetSpendingRuleKey(address string, ruleIndex uint32) []byte {
	key := append(SpendingRuleKey, []byte(address)...)
	return append(key, uint32ToBigEndian(ruleIndex)...)
}

// GetAutomationRuleKey returns the store key for automation rules
func GetAutomationRuleKey(address string, ruleIndex uint32) []byte {
	key := append(AutomationRuleKey, []byte(address)...)
	return append(key, uint32ToBigEndian(ruleIndex)...)
}

// GetDailySpendingKey returns the store key for daily spending tracking
// Format: prefix | address | day (unix timestamp / 86400)
func GetDailySpendingKey(address string, day int64) []byte {
	key := append(DailySpendingKey, []byte(address)...)
	return append(key, int64ToBigEndian(day)...)
}

// GetTransactionHistoryKey returns the store key for transaction history
func GetTransactionHistoryKey(address string, txIndex uint64) []byte {
	key := append(TransactionHistoryKey, []byte(address)...)
	return append(key, uint64ToBigEndian(txIndex)...)
}

// GetPendingAutomationKey returns the store key for pending automations
func GetPendingAutomationKey(address string) []byte {
	return append(PendingAutomationKey, []byte(address)...)
}

// GetEmergencyReserveKey returns the store key for emergency reserves
func GetEmergencyReserveKey(address string) []byte {
	return append(EmergencyReserveKey, []byte(address)...)
}

// GetAgentRegistryKey returns the store key for agent registry
func GetAgentRegistryKey(did string) []byte {
	return append(AgentRegistryKey, []byte(did)...)
}

// uint32ToBigEndian converts uint32 to 4 bytes big endian
func uint32ToBigEndian(n uint32) []byte {
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, n)
	return bz
}
