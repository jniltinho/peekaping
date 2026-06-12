# Makefile for Peekaping project
#
# This Makefile supports multiple Docker Compose configurations:
# - Development: dev-postgres, dev-sqlite, dev-mysql
# - Production:  prod-postgres, prod-sqlite, prod-mysql
# - Standalone:  postgres, mysql
#
# To change the default database, modify DEFAULT_DEV_DB or DEFAULT_PROD_DB below
# Example: make dev (uses default) vs make dev-mysql (specific database)

# Variables
GO_SERVER_DIR = apps/server
WEB_DIR = apps/web
BINARY_NAME = peekaping-server

# Docker Compose configurations
COMPOSE_DEV_POSTGRES = docker-compose.dev.postgres.yml
COMPOSE_DEV_SQLITE = docker-compose.dev.sqlite.yml
COMPOSE_DEV_MYSQL = docker-compose.dev.mysql.yml
COMPOSE_PROD_POSTGRES = docker-compose.prod.postgres.yml
COMPOSE_PROD_SQLITE = docker-compose.prod.sqlite.yml
COMPOSE_PROD_MYSQL = docker-compose.prod.mysql.yml
COMPOSE_POSTGRES = docker-compose.postgres.yml
COMPOSE_SQLITE = docker-compose.sqlite.yml
COMPOSE_MYSQL = docker-compose.mysql.yml

# Default configurations
DEFAULT_DEV_DB = postgres
DEFAULT_PROD_DB = postgres

# Default target
.DEFAULT_GOAL := help

# Help target - shows available commands
.PHONY: help
help: ## Show this help message
	@echo "🐳 DOCKER CONFIGURATIONS QUICK REFERENCE:"
	@echo "  \033[32mDevelopment:\033[0m   dev-postgres, dev-sqlite, dev-mysql"
	@echo "  \033[33mProduction:\033[0m    prod-postgres, prod-sqlite, prod-mysql"
	@echo "  \033[36mStandalone:\033[0m    postgres, mysql"
	@echo "  \033[35mSwitchers:\033[0m     switch-to-postgres, switch-to-sqlite"
	@echo "  \033[31mStop All:\033[0m      docker-down-all"
	@echo ""
	@echo "📋 Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: docker-configs
docker-configs: ## Show all available Docker Compose configurations
	@echo "📋 Available Docker Compose Configurations:"
	@echo ""
	@echo "🔧 \033[32mDEVELOPMENT ENVIRONMENTS:\033[0m"
	@echo "  make dev-postgres    # $(COMPOSE_DEV_POSTGRES)"
	@echo "  make dev-sqlite      # $(COMPOSE_DEV_SQLITE)"
	@echo "  make dev-mysql       # $(COMPOSE_DEV_MYSQL)"
	@echo ""
	@echo "🚀 \033[33mPRODUCTION ENVIRONMENTS:\033[0m"
	@echo "  make prod-postgres   # $(COMPOSE_PROD_POSTGRES)"
	@echo "  make prod-sqlite     # $(COMPOSE_PROD_SQLITE)"
	@echo "  make prod-mysql      # $(COMPOSE_PROD_MYSQL)"
	@echo ""
	@echo "🎯 \033[36mSTANDALONE ENVIRONMENTS:\033[0m"
	@echo "  make postgres        # $(COMPOSE_POSTGRES)"
	@echo "  make mysql           # $(COMPOSE_MYSQL)"
	@echo ""
	@echo "⚡ \033[35mQUICK SWITCHERS:\033[0m"
	@echo "  make switch-to-postgres   # Stop all → Start PostgreSQL dev"
	@echo "  make switch-to-sqlite     # Stop all → Start SQLite dev"
	@echo ""
	@echo "🔍 \033[34mUTILITY COMMANDS:\033[0m"
	@echo "  make docker-status        # Show status of all configurations"
	@echo "  make docker-ps            # Show running containers"
	@echo "  make docker-down-all      # Stop all configurations"


# Docker targets - Development Environment
.PHONY: docker-dev-postgres
docker-dev-postgres: ## Start development environment with PostgreSQL
	@echo "Starting development environment with PostgreSQL..."
	docker-compose -f $(COMPOSE_DEV_POSTGRES) up -d --build

.PHONY: docker-dev-sqlite
docker-dev-sqlite: ## Start development environment with SQLite
	@echo "Starting development environment with SQLite..."
	docker-compose -f $(COMPOSE_DEV_SQLITE) up -d --build

