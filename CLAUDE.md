# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**llmio** is a Go rewrite of Loomio, a collaborative decision-making platform. This is a module-by-module migration from Ruby on Rails to Go, maintaining API compatibility with the existing Vue.js frontend. The original Loomio codebase is included as git submodules in `orig/` for reference.

## Development Commands

```bash
# Setup environment (installs Go, pnpm, Ruby, lefthook)
mise trust && mise install

# Run linter (auto-runs on pre-commit via lefthook)
golangci-lint run --fix --timeout 2m

# Run tests
go test ./...

# Run single test
go test -run TestName ./path/to/package

# Install git hooks
lefthook install
```

## Project Structure

```
llmio/
├── orig/                      # Original Loomio repos (git submodules)
│   ├── loomio/               # Rails 8.0 main app - API contract reference
│   ├── loomio_channel_server/ # Node.js real-time server
│   └── loomio-deploy/        # Docker deployment configs
├── discovery/                 # Third-party investigation deliverables
├── research/                  # Architecture investigation docs
│   └── investigation/        # Themed docs (models, api, database, etc.)
│   └── follow_up/            # Follow-up investigation questions
│   └── synthesis/            # Synthesized findings
│   └── final_follow_up/      # Final questions for third party
│   └── final_synthesis/      # Implementation-ready findings
├── .specify/memory/          # Project constitution and governance
├── .golangci.yml             # Linter config
├── .lefthook.yml             # Git hooks config
└── mise.toml                 # Dev environment config
```

## Core Principles (from Constitution)

1. **TDD Required**: Tests MUST be written before implementation. Commit history must show test commits preceding implementation.

2. **API Compatibility**: Vue frontend must continue working. Serializer output must match `orig/loomio/app/serializers/`.

3. **Minimal Dependencies**: Use Go stdlib over external packages. Approved stack only:
   - Router: `chi/v5`
   - Database: `pgx/v5` + `sqlc`
   - Jobs: `river`
   - WebSocket: `nhooyr.io/websocket`
   - Redis: `go-redis/v9`
   - Logging: `log/slog` (stdlib)
   - Testing: `testify`

4. **Type Safety**: Use `sqlc` for compile-time SQL. Avoid `any`. Use custom types for domain concepts (e.g., `type UserID int64`).

5. **Observability**: Structured logging via `log/slog`. Propagate request IDs. Wrap errors with context.

## Commit Messages

Use conventional commits format. Allowed types:
`build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `plan`, `refactor`, `revert`, `style`, `test`

Validated via commitlint on commit-msg hook.

## Key Reference Documents

- **API contracts**: `orig/loomio/app/serializers/` and `orig/loomio/config/routes.rb`
- **Database schema**: `research/schema_dump.sql` (57 tables)
- **Architecture overview**: `research/investigation/overview.md`
- **Project governance**: `.specify/memory/constitution.md`
- **Implementation synthesis**: `research/final_synthesis/` (OAuth, permissions, rate limiting, webhooks, etc.)
- **Open questions**: `research/final_follow_up/` (pending third-party clarification)
- **Comparison summaries**: `research/FOLLOW_UP_SUMMARY.md` and `research/SYNTHESIS_SUMMARY.md` (entry points for LLMs)

## Architecture Notes

Loomio is an event-driven system with:
- 9 poll types: `proposal`, `poll`, `count`, `score`, `ranked_choice`, `meeting`, `dot_vote`, `check`, `question`
- 42 event kinds (14 webhook-eligible)
- 38 background job workers
- Redis pub/sub for real-time features
- PostgreSQL with citext, hstore, pgcrypto extensions

Key Rails patterns to understand when investigating:
- `app/models/concerns/searchable.rb` - Full-text search via pg_search
- `app/services/message_channel_service.rb` - Redis pub/sub for real-time
- `app/workers/` - 38 Sidekiq workers (see `research/final_synthesis/infrastructure_jobs.md`)
- `app/models/concerns/` - Shared behaviors (HasRichText, HasEvents, Reactable, etc.)

## Known Issues in Original Codebase (Must Fix in Go)

- **OAuth CSRF**: Missing state parameter validation in `identities_controller.rb` - implement proper state handling
- **Rate limit responses**: `ThrottleService` returns 500 instead of 429 - return proper status with Retry-After header
- **Redis TTL**: Rate limit keys may not expire - always set TTL on Redis keys
