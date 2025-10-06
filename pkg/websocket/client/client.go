package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"go-backend/pkg/utils"
)

// SocketClient WebSocket 客户端实现
type SocketClient struct {
	conn       *websocket.Conn
	state      WebSocketState
	stateMutex sync.RWMutex
	options    SocketOptions

	// 订阅管理
	subscriptions     map[string][]*SubscriptionRecord
	subscriptionMutex sync.RWMutex

	// 状态变化回调
	stateCallbacks     map[string]StateChangeCallback
	stateCallbackMutex sync.RWMutex

	// 重连和心跳
	reconnectTimer  *time.Timer
	heartbeatTicker *time.Ticker
	heartbeatCancel context.CancelFunc

	// 指数退避算法
	currentBackoffDelay time.Duration
	baseBackoffDelay    time.Duration
	maxBackoffDelay     time.Duration

	// 控制
	isManualDisconnect bool
	stopChan           chan struct{}
	doneChan           chan struct{}
	mutex              sync.Mutex

	// WebSocket写入保护
	writeMutex sync.Mutex

	// 内部系统订阅
	disconnectUnsub UnsubscribeFunction
	errorUnsub      UnsubscribeFunction
}

// NewSocketClient 创建新的WebSocket客户端
func NewSocketClient(options SocketOptions) *SocketClient {
	// 设置默认值
	if options.HeartbeatInterval == 0 {
		options.HeartbeatInterval = 30 * time.Second
	}

	client := &SocketClient{
		state:               Disconnected,
		options:             options,
		subscriptions:       make(map[string][]*SubscriptionRecord),
		stateCallbacks:      make(map[string]StateChangeCallback),
		baseBackoffDelay:    500 * time.Millisecond,
		maxBackoffDelay:     16 * time.Second,
		currentBackoffDelay: 500 * time.Millisecond,
		stopChan:            make(chan struct{}),
		doneChan:            make(chan struct{}),
	}

	return client
}

// State 获取当前连接状态
func (c *SocketClient) State() WebSocketState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

// Connect 连接到WebSocket服务器
func (c *SocketClient) Connect(token ...string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state == Connected {
		c.log("Already connected")
		return nil
	}

	if c.state == Connecting {
		c.log("Already connecting")
		return nil
	}

	// 确定使用的token
	authToken := c.options.Token
	if len(token) > 0 && token[0] != "" {
		authToken = token[0]
	}

	if authToken == "" {
		return fmt.Errorf("token is required for WebSocket connection")
	}

	if c.options.URL == "" {
		return fmt.Errorf("WebSocket URL is required")
	}

	// 重置手动断开标记
	c.isManualDisconnect = false
	c.setState(Connecting)

	// 设置内部订阅
	c.setupInternalSubscriptions()

	// 构建WebSocket URL
	u, err := url.Parse(c.options.URL)
	if err != nil {
		c.setState(Error)
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("token", authToken)
	u.RawQuery = q.Encode()

	// 连接WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		c.setState(Error)
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.setState(Connected)
	c.resetBackoffDelay()

	// 启动消息处理和心跳
	go c.handleMessages()
	c.startHeartbeat()
	c.resubscribeAll()

	c.log("WebSocket connected")
	return nil
}

// Disconnect 断开连接
func (c *SocketClient) Disconnect() <-chan struct{} {
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.isManualDisconnect = true
		c.clearReconnectTimer()
		c.stopHeartbeat()

		if c.conn != nil {
			// 使用写入锁保护WebSocket关闭消息
			c.writeMutex.Lock()
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Manual disconnect"))
			c.writeMutex.Unlock()

			c.conn.Close()
		}

		c.setState(Disconnected)
	}()

	return doneChan
}

// Subscribe 订阅主题
func (c *SocketClient) Subscribe(topic string, handler MessageHandler) UnsubscribeFunction {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	id := c.generateID()
	record := &SubscriptionRecord{
		Topic:   topic,
		Handler: handler,
		ID:      id,
	}

	// 检查是否是该主题的第一个订阅
	isFirstSubscription := len(c.subscriptions[topic]) == 0

	// 保存订阅记录
	c.subscriptions[topic] = append(c.subscriptions[topic], record)

	// 只有在第一次订阅该主题且已连接时，才发送订阅请求到服务器
	if isFirstSubscription && c.State() == Connected {
		c.sendSubscribeMessage(topic)
	}

	c.log(fmt.Sprintf("Subscribed to topic: %s (handlers: %d)", topic, len(c.subscriptions[topic])))

	// 返回取消订阅函数
	return func() {
		c.unsubscribeByID(id)
	}
}

