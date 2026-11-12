GO ?= go
APP_NAME := agentmesh
BIN_DIR := bin
ARGS ?=

.DEFAULT_GOAL := help
.PHONY: init wire generate fmt lint build build-cli build-server run run-cli all help

init: ## Download project and code generation tool dependencies
	$(GO) mod download

wire: ## Generate Wire dependency injection code
	$(GO) tool wire ./internal/server

generate: wire ## Generate project code (Wire)

fmt: ## Format Go code
	$(GO) fmt ./...

lint: generate ## Check code with go vet
	$(GO) vet ./...

build: build-cli build-server ## Build the CLI and HTTP server

build-cli: ## Build the CLI: bin/agentmesh
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd/agentmesh

build-server: generate ## Generate Wire code and build the HTTP server: bin/agentmesh-server
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME)-server ./cmd/server

run: generate ## Generate Wire code and start the HTTP server
	$(GO) run ./cmd/server $(ARGS)

run-cli: ## Run the CLI, e.g. make run-cli ARGS="ping"
	$(GO) run ./cmd/agentmesh $(ARGS)

all: build ## Generate code and build

help: ## Show available commands
	@printf 'Usage: make [target]\n\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z_-]+:.*## / { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
