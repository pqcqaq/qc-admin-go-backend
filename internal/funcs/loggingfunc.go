package funcs

import (
	"context"
	schemaLogging "go-backend/database/ent/logging"
	"go-backend/pkg/database"
	"go-backend/pkg/logging"
)

// CreateAsyncLoggingFunc 异步创建日志记录
func CreateAsyncLoggingFunc(level string, logType string, message string, method, path, ip, query string, code int, user_agent string, data map[string]interface{}, stack string) {

	ctx := context.Background()
	if database.Client == nil {
		logging.Error("Database client is not initialized")
		return // 数据库客户端未初始化
	}
	go func() {
		builderLevel := schemaLogging.Level(level)
		builderLogType := schemaLogging.Type(logType)

		builder := database.Client.Logging.Create()
		builder.SetLevel(builderLevel)
		builder.SetType(builderLogType)
		builder.SetMessage(message)
		if method != "" {
			builder.SetMethod(method)
		}
		if path != "" {
			builder.SetPath(path)
		}
		if ip != "" {
			builder.SetIP(ip)
		}
		if query != "" {
			builder.SetQuery(query)
		}
		if code > 0 {
			builder.SetCode(code)
		}
		if user_agent != "" {
			builder.SetUserAgent(user_agent)
		}
		if data != nil {
			builder.SetData(data)
		}
		if stack != "" {
			builder.SetStack(stack)
		}

		if _, err := builder.Save(ctx); err != nil {
			logging.Error("Failed to create async logging: %v", err)
		}

		logging.Debug("Async logging created successfully")
		ctx.Done() // 确保上下文被取消
	}()
}
