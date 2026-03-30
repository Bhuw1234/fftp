// Package dpc_connector connects Bacalhau job completion to DPC blockchain rewards
package dpc_connector

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Config holds the DPC connector configuration
type Config struct {
	// Enabled determines if DPC rewards are enabled
	Enabled bool `json:"enabled"`

	// RPCEndpoint is the DPC blockchain RPC endpoint
	RPCEndpoint string `json:"rpc_endpoint"`

	// ChainID is the DPC chain ID
	ChainID string `json:"chain_id"`

	// NodeAddress is the compute node's wallet address
	NodeAddress string `json:"node_address"`

	// MinimumJobDuration is the minimum job duration in seconds to earn rewards
	MinimumJobDuration int64 `json:"minimum_job_duration"`

	// RewardMultiplier allows adjusting rewards (default 1.0)
	RewardMultiplier float64 `json:"reward_multiplier"`
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Enabled:            false,
		RPCEndpoint:        "http://localhost:26657",
		ChainID:            "dpc-testnet-1",
		NodeAddress:        "",
		MinimumJobDuration: 1,
		RewardMultiplier:   1.0,
	}
}

// Connector handles communication between Bacalhau and DPC blockchain
type Connector struct {
	config     Config
	httpClient *http.Client
}

// New creates a new DPC connector
func New(cfg Config) *Connector {
	return &Connector{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// JobCompletionEvent represents a job completion event from Bacalhau
type JobCompletionEvent struct {
	JobID         string    `json:"job_id"`
	ExecutionID   string    `json:"execution_id"`
	NodeID        string    `json:"node_id"`
	ComputeUnits  uint64    `json:"compute_units"`
	ExecutionTime int64     `json:"execution_time"` // in seconds
	OutputHash    string    `json:"output_hash"`
	Success       bool      `json:"success"`
	CompletedAt   time.Time `json:"completed_at"`
}

// SubmitProofRequest is the request to submit a proof to DPC
type SubmitProofRequest struct {
	JobID         string `json:"job_id"`
	NodeAddress   string `json:"node_address"`
	ComputeUnits  uint64 `json:"compute_units"`
	ExecutionTime int64  `json:"execution_time"`
	OutputHash    string `json:"output_hash"`
	Signature     string `json:"signature"`
}

// SubmitProofResponse is the response from DPC
type SubmitProofResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Reward string `json:"reward"`
	Error  string `json:"error,omitempty"`
}

// OnJobComplete is called when a Bacalhau job completes
// This is the main integration point
func (c *Connector) OnJobComplete(ctx context.Context, event JobCompletionEvent) error {
	if !c.config.Enabled {
		log.Ctx(ctx).Debug().Msg("DPC rewards disabled, skipping")
		return nil
	}

	// Skip failed jobs
	if !event.Success {
		log.Ctx(ctx).Debug().Str("job_id", event.JobID).Msg("Job failed, no DPC reward")
		return nil
	}

	// Skip short jobs
	if event.ExecutionTime < c.config.MinimumJobDuration {
		log.Ctx(ctx).Debug().
			Str("job_id", event.JobID).
			Int64("duration", event.ExecutionTime).
			Msg("Job too short for DPC reward")
		return nil
	}

	// Calculate compute units if not provided
	if event.ComputeUnits == 0 {
		event.ComputeUnits = c.estimateComputeUnits(event)
	}

	// Generate signature for the proof
	signature := c.generateSignature(event)

	// Submit proof to DPC blockchain
	req := SubmitProofRequest{
		JobID:         event.JobID,
		NodeAddress:   c.config.NodeAddress,
		ComputeUnits:  event.ComputeUnits,
		ExecutionTime: event.ExecutionTime,
		OutputHash:    event.OutputHash,
		Signature:     signature,
	}

	resp, err := c.SubmitProof(ctx, req)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("job_id", event.JobID).Msg("Failed to submit DPC proof")
		return fmt.Errorf("failed to submit DPC proof: %w", err)
	}

	if resp.Error != "" {
		log.Ctx(ctx).Error().Str("error", resp.Error).Str("job_id", event.JobID).Msg("DPC proof rejected")
		return fmt.Errorf("DPC proof rejected: %s", resp.Error)
	}

	log.Ctx(ctx).Info().
		Str("job_id", event.JobID).
		Str("reward", resp.Reward).
		Str("status", resp.Status).
		Msg("DPC reward earned")

	return nil
}

// SubmitProof submits a compute proof to the DPC blockchain
func (c *Connector) SubmitProof(ctx context.Context, req SubmitProofRequest) (*SubmitProofResponse, error) {
	// Prepare the transaction
	txData, err := json.Marshal(map[string]interface{}{
		"type": "submit_proof",
		"data": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proof: %w", err)
	}

	// Broadcast transaction to DPC
	url := fmt.Sprintf("%s/tx", c.config.RPCEndpoint)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(txData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result SubmitProofResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetRewardBalance gets the pending reward balance for the node
func (c *Connector) GetRewardBalance(ctx context.Context) (string, error) {
	if !c.config.Enabled {
		return "0", nil
	}

	url := fmt.Sprintf("%s/query/reward/%s", c.config.RPCEndpoint, c.config.NodeAddress)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to query reward: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Amount string `json:"amount"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Amount, nil
}

// ClaimRewards claims pending rewards
func (c *Connector) ClaimRewards(ctx context.Context) (string, error) {
	if !c.config.Enabled {
		return "0", nil
	}

	txData, err := json.Marshal(map[string]interface{}{
		"type": "claim_reward",
		"data": map[string]string{
			"node_address": c.config.NodeAddress,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal claim: %w", err)
	}

	url := fmt.Sprintf("%s/tx", c.config.RPCEndpoint)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(txData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Amount string `json:"amount"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	log.Ctx(ctx).Info().Str("amount", result.Amount).Msg("DPC rewards claimed")

	return result.Amount, nil
}

// estimateComputeUnits estimates compute units based on job execution time
func (c *Connector) estimateComputeUnits(event JobCompletionEvent) uint64 {
	// Simple estimation: 1 compute unit per second of execution
	// This can be made more sophisticated based on CPU/GPU usage
	baseUnits := uint64(event.ExecutionTime)

	// Apply reward multiplier
	if c.config.RewardMultiplier != 0 && c.config.RewardMultiplier != 1.0 {
		baseUnits = uint64(float64(baseUnits) * c.config.RewardMultiplier)
	}

	return baseUnits
}

// generateSignature generates a signature for the proof
func (c *Connector) generateSignature(event JobCompletionEvent) string {
	// Create a deterministic hash of the job completion event
	data := fmt.Sprintf("%s:%s:%s:%d:%d",
		event.JobID,
		c.config.NodeAddress,
		event.OutputHash,
		event.ComputeUnits,
		event.ExecutionTime,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
