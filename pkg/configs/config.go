package configs

import (
	"fmt"

	"go-backend/pkg/logging"

	"github.com/spf13/viper"
)

// AppConfig 应用配置
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // gin模式: debug, release, test
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别: debug, info, warn, error, fatal
	Prefix string `mapstructure:"prefix"` // 日志前缀
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

var config *AppConfig

// LoadConfig 从YAML文件加载配置
func LoadConfig(configPath string) (*AppConfig, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		logging.Warn("Warning: Config file not found, using defaults: %v", err)
	}

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	config = &AppConfig{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return config, nil
}

// GetConfig 获取当前配置
func GetConfig() *AppConfig {
	if config == nil {
		logging.Fatal("Config not loaded. Call LoadConfig first.")
	}
	return config
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("server.mode", "debug")

	// 数据库默认配置
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "file:ent.db?cache=shared&_fk=1")

	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.prefix", "APP")

	// Redis默认配置
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.read_timeout", 3)
	viper.SetDefault("redis.write_timeout", 3)
	viper.SetDefault("redis.idle_timeout", 300)
}
