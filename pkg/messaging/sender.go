package messaging

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	"go-backend/pkg/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// Publish 发布消息（使用 msgpack 序列化）
func Publish(ctx context.Context, task MessageStruct) (string, error) {
	err := prehandleTask(ctx, &task)

	if err != nil {
		return "", fmt.Errorf("预处理任务失败: %w", err)
	}

	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey

	// 使用 msgpack 序列化
	data, err := msgpack.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("msgpack 序列化失败: %w", err)
	}

	// 转为 base64 存储（Redis Stream 的值必须是字符串）
	encoded := base64.StdEncoding.EncodeToString(data)

	// 添加到 Stream
	result, err := caching.GetInstanceUnsafe().XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": encoded,
		},
	}).Result()

	if err != nil {
		return "", fmt.Errorf("发布到 Stream 失败: %w", err)
	}

	logger.Info("✓ 发布消息成功 - ID: %s, TaskID: %s", result, task.id)
	return result, nil
}

func prehandleTask(_ context.Context, task *MessageStruct) error {
	task.id = generateUniqueID()
	if task.Priority < 0 {
		task.Priority = 0
	}
	task.createdAt = time.Now()
	return nil
}

func generateUniqueID() string {
	return utils.UUIDString()
}
