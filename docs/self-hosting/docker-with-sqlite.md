# Docker + SQLite Setup

SQLite is the simplest way to run Monitoring — everything runs in a single container with no external database required. Schema migrations run automatically on first start.

## Quick start

```bash
docker run -d --restart=always \
  -p 8383:8383 \
  -e DB_TYPE=sqlite \
  -e DB_NAME=/app/data/monitoring.db \
  -v $(pwd)/.data/sqlite:/app/data \
  --name monitoring \
  jniltinho/monitoring-bundle-sqlite:latest
```

Open `http://localhost:8383` in your browser.

## Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  monitoring:
    image: jniltinho/monitoring-bundle-sqlite:latest
    restart: unless-stopped
    ports:
      - "8383:8383"
    environment:
      - DB_TYPE=sqlite
      - DB_NAME=/app/data/monitoring.db
    volumes:
      - ./.data/sqlite:/app/data
      - ./.data/logs:/var/log/supervisor
    container_name: monitoring-bundle-sqlite
```

```bash
docker compose up -d
```

## Volumes

| Path (container) | Purpose |
|---|---|
| `/app/data` | SQLite database file |
| `/var/log/supervisor` | Application logs |

## Notes

- **Migrations** run automatically at startup via `monitoring migrate up` (GORM AutoMigrate). No manual step needed.
- No Redis required for single-node SQLite deployments.
- For additional configuration see [Configuration](../configuration.md).
