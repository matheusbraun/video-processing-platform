package logging

import (
	"log"
	"os"
)

// Logger is a simple structured logger
type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	warnLog  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string) *Logger {
	prefix := "[" + serviceName + "] "

	return &Logger{
		infoLog:  log.New(os.Stdout, prefix+"INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, prefix+"ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:  log.New(os.Stdout, prefix+"WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.infoLog.Println(v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLog.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.errorLog.Println(v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	l.warnLog.Println(v...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.warnLog.Printf(format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLog.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLog.Fatalf(format, v...)
}

var defaultLogger = NewLogger("app")

func Info(msg string, args ...interface{}) {
	defaultLogger.Infof("%s %v", msg, args)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Errorf("%s %v", msg, args)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warnf("%s %v", msg, args)
}

func Fatal(msg string, args ...interface{}) {
	defaultLogger.Fatalf("%s %v", msg, args)
}
