package channel

import (
	"context"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
)

func CreateChannelFactory(options ChannelFactoryOptions) *ChannelFactory {
	return &ChannelFactory{
		toClient: options.ToClient,
		toServer: options.ToServer,
		onClose:  options.OnClose,
		sendCtx:  context.Background(),
	}
}

func (c *ChannelFactory) StartNewChannel(topic, id string, userId uint64) *Channel {
	newChannel := Channel{
		ID:        id,
		Topic:     topic,
		CreatorId: userId,
		factory:   c,
	}
	newChannel.SetStatus(Channel_Started)
	c.notifyServer(topic, id, userId, messaging.ChannelActionCreate)
	return &newChannel
}

func (c *ChannelFactory) notifyServer(topic, id string, userId uint64, action messaging.ChannelAction) {
	_, err := messaging.Publish(c.sendCtx, messaging.MessageStruct{
		Type: messaging.ChannelToServer,
		Payload: messaging.ChannelMessagePayLoad{
			ID:     id,
			Topic:  topic,
			UserID: userId,
			Action: action,
		},
	})
	if err != nil {
		logger.Error("Channel 创建时通知创建消息失败: %v, action: %d", err, action)
	}
}

func (c *Channel) NewMessage(data any) *ChannelMsg {
	return &ChannelMsg{
		Data:    data,
		channel: c,

		timestamp: utils.Now().Unix(),
	}
}

func (c *ChannelMsg) GetChannelId() string {
	return c.channel.ID
}

func (c *ChannelMsg) GetChannelCreatorId() uint64 {
	return c.channel.CreatorId
}

func (c *ChannelMsg) ToClient() error {
	c.channel.history = append(c.channel.history, *c)
	err := c.channel.factory.toClient(*c)

	if err != nil {
		logger.Error("向客户端发送消息失败: %v", err)
	}

	return err
}

func (c *ChannelMsg) ToServer() error {
	c.channel.history = append(c.channel.history, *c)
	err := c.channel.factory.toServer(*c)

	if err != nil {
		logger.Error("向服务器发送消息失败: %v", err)
	}

	return err
}

func (c *Channel) Close() error {
	c.factory.notifyServer(c.Topic, c.ID, c.CreatorId, messaging.ChannelActionClose)
	if c.factory.onClose != nil {
		err := c.factory.onClose(c)
		if err != nil {
			logger.Error("Channel 关闭时出错: %v", err)
			return err
		}
	} else {
		logger.Warn("Channel 关闭时没有设置 onClose 回调")
	}
	// 这里暂时不做处理,在ws中发送关闭消息
	c.SetStatus(Channel_Closed)
	return nil
}
