## Context

The Peekaping application currently supports four database backends for its own persistence:

- SQLite (file-based, special-cased in many places for WAL, single-writer, etc.)
- PostgreSQL (recommended for production, full feature set)
- MySQL / MariaDB (recently brought to full parity including bundle Docker images, startup scripts, and self-hosting docs)
- MongoDB (document store, schemaless, no migrations, completely separate connection/client code, repository implementations, and deployment artifacts)

MongoDB support lives in:
- `apps/server/internal/infra/mongo.go` (ProvideMongoDB + command monitor)
- `apps/server/internal/config/config.go` + `validators.go` (valid types + custom rules)
- `apps/server/internal/utils/dig.go` (RegisterRepositoryByDBType switch)
- Per-module `*.mongo.repository.go` files (e.g. monitor) that implement the same interfaces as their SQL counterparts using the mongo driver and bson models.
- Graceful shutdown paths (still partially in sql.go).
- A complete parallel set of bundle artifacts (Dockerfile, supervisord, startup.sh) and compose files (base + dev + prod) plus a dedicated self-hosting guide.
- Extensive references in README, docs, Makefile, charts, and architecture tables.
- Default dev/prod DB in Makefile is currently mongo.

The recent `adicionar-suporte-mariadb` change completed SQL coverage. Keeping MongoDB now represents pure maintenance drag with little incremental value for most users. The proposal therefore calls for its complete removal.

## Goals / Non-Goals

**Goals:**
- Remove "mongo" and "mongodb" from every valid DB_TYPE list, validator, error message, and runtime switch.
- Delete the dedicated MongoDB infrastructure code (infra/mongo.go, mongo-specific repos, registration helper branches).
- Delete every MongoDB-specific deployment artifact (Dockerfile, startup script, supervisord config, all compose variants).
- Delete the dedicated self-hosting guide and remove MongoDB storage references from all user-facing documentation and build tooling.
- Simplify DI and repository provisioning to be SQL-only.
- Leave the "MongoDB monitor type" (the ability to monitor external MongoDB servers) completely untouched.
- Provide clear breaking-change guidance and migration steps for existing MongoDB users.
- Produce a clean codebase that only needs to reason about Bun + the three SQL variants.

**Non-Goals:**
- Building any data-migration tooling or automated import from MongoDB to SQL.
- Keeping the mongo DB_TYPE values as deprecated/alias shims.
- Removing or altering the MongoDB *monitor executor* (separate concern from the app's own DB).
- Changing public API contracts, response shapes, or any non-persistence module behavior.
- Updating the React frontend or any client generation.
- Preserving the old mongo bundle images under new tags or compatibility layers.

## Decisions

### 1. Deletion strategy for code and files
**Decision**: Physically delete the mongo-specific files and directories rather than commenting them out or leaving stub implementations.

**Rationale**: A clean break reduces future confusion, shrinks the repo, removes the mongo-driver dependency (once no longer referenced), and forces any missed references to surface as compile errors. Comments and stubs tend to rot.

**Alternatives considered**:
- Keep files behind build tags → rejected (still ships dead weight and complexity).
- Deprecate with warnings for one release → rejected (proposal explicitly wants a clean removal; deprecation would require more ongoing work).

### 2. Handling of per-module mongo repositories
**Decision**: Delete every `*.mongo.repository.go` (monitor and any other modules that have them) together with the registration switch in dig.go. After removal, all repositories will be provided via the SQL constructor path unconditionally (or the helper function itself can be inlined/removed if it becomes a no-op).

**Rationale**: These files exist solely to satisfy the previous dual-backend strategy. Once the mongo backend is gone they are dead code.

### 3. Default database in Makefile and examples
**Decision**: Change `DEFAULT_DEV_DB` and `DEFAULT_PROD_DB` from `mongo` to `postgres` (the documented recommended production choice). Keep sqlite as the easy local default where appropriate.

**Rationale**: Postgres is the most fully featured and commonly recommended backend after the recent SQL unification work.

### 4. Mongo driver dependency
**Decision**: After deleting infra/mongo.go and the mongo repo implementations, run `go mod tidy`. If any monitor-type code that talks *to* external MongoDB instances still imports the driver, keep the dependency (isolated to the monitor package). Otherwise remove it entirely from go.mod.

**Rationale**: The driver was only pulled in for the app's own MongoDB backend plus (potentially) monitor targets. We want to drop it for the backend use case.

### 5. Documentation removal vs. historical note
**Decision**: Delete `docker-with-mongo.md` outright. In README and other docs, replace MongoDB storage claims with the new SQL-only list and add a short "Previously supported" note only if it improves the migration story. Do not keep the full guide.

**Rationale**: Keeping the guide would imply ongoing support. A short migration call-out in the proposal/spec and a deprecation note in the change log is sufficient.

## Risks / Trade-offs

- [Existing MongoDB users experience a hard break on upgrade] → Mitigation: prominent **BREAKING** markers in proposal, spec (with Reason + Migration for every removed requirement), README, and release notes. Clear instructions to dump data and re-import into a supported SQL backend before upgrading.
- [Incomplete removal leaves references or compile errors] → Mitigation: systematic `grep -ri mongo` (excluding monitor-type code and docs that legitimately mention external Mongo targets), followed by `go build ./...` and `go mod tidy`. The design favors deletion so that the compiler becomes the safety net.
- [Docker / compose examples in user projects or third-party docs become stale] → Mitigation: the change is additive in the sense that we are publishing fewer options; the three SQL guides will be the only ones promoted.
- [Reduced "flexible storage" marketing claim] → Trade-off accepted: three well-supported SQL options (with full bundle parity) is simpler and more maintainable than four heterogeneous backends. The proposal explicitly calls this out.

## Migration Plan

For users on MongoDB today:
1. Use `mongodump` (or equivalent) to export data while the mongo-backed instance is still running.
2. Provision a new supported SQL database (Postgres recommended).
3. Point a new/fresh Peekaping installation at the SQL database (`DB_TYPE=postgres` etc.) with appropriate credentials.
4. Let the migrate container run (or run `bun db migrate` manually).
5. Import the dumped data using appropriate SQL tools or custom scripts (no automated importer is provided).
6. Verify monitors, notifications, and status pages.
7. Switch production traffic / update compose files / rebuild bundles as needed.

Rollback for the project itself (if the removal decision is later reversed): the change is purely subtractive. Re-adding mongo would require restoring the deleted files from git history and re-introducing the registration branches. No data-model changes are made to the SQL side.

## Open Questions

- Exact list of modules that have `*.mongo.repository.go` files (monitor is confirmed; a full grep during implementation will reveal others such as heartbeat, notification history, etc.).
- Whether any shared bson utility code in `internal/utils/bson.go` can be deleted or is used by monitor-type executors that talk to external Mongo instances.
- Whether the `go.mongodb.org/mongo-driver` can be fully removed from go.mod or must remain for monitor targets.
- Should the default in development examples switch all the way to sqlite for "zero-config" friendliness, or stay on postgres? (Current Makefile used mongo; proposal suggests postgres as the new default.)