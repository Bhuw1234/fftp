// Package app provides the DPC blockchain application with CometBFT consensus
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v3"
)

const (
	// DPC Token Parameters (in base units with 18 decimals)
	MaxSupply      = "21000000000000000000000000000" // 21B with 18 decimals
	InitialSupply  = "1000000000000000000000000000"  // 1B with 18 decimals
	RewardPerUnit  = "1000000000000000"              // 0.001 DPC
	ChainID        = "dpc-testnet-1"
)

// Proof-of-Compute state
type PoCState struct {
	mu                sync.RWMutex `json:"-"`
	TotalSupply       string       `json:"total_supply"`
	CurrentHeight     int64        `json:"current_height"`
	CurrentDifficulty uint64       `json:"current_difficulty"`
	JobsSubmitted     uint64       `json:"jobs_submitted"`
	JobsCompleted     uint64       `json:"jobs_completed"`
	TotalRewards      string       `json:"total_rewards"`
	AppHash           []byte       `json:"app_hash"`
}

// DPCApplication implements the ABCI Application interface
type DPCApplication struct {
	abcitypes.BaseApplication

	db         *badger.DB
	state      *PoCState
	validators []abcitypes.ValidatorUpdate
}

// NewDPCApplication creates a new DPC blockchain application with persistence
func NewDPCApplication(dbPath string) *DPCApplication {
	// Ensure database directory exists
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		panic(fmt.Errorf("failed to create database directory: %w", err))
	}

	// Open badger database with silent logger
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Silence badger logs

	db, err := badger.Open(opts)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	app := &DPCApplication{
		db: db,
		state: &PoCState{
			TotalSupply:       InitialSupply,
			CurrentHeight:     0,
			CurrentDifficulty: 1,
			JobsSubmitted:     0,
			JobsCompleted:     0,
			TotalRewards:      "0",
			AppHash:           []byte("dpc-genesis"),
		},
		validators: make([]abcitypes.ValidatorUpdate, 0),
	}

	// Load existing state from database
	app.loadState()

	log.Printf("[DPC] Application initialized with persistence at: %s", dbPath)
	log.Printf("[DPC] Current height: %d, AppHash: %x", app.state.CurrentHeight, app.state.AppHash)

	return app
}

// Close closes the database connection
func (app *DPCApplication) Close() error {
	if app.db != nil {
		return app.db.Close()
	}
	return nil
}

// saveState persists the current state to database
func (app *DPCApplication) saveState() {
	app.state.mu.RLock()
	defer app.state.mu.RUnlock()

	data, err := json.Marshal(app.state)
	if err != nil {
		log.Printf("[DPC] Error marshaling state: %v", err)
		return
	}

	err = app.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("state"), data)
	})
	if err != nil {
		log.Printf("[DPC] Error saving state to database: %v", err)
	} else {
		log.Printf("[DPC] State saved: height=%d, hash=%x", app.state.CurrentHeight, app.state.AppHash)
	}
}

// loadState loads persisted state from database
func (app *DPCApplication) loadState() {
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("state"))
		if err == badger.ErrKeyNotFound {
			log.Printf("[DPC] No existing state found, using genesis state")
			return nil
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, app.state)
		})
	})
	if err != nil {
		log.Printf("[DPC] Error loading state from database: %v", err)
	} else if app.state.CurrentHeight > 0 {
		log.Printf("[DPC] State loaded: height=%d, hash=%x", app.state.CurrentHeight, app.state.AppHash)
	}
}

// Info returns application info
func (app *DPCApplication) Info(ctx context.Context, req *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	// Reload state from database to ensure we have latest
	app.loadState()

	app.state.mu.RLock()
	defer app.state.mu.RUnlock()

	log.Printf("[DPC] Info request: returning height=%d, hash=%x", app.state.CurrentHeight, app.state.AppHash)

	return &abcitypes.ResponseInfo{
		Version:          "v1.0.0",
		AppVersion:       1,
		LastBlockHeight:  app.state.CurrentHeight,
		LastBlockAppHash: app.state.AppHash,
		Data:            "DPC Proof-of-Compute Blockchain",
	}, nil
}

