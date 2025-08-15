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
		attachments.GET("/paginated", attachmentHandler.GetAttachmentsWithPagination) // 分页查询路由
		attachments.GET("/:id", attachmentHandler.GetAttachment)
		attachments.GET("/:id/url", attachmentHandler.GetAttachmentURL) // 获取文件预签名URL
		attachments.POST("", attachmentHandler.CreateAttachment)
		attachments.PUT("/:id", attachmentHandler.UpdateAttachment)
		attachments.DELETE("/:id", attachmentHandler.DeleteAttachment)
	}
}
