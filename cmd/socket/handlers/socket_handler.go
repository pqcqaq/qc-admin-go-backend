package handlers

import (
	"fmt"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/utils"
	"go-backend/pkg/websocket"
	"go-backend/pkg/websocket/types"
)

var sender types.MessageSender

func SetSender(s types.MessageSender) {
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

func testChatHandler(message messaging.MessageStruct) error {
	socketMsgMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		logging.Error("Invalid message payload type")
		return fmt.Errorf("invalid message payload type")
	}

	var socketMsg messaging.UserMessagePayload
	err := utils.MapToStruct(socketMsgMap, &socketMsg)
	if err != nil {
		logging.Error("Failed to convert payload to SocketMessagePayload: %v", err)
		return fmt.Errorf("failed to convert payload to SocketMessagePayload: %w", err)
	}

	logging.WithName("Test_Chat_Handler").Info("Received test chat message from user %d: %s, content is: %v", socketMsg.UserId, socketMsg.MessageId, socketMsg.Data)

	return nil
}

func createChannelUserMsgHandler(ws *websocket.WsServer) messaging.MessageHandler {
	return func(message messaging.MessageStruct) error {
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

		channel := ws.GetChannelById(socketMsg.ID)
		if channel == nil {
			return fmt.Errorf("channel %s not found", socketMsg.ID)
		}

		if socketMsg.Action == messaging.ChannelActionClose {
			channel.Close()
			return nil
		}

		if socketMsg.Action == messaging.ChannelActionErr {
			clients := ws.GetClientFromChannelId(channel.ID)
			for _, client := range clients {
				client.SendChannelError(channel.ID, websocket.ErrInternalServer, fmt.Errorf("%s", socketMsg.Data))
			}
			return nil
		}

		msg := channel.NewMessage(socketMsg.Data)
		if msg != nil {
			msg.ToClient()
		}
		return nil
	}
}

func registerSocketHandler(ws *websocket.WsServer) {
	messaging.RegisterHandler(messaging.ServerToUserSocket, handleMessage)
	messaging.RegisterHandler(messaging.UserToServerSocket, testChatHandler)
	messaging.RegisterHandler(messaging.ChannelToUser, createChannelUserMsgHandler(ws))
	RegisterIsolate()
	RegisterChat()
	RegisterPanicTest()
}
