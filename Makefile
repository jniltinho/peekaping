# Makefile for Peekaping project

# Variables
GO_SERVER_DIR = apps/server
WEB_DIR = apps/web
BINARY_NAME = peekaping-server

VERSION    := $(or $(shell git describe --tags --abbrev=0 2>/dev/null),dev)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

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
build-bundle-sqlite: ## Build bundle SQLite image
	docker build -f Dockerfile.bundle.sqlite \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t peekaping-bundle-sqlite:$(VERSION) .

.PHONY: build-bundle-postgres
build-bundle-postgres: ## Build bundle PostgreSQL image
	docker build -f Dockerfile.bundle.postgres \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t peekaping-bundle-postgres:$(VERSION) .

.PHONY: build-bundle-mysql
build-bundle-mysql: ## Build bundle MySQL image
	docker build -f Dockerfile.bundle.mysql \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t peekaping-bundle-mysql:$(VERSION) .


# ── Run bundle ────────────────────────────────────────────────────────────────

.PHONY: run-bundle-sqlite
run-bundle-sqlite: ## Run peekaping-bundle-sqlite:$(VERSION) container
	docker run --rm --name peekaping -p 8383:8383 \
		-v peekaping-data:/app/data \
		-e TZ=America/Sao_Paulo \
		-e MODE=dev \
		-e ENGINE_WORKERS=10 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		-e ENGINE_USE_REDIS=true \
		peekaping-bundle-sqlite:$(VERSION)

.PHONY: run-bundle-postgres
run-bundle-postgres: ## Run peekaping-bundle-postgres:$(VERSION) container
	docker run --rm --name peekaping -p 8383:8383 \
		-v peekaping-postgres-data:/var/lib/postgresql/data \
		-e TZ=America/Sao_Paulo \
		-e MODE=dev \
		-e DB_TYPE=postgres \
		-e DB_HOST=localhost \
		-e DB_PORT=5432 \
		-e DB_NAME=peekaping \
		-e DB_USER=peekaping \
		-e DB_PASS=password \
		-e ENGINE_WORKERS=10 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		-e ENGINE_USE_REDIS=true \
		peekaping-bundle-postgres:$(VERSION)

.PHONY: run-bundle-mysql
run-bundle-mysql: ## Run peekaping-bundle-mysql:$(VERSION) container
	docker run --rm --name peekaping -p 8383:8383 \
		-v peekaping-mysql-data:/var/lib/mysql \
		-e TZ=America/Sao_Paulo \
		-e MODE=dev \
		-e DB_TYPE=mysql \
		-e DB_HOST=localhost \
		-e DB_PORT=3306 \
		-e DB_NAME=peekaping \
		-e DB_USER=peekaping \
		-e DB_PASS=password \
		-e ENGINE_WORKERS=10 \
		-e ENGINE_SCHEDULER_INTERVAL=10s \
		-e ENGINE_USE_REDIS=true \
		peekaping-bundle-mysql:$(VERSION)


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

.PHONY: migrate-up
migrate-up: ## Apply / update database schema (GORM AutoMigrate)
	cd apps/server && go run . migrate up

.PHONY: migrate-status
migrate-status: ## Check database connectivity
	cd apps/server && go run . migrate status


# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build the unified peekaping binary
	cd $(GO_SERVER_DIR) && go build -o ../../bin/monitoring .

.PHONY: build-api
build-api: build ## Build the unified peekaping binary (alias)

.PHONY: build-engine
build-engine: build ## Build the unified peekaping binary (alias)

.PHONY: build-bun
build-bun: build ## Build the unified peekaping binary (alias)

.PHONY: embed-web
embed-web: ## Build frontend and embed into Go API binary
	pnpm --filter=web run build
	rm -rf apps/server/internal/frontend/web/*
	cp -r apps/web/dist/. apps/server/internal/frontend/web/


# ── Run ───────────────────────────────────────────────────────────────────────

.PHONY: run-api
run-api: ## Run the API server
	cd $(GO_SERVER_DIR) && go run . api

.PHONY: run-engine
run-engine: ## Run the engine service
	cd $(GO_SERVER_DIR) && go run . engine

.PHONY: dev
dev: ## Start full development environment
	pnpm run dev dev:api dev:engine docs:watch


# ── Test & lint ───────────────────────────────────────────────────────────────

.PHONY: test-server
test-server: ## Run server tests
	cd apps/server && go test -v ./internal/...

.PHONY: lint-web
lint-web: ## Lint and build the web app
	cd apps/web && pnpm lint && pnpm build


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
	pnpm install
	cd apps/server && go mod tidy