// Unsubscribe 取消订阅
func (c *SocketClient) Unsubscribe(topic string, handler ...MessageHandler) {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	records, exists := c.subscriptions[topic]
	if !exists {
		return
	}

	if len(handler) > 0 && handler[0] != nil {
		// 取消特定处理器的订阅
		targetHandler := handler[0]
		for i, record := range records {
			// 由于Go中函数比较的限制，这里使用指针比较
			// 实际应用中可能需要其他方式来标识处理器
			if fmt.Sprintf("%p", record.Handler) == fmt.Sprintf("%p", targetHandler) {
				records = append(records[:i], records[i+1:]...)
				if len(records) == 0 {
					delete(c.subscriptions, topic)
					c.sendUnsubscribeMessage(topic)
				} else {
					c.subscriptions[topic] = records
				}
				break
			}
		}
	} else {
		// 取消该主题的所有订阅
		delete(c.subscriptions, topic)
		c.sendUnsubscribeMessage(topic)
	}

	c.log(fmt.Sprintf("Unsubscribed from topic: %s", topic))
}

// UnsubscribeAll 取消所有订阅
func (c *SocketClient) UnsubscribeAll() {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	for topic := range c.subscriptions {
		c.sendUnsubscribeMessage(topic)
	}
	c.subscriptions = make(map[string][]*SubscriptionRecord)
	c.log("Unsubscribed from all topics")
}

// OnStateChange 监听连接状态变化
func (c *SocketClient) OnStateChange(callback StateChangeCallback) UnsubscribeFunction {
	c.stateCallbackMutex.Lock()
	defer c.stateCallbackMutex.Unlock()

	id := c.generateID()
	c.stateCallbacks[id] = callback

	return func() {
		c.stateCallbackMutex.Lock()
		defer c.stateCallbackMutex.Unlock()
		delete(c.stateCallbacks, id)
	}
}

// SendMessage 发送消息
func (c *SocketClient) SendMessage(topic string, data interface{}) error {
	message := ClientMessage{
		Action: "msg",
		Topic:  topic,
		Data:   data,
	}
	return c.sendMessage(message)
}

// CreateChannel 创建频道
func (c *SocketClient) CreateChannel(topic string, handler ChannelMessageHandler, errHandler ...ChannelCloseHandler) (*Channel, error) {
	return c.createChannelWithTimeout(topic, handler, 3*time.Second, errHandler...)
}

// createChannelWithTimeout 创建频道（带超时）
func (c *SocketClient) createChannelWithTimeout(topic string, handler ChannelMessageHandler, timeout time.Duration, errHandler ...ChannelCloseHandler) (*Channel, error) {
	resultChan := make(chan *Channel, 1)
	errorChan := make(chan error, 1)

	var channelID string
	var channelCreated bool
	var createTopicUnsub UnsubscribeFunction
	mutex := sync.Mutex{}

	// 订阅频道创建响应
	createTopicUnsub = c.Subscribe(fmt.Sprintf("%s.cre", topic), func(data interface{}, responseTopic string) {
		mutex.Lock()
		defer mutex.Unlock()

		if channelCreated {
			return // 已经处理过了
		}

		// 解析频道创建响应
		var createRes ChannelCreateRes
		if dataMap, ok := data.(map[string]interface{}); ok {
			if chID, exists := dataMap["channelId"]; exists {
				if chIDStr, ok := chID.(string); ok {
					createRes.ChannelID = &chIDStr
				}
			}
			if errData, exists := dataMap["error"]; exists {
				if errMap, ok := errData.(map[string]interface{}); ok {
					errorData := &ErrorMsgData{}
					if code, codeExists := errMap["code"]; codeExists {
						if codeStr, ok := code.(string); ok {
							errorData.Code = codeStr
						}
					}
					if detail, detailExists := errMap["detail"]; detailExists {
						if detailStr, ok := detail.(string); ok {
							errorData.Detail = detailStr
						}
					}
					createRes.Error = errorData
				}
			}
		}

		channelCreated = true
		createTopicUnsub()

		if createRes.Error != nil {
			c.log(fmt.Sprintf("Channel creation error: %+v", createRes.Error))
			select {
			case errorChan <- fmt.Errorf("channel creation failed: %s %s", createRes.Error.Code, createRes.Error.Detail):
			default:
			}
			return
		}

		if createRes.ChannelID != nil {
			channelID = *createRes.ChannelID
			c.log(fmt.Sprintf("Channel created with ID: %s", channelID))

			// 创建频道实例
			channel := c.setupChannel(channelID, handler, errHandler...)
			select {
			case resultChan <- channel:
			default:
			}
		}
	})

	// 发送频道创建请求
	createMessage := ClientMessage{
		Action: "channel_start",
		Topic:  topic,
	}
	if err := c.sendMessage(createMessage); err != nil {
		createTopicUnsub()
		return nil, fmt.Errorf("failed to send channel creation request: %w", err)
	}

	// 等待结果或超时
	select {
	case channel := <-resultChan:
		return channel, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(timeout):
		createTopicUnsub()
		return nil, fmt.Errorf("channel creation timed out")
	}
}

