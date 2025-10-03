package messaging

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
