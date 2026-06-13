# Configuration

All configuration is done via environment variables. Copy `.env.prod.example` to `.env` and adjust as needed.

## Database

| Variable | Default | Description |
|---|---|---|
| `DB_TYPE` | `sqlite` | Database driver: `sqlite`, `postgres`, `postgresql`, `mysql`, `mariadb` |
| `DB_HOST` | — | Database host (required for postgres/mysql) |
| `DB_PORT` | — | Database port (required for postgres/mysql) |
| `DB_USER` | — | Database user (required for postgres/mysql) |
| `DB_PASS` | — | Database password (required for postgres/mysql) |
| `DB_NAME` | `monitoring.db` | Database name or SQLite file path |

## Server

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8383` | Port the API + embedded frontend listens on |
| `CLIENT_URL` | `http://localhost:3000` | Allowed CORS origin (dev only) |
| `MODE` | `dev` | Runtime mode: `dev`, `prod`, or `test` |
| `LOG_LEVEL` | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `TZ` | `UTC` | Server timezone (e.g. `America/New_York`) |
| `SERVICE_NAME` | `monitoring:api` | Service identifier shown in logs |

## Redis

Redis is required for multi-node deployments. Single-node / SQLite setups can omit it by not setting `REDIS_HOST`.

| Variable | Default | Description |
|---|---|---|
| `REDIS_HOST` | `redis` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password (leave empty if none) |
| `REDIS_DB` | `0` | Redis database index (0–15) |

## Queue / Engine

| Variable | Default | Description |
|---|---|---|
| `QUEUE_CONCURRENCY` | `128` | Max concurrent task workers in the engine |
| `PRODUCER_CONCURRENCY` | `10` | Concurrent goroutines that claim and enqueue monitors |

## Brute-force protection

Failed login attempts are tracked per IP/username. The account is locked after `BRUTEFORCE_MAX_ATTEMPTS` failures within `BRUTEFORCE_WINDOW`.

| Variable | Default | Description |
|---|---|---|
| `BRUTEFORCE_MAX_ATTEMPTS` | `20` | Max failed logins before lockout |
| `BRUTEFORCE_WINDOW` | `1m` | Time window for counting failures (e.g. `5m`, `1h`) |
| `BRUTEFORCE_LOCKOUT` | `1m` | Lock duration after exceeding limit |
