# Go Backend 模板使用手册

## 📖 模板概述

这是一个生产就绪的Go后端项目模板，集成了现代Web开发的最佳实践和常用功能。模板包含完整的用户管理、文件处理、缓存、日志等基础设施，并提供了扫描管理作为业务逻辑示例。

## 🎯 目标场景

### 适用于以下项目类型

- ✅ **RESTful API服务** - 提供标准的CRUD操作
- ✅ **企业管理系统** - 内置用户权限、文件管理等功能
- ✅ **内容管理平台** - 支持富媒体内容和文件存储
- ✅ **数据处理平台** - 包含数据导入导出功能
- ✅ **微服务架构** - 模块化设计，易于拆分

### 技术要求

- Go 1.21+
- Redis (可选，用于缓存)
- 数据库: SQLite/MySQL/PostgreSQL
- AWS S3 (可选，用于文件存储)

## 🚀 快速开始

### 1. 获取模板

```bash
# 克隆模板仓库
git clone <your-template-repo> my-new-project
cd my-new-project

# 删除原有git历史，开始新项目
rm -rf .git
git init
```

### 2. 自定义项目

```bash
# 修改模块名 (将 go-backend 替换为您的项目名)
# 修改 go.mod
module my-awesome-api

# 批量替换代码中的导入路径
find . -name "*.go" -exec sed -i 's/go-backend/my-awesome-api/g' {} \;
```

### 3. 配置环境

```bash
# 复制配置文件
cp config.yaml config.local.yaml

# 编辑配置文件
vim config.local.yaml
```

### 4. 安装依赖并启动

```bash
# 下载依赖
go mod download

# 生成数据库代码
go generate ./database/generate.go

# 启动开发服务器
go run main.go -c config.local.yaml
```

## 📝 定制指南

### 修改项目名称和信息

1. **更新 go.mod**
   ```go.mod
   module your-project-name
   
   go 1.21
   ```

2. **更新 main.go 中的导入**
   ```go
   import (
       "your-project-name/internal/routes"
       "your-project-name/pkg/configs"
       // ...
   )
   ```

3. **批量替换导入路径**
   ```bash
   # 使用sed或IDE的全局替换功能
   find . -name "*.go" -exec sed -i 's/go-backend/your-project-name/g' {} \;
   ```

### 替换示例业务逻辑

模板中的"扫描管理"是一个示例业务，您可以将其替换为自己的业务逻辑：

1. **删除示例代码**
   ```bash
   # 删除扫描相关文件
   rm database/schema/scan.go
   rm internal/handlers/scan_handler.go
   rm internal/routes/scan.go
   rm internal/funcs/scanfunc.go
   rm shared/models/scan.go
   ```

2. **创建新业务模块**
   ```bash
   # 例如：产品管理模块
   touch database/schema/product.go
   touch internal/handlers/product_handler.go
   touch internal/routes/product.go
   touch internal/funcs/productfunc.go
   touch shared/models/product.go
   ```

3. **更新路由注册**
   ```go
   // internal/routes/routes.go
   func (r *Router) SetupRoutes(engine *gin.Engine) {
       // ... 其他路由
       r.setupProductRoutes(api)  // 替换 setupScanRoutes
   }
   ```

## 🏗️ 架构说明

### 目录结构说明

```
项目根目录/
├── 配置层 (config.*.yaml)     # 多环境配置文件
├── 入口层 (main.go)           # 应用程序入口
├── 数据层 (database/)         # 数据模型和ORM
├── 业务层 (internal/)         # 核心业务逻辑
├── 工具层 (pkg/)             # 可复用的工具包
├── 模型层 (shared/)          # 共享数据结构
└── 部署层 (docker-compose/)  # 容器化配置
```

### 分层设计原则

1. **controller层** (`internal/handlers/`)
   - 处理HTTP请求和响应
   - 参数验证和转换
   - 调用业务逻辑层

2. **service层** (`internal/funcs/`)
   - 核心业务逻辑
   - 数据处理和转换
   - 调用数据访问层

3. **repository层** (Ent ORM)
   - 数据持久化
   - 数据库操作
   - 缓存管理

