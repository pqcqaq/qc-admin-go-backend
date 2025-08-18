package models

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	SenderType string `json:"senderType" binding:"required,oneof=email phone sms"` // 发送方式
	Purpose    string `json:"purpose" binding:"required"`                          // 用途
	Identifier string `json:"identifier" binding:"required"`                       // 标识符
}

// SendVerifyCodeResponse 发送验证码响应
type SendVerifyCodeResponse struct {
	Message string `json:"message"`
}

// VerifyCodeRequest 验证验证码请求
type VerifyCodeRequest struct {
	SenderType string `json:"senderType" binding:"required,oneof=email phone sms"` // 发送方式
	Purpose    string `json:"purpose" binding:"required"`                          // 用途
	Identifier string `json:"identifier" binding:"required"`                       // 标识符
	Code       string `json:"code" binding:"required"`                             // 验证码
}

// VerifyCodeResponse 验证验证码响应
type VerifyCodeResponse struct {
	Message string `json:"message"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	CredentialType string `json:"credentialType" binding:"required,oneof=password email phone oauth totp"` // 认证类型
	Identifier     string `json:"identifier" binding:"required"`                                           // 标识符
	Secret         string `json:"secret,omitempty"`                                                        // 密码（密码登录时必需）
	VerifyCode     string `json:"verifyCode,omitempty"`                                                    // 验证码（非密码登录时必需）
}

// LoginResponse 登录响应
type LoginResponse struct {
	User    UserInfo `json:"user"`
	Token   string   `json:"token,omitempty"` // JWT token (如果有的话)
	Message string   `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	CredentialType string `json:"credentialType" binding:"required,oneof=password email phone oauth totp"` // 认证类型
	Identifier     string `json:"identifier" binding:"required"`                                           // 标识符
	Secret         string `json:"secret,omitempty"`                                                        // 密码
	VerifyCode     string `json:"verifyCode,omitempty"`                                                    // 验证码（非密码注册时必需）
	Username       string `json:"username" binding:"required"`                                             // 用户名
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User    UserInfo `json:"user"`
	Token   string   `json:"token,omitempty"` // JWT token (如果有的话)
	Message string   `json:"message"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	CredentialType string `json:"credentialType" binding:"required,oneof=password email phone oauth totp"` // 认证类型
	Identifier     string `json:"identifier" binding:"required"`                                           // 标识符
	NewPassword    string `json:"newPassword" binding:"required"`                                          // 新密码
	VerifyCode     string `json:"verifyCode,omitempty"`                                                    // 验证码（非密码重置时必需）
	OldPassword    string `json:"oldPassword,omitempty"`                                                   // 原密码（密码重置时必需）
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID          uint64                `json:"id"`
	Name        string                `json:"name"`
	Status      string                `json:"status"`
	CreateTime  string                `json:"createTime"`
	UpdateTime  string                `json:"updateTime"`
	Roles       []*RoleResponse       `json:"roles,omitempty"`       // 用户角色
	Permissions []*PermissionResponse `json:"permissions,omitempty"` // 用户权限（通过角色继承）
}
