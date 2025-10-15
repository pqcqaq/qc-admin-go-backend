package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	Socket   SocketConfig   `mapstructure:"socket"`
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

	// 处理配置导入
	if err := processConfigImports(filepath.Dir(configPath)); err != nil {
		return nil, fmt.Errorf("处理配置导入失败: %w", err)
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

// processConfigImports 处理配置文件导入
func processConfigImports(baseDir string) error {
	imports := viper.GetStringSlice("config.import")
	if len(imports) == 0 {
		return nil
	}

	logger := logging.WithName("Config Resolver")
	for _, importPath := range imports {

		logger.Info("Importing config for: %s", importPath)

		// 解析配置变量引用
		resolvedPath := ResolveConfigVariables(importPath)

		// 去除file:前缀
		resolvedPath = strings.TrimPrefix(resolvedPath, "file:")

		// 处理相对路径
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(baseDir, resolvedPath)
		}

		if err := mergeConfigFile(resolvedPath); err != nil {
			logger.Error("导入配置文件失败 %s", resolvedPath)
			return fmt.Errorf("导入配置文件失败 %s: %w", resolvedPath, err)
		}
		logger.Info("Successfully import config from: %s", resolvedPath)
	}

	return nil
}

// mergeConfigFile 合并配置文件
func mergeConfigFile(configPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 创建临时viper实例
	tempViper := viper.New()
	tempViper.SetConfigFile(configPath)

	if err := tempViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 合并到主viper实例
	return viper.MergeConfigMap(tempViper.AllSettings())
}

// HasPlaceholder 检查字符串是否包含配置变量占位符
func HasPlaceholder(s string) bool {
	return strings.Contains(s, "${") && strings.Contains(s, "}")
}

// ReplaceKey 替换字符串中的第一个配置变量占位符
func ReplaceKey(s string) string {
	start := strings.Index(s, "${")
	if start == -1 {
		return s
	}

	end := strings.Index(s[start:], "}")
	if end == -1 {
		return s
	}
	end += start

	// 提取配置键名
	configKey := s[start+2 : end]

	// 先尝试从当前配置中获取值
	configValue := viper.GetString(configKey)

	// 如果配置中没有，再尝试环境变量
	if configValue == "" {
		configValue = os.Getenv(configKey)
	}

	// 如果还是没有，记录错误
	if configValue == "" {
		logging.Error("Cannot find Config Key: %s", configKey)
	}

	// 替换占位符并返回
	return s[:start] + configValue + s[end+1:]
}

// ResolveConfigVariables 解析配置变量引用
func ResolveConfigVariables(path string) string {
	resolved := path

	// 循环处理所有占位符，每次替换一个
	for HasPlaceholder(resolved) {
		resolved = ReplaceKey(resolved)
	}

	return resolved
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

	// SocketIO默认配置
	setSocketConfigDefaults()
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
