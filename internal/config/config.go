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

// SessionConfig holds session management settings.
type SessionConfig struct {
	Duration        time.Duration `mapstructure:"duration"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
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
