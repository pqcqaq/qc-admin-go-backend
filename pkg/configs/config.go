package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-backend/pkg/logging"

	"github.com/spf13/viper"
)

// AppConfig 应用配置
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Redis    RedisConfig    `mapstructure:"redis"`
	S3       S3Config       `mapstructure:"s3"`
	Email    EmailConfig    `mapstructure:"email"`
	SMS      SMSConfig      `mapstructure:"sms"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
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

	// S3默认配置
	setS3ConfigDefaults()

	// 邮件默认配置
	setEmailConfigDefaults()

	// 短信默认配置
	setSMSConfigDefaults()

	// JWT默认配置
	setJWTConfigDefaults()

	// OpenAI默认配置
	setOpenAIConfigDefaults()
}

// ResolveConfigPath 解析配置文件路径，支持相对路径和绝对路径
func ResolveConfigPath(configPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(configPath) {
		// 检查文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return "", fmt.Errorf("配置文件不存在: %s", configPath)
		}
		return configPath, nil
	}

	// 相对路径：相对于当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}

	resolvedPath := filepath.Join(workDir, configPath)

	// 检查文件是否存在（可选，因为Viper会处理文件不存在的情况）
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		log.Printf("警告: 配置文件不存在 %s, 将使用默认值", resolvedPath)
	}

	return resolvedPath, nil
}

// ResolveStaticPath 解析静态文件目录路径，支持相对路径和绝对路径
func ResolveStaticPath(staticPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(staticPath) {
		// 检查目录是否存在
		if _, err := os.Stat(staticPath); os.IsNotExist(err) {
			return "", fmt.Errorf("静态文件目录不存在: %s", staticPath)
		}
		return staticPath, nil
	}

	// 相对路径：相对于当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}

	resolvedPath := filepath.Join(workDir, staticPath)

	// 检查目录是否存在，如果不存在则创建
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		log.Printf("静态文件目录不存在，正在创建: %s", resolvedPath)
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return "", fmt.Errorf("创建静态文件目录失败: %w", err)
		}
	}

	return resolvedPath, nil
}
