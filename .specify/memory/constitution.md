<!--
SYNC IMPACT REPORT
==================
Version change: 1.1.0 → 1.2.0 (MINOR: add /speckit.analyze remediation policy)

Modified principles: None

Added sections:
- "Specification Analysis (`/speckit.analyze`)" under Development Workflow

Removed sections: None

Templates requiring updates:
- .specify/templates/plan-template.md: ✅ compatible (no changes needed)
- .specify/templates/spec-template.md: ✅ compatible (no changes needed)
- .specify/templates/tasks-template.md: ✅ compatible (no changes needed)

Deferred items: None

Amendment rationale:
The 004-groups-memberships analysis revealed that automatically suggesting
remediation edits for MEDIUM+ severity issues improves specification quality
and catches gaps before implementation begins. This policy ensures consistent
handling of analysis findings across all features.

Previous version (1.1.0) changes:
- II. Spec-First API Design → II. Huma-First API Design (renamed + workflow clarified)
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

### II. Huma-First API Design

Huma operations are the single source of truth for API contracts:

1. **Define types in Go** - Request/response structs with Huma tags define the contract
2. **Huma generates OpenAPI** - Runtime `/openapi.json` endpoint is the canonical spec
3. **Document in `openapi/`** - YAML files in `openapi/paths/` serve as human-readable documentation and design artifacts (not code generation source)
4. **Type chain to frontend** - Generated TypeScript types from Huma's live OpenAPI output

**Workflow**:
1. Design endpoint in `openapi/paths/*.yaml` (documentation + design review)
2. Implement Go types and Huma operation (source of truth)
3. Huma validates requests/responses at runtime
4. Frontend generates types from Huma's `/openapi.json`

**Rationale**: Huma's Go-native approach provides better IDE support, compile-time checking, and eliminates drift between spec and implementation. The OpenAPI YAML files remain valuable for design discussion and documentation.

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
5. **Timing-safe operations** - Use constant-time comparison for auth, dummy hashes for non-existent users

**Rationale**: Original codebase has documented security gaps. Rewrite MUST fix them, not perpetuate them.

### IV. Full-Stack Type Safety

Types flow from database to browser with compile-time guarantees:

1. **sqlc for database access** - SQL queries generate type-safe Go code
2. **No `interface{}` for domain data** - Explicit types required (use `any` sparingly, only for truly dynamic data)
3. **TypeScript strict mode** - No `any` types in SvelteKit codebase
4. **Single type source** - Go structs → Huma OpenAPI → TypeScript (generated, not duplicated)

**Rationale**: Type errors caught at compile time prevent runtime failures in production.

### V. Simplicity & YAGNI

Build only what is needed, in the simplest way possible:

1. **No premature abstraction** - Three similar lines are better than a premature helper
2. **No speculative features** - If not in the spec, don't build it
3. **Minimal dependencies** - Prefer stdlib over third-party when equivalent
4. **Delete, don't deprecate** - Unused code is removed, not commented
5. **Deferred complexity** - Collaborative editing (Yjs), Matrix chatbot deferred to later phases
6. **MVP-first infrastructure** - Start with simple solutions (e.g., in-memory sessions), upgrade when needed

**Rationale**: Rewrites fail from scope creep. Each addition must justify its presence.

## Technology Constraints

**Backend Stack** (mandatory):
- Go 1.25+ with Huma web framework
- PostgreSQL 18 with sqlc + pgx/v5
- goose for embedded migrations
- River for Postgres-backed background jobs (when needed)
- Redis for cache/pubsub (when needed)

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

1. **Design**: Write OpenAPI spec in `openapi/` for design review
2. **Test**: Write failing tests (Go, pgTap, Vitest, Playwright as appropriate)
3. **Implement**: Build Huma operations with Go types (becomes source of truth)
4. **Validate**: Run full test suite, security checks
5. **Verify**: Confirm Huma's generated OpenAPI matches design intent
6. **Commit**: Conventional commits, small focused changes

### Code Review Requirements

- All PRs require passing CI (tests, linting, type checks)
- Security-related changes require explicit security review
- Constitution violations block merge

### Specification Analysis (`/speckit.analyze`)

- After running `/speckit.analyze`, automatically suggest remediation edits for ALL issues with severity above LOW (i.e., MEDIUM, HIGH, CRITICAL)
- LOW severity issues are reported but remediation is optional
- User must approve remediation edits before they are applied
- CRITICAL issues MUST be resolved before `/speckit.implement`

### Project Structure

```
cmd/server/       # Application entrypoint
cmd/migrate/      # Database migration tool
internal/         # Private application code
  api/            # Huma operations + middleware
  auth/           # Sessions, password hashing, tokens
  db/             # sqlc generated code + pool
  jobs/           # River job definitions (when needed)
  mail/           # Email sending (when needed)
  realtime/       # SSE/WebSocket handlers (when needed)
web/              # SvelteKit frontend (when needed)
migrations/       # SQL migrations (embedded via goose)
openapi/          # Design documentation (human-readable specs)
tests/            # Additional test files
  pgtap/          # PostgreSQL schema tests
```

## Governance

### Amendment Process

1. Propose change with rationale in PR
2. Document what principle changes and why
3. Assess impact on existing features
4. Update version (MAJOR.MINOR.PATCH):
   - MAJOR: Principle removal or incompatible redefinition
   - MINOR: New principle or material expansion/clarification
   - PATCH: Typo fixes or non-semantic refinement
5. Update all dependent templates

### Compliance

- All PRs MUST pass Constitution Check (see plan-template.md)
- Violations require explicit justification in Complexity Tracking table
- This constitution supersedes conflicting practices

**Version**: 1.2.0 | **Ratified**: 2026-02-01 | **Last Amended**: 2026-02-02
