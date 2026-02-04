# Quickstart: Groups & Memberships

**Feature**: 004-groups-memberships
**Date**: 2026-02-02

## Overview

This feature adds organizational containers (Groups) with hierarchy support and permission-based membership management. Groups are the foundational unit for collaborative decision-making in Loomio.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Group** | Container with name, handle, description, and 11 permission flags |
| **Membership** | User's relationship to a group (role: admin/member, status: pending/active) |
| **Subgroup** | Child group linked to a parent via `parent_id` |
| **Audit Log** | PostgreSQL trigger-based logging of all changes |

## API Endpoints Summary

### Groups

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/groups` | Create group (auto-makes creator admin) |
| GET | `/api/v1/groups` | List user's groups |
| GET | `/api/v1/groups/{id}` | Get group details |
| PATCH | `/api/v1/groups/{id}` | Update group (admin only) |
| POST | `/api/v1/groups/{id}/archive` | Archive group |
| POST | `/api/v1/groups/{id}/unarchive` | Unarchive group |
| POST | `/api/v1/groups/{id}/subgroups` | Create subgroup |
| GET | `/api/v1/group-by-handle/{handle}` | Lookup by handle |

### Memberships

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/groups/{groupId}/memberships` | List group members |
| POST | `/api/v1/groups/{groupId}/memberships` | Invite user |
| DELETE | `/api/v1/memberships/{id}` | Remove member (admin only) |
| POST | `/api/v1/memberships/{id}/accept` | Accept invitation |
| POST | `/api/v1/memberships/{id}/promote` | Promote to admin |
| POST | `/api/v1/memberships/{id}/demote` | Demote to member |
| GET | `/api/v1/users/me/invitations` | List pending invitations |

## Common Workflows

### 1. Create a Group and Invite Members

```bash
# Create group (returns group with creator as admin)
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Cookie: loomio_session=..." \
  -H "Content-Type: application/json" \
  -d '{"name": "Climate Action Team", "description": "Working on climate initiatives"}'

# Response: {"group": {"id": 1, "name": "Climate Action Team", "handle": "climate-action-team", ...}}

# Invite a user (note: path uses {groupId} not {id})
curl -X POST http://localhost:8080/api/v1/groups/1/memberships \
  -H "Cookie: loomio_session=..." \
  -H "Content-Type: application/json" \
  -d '{"user_id": 42, "role": "member"}'

# Response: {"membership": {"id": 1, "user_id": 42, "role": "member", "accepted_at": null, ...}}
```

### 2. Accept an Invitation

```bash
# List pending invitations
curl http://localhost:8080/api/v1/users/me/invitations \
  -H "Cookie: loomio_session=..."

# Response: {"invitations": [{"id": 1, "group": {...}, "inviter": {...}, ...}]}

# Accept invitation
curl -X POST http://localhost:8080/api/v1/memberships/1/accept \
  -H "Cookie: loomio_session=..."

# Response: {"membership": {"id": 1, "accepted_at": "2026-02-02T...", ...}}
```

### 3. Configure Group Permissions

```bash
# Update permission flags (admin only)
curl -X PATCH http://localhost:8080/api/v1/groups/1 \
  -H "Cookie: loomio_session=..." \
  -H "Content-Type: application/json" \
  -d '{"members_can_add_members": true, "members_can_start_discussions": true}'
```

## Permission Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `members_can_add_members` | true | Members can invite others |
| `members_can_add_guests` | true | Members can add discussion guests |
| `members_can_start_discussions` | true | Members can create discussions |
| `members_can_raise_motions` | true | Members can create polls |
| `members_can_edit_discussions` | false | Members can edit discussion titles |
| `members_can_edit_comments` | true | Members can edit own comments |
| `members_can_delete_comments` | true | Members can delete own comments |
| `members_can_announce` | false | Members can send announcements |
| `members_can_create_subgroups` | false | Members can create subgroups |
| `admins_can_edit_user_content` | false | Admins can edit any content |
| `parent_members_can_see_discussions` | false | Parent members see subgroup content |

