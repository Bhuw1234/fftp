// Package dpc_connector provides integration with Bacalhau compute nodes
package dpc_connector

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bacalhau-project/bacalhau/pkg/models"
)

// Integration provides the glue between Bacalhau and DPC rewards
type Integration struct {
	connector *Connector
	hook      *JobCompletionHook
	config    Config
}

// NewIntegration creates a new Bacalhau-DPC integration
func NewIntegration(cfg Config) *Integration {
	connector := New(cfg)
	hookCfg := DefaultJobHookConfig()
	hookCfg.Connector = connector

	return &Integration{
		connector: connector,
		hook:      NewJobCompletionHook(hookCfg),
		config:    cfg,
	}
}

// Start begins the integration
func (i *Integration) Start(ctx context.Context) {
	i.hook.Start(ctx)
	log.Ctx(ctx).Info().
		Str("rpc", i.config.RPCEndpoint).
		Str("node", i.config.NodeAddress).
		Msg("DPC integration started")
}

// Stop stops the integration
func (i *Integration) Stop() {
	i.hook.Stop()
}

// OnExecutionComplete handles execution completion events from Bacalhau
// This should be called from a watcher or event handler
func (i *Integration) OnExecutionComplete(ctx context.Context, execution models.Execution) error {
	if !i.config.Enabled {
		return nil
	}

	// Only process completed executions
	if execution.ComputeState.StateType != models.ExecutionStateCompleted {
		return nil
	}

	// Create job completion event
	event := JobCompletionEvent{
		JobID:       execution.Job.ID,
		ExecutionID: execution.ID,
		NodeID:      execution.NodeID,
		Success:     true,
		CompletedAt: time.Now(),
	}

	// Get execution time from run output
	if execution.RunOutput != nil {
		event.ExecutionTime = execution.RunOutput.GetExecutionDuration()
	}

	// Get output hash
	if execution.PublishedResult != nil {
		event.OutputHash = execution.PublishedResult.String()
	}

	// Estimate compute units based on job spec
	event.ComputeUnits = i.estimateComputeUnits(execution)

	return i.hook.OnJobComplete(ctx, event)
}

// OnExecutionFailed handles execution failure events
func (i *Integration) OnExecutionFailed(ctx context.Context, execution models.Execution) error {
	if !i.config.Enabled {
		return nil
	}

	event := JobCompletionEvent{
		JobID:       execution.Job.ID,
		ExecutionID: execution.ID,
		NodeID:      execution.NodeID,
		Success:     false,
		CompletedAt: time.Now(),
	}

	return i.hook.OnJobComplete(ctx, event)
}

// estimateComputeUnits estimates compute units from an execution
func (i *Integration) estimateComputeUnits(execution models.Execution) uint64 {
	// Base estimation from job task
	task := execution.Job.Task()
	if task == nil {
		return 1
	}

	// Calculate from resource requirements
	var units uint64 = 1

	// CPU contribution
	if task.ResourcesConfig.CPU > 0 {
		units += uint64(task.ResourcesConfig.CPU * 100)
	}

	// Memory contribution (per GB)
	if task.ResourcesConfig.Memory > 0 {
		units += uint64(task.ResourcesConfig.Memory / (1024 * 1024 * 1024))
	}

	// GPU contribution (higher weight)
	if task.ResourcesConfig.GPU > 0 {
		units += uint64(task.ResourcesConfig.GPU * 1000)
	}

	// Apply reward multiplier
	if i.config.RewardMultiplier != 0 && i.config.RewardMultiplier != 1.0 {
		units = uint64(float64(units) * i.config.RewardMultiplier)
	}

	return units
}

// GetPendingRewards returns the pending reward balance
func (i *Integration) GetPendingRewards(ctx context.Context) (string, error) {
	return i.connector.GetRewardBalance(ctx)
}

// ClaimRewards claims pending rewards
func (i *Integration) ClaimRewards(ctx context.Context) (string, error) {
	return i.connector.ClaimRewards(ctx)
}

// Stats returns integration statistics
func (i *Integration) Stats() map[string]interface{} {
	return i.hook.Stats()
}