.PHONY: docker-dev-mysql
docker-dev-mysql: ## Start development environment with MySQL/MariaDB
	@echo "Starting development environment with MySQL/MariaDB..."
	docker-compose -f $(COMPOSE_DEV_MYSQL) up -d --build


# Docker targets - Production Environment
.PHONY: docker-prod-postgres
docker-prod-postgres: ## Start production environment with PostgreSQL
	@echo "Starting production environment with PostgreSQL..."
	docker-compose -f $(COMPOSE_PROD_POSTGRES) up -d

.PHONY: docker-prod-sqlite
docker-prod-sqlite: ## Start production environment with SQLite
	@echo "Starting production environment with SQLite..."
	docker-compose -f $(COMPOSE_PROD_SQLITE) up -d

.PHONY: docker-prod-mysql
docker-prod-mysql: ## Start production environment with MySQL/MariaDB
	@echo "Starting production environment with MySQL/MariaDB..."
	docker-compose -f $(COMPOSE_PROD_MYSQL) up -d


# Docker targets - Standard Configurations
.PHONY: docker-postgres
docker-postgres: ## Start PostgreSQL environment
	@echo "Starting PostgreSQL environment..."
	docker-compose -f $(COMPOSE_POSTGRES) up -d

.PHONY: docker-sqlite
docker-sqlite: ## Start SQLite environment
	@echo "Starting SQLite environment..."
	docker-compose -f $(COMPOSE_SQLITE) up -d

.PHONY: docker-mysql
docker-mysql: ## Start MySQL/MariaDB environment
	@echo "Starting MySQL/MariaDB environment..."
	docker-compose -f $(COMPOSE_MYSQL) up -d

# Docker targets - Service Management
.PHONY: down-dev-postgres
down-dev-postgres: ## Stop development PostgreSQL services
	@echo "Stopping development PostgreSQL services..."
	docker-compose -f $(COMPOSE_DEV_POSTGRES) down

.PHONY: down-dev-sqlite
down-dev-sqlite: ## Stop development SQLite services
	@echo "Stopping development SQLite services..."
	docker-compose -f $(COMPOSE_DEV_SQLITE) down

.PHONY: down-prod-postgres
down-prod-postgres: ## Stop production PostgreSQL services
	@echo "Stopping production PostgreSQL services..."
	docker-compose -f $(COMPOSE_PROD_POSTGRES) down

.PHONY: down-prod-sqlite
down-prod-sqlite: ## Stop production SQLite services
	@echo "Stopping production SQLite services..."
	docker-compose -f $(COMPOSE_PROD_SQLITE) down

.PHONY: down-postgres
down-postgres: ## Stop PostgreSQL services
	@echo "Stopping PostgreSQL services..."
	docker-compose -f $(COMPOSE_POSTGRES) down

.PHONY: down-sqlite
down-sqlite: ## Stop SQLite services
	@echo "Stopping SQLite services..."
	docker-compose -f $(COMPOSE_SQLITE) down

.PHONY: down-mysql
down-mysql: ## Stop MySQL/MariaDB services
	@echo "Stopping MySQL/MariaDB services..."
	docker-compose -f $(COMPOSE_MYSQL) down

.PHONY: down-dev-mysql
down-dev-mysql: ## Stop development MySQL/MariaDB services
	@echo "Stopping development MySQL/MariaDB services..."
	docker-compose -f $(COMPOSE_DEV_MYSQL) down

.PHONY: down-prod-mysql
down-prod-mysql: ## Stop production MySQL/MariaDB services
	@echo "Stopping production MySQL/MariaDB services..."
	docker-compose -f $(COMPOSE_PROD_MYSQL) down

.PHONY: docker-down
docker-down: down-dev-$(DEFAULT_DEV_DB) ## Stop default development services

.PHONY: docker-down-all
docker-down-all: ## Stop all Docker Compose services
	@echo "Stopping all Docker services..."
	@docker-compose -f $(COMPOSE_DEV_POSTGRES) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_DEV_SQLITE) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_DEV_MYSQL) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_PROD_POSTGRES) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_PROD_SQLITE) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_PROD_MYSQL) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_POSTGRES) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_SQLITE) down 2>/dev/null || true
	@docker-compose -f $(COMPOSE_MYSQL) down 2>/dev/null || true

# Database targets
.PHONY: migrate-init
migrate-init: ## Run database migrations init
	@echo "Running database migrations init..."
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db init

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db migrate

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	@echo "Rolling back database migrations..."
	cd apps/server && ../../scripts/tool.sh go run cmd/bun/main.go db rollback


