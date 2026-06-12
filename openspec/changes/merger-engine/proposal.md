## Why

The current Peekaping deployment requires four separate services (api, producer, worker, ingester) plus mandatory Redis, making small and medium self-hosted deployments unnecessarily complex to operate. Unifying producer + worker + ingester into a single `engine` binary eliminates Redis as a required dependency, reduces infrastructure overhead, and makes the project accessible to operators who want a simpler setup.

## What Changes

- **New binary `cmd/engine`** replaces the three separate `cmd/producer`, `cmd/worker`, and `cmd/ingester` binaries
- **New internal package `internal/engine`** containing scheduler, worker pool, ingester, and internal queue abstractions
- **Redis becomes optional** (`ENGINE_USE_REDIS=false` uses Go channels; `true` keeps Redis as an optional external queue)
- **Docker Compose simplified** from 5 services (api, producer, worker, ingester, migrate) to 3 (api, engine, migrate)
- **Deprecated** `cmd/producer`, `cmd/worker`, `cmd/ingester` — removed from production compose files, kept in repo temporarily for rollback
- **New Dockerfiles** `Dockerfile.engine` + updated bundle compose files for SQLite, PostgreSQL, MySQL
- **Documentation updated** to reflect Medium Mode architecture

## Capabilities

### New Capabilities
- `engine-core`: Unified engine binary that combines scheduling, health-check execution, and result ingestion into a single process with configurable worker pool and optional Redis queue

### Modified Capabilities
<!-- No existing spec-level requirements are changing — this is a deployment/infrastructure consolidation. -->

## Impact

- **Removed runtime dependencies**: Redis no longer required in default (medium-mode) deploys
- **Affected binaries**: `cmd/producer`, `cmd/worker`, `cmd/ingester` deprecated; new `cmd/engine` added
- **Affected infra files**: All `docker-compose.*.yml` bundle/prod files, `Makefile`, supervisord conf files, startup scripts, Helm charts
- **Internal packages affected**: `internal/infra/queue.go` (Asynq/Redis), `internal/infra/redis_event_bus.go` — engine uses its own internal queue by default
- **No API surface changes**: The `cmd/api` REST/WebSocket service is unchanged
- **No database schema changes**: No migrations needed
- **Go module**: No new external dependencies required (Go channels for in-memory queue; existing `github.com/redis/go-redis` kept for optional Redis mode)
