package models

type ScanResponse struct {
	ID         string `json:"id"`
	CreateTime string `json:"createTime"`
	Content    string `json:"content"`
	Success    bool   `json:"success"`
	ImageId    string `json:"imageId,omitempty"`  // 关联的图片ID
	ImageUrl   string `json:"imageUrl,omitempty"` // 图片URL
}

type CreateScanRequest struct {
	Content string `json:"content" binding:"required"` // 扫描内容
	Success *bool  `json:"success" binding:"required"` // 扫描是否成功
	ImageId string `json:"imageId,omitempty"`          // 关联的图片ID
}

type UpdateScanRequest struct {
	Content string `json:"content" binding:"required"` // 扫描内容
	Success *bool  `json:"success" binding:"required"` // 扫描是否成功
	ImageId string `json:"imageId,omitempty"`          // 关联的图片ID
}

type PageScansRequest struct {
	PaginationRequest
	Content   string `form:"content" json:"content"`     // 按内容模糊搜索
	Success   *bool  `form:"success" json:"success"`     // 按扫描结果过滤
	BeginTime string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime   string `form:"endTime" json:"endTime"`     // 结束时间
}

type PageScansResponse struct {
	Data       []*ScanResponse `json:"data"` // 扫描记录列表
	Pagination Pagination      `json:"pagination"`
}
