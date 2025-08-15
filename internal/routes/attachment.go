package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupAttachmentRoutes 设置附件相关路由
func (r *Router) setupAttachmentRoutes(rg *gin.RouterGroup) {

	attachmentHandler := handlers.NewAttachmentHandler()

	attachments := rg.Group("/attachments")
	{
		attachments.GET("", attachmentHandler.GetAttachments)
		attachments.GET("/page", attachmentHandler.GetAttachmentsWithPagination) // 分页查询路由
		attachments.GET("/:id", attachmentHandler.GetAttachment)
		attachments.POST("", attachmentHandler.CreateAttachment)
		attachments.PUT("/:id", attachmentHandler.UpdateAttachment)
		attachments.DELETE("/:id", attachmentHandler.DeleteAttachment)
	}
}
