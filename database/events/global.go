package events

import (
	"context"
	"sync"
)

var (
	// globalEventBus 全局事件总线实例
	globalEventBus *EventBus
	// once 确保全局事件总线只初始化一次
	once sync.Once
)

// GlobalEventBus 获取全局事件总线实例（单例模式）
func GlobalEventBus() *EventBus {
	once.Do(func() {
		globalEventBus = NewEventBus()
	})
	return globalEventBus
}

// Subscribe 订阅全局事件
func Subscribe(eventType EventType, handler EventHandler) {
	GlobalEventBus().Subscribe(eventType, handler)
}

// SubscribeFunc 使用函数订阅全局事件
func SubscribeFunc(eventType EventType, handlerFunc func(ctx context.Context, event *Event) error) {
	GlobalEventBus().SubscribeFunc(eventType, handlerFunc)
}

// Unsubscribe 取消订阅全局事件
func Unsubscribe(eventType EventType, handler EventHandler) {
	GlobalEventBus().Unsubscribe(eventType, handler)
}

// Publish 发布全局事件
func Publish(ctx context.Context, event *Event) error {
	return GlobalEventBus().Publish(ctx, event)
}

// PublishSync 同步发布全局事件
func PublishSync(ctx context.Context, event *Event) error {
	return GlobalEventBus().PublishSync(ctx, event)
}

// PublishAsync 异步发布全局事件
func PublishAsync(ctx context.Context, event *Event) {
	GlobalEventBus().PublishAsync(ctx, event)
}

// SetLogger 设置全局事件总线的日志记录器
func SetLogger(logger LoggerInterface) {
	GlobalEventBus().SetLogger(logger)
}

// GetHandlerCount 获取全局事件总线中指定事件类型的处理器数量
func GetHandlerCount(eventType EventType) int {
	return GlobalEventBus().GetHandlerCount(eventType)
}

// Clear 清除全局事件总线中的所有事件处理器
func Clear() {
	GlobalEventBus().Clear()
}
