// Package dpc provides DPC reward integration for Bacalhau compute nodes
package dpc

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bacalhau-project/bacalhau/pkg/lib/watcher"
	"github.com/bacalhau-project/bacalhau/pkg/models"
)

// WatcherHandler handles execution events and submits DPC rewards
type WatcherHandler struct {
	integration *Integration
}

// NewWatcherHandler creates a new DPC watcher handler
func NewWatcherHandler(cfg Config) *WatcherHandler {
	return &WatcherHandler{
		integration: NewIntegration(cfg),
	}
}

// Start begins the DPC integration
func (h *WatcherHandler) Start(ctx context.Context) {
	h.integration.Start(ctx)
}

// Stop stops the DPC integration
func (h *WatcherHandler) Stop() {
	h.integration.Stop()
}

// HandleEvent implements the watcher.Handler interface
// This is the main entry point for Bacalhau execution events
func (h *WatcherHandler) HandleEvent(ctx context.Context, event watcher.Event) error {
	// Extract the execution upsert from the event
	upsert, ok := event.Object.(models.ExecutionUpsert)
	if !ok {
		return nil // Not an execution event, skip
	}

	execution := upsert.Current

	// Handle different execution states
	switch execution.ComputeState.StateType {
	case models.ExecutionStateCompleted:
		return h.handleCompleted(ctx, execution)
	case models.ExecutionStateFailed:
		return h.handleFailed(ctx, execution)
	default:
		return nil // Not a terminal state, skip
	}
}

// handleCompleted processes a completed execution and submits DPC reward
func (h *WatcherHandler) handleCompleted(ctx context.Context, execution models.Execution) error {
	log.Ctx(ctx).Debug().
		Str("job_id", execution.Job.ID).
		Str("execution_id", execution.ID).
		Msg("DPC: Processing completed execution")

	return h.integration.OnExecutionComplete(ctx, execution)
}

// handleFailed processes a failed execution (for tracking purposes)
func (h *WatcherHandler) handleFailed(ctx context.Context, execution models.Execution) error {
	log.Ctx(ctx).Debug().
		Str("job_id", execution.Job.ID).
		Str("execution_id", execution.ID).
		Msg("DPC: Processing failed execution")

	return h.integration.OnExecutionFailed(ctx, execution)
}

// GetPendingRewards returns the pending DPC reward balance
func (h *WatcherHandler) GetPendingRewards(ctx context.Context) (string, error) {
	return h.integration.GetPendingRewards(ctx)
}

// ClaimRewards claims pending DPC rewards
func (h *WatcherHandler) ClaimRewards(ctx context.Context) (string, error) {
	return h.integration.ClaimRewards(ctx)
}

// Stats returns DPC integration statistics
func (h *WatcherHandler) Stats() map[string]interface{} {
	return h.integration.Stats()
}

// GetIntegration returns the underlying integration for advanced usage
func (h *WatcherHandler) GetIntegration() *Integration {
	return h.integration
}

// Ensure WatcherHandler implements the necessary interface
var _ watcher.Handler = (*WatcherHandler)(nil)
