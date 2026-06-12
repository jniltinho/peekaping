## Context

Peekaping's bundle Docker images currently use a three-process supervisord setup: Redis, the Go API, and Caddy. Caddy's sole job in this context is to serve static frontend files and proxy `/api/*` to the Go API. The React/Vite frontend is compiled into `apps/web/dist` by the `web-builder` Docker stage and then copied into `/app/web` in the final runtime image.

Go 1.16+ provides the `embed` package which allows static files to be embedded into a binary at compile time via `//go:embed`. This lets the API binary carry its own frontend without any external file system or reverse proxy.

## Goals / Non-Goals

**Goals:**
- Embed `apps/web/dist` into the `api` binary using `//go:embed`
- Serve the SPA from the Go API with correct cache headers and `index.html` fallback
- Generate and serve a runtime `env.js` from the API (replacing the Caddy/shell startup step)
- Remove Caddy from bundle Docker images; simplify supervisord config to Redis + api + engine
- Update Dockerfile build order so `web-builder` dist is available before `go build`
- Expose the API directly on port 8034 (or a single configurable port) in the bundle image

**Non-Goals:**
- Removing Caddy from split/production compose deployments (Caddy stays for TLS + WebSocket in prod)
- Changing the frontend source (React/Vite) in any way
- Bundling Redis or the engine into the same binary

## Decisions

### 1. Embed in a dedicated file, not in server.go

A new file `apps/server/internal/frontend/frontend.go` holds the `//go:embed` directive and the `Handler()` factory function. This keeps `server.go` clean and makes the embed easy to stub in tests.

**Alternative considered:** embedding directly in `server.go` — rejected because the `//go:embed` requires the target directory to exist relative to the `.go` file that declares it, and it would clutter the server setup.

The actual embed path would be `//go:embed all:web` relative to `frontend.go`. The compiled `dist/` folder is copied into `apps/server/internal/frontend/web/` during the Docker build (or via a `make` target locally).

### 2. Route priority: API and WebSocket first, static handler last

Echo's router is first-match, so all `/api/*` and `/socket.io/*` routes must be registered before the catch-all static handler. The existing `ProvideServer` function already registers these before the final catch-all — only the Swagger placeholder and the final `handle {}` equivalent need adjusting.

**Static file handler approach:**
- Use `http.FS(embed.FS subpath)` wrapped in `echo.WrapHandler` for asset routes
- Register a custom `GET("/*")` route as the last route that:
  1. Tries to serve the file from the embed.FS
  2. Falls back to serving `index.html` (SPA behavior)

### 3. env.js served from Go, generated at server startup

`env.js` contains runtime configuration (`window.__CONFIG__`) and must not be embedded at compile time (it depends on the `CLIENT_URL` env var which is set at runtime). The API writes a fresh `env.js` to an in-memory variable or a small embedded template and serves it via a dedicated route, bypassing the embed.FS.

**Implementation:** A `GET /env.js` route in the API that renders:
```js
window.__CONFIG__ = { API_URL: "" };
```
This replaces the current shell heredoc in `startup.bundle.*.sh`.

### 4. Dockerfile build order change

The go-builder and web-builder stages currently run independently. To embed, `apps/server/internal/frontend/web/` must be populated before `go build`. The updated Dockerfile:

1. `web-builder` stage: compile frontend → `dist/`
2. `go-builder` stage: depends on `web-builder` output via `COPY --from=web-builder /app/apps/web/dist ./internal/frontend/web`
3. `go build ./cmd/api` — embeds the copied dist at compile time
4. Final stage: only copies `api`, `engine`, `bun` — no Caddy, no `/app/web` directory

### 5. Port change in bundle images: 8034 → exposed directly

With Caddy removed, the bundle images expose port 8034 (the Go API port) directly instead of 8383 (which was Caddy's port). The `EXPOSE` directive in the Dockerfile changes accordingly. Users who previously used port 8383 will need to map to 8034.

**Alternative:** Keep the port as 8383 by configuring the API to listen on 8383 via `SERVER_PORT=8383`. This is less disruptive and is the chosen approach to avoid user-facing breaking changes.

## Risks / Trade-offs

- **Binary size increase** → The `api` binary will grow by the size of the frontend dist (~1–3 MB compressed). Acceptable for the simplification gained.
- **Rebuild required for frontend changes** → Any frontend change requires rebuilding the Go binary. Acceptable for bundle/production use. Development mode still uses Vite's dev server with the proxy config.
- **env.js in-memory vs disk** → Serving env.js from Go memory (a string generated at startup) is simpler than writing it to disk, and avoids permission issues in read-only filesystems.
- **Caddy removal breaks split-compose users who map 8383** → Mitigated by keeping `SERVER_PORT=8383` as the default in bundle supervisord conf, so the URL stays the same for end users.

## Migration Plan

1. Add `apps/server/internal/frontend/` package with embed + handler
2. Update `server.go` to register the frontend handler as the last route
3. Add `GET /env.js` route to the API that serves runtime config
4. Update `Dockerfile.bundle.{sqlite,postgres,mysql}`: copy dist before go build, remove Caddy install
5. Update `supervisord.bundle.*.conf`: remove `[program:caddy]`
6. Update `startup.bundle.*.sh`: remove `env.js` generation (now done by the API)
7. Update `Caddyfile` documentation note (still used for split/prod)
8. Bump `EXPOSE` from 8383 to 8034 in bundle Dockerfiles (or set `SERVER_PORT=8383` in supervisord)

**Rollback:** Reverting commits to the Dockerfiles and `server.go` is sufficient. Caddy was installed from a stable repository and can be re-added to Dockerfiles in one line.

## Open Questions

- Should `SERVER_PORT` default to `8383` in bundle supervisord configs (preserve old external port) or `8034` (canonical API port)? → Chosen: keep 8383 via supervisord `environment=SERVER_PORT="8383"` to avoid breaking existing deployments.
- Should the split/prod compose files (which still use Caddy) have their Caddyfile updated to remove the now-redundant static file blocks? → Out of scope for this change; left for a follow-up.
