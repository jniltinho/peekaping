# Docker + PostgreSQL Setup

## Prerequisites

- Docker and Docker Compose installed
- A `.env` file with your database credentials (see [Configuration](../configuration.md))

## Quick start (bundle image — PostgreSQL embedded)

Create a `docker-compose.yml`:

```yaml
services:
  monitoring:
    image: jniltinho/monitoring-bundle-postgres:latest
    restart: unless-stopped
    ports:
      - "8383:8383"
    env_file:
      - .env
    environment:
      - DB_TYPE=postgres
      - DB_HOST=localhost
      - DB_PORT=5432
    volumes:
      - ./.data/postgres:/var/lib/postgresql/data
      - ./.data/logs:/var/log/supervisor
    container_name: monitoring-bundle-postgres
```

Minimum `.env`:

```env
DB_USER=postgres
DB_PASS=yourpassword
DB_NAME=monitoring
DB_HOST=localhost
DB_TYPE=postgres

SERVER_PORT=8383
```

```bash
docker compose up -d
```

Open `http://localhost:8383` in your browser.

## Microservices setup (advanced)

For a fully separated deployment (separate `api`, `engine`, `migrate` containers):

```bash
docker compose -f docker-compose.postgres.yml up -d
```

This starts: `database` (PostgreSQL 17), `redis`, `migrate`, `api`, `engine`.

## Migrations

Schema migrations run automatically inside the bundle image before the server starts.

To run them manually:

```bash
# Via Makefile
make migrate-up

# Via binary
cd apps/server
go build -o monitoring ./cmd/monitoring
./monitoring --env-file .env migrate up
```

See [Migration Setup](../../MIGRATION_SETUP.md) for full details.

## Environment variables

See [Configuration](../configuration.md) for the full reference.
