package messaging

import "time"

type MessageType string // 消息类型，也代表这个消息的处理者是谁，需要实现对应的消息发送和处理器

const (
	ToUserSocket MessageType = "socket.user" // 发送给用户的websocket消息
	ToWorker     MessageType = "worker"      // 发送给后台任务处理器的消息
)

type MessageStruct struct {
	id        string    `msgpack:"id"` // 这条消息的唯一标识
	createdAt time.Time `msgpack:"created_at"`

	Type     MessageType `msgpack:"type"`     // 消息类型
	Payload  any         `msgpack:"payload"`  // 根据消息的类型不同,这里会是不同的Payload结构体
	Priority int         `msgpack:"priority"` // 优先级，数字越大优先级越高
}
