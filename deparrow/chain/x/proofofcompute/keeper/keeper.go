// Package keeper implements the state keeper for the proofofcompute module
package keeper

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// Keeper manages the state for the proofofcompute module
type Keeper struct {
	db *badger.DB
}

// NewKeeper creates a new keeper instance
func NewKeeper(db *badger.DB) Keeper {
	return Keeper{db: db}
}

// GetJob retrieves a job by ID
func (k Keeper) GetJob(jobID string) (types.Job, bool) {
	var job types.Job
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyJob(jobID))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &job)
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting job %s: %v", jobID, err)
		return types.Job{}, false
	}

	return job, found
}

// SetJob stores a job
func (k Keeper) SetJob(job types.Job) error {
	bz, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyJob(job.ID), bz)
	})
}

// DeleteJob removes a job
func (k Keeper) DeleteJob(jobID string) error {
	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(types.KeyJob(jobID))
	})
}

// GetNextJobID returns the next available job ID
func (k Keeper) GetNextJobID() string {
	var jobID string = "1"

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.JobCounterKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			jobID = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting job counter: %v", err)
	}

	return jobID
}

// IncrementJobID increments and returns the next job ID
func (k Keeper) IncrementJobID() (string, error) {
	currentID := k.GetNextJobID()
	id, _ := strconv.ParseUint(currentID, 10, 64)
	nextID := strconv.FormatUint(id+1, 10)

	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.JobCounterKey, []byte(nextID))
	})

	if err != nil {
		return "", fmt.Errorf("failed to increment job ID: %w", err)
	}

	return currentID, nil
}

// GetAllJobs returns all jobs
func (k Keeper) GetAllJobs() []types.Job {
	var jobs []types.Job

	err := k.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := types.JobKey
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var job types.Job
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &job)
			})
			if err != nil {
				continue
			}
			jobs = append(jobs, job)
		}
		return nil
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting all jobs: %v", err)
	}

	return jobs
}

// GetJobsBySubmitter returns all jobs by a submitter
func (k Keeper) GetJobsBySubmitter(submitter string) []types.Job {
	allJobs := k.GetAllJobs()
	var result []types.Job
	for _, job := range allJobs {
		if job.Submitter == submitter {
			result = append(result, job)
		}
	}
	return result
}

// GetJobsByComputeNode returns all jobs executed by a compute node
func (k Keeper) GetJobsByComputeNode(nodeAddress string) []types.Job {
	allJobs := k.GetAllJobs()
	var result []types.Job
	for _, job := range allJobs {
		if job.ComputeNode == nodeAddress {
			result = append(result, job)
		}
	}
	return result
}

// GetProof retrieves a proof by job ID
func (k Keeper) GetProof(jobID string) (types.ComputeProof, bool) {
	var proof types.ComputeProof
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyProof(jobID))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &proof)
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting proof %s: %v", jobID, err)
		return types.ComputeProof{}, false
	}

	return proof, found
}

// SetProof stores a proof
func (k Keeper) SetProof(proof types.ComputeProof) error {
	bz, err := json.Marshal(proof)
	if err != nil {
		return fmt.Errorf("failed to marshal proof: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyProof(proof.JobID), bz)
	})
}

// GetPendingReward gets pending rewards for a node
func (k Keeper) GetPendingReward(nodeAddress string) (types.PendingReward, bool) {
	var reward types.PendingReward
	found := false

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.KeyPendingReward(nodeAddress))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &reward)
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting pending reward for %s: %v", nodeAddress, err)
		return types.PendingReward{}, false
	}

	return reward, found
}

// SetPendingReward sets pending rewards for a node
func (k Keeper) SetPendingReward(reward types.PendingReward) error {
	bz, err := json.Marshal(reward)
	if err != nil {
		return fmt.Errorf("failed to marshal pending reward: %w", err)
	}

	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.KeyPendingReward(reward.NodeAddress), bz)
	})
}

