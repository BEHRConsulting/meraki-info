// Package logger provides structured logging configuration
//
// It installs a JSON log/slog handler on stderr as the process-wide default
// logger. Logs go to stderr so that command output on stdout stays clean
// for piping into jq, CSV tools and the like.
package logger

import (
	"log/slog"
	"os"
)

// InitLogger initializes the structured logger with the specified log level
//
// Parameters:
//   - level: "debug", "info" or "error". Any other value, including "",
//     silently means "error", the quietest level, so a typo never makes
//     the tool noisier than its default.
//
// It replaces slog's default logger, so every slog.Info/Debug/Error call in
// the program uses the new level from then on.
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
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Set as default logger
	slog.SetDefault(logger)
}
