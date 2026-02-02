package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Session  SessionConfig  `mapstructure:"session"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// Validate checks if all configuration sections have valid values.
// Returns an error describing all validation failures across all sections.
func (c Config) Validate() error {
	var errs []error

	if err := c.Database.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Session.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	User              string        `mapstructure:"user"`
	Password          string        `mapstructure:"password"`
	Name              string        `mapstructure:"name"`
	SSLMode           string        `mapstructure:"sslmode"`
	MaxConns          int32         `mapstructure:"max_conns"`
	MinConns          int32         `mapstructure:"min_conns"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// SSLMode represents valid PostgreSQL SSL modes.
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

// Validate checks if the DatabaseConfig has valid values.
// Returns an error describing all validation failures, or nil if valid.
func (c DatabaseConfig) Validate() error {
	var errs []string

	if c.Host == "" {
		errs = append(errs, "database.host cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Sprintf("database.port must be 1-65535, got %d", c.Port))
	}
	if c.User == "" {
		errs = append(errs, "database.user cannot be empty")
	}
	if c.Name == "" {
		errs = append(errs, "database.name cannot be empty")
	}
	if !SSLMode(c.SSLMode).Valid() {
		errs = append(errs, fmt.Sprintf("database.sslmode must be one of: disable, allow, prefer, require, verify-ca, verify-full; got %q", c.SSLMode))
	}
	if c.MaxConns < 1 {
		errs = append(errs, fmt.Sprintf("database.max_conns must be positive, got %d", c.MaxConns))
	}
	if c.MinConns < 0 {
		errs = append(errs, fmt.Sprintf("database.min_conns cannot be negative, got %d", c.MinConns))
	}
	if c.MinConns > c.MaxConns {
		errs = append(errs, fmt.Sprintf("database.min_conns (%d) cannot exceed max_conns (%d)", c.MinConns, c.MaxConns))
	}
	if c.MaxConnLifetime <= 0 {
		errs = append(errs, fmt.Sprintf("database.max_conn_lifetime must be positive, got %v", c.MaxConnLifetime))
	}
	if c.MaxConnIdleTime <= 0 {
		errs = append(errs, fmt.Sprintf("database.max_conn_idle_time must be positive, got %v", c.MaxConnIdleTime))
	}
	if c.HealthCheckPeriod <= 0 {
		errs = append(errs, fmt.Sprintf("database.health_check_period must be positive, got %v", c.HealthCheckPeriod))
	}

	if len(errs) > 0 {
		return fmt.Errorf("database config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// DSN returns the PostgreSQL connection string.
// Passwords are single-quoted and escaped to handle special characters
// (spaces, single quotes, backslashes, equals signs).
func (c DatabaseConfig) DSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.SSLMode)
	if c.Password != "" {
		dsn += fmt.Sprintf(" password='%s'", escapePassword(c.Password))
	}
	return dsn
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
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Validate checks if the ServerConfig has valid values.
func (c ServerConfig) Validate() error {
	var errs []string

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Sprintf("server.port must be 1-65535, got %d", c.Port))
	}
	if c.ReadTimeout <= 0 {
		errs = append(errs, fmt.Sprintf("server.read_timeout must be positive, got %v", c.ReadTimeout))
	}
	if c.WriteTimeout <= 0 {
		errs = append(errs, fmt.Sprintf("server.write_timeout must be positive, got %v", c.WriteTimeout))
	}
	if c.IdleTimeout <= 0 {
		errs = append(errs, fmt.Sprintf("server.idle_timeout must be positive, got %v", c.IdleTimeout))
	}

	if len(errs) > 0 {
		return fmt.Errorf("server config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// SessionConfig holds session management settings.
type SessionConfig struct {
	Duration        time.Duration `mapstructure:"duration"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// Validate checks if the SessionConfig has valid values.
func (c SessionConfig) Validate() error {
	var errs []string

	if c.Duration <= 0 {
		errs = append(errs, fmt.Sprintf("session.duration must be positive, got %v", c.Duration))
	}
	if c.CleanupInterval <= 0 {
		errs = append(errs, fmt.Sprintf("session.cleanup_interval must be positive, got %v", c.CleanupInterval))
	}

	if len(errs) > 0 {
		return fmt.Errorf("session config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// LogLevel represents valid log levels.
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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Validate checks if the LoggingConfig has valid values.
func (c LoggingConfig) Validate() error {
	var errs []string

	if !LogLevel(c.Level).Valid() {
		errs = append(errs, fmt.Sprintf("logging.level must be one of: debug, info, warn, error; got %q", c.Level))
	}
	if !LogFormat(c.Format).Valid() {
		errs = append(errs, fmt.Sprintf("logging.format must be one of: json, text; got %q", c.Format))
	}
	// Output can be "stdout", "stderr", or any file path - we don't validate file existence here
	if c.Output == "" {
		errs = append(errs, "logging.output cannot be empty")
	}

	if len(errs) > 0 {
		return fmt.Errorf("logging config validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
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
	// Config file settings
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

	// Environment variable binding
	v.SetEnvPrefix("LOOMIO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into struct
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
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "loomio_development")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 2)
	v.SetDefault("database.max_conn_lifetime", time.Hour)
	v.SetDefault("database.max_conn_idle_time", 30*time.Minute)
	v.SetDefault("database.health_check_period", time.Minute)

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
