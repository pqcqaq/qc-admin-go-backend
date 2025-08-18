package sms

import (
	"context"
	"fmt"
	"time"

	"go-backend/pkg/configs"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// TencentProvider 腾讯云短信提供商
type TencentProvider struct {
	config *configs.TencentSMSConfig
	client *sms.Client
}

// NewTencentProvider 创建腾讯云短信提供商
func NewTencentProvider(config *configs.TencentSMSConfig) (*TencentProvider, error) {
	provider := &TencentProvider{
		config: config,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	// 创建认证信息
	credential := common.NewCredential(
		config.SecretID,
		config.SecretKey,
	)

	// 创建客户端配置
	cpf := profile.NewClientProfile()
	if config.Endpoint != "" {
		cpf.HttpProfile.Endpoint = config.Endpoint
	} else {
		cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	}

	// 设置地域
	region := config.Region
	if region == "" {
		region = "ap-guangzhou"
	}

	// 创建短信客户端
	client, err := sms.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("创建腾讯云短信客户端失败: %w", err)
	}

	provider.client = client

	if logger != nil {
		logger.Info("腾讯云短信提供商初始化成功: region=%s", region)
	}

	return provider, nil
}

// Name 返回提供商名称
func (p *TencentProvider) Name() string {
	return "tencent"
}

// ValidateConfig 验证配置是否有效
func (p *TencentProvider) ValidateConfig() error {
	if p.config.SecretID == "" {
		return fmt.Errorf("腾讯云短信配置错误: secret_id 不能为空")
	}
	if p.config.SecretKey == "" {
		return fmt.Errorf("腾讯云短信配置错误: secret_key 不能为空")
	}
	if p.config.AppID == "" {
		return fmt.Errorf("腾讯云短信配置错误: app_id 不能为空")
	}
	if p.config.SignName == "" {
		return fmt.Errorf("腾讯云短信配置错误: sign_name 不能为空")
	}
	if p.config.TemplateID == "" {
		return fmt.Errorf("腾讯云短信配置错误: template_id 不能为空")
	}
	return nil
}

// SendMessage 发送短信消息
func (p *TencentProvider) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// 设置默认超时时间
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}

	// 使用提供的参数或配置的默认值
	signName := req.SignName
	if signName == "" {
		signName = p.config.SignName
	}

	templateID := req.TemplateCode
	if templateID == "" {
		templateID = p.config.TemplateID
	}

	// 准备模板参数数组 (腾讯云按数字索引顺序传递参数)
	var templateParamSet []*string
	if len(req.TemplateParam) > 0 {
		// 腾讯云模板参数需要按顺序传递，这里按照常见的参数顺序
		if code, exists := req.TemplateParam["code"]; exists {
			templateParamSet = append(templateParamSet, common.StringPtr(code))
		}
		// 可以根据需要添加更多参数
		for key, value := range req.TemplateParam {
			if key != "code" { // 避免重复添加code
				templateParamSet = append(templateParamSet, common.StringPtr(value))
			}
		}
	}

	// 创建发送请求
	request := sms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{req.PhoneNumber})
	request.SmsSdkAppId = common.StringPtr(p.config.AppID)
	request.SignName = common.StringPtr(signName)
	request.TemplateId = common.StringPtr(templateID)
	if len(templateParamSet) > 0 {
		request.TemplateParamSet = templateParamSet
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// 使用带超时的上下文发送短信
	response, err := p.client.SendSmsWithContext(timeoutCtx, request)
	if err != nil {
		if logger != nil {
			logger.Error("腾讯云短信发送失败: phone=%s, error=%v", req.PhoneNumber, err)
		}
		return nil, fmt.Errorf("发送短信失败: %w", err)
	}

	// 构建响应
	result := &SendMessageResponse{
		Success:   false,
		MessageID: *response.Response.RequestId,
		Code:      "",
		Message:   "",
	}

	// 检查发送结果
	if len(response.Response.SendStatusSet) > 0 {
		status := response.Response.SendStatusSet[0]
		result.Code = *status.Code
		result.Message = *status.Message
		result.BizID = *status.SerialNo

		// 腾讯云成功状态码为 "Ok"
		if *status.Code == "Ok" {
			result.Success = true

			if logger != nil {
				logger.Info("腾讯云短信发送成功: phone=%s, serialNo=%s", req.PhoneNumber, result.BizID)
			}
		} else {
			if logger != nil {
				logger.Error("腾讯云短信发送失败: phone=%s, code=%s, message=%s",
					req.PhoneNumber, result.Code, result.Message)
			}
		}
	}

	return result, nil
}

// SendVerificationCode 发送验证码短信
func (p *TencentProvider) SendVerificationCode(ctx context.Context, req *VerificationCodeRequest) (*SendMessageResponse, error) {
	// 腾讯云通常只需要验证码参数
	sendReq := &SendMessageRequest{
		PhoneNumber: req.PhoneNumber,
		TemplateParam: map[string]string{
			"code": req.Code,
		},
		Timeout: req.Timeout,
	}

	return p.SendMessage(ctx, sendReq)
}

// Close 关闭客户端连接
func (p *TencentProvider) Close() error {
	// 腾讯云SDK不需要显式关闭连接
	if logger != nil {
		logger.Info("腾讯云短信客户端已关闭")
	}
	return nil
}
