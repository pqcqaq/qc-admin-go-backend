package models

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   *int   `json:"age,omitempty"`
	Phone string `json:"phone,omitempty"`
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
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   *int   `json:"age,omitempty"`
	Phone string `json:"phone,omitempty"`
}