// Example usage in compute node startup:
//
//	func setupDPCIntegration(ctx context.Context, nodeID string) *dpc_connector.Integration {
//		cfg := dpc_connector.Config{
//			Enabled:            os.Getenv("DPC_ENABLED") == "true",
//			RPCEndpoint:        getEnv("DPC_RPC_ENDPOINT", "http://localhost:26657"),
//			ChainID:            getEnv("DPC_CHAIN_ID", "dpc-testnet-1"),
//			NodeAddress:        os.Getenv("DPC_NODE_ADDRESS"),
//			MinimumJobDuration: 1,
//			RewardMultiplier:   1.0,
//		}
//
//		integration := dpc_connector.NewIntegration(cfg)
//		integration.Start(ctx)
//
//		return integration
//	}
//
// Then in the event watcher:
//
//	func (w *Watcher) handleExecutionComplete(ctx context.Context, exec models.Execution) {
//		if w.dpcIntegration != nil {
//			w.dpcIntegration.OnExecutionComplete(ctx, exec)
//		}
//	}

// RegisterWithWatcher registers the DPC integration with a Bacalhau watcher
func (i *Integration) RegisterWithWatcher(watcher EventWatcher) {
	watcher.RegisterHandler(i)
}

// EventWatcher is the interface for Bacalhau event watchers
type EventWatcher interface {
	RegisterHandler(handler ExecutionHandler)
}

// ExecutionHandler handles execution events
type ExecutionHandler interface {
	OnExecutionComplete(ctx context.Context, execution models.Execution) error
	OnExecutionFailed(ctx context.Context, execution models.Execution) error
}

// Ensure Integration implements ExecutionHandler
var _ ExecutionHandler = (*Integration)(nil)

// WatcherIntegration connects DPC to Bacalhau's event watcher system
type WatcherIntegration struct {
	integration *Integration
}

// NewWatcherIntegration creates a watcher-based integration
func NewWatcherIntegration(cfg Config) *WatcherIntegration {
	return &WatcherIntegration{
		integration: NewIntegration(cfg),
	}
}

// Start starts the watcher integration
func (w *WatcherIntegration) Start(ctx context.Context) {
	w.integration.Start(ctx)
}

// Stop stops the watcher integration
func (w *WatcherIntegration) Stop() {
	w.integration.Stop()
}

// HandleEvent implements the watcher.Handler interface
func (w *WatcherIntegration) HandleEvent(ctx context.Context, event interface{}) error {
	switch e := event.(type) {
	case models.Execution:
		if e.ComputeState.StateType == models.ExecutionStateCompleted {
			return w.integration.OnExecutionComplete(ctx, e)
		} else if e.ComputeState.StateType == models.ExecutionStateFailed {
			return w.integration.OnExecutionFailed(ctx, e)
		}
	case *models.Execution:
		if e.ComputeState.StateType == models.ExecutionStateCompleted {
			return w.integration.OnExecutionComplete(ctx, *e)
		} else if e.ComputeState.StateType == models.ExecutionStateFailed {
			return w.integration.OnExecutionFailed(ctx, *e)
		}
	}
	return nil
}

// GetIntegration returns the underlying integration
func (w *WatcherIntegration) GetIntegration() *Integration {
	return w.integration
}

// Example environment configuration:
/*
# DPC Integration Environment Variables

DPC_ENABLED=true
DPC_RPC_ENDPOINT=http://34.180.51.11:26657
DPC_CHAIN_ID=dpc-testnet-1
DPC_NODE_ADDRESS=<your-node-wallet-address>

# Optional
DPC_MIN_JOB_DURATION=1
DPC_REWARD_MULTIPLIER=1.0
*/

// Validate validates the configuration
func (c Config) Validate() error {
	if c.Enabled {
		if c.RPCEndpoint == "" {
			return fmt.Errorf("DPC RPC endpoint is required when DPC is enabled")
		}
		if c.ChainID == "" {
			return fmt.Errorf("DPC chain ID is required when DPC is enabled")
		}
		if c.NodeAddress == "" {
			return fmt.Errorf("DPC node address is required when DPC is enabled")
		}
	}
	return nil
}
