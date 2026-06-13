## ADDED Requirements

### Requirement: Single binary entrypoint
The system SHALL provide a single compiled binary named `peekaping` that exposes all server and migration functionality via Cobra subcommands. The binary SHALL be the sole Go entry point, replacing the separate `api`, `engine`, and `bun` binaries.

#### Scenario: Binary shows help with subcommands
- **WHEN** the user runs `peekaping --help`
- **THEN** the output lists three subcommands: `api`, `engine`, and `migrate`, with their descriptions

#### Scenario: Unknown subcommand
- **WHEN** the user runs `peekaping unknown`
- **THEN** the binary exits with a non-zero code and prints a usage error

### Requirement: API subcommand starts HTTP server
The system SHALL provide a `peekaping api` subcommand that starts the Echo v5 REST API server with the same behavior as the previous `api` binary.

#### Scenario: API server starts with default config
- **WHEN** the user runs `peekaping api` with required env vars set
- **THEN** the HTTP server starts on the configured port and responds to requests

#### Scenario: API server reads config from env vars
- **WHEN** `SERVER_PORT=9090` is set in the environment
- **THEN** `peekaping api` binds to port 9090

#### Scenario: API server reads config from .env file
- **WHEN** a `.env` file is present in the working directory with `SERVER_PORT=9090`
- **THEN** `peekaping api` binds to port 9090

### Requirement: Engine subcommand starts scheduler/worker
The system SHALL provide a `peekaping engine` subcommand that starts the unified scheduler, worker pool, and ingester with the same behavior as the previous `engine` binary.

#### Scenario: Engine starts with required config
- **WHEN** the user runs `peekaping engine` with DB and Redis env vars set
- **THEN** the engine starts processing monitor jobs

#### Scenario: Engine respects ENGINE_WORKERS env var
- **WHEN** `ENGINE_WORKERS=5` is set
- **THEN** `peekaping engine` starts 5 worker goroutines

### Requirement: Migrate subcommand wraps bun migration CLI
The system SHALL provide a `peekaping migrate` subcommand with sub-subcommands that replicate the full `bun db *` migration functionality, using the same underlying bun/migrate library.

#### Scenario: Migrate up runs pending migrations
- **WHEN** the user runs `peekaping migrate up`
- **THEN** all pending database migrations are applied and success is printed

#### Scenario: Migrate rollback reverts last group
- **WHEN** the user runs `peekaping migrate rollback`
- **THEN** the last migration group is rolled back

#### Scenario: Migrate init creates migration tables
- **WHEN** the user runs `peekaping migrate init`
- **THEN** the bun migration tracking tables are created in the database

#### Scenario: Migrate status shows migration state
- **WHEN** the user runs `peekaping migrate status`
- **THEN** the list of applied and unapplied migrations is printed

#### Scenario: Migrate create-go creates a Go migration file
- **WHEN** the user runs `peekaping migrate create-go <name>`
- **THEN** a new Go migration file is created in the migrations directory

### Requirement: Viper-based configuration loading
The system SHALL use Viper to load configuration from environment variables and a `.env` file. The root command SHALL accept an `--env-file` flag (default: `.env`) specifying the path to the env file. All subcommands SHALL inherit this flag.

#### Scenario: Config loaded from --env-file flag
- **WHEN** the user runs `peekaping --env-file /etc/peekaping.env api`
- **THEN** configuration is loaded from `/etc/peekaping.env`

#### Scenario: Env vars override .env file values
- **WHEN** `.env` has `SERVER_PORT=8034` but `SERVER_PORT=9000` is exported in the shell
- **THEN** `peekaping api` uses port 9000

### Requirement: Unified config struct with per-subcommand validation
The system SHALL use a single config struct that covers all subcommand settings. Validation SHALL only enforce required fields relevant to the active subcommand.

#### Scenario: API subcommand validates SERVER_PORT
- **WHEN** `peekaping api` is started without `SERVER_PORT` and no default is present
- **THEN** the binary exits with a config validation error mentioning `SERVER_PORT`

#### Scenario: Engine subcommand does not require SERVER_PORT
- **WHEN** `peekaping engine` is started without `SERVER_PORT`
- **THEN** the engine starts successfully (SERVER_PORT is only required by `api`)

### Requirement: Remove urfave/cli dependency
The system SHALL NOT depend on `github.com/urfave/cli/v2`. The migration CLI functionality SHALL be re-implemented using Cobra sub-subcommands.

#### Scenario: go.mod does not contain urfave/cli
- **WHEN** `go mod tidy` is run after the implementation
- **THEN** `github.com/urfave/cli/v2` does not appear in `go.mod` or `go.sum`
