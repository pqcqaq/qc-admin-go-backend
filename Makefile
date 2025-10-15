# Makefile for Go Backend

# 变量定义
BINARY_NAME=go-backend
MAIN_PATH=./main.go

# 数据库驱动构建标签
DB_TAGS?=all

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Database driver build tags:"
	@echo "  all        - 包含所有数据库驱动（默认）"
	@echo "  sqlite     - 仅包含 SQLite 驱动"
	@echo "  mysql      - 仅包含 MySQL 驱动"
	@echo "  postgres   - 仅包含 PostgreSQL 驱动"
	@echo "  clickhouse - 仅包含 ClickHouse 驱动"
	@echo "  sqlserver  - 仅包含 SQL Server 驱动"
	@echo "  oracle     - 仅包含 Oracle 驱动"
	@echo ""
	@echo "示例用法："
	@echo "  make build                    # 构建包含所有驱动的版本"
	@echo "  make build DB_TAGS=sqlite     # 构建仅包含 SQLite 驱动的版本"
	@echo "  make build DB_TAGS=\"mysql postgres\" # 构建包含 MySQL 和 PostgreSQL 驱动的版本"
	@echo "  make swagger                  # 生成 Swagger 文档"
	@echo "  make run-with-swagger         # 生成文档并启动服务器"
	@echo "  make init-full                # 完整初始化项目（包括 Swagger）"

# 初始化项目
.PHONY: init
init: ## 初始化项目（下载依赖和生成Ent代码）
	go mod tidy
	go install entgo.io/ent/cmd/ent@latest
	go generate ./ent

# 完整初始化项目
.PHONY: init-full
init-full: ## 完整初始化项目（包括依赖、Ent代码和Swagger文档）
	@echo "Initializing project..."
	go mod tidy
	go install entgo.io/ent/cmd/ent@latest
	@$(MAKE) install-swagger
	go generate ./ent
	@$(MAKE) swagger
	@echo "Project initialization completed!"

# 生成Ent代码
.PHONY: generate
generate: ## 生成Ent ORM代码
	go generate ./ent

# 生成Swagger文档
.PHONY: swagger
swagger: ## 生成Swagger API文档
	@echo "Generating Swagger documentation..."
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation generated successfully!"
	@echo "Swagger UI will be available at: http://localhost:8080/swagger/index.html"

# 安装Swagger工具
.PHONY: install-swagger
install-swagger: ## 安装Swagger CLI工具
	go install github.com/swaggo/swag/cmd/swag@latest

# 构建项目
.PHONY: build
build: ## 构建项目
	go build -tags "$(DB_TAGS)" -o $(BINARY_NAME) $(MAIN_PATH)

# 构建指定数据库驱动版本
.PHONY: build-sqlite
build-sqlite: ## 构建仅包含SQLite驱动的版本
	go build -tags "sqlite" -o $(BINARY_NAME)-sqlite $(MAIN_PATH)

.PHONY: build-mysql
build-mysql: ## 构建仅包含MySQL驱动的版本
	go build -tags "mysql" -o $(BINARY_NAME)-mysql $(MAIN_PATH)

.PHONY: build-postgres
build-postgres: ## 构建仅包含PostgreSQL驱动的版本
	go build -tags "postgres" -o $(BINARY_NAME)-postgres $(MAIN_PATH)

.PHONY: build-all-variants
build-all-variants: ## 构建所有单数据库驱动版本
	@echo "Building all database driver variants..."
	@$(MAKE) build-sqlite
	@$(MAKE) build-mysql
	@$(MAKE) build-postgres
	@echo "All variants built successfully!"

# 运行项目
.PHONY: run
run: ## 运行项目
	go run -tags "$(DB_TAGS)" $(MAIN_PATH)

# 运行项目并生成Swagger文档
.PHONY: run-with-swagger
run-with-swagger: ## 生成Swagger文档并运行项目
	@echo "Generating Swagger documentation..."
	@$(MAKE) swagger
	@echo "Starting server with Swagger documentation..."
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "Health Check: http://localhost:8080/health"
	go run -tags "$(DB_TAGS)" $(MAIN_PATH)

# 运行项目（带热重载）
.PHONY: dev
dev: ## 开发模式运行（需要安装air）
	air

# 清理构建文件
.PHONY: clean
clean: ## 清理构建文件
	rm -f $(BINARY_NAME)*
	rm -f ent.db

# 清理所有生成文件
.PHONY: clean-all
clean-all: ## 清理构建文件和生成的文档
	rm -f $(BINARY_NAME)*
	rm -f ent.db
	rm -rf docs/

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
