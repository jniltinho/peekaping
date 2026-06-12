# Docker + PostgreSQL Setup

## Prerequisites

- Docker and Docker Compose installed
- A `.env` file (copy from `.env.prod.example` and adjust values)

## Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  peekaping:
    image: 0xfurai/peekaping-bundle-postgres:latest
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
    container_name: peekaping-bundle-postgres
```

```bash
docker compose up -d
```

Open `http://localhost:8383` in your browser.

## Microservices setup (advanced)

For a fully separated deployment (separate api, engine, migrate containers) use the compose file from the `docker/` folder:

```bash
docker compose -f docker/docker-compose.prod.postgres.yml up -d
```

This starts: `database` (Postgres 17), `redis`, `migrate`, `api`, `engine`.

## Environment variables

Minimum required in `.env`:

```env
DB_USER=postgres
DB_PASS=yourpassword
DB_NAME=peekaping
DB_HOST=database
DB_TYPE=postgres

SERVER_PORT=8383
```

See [Configuration](../configuration.md) for the full reference.
