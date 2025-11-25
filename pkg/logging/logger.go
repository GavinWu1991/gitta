package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with convenience methods.
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new structured logger with the specified level.
func NewLogger(level string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewJSONLogger creates a JSON-formatted logger (for production).
func NewJSONLogger(level string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stderr, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithTraceID adds a trace ID to the logger context for request correlation.
func (l *Logger) WithTraceID(ctx context.Context, traceID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("trace_id", traceID),
	}
}

// WithContext adds context values to the logger.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace ID from context if present
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return l.WithTraceID(ctx, traceID)
	}
	return l
}
