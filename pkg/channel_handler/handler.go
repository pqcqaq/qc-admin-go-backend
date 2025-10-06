package channelhandler

import (
	"context"
	"fmt"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
)

func (ch *ChannelHandler) SetLogger(l Logger) {
	ch.logger = l
}

// 获取状态
func (i *IsolateChannel) GetStatus() IsolateChannelLifecycle {
	i.sMu.Lock()
	defer i.sMu.Unlock()
	return i.status
}

// 设置状态
func (i *IsolateChannel) setStatus(status IsolateChannelLifecycle) {
	i.sMu.Lock()
	defer i.sMu.Unlock()
	i.status = status
}

// 尝试读取消息
func (i *IsolateChannel) TryRead() (*IsolateChannelMsg, bool) {
	if i.GetStatus() == Channel_Closed {
		return nil, false
	}

	if i.GetStatus() != Channel_Running {
		return nil, false
	}

	// 非阻塞读取
	select {
	case msg := <-i.readChan:
		return msg, true
	default:
		return nil, false
	}
}

// 读取消息
func (i *IsolateChannel) Read() (*IsolateChannelMsg, error) {
	if i.GetStatus() == Channel_Closed {
		return nil, fmt.Errorf("channel %s is closed", i.ID)
	}

	if i.GetStatus() != Channel_Running {
		return nil, fmt.Errorf("channel %s is not running, cannot read", i.ID)
	}

	msg := <-i.readChan
	return msg, nil
}

// 关闭Channel
func (i *IsolateChannel) Close() error {
	if i.GetStatus() == Channel_Closed {
		return fmt.Errorf("channel %s is already closed", i.ID)
	}

	if i.GetStatus() != Channel_Running {
		return fmt.Errorf("channel %s is not running, cannot close", i.ID)
	}

	return i.factory.close(i.ID)
}

// 发送消息
func (i *IsolateChannel) Send(msg any) error {
	if i.GetStatus() == Channel_Closed {
		return fmt.Errorf("channel %s is closed", i.ID)
	}

	if i.GetStatus() != Channel_Running {
		return fmt.Errorf("channel %s is not running, cannot send", i.ID)
	}

	return i.factory.send(i.ID, msg)
}

// 等待关闭信号
func (i *IsolateChannel) Signal() error {
	if i.GetStatus() == Channel_Closed {
		return fmt.Errorf("channel %s is already closed", i.ID)
	}
	if i.GetStatus() != Channel_Running {
		return fmt.Errorf("channel %s is not running, cannot wait for signal", i.ID)
	}

	// 等待关闭信号
	<-i.closeChan
	return nil
}

func (ch *ChannelHandler) onReceiveStarted(msg messaging.ChannelMessagePayLoad) {
	if utils.MatchTopic(ch.topic, msg.Topic) {
		channel := ch.StartNewChannel(ch.topic, msg.ID, msg.SessionId, msg.UserID, msg.ClientId)
		ch.putChannel(channel)
		ch.logger.Info("New channel created: %s, topic: %s, userId: %d", msg.ID, msg.Topic, msg.UserID)
	}
}

func (ch *ChannelHandler) onReceiveMsg(msg messaging.ChannelMessagePayLoad) {
	if utils.MatchTopic(ch.topic, msg.Topic) {
		channel, exists := ch.getChannel(msg.ID)
		if !exists {
			ch.logger.Error("Channel %s not found for topic %s", msg.ID, msg.Topic)
			return
		}
		channel.readChan <- &IsolateChannelMsg{
			Data:      msg.Data,
			channel:   channel,
			timestamp: utils.Now().Unix(),
		}
	}
}

func (ch *ChannelHandler) onReceiveStoped(msg messaging.ChannelMessagePayLoad) {
	if utils.MatchTopic(ch.topic, msg.Topic) {
		channel, exists := ch.getChannel(msg.ID)
		if !exists {
			ch.logger.Error("Channel %s not found for topic %s", msg.ID, msg.Topic)
			return
		}

		// 设置状态为关闭
		channel.setStatus(Channel_Closed)

		// 发送关闭信号
		close(channel.closeChan)

		// 从管理器中删除 channel
		ch.deleteChannel(msg.ID)
		ch.logger.Info("Channel closed: %s, topic: %s", msg.ID, msg.Topic)
	}
}

func (ch *ChannelHandler) RegisterHandler() {
	messaging.RegisterHandler(messaging.ChannelToServer, func(message messaging.MessageStruct) error {
		socketMsgMap, ok := message.Payload.(map[string]interface{})
		if !ok {
			logging.Error("Invalid message payload type")
			return fmt.Errorf("invalid message payload type")
		}

		var socketMsg messaging.ChannelMessagePayLoad
		err := utils.MapToStruct(socketMsgMap, &socketMsg)
		if err != nil {
			logging.Error("Failed to convert payload to SocketMessagePayload: %v", err)
			return fmt.Errorf("failed to convert payload to SocketMessagePayload: %w", err)
		}

		// 根据不同的Action处理
		switch socketMsg.Action {
		case messaging.ChannelActionCreate:
			ch.onReceiveStarted(socketMsg)
		case messaging.ChannelActionMsg:
			ch.onReceiveMsg(socketMsg)
		case messaging.ChannelActionClose:
			ch.onReceiveStoped(socketMsg)
		}

		return nil
	})
}

func (ch *ChannelHandler) putChannel(channel *IsolateChannel) {
	ch.cMu.Lock()
	defer ch.cMu.Unlock()
	// 放入map
	ch.channels[channel.ID] = channel
}

func (ch *ChannelHandler) getChannel(channelId string) (*IsolateChannel, bool) {
	ch.cMu.Lock()
	defer ch.cMu.Unlock()
	channel, exists := ch.channels[channelId]
	return channel, exists
}

func (ch *ChannelHandler) deleteChannel(channelId string) {
	ch.cMu.Lock()
	defer ch.cMu.Unlock()
	delete(ch.channels, channelId)
}

func NewChannelHandler(options CreateChannelHandlerOptions) *ChannelHandler {
	return &ChannelHandler{
		topic:      options.Topic,
		onReceived: options.NewChannelReceived,
		send:       options.SendMessage,
		close:      options.CloseChannel,
		logger:     options.Logger,

		channels: make(map[string]*IsolateChannel),
	}
}

func (ch *ChannelHandler) StartNewChannel(topic, channelId, sessionId string, creatorId, clientId uint64) *IsolateChannel {
	channel := &IsolateChannel{
		ID:        channelId,
		Topic:     topic,
		CreatorId: creatorId,
		SessionId: sessionId,
		ClientId:  clientId,
		history:   make([]IsolateChannelMsg, 0),
		status:    Channel_Running,
		readChan:  make(chan *IsolateChannelMsg, 100),
		factory:   ch,
		closeChan: make(chan struct{}),
	}
	go ch.onReceived(channel)
	return channel
}

func NewCloseChannelHandler(ctx context.Context) ChannelCloser {
	return func(channelId string) error {
		messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.ChannelToUser,
			Payload: messaging.ChannelMessagePayLoad{
				ID:     channelId,
				Action: messaging.ChannelActionClose,
			},
		})
		return nil
	}
}

func NewMessageSender(ctx context.Context) ChannelSender {
	return func(channelId string, msg any) error {
		messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.ChannelToUser,
			Payload: messaging.ChannelMessagePayLoad{
				ID:     channelId,
				Data:   msg,
				Action: messaging.ChannelActionMsg,
			},
		})
		return nil
	}
}
