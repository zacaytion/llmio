<!--
SYNC IMPACT REPORT
==================
Version change: 0.0.0 → 1.0.0 (initial ratification)

Modified principles: N/A (new constitution)

Added sections:
- I. Test-First Development (NON-NEGOTIABLE)
- II. Spec-First API Design
- III. Security-First
- IV. Full-Stack Type Safety
- V. Simplicity & YAGNI
- Technology Constraints (new section)
- Development Workflow (new section)
- Governance

Removed sections: N/A (new constitution)

Templates requiring updates:
- .specify/templates/plan-template.md: ✅ compatible (Constitution Check section exists)
- .specify/templates/spec-template.md: ✅ compatible (Requirements section aligns)
- .specify/templates/tasks-template.md: ✅ compatible (TDD workflow matches)

Deferred items: None
-->

# Loomio Rewrite Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

All feature development MUST follow Test-Driven Development:

1. **Write failing tests first** - No implementation code before tests exist
2. **Red-Green-Refactor cycle** - Tests fail → implement minimum to pass → refactor
3. **Test coverage requirements**:
   - Go backend: stdlib `testing`, table-driven tests
   - Database logic: pgTap for schema/function tests
   - Frontend units: Vitest
   - Components: Storybook tests
   - E2E: Playwright for critical user journeys

**Rationale**: Clean slate rewrite demands confidence in every change. Tests are the specification; implementation follows.

### II. Spec-First API Design

OpenAPI specifications are the single source of truth for all API contracts:

1. **Edit spec first** - All API changes start in `openapi/` YAML files
2. **Generate, don't hand-write** - Go types generated via oapi-codegen
3. **Huma validates at runtime** - Framework enforces spec compliance
4. **Type chain to frontend** - Generated TypeScript types from Huma's live OpenAPI output

**Rationale**: 204 endpoints already specified in discovery. Spec-first preserves that investment and ensures frontend/backend alignment.

### III. Security-First

Address all known security issues before adding new features:

1. **CRITICAL issues block production** - OAuth CSRF, rate limiting, webhook signing
2. **No security shortcuts** - Every exception requires documented justification
3. **Preserve existing security patterns**:
   - HTML sanitization (whitelist-based)
   - Authorization on all mutations
   - CSRF validation
   - Cryptographic tokens for sensitive operations
4. **Rate limit all endpoints** - Including bot APIs (`/api/b2/`, `/api/b3/`)

**Rationale**: Original codebase has documented security gaps. Rewrite MUST fix them, not perpetuate them.

### IV. Full-Stack Type Safety

Types flow from database to browser with compile-time guarantees:

1. **sqlc for database access** - SQL queries generate type-safe Go code
2. **No `interface{}` for domain data** - Explicit types required
3. **TypeScript strict mode** - No `any` types in SvelteKit codebase
4. **Single type source** - OpenAPI → Go structs → TypeScript (generated, not duplicated)

**Rationale**: Type errors caught at compile time prevent runtime failures in production.

### V. Simplicity & YAGNI

Build only what is needed, in the simplest way possible:

1. **No premature abstraction** - Three similar lines are better than a premature helper
2. **No speculative features** - If not in the spec, don't build it
3. **Minimal dependencies** - Prefer stdlib over third-party when equivalent
4. **Delete, don't deprecate** - Unused code is removed, not commented
5. **Deferred complexity** - Collaborative editing (Yjs), Matrix chatbot deferred to later phases

**Rationale**: Rewrites fail from scope creep. Each addition must justify its presence.

## Technology Constraints

**Backend Stack** (mandatory):
- Go 1.25+ with Huma web framework
- PostgreSQL 18 with sqlc + pgx/v5
- goose for embedded migrations
- River for Postgres-backed background jobs
- Redis for cache/pubsub

**Frontend Stack** (mandatory):
- SvelteKit with TypeScript strict mode
- Uppy for file uploads (presigned URLs)
- Vitest + Storybook + Playwright for testing

**Real-time**:
- SSE for notifications and vote updates
- WebSockets for bidirectional features (chat)

Deviations require constitution amendment with documented justification.

## Development Workflow

### Feature Implementation Flow

1. **Spec**: Write/update OpenAPI spec in `openapi/`
2. **Generate**: Run oapi-codegen to produce Go types
3. **Test**: Write failing tests (Go, pgTap, Vitest, Playwright as appropriate)
4. **Implement**: Build minimum code to pass tests
5. **Validate**: Run full test suite, security checks
6. **Commit**: Conventional commits, small focused changes

### Code Review Requirements

- All PRs require passing CI (tests, linting, type checks)
- Security-related changes require explicit security review
- Constitution violations block merge

### Project Structure

```
cmd/server/       # Application entrypoint
internal/         # Private application code
  api/            # Huma operations
  auth/           # Sessions, OAuth, SAML
  db/             # sqlc generated code
  jobs/           # River job definitions
  mail/           # Email sending
  realtime/       # SSE/WebSocket handlers
web/              # SvelteKit frontend
migrations/       # SQL migrations (embedded)
openapi/          # Source OpenAPI specs
generated/        # Generated code (Go, TypeScript)
```

## Governance

### Amendment Process

1. Propose change with rationale in PR
2. Document what principle changes and why
3. Assess impact on existing features
4. Update version (MAJOR.MINOR.PATCH):
   - MAJOR: Principle removal or incompatible redefinition
   - MINOR: New principle or material expansion
   - PATCH: Clarification or non-semantic refinement
5. Update all dependent templates

### Compliance

- All PRs MUST pass Constitution Check (see plan-template.md)
- Violations require explicit justification in Complexity Tracking table
- This constitution supersedes conflicting practices

**Version**: 1.0.0 | **Ratified**: 2026-02-01 | **Last Amended**: 2026-02-01
