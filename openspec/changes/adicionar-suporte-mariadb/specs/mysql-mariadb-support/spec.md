# MySQL / MariaDB Database Support

This specification defines the requirements for first-class support of MySQL and MariaDB as database backends via `DB_TYPE=mysql` or `DB_TYPE=mariadb`. MariaDB MUST be treated as wire-compatible with MySQL for connection purposes.

## ADDED Requirements

### Requirement: DB_TYPE validation accepts mysql and mariadb

The configuration system SHALL accept "mysql" and "mariadb" (case-insensitive) as valid values for the `DB_TYPE` environment variable / config field, in addition to existing supported types.

#### Scenario: Valid mariadb DB_TYPE passes validation
- **WHEN** the application starts with `DB_TYPE=mariadb` (and required DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME provided)
- **THEN** config loading and custom validation succeed without "db_type" validation error.

#### Scenario: Valid mysql DB_TYPE passes validation
- **WHEN** the application starts with `DB_TYPE=mysql` (and required connection parameters)
- **THEN** config loading succeeds.

#### Scenario: Invalid db type is rejected
- **WHEN** `DB_TYPE=foo` is provided
- **THEN** validation fails with a clear message listing the allowed types including mysql and mariadb.

### Requirement: SQL provider connects using mysql dialect for mariadb

The infrastructure layer SHALL successfully establish a Bun *bun.DB connection for `DB_TYPE=mariadb` using the same MySQL driver and dialect as `mysql`.

#### Scenario: Application starts with external MariaDB
- **WHEN** `DB_TYPE=mariadb` and correct credentials for a reachable MariaDB server are supplied
- **THEN** ProvideSQLDB returns a working *bun.DB, connection pings successfully, and subsequent queries/migrations execute.

#### Scenario: Application starts with external MySQL
- **WHEN** `DB_TYPE=mysql` and correct credentials for a reachable MySQL server are supplied
- **THEN** the same connection success as above.

### Requirement: Database custom rules enforce credentials for mysql family

ValidateDatabaseCustomRules SHALL require DB_HOST, DB_PORT, DB_USER, DB_PASS (and numeric port) for both "mysql" and "mariadb", exactly as for "postgres".

#### Scenario: Missing credentials for mariadb rejected
- **WHEN** `DB_TYPE=mariadb` but DB_HOST is empty
- **THEN** ValidateDatabaseCustomRules returns an error mentioning DB_HOST is required for mariadb.

### Requirement: Graceful shutdown handles mysql/mariadb

GracefulDatabaseShutdown SHALL treat "mariadb" and "mysql" as SQL databases (delegating to the shared graceful shutdown path, which is a no-op for non-sqlite SQL dbs).

#### Scenario: Shutdown with mysql family
- **WHEN** the process shuts down with DB_TYPE set to mysql or mariadb
- **THEN** no error is raised from the mysql-specific shutdown path and the container exits cleanly.

### Requirement: Bundle (monolithic) deployment artifacts exist for MySQL

The project SHALL provide a complete set of artifacts allowing `docker run ... peekaping-bundle-mysql` (or equivalent tag) that embeds MariaDB/MySQL server, Redis, the Go services, web UI, and runs migrations on first start — exactly parallel to the postgres, sqlite and mongo bundles.

#### Scenario: Monolithic docker run for bundle-mysql succeeds
- **WHEN** a user follows documented `docker run` command for the mysql bundle image with appropriate -e DB_* vars and volume mount
- **THEN** the container starts, initializes the embedded DB (if first run), runs migrations, starts all services, and serves the UI on the expected port.

### Requirement: Docker Compose microservice files exist for MySQL/MariaDB

The project SHALL ship `docker-compose.mysql.yml` (and dev/prod variants) plus example `.env` guidance so users can run the api/web/worker etc. against an external or co-located MariaDB container using the official mariadb or mysql image.

#### Scenario: Compose up with external mysql database
- **WHEN** `docker-compose -f docker-compose.mysql.yml up` is executed with a valid .env pointing DB_TYPE=mysql or mariadb
- **THEN** the database service (if included) or external DB becomes healthy, migrate runs, and api becomes reachable.

### Requirement: Self-hosting documentation for MySQL/MariaDB

A new documentation page `docs/self-hosting/docker-with-mysql.md` SHALL exist with instructions for both monolithic bundle mode and microservice compose mode, including recommended environment variables, volume mounts, healthchecks, and notes about MariaDB vs MySQL image choice. The page SHALL be linked from README and the self-hosting index.

#### Scenario: User discovers and follows mysql self-hosting guide
- **WHEN** a user visits the docs site or README links for database options
- **THEN** they find a working "Docker + MySQL" guide with copy-pasteable commands that result in a running Peekaping instance backed by MariaDB.

### Requirement: All references updated for consistency

README, development docs, architecture component docs, MIGRATION_SETUP, and Makefile SHALL list mysql/mariadb as supported DB_TYPE options with the same prominence as postgres, sqlite, mongo. The incomplete "MySQL/MariaDB -" bullet in the features list SHALL be completed or removed.

#### Scenario: Documentation accurately reflects supported databases
- **WHEN** a reader scans README features, quick-start, or architecture tables
- **THEN** mysql and mariadb appear in DB_TYPE lists and no broken/incomplete items remain for this database family.

## Notes for Implementers

- Reuse the existing `_ "github.com/go-sql-driver/mysql"` import and `mysqldialect` — do not add a separate mariadb driver.
- In Go code, treat "mariadb" by falling through to the "mysql" connection logic (or duplicate the small case with a log message "Connecting via MySQL protocol (mariadb)").
- For the bundle Dockerfile, base it on a Debian/Ubuntu image that can install `mariadb-server` (or `mysql-server`); use the same multi-stage pattern as the postgres bundle.
- The startup script for the bundle must perform DB initialization, user/db creation, and wait logic equivalent to the postgres one (MariaDB uses `mysql` client and `mariadb-admin` or `mysqladmin` for readiness).
- Migrations are transactional-sql; note that some MySQL DDL is not transactional (documented already in migrate.md).
- Healthchecks in compose files for the DB service should use the appropriate mysqladmin ping or equivalent for the chosen image.
- No changes are required to the actual application queries or schema (Bun + migrations abstract the differences).
- After adding files, run `make` targets or the equivalent compose commands locally to verify the bundle and compose paths start cleanly.