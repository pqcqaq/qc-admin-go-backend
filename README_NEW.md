# Go Backend Project

这是一个结构化的Go后端项目，使用Gin框架和Ent ORM。

## 项目结构

```
go-backend/
├── configs/                    # 配置管理
│   └── config.go              # 应用配置
├── ent/                       # Ent ORM生成的代码
│   ├── client.go
│   ├── schema/
│   └── ...
├── internal/                  # 内部应用逻辑
│   ├── handlers/              # HTTP处理器
│   │   ├── user_handler.go    # 用户处理器
│   │   └── health_handler.go  # 健康检查处理器
│   ├── routes/                # 路由配置
│   │   └── routes.go          # 路由设置
│   └── services/              # 业务逻辑服务
│       └── user_service.go    # 用户服务
├── pkg/                       # 公共包
│   └── database/              # 数据库配置
│       └── database.go
├── shared/                    # 共享类型和模型
│   └── models/
│       └── user.go            # 用户相关模型
├── main.go                    # 应用入口
├── go.mod                     # Go模块文件
├── go.sum                     # Go依赖锁定文件
├── Makefile                   # 构建脚本
└── README.md                  # 项目说明
```

## 架构设计

### 分层架构

1. **Handler Layer (处理器层)**: 处理HTTP请求和响应
2. **Service Layer (服务层)**: 包含业务逻辑
3. **Repository Layer (仓储层)**: 由Ent ORM提供，处理数据访问

### 目录说明

- **`configs/`**: 应用配置管理，包括数据库配置、服务器配置等
- **`internal/`**: 内部应用代码，不对外暴露
  - **`handlers/`**: HTTP请求处理器，负责处理路由和HTTP相关逻辑
  - **`services/`**: 业务逻辑服务，包含核心业务逻辑
  - **`routes/`**: 路由配置，集中管理所有API路由
- **`pkg/`**: 可重用的公共包，可以被其他项目引用
- **`shared/`**: 共享的类型定义、模型等
- **`ent/`**: Ent ORM生成的代码

## 快速开始

### 安装依赖

```bash
make init
```

### 运行应用

```bash
# 开发模式运行
make run

# 或者使用热重载（需要先安装air）
go install github.com/cosmtrek/air@latest
make dev
```

### 构建应用

```bash
make build
```

## API 端点

### 健康检查
- `GET /health` - 健康检查

### 用户管理
- `GET /api/v1/users` - 获取所有用户
- `GET /api/v1/users/:id` - 获取特定用户
- `POST /api/v1/users` - 创建用户
- `PUT /api/v1/users/:id` - 更新用户
- `DELETE /api/v1/users/:id` - 删除用户

## 开发指南

### 添加新的功能模块

1. 在 `ent/schema/` 中定义新的数据模型
2. 在 `shared/models/` 中定义请求/响应模型
3. 在 `internal/services/` 中实现业务逻辑
4. 在 `internal/handlers/` 中实现HTTP处理器
5. 在 `internal/routes/` 中添加路由配置

### 数据库操作

```bash
# 生成Ent代码
make generate

# 重置数据库
make reset-db
```

### 代码质量

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 代码检查（需要安装golangci-lint）
make lint
```

## 配置

应用配置在 `configs/config.go` 中定义，包括：

- 服务器端口配置
- 数据库连接配置
- Gin运行模式配置

## 依赖项

- **Gin**: HTTP Web框架
- **Ent**: ORM框架
- **SQLite**: 数据库（可替换为其他数据库）

## 许可证

MIT License
