package sms

import (
	"context"
	"time"
)

// SMSProvider 短信提供商接口
type SMSProvider interface {
	// Name 返回提供商名称
	Name() string

	// SendMessage 发送短信消息
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)

	// SendVerificationCode 发送验证码短信
	SendVerificationCode(ctx context.Context, req *VerificationCodeRequest) (*SendMessageResponse, error)

	// ValidateConfig 验证配置是否有效
	ValidateConfig() error

	// Close 关闭客户端连接
	Close() error
}

// SendMessageRequest 发送短信请求
type SendMessageRequest struct {
	PhoneNumber   string            `json:"phone_number"`   // 手机号码
	TemplateCode  string            `json:"template_code"`  // 模板代码/ID
	SignName      string            `json:"sign_name"`      // 短信签名
	TemplateParam map[string]string `json:"template_param"` // 模板参数
	Content       string            `json:"content"`        // 短信内容 (某些提供商支持)
	Timeout       time.Duration     `json:"timeout"`        // 超时时间
}

// VerificationCodeRequest 验证码请求
type VerificationCodeRequest struct {
	PhoneNumber string        `json:"phone_number"` // 手机号码
	Code        string        `json:"code"`         // 验证码
	Purpose     string        `json:"purpose"`      // 用途 (register, login, reset_password)
	Timeout     time.Duration `json:"timeout"`      // 超时时间
}

// SendMessageResponse 发送短信响应
type SendMessageResponse struct {
	Success   bool   `json:"success"`    // 是否成功
	MessageID string `json:"message_id"` // 消息ID
	BizID     string `json:"biz_id"`     // 业务ID
	Code      string `json:"code"`       // 响应代码
	Message   string `json:"message"`    // 响应消息
}

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderAliyun  ProviderType = "aliyun"  // 阿里云
	ProviderTencent ProviderType = "tencent" // 腾讯云
	ProviderHTTP    ProviderType = "http"    // HTTP接口
	ProviderMock    ProviderType = "mock"    // Mock提供商（用于测试）
)

// IsValid 检查提供商类型是否有效
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderAliyun, ProviderTencent, ProviderHTTP, ProviderMock:
		return true
	default:
		return false
	}
}

// String 返回提供商类型字符串
func (p ProviderType) String() string {
	return string(p)
}
