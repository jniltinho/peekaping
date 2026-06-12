## Context

The Peekaping server already imports the MySQL driver and registers the mysql dialect for Bun:

- `apps/server/internal/infra/sql.go` has a `case "mysql":` that builds a DSN and uses `mysqldialect`.
- `internal/config/validators.go` already lists `"mysql"` in `validateDBType`.
- `config.go` has a case for "mysql" inside `ValidateDatabaseCustomRules` (but not "mariadb").
- The error message in `formatValidationError` already advertises mysql.
- `GracefulDatabaseShutdown` already routes mysql through the SQL path.
- MIGRATION_SETUP.md and architecture docs claim MySQL support at a high level.
- The features list in README mentions "MySQL/MariaDB -" (incomplete).
- Docker / deployment reality: only sqlite, postgres, and mongo have full bundle Dockerfiles (`Dockerfile.bundle.*`), startup shell scripts, supervisord configs, dedicated compose files, and matching self-hosting markdown guides. No equivalent exists for mysql.

The gap is therefore:
1. `mariadb` is not a legal DB_TYPE anywhere.
2. No production-grade "one docker run" or "docker compose" turnkey artifacts for the mysql family.
3. Inconsistent / incomplete user-facing documentation.

This change makes `mariadb` (and `mysql`) first-class citizens on par with the other three families.

## Goals / Non-Goals

**Goals:**
- Make `DB_TYPE=mariadb` (and `mysql`) fully functional for both microservice (external DB) and bundle (embedded DB) deployments.
- Deliver the same artifact surface as the other DBs: bundle Dockerfile + startup + supervisord + compose yml files (dev/prod + base) + self-hosting doc page.
- Update validation, connection, and shutdown code so that "mariadb" is accepted and treated identically to "mysql" at the wire level.
- Bring documentation, README, Makefile, and architecture tables to consistency.
- Preserve all existing behavior for postgres, sqlite, and mongo.

**Non-Goals:**
- Introduce a dedicated `github.com/go-mysql-org/go-mysql-driver` or other mariadb-specific driver (the standard `go-sql-driver/mysql` is sufficient and already present).
- Add a MariaDB subchart to the Helm chart (documenting external usage is enough; a subchart can be a later increment).
- Change any application-level queries, monitor implementations, or business logic.
- Rewrite migrations or add mysql-specific migration variants.
- Implement advanced mysql-only features (e.g. GTID, replication-aware startup).
- Performance or compatibility matrix testing beyond "it starts and migrations apply".
- Altering the public API or any existing DB contract.

## Decisions

### 1. How to handle "mariadb" vs "mysql" in code
**Decision**: Add "mariadb" to the valid list in `validateDBType`. In `ValidateDatabaseCustomRules` and the big switch in `ProvideSQLDB`, treat "mariadb" exactly like "mysql" (either via an explicit case that falls into the same block or by normalizing the string early to "mysql" for the rest of the call path). Log at info level "Connecting to MySQL-compatible database (type: mariadb)" when the original value was mariadb.

**Rationale**: MariaDB speaks the MySQL protocol; the DSN, driver, and dialect are identical. Keeping the internal string as "mysql" for the rest of the system (e.g. the sqlite-only graceful shutdown guard) minimizes risk. Explicit case keeps logging honest.

**Alternatives considered**:
- Always normalize to "mysql" on load → rejected; loses the user's stated preference in logs and error messages.
- Treat mariadb as completely separate code path → rejected; duplicates the mysql block for no technical gain.

### 2. Base image and packaging for the bundle
**Decision**: Create `Dockerfile.bundle.mysql` following the exact multi-stage pattern of `Dockerfile.bundle.postgres` (go build stages + web build + final debian-slim stage that installs the DB server + redis + supervisor + caddy). Install `mariadb-server` (or `default-mysql-server`) + client from the distro repos. Use a similar supervisord program section for the DB.

**Rationale**: Consistency with the three existing bundles reduces cognitive load for maintainers and users. Debian bookworm (used by postgres bundle) has mature mariadb packages.

**Alternatives considered**:
- Use the official mariadb docker image as base and COPY our binaries in → possible but would require more complex entrypoint merging and would diverge from the "we control the supervisor + init" model used by the other bundles.
- Support only external DB for mysql (no bundle) → rejected by the proposal; users expect the same "docker run one image" experience advertised for the others.

