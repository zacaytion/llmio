# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **monorepo for rewriting Loomio** - a collaborative decision-making platform. The `orig/` directory contains the existing Rails 8 + Vue 3 codebase being analyzed; `discovery/` contains comprehensive specifications for the rewrite.

**Primary LLM Reference:** `discovery/loomio_rewrite_context.md` (~25K tokens) - read this first for complete context.

## Development Commands

### Go Backend (new rewrite)

```bash
go test ./... -v              # Run all tests
go run ./cmd/server           # Start API server (port 8080)
go run ./cmd/migrate up       # Run database migrations
go run ./cmd/migrate status   # Check migration status
golangci-lint run ./...       # Lint (output in .var/log/golangci-lint.log)
sqlc generate                 # Regenerate DB types from queries
```

### Go Gotchas

- golangci-lint v2 writes to `.var/log/golangci-lint.log` (see `.golangci.yml` output.formats.tab.path)
- golangci-lint autofix: use `golangci-lint run ./... --fix` to auto-fix gofmt/goimports issues
- Import ordering: stdlib first, blank line, then third-party (goimports enforces this)
- Config has `interface{}` → `any` rewrite rule; use `any` not `interface{}`
- errcheck requires `defer func() { _ = conn.Close() }()` not `defer conn.Close()`
- errcheck requires checked type assertions: `val, ok := x.(*Type)` not `val := x.(*Type)`
- Avoid naming local variables `api` when importing `internal/api` package
- Port 8080 typically occupied by `gvproxy` (Docker/Podman); use `PORT=8081` for Go server
- Viper requires `mapstructure` tags on config structs (not `yaml` or `json`)
- Viper durations: use `time.Duration` values directly in `SetDefault()`, not strings
- Viper file not found: use `if !errors.As(err, &viper.ConfigFileNotFoundError{})` to ignore missing files
- PostgreSQL DSN passwords: single-quote and escape (`\\` then `\'`) for special chars
- Dead code removal: Check test files (`*_test.go`) before removing package-level vars - tests may depend on them
- Import cycle: `internal/validation` can't import `internal/config` (config imports validation); duplicate switch logic is intentional
- Viper env prefix: Go app uses `LOOMIO_DATABASE_*` env vars (e.g., `LOOMIO_DATABASE_PASSWORD`), not `DB_*`

### Error Handling Patterns

- Use `db.IsNotFound(err)` to check for `pgx.ErrNoRows` (not just `err != nil`)
- DB errors → 500; NotFound → context-specific (401 for auth, 404 for resources)
- Crypto/rand failures should panic (not silent fallback) - security-critical
- Auth failures: log reason server-side, return generic "Invalid credentials" to client
- Always call `LogDBError(ctx, "OperationName", err)` before returning 500 for DB errors
- Validator registration: Use `mustRegister()` pattern that panics on failure (startup-time safety)

### Logging Patterns

- Use `logging.SetupDefaultWithCleanup()` to get cleanup function for file handles
- Cleanup function is no-op for stdout/stderr, closes file for file output

### Security Patterns

- Timing attacks: dummy password hash must be valid Argon2id (verification detects invalid formats)
- Generate dummy hash at `init()` with `auth.HashPassword("placeholder")` for consistent timing
- Tests must use same pattern: `testDummyHash` generated via `auth.HashPassword()` in test `init()`

### Goose Migrations

- Goose treats ALL `*.sql` files in `migrations/` as migrations based on numeric prefix
- pgTap schema tests belong in `tests/pgtap/`, NOT in migrations directory

### Makefile & Containers

- `make up/down/logs` - Container lifecycle; `make server/migrate ARGS="..."` - Run binaries with args
- `make clean` - Remove everything (volumes, binaries, caches); `make clean-go-{build,test,mod,fuzz,all}` - Go caches
- Container env vars (`POSTGRES_*`) differ from Go app env vars (`LOOMIO_DATABASE_*`); both documented in `.env.example`
- PostgreSQL 18 volume mount: use `/var/lib/postgresql` (not `/var/lib/postgresql/data`) for `pg_upgrade --link` compatibility
- Podman Compose volumes: prefixed with project directory name (e.g., `llmio_postgres_data`)

### PostgreSQL Access

**Via container (recommended):**
- Use `psql-18` (not `psql`) with `-h localhost -U postgres` and `PGPASSWORD=postgres`
- Example: `PGPASSWORD=postgres psql-18 -h localhost -U postgres -d loomio_development -c "SELECT 1;"`
- Credentials match `.env.example` defaults: `POSTGRES_USER=postgres`, `POSTGRES_PASSWORD=postgres`

