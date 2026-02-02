# Data Model: Configuration System

**Feature**: 002-config-system
**Date**: 2026-02-02

## Overview

This feature introduces configuration entities (Go structs) but no database changes. All configuration is loaded at application startup and held in memory.

## Config Struct Hierarchy

```go
// Config holds all application configuration.
// Loaded once at startup via config.Load().
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    Server   ServerConfig   `mapstructure:"server"`
    Session  SessionConfig  `mapstructure:"session"`
    Logging  LoggingConfig  `mapstructure:"logging"`
}
```

## Entity: DatabaseConfig

**Purpose**: PostgreSQL connection and pool settings

```go
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
```

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| Host | string | "localhost" | Non-empty |
| Port | int | 5432 | 1-65535 |
| User | string | "postgres" | Non-empty |
| Password | string | "" | None (can be empty for local dev) |
| Name | string | "loomio_development" | Non-empty |
| SSLMode | string | "disable" | One of: disable, require, verify-ca, verify-full |
| MaxConns | int32 | 25 | > 0 |
| MinConns | int32 | 2 | >= 0, <= MaxConns |
| MaxConnLifetime | duration | 1h | > 0 |
| MaxConnIdleTime | duration | 30m | > 0 |
| HealthCheckPeriod | duration | 1m | > 0 |

**Derived Method**:
```go
// DSN returns the PostgreSQL connection string.
func (c DatabaseConfig) DSN() string {
    dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Name, c.SSLMode)
    if c.Password != "" {
        dsn += fmt.Sprintf(" password=%s", c.Password)
    }
    return dsn
}
```

## Entity: ServerConfig

**Purpose**: HTTP server settings

```go
type ServerConfig struct {
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
    IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}
```

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| Port | int | 8080 | 1-65535 |
| ReadTimeout | duration | 15s | > 0 |
| WriteTimeout | duration | 15s | > 0 |
| IdleTimeout | duration | 60s | > 0 |

## Entity: SessionConfig

**Purpose**: User session management settings

```go
type SessionConfig struct {
    Duration        time.Duration `mapstructure:"duration"`
    CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}
```

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| Duration | duration | 168h (7 days) | > 0 |
| CleanupInterval | duration | 10m | > 0 |

## Entity: LoggingConfig

**Purpose**: Structured logging settings

```go
type LoggingConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
    Output string `mapstructure:"output"`
}
```

| Field | Type | Default | Validation |
|-------|------|---------|------------|
| Level | string | "info" | One of: debug, info, warn, error |
| Format | string | "json" | One of: text, json |
| Output | string | "stdout" | stdout, stderr, or valid file path |

## YAML Schema

```yaml
# config.yaml schema
database:
  host: string          # default: localhost
  port: integer         # default: 5432
  user: string          # default: postgres
  password: string      # default: ""
  name: string          # default: loomio_development
  sslmode: string       # default: disable
  max_conns: integer    # default: 25
  min_conns: integer    # default: 2
  max_conn_lifetime: duration  # default: 1h
  max_conn_idle_time: duration # default: 30m
  health_check_period: duration # default: 1m

server:
  port: integer         # default: 8080
  read_timeout: duration   # default: 15s
  write_timeout: duration  # default: 15s
  idle_timeout: duration   # default: 60s

session:
  duration: duration           # default: 168h
  cleanup_interval: duration   # default: 10m

logging:
  level: string         # default: info
  format: string        # default: json
  output: string        # default: stdout
```

## Relationships

```
Config (root)
├── DatabaseConfig (1:1) → used by db.NewPoolFromConfig()
├── ServerConfig (1:1) → used by http.Server in cmd/server
├── SessionConfig (1:1) → used by auth.NewSessionStoreWithConfig()
└── LoggingConfig (1:1) → used by logging.Setup()
```

## State Transitions

Configuration is immutable after load:

```
[No Config] → Load() → [Config Loaded] → [Application Running]
                ↑
                └── Error on invalid config (application exits)
```

## No Database Changes

This feature does not modify the PostgreSQL schema. All configuration is:
- Loaded from files/env/CLI at startup
- Stored in memory as Go structs
- Passed to constructors that need settings
