// Package dpc_connector provides hooks for Bacalhau job events
package dpc_connector

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// JobHookConfig configures the job completion hook
type JobHookConfig struct {
	// Connector is the DPC connector instance
	Connector *Connector

	// BatchSize for batching reward submissions (0 = no batching)
	BatchSize int `json:"batch_size"`

	// BatchTimeout is the maximum time to wait before submitting a batch
	BatchTimeout time.Duration `json:"batch_timeout"`

	// RetryAttempts is the number of retry attempts for failed submissions
	RetryAttempts int `json:"retry_attempts"`

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration `json:"retry_delay"`
}

// DefaultJobHookConfig returns default hook configuration
func DefaultJobHookConfig() JobHookConfig {
	return JobHookConfig{
		BatchSize:     10,
		BatchTimeout:  30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
	}
}

// JobCompletionHook hooks into Bacalhau job completion events
type JobCompletionHook struct {
	config JobHookConfig

	// pendingEvents stores events waiting to be batched
	pendingEvents []JobCompletionEvent
	mu            sync.Mutex

	// stopChan signals the batch processor to stop
	stopChan chan struct{}

	// eventChan receives job completion events
	eventChan chan JobCompletionEvent
}

// NewJobCompletionHook creates a new job completion hook
func NewJobCompletionHook(cfg JobHookConfig) *JobCompletionHook {
	return &JobCompletionHook{
		config:       cfg,
		pendingEvents: make([]JobCompletionEvent, 0, cfg.BatchSize),
		stopChan:     make(chan struct{}),
		eventChan:    make(chan JobCompletionEvent, 100),
	}
}

// Start begins processing job completion events
func (h *JobCompletionHook) Start(ctx context.Context) {
	go h.processEvents(ctx)
	go h.batchProcessor(ctx)

	log.Ctx(ctx).Info().Msg("DPC job completion hook started")
}

// Stop stops the hook
func (h *JobCompletionHook) Stop() {
	close(h.stopChan)
}

// OnJobComplete is called by Bacalhau when a job completes
// This is the integration point with Bacalhau's event system
func (h *JobCompletionHook) OnJobComplete(ctx context.Context, event JobCompletionEvent) error {
	select {
	case h.eventChan <- event:
		return nil
	default:
		log.Ctx(ctx).Warn().
			Str("job_id", event.JobID).
			Msg("DPC event channel full, event dropped")
		return nil
	}
}

// processEvents processes incoming job completion events
func (h *JobCompletionHook) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case event := <-h.eventChan:
			if h.config.BatchSize > 1 {
				h.addToBatch(ctx, event)
			} else {
				h.processSingleEvent(ctx, event)
			}
		}
	}
}

// addToBatch adds an event to the pending batch
func (h *JobCompletionHook) addToBatch(ctx context.Context, event JobCompletionEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.pendingEvents = append(h.pendingEvents, event)

	// Submit batch if full
	if len(h.pendingEvents) >= h.config.BatchSize {
		h.submitBatch(ctx)
	}
}

// batchProcessor periodically submits batches
func (h *JobCompletionHook) batchProcessor(ctx context.Context) {
	if h.config.BatchSize <= 1 {
		return
	}

	ticker := time.NewTicker(h.config.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.flushBatch(ctx)
			return
		case <-h.stopChan:
			h.flushBatch(ctx)
			return
		case <-ticker.C:
			h.flushBatch(ctx)
		}
	}
}

// flushBatch submits all pending events
func (h *JobCompletionHook) flushBatch(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.pendingEvents) == 0 {
		return
	}

	h.submitBatch(ctx)
}

// submitBatch submits the current batch of events
func (h *JobCompletionHook) submitBatch(ctx context.Context) {
	events := h.pendingEvents
	h.pendingEvents = make([]JobCompletionEvent, 0, h.config.BatchSize)

	log.Ctx(ctx).Debug().Int("count", len(events)).Msg("Submitting DPC reward batch")

	for _, event := range events {
		h.processSingleEvent(ctx, event)
	}
}

// processSingleEvent processes a single job completion event with retries
func (h *JobCompletionHook) processSingleEvent(ctx context.Context, event JobCompletionEvent) {
	var lastErr error

	for attempt := 0; attempt <= h.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			log.Ctx(ctx).Debug().
				Str("job_id", event.JobID).
				Int("attempt", attempt).
				Msg("Retrying DPC reward submission")

			select {
			case <-ctx.Done():
				return
			case <-time.After(h.config.RetryDelay):
			}
		}

		err := h.config.Connector.OnJobComplete(ctx, event)
		if err == nil {
			return
		}
		lastErr = err
	}

	log.Ctx(ctx).Error().
		Err(lastErr).
		Str("job_id", event.JobID).
		Int("attempts", h.config.RetryAttempts+1).
		Msg("Failed to submit DPC reward after all attempts")
}

// Stats returns statistics about the hook
func (h *JobCompletionHook) Stats() map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	return map[string]interface{}{
		"pending_events": len(h.pendingEvents),
		"batch_size":     h.config.BatchSize,
		"channel_size":   len(h.eventChan),
	}
}
