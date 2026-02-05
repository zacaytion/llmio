package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zacaytion/llmio/internal/validation"
)

// T005: Test for Config struct existence.
func Test_ConfigStructs(t *testing.T) {
	// Verify structs can be instantiated with expected fields
	cfg := Config{
		PG: PGConfig{
			Host:              "localhost",
			Port:              5432,
			Database:          "testdb",
			SSLMode:           "disable",
			MaxConns:          25,
			MinConns:          2,
			MaxConnLifetime:   time.Hour,
			MaxConnIdleTime:   30 * time.Minute,
			HealthCheckPeriod: time.Minute,
			UserAdmin:         "postgres",
			PassAdmin:         "secret",
		},
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Session: SessionConfig{
			Duration:        168 * time.Hour,
			CleanupInterval: 10 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}

	if cfg.PG.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.PG.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Server.Port)
	}
	if cfg.Session.Duration != 168*time.Hour {
		t.Errorf("expected 168h, got %v", cfg.Session.Duration)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected json, got %s", cfg.Logging.Format)
	}
}

// clearLlmioEnvVars clears all LLMIO_* environment variables for the duration of the test.
// This ensures tests for defaults aren't affected by env vars set in the shell.
func clearLlmioEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"LLMIO_PG_HOST",
		"LLMIO_PG_PORT",
		"LLMIO_PG_DATABASE",
		"LLMIO_PG_SSLMODE",
		"LLMIO_PG_MAX_CONNS",
		"LLMIO_PG_MIN_CONNS",
		"LLMIO_PG_MAX_CONN_LIFETIME",
		"LLMIO_PG_MAX_CONN_IDLE_TIME",
		"LLMIO_PG_HEALTH_CHECK_PERIOD",
		"LLMIO_PG_USER_ADMIN",
		"LLMIO_PG_PASS_ADMIN",
		"LLMIO_PG_USER_MIGRATION",
		"LLMIO_PG_PASS_MIGRATION",
		"LLMIO_PG_USER_APP",
		"LLMIO_PG_PASS_APP",
		"LLMIO_SERVER_PORT",
		"LLMIO_SERVER_READ_TIMEOUT",
		"LLMIO_SERVER_WRITE_TIMEOUT",
		"LLMIO_SERVER_IDLE_TIMEOUT",
		"LLMIO_SESSION_DURATION",
		"LLMIO_SESSION_CLEANUP_INTERVAL",
		"LLMIO_LOGGING_LEVEL",
		"LLMIO_LOGGING_FORMAT",
		"LLMIO_LOGGING_OUTPUT",
	}
	for _, env := range envVars {
		t.Setenv(env, "")
	}
}

// T006: Test for Load() with defaults.
//
//nolint:gocyclo // Test function validating many config defaults - complexity is intentional
func Test_Load_Defaults(t *testing.T) {
	clearLlmioEnvVars(t) // Clear any env vars that might override defaults

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// PostgreSQL defaults
	if cfg.PG.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.PG.Host)
	}
	if cfg.PG.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.PG.Port)
	}
	if cfg.PG.UserAdmin != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.PG.UserAdmin)
	}
	if cfg.PG.Database != "loomio_development" {
		t.Errorf("expected loomio_development, got %s", cfg.PG.Database)
	}
	if cfg.PG.SSLMode != "disable" {
		t.Errorf("expected disable, got %s", cfg.PG.SSLMode)
	}
	if cfg.PG.MaxConns != 25 {
		t.Errorf("expected 25, got %d", cfg.PG.MaxConns)
	}
	if cfg.PG.MinConns != 2 {
		t.Errorf("expected 2, got %d", cfg.PG.MinConns)
	}
	if cfg.PG.MaxConnLifetime != time.Hour {
		t.Errorf("expected 1h, got %v", cfg.PG.MaxConnLifetime)
	}
	if cfg.PG.MaxConnIdleTime != 30*time.Minute {
		t.Errorf("expected 30m, got %v", cfg.PG.MaxConnIdleTime)
	}
	if cfg.PG.HealthCheckPeriod != time.Minute {
		t.Errorf("expected 1m, got %v", cfg.PG.HealthCheckPeriod)
	}

	// Server defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("expected 15s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 15*time.Second {
		t.Errorf("expected 15s, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Server.IdleTimeout != 60*time.Second {
		t.Errorf("expected 60s, got %v", cfg.Server.IdleTimeout)
	}

	// Session defaults
	if cfg.Session.Duration != 168*time.Hour {
		t.Errorf("expected 168h, got %v", cfg.Session.Duration)
	}
	if cfg.Session.CleanupInterval != 10*time.Minute {
		t.Errorf("expected 10m, got %v", cfg.Session.CleanupInterval)
	}

	// Logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("expected info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected json, got %s", cfg.Logging.Format)
	}
	if cfg.Logging.Output != "stdout" {
		t.Errorf("expected stdout, got %s", cfg.Logging.Output)
	}
}

