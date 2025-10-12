package components

import "github.com/spf13/viper"

type MonitorConfig struct {
	Enabled       bool  `mapstructure:"enabled"`        // 是否启用系统监控
	Interval      int64 `mapstructure:"interval"`       // 监控数据采集间隔（秒）
	RetentionDays int   `mapstructure:"retention_days"` // 监控数据保留天数
}

func setMinitorConfigDefaults() {
	viper.SetDefault("server.components.monitor.enabled", true)
	viper.SetDefault("server.components.monitor.interval", 60)      // 60秒
	viper.SetDefault("server.components.monitor.retention_days", 7) // 7天
}
