package models

type LoggingResponse struct {
	ID         string                 `json:"id"`
	Level      string                 `json:"level"`
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	Query      string                 `json:"query,omitempty"`
	Code       int                    `json:"code,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Stack      string                 `json:"stack,omitempty"`
	CreateTime string                 `json:"createTime"`
	UpdateTime string                 `json:"updateTime"`
}

type CreateLoggingRequest struct {
	Level     string                 `json:"level" binding:"omitempty,oneof=debug info error warn fatal"`
	Type      string                 `json:"type" binding:"omitempty,oneof=Error Panic manul"`
	Message   string                 `json:"message" binding:"required,max=500"`
	Method    string                 `json:"method,omitempty" binding:"omitempty,max=127"`
	Path      string                 `json:"path,omitempty" binding:"omitempty,max=255"`
	IP        string                 `json:"ip,omitempty" binding:"omitempty,max=45"`
	Query     string                 `json:"query,omitempty" binding:"omitempty,max=1000"`
	Code      *int                   `json:"code,omitempty" binding:"omitempty,min=1"`
	UserAgent string                 `json:"userAgent,omitempty" binding:"omitempty,max=512"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Stack     string                 `json:"stack,omitempty" binding:"omitempty,max=8192"`
}

type UpdateLoggingRequest struct {
	Level     string                 `json:"level" binding:"omitempty,oneof=debug info error warn fatal"`
	Type      string                 `json:"type" binding:"omitempty,oneof=Error Panic manul"`
	Message   string                 `json:"message" binding:"required,max=500"`
	Method    string                 `json:"method,omitempty" binding:"omitempty,max=127"`
	Path      string                 `json:"path,omitempty" binding:"omitempty,max=255"`
	IP        string                 `json:"ip,omitempty" binding:"omitempty,max=45"`
	Query     string                 `json:"query,omitempty" binding:"omitempty,max=1000"`
	Code      *int                   `json:"code,omitempty" binding:"omitempty,min=1"`
	UserAgent string                 `json:"userAgent,omitempty" binding:"omitempty,max=512"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Stack     string                 `json:"stack,omitempty" binding:"omitempty,max=8192"`
}

type PageLoggingRequest struct {
	PaginationRequest
	Level     string `form:"level" json:"level" binding:"omitempty,oneof=debug info error warn fatal"`
	Type      string `form:"type" json:"type" binding:"omitempty,oneof=Error Panic manul"`
	Message   string `form:"message" json:"message"`     // 按消息内容模糊搜索
	Method    string `form:"method" json:"method"`       // 按HTTP方法过滤
	Path      string `form:"path" json:"path"`           // 按路径模糊搜索
	IP        string `form:"ip" json:"ip"`               // 按IP过滤
	Code      *int   `form:"code" json:"code"`           // 按状态码过滤
	BeginTime string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime   string `form:"endTime" json:"endTime"`     // 结束时间
}

type PageLoggingResponse struct {
	Data       []*LoggingResponse `json:"data"`
	Pagination Pagination         `json:"pagination"`
}
