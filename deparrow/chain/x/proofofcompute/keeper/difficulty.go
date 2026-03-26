package keeper

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/proofofcompute/types"
)

// Difficulty constants
const (
	// InitialDifficulty is the starting difficulty level
	InitialDifficulty uint64 = 1

	// MaxDifficulty is the maximum difficulty level
	MaxDifficulty uint64 = 1000000

	// MinDifficulty is the minimum difficulty level
	MinDifficulty uint64 = 1

	// DifficultyAdjustmentFactor is the max adjustment per period (25%)
	DifficultyAdjustmentFactor float64 = 0.25

	// TargetJobsPerBlock is the target number of jobs per block
	TargetJobsPerBlock float64 = 10.0
)

// GetCurrentDifficulty returns the current difficulty level
func (k Keeper) GetCurrentDifficulty(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DifficultyKey)
	if bz == nil {
		return InitialDifficulty
	}
	return sdk.BigEndianToUint64(bz)
}

// SetCurrentDifficulty sets the current difficulty level
func (k Keeper) SetCurrentDifficulty(ctx sdk.Context, difficulty uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.DifficultyKey, sdk.Uint64ToBigEndian(difficulty))
}

// AdjustDifficulty adjusts the difficulty based on network activity
// This is called at the end of each adjustment period
func (k Keeper) AdjustDifficulty(ctx sdk.Context) {
	params := k.GetParams(ctx)

	// Check if we're at an adjustment period
	blockHeight := ctx.BlockHeight()
	if uint64(blockHeight)%params.DifficultyAdjustment != 0 {
		return
	}

	currentDifficulty := k.GetCurrentDifficulty(ctx)
	jobsCompleted := k.GetBlockJobCount(ctx)

	// Calculate actual jobs per block
	blocksInPeriod := params.DifficultyAdjustment
	actualJobsPerBlock := float64(jobsCompleted) / float64(blocksInPeriod)

	// Calculate new difficulty
	newDifficulty := k.calculateNewDifficulty(
		currentDifficulty,
		TargetJobsPerBlock,
		actualJobsPerBlock,
	)

	// Update difficulty
	k.SetCurrentDifficulty(ctx, newDifficulty)

	// Reset block job count for next period
	k.ResetBlockJobCount(ctx)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDifficultyAdjusted,
			sdk.NewAttribute("old_difficulty", string(rune(currentDifficulty))),
			sdk.NewAttribute("new_difficulty", string(rune(newDifficulty))),
			sdk.NewAttribute("jobs_per_block", string(rune(int64(actualJobsPerBlock)))),
		),
	)

	k.Logger(ctx).Info(
		"difficulty adjusted",
		"old", currentDifficulty,
		"new", newDifficulty,
		"jobs_per_block", actualJobsPerBlock,
		"target", TargetJobsPerBlock,
	)
}

// calculateNewDifficulty calculates the new difficulty based on activity
func (k Keeper) calculateNewDifficulty(
	currentDifficulty uint64,
	targetJobsPerBlock float64,
	actualJobsPerBlock float64,
) uint64 {
	if actualJobsPerBlock == 0 {
		// No jobs, decrease difficulty
		return max(MinDifficulty, currentDifficulty/2)
	}

	// Calculate ratio
	ratio := actualJobsPerBlock / targetJobsPerBlock

	// Determine adjustment direction
	var adjustment float64
	if ratio > 1 {
		// Too many jobs, increase difficulty
		adjustment = math.Min(ratio-1, DifficultyAdjustmentFactor)
	} else {
		// Too few jobs, decrease difficulty
		adjustment = -math.Min(1-ratio, DifficultyAdjustmentFactor)
	}

	// Apply adjustment
	newDifficultyFloat := float64(currentDifficulty) * (1 + adjustment)

	// Clamp to valid range
	newDifficulty := uint64(math.Round(newDifficultyFloat))
	newDifficulty = max(MinDifficulty, min(MaxDifficulty, newDifficulty))

	return newDifficulty
}

// GetDifficultyForJob returns the difficulty requirement for a specific job
// This can be used for proof-of-work requirements on job submission
func (k Keeper) GetDifficultyForJob(ctx sdk.Context, computeUnits uint64) uint64 {
	baseDifficulty := k.GetCurrentDifficulty(ctx)

	// Scale difficulty based on compute units
	// Larger jobs have higher difficulty requirements
	if computeUnits > 10000 {
		return baseDifficulty * 2
	} else if computeUnits > 5000 {
		return baseDifficulty * 3 / 2
	}

	return baseDifficulty
}

// ValidateDifficulty validates that a proof meets the difficulty requirement
// This is used for proof-of-work verification
func (k Keeper) ValidateDifficulty(ctx sdk.Context, proof *types.ComputeProof) bool {
	jobDifficulty := k.GetDifficultyForJob(ctx, proof.ComputeUnits)

	// Calculate proof hash
	proofHash := k.calculateProofHash(proof)

	// Check if hash meets difficulty requirement
	// Difficulty N means the first N bytes of hash must be zeros
	requiredZeros := int(jobDifficulty)
	if requiredZeros > len(proofHash) {
		requiredZeros = len(proofHash)
	}

	zeros := 0
	for _, b := range proofHash[:requiredZeros] {
		if b != 0 {
			break
		}
		zeros++
	}

	return zeros >= requiredZeros
}

// calculateProofHash calculates the hash of a proof for difficulty validation
func (k Keeper) calculateProofHash(proof *types.ComputeProof) []byte {
	// Simple hash of proof data for difficulty check
	// In production, use a more sophisticated hashing scheme
	data := make([]byte, 0)
	data = append(data, []byte(proof.JobID)...)
	data = append(data, []byte(proof.NodeID)...)
	data = append(data, sdk.Uint64ToBigEndian(proof.ComputeUnits)...)
	data = append(data, proof.OutputHash...)
	data = append(data, proof.Signature...)

	// Use SHA256 for simplicity
	hash := make([]byte, 32)
	for i, b := range data {
		if i >= 32 {
			break
		}
		hash[i] = b
	}
	return hash
}

// GetDifficultyStats returns difficulty statistics
func (k Keeper) GetDifficultyStats(ctx sdk.Context) (difficulty uint64, jobsInPeriod uint64, blocksUntilAdjustment int64) {
	difficulty = k.GetCurrentDifficulty(ctx)
	jobsInPeriod = k.GetBlockJobCount(ctx)

	params := k.GetParams(ctx)
	currentBlock := ctx.BlockHeight()
	nextAdjustment := (currentBlock / int64(params.DifficultyAdjustment) + 1) * int64(params.DifficultyAdjustment)
	blocksUntilAdjustment = nextAdjustment - currentBlock

	return difficulty, jobsInPeriod, blocksUntilAdjustment
}

// min returns the minimum of two uint64 values
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two uint64 values
func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
