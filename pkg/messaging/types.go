package messaging

import "time"

type MessageType string // 消息类型，也代表这个消息的处理者是谁，需要实现对应的消息发送和处理器

const (
	ServerToUserSocket MessageType = "socket.user"        // 发送给用户的websocket消息
	UserToServerSocket MessageType = "socket.server"      // 用户通过websocket发送给服务器的消息
	ChannelToServer    MessageType = "channel.server"     // 发送给频道处理器的消息
	ChannelToUser      MessageType = "channel.user"       // 发送给频道用户的消息
	ServerToWorker     MessageType = "worker"             // 发送给后台任务处理器的消息
	ChannelOpenCheck   MessageType = "channel.open_check" // 请求创建频道的消息
	ChannelOpenRes     MessageType = "channel.open_res"   // 频道创建结果的响应
	SubscribeCheck     MessageType = "subscribe.check"    // 订阅频道的权限检查
	SubscribeRes       MessageType = "subscribe.res"      // 订阅频道的权限检查结果
)

type TopicPayload interface {
}

type MessageStruct struct {
	id        string    `msgpack:"id"` // 这条消息的唯一标识
	createdAt time.Time `msgpack:"created_at"`

	Type     MessageType  `msgpack:"type"`     // 消息类型
	Payload  TopicPayload `msgpack:"payload"`  // 根据消息的类型不同,这里会是不同的Payload结构体
	Priority int          `msgpack:"priority"` // 优先级，数字越大优先级越高
}

type ChannelOpenCheckPayload struct {
	ChannelID string `msgpack:"channel_id" json:"channel_id"` // 频道ID
	Topic     string `msgpack:"topic" json:"topic"`           // 频道主题
	UserID    uint64 `msgpack:"user_id" json:"user_id"`       // 频道所属用户ID
	SessionId string `msgpack:"session_id" json:"session_id"` // 频道所属会话ID
	ClientId  uint64 `msgpack:"client_id" json:"client_id"`   // 频道所属客户端ID
	Allowed   bool   `msgpack:"allowed" json:"allowed"`       // 是否允许创建频道
	Timestamp int64  `msgpack:"timestamp" json:"timestamp"`   // 消息发送的时间戳, 用于判断是否超时
}

type SubscribeCheckPayload struct {
	Topic     string `msgpack:"topic" json:"topic"`           // 频道主题
	UserID    uint64 `msgpack:"user_id" json:"user_id"`       // 频道所属用户ID
	SessionId string `msgpack:"session_id" json:"session_id"` // 频道所属会话ID
	ClientId  uint64 `msgpack:"client_id" json:"client_id"`   // 频道所属客户端ID
	Allowed   bool   `msgpack:"allowed" json:"allowed"`       // 是否允许订阅频道
	Timestamp int64  `msgpack:"timestamp" json:"timestamp"`   // 消息发送的时间戳, 用于判断是否超时
}

type SocketMessagePayload struct {
	UserId *uint64 `msgpack:"user_id" json:"user_id"` // 接收消息的用户ID, 如果为空则表示发送给所有用户
	Topic  string  `msgpack:"topic" json:"topic"`     // 订阅的主题
	// 注意！！ 因为分为了两个服务，在传递之后json标签会丢失，若需要传输struct，请先转换成map格式
	Data any `msgpack:"data" json:"data"` // 具体消息内容
}

type UserMessagePayload struct {
	Topic    string `msgpack:"message_id" json:"message_id"` // 这条消息的唯一标识, 用于回复
	UserId   uint64 `msgpack:"user_id" json:"user_id"`       // 消息的用户ID
	ClientId uint64 `msgpack:"client_id" json:"client_id"`   // 消息的客户端ID
	Data     any    `msgpack:"data" json:"data"`             // 具体消息内容
}

type ChannelMessagePayLoad struct {
	ID        string        `msgpack:"id" json:"id"` // 频道ID
	Topic     string        `msgpack:"topic,omitempty" json:"topic,omitempty"`
	UserID    uint64        `msgpack:"user_id,omitempty" json:"user_id,omitempty"`
	SessionId string        `msgpack:"session_id,omitempty" json:"session_id,omitempty"`
	ClientId  uint64        `msgpack:"client_id,omitempty" json:"client_id,omitempty"`
	Action    ChannelAction `msgpack:"action,omitempty" json:"action,omitempty"`
	Data      any           `msgpack:"data" json:"data"`
}

type ChannelAction int

const (
	ChannelActionCreate ChannelAction = iota
	ChannelActionMsg
	ChannelActionErr
	ChannelActionClose
)
