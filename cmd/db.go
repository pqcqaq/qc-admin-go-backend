package cmd

import (
	"fmt"

	"go-backend/pkg/configs"
	pkgdatabase "go-backend/pkg/database"
	"go-backend/pkg/logging"

	"github.com/spf13/cobra"
)

// dbCmd 数据库相关命令
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "数据库操作命令",
	Long:  "提供数据库相关的操作，如迁移、检查等",
}

// migrateDbCmd 数据库迁移命令
var migrateDbCmd = &cobra.Command{
	Use:   "migrate",
	Short: "执行数据库迁移",
	Long:  "执行数据库模式迁移，将数据库结构更新到最新版本",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		resolvedConfigPath, err := configs.ResolveConfigPath(configFile)
		if err != nil {
			return fmt.Errorf("解析配置文件路径失败: %w", err)
		}

		config, err := configs.LoadConfig(resolvedConfigPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 强制启用自动迁移
		config.Database.AutoMigrate = true
		config.Database.SkipMigrateCheck = false

		config.Database.ConnectionCheckInterval = 0 // 禁用连接检查间隔

		// 设置日志
		logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
		logging.SetPrefix(config.Logging.Prefix)
		pkgdatabase.SetLogger(logging.WithName("Database"))

		logging.Info("正在执行数据库迁移...")

		// 初始化数据库
		dbClient := pkgdatabase.InitInstance(&config.Database)
		defer dbClient.Close()

		logging.Info("数据库迁移完成")
		return nil
	},
}

// checkDbCmd 检查数据库连接
var checkDbCmd = &cobra.Command{
	Use:   "check",
	Short: "检查数据库连接",
	Long:  "检查数据库连接是否正常",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		resolvedConfigPath, err := configs.ResolveConfigPath(configFile)
		if err != nil {
			return fmt.Errorf("解析配置文件路径失败: %w", err)
		}

		config, err := configs.LoadConfig(resolvedConfigPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 跳过迁移检查，只测试连接
		config.Database.SkipMigrateCheck = true
		config.Database.AutoMigrate = false

		config.Database.ConnectionCheckInterval = 0 // 禁用连接检查间隔

		// 设置日志
		logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
		logging.SetPrefix(config.Logging.Prefix)
		pkgdatabase.SetLogger(logging.WithName("Database"))

		logging.Info("正在检查数据库连接...")

		// 测试数据库连接
		dbClient := pkgdatabase.InitInstance(&config.Database)
		defer dbClient.Close()

		// 执行一个简单的查询来验证连接
		ctx := cmd.Context()
		tx, err := dbClient.Tx(ctx)

		defer tx.Commit()

		if err != nil {
			return fmt.Errorf("数据库连接测试失败: %w", err)
		}
		_, err = tx.ExecContext(ctx, "SELECT 1")
		if err != nil {
			return fmt.Errorf("数据库连接测试失败: %w", err)
		}

		logging.Info("数据库连接正常\n")
		return nil
	},
}

func init() {
	// 添加db命令到root
	rootCmd.AddCommand(dbCmd)

	// 添加子命令到db
	dbCmd.AddCommand(migrateDbCmd)
	dbCmd.AddCommand(checkDbCmd)
}