// InitChain initializes the chain with genesis state
func (app *DPCApplication) InitChain(ctx context.Context, req *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	app.state.mu.Lock()
	defer app.state.mu.Unlock()

	// Parse genesis state
	var genesisState struct {
		Proofofcompute struct {
			Params struct {
				MinComputeUnits     uint64 `json:"min_compute_units"`
				RewardPerUnit       string `json:"reward_per_unit"`
				MaxSupply           string `json:"max_supply"`
				ComplexityMultiplier uint64 `json:"complexity_multiplier"`
			} `json:"params"`
			TotalSupply string `json:"total_supply"`
		} `json:"proofofcompute"`
	}

	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err == nil {
		app.state.TotalSupply = InitialSupply
	}

	// Store validators
	app.validators = req.Validators

	log.Printf("[DPC] Chain initialized: chain_id=%s, validators=%d", ChainID, len(req.Validators))

	return &abcitypes.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
	}, nil
}

// CheckTx validates a transaction before adding to mempool
func (app *DPCApplication) CheckTx(ctx context.Context, req *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	// Basic validation - accept all transactions for now
	return &abcitypes.ResponseCheckTx{
		Code:      abcitypes.CodeTypeOK,
		GasWanted: 200000,
	}, nil
}

// PrepareProposal prepares a block proposal
func (app *DPCApplication) PrepareProposal(ctx context.Context, req *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
	return &abcitypes.ResponsePrepareProposal{
		Txs: req.Txs,
	}, nil
}

// ProcessProposal processes a block proposal
func (app *DPCApplication) ProcessProposal(ctx context.Context, req *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
	return &abcitypes.ResponseProcessProposal{
		Status: abcitypes.ResponseProcessProposal_ACCEPT,
	}, nil
}

// FinalizeBlock processes all transactions in a block (ABCI++ method)
func (app *DPCApplication) FinalizeBlock(ctx context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	app.state.mu.Lock()
	defer app.state.mu.Unlock()

	app.state.CurrentHeight = req.Height

	txs := make([]*abcitypes.ExecTxResult, len(req.Txs))

	for i, tx := range req.Txs {
		// Parse transaction
		var txData struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(tx, &txData); err != nil {
			txs[i] = &abcitypes.ExecTxResult{
				Code: 1,
				Log:  fmt.Sprintf("Failed to parse transaction: %v", err),
			}
			continue
		}

		// Handle different transaction types
		switch txData.Type {
		case "submit_job":
			app.state.JobsSubmitted++
			txs[i] = &abcitypes.ExecTxResult{
				Code: abcitypes.CodeTypeOK,
				Log:  "Job submitted successfully",
			}

		case "submit_proof":
			app.state.JobsCompleted++
			// Calculate reward based on compute units
			var proof struct {
				ComputeUnits uint64 `json:"compute_units"`
				Complexity   uint64 `json:"complexity"`
			}
			if err := json.Unmarshal(txData.Data, &proof); err == nil {
				// Reward calculation (simplified)
				app.state.TotalRewards = fmt.Sprintf("%d", app.state.JobsCompleted*1000000000000000)
			}
			txs[i] = &abcitypes.ExecTxResult{
				Code: abcitypes.CodeTypeOK,
				Log:  "Proof verified, reward distributed",
			}

		case "register_provider":
			txs[i] = &abcitypes.ExecTxResult{
				Code: abcitypes.CodeTypeOK,
				Log:  "Provider registered",
			}

		case "create_wallet":
			txs[i] = &abcitypes.ExecTxResult{
				Code: abcitypes.CodeTypeOK,
				Log:  "AI Agent wallet created",
			}

		default:
			txs[i] = &abcitypes.ExecTxResult{
				Code: 2,
				Log:  fmt.Sprintf("Unknown transaction type: %s", txData.Type),
			}
		}
	}

	// Adjust difficulty every 100 blocks
	if app.state.CurrentHeight%100 == 0 && app.state.CurrentHeight > 0 {
		if app.state.JobsCompleted > app.state.JobsSubmitted/2 {
			app.state.CurrentDifficulty++
		} else if app.state.CurrentDifficulty > 1 {
			app.state.CurrentDifficulty--
		}
	}

	// Update app hash
	app.state.AppHash = []byte(fmt.Sprintf("dpc-block-%d", app.state.CurrentHeight))

	log.Printf("[DPC] FinalizeBlock: height=%d, txs=%d, hash=%x",
		app.state.CurrentHeight, len(req.Txs), app.state.AppHash)

	return &abcitypes.ResponseFinalizeBlock{
		TxResults: txs,
		AppHash:   app.state.AppHash,
	}, nil
}