## Authorization Rules

### Who Can Do What

| Action | Admin | Member | Non-member |
|--------|-------|--------|------------|
| View group | ✅ | ✅ | ❌ |
| Update group settings | ✅ | ❌ | ❌ |
| Archive/unarchive group | ✅ | ❌ | ❌ |
| Invite members | ✅ | if `members_can_add_members` | ❌ |
| Remove members | ✅ | ❌ | ❌ |
| Promote/demote | ✅ | ❌ | ❌ |
| Create subgroups | ✅ | if `members_can_create_subgroups` | ❌ |

## Database Schema

### Core Tables

```sql
-- Groups table
CREATE TABLE groups (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    handle          CITEXT NOT NULL UNIQUE,
    description     TEXT,
    parent_id       BIGINT REFERENCES groups(id),
    created_by_id   BIGINT NOT NULL REFERENCES users(id),
    archived_at     TIMESTAMPTZ,
    -- 11 permission flags (all BOOLEAN NOT NULL DEFAULT)
    members_can_add_members BOOLEAN NOT NULL DEFAULT TRUE,
    -- ... other flags ...
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Memberships table
CREATE TABLE memberships (
    id          BIGSERIAL PRIMARY KEY,
    group_id    BIGINT NOT NULL REFERENCES groups(id),
    user_id     BIGINT NOT NULL REFERENCES users(id),
    role        TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    inviter_id  BIGINT NOT NULL REFERENCES users(id),
    accepted_at TIMESTAMPTZ,  -- NULL = pending invitation
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, user_id)
);
```

### Audit Schema

```sql
-- Audit log (supa_audit pattern)
CREATE TABLE audit.record_version (
    id          BIGSERIAL PRIMARY KEY,
    op          audit.operation NOT NULL,  -- INSERT/UPDATE/DELETE
    ts          TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    xact_id     BIGINT NOT NULL DEFAULT txid_current(),
    table_oid   OID NOT NULL,
    table_name  NAME NOT NULL,
    record      JSONB,      -- New state
    old_record  JSONB,      -- Old state
    actor_id    BIGINT      -- User who made change
);
```

## Testing

### Run Tests

```bash
# Go API tests
go test ./internal/api/... -v -run TestGroup
go test ./internal/api/... -v -run TestMembership

# Database trigger tests (pgTap)
make test-pgtap
```

### Key Test Cases

1. **Group Creation**: Verify creator becomes admin
2. **Handle Generation**: Verify slug generation from name
3. **Handle Uniqueness**: Case-insensitive collision detection
4. **Last Admin Protection**: Cannot demote/remove last admin
5. **Permission Enforcement**: Members blocked when flag is false
6. **Audit Logging**: Changes create audit.record_version entries

## Error Codes

| HTTP | Error | When |
|------|-------|------|
| 401 | `unauthorized` | Not authenticated |
| 403 | `forbidden` | Not a member or insufficient role |
| 404 | `not_found` | Group or membership doesn't exist |
| 409 | `conflict` | Handle taken / already member / last admin |
| 422 | `validation_error` | Invalid input data |

## Migration Commands

```bash
# Run migrations
go run ./cmd/migrate up

# Check migration status
go run ./cmd/migrate status

# Rollback last migration
go run ./cmd/migrate down
```

## Files Changed

### New Files

```
internal/api/groups.go          # Group handlers
internal/api/groups_test.go     # Group tests
internal/api/memberships.go     # Membership handlers
internal/api/memberships_test.go
queries/groups.sql              # sqlc queries
queries/memberships.sql
migrations/002_create_audit_schema.sql
migrations/003_create_groups.sql
migrations/004_create_memberships.sql
migrations/005_enable_auditing.sql
tests/pgtap/003_groups_test.sql
tests/pgtap/004_memberships_test.sql
```

### Modified Files

```
internal/api/dto.go             # Add GroupDTO, MembershipDTO
internal/db/models.go           # sqlc regenerated
cmd/server/main.go              # Register new handlers
```
