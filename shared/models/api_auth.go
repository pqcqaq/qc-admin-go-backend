package models

type PermissionsList struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// APIAuthResponse API认证响应结构
type APIAuthResponse struct {
	ID          string                 `json:"id"`
	CreateTime  string                 `json:"createTime"`
	UpdateTime  string                 `json:"updateTime"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"` // http or websocket
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	IsPublic    bool                   `json:"isPublic"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Permissions []*PermissionResponse  `json:"permissions,omitempty"`
}

// CreateAPIAuthRequest 创建API认证请求结构
type CreateAPIAuthRequest struct {
	Name        string                 `json:"name" binding:"required"`     // API名称
	Description string                 `json:"description,omitempty"`       // API描述
	Method      string                 `json:"method" binding:"required"`   // HTTP方法
	Path        string                 `json:"path" binding:"required"`     // API路径
	Type        string                 `json:"type" binding:"required"`     // API类型，http或websocket
	IsPublic    *bool                  `json:"isPublic" binding:"required"` // 是否为公开API
	IsActive    *bool                  `json:"isActive" binding:"required"` // 是否启用
	Metadata    map[string]interface{} `json:"metadata,omitempty"`          // 额外的元数据信息
	Permissions []*PermissionsList     `json:"permissions,omitempty"`       // 关联的权限ID列表
}

// UpdateAPIAuthRequest 更新API认证请求结构
type UpdateAPIAuthRequest struct {
	Name        string                 `json:"name" binding:"required"`     // API名称
	Description string                 `json:"description,omitempty"`       // API描述
	Method      string                 `json:"method" binding:"required"`   // HTTP方法
	Path        string                 `json:"path" binding:"required"`     // API路径
	Type        string                 `json:"type" binding:"required"`     // API类型，http或websocket
	IsPublic    *bool                  `json:"isPublic" binding:"required"` // 是否为公开API
	IsActive    *bool                  `json:"isActive" binding:"required"` // 是否启用
	Metadata    map[string]interface{} `json:"metadata,omitempty"`          // 额外的元数据信息
	Permissions []*PermissionsList     `json:"permissions,omitempty"`       // 关联的权限ID列表
}

// PageAPIAuthRequest 分页查询API认证请求结构
type PageAPIAuthRequest struct {
	PaginationRequest
	Name      string `form:"name" json:"name"`           // 按名称模糊搜索
	Method    string `form:"method" json:"method"`       // 按HTTP方法过滤
	Path      string `form:"path" json:"path"`           // 按路径模糊搜索
	Type      string `form:"type" json:"type"`           // 按API类型过滤
	IsPublic  *bool  `form:"isPublic" json:"isPublic"`   // 按是否公开过滤
	IsActive  *bool  `form:"isActive" json:"isActive"`   // 按是否启用过滤
	BeginTime string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime   string `form:"endTime" json:"endTime"`     // 结束时间
}

// PageAPIAuthResponse 分页查询API认证响应结构
type PageAPIAuthResponse struct {
	Data       []*APIAuthResponse `json:"data"` // API认证列表
	Pagination Pagination         `json:"pagination"`
}
