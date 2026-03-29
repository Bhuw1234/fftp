// Package keeper implements query handling for the agentwallet module
package keeper

import (
	"encoding/json"

	"github.com/deparrow/dpc/x/agentwallet/types"
)

// QueryWalletResponse is the response for wallet query
type QueryWalletResponse struct {
	Wallet types.AgentWallet `json:"wallet"`
}

// QueryBalanceResponse is the response for balance query
type QueryBalanceResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// QueryParamsResponse is the response for params query
type QueryParamsResponse struct {
	Params types.Params `json:"params"`
}

// HandleQuery handles query requests
func (k Keeper) HandleQuery(path string, data []byte) ([]byte, error) {
	switch path {
	case "/agentwallet/wallet":
		var req struct {
			Address string `json:"address"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryWallet(req.Address)

	case "/agentwallet/balance":
		var req struct {
			Address string `json:"address"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return k.queryBalance(req.Address)

	case "/agentwallet/wallets":
		return k.queryWallets()

	case "/agentwallet/params":
		return k.queryParams()

	default:
		return nil, nil // Not handled by this module
	}
}

func (k Keeper) queryWallet(address string) ([]byte, error) {
	wallet, found := k.GetWallet(address)
	if !found {
		return nil, types.ErrWalletNotFound
	}
	resp := QueryWalletResponse{Wallet: wallet}
	return json.Marshal(resp)
}

func (k Keeper) queryBalance(address string) ([]byte, error) {
	wallet, found := k.GetWallet(address)
	if !found {
		return nil, types.ErrWalletNotFound
	}
	resp := QueryBalanceResponse{
		Address: address,
		Balance: wallet.Balance.Amount,
	}
	return json.Marshal(resp)
}

func (k Keeper) queryWallets() ([]byte, error) {
	wallets := k.GetAllWallets()
	resp := struct {
		Wallets []types.AgentWallet `json:"wallets"`
	}{Wallets: wallets}
	return json.Marshal(resp)
}

func (k Keeper) queryParams() ([]byte, error) {
	params := k.GetParams()
	resp := QueryParamsResponse{Params: params}
	return json.Marshal(resp)
}
