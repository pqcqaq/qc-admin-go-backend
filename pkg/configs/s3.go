package configs

import "github.com/spf13/viper"

// S3Config S3配置
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`          // S3端点URL
	PublicEndpoint  string `mapstructure:"public_endpoint"`   // 公共访问端点URL
	Region          string `mapstructure:"region"`            // AWS区域
	AccessKeyID     string `mapstructure:"access_key_id"`     // AWS访问密钥ID
	SecretAccessKey string `mapstructure:"secret_access_key"` // AWS访问密钥
	SessionToken    string `mapstructure:"session_token"`     // AWS会话令牌（可选）
	Bucket          string `mapstructure:"bucket"`            // 默认存储桶
	UseSSL          bool   `mapstructure:"use_ssl"`           // 是否使用HTTPS
	ForcePathStyle  bool   `mapstructure:"force_path_style"`  // 是否强制使用路径样式URL
	DisableSSL      bool   `mapstructure:"disable_ssl"`       // 是否禁用SSL
	Timeout         int    `mapstructure:"timeout"`           // 超时时间（秒）
	MaxRetries      int    `mapstructure:"max_retries"`       // 最大重试次数
}

// setS3ConfigDefaults 设置S3默认配置
func setS3ConfigDefaults() {
	viper.SetDefault("s3.endpoint", "")
	viper.SetDefault("s3.public_endpoint", "")
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.access_key_id", "")
	viper.SetDefault("s3.secret_access_key", "")
	viper.SetDefault("s3.session_token", "")
	viper.SetDefault("s3.bucket", "default-bucket")
	viper.SetDefault("s3.use_ssl", true)
	viper.SetDefault("s3.force_path_style", false)
	viper.SetDefault("s3.disable_ssl", false)
	viper.SetDefault("s3.timeout", 30)
	viper.SetDefault("s3.max_retries", 3)
}
