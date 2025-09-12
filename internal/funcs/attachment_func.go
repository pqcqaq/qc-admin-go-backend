package funcs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/attachment"
	"go-backend/pkg/configs"
	"go-backend/pkg/database"
	"go-backend/pkg/s3"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// GetAllAttachments 获取所有附件
func GetAllAttachments(ctx context.Context) ([]*ent.Attachment, error) {
	return database.Client.Attachment.Query().All(ctx)
}

// GetAttachmentByID 根据ID获取附件
func GetAttachmentByID(ctx context.Context, id uint64) (*ent.Attachment, error) {
	attachment, err := database.Client.Attachment.Query().Where(attachment.ID(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("attachment with id %d not found", id)
		}
		return nil, err
	}
	return attachment, nil
}

// CreateAttachment 创建附件
func CreateAttachment(ctx context.Context, req *models.CreateAttachmentRequest) (*ent.Attachment, error) {
	builder := database.Client.Attachment.Create().
		SetFilename(req.Filename).
		SetPath(req.Path).
		SetContentType(req.ContentType).
		SetSize(req.Size).
		SetBucket(req.Bucket)

	// 设置可选字段
	if req.URL != "" {
		builder = builder.SetURL(req.URL)
	}
	if req.Etag != "" {
		builder = builder.SetEtag(req.Etag)
	}
	if req.StorageProvider != "" {
		builder = builder.SetStorageProvider(req.StorageProvider)
	}
	if req.Metadata != nil {
		builder = builder.SetMetadata(req.Metadata)
	}
	if req.Status != "" {
		builder = builder.SetStatus(attachment.Status(req.Status))
	}
	if req.UploadSessionID != "" {
		builder = builder.SetUploadSessionID(req.UploadSessionID)
	}
	if req.Tag1 != "" {
		builder = builder.SetTag1(req.Tag1)
	}
	if req.Tag2 != "" {
		builder = builder.SetTag2(req.Tag2)
	}
	if req.Tag3 != "" {
		builder = builder.SetTag3(req.Tag3)
	}

	return builder.Save(ctx)
}

