package models

// RoleInfo 角色信息（用于客户端设备响应中）
type RoleInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ClientDeviceResponse 客户端设备响应
type ClientDeviceResponse struct {
	ID                 string     `json:"id"`
	CreateTime         string     `json:"createTime"`
	UpdateTime         string     `json:"updateTime"`
	Name               string     `json:"name"`
	Code               string     `json:"code"`
	Description        string     `json:"description"`
	Enabled            bool       `json:"enabled"`
	AccessTokenExpiry  uint64     `json:"accessTokenExpiry"`
	RefreshTokenExpiry uint64     `json:"refreshTokenExpiry"`
	Anonymous          bool       `json:"anonymous"`
	Roles              []RoleInfo `json:"roles,omitempty"` // 关联的角色列表
}

// CreateClientDeviceRequest 创建客户端设备请求
type CreateClientDeviceRequest struct {
	Name               string   `json:"name" binding:"required"` // 设备名称
	Enabled            *bool    `json:"enabled"`
	Description        string   `json:"description"`                           // 是否启用，可选，默认true
	AccessTokenExpiry  uint64   `json:"accessTokenExpiry" binding:"required"`  // accessToken超时时间(ms)
	RefreshTokenExpiry uint64   `json:"refreshTokenExpiry" binding:"required"` // refreshToken超时时间(ms)
	Anonymous          *bool    `json:"anonymous"`                             // 允许所有角色登录，可选，默认true
	RoleIds            []string `json:"roleIds,omitempty"`                     // 关联的角色ID列表
}

// UpdateClientDeviceRequest 更新客户端设备请求
type UpdateClientDeviceRequest struct {
	Name               string   `json:"name" binding:"required"` // 设备名称
	Enabled            *bool    `json:"enabled" binding:"required"`
	Description        string   `json:"description"`                           // 是否启用
	AccessTokenExpiry  *uint64  `json:"accessTokenExpiry" binding:"required"`  // accessToken超时时间(ms)
	RefreshTokenExpiry *uint64  `json:"refreshTokenExpiry" binding:"required"` // refreshToken超时时间(ms)
	Anonymous          *bool    `json:"anonymous" binding:"required"`          // 允许所有角色登录
	RoleIds            []string `json:"roleIds,omitempty"`                     // 关联的角色ID列表
}

// PageClientDevicesRequest 分页查询客户端设备请求
type PageClientDevicesRequest struct {
	PaginationRequest
	Name      string `form:"name" json:"name"`           // 按名称模糊搜索
	Code      string `form:"code" json:"code"`           // 按code精确搜索
	Enabled   *bool  `form:"enabled" json:"enabled"`     // 按启用状态过滤
	Anonymous *bool  `form:"anonymous" json:"anonymous"` // 按匿名登录过滤
	BeginTime string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime   string `form:"endTime" json:"endTime"`     // 结束时间
}

// PageClientDevicesResponse 分页查询客户端设备响应
type PageClientDevicesResponse struct {
	Data       []*ClientDeviceResponse `json:"data"` // 客户端设备列表
	Pagination Pagination              `json:"pagination"`
}

// CheckClientAccessRequest 检查客户端访问权限请求
type CheckClientAccessRequest struct {
	Code   string   `json:"code" binding:"required"`   // 客户端code
	UserId string   `json:"userId" binding:"required"` // 用户ID
	Roles  []string `json:"roles"`                     // 用户拥有的角色ID列表
}

// CheckClientAccessResponse 检查客户端访问权限响应
type CheckClientAccessResponse struct {
	Allowed bool   `json:"allowed"` // 是否允许访问
	Reason  string `json:"reason"`  // 不允许访问的原因
}

// ClientDeviceByCodeResponse 根据code获取客户端设备响应
type ClientDeviceByCodeResponse struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Code               string     `json:"code"`
	Description        string     `json:"description"`
	Enabled            bool       `json:"enabled"`
	AccessTokenExpiry  uint64     `json:"accessTokenExpiry"`
	RefreshTokenExpiry uint64     `json:"refreshTokenExpiry"`
	Anonymous          bool       `json:"anonymous"`
	Roles              []RoleInfo `json:"roles,omitempty"` // 关联的角色列表
}
