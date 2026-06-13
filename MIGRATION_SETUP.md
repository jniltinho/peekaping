# Database Migrations

Peekaping uses **GORM AutoMigrate** to manage the database schema. On every startup, the migration step creates any missing tables, adds new columns, and creates indexes automatically — no manual tracking files required.

## How it works

1. The `monitoring migrate up` command connects to the configured database and calls GORM AutoMigrate across all 19 tables.
2. AutoMigrate is **idempotent** — running it against an existing, up-to-date database is a no-op.
3. Tables and columns that exist in the database but are not in the schema are left untouched (AutoMigrate never drops columns or tables).

## Running migrations

### Via Docker (bundle images)

Migrations run automatically before the server starts inside every bundle image (`peekaping-bundle-sqlite`, `peekaping-bundle-postgres`, `peekaping-bundle-mysql`). No manual step is needed.

### Via Makefile (development)

```bash
# Apply / update schema
make migrate-up

# Check database connectivity
make migrate-status
```

### Via binary

```bash
cd apps/server

# Build the unified binary
go build -o monitoring ./cmd/monitoring

# Apply / update schema
./monitoring migrate up

# Check database connectivity
./monitoring migrate status
```

Pass `--env-file /path/to/.env` to load configuration from a specific file:

```bash
./monitoring --env-file /app/.env migrate up
```

## Available subcommands

| Command | Description |
|---|---|
| `monitoring migrate up` | Create missing tables, add missing columns, create indexes |
| `monitoring migrate status` | Verify database connectivity and report DB type |

## Supported databases

| DB_TYPE | Notes |
|---|---|
| `sqlite` | Default. Single file, no server required. |
| `postgres` / `postgresql` | Recommended for production. |
| `mysql` / `mariadb` | MySQL 8+ or MariaDB 10.5+. |

## Migration container (microservices deployment)

The `apps/server/infra/Dockerfile.migrate` builds a minimal image that runs `monitoring migrate up` and exits. It is used in the microservices Docker Compose files:

```yaml
migrate:
  build:
    context: apps/server
    dockerfile: infra/Dockerfile.migrate
  env_file: .env
  restart: "no"
  depends_on:
    database:
      condition: service_healthy
```

The server service should declare `depends_on: migrate: condition: service_completed_successfully` so it waits for the migration to finish before accepting traffic.

## Makefile targets reference

| Target | Description |
|---|---|
| `make migrate-up` | Run `monitoring migrate up` via `go run` |
| `make migrate-status` | Run `monitoring migrate status` via `go run` |
| `make build` | Compile `monitoring` binary to `bin/monitoring` |