// UpdateAttachment 更新附件
func UpdateAttachment(ctx context.Context, id uint64, req *models.UpdateAttachmentRequest) (*ent.Attachment, error) {
	// 首先检查附件是否存在
	exists, err := database.Client.Attachment.Query().Where(attachment.ID(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("attachment with id %d not found", id)
	}

	builder := database.Client.Attachment.UpdateOneID(id)

	// 更新字段
	if req.Filename != "" {
		builder = builder.SetFilename(req.Filename)
	}
	if req.Path != "" {
		builder = builder.SetPath(req.Path)
	}
	if req.URL != "" {
		builder = builder.SetURL(req.URL)
	}
	if req.ContentType != "" {
		builder = builder.SetContentType(req.ContentType)
	}
	if req.Size > 0 {
		builder = builder.SetSize(req.Size)
	}
	if req.Etag != "" {
		builder = builder.SetEtag(req.Etag)
	}
	if req.Bucket != "" {
		builder = builder.SetBucket(req.Bucket)
	}
	if req.StorageProvider != "" {
		builder = builder.SetStorageProvider(req.StorageProvider)
	}
	if req.Metadata != nil {
		builder = builder.SetMetadata(req.Metadata)
	}
	if req.Status != "" {
		builder = builder.SetStatus(attachment.Status(req.Status))
	}
	if req.UploadSessionID != "" {
		builder = builder.SetUploadSessionID(req.UploadSessionID)
	}
	if req.Tag1 != "" {
		builder = builder.SetTag1(req.Tag1)
	}
	if req.Tag2 != "" {
		builder = builder.SetTag2(req.Tag2)
	}
	if req.Tag3 != "" {
		builder = builder.SetTag3(req.Tag3)
	}

	return builder.Save(ctx)
}

// DeleteAttachment 删除附件
func DeleteAttachment(ctx context.Context, id uint64) error {
	// 首先检查附件是否存在
	exists, err := database.Client.Attachment.Query().Where(attachment.ID(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("attachment with id %d not found", id)
	}

	return database.Client.Attachment.DeleteOneID(id).Exec(ctx)
}

// GetAttachmentsWithPagination 分页获取附件列表
func GetAttachmentsWithPagination(ctx context.Context, req *models.GetAttachmentsRequest) (*models.AttachmentsListResponse, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.OrderBy == "" {
		req.OrderBy = "create_time"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	// 构建查询条件
	query := database.Client.Attachment.Query()

	// 添加搜索条件
	if req.Filename != "" {
		query = query.Where(attachment.FilenameContains(req.Filename))
	}
	if req.ContentType != "" {
		query = query.Where(attachment.ContentTypeContains(req.ContentType))
	}
	if req.Status != "" {
		query = query.Where(attachment.StatusEQ(attachment.Status(req.Status)))
	}
	if req.Bucket != "" {
		query = query.Where(attachment.BucketContains(req.Bucket))
	}
	if req.StorageProvider != "" {
		query = query.Where(attachment.StorageProviderContains(req.StorageProvider))
	}
	if req.Tags != "" {
		query = query.Where(attachment.Or(attachment.Tag1Contains(req.Tags), attachment.Tag2Contains(req.Tags), attachment.Tag3Contains(req.Tags)))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count attachments: %w", err)
	}

	// 计算分页信息
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	offset := (req.Page - 1) * req.PageSize

	// 添加排序和分页
	switch strings.ToLower(req.OrderBy) {
	case "id":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldID))
		} else {
			query = query.Order(ent.Asc(attachment.FieldID))
		}
	case "filename":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldFilename))
		} else {
			query = query.Order(ent.Asc(attachment.FieldFilename))
		}
	case "content_type":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldContentType))
		} else {
			query = query.Order(ent.Asc(attachment.FieldContentType))
		}
	case "size":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldSize))
		} else {
			query = query.Order(ent.Asc(attachment.FieldSize))
		}
	case "status":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldStatus))
		} else {
			query = query.Order(ent.Asc(attachment.FieldStatus))
		}
	case "create_time", "created_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldCreateTime))
		} else {
			query = query.Order(ent.Asc(attachment.FieldCreateTime))
		}
	case "update_time", "updated_at":
		if strings.ToLower(req.Order) == "desc" {
			query = query.Order(ent.Desc(attachment.FieldUpdateTime))
		} else {
			query = query.Order(ent.Asc(attachment.FieldUpdateTime))
		}
	default:
		// 默认按创建时间降序排列
		query = query.Order(ent.Desc(attachment.FieldCreateTime))
	}

	// 执行分页查询
	attachments, err := query.
		Offset(offset).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}

	// 转换为响应格式
	attachmentResponses := make([]*models.AttachmentResponse, len(attachments))
	for i, a := range attachments {
		attachmentResponses[i] = &models.AttachmentResponse{
			ID:              utils.Uint64ToString(a.ID),
			CreateTime:      a.CreateTime.Format(time.RFC3339),
			UpdateTime:      a.UpdateTime.Format(time.RFC3339),
			Filename:        a.Filename,
			Path:            a.Path,
			URL:             a.URL,
			ContentType:     a.ContentType,
			Size:            a.Size,
			Etag:            a.Etag,
			Bucket:          a.Bucket,
			StorageProvider: a.StorageProvider,
			Metadata:        a.Metadata,
			Status:          string(a.Status),
			UploadSessionID: a.UploadSessionID,
			Tag1:            a.Tag1,
			Tag2:            a.Tag2,
			Tag3:            a.Tag3,
		}
	}

	// 构建分页信息
	pagination := models.Pagination{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      int64(total),
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &models.AttachmentsListResponse{
		Data:       attachmentResponses,
		Pagination: pagination,
	}, nil
}

