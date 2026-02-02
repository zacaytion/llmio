# Research: Local Development Workflow

**Feature**: 003-dev-workflow
**Date**: 2026-02-02

## Overview

This feature creates developer tooling (compose.yml + Makefile), not application code. Research focuses on best practices for the specific tools being configured.

## Podman Compose Configuration

### Decision: Use compose.yml (not docker-compose.yml)

**Rationale**: Podman Compose supports `compose.yml` as the default filename. Using this name makes the intent clear and avoids confusion with Docker-specific tooling.

**Alternatives considered**:
- `docker-compose.yml` - Works with Podman but implies Docker dependency
- `podman-compose.yml` - Non-standard, not recognized by default

### Decision: Named volumes for persistence

**Rationale**: Named volumes (`postgres_data`, `pgadmin_data`) persist data across container restarts without cluttering the repository with data directories.

**Alternatives considered**:
- Bind mounts to `./data/` - Pollutes repository, requires gitignore management
- Anonymous volumes - Data lost on `podman compose down`

### Decision: Health checks with depends_on condition

**Rationale**: Using `depends_on: postgres: condition: service_healthy` ensures PgAdmin only starts after PostgreSQL is accepting connections, preventing connection errors on startup.

**Alternatives considered**:
- Simple `depends_on` without condition - Race condition, PgAdmin may fail to connect
- External health check scripts - Over-engineering for local dev

## PgAdmin Pre-configuration

### Decision: servers.json volume mount

**Rationale**: PgAdmin reads `/pgadmin4/servers.json` on startup to auto-register database connections. Mounting a pre-configured file eliminates manual setup for every developer.

**Configuration**:
```json
{
  "Servers": {
    "1": {
      "Name": "loomio-dev",
      "Group": "Local",
      "Host": "postgres",
      "Port": 5432,
      "MaintenanceDB": "loomio_development",
      "Username": "postgres",
      "SSLMode": "prefer"
    }
  }
}
```

**Note**: Password is NOT stored in servers.json. PgAdmin prompts on first connection, then caches in its internal storage (persisted via named volume).

## Makefile Self-Documentation

### Decision: awk-based help target

**Rationale**: The `##` comment pattern with awk parsing is a well-established convention for self-documenting Makefiles. It works without additional dependencies and provides grouped, colored output.

**Pattern**:
```makefile
target: ## Description
	@command

help: ## Show help
	@awk 'BEGIN {FS = ":.*##"} ...' $(MAKEFILE_LIST)
```

**Alternatives considered**:
- Manual help text - Drifts out of sync with actual targets
- External documentation - Developers won't read it

### Decision: Proper prerequisites with directory targets

**Rationale**: Make's prerequisite system ensures directories exist before commands that need them, without redundant `mkdir -p` in every target.

**Pattern**:
```makefile
.var/coverage:
	mkdir -p .var/coverage

test: .var/coverage
	go test -coverprofile=.var/coverage/coverage.out ./...
```

## Environment Variables

### Decision: .env file with .env.example template

**Rationale**: Standard pattern for 12-factor apps. `.env.example` documents required variables; `.env` contains actual values and is gitignored.

**Default values in .env.example**:
```bash
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=loomio_development
PGADMIN_DEFAULT_EMAIL=admin@local.dev
PGADMIN_DEFAULT_PASSWORD=admin
```

**Note**: These are development-only credentials. Production uses different secrets management.

## Lint Output Strategy

### Decision: tee to terminal AND log file

**Rationale**: Developers need immediate feedback (terminal), but also a persistent log for reference or CI integration. Using `tee` satisfies both.

**Implementation**:
```makefile
lint: .var/log
	golangci-lint run ./... 2>&1 | tee .var/log/golangci-lint.log
```

**Note**: This overrides the default golangci-lint output config which writes only to the log file.

## Coverage Workflow

### Decision: Single coverage.out file (no timestamps)

**Rationale**: Simpler than timestamped files. The prerequisite system ensures coverage-view uses fresh data.

**Workflow**:
1. `make test` → generates `.var/coverage/coverage.out`
2. `make coverage-view` → requires coverage.out as prerequisite, generates HTML, opens browser

## Open Questions Resolved

No NEEDS CLARIFICATION items from Technical Context - all decisions made during brainstorming session.
