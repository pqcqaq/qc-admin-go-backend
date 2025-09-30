package models

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	SenderType string `json:"senderType" binding:"required,oneof=email phone sms"` // 发送方式
	Purpose    string `json:"purpose" binding:"required"`                          // 用途
	Identifier string `json:"identifier" binding:"required"`                       // 标识符
	ClientCode string `json:"clientCode,omitempty"`
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
	ClientCode     string `json:"clientCode,omitempty" binding:"required"`
	RememberMe     *bool  `json:"rememberMe"` // 记住我
}

// LoginResponse 登录响应
type LoginResponse struct {
	User    UserInfo  `json:"user"`
	Token   TokenInfo `json:"token,omitempty"` // JWT token (如果有的话)
	Message string    `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	CredentialType string `json:"credentialType" binding:"required,oneof=password email phone oauth totp"` // 认证类型
	Identifier     string `json:"identifier" binding:"required"`                                           // 标识符
	Secret         string `json:"secret,omitempty"`                                                        // 密码
	VerifyCode     string `json:"verifyCode,omitempty"`                                                    // 验证码（非密码注册时必需）
	Username       string `json:"username" binding:"required"`                                             // 用户名
	ClientCode     string `json:"clientCode,omitempty" binding:"required"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User    UserInfo  `json:"user"`
	Token   TokenInfo `json:"token,omitempty"` // JWT token (如果有的话)
	Message string    `json:"message"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	CredentialType string `json:"credentialType" binding:"required,oneof=password email phone oauth totp"` // 认证类型
	Identifier     string `json:"identifier" binding:"required"`                                           // 标识符
	NewPassword    string `json:"newPassword" binding:"required"`                                          // 新密码
	VerifyCode     string `json:"verifyCode,omitempty"`                                                    // 验证码（非密码重置时必需）
	OldPassword    string `json:"oldPassword,omitempty"`                                                   // 原密码（密码重置时必需）
	ClientCode     string `json:"clientCode,omitempty" binding:"required"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type TokenInfo struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	AccessExpiredIn  uint64 `json:"accessExpiredIn"`
	RefreshExpiredIn uint64 `json:"refreshExpiredIn"`
}

// RefreshTokenRequest 刷新Token请求结构体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Status      string                `json:"status"`
	CreateTime  string                `json:"createTime"`
	UpdateTime  string                `json:"updateTime"`
	Avatar      string                `json:"avatar,omitempty"`      // 头像
	Roles       []*RoleResponse       `json:"roles,omitempty"`       // 用户角色
	Sex         string                `json:"sex,omitempty"`         // 性别
	Age         int                   `json:"age,omitempty"`         // 年龄
	Permissions []*PermissionResponse `json:"permissions,omitempty"` // 用户权限（通过角色继承）
}

// LoginRecordResponse 登录记录响应
type LoginRecordResponse struct {
	ID             uint64                 `json:"id"`
	UserID         uint64                 `json:"userId"`
	Identifier     string                 `json:"identifier"`
	CredentialType string                 `json:"credentialType"`
	IPAddress      string                 `json:"ipAddress"`
	UserAgent      string                 `json:"userAgent,omitempty"`
	DeviceInfo     string                 `json:"deviceInfo,omitempty"`
	Location       string                 `json:"location,omitempty"`
	Status         string                 `json:"status"`
	FailureReason  string                 `json:"failureReason,omitempty"`
	SessionID      string                 `json:"sessionId,omitempty"`
	LoginTime      string                 `json:"loginTime"`
	LogoutTime     string                 `json:"logoutTime,omitempty"`
	Duration       int                    `json:"duration,omitempty"` // 会话持续时间(秒)
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
