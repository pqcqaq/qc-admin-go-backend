package sms

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-backend/pkg/configs"
)

// MockProvider Mock短信提供商，用于测试，将验证码打印到控制台
type MockProvider struct {
	config *configs.MockSMSConfig
}

// NewMockProvider 创建Mock短信提供商
func NewMockProvider(config *configs.MockSMSConfig) (*MockProvider, error) {
	provider := &MockProvider{
		config: config,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// Name 返回提供商名称
func (p *MockProvider) Name() string {
	return "Mock SMS Provider"
}

// SendMessage 发送短信消息
func (p *MockProvider) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("发送短信请求不能为空")
	}

	// 打印到控制台
	log.Printf("=== Mock SMS ===")
	log.Printf("发送短信到: %s", req.PhoneNumber)
	log.Printf("模板代码: %s", req.TemplateCode)
	log.Printf("签名: %s", req.SignName)
	log.Printf("模板参数: %+v", req.TemplateParam)
	if req.Content != "" {
		log.Printf("短信内容: %s", req.Content)
	}
	log.Printf("===============")

	// 模拟成功响应
	return &SendMessageResponse{
		Success:   true,
		MessageID: fmt.Sprintf("mock_msg_%d", time.Now().Unix()),
		BizID:     fmt.Sprintf("mock_biz_%d", time.Now().Unix()),
		Code:      "OK",
		Message:   "Mock短信发送成功",
	}, nil
}

// SendVerificationCode 发送验证码短信
func (p *MockProvider) SendVerificationCode(ctx context.Context, req *VerificationCodeRequest) (*SendMessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("验证码请求不能为空")
	}

	// 打印验证码到控制台
	log.Printf("==================== Mock验证码 ====================")
	log.Printf("   手机号码: %s", req.PhoneNumber)
	log.Printf("   验证码: %s", req.Code)
	log.Printf("   用途: %s", req.Purpose)
	log.Printf("   超时时间: %v", req.Timeout)
	log.Printf("==================================================")

	// 如果启用了详细输出，显示更多信息
	if p.config != nil && p.config.Verbose {
		log.Printf("Mock配置: 详细模式已启用")
		log.Printf("发送时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// 模拟成功响应
	return &SendMessageResponse{
		Success:   true,
		MessageID: fmt.Sprintf("mock_verification_%d", time.Now().Unix()),
		BizID:     fmt.Sprintf("mock_verification_biz_%d", time.Now().Unix()),
		Code:      "OK",
		Message:   "Mock验证码发送成功",
	}, nil
}

// ValidateConfig 验证配置是否有效
func (p *MockProvider) ValidateConfig() error {
	// Mock provider 不需要特殊配置验证
	return nil
}

// Close 关闭客户端连接
func (p *MockProvider) Close() error {
	log.Printf("Mock SMS Provider 已关闭")
	return nil
}
