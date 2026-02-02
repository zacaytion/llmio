# Implementation Plan: User Authentication

**Branch**: `001-user-auth` | **Date**: 2026-02-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-user-auth/spec.md`

## Summary

Implement email/password authentication with Argon2id password hashing and in-memory session storage. This provides the foundational user registration, login, logout, and session management required before any other features can be built.

## Technical Context

**Language/Version**: Go 1.25+ with Huma web framework
**Primary Dependencies**: Huma, pgx/v5, sqlc, golang.org/x/crypto/argon2
**Storage**: PostgreSQL 18 for users; in-memory Go map for sessions (MVP)
**Testing**: Go stdlib `testing` with table-driven tests, pgTap for schema
**Target Platform**: Linux server (Docker)
**Project Type**: Web application (Go backend + SvelteKit frontend)
**Performance Goals**: 100 concurrent logins without degradation (SC-007), login within 5 seconds (SC-002)
**Constraints**: <3 second response for invalid logins (SC-006), sessions lost on restart (acceptable per spec)
**Scale/Scope**: MVP, single server, ~1000 users initially

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | PASS | All code via TDD; table-driven Go tests + pgTap for schema |
| II. Spec-First API Design | PASS | OpenAPI contracts in `/contracts/`; Huma validates and generates types at runtime |
| III. Security-First | PASS | Argon2id hashing, no account enumeration, constant-time comparison |
| IV. Full-Stack Type Safety | PASS | sqlc generates Go types; TypeScript from OpenAPI |
| V. Simplicity & YAGNI | PASS | In-memory sessions per spec; no Redis initially |

**Pre-design gate: PASSED**

## Project Structure

### Documentation (this feature)

```text
specs/001-user-auth/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── auth.yaml        # OpenAPI spec for auth endpoints
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
# Go backend (following constitution project structure)
cmd/server/              # Application entrypoint
internal/
  api/                   # Huma operations
    auth.go              # Register, Login, Logout handlers
  auth/                  # Sessions, password hashing
    session.go           # In-memory session store
    password.go          # Argon2id hashing utilities
  db/                    # sqlc generated code
    queries/
      users.sql          # User CRUD queries
migrations/              # SQL migrations (embedded via goose)
  001_create_users.sql   # Users table with auth fields
openapi/                 # Source OpenAPI specs
  paths/
    auth.yaml            # Authentication endpoints

# SvelteKit frontend (out of scope for this feature per spec)
web/                     # Frontend forms are separate feature

# Tests
internal/
  api/
    auth_test.go         # Handler tests
  auth/
    session_test.go      # Session store tests
    password_test.go     # Argon2id tests
  db/
    users_test.go        # Query tests (integration)
```

**Structure Decision**: Following constitution's mandated project structure. Backend-only for this feature; frontend forms are a declared dependency in spec.

## Complexity Tracking

> No violations. Design follows constitution principles.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | - | - |
