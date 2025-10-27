package configs

import (
	"fmt"
	"os"
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

// 因为这个包在初始化时可能会被调用，所以提供一个默认的空实现，防止nil引用
type noopLogger struct {
	name string
}

func (n *noopLogger) Debug(fmtStr string, args ...any) {
	fmt.Fprintln(os.Stdout, "DEBUG: "+n.name+fmt.Sprintf(fmtStr, args...))
}
func (n *noopLogger) Info(fmtStr string, args ...any) {
	fmt.Fprintln(os.Stdout, "INFO: "+n.name+fmt.Sprintf(fmtStr, args...))
}
func (n *noopLogger) Warn(fmtStr string, args ...any) {
	fmt.Fprintln(os.Stdout, "WARN: "+n.name+fmt.Sprintf(fmtStr, args...))
}
func (n *noopLogger) Error(fmtStr string, args ...any) {
	fmt.Fprintln(os.Stderr, "ERROR: "+n.name+fmt.Sprintf(fmtStr, args...))
}
func (n *noopLogger) Fatal(fmtStr string, args ...any) {
	fmt.Fprintln(os.Stderr, "FATAL: "+n.name+fmt.Sprintf(fmtStr, args...))
	panic("fatal error occurred: " + fmt.Sprintf(fmtStr, args...))
}

func init() {
	logger = &noopLogger{
		name: "GLOBAL",
	}
}
