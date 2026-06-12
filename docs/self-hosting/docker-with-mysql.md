# Docker + MySQL/MariaDB Setup

## Prerequisites

- Docker and Docker Compose installed
- A `.env` file (copy from `.env.prod.example` and adjust values)

## Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  peekaping:
    image: 0xfurai/peekaping-bundle-mysql:latest
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

```bash
docker compose up -d
```

Open `http://localhost:8383` in your browser.

## Microservices setup (advanced)

```bash
docker compose -f docker/docker-compose.prod.mysql.yml up -d
```

This starts: `database` (MariaDB 11), `redis`, `migrate`, `api`, `engine`.

## Environment variables

Minimum required in `.env`:

```env
DB_USER=peekaping
DB_PASS=yourpassword
DB_NAME=peekaping
DB_HOST=database
DB_TYPE=mysql

SERVER_PORT=8383
```

See [Configuration](../configuration.md) for the full reference.
