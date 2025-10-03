package handlers

import (
	"fmt"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
)

var sender MessageSender

func SetSender(s MessageSender) {
	sender = s
}

func handleMessage(message messaging.MessageStruct) error {
	socketMsgMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		logging.Error("Invalid message payload type")
		return fmt.Errorf("invalid message payload type")
	}

	var socketMsg messaging.SocketMessagePayload
	err := utils.MapToStruct(socketMsgMap, &socketMsg)
	if err != nil {
		logging.Error("Failed to convert payload to SocketMessagePayload: %v", err)
		return fmt.Errorf("failed to convert payload to SocketMessagePayload: %w", err)
	}

	// 发送消息到wsSender
	if sender != nil {
		err := sender(socketMsg)
		if err != nil {
			logging.Error("Failed to send message via WebSocket: %v", err)
			return err
		}
	} else {
		logging.Error("Message sender is not set")
		return fmt.Errorf("message sender is not set")
	}
	return nil
}

func registerSocketHandler() {
	messaging.RegisterHandler(messaging.ServerToUserSocket, handleMessage)
}
