package handlers

import (
	"context"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// AttachmentHandler 附件处理器
type AttachmentHandler struct {
}

// NewAttachmentHandler 创建新的附件处理器
func NewAttachmentHandler() *AttachmentHandler {
	return &AttachmentHandler{}
}

// GetAttachments 获取所有附件
func (h *AttachmentHandler) GetAttachments(c *gin.Context) {
	attachments, err := funcs.GetAllAttachments(context.Background())
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取附件列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachments,
		"count":   len(attachments),
	})
}

// GetAttachmentsWithPagination 分页获取附件列表
func (h *AttachmentHandler) GetAttachmentsWithPagination(c *gin.Context) {
	var req models.GetAttachmentsRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "desc"
	req.OrderBy = "create_time"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 调用服务层方法
	result, err := funcs.GetAttachmentsWithPagination(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取附件列表失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetAttachment 根据ID获取附件
func (h *AttachmentHandler) GetAttachment(c *gin.Context) {
	idStr := c.Param("id")

	// 验证ID参数
	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("附件ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	attachment, err := funcs.GetAttachmentByID(context.Background(), id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询附件失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachment,
	})
}

// CreateAttachment 创建附件
func (h *AttachmentHandler) CreateAttachment(c *gin.Context) {
	var req models.CreateAttachmentRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.Filename == "" {
		middleware.ThrowError(c, middleware.BadRequestError("文件名不能为空", nil))
		return
	}

	if req.Path == "" {
		middleware.ThrowError(c, middleware.BadRequestError("文件路径不能为空", nil))
		return
	}

	if req.ContentType == "" {
		middleware.ThrowError(c, middleware.BadRequestError("文件类型不能为空", nil))
		return
	}

	if req.Bucket == "" {
		middleware.ThrowError(c, middleware.BadRequestError("存储桶不能为空", nil))
		return
	}

	if req.Size <= 0 {
		middleware.ThrowError(c, middleware.BadRequestError("文件大小必须大于0", nil))
		return
	}

	attachment, err := funcs.CreateAttachment(context.Background(), &req)
	if err != nil {
		// 根据错误内容判断错误类型
		if err.Error() == "attachment already exists" {
			middleware.ThrowError(c, middleware.NewCustomError(middleware.ErrCodeConflict, "附件已存在", map[string]interface{}{
				"path": req.Path,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("创建附件失败", err.Error()))
		}
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    attachment,
		"message": "附件创建成功",
	})
}

// UpdateAttachment 更新附件
func (h *AttachmentHandler) UpdateAttachment(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	attachment, err := funcs.UpdateAttachment(context.Background(), id, &req)
	if err != nil {
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新附件失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachment,
		"message": "附件更新成功",
	})
}

// DeleteAttachment 删除附件
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]interface{}{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteAttachment(context.Background(), id)
	if err != nil {
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]interface{}{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除附件失败", err.Error()))
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "附件删除成功",
	})
}
