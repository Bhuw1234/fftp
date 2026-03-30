// Package dpc provides integration with Bacalhau compute nodes
package dpc

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

	// Estimate execution time - default to 1 second if not available
	// In production, this could be tracked separately
	event.ExecutionTime = 1

	// Get output hash from published result
	if execution.PublishedResult != nil {
		event.OutputHash = execution.PublishedResult.Type
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

	// Parse and add CPU contribution
	if task.ResourcesConfig.CPU != "" {
		resources, err := task.ResourcesConfig.ToResources()
		if err == nil && resources.CPU > 0 {
			units += uint64(resources.CPU * 100)
		}
	}

	// Add memory contribution (per GB)
	if task.ResourcesConfig.Memory != "" {
		resources, err := task.ResourcesConfig.ToResources()
		if err == nil && resources.Memory > 0 {
			units += resources.Memory / (1024 * 1024 * 1024)
		}
	}

	// Add GPU contribution (higher weight)
	if task.ResourcesConfig.GPU != "" {
		resources, err := task.ResourcesConfig.ToResources()
		if err == nil && resources.GPU > 0 {
			units += resources.GPU * 1000
		}
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
