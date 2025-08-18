package sms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-backend/pkg/configs"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// SMSClient 短信客户端结构
type SMSClient struct {
	config   *configs.SMSConfig
	provider SMSProvider
	factory  ProviderFactory
}

// 单例相关变量
var (
	Client *SMSClient
	once   sync.Once
	mu     sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的短信客户端
func NewClient(smsConfig *configs.SMSConfig) (*SMSClient, error) {
	if smsConfig == nil {
		return nil, fmt.Errorf("短信配置不能为空")
	}

	// 创建提供商工厂
	factory := NewProviderFactory()

	// 创建短信提供商
	provider, err := CreateProviderFromConfig(smsConfig)
	if err != nil {
		return nil, fmt.Errorf("创建短信提供商失败: %w", err)
	}

	client := &SMSClient{
		config:   smsConfig,
		provider: provider,
		factory:  factory,
	}

	if logger != nil {
		logger.Info("短信服务初始化成功: provider=%s", smsConfig.Provider)
	}

	return client, nil
}

// GetClient 获取单例短信客户端
func GetClient() *SMSClient {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// InitializeClient 初始化单例短信客户端
func InitializeClient(smsConfig *configs.SMSConfig) error {
	var err error
	once.Do(func() {
		Client, err = NewClient(smsConfig)
	})
	return err
}

// GetProvider 获取当前使用的提供商
func (c *SMSClient) GetProvider() SMSProvider {
	return c.provider
}

// GetProviderName 获取当前提供商名称
func (c *SMSClient) GetProviderName() string {
	if c.provider != nil {
		return c.provider.Name()
	}
	return "unknown"
}

// SwitchProvider 切换提供商 (动态切换)
func (c *SMSClient) SwitchProvider(providerType ProviderType) error {
	if !providerType.IsValid() {
		return fmt.Errorf("不支持的提供商类型: %s", providerType)
	}

	// 创建新的提供商
	newProvider, err := c.factory.CreateProvider(providerType, c.config)
	if err != nil {
		return fmt.Errorf("切换提供商失败: %w", err)
	}

	// 关闭旧提供商
	if c.provider != nil {
		if err := c.provider.Close(); err != nil {
			if logger != nil {
				logger.Error("关闭旧提供商失败: %v", err)
			}
		}
	}

	// 切换到新提供商
	c.provider = newProvider
	c.config.Provider = providerType.String()

	if logger != nil {
		logger.Info("已切换到新的短信提供商: %s", providerType)
	}

	return nil
}

// SendMessage 发送短信消息
func (c *SMSClient) SendMessage(phoneNumber, templateCode, signName string, templateParam map[string]string) error {
	return c.SendMessageWithContext(context.Background(), phoneNumber, templateCode, signName, templateParam)
}

// SendMessageWithContext 发送短信消息 (带上下文)
func (c *SMSClient) SendMessageWithContext(ctx context.Context, phoneNumber, templateCode, signName string, templateParam map[string]string) error {
	if c.provider == nil {
		return fmt.Errorf("短信提供商未初始化")
	}

	req := &SendMessageRequest{
		PhoneNumber:   phoneNumber,
		TemplateCode:  templateCode,
		SignName:      signName,
		TemplateParam: templateParam,
		Timeout:       30 * time.Second,
	}

	response, err := c.provider.SendMessage(ctx, req)
	if err != nil {
		return fmt.Errorf("发送短信失败: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("短信发送失败: %s", response.Message)
	}

	return nil
}

// SendVerificationCode 发送验证码短信
func (c *SMSClient) SendVerificationCode(phoneNumber, code, purpose string) error {
	return c.SendVerificationCodeWithContext(context.Background(), phoneNumber, code, purpose)
}

// SendVerificationCodeWithContext 发送验证码短信 (带上下文)
func (c *SMSClient) SendVerificationCodeWithContext(ctx context.Context, phoneNumber, code, purpose string) error {
	if c.provider == nil {
		return fmt.Errorf("短信提供商未初始化")
	}

	req := &VerificationCodeRequest{
		PhoneNumber: phoneNumber,
		Code:        code,
		Purpose:     purpose,
		Timeout:     30 * time.Second,
	}

	response, err := c.provider.SendVerificationCode(ctx, req)
	if err != nil {
		return fmt.Errorf("发送验证码短信失败: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("验证码短信发送失败: %s", response.Message)
	}

	return nil
}

// SendSimpleVerificationCode 发送简单验证码短信 (只包含验证码)
func (c *SMSClient) SendSimpleVerificationCode(phoneNumber, code string) error {
	return c.SendVerificationCode(phoneNumber, code, "")
}

// Close 关闭短信客户端
func (c *SMSClient) Close() error {
	if c.provider != nil {
		return c.provider.Close()
	}
	return nil
}

// 全局便捷函数

// SendVerificationCode 发送验证码短信 (全局函数)
func SendVerificationCode(phoneNumber, code, purpose string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("短信客户端未初始化")
	}
	return client.SendVerificationCode(phoneNumber, code, purpose)
}

// SendSimpleVerificationCode 发送简单验证码短信 (全局函数)
func SendSimpleVerificationCode(phoneNumber, code string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("短信客户端未初始化")
	}
	return client.SendSimpleVerificationCode(phoneNumber, code)
}

// SendMessage 发送短信消息 (全局函数)
func SendMessage(phoneNumber, templateCode, signName string, templateParam map[string]string) error {
	client := GetClient()
	if client == nil {
		return fmt.Errorf("短信客户端未初始化")
	}
	return client.SendMessage(phoneNumber, templateCode, signName, templateParam)
}

// GetSupportedProviders 获取支持的提供商列表
func GetSupportedProviders() []ProviderType {
	factory := NewProviderFactory()
	return factory.GetSupportedProviders()
}
