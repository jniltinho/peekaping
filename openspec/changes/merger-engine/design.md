## Context

Peekaping currently runs four Go binaries in production: `api`, `producer`, `worker`, `ingester`, plus Redis as a mandatory queue broker. The producer polls the database for due monitors and enqueues Asynq jobs to Redis. The worker dequeues and executes health checks. The ingester consumes results from Redis and persists them. For small and medium deployments this is operationally heavy — Redis adds an external dependency with its own HA and persistence concerns.

The goal is to fold producer + worker + ingester into a single `engine` binary using Go channels as the default internal queue, making Redis optional. The API remains untouched.

Current repo state: Go 1.24, Echo v5, Bun ORM, Asynq (Redis-backed), Uber Dig DI, all MongoDB code removed.

## Goals / Non-Goals

**Goals:**
- Produce a new `cmd/engine` binary that fully replaces `cmd/producer` + `cmd/worker` + `cmd/ingester`
- Make Redis optional — default mode uses in-process Go channel queues
- Keep the `cmd/api` service completely unchanged
- No database schema migrations required
- Maintain all existing health-check executor types (HTTP, TCP, DNS, ping, gRPC, MQTT, Kafka, RabbitMQ, Redis, SQL)
- Graceful shutdown (drain in-flight checks before exit)
- Retain optional Redis mode for future horizontal scaling

**Non-Goals:**
- Embedding the engine into the API binary (single-binary mode — possible future step)
- Horizontal engine scaling (multiple engine instances) in this change
- Replacing Asynq internals if Redis mode is enabled (keep Asynq as-is for that path)
- Any frontend or API changes
- New monitor types

## Decisions

### D1 — Internal queue via Go channels
**Decision**: Default queue is a pair of buffered Go channels (`chan CheckJob`, `chan CheckResult`).
**Rationale**: Zero external dependencies, idiomatic Go, sufficient for a single-process engine. Channels are safe under concurrent producer/consumer usage. Jobs lost on restart are acceptable — the scheduler recalculates due monitors on the next tick.
**Alternative considered**: Embedded SQLite queue (like `asynq`'s in-memory mode) — adds complexity with no benefit at single-process scale.

### D2 — Scheduler polls DB directly (no Redis sorted sets)
**Decision**: The engine's scheduler goroutine polls the database with `SELECT ... WHERE next_check_at <= NOW()` on a configurable interval (default 10s), then pushes jobs to the internal channel queue.
**Rationale**: Direct DB polling is simpler and already works in the existing producer. The `SELECT ... FOR UPDATE SKIP LOCKED` pattern (PostgreSQL/MySQL) or a simple lock column (SQLite) prevents duplicate scheduling if multiple engine instances ever run.
**Alternative considered**: Keeping Redis sorted sets — reintroduces the mandatory Redis dependency.

### D3 — Worker pool size is configurable
**Decision**: `ENGINE_WORKERS` env var (default 10) controls goroutine count. Each goroutine pulls from the job channel and calls the existing executor dispatch logic.
**Rationale**: Existing executor code in `internal/modules/healthcheck/executor/` is already well-structured and can be called directly without Asynq wrapping.

### D4 — Ingester runs as a goroutine draining the result channel
**Decision**: A single ingester goroutine reads `CheckResult` values from the result channel, persists heartbeats, updates monitor status, and fires notifications — identical to what the current `cmd/ingester` does but in-process.
**Rationale**: Serialized ingestion per result is correct and avoids concurrent writes to the same monitor's status row.

### D5 — Redis mode preserved via interface abstraction
**Decision**: Define `Queue` and `ResultQueue` interfaces. Provide `MemoryQueue` (channels) and `RedisQueue` (wrapping Asynq) implementations selected by `ENGINE_USE_REDIS`.
**Rationale**: Allows future horizontal scale-out without changing the engine's core logic.

### D6 — Old binaries deprecated, not immediately deleted
**Decision**: `cmd/producer`, `cmd/worker`, `cmd/ingester` are kept in the repo but removed from all Docker Compose and Makefile build targets. A comment marks them deprecated.
**Rationale**: Provides a safe rollback path. Final removal is a follow-up commit after validation.

### D7 — Uber Dig DI wiring for engine
**Decision**: `cmd/engine/main.go` uses the same Uber Dig container pattern as `cmd/api`. `internal/engine` exposes a `RegisterDependencies` function.
**Rationale**: Consistency with existing codebase patterns. Makes it easy to share repository/service instances already constructed by Dig.

## Risks / Trade-offs

- **Jobs lost on restart** → Scheduler recalculates all due monitors on next tick (within `ENGINE_SCHEDULER_INTERVAL`). Acceptable for monitoring use case; not acceptable for guaranteed-delivery queues.
- **Channel buffer overflow under heavy load** → Configurable buffer size (`ENGINE_QUEUE_BUFFER`, default 1000). If buffer fills, scheduler blocks until workers drain it — natural back-pressure.
- **Single engine instance only in default mode** → Multiple engine instances with in-memory queues would cause duplicate checks. Mitigation: document that horizontal engine scaling requires `ENGINE_USE_REDIS=true`.
- **Asynq dependency still in go.mod** → Only loaded when `ENGINE_USE_REDIS=true`. Can be removed in a future cleanup.
- **SQLite concurrency** → SQLite under concurrent writes from ingester + API can cause `SQLITE_BUSY`. Mitigation: WAL mode already enabled in `internal/infra/sql.go`; ingester is single-goroutine so write contention is low.

## Migration Plan

1. Build `cmd/engine` alongside existing binaries (no removal yet)
2. Update `docker-compose.*.yml` prod/bundle files to replace producer+worker+ingester with engine
3. Update `Makefile` with `build-engine` target
4. Update supervisord bundle configs
5. Update startup scripts
6. Helm chart: replace 3 deployments with 1 engine deployment
7. Smoke-test: bring up with SQLite, PostgreSQL, MySQL
8. Mark old cmd directories as deprecated
9. Update docs (README, architecture pages)

**Rollback**: Revert compose files to use producer+worker+ingester images. No DB migration to reverse.

## Open Questions

- Should `ENGINE_SCHEDULER_INTERVAL` default to 10s (same as current producer) or shorter?
- Should the engine expose a `/healthz` HTTP endpoint for liveness probes? (Recommended yes, simple port like 8035)
- Should the engine emit Prometheus metrics at `/metrics`? (Deferred to follow-up)
