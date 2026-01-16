package logging

import (
	"log/slog"
	"os"
)

// Logger is a structured JSON logger using slog
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new logger instance with structured JSON output
func NewLogger(serviceName string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	return &Logger{
		logger: logger,
	}
}

// Info logs an info message with structured fields
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Error logs an error message with structured fields
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// Warn logs a warning message with structured fields
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Debug logs a debug message with structured fields
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// With returns a new logger with additional fields
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With(slog.String("error", err.Error())),
	}
}

var defaultLogger = NewLogger("app")

func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}
