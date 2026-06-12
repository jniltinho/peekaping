## Why

Currently the frontend (React/Vite) is served by Caddy as a separate process inside the bundle Docker image, requiring the full Caddy installation and a multi-process supervisor setup just to serve static files. Embedding the compiled frontend directly into the Go API binary eliminates this dependency, produces a simpler deployment artifact, and makes the API self-contained for development and lightweight deployments.

## What Changes

- Add `//go:embed` directive in the API server to bundle `apps/web/dist` into the binary at compile time
- Register an Echo file-server handler that serves the embedded SPA (with fallback to `index.html` for unknown paths)
- Update the Dockerfile build pipeline so the frontend is compiled before the Go binary (the `dist/` folder must exist at `go build` time)
- Remove Caddy from the bundle Docker images (the Go API now handles static file serving)
- Remove supervisord `[program:caddy]` and the Caddyfile dependency from bundle images
- Update `docker-compose.prod.*` to expose the API port (8034) directly or via a minimal reverse proxy

## Capabilities

### New Capabilities

- `frontend-embed`: Embedded SPA serving inside the Go API — the binary serves the React frontend from `embed.FS`, with SPA fallback (`index.html` for all non-API, non-asset paths), cache headers for hashed assets, and `no-store` for `index.html` and `env.js`

### Modified Capabilities

*(none — no existing spec-level requirements change)*

## Impact

- **`apps/server/internal/server.go`**: add `go:embed` directive and Echo static-file + SPA routes
- **`apps/server/cmd/api/`**: no change (server construction is unchanged)
- **`Dockerfile.bundle.sqlite/postgres/mysql`**: add a build-order dependency so `web-builder` dist is copied into the Go source tree before `go build`; remove Caddy installation and runtime copy
- **`supervisord.bundle.*.conf`**: remove `[program:caddy]` section
- **`Caddyfile`**: still used in split (prod) deployments; untouched for the split compose files
- **`apps/web/src/`** or **`apps/server/cmd/api/`**: add an `env.js` generation step (currently done by the Caddy startup script) that must now happen at runtime inside the Go server or via the startup script
- No changes to database, Redis, engine, or notification modules
