package events

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"
)

// EventMixin 事件驱动的mixin，可以被其他schema继承
type EventMixin struct {
	mixin.Schema
}

// Hooks 返回事件驱动的钩子
func (EventMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		// Pre-mutation hook - 在变更前触发
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// 发布 pre 事件
				preEvent := NewEvent(getPreEventType(m.Op()), ctx, m)
				if err := Publish(ctx, preEvent); err != nil {
					return nil, err
				}

				// 执行实际的变更
				value, err := next.Mutate(ctx, m)

				// 发布 post 事件
				postEvent := NewEventWithValue(getPostEventType(m.Op()), ctx, m, nil, value, err)
				// Post 事件异步执行，不阻塞主流程
				PublishAsync(ctx, postEvent)

				return value, err
			})
		},
	}
}

// getPreEventType 根据操作类型获取对应的 pre 事件类型
func getPreEventType(op ent.Op) EventType {
	switch op {
	case ent.OpCreate:
		return EventTypePreCreate
	case ent.OpUpdate, ent.OpUpdateOne:
		return EventTypePreUpdate
	case ent.OpDelete, ent.OpDeleteOne:
		return EventTypePreDelete
	default:
		return EventTypePreMutate
	}
}

// getPostEventType 根据操作类型获取对应的 post 事件类型
func getPostEventType(op ent.Op) EventType {
	switch op {
	case ent.OpCreate:
		return EventTypePostCreate
	case ent.OpUpdate, ent.OpUpdateOne:
		return EventTypePostUpdate
	case ent.OpDelete, ent.OpDeleteOne:
		return EventTypePostDelete
	default:
		return EventTypePostMutate
	}
}

// ConditionalEventHandler 条件事件处理器，只处理特定条件的事件
type ConditionalEventHandler struct {
	Handler     EventHandlerFunc
	EventTypes  []EventType
	EntityTypes []string
	Operations  []ent.Op
}

// Handle 实现 EventHandler 接口
func (h ConditionalEventHandler) Handle(ctx context.Context, event *Event) error {
	return h.Handler(ctx, event)
}

// SupportsEvent 检查是否支持指定的事件
func (h ConditionalEventHandler) SupportsEvent(eventType EventType, entityType string, operation ent.Op) bool {
	// 检查事件类型
	if len(h.EventTypes) > 0 {
		found := false
		for _, et := range h.EventTypes {
			if et == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查实体类型
	if len(h.EntityTypes) > 0 {
		found := false
		for _, et := range h.EntityTypes {
			if et == entityType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查操作类型
	if len(h.Operations) > 0 {
		found := false
		for _, op := range h.Operations {
			if op == operation {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// NewConditionalHandler 创建条件事件处理器
func NewConditionalHandler(
	handler EventHandlerFunc,
	eventTypes []EventType,
	entityTypes []string,
	operations []ent.Op,
) *ConditionalEventHandler {
	return &ConditionalEventHandler{
		Handler:     handler,
		EventTypes:  eventTypes,
		EntityTypes: entityTypes,
		Operations:  operations,
	}
}
