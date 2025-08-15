package models

// ApiResponse 通用API响应结构
type ApiResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     string            `json:"status"`
	Message    string            `json:"message"`
	Components map[string]string `json:"components,omitempty"`
}

// PaginationRequest 分页请求结构
type PaginationRequest struct {
	Page     int    `form:"page" json:"page" binding:"min=1"`                   // 页码，从1开始
	PageSize int    `form:"page_size" json:"page_size" binding:"min=1,max=100"` // 每页数量，最大100
	OrderBy  string `form:"order_by" json:"order_by"`                           // 排序字段
	Order    string `form:"order" json:"order" binding:"oneof=asc desc"`        // 排序方向：asc 或 desc
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Data       interface{} `json:"data"`       // 数据列表
	Pagination Pagination  `json:"pagination"` // 分页信息
}

// Pagination 分页信息
type Pagination struct {
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页数量
	Total      int64 `json:"total"`       // 总记录数
	TotalPages int   `json:"total_pages"` // 总页数
	HasNext    bool  `json:"has_next"`    // 是否有下一页
	HasPrev    bool  `json:"has_prev"`    // 是否有上一页
}
