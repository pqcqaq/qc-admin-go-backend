package main

import (
	"context"
	"fmt"
	"go-backend/internal/funcs"
	"go-backend/internal/subscription"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
)

const (
	ChannelStartMethodStr = "ChannelStart"
	SubscribeMethodStr    = "Subscribe"
)

func setupHandlers(ctx context.Context) {
	consumer := messaging.NewMessageConsumer(
		"qc-admin_api_server",
		messaging.ChannelOpenCheck,
		messaging.SubscribeCheck,
	)
	consumer.CreateGroup(ctx)

	// 启动Stream清理器
	cleaner := messaging.NewStreamCleaner(messaging.ChannelOpenCheck)
	cleaner.StartCleanup(ctx)

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

		allowed, _ := funcs.IsTopicAllowed(socketMsg.Topic, socketMsg.UserID, ChannelStartMethodStr)

		// 这里发送创建请求, 若五秒钟之后还没应答则创建失败
		_, err = messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.ChannelOpenRes,
			Payload: messaging.ChannelOpenCheckPayload{
				ChannelID: socketMsg.ChannelID,
				Topic:     socketMsg.Topic,
				UserID:    socketMsg.UserID,
				SessionId: socketMsg.SessionId,
				ClientId:  socketMsg.ClientId,
				Allowed:   allowed,
				Timestamp: utils.Now().Unix(),
			},
		})

		if err != nil {
			logging.Error("Failed to publish channel open response: %v", err)
			return fmt.Errorf("failed to publish channel open response: %w", err)
		}

		return nil
	})

	messaging.RegisterHandler(messaging.SubscribeCheck, func(message messaging.MessageStruct) error {
		socketMsgMap, ok := message.Payload.(map[string]interface{})
		if !ok {
			logging.Error("Invalid message payload type")
			return fmt.Errorf("invalid message payload type")
		}

		var socketMsg messaging.SubscribeCheckPayload
		err := utils.MapToStruct(socketMsgMap, &socketMsg)
		if err != nil {
			logging.Error("Failed to convert payload to SocketMessagePayload: %v", err)
			return fmt.Errorf("failed to convert payload to SocketMessagePayload: %w", err)
		}

		// 若已经超时五秒钟则不管了
		if utils.Now().Unix()-socketMsg.Timestamp > 5 {
			logging.Warn("Channel open check message timed out for subscribe to topic: %s", socketMsg.Topic)
			return nil
		}

		logging.Info("Received subscribe check to topic: %s, userID: %d, sessionId: %s, clientId: %d", socketMsg.Topic, socketMsg.UserID, socketMsg.SessionId, socketMsg.ClientId)

		allowed, _ := funcs.IsTopicAllowed(socketMsg.Topic, socketMsg.UserID, SubscribeMethodStr)

		var data any = nil
		if !allowed {
			logging.Warn("Subscription to topic %s denied for user %d", socketMsg.Topic, socketMsg.UserID)
		} else {
			data, err = subscription.PublishSubscribeSuccessMessage(socketMsg.Topic, subscription.SubscribeSuccessPayload{
				Topic:     socketMsg.Topic,
				UserID:    socketMsg.UserID,
				SessionId: socketMsg.SessionId,
				ClientId:  socketMsg.ClientId,
			})
			if err != nil {
				logging.Warn("Failed to publish subscribe success message: %v", err)
			}
		}

		// 这里发送创建请求, 若五秒钟之后还没应答则创建失败
		_, err = messaging.Publish(ctx, messaging.MessageStruct{
			Type: messaging.SubscribeRes,
			Payload: messaging.SubscribeCheckPayload{
				Topic:     socketMsg.Topic,
				UserID:    socketMsg.UserID,
				SessionId: socketMsg.SessionId,
				ClientId:  socketMsg.ClientId,
				Allowed:   allowed,
				Timestamp: utils.Now().Unix(),
				Data:      data,
			},
		})

		if err != nil {
			logging.Error("Failed to publish subscribe response: %v", err)
			return fmt.Errorf("failed to publish subscribe response: %w", err)
		}

		return nil
	})

	consumer.Consume(ctx)
}
