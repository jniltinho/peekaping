# Docker + SQLite Setup

SQLite is the simplest way to run Peekaping — everything runs in a single container with no external database required.

## Quick start

```bash
docker run -d --restart=always \
  -p 8383:8383 \
  -e DB_NAME=/app/data/peekaping.db \
  -v $(pwd)/.data/sqlite:/app/data \
  --name peekaping \
  0xfurai/peekaping-bundle-sqlite:latest
```

Open `http://localhost:8383` in your browser.

## Docker Compose

Create a `docker-compose.yml`:

```yaml
services:
  peekaping:
    image: 0xfurai/peekaping-bundle-sqlite:latest
    restart: unless-stopped
    ports:
      - "8383:8383"
    environment:
      - DB_TYPE=sqlite
      - DB_NAME=/app/data/peekaping.db
    volumes:
      - ./.data/sqlite:/app/data
      - ./.data/logs:/var/log/supervisor
    container_name: peekaping-bundle-sqlite
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

- No Redis required — the engine runs without a queue when `ENGINE_USE_REDIS=false` (default).
- For additional configuration see [Configuration](../configuration.md).
