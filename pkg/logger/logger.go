package logger

import (
	"encoding/json"
	"log"
	"time"

	"github.com/chthon/shortlink/pkg/config"
)

// Logger interface
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// SimpleLogger implements Logger interface
type SimpleLogger struct {
	level  string
	format string
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config, serviceName string) Logger {
	return &SimpleLogger{
		level:  cfg.Logging.Level,
		format: cfg.Logging.Format,
	}
}

// Info logs info level messages
func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	l.log("INFO", msg, fields...)
}

// Error logs error level messages
func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	l.log("ERROR", msg, fields...)
}

// Debug logs debug level messages
func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	if l.level == "debug" {
		l.log("DEBUG", msg, fields...)
	}
}

// Warn logs warning level messages
func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	l.log("WARN", msg, fields...)
}

// log internal logging function
func (l *SimpleLogger) log(level, msg string, fields ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   msg,
		Fields:    make(map[string]interface{}),
	}

	// Parse fields as key-value pairs
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			if key, ok := fields[i].(string); ok {
				entry.Fields[key] = fields[i+1]
			}
		}
	}

	if l.format == "json" {
		jsonBytes, _ := json.Marshal(entry)
		log.Println(string(jsonBytes))
	} else {
		log.Printf("[%s] %s: %s", level, entry.Timestamp.Format(time.RFC3339), msg)
	}
}

// RequestLogger middleware for logging HTTP requests
func RequestLogger(logger Logger, serviceName string) func(c interface{}) {
	return func(c interface{}) {
		start := time.Now()

		// This is a placeholder - in real implementation, you'd extract
		// request details from the specific framework's context
		logger.Info("HTTP Request",
			"service", serviceName,
			"duration", time.Since(start),
			"timestamp", start.UTC(),
		)
	}
}

// Default logger instance
var defaultLogger Logger

// InitDefaultLogger initializes the default logger
func InitDefaultLogger(cfg *config.Config, serviceName string) {
	defaultLogger = NewLogger(cfg, serviceName)
}

// Info logs to default logger
func Info(msg string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, fields...)
	} else {
		log.Printf("[INFO] %s", msg)
	}
}

// Error logs to default logger
func Error(msg string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, fields...)
	} else {
		log.Printf("[ERROR] %s", msg)
	}
}

// Debug logs to default logger
func Debug(msg string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, fields...)
	} else {
		log.Printf("[DEBUG] %s", msg)
	}
}

// Warn logs to default logger
func Warn(msg string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, fields...)
	} else {
		log.Printf("[WARN] %s", msg)
	}
}
