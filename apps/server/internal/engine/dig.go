package engine

import (
	"peekaping/internal/modules/certificate"
	"peekaping/internal/modules/events"
	"peekaping/internal/modules/healthcheck"
	"peekaping/internal/modules/healthcheck/executor"
	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/ingester"
	"peekaping/internal/modules/maintenance"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/modules/monitor_maintenance"
	"peekaping/internal/modules/proxy"

	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// RegisterDependencies wires all engine components into the Uber Dig container.
func RegisterDependencies(container *dig.Container, cfg Config) {
	// Queues — created once and shared between scheduler, workers, and ingester.
	container.Provide(func(client *redis.Client) Queue {
		if cfg.UseRedis {
			return NewRedisQueue(client)
		}
		return NewMemoryQueue(cfg.QueueBuffer)
	})
	container.Provide(func(client *redis.Client) ResultQueue {
		if cfg.UseRedis {
			return NewRedisResultQueue(client)
		}
		return NewMemoryResultQueue(cfg.QueueBuffer)
	})

	// IngesterTaskHandler reused from the ingester package.
	container.Provide(func(
		heartbeatSvc heartbeat.Service,
		certSvc certificate.Service,
		monMaintenanceSvc monitor_maintenance.Service,
		bus events.EventBus,
		logger *zap.SugaredLogger,
	) *ingester.IngesterTaskHandler {
		return ingester.NewIngesterTaskHandler(heartbeatSvc, certSvc, monMaintenanceSvc, bus, logger)
	})

	container.Provide(func(
		monitorSvc monitor.Service,
		proxySvc proxy.Service,
		maintenanceSvc maintenance.Service,
		heartbeatSvc heartbeat.Service,
		q Queue,
		logger *zap.SugaredLogger,
	) *Scheduler {
		return NewScheduler(monitorSvc, proxySvc, maintenanceSvc, heartbeatSvc, q, cfg, logger)
	})

	container.Provide(func(
		supervisor *healthcheck.HealthCheckSupervisor,
		execReg *executor.ExecutorRegistry,
		q Queue,
		rq ResultQueue,
		logger *zap.SugaredLogger,
	) *WorkerPool {
		return NewWorkerPool(supervisor, execReg, q, rq, cfg, logger)
	})

	container.Provide(func(
		handler *ingester.IngesterTaskHandler,
		rq ResultQueue,
		logger *zap.SugaredLogger,
	) *Ingester {
		return NewIngester(handler, rq, logger)
	})

	container.Provide(func(
		scheduler *Scheduler,
		wp *WorkerPool,
		ing *Ingester,
		q Queue,
		rq ResultQueue,
		logger *zap.SugaredLogger,
	) *Engine {
		return New(scheduler, wp, ing, q, rq, logger)
	})
}
