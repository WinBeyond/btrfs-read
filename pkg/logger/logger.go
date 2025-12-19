package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel = LevelInfo
	debugLogger  *log.Logger
	infoLogger   *log.Logger
	warnLogger   *log.Logger
	errorLogger  *log.Logger
)

func init() {
	debugLogger = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(os.Stderr, "[INFO]  ", log.Ldate|log.Ltime)
	warnLogger = log.New(os.Stderr, "[WARN]  ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	currentLevel = level
}

// SetLevelFromString 从字符串设置日志级别
func SetLevelFromString(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = LevelDebug
	case "info":
		currentLevel = LevelInfo
	case "warn", "warning":
		currentLevel = LevelWarn
	case "error":
		currentLevel = LevelError
	default:
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", level)
	}
	return nil
}

// SetOutput 设置日志输出
func SetOutput(w io.Writer) {
	debugLogger.SetOutput(w)
	infoLogger.SetOutput(w)
	warnLogger.SetOutput(w)
	errorLogger.SetOutput(w)
}

// Debug 输出调试日志
func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info 输出信息日志
func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warn 输出警告日志
func Warn(format string, v ...interface{}) {
	if currentLevel <= LevelWarn {
		warnLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error 输出错误日志
func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// GetLevel 获取当前日志级别
func GetLevel() LogLevel {
	return currentLevel
}
