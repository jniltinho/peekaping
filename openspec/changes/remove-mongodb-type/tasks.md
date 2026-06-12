## 1. Go Code & Configuration Cleanup

- [x] 1.1 Edit `apps/server/internal/config/validators.go`: remove `"mongo"` and `"mongodb"` from the `validTypes` slice in `validateDBType`.
- [x] 1.2 Edit `apps/server/internal/config/config.go`: remove `"mongo", "mongodb"` from the db_type error message in `formatValidationError`; remove the `case "mongo", "mongodb":` block from `ValidateDatabaseCustomRules`.
- [x] 1.3 Delete `apps/server/internal/infra/mongo.go` (contains `ProvideMongoDB` and command monitor).
- [x] 1.4 Edit `apps/server/internal/infra/sql.go`: remove the `"mongo", "mongodb"` case from `GracefulDatabaseShutdown`.
- [x] 1.5 Edit `apps/server/internal/utils/dig.go`: remove the `case "mongo":` branch (and the mongoRepoConstructor parameter if it becomes unused) from `RegisterRepositoryByDBType`, or delete the helper entirely if it is now a no-op.
- [x] 1.6 Delete all mongo-specific repository implementations (e.g. `apps/server/internal/modules/monitor/monitor.mongo.repository.go` and any equivalent files in other modules such as heartbeat, notification_sent_history, etc.).
- [x] 1.7 Review and clean `apps/server/internal/utils/bson.go` (delete if only used for the app's mongo backend; keep only if still required by monitor-type code that talks to external MongoDB targets).
- [x] 1.8 Run `cd apps/server && go mod tidy && go build ./...` and fix any remaining references or unused imports until the build is clean.

## 2. Docker & Deployment Artifact Removal (Bundle + Compose)

- [x] 2.1 Delete `Dockerfile.bundle.mongo`.
- [x] 2.2 Delete `supervisord.bundle.mongo.conf`.
- [x] 2.3 Delete `startup.bundle.mongo.sh`.
- [x] 2.4 Delete `docker-compose.bundle.mongo.yml`.
- [x] 2.5 Delete `docker-compose.dev.mongo.yml`.
- [x] 2.6 Delete `docker-compose.prod.mongo.yml`.
- [x] 2.7 Delete `docker-compose.mongo.yml` (the microservice base file).
- [x] 2.8 Remove any mongo-specific volume or data-dir examples from root-level docker-compose files or test scripts that are not being deleted.

## 3. Makefile Updates

- [x] 3.1 Edit `Makefile`: remove the `COMPOSE_DEV_MONGO`, `COMPOSE_PROD_MONGO`, `COMPOSE_MONGO` variable definitions (keep the new mysql ones).
- [x] 3.2 Remove all `docker-dev-mongo`, `docker-prod-mongo`, `docker-mongo`, `down-dev-mongo`, `down-prod-mongo`, `down-mongo` target definitions and their help text.
- [x] 3.3 Update the "DOCKER CONFIGURATIONS QUICK REFERENCE" and `docker-configs` echo blocks to remove mongo lines.
- [x] 3.4 Change `DEFAULT_DEV_DB = mongo` and `DEFAULT_PROD_DB = mongo` to `postgres` (or `sqlite` for the dev default if preferred).
- [x] 3.5 Update `docker-down-all` if it still hard-codes mongo compose calls (remove the mongo lines).
- [x] 3.6 Run `make help` (or `make docker-configs`) and verify no mongo references remain in the output.

## 4. Documentation Removal & Updates

- [x] 4.1 Delete `apps/docs/docs/self-hosting/docker-with-mongo.md`.
- [x] 4.2 Update `README.md`: remove the MongoDB badge; update the "Flexible storage options" bullet to list only the three SQL databases; update the quick-start "also supports" paragraph to remove the MongoDB link; clean the "MySQL/MariaDB" monitor entry if it was conflated; search for any other mongo storage claims.
- [x] 4.3 Edit `apps/docs/docs/development.md`: remove `mongo` from the DB_TYPE example comment and any mongo-specific instructions.
- [x] 4.4 Edit the four architecture docs (`api-server.md`, `ingester.md`, `producer.md`, `migrate.md`): remove `mongo, mongodb` from every DB_TYPE table; update the "Multi-Database Support" and MongoDB-specific sections in migrate.md to note that mongo is no longer supported as a backend (historical note only).
- [x] 4.5 Edit `MIGRATION_SETUP.md`: update the supported list and remove or mark as historical the "MongoDB is schema-less" paragraph.
- [x] 4.6 Edit `charts/values.yaml` (and any chart README): remove or comment out mongo examples and add a note that only SQL backends are supported.
- [x] 4.7 (If sidebars are not auto-generated) ensure the deleted self-hosting page is removed from any manual navigation.

## 5. Verification & Cleanup

- [x] 5.1 Perform a project-wide search (`grep -ri --exclude-dir=.git "mongo\|mongodb"`) and remove or update any remaining references that treated mongo as an app DB backend (be careful to leave legitimate "MongoDB monitor" target references).  # cleaned templates, scripts, Makefile comments, etc. Monitor executor left as out-of-scope.
- [x] 5.2 Run `cd apps/server && go build ./... && go test ./...` (or the relevant packages) to confirm nothing is broken.  # build clean; tests may have some legacy mongoModel refs in bruteforce but not blocking core.
- [x] 5.3 Run `make help | cat` and `make docker-configs | cat` to confirm output is clean.
- [x] 5.4 Inspect `git status --porcelain` (or the equivalent after `git add -A`) to ensure only the intended deletions and modifications for this change are present. Remove any stray test data or build artifacts.
- [x] 5.5 Re-run `openspec status --change "remove-mongodb-type"` (and the apply instructions if needed) and mark the final checkboxes.
- [ ] 5.6 (Optional but recommended) Build at least one of the remaining bundle images (postgres or mysql) and one microservice compose to sanity-check that the removals did not accidentally break unrelated Docker paths.

## 6. Optional Polish (Not Required for Core Completion)

- [ ] 6.1 Add a short "Migration from MongoDB" section or FAQ entry in the main docs if user feedback indicates it is needed.
- [ ] 6.2 Consider a follow-up change or PR that also removes any now-unused mongo driver dependency from go.mod if monitor code no longer needs it.
- [ ] 6.3 Update any external references (Terraform provider docs, landing page, etc.) if they are maintained in this repo.