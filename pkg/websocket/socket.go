package websocket

import (
	"context"
	"fmt"
	"go-backend/pkg/configs"
	"go-backend/pkg/jwt"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
	"go-backend/pkg/websocket/channel"
	"go-backend/pkg/websocket/types"
	"net/http"
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

func (s *WsServer) GetChannelById(channelId string) *channel.Channel {
	s.cIMu.Lock()
	defer s.cIMu.Unlock()
	return s.channelIdMap[channelId]
}

type WsServerOptions struct {
	AllowOrigins   []string
	ChannelFactory *channel.ChannelFactory
}

func NewWsServer(options WsServerOptions) *WsServer {
	allowList := options.AllowOrigins
	allowAll := len(allowList) == 0
	return &WsServer{
		allowList:           allowList,
		allowAll:            allowAll,
		connectedClients:    make(map[*ClientConnWrapper]bool),
		userClients:         make(map[uint64]map[*ClientConnWrapper]bool),
		clientSubscriptions: make(map[*ClientConnWrapper]map[string]bool),
		channelsClient:      make(map[string]map[*ClientConnWrapper]bool),
		channelIdMap:        make(map[string]*channel.Channel),
		sendCtx:             context.Background(),
		channelFactory:      options.ChannelFactory,
	}
}

func (s *WsServer) LockAll() {
	s.cMu.Lock()
	s.uCMu.Lock()
	s.cSMu.Lock()
}

func (s *WsServer) UnlockAll() {
	s.cSMu.Unlock()
	s.uCMu.Unlock()
	s.cMu.Unlock()
}

// cleanupExpiredClients 清理超时的客户端连接
func (s *WsServer) cleanupExpiredClients() {
	timeoutDuration := time.Duration(configs.GetConfig().Socket.PingTimeout) * time.Second
	now := time.Now()

	s.LockAll()
	defer s.UnlockAll()

	var expiredClients []*ClientConnWrapper
	for client := range s.connectedClients {
		if now.Sub(client.GetLastPong()) > timeoutDuration {
			expiredClients = append(expiredClients, client)
		}
	}

	// 清理过期的客户端
	for _, client := range expiredClients {
		// 关闭WebSocket连接
		client.Conn.Close()

		// 使用统一的清理方法
		s.removeClient(client, "ping timeout")
	}
}

// startCleanupRoutine 启动定期清理过期客户端的goroutine
func (s *WsServer) startCleanupRoutine() {
	cleanupInterval := time.Duration(configs.GetConfig().Socket.PingTimeout/2) * time.Second
	if cleanupInterval < 10*time.Second {
		cleanupInterval = 10 * time.Second // 最小清理间隔为10秒
	}

	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for range ticker.C {
			s.cleanupExpiredClients()
		}
	}()

	logger.Info("Started cleanup routine with interval: %v", cleanupInterval)
}

// removeClient 从所有数据结构中移除客户端
func (s *WsServer) removeClient(client *ClientConnWrapper, reason string) {
	logger.Info("Removing client %s (user %d), reason: %s", client.id, client.UserId, reason)

	// 从全局数据结构中移除
	delete(s.connectedClients, client)

	// 从userClients中移除
	if s.userClients[client.UserId] != nil {
		delete(s.userClients[client.UserId], client)
		// 如果该用户没有其他客户端了，删除整个用户记录
		if len(s.userClients[client.UserId]) == 0 {
			delete(s.userClients, client.UserId)
		}
	}

	// 从订阅记录中移除
	delete(s.clientSubscriptions, client)

	// 清理客户端的频道
	for channelId := range client.channels {
		// 关闭频道并清理资源
		if channel := client.channels[channelId]; channel != nil {
			channel.Close()
		}
		delete(client.channels, channelId)
	}

	// 从频道客户端映射中移除
	for channelId, clientsMap := range s.channelsClient {
		if clientsMap != nil {
			delete(clientsMap, client)
			// 如果该频道没有其他客户端了，删除整个频道记录
			if len(clientsMap) == 0 {
				delete(s.channelsClient, channelId)

				// 同时从channelIdMap中移除该频道
				s.cIMu.Lock()
				delete(s.channelIdMap, channelId)
				s.cIMu.Unlock()
			}
		}
	}
}

