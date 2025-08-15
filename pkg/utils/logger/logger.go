package logger

import (
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Logger struct {
	level LogLevel
	prefix string
	debugLog *log.Logger
	infoLog *log.Logger
	warnLog *log.Logger
	errorLog *log.Logger
	fatalLog *log.Logger
}

// New creates a new logger instance
func New(level LogLevel, prefix string) *Logger {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	return &Logger{
		level:    level,
		prefix:   prefix,
		debugLog: log.New(os.Stdout, "🐛 DEBUG ", flags),
		infoLog:  log.New(os.Stdout, "ℹ️  INFO  ", flags),
		warnLog:  log.New(os.Stdout, "⚠️  WARN  ", flags),
		errorLog: log.New(os.Stderr, "❌ ERROR ", flags),
		fatalLog: log.New(os.Stderr, "💀 FATAL ", flags),
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debugLog.Printf(format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.infoLog.Printf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warnLog.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errorLog.Printf(format, v...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.fatalLog.Printf(format, v...)
		os.Exit(1)
	}
}

// HTTPRequest logs an HTTP request
func (l *Logger) HTTPRequest(method, path, remoteAddr string, statusCode int, duration time.Duration, userAgent string) {
	var emoji string
	switch {
	case statusCode >= 200 && statusCode < 300:
		emoji = "✅"
	case statusCode >= 300 && statusCode < 400:
		emoji = "🔄"
	case statusCode >= 400 && statusCode < 500:
		emoji = "⚠️"
	case statusCode >= 500:
		emoji = "❌"
	default:
		emoji = "❓"
	}

	l.Info("%s %s %s | %d | %v | %s | %s", 
		emoji, method, path, statusCode, duration, remoteAddr, userAgent)
}

// Database logs database operations
func (l *Logger) Database(operation, table string, duration time.Duration, err error) {
	if err != nil {
		l.Error("🗄️  DB %s on %s failed after %v: %v", operation, table, duration, err)
	} else {
		l.Debug("🗄️  DB %s on %s completed in %v", operation, table, duration)
	}
}

// Auth logs authentication events
func (l *Logger) Auth(event, user string, success bool, details ...interface{}) {
	var emoji string
	if success {
		emoji = "��"
		l.Info("%s AUTH %s for user %s", emoji, event, user)
	} else {
		emoji = "��"
		l.Warn("%s AUTH %s failed for user %s: %v", emoji, event, user, details)
	}
}

// Global logger instance
var Global *Logger

// Initialize the global logger
func init(){
	// Default to INFO level, can be overridden by environment variable
	level := INFO
	if os.Getenv("LOG_LEVEL")  == "DEBUG" {
		level = DEBUG
	}
	
	Global = New(level, "[FLUXIO]")
}

// Convenience functions for global logger
func Debug(format string, v ...interface{}) { Global.Debug(format, v...) }
func Info(format string, v ...interface{})  { Global.Info(format, v...) }
func Warn(format string, v ...interface{})  { Global.Warn(format, v...) }
func Error(format string, v ...interface{}) { Global.Error(format, v...) }
func Fatal(format string, v ...interface{}) { Global.Fatal(format, v...) }

// HTTPRequest logs an HTTP request using the global logger
func HTTPRequest(method, path, remoteAddr string, statusCode int, duration time.Duration, userAgent string) {
	Global.HTTPRequest(method, path, remoteAddr, statusCode, duration, userAgent)
}

// Database logs database operations using the global logger
func Database(operation, table string, duration time.Duration, err error) {
	Global.Database(operation, table, duration, err)
}

// Auth logs authentication events using the global logger
func Auth(event, user string, success bool, details ...interface{}) {
	Global.Auth(event, user, success, details...)
}