4. **model层** (`shared/models/`)
   - 数据传输对象 (DTO)
   - 请求/响应结构体
   - 业务实体定义

## 🔧 常见定制场景

### 1. 添加认证和授权

```go
// internal/middleware/auth.go
package middleware

import (
    "strings"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 验证JWT token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
            return []byte("your-secret-key"), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
            c.Abort()
            return
        }

        // 将用户信息存储到上下文
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"])
            c.Set("username", claims["username"])
        }

        c.Next()
    }
}
```

### 2. 添加数据库关联

```go
// database/schema/order.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
)

type Order struct {
    ent.Schema
}

func (Order) Fields() []ent.Field {
    return []ent.Field{
        field.String("order_no").Unique(),
        field.Float("total_amount"),
        field.Enum("status").Values("pending", "paid", "cancelled"),
    }
}

func (Order) Edges() []ent.Edge {
    return []ent.Edge{
        // 订单属于某个用户
        edge.From("user", User.Type).
            Ref("orders").
            Unique(),
        // 订单包含多个商品
        edge.To("items", OrderItem.Type),
    }
}
```

### 3. 添加缓存支持

```go
// internal/funcs/productfunc.go
package funcs

import (
    "context"
    "encoding/json"
    "time"
    "your-project/pkg/caching"
)

func GetProductWithCache(ctx context.Context, productID uint64) (*Product, error) {
    cacheKey := fmt.Sprintf("product:%d", productID)
    
    // 尝试从缓存获取
    if cached, err := caching.Get(ctx, cacheKey); err == nil {
        var product Product
        if err := json.Unmarshal([]byte(cached), &product); err == nil {
            return &product, nil
        }
    }
    
    // 从数据库获取
    product, err := GetProductFromDB(ctx, productID)
    if err != nil {
        return nil, err
    }
    
    // 存入缓存
    if data, err := json.Marshal(product); err == nil {
        caching.Set(ctx, cacheKey, string(data), 5*time.Minute)
    }
    
    return product, nil
}
```

### 4. 添加消息队列

```go
// pkg/queue/queue.go
package queue

import (
    "encoding/json"
    "github.com/streadway/amqp"
)

type TaskQueue struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func NewTaskQueue() (*TaskQueue, error) {
    conn, err := amqp.Dial("amqp://localhost")
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    
    return &TaskQueue{conn: conn, channel: ch}, nil
}

func (tq *TaskQueue) PublishTask(queueName string, task any) error {
    data, err := json.Marshal(task)
    if err != nil {
        return err
    }
    
    return tq.channel.Publish(
        "",        // exchange
        queueName, // routing key
        false,     // mandatory
        false,     // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        data,
        },
    )
}
```

## 📊 性能优化建议

### 1. 数据库优化

```go
// 使用批量操作
func CreateMultipleUsers(ctx context.Context, users []*User) error {
    bulk := make([]*ent.UserCreate, len(users))
    for i, user := range users {
        bulk[i] = client.User.Create().
            SetName(user.Name).
            SetEmail(user.Email)
    }
    
    _, err := client.User.CreateBulk(bulk...).Save(ctx)
    return err
}

// 使用select优化查询
func GetUserWithPosts(ctx context.Context, userID uint64) (*User, error) {
    return client.User.Query().
        Where(user.ID(userID)).
        WithPosts(func(q *ent.PostQuery) {
            q.Select(post.FieldTitle, post.FieldCreatedAt)
        }).
        Only(ctx)
}
```

### 2. 缓存策略

```go
// 多级缓存
type CacheService struct {
    local  *sync.Map          // 本地缓存
    redis  *redis.Client      // 分布式缓存
}

func (cs *CacheService) Get(key string) (any, bool) {
    // 先查本地缓存
    if value, ok := cs.local.Load(key); ok {
        return value, true
    }
    
    // 再查Redis
    if value, err := cs.redis.Get(context.Background(), key).Result(); err == nil {
        cs.local.Store(key, value) // 回写本地缓存
        return value, true
    }
    
    return nil, false
}
```

### 3. 连接池配置

