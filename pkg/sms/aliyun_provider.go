package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-backend/pkg/configs"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	tea "github.com/alibabacloud-go/tea/tea"
)

// AliyunProvider 阿里云短信提供商
type AliyunProvider struct {
	config *configs.AliyunSMSConfig
	client *dysmsapi.Client
}

// NewAliyunProvider 创建阿里云短信提供商
func NewAliyunProvider(config *configs.AliyunSMSConfig) (*AliyunProvider, error) {
	provider := &AliyunProvider{
		config: config,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	// 创建阿里云短信客户端配置
	openAPIConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
	}

	// 设置端点
	if config.Endpoint != "" {
		openAPIConfig.Endpoint = tea.String(config.Endpoint)
	} else {
		openAPIConfig.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	}

	// 创建短信客户端
	client, err := dysmsapi.NewClient(openAPIConfig)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云短信客户端失败: %w", err)
	}

	provider.client = client

	if logger != nil {
		logger.Info("阿里云短信提供商初始化成功: endpoint=%s", config.Endpoint)
	}

	return provider, nil
}

// Name 返回提供商名称
func (p *AliyunProvider) Name() string {
	return "aliyun"
}

// ValidateConfig 验证配置是否有效
func (p *AliyunProvider) ValidateConfig() error {
	if p.config.AccessKeyID == "" {
		return fmt.Errorf("阿里云短信配置错误: access_key_id 不能为空")
	}
	if p.config.AccessKeySecret == "" {
		return fmt.Errorf("阿里云短信配置错误: access_key_secret 不能为空")
	}
	if p.config.SignName == "" {
		return fmt.Errorf("阿里云短信配置错误: sign_name 不能为空")
	}
	if p.config.TemplateCode == "" {
		return fmt.Errorf("阿里云短信配置错误: template_code 不能为空")
	}
	return nil
}

// SendMessage 发送短信消息
func (p *AliyunProvider) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// 设置默认超时时间
	if req.Timeout == 0 {
		req.Timeout = 30 * time.Second
	}

	// 使用提供的参数或配置的默认值
	signName := req.SignName
	if signName == "" {
		signName = p.config.SignName
	}

	templateCode := req.TemplateCode
	if templateCode == "" {
		templateCode = p.config.TemplateCode
	}

	// 准备模板参数
	var templateParamStr string
	if len(req.TemplateParam) > 0 {
		paramBytes, err := json.Marshal(req.TemplateParam)
		if err != nil {
			return nil, fmt.Errorf("序列化模板参数失败: %w", err)
		}
		templateParamStr = string(paramBytes)
	}

	// 创建发送请求
	sendSmsRequest := &dysmsapi.SendSmsRequest{
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		PhoneNumbers:  tea.String(req.PhoneNumber),
		TemplateParam: tea.String(templateParamStr),
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// 发送短信
	response, err := p.client.SendSms(sendSmsRequest)
	if err != nil {
		if logger != nil {
			logger.Error("阿里云短信发送失败: phone=%s, error=%v", req.PhoneNumber, err)
		}
		return nil, fmt.Errorf("发送短信失败: %w", err)
	}

	// 检查上下文是否超时
	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("发送短信超时")
	default:
	}

	// 构建响应
	result := &SendMessageResponse{
		Success:   false,
		MessageID: "",
		BizID:     tea.StringValue(response.Body.BizId),
		Code:      tea.StringValue(response.Body.Code),
		Message:   tea.StringValue(response.Body.Message),
	}

	// 检查响应结果
	if response.Body.Code != nil && *response.Body.Code == "OK" {
		result.Success = true
		result.MessageID = tea.StringValue(response.Body.RequestId)

		if logger != nil {
			logger.Info("阿里云短信发送成功: phone=%s, bizId=%s", req.PhoneNumber, result.BizID)
		}
	} else {
		if logger != nil {
			logger.Error("阿里云短信发送失败: phone=%s, code=%s, message=%s",
				req.PhoneNumber, result.Code, result.Message)
		}
	}

	return result, nil
}

// SendVerificationCode 发送验证码短信
func (p *AliyunProvider) SendVerificationCode(ctx context.Context, req *VerificationCodeRequest) (*SendMessageResponse, error) {
	// 根据用途生成不同的参数
	var purposeText string
	switch req.Purpose {
	case "register":
		purposeText = "注册"
	case "login":
		purposeText = "登录"
	case "reset_password":
		purposeText = "重置密码"
	default:
		purposeText = "验证"
	}

	// 构建发送请求
	sendReq := &SendMessageRequest{
		PhoneNumber: req.PhoneNumber,
		TemplateParam: map[string]string{
			"code":    req.Code,
			"purpose": purposeText,
		},
		Timeout: req.Timeout,
	}

	return p.SendMessage(ctx, sendReq)
}

// Close 关闭客户端连接
func (p *AliyunProvider) Close() error {
	// 阿里云SDK不需要显式关闭
	return nil
}