// generateUploadSessionID 生成上传会话ID
func generateUploadSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// PrepareUpload 准备文件上传
func PrepareUpload(ctx context.Context, req *models.PrepareUploadRequest) (*models.PrepareUploadResponse, error) {
	// 生成上传会话ID
	uploadSessionID := generateUploadSessionID()

	// 生成文件路径
	timestamp := time.Now().Format("2006/01/02")
	filePath := fmt.Sprintf("uploads/%s/%s_%s", timestamp, uploadSessionID, req.Filename)

	// 获取配置中的默认bucket
	config := configs.GetConfig()
	bucket := req.Bucket
	if bucket == "" {
		bucket = config.S3.Bucket
	}

	// 在数据库中创建uploading状态的附件记录
	attachmentRecord, err := database.Client.Attachment.Create().
		SetFilename(req.Filename).
		SetPath(filePath).
		SetContentType(req.ContentType).
		SetSize(req.Size).
		SetBucket(bucket).
		SetStatus(attachment.StatusUploading).
		SetUploadSessionID(uploadSessionID).
		SetTag1(req.Tag1).
		SetTag2(req.Tag2).
		SetTag3(req.Tag3).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	// 使用S3客户端生成预签名URL
	s3Client := s3.GetClient()
	uploadURL, err := s3Client.GetPresignedPutURL(bucket, filePath, time.Hour, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &models.PrepareUploadResponse{
		UploadURL:       uploadURL,
		UploadSessionID: uploadSessionID,
		ExpiresAt:       time.Now().Add(time.Hour).Unix(),
		AttachmentID:    utils.ToString(attachmentRecord.ID),
	}, nil
}

// ConfirmUpload 确认文件上传完成
func ConfirmUpload(ctx context.Context, req *models.ConfirmUploadRequest) (*ent.Attachment, error) {
	// 根据上传会话ID查找附件
	attachmentRecord, err := database.Client.Attachment.Query().
		Where(attachment.UploadSessionIDEQ(req.UploadSessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("upload session not found")
		}
		return nil, fmt.Errorf("failed to query attachment: %w", err)
	}

	// 检查附件状态
	if attachmentRecord.Status != attachment.StatusUploading {
		return nil, fmt.Errorf("invalid upload session status: %s", attachmentRecord.Status)
	}

	// 获取S3客户端和配置
	s3Client := s3.GetClient()
	config := configs.GetConfig()

	// 从S3获取文件的真实信息
	fileInfo, err := s3Client.GetFileInfo(attachmentRecord.Bucket, attachmentRecord.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info from S3: %w", err)
	}

	// 生成文件的访问URL
	var fileURL string
	if config.S3.Endpoint != "" {
		var trueEndpoint string
		if utils.IsEmpty(config.S3.PublicEndpoint) {
			trueEndpoint = config.S3.Endpoint
		} else {
			trueEndpoint = config.S3.PublicEndpoint
		}
		// 使用自定义端点（如MinIO）
		fileURL = fmt.Sprintf("%s/%s/%s", trueEndpoint, attachmentRecord.Bucket, attachmentRecord.Path)
	} else {
		// 使用AWS S3标准URL格式
		fileURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", attachmentRecord.Bucket, config.S3.Region, attachmentRecord.Path)
	}

	// 更新附件状态为已完成，使用从S3获取的真实数据
	builder := database.Client.Attachment.UpdateOneID(attachmentRecord.ID).
		SetStatus(attachment.StatusUploaded).
		SetURL(fileURL).
		SetStorageProvider("s3")

	// 使用S3返回的真实文件大小
	if fileInfo.ContentLength != nil {
		builder = builder.SetSize(*fileInfo.ContentLength)
	}

	// 使用S3返回的ETag（优先级高于客户端传递的）
	if fileInfo.ETag != nil {
		etag := *fileInfo.ETag
		// 移除ETag的双引号（如果存在）
		if len(etag) > 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
			etag = etag[1 : len(etag)-1]
		}
		builder = builder.SetEtag(etag)
	} else if req.Etag != "" {
		// 如果S3没有返回ETag，才使用客户端提供的
		builder = builder.SetEtag(req.Etag)
	}

	// 同步内容类型（如果S3有提供）
	if fileInfo.ContentType != nil && *fileInfo.ContentType != "" {
		builder = builder.SetContentType(*fileInfo.ContentType)
	}

	// 同步最后修改时间
	if fileInfo.LastModified != nil {
		// 可以将S3的最后修改时间存储到元数据中
		metadata := attachmentRecord.Metadata
		if metadata == nil {
			metadata = make(map[string]any)
		}
		metadata["s3_last_modified"] = fileInfo.LastModified.Format(time.RFC3339)
		builder = builder.SetMetadata(metadata)
	}

	// 验证文件大小（如果客户端提供了实际大小，检查是否匹配）
	if req.ActualSize > 0 && fileInfo.ContentLength != nil && req.ActualSize != *fileInfo.ContentLength {
		return nil, fmt.Errorf("file size mismatch: client reported %d bytes, S3 has %d bytes", req.ActualSize, *fileInfo.ContentLength)
	}

	return builder.Save(ctx)
}

// GetAttachmentByUploadSessionID 根据上传会话ID获取附件
func GetAttachmentByUploadSessionID(ctx context.Context, uploadSessionID string) (*ent.Attachment, error) {
	attachment, err := database.Client.Attachment.Query().
		Where(attachment.UploadSessionIDEQ(uploadSessionID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("upload session not found")
		}
		return nil, err
	}
	return attachment, nil
}
