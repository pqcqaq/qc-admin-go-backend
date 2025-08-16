package database

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	database "go-backend/database/ent"
	_ "go-backend/database/ent/runtime"

	"go-backend/pkg/configs"
	"go-backend/pkg/database/drivers"
	"go-backend/pkg/logging"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// DatabaseConfig 数据库配置结构

// 单例相关变量
var (
	once sync.Once
	mu   sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

var Client *database.Client

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的数据库客户端
func NewClient(config *configs.DatabaseConfig) (*database.Client, error) {
	// 验证驱动是否被支持
	if !drivers.IsDriverSupported(config.Driver) {
		supportedDrivers := drivers.GetSupportedDrivers()
		return nil, fmt.Errorf("unsupported database driver: %s. Supported drivers: %v", config.Driver, supportedDrivers)
	}

	client, err := database.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 如果配置了自动迁移，则创建数据库模式
	if !config.SkipMigrateCheck {
		if config.AutoMigrate {
			if err := client.Schema.Create(context.Background()); err != nil {
				client.Close()
				return nil, fmt.Errorf("failed to create database schema: %w", err)
			}
		} else {
			// 如果没有配置自动迁移，则检查是否需要迁移
			need, err := checkMigrationNeeded(client)
			if err != nil {
				logger.Error("Migration check failed: %v", err)
				panic("Migration check failed")
			}
			if need != nil && *need {
				panic("Database schema is not up to date, please run migrations, set `auto_migrate` to true or set `skip_migrate` to true in config")
			}
		}
	} else {
		logger.Warn("Skipping migration check as per configuration")
	}

	logger.Info("Database connected successfully with driver: %s", config.Driver)

	return client, nil
}

// MustNewClient 创建数据库客户端，失败时panic
func MustNewClient(config *configs.DatabaseConfig) *database.Client {
	client, err := NewClient(config)
	if err != nil {
		logger.Fatal("failed to create database client: %v", err)
	}
	return client
}

// InitInstance 初始化数据库客户端单例实例（只执行一次）
func InitInstance(config *configs.DatabaseConfig) *database.Client {
	once.Do(func() {
		supports := GetSupportedDrivers()
		logging.Info(
			`
===========Build Dependencies============
Supported database drivers:
	%v`,
			strings.Join(supports, ", "),
		)
		Client = MustNewClient(config)
	})
	return Client
}

// GetInstanceUnsafe 获取已初始化的数据库客户端单例实例（不进行任何检查）
// 注意：使用前请确保已经调用过 InitInstance 或 GetInstance
func GetInstanceUnsafe() *database.Client {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// CloseInstance 关闭数据库连接并重置单例实例
func CloseInstance() error {
	mu.Lock()
	defer mu.Unlock()

	if Client != nil {
		err := Client.Close()
		Client = nil
		// 重置 once，允许重新初始化
		once = sync.Once{}
		logger.Info("Database connection closed and instance reset")
		return err
	}
	return nil
}

// IsInstanceInitialized 检查单例实例是否已经初始化
func IsInstanceInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return Client != nil
}

// IsAlive 检查数据库连接是否仍然有效
func IsAlive() bool {
	if Client == nil {
		return false
	}
	ctx := context.Background()
	tx, err := Client.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Database connection is not alive: %v", err)
		return false
	}
	tx.Rollback() // 回滚事务以释放连接
	logger.Info("Database connection is alive")
	return true
}

// GetSupportedDrivers 获取支持的数据库驱动列表
func GetSupportedDrivers() []string {
	return drivers.GetSupportedDrivers()
}

// ListLoadedDrivers 列出已加载的数据库驱动
func ListLoadedDrivers() string {
	return drivers.ListDrivers()
}

// ValidateDriver 验证给定的驱动是否被支持
func ValidateDriver(driverName string) error {
	if !drivers.IsDriverSupported(driverName) {
		supportedDrivers := drivers.GetSupportedDrivers()
		return fmt.Errorf("unsupported database driver: %s. Supported drivers: %v", driverName, supportedDrivers)
	}
	return nil
}

// checkMigrationNeeded 检查数据库是否需要迁移
// 如果需要迁移，将输出迁移SQL语句到文件
func checkMigrationNeeded(client *database.Client) (*bool, error) {
	ctx := context.Background()

	var needMigration bool
	// 使用WriteTo方法检查是否有待执行的迁移
	var buf bytes.Buffer
	if err := client.Schema.WriteTo(ctx, &buf); err != nil {
		return nil, fmt.Errorf("failed to check migration status: %w", err)
	}

	migrationSQL := buf.String()
	if migrationSQL != "" {
		needMigration = true

		// 创建migration目录
		migrationDir := "migration"
		if err := os.MkdirAll(migrationDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create migration directory: %w", err)
		}

		// 生成文件名，格式：migration_YYYYMMDD_HHMMSS.sql
		timestamp := time.Now().Format("20060102_150405")
		fileName := fmt.Sprintf("migration_%s.sql", timestamp)
		filePath := filepath.Join(migrationDir, fileName)

		// 写入SQL文件
		if err := os.WriteFile(filePath, []byte(migrationSQL), 0644); err != nil {
			return nil, fmt.Errorf("failed to write migration file: %w", err)
		}

		if logger != nil {
			logger.Error("Database migration needed. Migration SQL saved to: %s", filePath)
		}
	} else {
		needMigration = false
		// 如果没有迁移SQL，说明数据库是最新的
		if logger != nil {
			logger.Info("Database schema is up to date, no migration needed")
		}
	}

	return &needMigration, nil
}
