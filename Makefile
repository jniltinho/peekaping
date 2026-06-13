# Makefile for Monitoring project

GO_SERVER_DIR := apps/server

VERSION    := $(or $(shell git describe --tags --abbrev=0 2>/dev/null),dev)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BUILD_ARGS := \
	--build-arg VERSION=$(VERSION) \
	--build-arg BUILD_TIME=$(BUILD_TIME) \
	--build-arg GIT_COMMIT=$(GIT_COMMIT)

COMPOSE_SQLITE   := docker-compose.bundle.sqlite.yml
COMPOSE_POSTGRES := docker-compose.bundle.postgres.yml
COMPOSE_MYSQL    := docker-compose.bundle.mysql.yml

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-26s\033[0m %s\n", $$1, $$2}'


# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build the monitoring binary (UPX if available)
	$(MAKE) -C $(GO_SERVER_DIR) build

.PHONY: embed-web
embed-web: ## Build frontend and embed into Go binary
	pnpm --filter=web build
	rm -rf $(GO_SERVER_DIR)/internal/frontend/web/*
	cp -r apps/web/dist/. $(GO_SERVER_DIR)/internal/frontend/web/

.PHONY: build-all
build-all: embed-web build ## Embed frontend then build binary


# ── Run (development) ─────────────────────────────────────────────────────────

.PHONY: dev
dev: ## Start the web frontend dev server
	pnpm --filter=web dev

.PHONY: run-api
run-api: ## Run the API server (go run)
	cd $(GO_SERVER_DIR) && go run . api

.PHONY: run-engine
run-engine: ## Run the engine (go run)
	cd $(GO_SERVER_DIR) && go run . engine


# ── Migrations ────────────────────────────────────────────────────────────────

.PHONY: migrate-up
migrate-up: ## Apply database schema (GORM AutoMigrate)
	cd $(GO_SERVER_DIR) && go run . migrate up

.PHONY: migrate-status
migrate-status: ## Check database connectivity
	cd $(GO_SERVER_DIR) && go run . migrate status


# ── Bundle Docker images (all-in-one) ─────────────────────────────────────────

.PHONY: build-bundle-sqlite
build-bundle-sqlite: ## Build bundle SQLite image
	docker build -f Dockerfile.bundle.sqlite $(BUILD_ARGS) -t monitoring-bundle-sqlite:$(VERSION) .

.PHONY: build-bundle-postgres
build-bundle-postgres: ## Build bundle PostgreSQL image
	docker build -f Dockerfile.bundle.postgres $(BUILD_ARGS) -t monitoring-bundle-postgres:$(VERSION) .

.PHONY: build-bundle-mysql
build-bundle-mysql: ## Build bundle MySQL image
	docker build -f Dockerfile.bundle.mysql $(BUILD_ARGS) -t monitoring-bundle-mysql:$(VERSION) .

.PHONY: build-bundle-all
build-bundle-all: build-bundle-sqlite build-bundle-postgres build-bundle-mysql ## Build all bundle images


# ── Infra Docker images (individual services) ─────────────────────────────────

.PHONY: build-infra-api
build-infra-api: ## Build infra API service image
	docker build -f $(GO_SERVER_DIR)/infra/Dockerfile.api $(BUILD_ARGS) \
		-t monitoring-api:$(VERSION) $(GO_SERVER_DIR)

.PHONY: build-infra-engine
build-infra-engine: ## Build infra engine service image
	docker build -f $(GO_SERVER_DIR)/infra/Dockerfile.engine $(BUILD_ARGS) \
		-t monitoring-engine:$(VERSION) $(GO_SERVER_DIR)

.PHONY: build-infra-migrate
build-infra-migrate: ## Build infra migrate service image
	docker build -f $(GO_SERVER_DIR)/infra/Dockerfile.migrate $(BUILD_ARGS) \
		-t monitoring-migrate:$(VERSION) $(GO_SERVER_DIR)

.PHONY: build-infra-all
build-infra-all: build-infra-api build-infra-engine build-infra-migrate ## Build all infra service images


# ── Run bundle containers ─────────────────────────────────────────────────────

.PHONY: run-bundle-sqlite
run-bundle-sqlite: ## Run monitoring-bundle-sqlite container
	docker run --rm --name monitoring -p 8383:8383 \
		-v monitoring-data:/app/data \
		-e TZ=America/Sao_Paulo \
		-e DB_NAME=/app/data/monitoring.db \
		-e ENGINE_WORKERS=4 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		monitoring-bundle-sqlite:$(VERSION)

.PHONY: run-bundle-postgres
run-bundle-postgres: ## Run monitoring-bundle-postgres container
	docker run --rm --name monitoring -p 8383:8383 \
		-v monitoring-postgres-data:/var/lib/postgresql/data \
		-e TZ=America/Sao_Paulo \
		-e DB_TYPE=postgres \
		-e DB_HOST=localhost \
		-e DB_PORT=5432 \
		-e DB_NAME=monitoring \
		-e DB_USER=monitoring \
		-e DB_PASS=password \
		-e ENGINE_WORKERS=4 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		monitoring-bundle-postgres:$(VERSION)

.PHONY: run-bundle-mysql
run-bundle-mysql: ## Run monitoring-bundle-mysql container
	docker run --rm --name monitoring -p 8383:8383 \
		-v monitoring-mysql-data:/var/lib/mysql \
		-e TZ=America/Sao_Paulo \
		-e DB_TYPE=mysql \
		-e DB_HOST=localhost \
		-e DB_PORT=3306 \
		-e DB_NAME=monitoring \
		-e DB_USER=monitoring \
		-e DB_PASS=password \
		-e ENGINE_WORKERS=4 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		monitoring-bundle-mysql:$(VERSION)


# ── Docker Compose ────────────────────────────────────────────────────────────

.PHONY: docker-up-sqlite
docker-up-sqlite: ## Start SQLite bundle via docker compose
	docker compose -f $(COMPOSE_SQLITE) up -d --build

.PHONY: docker-up-postgres
docker-up-postgres: ## Start PostgreSQL bundle via docker compose
	docker compose -f $(COMPOSE_POSTGRES) up -d --build

.PHONY: docker-up-mysql
docker-up-mysql: ## Start MySQL bundle via docker compose
	docker compose -f $(COMPOSE_MYSQL) up -d --build

.PHONY: docker-down-sqlite
docker-down-sqlite: ## Stop SQLite bundle
	docker compose -f $(COMPOSE_SQLITE) down

.PHONY: docker-down-postgres
docker-down-postgres: ## Stop PostgreSQL bundle
	docker compose -f $(COMPOSE_POSTGRES) down

.PHONY: docker-down-mysql
docker-down-mysql: ## Stop MySQL bundle
	docker compose -f $(COMPOSE_MYSQL) down

.PHONY: docker-down-all
docker-down-all: ## Stop all bundles
	@docker compose -f $(COMPOSE_SQLITE) down 2>/dev/null || true
	@docker compose -f $(COMPOSE_POSTGRES) down 2>/dev/null || true
	@docker compose -f $(COMPOSE_MYSQL) down 2>/dev/null || true


# ── Test & lint ───────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run server tests
	cd $(GO_SERVER_DIR) && go test -v ./internal/...

.PHONY: lint
lint: ## Lint the web app
	pnpm --filter=web lint


# ── Setup & install ───────────────────────────────────────────────────────────

.PHONY: install
install: ## Install all dependencies
	pnpm install
	cd $(GO_SERVER_DIR) && go mod tidy

.PHONY: setup
setup: ## Setup dev environment via asdf
	@if command -v asdf >/dev/null 2>&1; then \
		asdf plugin add golang || true; \
		asdf plugin add nodejs || true; \
		asdf plugin add pnpm || true; \
		asdf install; \
	else \
		echo "asdf not found — install Go 1.26, Node.js 22, pnpm 9 manually."; \
	fi
