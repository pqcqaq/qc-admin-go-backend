package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupLoginRecordRoutes 设置登录记录相关路由
func (r *Router) setupLoginRecordRoutes(rg *gin.RouterGroup) {
	loginRecordHandler := handlers.NewLoginRecordHandler()

	// 用户可以查看自己的登录记录
	auth := rg.Group("/auth")
	auth.GET("/login-records", loginRecordHandler.GetUserLoginRecords)

	// 管理员可以查看所有登录记录
	admin := rg.Group("/admin")
	admin.GET("/login-records", loginRecordHandler.GetLoginRecords)
}
