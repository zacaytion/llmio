package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/zacaytion/llmio/internal/validation"
)

// Config holds all application configuration.
type Config struct {
	PG      PGConfig      `mapstructure:"pg"`
	Server  ServerConfig  `mapstructure:"server"`
	Session SessionConfig `mapstructure:"session"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// Validate checks if all configuration sections have valid values.
// Uses go-playground/validator for declarative struct validation.
func (c Config) Validate() error {
	return validation.Validate(c)
}

// PGConfig holds PostgreSQL connection settings.
// Environment variables use LLMIO_PG_* prefix (e.g., LLMIO_PG_HOST, LLMIO_PG_PORT).
type PGConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Database string `mapstructure:"database" validate:"required"`
	SSLMode  string `mapstructure:"sslmode" validate:"required,sslmode"`

	// Connection pool settings
	MaxConns          int32         `mapstructure:"max_conns" validate:"required,min=1"`
	MinConns          int32         `mapstructure:"min_conns" validate:"min=0,ltefield=MaxConns"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime" validate:"required,gt=0"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time" validate:"required,gt=0"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period" validate:"required,gt=0"`

	// Three-role credential model:
	// - Admin: superuser for initial setup (postgres)
	// - Migration: DDL privileges for schema changes (loomio_migration)
	// - App: DML privileges for runtime queries (loomio_app)

	// Admin credentials (superuser, used for role creation in migrations)
	UserAdmin string `mapstructure:"user_admin"`
	PassAdmin string `mapstructure:"pass_admin"`

	// Migration credentials (used by cmd/migrate for DDL operations)
	UserMigration string `mapstructure:"user_migration"`
	PassMigration string `mapstructure:"pass_migration"`

	// App credentials (used by cmd/server for runtime queries)
	UserApp string `mapstructure:"user_app"`
	PassApp string `mapstructure:"pass_app"`
}

// AdminDSN returns the PostgreSQL connection string for admin operations.
// Uses UserAdmin/PassAdmin credentials (superuser for role creation).
func (c PGConfig) AdminDSN() string {
	return c.buildDSN(c.UserAdmin, c.PassAdmin)
}

// MigrationDSN returns the PostgreSQL connection string for migrations.
// Uses UserMigration/PassMigration credentials (DDL privileges).
// Falls back to AdminDSN if migration credentials not set.
func (c PGConfig) MigrationDSN() string {
	if c.UserMigration == "" {
		return c.AdminDSN()
	}
	return c.buildDSN(c.UserMigration, c.PassMigration)
}

// AppDSN returns the PostgreSQL connection string for runtime queries.
// Uses UserApp/PassApp credentials (DML privileges only).
func (c PGConfig) AppDSN() string {
	return c.buildDSN(c.UserApp, c.PassApp)
}

// buildDSN constructs a PostgreSQL connection string with the given credentials.
func (c PGConfig) buildDSN(user, password string) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, user, c.Database, c.SSLMode)
	if password != "" {
		dsn += fmt.Sprintf(" password='%s'", escapePassword(password))
	}
	return dsn
}

// SSLMode represents valid PostgreSQL SSL modes.
// Note: This type is defined for documentation and type-safe usage in code,
// but PGConfig uses string for SSLMode to simplify Viper unmarshaling.
// Validation is handled by the "sslmode" custom validator in internal/validation.
type SSLMode string

// Valid SSL modes for PostgreSQL connections.
const (
	SSLModeDisable    SSLMode = "disable"
	SSLModeAllow      SSLMode = "allow"
	SSLModePrefer     SSLMode = "prefer"
	SSLModeRequire    SSLMode = "require"
	SSLModeVerifyCA   SSLMode = "verify-ca"
	SSLModeVerifyFull SSLMode = "verify-full"
)

// Valid returns true if the SSLMode is a recognized PostgreSQL SSL mode.
func (m SSLMode) Valid() bool {
	switch m {
	case SSLModeDisable, SSLModeAllow, SSLModePrefer, SSLModeRequire, SSLModeVerifyCA, SSLModeVerifyFull:
		return true
	default:
		return false
	}
}

