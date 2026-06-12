package engine

import (
	"context"

	"go.uber.org/zap"
)

// Engine coordinates the Scheduler, WorkerPool, and Ingester subsystems.
type Engine struct {
	scheduler   *Scheduler
	workerPool  *WorkerPool
	ingester    *Ingester
	jobQueue    Queue
	resultQueue ResultQueue
	logger      *zap.SugaredLogger
}

// New assembles an Engine from its component parts.
func New(
	scheduler *Scheduler,
	workerPool *WorkerPool,
	ing *Ingester,
	jobQueue Queue,
	resultQueue ResultQueue,
	logger *zap.SugaredLogger,
) *Engine {
	return &Engine{
		scheduler:   scheduler,
		workerPool:  workerPool,
		ingester:    ing,
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		logger:      logger.With("component", "engine"),
	}
}

// Start launches all subsystems and blocks until ctx is cancelled.
func (e *Engine) Start(ctx context.Context) error {
	e.logger.Info("Engine starting")

	go e.scheduler.Start(ctx)
	go e.ingester.Start(ctx)
	e.workerPool.Start(ctx) // blocks until all workers drain

	// After workers are done, close the result queue so the ingester finishes draining.
	e.resultQueue.Close()

	e.logger.Info("Engine stopped")
	return nil
}
