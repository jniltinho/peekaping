## Why

Maintaining parallel support for MongoDB (as a schemaless document store) in addition to the SQL backends (SQLite, PostgreSQL, MySQL/MariaDB via Bun) imposes a high maintenance cost across the codebase, deployment artifacts, documentation, and testing. Separate connection logic, repository registration, bundle Dockerfiles/startup/supervisord, compose files, graceful shutdown paths, and extensive docs duplication exist only because of MongoDB. After recently completing comprehensive MariaDB/MySQL support (including bundles), the cost/benefit of keeping four DB families no longer justifies the complexity. Removing MongoDB simplifies the project to a consistent SQL-only storage layer.

## What Changes

- **BREAKING**: Remove "mongo" and "mongodb" as accepted values for `DB_TYPE`. Validation and custom rules will reject them with clear errors.

- Remove all MongoDB-specific infrastructure: `ProvideMongoDB`, client connection/options, command monitoring, graceful disconnect logic, and conditional repository registration in DI helpers.

- Clean up or delete mongo-specific repository code (e.g. monitor mongo repository) and any bson utilities that were mongo-only.

- **BREAKING / removal**: Completely delete all MongoDB deployment artifacts (Dockerfile.bundle.mongo, supervisord.bundle.mongo.conf, startup.bundle.mongo.sh, all `docker-compose.*.mongo.yml` variants) and associated example volumes/scripts.

- Remove `docker-with-mongo.md`; strip all MongoDB references from README (badges, storage paragraph, quick-start links, monitor list if applicable), development.md, architecture component docs, MIGRATION_SETUP.md, Makefile (COMPOSE_ vars, targets, help text, down-all), and Helm chart values/comments.

- Update Graceful shutdown, DI registration, and any remaining conditionals that branched on mongo/mongodb.

- The migrate component (already documented as not applying to MongoDB) will have references cleaned.

- No changes to the public HTTP API, monitor *execution types* (the ability to monitor external MongoDB instances remains a separate feature), or non-storage code.

## Capabilities

### New Capabilities
- `remove-mongodb-support`: Defines the removal of MongoDB as a supported application database backend (`DB_TYPE`). Going forward the system is strictly SQL-only (sqlite, postgres, mysql, mariadb). This capability will be captured in a new spec that declares the supported backends and removal of the mongo family.

### Modified Capabilities
(none — there are no existing specs in the top-level `openspec/specs/` that currently define database backend requirements. The new spec created by this change will establish the post-removal state.)

## Impact

**Primary affected areas**:
- `apps/server/internal/config/` — validators.go and config.go (remove strings and cases from validTypes, error messages, and ValidateDatabaseCustomRules).
- `apps/server/internal/infra/` — delete mongo.go entirely; remove mongo handling from sql.go graceful shutdown and any shared shutdown.
- `apps/server/internal/utils/` — dig.go (remove mongo branch in RegisterRepositoryByDBType); review bson.go and other utils.
- `apps/server/internal/modules/monitor/` — review and remove mongo-specific repository (monitor.mongo.repository.go and related) if they are storage-backend implementations rather than monitor-type executors.
- Root Docker artifacts — delete Dockerfile.bundle.mongo, supervisord.bundle.mongo.conf, startup.bundle.mongo.sh, docker-compose.bundle.mongo.yml, docker-compose.dev.mongo.yml, docker-compose.prod.mongo.yml, docker-compose.mongo.yml (and any test variants).
- `apps/docs/docs/self-hosting/` — delete docker-with-mongo.md.
- Documentation updates across `apps/docs/docs/architecture/*.md`, development.md, MIGRATION_SETUP.md, README.md (multiple sections), and sidebars.
- `Makefile` — remove all mongo-related COMPOSE_ variables, docker-*/down-* targets, help text, and down-all calls.
- `charts/values.yaml` and any Helm docs — remove or comment mongo examples.
- Minor: any CI, test_utils, or example scripts that reference mongo variants.

**Downstream**:
- Users currently running on MongoDB must migrate their data to one of the remaining SQL backends before upgrading (dump/restore + let migrations run). No automatic migration path is provided.
- The set of published bundle images and "official" compose examples shrinks from four DB families to three.
- Future contributors only maintain SQL + Bun paths (plus the three bundle variants).
- Reduced Docker image sizes/variants for the mongo bundle (no longer built/published).

**Out of scope**:
- Removing the "MongoDB" *monitor type* (the ability for Peekaping to monitor the health/uptime of external MongoDB servers as a target). That is independent of the application's own persistence layer.
- Changes to any other monitor implementations, notification channels, status pages, or the React frontend.
- Rewriting history or providing a data-migration CLI (users use standard mongodump + mysql/pg tools + the existing migrate container).
- Keeping deprecated/compatibility shims for mongo DB_TYPE (the removal is clean).