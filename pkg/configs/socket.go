package configs

import "github.com/spf13/viper"

// SocketConfig websocket服务器配置
type SocketConfig struct {
	Port            string   `mapstructure:"port"` // 服务器监听端口
	ReadBufferSize  int      `mapstructure:"read_buffer_size"`
	WriteBufferSize int      `mapstructure:"write_buffer_size"`
	AllowOrigins    []string `mapstructure:"allow_origins"` // 允许的来源列表，空表示允许所有
	PingTimeout     int      `mapstructure:"ping_timeout"`  // 心跳超时时间，单位秒
}

func setSocketConfigDefaults() {
	viper.SetDefault("socket.port", "localhost:8088")
	viper.SetDefault("socket.read_buffer_size", 1024)
	viper.SetDefault("socket.write_buffer_size", 1024)
	viper.SetDefault("socket.allow_origins", []string{}) // 默认允许所有来源
	viper.SetDefault("socket.ping_timeout", 60)          // 默认60秒心跳超时
}
