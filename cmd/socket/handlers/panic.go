package handlers

import (
	"context"
	"fmt"
	channelhandler "go-backend/pkg/channel_handler"
	"go-backend/pkg/logging"
)

func RegisterPanicTest() {
	ctx := context.Background()
	handler := channelhandler.NewChannelHandler(channelhandler.CreateChannelHandlerOptions{
		Topic:       "test_panic/#",
		SendMessage: channelhandler.NewMessageSender(ctx),
		NewChannelReceived: func(channel *channelhandler.IsolateChannel) error {
			// 前五次循环把收到的内容原样返回,第六次抛出panic
			count := 0
			for {
				if count >= 6 {
					logging.Info("Channel %s reached max message count, closing.", channel.ID)
					// 故意制造panic
					channel.Panic(fmt.Errorf("Intentional panic for testing on channel %s", channel.ID))
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
		},
		CloseChannel: channelhandler.NewCloseChannelHandler(ctx),
		ErrSender:    channelhandler.NewErrorSender(ctx),
	})

	handler.SetLogger(logging.WithName("ChatChannelHandler"))
	handler.RegisterHandler()
}
