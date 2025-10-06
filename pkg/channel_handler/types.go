package channelhandler

import (
	"sync"
)

type IsolateChannelLifecycle int

type Logger interface {
	Debug(fmt string, args ...any)
	Info(fmt string, args ...any)
	Warn(fmt string, args ...any)
	Error(fmt string, args ...any)
	Fatal(fmt string, args ...any)
}

const (
	Channel_Ready IsolateChannelLifecycle = iota
	Channel_Running
	Channel_Closed
)

type IsolateChannelMsg struct {
	Data      any `json:"data"`
	timestamp int64

	channel *IsolateChannel `json:"-"`
}

type IsolateChannel struct {
	ID        string
	Topic     string
	CreatorId uint64
	SessionId string
	ClientId  uint64

	history []IsolateChannelMsg
	status  IsolateChannelLifecycle

	sMu      sync.Mutex
	readChan chan (*IsolateChannelMsg)

	closeChan chan struct{}

	factory *ChannelHandler
}

type ChannelReceiver func(channel *IsolateChannel) error
type ChannelSender func(channelId string, msg any) error
type ChannelCloser func(channelId string) error
type ChannelError func(channelId string, err error) error

type CreateChannelHandlerOptions struct {
	Topic              string
	NewChannelReceived ChannelReceiver
	SendMessage        ChannelSender
	CloseChannel       ChannelCloser
	ErrSender          ChannelError
	Logger             Logger
}

type ChannelHandler struct {
	logger Logger

	topic       string
	onReceived  ChannelReceiver
	send        ChannelSender
	close       ChannelCloser
	errorSender ChannelError

	// 记录当前 所有活跃的频道
	channels map[string]*IsolateChannel
	cMu      sync.Mutex
}