// T033: Test for environment variable override (LLMIO_*).
func Test_Load_EnvVarOverride(t *testing.T) {
	// Set environment variable
	t.Setenv("LLMIO_SERVER_PORT", "9001")
	t.Setenv("LLMIO_PG_DATABASE", "loomio_from_env")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env var should override default
	if cfg.Server.Port != 9001 {
		t.Errorf("expected 9001 (from LLMIO_SERVER_PORT), got %d", cfg.Server.Port)
	}
	if cfg.PG.Database != "loomio_from_env" {
		t.Errorf("expected loomio_from_env (from LLMIO_PG_DATABASE), got %s", cfg.PG.Database)
	}
}

// T028: Test for CLI flag override of config file value.
func Test_LoadWithViper_CLIOverride(t *testing.T) {
	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
server:
  port: 8080
pg:
  database: loomio_development
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Create a viper instance with CLI flag bound
	v := NewViper()

	// Simulate CLI flag override
	v.Set("server.port", 9000)

	cfg, err := LoadWithViper(v, configPath)
	if err != nil {
		t.Fatalf("LoadWithViper failed: %v", err)
	}

	// CLI flag should override config file value
	if cfg.Server.Port != 9000 {
		t.Errorf("expected 9000 (CLI override), got %d", cfg.Server.Port)
	}

	// Config file value should still apply where not overridden
	if cfg.PG.Database != "loomio_development" {
		t.Errorf("expected loomio_development, got %s", cfg.PG.Database)
	}
}

// T023: Test for config.test.yaml loading different database name.
func Test_Load_TestConfig(t *testing.T) {
	clearLlmioEnvVars(t) // Clear env vars so file values are used

	// Create a temporary test config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.test.yaml"

	yamlContent := `
pg:
  host: localhost
  port: 5432
  user_admin: testuser
  pass_admin: testpass
  database: loomio_test
  max_conns: 5
  min_conns: 1
server:
  port: 8081
session:
  duration: 1h
  cleanup_interval: 1m
logging:
  level: warn
  format: text
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify test database name is different from development
	if cfg.PG.Database != "loomio_test" {
		t.Errorf("expected loomio_test, got %s", cfg.PG.Database)
	}
	if cfg.PG.Database == "loomio_development" {
		t.Error("test config should NOT use development database")
	}

	// Verify other test-specific settings
	if cfg.PG.MaxConns != 5 {
		t.Errorf("expected 5, got %d", cfg.PG.MaxConns)
	}
	if cfg.Server.Port != 8081 {
		t.Errorf("expected 8081, got %d", cfg.Server.Port)
	}
	if cfg.Session.Duration != time.Hour {
		t.Errorf("expected 1h, got %v", cfg.Session.Duration)
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("expected warn, got %s", cfg.Logging.Level)
	}
}

// T013: Test for YAML file loading.
func Test_Load_YAMLFile(t *testing.T) {
	clearLlmioEnvVars(t) // Clear env vars so file values are used

	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
pg:
  host: testhost
  port: 5433
  user_admin: testuser
  pass_admin: testpass
  database: testdb
server:
  port: 9000
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify YAML values override defaults
	if cfg.PG.Host != "testhost" {
		t.Errorf("expected testhost, got %s", cfg.PG.Host)
	}
	if cfg.PG.Port != 5433 {
		t.Errorf("expected 5433, got %d", cfg.PG.Port)
	}
	if cfg.PG.UserAdmin != "testuser" {
		t.Errorf("expected testuser, got %s", cfg.PG.UserAdmin)
	}
	if cfg.PG.PassAdmin != "testpass" {
		t.Errorf("expected testpass, got %s", cfg.PG.PassAdmin)
	}
	if cfg.PG.Database != "testdb" {
		t.Errorf("expected testdb, got %s", cfg.PG.Database)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected 9000, got %d", cfg.Server.Port)
	}

	// Verify defaults still apply for unspecified values
	if cfg.PG.SSLMode != "disable" {
		t.Errorf("expected disable (default), got %s", cfg.PG.SSLMode)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected json (default), got %s", cfg.Logging.Format)
	}
}

// T058: Test that invalid YAML returns error with file information.
func Test_Load_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tmpDir := t.TempDir()
	configPath := tmpDir + "/invalid.yaml"

	invalidYAML := `
pg:
  host: localhost
  port: invalid_not_a_number
  user testuser  # Missing colon - syntax error
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	// Error should contain file path or indicate YAML parsing issue
	errStr := err.Error()
	if !containsAny(errStr, []string{"config", "yaml", "unmarshal", "reading"}) {
		t.Errorf("error should indicate config/yaml issue, got: %s", errStr)
	}
}

