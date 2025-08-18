package sms

import (
	"fmt"
	"strings"

	"go-backend/pkg/configs"
)

// ProviderFactory 短信提供商工厂接口
type ProviderFactory interface {
	CreateProvider(providerType ProviderType, config *configs.SMSConfig) (SMSProvider, error)
	GetSupportedProviders() []ProviderType
}

// DefaultProviderFactory 默认提供商工厂
type DefaultProviderFactory struct{}

// NewProviderFactory 创建提供商工厂
func NewProviderFactory() ProviderFactory {
	return &DefaultProviderFactory{}
}

// CreateProvider 创建指定类型的短信提供商
func (f *DefaultProviderFactory) CreateProvider(providerType ProviderType, config *configs.SMSConfig) (SMSProvider, error) {
	if !providerType.IsValid() {
		return nil, fmt.Errorf("不支持的短信提供商类型: %s", providerType)
	}

	switch providerType {
	case ProviderAliyun:
		return NewAliyunProvider(&config.Aliyun)
	case ProviderTencent:
		return NewTencentProvider(&config.Tencent)
	case ProviderHTTP:
		return NewHTTPProvider(&config.HTTP)
	case ProviderMock:
		return NewMockProvider(&config.Mock)
	default:
		return nil, fmt.Errorf("不支持的短信提供商类型: %s", providerType)
	}
}

// GetSupportedProviders 获取支持的提供商列表
func (f *DefaultProviderFactory) GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderAliyun,
		ProviderTencent,
		ProviderHTTP,
		ProviderMock,
	}
}

// CreateProviderFromConfig 从配置创建提供商
func CreateProviderFromConfig(config *configs.SMSConfig) (SMSProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("短信配置不能为空")
	}

	providerType := ProviderType(strings.ToLower(config.Provider))
	if !providerType.IsValid() {
		return nil, fmt.Errorf("不支持的短信提供商: %s", config.Provider)
	}

	factory := NewProviderFactory()
	return factory.CreateProvider(providerType, config)
}
