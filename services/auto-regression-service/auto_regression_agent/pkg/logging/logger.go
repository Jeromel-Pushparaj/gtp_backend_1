// Package logging provides structured logging for OpenTest
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// ParseLevel parses a log level string
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	SetLevel(level Level)
}

// LogConfig configures the logger
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, stderr, or file path
}

// StandardLogger implements Logger with standard output
type StandardLogger struct {
	level  Level
	format string
	output io.Writer
	fields []Field
	mu     sync.Mutex
}

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg LogConfig) Logger {
	level := ParseLevel(cfg.Level)
	format := cfg.Format
	if format == "" {
		format = "json"
	}

	var output io.Writer
	switch cfg.Output {
	case "", "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			output = os.Stdout
		} else {
			output = f
		}
	}

	return &StandardLogger{
		level:  level,
		format: format,
		output: output,
		fields: nil,
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger() Logger {
	return NewLogger(LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
}

// SetLevel sets the log level
func (l *StandardLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// With returns a new logger with the given fields
func (l *StandardLogger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &StandardLogger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: newFields,
	}
}

// Debug logs a debug message
func (l *StandardLogger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info logs an info message
func (l *StandardLogger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (l *StandardLogger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error logs an error message
func (l *StandardLogger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// log writes a log entry
func (l *StandardLogger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Combine default fields with provided fields
	allFields := make([]Field, len(l.fields)+len(fields))
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)

	if l.format == "json" {
		l.logJSON(level, msg, allFields)
	} else {
		l.logText(level, msg, allFields)
	}
}

// logJSON writes a JSON log entry
func (l *StandardLogger) logJSON(level Level, msg string, fields []Field) {
	entry := map[string]interface{}{
		"time":    time.Now().Format(time.RFC3339),
		"level":   levelNames[level],
		"message": msg,
	}

	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.output.Write(data)
	l.output.Write([]byte("\n"))
}

// logText writes a text log entry
func (l *StandardLogger) logText(level Level, msg string, fields []Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s [%s] %s", timestamp, levelNames[level], msg)

	for _, f := range fields {
		line += fmt.Sprintf(" %s=%v", f.Key, f.Value)
	}

	l.output.Write([]byte(line + "\n"))
}

// Global logger instance
var globalLogger Logger = NewDefaultLogger()

// SetGlobalLogger sets the global logger
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() Logger {
	return globalLogger
}

// Global convenience functions
func Debug(msg string, fields ...Field) { globalLogger.Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { globalLogger.Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { globalLogger.Warn(msg, fields...) }
func Err(msg string, fields ...Field)   { globalLogger.Error(msg, fields...) }
