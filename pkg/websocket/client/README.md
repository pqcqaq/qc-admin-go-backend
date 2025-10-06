# Go WebSocket Client

这是一个功能完整的Go语言WebSocket客户端，移植自TypeScript版本，支持主题订阅、自动重连、心跳检测和双向通信频道等功能。

## 特性

- ✅ **WebSocket连接管理** - 自动连接、断开和重连
- ✅ **主题订阅系统** - 支持MQTT风格的主题匹配 (`+` 单层通配符, `#` 多层通配符)
- ✅ **自动重连机制** - 指数退避算法，智能重连
- ✅ **心跳检测** - 自动检测连接状态
- ✅ **Token刷新** - 自动处理token过期和刷新
- ✅ **双向通信频道** - 支持创建专用通信频道
- ✅ **错误处理** - 完善的错误处理和回调机制
- ✅ **线程安全** - 并发安全的设计
- ✅ **调试日志** - 可选的详细调试信息

## 依赖

- `github.com/gorilla/websocket` - WebSocket连接
- `github.com/google/uuid` - UUID生成
- `go-backend/pkg/utils` - 主题匹配工具

## 快速开始

### 基本使用

```go
package main

import (
    "log"
    "time"
    "go-backend/pkg/websocket/client"
)

func main() {
    // 创建客户端
    wsClient := client.NewSocketClient(client.SocketOptions{
        URL:               "ws://localhost:8080/ws",
        Token:             "your-auth-token",
        HeartbeatInterval: 30 * time.Second,
        Debug:             true,
    })

    // 连接到服务器
    if err := wsClient.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer wsClient.Disconnect()

    // 订阅消息
    unsub := wsClient.Subscribe("notifications", func(data interface{}, topic string) {
        log.Printf("Received: %+v on topic: %s", data, topic)
    })
    defer unsub()

    // 发送消息
    wsClient.SendMessage("ping", "Hello Server!")

    // 保持连接
    time.Sleep(10 * time.Second)
}
```

### 高级配置

```go
options := client.SocketOptions{
    URL:               "ws://localhost:8080/ws",
    Token:             "your-auth-token",
    HeartbeatInterval: 30 * time.Second,
    Debug:             true,
    RefreshToken: func() (string, error) {
        // 实现你的token刷新逻辑
        newToken, err := refreshTokenFromAPI()
        return newToken, err
    },
    ErrorHandler: func(msg client.ErrorMsgData) {
        log.Printf("WebSocket error: %s - %s", msg.Code, msg.Detail)
    },
}

wsClient := client.NewSocketClient(options)
```

## 主题订阅

支持MQTT风格的主题匹配：

```go
// 精确匹配
wsClient.Subscribe("user/123/message", handler)

// 单层通配符 (+)
wsClient.Subscribe("user/+/message", handler) // 匹配 user/123/message, user/456/message

// 多层通配符 (#)
wsClient.Subscribe("system/#", handler) // 匹配 system/alert, system/alert/critical

// 监听状态变化
stateUnsub := wsClient.OnStateChange(func(state client.WebSocketState) {
    log.Printf("Connection state: %s", state.String())
})
defer stateUnsub()
```

## 双向通信频道

创建专用的双向通信频道：

```go
// 创建频道
channel, err := wsClient.CreateChannel("chat/room1",
    func(data interface{}) {
        log.Printf("Channel message: %+v", data)
    },
    func(reason client.ErrorMsgData) {
        log.Printf("Channel error: %s", reason.Detail)
    },
)
if err != nil {
    log.Fatalf("Failed to create channel: %v", err)
}

// 通过频道发送消息
channel.Send(map[string]interface{}{
    "type":    "chat",
    "message": "Hello from channel!",
})

// 设置关闭处理器
channel.OnClose(func(reason client.ErrorMsgData) {
    log.Printf("Channel closed: %s", reason.Detail)
})

// 关闭频道
defer channel.Close()
```

## 错误处理和重连

客户端具有内置的错误处理和自动重连机制：

```go
// 自动处理token过期
options.RefreshToken = func() (string, error) {
    // 调用你的API获取新token
    return getNewTokenFromAPI()
}

// 处理错误消息
options.ErrorHandler = func(msg client.ErrorMsgData) {
    switch msg.Code {
    case "RATE_LIMIT":
        log.Printf("Rate limited: %s", msg.Detail)
    case "INVALID_TOPIC":
        log.Printf("Invalid topic: %s", msg.Detail)
    default:
        log.Printf("Error: %s - %s", msg.Code, msg.Detail)
    }
}
```

## 连接状态

客户端支持以下连接状态：

- `Connecting` - 正在连接
- `Connected` - 已连接
- `Disconnected` - 已断开
- `Reconnecting` - 重连中
- `Error` - 错误状态

```go
// 检查当前状态
state := wsClient.State()
log.Printf("Current state: %s", state.String())

// 监听状态变化
unsub := wsClient.OnStateChange(func(state client.WebSocketState) {
    switch state {
    case client.Connected:
        log.Println("Connected to server")
    case client.Disconnected:
        log.Println("Disconnected from server")
    case client.Reconnecting:
        log.Println("Attempting to reconnect...")
    }
})
defer unsub()
```

## API 参考

### SocketClient 方法

- `Connect(token ...string) error` - 连接到WebSocket服务器
- `Disconnect()` - 断开连接
- `Subscribe(topic string, handler MessageHandler) UnsubscribeFunction` - 订阅主题
- `Unsubscribe(topic string, handler ...MessageHandler)` - 取消订阅
- `UnsubscribeAll()` - 取消所有订阅
- `SendMessage(topic string, data interface{}) error` - 发送消息
- `CreateChannel(topic string, handler ChannelMessageHandler, errHandler ...ChannelCloseHandler) (*Channel, error)` - 创建频道
- `State() WebSocketState` - 获取连接状态
- `OnStateChange(callback StateChangeCallback) UnsubscribeFunction` - 监听状态变化

### Channel 方法

- `Send(data interface{}) error` - 发送频道消息
- `Close()` - 关闭频道
- `OnClose(handler ChannelCloseHandler)` - 设置关闭处理器

## 线程安全

所有的客户端方法都是线程安全的，可以在多个goroutine中并发使用。

## 调试

启用调试日志来查看详细的连接和消息信息：

```go
options := client.SocketOptions{
    Debug: true, // 启用调试日志
    // ... 其他配置
}
```

## 注意事项

1. 确保在应用程序退出前调用 `Disconnect()` 方法
2. 记住调用 `UnsubscribeFunction` 来清理订阅
3. 频道使用完毕后记得调用 `Close()` 方法
4. 在生产环境中可以关闭调试日志以提高性能

## 完整示例

查看 `example.go` 文件获取更多使用示例。