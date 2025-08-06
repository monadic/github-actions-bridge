package bridge

import (
	"fmt"
	"log"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LogLevelDebug is for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for general informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
)

// Logger provides structured logging for the bridge
type Logger struct {
	level  LogLevel
	prefix string
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *Logger {
	level := LogLevelInfo
	
	// Check environment variable for log level
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		switch lvl {
		case "debug", "DEBUG":
			level = LogLevelDebug
		case "warn", "WARN":
			level = LogLevelWarn
		case "error", "ERROR":
			level = LogLevelError
		}
	}
	
	return &Logger{
		level:  level,
		prefix: prefix,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		msg := fmt.Sprintf(format, args...)
		log.Printf("[DEBUG] [%s] %s", l.prefix, msg)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		msg := fmt.Sprintf(format, args...)
		log.Printf("[INFO] [%s] %s", l.prefix, msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		msg := fmt.Sprintf(format, args...)
		log.Printf("[WARN] [%s] %s", l.prefix, msg)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		msg := fmt.Sprintf(format, args...)
		log.Printf("[ERROR] [%s] %s", l.prefix, msg)
	}
}

// WorkflowExecutionLog logs workflow execution details
func (l *Logger) WorkflowExecutionLog(execID, unitID string, status string, duration string) {
	l.Info("Workflow execution: id=%s unit=%s status=%s duration=%s", 
		execID, unitID, status, duration)
}

// SecurityLog logs security-related events
func (l *Logger) SecurityLog(event string, details map[string]interface{}) {
	detailStr := ""
	for k, v := range details {
		detailStr += fmt.Sprintf(" %s=%v", k, v)
	}
	l.Warn("Security event: %s%s", event, detailStr)
}