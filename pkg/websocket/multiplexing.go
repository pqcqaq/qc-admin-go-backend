package websocket

import (
	"context"
	"fmt"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
	"go-backend/pkg/websocket/channel"
)

// 当一个Channel被创建,其有两条路径,一是从客户端发送到服务器,而是从服务器发送到客户端
// 1. 数据从服务器发送到客户端时, 本质是由其他服务发起ChannelToUser消息
//    然后该消息到消息队列被WebSocket服务器处理
//    然后由WebSocket服务器发送到对应的客户端
// 2. 数据从客户端发送到服务器时, 本质是由客户端发送ChannelMsg消息
//    然后由WebSocket服务器将该消息发布到消息队列
//    然后其他服务订阅该消息并处理
// 所以对于每个Channel,都需要有两个发送器
// 1. ToClientSender, 用于将消息发送到客户端
//    ChannelToUser  ->  Redis  ->  WebSocket Server  ->  Client
// 2. ToServerSender, 用于将消息发送到服务器
//    Client  ->  WebSocket Server  ->  Redis  ->  ChannelToServer
// ChannelHandler 在其中,接收到用户的Channel消息之后,还可以将消息通过这个Channel发送到用户
// 作为其他服务, 监听ChannelToServer, 并且可以发送ChannelToUser消息,实现与websocket的完全隔离

/*
 * CreateToClientChannelSender 创建一个用于将消息发送到客户端的发送器
 * @param ctx 上下文
 */
func (s *WsServer) CreateToClientChannelSender(ctx context.Context) channel.ToClientSender {
	return func(msg channel.ChannelMsg) error {
		id := msg.GetChannelId()
		clients := s.GetClientFromChannelId(id)
		for _, client := range clients {
			client.SendChannelMsg(id, msg)
		}
		return nil
	}
}

/*
 * CreateToServerChannelSender 创建一个用于将消息发送到服务器的发送器
 * @param ctx 上下文
 */
func (s *WsServer) CreateToServerChannelSender(ctx context.Context) channel.ToServerSender {
	return func(msg channel.ChannelMsg) error {
		channelId := msg.GetChannelId()
		userId := msg.GetChannelCreatorId()

		channel := s.GetChannelById(channelId)

		_, err := messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.ChannelToServer,
			Payload: messaging.ChannelMessagePayLoad{
				ID:     channelId,
				Topic:  channel.Topic,
				UserID: userId,
				Action: messaging.ChannelActionMsg,
				Data:   msg.Data,
			},
		})
		return err
	}
}

/*
 * CreateChannelCloser 创建一个频道关闭器
 * @param ctx 上下文
 */
func (s *WsServer) CreateChannelCloser(ctx context.Context) channel.ChannelCloser {
	return func(ch *channel.Channel) error {
		s.cIMu.Lock()
		delete(s.channelIdMap, ch.ID)
		s.cIMu.Unlock()

		// 获取所有订阅该频道的客户端
		clients := s.GetClientFromChannelId(ch.ID)
		for _, client := range clients {
			// 从客户端和全局映射中移除频道
			s.cCMu.Lock()
			delete(client.channels, ch.ID)
			if s.channelsClient[ch.ID] != nil {
				delete(s.channelsClient[ch.ID], client)
				if len(s.channelsClient[ch.ID]) == 0 {
					delete(s.channelsClient, ch.ID)
				}
			}
			s.cCMu.Unlock()

			// 通知客户端频道已关闭
			client.SendChannelClosed(ch.ID)
		}
		return nil
	}
}

func (s *WsServer) SetChannelFactory(factory *channel.ChannelFactory) {
	s.channelFactory = factory
}

func (c *ClientConnWrapper) AddChannel(channel *channel.Channel) {
	c.cMu.Lock()
	c.channels[channel.ID] = channel
	c.cMu.Unlock()
}

func (c *WsServer) CreateChannelId(clientId string, userId uint64, clientDeviceId uint64, msg ClientMessage) string {
	return utils.StringShorten(fmt.Sprintf("%s_%d_%d_%s", clientId, userId, clientDeviceId, msg.Topic), 8)
}

