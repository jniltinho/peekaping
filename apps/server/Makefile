# Makefile for building Peekaping server binaries
#
# Usage:
#   make            # Build all binaries (default)
#   make api        # Build only the api binary
#   make engine     # Build only the engine binary
#   make bun        # Build only the bun migration tool
#   make clean      # Remove all built binaries

# Go parameters
GO            := go
CGO_ENABLED   := 0
LDFLAGS       := -s -w
GOFLAGS       := -trimpath

# Directories
BIN_DIR       := bin

# Binaries
API_BINARY    := $(BIN_DIR)/api
ENGINE_BINARY := $(BIN_DIR)/engine
BUN_BINARY    := $(BIN_DIR)/bun

.PHONY: all build help clean api engine bun

# Default target
.DEFAULT_GOAL := all

all: build ## Build all binaries (default)

build: api engine bun ## Build all binaries

# Individual build targets
api: ## Build the API server (includes embedded frontend if web/ is populated)
	@mkdir -p $(BIN_DIR)
	@echo "Building api..."
	CGO_ENABLED=$(CGO_ENABLED) GOFLAGS="$(GOFLAGS)" $(GO) build -o $(API_BINARY) -ldflags="$(LDFLAGS)" ./cmd/api

engine: ## Build the engine (scheduler + worker pool + ingester)
	@mkdir -p $(BIN_DIR)
	@echo "Building engine..."
	CGO_ENABLED=$(CGO_ENABLED) GOFLAGS="$(GOFLAGS)" $(GO) build -o $(ENGINE_BINARY) -ldflags="$(LDFLAGS)" ./cmd/engine

bun: ## Build the bun migration tool
	@mkdir -p $(BIN_DIR)
	@echo "Building bun..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BUN_BINARY) -ldflags="$(LDFLAGS)" ./cmd/bun

clean: ## Remove all built binaries
	@echo "Cleaning binaries..."
	@rm -rf $(BIN_DIR)/*
	@echo "Done."

help: ## Show this help message
	@echo "Peekaping Server - Build targets"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make           # Build all binaries"
	@echo "  make api       # Build only api"
	@echo "  make engine    # Build only engine"
	@echo "  make clean     # Remove bin/ contents"
