package websocket

import (
	"context"
	"go-backend/pkg/websocket/channel"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Logger interface {
	Info(format string, v ...any)
	Error(format string, v ...any)
	Debug(format string, v ...any)
	Warn(format string, v ...any)
}

var logger Logger

func SetLogger(l Logger) {
	logger = l
}

// 关键的数据结构：
//  全部客户端记录,
//  用户id -> 客户端列表映射（一个用户可以有多个客户端同时连接）
//  客户端id -> 订阅的频道列表映射

type ClientMessage struct {
	Action string `json:"action"`          // "subscribe" or "unsubscribe"
	Topic  string `json:"topic,omitempty"` // 频道名称
	Data   any    `json:"data,omitempty"`  // 消息内容（仅在发布消息时使用）
}

type WsServer struct {
	// 操作全局结构体的锁
	cMu  sync.Mutex
	uCMu sync.Mutex
	cSMu sync.Mutex
	cCMu sync.Mutex
	cIMu sync.Mutex

	// 发送消息的ctx
	sendCtx context.Context

	allowList []string
	allowAll  bool

	connectedClients    map[*ClientConnWrapper]bool            // 连接的客户端
	userClients         map[uint64]map[*ClientConnWrapper]bool // 用户ID -> 客户端列表映射
	clientSubscriptions map[*ClientConnWrapper]map[string]bool // 客户端 -> 订阅的频道列表映射
	channelsClient      map[string]map[*ClientConnWrapper]bool // 频道ID -> 订阅的客户端列表映射

	// channel的ID必然不可能是重复的,因为它是由用户ID,客户端ID,和频道主题三部分组成,而且本身就不可能重复创建
	channelIdMap map[string]*channel.Channel

	// channel factory
	channelFactory *channel.ChannelFactory
}

type WsServerOptions struct {
	AllowOrigins   []string
	ChannelFactory *channel.ChannelFactory
}

type ClientConnWrapper struct {
	id       string
	Conn     *websocket.Conn
	UserId   uint64
	ClientId uint64
	lastPong time.Time

	channels map[string]*channel.Channel

	// channels Lock
	cMu sync.Mutex
	// lastPong Lock
	lastPongMu sync.RWMutex
	// WebSocket write Lock
	writeMu sync.Mutex
}
