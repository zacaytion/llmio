# Implementation Plan: Discussions & Comments

**Branch**: `005-discussions` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)

## Summary

Add threaded discussions with comments to enable group conversations. Users create discussions within groups (subject to permission flags) or as direct discussions with specific participants. Comments nest up to a configurable depth, support editing and soft deletion, and track per-user read state.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: Huma (API), pgx/v5 (database), sqlc (type-safe queries)
**Storage**: PostgreSQL 18 (existing)
**Testing**: Go stdlib `testing` with table-driven tests; pgTap for schema tests
**Target Platform**: Linux server (Docker/Podman)
**Project Type**: Web API (backend only for this feature)
**Performance Goals**: 95% of read-state updates within 500ms (SC-004)
**Constraints**: Comment threading up to 10 levels (respecting max_depth)
**Scale/Scope**: Supports existing user base; no specific scale target for MVP

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-First Development | ✅ Pass | pgTap for schema, Go table-driven tests for API |
| II. Huma-First API Design | ✅ Pass | Huma operations define contracts; OpenAPI for docs |
| III. Security-First | ✅ Pass | Permission checks via group flags; author/admin deletion |
| IV. Full-Stack Type Safety | ✅ Pass | sqlc generates types; Huma validates requests |
| V. Simplicity & YAGNI | ✅ Pass | No real-time, no notifications—both deferred |

## Project Structure

### Documentation (this feature)

```text
specs/005-discussions/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (OpenAPI YAML)
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── discussions.go     # Huma operations for discussions
│   └── comments.go        # Huma operations for comments
├── db/
│   ├── queries/
│   │   ├── discussions.sql  # sqlc queries
│   │   └── comments.sql     # sqlc queries
│   └── *.sql.go             # sqlc generated output (per sqlc.yaml: out: "internal/db")
└── discussion/              # Domain logic (service layer)
    ├── service.go           # DiscussionService
    ├── comment_service.go   # CommentService
    └── permissions.go       # Permission checks

migrations/
├── 002_create_audit_schema.sql        # Feature 004 (dependency)
├── 003_create_groups.sql              # Feature 004 (dependency)
├── 004_create_memberships.sql         # Feature 004 (dependency)
├── 005_enable_auditing.sql            # Feature 004 (dependency)
├── 006_create_discussions.sql         # This feature
├── 007_create_comments.sql            # This feature
└── 008_create_discussion_readers.sql  # This feature

tests/pgtap/
├── 006_discussions_test.sql           # Schema constraints
├── 007_comments_test.sql              # Comment threading tests
└── 008_discussion_readers_test.sql    # Read tracking tests
```

**Structure Decision**: Follows existing `internal/` layout from Features 001-003. Adds `internal/discussion/` for domain logic, keeping API handlers thin.

## Complexity Tracking

No constitution violations. Table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
