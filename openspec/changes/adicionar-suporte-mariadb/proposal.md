## Why

The application core (config validation + Bun SQL provider) already contains partial support for MySQL (`DB_TYPE=mysql`), but `mariadb` is not recognized as a valid value. There are no equivalent bundle Dockerfiles, startup scripts, supervisord configs, docker-compose files, or self-hosting guides for MySQL/MariaDB (unlike the complete coverage for SQLite, PostgreSQL and MongoDB). Documentation is inconsistent (e.g. incomplete bullet in README features, missing dedicated setup page) and deployment options are incomplete. Users who prefer or are required to use MariaDB/MySQL cannot easily adopt the project with the same convenience as other databases.

## What Changes

- Accept `mariadb` (and keep/ensure `mysql`) as valid `DB_TYPE` values in validators, custom rules, connection logic, and shutdown helpers.
- Normalize or add explicit `mariadb` case in `ProvideSQLDB` / `ValidateDatabaseCustomRules` / `Graceful...` (reuse the existing MySQL driver/dialect since MariaDB is wire-compatible).
- Add complete "bundle" (monolithic easy single-container) support: `Dockerfile.bundle.mysql`, `supervisord.bundle.mysql.conf`, `startup.bundle.mysql.sh`, `docker-compose.bundle.mysql.yml`.
- Add microservice docker-compose files for external MySQL/MariaDB: `docker-compose.mysql.yml` + dev/prod variants to match existing patterns.
- Add self-hosting documentation: `apps/docs/docs/self-hosting/docker-with-mysql.md` (covering both monolithic bundle and microservice compose, with MariaDB recommended image where appropriate).
- Update all references: README.md (fix features list, quick-start mentions), development.md, architecture/*.md (consistent DB_TYPE tables), MIGRATION_SETUP.md, Makefile (new COMPOSE_* vars and targets), docs sidebars if auto-generated.
- Update Helm chart values / docs to document external MySQL usage (or add mariadb subchart dependency if decided).
- Ensure migrations (bun) work with mysql/mariadb (already claimed in migrate.md; validate coverage).
- No changes to public API or existing database behavior for other types.

## Capabilities

### New Capabilities
- `mysql-mariadb-support`: Defines the requirements for first-class MySQL and MariaDB support as a `DB_TYPE`, including validation, runtime connection, bundle and compose deployment artifacts, and documentation parity with Postgres/SQLite/Mongo.

### Modified Capabilities
(none — this is additive support for a new database family; no existing top-level spec in `openspec/specs/` is modified)

## Impact

**Primary affected areas**:
- `apps/server/internal/config/` — validators.go (add "mariadb"), config.go (update error message + ValidateDatabaseCustomRules to handle "mariadb").
- `apps/server/internal/infra/sql.go` — extend switch and shutdown functions to treat "mariadb" like "mysql".
- New Docker artifacts under repo root (modeled exactly after postgres/sqlite/mongo bundles).
- New docs file + updates to existing docs and README.
- `Makefile` — add variables and phony targets for mysql dev/prod/bundle.
- Possibly `charts/values.yaml` and related templates for external DB guidance.
- CI / test scripts if any DB-specific matrix exists (e.g. test_utils/ similar to mssql).

**Downstream**:
- Users gain `docker run ... 0xfurai/peekaping-bundle-mysql:latest` (or mariadb image tag) and compose-based setups.
- Documentation site will have a new "Docker + MySQL" page.
- Future maintenance must keep the four DB families in sync for bundle + docs.

**Out of scope**:
- Changing the MySQL driver or introducing a dedicated mariadb driver (the existing `go-sql-driver/mysql` + `mysqldialect` is sufficient and already imported).
- Adding native MariaDB healthchecks or special SQL in application code (standard SQL + Bun handles it).
- Full Helm subchart for embedded MariaDB (external DB documented; can be follow-up).
- Performance or feature parity testing specific to MariaDB vs MySQL (basic connectivity + migration coverage is the goal).