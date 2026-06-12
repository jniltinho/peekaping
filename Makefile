# Makefile for Peekaping project

# Variables
GO_SERVER_DIR = apps/server
WEB_DIR = apps/web
BINARY_NAME = peekaping-server

# Docker Compose bundle configurations
COMPOSE_POSTGRES = docker-compose.bundle.postgres.yml
COMPOSE_SQLITE   = docker-compose.bundle.sqlite.yml
COMPOSE_MYSQL    = docker-compose.bundle.mysql.yml

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'


# ── Docker build ──────────────────────────────────────────────────────────────

.PHONY: build-bundle-sqlite
build-bundle-sqlite: ## Build bundle SQLite image (peekaping-bundle-sqlite:dev)
	docker build --no-cache -f Dockerfile.bundle.sqlite -t peekaping-bundle-sqlite:dev .

.PHONY: build-bundle-postgres
build-bundle-postgres: ## Build bundle PostgreSQL image (peekaping-bundle-postgres:dev)
	docker build --no-cache -f Dockerfile.bundle.postgres -t peekaping-bundle-postgres:dev .

.PHONY: build-bundle-mysql
build-bundle-mysql: ## Build bundle MySQL image (peekaping-bundle-mysql:dev)
	docker build --no-cache -f Dockerfile.bundle.mysql -t peekaping-bundle-mysql:dev .


# ── Docker ────────────────────────────────────────────────────────────────────

.PHONY: docker-postgres
docker-postgres: ## Start bundle with PostgreSQL
	docker-compose -f $(COMPOSE_POSTGRES) up -d --build

.PHONY: docker-sqlite
docker-sqlite: ## Start bundle with SQLite
	docker-compose -f $(COMPOSE_SQLITE) up -d --build

.PHONY: docker-mysql
docker-mysql: ## Start bundle with MySQL/MariaDB
	docker-compose -f $(COMPOSE_MYSQL) up -d --build

.PHONY: down-postgres
down-postgres: ## Stop PostgreSQL bundle
	docker-compose -f $(COMPOSE_POSTGRES) down

.PHONY: down-sqlite
down-sqlite: ## Stop SQLite bundle
	docker-compose -f $(COMPOSE_SQLITE) down

.PHONY: down-mysql
down-mysql: ## Stop MySQL/MariaDB bundle
	docker-compose -f $(COMPOSE_MYSQL) down

.PHONY: docker-down-all
docker-down-all: ## Stop all bundle services
	@docker-compose -f $(COMPOSE_POSTGRES) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_SQLITE) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_MYSQL) down 2>/dev/null || true


# ── Database migrations ───────────────────────────────────────────────────────

.PHONY: migrate-init
migrate-init: ## Init database migrations
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db init

.PHONY: migrate-up
migrate-up: ## Run database migrations
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db migrate

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db rollback


# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build-api
build-api: ## Build the API binary
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/api ./cmd/api

.PHONY: build-engine
build-engine: ## Build the engine binary (scheduler + worker pool + ingester)
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/engine ./cmd/engine

.PHONY: build-bun
build-bun: ## Build the bun (migration) binary
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/bun ./cmd/bun

.PHONY: embed-web
embed-web: ## Build frontend and embed into Go API binary
	pnpm --filter=web run build
	rm -rf apps/server/internal/frontend/web/*
	cp -r apps/web/dist/. apps/server/internal/frontend/web/


# ── Run ───────────────────────────────────────────────────────────────────────

.PHONY: run-api
run-api: ## Run the API server
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go run ./cmd/api/main.go

.PHONY: run-engine
run-engine: ## Run the engine service
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go run ./cmd/engine/main.go

.PHONY: dev
dev: ## Start full development environment
	./scripts/tool.sh pnpm run dev dev:api dev:engine docs:watch


# ── Test & lint ───────────────────────────────────────────────────────────────

.PHONY: test-server
test-server: ## Run server tests
	cd apps/server && ../../scripts/tool.sh go test -v ./internal/...

.PHONY: lint-web
lint-web: ## Lint and build the web app
	cd apps/web && ../../scripts/tool.sh pnpm lint && ../../scripts/tool.sh pnpm build


# ── Setup ─────────────────────────────────────────────────────────────────────

.PHONY: setup
setup: ## Setup development environment via asdf
	@if command -v asdf >/dev/null 2>&1; then \
		asdf plugin add golang || true; \
		asdf plugin add nodejs || true; \
		asdf plugin add pnpm || true; \
		asdf install; \
	else \
		echo "asdf not found — install Go 1.26, Node.js 22, pnpm 9 manually."; \
	fi

.PHONY: install
install: ## Install all dependencies
	./scripts/tool.sh pnpm install
	cd apps/server && ../../scripts/tool.sh go mod tidy