// containsAny checks if s contains any of the substrings (case-insensitive).
func containsAny(s string, substrings []string) bool {
	sLower := strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(sLower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// T007: Test for PGConfig DSN methods.
func Test_PGConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   PGConfig
		method   string // "admin", "app", or "migration"
		expected string
	}{
		{
			name: "admin without password",
			config: PGConfig{
				Host:      "localhost",
				Port:      5432,
				Database:  "testdb",
				SSLMode:   "disable",
				UserAdmin: "postgres",
			},
			method:   "admin",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable",
		},
		{
			name: "admin with password",
			config: PGConfig{
				Host:      "db.example.com",
				Port:      5433,
				Database:  "proddb",
				SSLMode:   "require",
				UserAdmin: "admin",
				PassAdmin: "secret",
			},
			method:   "admin",
			expected: "host=db.example.com port=5433 user=admin dbname=proddb sslmode=require password='secret'",
		},
		{
			name: "app credentials",
			config: PGConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				SSLMode:  "disable",
				UserApp:  "loomio_app",
				PassApp:  "apppass",
			},
			method:   "app",
			expected: "host=localhost port=5432 user=loomio_app dbname=testdb sslmode=disable password='apppass'",
		},
		{
			name: "migration credentials",
			config: PGConfig{
				Host:          "localhost",
				Port:          5432,
				Database:      "testdb",
				SSLMode:       "disable",
				UserMigration: "loomio_migration",
				PassMigration: "migratepass",
			},
			method:   "migration",
			expected: "host=localhost port=5432 user=loomio_migration dbname=testdb sslmode=disable password='migratepass'",
		},
		{
			name: "migration falls back to admin",
			config: PGConfig{
				Host:      "localhost",
				Port:      5432,
				Database:  "testdb",
				SSLMode:   "disable",
				UserAdmin: "postgres",
				PassAdmin: "adminpass",
				// UserMigration not set
			},
			method:   "migration",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='adminpass'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.method {
			case "admin":
				got = tt.config.AdminDSN()
			case "app":
				got = tt.config.AppDSN()
			case "migration":
				got = tt.config.MigrationDSN()
			}
			if got != tt.expected {
				t.Errorf("%sDSN() = %q, want %q", tt.method, got, tt.expected)
			}
		})
	}
}

