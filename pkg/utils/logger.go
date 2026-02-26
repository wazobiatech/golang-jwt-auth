package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLogLevel parses a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"msg"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// Logger provides structured logging capabilities
type Logger struct {
	service  string
	level    LogLevel
	output   *log.Logger
	mutex    sync.RWMutex
}

var (
	defaultLogger *Logger
	loggerOnce    sync.Once
)

// NewLogger creates a new logger instance
func NewLogger(service string) *Logger {
	config := GetConfig()
	level := ParseLogLevel(config.LogLevel)

	return &Logger{
		service: service,
		level:   level,
		output:  log.New(os.Stdout, "", 0), // No prefix, we handle formatting
	}
}

// GetDefaultLogger returns the singleton default logger
func GetDefaultLogger() *Logger {
	loggerOnce.Do(func() {
		defaultLogger = NewLogger("auth-middleware")
	})
	return defaultLogger
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.level
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.GetLevel()
}

// log writes a log entry with the specified level
func (l *Logger) log(level LogLevel, message string, meta map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Level:     level.String(),
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   l.service,
	}

	if meta != nil && len(meta) > 0 {
		entry.Meta = meta
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.output.Printf("[%s] %s: %s (JSON marshal error: %v)", 
			level.String(), l.service, message, err)
		return
	}

	l.output.Println(string(jsonBytes))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, meta map[string]interface{}) {
	l.log(LevelDebug, message, meta)
}

// Info logs an info message
func (l *Logger) Info(message string, meta map[string]interface{}) {
	l.log(LevelInfo, message, meta)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, meta map[string]interface{}) {
	l.log(LevelWarn, message, meta)
}

// Error logs an error message
func (l *Logger) Error(message string, meta map[string]interface{}) {
	l.log(LevelError, message, meta)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string, meta map[string]interface{}) {
	l.log(LevelFatal, message, meta)
	os.Exit(1)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...), nil)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...), nil)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...), nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...), nil)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format, args...), nil)
}

// WithMeta creates a new logger with default metadata
func (l *Logger) WithMeta(meta map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger:      l,
		defaultMeta: meta,
	}
}

// ContextLogger wraps a Logger with default metadata
type ContextLogger struct {
	logger      *Logger
	defaultMeta map[string]interface{}
}

// mergeMeta combines default metadata with provided metadata
func (cl *ContextLogger) mergeMeta(meta map[string]interface{}) map[string]interface{} {
	if cl.defaultMeta == nil {
		return meta
	}

	merged := make(map[string]interface{})
	
	// Copy default metadata
	for k, v := range cl.defaultMeta {
		merged[k] = v
	}
	
	// Override with provided metadata
	if meta != nil {
		for k, v := range meta {
			merged[k] = v
		}
	}
	
	return merged
}

// Debug logs a debug message with merged metadata
func (cl *ContextLogger) Debug(message string, meta map[string]interface{}) {
	cl.logger.Debug(message, cl.mergeMeta(meta))
}

// Info logs an info message with merged metadata
func (cl *ContextLogger) Info(message string, meta map[string]interface{}) {
	cl.logger.Info(message, cl.mergeMeta(meta))
}

// Warn logs a warning message with merged metadata
func (cl *ContextLogger) Warn(message string, meta map[string]interface{}) {
	cl.logger.Warn(message, cl.mergeMeta(meta))
}

// Error logs an error message with merged metadata
func (cl *ContextLogger) Error(message string, meta map[string]interface{}) {
	cl.logger.Error(message, cl.mergeMeta(meta))
}

// Fatal logs a fatal message with merged metadata and exits the program
func (cl *ContextLogger) Fatal(message string, meta map[string]interface{}) {
	cl.logger.Fatal(message, cl.mergeMeta(meta))
}

// Package-level convenience functions using the default logger

// Debug logs a debug message using the default logger
func Debug(message string, meta map[string]interface{}) {
	GetDefaultLogger().Debug(message, meta)
}

// Info logs an info message using the default logger
func Info(message string, meta map[string]interface{}) {
	GetDefaultLogger().Info(message, meta)
}

// Warn logs a warning message using the default logger
func Warn(message string, meta map[string]interface{}) {
	GetDefaultLogger().Warn(message, meta)
}

// Error logs an error message using the default logger
func Error(message string, meta map[string]interface{}) {
	GetDefaultLogger().Error(message, meta)
}

// Fatal logs a fatal message using the default logger and exits the program
func Fatal(message string, meta map[string]interface{}) {
	GetDefaultLogger().Fatal(message, meta)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}

// Infof logs a formatted info message using the default logger
func Infof(format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

// Fatalf logs a formatted fatal message using the default logger and exits the program
func Fatalf(format string, args ...interface{}) {
	GetDefaultLogger().Fatalf(format, args...)
}