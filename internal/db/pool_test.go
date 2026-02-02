package db

import (
	"context"
	"testing"
	"time"

	"github.com/zacaytion/llmio/internal/config"
)

// T014: Test for NewPoolFromConfig.
func TestNewPoolFromConfig(t *testing.T) {
	// This test verifies the function signature and config mapping.
	// It does NOT actually connect to a database (that would be an integration test).

	cfg := config.DatabaseConfig{
		Host:              "testhost",
		Port:              5433,
		User:              "testuser",
		Password:          "testpass",
		Name:              "testdb",
		SSLMode:           "disable",
		MaxConns:          10,
		MinConns:          1,
		MaxConnLifetime:   2 * time.Hour,
		MaxConnIdleTime:   15 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	// We can't actually connect without a real database,
	// but we can verify the function exists and accepts the config.
	// The actual connection test would be in integration tests.

	// For unit testing, we verify the DSN is correctly built from config
	// Note: Passwords are now single-quoted for special character support
	expectedDSN := "host=testhost port=5433 user=testuser dbname=testdb sslmode=disable password='testpass'"
	if cfg.DSN() != expectedDSN {
		t.Errorf("Config DSN = %q, want %q", cfg.DSN(), expectedDSN)
	}

	// Verify pool settings are accessible
	if cfg.MaxConns != 10 {
		t.Errorf("MaxConns = %d, want 10", cfg.MaxConns)
	}
	if cfg.MinConns != 1 {
		t.Errorf("MinConns = %d, want 1", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 2*time.Hour {
		t.Errorf("MaxConnLifetime = %v, want 2h", cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime != 15*time.Minute {
		t.Errorf("MaxConnIdleTime = %v, want 15m", cfg.MaxConnIdleTime)
	}
	if cfg.HealthCheckPeriod != 30*time.Second {
		t.Errorf("HealthCheckPeriod = %v, want 30s", cfg.HealthCheckPeriod)
	}
}

// TestNewPoolFromConfig_FunctionExists verifies the function signature compiles.
// This will fail to compile if NewPoolFromConfig doesn't exist with the right signature.
func TestNewPoolFromConfig_FunctionExists(t *testing.T) {
	// This is a compile-time check. If NewPoolFromConfig doesn't exist,
	// the test file won't compile.
	var _ = NewPoolFromConfig // Reference the function to ensure it exists
}

// T137: Test that NewPoolFromConfig returns error for invalid connection.
// This is an integration test that verifies error handling for unreachable hosts.
func TestNewPoolFromConfig_ConnectionError(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:              "nonexistent.invalid.host.example.com",
		Port:              5432,
		User:              "testuser",
		Password:          "testpass",
		Name:              "testdb",
		SSLMode:           "disable",
		MaxConns:          10,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   15 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := NewPoolFromConfig(ctx, cfg)
	if err == nil {
		t.Error("expected error when connecting to invalid host, got nil")
	}
}
