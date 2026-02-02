package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zacaytion/llmio/internal/validation"
)

// T005: Test for Config struct existence.
func TestConfigStructs(t *testing.T) {
	// Verify structs can be instantiated with expected fields
	cfg := Config{
		Database: DatabaseConfig{
			Host:              "localhost",
			Port:              5432,
			User:              "postgres",
			Password:          "secret",
			Name:              "testdb",
			SSLMode:           "disable",
			MaxConns:          25,
			MinConns:          2,
			MaxConnLifetime:   time.Hour,
			MaxConnIdleTime:   30 * time.Minute,
			HealthCheckPeriod: time.Minute,
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

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Database.Host)
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

// T006: Test for Load() with defaults.
//
//nolint:gocyclo // Test function validating many config defaults - complexity is intentional
func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Database defaults
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.Database.User)
	}
	if cfg.Database.Name != "loomio_development" {
		t.Errorf("expected loomio_development, got %s", cfg.Database.Name)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("expected disable, got %s", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxConns != 25 {
		t.Errorf("expected 25, got %d", cfg.Database.MaxConns)
	}
	if cfg.Database.MinConns != 2 {
		t.Errorf("expected 2, got %d", cfg.Database.MinConns)
	}
	if cfg.Database.MaxConnLifetime != time.Hour {
		t.Errorf("expected 1h, got %v", cfg.Database.MaxConnLifetime)
	}
	if cfg.Database.MaxConnIdleTime != 30*time.Minute {
		t.Errorf("expected 30m, got %v", cfg.Database.MaxConnIdleTime)
	}
	if cfg.Database.HealthCheckPeriod != time.Minute {
		t.Errorf("expected 1m, got %v", cfg.Database.HealthCheckPeriod)
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

// T033: Test for environment variable override (LOOMIO_*).
func TestLoad_EnvVarOverride(t *testing.T) {
	// Set environment variable
	t.Setenv("LOOMIO_SERVER_PORT", "9001")
	t.Setenv("LOOMIO_DATABASE_NAME", "loomio_from_env")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env var should override default
	if cfg.Server.Port != 9001 {
		t.Errorf("expected 9001 (from LOOMIO_SERVER_PORT), got %d", cfg.Server.Port)
	}
	if cfg.Database.Name != "loomio_from_env" {
		t.Errorf("expected loomio_from_env (from LOOMIO_DATABASE_NAME), got %s", cfg.Database.Name)
	}
}

// T028: Test for CLI flag override of config file value.
func TestLoadWithViper_CLIOverride(t *testing.T) {
	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
server:
  port: 8080
database:
  name: loomio_development
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
	if cfg.Database.Name != "loomio_development" {
		t.Errorf("expected loomio_development, got %s", cfg.Database.Name)
	}
}

// T023: Test for config.test.yaml loading different database name.
func TestLoad_TestConfig(t *testing.T) {
	// Create a temporary test config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.test.yaml"

	yamlContent := `
database:
  host: localhost
  port: 5432
  user: testuser
  password: testpass
  name: loomio_test
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
	if cfg.Database.Name != "loomio_test" {
		t.Errorf("expected loomio_test, got %s", cfg.Database.Name)
	}
	if cfg.Database.Name == "loomio_development" {
		t.Error("test config should NOT use development database")
	}

	// Verify other test-specific settings
	if cfg.Database.MaxConns != 5 {
		t.Errorf("expected 5, got %d", cfg.Database.MaxConns)
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
func TestLoad_YAMLFile(t *testing.T) {
	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
database:
  host: testhost
  port: 5433
  user: testuser
  password: testpass
  name: testdb
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
	if cfg.Database.Host != "testhost" {
		t.Errorf("expected testhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected 5433, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("expected testuser, got %s", cfg.Database.User)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("expected testpass, got %s", cfg.Database.Password)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("expected testdb, got %s", cfg.Database.Name)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected 9000, got %d", cfg.Server.Port)
	}

	// Verify defaults still apply for unspecified values
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("expected disable (default), got %s", cfg.Database.SSLMode)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected json (default), got %s", cfg.Logging.Format)
	}
}

// T058: Test that invalid YAML returns error with file information.
func TestLoad_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tmpDir := t.TempDir()
	configPath := tmpDir + "/invalid.yaml"

	invalidYAML := `
database:
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

// T007: Test for DatabaseConfig.DSN() method.
func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "without password",
			config: DatabaseConfig{
				Host:    "localhost",
				Port:    5432,
				User:    "postgres",
				Name:    "testdb",
				SSLMode: "disable",
			},
			expected: "host=localhost port=5432 user=postgres dbname=testdb sslmode=disable",
		},
		{
			name: "with password",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				Name:     "proddb",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5433 user=admin dbname=proddb sslmode=require password='secret'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.DSN()
			if got != tt.expected {
				t.Errorf("DSN() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// T096: Test the full priority chain: CLI > env > file > defaults.
func TestLoad_FullPriorityChain(t *testing.T) {
	// Create a temporary YAML config file with specific values
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
server:
  port: 8000
database:
  name: loomio_from_file
  host: filehost
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Set environment variable to override file
	t.Setenv("LOOMIO_DATABASE_NAME", "loomio_from_env")

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

	// 2. Env (database.name) should override file value
	if cfg.Database.Name != "loomio_from_env" {
		t.Errorf("expected database.name='loomio_from_env' (env override), got %s", cfg.Database.Name)
	}

	// 3. File (database.host) should override default
	if cfg.Database.Host != "filehost" {
		t.Errorf("expected database.host='filehost' (file), got %s", cfg.Database.Host)
	}

	// 4. Default (database.port) should apply when nothing else specified
	if cfg.Database.Port != 5432 {
		t.Errorf("expected database.port=5432 (default), got %d", cfg.Database.Port)
	}
}

// T097: Test that environment variables override config file values.
func TestLoad_EnvOverridesConfigFile(t *testing.T) {
	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	yamlContent := `
database:
  host: config-file-host
  name: config-file-db
server:
  port: 7000
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Set env vars to override config file
	t.Setenv("LOOMIO_DATABASE_HOST", "env-host")
	t.Setenv("LOOMIO_SERVER_PORT", "7777")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env should override file
	if cfg.Database.Host != "env-host" {
		t.Errorf("expected database.host='env-host' (env override), got %s", cfg.Database.Host)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("expected server.port=7777 (env override), got %d", cfg.Server.Port)
	}

	// File value should still apply where env not set
	if cfg.Database.Name != "config-file-db" {
		t.Errorf("expected database.name='config-file-db' (file), got %s", cfg.Database.Name)
	}
}

// T099: Test DatabaseConfig validation catches invalid values.
// Uses go-playground/validator; tests check for field name in error.
func TestDatabaseConfig_Validate(t *testing.T) {
	validConfig := DatabaseConfig{
		Host:              "localhost",
		Port:              5432,
		User:              "postgres",
		Name:              "testdb",
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
		modify    func(*DatabaseConfig)
		wantField string // field name expected in validation error
	}{
		{
			name:      "empty host",
			modify:    func(c *DatabaseConfig) { c.Host = "" },
			wantField: "Host",
		},
		{
			name:      "invalid port zero",
			modify:    func(c *DatabaseConfig) { c.Port = 0 },
			wantField: "Port",
		},
		{
			name:      "invalid port negative",
			modify:    func(c *DatabaseConfig) { c.Port = -1 },
			wantField: "Port",
		},
		{
			name:      "invalid port too high",
			modify:    func(c *DatabaseConfig) { c.Port = 65536 },
			wantField: "Port",
		},
		{
			name:      "empty user",
			modify:    func(c *DatabaseConfig) { c.User = "" },
			wantField: "User",
		},
		{
			name:      "empty name",
			modify:    func(c *DatabaseConfig) { c.Name = "" },
			wantField: "Name",
		},
		{
			name:      "invalid sslmode",
			modify:    func(c *DatabaseConfig) { c.SSLMode = "invalid" },
			wantField: "SSLMode",
		},
		{
			name:      "max_conns zero",
			modify:    func(c *DatabaseConfig) { c.MaxConns = 0 },
			wantField: "MaxConns",
		},
		{
			name:      "min_conns negative",
			modify:    func(c *DatabaseConfig) { c.MinConns = -1 },
			wantField: "MinConns",
		},
		{
			name:      "min_conns exceeds max_conns",
			modify:    func(c *DatabaseConfig) { c.MinConns = 30; c.MaxConns = 25 },
			wantField: "MinConns",
		},
		{
			name:      "max_conn_lifetime zero",
			modify:    func(c *DatabaseConfig) { c.MaxConnLifetime = 0 },
			wantField: "MaxConnLifetime",
		},
		{
			name:      "max_conn_idle_time negative",
			modify:    func(c *DatabaseConfig) { c.MaxConnIdleTime = -time.Second },
			wantField: "MaxConnIdleTime",
		},
		{
			name:      "health_check_period zero",
			modify:    func(c *DatabaseConfig) { c.HealthCheckPeriod = 0 },
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
func TestServerConfig_Validate(t *testing.T) {
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
func TestSessionConfig_Validate(t *testing.T) {
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
func TestLoggingConfig_Validate(t *testing.T) {
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
func TestSSLMode_Valid(t *testing.T) {
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
func TestLogLevel_Valid(t *testing.T) {
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
func TestLogFormat_Valid(t *testing.T) {
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
func TestLoad_ValidationFailure(t *testing.T) {
	// Create a config file with invalid values
	tmpDir := t.TempDir()
	configPath := tmpDir + "/invalid_config.yaml"

	yamlContent := `
database:
  host: localhost
  port: 0  # Invalid: must be 1-65535
  user: postgres
  name: testdb
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
func TestLoad_UnmarshalError(t *testing.T) {
	// Create a temporary file with type mismatch (string for integer field)
	tmpDir := t.TempDir()
	configPath := tmpDir + "/type_error.yaml"

	// Duration fields expect duration strings, but we'll provide invalid format
	invalidYAML := `
database:
  host: localhost
  port: 5432
  user: postgres
  name: testdb
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
func TestDatabaseConfig_DSN_SpecialChars(t *testing.T) {
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
			cfg := DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: tt.password,
				Name:     "testdb",
				SSLMode:  "disable",
			}
			got := cfg.DSN()
			if got != tt.expected {
				t.Errorf("DSN() = %q, want %q", got, tt.expected)
			}
		})
	}
}
