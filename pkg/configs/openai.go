package configs

import "github.com/spf13/viper"

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	Enable      bool   `mapstructure:"enable"`       // 是否启用OpenAI功能
	APIKey      string `mapstructure:"api_key"`      // OpenAI API密钥
	BaseURL     string `mapstructure:"base_url"`     // API基础URL
	Model       string `mapstructure:"model"`        // 默认模型
	MaxTokens   int    `mapstructure:"max_tokens"`   // 最大token数
	Temperature float32 `mapstructure:"temperature"` // 温度参数
	Timeout     int    `mapstructure:"timeout"`      // 请求超时时间(秒)
	Proxy       string `mapstructure:"proxy"`        // 代理地址
	OrgID       string `mapstructure:"org_id"`       // 组织ID
}

// setOpenAIConfigDefaults 设置OpenAI默认配置
func setOpenAIConfigDefaults() {
	viper.SetDefault("openai.enable", false)
	viper.SetDefault("openai.api_key", "")
	viper.SetDefault("openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("openai.model", "gpt-3.5-turbo")
	viper.SetDefault("openai.max_tokens", 1500)
	viper.SetDefault("openai.temperature", 0.7)
	viper.SetDefault("openai.timeout", 30)
	viper.SetDefault("openai.proxy", "")
	viper.SetDefault("openai.org_id", "")
}

// GetOpenAIConfig 获取OpenAI配置
func GetOpenAIConfig() *OpenAIConfig {
	setOpenAIConfigDefaults()
	
	return &OpenAIConfig{
		Enable:      viper.GetBool("openai.enable"),
		APIKey:      viper.GetString("openai.api_key"),
		BaseURL:     viper.GetString("openai.base_url"),
		Model:       viper.GetString("openai.model"),
		MaxTokens:   viper.GetInt("openai.max_tokens"),
		Temperature: float32(viper.GetFloat64("openai.temperature")),
		Timeout:     viper.GetInt("openai.timeout"),
		Proxy:       viper.GetString("openai.proxy"),
		OrgID:       viper.GetString("openai.org_id"),
	}
}