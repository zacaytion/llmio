# Research: Configuration System

**Feature**: 002-config-system
**Date**: 2026-02-02

## Technology Decisions

### 1. Configuration Library: Viper

**Decision**: Use `github.com/spf13/viper` for configuration management

**Rationale**:
- Industry standard for Go configuration (30k+ GitHub stars)
- Native support for YAML, JSON, TOML, env vars, CLI flags
- Built-in priority merging (exactly what spec requires)
- Automatic env var binding with prefix support
- Works seamlessly with Cobra for CLI integration

**Alternatives Considered**:
- **Manual implementation**: Would duplicate Viper's tested logic; higher maintenance burden
- **envconfig**: Only handles env vars, no file support
- **koanf**: Good alternative but less ecosystem support, fewer examples

### 2. CLI Framework: Cobra

**Decision**: Use `github.com/spf13/cobra` for CLI structure

**Rationale**:
- De facto standard for Go CLI apps (30k+ GitHub stars)
- Native integration with Viper for flag binding
- Subcommand support needed for migrate (up/down/status)
- Automatic help generation
- Already an indirect dependency (via goose)

**Alternatives Considered**:
- **flag package**: No subcommand support, manual help text
- **urfave/cli**: Good but less Viper integration
- **kong**: Newer, less ecosystem support

### 3. Logging: log/slog

**Decision**: Use Go stdlib `log/slog` (Go 1.21+)

**Rationale**:
- Stdlib = no external dependency
- Structured logging with JSON support
- Level-based filtering built-in
- Handler interface for customization
- Constitution Principle V (Simplicity) - prefer stdlib

**Alternatives Considered**:
- **zerolog**: Faster but external dependency
- **zap**: More features but external dependency, complex API
- **logrus**: Older, maintenance mode

## Best Practices Research

### Viper Configuration Pattern

```go
func Load(configPath string) (*Config, error) {
    v := viper.New()

    // 1. Set defaults first
    setDefaults(v)

    // 2. Config file (optional)
    if configPath != "" {
        v.SetConfigFile(configPath)
    } else {
        v.SetConfigName("config")
        v.SetConfigType("yaml")
        v.AddConfigPath(".")
    }
    v.ReadInConfig() // Ignore error if not found

    // 3. Environment variables
    v.SetEnvPrefix("LOOMIO")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // 4. Unmarshal to struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### Cobra + Viper Integration

```go
var cfgFile string
var cfg *Config

var rootCmd = &cobra.Command{
    Use:   "server",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error
        cfg, err = config.Load(cfgFile)
        return err
    },
    RunE: runServer,
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

    // Bind flags to viper for priority override
    rootCmd.Flags().Int("port", 0, "server port")
    viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port"))
}
```

### slog Configuration Pattern

```go
func Setup(cfg LoggingConfig) error {
    var output io.Writer = os.Stdout
    if cfg.Output == "stderr" {
        output = os.Stderr
    } else if cfg.Output != "stdout" && cfg.Output != "" {
        f, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            slog.Warn("failed to open log file, falling back to stdout", "path", cfg.Output, "error", err)
            output = os.Stdout
        } else {
            output = f
        }
    }

    level := slog.LevelInfo
    switch cfg.Level {
    case "debug": level = slog.LevelDebug
    case "warn": level = slog.LevelWarn
    case "error": level = slog.LevelError
    }

    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: level}
    if cfg.Format == "json" {
        handler = slog.NewJSONHandler(output, opts)
    } else {
        handler = slog.NewTextHandler(output, opts)
    }

    slog.SetDefault(slog.New(handler))
    return nil
}
```

## Existing Codebase Patterns

### Current Config Locations

| Component | File | Line | Current Pattern |
|-----------|------|------|-----------------|
| DB config | `internal/db/pool.go` | 23-33 | `DefaultConfig()` + `getEnv()` |
| Server port | `cmd/server/main.go` | 27 | `getEnv("PORT", "8080")` |
| HTTP timeouts | `cmd/server/main.go` | 64-66 | Hardcoded constants |
| Session duration | `internal/auth/session.go` | 14 | `const SessionDuration` |
| Cleanup interval | `cmd/server/main.go` | 96 | Hardcoded in `time.NewTicker()` |
| Pool settings | `internal/db/pool.go` | 55-59 | Hardcoded in `NewPool()` |

### Legacy Environment Variables

These must continue working (FR-006):
- `DB_HOST` → `database.host`
- `DB_USER` → `database.user`
- `DB_PASSWORD` → `database.password`
- `DB_NAME` → `database.name`
- `DB_SSLMODE` → `database.sslmode`
- `PORT` → `server.port`

## Migration Strategy

1. **Create new config package** - No breaking changes
2. **Add NewPoolFromConfig()** - Keep old NewPool() for compatibility
3. **Add NewSessionStoreWithConfig()** - Keep old NewSessionStore()
4. **Update cmd/server** - Load config, pass to constructors
5. **Update cmd/migrate** - Same config loading
6. **Update logging** - Replace log with slog calls
7. **Remove old getEnv** - After all callers migrated

## Testing Approach

### Unit Tests for Config Loading

```go
func TestLoad_Defaults(t *testing.T) {
    cfg, err := Load("")
    require.NoError(t, err)
    assert.Equal(t, "localhost", cfg.Database.Host)
    assert.Equal(t, 5432, cfg.Database.Port)
    assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoad_EnvOverride(t *testing.T) {
    t.Setenv("LOOMIO_SERVER_PORT", "9000")
    cfg, err := Load("")
    require.NoError(t, err)
    assert.Equal(t, 9000, cfg.Server.Port)
}

func TestLoad_LegacyEnvVars(t *testing.T) {
    t.Setenv("DB_HOST", "db.example.com")
    cfg, err := Load("")
    require.NoError(t, err)
    assert.Equal(t, "db.example.com", cfg.Database.Host)
}
```

## References

- [Viper Documentation](https://github.com/spf13/viper)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Go slog Package](https://pkg.go.dev/log/slog)
- [12-Factor App Config](https://12factor.net/config)
