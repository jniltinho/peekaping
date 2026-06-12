## 1. Frontend Embed Package

- [x] 1.1 Create directory `apps/server/internal/frontend/` and add `apps/server/internal/frontend/web/.gitkeep` (the embed target; populated at build time)
- [x] 1.2 Create `apps/server/internal/frontend/frontend.go` with `//go:embed all:web` directive and `Handler(cfg *config.Config) echo.HandlerFunc` that serves files from the embedded FS with SPA fallback to `index.html`
- [x] 1.3 Add cache-header logic in the handler: `Cache-Control: public, max-age=31536000, immutable` for hashed assets (paths matching `*.js`, `*.css`, `*.woff*`, `*.svg`, `*.png`, `*.jpg`, `*.ico`), `Cache-Control: no-store` for `index.html`

## 2. env.js Runtime Endpoint

- [x] 2.1 Add `GET /env.js` route in `apps/server/internal/server.go` (before the static catch-all) that returns a JS snippet setting `window.__CONFIG__ = { API_URL: "" }` with `Content-Type: application/javascript` and `Cache-Control: no-store, no-cache, must-revalidate`
- [x] 2.2 The endpoint must respond even when the embed.FS web directory is empty (i.e., during local development without a built frontend)

## 3. Server Route Wiring

- [x] 3.1 In `apps/server/internal/server.go`, import the new `frontend` package and register `frontend.Handler(cfg)` as the last catch-all route: `e.GET("/*", frontend.Handler(cfg))`
- [x] 3.2 Verify route registration order: all `/api/*`, `/socket.io/*`, `/swagger/*`, and `/env.js` routes must be registered before the catch-all

## 4. Dockerfile Build Pipeline

- [x] 4.1 Update `Dockerfile.bundle.sqlite`: reorder stages so `web-builder` runs first; add `COPY --from=web-builder /app/apps/web/dist ./internal/frontend/web` in `go-builder` before `go build ./cmd/api`
- [x] 4.2 Update `Dockerfile.bundle.postgres`: same change as 4.1
- [x] 4.3 Update `Dockerfile.bundle.mysql`: same change as 4.1
- [x] 4.4 Remove the `COPY --from=web-builder /app/apps/web/dist /app/web` line and the `RUN mkdir -p /app/web` line from the final stage of all three bundle Dockerfiles
- [x] 4.5 Remove Caddy installation block (`curl ... caddy ... apt-get install -y caddy`) from all three bundle Dockerfiles
- [x] 4.6 Remove `COPY Caddyfile /etc/caddy/Caddyfile` from all three bundle Dockerfiles
- [x] 4.7 Change `EXPOSE 8383` to `EXPOSE 8034` in all three bundle Dockerfiles (or keep 8383 if supervisord sets `SERVER_PORT=8383`)

## 5. Supervisord Configuration

- [x] 5.1 Update `supervisord.bundle.sqlite.conf`: remove `[program:caddy]` section; add `SERVER_PORT="8383"` to the `[program:api]` environment so the external port stays the same
- [x] 5.2 Update `supervisord.bundle.postgres.conf`: same change as 5.1
- [x] 5.3 Update `supervisord.bundle.mysql.conf`: same change as 5.1

## 6. Startup Scripts

- [x] 6.1 Update `startup.bundle.sqlite.sh`: remove the `env.js` generation block (the heredoc writing to `/app/web/env.js` and the `chmod` line that follows)
- [x] 6.2 Update `startup.bundle.postgres.sh`: same change as 6.1
- [x] 6.3 Update `startup.bundle.mysql.sh`: same change as 6.1

## 7. Local Development Support

- [x] 7.1 Add a `make embed-web` (or similar) Makefile target that runs `pnpm --filter=web run build` and copies `apps/web/dist/` to `apps/server/internal/frontend/web/`, enabling local testing of the embed without Docker
- [x] 7.2 Add `apps/server/internal/frontend/web/` to `.gitignore` (build artifact, not source)

## 8. Validation

- [x] 8.1 Run `go build ./apps/server/cmd/api` after `make embed-web` — binary must compile with no embed errors
- [x] 8.2 Run `docker build -f Dockerfile.bundle.sqlite -t peekaping-bundle-sqlite:embed-test .` — build must succeed
- [x] 8.3 Run `docker run -p 8383:8034 peekaping-bundle-sqlite:embed-test` (or `SERVER_PORT=8383`) and verify `http://localhost:8383/` serves the React SPA
- [x] 8.4 Verify `http://localhost:8383/api/v1/version` returns JSON (API routes not shadowed by static handler)
- [x] 8.5 Verify `http://localhost:8383/env.js` returns the runtime config JS
- [x] 8.6 Verify `http://localhost:8383/monitors` returns `index.html` (SPA fallback works)
- [x] 8.7 Confirm `supervisorctl status` inside the container shows `redis`, `api`, `engine` only — no `caddy`
