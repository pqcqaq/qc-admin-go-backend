package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/configs"
	"go-backend/pkg/s3"
	"go-backend/pkg/utils"
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
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	attachment, err := funcs.GetAttachmentByID(context.Background(), id)
	if err != nil {
		// 根据错误类型抛出不同的自定义错误
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]any{
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
			middleware.ThrowError(c, middleware.NewCustomError(middleware.ErrCodeConflict, "附件已存在", map[string]any{
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
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]any{
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
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]any{
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
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	err = funcs.DeleteAttachment(context.Background(), id)
	if err != nil {
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]any{
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

// GetAttachmentURL 获取附件的预签名URL
func (h *AttachmentHandler) GetAttachmentURL(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("附件ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	// 获取附件信息
	attachment, err := funcs.GetAttachmentByID(context.Background(), id)
	if err != nil {
		if err.Error() == "attachment not found" {
			middleware.ThrowError(c, middleware.NotFoundError("附件不存在", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询附件失败", err.Error()))
		}
		return
	}

	// 获取S3客户端
	s3Client := s3.GetClient()
	if s3Client == nil {
		middleware.ThrowError(c, middleware.InternalServerError("S3服务不可用", nil))
		return
	}

	// 生成预签名URL（有效期1小时）
	url, err := s3Client.GetFileURL(attachment.Bucket, attachment.Path, time.Hour)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("生成文件URL失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"id":       attachment.ID,
			"filename": attachment.Filename,
			"url":      url,
			"expires":  time.Now().Add(time.Hour).Unix(),
		},
	})
}

// PrepareUpload 准备文件上传，返回上传凭证
func (h *AttachmentHandler) PrepareUpload(c *gin.Context) {
	var req models.PrepareUploadRequest

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

	if req.ContentType == "" {
		middleware.ThrowError(c, middleware.BadRequestError("文件类型不能为空", nil))
		return
	}

	if req.Size <= 0 {
		middleware.ThrowError(c, middleware.BadRequestError("文件大小必须大于0", nil))
		return
	}

	// 调用业务逻辑
	result, err := funcs.PrepareUpload(context.Background(), &req)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("准备上传失败", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
		"message": "上传凭证生成成功",
	})
}

// ConfirmUpload 确认文件上传完成
func (h *AttachmentHandler) ConfirmUpload(c *gin.Context) {
	var req models.ConfirmUploadRequest

	// 数据绑定验证
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	// 业务逻辑验证
	if req.UploadSessionID == "" {
		middleware.ThrowError(c, middleware.BadRequestError("上传会话ID不能为空", nil))
		return
	}

	// 调用业务逻辑
	attachment, err := funcs.ConfirmUpload(context.Background(), &req)
	if err != nil {
		if err.Error() == "upload session not found" {
			middleware.ThrowError(c, middleware.NotFoundError("上传会话不存在", map[string]any{
				"upload_session_id": req.UploadSessionID,
			}))
		} else if err.Error() == "invalid upload session status" {
			middleware.ThrowError(c, middleware.BadRequestError("无效的上传会话状态", map[string]any{
				"upload_session_id": req.UploadSessionID,
			}))
		} else {
			middleware.ThrowError(c, middleware.InternalServerError("确认上传失败", err.Error()))
		}
		return
	}

	// 转换为响应格式
	attachmentResponse := &models.AttachmentResponse{
		ID:              utils.Uint64ToString(attachment.ID),
		CreateTime:      attachment.CreateTime.Format(time.RFC3339),
		UpdateTime:      attachment.UpdateTime.Format(time.RFC3339),
		Filename:        attachment.Filename,
		Path:            attachment.Path,
		URL:             attachment.URL,
		ContentType:     attachment.ContentType,
		Size:            attachment.Size,
		Etag:            attachment.Etag,
		Bucket:          attachment.Bucket,
		StorageProvider: attachment.StorageProvider,
		Metadata:        attachment.Metadata,
		Status:          string(attachment.Status),
		UploadSessionID: attachment.UploadSessionID,
		Tag1:            attachment.Tag1,
		Tag2:            attachment.Tag2,
		Tag3:            attachment.Tag3,
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachmentResponse,
		"message": "文件上传确认成功",
	})
}

// DirectUpload 直接上传文件表单
func (h *AttachmentHandler) DirectUpload(c *gin.Context) {
	// 解析表单数据
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("解析表单数据失败", err.Error()))
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("获取上传文件失败", err.Error()))
		return
	}
	defer file.Close()

	// 验证文件
	if header.Size <= 0 {
		middleware.ThrowError(c, middleware.BadRequestError("文件大小无效", nil))
		return
	}

	// 获取其他表单参数
	bucket := c.PostForm("bucket")
	if bucket == "" {
		config := configs.GetConfig()
		bucket = config.S3.Bucket
	}

	tag1 := c.PostForm("tag1")
	tag2 := c.PostForm("tag2")
	tag3 := c.PostForm("tag3")

	// 生成文件路径
	timestamp := time.Now().Format("2006/01/02")
	fileName := header.Filename
	filePath := fmt.Sprintf("uploads/%s/%d_%s", timestamp, time.Now().UnixNano(), fileName)

	// 获取S3客户端
	s3Client := s3.GetClient()
	if s3Client == nil {
		middleware.ThrowError(c, middleware.InternalServerError("S3服务不可用", nil))
		return
	}

	// 上传文件到S3
	uploadResult, err := s3Client.UploadFile(bucket, filePath, file, header.Header.Get("Content-Type"))
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("文件上传失败", err.Error()))
		return
	}

	// 创建附件记录
	createReq := &models.CreateAttachmentRequest{
		Filename:        fileName,
		Path:            filePath,
		ContentType:     header.Header.Get("Content-Type"),
		Size:            header.Size,
		Bucket:          bucket,
		StorageProvider: "s3",
		Status:          "uploaded",
		Tag1:            tag1,
		Tag2:            tag2,
		Tag3:            tag3,
	}

	// 如果上传结果包含ETag，设置它
	if uploadResult.ETag != nil {
		createReq.Etag = *uploadResult.ETag
	}

	// 如果上传结果包含Location，设置它
	if uploadResult.Location != "" {
		createReq.URL = uploadResult.Location
	}

	// 保存到数据库
	attachment, err := funcs.CreateAttachment(context.Background(), createReq)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("保存附件信息失败", err.Error()))
		return
	}

	// 转换为响应格式
	attachmentResponse := &models.AttachmentResponse{
		ID:              utils.Uint64ToString(attachment.ID),
		CreateTime:      attachment.CreateTime.Format(time.RFC3339),
		UpdateTime:      attachment.UpdateTime.Format(time.RFC3339),
		Filename:        attachment.Filename,
		Path:            attachment.Path,
		URL:             attachment.URL,
		ContentType:     attachment.ContentType,
		Size:            attachment.Size,
		Etag:            attachment.Etag,
		Bucket:          attachment.Bucket,
		StorageProvider: attachment.StorageProvider,
		Metadata:        attachment.Metadata,
		Status:          string(attachment.Status),
		UploadSessionID: attachment.UploadSessionID,
		Tag1:            attachment.Tag1,
		Tag2:            attachment.Tag2,
		Tag3:            attachment.Tag3,
	}

	response := &models.DirectUploadResponse{
		Success:    true,
		Message:    "文件上传成功",
		Attachment: attachmentResponse,
	}

	c.JSON(200, response)
}
