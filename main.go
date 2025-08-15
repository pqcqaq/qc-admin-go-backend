package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-backend/internal/routes"
	"go-backend/pkg/configs"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	// 定义命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
	flag.StringVar(&configPath, "c", "config.yaml", "配置文件路径（简写）")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用方法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s                           # 使用默认配置文件 config.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c config.dev.yaml        # 使用开发环境配置\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config config.prod.yaml # 使用生产环境配置\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c /path/to/config.yaml   # 使用绝对路径配置\n", os.Args[0])
	}

	// 解析命令行参数
	flag.Parse()

	// 处理配置文件路径（支持相对路径和绝对路径）
	resolvedConfigPath, err := resolveConfigPath(configPath)
	if err != nil {
		log.Fatalf("Failed to resolve config path: %v", err)
	}

	// 加载配置
	config, err := configs.LoadConfig(resolvedConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志系统
	logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
	logging.SetPrefix(config.Logging.Prefix)

	// 设置数据库包的logger
	database.SetLogger(logging.GetInstance())

	logging.Info("Config loaded successfully from: %s", resolvedConfigPath)
	logging.Info("Log level set to: %s", config.Logging.Level)

	// 设置Gin模式
	gin.SetMode(config.Server.Mode)
	logging.Info("Gin mode set to: %s", config.Server.Mode)

	// 创建数据库连接
	client := database.InitInstance(&config.Database)
	defer client.Close()
	logging.Info("Database connection established")

	// 创建Gin引擎
	engine := gin.Default()

	// 设置路由
	router := routes.NewRouter()
	router.SetupRoutes(engine)
	logging.Info("Routes configured")

	// 启动服务器
	logging.Info("Server starting on %s", config.Server.Port)
	logging.Fatal("Server failed to start: %v", engine.Run(config.Server.Port))
}

// resolveConfigPath 解析配置文件路径，支持相对路径和绝对路径
func resolveConfigPath(configPath string) (string, error) {
	// 如果是绝对路径，直接返回
	if filepath.IsAbs(configPath) {
		// 检查文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found: %s", configPath)
		}
		return configPath, nil
	}

	// 相对路径：相对于当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	resolvedPath := filepath.Join(workDir, configPath)

	// 检查文件是否存在（可选，因为Viper会处理文件不存在的情况）
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		log.Printf("Warning: Config file not found at %s, will use defaults", resolvedPath)
	}

	return resolvedPath, nil
}