// setupChannel 设置频道实例
func (c *SocketClient) setupChannel(channelID string, handler ChannelMessageHandler, errHandler ...ChannelCloseHandler) *Channel {
	// 创建等待通道，用于通知频道结束
	waitChan := make(chan struct{})
	var waitOnce sync.Once

	// 订阅频道消息
	messageUnsub := c.Subscribe(channelID, func(data interface{}, topic string) {
		handler(data)
	})

	// 订阅频道错误消息
	var errorUnsub UnsubscribeFunction
	if len(errHandler) > 0 && errHandler[0] != nil {
		errorUnsub = c.Subscribe(fmt.Sprintf("%s.err", channelID), func(data interface{}, topic string) {
			if dataMap, ok := data.(map[string]interface{}); ok {
				errorData := ErrorMsgData{}
				if code, exists := dataMap["code"]; exists {
					if codeStr, ok := code.(string); ok {
						errorData.Code = codeStr
					}
				}
				if detail, exists := dataMap["detail"]; exists {
					if detailStr, ok := detail.(string); ok {
						errorData.Detail = detailStr
					}
				}
				errHandler[0](errorData)
			}
		})
	}

	var closeHandler ChannelCloseHandler
	var closeUnsub UnsubscribeFunction

	// 创建发送函数
	send := func(data interface{}) error {
		message := ClientMessage{
			Action: "channel",
			Topic:  channelID,
			Data:   data,
		}
		return c.sendMessage(message)
	}

	// 创建关闭函数
	closer := func() <-chan struct{} {
		doneChan := make(chan struct{})

		go func() {
			defer close(doneChan)

			// 只有在连接状态下才发送关闭消息
			if c.State() == Connected {
				closeMessage := ClientMessage{
					Action: "channel_close",
					Topic:  channelID,
				}
				// 忽略发送错误，因为连接可能已经断开
				c.sendMessage(closeMessage)
			}

			// 清理订阅
			messageUnsub()
			if errorUnsub != nil {
				errorUnsub()
			}
			if closeUnsub != nil {
				closeUnsub()
			}
			c.Unsubscribe(channelID)

			// 通知等待者频道已结束
			waitOnce.Do(func() {
				close(waitChan)
			})
		}()

		return doneChan
	}

	// 创建等待函数
	waiter := func() <-chan struct{} {
		return waitChan
	}

	// 创建设置关闭处理器的函数
	onClose := func(handler ChannelCloseHandler) {
		closeHandler = handler

		// 订阅关闭消息
		closeUnsub = c.Subscribe(fmt.Sprintf("%s.clo", channelID), func(data interface{}, topic string) {
			if closeHandler != nil {
				if dataMap, ok := data.(map[string]interface{}); ok {
					errorData := ErrorMsgData{}
					if code, exists := dataMap["code"]; exists {
						if codeStr, ok := code.(string); ok {
							errorData.Code = codeStr
						}
					}
					if detail, exists := dataMap["detail"]; exists {
						if detailStr, ok := detail.(string); ok {
							errorData.Detail = detailStr
						}
					}
					closeHandler(errorData)
				}
			}

			// 清理资源
			closeUnsub()
			if errorUnsub != nil {
				errorUnsub()
			}
			messageUnsub()
			c.Unsubscribe(channelID)

			// 通知等待者频道已结束
			waitOnce.Do(func() {
				close(waitChan)
			})
		})
	}

	return &Channel{
		Send:    send,
		Close:   closer,
		Wait:    waiter,
		OnClose: onClose,
	}
}

