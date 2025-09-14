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

	"entgo.io/ent/dialect/sql"
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

// 连接检查相关变量
var (
	connectionCheckTicker *time.Ticker
	connectionCheckStop   chan bool
	connectionCheckMu     sync.Mutex
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

	drv, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db := drv.DB()
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	var client *database.Client
	if config.Debug {
		logger.Info("Database debug mode enabled")
		client = database.NewClient(database.Driver(drv), database.Debug())
	} else {
		client = database.NewClient(database.Driver(drv))
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

	// 启动连接检查协程
	if config.ConnectionCheckInterval > 0 {
		startConnectionCheck(client, config.ConnectionCheckInterval)
	}

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

	// 停止连接检查
	stopConnectionCheck()

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
	res, err := Client.ExecContext(ctx, "SELECT 1")
	if err != nil {
		logger.Error("Database connection is dead: %v", err)
		return false
	}
	if res == nil {
		return false
	}
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

// startConnectionCheck 启动数据库连接检查协程
func startConnectionCheck(client *database.Client, interval time.Duration) {
	connectionCheckMu.Lock()
	defer connectionCheckMu.Unlock()

	// 如果已经有检查在运行，先停止它
	if connectionCheckTicker != nil {
		stopConnectionCheckUnsafe()
	}

	connectionCheckTicker = time.NewTicker(interval)
	connectionCheckStop = make(chan bool, 1)

	go func() {
		if logger != nil {
			logger.Info("Database connection check started with interval: %v", interval)
		}

		for {
			select {
			case <-connectionCheckTicker.C:
				checkDatabaseConnection(client)
			case <-connectionCheckStop:
				connectionCheckTicker.Stop()
				if logger != nil {
					logger.Info("Database connection check stopped")
				}
				return
			}
		}
	}()
}

// stopConnectionCheck 停止数据库连接检查协程
func stopConnectionCheck() {
	connectionCheckMu.Lock()
	defer connectionCheckMu.Unlock()
	stopConnectionCheckUnsafe()
}

// stopConnectionCheckUnsafe 不加锁的停止连接检查（内部使用）
func stopConnectionCheckUnsafe() {
	if connectionCheckTicker != nil {
		connectionCheckTicker.Stop()
		connectionCheckTicker = nil
	}
	if connectionCheckStop != nil {
		select {
		case connectionCheckStop <- true:
		default:
		}
		close(connectionCheckStop)
		connectionCheckStop = nil
	}
}

// checkDatabaseConnection 检查数据库连接状态
func checkDatabaseConnection(client *database.Client) {
	if client == nil {
		if logger != nil {
			logger.Error("Database client is nil during connection check")
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var alive bool = true
	if Client == nil {
		return
	}
	res, err := Client.ExecContext(ctx, "SELECT 1")
	if err != nil {
		alive = false
	}
	if res == nil {
		alive = false
	}

	if !alive {
		if logger != nil {
			logger.Error("Database connection is dead, check your database server!")
		}
	} else {
		if logger != nil {
			logger.Info("Database connection is alive")
		}
	}
}

// ExportAllTablesGlobal 导出所有表数据到JSON文件（使用全局客户端实例）
func ExportAllTablesGlobal(outputDir string) (*ExportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	config := &ExportConfig{
		OutputDir:    outputDir,
		PrettyFormat: true,
		Context:      context.Background(),
	}

	if config.OutputDir == "" {
		config.OutputDir = "./exports"
	}

	return ExportAllTables(Client, config)
}

// ExportSpecificTablesGlobal 使用全局客户端导出指定表
func ExportSpecificTablesGlobal(entityNames []string, outputDir string) (*ExportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	return ExportSpecificTables(Client, entityNames, outputDir)
}

// ImportAllTablesGlobalFromDatabase 导入所有表数据到数据库（使用全局客户端实例）
func ImportAllTablesGlobalFromDatabase(inputDir string) (*ImportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	return ImportAllTablesGlobal(inputDir)
}

// ImportSpecificTablesGlobalFromDatabase 使用全局客户端导入指定表
func ImportSpecificTablesGlobalFromDatabase(entityNames []string, inputDir string) (*ImportResult, error) {
	if Client == nil {
		return nil, fmt.Errorf("database client is not initialized, call InitInstance first")
	}

	return ImportSpecificTablesGlobal(entityNames, inputDir)
}
