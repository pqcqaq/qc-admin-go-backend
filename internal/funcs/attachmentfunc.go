package funcs

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/attachment"
	"go-backend/pkg/database"
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
	if req.Tag1 != "" {
		query = query.Where(attachment.Tag1Contains(req.Tag1))
	}
	if req.Tag2 != "" {
		query = query.Where(attachment.Tag2Contains(req.Tag2))
	}
	if req.Tag3 != "" {
		query = query.Where(attachment.Tag3Contains(req.Tag3))
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
			ID:              a.ID,
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
