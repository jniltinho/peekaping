## Context

The entire HTTP delivery layer of the Peekaping API server (`apps/server/internal`) is implemented using Gin v1.10.0:

- Central engine + global middleware setup lives in `internal/server.go`.
- ~12 functional modules expose routes via a consistent `XxxRoute.ConnectRoute(*gin.RouterGroup, *XxxController)` pattern.
- Controllers are thin adapters: they extract params/queries, bind JSON, call services, and return standardized `utils.ApiResponse` via `ctx.JSON(...)`.
- Three layers of auth middleware + a sophisticated brute-force guard are Gin-specific.
- A push ingestion endpoint and socket.io compatibility shim perform raw `ResponseWriter`/`Request` delegation.
- Query extraction helpers in `internal/utils/query.go` are written against `*gin.Context`.
- The official module generator templates (`templates/module/*.tmpl`) emit Gin code.
- All related tests use `gin.CreateTestContext` + `httptest` recorder patterns.
- Swagger is served via `swaggo/gin-swagger` on top of swag-generated annotations (the annotations themselves are mostly router-agnostic).

The `ProvideServer` constructor (injected via Uber Dig) receives every Route and Controller explicitly. `main.go` only cares about calling `.Run(port)` on the router and graceful shutdown.

No other part of the system (healthcheck executors, producer, ingester, services, repositories, events) has any knowledge of the web framework.

## Goals / Non-Goals

**Goals:**
- Replace Gin with Echo Framework v5 as the sole HTTP router and context abstraction for the API server.
- Port all existing functionality (routing, CRUD per module + extra endpoints, dual JWT+API-key auth, brute-force protection on login, push endpoint, socket.io shims, CORS, recovery, health/version endpoints, Swagger UI) with identical external behavior.
- Update the module code generator so that newly scaffolded modules produce Echo code.
- Keep the public HTTP API contract, response shapes, authentication semantics, and error messages unchanged.
- Produce a clean, maintainable Echo implementation that is at least as easy to extend as the current Gin one.
- Update or remove all Gin references in go.mod, tests, and templates.

