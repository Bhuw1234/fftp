// Package keeper implements the state keeper for the computemarket module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/computemarket/types"
)

// Keeper manages the state for the computemarket module
type Keeper struct {
	db *badger.DB
}

// NewKeeper creates a new keeper instance
func NewKeeper(db *badger.DB) Keeper {
	return Keeper{db: db}
}

// Provider methods

// GetProvider retrieves a provider by address
func (k Keeper) GetProvider(address string) (types.Provider, bool) {
	var provider types.Provider
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyProvider(address))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &provider)
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting provider %s: %v", address, err)
		return types.Provider{}, false
	}

	return provider, found
}

// SetProvider stores a provider
func (k Keeper) SetProvider(provider types.Provider) error {
	bz, err := json.Marshal(provider)
	if err != nil {
		return fmt.Errorf("failed to marshal provider: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyProvider(provider.Address), bz)
	})
}

// GetAllProviders returns all providers
func (k Keeper) GetAllProviders() []types.Provider {
	var providers []types.Provider

	err := k.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := types.ProviderKey
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var provider types.Provider
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &provider)
			})
			if err != nil {
				continue
			}
			providers = append(providers, provider)
		}
		return nil
	})

	if err != nil {
		log.Printf("[computemarket] Error getting all providers: %v", err)
	}

	return providers
}

// GetActiveProviders returns all active providers
func (k Keeper) GetActiveProviders() []types.Provider {
	allProviders := k.GetAllProviders()
	var active []types.Provider
	for _, p := range allProviders {
		if p.Status == types.ProviderStatusActive {
			active = append(active, p)
		}
	}
	return active
}

// Escrow methods

// GetEscrow retrieves an escrow by ID
func (k Keeper) GetEscrow(escrowID string) (types.Escrow, bool) {
	var escrow types.Escrow
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyEscrow(escrowID))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &escrow)
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting escrow %s: %v", escrowID, err)
		return types.Escrow{}, false
	}

	return escrow, found
}

// SetEscrow stores an escrow
func (k Keeper) SetEscrow(escrow types.Escrow) error {
	bz, err := json.Marshal(escrow)
	if err != nil {
		return fmt.Errorf("failed to marshal escrow: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyEscrow(escrow.ID), bz)
	})
}

// GetNextEscrowID returns the next available escrow ID
func (k Keeper) GetNextEscrowID() string {
	var escrowID string = "1"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.EscrowCounterKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			escrowID = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting escrow counter: %v", err)
	}

	return escrowID
}

// IncrementEscrowID increments and returns the next escrow ID
func (k Keeper) IncrementEscrowID() (string, error) {
	currentID := k.GetNextEscrowID()
	id, _ := strconv.ParseUint(currentID, 10, 64)
	nextID := strconv.FormatUint(id+1, 10)

	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.EscrowCounterKey, []byte(nextID))
	})

	if err != nil {
		return "", fmt.Errorf("failed to increment escrow ID: %w", err)
	}

	return currentID, nil
}

// GetEscrowsByProvider returns all escrows for a provider
func (k Keeper) GetEscrowsByProvider(provider string) []types.Escrow {
	allEscrows := k.GetAllEscrows()
	var result []types.Escrow
	for _, e := range allEscrows {
		if e.Provider == provider {
			result = append(result, e)
		}
	}
	return result
}

// GetAllEscrows returns all escrows
func (k Keeper) GetAllEscrows() []types.Escrow {
	var escrows []types.Escrow

	err := k.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := types.EscrowKey
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var escrow types.Escrow
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &escrow)
			})
			if err != nil {
				continue
			}
			escrows = append(escrows, escrow)
		}
		return nil
	})

	if err != nil {
		log.Printf("[computemarket] Error getting all escrows: %v", err)
	}

	return escrows
}

// Dispute methods

// GetDispute retrieves a dispute by ID
func (k Keeper) GetDispute(disputeID string) (types.Dispute, bool) {
	var dispute types.Dispute
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyDispute(disputeID))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &dispute)
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting dispute %s: %v", disputeID, err)
		return types.Dispute{}, false
	}

	return dispute, found
}

// SetDispute stores a dispute
func (k Keeper) SetDispute(dispute types.Dispute) error {
	bz, err := json.Marshal(dispute)
	if err != nil {
		return fmt.Errorf("failed to marshal dispute: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyDispute(dispute.ID), bz)
	})
}

// GetNextDisputeID returns the next available dispute ID
func (k Keeper) GetNextDisputeID() string {
	var disputeID string = "1"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.DisputeCounterKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			disputeID = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting dispute counter: %v", err)
	}

	return disputeID
}

// IncrementDisputeID increments and returns the next dispute ID
func (k Keeper) IncrementDisputeID() (string, error) {
	currentID := k.GetNextDisputeID()
	id, _ := strconv.ParseUint(currentID, 10, 64)
	nextID := strconv.FormatUint(id+1, 10)

	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.DisputeCounterKey, []byte(nextID))
	})

	if err != nil {
		return "", fmt.Errorf("failed to increment dispute ID: %w", err)
	}

	return currentID, nil
}

// GetDisputesByProvider returns all disputes for a provider
func (k Keeper) GetDisputesByProvider(provider string) []types.Dispute {
	allDisputes := k.GetAllDisputes()
	var result []types.Dispute
	for _, d := range allDisputes {
		if d.Provider == provider {
			result = append(result, d)
		}
	}
	return result
}

// GetAllDisputes returns all disputes
func (k Keeper) GetAllDisputes() []types.Dispute {
	var disputes []types.Dispute

	err := k.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := types.DisputeKey
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var dispute types.Dispute
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &dispute)
			})
			if err != nil {
				continue
			}
			disputes = append(disputes, dispute)
		}
		return nil
	})

	if err != nil {
		log.Printf("[computemarket] Error getting all disputes: %v", err)
	}

	return disputes
}

// Stats methods

// GetTotalStaked returns the total staked amount
func (k Keeper) GetTotalStaked() string {
	var staked string = "0"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.TotalStakedKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			staked = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting total staked: %v", err)
	}

	return staked
}

// SetTotalStaked sets the total staked amount
func (k Keeper) SetTotalStaked(amount string) error {
	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.TotalStakedKey, []byte(amount))
	})
}

// GetTotalEscrowed returns the total escrowed amount
func (k Keeper) GetTotalEscrowed() string {
	var escrowed string = "0"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.TotalEscrowedKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			escrowed = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[computemarket] Error getting total escrowed: %v", err)
	}

	return escrowed
}

// SetTotalEscrowed sets the total escrowed amount
func (k Keeper) SetTotalEscrowed(amount string) error {
	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.TotalEscrowedKey, []byte(amount))
	})
}

// GetParams returns the module parameters
func (k Keeper) GetParams() types.Params {
	return types.DefaultParams()
}

// GetStats returns module statistics
func (k Keeper) GetStats() map[string]interface{} {
	providers := k.GetAllProviders()
	var activeCount uint64
	for _, p := range providers {
		if p.Status == types.ProviderStatusActive {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_staked":     k.GetTotalStaked(),
		"total_escrowed":   k.GetTotalEscrowed(),
		"active_providers": activeCount,
		"total_providers":  uint64(len(providers)),
	}
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
