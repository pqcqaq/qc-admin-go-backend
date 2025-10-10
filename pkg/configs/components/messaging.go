package components

import "github.com/spf13/viper"

type MessagingConfig struct {
	Enabled     bool          `mapstructure:"enabled"`      // 是否启用消息处理器
	GroupName   string        `mapstructure:"group_name"`   // 消费者组名称
	StreamKey   string        `mapstructure:"stream_key"`   // Redis 流键名
	MaxRetries  int64         `mapstructure:"max_retries"`  // 最大重试次数
	ReadTimeout int64         `mapstructure:"read_timeout"` // 读取消息的阻塞超时时间（毫秒）
	ReadCount   int64         `mapstructure:"read_count"`   // 每次读取的消息数量
	IdleTimeout int64         `mapstructure:"idle_timeout"` // 消息空闲超时时间（毫秒）
	Cleanup     CleanupConfig `mapstructure:"cleanup"`      // 清理配置
}

type CleanupConfig struct {
	Enabled          bool  `mapstructure:"enabled"`             // 是否启用清理功能
	Interval         int64 `mapstructure:"interval"`            // 清理间隔（秒）
	MaxLen           int64 `mapstructure:"max_len"`             // Stream最大长度
	MaxAge           int64 `mapstructure:"max_age"`             // 消息最大保留时间（秒）
	DeadLetterMaxAge int64 `mapstructure:"dead_letter_max_age"` // 死信队列最大保留时间（秒）
}

func setMessagingConfigDefaults() {
	// 消息处理器默认配置
	viper.SetDefault("server.components.messaging.enabled", true)
	viper.SetDefault("server.components.messaging.group_name", "qc_admin_default_group")
	viper.SetDefault("server.components.messaging.stream_key", "qc_admin_stream")
	viper.SetDefault("server.components.messaging.max_retries", 1)
	viper.SetDefault("server.components.messaging.read_timeout", 2000) // 2s
	viper.SetDefault("server.components.messaging.read_count", 1)
	viper.SetDefault("server.components.messaging.idle_timeout", 60000) // 60s

	// 清理配置默认值
	viper.SetDefault("server.components.messaging.cleanup.enabled", true)
	viper.SetDefault("server.components.messaging.cleanup.interval", 3600)               // 1小时
	viper.SetDefault("server.components.messaging.cleanup.max_len", 0)                   // 0表示不使用长度清理
	viper.SetDefault("server.components.messaging.cleanup.max_age", 0)                   // 0表示不使用时间清理
	viper.SetDefault("server.components.messaging.cleanup.dead_letter_max_age", 2592000) // 30天
}
