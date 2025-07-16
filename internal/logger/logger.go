// Package logger provides structured logging configuration
package logger

import (
	"log/slog"
	"os"
)

// InitLogger initializes the structured logger with the specified log level
func InitLogger(level string) {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelError
	}

	// Create a new logger with JSON output for structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Set as default logger
	slog.SetDefault(logger)
}
