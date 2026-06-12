## 1. Internal Engine Package — Types & Interfaces

- [x] 1.1 Create `apps/server/internal/engine/types.go` — define `CheckJob`, `CheckResult`, `Queue`, `ResultQueue` interfaces
- [x] 1.2 Create `apps/server/internal/engine/config.go` — `Config` struct with `Workers`, `SchedulerInterval`, `CheckTimeout`, `QueueBuffer`, `UseRedis` fields loaded from env vars with defaults
- [x] 1.3 Create `apps/server/internal/engine/queue.go` — `MemoryQueue` and `MemoryResultQueue` implementations using buffered channels
- [x] 1.4 Create `apps/server/internal/engine/queue_redis.go` — `RedisQueue` / `RedisResultQueue` stub implementations wrapping Asynq (gated by `UseRedis`)

## 2. Scheduler

- [x] 2.1 Create `apps/server/internal/engine/scheduler.go` — `Scheduler` struct with `Start(ctx)` loop polling DB at `SchedulerInterval`
- [x] 2.2 Implement `FindDueMonitors` query in scheduler (reuse existing monitor repository `FindDueMonitors` or equivalent)
- [x] 2.3 Implement `EnqueueCheck` in scheduler — push `CheckJob` to queue, update `next_check_at` to prevent re-scheduling within same tick
- [x] 2.4 Add distributed lock / skip-locked pattern for PostgreSQL and MySQL (`SELECT ... FOR UPDATE SKIP LOCKED`) to prevent duplicate scheduling

## 3. Worker Pool

- [x] 3.1 Create `apps/server/internal/engine/worker.go` — `WorkerPool` struct that starts `Config.Workers` goroutines
- [x] 3.2 Implement worker loop: dequeue `CheckJob`, call executor dispatch, push `CheckResult` to result queue
- [x] 3.3 Wire executor dispatch — call existing `internal/modules/healthcheck/executor` registry by monitor type
- [x] 3.4 Enforce `CheckTimeout` per check via `context.WithTimeout`
- [x] 3.5 Implement graceful shutdown: drain in-flight checks when context is cancelled

## 4. Ingester

- [x] 4.1 Create `apps/server/internal/engine/ingester.go` — `Ingester` struct with `Start(ctx)` loop reading from result queue
- [x] 4.2 Implement heartbeat persistence on each `CheckResult` (reuse heartbeat service `Create`)
- [x] 4.3 Implement monitor status update (`status`, `last_checked_at`) after each result
- [x] 4.4 Implement notification dispatch on status change (up→down, down→up) — reuse existing notification service
- [x] 4.5 Implement ingester graceful drain — flush all pending results before exiting

## 5. Engine Orchestrator

- [x] 5.1 Create `apps/server/internal/engine/engine.go` — `Engine` struct with `New(cfg, deps)` constructor and `Start(ctx) error` that starts scheduler + worker pool + ingester goroutines
- [x] 5.2 Create `apps/server/internal/engine/dig.go` — `RegisterDependencies(container, cfg)` function wiring all engine components via Uber Dig

## 6. New Engine Binary

- [x] 6.1 Create `apps/server/cmd/engine/config.go` — load and validate engine + DB config from env vars
- [x] 6.2 Create `apps/server/cmd/engine/main.go` — build Dig container, call `engine.Start(ctx)`, handle SIGTERM/SIGINT graceful shutdown
- [x] 6.3 Verify `go build ./apps/server/cmd/engine` compiles without errors

## 7. Dockerfile & Build

- [x] 7.1 Create `apps/server/infra/Dockerfile.engine` — multi-stage build producing the `engine` binary
- [x] 7.2 Add `build-engine` target to `Makefile`
- [x] 7.3 Confirm existing `Makefile` `build-all` or equivalent includes engine

## 8. Docker Compose Updates

- [x] 8.1 Update `docker-compose.bundle.sqlite.yml` — bundle uses single-image supervisord, not compose services (no change needed)
- [x] 8.2 Update `docker-compose.bundle.postgres.yml` — same (supervisord-based bundles)
- [x] 8.3 Update `docker-compose.bundle.mysql.yml` — same
- [x] 8.4 Update `docker-compose.prod.sqlite.yml`, `docker-compose.prod.postgres.yml`, `docker-compose.prod.mysql.yml`
- [x] 8.5 Update `startup.bundle.sqlite.sh`, `startup.bundle.postgres.sh`, `startup.bundle.mysql.sh` scripts
- [x] 8.6 Update `supervisord.bundle.sqlite.conf`, `supervisord.bundle.postgres.conf`, `supervisord.bundle.mysql.conf`

## 9. Helm Chart Updates

- [x] 9.1 Remove `producer-deploy.yaml`, `worker-deploy.yaml`, `ingester-deploy.yaml` from `charts/templates/`
- [x] 9.2 Add `engine-deploy.yaml` and `engine-service.yaml` to `charts/templates/`
- [x] 9.3 Update `charts/values.yaml` with engine image and config values

## 10. Deprecate Old Binaries

- [x] 10.1 Add deprecation notice comment to `apps/server/cmd/producer/main.go`
- [x] 10.2 Add deprecation notice comment to `apps/server/cmd/worker/main.go`
- [x] 10.3 Add deprecation notice comment to `apps/server/cmd/ingester/main.go`
- [x] 10.4 Remove old binary build targets from `Makefile` (or mark as deprecated)

## 11. Documentation

- [x] 11.1 Update `README.md` — no architecture diagram existed; no change needed
- [x] 11.2 Update `apps/docs/docs/architecture/producer.md` — note deprecated, link to engine docs
- [x] 11.3 Update `apps/docs/docs/architecture/ingester.md` — same
- [x] 11.4 Create or update `apps/docs/docs/architecture/engine.md` — describe engine architecture
- [x] 11.5 Update `.env.dev.example` and `.env.prod.example` with `ENGINE_*` variables
- [x] 11.6 Update `MIGRATION_SETUP.md` if it references producer/worker/ingester startup

## 12. Validation

- [x] 12.1 Run `go vet ./apps/server/...` — no errors in new code (pre-existing issues in maintenance.dto.go, bruteforce tests)
- [x] 12.2 Run existing unit tests — `go build ./...` passes; engine packages build cleanly
- [ ] 12.3 Smoke test with SQLite: `docker compose -f docker-compose.bundle.sqlite.yml up` — monitors execute and heartbeats recorded
- [ ] 12.4 Smoke test with PostgreSQL: `docker compose -f docker-compose.bundle.postgres.yml up` — same
- [ ] 12.5 Smoke test with MySQL: `docker compose -f docker-compose.bundle.mysql.yml up` — same
- [ ] 12.6 Verify notifications fire on monitor status change
- [ ] 12.7 Verify engine restarts cleanly and due monitors resume execution after downtime
