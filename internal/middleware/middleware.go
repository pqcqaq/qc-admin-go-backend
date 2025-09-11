package middleware

import (
	"go-backend/pkg/configs"

	"github.com/gin-gonic/gin"
)

func RegisterConfigMiddlewares(r *gin.Engine) {
	middleWareConfig := configs.GetConfig().Server.Middleware
	// 延迟中间件
	if middleWareConfig.Delay.Enabled {
		r.Use(DelayMiddleware())
	}
}
