package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Fields represents log fields
type Fields map[string]any

// Logger provides structured logging
type Logger struct {
	prefix   string
	fields   Fields
	minLevel Level
}

// Default returns a new logger with default settings
func Default() *Logger {
	return &Logger{
		minLevel: LevelInfo,
	}
}

// New creates a new logger with the given prefix
func New(prefix string) *Logger {
	return &Logger{
		prefix:   prefix,
		minLevel: LevelInfo,
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields Fields) *Logger {
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{
		prefix:   l.prefix,
		fields:   merged,
		minLevel: l.minLevel,
	}
}

// WithField returns a new logger with a single additional field
func (l *Logger) WithField(key string, value any) *Logger {
	return l.WithFields(Fields{key: value})
}

// WithRequestID returns a new logger with request ID
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// WithCorrelationID returns a new logger with correlation ID
func (l *Logger) WithCorrelationID(corrID string) *Logger {
	return l.WithField("correlation_id", corrID)
}

// WithUserID returns a new logger with user ID
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithField("user_id", userID)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.minLevel = level
}

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Prefix    string `json:"prefix,omitempty"`
	Fields    Fields `json:"fields,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Func      string `json:"func,omitempty"`
}

func (l *Logger) log(level Level, msg string, fields Fields) {
	if level < l.minLevel {
		return
	}

	// Merge fields
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}

	// Get caller info
	pc, file, line, ok := runtime.Caller(2)
	funcName := ""
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}

	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   msg,
		Prefix:    l.prefix,
		Fields:    merged,
		File:      file,
		Line:      line,
		Func:      funcName,
	}

	// Output as JSON
	output, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Println(string(output))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelDebug, msg, f)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelInfo, msg, f)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelWarn, msg, f)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelError, msg, f)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatal logs an error message and exits
func (l *Logger) Fatal(msg string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelError, msg, f)
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits
func (l *Logger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}
