package database

import (
	"context"
	"log"
	"sync"

	database "go-backend/database/ent"
	_ "go-backend/database/ent/runtime"

	"go-backend/pkg/configs"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
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
	client, err := database.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, err
	}

	// 运行自动迁移
	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		return nil, err
	}

	if logger != nil {
		logger.Info("Database connected successfully with driver: %s", config.Driver)
	} else {
		log.Printf("Database connected successfully with driver: %s", config.Driver)
	}

	return client, nil
}

// MustNewClient 创建数据库客户端，失败时panic
func MustNewClient(config *configs.DatabaseConfig) *database.Client {
	client, err := NewClient(config)
	if err != nil {
		if logger != nil {
			logger.Fatal("failed to create database client: %v", err)
		} else {
			log.Fatalf("failed to create database client: %v", err)
		}
	}
	return client
}

// InitInstance 初始化数据库客户端单例实例（只执行一次）
func InitInstance(config *configs.DatabaseConfig) *database.Client {
	once.Do(func() {
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
		if logger != nil {
			logger.Info("Database connection closed and instance reset")
		} else {
			log.Println("Database connection closed and instance reset")
		}
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
		if logger != nil {
			logger.Error("Database connection is not alive: %v", err)
		} else {
			log.Printf("Database connection is not alive: %v", err)
		}
		return false
	}
	tx.Rollback() // 回滚事务以释放连接
	if logger != nil {
		logger.Info("Database connection is alive")
	} else {
		log.Println("Database connection is alive")
	}
	return true
}