func (s *WsServer) handleClientMessage(client *ClientConnWrapper, msg ClientMessage) error {
	switch msg.Action {
	case "ping":
		// 响应心跳
		return client.Pong()
	case "subscribe":
		s.cSMu.Lock()
		if s.clientSubscriptions[client] == nil {
			s.clientSubscriptions[client] = make(map[string]bool)
		}
		s.clientSubscriptions[client][msg.Topic] = true
		s.cSMu.Unlock()
		logger.Info("Client %s subscribed to topic %s", client.id, msg.Topic)
	case "unsubscribe":
		s.cSMu.Lock()
		if s.clientSubscriptions[client] != nil {
			delete(s.clientSubscriptions[client], msg.Topic)
		}
		s.cSMu.Unlock()
		logger.Info("Client %s unsubscribed from topic %s", client.id, msg.Topic)
	case "msg":

		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'msg'"))
			return fmt.Errorf("topic is required for action 'msg'")
		}

		messaging.Publish(s.sendCtx, messaging.MessageStruct{
			Type: messaging.UserToServerSocket,
			Payload: messaging.UserMessagePayload{
				MessageId: msg.Topic,
				UserId:    client.UserId,
				Data:      msg.Data,
				ClientId:  client.ClientId,
			},
		})
	case "channel_start":
		if s.channelFactory == nil {
			client.SendChannelCreatedFailed(msg.Topic, ErrInternalServer, fmt.Errorf("channel factory is not set"))
			return fmt.Errorf("channel factory is not set")
		}

		if utils.IsEmpty(msg.Topic) {
			// client.SendChannelError(msg.Topic, ErrEmptyTopic, fmt.Errorf("topic is required for action 'channel_start'"))
			// topic为空时,根本就不知道id是什么,没办法发送channel的错误信息
			client.SendErrorMsg(ErrEmptyTopic, fmt.Errorf("topic is required for action 'channel_start'"))
			return fmt.Errorf("topic is required for action 'channel_start'")
		}

		channelId := s.CreateChannelId(client.UserId, client.ClientId, msg)

		// 这里如果允许加入的话,就可以支持多客户端加入到一个频道,但是这样会不会有安全性问题?
		if s.GetChannelById(channelId) != nil {
			client.SendChannelCreatedFailed(msg.Topic, ErrChannelExists, fmt.Errorf("channel %s already exists", channelId))
			return fmt.Errorf("channel %s already exists", channelId)
		}
		channel := s.channelFactory.StartNewChannel(msg.Topic, channelId, client.UserId)

		// 将新创建的频道加入到全局映射中
		s.cIMu.Lock()
		s.channelIdMap[channelId] = channel
		s.cIMu.Unlock()

		client.AddChannel(channel)
		s.AddChannelClientMapping(channel.ID, client)
		client.SendChannelCreatedSuccess(msg.Topic, channel)

	case "channel":
		if s.channelFactory == nil {
			client.SendErrorMsg(ErrInternalServer, fmt.Errorf("channel factory is not set"))
			return fmt.Errorf("channel factory is not set")
		}

		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'channel_start'"))
			return fmt.Errorf("topic is required for action 'channel_start'")
		}

		// 这里是直接将topic用作id
		channel := s.GetChannelById(msg.Topic)
		if channel == nil {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("channel %s not found", msg.Topic))
			return fmt.Errorf("channel %s not found", msg.Topic)
		}

		channel.NewMessage(msg.Data).ToServer()

	case "channel_close":
		if s.channelFactory == nil {
			client.SendErrorMsg(ErrInternalServer, fmt.Errorf("channel factory is not set"))
			return fmt.Errorf("channel factory is not set")
		}
		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'channel_close'"))
			return fmt.Errorf("topic is required for action 'channel_close'")
		}
		channel := s.GetChannelById(msg.Topic)
		if channel == nil {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("channel %s not found", msg.Topic))
			return fmt.Errorf("channel %s not found", msg.Topic)
		}
		channel.Close()
		logger.Info("Client %s closed channel %s", client.id, channel.ID)
	default:
		logger.Warn("Unknown action from client %s: %s", client.id, msg.Action)
	}
	return nil
}

func (s *WsServer) generateSessionID() string {
	return utils.UUIDString()
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
}

func (c *ClientConnWrapper) Pong() error {
	c.lastPongMu.Lock()
	c.lastPong = time.Now()
	c.lastPongMu.Unlock()
	return c.SendMessage(ClientMessage{
		Action: "pong",
	})
}

// GetLastPong 安全地获取最后一次pong时间
func (c *ClientConnWrapper) GetLastPong() time.Time {
	c.lastPongMu.RLock()
	defer c.lastPongMu.RUnlock()
	return c.lastPong
}

