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
	setServerConfigDefaults()

	// 数据库默认配置
	setDatabaseConfigDefaults()

	// 日志默认配置
	setLoggingConfigDefaults()

	// Redis默认配置
	setRedisConfigDefaults()
}
