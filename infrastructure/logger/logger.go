package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	UserIDKey        contextKey = "user_id"
)

// Logger wraps logrus with context-aware logging
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new structured logger
func NewLogger() *Logger {
	log := logrus.New()

	// Set JSON format for production
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	// Set level based on environment
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	return &Logger{Logger: log}
}

// WithContext creates a logger with context fields
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithFields(logrus.Fields{})

	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		entry = entry.WithField("user_id", userID)
	}

	return entry
}

// WithFields creates a logger with additional fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// RequestLogger logs HTTP requests
func (l *Logger) RequestLogger(ctx context.Context, method, path string, statusCode int, duration string) {
	l.WithContext(ctx).WithFields(logrus.Fields{
		"method":    method,
		"path":      path,
		"status":    statusCode,
		"duration":  duration,
		"component": "http",
	}).Info("HTTP request processed")
}

// ErrorLogger logs errors with context
func (l *Logger) ErrorLogger(ctx context.Context, err error, message string, fields logrus.Fields) {
	entry := l.WithContext(ctx).WithError(err)
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	entry.Error(message)
}

// DatabaseLogger logs database operations
func (l *Logger) DatabaseLogger(ctx context.Context, operation, table string, duration string, err error) {
	entry := l.WithContext(ctx).WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration,
		"component": "database",
	})

	if err != nil {
		entry.WithError(err).Error("Database operation failed")
	} else {
		entry.Debug("Database operation completed")
	}
}

// ContextWithCorrelationID adds correlation ID to context
func ContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// ContextWithUserID adds user ID to context
func ContextWithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
