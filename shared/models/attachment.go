package models

// CreateAttachmentRequest 创建附件请求结构
type CreateAttachmentRequest struct {
	Filename        string                 `json:"filename" binding:"required"`
	Path            string                 `json:"path" binding:"required"`
	URL             string                 `json:"url,omitempty"`
	ContentType     string                 `json:"content_type" binding:"required"`
	Size            int64                  `json:"size" binding:"required,min=1"`
	Etag            string                 `json:"etag,omitempty"`
	Bucket          string                 `json:"bucket" binding:"required"`
	StorageProvider string                 `json:"storage_provider,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Status          string                 `json:"status,omitempty"`
	UploadSessionID string                 `json:"upload_session_id,omitempty"`
	Tag1            string                 `json:"tag1,omitempty"`
	Tag2            string                 `json:"tag2,omitempty"`
	Tag3            string                 `json:"tag3,omitempty"`
}

// UpdateAttachmentRequest 更新附件请求结构
type UpdateAttachmentRequest struct {
	Filename        string                 `json:"filename,omitempty"`
	Path            string                 `json:"path,omitempty"`
	URL             string                 `json:"url,omitempty"`
	ContentType     string                 `json:"content_type,omitempty"`
	Size            int64                  `json:"size,omitempty"`
	Etag            string                 `json:"etag,omitempty"`
	Bucket          string                 `json:"bucket,omitempty"`
	StorageProvider string                 `json:"storage_provider,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Status          string                 `json:"status,omitempty"`
	UploadSessionID string                 `json:"upload_session_id,omitempty"`
	Tag1            string                 `json:"tag1,omitempty"`
	Tag2            string                 `json:"tag2,omitempty"`
	Tag3            string                 `json:"tag3,omitempty"`
}

// AttachmentResponse 附件响应结构
type AttachmentResponse struct {
	ID              uint64                 `json:"id"`
	CreateTime      string                 `json:"create_time"`
	UpdateTime      string                 `json:"update_time"`
	Filename        string                 `json:"filename"`
	Path            string                 `json:"path"`
	URL             string                 `json:"url,omitempty"`
	ContentType     string                 `json:"content_type"`
	Size            int64                  `json:"size"`
	Etag            string                 `json:"etag,omitempty"`
	Bucket          string                 `json:"bucket"`
	StorageProvider string                 `json:"storage_provider"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Status          string                 `json:"status"`
	UploadSessionID string                 `json:"upload_session_id,omitempty"`
	Tag1            string                 `json:"tag1,omitempty"`
	Tag2            string                 `json:"tag2,omitempty"`
	Tag3            string                 `json:"tag3,omitempty"`
}

// GetAttachmentsRequest 获取附件列表请求结构
type GetAttachmentsRequest struct {
	PaginationRequest
	Filename        string `form:"filename" json:"filename"`                 // 按文件名模糊搜索
	ContentType     string `form:"content_type" json:"content_type"`         // 按内容类型过滤
	Status          string `form:"status" json:"status"`                     // 按状态过滤
	Bucket          string `form:"bucket" json:"bucket"`                     // 按存储桶过滤
	StorageProvider string `form:"storage_provider" json:"storage_provider"` // 按存储提供商过滤
	Tag1            string `form:"tag1" json:"tag1"`                         // 按标签1过滤
	Tag2            string `form:"tag2" json:"tag2"`                         // 按标签2过滤
	Tag3            string `form:"tag3" json:"tag3"`                         // 按标签3过滤
}

// AttachmentsListResponse 附件列表响应结构
type AttachmentsListResponse struct {
	Data       []*AttachmentResponse `json:"data"`
	Pagination Pagination            `json:"pagination"`
}