// SetLastPong 安全地设置最后一次pong时间
func (c *ClientConnWrapper) SetLastPong(t time.Time) {
	c.lastPongMu.Lock()
	defer c.lastPongMu.Unlock()
	c.lastPong = t
}

func (c *ClientConnWrapper) SendMessage(message any) error {
	return c.Conn.WriteJSON(message)
}

func (c *ClientConnWrapper) SendErrorMsg(code ErroeCode, err error) {
	response := map[string]interface{}{
		"topic": "?er",
		"data": map[string]interface{}{
			"code":   code,
			"detail": err.Error(),
		},
		"timestamp": time.Now().Unix(),
	}
	c.Conn.WriteJSON(response)
}

func (s *WsServer) handleClientConnection(w http.ResponseWriter, r *http.Request) {
	wsConfig := configs.GetConfig().Socket

	if s.allowAll {
		logger.Warn("Allowing all origins for WebSocket connections")
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  wsConfig.ReadBufferSize,
		WriteBufferSize: wsConfig.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			if s.allowAll {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, o := range s.allowList {
				if o == origin {
					return true
				}
			}
			logger.Warn("Blocking WebSocket connection from origin: %s", origin)
			return false
		},
	}

	// 升级HTTP连接为WebSocket连接
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket升级失败: %v", err)
		return
	}
	defer ws.Close()

	// 从请求url中获取token
	token := r.URL.Query().Get("token")
	claims, err := jwt.GetService().ValidateToken(token)

	if err != nil {
		logger.Warn("无效的JWT令牌: %v", err)
		// 发送自定义 JSON 错误消息
		response := map[string]interface{}{
			"topic": "?dc",
			"data": map[string]interface{}{
				"code":   ErrTokenExpired,
				"detail": err.Error(),
			},
			"timestamp": time.Now().Unix(),
		}

		ws.WriteJSON(response)
		return
	}

	s.LockAll()
	// 注册新的客户端
	client := &ClientConnWrapper{
		id:       s.generateSessionID(),
		Conn:     ws,
		UserId:   claims.UserID,
		ClientId: claims.ClientDeviceId,

		channels: make(map[string]*channel.Channel, 16),
	}
	// 安全地设置初始 lastPong 时间
	client.SetLastPong(time.Now())

	s.connectedClients[client] = true
	if s.userClients[client.UserId] == nil {
		s.userClients[client.UserId] = make(map[*ClientConnWrapper]bool)
	}
	s.userClients[client.UserId][client] = true
	s.clientSubscriptions[client] = make(map[string]bool)

	s.UnlockAll()

	for {
		var msg ClientMessage
		// 读取客户端发送的消息
		err := ws.ReadJSON(&msg)
		if err != nil {
			logger.Warn("读取客户端消息失败: %v", err)
			s.LockAll()
			// 客户端断开连接，清理数据结构
			s.removeClient(client, "connection error or disconnect")
			s.UnlockAll()
			break
		}

		// 处理客户端消息
		if err := s.handleClientMessage(client, msg); err != nil {
			logger.Error("处理客户端消息失败: %v", err)
		}
	}
}

func (s *WsServer) Start(address string) error {
	// 启动清理过期客户端的协程
	s.startCleanupRoutine()

	http.HandleFunc("/ws", s.handleClientConnection)
	logger.Info("WebSocket Server started on %s", address)
	return http.ListenAndServe(address, nil)
}

func (s *WsServer) CreateSender() types.MessageSender {
	return func(message messaging.SocketMessagePayload) error {
		topic := message.Topic
		userId := message.UserId

		var userC map[*ClientConnWrapper]bool
		if userId != nil {
			userC = s.userClients[*userId]
		} else {
			userC = s.connectedClients
		}
		for c := range userC {
			subs := s.clientSubscriptions[c]

			subsList := make([]string, 0, len(subs))
			for sub := range subs {
				subsList = append(subsList, sub)
			}

			if utils.IsAnyMatch(subsList, topic) {
				logger.Info("Sending message to user %d on topic %s", *userId, topic)
				c.SendMessage(message)
			}
		}
		return nil
	}
}

// 优雅停机
func (s *WsServer) Shutdown() {
	s.LockAll()
	defer s.UnlockAll()
	for client := range s.connectedClients {
		// 释放所有的channel
		for _, channel := range client.channels {
			client.SendChannelClosed(channel.ID)
		}
		// 关闭连接
		client.Conn.Close()
		s.removeClient(client, "server shutdown")
	}
	logger.Info("WebSocket server shutdown complete")
}
