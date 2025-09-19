package models

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Name       string `json:"name" binding:"required"`
	Age        *int   `json:"age,omitempty"`
	Sex        string `json:"sex,omitempty"`
	Status     string `json:"status,omitempty"`     // 用户状态
	CreateTime string `json:"createTime,omitempty"` // 创建时间
	UpdateTime string `json:"updateTime,omitempty"` // 更新时间
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Age      *int   `json:"age,omitempty"`
	Sex      string `json:"sex,omitempty"`
	Status   string `json:"status,omitempty"`   // 用户状态
	AvatarId string `json:"avatarId,omitempty"` // 头像ID
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Age        *int     `json:"age,omitempty"`
	Sex        string   `json:"sex,omitempty"`
	Status     string   `json:"status,omitempty"`     // 用户状态
	CreateTime string   `json:"createTime,omitempty"` // 创建时间
	UpdateTime string   `json:"updateTime,omitempty"` // 更新时间
	Avatar     string   `json:"avatar,omitempty"`     // 头像
	Roles      []string `json:"roles,omitempty"`      // 角色列表
}

type UserFromRoleResponse struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Age        *int           `json:"age,omitempty"`
	Sex        string         `json:"sex,omitempty"`
	Status     string         `json:"status,omitempty"`     // 用户状态
	CreateTime string         `json:"createTime,omitempty"` // 创建时间
	UpdateTime string         `json:"updateTime,omitempty"` // 更新时间
	Avatar     string         `json:"avatar,omitempty"`     // 头像
	Roles      []string       `json:"roles,omitempty"`      // 角色列表
	OtherRoles []RoleResponse `json:"otherRoles,omitempty"` // 角色列表
}

// GetUsersRequest 获取用户列表请求结构
type GetUsersRequest struct {
	PaginationRequest
	Name          string `form:"name" json:"name"`                   // 按姓名模糊搜索
	Sex           string `json:"sex,omitempty"`                      // 性别筛选
	Status        string `json:"status,omitempty"`                   // 用户状态
	Keyword       string `form:"keyword" json:"keyword"`             // 搜索关键字（用户名、昵称）
	ExcludeRoleId string `form:"excludeRoleId" json:"excludeRoleId"` // 排除拥有指定角色的用户
}

// UsersListResponse 用户列表响应结构
type UsersListResponse struct {
	Data       []*UserResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
