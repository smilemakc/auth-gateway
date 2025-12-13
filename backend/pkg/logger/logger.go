package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Logger provides structured logging
type Logger struct {
	level      LogLevel
	service    string
	jsonOutput bool
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new logger instance
func New(service string, level LogLevel, jsonOutput bool) *Logger {
	return &Logger{
		level:      level,
		service:    service,
		jsonOutput: jsonOutput,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	if l.shouldLog(DebugLevel) {
		l.log(DebugLevel, message, fields...)
	}
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	if l.shouldLog(InfoLevel) {
		l.log(InfoLevel, message, fields...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	if l.shouldLog(WarnLevel) {
		l.log(WarnLevel, message, fields...)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	if l.shouldLog(ErrorLevel) {
		l.log(ErrorLevel, message, fields...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	l.log(FatalLevel, message, fields...)
	os.Exit(1)
}

// log performs the actual logging
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Service:   l.service,
		Message:   message,
	}

	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	if l.jsonOutput {
		output, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Error marshaling log entry: %v", err)
			return
		}
		fmt.Println(string(output))
	} else {
		l.printPlain(entry)
	}
}

// printPlain prints a log entry in plain text format
func (l *Logger) printPlain(entry LogEntry) {
	fieldsStr := ""
	if entry.Fields != nil {
		fieldsJSON, _ := json.Marshal(entry.Fields)
		fieldsStr = " " + string(fieldsJSON)
	}

	fmt.Printf("[%s] %s [%s] %s%s\n",
		entry.Timestamp,
		entry.Level,
		entry.Service,
		entry.Message,
		fieldsStr,
	)
}

// shouldLog checks if the log level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DebugLevel: 0,
		InfoLevel:  1,
		WarnLevel:  2,
		ErrorLevel: 3,
		FatalLevel: 4,
	}

	return levels[level] >= levels[l.level]
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// This is a simplified version - in production, you might want to maintain fields
	return l
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Default logger instance
var defaultLogger = New("auth-gateway", InfoLevel, true)

// Global logging functions

// Debug logs a debug message using the default logger
func Debug(message string, fields ...map[string]interface{}) {
	defaultLogger.Debug(message, fields...)
}

// Info logs an info message using the default logger
func Info(message string, fields ...map[string]interface{}) {
	defaultLogger.Info(message, fields...)
}

// Warn logs a warning message using the default logger
func Warn(message string, fields ...map[string]interface{}) {
	defaultLogger.Warn(message, fields...)
}

// Error logs an error message using the default logger
func Error(message string, fields ...map[string]interface{}) {
	defaultLogger.Error(message, fields...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(message string, fields ...map[string]interface{}) {
	defaultLogger.Fatal(message, fields...)
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}