// T096: Test the full priority chain: CLI > env > file > defaults.
func Test_Load_FullPriorityChain(t *testing.T) {
	clearLlmioEnvVars(t) // Clear all env vars first

	// Create a temporary YAML config file with specific values
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
server:
  port: 8000
pg:
  database: loomio_from_file
  host: filehost
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Set environment variable to override file (after clearing)
	t.Setenv("LLMIO_PG_DATABASE", "loomio_from_env")

	// Create a viper instance with CLI flag to override env
	v := NewViper()

	// Simulate CLI flag override (highest priority)
	v.Set("server.port", 9999)

	cfg, err := LoadWithViper(v, configPath)
	if err != nil {
		t.Fatalf("LoadWithViper failed: %v", err)
	}

	// Verify priority chain:
	// 1. CLI (server.port) should be 9999 (not 8000 from file)
	if cfg.Server.Port != 9999 {
		t.Errorf("expected server.port=9999 (CLI override), got %d", cfg.Server.Port)
	}

	// 2. Env (pg.database) should override file value
	if cfg.PG.Database != "loomio_from_env" {
		t.Errorf("expected pg.database='loomio_from_env' (env override), got %s", cfg.PG.Database)
	}

	// 3. File (pg.host) should override default
	if cfg.PG.Host != "filehost" {
		t.Errorf("expected pg.host='filehost' (file), got %s", cfg.PG.Host)
	}

	// 4. Default (pg.port) should apply when nothing else specified
	if cfg.PG.Port != 5432 {
		t.Errorf("expected pg.port=5432 (default), got %d", cfg.PG.Port)
	}
}

// T097: Test that environment variables override config file values.
func Test_Load_EnvOverridesConfigFile(t *testing.T) {
	clearLlmioEnvVars(t) // Clear all env vars first

	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
pg:
  host: config-file-host
  database: config-file-db
server:
  port: 7000
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Set env vars to override config file (after clearing)
	t.Setenv("LLMIO_PG_HOST", "env-host")
	t.Setenv("LLMIO_SERVER_PORT", "7777")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env should override file
	if cfg.PG.Host != "env-host" {
		t.Errorf("expected pg.host='env-host' (env override), got %s", cfg.PG.Host)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("expected server.port=7777 (env override), got %d", cfg.Server.Port)
	}

	// File value should still apply where env not set
	if cfg.PG.Database != "config-file-db" {
		t.Errorf("expected pg.database='config-file-db' (file), got %s", cfg.PG.Database)
	}
}

// T099: Test PGConfig validation catches invalid values.
// Uses go-playground/validator; tests check for field name in error.
func Test_PGConfig_Validate(t *testing.T) {
	validConfig := PGConfig{
		Host:              "localhost",
		Port:              5432,
		Database:          "testdb",
		SSLMode:           "disable",
		MaxConns:          25,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}

	// Valid config should pass
	if err := validation.Validate(validConfig); err != nil {
		t.Errorf("valid config should pass validation, got: %v", err)
	}

	tests := []struct {
		name      string
		modify    func(*PGConfig)
		wantField string // field name expected in validation error
	}{
		{
			name:      "empty host",
			modify:    func(c *PGConfig) { c.Host = "" },
			wantField: "Host",
		},
		{
			name:      "invalid port zero",
			modify:    func(c *PGConfig) { c.Port = 0 },
			wantField: "Port",
		},
		{
			name:      "invalid port negative",
			modify:    func(c *PGConfig) { c.Port = -1 },
			wantField: "Port",
		},
		{
			name:      "invalid port too high",
			modify:    func(c *PGConfig) { c.Port = 65536 },
			wantField: "Port",
		},
		{
			name:      "empty database",
			modify:    func(c *PGConfig) { c.Database = "" },
			wantField: "Database",
		},
		{
			name:      "invalid sslmode",
			modify:    func(c *PGConfig) { c.SSLMode = "invalid" },
			wantField: "SSLMode",
		},
		{
			name:      "max_conns zero",
			modify:    func(c *PGConfig) { c.MaxConns = 0 },
			wantField: "MaxConns",
		},
		{
			name:      "min_conns negative",
			modify:    func(c *PGConfig) { c.MinConns = -1 },
			wantField: "MinConns",
		},
		{
			name:      "min_conns exceeds max_conns",
			modify:    func(c *PGConfig) { c.MinConns = 30; c.MaxConns = 25 },
			wantField: "MinConns",
		},
		{
			name:      "max_conn_lifetime zero",
			modify:    func(c *PGConfig) { c.MaxConnLifetime = 0 },
			wantField: "MaxConnLifetime",
		},
		{
			name:      "max_conn_idle_time negative",
			modify:    func(c *PGConfig) { c.MaxConnIdleTime = -time.Second },
			wantField: "MaxConnIdleTime",
		},
		{
			name:      "health_check_period zero",
			modify:    func(c *PGConfig) { c.HealthCheckPeriod = 0 },
			wantField: "HealthCheckPeriod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig // copy
			tt.modify(&cfg)
			err := validation.Validate(cfg)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error should reference field %q, got: %v", tt.wantField, err)
			}
		})
	}
}

