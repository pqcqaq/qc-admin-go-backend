# 错误处理中间件使用指南

## 概述

本项目实现了一个统一的错误处理中间件系统，可以优雅地处理应用程序中的各种错误，包括panic、业务逻辑错误、验证错误等。

## 文件结构

```
internal/middleware/
├── errors.go          # 自定义错误类型和预定义错误
├── error_handler.go   # 错误处理中间件实现
└── examples.go        # 使用示例和最佳实践
```

## 主要特性

### 1. 自定义错误类型
- **CustomError**: 包含错误代码、消息、数据和堆栈信息
- **预定义错误代码**: 通用错误和业务特定错误
- **便利函数**: 快速创建常见类型的错误

### 2. 错误处理中间件
- **ErrorHandlerMiddleware**: 处理panic恢复
- **ErrorHandler**: 处理通过gin.Context.Error()抛出的错误
- **统一响应格式**: 标准化的错误响应结构

### 3. 日志记录
- **结构化日志**: JSON格式的错误日志
- **上下文信息**: 包含请求路径、方法、IP等信息
- **堆栈跟踪**: 在开发环境中包含详细堆栈信息

## 快速开始

### 1. 注册中间件

在路由设置中注册错误处理中间件：

```go
func (r *Router) SetupRoutes(engine *gin.Engine) {
    // 注册错误处理中间件
    engine.Use(middleware.ErrorHandlerMiddleware()) // 处理panic恢复
    engine.Use(middleware.ErrorHandler())           // 处理gin.Error
    
    // ... 其他路由设置
}
```

### 2. 在处理器中使用

#### 方式一：使用 ThrowError (推荐)

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        middleware.ThrowError(c, middleware.BadRequestError("用户ID不能为空", nil))
        return
    }
    
    user, err := h.userService.GetUser(id)
    if err != nil {
        middleware.ThrowError(c, middleware.UserNotFoundError(map[string]any{
            "id": id,
        }))
        return
    }
    
    c.JSON(200, gin.H{"success": true, "data": user})
}
```

#### 方式二：使用 Panic

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        middleware.PanicWithError(middleware.ValidationError("请求数据格式错误", err.Error()))
    }
    
    // ... 业务逻辑
}
```

## 预定义错误类型

### 通用错误
- `ErrCodeBadRequest (400)`: 请求参数错误
- `ErrCodeUnauthorized (401)`: 未授权
- `ErrCodeForbidden (403)`: 禁止访问
- `ErrCodeNotFound (404)`: 资源未找到
- `ErrCodeInternal (500)`: 内部服务器错误

### 业务错误
- `ErrCodeUserNotFound (1001)`: 用户不存在
- `ErrCodeUserExists (1002)`: 用户已存在
- `ErrCodeInvalidUserData (1003)`: 用户数据无效
- `ErrCodeDatabaseError (2001)`: 数据库错误
- `ErrCodeValidationError (3001)`: 数据验证错误

## 便利函数

### 创建常见错误

```go
// 400 错误
middleware.BadRequestError("参数错误", data)

// 404 错误
middleware.NotFoundError("资源未找到", data)

// 401 错误
middleware.UnauthorizedError("未授权", data)

// 500 错误
middleware.InternalServerError("服务器错误", data)

// 数据库错误
middleware.DatabaseError("数据库操作失败", data)

// 验证错误
middleware.ValidationError("数据验证失败", data)

// 用户相关错误
middleware.UserNotFoundError(data)
middleware.UserExistsError(data)
```

### 创建自定义错误

```go
// 基本自定义错误
customErr := middleware.NewCustomError(5001, "自定义业务错误", data)

// 带堆栈信息的错误
customErr := middleware.NewCustomErrorWithStack(5001, "严重错误", data)
```

## 错误响应格式

所有错误都将以统一的JSON格式返回：

```json
{
  "success": false,
  "code": 1001,
  "message": "用户不存在",
  "data": {
    "id": 123
  },
  "timestamp": "2025-08-15T10:30:00Z",
  "path": "/api/v1/users/123",
  "stack": "goroutine 1 [running]:\n..." // 仅在开发环境
}
```

## 日志格式

错误日志以JSON格式记录：

```json
{
  "type": "Error",
  "code": 1001,
  "message": "用户不存在",
  "method": "GET",
  "path": "/api/v1/users/123",
  "query": "",
  "user_agent": "Mozilla/5.0...",
  "ip": "127.0.0.1",
  "timestamp": "2025-08-15T10:30:00Z",
  "data": {"id": 123},
  "stack": "..."
}
```

## 最佳实践

### 1. 错误分层
- **Controller层**: 参数验证、权限检查
- **Service层**: 业务逻辑错误
- **Repository层**: 数据访问错误

### 2. 错误信息
- 使用清晰、用户友好的错误消息
- 包含有助于调试的上下文数据
- 避免暴露敏感信息

### 3. 错误代码
- 使用有意义的错误代码
- 保持错误代码的一致性
- 为不同的业务场景定义特定的错误代码

### 4. 性能考虑
- 避免在热路径中使用panic
- 优先使用ThrowError而不是PanicWithError
- 在生产环境中禁用详细的堆栈信息

## 测试错误处理

可以使用演示端点测试不同类型的错误：

```bash
# 测试验证错误
curl "http://localhost:8080/api/v1/demo-errors?type=validation"

# 测试未找到错误
curl "http://localhost:8080/api/v1/demo-errors?type=notfound"

# 测试panic错误
curl "http://localhost:8080/api/v1/demo-errors?type=panic"

# 测试业务逻辑错误
curl "http://localhost:8080/api/v1/demo-errors?type=business"
```

## 扩展错误类型

要添加新的错误类型，请在`errors.go`中：

1. 定义新的错误代码常量
2. 在`ErrorMessages`映射中添加错误消息
3. 创建便利函数（可选）

```go
const (
    ErrCodeCustomBusiness = 6001
)

var ErrorMessages = map[int]string{
    ErrCodeCustomBusiness: "自定义业务错误",
}

func CustomBusinessError(message string, data any) *CustomError {
    if message == "" {
        message = GetErrorMessage(ErrCodeCustomBusiness)
    }
    return NewCustomError(ErrCodeCustomBusiness, message, data)
}
```
