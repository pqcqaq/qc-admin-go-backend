package configs

import "github.com/spf13/viper"

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别: debug, info, warn, error, fatal
	Prefix string `mapstructure:"prefix"` // 日志前缀
}

func setLoggingConfigDefaults() {
	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.prefix", "APP")
}
