package handlers

import (
	"context"
	channelhandler "go-backend/pkg/channel_handler"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
)

func RegisterIsolate() {

	channelhandler.SetLogger(logging.WithName("ChannelHandler"))

	ctx := context.Background()
	handler := channelhandler.NewChannelHandler(channelhandler.CreateChannelHandlerOptions{
		Topic: "test_handler/#",
		SendMessage: func(channelId string, msg any) error {
			messaging.Publish(ctx, messaging.MessageStruct{
				Type: messaging.ChannelToUser,
				Payload: messaging.ChannelMessagePayLoad{
					ID:     channelId,
					Data:   msg,
					Action: messaging.ChannelActionMsg,
				},
			})
			return nil
		},
		NewChannelReceived: func(channel *channelhandler.IsolateChannel) error {
			count := 0
			for {
				if count >= 5 {
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
		CloseChannel: func(channelId string) error {
			messaging.Publish(ctx, messaging.MessageStruct{
				Type: messaging.ChannelToUser,
				Payload: messaging.ChannelMessagePayLoad{
					ID:     channelId,
					Action: messaging.ChannelActionClose,
				},
			})
			return nil
		},
	})

	handler.RegisterHandler()
}
