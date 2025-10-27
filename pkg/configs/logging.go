package configs

import "github.com/spf13/viper"

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level   string               `mapstructure:"level"`   // 日志级别: debug, info, warn, error, fatal
	Prefix  string               `mapstructure:"prefix"`  // 日志前缀
	Console LoggingConsoleConfig `mapstructure:"console"` // 控制台日志配置
	File    LoggingFileConfig    `mapstructure:"file"`    // 文件日志配置
}

type LoggingConsoleConfig struct {
	Enabled bool `mapstructure:"enabled"` // 是否启用控制台日志
}

type LoggingFileConfig struct {
	Enabled      bool                  `mapstructure:"enabled"`        // 是否启用文件日志
	Path         string                `mapstructure:"path"`           // 日志文件存储路径
	SplitByLevel bool                  `mapstructure:"split_by_level"` // 是否按级别分割日志文件
	Filenames    LoggingFileNameConfig `mapstructure:"filenames"`      // 各级别日志文件名
	MaxSize      int                   `mapstructure:"max_size"`       // 单个日志文件最大尺寸（MB）
	MaxBackups   int                   `mapstructure:"max_backups"`    // 最大备份数量
	MaxAge       int                   `mapstructure:"max_age"`        // 最大保存天数
	Compress     bool                  `mapstructure:"compress"`       // 是否压缩备份日志
}

type LoggingFileNameConfig struct {
	Debug string `mapstructure:"debug"`
	Info  string `mapstructure:"info"`
	Warn  string `mapstructure:"warn"`
	Error string `mapstructure:"error"`
	Fatal string `mapstructure:"fatal"`
	All   string `mapstructure:"all"`
}

func setLoggingConfigDefaults() {
	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.prefix", "APP")
	viper.SetDefault("logging.console.enabled", true)
	viper.SetDefault("logging.file.enabled", false)
	viper.SetDefault("logging.file.path", "./logs")
	viper.SetDefault("logging.file.split_by_level", true)
	viper.SetDefault("logging.file.filenames.debug", "debug.log")
	viper.SetDefault("logging.file.filenames.info", "info.log")
	viper.SetDefault("logging.file.filenames.warn", "warn.log")
	viper.SetDefault("logging.file.filenames.error", "error.log")
	viper.SetDefault("logging.file.filenames.fatal", "fatal.log")
	viper.SetDefault("logging.file.filenames.all", "all.log")
	viper.SetDefault("logging.file.max_size", 100)
	viper.SetDefault("logging.file.max_backups", 10)
	viper.SetDefault("logging.file.max_age", 7)
	viper.SetDefault("logging.file.compress", false)
}
