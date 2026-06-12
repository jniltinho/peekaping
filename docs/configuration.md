# Configuration

All configuration is done via environment variables. Copy `.env.prod.example` to `.env` and adjust as needed.

## Database

| Variable | Default | Description |
|---|---|---|
| `DB_TYPE` | `sqlite` | Database driver: `sqlite`, `postgres`, `mysql` |
| `DB_HOST` | — | Database host |
| `DB_PORT` | — | Database port |
| `DB_USER` | — | Database user |
| `DB_PASS` | — | Database password |
| `DB_NAME` | — | Database name (or file path for SQLite) |

## Server

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8383` | Port the API + frontend listens on |
| `CLIENT_URL` | `http://localhost:5173` | Allowed CORS origin (dev only) |
| `MODE` | `prod` | Logging mode: `dev` or `prod` |
| `TZ` | `America/New_York` | Server timezone |

## Redis

Redis is optional. Set `ENGINE_USE_REDIS=false` to run without it (SQLite / single-node setups).

| Variable | Default | Description |
|---|---|---|
| `REDIS_HOST` | `redis` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password (leave empty if none) |
| `REDIS_DB` | `0` | Redis database index |

## Engine

The `engine` binary replaces the former `producer` + `worker` + `ingester` trio.

| Variable | Default | Description |
|---|---|---|
| `ENGINE_WORKERS` | `10` | Number of parallel check workers |
| `ENGINE_SCHEDULER_INTERVAL` | `10s` | How often the scheduler enqueues due monitors |
| `ENGINE_CHECK_TIMEOUT` | `30s` | Maximum duration for a single monitor check |
| `ENGINE_QUEUE_BUFFER` | `1000` | In-memory queue buffer size |
| `ENGINE_USE_REDIS` | `false` | Use Redis-backed queue (required for multi-node) |
