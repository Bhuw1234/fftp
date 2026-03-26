package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// MatchJob finds the best provider for a job based on requirements
func (k Keeper) MatchJob(ctx sdk.Context, jobID string, requirements types.ProviderCapabilities, excludeProviders []string) (*types.JobMatch, error) {
	// Get all active providers
	activeProviders := k.GetActiveProviders(ctx)
	if len(activeProviders) == 0 {
		return nil, types.ErrNoProvidersAvailable
	}

	// Create exclusion set
	excludeSet := make(map[string]bool)
	for _, addr := range excludeProviders {
		excludeSet[addr] = true
	}

	// Score and filter providers
	var candidates []providerScore
	for _, addr := range activeProviders {
		// Skip excluded providers
		if excludeSet[addr] {
			continue
		}

		provider, err := k.GetProvider(ctx, addr)
		if err != nil {
			continue
		}

		// Skip inactive or slashed providers
		if !provider.IsActive() {
			continue
		}

		// Check capabilities match
		if !provider.Capabilities.Matches(requirements) {
			continue
		}

		// Check reputation threshold
		params := k.GetParams(ctx)
		if provider.ReputationScore < params.MinReputation {
			continue
		}

		// Calculate match score
		score := provider.MatchScore(requirements)
		if score > 0 {
			candidates = append(candidates, providerScore{
				address: addr,
				score:   score,
				provider: provider,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, types.ErrNoProvidersAvailable
	}

	// Sort by score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// Select best provider
	best := candidates[0]

	// Create job match
	match := types.NewJobMatch(jobID, best.address, best.score, requirements)
	if err := k.SetJobMatch(ctx, match); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeJobMatched,
			sdk.NewAttribute(types.AttributeKeyJobID, jobID),
			sdk.NewAttribute(types.AttributeKeyProvider, best.address),
			sdk.NewAttribute(types.AttributeKeyMatchScore, string(rune(best.score))),
		),
	)

	return match, nil
}

// providerScore holds a provider and their match score
type providerScore struct {
	address  string
	score    uint32
	provider *types.ProviderExtended
}

// SetJobMatch stores a job match
func (k Keeper) SetJobMatch(ctx sdk.Context, match *types.JobMatch) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(match)
	store.Set(types.GetJobMatchKey(match.JobID), bz)
	return nil
}

// GetJobMatch retrieves a job match by job ID
func (k Keeper) GetJobMatch(ctx sdk.Context, jobID string) (*types.JobMatch, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetJobMatchKey(jobID)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrJobMatchNotFound
	}

	var match types.JobMatch
	k.cdc.MustUnmarshal(bz, &match)
	return &match, nil
}

// DeleteJobMatch removes a job match
func (k Keeper) DeleteJobMatch(ctx sdk.Context, jobID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetJobMatchKey(jobID))
}

// GetProvidersForRequirements finds all providers matching requirements
func (k Keeper) GetProvidersForRequirements(ctx sdk.Context, requirements types.ProviderCapabilities) ([]types.ProviderExtended, error) {
	activeProviders := k.GetActiveProviders(ctx)
	var matches []types.ProviderExtended

	for _, addr := range activeProviders {
		provider, err := k.GetProvider(ctx, addr)
		if err != nil {
			continue
		}

		if provider.IsActive() && provider.Capabilities.Matches(requirements) {
			matches = append(matches, *provider)
		}
	}

	return matches, nil
}

// MatchByRegion finds providers in specific regions
func (k Keeper) MatchByRegion(ctx sdk.Context, requirements types.ProviderCapabilities, preferredRegions []string) ([]types.ProviderExtended, error) {
	providers, err := k.GetProvidersForRequirements(ctx, requirements)
	if err != nil {
		return nil, err
	}

	// Filter by preferred regions
	var matches []types.ProviderExtended
	for _, provider := range providers {
		for _, region := range preferredRegions {
			for _, providerRegion := range provider.Capabilities.Regions {
				if providerRegion == region {
					matches = append(matches, provider)
					break
				}
			}
		}
	}

	return matches, nil
}

// MatchByReputation finds providers with minimum reputation
func (k Keeper) MatchByReputation(ctx sdk.Context, requirements types.ProviderCapabilities, minReputation uint32) ([]types.ProviderExtended, error) {
	providers, err := k.GetProvidersForRequirements(ctx, requirements)
	if err != nil {
		return nil, err
	}

	// Filter by reputation
	var matches []types.ProviderExtended
	for _, provider := range providers {
		if provider.ReputationScore >= minReputation {
			matches = append(matches, provider)
		}
	}

	// Sort by reputation (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ReputationScore > matches[j].ReputationScore
	})

	return matches, nil
}

// MatchByStake finds providers with minimum stake
func (k Keeper) MatchByStake(ctx sdk.Context, requirements types.ProviderCapabilities, minStake sdk.Int) ([]types.ProviderExtended, error) {
	providers, err := k.GetProvidersForRequirements(ctx, requirements)
	if err != nil {
		return nil, err
	}

	// Filter by stake
	var matches []types.ProviderExtended
	for _, provider := range providers {
		if provider.StakedAmount.Amount.GTE(minStake) {
			matches = append(matches, provider)
		}
	}

	// Sort by stake (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].StakedAmount.Amount.GT(matches[j].StakedAmount.Amount)
	})

	return matches, nil
}

// FindBestProviders returns the top N providers for a job
func (k Keeper) FindBestProviders(ctx sdk.Context, requirements types.ProviderCapabilities, topN int) ([]types.JobMatch, error) {
	providers, err := k.GetProvidersForRequirements(ctx, requirements)
	if err != nil {
		return nil, err
	}

	if len(providers) == 0 {
		return nil, types.ErrNoProvidersAvailable
	}

	// Calculate scores and sort
	var scored []providerScore
	for _, provider := range providers {
		score := provider.MatchScore(requirements)
		if score > 0 {
			scored = append(scored, providerScore{
				address:  provider.Address,
				score:    score,
				provider: &provider,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top N
	if topN > len(scored) {
		topN = len(scored)
	}

	var matches []types.JobMatch
	for i := 0; i < topN; i++ {
		matches = append(matches, types.JobMatch{
			Provider:     scored[i].address,
			Score:        scored[i].score,
			Requirements: requirements,
		})
	}

	return matches, nil
}

// CalculateMatchScore calculates a match score for a provider
func CalculateMatchScore(provider *types.ProviderExtended, requirements types.ProviderCapabilities) uint32 {
	return provider.MatchScore(requirements)
}
