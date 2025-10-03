package handlers

import (
	"fmt"
	"go-backend/internal/middleware"
	"go-backend/pkg/messaging"

	"github.com/gin-gonic/gin"
)

type TestHandler struct {
}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

func (h *TestHandler) TestSendUserSocketMsg(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	userId, ex := middleware.GetUserIDFromContext(ctx)
	if !ex {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	messaging.Publish(ctx, messaging.MessageStruct{
		Type: messaging.ServerToUserSocket,
		Payload: messaging.SocketMessagePayload{
			UserId: &userId,
			Topic:  "test_topic",
			Data:   fmt.Sprintf("Hello User %d! This is a test message from server.", userId),
		},
		Priority: 1,
	})
	c.JSON(200, gin.H{"status": "message sent"})
}
