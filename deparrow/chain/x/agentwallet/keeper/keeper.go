// Package keeper implements the state keeper for the agentwallet module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/agentwallet/types"
)

// Keeper manages the state for the agentwallet module
type Keeper struct {
	db *badger.DB
}

// NewKeeper creates a new keeper instance
func NewKeeper(db *badger.DB) Keeper {
	return Keeper{db: db}
}

// GetWallet retrieves a wallet by address
func (k Keeper) GetWallet(address string) (types.AgentWallet, bool) {
	var wallet types.AgentWallet
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyWallet(address))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &wallet)
		})
	})

	if err != nil {
		log.Printf("[agentwallet] Error getting wallet %s: %v", address, err)
		return types.AgentWallet{}, false
	}

	return wallet, found
}

// SetWallet stores a wallet
func (k Keeper) SetWallet(wallet types.AgentWallet) error {
	bz, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyWallet(wallet.Address), bz)
	})
}

// GetAllWallets returns all wallets
func (k Keeper) GetAllWallets() []types.AgentWallet {
	var wallets []types.AgentWallet

	err := k.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := types.WalletKey
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var wallet types.AgentWallet
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &wallet)
			})
			if err != nil {
				continue
			}
			wallets = append(wallets, wallet)
		}
		return nil
	})

	if err != nil {
		log.Printf("[agentwallet] Error getting all wallets: %v", err)
	}

	return wallets
}

// GetNextDIDSuffix returns the next DID suffix number
func (k Keeper) GetNextDIDSuffix() string {
	var suffix string = "1"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.DIDCounterKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			suffix = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[agentwallet] Error getting DID counter: %v", err)
	}

	return suffix
}

// IncrementDIDSuffix increments and returns the next DID suffix
func (k Keeper) IncrementDIDSuffix() (string, error) {
	currentSuffix := k.GetNextDIDSuffix()
	id, _ := strconv.ParseUint(currentSuffix, 10, 64)
	nextSuffix := strconv.FormatUint(id+1, 10)

	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.DIDCounterKey, []byte(nextSuffix))
	})

	if err != nil {
		return "", fmt.Errorf("failed to increment DID suffix: %w", err)
	}

	return currentSuffix, nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams() types.Params {
	return types.DefaultParams()
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
