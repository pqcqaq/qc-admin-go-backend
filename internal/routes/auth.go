package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupAuthRoutes 设置认证相关路由
func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	authHandler := handlers.NewAuthHandler()

	auth := rg.Group("/auth")
	{
		// 公开路由（不需要认证）
		auth.POST("/send-verify-code", authHandler.SendVerifyCode)
		auth.POST("/verify-code", authHandler.VerifyCode)
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/reset-password", authHandler.ResetPassword)

		// 需要认证（配置为非public）的路由
		auth.POST("/refresh-token", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/user-info", authHandler.GetUserInfo)
		auth.GET("/user-menu-tree", authHandler.GetUserMenuTree)
	}
}
