package cmd

import (
	"context"
	"fmt"
	"go-backend/pkg/configs"
	pkgdatabase "go-backend/pkg/database"
	"go-backend/pkg/logging"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// export命令的参数变量
var (
	outputDir     string
	include       string
	exclude       string
	pretty        bool
	timeout       time.Duration
	showResult    bool
	excludeFields string
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

var exportDbCmd = &cobra.Command{
	Use:   "export",
	Short: "导出数据库表数据到JSON文件",
	Long:  "将数据库中的表数据导出为JSON文件，支持多种配置选项",
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

		// 设置日志
		logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
		logging.SetPrefix(config.Logging.Prefix)
		pkgdatabase.SetLogger(logging.WithName("Database"))

		// 创建数据库客户端
		client, err := pkgdatabase.NewClient(&config.Database)
		if err != nil {
			return fmt.Errorf("创建数据库客户端失败: %v", err)
		}
		defer client.Close()

		// 创建导出配置
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		exportConfig := &pkgdatabase.ExportConfig{
			OutputDir:    outputDir,
			PrettyFormat: pretty,
			Context:      ctx,
		}

		// 处理包含和排除列表
		if include != "" {
			exportConfig.IncludeEntities = strings.Split(include, ",")
			for i, entity := range exportConfig.IncludeEntities {
				exportConfig.IncludeEntities[i] = strings.TrimSpace(entity)
			}
		}

		if exclude != "" {
			exportConfig.ExcludeEntities = strings.Split(exclude, ",")
			for i, entity := range exportConfig.ExcludeEntities {
				exportConfig.ExcludeEntities[i] = strings.TrimSpace(entity)
			}
		}

		if excludeFields != "" {
			exportConfig.ExcludeFields = strings.Split(excludeFields, ",")
			for i, field := range exportConfig.ExcludeFields {
				exportConfig.ExcludeFields[i] = strings.TrimSpace(field)
			}
		}

		// 执行导出
		fmt.Println("开始导出数据库表...")
		result, err := pkgdatabase.ExportAllTables(client, exportConfig)
		if err != nil {
			return fmt.Errorf("导出表失败: %v", err)
		}

		// 显示导出结果
		fmt.Printf("导出完成！\n")
		fmt.Printf("总实体数: %d\n", result.TotalEntities)
		fmt.Printf("成功: %d\n", result.SuccessCount)
		fmt.Printf("失败: %d\n", result.FailedCount)
		fmt.Printf("输出目录: %s\n", result.OutputDirectory)

		if showResult {
			fmt.Println("\n详细结果:")
			for _, entityResult := range result.Results {
				if entityResult.Success {
					fmt.Printf("✓ %s: %d 条记录 -> %s\n",
						entityResult.EntityName, entityResult.RecordCount, entityResult.FilePath)
				} else {
					fmt.Printf("✗ %s: %s\n", entityResult.EntityName, entityResult.Error)
				}
			}
		}

		return nil
	},
}

func init() {
	// 添加db命令到root
	rootCmd.AddCommand(dbCmd)

	// 添加子命令到db
	dbCmd.AddCommand(migrateDbCmd)
	dbCmd.AddCommand(checkDbCmd)
	dbCmd.AddCommand(exportDbCmd)

	// 为export命令添加参数
	exportDbCmd.Flags().StringVarP(&outputDir, "output", "o", "./exports", "输出目录")
	exportDbCmd.Flags().StringVarP(&include, "include", "i", "", "仅导出指定的实体，用逗号分隔")
	exportDbCmd.Flags().StringVarP(&exclude, "exclude", "e", "", "排除指定的实体，用逗号分隔")
	exportDbCmd.Flags().BoolVarP(&pretty, "pretty", "p", true, "是否格式化JSON输出")
	exportDbCmd.Flags().DurationVarP(&timeout, "timeout", "t", time.Minute*10, "导出超时时间")
	exportDbCmd.Flags().BoolVarP(&showResult, "result", "r", false, "是否显示详细导出结果")
	exportDbCmd.Flags().StringVarP(&excludeFields, "exclude-fields", "f", "", "导出时排除指定的字段，多个字段用逗号分隔")
}