// 处理接收到的消息
func (c *SocketClient) handleMessages() {
	defer close(c.doneChan)

	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.conn == nil {
				return
			}

			_, messageData, err := c.conn.ReadMessage()
			if err != nil {
				c.log(fmt.Sprintf("WebSocket read error: %v", err))
				c.setState(Disconnected)
				c.scheduleReconnect()
				return
			}

			c.handleMessage(messageData)
		}
	}
}

// 处理单个消息
func (c *SocketClient) handleMessage(data []byte) {
	var message SocketMessagePayload
	if err := json.Unmarshal(data, &message); err != nil {
		c.log(fmt.Sprintf("Error parsing message: %v", err))
		return
	}

	c.log(fmt.Sprintf("Received message: %+v", message))

	// 获取所有订阅的主题
	c.subscriptionMutex.RLock()
	defer c.subscriptionMutex.RUnlock()

	// 找到匹配的订阅并分发消息
	for subscribedTopic, records := range c.subscriptions {
		if utils.MatchTopic(subscribedTopic, message.Topic) {
			for _, record := range records {
				// 在goroutine中执行处理器，避免阻塞消息循环
				go func(handler MessageHandler, data interface{}, topic string) {
					defer func() {
						if r := recover(); r != nil {
							c.log(fmt.Sprintf("Error in message handler: %v", r))
						}
					}()
					handler(data, topic)
				}(record.Handler, message.Data, message.Topic)
			}
		}
	}
}

// 发送订阅消息到服务器
func (c *SocketClient) sendSubscribeMessage(topic string) {
	message := ClientMessage{
		Action: "subscribe",
		Topic:  topic,
	}
	c.sendMessage(message)
}

// 发送取消订阅消息到服务器
func (c *SocketClient) sendUnsubscribeMessage(topic string) {
	message := ClientMessage{
		Action: "unsubscribe",
		Topic:  topic,
	}
	c.sendMessage(message)
}

// 发送消息到服务器
func (c *SocketClient) sendMessage(message ClientMessage) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// 检查连接状态
	if c.State() != Connected || c.conn == nil {
		return fmt.Errorf("WebSocket not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.log(fmt.Sprintf("Sent message: %+v", message))
	return nil
}

// 重新订阅所有主题（用于重连后）
func (c *SocketClient) resubscribeAll() {
	c.subscriptionMutex.RLock()
	defer c.subscriptionMutex.RUnlock()

	for topic := range c.subscriptions {
		c.sendSubscribeMessage(topic)
	}
}

// 设置连接状态
func (c *SocketClient) setState(state WebSocketState) {
	c.stateMutex.Lock()
	oldState := c.state
	c.state = state
	c.stateMutex.Unlock()

	if oldState != state {
		c.log(fmt.Sprintf("State changed to: %s", state.String()))

		// 通知状态变化回调
		c.stateCallbackMutex.RLock()
		defer c.stateCallbackMutex.RUnlock()

		for _, callback := range c.stateCallbacks {
			go func(cb StateChangeCallback) {
				defer func() {
					if r := recover(); r != nil {
						c.log(fmt.Sprintf("Error in state change callback: %v", r))
					}
				}()
				cb(state)
			}(callback)
		}
	}
}

// 安排重连
func (c *SocketClient) scheduleReconnect() {
	// 如果是手动断开，则不进行重连
	if c.isManualDisconnect {
		c.log("Manually disconnected, not scheduling reconnect")
		return
	}

	c.setState(Reconnecting)

	c.reconnectTimer = time.AfterFunc(c.currentBackoffDelay, func() {
		c.log(fmt.Sprintf("Attempting to reconnect (delay: %v)", c.currentBackoffDelay))
		if err := c.Connect(); err != nil {
			c.log(fmt.Sprintf("Reconnect failed: %v", err))
			c.increaseBackoffDelay()
		}
	})
}

// 清除重连定时器
func (c *SocketClient) clearReconnectTimer() {
	if c.reconnectTimer != nil {
		c.reconnectTimer.Stop()
		c.reconnectTimer = nil
	}
}

// 开始心跳
func (c *SocketClient) startHeartbeat() {
	c.stopHeartbeat()

	ctx, cancel := context.WithCancel(context.Background())
	c.heartbeatCancel = cancel

	c.heartbeatTicker = time.NewTicker(c.options.HeartbeatInterval)

	go func() {
		defer c.heartbeatTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.heartbeatTicker.C:
				if c.State() == Connected {
					pingMessage := ClientMessage{
						Action: "ping",
						Topic:  "",
					}
					c.sendMessage(pingMessage)
				}
			}
		}
	}()
}

