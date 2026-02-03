# Implementation Plan: Groups & Memberships

**Branch**: `004-groups-memberships` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-groups-memberships/spec.md`

## Summary

Build organizational containers (Groups) with hierarchy support (subgroups) and permission-based membership system. Groups are the foundational unit for collaborative decision-making - containing members with roles (admin/member), 11 configurable permission flags, and audit logging via PostgreSQL triggers using the supa_audit pattern.

## Technical Context

**Language/Version**: Go 1.25+ (matches existing codebase)
**Primary Dependencies**: Huma web framework, pgx/v5, sqlc, go-playground/validator/v10
**Storage**: PostgreSQL 18 with CITEXT extension (case-insensitive handles), pgx/v5 + sqlc
**Testing**: Go stdlib `testing` with table-driven tests, pgTap for database triggers/constraints
**Target Platform**: Linux server (API backend)
**Project Type**: Web application (Go backend, extends existing internal/ structure)
**Performance Goals**: Group operations < 100ms p95, permission checks < 10ms
**Constraints**: Last-admin protection enforced at DB level, audit logging via triggers (not app layer)
**Scale/Scope**: 10K+ groups, millions of memberships, 11 permission flags per group

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Status | Notes |
|-----------|-------------|--------|-------|
| I. Test-First Development | Write failing tests before implementation | ✅ COMPLIANT | pgTap for DB triggers, Go table-driven tests for API |
| II. Huma-First API Design | Define types in Go, Huma generates OpenAPI | ✅ COMPLIANT | Group/Membership DTOs → Huma operations |
| III. Security-First | Authorization on all mutations, no shortcuts | ✅ COMPLIANT | Admin-only mutations, permission flag enforcement |
| IV. Full-Stack Type Safety | sqlc for DB, explicit types, no `interface{}` | ✅ COMPLIANT | sqlc generates typed Group/Membership models |
| V. Simplicity & YAGNI | Build only what's needed | ✅ COMPLIANT | No email notifications (deferred), MVP permission checks |

**Pre-Design Gate**: ✅ PASSED - All principles aligned

### Post-Design Evaluation

| Principle | Design Artifacts | Status | Verification |
|-----------|------------------|--------|--------------|
| I. Test-First | data-model.md defines test cases for triggers | ✅ PASS | pgTap tests planned for last-admin, audit triggers |
| II. Huma-First | contracts/groups.yaml documents API design | ✅ PASS | Go types will be source of truth; YAML is documentation |
| III. Security-First | Authorization rules defined in quickstart.md | ✅ PASS | Admin/member role checks on all mutations |
| IV. Type Safety | data-model.md defines typed entities | ✅ PASS | sqlc generates Group, Membership, audit types |
| V. YAGNI | Only spec requirements implemented | ✅ PASS | No email notifications, no extra features |

**Post-Design Gate**: ✅ PASSED - Design artifacts align with constitution

## Project Structure

### Documentation (this feature)

```text
specs/004-groups-memberships/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── groups.yaml      # OpenAPI spec for groups/memberships endpoints
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── api/
│   ├── groups.go          # Huma operations for groups CRUD
│   ├── groups_test.go     # API integration tests
│   ├── memberships.go     # Huma operations for memberships
│   └── memberships_test.go
├── db/
│   ├── models.go          # sqlc-generated (Group, Membership types)
│   ├── groups.sql.go      # sqlc-generated queries
│   └── memberships.sql.go # sqlc-generated queries

migrations/
├── 001_create_users.sql           # (existing)
├── 002_create_audit_schema.sql    # Audit infrastructure (supa_audit pattern)
├── 003_create_groups.sql          # Groups table + triggers
├── 004_create_memberships.sql     # Memberships table + last-admin trigger
└── 005_enable_auditing.sql        # Enable audit triggers on groups + memberships

queries/
├── groups.sql             # sqlc query definitions
└── memberships.sql        # sqlc query definitions

tests/pgtap/
├── 002_audit_schema_test.sql      # Audit trigger tests
├── 003_groups_test.sql            # Group constraints/triggers
└── 004_memberships_test.sql       # Last-admin protection tests
```

**Structure Decision**: Extends existing Go backend structure with new internal/api handlers and internal/db models. Audit infrastructure is a shared schema (002_) used by both groups and memberships tables.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*No violations - all complexity is justified by spec requirements:*

| Design Choice | Justification |
|---------------|---------------|
| PostgreSQL triggers for audit | TC-001/TC-002 requirement; ensures audit capture even if app has bugs |
| Last-admin trigger | TC-005 requirement; DB-level enforcement prevents race conditions |
| 11 permission flags | FR-019 requirement; matches original Loomio's proven governance model |
