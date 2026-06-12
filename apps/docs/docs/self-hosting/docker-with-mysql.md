---
sidebar_position: 4
---

# Docker + MySQL / MariaDB

MariaDB is a community-developed, commercially supported fork of MySQL and is **wire-compatible** with the MySQL protocol. Peekaping uses the standard MySQL Go driver and dialect, so `DB_TYPE=mysql` **or** `DB_TYPE=mariadb` both work against a MariaDB or MySQL server.

We recommend the `mariadb` image for the open-source friendly license and excellent compatibility.

## Monolithic mode

The simplest mode of operation is the monolithic deployment mode. This mode runs all of Peekaping microservice components (db + api + web + gateway) inside a single process as a single Docker image.

```bash
docker run -d --restart=always \
  -p 8383:8383 \
  -e DB_NAME=peekaping \
  -e DB_USER=peekaping \
  -e DB_PASS=secure_test_password_123 \
  -v $(pwd)/.data/mysql:/var/lib/mysql \
  --name peekaping \
  0xfurai/peekaping-bundle-mysql:latest
```

To add custom caddy file add
```
-v ./custom-Caddyfile:/etc/caddy/Caddyfile:ro
```

Use `DB_TYPE=mariadb` (or `mysql`) if you want to be explicit in your `.env` or environment.

If you need more granular control on system components read [Microservice mode section](#microservice-mode)

## Microservice mode

### Prerequisites

- Docker Compose 2.0+

### 1. Create Project Structure

Create a new directory for your Peekaping installation and set up the following structure:

```
peekaping/
├── .env
├── docker-compose.yml
└── nginx.conf
```

### 2. Create Configuration Files

#### `.env` file

Create a `.env` file with your configuration:

```env
# Database Configuration
DB_USER=peekaping
DB_PASS=your-secure-password-here
DB_NAME=peekaping
DB_HOST=database
DB_PORT=3306
DB_TYPE=mariadb   # or mysql — both are supported

# Server Configuration
SERVER_PORT=8034
CLIENT_URL="http://localhost:8383"

# Application Settings
MODE=prod
TZ="America/New_York"

# JWT settings are automatically managed in the database
# Default settings are initialized on first startup:
# - Access token expiration: 15 minutes
# - Refresh token expiration: 720 hours (30 days)
# - Secret keys are automatically generated securely
```
:::info JWT Settings
JWT settings (access/refresh token expiration times and secret keys) are now automatically managed in the database. Default secure settings are initialized on first startup, and secret keys are generated automatically.
:::
:::warning Important Security Notes
- **Change all default passwords and secret keys**
- Use strong, unique passwords for the database
- Consider using environment-specific secrets management
:::

#### `docker-compose.yml` file

Create a `docker-compose.yml` file:

```yaml
networks:
  appnet:

services:
  database:
    image: mariadb:11
    restart: unless-stopped
    env_file:
      - .env
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_PASS:-password}
      MYSQL_DATABASE: ${DB_NAME:-peekaping}
      MYSQL_USER: ${DB_USER:-peekaping}
      MYSQL_PASSWORD: ${DB_PASS:-password}
    volumes:
      - ./.data/mysql:/var/lib/mysql
    networks:
      - appnet
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 30s
      timeout: 2s
      retries: 5
      start_period: 5s

  redis:
    image: redis:7
    restart: unless-stopped
    networks:
      - appnet
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 2s
      retries: 5
      start_period: 5s

  migrate:
    image: 0xfurai/peekaping-migrate:latest
    restart: "no"
    env_file:
      - .env
    depends_on:
      database:
        condition: service_healthy
    networks:
      - appnet

  api:
    image: 0xfurai/peekaping-api:latest
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      database:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - appnet
    healthcheck:
      test:
        [
          "CMD-SHELL",
          "wget -qO - http://localhost:8034/api/v1/health || exit 1",
        ]
      interval: 30s
      timeout: 2s
      retries: 5
      start_period: 5s

  producer:
    image: 0xfurai/peekaping-producer:latest
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      database:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - appnet

  worker:
    image: 0xfurai/peekaping-worker:latest
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      redis:
        condition: service_healthy
    networks:
      - appnet

  ingester:
    image: 0xfurai/peekaping-ingester:latest
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      database:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - appnet

  web:
    image: 0xfurai/peekaping-web:latest
    depends_on:
      api:
        condition: service_healthy
    networks:
      - appnet
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:80 || exit 1"]
      interval: 30s
      timeout: 2s
      retries: 5
      start_period: 5s

  gateway:
    image: nginx:latest
    ports:
      - "8383:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      api:
        condition: service_healthy
      web:
        condition: service_healthy
    networks:
      - appnet
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:80 || exit 1"]
      interval: 30s
      timeout: 2s
      retries: 5
      start_period: 5s
```

To add custom caddy file (when using the bundle image) or adjust ports, edit the relevant service.

### 3. Start the stack

```bash
docker compose up -d --build
```

The `migrate` container will run once and exit after applying migrations. The API and other services will become healthy after the database is ready.

### 4. Access the UI

Open http://localhost:8383 (or the port you mapped for the gateway/web).

Default admin credentials are created on first run (check logs or the initial setup flow).

## Environment variables for MySQL/MariaDB

| Variable   | Description                              | Example / Default |
|------------|------------------------------------------|-------------------|
| DB_TYPE    | `mysql` or `mariadb` (both accepted)     | mariadb           |
| DB_HOST    | Database hostname (use `database` in compose) | database       |
| DB_PORT    | Database port                            | 3306              |
| DB_NAME    | Database name                            | peekaping         |
| DB_USER    | Application database user                | peekaping         |
| DB_PASS    | Application database password            | (strong secret)   |

The official `mariadb` and `mysql` images also understand `MYSQL_*` variables for initial setup (root password, database, user). The compose examples above map from your `DB_*` variables.

## Healthchecks & readiness

- Database: `mysqladmin ping`
- API: internal `/api/v1/health`
- Web: HTTP check on the gateway port

These are already included in the example compose files.

## Backup & data

The database data lives in the volume `.data/mysql` (or the path you configured). Back it up with standard `mysqldump` or by copying the volume data when the database container is stopped.

Example dump (from host or another container):

```bash
docker exec -i <mariadb-container> mysqldump -u root -p --all-databases > backup.sql
```

## Switching from another database

Because the schema is managed by the same bun migrations for all SQL backends, you can point a fresh installation at an existing MySQL/MariaDB server that has had migrations run (or let the migrate container do it).

There is no automatic data migration tool between database families (SQLite ↔ Postgres ↔ MySQL). Use standard dump/restore tools if you need to move data.

## Related pages

- [Docker + PostgreSQL](./docker-with-postgres)
- [Docker + SQLite](./docker-with-sqlite)
- [Docker + MongoDB](./docker-with-mongo)
- [Migration setup](../../architecture/migrate)
