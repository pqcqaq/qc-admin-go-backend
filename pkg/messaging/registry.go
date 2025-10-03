package messaging

import "sync"

// Message 发送和处理的注册器

type MessageHandler func(message MessageStruct) error
type HandlerRemover func() // 用于移除注册的处理器

var (
	handlers = make(map[MessageType][]*handlerWrapper)
	mu       sync.RWMutex
)

// handlerWrapper 包装 handler,使其可以通过 ID 识别
type handlerWrapper struct {
	id      int64
	handler MessageHandler
}

var handlerIDCounter int64

func RegisterHandler(messageType MessageType, handler MessageHandler) HandlerRemover {
	mu.Lock()
	defer mu.Unlock()

	// 生成唯一 ID
	handlerIDCounter++
	id := handlerIDCounter

	wrapper := &handlerWrapper{
		id:      id,
		handler: handler,
	}

	handlers[messageType] = append(handlers[messageType], wrapper)

	// 返回移除函数
	return func() {
		mu.Lock()
		defer mu.Unlock()

		hs := handlers[messageType]
		for i, h := range hs {
			if h.id == id {
				handlers[messageType] = append(hs[:i], hs[i+1:]...)
				break
			}
		}
	}
}

// GetHandlers 获取指定类型的所有处理器(用于消息分发)
func GetHandlers(messageType MessageType) []MessageHandler {
	mu.RLock()
	defer mu.RUnlock()

	wrappers := handlers[messageType]
	result := make([]MessageHandler, len(wrappers))
	for i, w := range wrappers {
		result[i] = w.handler
	}
	return result
}
