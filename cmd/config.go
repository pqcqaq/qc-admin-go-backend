package cmd

import (
	"fmt"

	"go-backend/pkg/configs"
	"go-backend/pkg/logging"

	"github.com/spf13/cobra"
)

// configCmd 配置相关命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置文件相关操作",
	Long:  "提供配置文件的验证、显示等操作",
}

// validateCmd 验证配置文件
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证配置文件",
	Long:  "验证配置文件的语法和配置项是否正确",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolvedConfigPath, err := configs.ResolveConfigPath(configFile)
		if err != nil {
			return fmt.Errorf("解析配置文件路径失败: %w", err)
		}

		_, err = configs.LoadConfig(resolvedConfigPath)
		if err != nil {
			return fmt.Errorf("配置文件验证失败: %w", err)
		}

		logging.Info("配置文件验证通过: %s\n", resolvedConfigPath)
		return nil
	},
}

func init() {
	// 添加config命令到root
	rootCmd.AddCommand(configCmd)

	// 添加子命令到config
	configCmd.AddCommand(validateCmd)
}
