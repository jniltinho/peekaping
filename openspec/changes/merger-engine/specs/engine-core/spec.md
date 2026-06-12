## ADDED Requirements

### Requirement: Engine binary unifies scheduling, execution, and ingestion
The system SHALL provide a single `cmd/engine` binary that replaces `cmd/producer`, `cmd/worker`, and `cmd/ingester`. The engine SHALL start all three subsystems (scheduler, worker pool, ingester) within a single process and coordinate them via internal queues.

#### Scenario: Engine starts all subsystems
- **WHEN** the engine binary is launched with valid configuration
- **THEN** the scheduler goroutine starts polling the database for due monitors
- **THEN** the worker pool starts with `ENGINE_WORKERS` goroutines consuming from the job queue
- **THEN** the ingester goroutine starts consuming from the result queue

#### Scenario: Engine shuts down gracefully
- **WHEN** the engine receives SIGTERM or SIGINT
- **THEN** the scheduler stops enqueuing new jobs
- **THEN** the worker pool finishes all in-flight checks before exiting
- **THEN** the ingester drains and persists all pending results before exiting

### Requirement: In-memory queue as default
The system SHALL use buffered Go channels as the default job and result queues when `ENGINE_USE_REDIS=false` (the default).

#### Scenario: Jobs flow through in-memory queue
- **WHEN** the scheduler finds a due monitor and `ENGINE_USE_REDIS=false`
- **THEN** the check job is pushed to an in-process channel without any Redis connection

#### Scenario: Buffer back-pressure
- **WHEN** the job channel buffer is full (all `ENGINE_WORKERS` goroutines are busy)
- **THEN** the scheduler blocks on the channel send until a worker frees up
- **THEN** no jobs are dropped or lost

### Requirement: Redis queue as optional mode
The system SHALL support Redis-backed queues when `ENGINE_USE_REDIS=true`, using the existing Asynq infrastructure.

#### Scenario: Redis mode selected at startup
- **WHEN** the engine starts with `ENGINE_USE_REDIS=true` and valid `REDIS_URL`
- **THEN** jobs are enqueued to Redis via Asynq
- **THEN** workers consume jobs from Redis instead of in-memory channels

#### Scenario: Redis mode requires REDIS_URL
- **WHEN** the engine starts with `ENGINE_USE_REDIS=true` but no `REDIS_URL` configured
- **THEN** the engine exits with a clear error message indicating the missing configuration

### Requirement: Configurable worker pool size
The system SHALL allow the number of concurrent health-check workers to be configured via `ENGINE_WORKERS` environment variable (default: 10).

#### Scenario: Default worker count
- **WHEN** the engine starts without `ENGINE_WORKERS` set
- **THEN** exactly 10 worker goroutines are started

#### Scenario: Custom worker count
- **WHEN** the engine starts with `ENGINE_WORKERS=25`
- **THEN** exactly 25 worker goroutines are started

### Requirement: Scheduler polls for due monitors
The scheduler SHALL query the database at each tick interval for monitors whose `next_check_at` is in the past, and enqueue a check job for each.

#### Scenario: Due monitors enqueued
- **WHEN** the scheduler tick fires and monitors exist with `next_check_at <= now()`
- **THEN** each due monitor is enqueued as a `CheckJob` exactly once per tick
- **THEN** the monitor's `next_check_at` is updated to prevent re-scheduling within the same tick

#### Scenario: No due monitors
- **WHEN** the scheduler tick fires and no monitors are due
- **THEN** no jobs are enqueued and the scheduler waits for the next tick

#### Scenario: Configurable scheduler interval
- **WHEN** `ENGINE_SCHEDULER_INTERVAL` is set (e.g., `10s`)
- **THEN** the scheduler polls the database at that interval

### Requirement: Worker executes health checks using existing executors
Each worker goroutine SHALL dequeue a `CheckJob` and invoke the appropriate executor from `internal/modules/healthcheck/executor/` based on monitor type.

#### Scenario: HTTP check executed
- **WHEN** a worker dequeues a job for an HTTP monitor
- **THEN** the HTTP executor is called with the monitor's configuration
- **THEN** a `CheckResult` is produced with success flag, status code, and latency

#### Scenario: Check timeout enforced
- **WHEN** a health check exceeds `ENGINE_CHECK_TIMEOUT`
- **THEN** the check is cancelled via context and a failed `CheckResult` is produced with a timeout error

### Requirement: Ingester persists results and fires notifications
The ingester goroutine SHALL consume each `CheckResult`, persist a heartbeat record, update the monitor's current status, and dispatch notifications when status changes.

#### Scenario: Successful check persisted
- **WHEN** the ingester receives a successful `CheckResult`
- **THEN** a heartbeat record is inserted with `status=up`, latency, and timestamp
- **THEN** the monitor's `status` and `last_checked_at` fields are updated

#### Scenario: Failing check triggers notification
- **WHEN** the ingester receives a failed `CheckResult` and the monitor was previously `up`
- **THEN** a heartbeat record is inserted with `status=down`
- **THEN** all configured notification channels for that monitor are triggered

#### Scenario: Recovery triggers notification
- **WHEN** the ingester receives a successful `CheckResult` and the monitor was previously `down`
- **THEN** a heartbeat record is inserted with `status=up`
- **THEN** all configured notification channels are triggered with a recovery message

### Requirement: Engine configuration via environment variables
The engine SHALL be fully configured via environment variables with sensible defaults.

#### Scenario: All required config present
- **WHEN** `DB_DRIVER`, `DB_DSN` (or equivalent DB vars) are set
- **THEN** the engine starts without error

#### Scenario: Config defaults applied
- **WHEN** optional env vars are omitted
- **THEN** `ENGINE_WORKERS=10`, `ENGINE_SCHEDULER_INTERVAL=10s`, `ENGINE_CHECK_TIMEOUT=30s`, `ENGINE_USE_REDIS=false`, `ENGINE_QUEUE_BUFFER=1000` are used
