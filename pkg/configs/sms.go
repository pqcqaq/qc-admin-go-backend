package configs

import "github.com/spf13/viper"

// SMSConfig 短信配置
type SMSConfig struct {
	Provider string                 `mapstructure:"provider"` // 短信服务提供商 (aliyun, tencent, http)
	Aliyun   AliyunSMSConfig        `mapstructure:"aliyun"`   // 阿里云短信配置
	Tencent  TencentSMSConfig       `mapstructure:"tencent"`  // 腾讯云短信配置
	HTTP     HTTPSMSConfig          `mapstructure:"http"`     // HTTP短信配置
	Extra    map[string]interface{} `mapstructure:"extra"`    // 额外配置
}

// AliyunSMSConfig 阿里云短信配置
type AliyunSMSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`     // 访问密钥ID
	AccessKeySecret string `mapstructure:"access_key_secret"` // 访问密钥Secret
	SignName        string `mapstructure:"sign_name"`         // 短信签名
	TemplateCode    string `mapstructure:"template_code"`     // 短信模板代码
	RegionID        string `mapstructure:"region_id"`         // 地域ID
	Endpoint        string `mapstructure:"endpoint"`          // 服务端点
}

// TencentSMSConfig 腾讯云短信配置
type TencentSMSConfig struct {
	SecretID   string `mapstructure:"secret_id"`   // 腾讯云SecretId
	SecretKey  string `mapstructure:"secret_key"`  // 腾讯云SecretKey
	Region     string `mapstructure:"region"`      // 地域
	AppID      string `mapstructure:"app_id"`      // 短信SdkAppId
	SignName   string `mapstructure:"sign_name"`   // 短信签名
	TemplateID string `mapstructure:"template_id"` // 短信模板ID
	Endpoint   string `mapstructure:"endpoint"`    // 服务端点
}

// HTTPSMSConfig HTTP短信配置
type HTTPSMSConfig struct {
	URL        string            `mapstructure:"url"`         // HTTP API地址
	Method     string            `mapstructure:"method"`      // HTTP方法 (GET, POST)
	Headers    map[string]string `mapstructure:"headers"`     // HTTP头
	AuthType   string            `mapstructure:"auth_type"`   // 认证类型 (basic, bearer, api_key)
	Username   string            `mapstructure:"username"`    // 用户名 (basic auth)
	Password   string            `mapstructure:"password"`    // 密码 (basic auth)
	Token      string            `mapstructure:"token"`       // Token (bearer auth)
	APIKey     string            `mapstructure:"api_key"`     // API Key
	APISecret  string            `mapstructure:"api_secret"`  // API Secret
	SignName   string            `mapstructure:"sign_name"`   // 短信签名
	TemplateID string            `mapstructure:"template_id"` // 模板ID
	Timeout    int               `mapstructure:"timeout"`     // 超时时间(秒)
}

// setSMSConfigDefaults 设置短信默认配置
func setSMSConfigDefaults() {
	viper.SetDefault("sms.provider", "aliyun")

	// 阿里云默认配置
	viper.SetDefault("sms.aliyun.access_key_id", "")
	viper.SetDefault("sms.aliyun.access_key_secret", "")
	viper.SetDefault("sms.aliyun.sign_name", "")
	viper.SetDefault("sms.aliyun.template_code", "")
	viper.SetDefault("sms.aliyun.region_id", "cn-hangzhou")
	viper.SetDefault("sms.aliyun.endpoint", "dysmsapi.aliyuncs.com")

	// 腾讯云默认配置
	viper.SetDefault("sms.tencent.secret_id", "")
	viper.SetDefault("sms.tencent.secret_key", "")
	viper.SetDefault("sms.tencent.region", "ap-guangzhou")
	viper.SetDefault("sms.tencent.app_id", "")
	viper.SetDefault("sms.tencent.sign_name", "")
	viper.SetDefault("sms.tencent.template_id", "")
	viper.SetDefault("sms.tencent.endpoint", "sms.tencentcloudapi.com")

	// HTTP默认配置
	viper.SetDefault("sms.http.url", "")
	viper.SetDefault("sms.http.method", "POST")
	viper.SetDefault("sms.http.auth_type", "api_key")
	viper.SetDefault("sms.http.timeout", 30)
}
