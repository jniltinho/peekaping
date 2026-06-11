## Why

The API server module in `apps/server` is built entirely on the Gin web framework (v1.10). All HTTP routing, middleware, request handling, response formatting, and the code generator templates are tightly coupled to Gin's types (`*gin.Context`, `*gin.RouterGroup`, `gin.HandlerFunc`, `gin.H`, `ShouldBindJSON`, etc.).

This coupling creates long-term maintenance friction:
- The module code generator (`scripts/generate/generate_module.go` + templates) hardcodes Gin-specific code for every new monitor/notification/etc module.
- Custom query helpers, auth chain middleware, brute-force guard, push endpoint, and socket.io compatibility shims are all Gin-aware.
- Every controller method and route registration file carries Gin imports and idioms.
- Swagger serving is done via `swaggo/gin-swagger`.

Migrating to Echo (targeting v5) modernizes the HTTP layer, aligns with the project's stated goal of an "easily extensible server architecture", and reduces the surface area that must be touched when adding new endpoints or evolving the framework. Echo offers a clean middleware model, strong performance characteristics, and a more explicit handler contract (`func(echo.Context) error`).

## What Changes

- **BREAKING (internal)**: Complete replacement of the Gin HTTP framework with Echo Framework v5 for the API server.
- All route registration logic, controllers, and HTTP-specific middleware are rewritten to use Echo primitives.
- Framework-specific utilities (`internal/utils/query.go`) are either removed or adapted.
- The module scaffolding templates are updated to generate Echo-compatible code going forward.
- Dependency changes: remove `gin-gonic/gin`, `gin-contrib/*`, `swaggo/gin-swagger`; add Echo and appropriate Echo swagger integration.
- Test code that constructs Gin test contexts or uses Gin-specific test helpers is updated.
- Socket.io compatibility routes and the push endpoint are ported (they already do a lot of raw `http.ResponseWriter`/`Request` work, which helps).
- The single `ProvideServer` DI function and `internal.Server` struct are updated (still exposes the router/engine for `main.go` to call `.Run` or equivalent).
- No changes to external HTTP API contract, response shapes (`ApiResponse`), authentication semantics, or business logic in services.

## Capabilities

### New Capabilities

(none — this is an internal implementation migration of the existing HTTP delivery mechanism)

### Modified Capabilities

(none — the observable API behavior and requirements do not change; only the framework implementing the routes changes. No delta specs required in `openspec/specs/`.)

## Impact

**Primary affected area**: `apps/server/internal/` (the entire "API module")

- `internal/server.go` — engine creation, global middleware (CORS), route mounting, swagger mount, socket.io shims.
- `internal/utils/query.go` — Gin-specific query extraction helpers (used by almost every controller).
- Every `*/<module>.route.go` (auth, monitor, notification_channel, proxy, setting, maintenance, status_page, tag, badge, api_key, etc.).
- Every `*/<module>.controller.go` (handler methods using `*gin.Context`).
- `internal/modules/middleware/auth_chain.go` and per-module auth/api_key middleware.
- `internal/modules/bruteforce/bruteforce.go` (guard middleware).
- `internal/modules/healthcheck/push_handler.go`.
- `internal/modules/websocket/websocket.go` (indirectly, via raw handler).
- All unit/integration tests that exercise HTTP handlers or middleware.
- `apps/server/templates/module/route.go.tmpl` and `controller.go.tmpl`.
- `apps/server/go.mod` / `go.sum` and generated `docs/` swagger files (regeneration step).
- `cmd/api/main.go` (minor — still invokes the server and calls Run).
- Potential updates to development docs and Makefile if any Gin-specific dev tooling exists.

**Downstream**:
- Future modules created via the generator will follow the new Echo pattern.
- Any external consumers of the `internal.Server.Router` type will see an `*echo.Echo` (or wrapper) instead of `*gin.Engine`.
- CI / Docker builds will pull the new dependency set.

**Out of scope for this change**:
- Changes to services, repositories, domain models, or any non-HTTP code.
- Introduction of new API endpoints or behavioral changes.
- Multi-user / RBAC / incidents features (those remain separate roadmap items).
- Database or queue layer.
- Frontend client generation (the OpenAPI contract should stay stable).