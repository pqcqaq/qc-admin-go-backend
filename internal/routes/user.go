package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupUserRoutes 设置用户相关路由
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {

	userHandler := handlers.NewUserHandler()

	users := rg.Group("/users")
	{
		users.GET("", userHandler.GetUsers)
		users.GET("/page", userHandler.GetUsersWithPagination)
		users.GET("/:id", userHandler.GetUser)
		users.POST("", userHandler.CreateUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}
}
