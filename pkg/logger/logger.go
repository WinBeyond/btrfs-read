package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel represents a logging level.
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

// SetLevel sets the log level.
func SetLevel(level LogLevel) {
	currentLevel = level
}

// SetLevelFromString sets the log level from a string.
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

// SetOutput sets the log output.
func SetOutput(w io.Writer) {
	debugLogger.SetOutput(w)
	infoLogger.SetOutput(w)
	warnLogger.SetOutput(w)
	errorLogger.SetOutput(w)
}

// Debug outputs a debug log.
func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Info outputs an info log.
func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		infoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warn outputs a warning log.
func Warn(format string, v ...interface{}) {
	if currentLevel <= LevelWarn {
		warnLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Error outputs an error log.
func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		errorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// GetLevel returns the current log level.
func GetLevel() LogLevel {
	return currentLevel
}
