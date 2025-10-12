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
	"time"

	"github.com/gorilla/websocket"
)

var internalExts = []string{".err", ".res", ".cre", ".clo"}

func (s *WsServer) GetChannelById(channelId string) *channel.Channel {
	s.cIMu.Lock()
	defer s.cIMu.Unlock()
	return s.channelIdMap[channelId]
}

func NewWsServer(options WsServerOptions) *WsServer {
	allowList := options.AllowOrigins
	allowAll := len(allowList) == 0
	wsServer := &WsServer{
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
	wsServer.startChannelOpenListener()
	wsServer.startSubscribeListener()
	return wsServer
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
	s.cCMu.Lock()
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
	s.cCMu.Unlock()
}

func (s *WsServer) handleClientMessage(client *ClientConnWrapper, msg ClientMessage) error {
	switch msg.Action {
	case "ping":
		// 响应心跳
		return client.Pong()
	case "subscribe":
		// 若是以字母或者数字开头的，则是内部订阅，可以直接订阅
		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'subscribe'"))
			return fmt.Errorf("topic is required for action 'subscribe'")
		}

		logger.Info("Client %s requests to subscribe to internal topic %s", client.id, msg.Topic)
		// 内部的频道订阅请求, 直接订阅
		// 规则: 以字母或者数字开头的,或者以.err/.res/.cre/.clo结尾的都是内部频道
		// 其他的都需要后台服务进行权限验证
		// 这样设计的目的是为了让一些公共频道可以直接订阅,而不需要每次都经过后台服务验证,提高效率
		// 例如用户的个人消息频道,系统公告频道等
		// 当然,如果有安全性要求的频道,还是需要经过后台服务验证的
		// 例如用户的订单消息频道,支付消息频道等
		// 这些频道一般都是以用户ID开头的,这样就可以避免普通用户订阅到其他用户的频道
		// 这样就可以保证只有用户12345自己可以订阅到这些频道,而其他用户无法订阅到
		if !utils.StartsWithAlphanumeric(msg.Topic) || utils.IsEndWith(msg.Topic, internalExts...) {
			s.subsTopic(client, msg.Topic)
			return nil
		}

		// 发送订阅请求到后台服务进行权限验证
		_, err := messaging.Publish(s.sendCtx, messaging.MessageStruct{
			Type: messaging.SubscribeCheck,
			Payload: messaging.SubscribeCheckPayload{
				Topic:     msg.Topic,
				UserID:    client.UserId,
				SessionId: client.id,
				ClientId:  client.ClientId,
				Allowed:   false, // 初始为不允许, 需要后台服务确认
				Timestamp: utils.Now().Unix(),
			},
		})

		if err != nil {
			client.SendSubsFailed(msg.Topic, fmt.Errorf("failed to publish subscribe check: %w", err))
			return fmt.Errorf("failed to publish subscribe check: %w", err)
		}

		// 五秒钟之后还没创建成功则表示失败
		time.AfterFunc(5*time.Second, func() {
			s.cSMu.Lock()
			if s.clientSubscriptions[client] == nil || !s.clientSubscriptions[client][msg.Topic] {
				client.SendSubsFailed(msg.Topic, fmt.Errorf("subscribe to topic %s timed out", msg.Topic))
			}
			s.cSMu.Unlock()
		})

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

		_, err := messaging.Publish(s.sendCtx, messaging.MessageStruct{
			Type: messaging.UserToServerSocket,
			Payload: messaging.UserMessagePayload{
				Topic:    msg.Topic,
				UserId:   client.UserId,
				Data:     msg.Data,
				ClientId: client.ClientId,
			},
		})

		if err != nil {
			client.SendErrorMsg(ErrInternalServer, fmt.Errorf("failed to publish user message: %w", err))
			return fmt.Errorf("failed to publish user message: %w", err)
		}

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

		channelId := s.CreateChannelId(client.id, client.UserId, client.ClientId, msg)

		// 这里如果允许加入的话,就可以支持多客户端加入到一个频道,但是这样会不会有安全性问题?
		if s.GetChannelById(channelId) != nil {
			client.SendChannelCreatedFailed(msg.Topic, ErrChannelExists, fmt.Errorf("channel %s already exists", channelId))
			return fmt.Errorf("channel %s already exists", channelId)
		}
		// s.createNewChannel(msg, channelId, client)

		// 这里发送创建请求, 若五秒钟之后还没应答则创建失败
		logger.Info("Client %s requests to create channel %s for topic %s", client.id, channelId, msg.Topic)
		_, err := messaging.Publish(s.sendCtx, messaging.MessageStruct{
			Type: messaging.ChannelOpenCheck,
			Payload: messaging.ChannelOpenCheckPayload{
				ChannelID: channelId,
				Topic:     msg.Topic,
				UserID:    client.UserId,
				SessionId: client.id,
				ClientId:  client.ClientId,
				Allowed:   false, // 初始为不允许, 需要后台服务确认
				Timestamp: utils.Now().Unix(),
			},
		})

		if err != nil {
			client.SendChannelCreatedFailed(msg.Topic, ErrInternalServer, fmt.Errorf("failed to publish channel open check: %w", err))
			return fmt.Errorf("failed to publish channel open check: %w", err)
		}

		time.AfterFunc(5*time.Second, func() {
			// 五秒钟之后还没创建成功则表示失败
			if s.GetChannelById(channelId) == nil {
				client.SendChannelCreatedFailed(msg.Topic, ErrChannelCreateTimeout, fmt.Errorf("channel %s creation timed out", channelId))
			}
		})

	case "channel":
		if s.channelFactory == nil {
			client.SendErrorMsg(ErrInternalServer, fmt.Errorf("channel factory is not set"))
			return fmt.Errorf("channel factory is not set")
		}

		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'channel_start'"))
			return nil
		}

		// 这里是直接将topic用作id
		channel := s.GetChannelById(msg.Topic)
		if channel == nil {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("channel %s not found", msg.Topic))
			return nil
		}

		channel.NewMessage(msg.Data).ToServer()

	case "channel_close":
		if s.channelFactory == nil {
			client.SendErrorMsg(ErrInternalServer, fmt.Errorf("channel factory is not set"))
			return fmt.Errorf("channel factory is not set")
		}
		if utils.IsEmpty(msg.Topic) {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("topic is required for action 'channel_close'"))
			return nil
		}
		channel := s.GetChannelById(msg.Topic)
		if channel == nil {
			client.SendErrorMsg(ErrInvalidMessageId, fmt.Errorf("channel %s not found", msg.Topic))
			return nil
		}
		channel.Close()
		logger.Info("Client %s closed channel %s", client.id, channel.ID)
	default:
		logger.Warn("Unknown action from client %s: %s", client.id, msg.Action)
	}
	return nil
}