```go
// pkg/database/database.go
func NewDatabase(config DatabaseConfig) (*ent.Client, error) {
    db, err := sql.Open(config.Driver, config.Source)
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    db.SetMaxOpenConns(25)               // 最大打开连接数
    db.SetMaxIdleConns(5)                // 最大空闲连接数
    db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生存时间
    
    return ent.NewClient(ent.Driver(sql.OpenDB(config.Driver, db))), nil
}
```

## 🧪 测试策略

### 1. 单元测试

```go
// internal/handlers/user_handler_test.go
func TestUserHandler_CreateUser(t *testing.T) {
    // 准备测试数据
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // 模拟请求
    reqBody := `{"name":"test user","email":"test@example.com"}`
    c.Request = httptest.NewRequest("POST", "/users", strings.NewReader(reqBody))
    c.Request.Header.Set("Content-Type", "application/json")
    
    // 执行测试
    handler := NewUserHandler()
    handler.CreateUser(c)
    
    // 验证结果
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### 2. 集成测试

```go
// tests/integration/user_test.go
func TestUserAPI(t *testing.T) {
    // 启动测试服务器
    router := setupTestRouter()
    server := httptest.NewServer(router)
    defer server.Close()
    
    // 测试创建用户
    resp, err := http.Post(server.URL+"/api/v1/users", "application/json", 
        strings.NewReader(`{"name":"test","email":"test@example.com"}`))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

## 🚀 部署建议

### 1. Docker部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config.prod.yaml ./config.yaml

CMD ["./main"]
```

### 2. Kubernetes部署

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-backend
  template:
    metadata:
      labels:
        app: go-backend
    spec:
      containers:
      - name: go-backend
        image: your-registry/go-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

### 3. CI/CD管道

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./...
      
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build Docker image
        run: docker build -t go-backend .
      - name: Push to registry
        run: docker push your-registry/go-backend:latest
```

## 🔍 监控和维护

### 1. 健康检查

模板已内置健康检查端点 `/health`，可以扩展为更详细的检查：

```go
// internal/handlers/health_handler.go
func (h *HealthHandler) DetailedHealth(c *gin.Context) {
    status := gin.H{
        "status": "ok",
        "timestamp": time.Now(),
        "checks": gin.H{},
    }
    
    // 检查数据库连接
    if err := database.Ping(); err != nil {
        status["checks"].(gin.H)["database"] = "error"
        status["status"] = "error"
    } else {
        status["checks"].(gin.H)["database"] = "ok"
    }
    
    // 检查Redis连接
    if err := redis.Ping().Err(); err != nil {
        status["checks"].(gin.H)["redis"] = "error"
        status["status"] = "error"
    } else {
        status["checks"].(gin.H)["redis"] = "ok"
    }
    
    if status["status"] == "error" {
        c.JSON(http.StatusServiceUnavailable, status)
    } else {
        c.JSON(http.StatusOK, status)
    }
}
```

### 2. 日志集中化

```go
// pkg/logging/structured.go
func LogRequest(c *gin.Context) {
    logger.Info("http request",
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
        zap.String("ip", c.ClientIP()),
        zap.String("user_agent", c.Request.UserAgent()),
        zap.Duration("duration", time.Since(start)),
        zap.Int("status", c.Writer.Status()),
    )
}
```

## 📚 最佳实践

1. **代码组织**：保持清晰的分层架构
2. **错误处理**：使用统一的错误处理机制
3. **配置管理**：环境变量优先，配置文件作为默认值
4. **安全性**：输入验证、SQL注入防护、XSS防护
5. **性能**：合理使用缓存、数据库连接池、异步处理
6. **可维护性**：完善的测试覆盖、清晰的文档、代码审查

## ❓ 常见问题

### Q: 如何更换数据库？
A: 修改配置文件中的数据库驱动和连接字符串，Ent ORM支持多种数据库。

### Q: 如何添加认证？
A: 实现JWT中间件，参考上述认证示例代码。

### Q: 如何部署到生产环境？
A: 使用Docker容器化部署，配置好生产环境的配置文件。

### Q: 如何扩展新的API？
A: 按照模板的分层架构，依次添加schema、handler、route和func文件。

---

🎉 现在您已经掌握了使用这个Go Backend模板的所有知识！开始构建您的项目吧！