**Go app connection:**
- Uses `LOOMIO_DATABASE_*` env vars (Viper prefix), not `POSTGRES_*`
- Example: `LOOMIO_DATABASE_USER=postgres LOOMIO_DATABASE_PASSWORD=postgres make server`

### Version Management (mise)

This project uses [mise](https://mise.jdx.dev/) for tool version management with experimental monorepo mode.

```bash
mise install              # Install all tools (Go, Ruby 3.4.7, Node 22, pnpm)
```

### Backend (Rails API) - from `orig/loomio/`

```bash
bundle install            # Install Ruby dependencies
rails s                   # Start Rails server (port 3000)
rails c                   # Rails console
bundle exec rspec         # Run all tests
bundle exec rspec spec/path/to/file_spec.rb      # Run single test file
bundle exec rspec spec/path/to/file_spec.rb:42   # Run specific line
```

### Frontend (Vue 3 SPA) - from `orig/loomio/`

```bash
mise run frontend-serve   # Hot-reload dev server (port 5173 → proxies to 3000)
mise run frontend-build   # Production build → public/client3
mise run frontend-test    # E2E tests (Nightwatch)
mise run pnpm-install     # Install frontend dependencies
```

Or directly from `orig/loomio/vue/`:
```bash
pnpm install && pnpm run serve
```

### WebSocket Server - from `orig/loomio_channel_server/`

```bash
mise run serve-ws         # Start Socket.io server
mise run serve-hocuspocus # Start collaborative editing server
```

### Database Setup

```bash
createdb loomio_development
cd orig/loomio && rake db:setup
```

## Git Workflow

Conventional commits enforced via commitlint. Valid types:
`build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `plan`, `refactor`, `revert`, `style`, `test`

Pre-commit hooks run multiple linters in parallel via Lefthook.

**Pre-commit gotcha**: Linter runs on ALL Go files, not just staged ones. Pre-existing errors anywhere block commits.

### Additional Linters

```bash
make lint-all         # Run all linters
make lint-files       # File naming conventions (ls-lint)
make lint-md          # Markdown style (markdownlint)
make lint-makefile    # Makefile quality (checkmake)
make lint-migrations  # SQL migration safety (squawk)
```

### Linter Gotchas

- squawk: npm package broken; use `pipx:squawk-cli` via mise, invoke with `mise exec -- squawk`
- ls-lint: Regex rules match filename only (not extension); use `regex:^\d{3}_[a-z_]+$` not `^\d{3}_[a-z_]+\.sql$`
- checkmake: `minphony`/`phonydeclared` rules have false positives with file-target patterns; disable in `checkmake.ini`
- markdownlint: Exclude generated/legacy dirs in `.markdownlint-cli2.yaml` globs (discovery, specs, .specify)

## Architecture

### Core Domain (see `discovery/specifications/` for details)

| Concept | Description |
|---------|-------------|
| **Group** | Organization with members and permission flags |
| **Discussion** | Conversation thread (can belong to group or be "direct") |
| **Poll** | Decision tool (proposal, ranked choice, dot vote, etc.) |
| **Stance** | User's vote/position on a poll |
| **Event** | Activity record driving timelines, notifications, webhooks |

### Key Patterns

1. **Service Layer** - All mutations flow through `*Service` classes (`PollService.create`, `DiscussionService.update`)
2. **Event Sourcing** - Actions create Event records that publish to Redis → Socket.io → Vue clients
3. **Permission Flags** - Groups have 12 `members_can_*` boolean flags controlling capabilities
4. **Client-side ORM** - LokiJS mirrors Rails models with 28 record interfaces

### Request Flow

```
Vue SPA → REST /api/v1/* → Controller → authorize!(CanCanCan) → *Service.action() → Event.publish!
                                                                                        ↓
Vue SPA ← Socket.io (records) ← Redis pub/sub ← PublishEventWorker
```

### Directory Structure

```
discovery/                 # Rewrite specifications (read first!)
  ├── loomio_rewrite_context.md  # Executive summary (25K tokens)
  ├── specifications/            # 26 detailed spec files
  ├── openapi/                   # API documentation (~204 endpoints)
  └── schemas/                   # Database and request/response schemas

orig/loomio/               # Rails 8 API + Vue 3 frontend
  ├── app/
  │   ├── controllers/api/v1/    # REST endpoints (~30 controllers)
  │   ├── models/                # ActiveRecord + concerns
  │   ├── services/              # Business logic (*Service classes)
  │   └── workers/               # Sidekiq jobs
  └── vue/src/
      ├── components/            # 217 Vue components
      └── shared/
          ├── services/          # 35 services (records.js, session.js)
          ├── models/            # 31 client-side models
          └── interfaces/        # 28 LokiJS record interfaces

orig/loomio_channel_server/  # Node.js WebSocket server
```

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Rails 8 API-only, Ruby 3.4.7 |
| Frontend | Vue 3, Vite, Vuetify |
| Database | PostgreSQL with pg_search |
| Queue | Sidekiq + Redis |
| Real-time | Socket.io, Hocuspocus + Yjs |
| Client State | LokiJS in-memory DB |
| Testing | RSpec (backend), Nightwatch (E2E) |

## Spec-First Development Workflow

This project uses speckit commands for spec-first TDD development. See `docs/spec-first-tdd-workflow.md` for full guide.

### Speckit Commands (in order)

| Command | Purpose |
|---------|---------|
| `/speckit.specify <description>` | Create feature spec from description |
| `/speckit.clarify` | Fill gaps in spec via questions |
| `/speckit.plan` | Generate technical design from spec |
| `/speckit.tasks` | Create ordered task list with TDD |
| `/speckit.analyze` | Validate spec ↔ plan ↔ tasks consistency |
| `/speckit.implement` | Execute tasks with TDD discipline |

### Key Files

- `.specify/memory/constitution.md` - Project principles (TDD mandatory, spec-first API, security-first)
- `.specify/templates/` - Templates for spec, plan, tasks
- `specs/N-feature-name/` - Feature artifacts (spec.md, plan.md, tasks.md)

### Speckit Workflow Notes

- Code review fixes: Add as new phase in `specs/$feature/tasks.md`, not separate `tasks-fixes.md`
- Task IDs must be unique: Continue numbering from last task (e.g., T055+ if T054 exists)

### Superpowers Integration

- `/superpowers:brainstorming` - FIRST step before any feature work; explores requirements via Q&A
- `/superpowers:writing-plans` - Creates detailed bite-sized task plans from specs
- `/superpowers:test-driven-development` - During implementation, enforce Red-Green-Refactor
- `/superpowers:verification-before-completion` - Before claiming done, verify tests pass

### Design Document Storage

- **Always use `/speckit.specify`** to create feature specs and designs - NOT `docs/plans/`
- All design artifacts go in `specs/N-feature-name/` via speckit workflow

### Feature Branch Setup

```bash
.specify/scripts/bash/create-new-feature.sh --json --number N --short-name "feature-name" "Description"
```

### Discovery Reference (check when designing new features)

- `discovery/openapi/paths/` - Existing API endpoint patterns (auth.yaml, users.yaml, etc.)
- `discovery/schemas/request_schemas/` - Request parameter schemas by controller
- `discovery/schemas/response_schemas/` - Response serializer schemas
- `discovery/schema_dump.sql` - Full PostgreSQL schema (users table at line 2278)

## Active Technologies
- Go 1.25+ with Huma web framework + Huma, pgx/v5, sqlc, golang.org/x/crypto/argon2 (001-user-auth)
- goose/v3 for database migrations (001-user-auth)
- PostgreSQL 18 for users; in-memory Go map for sessions (MVP) (001-user-auth)
- Go 1.25+ + Viper (config), Cobra (CLI), log/slog (stdlib logging) (002-config-system)
- PostgreSQL 18 (existing), YAML config files (new) (002-config-system)
- N/A (shell scripts, Makefile, YAML configuration) + Podman, Podman Compose, golangci-lint, goimports (003-dev-workflow)
- PostgreSQL 18 (container), Redis 8 (container) (003-dev-workflow)

## Validation Patterns

**Use `go-playground/validator/v10`** for all struct validation (config, API requests, domain entities).

```go
// Add validate tags to structs
type Config struct {
    Port int `validate:"required,min=1,max=65535"`
    Host string `validate:"required"`
}

// Use internal/validation package for shared validator instance
import "github.com/zacaytion/llmio/internal/validation"

func Load() (*Config, error) {
    cfg := &Config{...}
    if err := validation.Validate(cfg); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    return cfg, nil
}
```

**Custom validators** for domain-specific rules (e.g., `sslmode`, `loglevel`) are registered in `internal/validation/validator.go`.

**Cross-field validation**: Use `ltefield`/`gtefield` tags (e.g., `validate:"ltefield=MaxConns"` ensures MinConns ≤ MaxConns).

## Recent Changes
- 003-dev-workflow: Added N/A (shell scripts, Makefile, YAML configuration) + Podman, Podman Compose, golangci-lint, goimports
- 001-user-auth: Added Go 1.25+ with Huma web framework + Huma, pgx/v5, sqlc, golang.org/x/crypto/argon2