func (s *WsServer) subsTopic(client *ClientConnWrapper, topic string) {
	s.cSMu.Lock()
	if s.clientSubscriptions[client] == nil {
		s.clientSubscriptions[client] = make(map[string]bool)
	}
	s.clientSubscriptions[client][topic] = true
	s.cSMu.Unlock()
	logger.Info("Client %s subscribed to topic %s", client.id, topic)
	if utils.StartsWithAlphanumeric(topic) && !utils.IsEndWith(topic, internalExts...) {
		client.SendSubsSuccess(topic)
	}
}

func (s *WsServer) GetClientFromSessionId(sessionId string) *ClientConnWrapper {
	s.cMu.Lock()
	defer s.cMu.Unlock()
	for client := range s.connectedClients {
		if client.id == sessionId {
			return client
		}
	}
	return nil
}

func (s *WsServer) generateSessionID() string {
	return utils.UUIDString()
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
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
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
	c.SendMessage(response)
}

func (s *WsServer) handleClientConnection(w http.ResponseWriter, r *http.Request) {
	// 检查是否是 WebSocket 升级请求
	if r.Header.Get("Upgrade") != "websocket" {
		// 对于普通 HTTP 请求
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "HTTP 403 - This endpoint is for WebSocket connections only.")
		return
	}

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

		// 在此阶段直接使用ws写入是安全的，因为只有一个goroutine
		ws.WriteJSON(response)
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, utils.MapToString(response)))
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

	client.SendConnectedSuccess()
	logger.Info("New client connected: %s (user %d)", client.id, client.UserId)

	for {
		var msg ClientMessage
		// 读取客户端发送的消息
		err := ws.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Info("Client %s disconnected normally", client.id)
			} else {
				logger.Warn("读取客户端消息失败: %v", err)
			}
			s.LockAll()
			// 客户端断开连接，清理数据结构
			s.removeClient(client, "connection error or disconnect")
			s.UnlockAll()
			break
		}

		// 处理客户端消息
		if err := s.handleClientMessage(client, msg); err != nil {
			logger.Error("处理客户端消息失败: %v, data is :%+v", err, msg.Data)
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
			logger.Info("Broadcasting message on topic %s", topic)
			userC = s.connectedClients
		}
		if len(userC) == 0 {
			return nil
		}
		for c := range userC {
			subs := s.clientSubscriptions[c]

			subsList := make([]string, 0, len(subs))
			for sub := range subs {
				subsList = append(subsList, sub)
			}

			if utils.IsAnyMatch(subsList, topic) {
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
