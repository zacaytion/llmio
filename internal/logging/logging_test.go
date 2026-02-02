package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/zacaytion/llmio/internal/config"
)

// T036: Test for logging.Setup() with JSON format.
func TestSetup_JSONFormat(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	// Capture stdout
	var buf bytes.Buffer
	logger := Setup(cfg, &buf)

	logger.Info("test message", "key", "value")

	// Verify output is valid JSON
	output := buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}

	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("expected valid JSON, got: %s, error: %v", output, err)
	}

	// Verify expected fields
	if logEntry["msg"] != "test message" {
		t.Errorf("expected msg='test message', got %v", logEntry["msg"])
	}
	if logEntry["key"] != "value" {
		t.Errorf("expected key='value', got %v", logEntry["key"])
	}
}

// T037: Test for log level filtering.
func TestSetup_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logFunc   func(*slog.Logger)
		shouldLog bool
	}{
		{
			name:      "info level logs info",
			level:     "info",
			logFunc:   func(l *slog.Logger) { l.Info("info message") },
			shouldLog: true,
		},
		{
			name:      "info level filters debug",
			level:     "info",
			logFunc:   func(l *slog.Logger) { l.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "warn level filters info",
			level:     "warn",
			logFunc:   func(l *slog.Logger) { l.Info("info message") },
			shouldLog: false,
		},
		{
			name:      "warn level logs warn",
			level:     "warn",
			logFunc:   func(l *slog.Logger) { l.Warn("warn message") },
			shouldLog: true,
		},
		{
			name:      "error level logs error",
			level:     "error",
			logFunc:   func(l *slog.Logger) { l.Error("error message") },
			shouldLog: true,
		},
		{
			name:      "debug level logs debug",
			level:     "debug",
			logFunc:   func(l *slog.Logger) { l.Debug("debug message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggingConfig{
				Level:  tt.level,
				Format: "json",
				Output: "stdout",
			}

			var buf bytes.Buffer
			logger := Setup(cfg, &buf)
			tt.logFunc(logger)

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("expected shouldLog=%v, got hasOutput=%v, output=%q", tt.shouldLog, hasOutput, buf.String())
			}
		})
	}
}

// T038: Test for log file fallback to stdout.
func TestSetup_FileFallbackToStdout(t *testing.T) {
	// Use a path that doesn't exist and can't be created
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "/nonexistent/path/that/cannot/exist/logfile.log",
	}

	// Setup should not panic, should fall back to stdout
	var buf bytes.Buffer
	logger := Setup(cfg, &buf)

	// Should still be able to log
	logger.Info("fallback test")

	// Output should go to our buffer (simulating stdout fallback)
	if buf.Len() == 0 {
		t.Error("expected log output after fallback, got nothing")
	}
}

// TestSetup_TextFormat verifies text format output.
func TestSetup_TextFormat(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}

	var buf bytes.Buffer
	logger := Setup(cfg, &buf)

	logger.Info("text message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Fatal("expected log output, got empty string")
	}

	// Text format should NOT be valid JSON
	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err == nil {
		t.Error("text format should not be valid JSON")
	}

	// Should contain the message
	if !strings.Contains(output, "text message") {
		t.Errorf("expected output to contain 'text message', got: %s", output)
	}
}

// TestSetup_FileOutput verifies file output works.
func TestSetup_FileOutput(t *testing.T) {
	tmpFile := t.TempDir() + "/test.log"

	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: tmpFile,
	}

	// Pass nil for fallback since we expect file to work
	logger := Setup(cfg, os.Stdout)

	logger.Info("file output test")

	// Read the file - tmpFile is from t.TempDir(), a controlled test path
	content, err := os.ReadFile(tmpFile) //nolint:gosec // G304: tmpFile is from t.TempDir(), not untrusted input
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected log content in file, got empty")
	}

	// Verify it's valid JSON
	var logEntry map[string]any
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Errorf("expected valid JSON in file, got: %s, error: %v", content, err)
	}
}

// T055: Test that SetupWithCleanup returns a closer for file handles.
func TestSetupWithCleanup_ReturnsCloser(t *testing.T) {
	tmpFile := t.TempDir() + "/test_cleanup.log"

	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: tmpFile,
	}

	logger, closer := SetupWithCleanup(cfg, os.Stdout)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// When logging to a file, closer should be non-nil
	if closer == nil {
		t.Fatal("expected non-nil closer when logging to file")
	}

	// Log something
	logger.Info("test message before close")

	// Close should not error
	if err := closer(); err != nil {
		t.Errorf("closer returned error: %v", err)
	}
}

// TestSetupWithCleanup_StdoutNoClose verifies stdout doesn't need closing.
func TestSetupWithCleanup_StdoutNoClose(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	var buf bytes.Buffer
	logger, closer := SetupWithCleanup(cfg, &buf)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// For stdout/stderr, closer should be a no-op (not nil for consistency)
	if closer == nil {
		t.Fatal("expected non-nil closer even for stdout (should be no-op)")
	}

	// Calling closer should not error
	if err := closer(); err != nil {
		t.Errorf("closer returned error: %v", err)
	}
}
