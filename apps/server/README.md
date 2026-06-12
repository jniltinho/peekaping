# Peekaping Server

Go backend for Peekaping — uptime monitoring system.

## Binaries

| Binary | Command | Description |
|--------|---------|-------------|
| `api` | `cmd/api` | HTTP API server (Echo v5, port 8034). Serves REST endpoints and embeds the React frontend. |
| `engine` | `cmd/engine` | Unified scheduler + worker pool + ingester. Replaces the old `producer`, `worker`, and `ingester` binaries. |
| `bun` | `cmd/bun` | Database migration tool (uptrace/bun CLI). |

> `cmd/producer`, `cmd/worker`, `cmd/ingester` are kept for rollback only — use `engine` for new deployments.

## Tech Stack

- **Go 1.26**
- **HTTP framework**: Echo v5 (`github.com/labstack/echo/v5`)
- **ORM**: uptrace/bun (SQLite / PostgreSQL / MySQL / MariaDB)
- **DI**: Uber Dig
- **Queue**: Asynq (Redis-backed, optional in engine via `ENGINE_USE_REDIS=false`)
- **Event bus**: Redis Pub/Sub

## Build

```bash
# All binaries
make

# Individual
make api
make engine
make bun

# Embed frontend before building api (local testing)
# Run from repo root:
make embed-web   # builds web and copies dist → internal/frontend/web/
make api
```

## Frontend Embed

The `api` binary embeds the compiled React frontend via `//go:embed` (`internal/frontend/`). In Docker, the `web-builder` stage runs first and its `dist/` is copied into the Go source tree before `go build ./cmd/api`.

In development, the frontend runs via Vite dev server (`pnpm run dev`) which proxies `/api` and `/socket.io` to `localhost:8034` — no embed needed.

## Key Configuration (api)

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8034` | HTTP listen port |
| `CLIENT_URL` | `http://localhost:3000` | Allowed origin / client URL |
| `DB_TYPE` | `sqlite` | `sqlite`, `postgres`, `mysql`, `mariadb` |
| `DB_NAME` | `peekaping.db` | Database name or SQLite path |
| `REDIS_HOST` | `redis` | Redis host (event bus) |
| `MODE` | `dev` | `dev` or `prod` |

## Key Configuration (engine)

| Variable | Default | Description |
|----------|---------|-------------|
| `ENGINE_WORKERS` | `10` | Concurrent health-check workers |
| `ENGINE_SCHEDULER_INTERVAL` | `10s` | DB poll interval for due monitors |
| `ENGINE_CHECK_TIMEOUT` | `30s` | Per-check wall-clock timeout |
| `ENGINE_QUEUE_BUFFER` | `1000` | In-memory channel buffer size |
| `ENGINE_USE_REDIS` | `false` | Use Redis-backed queue (future) |

## Project Structure

```
apps/server/
├── cmd/
│   ├── api/          HTTP API server entrypoint
│   ├── engine/       Engine (scheduler + worker + ingester) entrypoint
│   └── bun/          Migration tool entrypoint
├── internal/
│   ├── engine/       Engine core: scheduler, worker pool, ingester, queues
│   ├── frontend/     Embedded React SPA (//go:embed all:web)
│   ├── modules/      Feature modules (monitor, heartbeat, auth, …)
│   ├── config/       Shared config structs and validators
│   ├── infra/        DB, Redis, event bus providers
│   └── server.go     Echo router setup and DI wiring
└── Makefile
```

## Docs

Swagger UI (dev): http://localhost:8034/swagger/index.html

Architecture docs: https://docs.peekaping.com/architecture
