<!--
Sync Impact Report
==================
Version: (new) 1.0.0
Changes: Initial ratification - all sections new
Added Principles:
  - I. Test-Driven Development (NON-NEGOTIABLE)
  - II. API Compatible Evolution
  - III. Minimal Dependencies
  - IV. Type Safety First
  - V. Observability Built-In
  - VI. Incremental Delivery
  - VII. Simplicity (YAGNI)
Added Sections:
  - Technology Stack (approved dependencies)
  - Development Workflow (quality gates)
Template Consistency Check:
  - .specify/templates/plan-template.md: Constitution Check section exists (line 30-34)
  - .specify/templates/spec-template.md: Acceptance scenarios support testing
  - .specify/templates/tasks-template.md: TDD guidance present (line 84-85, 180-181)
Follow-up TODOs: None
-->

# llmio Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)

Tests MUST be written before implementation code. This principle cannot be bypassed.

- **Red Phase**: Write tests that fail. Tests MUST fail before any implementation begins.
- **Green Phase**: Write the minimal implementation to make tests pass. No extra features.
- **Refactor Phase**: Clean up code only after all tests pass. Do not add behavior.
- **Merge Gate**: No code merges without passing tests and coverage maintained or improved.
- **Verification**: Commit history MUST show test commits preceding implementation commits.

### II. API Compatible Evolution

Existing Loomio API contracts are the baseline. The Vue frontend must continue to function.

- Breaking changes MUST include: documented migration path, API version bump, deprecation period.
- New endpoints MAY extend but MUST NOT contradict existing behavior.
- Serializer output formats MUST match original Loomio contracts (see `orig/loomio/app/serializers/`).
- API versioning (v1, b1, b2, b3) MUST be preserved during migration.
- Frontend compatibility is verified via the original Vue app in `orig/loomio/vue/`.

### III. Minimal Dependencies

Prefer Go standard library over external libraries. Every dependency is a maintenance burden.

- External dependencies MUST provide significant value beyond convenience.
- Each new dependency requires explicit justification in code review.
- Approved core stack (per ADR-003): Chi, sqlc, pgx, River, nhooyr.io/websocket, go-redis/v9.
- Logging MUST use stdlib `log/slog` - no external logging frameworks.
- Rejected patterns: ORMs, heavy frameworks, dependencies that duplicate stdlib functionality.

### IV. Type Safety First

Leverage Go's type system and compile-time verification to prevent runtime errors.

- Database queries MUST use sqlc for compile-time SQL verification.
- Avoid `any` and type assertions where stronger types are possible.
- Database schema is the source of truth - models derive from it, not vice versa.
- Use custom types for domain concepts (e.g., `type UserID int64` not bare `int64`).
- Nil safety: prefer value types, use pointer types only when nil is meaningful.

### V. Observability Built-In

All production code MUST be debuggable and traceable without additional instrumentation.

- Structured logging via `log/slog` with consistent field names across services.
- All HTTP handlers MUST propagate request IDs for distributed tracing.
- Errors MUST include context: wrap errors with `fmt.Errorf("context: %w", err)`.
- Metrics required for: request latency, error rates, queue depths, database query times.
- Log levels: DEBUG for development, INFO for production operations, ERROR for failures.

### VI. Incremental Delivery

Ship working code early and often. Each module MUST be independently deployable and testable.

- Follow hybrid migration strategy (ADR-001): module-by-module, not big bang.
- Each deliverable MUST provide user value, not just infrastructure.
- Feature flags for gradual rollout when changes affect existing users.
- Rollback path MUST exist for every deployment.
- Priority order: low-risk modules first (records.js, bots.js), then read endpoints, then write endpoints.

### VII. Simplicity (YAGNI)

Start with the simplest solution that works. Complexity MUST be justified by concrete, current needs.

- Do not build for hypothetical future requirements.
- If a simpler alternative exists, use it until proven insufficient.
- Prefer explicit code over clever abstractions.
- Three lines of duplicated code are better than one premature abstraction.
- When in doubt, leave it out.

## Technology Stack

Approved dependencies and their purposes (per ADR-003):

| Purpose | Package | Rationale |
|---------|---------|-----------|
| HTTP Router | `github.com/go-chi/chi/v5` | Lightweight, stdlib-aligned, composable middleware |
| Database Driver | `github.com/jackc/pgx/v5` | Full PostgreSQL feature support |
| SQL Generation | `github.com/sqlc-dev/sqlc` | Compile-time type-safe queries |
| Background Jobs | `github.com/riverqueue/river` | PostgreSQL-backed, no Redis required |
| WebSocket | `nhooyr.io/websocket` | Modern, context-aware, minimal API |
| Redis Client | `github.com/redis/go-redis/v9` | Caching and pub/sub for real-time features |
| Logging | `log/slog` (stdlib) | Built-in, sufficient, no external deps |
| Testing | `github.com/stretchr/testify` | De facto standard assertions |

Adding dependencies outside this list requires constitution amendment.

## Development Workflow

### Quality Gates

1. **Pre-commit**: golangci-lint MUST pass (auto-run via lefthook)
2. **Commit Message**: MUST follow conventional commits format
3. **PR Review**: Reviewer MUST verify TDD compliance (test commits precede implementation)
4. **CI Pipeline**: All tests MUST pass, coverage MUST not decrease
5. **Merge**: Squash merge to main, deployment follows

### Code Review Checklist

- [ ] Tests written first and failing before implementation?
- [ ] API changes backward compatible or properly versioned?
- [ ] New dependencies justified and approved?
- [ ] Types used appropriately (no unnecessary `any`)?
- [ ] Observability in place (logging, tracing, error context)?
- [ ] Simplest solution that meets requirements?

## Governance

This constitution supersedes all other development practices for the llmio project.

### Amendment Process

1. Propose change via PR with rationale
2. Review by project maintainers
3. Approval requires explicit consensus
4. Version bump according to semver:
   - MAJOR: Principle removal or fundamental redefinition
   - MINOR: New principle or section added
   - PATCH: Clarification or wording improvement
5. Update `LAST_AMENDED_DATE` and `CONSTITUTION_VERSION`

### Compliance

- All PRs MUST be reviewed against this constitution
- Constitution Check in plan-template.md MUST be completed before implementation
- Violations MUST be documented in Complexity Tracking if justified
- Unjustified violations block merge

**Version**: 1.0.0 | **Ratified**: 2025-01-30 | **Last Amended**: 2025-01-30
