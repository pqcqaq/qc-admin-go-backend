package models

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Age        *int   `json:"age,omitempty"`
	Phone      string `json:"phone,omitempty"`
	CreateTime string `json:"createTime,omitempty"` // 创建时间
	UpdateTime string `json:"updateTime,omitempty"` // 更新时间
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Age   *int   `json:"age,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Age        *int   `json:"age,omitempty"`
	Phone      string `json:"phone,omitempty"`
	CreateTime string `json:"createTime,omitempty"` // 创建时间
	UpdateTime string `json:"updateTime,omitempty"` // 更新时间
}

// GetUsersRequest 获取用户列表请求结构
type GetUsersRequest struct {
	PaginationRequest
	Name  string `form:"name" json:"name"`   // 按姓名模糊搜索
	Email string `form:"email" json:"email"` // 按邮箱模糊搜索
}

// UsersListResponse 用户列表响应结构
type UsersListResponse struct {
	Data       []*UserResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
