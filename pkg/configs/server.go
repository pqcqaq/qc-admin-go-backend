package configs

import "github.com/spf13/viper"

// ServerConfig 服务器配置
type ServerConfig struct {
	Port   string       `mapstructure:"port"`
	Mode   string       `mapstructure:"mode"`   // gin模式: debug, release, test
	Static StaticConfig `mapstructure:"static"` // 静态文件服务配置
	Debug  bool         `mapstructure:"debug"`  // 是否启用调试模式
	CORS   CORSConfig   `mapstructure:"cors"`   // 跨域配置
}

type StaticConfig struct {
	Enabled bool   `mapstructure:"enabled"` // 是否启用静态文件服务
	Root    string `mapstructure:"root"`    // 静态文件根目录
	Path    string `mapstructure:"path"`    // 静态文件访问路径
}

type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`           // 是否启用CORS
	AllowAllOrigins  bool     `mapstructure:"allow_all_origins"` // 是否允许所有来源
	AllowOrigins     []string `mapstructure:"allow_origins"`     // 允许的来源列表
	AllowMethods     []string `mapstructure:"allow_methods"`     // 允许的HTTP方法
	AllowHeaders     []string `mapstructure:"allow_headers"`     // 允许的请求头
	ExposeHeaders    []string `mapstructure:"expose_headers"`    // 暴露的响应头
	AllowCredentials bool     `mapstructure:"allow_credentials"` // 是否允许携带凭证
	MaxAge           int      `mapstructure:"max_age"`           // 预检请求缓存时间（秒）
}

func setServerConfigDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.debug", false)
	viper.SetDefault("server.static.enabled", true)
	viper.SetDefault("server.static.root", "../public")
	viper.SetDefault("server.static.path", "/static")

	// CORS默认配置
	viper.SetDefault("server.cors.enabled", true)
	viper.SetDefault("server.cors.allow_all_origins", false)
	viper.SetDefault("server.cors.allow_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("server.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("server.cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"})
	viper.SetDefault("server.cors.expose_headers", []string{})
	viper.SetDefault("server.cors.allow_credentials", true)
	viper.SetDefault("server.cors.max_age", 86400) // 24小时
}
