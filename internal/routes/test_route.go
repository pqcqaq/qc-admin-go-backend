package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func (r *Router) setupTestRoutes(rg *gin.RouterGroup) {

	testHandler := handlers.NewTestHandler()
	test := rg.Group("/test")
	{
		test.GET("/send-user-socket-msg", testHandler.TestSendUserSocketMsg)
	}
}
