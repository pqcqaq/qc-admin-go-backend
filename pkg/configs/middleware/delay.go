package middleware

import (
	"time"

	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置
type DelayConfig struct {
	Enabled bool          `mapstructure:"enabled"` // 是否启用延迟中间件\
	Min     time.Duration `mapstructure:"min"`     // 最小延迟时间
	Max     time.Duration `mapstructure:"max"`     // 最大延迟时间
}

func setDelayConfigDefaults() {
	// 延迟中间件默认配置
	viper.SetDefault("server.middleware.delay.enabled", false)            // 默认不启用延迟中间件
	viper.SetDefault("server.middleware.delay.min", 100*time.Millisecond) // 默认最小延迟时间为100毫秒
	viper.SetDefault("server.middleware.delay.max", 1*time.Second)        // 默认最大延迟时间为1秒
}
