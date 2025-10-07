package client

import "time"

// WebSocketState WebSocket 连接状态枚举
type WebSocketState int

const (
	Connecting   WebSocketState = iota // 连接中
	Connected                          // 已连接
	Disconnected                       // 已断开
	Reconnecting                       // 重连中
	Error                              // 错误状态
)

// String 返回状态的字符串表示
func (s WebSocketState) String() string {
	switch s {
	case Connecting:
		return "connecting"
	case Connected:
		return "connected"
	case Disconnected:
		return "disconnected"
	case Reconnecting:
		return "reconnecting"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// ClientMessage 客户端发送的消息结构
type ClientMessage struct {
	Action string      `json:"action"` // subscribe, unsubscribe, ping, pong, msg, channel_start, channel, channel_close
	Topic  string      `json:"topic"`
	Data   interface{} `json:"data,omitempty"`
}

// SocketMessagePayload 服务器发送的消息结构
type SocketMessagePayload struct {
	Topic     string      `json:"topic"`
	UserID    uint64      `json:"userId"`
	Data      interface{} `json:"data"`
	Timestamp uint64      `json:"timestamp,omitempty"`
}

// MessageHandler 消息处理器函数类型
type MessageHandler func(data interface{}, topic string)

// HandlerWrapper 处理器包装器，用于安全地管理和标识处理器
type HandlerWrapper struct {
	ID      string
	Handler MessageHandler
}

// StateChangeCallback 状态变化回调函数类型
type StateChangeCallback func(state WebSocketState)

// UnsubscribeFunction 取消订阅函数类型
type UnsubscribeFunction func()

// RefreshTokenFunction token刷新函数类型
type RefreshTokenFunction func() (string, error)

// ErrorHandler 错误处理函数类型
type ErrorHandler func(msg ErrorMsgData)

// SocketOptions WebSocket 客户端配置选项
type SocketOptions struct {
	URL               string               // WebSocket服务器地址
	Token             string               // 认证token
	HeartbeatInterval time.Duration        // 心跳间隔，默认30秒
	Debug             bool                 // 是否开启调试日志
	RefreshToken      RefreshTokenFunction // token刷新函数
	ErrorHandler      ErrorHandler         // 错误处理函数
}

// SubscriptionRecord 内部订阅记录
type SubscriptionRecord struct {
	Topic          string
	HandlerWrapper *HandlerWrapper
	ID             string
}

// ChannelOpenRecord 频道开放处理器记录
type ChannelOpenRecord struct {
	Topic   string
	Handler ChannelHandler
	ID      string
}

// DisConnectMsg 服务器断开连接消息
type DisConnectMsg struct {
	Topic string `json:"topic"` // "?dc"
	Data  struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	} `json:"data"`
}

// ErrorMsg 服务器错误消息
type ErrorMsg struct {
	Topic string       `json:"topic"` // "?er"
	Data  ErrorMsgData `json:"data"`
}

// ErrorMsgData 错误消息数据
type ErrorMsgData struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// ChannelCreateRes 频道创建响应
type ChannelCreateRes struct {
	ChannelID *string       `json:"channelId,omitempty"`
	Error     *ErrorMsgData `json:"error,omitempty"`
}

// Channel 频道相关类型
type ChannelHandler func(channel Channel)
type ChannelSender func(data interface{}) error
type ChannelMessageHandler func(data interface{})
type ChannelCloser func() <-chan struct{}
type ChannelCloseHandler func(reason ErrorMsgData)
type ChannelWaiter func() <-chan struct{}

// Channel 频道实例
type Channel struct {
	id    string
	topic string

	Send    ChannelSender
	Close   ChannelCloser
	Wait    ChannelWaiter
	OnClose func(handler ChannelCloseHandler)
}

func (c Channel) ID() string {
	return c.id
}

func (c Channel) Topic() string {
	return c.topic
}

// ISocketClient WebSocket 客户端接口
type ISocketClient interface {
	// State 获取当前连接状态
	State() WebSocketState

	// Connect 连接到WebSocket服务器
	Connect(token ...string) error

	// Disconnect 断开连接
	Disconnect() <-chan struct{}

	// Subscribe 订阅主题
	Subscribe(topic string, handler MessageHandler) UnsubscribeFunction

	// Unsubscribe 取消订阅
	Unsubscribe(topic string, handler ...MessageHandler)

	// UnsubscribeAll 取消所有订阅
	UnsubscribeAll()

	// OnStateChange 监听连接状态变化
	OnStateChange(callback StateChangeCallback) UnsubscribeFunction

	// SendMessage 发送消息
	SendMessage(topic string, data interface{}) error

	// CreateChannel 创建频道
	CreateChannel(topic string, handler ChannelMessageHandler, errHandler ...ChannelCloseHandler) (*Channel, error)

	// RegisterChannelOpen 注册频道开放处理器，当服务器主动创建频道时会调用
	RegisterChannelOpen(topic string, handler ChannelHandler) UnsubscribeFunction
}
