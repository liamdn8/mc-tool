package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
)

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	format string // text or json
	output io.Writer
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger
func NewLogger(level, format string) *Logger {
	return &Logger{
		level:  LogLevel(level),
		format: format,
		output: os.Stdout,
	}
}

// Debug logs debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DEBUG, message, fields...)
}

// Info logs info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(INFO, message, fields...)
}

// Warn logs warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WARN, message, fields...)
}

// Error logs error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.log(ERROR, message, fields...)
}

func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	// Check if we should log based on level
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   message,
	}

	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	if l.format == "json" {
		l.logJSON(entry)
	} else {
		l.logText(entry)
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.level]
}

func (l *Logger) logJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}
	fmt.Fprintln(l.output, string(data))
}

func (l *Logger) logText(entry LogEntry) {
	fieldsStr := ""
	if entry.Fields != nil {
		fieldsData, _ := json.Marshal(entry.Fields)
		fieldsStr = " " + string(fieldsData)
	}
	fmt.Fprintf(l.output, "%s [%s] %s%s\n", entry.Timestamp, entry.Level, entry.Message, fieldsStr)
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level, format string) {
	globalLogger = NewLogger(level, format)
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewLogger("info", "text")
	}
	return globalLogger
}
