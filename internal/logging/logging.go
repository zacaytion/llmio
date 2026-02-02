// Package logging provides structured logging configuration using log/slog.
//
// # File Handle Management
//
// When logging to files, the Setup() and SetupDefault() functions open file handles
// that are intended to remain open for the application lifetime. For short-lived
// processes or tests, use SetupWithCleanup() which returns a cleanup function.
//
// The design assumes logging is configured once at startup:
//   - SetupDefault() - For typical server applications (file stays open until exit)
//   - SetupWithCleanup() - For CLI tools or tests that need explicit cleanup
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/zacaytion/llmio/internal/config"
)

// CloseFunc is a function that cleans up logging resources.
// Returns nil on success or an error if cleanup failed.
type CloseFunc func() error

// Setup creates and returns a configured slog.Logger based on LoggingConfig.
// The fallbackWriter is used when the configured output cannot be opened.
// Returns the configured logger (does not set as default - caller should do that).
//
// Note: When logging to a file, the file handle is not explicitly closed.
// For applications that need to close the file handle, use SetupWithCleanup instead.
func Setup(cfg config.LoggingConfig, fallbackWriter io.Writer) *slog.Logger {
	level := parseLevel(cfg.Level)
	writer := getWriter(cfg.Output, fallbackWriter)
	handler := createHandler(cfg.Format, writer, level)
	return slog.New(handler)
}

// SetupWithCleanup creates a logger and returns a cleanup function for the file handle.
// The cleanup function should be called when logging is no longer needed (e.g., on shutdown).
// For stdout/stderr output, the cleanup function is a no-op but still safe to call.
func SetupWithCleanup(cfg config.LoggingConfig, fallbackWriter io.Writer) (*slog.Logger, CloseFunc) {
	level := parseLevel(cfg.Level)
	writer, closer := getWriterWithCleanup(cfg.Output, fallbackWriter)
	handler := createHandler(cfg.Format, writer, level)
	return slog.New(handler), closer
}

// SetupDefault creates a logger and sets it as the default slog logger.
// The file handle (if logging to a file) remains open for the application lifetime.
// For explicit cleanup, use SetupWithCleanup instead.
func SetupDefault(cfg config.LoggingConfig) {
	logger := Setup(cfg, os.Stdout)
	slog.SetDefault(logger)
}

// SetupDefaultWithCleanup creates a logger, sets it as default, and returns a cleanup function.
// Call the cleanup function on application shutdown to close any file handles.
func SetupDefaultWithCleanup(cfg config.LoggingConfig) CloseFunc {
	logger, closer := SetupWithCleanup(cfg, os.Stdout)
	slog.SetDefault(logger)
	return closer
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
	writer, _ := getWriterWithCleanup(output, overrideWriter)
	return writer
}

// getWriterWithCleanup returns the writer and a cleanup function.
// The cleanup function closes the file handle if one was opened.
func getWriterWithCleanup(output string, overrideWriter io.Writer) (io.Writer, CloseFunc) {
	noopCloser := func() error { return nil }

	switch strings.ToLower(output) {
	case "stdout", "":
		if overrideWriter != nil {
			return overrideWriter, noopCloser
		}
		return os.Stdout, noopCloser
	case "stderr":
		if overrideWriter != nil {
			return overrideWriter, noopCloser
		}
		return os.Stderr, noopCloser
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
				return overrideWriter, noopCloser
			}
			return os.Stdout, noopCloser
		}
		return file, func() error { return file.Close() }
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
