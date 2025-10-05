package channelhandler

type IsolateChannelLifecycle int

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

const (
	Channel_Started IsolateChannelLifecycle = iota
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

	history []IsolateChannelMsg
	status  IsolateChannelLifecycle
}
