# Implementation Plan: Configuration System

**Branch**: `002-config-system` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-config-system/spec.md`

## Summary

Add a centralized configuration system using Viper + Cobra + log/slog to eliminate hardcoded values, reduce dev friction, and enable separate test database setup. Configuration loads from YAML files, environment variables, and CLI flags with priority: CLI > env > file > defaults.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: Viper (config), Cobra (CLI), log/slog (stdlib logging)
**Storage**: PostgreSQL 18 (existing), YAML config files (new)
**Testing**: Go stdlib `testing`, table-driven tests
**Target Platform**: Linux/macOS server
**Project Type**: Single Go module with multiple commands
**Performance Goals**: N/A (configuration loading is startup-only)
**Constraints**: None
**Scale/Scope**: 2 commands (server, migrate), ~10 config files touched

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | ✅ PASS | Config loading will have unit tests |
| II. Huma-First API Design | ✅ N/A | No new API endpoints |
| III. Security-First | ✅ PASS | Credentials in .gitignore, not committed |
| IV. Full-Stack Type Safety | ✅ PASS | Typed config structs |
| V. Simplicity & YAGNI | ⚠️ CHECK | Viper/Cobra adds dependencies - justified by FR-001 through FR-015 |

**Gate Result**: PASS (Viper/Cobra justified by explicit spec requirements for YAML + CLI flags)

## Project Structure

### Documentation (this feature)

```text
specs/002-config-system/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
├── server/main.go       # MODIFY: Add Cobra, load config
└── migrate/main.go      # MODIFY: Add Cobra, load config

internal/
├── config/              # NEW: Configuration package
│   └── config.go        # Config structs + Viper loading
├── logging/             # NEW: Logging package
│   └── logging.go       # slog setup from config
├── api/
│   ├── logging.go       # MODIFY: Use slog instead of log
│   └── middleware.go    # MODIFY: Use slog instead of log
├── auth/
│   └── session.go       # MODIFY: Configurable duration
└── db/
    └── pool.go          # MODIFY: NewPoolFromConfig()

config.yaml              # NEW: Development config
config.example.yaml      # NEW: Example config (committed)
config.test.yaml         # NEW: Test config
.gitignore               # MODIFY: Add config.yaml, config.local.yaml
```

**Structure Decision**: Single Go module, config package under `internal/`, logging as separate package per spec clarification.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Viper dependency | YAML + env + CLI priority merging | Manual implementation would duplicate Viper's tested logic |
| Cobra dependency | CLI flags for all config values | flag package lacks subcommands needed for migrate |

## Design Decisions

### Config Struct Hierarchy

```go
type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    Session  SessionConfig
    Logging  LoggingConfig
}
```

### Environment Variable Mapping

| Config Path | Environment Variable |
|-------------|---------------------|
| database.host | LOOMIO_DATABASE_HOST |
| database.port | LOOMIO_DATABASE_PORT |
| database.user | LOOMIO_DATABASE_USER |
| database.password | LOOMIO_DATABASE_PASSWORD |
| database.name | LOOMIO_DATABASE_NAME |
| database.sslmode | LOOMIO_DATABASE_SSLMODE |
| server.port | LOOMIO_SERVER_PORT |

### CLI Flag Naming

Server command flags:
- `--config` - config file path
- `--port` - server port
- `--http-read-timeout`, `--http-write-timeout`, `--http-idle-timeout`
- `--db-host`, `--db-port`, `--db-user`, `--db-password`, `--db-name`, `--db-sslmode`
- `--db-max-conns`, `--db-min-conns`, `--db-max-conn-lifetime`, `--db-max-conn-idle-time`, `--db-health-check-period`
- `--session-duration`, `--session-cleanup-interval`
- `--log-level`, `--log-format`, `--log-output`

### Default Values

| Setting | Default | Source |
|---------|---------|--------|
| database.host | localhost | Existing behavior |
| database.port | 5432 | Existing behavior |
| database.user | postgres | Existing behavior |
| database.name | loomio_development | Existing behavior |
| database.sslmode | disable | Existing behavior |
| database.max_conns | 25 | Existing pool.go:55 |
| database.min_conns | 2 | Existing pool.go:56 |
| database.max_conn_lifetime | 1h | Existing pool.go:57 |
| database.max_conn_idle_time | 30m | Existing pool.go:58 |
| database.health_check_period | 1m | Existing pool.go:59 |
| server.port | 8080 | Existing main.go:27 |
| server.read_timeout | 15s | Existing main.go:64 |
| server.write_timeout | 15s | Existing main.go:65 |
| server.idle_timeout | 60s | Existing main.go:66 |
| session.duration | 168h | Existing session.go:14 |
| session.cleanup_interval | 10m | Existing main.go:96 |
| logging.level | info | New default |
| logging.format | json | Per spec FR-012 |
| logging.output | stdout | New default |

## Files to Create

| File | Purpose | Lines (est) |
|------|---------|-------------|
| `internal/config/config.go` | Config structs, Load(), defaults, env binding | ~200 |
| `internal/config/config_test.go` | Unit tests for config loading | ~150 |
| `internal/logging/logging.go` | slog Setup() from LoggingConfig | ~80 |
| `internal/logging/logging_test.go` | Unit tests for logging setup | ~60 |
| `config.example.yaml` | Example config with placeholders | ~40 |
| `config.test.yaml` | Test environment config | ~30 |

## Files to Modify

| File | Changes | Impact |
|------|---------|--------|
| `cmd/server/main.go` | Cobra root command, config loading, remove getEnv | Major rewrite |
| `cmd/migrate/main.go` | Cobra commands (up/down/status), config loading | Major rewrite |
| `internal/db/pool.go` | Add NewPoolFromConfig(), keep NewPool() for compat | Add function |
| `internal/auth/session.go` | Add NewSessionStoreWithConfig(), store duration | Add function |
| `internal/api/logging.go` | Replace log.Printf with slog calls | All functions |
| `internal/api/middleware.go` | Replace log.Printf with slog.Info | LoggingMiddleware |
| `go.mod` | Add viper dependency | 1 line |
| `.gitignore` | Add config.yaml, config.local.yaml | 2 lines |

## Testing Strategy

1. **Config Loading Tests** (`internal/config/config_test.go`):
   - Test defaults applied when no config
   - Test YAML file loading
   - Test env var override
   - Test CLI flag override (via viper binding)
   - Test legacy env var compatibility
   - Test invalid YAML error handling

2. **Logging Setup Tests** (`internal/logging/logging_test.go`):
   - Test JSON format output
   - Test text format output
   - Test log level filtering
   - Test file output fallback to stdout

3. **Integration Tests**:
   - Server starts with config file
   - Server starts with env vars only
   - Server starts with no config (defaults)
   - Migrate uses same config as server

## Verification Checklist

- [ ] `go test ./...` passes
- [ ] `golangci-lint run ./...` passes
- [ ] Server starts with `go run ./cmd/server`
- [ ] Server starts with `go run ./cmd/server --config config.example.yaml`
- [ ] Server starts with `LOOMIO_SERVER_PORT=9000 go run ./cmd/server`
- [ ] Server starts with `DB_HOST=localhost go run ./cmd/server` (legacy)
- [ ] Migrate works: `go run ./cmd/migrate --config config.test.yaml status`
- [ ] JSON logs appear when log format is json
- [ ] Test database isolation verified
