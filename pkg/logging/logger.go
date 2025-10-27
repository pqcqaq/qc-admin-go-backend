package logging

import (
	"fmt"
	"go-backend/pkg/configs"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"gopkg.in/natefinch/lumberjack.v2"
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

// LoggerInterface 定义日志接口
type LoggerInterface interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

// Logger 结构体定义
type Logger struct {
	level       LogLevel
	prefix      string
	component   string
	consoleOut  io.Writer
	fileWriters map[LogLevel]*lumberjack.Logger
	config      *configs.LoggingConfig
	mu          sync.Mutex
}

// 确保Logger实现了LoggerInterface接口
var _ LoggerInterface = (*Logger)(nil)

// 颜色定义
var (
	debugColor  = color.New(color.FgCyan)
	infoColor   = color.New(color.FgGreen)
	warnColor   = color.New(color.FgYellow)
	errorColor  = color.New(color.FgRed)
	fatalColor  = color.New(color.FgRed, color.Bold)
	timeColor   = color.New(color.FgBlue)
	prefixColor = color.New(color.FgMagenta)
)

// 全局logger实例
var defaultLogger *Logger

func PreHandle(config *configs.LoggingConfig) {
	config.File.Path = configs.ResolveConfigVariables(config.File.Path)
	config.File.Filenames.Debug = configs.ResolveConfigVariables(config.File.Filenames.Debug)
	config.File.Filenames.Info = configs.ResolveConfigVariables(config.File.Filenames.Info)
	config.File.Filenames.Warn = configs.ResolveConfigVariables(config.File.Filenames.Warn)
	config.File.Filenames.Error = configs.ResolveConfigVariables(config.File.Filenames.Error)
	config.File.Filenames.Fatal = configs.ResolveConfigVariables(config.File.Filenames.Fatal)
	config.File.Filenames.All = configs.ResolveConfigVariables(config.File.Filenames.All)
}

// NewLogger 创建一个新的 Logger 实例
func NewLogger(config *configs.LoggingConfig) *Logger {
	defaultLogger = &Logger{
		level:       ParseLogLevel(config.Level),
		prefix:      config.Prefix,
		component:   "GLOBAL",
		config:      config,
		fileWriters: make(map[LogLevel]*lumberjack.Logger),
	}

	if config.Console.Enabled {
		defaultLogger.consoleOut = os.Stdout
	}

	if config.File.Enabled {
		if err := os.MkdirAll(config.File.Path, os.ModePerm); err != nil {
			panic(fmt.Errorf("failed to create log directory: %w", err))
		}

		if config.File.SplitByLevel {
			// 按级别创建不同的 writer
			defaultLogger.fileWriters[DEBUG] = defaultLogger.createLumberjackLogger(config.File.Filenames.Debug)
			defaultLogger.fileWriters[INFO] = defaultLogger.createLumberjackLogger(config.File.Filenames.Info)
			defaultLogger.fileWriters[WARN] = defaultLogger.createLumberjackLogger(config.File.Filenames.Warn)
			defaultLogger.fileWriters[ERROR] = defaultLogger.createLumberjackLogger(config.File.Filenames.Error)
			defaultLogger.fileWriters[FATAL] = defaultLogger.createLumberjackLogger(config.File.Filenames.Fatal)
		} else {
			// 所有级别的日志写入同一个文件
			allWriter := defaultLogger.createLumberjackLogger(config.File.Filenames.All)
			defaultLogger.fileWriters[DEBUG] = allWriter
			defaultLogger.fileWriters[INFO] = allWriter
			defaultLogger.fileWriters[WARN] = allWriter
			defaultLogger.fileWriters[ERROR] = allWriter
			defaultLogger.fileWriters[FATAL] = allWriter
		}
	}

	return defaultLogger
}

// createLumberjackLogger 创建一个 lumberjack 实例
func (l *Logger) createLumberjackLogger(filename string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   filepath.Join(l.config.File.Path, filename),
		MaxSize:    l.config.File.MaxSize,
		MaxBackups: l.config.File.MaxBackups,
		MaxAge:     l.config.File.MaxAge,
		Compress:   l.config.File.Compress,
		LocalTime:  true,
	}
}

func WithName(name string) *Logger {
	return defaultLogger.WithComponent(name)
}

func (l *Logger) WithComponent(component string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 返回一个新的Logger实例，继承父级的配置，但有新的component
	return &Logger{
		level:       l.level,
		prefix:      l.prefix,
		component:   component,
		consoleOut:  l.consoleOut,
		fileWriters: l.fileWriters,
		config:      l.config,
	}
}

// 格式化时间戳
func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

// 通用日志输出函数
func (l *Logger) log(level LogLevel, levelName string, colorFunc *color.Color, format string, args ...any) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	timestamp := formatTime()

	// 控制台输出
	if l.consoleOut != nil {
		consoleTimestamp := timeColor.Sprintf("[%s]", timestamp)
		consolePrefix := prefixColor.Sprintf("[%s]", l.prefix)
		levelStr := colorFunc.Sprintf("[%s]", levelName)

		if l.component != "" {
			consolePrefix = fmt.Sprintf("%s [%s]", consolePrefix, l.component)
		}

		fmt.Fprintf(l.consoleOut, "%s %s %s %s\n", consoleTimestamp, consolePrefix, levelStr, message)
	}

	// 文件输出
	if l.config.File.Enabled {
		fileWriter, ok := l.fileWriters[level]
		if ok && fileWriter != nil {
			// 文件日志不带颜色
			filePrefix := fmt.Sprintf("[%s]", l.prefix)
			if l.component != "" {
				filePrefix = fmt.Sprintf("%s [%s]", filePrefix, l.component)
			}
			fileLevelStr := fmt.Sprintf("[%s]", levelName)

			logLine := fmt.Sprintf("[%s] %s %s %s\n", timestamp, filePrefix, fileLevelStr, message)

			// lumberjack 会自动处理按日期/大小轮转
			_, _ = fileWriter.Write([]byte(logLine))
		}
	}
}

// Debug 输出调试信息
func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, "DEBUG", debugColor, format, args...)
}

// Info 输出信息
func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, "INFO", infoColor, format, args...)
}

// Warn 输出警告
func (l *Logger) Warn(format string, args ...any) {
	l.log(WARN, "WARN", warnColor, format, args...)
}

// Error 输出错误
func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, "ERROR", errorColor, format, args...)
}

// Fatal 输出致命错误并退出程序
func (l *Logger) Fatal(format string, args ...any) {
	l.log(FATAL, "FATAL", fatalColor, format, args...)
	os.Exit(1)
}

func GetInstance() LoggerInterface {
	return defaultLogger
}

// 全局便捷函数
func SetLevel(level LogLevel) {
	defaultLogger.mu.Lock()
	defaultLogger.level = level
	defaultLogger.mu.Unlock()
}

func SetPrefix(prefix string) {
	defaultLogger.mu.Lock()
	defaultLogger.prefix = prefix
	defaultLogger.mu.Unlock()
}

func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...any) {
	defaultLogger.Fatal(format, args...)
}

// ParseLogLevel 从字符串解析日志级别
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
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
