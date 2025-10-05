package messaging

import "time"

type MessageType string // 消息类型，也代表这个消息的处理者是谁，需要实现对应的消息发送和处理器

const (
	ServerToUserSocket MessageType = "socket.user"    // 发送给用户的websocket消息
	UserToServerSocket MessageType = "socket.server"  // 用户通过websocket发送给服务器的消息
	ChannelToServer    MessageType = "channel.server" // 发送给频道处理器的消息
	ChannelToUser      MessageType = "channel.user"   // 发送给频道用户的消息
	ServerToWorker     MessageType = "worker"         // 发送给后台任务处理器的消息
)

type MessageStruct struct {
	id        string    `msgpack:"id"` // 这条消息的唯一标识
	createdAt time.Time `msgpack:"created_at"`

	Type     MessageType `msgpack:"type"`     // 消息类型
	Payload  any         `msgpack:"payload"`  // 根据消息的类型不同,这里会是不同的Payload结构体
	Priority int         `msgpack:"priority"` // 优先级，数字越大优先级越高
}

type SocketMessagePayload struct {
	UserId *uint64 `msgpack:"user_id" json:"user_id"` // 接收消息的用户ID, 如果为空则表示发送给所有用户
	Topic  string  `msgpack:"topic" json:"topic"`     // 订阅的主题
	Data   any     `msgpack:"data" json:"data"`       // 具体消息内容
}

type UserMessagePayload struct {
	MessageId string `msgpack:"message_id" json:"message_id"` // 这条消息的唯一标识, 用于回复
	UserId    uint64 `msgpack:"user_id" json:"user_id"`       // 消息的用户ID
	ClientId  uint64 `msgpack:"client_id" json:"client_id"`   // 消息的客户端ID
	Data      any    `msgpack:"data" json:"data"`             // 具体消息内容
}

type ChannelAction int

const (
	ChannelActionCreate ChannelAction = iota
	ChannelActionMsg
	ChannelActionClose
)

type ChannelMessagePayLoad struct {
	ID     string        `msgpack:"id" json:"id"` // 频道ID
	Topic  string        `msgpack:"topic,omitempty" json:"topic,omitempty"`
	UserID uint64        `msgpack:"user_id,omitempty" json:"user_id,omitempty"`
	Action ChannelAction `msgpack:"action,omitempty" json:"action,omitempty"`
	Data   any           `msgpack:"data" json:"data"`
}
