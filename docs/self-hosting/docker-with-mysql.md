# Docker + MySQL/MariaDB Setup

## Prerequisites

- Docker and Docker Compose installed
- A `.env` file with your database credentials (see [Configuration](../configuration.md))

## Quick start (bundle image — MySQL embedded)

Create a `docker-compose.yml`:

```yaml
services:
  peekaping:
    image: jniltinho/peekaping-bundle-mysql:latest
    restart: unless-stopped
    ports:
      - "8383:8383"
    env_file:
      - .env
    environment:
      - DB_TYPE=mysql
      - DB_HOST=localhost
      - DB_PORT=3306
    volumes:
      - ./.data/mysql:/var/lib/mysql
      - ./.data/logs:/var/log/supervisor
    container_name: peekaping-bundle-mysql
```

Minimum `.env`:

```env
DB_USER=peekaping
DB_PASS=yourpassword
DB_NAME=peekaping
DB_HOST=localhost
DB_TYPE=mysql

SERVER_PORT=8383
```

```bash
docker compose up -d
```

Open `http://localhost:8383` in your browser.

> **MariaDB**: set `DB_TYPE=mariadb` (uses the MySQL wire protocol — same driver, same options).

## Microservices setup (advanced)

For a fully separated deployment (separate `api`, `engine`, `migrate` containers):

```bash
docker compose -f docker-compose.mysql.yml up -d
```

This starts: `database` (MariaDB), `redis`, `migrate`, `api`, `engine`.

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
