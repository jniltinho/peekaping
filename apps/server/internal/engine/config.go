package engine

import "time"

// Config holds engine-specific runtime configuration.
type Config struct {
	// Number of concurrent worker goroutines (ENGINE_WORKERS, default 10).
	Workers int

	// How often the scheduler polls the DB for due monitors (ENGINE_SCHEDULER_INTERVAL, default 10s).
	SchedulerInterval time.Duration

	// Maximum per-check wall-clock timeout as a safety net (ENGINE_CHECK_TIMEOUT, default 30s).
	CheckTimeout time.Duration

	// Buffered channel capacity for job and result queues (ENGINE_QUEUE_BUFFER, default 1000).
	QueueBuffer int

	// When true, use Redis/Asynq queues instead of in-memory channels (ENGINE_USE_REDIS, default false).
	UseRedis bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Workers:           10,
		SchedulerInterval: 10 * time.Second,
		CheckTimeout:      30 * time.Second,
		QueueBuffer:       1000,
		UseRedis:          false,
	}
}
