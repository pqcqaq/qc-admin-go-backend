package main

import (
	"os"

	"go-backend/pkg/configs"

	"github.com/spf13/cobra"

	// 导入ent runtime以注册schema hooks

	_ "go-backend/database/ent/runtime"
)

var (
	// 全局配置变量
	configFile string
	serverPort string
	logLevel   string
)

// rootCmd 代表没有调用子命令时的基础命令
var rootCmd = &cobra.Command{
	Use:           "go-backend",
	Short:         "Go Backend Socket服务器",
	Long:          `go-backend是一个基于Go语言的后端Socket服务器`,
	RunE:          runServer,
	SilenceUsage:  true, // 运行失败时不显示usage信息
	SilenceErrors: true, // 不自动打印错误信息
}

// Execute 添加所有子命令到root命令并设置flags
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 持久化flags，所有子命令都可以使用
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "日志级别 (debug|info|warn|error)")

	// 本地flags，只有root命令可以使用
	rootCmd.Flags().StringVarP(&serverPort, "port", "p", "", "服务器端口 (覆盖配置文件)")
}

// applyFlagOverrides 应用命令行参数覆盖配置
func applyFlagOverrides(config *configs.AppConfig) {
	// 覆盖服务器端口
	if serverPort != "" {
		config.Server.Port = serverPort
		if config.Server.Port[0] != ':' {
			config.Server.Port = ":" + config.Server.Port
		}
	}

	// 覆盖日志级别
	if logLevel != "" {
		config.Logging.Level = logLevel
	}
}
