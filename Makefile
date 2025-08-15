# Makefile for Go Backend

# 变量定义
BINARY_NAME=go-backend
MAIN_PATH=./main.go

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# 初始化项目
.PHONY: init
init: ## 初始化项目（下载依赖和生成Ent代码）
	go mod tidy
	go install entgo.io/ent/cmd/ent@latest
	go generate ./ent

# 生成Ent代码
.PHONY: generate
generate: ## 生成Ent ORM代码
	go generate ./ent

# 构建项目
.PHONY: build
build: ## 构建项目
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# 运行项目
.PHONY: run
run: ## 运行项目
	go run $(MAIN_PATH)

# 运行项目（带热重载）
.PHONY: dev
dev: ## 开发模式运行（需要安装air）
	air

# 清理构建文件
.PHONY: clean
clean: ## 清理构建文件
	rm -f $(BINARY_NAME)
	rm -f ent.db

# 测试
.PHONY: test
test: ## 运行测试
	go test -v ./...

# 格式化代码
.PHONY: fmt
fmt: ## 格式化代码
	go fmt ./...

# 检查代码
.PHONY: lint
lint: ## 检查代码（需要安装golangci-lint）
	golangci-lint run

# 重置数据库
.PHONY: reset-db
reset-db: ## 重置数据库
	rm -f ent.db
