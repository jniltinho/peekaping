package engine

import (
	"context"
	"sync"
	"time"

	"peekaping/internal/modules/healthcheck"
	"peekaping/internal/modules/healthcheck/executor"

	"go.uber.org/zap"
)

// WorkerPool manages N goroutines that each pop CheckJobs, execute health checks,
// and push CheckResults to the result queue.
type WorkerPool struct {
	supervisor  *healthcheck.HealthCheckSupervisor
	execReg     *executor.ExecutorRegistry
	jobQueue    Queue
	resultQueue ResultQueue
	cfg         Config
	logger      *zap.SugaredLogger
	wg          sync.WaitGroup
}

// NewWorkerPool creates a WorkerPool.
func NewWorkerPool(
	supervisor *healthcheck.HealthCheckSupervisor,
	execReg *executor.ExecutorRegistry,
	jobQueue Queue,
	resultQueue ResultQueue,
	cfg Config,
	logger *zap.SugaredLogger,
) *WorkerPool {
	return &WorkerPool{
		supervisor:  supervisor,
		execReg:     execReg,
		jobQueue:    jobQueue,
		resultQueue: resultQueue,
		cfg:         cfg,
		logger:      logger.With("component", "engine.worker"),
	}
}

// Start launches cfg.Workers goroutines and blocks until ctx is cancelled and all workers exit.
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.logger.Infow("Worker pool started", "workers", wp.cfg.Workers)
	for i := 0; i < wp.cfg.Workers; i++ {
		wp.wg.Add(1)
		go wp.run(ctx, i)
	}
	wp.wg.Wait()
	wp.logger.Info("Worker pool stopped")
}

func (wp *WorkerPool) run(ctx context.Context, id int) {
	defer wp.wg.Done()
	wp.logger.Debugw("Worker started", "worker_id", id)
	for {
		job, err := wp.jobQueue.Pop(ctx)
		if err != nil {
			// Context cancelled or queue closed — exit cleanly.
			wp.logger.Debugw("Worker exiting", "worker_id", id, "reason", err)
			return
		}

		wp.process(ctx, job)
	}
}

func (wp *WorkerPool) process(ctx context.Context, job CheckJob) {
	mon := job.Monitor

	// Stale-task guard: skip if the job is older than 1.5× the monitor interval.
	staleThreshold := time.Duration(mon.Interval)*time.Second + time.Duration(mon.Interval/2)*time.Second
	if time.Since(job.ScheduledAt) > staleThreshold {
		wp.logger.Warnw("Skipping stale job",
			"monitor_id", mon.ID, "scheduled_at", job.ScheduledAt, "interval", mon.Interval)
		return
	}

	exec, ok := wp.execReg.GetExecutor(mon.Type)
	if !ok {
		wp.logger.Errorw("No executor found for monitor type",
			"monitor_id", mon.ID, "type", mon.Type)
		return
	}

	// Apply outer timeout as safety net.
	checkCtx, cancel := context.WithTimeout(ctx, wp.cfg.CheckTimeout)
	defer cancel()

	mon.LastHeartbeat = job.LastHeartbeat
	tickResult := wp.supervisor.HandleMonitorTick(checkCtx, mon, exec, job.Proxy, job.IsUnderMaintenance)
	if tickResult == nil {
		wp.logger.Debugw("Executor returned nil (no heartbeat needed)", "monitor_id", mon.ID)
		return
	}

	result := CheckResult{
		MonitorID:          mon.ID,
		MonitorName:        mon.Name,
		MonitorType:        mon.Type,
		MonitorInterval:    mon.Interval,
		MonitorTimeout:     mon.Timeout,
		MonitorMaxRetries:  mon.MaxRetries,
		MonitorRetryInt:    mon.RetryInterval,
		MonitorResendInt:   mon.ResendInterval,
		MonitorConfig:      mon.Config,
		Status:             tickResult.ExecutionResult.Status,
		Message:            tickResult.ExecutionResult.Message,
		PingMs:             tickResult.PingMs,
		StartTime:          tickResult.ExecutionResult.StartTime,
		EndTime:            tickResult.ExecutionResult.EndTime,
		IsUnderMaintenance: tickResult.IsUnderMaintenance,
		TLSInfo:            tickResult.ExecutionResult.TLSInfo,
		CheckCertExpiry:    job.CheckCertExpiry,
	}

	// Use a background context for pushing results so shutdown doesn't drop them.
	pushCtx, pushCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pushCancel()
	if err := wp.resultQueue.Push(pushCtx, result); err != nil {
		wp.logger.Errorw("Failed to push check result", "monitor_id", mon.ID, "error", err)
	}
}
