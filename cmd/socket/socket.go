package main

import (
	"go-backend/pkg/configs"
	"go-backend/pkg/jwt"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
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
	Action string `json:"action"` // "subscribe" or "unsubscribe"
	Topic  string `json:"topic"`  // 频道名称
	Data   string `json:"data"`   // 消息内容（仅在发布消息时使用）
}

type WsServer struct {
	// 操作全局结构体的锁
	ccMu sync.Mutex
	uCMu sync.Mutex
	cSMu sync.Mutex

	allowList []string
	allowAll  bool
}

func NewWsServer() *WsServer {
	allowList := configs.GetConfig().Socket.AllowOrigins
	allowAll := len(allowList) == 0
	return &WsServer{
		allowList: allowList,
		allowAll:  allowAll,
	}
}

func (s *WsServer) LockAll() {
	s.ccMu.Lock()
	s.uCMu.Lock()
	s.cSMu.Lock()
}

func (s *WsServer) UnlockAll() {
	s.cSMu.Unlock()
	s.uCMu.Unlock()
	s.ccMu.Unlock()
}

// cleanupExpiredClients 清理超时的客户端连接
func (s *WsServer) cleanupExpiredClients() {
	timeoutDuration := time.Duration(configs.GetConfig().Socket.PingTimeout) * time.Second
	now := time.Now()

	s.LockAll()
	defer s.UnlockAll()

	var expiredClients []*ClientConnWrapper
	for client := range connectedClients {
		if now.Sub(client.lastPong) > timeoutDuration {
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
	delete(connectedClients, client)

	// 从userClients中移除
	if userClients[client.UserId] != nil {
		delete(userClients[client.UserId], client)
		// 如果该用户没有其他客户端了，删除整个用户记录
		if len(userClients[client.UserId]) == 0 {
			delete(userClients, client.UserId)
		}
	}

	// 从订阅记录中移除
	delete(clientSubscriptions, client)
}

func (s *WsServer) handleClientMessage(client *ClientConnWrapper, msg ClientMessage) error {
	switch msg.Action {
	case "ping":
		// 响应心跳
		return client.Pong()
	case "subscribe":
		s.cSMu.Lock()
		if clientSubscriptions[client] == nil {
			clientSubscriptions[client] = make(map[string]bool)
		}
		clientSubscriptions[client][msg.Topic] = true
		s.cSMu.Unlock()
		logger.Info("Client %s subscribed to topic %s", client.id, msg.Topic)
	case "unsubscribe":
		s.cSMu.Lock()
		if clientSubscriptions[client] != nil {
			delete(clientSubscriptions[client], msg.Topic)
		}
		s.cSMu.Unlock()
		logger.Info("Client %s unsubscribed from topic %s", client.id, msg.Topic)
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
	lastPong time.Time
}

func (c *ClientConnWrapper) Pong() error {
	c.lastPong = time.Now()
	return c.SendMessage(ClientMessage{
		Action: "pong",
		Topic:  "",
	})
}

func (c *ClientConnWrapper) SendMessage(message any) error {
	return c.Conn.WriteJSON(message)
}

var connectedClients = make(map[*ClientConnWrapper]bool) // 连接的客户端
var userClients = make(map[uint64]map[*ClientConnWrapper]bool)
var clientSubscriptions = make(map[*ClientConnWrapper]map[string]bool)

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
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "无效的令牌"))
		return
	}

	s.LockAll()
	// 注册新的客户端
	client := &ClientConnWrapper{
		id:       s.generateSessionID(),
		Conn:     ws,
		UserId:   claims.UserID,
		lastPong: time.Now(),
	}
	connectedClients[client] = true
	if userClients[client.UserId] == nil {
		userClients[client.UserId] = make(map[*ClientConnWrapper]bool)
	}
	userClients[client.UserId][client] = true
	clientSubscriptions[client] = make(map[string]bool)

	s.UnlockAll()

	for {
		var msg ClientMessage
		// 读取客户端发送的消息
		err := ws.ReadJSON(&msg)
		if err != nil {
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

func SenderFunc(message messaging.SocketMessagePayload) error {
	topic := message.Topic
	userId := message.UserId

	var userC map[*ClientConnWrapper]bool
	if userId != nil {
		userC = userClients[*userId]
	} else {
		userC = connectedClients
	}
	for c := range userC {
		subs := clientSubscriptions[c]

		subsList := make([]string, 0, len(subs))
		for sub := range subs {
			subsList = append(subsList, sub)
		}

		if utils.IsAnyMatch(subsList, topic) {
			logger.Info("Sending message to user %d on topic %s", userId, topic)
			c.SendMessage(message)
		}
	}
	return nil
}