// T100: Test ServerConfig validation catches invalid values.
func Test_ServerConfig_Validate(t *testing.T) {
	validConfig := ServerConfig{
		Port:         8080,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := validation.Validate(validConfig); err != nil {
		t.Errorf("valid config should pass validation, got: %v", err)
	}

	tests := []struct {
		name      string
		modify    func(*ServerConfig)
		wantField string
	}{
		{
			name:      "invalid port zero",
			modify:    func(c *ServerConfig) { c.Port = 0 },
			wantField: "Port",
		},
		{
			name:      "read_timeout zero",
			modify:    func(c *ServerConfig) { c.ReadTimeout = 0 },
			wantField: "ReadTimeout",
		},
		{
			name:      "write_timeout negative",
			modify:    func(c *ServerConfig) { c.WriteTimeout = -time.Second },
			wantField: "WriteTimeout",
		},
		{
			name:      "idle_timeout zero",
			modify:    func(c *ServerConfig) { c.IdleTimeout = 0 },
			wantField: "IdleTimeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig
			tt.modify(&cfg)
			err := validation.Validate(cfg)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error should reference field %q, got: %v", tt.wantField, err)
			}
		})
	}
}

// T101: Test SessionConfig validation catches invalid values.
func Test_SessionConfig_Validate(t *testing.T) {
	validConfig := SessionConfig{
		Duration:        168 * time.Hour,
		CleanupInterval: 10 * time.Minute,
	}

	if err := validation.Validate(validConfig); err != nil {
		t.Errorf("valid config should pass validation, got: %v", err)
	}

	tests := []struct {
		name      string
		modify    func(*SessionConfig)
		wantField string
	}{
		{
			name:      "duration zero",
			modify:    func(c *SessionConfig) { c.Duration = 0 },
			wantField: "Duration",
		},
		{
			name:      "cleanup_interval negative",
			modify:    func(c *SessionConfig) { c.CleanupInterval = -time.Minute },
			wantField: "CleanupInterval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig
			tt.modify(&cfg)
			err := validation.Validate(cfg)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error should reference field %q, got: %v", tt.wantField, err)
			}
		})
	}
}

// T102: Test LoggingConfig validation catches invalid values.
func Test_LoggingConfig_Validate(t *testing.T) {
	validConfig := LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	if err := validation.Validate(validConfig); err != nil {
		t.Errorf("valid config should pass validation, got: %v", err)
	}

	tests := []struct {
		name      string
		modify    func(*LoggingConfig)
		wantField string
	}{
		{
			name:      "invalid level",
			modify:    func(c *LoggingConfig) { c.Level = "invalid" },
			wantField: "Level",
		},
		{
			name:      "invalid format",
			modify:    func(c *LoggingConfig) { c.Format = "invalid" },
			wantField: "Format",
		},
		{
			name:      "empty output",
			modify:    func(c *LoggingConfig) { c.Output = "" },
			wantField: "Output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig
			tt.modify(&cfg)
			err := validation.Validate(cfg)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error should reference field %q, got: %v", tt.wantField, err)
			}
		})
	}
}

