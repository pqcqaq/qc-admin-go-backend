package messaging

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	redisVersion     *RedisVersion
	redisVersionOnce sync.Once
)

// RedisVersion Redis版本信息
type RedisVersion struct {
	Major int
	Minor int
	Patch int
}

// IsAtLeast 检查是否至少是指定版本
func (v *RedisVersion) IsAtLeast(major, minor, patch int) bool {
	if v.Major > major {
		return true
	}
	if v.Major == major && v.Minor > minor {
		return true
	}
	if v.Major == major && v.Minor == minor && v.Patch >= patch {
		return true
	}
	return false
}

// GetRedisVersion 获取Redis版本（单例模式）
func GetRedisVersion(ctx context.Context) (*RedisVersion, error) {
	var err error
	redisVersionOnce.Do(func() {
		client := caching.GetInstanceUnsafe()
		info, infoErr := client.Info(ctx, "server").Result()
		if infoErr != nil {
			err = fmt.Errorf("获取Redis信息失败: %w", infoErr)
			return
		}

		// 解析版本号
		lines := strings.Split(info, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "redis_version:") {
				versionStr := strings.TrimPrefix(line, "redis_version:")
				versionStr = strings.TrimSpace(versionStr)

				parts := strings.Split(versionStr, ".")
				if len(parts) >= 3 {
					major, _ := strconv.Atoi(parts[0])
					minor, _ := strconv.Atoi(parts[1])
					patch, _ := strconv.Atoi(parts[2])

					redisVersion = &RedisVersion{
						Major: major,
						Minor: minor,
						Patch: patch,
					}
					logger.Info("检测到Redis版本: %d.%d.%d", major, minor, patch)
					return
				}
			}
		}
		err = fmt.Errorf("无法解析Redis版本信息")
	})

	if err != nil {
		return nil, err
	}
	return redisVersion, nil
}

type MessageCunsumer struct {
	consumerName string
}

func NewMessageConsumer(consumerName string) *MessageCunsumer {
	return &MessageCunsumer{consumerName: consumerName}
}

// CreateGroup 创建消费者组
func CreateGroup(ctx context.Context) error {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	// 尝试获取Redis版本
	_, err := GetRedisVersion(ctx)
	if err != nil {
		logger.Warn("无法获取Redis版本，将使用兼容模式: %v", err)
	}

	// 尝试创建消费者组，如果已存在会返回错误，可以忽略
	err = client.XGroupCreateMkStream(ctx, streamKey, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}
	return nil
}

// 消息分发函数
func messageDispatcher(message MessageStruct) error {
	handlers := GetHandlers(message.Type)
	if len(handlers) == 0 {
		logger.Warn("没有注册处理器来处理消息类型: %s", message.Type)
		return fmt.Errorf("没有注册处理器来处理消息类型: %s", message.Type)
	}
	anySuccess := false
	for _, handler := range handlers {
		if err := handler(message); err != nil {
			logger.Error("处理消息 %s 失败: %v", message.id, err)
		} else {
			anySuccess = true
		}
	}
	if !anySuccess {
		return fmt.Errorf("所有处理器均未成功处理消息 %s", message.id)
	}
	return nil
}

// Consume 开始消费消息
func (c *MessageCunsumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 读取新消息
			if err := c.readNewMessages(ctx, messageDispatcher); err != nil {
				logger.Error("[%s] 读取新消息错误: %v", c.consumerName, err)
				time.Sleep(time.Second)
			}

			// 处理待处理消息（pending messages）
			if err := c.processPendingMessages(ctx, messageDispatcher); err != nil {
				logger.Error("[%s] 处理待处理消息错误: %v", c.consumerName, err)
			}
		}
	}
}

// readNewMessages 读取新消息
func (c *MessageCunsumer) readNewMessages(ctx context.Context, handler func(MessageStruct) error) error {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	readTimeout := configs.GetConfig().Server.Components.Messaging.ReadTimeout
	readCount := configs.GetConfig().Server.Components.Messaging.ReadCount
	client := caching.GetInstanceUnsafe()

	streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: c.consumerName,
		Streams:  []string{streamKey, ">"}, // ">" 表示只读取新消息
		Count:    readCount,
		Block:    time.Duration(readTimeout) * time.Millisecond,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil // 没有新消息
		}
		return err
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			c.processMessage(ctx, message, handler)
		}
	}

	return nil
}