# Quick database environment switchers
.PHONY: switch-to-postgres
switch-to-postgres: docker-down-all dev-postgres ## Switch to PostgreSQL development environment
	@echo "Switched to PostgreSQL development environment"

.PHONY: switch-to-sqlite
switch-to-sqlite: docker-down-all dev-sqlite ## Switch to SQLite development environment
	@echo "Switched to SQLite development environment"

.PHONY: test-server
test-server: ## Test the server
	@echo "Testing the server..."
	cd apps/server && ../../scripts/tool.sh go test -v ./internal/...

.PHONY: lint-web
lint-web: ## Test the web
	@echo "Testing the web..."
	cd apps/web && ../../scripts/tool.sh pnpm lint && ../../scripts/tool.sh pnpm build

# Producer targets (deprecated — use engine instead)
.PHONY: build-producer
build-producer: ## DEPRECATED: Build the producer binary (use build-engine)
	@echo "Building producer..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/producer ./cmd/producer

.PHONY: run-producer
run-producer: ## DEPRECATED: Run the producer service (use run-engine)
	@echo "Running producer..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go run ./cmd/producer/main.go

# Ingester targets (deprecated — use engine instead)
.PHONY: build-ingester
build-ingester: ## DEPRECATED: Build the ingester binary (use build-engine)
	@echo "Building ingester..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/ingester ./cmd/ingester

.PHONY: run-ingester
run-ingester: ## DEPRECATED: Run the ingester service (use run-engine)
	@echo "Running ingester..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go run ./cmd/ingester/main.go

# Embed web frontend into Go API for local testing without Docker
.PHONY: embed-web
embed-web: ## Build frontend and copy dist into Go embed target (apps/server/internal/frontend/web/)
	@echo "Building web frontend..."
	pnpm --filter=web run build
	@echo "Copying dist to embed target..."
	rm -rf apps/server/internal/frontend/web/*
	cp -r apps/web/dist/. apps/server/internal/frontend/web/
	@echo "Frontend embedded. Run 'make build-api' to compile the API with the embedded frontend."

# Engine targets (replaces producer + worker + ingester)
.PHONY: build-engine
build-engine: ## Build the engine binary (scheduler + worker pool + ingester)
	@echo "Building engine..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go build -o ../../bin/engine ./cmd/engine

.PHONY: run-engine
run-engine: ## Run the engine service
	@echo "Running engine..."
	cd $(GO_SERVER_DIR) && ../../scripts/tool.sh go run ./cmd/engine/main.go

.PHONY: setup
setup: ## Setup development environment (asdf or manual)
	@echo "🚀 Setting up Peekaping development environment..."
	@if command -v asdf >/dev/null 2>&1; then \
		echo "✅ asdf found - using asdf for tool management"; \
		echo "📦 Adding asdf plugins..."; \
		asdf plugin add golang || true; \
		asdf plugin add nodejs || true; \
		asdf plugin add pnpm || true; \
		echo "🔧 Installing tools with asdf..."; \
		asdf install; \
		echo "✅ Setup complete! Tools installed via asdf:"; \
		echo "  - Go: $$(asdf current golang)"; \
		echo "  - Node.js: $$(asdf current nodejs)"; \
		echo "  - pnpm: $$(asdf exec pnpm --version)"; \
	else \
		echo "⚠️  asdf not found - you'll need to install tools manually"; \
		echo ""; \
		echo "Required tools:"; \
		echo "  - Go 1.24.1"; \
		echo "  - Node.js 22.0.0"; \
		echo "  - pnpm 9.0.0"; \
		echo ""; \
		echo "Installation options:"; \
		echo "  1. Install asdf: https://asdf-vm.com/guide/getting-started.html"; \
		echo "  2. Install tools manually from their official websites"; \
		echo ""; \
		echo "If you install asdf, run 'make setup' again to automatically install tools."; \
	fi
	@echo ""
	@echo "🎉 Development environment setup complete!"
	@echo "Run 'make help' to see available commands."

.PHONY: install
install: ## Install all dependencies (pnpm install + go mod tidy)
	@echo "📦 Installing all project dependencies..."
	@echo "Installing Node.js dependencies..."
	./scripts/tool.sh pnpm install
	@echo "Tidying Go modules..."
	cd apps/server && ../../scripts/tool.sh go mod tidy
	@echo "✅ All dependencies installed successfully!"

.PHONY: dev
dev: ## Start development environment
	./scripts/tool.sh pnpm run dev dev:api dev:ingester dev:producer dev:worker docs:watch