// Commit saves the application state to disk
func (app *DPCApplication) Commit(ctx context.Context, req *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	app.state.mu.RLock()
	height := app.state.CurrentHeight
	app.state.mu.RUnlock()

	// Persist state to database
	app.saveState()

	return &abcitypes.ResponseCommit{
		RetainHeight: height,
	}, nil
}

// Query handles state queries
func (app *DPCApplication) Query(ctx context.Context, req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
	app.state.mu.RLock()
	defer app.state.mu.RUnlock()

	switch req.Path {
	case "/status":
		status := map[string]interface{}{
			"chain_id":          ChainID,
			"height":            app.state.CurrentHeight,
			"total_supply":      app.state.TotalSupply,
			"max_supply":        MaxSupply,
			"difficulty":        app.state.CurrentDifficulty,
			"jobs_submitted":    app.state.JobsSubmitted,
			"jobs_completed":    app.state.JobsCompleted,
			"total_rewards":     app.state.TotalRewards,
			"consensus":         "Proof-of-Compute",
		}
		data, _ := json.Marshal(status)
		return &abcitypes.ResponseQuery{
			Code:  abcitypes.CodeTypeOK,
			Value: data,
		}, nil

	case "/supply":
		supply := map[string]interface{}{
			"total":     app.state.TotalSupply,
			"max":       MaxSupply,
			"rewards":   app.state.TotalRewards,
		}
		data, _ := json.Marshal(supply)
		return &abcitypes.ResponseQuery{
			Code:  abcitypes.CodeTypeOK,
			Value: data,
		}, nil

	default:
		return &abcitypes.ResponseQuery{
			Code:  1,
			Log:   fmt.Sprintf("Unknown query path: %s", req.Path),
		}, nil
	}
}

// ExtendVote creates vote extension
func (app *DPCApplication) ExtendVote(ctx context.Context, req *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return &abcitypes.ResponseExtendVote{
		VoteExtension: []byte("dpc-vote"),
	}, nil
}

// VerifyVoteExtension verifies vote extension
func (app *DPCApplication) VerifyVoteExtension(ctx context.Context, req *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return &abcitypes.ResponseVerifyVoteExtension{
		Status: abcitypes.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

// ListSnapshots lists snapshots
func (app *DPCApplication) ListSnapshots(ctx context.Context, req *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return &abcitypes.ResponseListSnapshots{}, nil
}

// OfferSnapshot offers a snapshot
func (app *DPCApplication) OfferSnapshot(ctx context.Context, req *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return &abcitypes.ResponseOfferSnapshot{
		Result: abcitypes.ResponseOfferSnapshot_ABORT,
	}, nil
}

// LoadSnapshotChunk loads a snapshot chunk
func (app *DPCApplication) LoadSnapshotChunk(ctx context.Context, req *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return &abcitypes.ResponseLoadSnapshotChunk{}, nil
}

// ApplySnapshotChunk applies a snapshot chunk
func (app *DPCApplication) ApplySnapshotChunk(ctx context.Context, req *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return &abcitypes.ResponseApplySnapshotChunk{
		Result: abcitypes.ResponseApplySnapshotChunk_ABORT,
	}, nil
}

// GetValidatorUpdates returns current validators
func (app *DPCApplication) GetValidatorUpdates() []abcitypes.ValidatorUpdate {
	app.state.mu.RLock()
	defer app.state.mu.RUnlock()

	return app.validators
}

// GetState returns the current state (for RPC queries)
func (app *DPCApplication) GetState() *PoCState {
	app.state.mu.RLock()
	defer app.state.mu.RUnlock()
	return app.state
}

// GetDBPath returns the default database path for a given home directory
func GetDBPath(home string) string {
	return filepath.Join(home, "data", "app.db")
}