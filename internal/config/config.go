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
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Session  SessionConfig  `mapstructure:"session"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// Validate checks if all configuration sections have valid values.
// Uses go-playground/validator for declarative struct validation.
func (c Config) Validate() error {
	return validation.Validate(c)
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host              string        `mapstructure:"host" validate:"required"`
	Port              int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	User              string        `mapstructure:"user" validate:"required"`
	Password          string        `mapstructure:"password"`
	Name              string        `mapstructure:"name" validate:"required"`
	SSLMode           string        `mapstructure:"sslmode" validate:"required,sslmode"`
	MaxConns          int32         `mapstructure:"max_conns" validate:"required,min=1"`
	MinConns          int32         `mapstructure:"min_conns" validate:"min=0,ltefield=MaxConns"`
	MaxConnLifetime   time.Duration `mapstructure:"max_conn_lifetime" validate:"required,gt=0"`
	MaxConnIdleTime   time.Duration `mapstructure:"max_conn_idle_time" validate:"required,gt=0"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period" validate:"required,gt=0"`
}

// SSLMode represents valid PostgreSQL SSL modes.
// Note: This type is defined for documentation and type-safe usage in code,
// but DatabaseConfig uses string for SSLMode to simplify Viper unmarshaling.
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

// DSN returns the PostgreSQL connection string.
// Passwords are single-quoted to handle special characters (spaces, equals signs).
// Backslashes and single quotes within the password are escaped.
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

	v.SetEnvPrefix("LOOMIO")
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
