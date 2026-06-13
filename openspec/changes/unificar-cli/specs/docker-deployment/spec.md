## ADDED Requirements

### Requirement: Dockerfiles build single peekaping binary
All bundle Dockerfiles (`Dockerfile.bundle.postgres`, `Dockerfile.bundle.mysql`, `Dockerfile.bundle.sqlite`) SHALL build a single `peekaping` binary from `./cmd/peekaping` instead of three separate binaries (`api`, `engine`, `bun`).

#### Scenario: Docker build produces one binary
- **WHEN** `docker build -f Dockerfile.bundle.mysql .` completes successfully
- **THEN** only `/app/server/peekaping` exists in the final image (no `/app/server/api`, `/app/server/engine`, or `/app/server/bun`)

#### Scenario: go-builder stage has single build command
- **WHEN** the Dockerfile go-builder stage is inspected
- **THEN** there is exactly one `go build` command targeting `./cmd/peekaping`

### Requirement: Supervisord configs use peekaping subcommands
All supervisord configuration files (`supervisord.bundle.postgres.conf`, `supervisord.bundle.mysql.conf`, `supervisord.bundle.sqlite.conf`) SHALL reference `peekaping api` and `peekaping engine` instead of `api` and `engine` binaries.

#### Scenario: Supervisord starts API via subcommand
- **WHEN** the bundle container starts
- **THEN** supervisord launches `/app/server/peekaping api` for the API service

#### Scenario: Supervisord starts engine via subcommand
- **WHEN** the bundle container starts
- **THEN** supervisord launches `/app/server/peekaping engine` for the engine service

### Requirement: Startup scripts use peekaping migrate for migrations
All startup and migration scripts (`startup.bundle.*.sh`, `scripts/run-migrations.sh`) SHALL use `peekaping migrate` subcommands instead of `bun db *` commands.

#### Scenario: Migrations run via peekaping migrate
- **WHEN** the bundle container starts and `run-migrations.sh` executes
- **THEN** the script calls `peekaping migrate init` followed by `peekaping migrate up`

#### Scenario: Migration script exits on failure
- **WHEN** `peekaping migrate up` fails (e.g., DB not reachable)
- **THEN** the script exits with a non-zero code and the container startup fails fast

### Requirement: Makefile targets use peekaping binary
The Makefile SHALL provide updated targets that build and run the unified binary, replacing the separate `build-api`, `build-engine`, `build-bun` targets.

#### Scenario: Single build target compiles peekaping
- **WHEN** the user runs `make build`
- **THEN** a single `peekaping` binary is produced in `apps/server/`

#### Scenario: Subcommand run targets invoke peekaping
- **WHEN** the user runs `make run-api`
- **THEN** `./peekaping api` is executed with the appropriate environment

#### Scenario: Migrate target runs migrations locally
- **WHEN** the user runs `make migrate-up`
- **THEN** `./peekaping migrate up` is executed against the local database
