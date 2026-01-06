package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

// Init initializes the global logger based on the environment
func Init(env string) {
	once.Do(func() {
		var handler slog.Handler
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}

		if env == "production" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}

		logger = slog.New(handler)
		slog.SetDefault(logger)
	})
}

// Get returns the global logger instance
func Get() *slog.Logger {
	if logger == nil {
		// Fallback to default if not initialized
		return slog.Default()
	}
	return logger
}

// Info logs an informational message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}
