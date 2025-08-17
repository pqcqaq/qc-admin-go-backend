# 支持 Select 事件的实现

## 概述

现在已经实现了对 ENT 查询操作（包括 Select）的事件支持。该实现包含：

1. **EventMixin** - 用于处理 mutation 事件（Create、Update、Delete）
2. **QueryEventMixin** - 用于处理查询事件（Select、First、All 等）

## 使用方法

### 1. 在 Schema 中添加 Mixin

```go
package schema

import (
    "go-backend/database/events"
    "entgo.io/ent"
    "entgo.io/ent/schema"
)

// User holds the schema definition for the User entity.
type User struct {
    ent.Schema
}

// Mixin returns User mixed-in fields.
func (User) Mixin() []ent.Mixin {
    return []ent.Mixin{
        // 用于 mutation 事件（Create、Update、Delete）
        events.EventMixin{},
        
        // 用于查询事件（Select、First、All 等）
        events.QueryEventMixin{},
        
        // 其他 mixin...
    }
}
```

### 2. 注册事件处理器

```go
package main

import (
    "context"
    "log"
    "go-backend/database/events"
)

func main() {
    // 注册查询事件处理器
    events.Register(&events.EventHandlerFunc{
        handler: func(ctx context.Context, event *events.Event) error {
            switch event.Type {
            case events.EventTypePreSelect:
                log.Printf("准备执行查询: 实体=%s, 操作=%v", event.EntityType, event.Operation)
                // 可以在这里添加查询前的逻辑，比如权限检查、参数验证等
                
            case events.EventTypePostSelect:
                log.Printf("查询完成: 实体=%s, 错误=%v", event.EntityType, event.Error)
                // 可以在这里添加查询后的逻辑，比如日志记录、缓存等
            }
            return nil
        },
        supportedEvents: events.SupportedEvents{
            "User": events.SupportedEntityEvents{
                EventTypes: []events.EventType{
                    events.EventTypePreSelect,
                    events.EventTypePostSelect,
                },
                Operations: []ent.Op{events.OpSelect},
            },
        },
        name: "QueryEventHandler",
    })
}
```

### 3. 使用条件事件处理器

```go
// 只处理特定实体的查询事件
handler := events.NewConditionalHandler(
    events.EventHandlerFunc{
        handler: func(ctx context.Context, event *events.Event) error {
            if event.Type == events.EventTypePreSelect {
                // 在查询前添加默认过滤条件
                log.Printf("拦截查询操作: %s", event.EntityType)
            }
            return nil
        },
    },
    []events.EventType{events.EventTypePreSelect, events.EventTypePostSelect},
    []string{"User", "Role"},      // 只处理 User 和 Role 实体
    []ent.Op{events.OpSelect},     // 只处理查询操作
)

events.Register(handler)
```

## 事件类型

### Mutation 事件（通过 EventMixin）
- `EventTypePreCreate` - Create 操作前
- `EventTypePostCreate` - Create 操作后
- `EventTypePreUpdate` - Update 操作前
- `EventTypePostUpdate` - Update 操作后
- `EventTypePreDelete` - Delete 操作前
- `EventTypePostDelete` - Delete 操作后

### 查询事件（通过 QueryEventMixin）
- `EventTypePreSelect` - 查询操作前
- `EventTypePostSelect` - 查询操作后

## 注意事项

1. **功能要求**: 项目必须启用 `intercept` 功能标志（已启用）
2. **导入 runtime**: 如果使用 Schema 级别的拦截器，需要导入 `ent/runtime` 包
3. **异步处理**: Post 事件默认异步处理，不会阻塞主流程
4. **错误处理**: Pre 事件中的错误会阻止操作继续执行

## 高级用法

### 查询权限控制
```go
events.Register(&events.EventHandlerFunc{
    handler: func(ctx context.Context, event *events.Event) error {
        if event.Type == events.EventTypePreSelect {
            // 检查用户权限
            user := getUserFromContext(ctx)
            if !user.HasPermission(event.EntityType, "read") {
                return fmt.Errorf("没有读取 %s 的权限", event.EntityType)
            }
        }
        return nil
    },
    // ... 配置支持的事件
})
```

### 查询日志记录
```go
events.Register(&events.EventHandlerFunc{
    handler: func(ctx context.Context, event *events.Event) error {
        if event.Type == events.EventTypePostSelect {
            duration := time.Since(event.Timestamp)
            log.Printf("查询性能: 实体=%s, 耗时=%v", event.EntityType, duration)
        }
        return nil
    },
    // ... 配置支持的事件
})
```

这样，你的应用就完全支持了 Select 和其他查询操作的事件处理！
