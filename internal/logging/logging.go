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
//
// # Error Handling
//
// SetupWithCleanup returns an error if the configured log file cannot be opened.
// This prevents silent failures where misconfigured logging goes unnoticed.
// SetupDefault and Setup fall back to stdout on file errors (legacy behavior).
package logging

import (
	"fmt"
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
// This function falls back to stdout on file open errors (use SetupWithCleanup for strict mode).
func Setup(cfg config.LoggingConfig, fallbackWriter io.Writer) *slog.Logger {
	level := parseLevel(cfg.Level)
	writer := getWriter(cfg.Output, fallbackWriter)
	handler := createHandler(cfg.Format, writer, level)
	return slog.New(handler)
}

// SetupWithCleanup creates a logger and returns a cleanup function for the file handle.
// The cleanup function should be called when logging is no longer needed (e.g., on shutdown).
// For stdout/stderr output, the cleanup function is a no-op but still safe to call.
//
// Returns an error if the configured file output cannot be opened. This ensures
// misconfigurations are caught at startup rather than silently falling back.
func SetupWithCleanup(cfg config.LoggingConfig, fallbackWriter io.Writer) (*slog.Logger, CloseFunc, error) {
	level := parseLevel(cfg.Level)
	writer, closer, err := getWriterWithCleanupStrict(cfg.Output, fallbackWriter)
	if err != nil {
		return nil, nil, err
	}
	handler := createHandler(cfg.Format, writer, level)
	return slog.New(handler), closer, nil
}

// SetupDefault creates a logger and sets it as the default slog logger.
// The file handle (if logging to a file) remains open for the application lifetime.
// For explicit cleanup, use SetupWithCleanup instead.
// Falls back to stdout if the configured file cannot be opened.
func SetupDefault(cfg config.LoggingConfig) {
	logger := Setup(cfg, os.Stdout)
	slog.SetDefault(logger)
}

// SetupDefaultWithCleanup creates a logger, sets it as default, and returns a cleanup function.
// Call the cleanup function on application shutdown to close any file handles.
// Returns an error if the configured file output cannot be opened.
func SetupDefaultWithCleanup(cfg config.LoggingConfig) (CloseFunc, error) {
	logger, closer, err := SetupWithCleanup(cfg, os.Stdout)
	if err != nil {
		return nil, err
	}
	slog.SetDefault(logger)
	return closer, nil
}

// parseLevel converts a string level to slog.Level.
// Invalid levels default to info with a warning logged.
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
		if level != "" {
			// Log warning for invalid level (uses current default logger)
			slog.Warn("invalid log level, defaulting to info",
				"configured", level,
				"valid_levels", "debug, info, warn, error",
			)
		}
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
// Falls back to overrideWriter or stdout on file open errors (legacy behavior).
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

// getWriterWithCleanupStrict returns the writer and a cleanup function, or an error.
// Unlike getWriterWithCleanup, this function returns an error instead of falling back
// when the configured file cannot be opened.
func getWriterWithCleanupStrict(output string, overrideWriter io.Writer) (io.Writer, CloseFunc, error) {
	noopCloser := func() error { return nil }

	switch strings.ToLower(output) {
	case "stdout", "":
		if overrideWriter != nil {
			return overrideWriter, noopCloser, nil
		}
		return os.Stdout, noopCloser, nil
	case "stderr":
		if overrideWriter != nil {
			return overrideWriter, noopCloser, nil
		}
		return os.Stderr, noopCloser, nil
	default:
		// Treat as file path - path comes from trusted config, not user input
		file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644) //nolint:gosec // G304: path comes from config file, not untrusted input
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file %q: %w", output, err)
		}
		return file, func() error { return file.Close() }, nil
	}
}

// createHandler creates the appropriate slog.Handler based on format.
// Invalid formats default to JSON with a warning logged.
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
		// Log warning for invalid format (uses current default logger)
		slog.Warn("invalid log format, defaulting to json",
			"configured", format,
			"valid_formats", "json, text",
		)
		return slog.NewJSONHandler(writer, opts)
	}
}
