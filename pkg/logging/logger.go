package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// LoggerInterface 定义日志接口，供其他包使用
type LoggerInterface interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
}

// Logger 结构体定义
type Logger struct {
	level  LogLevel
	prefix string
}

// 确保Logger实现了LoggerInterface接口
var _ LoggerInterface = (*Logger)(nil)

// 颜色定义
var (
	debugColor = color.New(color.FgCyan)
	infoColor  = color.New(color.FgGreen)
	warnColor  = color.New(color.FgYellow)
	errorColor = color.New(color.FgRed)
	fatalColor = color.New(color.FgRed, color.Bold)
	timeColor  = color.New(color.FgBlue)
	prefixColor = color.New(color.FgMagenta)
)

// 全局logger实例
var defaultLogger *Logger

// 初始化默认logger
func init() {
	defaultLogger = &Logger{
		level:  INFO,
		prefix: "APP",
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetPrefix 设置日志前缀
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// 格式化时间戳
func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 通用日志输出函数
func (l *Logger) log(level LogLevel, levelName string, colorFunc *color.Color, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := timeColor.Sprintf("[%s]", formatTime())
	prefix := prefixColor.Sprintf("[%s]", l.prefix)
	levelStr := colorFunc.Sprintf("[%s]", levelName)
	
	message := fmt.Sprintf(format, args...)
	
	fmt.Printf("%s %s %s %s\n", timestamp, prefix, levelStr, message)
}

// Debug 输出调试信息
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, "DEBUG", debugColor, format, args...)
}

// Info 输出信息
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, "INFO", infoColor, format, args...)
}

// Warn 输出警告
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, "WARN", warnColor, format, args...)
}

// Error 输出错误
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, "ERROR", errorColor, format, args...)
}

// Fatal 输出致命错误并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, "FATAL", fatalColor, format, args...)
	os.Exit(1)
}

func GetInstance() LoggerInterface {
	return defaultLogger
}

// 全局便捷函数
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func SetPrefix(prefix string) {
	defaultLogger.SetPrefix(prefix)
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// ParseLogLevel 从字符串解析日志级别
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug", "DEBUG":
		return DEBUG
	case "info", "INFO":
		return INFO
	case "warn", "WARN", "warning", "WARNING":
		return WARN
	case "error", "ERROR":
		return ERROR
	case "fatal", "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// LogLevelString 将日志级别转换为字符串
func LogLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "INFO"
	}
}
