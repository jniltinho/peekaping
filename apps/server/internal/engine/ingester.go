package engine

import (
	"context"
	"time"

	"peekaping/internal/modules/ingester"

	"go.uber.org/zap"
)

// Ingester drains the ResultQueue and persists each result via the existing IngesterTaskHandler.
type Ingester struct {
	handler     *ingester.IngesterTaskHandler
	resultQueue ResultQueue
	logger      *zap.SugaredLogger
}

// NewIngester creates an Ingester.
func NewIngester(
	handler *ingester.IngesterTaskHandler,
	resultQueue ResultQueue,
	logger *zap.SugaredLogger,
) *Ingester {
	return &Ingester{
		handler:     handler,
		resultQueue: resultQueue,
		logger:      logger.With("component", "engine.ingester"),
	}
}

// Start reads results from the queue until ctx is done, then drains any remaining items.
func (ing *Ingester) Start(ctx context.Context) {
	ing.logger.Info("Ingester started")
	for {
		result, err := ing.resultQueue.Pop(ctx)
		if err != nil {
			// Context cancelled — drain whatever is left in the channel.
			ing.drain()
			ing.logger.Info("Ingester stopped")
			return
		}
		ing.process(context.Background(), result)
	}
}

// drain flushes any results still buffered after the context is cancelled.
// It uses a short-lived context so it doesn't block indefinitely.
func (ing *Ingester) drain() {
	drainCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		result, err := ing.resultQueue.Pop(drainCtx)
		if err != nil {
			return
		}
		ing.process(context.Background(), result)
	}
}

func (ing *Ingester) process(ctx context.Context, result CheckResult) {
	payload := &ingester.IngesterTaskPayload{
		MonitorID:          result.MonitorID,
		MonitorName:        result.MonitorName,
		MonitorType:        result.MonitorType,
		MonitorInterval:    result.MonitorInterval,
		MonitorTimeout:     result.MonitorTimeout,
		MonitorMaxRetries:  result.MonitorMaxRetries,
		MonitorRetryInt:    result.MonitorRetryInt,
		MonitorResendInt:   result.MonitorResendInt,
		MonitorConfig:      result.MonitorConfig,
		Status:             result.Status,
		Message:            result.Message,
		PingMs:             result.PingMs,
		StartTime:          result.StartTime,
		EndTime:            result.EndTime,
		IsUnderMaintenance: result.IsUnderMaintenance,
		TLSInfo:            result.TLSInfo,
		CheckCertExpiry:    result.CheckCertExpiry,
	}
	if err := ing.handler.ProcessHeartbeat(ctx, payload); err != nil {
		ing.logger.Errorw("Failed to process heartbeat", "monitor_id", result.MonitorID, "error", err)
	}
}