// 停止心跳
func (c *SocketClient) stopHeartbeat() {
	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
		c.heartbeatCancel = nil
	}
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		c.heartbeatTicker = nil
	}
}

// 根据ID取消订阅
func (c *SocketClient) unsubscribeByID(id string) {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	for topic, records := range c.subscriptions {
		for i, record := range records {
			if record.ID == id {
				records = append(records[:i], records[i+1:]...)
				if len(records) == 0 {
					delete(c.subscriptions, topic)
					c.sendUnsubscribeMessage(topic)
				} else {
					c.subscriptions[topic] = records
				}
				c.log(fmt.Sprintf("Unsubscribed by ID: %s from topic: %s", id, topic))
				return
			}
		}
	}
}

// 生成唯一ID
func (c *SocketClient) generateID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// 增加退避延迟时间（指数退避）
func (c *SocketClient) increaseBackoffDelay() {
	c.currentBackoffDelay = time.Duration(math.Min(
		float64(c.currentBackoffDelay*2),
		float64(c.maxBackoffDelay),
	))
	c.log(fmt.Sprintf("Backoff delay increased to: %v", c.currentBackoffDelay))
}

// 重置退避延迟时间
func (c *SocketClient) resetBackoffDelay() {
	c.currentBackoffDelay = c.baseBackoffDelay
	c.log(fmt.Sprintf("Backoff delay reset to: %v", c.currentBackoffDelay))
}

// 设置内部系统订阅
func (c *SocketClient) setupInternalSubscriptions() {
	// 清理之前的订阅
	if c.disconnectUnsub != nil {
		c.disconnectUnsub()
	}
	if c.errorUnsub != nil {
		c.errorUnsub()
	}

	// 订阅断开连接消息
	c.disconnectUnsub = c.Subscribe("?dc", func(data interface{}, topic string) {
		c.log(fmt.Sprintf("[SocketClient] Received disconnect message: %+v", data))

		// 尝试解析断开连接消息
		if dataMap, ok := data.(map[string]interface{}); ok {
			if code, exists := dataMap["code"]; exists && code == "TOKEN_EXPIRED" {
				c.Disconnect()
				c.log("Disconnected due to token expiration")

				if c.options.RefreshToken != nil {
					newToken, err := c.options.RefreshToken()
					if err != nil {
						c.log(fmt.Sprintf("Failed to refresh token: %v", err))
						return
					}
					if newToken == "" {
						c.log("No new token obtained, cannot reconnect")
						return
					}

					c.log("Token refreshed, reconnecting...")
					c.options.Token = newToken
					if err := c.Connect(newToken); err != nil {
						c.log(fmt.Sprintf("Reconnection failed: %v", err))
					}
				}
			}
		}
	})

	// 订阅错误消息
	c.errorUnsub = c.Subscribe("?er", func(data interface{}, topic string) {
		c.log(fmt.Sprintf("[SocketClient] Received error message: %+v", data))

		if c.options.ErrorHandler != nil {
			// 尝试将data转换为ErrorMsgData
			if dataMap, ok := data.(map[string]interface{}); ok {
				errorData := ErrorMsgData{}
				if code, exists := dataMap["code"]; exists {
					if codeStr, ok := code.(string); ok {
						errorData.Code = codeStr
					}
				}
				if detail, exists := dataMap["detail"]; exists {
					if detailStr, ok := detail.(string); ok {
						errorData.Detail = detailStr
					}
				}
				c.options.ErrorHandler(errorData)
			}
		}
	})
}

// 日志输出
func (c *SocketClient) log(message string) {
	if c.options.Debug {
		log.Printf("[SocketClient] %s", message)
	}
}
