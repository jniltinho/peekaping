# Remove MongoDB Backend Support

This specification captures the requirements after the removal of MongoDB as an application database backend. The system SHALL support only SQL-based databases going forward.

## REMOVED Requirements

### Requirement: MongoDB as a valid application database backend

The system SHALL NOT accept `DB_TYPE=mongo` or `DB_TYPE=mongodb` (case-insensitive). These values are no longer valid database types for the application's own persistence layer.

#### Scenario: Startup with mongo DB_TYPE is rejected
- **WHEN** the application is started with `DB_TYPE=mongo` (or `mongodb`) and otherwise valid connection parameters
- **THEN** configuration validation fails with an error listing only the supported SQL types (postgres, postgresql, mysql, mariadb, sqlite).

#### Scenario: Graceful shutdown and connection code no longer reference mongo
- **WHEN** the application shuts down or initializes a database connection
- **THEN** no code paths for "mongo" or "mongodb" are executed; the mongo client is never instantiated.

### Requirement: MongoDB-specific repository implementations and DI registration

The DI container registration and per-module repository constructors SHALL no longer provide or select MongoDB-backed implementations. All data access SHALL use the SQL (Bun) repositories regardless of configuration.

#### Scenario: Repository registration for any module
- **WHEN** `RegisterRepositoryByDBType` (or equivalent) is called during container setup
- **THEN** only the SQL repository constructor is provided; the mongo constructor argument is unused and can be removed.

### Requirement: MongoDB connection provider and graceful shutdown handling

`ProvideMongoDB` and any mongo-specific shutdown logic SHALL be removed. The graceful database shutdown path SHALL only handle SQL database types.

#### Scenario: MongoDB client is never created
- **WHEN** the application boots with any supported DB_TYPE
- **THEN** no MongoDB client is created or connected; only `*bun.DB` (or equivalent SQL client) is provided for SQL types.

### Requirement: MongoDB bundle and compose deployment options

The project SHALL NOT provide `peekaping-bundle-mongo` image, `Dockerfile.bundle.mongo`, `startup.bundle.mongo.sh`, `supervisord.bundle.mongo.conf`, or any `docker-compose.*.mongo.yml` files. Users SHALL NOT be able to run a monolithic or microservice MongoDB-backed deployment using the official artifacts.

#### Scenario: Attempt to build or run mongo bundle
- **WHEN** a user tries to build or start using the (now-deleted) mongo bundle Dockerfile or compose files
- **THEN** the files do not exist; documentation directs users to one of the three supported SQL backends.

### Requirement: MongoDB self-hosting documentation and references

A dedicated `docker-with-mongo.md` page SHALL NOT exist. All references to MongoDB as a storage backend SHALL be removed from README, development guides, architecture docs, migration setup, Makefile help text, and Helm examples.

#### Scenario: User looks for MongoDB deployment instructions
- **WHEN** a user searches docs or README for "mongo" or "mongodb" in the context of running Peekaping itself
- **THEN** they find only references to using MongoDB as a *monitor target* (separate feature) or historical notes; storage options are limited to SQL databases with links to the three remaining guides.

## ADDED Requirements

### Requirement: SQL-only database backends are the supported set

The system SHALL officially support only the following values for `DB_TYPE`: `postgres`, `postgresql`, `mysql`, `mariadb`, `sqlite`.

#### Scenario: Documentation and validation accurately list supported types
- **WHEN** a user reads architecture tables, error messages, or help output
- **THEN** the listed types are exactly the SQL family (with mariadb as alias for mysql protocol) and mongo/mongodb are absent.

## Notes for Implementers

- Delete the entire `apps/server/internal/infra/mongo.go` (and its imports/usages).
- Simplify or remove `RegisterRepositoryByDBType` (or inline the SQL-only path); delete all `* .mongo.repository.go` files under modules (they implement the mongo side of the previous dual-backend strategy).
- Remove the `go.mongodb.org/mongo-driver` dependency from go.mod (after confirming no other usage remains — monitor types that talk *to* external Mongo instances may still need a client, but that should be isolated in the monitor executor, not the app DB layer).
- Delete all root-level mongo Docker and compose files listed in the proposal Impact section.
- Delete `apps/docs/docs/self-hosting/docker-with-mongo.md`.
- Perform global search/replace cleanups for DB_TYPE lists, help text, badges, quick-start paragraphs, and example .env snippets.
- The bson utils and any mongoModel types in module repos can be deleted if they were only for the backend.
- Default DB in Makefile examples can be switched to postgres or sqlite.
- Update MIGRATION_SETUP.md to remove the "MongoDB is schema-less" note or mark it historical.
- This is a breaking change — the spec and proposal require clear **Reason** + **Migration** guidance for users (export from mongo → import to chosen SQL DB → run migrations).
- After deletion, run `go mod tidy` and full build + any available tests.
- No changes are required for monitor *execution* code that happens to support "mongodb" as a target type for user-defined monitors.