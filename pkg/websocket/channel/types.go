package channel

import (
	"context"
	"sync"
)

type Logger interface {
	Debug(fmt string, args ...any)
	Info(fmt string, args ...any)
	Warn(fmt string, args ...any)
	Error(fmt string, args ...any)
	Fatal(fmt string, args ...any)
}

var logger Logger

func SetLogger(l Logger) {
	logger = l
}

type ChannelMsg struct {
	Data      any `json:"data"`
	timestamp int64

	channel *Channel `json:"-"`
}

type ChannelLifecycle int

const (
	Channel_Started ChannelLifecycle = iota
	Channel_Running
	Channel_Closed
)

type Channel struct {
	ID        string        // 这次对话的ID
	Topic     string        // 频道主题
	CreatorId uint64        // 创建者的用户ID
	SessionId string        // 创建者的会话ID
	ClientId  uint64        // 创建者的客户端ID
	history   []*ChannelMsg // 历史消息

	factory *ChannelFactory

	sMu    sync.Mutex
	status ChannelLifecycle
}

func (c *Channel) GetStatus() ChannelLifecycle {
	c.sMu.Lock()
	defer c.sMu.Unlock()
	return c.status
}

func (c *Channel) SetStatus(status ChannelLifecycle) {
	c.sMu.Lock()
	c.status = status
	c.sMu.Unlock()
}

type ToClientSender func(msg ChannelMsg) error
type ToServerSender func(msg ChannelMsg) error
type ChannelCloser func(channel *Channel) error

type ChannelFactoryOptions struct {
	ToClient ToClientSender
	ToServer ToServerSender
	OnClose  func(channel *Channel) error
}

type ChannelFactory struct {
	// 消息发送接口
	toClient ToClientSender
	toServer ToServerSender
	// 关闭回调
	onClose ChannelCloser

	sendCtx context.Context
}
