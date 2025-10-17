.PHONY: help build run test clean dev fmt vet lint build-linux

# 变量定义
APP_NAME := emby-bot
BUILD_DIR := bin
MAIN_PATH := cmd/server/main.go
DATA_DIR := data
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

help: ## 显示帮助信息
	@echo "Emby Telegram Bot - 可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@echo "编译项目..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✓ 编译完成: $(BUILD_DIR)/$(APP_NAME)"

build-linux: ## 编译 Linux AMD64 版本
	@echo "编译 Linux AMD64 版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "✓ 编译完成: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

run: ## 运行项目
	@echo "启动 Bot..."
	@mkdir -p $(DATA_DIR)
	go run $(MAIN_PATH)

dev: ## 开发模式运行 (需要 air)
	@echo "开发模式启动..."
	@mkdir -p $(DATA_DIR)
	air

test: ## 运行测试
	@echo "运行测试..."
	go test -v -race ./...

test-cover: ## 运行测试并生成覆盖率报告
	@echo "生成测试覆盖率..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告: coverage.html"

clean: ## 清理编译文件和日志
	@echo "清理文件..."
	rm -rf $(BUILD_DIR)/
	rm -rf logs/
	rm -f coverage.out coverage.html
	@echo "✓ 清理完成"

clean-data: ## 清理数据库文件 (慎用!)
	@echo "警告: 即将删除数据库文件!"
	@read -p "确认删除? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	rm -rf $(DATA_DIR)/
	@echo "✓ 数据已清理"

deps: ## 下载并整理依赖
	@echo "下载依赖..."
	go mod download
	go mod tidy
	@echo "✓ 依赖已更新"

fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@echo "✓ 代码已格式化"

vet: ## 运行 go vet
	@echo "运行 go vet..."
	go vet ./...
	@echo "✓ go vet 检查通过"

lint: ## 代码检查 (需要 golangci-lint)
	@echo "代码检查..."
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run
	@echo "✓ 代码检查通过"

install: ## 安装到 GOPATH/bin
	@echo "安装..."
	go install $(MAIN_PATH)
	@echo "✓ 已安装到 GOPATH"

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest .
	@echo "✓ Docker 镜像构建完成"

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run --rm -it --env-file .env $(APP_NAME):latest

check: fmt vet ## 快速检查 (格式化 + vet)
	@echo "✓ 快速检查完成"

all: clean deps fmt vet build ## 完整构建流程
	@echo "✓ 完整构建完成"
