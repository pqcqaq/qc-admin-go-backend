package handlers

import (
	"context"
	channelhandler "go-backend/pkg/channel_handler"
	"go-backend/pkg/logging"
)

func RegisterIsolate() {
	ctx := context.Background()
	handler := channelhandler.NewChannelHandler(channelhandler.CreateChannelHandlerOptions{
		Topic:       "test_handler/#",
		SendMessage: channelhandler.NewMessageSender(ctx),
		NewChannelReceived: func(channel *channelhandler.IsolateChannel) error {
			count := 0
			for {
				if count >= 50 {
					logging.Info("Channel %s reached max message count, closing.", channel.ID)
					channel.Close()
					break
				}
				msg, err := channel.Read()
				if err != nil {
					return err
				}
				// 处理消息
				logging.Info("Received message on channel %s: %v", channel.ID, msg.Data)
				channel.Send(msg.Data)
				count++
			}
			return nil
		},
		CloseChannel: channelhandler.NewCloseChannelHandler(ctx),
	})

	handler.SetLogger(logging.WithName("ChannelHandler"))
	handler.RegisterHandler()
}
