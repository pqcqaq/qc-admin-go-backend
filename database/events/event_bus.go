package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent"
)

// EventType 定义事件类型
type EventType string

const (
	EventTypePreSelect  EventType = "pre_select"
	EventTypePostSelect EventType = "post_select"

	EventTypePreCreate  EventType = "pre_create"
	EventTypePostCreate EventType = "post_create"

	EventTypePreUpdate  EventType = "pre_update"
	EventTypePostUpdate EventType = "post_update"

	EventTypePreDelete  EventType = "pre_delete"
	EventTypePostDelete EventType = "post_delete"

	EventTypePreMutate  EventType = "pre_mutate"
	EventTypePostMutate EventType = "post_mutate"
)

// 查询操作的特殊操作类型
const (
	OpSelect ent.Op = 999 // 使用一个特殊的数值表示查询操作
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

// SupportedEvents 定义处理器支持的事件类型
type SupportedEvents map[string]SupportedEntityEvents

// SupportedEntityEvents 定义单个实体支持的事件和操作
type SupportedEntityEvents struct {
	EventTypes []EventType `json:"eventType"`
	Operations []ent.Op    `json:"operation"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	// Handle 处理事件，返回错误会阻止操作继续（仅对 pre 事件有效）
	Handle(ctx context.Context, event *Event) error
	// SupportsEvent 返回支持的事件配置
	SupportsEvent() SupportedEvents
	// Name 可选 处理器名称
	Name() string
}

// EventHandlerFunc 函数类型的事件处理器
type EventHandlerFunc struct {
	handler         func(ctx context.Context, event *Event) error
	supportedEvents SupportedEvents
	name            string
}

func (f *EventHandlerFunc) Handle(ctx context.Context, event *Event) error {
	return f.handler(ctx, event)
}

func (f *EventHandlerFunc) SupportsEvent() SupportedEvents {
	return f.supportedEvents
}

func (f *EventHandlerFunc) Name() string {
	return f.name
}

// NewEventHandlerFunc 创建一个新的EventHandlerFunc
func NewEventHandlerFunc(handler func(ctx context.Context, event *Event) error, supportedEvents SupportedEvents) *EventHandlerFunc {
	return &EventHandlerFunc{
		handler:         handler,
		supportedEvents: supportedEvents,
		name:            "AnonymousHandler",
	}
}

// HandlerRegistry 处理器注册信息
type HandlerRegistry struct {
	handler    EventHandler
	eventTypes map[EventType]bool
	operations map[ent.Op]bool
}

// EventBus 事件总线
type EventBus struct {
	// 按实体类型和事件类型分类存储处理器
	handlers map[string]map[EventType][]*HandlerRegistry
	mu       sync.RWMutex
	logger   LoggerInterface
}

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// NewEventBus 创建新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string]map[EventType][]*HandlerRegistry),
	}
}

// SetLogger 设置日志记录器
func (eb *EventBus) SetLogger(logger LoggerInterface) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.logger = logger
}

// Register 注册事件处理器
func (eb *EventBus) Register(handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	supportedEvents := handler.SupportsEvent()

	for entityType, entityEvents := range supportedEvents {
		if eb.handlers[entityType] == nil {
			eb.handlers[entityType] = make(map[EventType][]*HandlerRegistry)
		}

		// 创建处理器注册信息
		registry := &HandlerRegistry{
			handler:    handler,
			eventTypes: make(map[EventType]bool),
			operations: make(map[ent.Op]bool),
		}

		// 设置支持的事件类型
		for _, eventType := range entityEvents.EventTypes {
			registry.eventTypes[eventType] = true
		}

		// 设置支持的操作类型
		for _, operation := range entityEvents.Operations {
			registry.operations[operation] = true
		}

		// 为每种支持的事件类型注册处理器
		for _, eventType := range entityEvents.EventTypes {
			eb.handlers[entityType][eventType] = append(eb.handlers[entityType][eventType], registry)
		}
	}

	if eb.logger != nil {
		eb.logger.Info("Handler registered for entities: %v", supportedEvents)
	}

	name := handler.Name()
	if name != "" {
		eb.logger.Info("Handler %s registered for entities: %v", name, supportedEvents)
	} else {
		eb.logger.Info("Anonymous handler registered for entities: %v", supportedEvents)
	}
}

// Subscribe 订阅事件（向后兼容的方法）
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.Register(handler)
}

// SubscribeFunc 使用函数订阅事件（向后兼容的方法）
func (eb *EventBus) SubscribeFunc(eventType EventType, handlerFunc func(ctx context.Context, event *Event) error) {
	// 创建一个支持所有实体和操作的处理器
	supportedEvents := SupportedEvents{
		"*": SupportedEntityEvents{
			EventTypes: []EventType{eventType},
			Operations: []ent.Op{ent.OpCreate, ent.OpUpdate, ent.OpUpdateOne, ent.OpDelete, ent.OpDeleteOne},
		},
	}

	handlerWrapper := NewEventHandlerFunc(handlerFunc, supportedEvents)
	eb.Register(handlerWrapper)
}

// Unsubscribe 取消订阅事件（根据处理器类型）
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	supportedEvents := handler.SupportsEvent()

	for entityType, entityEvents := range supportedEvents {
		if entityHandlers, exists := eb.handlers[entityType]; exists {
			for _, evType := range entityEvents.EventTypes {
				if registries, exists := entityHandlers[evType]; exists {
					for i, registry := range registries {
						if registry.handler == handler {
							eb.handlers[entityType][evType] = append(registries[:i], registries[i+1:]...)
							if eb.logger != nil {
								eb.logger.Info("Handler unsubscribed from event type: %s, entity: %s", evType, entityType)
							}
							break
						}
					}
				}
			}
		}
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	eb.mu.RLock()

	var registries []*HandlerRegistry

	// 获取特定实体类型的处理器
	if entityHandlers, exists := eb.handlers[event.EntityType]; exists {
		if eventHandlers, exists := entityHandlers[event.Type]; exists {
			registries = append(registries, eventHandlers...)
		}
	}

	// 获取通用处理器（支持所有实体类型）
	if universalHandlers, exists := eb.handlers["*"]; exists {
		if eventHandlers, exists := universalHandlers[event.Type]; exists {
			registries = append(registries, eventHandlers...)
		}
	}

	eb.mu.RUnlock()

	for _, registry := range registries {
		// 检查操作类型是否支持
		if !registry.operations[event.Operation] {
			continue
		}

		eb.logger.Debug("Publishing event %s for entity %s with operation %s",
			event.Type, event.EntityType, event.Operation)

		// 对于 pre 事件，如果处理器返回错误，则停止执行
		if err := registry.handler.Handle(ctx, event); err != nil {
			if eb.logger != nil {
				eb.logger.Error("Event handler failed for %s:%s - %v",
					event.Type, event.EntityType, err)
			}

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

	count := 0
	for _, entityHandlers := range eb.handlers {
		if handlers, exists := entityHandlers[eventType]; exists {
			count += len(handlers)
		}
	}
	return count
}

// Clear 清除所有事件处理器
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers = make(map[string]map[EventType][]*HandlerRegistry)
	if eb.logger != nil {
		eb.logger.Info("All event handlers cleared")
	}
}

// isPreEvent 检查是否为预处理事件
func isPreEvent(eventType EventType) bool {
	return eventType == EventTypePreSelect ||
		eventType == EventTypePreCreate ||
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
