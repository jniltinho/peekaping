GO          := go
CGO_ENABLED := 1
BIN_DIR     := bin
BINARY      := $(BIN_DIR)/monitoring
PREFIX      := peekaping/version

VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
VER        := $(patsubst v%,%,$(VERSION))
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS    := -trimpath -ldflags "-s -w \
              -X $(PREFIX).Version=$(VER) \
              -X $(PREFIX).BuildTime=$(BUILD_TIME) \
              -X $(PREFIX).GitCommit=$(GIT_COMMIT)"

.PHONY: build run-api run-engine migrate-up migrate-status clean help

.DEFAULT_GOAL := build

build: ## Build the monitoring binary
	@mkdir -p $(BIN_DIR)
	@rm -f $(BINARY)
	@echo "Building monitoring $(VER) ($(GIT_COMMIT))..."
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINARY) $(LDFLAGS) .
	@if command -v upx > /dev/null 2>&1; then \
		echo "Compressing with UPX..."; \
		upx --best --lzma $(BINARY); \
	else \
		echo "UPX not found, skipping compression."; \
	fi

run-api: ## Run the API server
	$(GO) run . api

run-engine: ## Run the engine (scheduler/worker/ingester)
	$(GO) run . engine

migrate-up: ## Apply / update database schema (GORM AutoMigrate)
	$(GO) run . migrate up

migrate-status: ## Check database connectivity
	$(GO) run . migrate status

clean: ## Remove built binary
	@rm -f $(BINARY)

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