### 3. Startup script complexity
**Decision**: Port the structure of `startup.bundle.postgres.sh`:
- validate required DB_* envs
- generate /app/.env and web env.js
- initialize data dir + create DB + user (using `mysql` / `mariadb-admin` or SQL via the client)
- start the DB temporarily for migrations
- invoke run-migrations.sh
- stop the temp DB
- exec supervisord

Use `mysqladmin ping` or equivalent for readiness waits. Store the "initialized" marker file under the data volume.

**Rationale**: The postgres script is the most complete reference; SQLite one is simpler because it has no separate server process.

### 4. Compose file strategy (microservices)
**Decision**: Add `docker-compose.mysql.yml` (and the dev/prod variants) modeled directly on the postgres equivalents. The DB service will use the `mariadb:latest` (or `mysql:8`) image with appropriate env vars (`MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD`) and a healthcheck using `healthcheck` command or `mysqladmin`. The migrate and api services point DB_TYPE at the database hostname.

**Rationale**: Exact parity lets users copy the postgres instructions and only change the DB_TYPE and image lines.

### 5. Documentation and naming
**Decision**:
- Doc page: `docker-with-mysql.md` (title "Docker + MySQL", content explains MariaDB is the recommended drop-in and shows both image choices).
- In all DB_TYPE tables and comments, list the values as `postgres`, `mysql`, `mariadb`, `sqlite`, `mongo`, `mongodb`.
- In README features, replace the dangling "- MySQL/MariaDB -" with a proper entry (or integrate it cleanly with the other DB monitors if the list is about monitor types; here it appears to be DB support).
- Update the badge list at the top of README only if new official badges are added; otherwise leave the existing three DB badges.

**Rationale**: "mysql" is the config value that actually drives code; "mariadb" is the friendly name many users search for. Documenting both reduces confusion.

### 6. Makefile targets
**Decision**: Add variables `COMPOSE_DEV_MYSQL`, `COMPOSE_PROD_MYSQL`, `COMPOSE_MYSQL` and corresponding phony targets (`dev-mysql`, `prod-mysql`, etc.) following the existing postgres pattern.

## Risks / Trade-offs

- [Startup script for MariaDB is non-trivial] → Mitigation: copy the proven postgres script and adapt only the DB-specific commands (init, client, readiness, ownership). Test the happy path + re-init path manually before declaring done.
- [Some DDL in migrations is not transactional on MySQL] → Mitigation: this is already documented in migrate.md; the spec for the new capability references it. No new risk.
- [Bundle image size] → Trade-off accepted: the postgres bundle is already large because it embeds a full DB server; mysql will be comparable. Users who care about size should use the microservice compose + external managed DB.
- [Healthcheck differences between mysql vs mariadb images] → Mitigation: use a flexible healthcheck in compose (e.g. `mysqladmin ping -h localhost --silent` inside the container) or the official image's HEALTHCHECK if present; document the choice.
- [Version skew between Go driver and server] → Mitigation: pin a recent but not bleeding-edge mariadb/mysql server tag in the bundle Dockerfile and in compose examples (e.g. mariadb:11 or mysql:8.0). The go-sql-driver/mysql is already a transitive dep and well-maintained.

## Migration Plan

No data migration is required for existing users (they are not on mysql today).

For new mysql users:
1. Use the new bundle image or compose file.
2. First start will run the normal bun migration flow (identical to postgres).

Rollback: remove the new files; the change is purely additive. Existing sqlite/postgres/mongo users are unaffected.

If a user was somehow using DB_TYPE=mysql before (possible only in custom external-DB compose), the behavior is unchanged.

## Open Questions

- Exact server package name on Debian (`mariadb-server` vs `default-mysql-server` vs `mysql-server`) — will be resolved during Dockerfile authoring by testing the build.
- Whether to publish a separate `peekaping-bundle-mariadb` tag or just document that the mysql bundle actually contains MariaDB. (Proposal leans toward a single "mysql" bundle tag + docs saying "uses MariaDB by default for license and compatibility reasons".)
- Do we want an automated e2e test that brings up the mysql bundle (similar to existing sqlite test compose)? Out of scope for the initial change but worth a follow-up task.