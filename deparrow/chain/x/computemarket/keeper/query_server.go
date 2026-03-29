// Package keeper implements query handling for the computemarket module
package keeper

import (
	"encoding/json"

	"github.com/deparrow/dpc/x/computemarket/types"
)

// QueryProviderResponse is the response for provider query
type QueryProviderResponse struct {
	Provider types.Provider `json:"provider"`
}

// QueryProvidersResponse is the response for providers list query
type QueryProvidersResponse struct {
	Providers []types.Provider `json:"providers"`
}

// QueryEscrowResponse is the response for escrow query
type QueryEscrowResponse struct {
	Escrow types.Escrow `json:"escrow"`
}

// QueryDisputeResponse is the response for dispute query
type QueryDisputeResponse struct {
	Dispute types.Dispute `json:"dispute"`
}

// QueryParamsResponse is the response for params query
type QueryParamsResponse struct {
	Params types.Params `json:"params"`
}

// QueryStatsResponse is the response for stats query
type QueryStatsResponse struct {
	TotalStaked     string `json:"total_staked"`
	TotalEscrowed   string `json:"total_escrowed"`
	ActiveProviders uint64 `json:"active_providers"`
	TotalProviders  uint64 `json:"total_providers"`
}

// HandleQuery handles query requests
func (k Keeper) HandleQuery(path string, data []byte) ([]byte, error) {
	switch path {
	case "/computemarket/provider":
		var req struct {
			Address string `json:"address"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryProvider(req.Address)

	case "/computemarket/providers":
		return k.queryProviders()

	case "/computemarket/active_providers":
		return k.queryActiveProviders()

	case "/computemarket/escrow":
		var req struct {
			EscrowID string `json:"escrow_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryEscrow(req.EscrowID)

	case "/computemarket/escrows_by_provider":
		var req struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryEscrowsByProvider(req.Provider)

	case "/computemarket/dispute":
		var req struct {
			DisputeID string `json:"dispute_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryDispute(req.DisputeID)

	case "/computemarket/disputes_by_provider":
		var req struct {
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryDisputesByProvider(req.Provider)

	case "/computemarket/params":
		return k.queryParams()

	case "/computemarket/stats":
		return k.queryStats()

	default:
		return nil, nil // Not handled by this module
	}
}

func (k Keeper) queryProvider(address string) ([]byte, error) {
	provider, found := k.GetProvider(address)
	if !found {
		return nil, types.ErrProviderNotFound
	}
	resp := QueryProviderResponse{Provider: provider}
	return json.Marshal(resp)
}

func (k Keeper) queryProviders() ([]byte, error) {
	providers := k.GetAllProviders()
	resp := QueryProvidersResponse{Providers: providers}
	return json.Marshal(resp)
}

func (k Keeper) queryActiveProviders() ([]byte, error) {
	providers := k.GetActiveProviders()
	var addresses []string
	for _, p := range providers {
		addresses = append(addresses, p.Address)
	}
	resp := struct {
		Providers []string `json:"providers"`
	}{Providers: addresses}
	return json.Marshal(resp)
}

func (k Keeper) queryEscrow(escrowID string) ([]byte, error) {
	escrow, found := k.GetEscrow(escrowID)
	if !found {
		return nil, types.ErrEscrowNotFound
	}
	resp := QueryEscrowResponse{Escrow: escrow}
	return json.Marshal(resp)
}

func (k Keeper) queryEscrowsByProvider(provider string) ([]byte, error) {
	escrows := k.GetEscrowsByProvider(provider)
	resp := struct {
		Escrows []types.Escrow `json:"escrows"`
	}{Escrows: escrows}
	return json.Marshal(resp)
}

func (k Keeper) queryDispute(disputeID string) ([]byte, error) {
	dispute, found := k.GetDispute(disputeID)
	if !found {
		return nil, types.ErrDisputeNotFound
	}
	resp := QueryDisputeResponse{Dispute: dispute}
	return json.Marshal(resp)
}

func (k Keeper) queryDisputesByProvider(provider string) ([]byte, error) {
	disputes := k.GetDisputesByProvider(provider)
	resp := struct {
		Disputes []types.Dispute `json:"disputes"`
	}{Disputes: disputes}
	return json.Marshal(resp)
}

func (k Keeper) queryParams() ([]byte, error) {
	params := k.GetParams()
	resp := QueryParamsResponse{Params: params}
	return json.Marshal(resp)
}

func (k Keeper) queryStats() ([]byte, error) {
	stats := k.GetStats()
	resp := QueryStatsResponse{
		TotalStaked:     stats["total_staked"].(string),
		TotalEscrowed:   stats["total_escrowed"].(string),
		ActiveProviders: stats["active_providers"].(uint64),
		TotalProviders:  stats["total_providers"].(uint64),
	}
	return json.Marshal(resp)
}