// T105: Test SSLMode.Valid() for all known modes.
func Test_SSLMode_Valid(t *testing.T) {
	validModes := []SSLMode{
		SSLModeDisable,
		SSLModeAllow,
		SSLModePrefer,
		SSLModeRequire,
		SSLModeVerifyCA,
		SSLModeVerifyFull,
	}

	for _, mode := range validModes {
		if !mode.Valid() {
			t.Errorf("SSLMode %q should be valid", mode)
		}
	}

	invalidModes := []SSLMode{"invalid", "DISABLE", "ssl", ""}
	for _, mode := range invalidModes {
		if mode.Valid() {
			t.Errorf("SSLMode %q should be invalid", mode)
		}
	}
}

// T106: Test LogLevel.Valid() for all known levels.
func Test_LogLevel_Valid(t *testing.T) {
	validLevels := []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
	}

	for _, level := range validLevels {
		if !level.Valid() {
			t.Errorf("LogLevel %q should be valid", level)
		}
	}

	invalidLevels := []LogLevel{"invalid", "INFO", "warning", ""}
	for _, level := range invalidLevels {
		if level.Valid() {
			t.Errorf("LogLevel %q should be invalid", level)
		}
	}
}

// T107: Test LogFormat.Valid() for all known formats.
func Test_LogFormat_Valid(t *testing.T) {
	validFormats := []LogFormat{
		LogFormatJSON,
		LogFormatText,
	}

	for _, format := range validFormats {
		if !format.Valid() {
			t.Errorf("LogFormat %q should be valid", format)
		}
	}

	invalidFormats := []LogFormat{"invalid", "JSON", "xml", ""}
	for _, format := range invalidFormats {
		if format.Valid() {
			t.Errorf("LogFormat %q should be invalid", format)
		}
	}
}

// T103/T104: Test that Load() fails with invalid config values.
func Test_Load_ValidationFailure(t *testing.T) {
	clearLlmioEnvVars(t) // Clear env vars so invalid file values are used

	// Create a config file with invalid values
	tmpDir := t.TempDir()
	configPath := tmpDir + "/invalid_config.yaml"

	yamlContent := `
pg:
  host: localhost
  port: 0  # Invalid: must be 1-65535
  database: testdb
  sslmode: disable
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid port, got nil")
	}

	if !strings.Contains(err.Error(), "validation") {
		t.Errorf("error should mention validation, got: %v", err)
	}
	// Validator uses field name "Port" (title case)
	if !strings.Contains(err.Error(), "Port") {
		t.Errorf("error should mention Port, got: %v", err)
	}
}

// T138: Test that type coercion failures in YAML cause unmarshal errors.
// For example, setting port to a non-numeric string should fail.
func Test_Load_UnmarshalError(t *testing.T) {
	// Create a temporary file with type mismatch (string for integer field)
	tmpDir := t.TempDir()
	configPath := tmpDir + "/type_error.yaml"

	// Duration fields expect duration strings, but we'll provide invalid format
	invalidYAML := `
pg:
  host: localhost
  port: 5432
  database: testdb
  sslmode: disable
  max_conn_lifetime: "not_a_duration"  # Invalid duration format
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid duration format, got nil")
	}

	// Error should indicate unmarshal/decode issue
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "unmarshal") && !strings.Contains(errStr, "cannot parse") && !strings.Contains(errStr, "decode") {
		t.Errorf("error should indicate unmarshal issue, got: %s", err)
	}
}

// T064: Test for DSN with special characters in password.
func Test_PGConfig_DSN_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "password with spaces",
			password: "my secret password",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='my secret password'",
		},
		{
			name:     "password with single quote",
			password: "pass'word",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='pass\\'word'",
		},
		{
			name:     "password with equals sign",
			password: "pass=word",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='pass=word'",
		},
		{
			name:     "password with backslash",
			password: "pass\\word",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='pass\\\\word'",
		},
		{
			name:     "password with multiple special chars",
			password: "p@ss'w=rd\\123",
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable password='p@ss\\'w=rd\\\\123'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PGConfig{
				Host:      "localhost",
				Port:      5432,
				Database:  "testdb",
				SSLMode:   "disable",
				UserAdmin: "postgres",
				PassAdmin: tt.password,
			}
			got := cfg.AdminDSN()
			if got != tt.expected {
				t.Errorf("AdminDSN() = %q, want %q", got, tt.expected)
			}
		})
	}
}
