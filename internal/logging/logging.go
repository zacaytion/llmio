// Package logging provides structured logging configuration using log/slog.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/zacaytion/llmio/internal/config"
)

// Setup creates and returns a configured slog.Logger based on LoggingConfig.
// The fallbackWriter is used when the configured output cannot be opened.
// Returns the configured logger (does not set as default - caller should do that).
func Setup(cfg config.LoggingConfig, fallbackWriter io.Writer) *slog.Logger {
	level := parseLevel(cfg.Level)
	writer := getWriter(cfg.Output, fallbackWriter)
	handler := createHandler(cfg.Format, writer, level)
	return slog.New(handler)
}

// SetupDefault creates a logger and sets it as the default slog logger.
func SetupDefault(cfg config.LoggingConfig) {
	logger := Setup(cfg, os.Stdout)
	slog.SetDefault(logger)
}

// parseLevel converts a string level to slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getWriter returns the appropriate io.Writer for the output configuration.
// The overrideWriter, if non-nil, is used for stdout/stderr (useful for testing).
// Falls back to overrideWriter or stdout if the configured file cannot be opened.
func getWriter(output string, overrideWriter io.Writer) io.Writer {
	switch strings.ToLower(output) {
	case "stdout", "":
		if overrideWriter != nil {
			return overrideWriter
		}
		return os.Stdout
	case "stderr":
		if overrideWriter != nil {
			return overrideWriter
		}
		return os.Stderr
	default:
		// Treat as file path - path comes from trusted config, not user input
		file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644) //nolint:gosec // G304: path comes from config file, not untrusted input
		if err != nil {
			// Log the error and fall back
			slog.Warn("failed to open log file, falling back",
				"path", output,
				"error", err,
			)
			if overrideWriter != nil {
				return overrideWriter
			}
			return os.Stdout
		}
		return file
	}
}

// createHandler creates the appropriate slog.Handler based on format.
func createHandler(format string, writer io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(format) {
	case "text":
		return slog.NewTextHandler(writer, opts)
	case "json", "":
		return slog.NewJSONHandler(writer, opts)
	default:
		return slog.NewJSONHandler(writer, opts)
	}
}