// String returns the string representation of the SSLMode.
func (m SSLMode) String() string {
	return string(m)
}

// escapePassword escapes special characters in passwords for PostgreSQL DSN.
// Backslashes and single quotes must be escaped within single-quoted values.
func escapePassword(password string) string {
	// Escape backslashes first (must come before single quote escaping)
	escaped := strings.ReplaceAll(password, `\`, `\\`)
	// Escape single quotes
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return escaped
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,gt=0"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,gt=0"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required,gt=0"`
}

// SessionConfig holds session management settings.
type SessionConfig struct {
	Duration        time.Duration `mapstructure:"duration" validate:"required,gt=0"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval" validate:"required,gt=0"`
}

// LogLevel represents valid log levels.
// Note: This type is defined for documentation and type-safe usage in code,
// but LoggingConfig uses string for Level to simplify Viper unmarshaling.
// Validation is handled by the "loglevel" custom validator in internal/validation.
type LogLevel string

// Valid log levels.
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Valid returns true if the LogLevel is a recognized level.
func (l LogLevel) Valid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	default:
		return false
	}
}

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	return string(l)
}

// LogFormat represents valid log formats.
// Note: This type is defined for documentation and type-safe usage in code,
// but LoggingConfig uses string for Format to simplify Viper unmarshaling.
// Validation is handled by the "logformat" custom validator in internal/validation.
type LogFormat string

// Valid log formats.
const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// Valid returns true if the LogFormat is a recognized format.
func (f LogFormat) Valid() bool {
	switch f {
	case LogFormatJSON, LogFormatText:
		return true
	default:
		return false
	}
}

// String returns the string representation of the LogFormat.
func (f LogFormat) String() string {
	return string(f)
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,loglevel"`
	Format string `mapstructure:"format" validate:"required,logformat"`
	Output string `mapstructure:"output" validate:"required"`
}

// NewViper creates a new Viper instance with defaults set.
// Use this when you need to bind CLI flags before loading config.
func NewViper() *viper.Viper {
	v := viper.New()
	setDefaults(v)
	return v
}

// Load reads configuration from file, environment, and sets defaults.
// Priority: CLI flags > env vars > config file > defaults.
func Load(configPath string) (*Config, error) {
	v := NewViper()
	return LoadWithViper(v, configPath)
}

// LoadWithViper reads configuration using a pre-configured Viper instance.
// This allows CLI flags to be bound to Viper before loading.
// Priority: explicitly set values > env vars > config file > defaults.
func LoadWithViper(v *viper.Viper, configPath string) (*Config, error) {
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Read config file (optional - won't fail if not found)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Only return error for real errors (not "file not found")
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("LLMIO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// PostgreSQL defaults (env vars: LLMIO_PG_*)
	v.SetDefault("pg.host", "localhost")
	v.SetDefault("pg.port", 5432)
	v.SetDefault("pg.database", "loomio_development")
	v.SetDefault("pg.sslmode", "disable")

	// Connection pool defaults
	v.SetDefault("pg.max_conns", 25)
	v.SetDefault("pg.min_conns", 2)
	v.SetDefault("pg.max_conn_lifetime", time.Hour)
	v.SetDefault("pg.max_conn_idle_time", 30*time.Minute)
	v.SetDefault("pg.health_check_period", time.Minute)

	// Three-role credential model defaults
	// Admin: superuser for initial setup and role creation
	v.SetDefault("pg.user_admin", "postgres")
	v.SetDefault("pg.pass_admin", "")
	// Migration: DDL privileges for schema changes (used by cmd/migrate)
	v.SetDefault("pg.user_migration", "")
	v.SetDefault("pg.pass_migration", "")
	// App: DML privileges for runtime queries (used by cmd/server)
	v.SetDefault("pg.user_app", "")
	v.SetDefault("pg.pass_app", "")

	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)

	// Session defaults
	v.SetDefault("session.duration", 168*time.Hour)
	v.SetDefault("session.cleanup_interval", 10*time.Minute)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
}
