package models

// CreateAttachmentRequest 创建附件请求结构
type CreateAttachmentRequest struct {
	Filename        string         `json:"filename" binding:"required"`
	Path            string         `json:"path" binding:"required"`
	URL             string         `json:"url,omitempty"`
	ContentType     string         `json:"contentType" binding:"required"`
	Size            int64          `json:"size" binding:"required,min=1"`
	Etag            string         `json:"etag,omitempty"`
	Bucket          string         `json:"bucket" binding:"required"`
	StorageProvider string         `json:"storageProvider,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	Status          string         `json:"status,omitempty"`
	UploadSessionID string         `json:"uploadSessionId,omitempty"`
	Tag1            string         `json:"tag1,omitempty"`
	Tag2            string         `json:"tag2,omitempty"`
	Tag3            string         `json:"tag3,omitempty"`
}

// UpdateAttachmentRequest 更新附件请求结构
type UpdateAttachmentRequest struct {
	Filename        string         `json:"filename,omitempty"`
	Path            string         `json:"path,omitempty"`
	URL             string         `json:"url,omitempty"`
	ContentType     string         `json:"contentType,omitempty"`
	Size            int64          `json:"size,omitempty"`
	Etag            string         `json:"etag,omitempty"`
	Bucket          string         `json:"bucket,omitempty"`
	StorageProvider string         `json:"storageProvider,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	Status          string         `json:"status,omitempty"`
	UploadSessionID string         `json:"uploadSessionId,omitempty"`
	Tag1            string         `json:"tag1,omitempty"`
	Tag2            string         `json:"tag2,omitempty"`
	Tag3            string         `json:"tag3,omitempty"`
}

// AttachmentResponse 附件响应结构
type AttachmentResponse struct {
	ID              string         `json:"id"`
	CreateTime      string         `json:"createTime"`
	UpdateTime      string         `json:"updateTime"`
	Filename        string         `json:"filename"`
	Path            string         `json:"path"`
	URL             string         `json:"url,omitempty"`
	ContentType     string         `json:"contentType"`
	Size            int64          `json:"size"`
	Etag            string         `json:"etag,omitempty"`
	Bucket          string         `json:"bucket"`
	StorageProvider string         `json:"storageProvider"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	Status          string         `json:"status"`
	UploadSessionID string         `json:"uploadSessionId,omitempty"`
	Tag1            string         `json:"tag1,omitempty"`
	Tag2            string         `json:"tag2,omitempty"`
	Tag3            string         `json:"tag3,omitempty"`
}

// GetAttachmentsRequest 获取附件列表请求结构
type GetAttachmentsRequest struct {
	PaginationRequest
	Filename        string `form:"filename" json:"filename"`               // 按文件名模糊搜索
	ContentType     string `form:"contentType" json:"contentType"`         // 按内容类型过滤
	Status          string `form:"status" json:"status"`                   // 按状态过滤
	Bucket          string `form:"bucket" json:"bucket"`                   // 按存储桶过滤
	StorageProvider string `form:"storageProvider" json:"storageProvider"` // 按存储提供商过滤
	Tags            string `form:"tags" json:"tags"`                       // 按标签过滤
}

// AttachmentsListResponse 附件列表响应结构
type AttachmentsListResponse struct {
	Data       []*AttachmentResponse `json:"data"`
	Pagination Pagination            `json:"pagination"`
}

// PrepareUploadRequest 准备上传请求结构
type PrepareUploadRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	Size        int64  `json:"size" binding:"required,min=1"`
	Bucket      string `json:"bucket,omitempty"`
	Tag1        string `json:"tag1,omitempty"`
	Tag2        string `json:"tag2,omitempty"`
	Tag3        string `json:"tag3,omitempty"`
}

// PrepareUploadResponse 准备上传响应结构
type PrepareUploadResponse struct {
	UploadURL       string         `json:"uploadUrl"`
	UploadSessionID string         `json:"uploadSessionId"`
	Fields          map[string]any `json:"fields,omitempty"` // 用于表单上传的额外字段
	ExpiresAt       int64          `json:"expiresAt"`
	AttachmentID    string         `json:"attachmentId"`
}

// ConfirmUploadRequest 确认上传请求结构
type ConfirmUploadRequest struct {
	UploadSessionID string `json:"uploadSessionId" binding:"required"`
	Etag            string `json:"etag,omitempty"`
	ActualSize      int64  `json:"actualSize,omitempty"`
}

// DirectUploadResponse 直接上传响应结构
type DirectUploadResponse struct {
	Success    bool                `json:"success"`
	Message    string              `json:"message"`
	Attachment *AttachmentResponse `json:"attachment"`
}
