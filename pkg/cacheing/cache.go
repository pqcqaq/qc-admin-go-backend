package cacheing

import (
	"context"
	"log"
	"sync"
	"time"

	"go-backend/pkg/configs"

	"github.com/redis/go-redis/v9"
)

// LoggerInterface 定义日志接口，避免循环依赖
type LoggerInterface interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// 单例相关变量
var (
	once sync.Once
	mu   sync.RWMutex
)

// 全局logger实例
var logger LoggerInterface

var Client *redis.Client

// SetLogger 设置logger实例
func SetLogger(l LoggerInterface) {
	logger = l
}

// NewClient 创建新的Redis客户端
func NewClient(config *configs.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            config.Addr,
		Password:        config.Password,
		DB:              config.DB,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		ReadTimeout:     time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(config.WriteTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(config.IdleTimeout) * time.Second,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Info("Redis connected successfully to: %s", config.Addr)
	} else {
		log.Printf("Redis connected successfully to: %s", config.Addr)
	}

	return rdb, nil
}

// MustNewClient 创建Redis客户端，失败时panic
func MustNewClient(config *configs.RedisConfig) *redis.Client {
	client, err := NewClient(config)
	if err != nil {
		if logger != nil {
			logger.Fatal("failed to create Redis client: %v", err)
		} else {
			log.Fatalf("failed to create Redis client: %v", err)
		}
	}
	return client
}

// InitInstance 初始化Redis客户端单例实例（只执行一次）
func InitInstance(config *configs.RedisConfig) *redis.Client {
	once.Do(func() {
		Client = MustNewClient(config)
	})
	return Client
}

// GetInstanceUnsafe 获取已初始化的Redis客户端单例实例（不进行任何检查）
// 注意：使用前请确保已经调用过 InitInstance 或 GetInstance
func GetInstanceUnsafe() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return Client
}

// CloseInstance 关闭Redis连接并重置单例实例
func CloseInstance() error {
	mu.Lock()
	defer mu.Unlock()

	if Client != nil {
		err := Client.Close()
		Client = nil
		// 重置 once，允许重新初始化
		once = sync.Once{}
		if logger != nil {
			logger.Info("Redis connection closed and instance reset")
		} else {
			log.Println("Redis connection closed and instance reset")
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

// IsAlive 检查Redis连接是否仍然有效
func IsAlive() bool {
	if Client == nil {
		return false
	}
	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		if logger != nil {
			logger.Error("Redis connection is not alive: %v", err)
		} else {
			log.Printf("Redis connection is not alive: %v", err)
		}
		return false
	}
	if logger != nil {
		logger.Info("Redis connection is alive")
	} else {
		log.Println("Redis connection is alive")
	}
	return true
}
