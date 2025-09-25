package configs

import (
	"github.com/spf13/viper"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"` // JWT密钥
	Issuer    string `mapstructure:"issuer"`     // 签发者
}

// setJWTConfigDefaults 设置JWT默认配置
func setJWTConfigDefaults() {
	viper.SetDefault("jwt.secret_key", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("jwt.issuer", "go-backend")
}
