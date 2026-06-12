package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"peekaping/internal"
	"peekaping/internal/config"
	internalengine "peekaping/internal/engine"
	"peekaping/internal/infra"
	"peekaping/internal/modules/certificate"
	"peekaping/internal/modules/events"
	"peekaping/internal/modules/healthcheck"
	"peekaping/internal/modules/heartbeat"
	"peekaping/internal/modules/maintenance"
	"peekaping/internal/modules/monitor"
	"peekaping/internal/modules/monitor_maintenance"
	"peekaping/internal/modules/monitor_notification"
	"peekaping/internal/modules/monitor_tag"
	"peekaping/internal/modules/monitor_tls_info"
	"peekaping/internal/modules/notification_sent_history"
	"peekaping/internal/modules/proxy"
	"peekaping/internal/modules/setting"
	"peekaping/internal/modules/stats"
	"peekaping/internal/modules/tag"
	"peekaping/internal/version"
	"context"
	"fmt"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	log.Printf("Starting Peekaping Engine v%s", version.Version)

	cfg, err := LoadAndValidate("../..")
	if err != nil {
		log.Fatalf("Failed to load and validate Engine config: %v", err)
	}

	os.Setenv("TZ", cfg.Timezone)

	container := dig.New()

	internalCfg := cfg.ToInternalConfig()
	engineCfg := cfg.ToEngineConfig()

	container.Provide(func() *config.Config { return internalCfg })
	container.Provide(internal.ProvideLogger)
	container.Provide(infra.ProvideSQLDB)

	// Redis event bus — used for publishing heartbeat events to the API's WebSocket layer.
	// When ENGINE_USE_REDIS=false this still requires Redis for events; for a fully Redis-free
	// setup, a future in-process event bus can replace this provider.
	container.Provide(infra.ProvideRedisClient)
	container.Provide(infra.ProvideRedisEventBus)

	// Module dependencies needed by the engine subsystems.
	heartbeat.RegisterDependencies(container, internalCfg)
	healthcheck.RegisterDependencies(container)
	tag.RegisterDependencies(container, internalCfg)
	monitor_tag.RegisterDependencies(container, internalCfg)
	monitor.RegisterDependencies(container, internalCfg)
	proxy.RegisterDependencies(container, internalCfg)
	maintenance.RegisterDependencies(container, internalCfg)
	monitor_maintenance.RegisterDependencies(container, internalCfg)
	monitor_notification.RegisterDependencies(container, internalCfg)
	setting.RegisterDependencies(container, internalCfg)
	notification_sent_history.RegisterDependencies(container)
	monitor_tls_info.RegisterDependencies(container)
	certificate.RegisterDependencies(container)
	stats.RegisterDependencies(container, internalCfg)

	// Engine components.
	internalengine.RegisterDependencies(container, engineCfg)

	err = container.Invoke(func(
		eng *internalengine.Engine,
		statsSvc stats.Service,
		eventBus events.EventBus,
		logger *zap.SugaredLogger,
	) error {
		// Wire stats event handlers.
		if impl, ok := statsSvc.(*stats.ServiceImpl); ok {
			impl.RegisterEventHandlers(eventBus)
		}

		logger.Infow("Engine configured",
			"workers", engineCfg.Workers,
			"scheduler_interval", engineCfg.SchedulerInterval,
			"check_timeout", engineCfg.CheckTimeout,
			"use_redis", engineCfg.UseRedis,
		)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		logger.Info("Engine is running. Press Ctrl+C to stop.")
		if err := eng.Start(ctx); err != nil {
			return fmt.Errorf("engine error: %w", err)
		}

		if err := eventBus.Close(); err != nil {
			logger.Errorw("Failed to close event bus", "error", err)
		}

		if err := infra.GracefulDatabaseShutdown(container, internalCfg, logger); err != nil {
			logger.Errorw("Failed to shutdown database", "error", err)
		}

		logger.Info("Engine stopped gracefully")
		return nil
	})

	if err != nil {
		log.Fatalf("Engine error: %v", err)
	}
}
