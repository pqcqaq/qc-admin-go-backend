package messaging

import (
	"context"
	"fmt"
	"go-backend/pkg/caching"
	"go-backend/pkg/configs"
	"go-backend/pkg/utils"
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
	mType        []MessageType
	consumerName string
}

func NewMessageConsumer(consumerName string, mType ...MessageType) *MessageCunsumer {
	// 如果为空则不允许
	if len(mType) == 0 {
		panic("MessageConsumer must have at least one MessageType")
	}

	return &MessageCunsumer{
		mType:        mType,
		consumerName: consumerName,
	}
}

// CreateGroup 创建消费者组
func (mc *MessageCunsumer) CreateGroup(ctx context.Context) error {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	// 尝试获取Redis版本
	_, err := GetRedisVersion(ctx)
	if err != nil {
		logger.Warn("无法获取Redis版本，将使用兼容模式: %v", err)
	}

	// 尝试创建消费者组，如果已存在会返回错误，可以忽略
	for _, iType := range mc.mType {
		err = client.XGroupCreateMkStream(ctx, fmt.Sprintf("%s:%s", streamKey, iType), groupName, "0").Err()
	}
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
		return nil
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

	streamKeys := make([]string, 0, len(c.mType)*2)
	for _, mType := range c.mType {
		streamKeys = append(streamKeys, fmt.Sprintf("%s:%s", streamKey, mType))
	}
	// 每个流后面跟一个 ">" 表示从该流读取新消息
	for range c.mType {
		streamKeys = append(streamKeys, ">")
	}

	// 读取新消息
	streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: c.consumerName,
		Streams:  streamKeys, // ">" 表示只读取新消息
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
		// 从stream名称中提取消息类型
		streamParts := strings.Split(stream.Stream, ":")
		if len(streamParts) < 2 {
			logger.Error("[%s] 无效的stream名称: %s", c.consumerName, stream.Stream)
			continue
		}
		messageType := streamParts[len(streamParts)-1]

		for _, message := range stream.Messages {
			c.processMessage(ctx, message, messageType, handler)
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

	for _, iType := range c.mType {
		var pending []redis.XPendingExt

		if useIdleParam {
			// Redis 6.2.0+ 支持 Idle 参数
			newPending, err := client.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: fmt.Sprintf("%s:%s", streamKey, iType),
				Group:  groupName,
				Start:  "-",
				End:    "+",
				Count:  readCount,
				Idle:   time.Duration(idleTimeout) * time.Millisecond,
			}).Result()
			if err != nil {
				return err
			}
			pending = newPending
		} else {
			// 旧版本Redis，不使用 Idle 参数
			newPending, err := client.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: fmt.Sprintf("%s:%s", streamKey, iType),
				Group:  groupName,
				Start:  "-",
				End:    "+",
				Count:  readCount,
			}).Result()
			if err != nil {
				return err
			}
			pending = newPending
		}

		// 处理当前类型的pending消息
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
				c.moveToDeadLetter(ctx, p.ID, string(iType))
				continue
			}

			// 认领消息
			messages, err := client.XClaim(ctx, &redis.XClaimArgs{
				Stream:   fmt.Sprintf("%s:%s", streamKey, iType),
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
				c.processMessage(ctx, message, string(iType), handler)
			}
		}
	}

	return nil
}

// processMessage 处理单条消息
func (c *MessageCunsumer) processMessage(ctx context.Context, message redis.XMessage, messageType string, handler func(MessageStruct) error) {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	// 解码消息
	data, ok := message.Values["data"].(string)
	if !ok {
		logger.Error("[%s] 消息格式错误: %s", c.consumerName, message.ID)
		client.XAck(ctx, fmt.Sprintf("%s:%s", streamKey, messageType), groupName, message.ID)
		return
	}

	// msgpack 反序列化
	var messageStruct MessageStruct
	if err := msgpack.Unmarshal(utils.StringToByte(data), &messageStruct); err != nil {
		logger.Error("[%s] msgpack 反序列化失败: %v", c.consumerName, err)
		client.XAck(ctx, fmt.Sprintf("%s:%s", streamKey, messageType), groupName, message.ID)
		return
	}

	// 执行业务处理
	if err := handler(messageStruct); err != nil {
		logger.Error("[%s] 处理消息失败 %s: %v", c.consumerName, message.ID, err)
		// 不 ACK，让消息进入 pending 状态，等待重试
		return
	}

	// 处理成功，ACK 消息
	if err := client.XAck(ctx, fmt.Sprintf("%s:%s", streamKey, messageType), groupName, message.ID).Err(); err != nil {
		logger.Error("[%s] ACK 失败: %v", c.consumerName, err)
		return
	}
}

// moveToDeadLetter 将消息移至死信队列
func (c *MessageCunsumer) moveToDeadLetter(ctx context.Context, messageID string, messageType string) {
	groupName := configs.GetConfig().Server.Components.Messaging.GroupName
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	deadLetterKey := fmt.Sprintf("%s:%s:dead_letter", streamKey, messageType)

	// 获取原始消息
	messages, err := client.XRange(ctx, fmt.Sprintf("%s:%s", streamKey, messageType), messageID, messageID).Result()
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
	client.XAck(ctx, fmt.Sprintf("%s:%s", streamKey, messageType), groupName, messageID)
	logger.Warn("消息 %s 已移至死信队列", messageID)
}

// GetGroupInfo 获取消费者组信息
func (c *MessageCunsumer) GetGroupInfo(ctx context.Context) {
	streamKey := configs.GetConfig().Server.Components.Messaging.StreamKey
	client := caching.GetInstanceUnsafe()

	for _, mType := range c.mType {
		info, err := client.XInfoGroups(ctx, fmt.Sprintf("%s:%s", streamKey, mType)).Result()
		if err != nil {
			logger.Error("获取消费者组信息失败 (类型: %s): %v", mType, err)
			continue
		}

		for _, group := range info {
			logger.Info("消费者组 (类型: %s): %s, Pending: %d, LastDeliveredID: %s",
				mType, group.Name, group.Pending, group.LastDeliveredID)
		}
	}
}