func (c *ClientConnWrapper) SendConnectedSuccess() {
	response := map[string]interface{}{
		"action": "connected",
	}
	c.SendMessage(response)
}

func (c *ClientConnWrapper) SendChannelCreatedSuccess(id string, channel *channel.Channel) {
	response := map[string]interface{}{
		"topic": fmt.Sprintf("%s.cre", id),
		"data": map[string]interface{}{
			"channelId": channel.ID,
		},
		"timestamp": utils.Now().Unix(),
	}
	c.SendMessage(response)
}

func (c *ClientConnWrapper) SendChannelCreatedFailed(id string, code ErroeCode, err error) {
	response := map[string]interface{}{
		"topic": fmt.Sprintf("%s.cre", id),
		"data": map[string]interface{}{
			"error": map[string]interface{}{
				"code":   code,
				"detail": err.Error(),
			},
		},
		"timestamp": utils.Now().Unix(),
	}
	c.SendMessage(response)
}

func (c *ClientConnWrapper) SendChannelClosed(id string) {
	response := map[string]interface{}{
		"topic": fmt.Sprintf("%s.clo", id),
		"data": map[string]interface{}{
			"code":   200,
			"detail": "Channel closed by server.",
		},
		"timestamp": utils.Now().Unix(),
	}
	c.SendMessage(response)
}

func (c *ClientConnWrapper) SendChannelError(id string, code ErroeCode, err error) {
	response := map[string]interface{}{
		"topic": fmt.Sprintf("%s.err", id),
		"data": map[string]interface{}{
			"error": map[string]interface{}{
				"code":   code,
				"detail": err.Error(),
			},
		},
		"timestamp": utils.Now().Unix(),
	}
	c.SendMessage(response)
}

func (c *ClientConnWrapper) SendChannelMsg(id string, msg channel.ChannelMsg) {
	response := map[string]interface{}{
		"topic":     id,
		"data":      msg.Data,
		"timestamp": utils.Now().Unix(),
	}
	c.SendMessage(response)
}

// AddChannelClientMapping
func (s *WsServer) AddChannelClientMapping(channelId string, client *ClientConnWrapper) {
	s.cCMu.Lock()
	if s.channelsClient[channelId] == nil {
		s.channelsClient[channelId] = make(map[*ClientConnWrapper]bool)
	}
	s.channelsClient[channelId][client] = true
	s.cCMu.Unlock()
}

// GetClientFromChannelId
func (s *WsServer) GetClientFromChannelId(channelId string) []*ClientConnWrapper {
	s.cCMu.Lock()
	clientsMap := s.channelsClient[channelId]
	if clientsMap == nil {
		s.cCMu.Unlock()
		return []*ClientConnWrapper{}
	}

	// 在锁内创建副本，避免数据竞争
	clients := make([]*ClientConnWrapper, 0, len(clientsMap))
	for client := range clientsMap {
		clients = append(clients, client)
	}
	s.cCMu.Unlock()
	return clients
}

// RemoveChannelClientMapping 从频道客户端映射中移除指定客户端
func (s *WsServer) RemoveChannelClientMapping(channelId string, client *ClientConnWrapper) {
	s.cCMu.Lock()
	defer s.cCMu.Unlock()

	if s.channelsClient[channelId] != nil {
		delete(s.channelsClient[channelId], client)
		// 如果该频道没有其他客户端了，删除整个频道记录
		if len(s.channelsClient[channelId]) == 0 {
			delete(s.channelsClient, channelId)
		}
	}
}

func (s *WsServer) GetClientByUserId(userId uint64) []*ClientConnWrapper {
	s.cMu.Lock()
	defer s.cMu.Unlock()
	clients := make([]*ClientConnWrapper, 0)
	for client := range s.connectedClients {
		if client.UserId == userId {
			clients = append(clients, client)
		}
	}
	return clients
}
