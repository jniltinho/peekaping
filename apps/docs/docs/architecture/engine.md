---
sidebar_position: 6
---

# Engine

The Engine is the unified processing component of Peekaping that combines scheduling, health-check execution, and result ingestion into a single binary. It replaces the separate `producer`, `worker`, and `ingester` services.

## Role & Responsibilities

- **Scheduler**: Polls the database at a configurable interval and enqueues check jobs for monitors that are due
- **Worker Pool**: Executes health checks concurrently using the existing executor registry
- **Ingester**: Persists heartbeat records, detects status changes, and fires notifications

## Architecture

```
DB poll (every ENGINE_SCHEDULER_INTERVAL)
         |
     Scheduler
         |
    [job channel]
         |
    Worker Pool (ENGINE_WORKERS goroutines)
         |
    [result channel]
         |
      Ingester
         |
       Database + Event Bus
```

### Internal Queue (Default)

By default (`ENGINE_USE_REDIS=false`), the engine uses buffered Go channels as the job and result queues. This requires no external dependencies beyond the database.

If the channel buffer fills up (all workers busy), the scheduler applies natural back-pressure by blocking on the push until a worker frees up. No jobs are dropped.

### Redis Queue (Optional)

Set `ENGINE_USE_REDIS=true` to route jobs through a Redis-backed queue (future feature). This enables horizontal engine scaling at the cost of requiring Redis.

## Environment Variables

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_TYPE` | `sqlite` | Database type: `postgres`, `mysql`, `mariadb`, `sqlite` |
| `DB_HOST` | — | Database host (not required for SQLite) |
| `DB_PORT` | — | Database port (not required for SQLite) |
| `DB_NAME` | `peekaping.db` | Database name or SQLite file path |
| `DB_USER` | — | Database username (not required for SQLite) |
| `DB_PASS` | — | Database password (not required for SQLite) |

### Redis Configuration (used for event bus; optional for queues)

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `redis` | Redis server hostname |
| `REDIS_PORT` | `6379` | Redis server port |
| `REDIS_PASSWORD` | `""` | Redis password |
| `REDIS_DB` | `0` | Redis database number (0–15) |

### Engine Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ENGINE_WORKERS` | `10` | Number of concurrent health-check workers |
| `ENGINE_SCHEDULER_INTERVAL` | `10s` | How often the scheduler polls for due monitors |
| `ENGINE_CHECK_TIMEOUT` | `30s` | Maximum per-check wall-clock timeout |
| `ENGINE_QUEUE_BUFFER` | `1000` | Buffered channel capacity for job/result queues |
| `ENGINE_USE_REDIS` | `false` | Use Redis-backed queues instead of in-memory channels |

### General Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MODE` | `dev` | Runtime mode: `dev`, `prod`, or `test` |
| `LOG_LEVEL` | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `TZ` | `UTC` | Timezone |
| `SERVICE_NAME` | `peekaping:engine` | Service identifier for logging |

## Graceful Shutdown

On receiving `SIGTERM` or `SIGINT`:

1. The scheduler stops enqueueing new jobs
2. Worker goroutines finish their in-flight checks
3. The ingester drains any remaining buffered results (up to 10 s)
4. The event bus and database connections are closed

## Scaling

**Single-instance (default):** One engine process handles all monitors. Suitable for up to thousands of monitors with the default worker count.

**Vertical scaling:** Increase `ENGINE_WORKERS` and `ENGINE_QUEUE_BUFFER` to handle higher monitor volumes.

**Horizontal scaling:** Requires `ENGINE_USE_REDIS=true`. Multiple engine instances compete for Redis queue items, preventing duplicate checks.

## Related Components

- [API Server](./api-server.md) — manages monitor configurations consumed by the engine
