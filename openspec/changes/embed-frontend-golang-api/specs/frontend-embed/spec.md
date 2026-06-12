## ADDED Requirements

### Requirement: API binary embeds the compiled frontend
The Go API binary SHALL embed the contents of `apps/web/dist` using the `//go:embed` directive so that the frontend is included at compile time with no external filesystem dependency at runtime.

#### Scenario: Binary serves index.html at root
- **WHEN** a GET request is made to `/`
- **THEN** the server responds with the embedded `index.html` and HTTP 200

#### Scenario: Binary serves hashed asset files
- **WHEN** a GET request is made to a path matching a compiled asset (e.g., `/assets/index-abc123.js`)
- **THEN** the server responds with the file content, correct MIME type, and a `Cache-Control: public, max-age=31536000, immutable` header

#### Scenario: SPA fallback for unknown routes
- **WHEN** a GET request is made to a path that is not an API route and not a known static asset (e.g., `/monitors`, `/settings/profile`)
- **THEN** the server responds with the embedded `index.html` and HTTP 200 (enabling client-side routing)

#### Scenario: env.js served with no-cache headers
- **WHEN** a GET request is made to `/env.js`
- **THEN** the server responds with the current `env.js` content and `Cache-Control: no-store, no-cache, must-revalidate`

### Requirement: API binary generates env.js at startup
The API server SHALL write a runtime-generated `env.js` file (containing `window.__CONFIG__`) into a writable location and serve it, allowing runtime configuration without rebuilding the frontend.

#### Scenario: env.js reflects runtime API_URL
- **WHEN** the API server starts with `CLIENT_URL` set
- **THEN** the generated `env.js` sets `window.__CONFIG__.API_URL` to the configured value (or empty string as default)

#### Scenario: env.js is served fresh after restart
- **WHEN** the API server restarts with a different `CLIENT_URL`
- **THEN** subsequent requests to `/env.js` return the updated configuration

### Requirement: Dockerfile build order ensures frontend is embedded
The Dockerfile multi-stage build SHALL copy the compiled frontend `dist/` from the `web-builder` stage into the Go source tree before `go build`, so the `//go:embed` directive can reference the files.

#### Scenario: Go build fails if dist is missing
- **WHEN** the Go API binary is compiled without the `apps/web/dist` directory present
- **THEN** the build fails with a clear error indicating the embed path is missing

#### Scenario: Single Docker stage produces self-contained API
- **WHEN** `docker build` completes for any bundle Dockerfile variant
- **THEN** the resulting `api` binary serves both API and frontend without requiring Caddy

### Requirement: Bundle Docker images do not require Caddy
The bundle Docker images (`Dockerfile.bundle.sqlite`, `.postgres`, `.mysql`) SHALL NOT install or run Caddy, relying solely on the API binary for HTTP serving.

#### Scenario: Container starts without Caddy process
- **WHEN** the bundle container starts
- **THEN** `supervisorctl status` shows only `redis`, `api`, and `engine` programs — no `caddy` program

#### Scenario: Frontend accessible on API port
- **WHEN** a browser requests `http://localhost:8034/`
- **THEN** the React SPA is served directly by the Go API

### Requirement: API routes take precedence over static file serving
The API server SHALL register all `/api/*` and `/socket.io/*` routes before the static file handler so that API paths are never intercepted by the SPA fallback.

#### Scenario: API path not shadowed by SPA fallback
- **WHEN** a GET request is made to `/api/v1/monitors`
- **THEN** the response comes from the monitor controller, not from `index.html`

#### Scenario: WebSocket path not shadowed by static handler
- **WHEN** a WebSocket upgrade request is made to `/socket.io/...`
- **THEN** the request is handled by the WebSocket server, not the file server
