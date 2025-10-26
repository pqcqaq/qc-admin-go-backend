package models

// CreateAreaRequest 创建地区请求结构
type CreateAreaRequest struct {
	Name      string  `json:"name" binding:"required"`
	Level     string  `json:"level" binding:"required,oneof=country province city district street"`
	Depth     int     `json:"depth" binding:"required,min=0,max=4"`
	Spell     string  `json:"spell" binding:"required"`
	Code      string  `json:"code" binding:"required"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ParentId  string  `json:"parentId,omitempty"`
	Color     string  `json:"color,omitempty"`
}

// UpdateAreaRequest 更新地区请求结构
type UpdateAreaRequest struct {
	Name      string   `json:"name,omitempty"`
	Spell     string   `json:"spell,omitempty"`
	Level     string   `json:"level,omitempty"`
	Depth     *int     `json:"depth,omitempty"`
	Code      string   `json:"code,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	ParentId  string   `json:"parentId,omitempty"`
	Color     string   `json:"color,omitempty"`
}

// AreaResponse 地区响应结构
type AreaResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Spell      string          `json:"spell"`
	Level      string          `json:"level"`
	Depth      int             `json:"depth"`
	Code       string          `json:"code"`
	Latitude   float64         `json:"latitude"`
	Longitude  float64         `json:"longitude"`
	ParentId   string          `json:"parentId,omitempty"`
	Color      string          `json:"color,omitempty"`
	Parent     *AreaResponse   `json:"parent,omitempty"`
	Children   []*AreaResponse `json:"children,omitempty"`
	CreateTime string          `json:"createTime"`
	UpdateTime string          `json:"updateTime"`
}

// GetAreasRequest 获取地区列表请求结构
type GetAreasRequest struct {
	PaginationRequest
	Name     string `form:"name" json:"name"`         // 按名称模糊搜索
	Level    string `form:"level" json:"level"`       // 按层级类型搜索
	Depth    *int   `form:"depth" json:"depth"`       // 按深度搜索
	Code     string `form:"code" json:"code"`         // 按编码搜索
	ParentId string `form:"parentId" json:"parentId"` // 按父级ID搜索
}

// AreasListResponse 地区列表响应结构
type AreasListResponse struct {
	Data       []*AreaResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// AreaTreeResponse 地区树形结构响应
type AreaTreeResponse struct {
	Data []*AreaResponse `json:"data"` // 树形结构的地区
}
