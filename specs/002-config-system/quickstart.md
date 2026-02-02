# Quickstart: Configuration System

**Feature**: 002-config-system
**Date**: 2026-02-02

## Quick Start

### 1. Run with Defaults

No configuration needed - sensible defaults work out of the box:

```bash
go run ./cmd/server
```

Server starts on port 8080, connects to `localhost:5432/loomio_development`.

### 2. Run with Config File

Create `config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: z
  password: password
  name: loomio_development

server:
  port: 9000

logging:
  level: debug
```

```bash
go run ./cmd/server --config config.yaml
```

### 3. Run with Environment Variables

```bash
# New namespaced format
export LOOMIO_DATABASE_NAME=loomio_production
export LOOMIO_SERVER_PORT=8080
go run ./cmd/server

# Legacy format (backward compatible)
export DB_HOST=db.example.com
export DB_NAME=loomio_production
export PORT=8080
go run ./cmd/server
```

### 4. Run with CLI Flags

```bash
go run ./cmd/server --port 9000 --log-level debug --db-name loomio_test
```

### 5. Priority Order

CLI flags > Environment variables > Config file > Defaults

```bash
# Config file says port 8080, but CLI wins
go run ./cmd/server --config config.yaml --port 9000
# Server runs on port 9000
```

## Running Tests with Test Database

### Setup Test Database

```bash
PGPASSWORD=password psql-18 -h localhost -U z -c "CREATE DATABASE loomio_test;"
```

### Run Migrations

```bash
go run ./cmd/migrate --config config.test.yaml up
```

### Run Tests

```bash
# Use test config for integration tests
go test ./... -v
```

## Database Migration Commands

```bash
# Check migration status
go run ./cmd/migrate status

# Show current database version
go run ./cmd/migrate version

# Run pending migrations
go run ./cmd/migrate up

# Rollback last migration
go run ./cmd/migrate down

# Create a new migration file
go run ./cmd/migrate create add_users_table

# With explicit config
go run ./cmd/migrate --config config.test.yaml status
```

**Note**: This migrate tool provides a simplified subset of goose commands. For advanced operations like `up-by-one`, `up-to`, `down-to`, `redo`, or `reset`, use goose CLI directly:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir migrations postgres "user=z dbname=loomio_development" up-by-one
```

## Example Config Files

### config.example.yaml (committed to repo)

```yaml
# Example configuration - copy to config.yaml and customize
# DO NOT commit config.yaml with real credentials

database:
  host: localhost
  port: 5432
  user: postgres
  password: ""  # Set via env var LOOMIO_DATABASE_PASSWORD for security
  name: loomio_development
  sslmode: disable
  max_conns: 25
  min_conns: 2

server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

session:
  duration: 168h  # 7 days
  cleanup_interval: 10m

logging:
  level: info     # debug, info, warn, error
  format: json    # json, text
  output: stdout  # stdout, stderr, or file path
```

### config.test.yaml (test environment)

```yaml
database:
  host: localhost
  port: 5432
  user: z
  password: password
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
```

## All Available Flags

### Server Command

```
go run ./cmd/server [flags]

Flags:
      --config string                    Config file path
      --port int                         Server port (default 8080)
      --http-read-timeout duration       HTTP read timeout (default 15s)
      --http-write-timeout duration      HTTP write timeout (default 15s)
      --http-idle-timeout duration       HTTP idle timeout (default 60s)
      --db-host string                   Database host (default "localhost")
      --db-port int                      Database port (default 5432)
      --db-user string                   Database user (default "postgres")
      --db-password string               Database password
      --db-name string                   Database name (default "loomio_development")
      --db-sslmode string                Database SSL mode (default "disable")
      --db-max-conns int                 Max connections (default 25)
      --db-min-conns int                 Min connections (default 2)
      --db-max-conn-lifetime duration    Max connection lifetime (default 1h)
      --db-max-conn-idle-time duration   Max connection idle time (default 30m)
      --db-health-check-period duration  Health check period (default 1m)
      --session-duration duration        Session duration (default 168h)
      --session-cleanup-interval duration Session cleanup interval (default 10m)
      --log-level string                 Log level: debug, info, warn, error (default "info")
      --log-format string                Log format: json, text (default "json")
      --log-output string                Log output: stdout, stderr, or file path (default "stdout")
  -h, --help                             Help for server
```

### Migrate Command

```
go run ./cmd/migrate [command] [flags]

Commands:
  up        Run all pending migrations
  down      Rollback last migration
  status    Show migration status

Flags:
      --config string     Config file path
      --db-host string    Database host (default "localhost")
      --db-port int       Database port (default 5432)
      --db-user string    Database user (default "postgres")
      --db-password string Database password
      --db-name string    Database name (default "loomio_development")
      --db-sslmode string Database SSL mode (default "disable")
  -h, --help              Help for migrate
```

## Environment Variables Reference

| Config Path | New Env Var | Legacy Env Var |
|-------------|-------------|----------------|
| database.host | LOOMIO_DATABASE_HOST | DB_HOST |
| database.port | LOOMIO_DATABASE_PORT | DB_PORT |
| database.user | LOOMIO_DATABASE_USER | DB_USER |
| database.password | LOOMIO_DATABASE_PASSWORD | DB_PASSWORD |
| database.name | LOOMIO_DATABASE_NAME | DB_NAME |
| database.sslmode | LOOMIO_DATABASE_SSLMODE | DB_SSLMODE |
| server.port | LOOMIO_SERVER_PORT | PORT |
| logging.level | LOOMIO_LOGGING_LEVEL | - |
| logging.format | LOOMIO_LOGGING_FORMAT | - |
| logging.output | LOOMIO_LOGGING_OUTPUT | - |

## Troubleshooting

### Config File Not Found

```
Error: config file not found: config.yaml
```

Config files are optional. If not specified, defaults are used. To use a config file:

```bash
go run ./cmd/server --config /path/to/config.yaml
```

### Invalid YAML

```
Error: error reading config file: yaml: line 5: could not find expected ':'
```

Check YAML syntax. Common issues:
- Missing colons after keys
- Incorrect indentation (use spaces, not tabs)
- Unquoted special characters

### Database Connection Failed

```
Error: failed to connect to database: dial tcp: connection refused
```

Check:
1. PostgreSQL is running: `pg_isready -h localhost`
2. Credentials are correct
3. Database exists: `psql -l`

### Log File Not Writable

```
WARN: failed to open log file, falling back to stdout
```

Application continues with stdout logging. Fix the file path or permissions if file logging is needed.
