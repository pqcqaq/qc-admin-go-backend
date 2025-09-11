package funcs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-backend/pkg/caching"

	"github.com/redis/go-redis/v9"
)

// String operations 字符串操作

// Set 设置字符串值
func Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return caching.Client.Set(ctx, key, value, expiration).Err()
}

// Get 获取字符串值
func Get(ctx context.Context, key string) (string, error) {
	return caching.Client.Get(ctx, key).Result()
}

// GetWithDefault 获取字符串值，如果不存在返回默认值
func GetWithDefault(ctx context.Context, key string, defaultValue string) string {
	val, err := caching.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return defaultValue
	}
	if err != nil {
		return defaultValue
	}
	return val
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) bool {
	count, err := caching.Client.Exists(ctx, key).Result()
	return err == nil && count > 0
}

// Delete 删除键
func Delete(ctx context.Context, keys ...string) error {
	return caching.Client.Del(ctx, keys...).Err()
}

// SetExpire 设置键的过期时间
func SetExpire(ctx context.Context, key string, expiration time.Duration) error {
	return caching.Client.Expire(ctx, key, expiration).Err()
}

// GetTTL 获取键的剩余生存时间
func GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return caching.Client.TTL(ctx, key).Result()
}

// JSON operations JSON操作

// SetJSON 设置JSON值
func SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return caching.Client.Set(ctx, key, jsonData, expiration).Err()
}

// GetJSON 获取JSON值并反序列化
func GetJSON(ctx context.Context, key string, dest any) error {
	val, err := caching.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Hash operations 哈希操作

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, field string, value any) error {
	return caching.Client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希字段值
func HGet(ctx context.Context, key string, field string) (string, error) {
	return caching.Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return caching.Client.HGetAll(ctx, key).Result()
}

// HExists 检查哈希字段是否存在
func HExists(ctx context.Context, key string, field string) bool {
	exists, err := caching.Client.HExists(ctx, key, field).Result()
	return err == nil && exists
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return caching.Client.HDel(ctx, key, fields...).Err()
}

// HLen 获取哈希字段数量
func HLen(ctx context.Context, key string) (int64, error) {
	return caching.Client.HLen(ctx, key).Result()
}

// List operations 列表操作

// LPush 从左侧推入列表
func LPush(ctx context.Context, key string, values ...any) error {
	return caching.Client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func RPush(ctx context.Context, key string, values ...any) error {
	return caching.Client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func LPop(ctx context.Context, key string) (string, error) {
	return caching.Client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func RPop(ctx context.Context, key string) (string, error) {
	return caching.Client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func LLen(ctx context.Context, key string) (int64, error) {
	return caching.Client.LLen(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return caching.Client.LRange(ctx, key, start, stop).Result()
}

// Set operations 集合操作

// SAdd 添加元素到集合
func SAdd(ctx context.Context, key string, members ...any) error {
	return caching.Client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func SMembers(ctx context.Context, key string) ([]string, error) {
	return caching.Client.SMembers(ctx, key).Result()
}

// SIsMember 检查元素是否在集合中
func SIsMember(ctx context.Context, key string, member any) bool {
	isMember, err := caching.Client.SIsMember(ctx, key, member).Result()
	return err == nil && isMember
}

// SRem 从集合中移除元素
func SRem(ctx context.Context, key string, members ...any) error {
	return caching.Client.SRem(ctx, key, members...).Err()
}

// SCard 获取集合成员数量
func SCard(ctx context.Context, key string) (int64, error) {
	return caching.Client.SCard(ctx, key).Result()
}

// Sorted Set operations 有序集合操作

// ZAdd 添加元素到有序集合
func ZAdd(ctx context.Context, key string, score float64, member any) error {
	return caching.Client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZRange 按索引范围获取有序集合元素
func ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return caching.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数范围获取有序集合元素
func ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return caching.Client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// ZRem 从有序集合中移除元素
func ZRem(ctx context.Context, key string, members ...any) error {
	return caching.Client.ZRem(ctx, key, members...).Err()
}

// ZCard 获取有序集合成员数量
func ZCard(ctx context.Context, key string) (int64, error) {
	return caching.Client.ZCard(ctx, key).Result()
}

// ZScore 获取有序集合成员的分数
func ZScore(ctx context.Context, key string, member string) (float64, error) {
	return caching.Client.ZScore(ctx, key, member).Result()
}

// Advanced operations 高级操作

// Pipeline 创建管道
func Pipeline() redis.Pipeliner {
	return caching.Client.Pipeline()
}

// Transaction 执行事务
func Transaction(ctx context.Context, fn func(tx *redis.Tx) error, keys ...string) error {
	return caching.Client.Watch(ctx, fn, keys...)
}

// Subscribe 订阅频道
func Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return caching.Client.Subscribe(ctx, channels...)
}

// Publish 发布消息到频道
func Publish(ctx context.Context, channel string, message any) error {
	return caching.Client.Publish(ctx, channel, message).Err()
}

// SetNX 只有键不存在时才设置（分布式锁）
func SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return caching.Client.SetNX(ctx, key, value, expiration).Result()
}

// Incr 递增
func Incr(ctx context.Context, key string) (int64, error) {
	return caching.Client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return caching.Client.IncrBy(ctx, key, value).Result()
}

// Decr 递减
func Decr(ctx context.Context, key string) (int64, error) {
	return caching.Client.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减
func DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return caching.Client.DecrBy(ctx, key, value).Result()
}

// FlushDB 清空当前数据库
func FlushDB(ctx context.Context) error {
	return caching.Client.FlushDB(ctx).Err()
}

// Keys 获取匹配模式的所有键
func Keys(ctx context.Context, pattern string) ([]string, error) {
	return caching.Client.Keys(ctx, pattern).Result()
}

// Scan 扫描键（推荐使用，避免阻塞）
func Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return caching.Client.Scan(ctx, cursor, match, count).Result()
}
