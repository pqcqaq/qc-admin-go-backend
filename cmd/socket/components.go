package main

import (
	"fmt"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	"go-backend/pkg/jwt"
	"go-backend/pkg/logging"
	"go-backend/pkg/messaging"
	"go-backend/pkg/websocket"
	"sync"

	"github.com/redis/go-redis/v9"
)

// InitResults 保存初始化结果
type InitResults struct {
	redisClient *redis.Client
	errors      map[string]error
}

// setupLogging 设置日志系统
func setupLogging(config *configs.AppConfig) {
	logging.SetLevel(logging.ParseLogLevel(config.Logging.Level))
	logging.SetPrefix(config.Logging.Prefix)

	// 设置各个包的logger
	caching.SetLogger(logging.WithName("Caching"))
	messaging.SetLogger(logging.WithName("Messaging"))
	websocket.SetLogger(logging.WithName("WebSocket"))
}

// initializeComponents 并行初始化各个组件
func initializeComponents(config *configs.AppConfig) (*InitResults, error) {
	results := &InitResults{
		errors: make(map[string]error),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并行初始化Redis连接
	wg.Add(1)
	go func() {
		defer wg.Done()
		redisClient := caching.InitInstance(&config.Redis)
		mu.Lock()
		if redisClient == nil {
			results.errors["redis"] = fmt.Errorf("failed to initialize Redis client")
		}
		results.redisClient = redisClient
		mu.Unlock()
		logging.Info("Redis client initialized successfully with address: %s", config.Redis.Addr)
	}()

	// 并行初始化JWT服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := jwt.InitializeService(&config.JWT); err != nil {
			mu.Lock()
			results.errors["jwt"] = err
			mu.Unlock()
			logging.Warn("Failed to initialize JWT service: %v", err)
		} else {
			logging.Info("JWT service initialized successfully")
		}
	}()

	wg.Wait()

	var importantCmps []string = append(make([]string, 0), "redis")

	// // 检查关键组件的初始化错误
	for _, cmp := range importantCmps {
		if results.errors[cmp] != nil {
			return nil, fmt.Errorf(cmp+"服务初始化失败: %w", results.errors[cmp])
		}
	}

	return results, nil
}
