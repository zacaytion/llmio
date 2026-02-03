# Implementation Plan: Local Development Workflow

**Branch**: `003-dev-workflow` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-dev-workflow/spec.md`

## Summary

Create local development infrastructure with two components: (1) a `compose.yml` for Podman Compose defining PostgreSQL 18, Redis 8, PgAdmin4, and Mailpit services, and (2) a self-documenting Makefile providing build, run, test, lint, and container management targets. This is developer tooling, not application code.

## Technical Context

**Language/Version**: N/A (shell scripts, Makefile, YAML configuration)
**Primary Dependencies**: Podman, Podman Compose, golangci-lint, goimports
**Storage**: PostgreSQL 18 (container), Redis 8 (container)
**Testing**: Manual verification of Makefile targets
**Target Platform**: macOS (developer workstations)
**Project Type**: Configuration files (not source code)
**Performance Goals**: Services healthy within 60 seconds of `make up`
**Constraints**: Standard ports (5432, 6379, 5050, 8025, 1025) must be available
**Scale/Scope**: Single developer local environment

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Applies? | Status | Notes |
|-----------|----------|--------|-------|
| I. Test-First Development | Partial | PASS | No Go code to test; Makefile targets verified manually |
| II. Huma-First API Design | No | N/A | No API endpoints in this feature |
| III. Security-First | Yes | PASS | `.env` gitignored; credentials not in compose.yml |
| IV. Full-Stack Type Safety | No | N/A | No application code |
| V. Simplicity & YAGNI | Yes | PASS | Minimal targets, no over-engineering |

**Technology Constraints Check**:
- Uses PostgreSQL 18 ✓
- Uses Redis for cache/pubsub (when needed) ✓
- No deviations from mandatory stack

**Gate Result**: PASS - All applicable principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/003-dev-workflow/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
# Files created by this feature
compose.yml              # Podman Compose services
Makefile                 # Development commands
.env.example             # Environment template (committed)
.env                     # Local environment (gitignored)
bin/
└── .gitkeep             # Keep directory, ignore binaries
docker/
└── pgadmin/
    └── servers.json     # PgAdmin pre-configured connection
```

**Structure Decision**: Configuration files at repository root. No source code directories needed for this feature.

## Complexity Tracking

No violations to justify - this feature follows all applicable constitution principles.
