package configs

import "github.com/spf13/viper"

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // gin模式: debug, release, test
}

func setServerConfigDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.mode", "debug")
}
