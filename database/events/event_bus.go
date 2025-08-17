package events

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"entgo.io/ent"
)

// EventType 定义事件类型
type EventType string

const (
	EventTypePreCreate  EventType = "pre_create"
	EventTypePostCreate EventType = "post_create"
	EventTypePreUpdate  EventType = "pre_update"
	EventTypePostUpdate EventType = "post_update"
	EventTypePreDelete  EventType = "pre_delete"
	EventTypePostDelete EventType = "post_delete"
	EventTypePreMutate  EventType = "pre_mutate"
	EventTypePostMutate EventType = "post_mutate"
)

// Event 表示一个数据库事件
type Event struct {
	Type       EventType       // 事件类型
	EntityType string          // 实体类型 (如 "User", "Role" 等)
	Operation  ent.Op          // 操作类型 (Create, Update, Delete 等)
	Context    context.Context // 上下文
	Mutation   ent.Mutation    // 变更信息
	OldValue   ent.Value       // 变更前的值 (仅 post 事件有效)
	NewValue   ent.Value       // 变更后的值 (仅 post 事件有效)
	Error      error           // 错误信息 (仅 post 事件有效)
	Fields     map[string]any  // 变更的字段
	Timestamp  time.Time       // 事件时间
}

// EventHandler 事件处理器接口
type EventHandler interface {
	// Handle 处理事件，返回错误会阻止操作继续（仅对 pre 事件有效）
	Handle(ctx context.Context, event *Event) error
	// SupportsEvent 检查是否支持指定的事件类型和实体类型
	SupportsEvent(eventType EventType, entityType string, operation ent.Op) bool
}

// EventHandlerFunc 函数类型的事件处理器
type EventHandlerFunc func(ctx context.Context, event *Event) error

func (f EventHandlerFunc) Handle(ctx context.Context, event *Event) error {
	return f(ctx, event)
}

func (f EventHandlerFunc) SupportsEvent(eventType EventType, entityType string, operation ent.Op) bool {
	return true // 默认支持所有事件
}

// EventBus 事件总线
type EventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
	logger   LoggerInterface
}

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// NewEventBus 创建新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// SetLogger 设置日志记录器
func (eb *EventBus) SetLogger(logger LoggerInterface) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.logger = logger
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Info("Handler subscribed to event type: %s", eventType)
}

// SubscribeFunc 使用函数订阅事件
func (eb *EventBus) SubscribeFunc(eventType EventType, handlerFunc func(ctx context.Context, event *Event) error) {
	eb.Subscribe(eventType, EventHandlerFunc(handlerFunc))
}

// Unsubscribe 取消订阅事件（根据处理器类型）
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if reflect.DeepEqual(h, handler) {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			eb.logger.Info("Handler unsubscribed from event type: %s", eventType)
			break
		}
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	eb.mu.RUnlock()

	for _, handler := range handlers {
		if !handler.SupportsEvent(event.Type, event.EntityType, event.Operation) {
			continue
		}

		// 对于 pre 事件，如果处理器返回错误，则停止执行
		if err := handler.Handle(ctx, event); err != nil {
			eb.logger.Error("Event handler failed for %s:%s - %v",
				event.Type, event.EntityType, err)

			// Pre 事件的错误会阻止操作继续
			if isPreEvent(event.Type) {
				return fmt.Errorf("event handler failed: %w", err)
			}
			// Post 事件的错误只记录，不阻止
		}
	}

	return nil
}

// PublishSync 同步发布事件（阻塞直到所有处理器完成）
func (eb *EventBus) PublishSync(ctx context.Context, event *Event) error {
	return eb.Publish(ctx, event)
}

// PublishAsync 异步发布事件（非阻塞）
func (eb *EventBus) PublishAsync(ctx context.Context, event *Event) {
	go func() {
		if err := eb.Publish(ctx, event); err != nil {
			eb.logger.Error("Async event publication failed: %v", err)
		}
	}()
}

// GetHandlerCount 获取指定事件类型的处理器数量
func (eb *EventBus) GetHandlerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.handlers[eventType])
}

// Clear 清除所有事件处理器
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers = make(map[EventType][]EventHandler)
	eb.logger.Info("All event handlers cleared")
}

// isPreEvent 检查是否为预处理事件
func isPreEvent(eventType EventType) bool {
	return eventType == EventTypePreCreate ||
		eventType == EventTypePreUpdate ||
		eventType == EventTypePreDelete ||
		eventType == EventTypePreMutate
}

// extractEntityType 从 mutation 中提取实体类型
func extractEntityType(m ent.Mutation) string {
	return m.Type()
}

// extractFields 从 mutation 中提取变更的字段
func extractFields(m ent.Mutation) map[string]any {
	fields := make(map[string]any)

	// 获取变更的字段名
	fieldNames := m.Fields()
	for _, fieldName := range fieldNames {
		if value, exists := m.Field(fieldName); exists {
			fields[fieldName] = value
		}
	}

	return fields
}

// NewEvent 创建新事件
func NewEvent(eventType EventType, ctx context.Context, m ent.Mutation) *Event {
	return &Event{
		Type:       eventType,
		EntityType: extractEntityType(m),
		Operation:  m.Op(),
		Context:    ctx,
		Mutation:   m,
		Fields:     extractFields(m),
		Timestamp:  time.Now(),
	}
}

// NewEventWithValue 创建带值的新事件（用于 post 事件）
func NewEventWithValue(eventType EventType, ctx context.Context, m ent.Mutation, oldValue, newValue ent.Value, err error) *Event {
	event := NewEvent(eventType, ctx, m)
	event.OldValue = oldValue
	event.NewValue = newValue
	event.Error = err
	return event
}