// AddPendingReward adds to the pending reward for a node
func (k Keeper) AddPendingReward(nodeAddress, rewardAmount string) error {
	reward, found := k.GetPendingReward(nodeAddress)
	if !found {
		reward = types.PendingReward{
			NodeAddress:   nodeAddress,
			Amount:        "0",
			JobsCompleted: 0,
		}
	}

	// Add amounts (string to uint addition)
	currentAmount := parseUint64(reward.Amount)
	addAmount := parseUint64(rewardAmount)
	newAmount := currentAmount + addAmount

	reward.Amount = fmt.Sprintf("%d", newAmount)
	reward.JobsCompleted++

	return k.SetPendingReward(reward)
}

// ClearPendingReward clears pending rewards for a node and returns the amount
func (k Keeper) ClearPendingReward(nodeAddress string) (string, error) {
	reward, found := k.GetPendingReward(nodeAddress)
	if !found {
		return "0", nil
	}

	err := k.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(types.KeyPendingReward(nodeAddress))
	})

	if err != nil {
		return "", fmt.Errorf("failed to clear pending reward: %w", err)
	}

	return reward.Amount, nil
}

// GetTotalSupply returns the total minted supply
func (k Keeper) GetTotalSupply() string {
	var supply string = "1000000000000000000000000000" // Default 1B

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.TotalSupplyKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			supply = string(val)
			return nil
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting total supply: %v", err)
	}

	return supply
}

// SetTotalSupply sets the total minted supply
func (k Keeper) SetTotalSupply(supply string) error {
	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.TotalSupplyKey, []byte(supply))
	})
}

// AddToTotalSupply adds minted tokens to total supply
func (k Keeper) AddToTotalSupply(amount string) error {
	current := parseUint64(k.GetTotalSupply())
	add := parseUint64(amount)
	newSupply := current + add

	return k.SetTotalSupply(fmt.Sprintf("%d", newSupply))
}

// GetCurrentDifficulty returns the current difficulty
func (k Keeper) GetCurrentDifficulty() uint64 {
	var difficulty uint64 = 1

	err := k.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(types.DifficultyKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			d, err := strconv.ParseUint(string(val), 10, 64)
			if err != nil {
				return err
			}
			difficulty = d
			return nil
		})
	})

	if err != nil {
		log.Printf("[proofofcompute] Error getting difficulty: %v", err)
	}

	return difficulty
}

// SetCurrentDifficulty sets the current difficulty
func (k Keeper) SetCurrentDifficulty(difficulty uint64) error {
	return k.db.Update(func(txn *badger.Txn) error {
		return txn.Set(types.DifficultyKey, []byte(strconv.FormatUint(difficulty, 10)))
	})
}

// GetParams returns the module parameters
func (k Keeper) GetParams() types.Params {
	// Return default params for now
	// In a full implementation, these would be stored in state
	return types.DefaultParams()
}

// CalculateReward calculates the reward for a job
// Formula: DPC = 0.001 × Complexity × ComputeUnits
func (k Keeper) CalculateReward(computeUnits uint64, complexity uint32) string {
	// Reward per unit is 0.001 DPC = 1000000000000000 base units (18 decimals)
	rewardPerUnit := uint64(1000000000000000)
	reward := rewardPerUnit * uint64(complexity) * computeUnits
	return fmt.Sprintf("%d", reward)
}

// GetStats returns module statistics
func (k Keeper) GetStats() map[string]interface{} {
	jobs := k.GetAllJobs()
	var completed, pending, running int
	for _, job := range jobs {
		switch job.Status {
		case types.JobStatusCompleted:
			completed++
		case types.JobStatusPending:
			pending++
		case types.JobStatusRunning:
			running++
		}
	}

	return map[string]interface{}{
		"total_jobs":     len(jobs),
		"completed_jobs": completed,
		"pending_jobs":   pending,
		"running_jobs":   running,
		"total_supply":   k.GetTotalSupply(),
		"difficulty":     k.GetCurrentDifficulty(),
	}
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