// processPendingMessages 处理待处理的消息（超时未 ACK 的消息）
func (c *MessageCunsumer) processPendingMessages(ctx context.Context, handler func(MessageStruct) error) error {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	maxRetries := configs.GetConfig().Server.Components.Messaging.MaxRetries
	idleTimeout := configs.GetConfig().Server.Components.Messaging.IdleTimeout
	readCount := configs.GetConfig().Server.Components.Messaging.ReadCount
	client := caching.GetInstanceUnsafe()

	// 获取Redis版本
	version, err := GetRedisVersion(ctx)
	useIdleParam := false
	if err == nil && version.IsAtLeast(6, 2, 0) {
		useIdleParam = true
	}

	var pending []redis.XPendingExt

	if useIdleParam {
		// Redis 6.2.0+ 支持 Idle 参数
		pending, err = client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: streamKey,
			Group:  groupName,
			Start:  "-",
			End:    "+",
			Count:  readCount,
			Idle:   time.Duration(idleTimeout) * time.Millisecond,
		}).Result()
	} else {
		// 旧版本Redis，不使用 Idle 参数
		pending, err = client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: streamKey,
			Group:  groupName,
			Start:  "-",
			End:    "+",
			Count:  readCount,
		}).Result()
	}

	if err != nil {
		return err
	}

	for _, p := range pending {
		// 如果没有使用Idle参数，需要手动过滤
		if !useIdleParam {
			idleDuration := time.Duration(p.Idle.Milliseconds()) * time.Millisecond
			if idleDuration < time.Duration(idleTimeout)*time.Millisecond {
				continue // 跳过未超时的消息
			}
		}

		// 检查重试次数
		if p.RetryCount >= int64(maxRetries) {
			logger.Error("[%s] 消息 %s 重试次数已达上限，移至死信队列", c.consumerName, p.ID)
			c.moveToDeadLetter(ctx, p.ID)
			continue
		}

		// 认领消息
		messages, err := client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   streamKey,
			Group:    groupName,
			Consumer: c.consumerName,
			MinIdle:  time.Duration(idleTimeout) * time.Millisecond,
			Messages: []string{p.ID},
		}).Result()

		if err != nil {
			logger.Error("[%s] 认领消息失败: %v", c.consumerName, err)
			continue
		}

		for _, message := range messages {
			logger.Warn("[%s] 重试处理消息: %s (第 %d 次)", c.consumerName, message.ID, p.RetryCount+1)
			c.processMessage(ctx, message, handler)
		}
	}

	return nil
}

// processMessage 处理单条消息
func (c *MessageCunsumer) processMessage(ctx context.Context, message redis.XMessage, handler func(MessageStruct) error) {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	// 解码消息
	encoded, ok := message.Values["data"].(string)
	if !ok {
		logger.Error("[%s] 消息格式错误: %s", c.consumerName, message.ID)
		client.XAck(ctx, streamKey, groupName, message.ID)
		return
	}

	// base64 解码
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		logger.Error("[%s] base64 解码失败: %v", c.consumerName, err)
		client.XAck(ctx, streamKey, groupName, message.ID)
		return
	}

	// msgpack 反序列化
	var MessageStruct MessageStruct
	if err := msgpack.Unmarshal(data, &MessageStruct); err != nil {
		logger.Error("[%s] msgpack 反序列化失败: %v", c.consumerName, err)
		client.XAck(ctx, streamKey, groupName, message.ID)
		return
	}

	// 执行业务处理
	if err := handler(MessageStruct); err != nil {
		logger.Error("[%s] 处理消息失败 %s: %v", c.consumerName, message.ID, err)
		// 不 ACK，让消息进入 pending 状态，等待重试
		return
	}

	// 处理成功，ACK 消息
	if err := client.XAck(ctx, streamKey, groupName, message.ID).Err(); err != nil {
		logger.Error("[%s] ACK 失败: %v", c.consumerName, err)
		return
	}
}

// moveToDeadLetter 将消息移至死信队列
func (c *MessageCunsumer) moveToDeadLetter(ctx context.Context, messageID string) {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	deadLetterKey := streamKey + ":dead_letter"

	// 获取原始消息
	messages, err := client.XRange(ctx, streamKey, messageID, messageID).Result()
	if err != nil || len(messages) == 0 {
		logger.Error("无法获取消息 %s", messageID)
		return
	}

	// 添加到死信队列
	client.XAdd(ctx, &redis.XAddArgs{
		Stream: deadLetterKey,
		Values: messages[0].Values,
	})

	// ACK 原消息
	client.XAck(ctx, streamKey, groupName, messageID)
	logger.Warn("消息 %s 已移至死信队列", messageID)
}

// GetGroupInfo 获取消费者组信息
func GetGroupInfo(ctx context.Context) {
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	info, err := client.XInfoGroups(ctx, streamKey).Result()
	if err != nil {
		logger.Error("获取消费者组信息失败: %v", err)
		return
	}

	for _, group := range info {
		logger.Info("消费者组: %s, Pending: %d, LastDeliveredID: %s",
			group.Name, group.Pending, group.LastDeliveredID)
	}
}