**Non-Goals:**
- Changing any service/repository/domain logic or adding new API endpoints.
- Altering the OpenAPI/Swagger contract (regenerate docs after the port).
- Introducing a new abstraction layer over Echo (or keeping Gin behind an interface) — direct use of Echo types in the HTTP layer is acceptable.
- Performance benchmarking or optimization as primary driver (though Echo's characteristics are a secondary benefit).
- Multi-process or blue/green deployment strategy for the migration itself.
- Updating the frontend, Terraform provider, or any client beyond what the stable API contract already provides.

## Decisions

### 1. Big-bang migration vs incremental / dual-framework support
**Decision**: Perform a single, complete cut-over of the HTTP layer in this change.

**Rationale**: The handler surface is large (dozens of methods) but extremely uniform. Maintaining two parallel implementations (Gin + Echo) would require duplicating the generator templates, middleware, and test helpers for the duration of the migration — exactly the coupling we want to escape. The push and socket.io paths already do raw HTTP work, which reduces risk.

**Alternatives considered**:
- Strangler fig pattern with a feature flag or separate port → rejected (too much temporary complexity for an internal refactor).
- Interface-based router abstraction → rejected for now; would add indirection with little long-term value since we are choosing one framework.

### 2. Echo import path and version
**Decision**: Target `github.com/labstack/echo/v4` (current stable major) or the v5 path if a stable `github.com/labstack/echo/v5` is released and desired. Pin a concrete version in go.mod. Document the exact chosen module in the tasks.

**Rationale**: As of the time of writing, the widely-used and well-supported Echo is v4. User request specified "v5"; we will clarify the precise module during implementation. The porting patterns are very similar between recent Echo majors.

### 3. Handling of Gin-specific utilities
**Decision**:
- Port `GetQueryInt` / `GetQueryBool` to Echo equivalents inside the same `internal/utils/query.go` (or a small `query_echo.go` if we want to keep the file clean). Echo's `c.QueryParam(name)` + strconv is straightforward.
- The response helpers in `internal/utils/http.go` (ApiResponse, NewSuccess/FailResponse, URIParams, PaginatedQueryParams) are already framework-agnostic — leave untouched.
- Remove or deprecate any unused Gin binding structs if they exist.

### 4. Middleware porting strategy
**Decision**:
- Convert every `gin.HandlerFunc` to `echo.MiddlewareFunc` (i.e. `func(next echo.HandlerFunc) echo.HandlerFunc`).
- For the brute-force `Guard` (the most stateful middleware — it inspects `c.Writer.Status()` after the handler runs):
  - Wrap the response writer or use `c.Response().Status` / a deferred status capture after `next(c)`.
  - Keep the `KeyExtractor` and `OnBlocked` callbacks but change their signatures to accept `echo.Context`.
- AuthChain and per-module auth middlewares port by replacing `c.GetHeader`, `c.Set`, `c.Abort()`, `c.Next()`, `c.ClientIP()`, `c.Request` with Echo equivalents (`c.Request().Header`, `c.Set`, `c.Response().WriteHeader` + return error for abort, `c.Request().Context()`, etc.).
- Prefer Echo's built-in `middleware.CORS()` and `middleware.Recover()` over the gin-contrib packages.

### 5. Special endpoints (push + socket.io)
**Decision**:
- Port `RegisterPushEndpoint` to accept `*echo.Group` and use `c.Param`, `c.QueryParam`, `c.Bind` (or query helpers), and return `c.JSON(...)` (which returns error).
- For socket.io: use `e.Any("/socket.io/*", func(c echo.Context) error { wsServer.ServeHTTP(c.Response().Writer, c.Request()); return nil })` or equivalent raw access. Echo's context gives direct access to the underlying request/response.

### 6. Swagger / OpenAPI serving
**Decision**:
- Keep all existing swag `// @Router`, `@Summary`, `@Security` etc. annotations on controller methods (they are independent of the router library).
- Replace the serving code: remove `ginSwagger.WrapHandler`, use either `swaggo/echo-swagger` (if a maintained package exists for the chosen Echo major) or mount the generated `/swagger/doc.json` + serve the static Swagger UI assets via Echo's static file or a small custom handler.
- Regenerate swagger docs after the code changes (`swag init` or equivalent).

### 7. Module generator templates
**Decision**: Update both `route.go.tmpl` and `controller.go.tmpl` (and any other relevant templates) to emit Echo code as part of this change. This is mandatory so that future modules do not regress to Gin.

The generator script itself (`generate_module.go`) does not need logic changes — only the template content.

### 8. Server struct and startup
**Decision**:
- Change `Server.Router` type from `*gin.Engine` to `*echo.Echo` (or keep the field name for minimal diff in main.go and call it `Engine`).
- In `cmd/api/main.go`, change the start line from `server.Router.Run(port)` to `server.Router.Start(port)` (Echo's equivalent). Handle graceful shutdown similarly (Echo has `Shutdown` support).
- Keep the DI signature of `ProvideServer` as-is in spirit (still receives all the routes/controllers); only the internal wiring changes.

### 9. Testing approach
**Decision**:
- Update all `_test.go` files that create Gin test contexts to use Echo test patterns:
  ```go
  e := echo.New()
  req := httptest.NewRequest(...)
  rec := httptest.NewRecorder()
  c := e.NewContext(req, rec)
  ```
- For middleware tests, exercise the middleware directly against Echo contexts.
- Add at least one smoke test that the routes are registered (or rely on existing integration/e2e tests that hit the real server).
- Service-level tests that do not touch HTTP remain untouched.

### 10. Rollback / safety
Because this is a compile-time replacement, the main safety is:
- Keep the change in a feature branch.
- Run full test suite + manual smoke of all major endpoints (auth, monitors CRUD + extra actions, push, status pages, notifications, etc.).
- Use the existing e2e Playwright suite and any API integration tests.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Subtle behavior differences in binding, query parsing, error responses, or header handling | Systematic port + side-by-side manual testing of every route group. Keep the standardized response helpers. |
| Brute-force guard logic (post-handler status inspection) is tricky to port cleanly | Design a small response status capturer middleware or use Echo's `c.Response().Status` after next. Add unit tests for the guard. |
| Swagger UI serving becomes more manual or a less polished package | Fall back to serving the static spec + official Swagger dist via Echo static or a tiny handler. The JSON contract stays identical. |
| Generator templates get out of sync again in the future | The migration itself updates them; document in contributing guide that the templates must be kept in sync with the chosen framework. |
| Large number of files touched increases chance of missing a `gin.` reference | Use grep + build after each logical slice (utils → server core → middlewares → routes/controllers batch → special handlers → tests → generator). |
| Echo v5 module path or API differences from v4 | Pin exact version early in tasks; if v5 is unstable, document decision to use latest stable v4 and note the user request. |
| Socket.io library expectations on the handler | The current code already passes raw writer/req; the port should be mechanical. Test the realtime dashboard connection. |

## Migration Plan (high level)

1. **Preparation** — Update go.mod (add Echo, remove Gin + gin-contrib + gin-swagger), run `go mod tidy`.
2. **Foundation** — Port `internal/utils/query.go`, response helpers stay, create any small Echo adapters.
3. **Core server** — Rewrite `internal/server.go`: engine creation, CORS (use Echo middleware), health/version, route mounting, push registration, socket.io, swagger mount.
4. **Middleware layer** — Port auth middlewares, AuthChain, bruteforce guard + its tests.
5. **All route + controller modules** — Systematic port (can be done module-by-module or in batches).
6. **Special handlers** — Push and websocket shims.
7. **Generator templates** — Update tmpl files and verify by running the generator for a throwaway module.
8. **Tests** — Fix all Gin test code; run full `go test ./...`.
9. **Swagger regeneration** + manual verification of UI.
10. **main.go & startup** — Minor adaptation for `.Start` / shutdown.
11. **Cut-over & validation** — Build, run with docker-compose variants, exercise UI + API + push monitors + realtime updates. Update any docs that mention Gin.

Rollback is "revert the commit / branch" because there is no runtime coexistence.

## Open Questions

- Exact Echo module to depend on (`github.com/labstack/echo/v4` vs a v5 path). Confirm and pin.
- Is there a maintained `swaggo/echo-swagger` equivalent for the chosen major, or do we implement a small static + spec server?
- Do we want to keep the field name `Router` on `Server` struct (for minimal main.go change) even if it holds an `*echo.Echo`?
- Any desire to expose Echo's more advanced features (e.g. binder customization, route naming, better error handling middleware) as part of this migration, or strict port only?
- Should we add a small comment or godoc in the HTTP layer files noting "Echo implementation of the presentation layer" for future maintainers?