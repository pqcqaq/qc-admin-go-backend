package openai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go-backend/pkg/configs"

	"github.com/sashabaranov/go-openai"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
	Debug(format string, args ...any)
}

// OpenAIClient OpenAI客户端结构
type OpenAIClient struct {
	config *configs.OpenAIConfig
	client *openai.Client
}

// 单例相关变量
var (
	Client *OpenAIClient
	once   sync.Once
	mu     sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的OpenAI客户端
func NewClient(openaiConfig *configs.OpenAIConfig) (*OpenAIClient, error) {
	if !openaiConfig.Enable {
		return nil, fmt.Errorf("OpenAI功能未启用")
	}

	if openaiConfig.APIKey == "" {
		return nil, fmt.Errorf("OpenAI配置不完整: api_key 不能为空")
	}

	// 创建HTTP客户端配置
	httpConfig := openai.DefaultConfig(openaiConfig.APIKey)

	// 设置基础URL
	if openaiConfig.BaseURL != "" {
		httpConfig.BaseURL = openaiConfig.BaseURL
	}

	// 设置组织ID
	if openaiConfig.OrgID != "" {
		httpConfig.OrgID = openaiConfig.OrgID
	}

	// 设置代理
	if openaiConfig.Proxy != "" {
		proxyURL, err := url.Parse(openaiConfig.Proxy)
		if err != nil {
			return nil, fmt.Errorf("代理地址格式错误: %w", err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		httpConfig.HTTPClient = &http.Client{
			Transport: transport,
			Timeout:   time.Duration(openaiConfig.Timeout) * time.Second,
		}
	} else {
		// 设置超时时间
		httpConfig.HTTPClient = &http.Client{
			Timeout: time.Duration(openaiConfig.Timeout) * time.Second,
		}
	}

	client := &OpenAIClient{
		config: openaiConfig,
		client: openai.NewClientWithConfig(httpConfig),
	}

	// 测试连接
	if err := client.TestConnection(); err != nil {
		return nil, fmt.Errorf("OpenAI服务连接测试失败: %w", err)
	}

	if logger != nil {
		logger.Info("OpenAI服务初始化成功: model=%s, base_url=%s", openaiConfig.Model, openaiConfig.BaseURL)
	}

	return client, nil
}

// GetClient 获取单例OpenAI客户端
func GetClient() *OpenAIClient {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// InitializeClient 初始化单例OpenAI客户端
func InitializeClient(openaiConfig *configs.OpenAIConfig) error {
	if !openaiConfig.Enable {
		logger.Info("OpenAiClient is not enabled skip init...")
		return nil
	}
	var err error
	once.Do(func() {
		Client, err = NewClient(openaiConfig)
	})
	return err
}

// TestConnection 测试OpenAI服务连接
func (c *OpenAIClient) TestConnection() error {
	if !c.config.Enable {
		if logger != nil {
			logger.Info("OpenAI服务未启用，跳过连接测试")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 发送一个简单的请求来测试连接
	_, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "t",
			},
		},
		MaxTokens: 1,
	})

	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	Model       string                         `json:"model,omitempty"`
	MaxTokens   int                            `json:"max_tokens,omitempty"`
	Temperature float32                        `json:"temperature,omitempty"`
	Stream      bool                           `json:"stream,omitempty"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	Content string `json:"content"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// CreateChatCompletion 创建聊天补全
func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 使用默认值填充请求
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = c.config.Temperature
	}

	// 创建OpenAI请求
	openaiReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	// 发送请求
	resp, err := c.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		if logger != nil {
			logger.Error("OpenAI聊天补全失败: %v", err)
		}
		return nil, fmt.Errorf("创建聊天补全失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI返回空响应")
	}

	response := &ChatResponse{
		Content: resp.Choices[0].Message.Content,
	}
	response.Usage.PromptTokens = resp.Usage.PromptTokens
	response.Usage.CompletionTokens = resp.Usage.CompletionTokens
	response.Usage.TotalTokens = resp.Usage.TotalTokens

	if logger != nil {
		logger.Debug("OpenAI聊天补全成功: tokens=%d", response.Usage.TotalTokens)
	}

	return response, nil
}

// SimpleChat 简单聊天接口
func (c *OpenAIClient) SimpleChat(ctx context.Context, message string) (string, error) {
	req := ChatRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// MultiRoundChat 多轮对话
func (c *OpenAIClient) MultiRoundChat(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	req := ChatRequest{
		Messages: messages,
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// CreateEmbedding 创建嵌入向量
func (c *OpenAIClient) CreateEmbedding(ctx context.Context, input string, model string) ([]float32, error) {
	if model == "" {
		model = "text-embedding-ada-002" // 默认嵌入模型
	}

	req := openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.EmbeddingModel(model),
	}

	resp, err := c.client.CreateEmbeddings(ctx, req)
	if err != nil {
		if logger != nil {
			logger.Error("OpenAI创建嵌入失败: %v", err)
		}
		return nil, fmt.Errorf("创建嵌入失败: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("OpenAI返回空嵌入数据")
	}

	return resp.Data[0].Embedding, nil
}

// 全局便捷函数

// SimpleChat 简单聊天 (全局函数)
func SimpleChat(ctx context.Context, message string) (string, error) {
	client := GetClient()
	if client == nil {
		return "", fmt.Errorf("OpenAI客户端未初始化")
	}
	return client.SimpleChat(ctx, message)
}

// MultiRoundChat 多轮对话 (全局函数)
func MultiRoundChat(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	client := GetClient()
	if client == nil {
		return "", fmt.Errorf("OpenAI客户端未初始化")
	}
	return client.MultiRoundChat(ctx, messages)
}

// CreateChatCompletion 创建聊天补全 (全局函数)
func CreateChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("OpenAI客户端未初始化")
	}
	return client.CreateChatCompletion(ctx, req)
}

// CreateEmbedding 创建嵌入向量 (全局函数)
func CreateEmbedding(ctx context.Context, input string, model string) ([]float32, error) {
	client := GetClient()
	if client == nil {
		return nil, fmt.Errorf("OpenAI客户端未初始化")
	}
	return client.CreateEmbedding(ctx, input, model)
}

// GetConfig 获取配置信息
func GetConfig() *configs.OpenAIConfig {
	client := GetClient()
	if client == nil {
		return nil
	}
	return client.config
}

// IsEnabled 检查OpenAI是否启用
func IsEnabled() bool {
	client := GetClient()
	if client == nil {
		return false
	}
	return client.config.Enable
}
