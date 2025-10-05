package handlers

import "go-backend/pkg/websocket"

func RegisterHandlers(ws *websocket.WsServer) {
	registerSocketHandler(ws)
}
