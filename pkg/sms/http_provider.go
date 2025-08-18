package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-backend/pkg/configs"
)

// HTTPProvider HTTP短信提供商
type HTTPProvider struct {
	config     *configs.HTTPSMSConfig
	httpClient *http.Client
}

// NewHTTPProvider 创建HTTP短信提供商
func NewHTTPProvider(config *configs.HTTPSMSConfig) (*HTTPProvider, error) {
	provider := &HTTPProvider{
		config: config,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	// 创建HTTP客户端
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	provider.httpClient = &http.Client{
		Timeout: timeout,
	}

	if logger != nil {
		logger.Info("HTTP短信提供商初始化成功: url=%s", config.URL)
	}

	return provider, nil
}

// Name 返回提供商名称
func (p *HTTPProvider) Name() string {
	return "http"
}

// ValidateConfig 验证配置是否有效
func (p *HTTPProvider) ValidateConfig() error {
	if p.config.URL == "" {
		return fmt.Errorf("HTTP短信配置错误: url 不能为空")
	}

	// 验证HTTP方法
	method := strings.ToUpper(p.config.Method)
	if method == "" {
		p.config.Method = "POST"
	} else if method != "GET" && method != "POST" && method != "PUT" {
		return fmt.Errorf("HTTP短信配置错误: 不支持的HTTP方法 %s", method)
	}

	// 验证认证类型
	if p.config.AuthType != "" {
		switch p.config.AuthType {
		case "basic", "bearer", "api_key":
			// 支持的认证类型
		default:
			return fmt.Errorf("HTTP短信配置错误: 不支持的认证类型 %s", p.config.AuthType)
		}
	}

	return nil
}

// SendMessage 发送短信消息
func (p *HTTPProvider) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// 设置默认超时时间
	if req.Timeout == 0 {
		req.Timeout = time.Duration(p.config.Timeout) * time.Second
		if req.Timeout == 0 {
			req.Timeout = 30 * time.Second
		}
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

	// 构建请求参数
	requestData := map[string]interface{}{
		"phone_number":   req.PhoneNumber,
		"template_id":    templateID,
		"sign_name":      signName,
		"template_param": req.TemplateParam,
		"content":        req.Content,
	}

	// 创建HTTP请求
	httpReq, err := p.createHTTPRequest(ctx, requestData)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()
	httpReq = httpReq.WithContext(timeoutCtx)

	// 发送HTTP请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		if logger != nil {
			logger.Error("HTTP短信发送失败: phone=%s, error=%v", req.PhoneNumber, err)
		}
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	result, err := p.parseResponse(resp.StatusCode, body)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Success && logger != nil {
		logger.Info("HTTP短信发送成功: phone=%s, messageId=%s", req.PhoneNumber, result.MessageID)
	} else if !result.Success && logger != nil {
		logger.Error("HTTP短信发送失败: phone=%s, code=%s, message=%s",
			req.PhoneNumber, result.Code, result.Message)
	}

	return result, nil
}

// createHTTPRequest 创建HTTP请求
func (p *HTTPProvider) createHTTPRequest(ctx context.Context, data map[string]interface{}) (*http.Request, error) {
	method := strings.ToUpper(p.config.Method)
	if method == "" {
		method = "POST"
	}

	var req *http.Request
	var err error

	if method == "GET" {
		// GET请求，将参数添加到URL
		var u *url.URL
		u, err = url.Parse(p.config.URL)
		if err != nil {
			return nil, fmt.Errorf("解析URL失败: %w", err)
		}

		values := u.Query()
		for key, value := range data {
			if value != nil {
				values.Add(key, fmt.Sprintf("%v", value))
			}
		}
		u.RawQuery = values.Encode()

		req, err = http.NewRequestWithContext(ctx, method, u.String(), nil)
	} else {
		// POST/PUT请求，将参数放在请求体中
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, method, p.config.URL, bytes.NewBuffer(jsonData))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证
	if err := p.setAuthentication(req); err != nil {
		return nil, fmt.Errorf("设置认证失败: %w", err)
	}

	// 设置自定义头部
	for key, value := range p.config.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// setAuthentication 设置认证信息
func (p *HTTPProvider) setAuthentication(req *http.Request) error {
	switch p.config.AuthType {
	case "basic":
		if p.config.Username == "" || p.config.Password == "" {
			return fmt.Errorf("basic认证需要username和password")
		}
		req.SetBasicAuth(p.config.Username, p.config.Password)

	case "bearer":
		if p.config.Token == "" {
			return fmt.Errorf("bearer认证需要token")
		}
		req.Header.Set("Authorization", "Bearer "+p.config.Token)

	case "api_key":
		if p.config.APIKey == "" {
			return fmt.Errorf("API Key认证需要api_key")
		}
		// 常见的API Key认证方式
		req.Header.Set("X-API-Key", p.config.APIKey)
		if p.config.APISecret != "" {
			req.Header.Set("X-API-Secret", p.config.APISecret)
		}
	}

	return nil
}

// parseResponse 解析HTTP响应
func (p *HTTPProvider) parseResponse(statusCode int, body []byte) (*SendMessageResponse, error) {
	result := &SendMessageResponse{
		Success: statusCode >= 200 && statusCode < 300,
		Code:    fmt.Sprintf("%d", statusCode),
		Message: string(body),
	}

	// 尝试解析JSON响应
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err == nil {
		// 标准化字段映射 (可根据实际API调整)
		if code, ok := jsonResp["code"]; ok {
			result.Code = fmt.Sprintf("%v", code)
		}
		if message, ok := jsonResp["message"]; ok {
			result.Message = fmt.Sprintf("%v", message)
		}
		if messageID, ok := jsonResp["message_id"]; ok {
			result.MessageID = fmt.Sprintf("%v", messageID)
		}
		if bizID, ok := jsonResp["biz_id"]; ok {
			result.BizID = fmt.Sprintf("%v", bizID)
		}
		if success, ok := jsonResp["success"]; ok {
			if successBool, ok := success.(bool); ok {
				result.Success = successBool
			}
		}
	}

	return result, nil
}

// SendVerificationCode 发送验证码短信
func (p *HTTPProvider) SendVerificationCode(ctx context.Context, req *VerificationCodeRequest) (*SendMessageResponse, error) {
	// HTTP提供商支持灵活的参数配置
	templateParam := map[string]string{
		"code":    req.Code,
		"purpose": req.Purpose,
	}

	sendReq := &SendMessageRequest{
		PhoneNumber:   req.PhoneNumber,
		TemplateParam: templateParam,
		Timeout:       req.Timeout,
	}

	return p.SendMessage(ctx, sendReq)
}

// Close 关闭客户端连接
func (p *HTTPProvider) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}
