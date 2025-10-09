package main

import (
	"context"
	"fmt"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
)

func setupHandlers(ctx context.Context) {
	consumer := messaging.NewMessageConsumer(
		"qc-admin_api_server",
		messaging.ChannelOpenCheck,
	)
	consumer.CreateGroup(ctx)
	// 注册处理器来处理创建channel的请求
	logging.Info("Register channel open check handler")
	messaging.RegisterHandler(messaging.ChannelOpenCheck, func(message messaging.MessageStruct) error {
		socketMsgMap, ok := message.Payload.(map[string]interface{})
		if !ok {
			logging.Error("Invalid message payload type")
			return fmt.Errorf("invalid message payload type")
		}

		var socketMsg messaging.ChannelOpenCheckPayload
		err := utils.MapToStruct(socketMsgMap, &socketMsg)
		if err != nil {
			logging.Error("Failed to convert payload to SocketMessagePayload: %v", err)
			return fmt.Errorf("failed to convert payload to SocketMessagePayload: %w", err)
		}

		// 若已经超时五秒钟则不管了
		if utils.Now().Unix()-socketMsg.Timestamp > 5 {
			logging.Warn("Channel open check message timed out for channel ID: %s", socketMsg.ChannelID)
			return nil
		}

		logging.Info("Received channel open check for channel ID: %s, topic: %s, userID: %d, sessionId: %s, clientId: %d", socketMsg.ChannelID, socketMsg.Topic, socketMsg.UserID, socketMsg.SessionId, socketMsg.ClientId)
		// 这里发送创建请求, 若五秒钟之后还没应答则创建失败
		messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.ChannelOpenRes,
			Payload: messaging.ChannelOpenCheckPayload{
				ChannelID: socketMsg.ChannelID,
				Topic:     socketMsg.Topic,
				UserID:    socketMsg.UserID,
				SessionId: socketMsg.SessionId,
				ClientId:  socketMsg.ClientId,
				Allowed:   true, // 初始为不允许, 需要后台服务确认
				Timestamp: utils.Now().Unix(),
			},
		})

		return nil
	})
	consumer.Consume(ctx)
}
