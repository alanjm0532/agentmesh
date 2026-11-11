GO ?= go
APP_NAME := agentmesh
BIN_DIR := bin

.DEFAULT_GOAL := help
.PHONY: init wire generate fmt lint build run all help

init: ## 下载项目和代码生成工具依赖
	$(GO) mod download

wire: ## 生成 Wire 依赖注入代码
	$(GO) tool wire ./cmd/server

generate: wire ## 生成项目代码（当前为 Wire）

fmt: ## 格式化 Go 代码
	$(GO) fmt ./...

lint: generate ## 使用 go vet 检查代码
	$(GO) vet ./...

build: generate ## 生成 Wire 代码并构建 bin/agentmesh
	mkdir -p $(BIN_DIR)
	$(GO) build -tags="sonic avx" -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

run: generate ## 生成 Wire 代码并启动服务
	$(GO) run ./cmd/server

all: build ## 生成代码并构建

help: ## 显示可用命令
	@printf 'Usage: make [target]\n\n'
	@awk 'BEGIN { FS = ":.*## " } /^[a-zA-Z_-]+:.*## / { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
