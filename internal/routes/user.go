package routes

import (
	"go-backend/internal/handlers"
	"go-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// setupUserRoutes 设置用户相关路由
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {

	userHandler := handlers.NewUserHandler()

	users := rg.Group("/users")
	// 全部需要auth中间件保护
	users.Use(middleware.JWTAuthMiddleware())
	{
		users.GET("", userHandler.GetUsers)
		users.GET("/page", userHandler.GetUsersWithPagination)
		users.GET("/:id", userHandler.GetUser)
		users.POST("", userHandler.CreateUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
		// 配置用户头像
		users.POST("/:id/avatar/:attachment_id", userHandler.SetUserAvatar)
	}
}
