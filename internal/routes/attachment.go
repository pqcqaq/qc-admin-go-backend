package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupAttachmentRoutes 设置附件相关路由
func (r *Router) setupAttachmentRoutes(rg *gin.RouterGroup) {

	attachmentHandler := handlers.NewAttachmentHandler()

	attachments := rg.Group("/attachments")
	// 全部需要auth中间件保护
	attachments.GET("", attachmentHandler.GetAttachments)
	attachments.GET("/page", attachmentHandler.GetAttachmentsWithPagination) // 分页查询路由
	attachments.GET("/:id", attachmentHandler.GetAttachment)
	attachments.GET("/:id/url", attachmentHandler.GetAttachmentURL) // 获取文件预签名URL
	attachments.POST("", attachmentHandler.CreateAttachment)
	attachments.PUT("/:id", attachmentHandler.UpdateAttachment)
	attachments.DELETE("/:id", attachmentHandler.DeleteAttachment)

	// 第一类功能：分离式上传
	attachments.POST("/prepare-upload", attachmentHandler.PrepareUpload) // 准备上传，返回上传凭证
	attachments.POST("/confirm-upload", attachmentHandler.ConfirmUpload) // 确认上传完成

	// 第二类功能：直接上传
	attachments.POST("/direct-upload", attachmentHandler.DirectUpload) // 直接上传文件表单
}
