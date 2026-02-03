# Data Model: Groups & Memberships

**Feature**: 004-groups-memberships
**Date**: 2026-02-02

## Entity Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          USERS (existing)                       │
│  id, email, name, username, password_hash, ...                  │
└─────────────────────────────────────────────────────────────────┘
         │                                    │
         │ creates                            │ is_member_of
         ▼                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                            GROUPS                               │
│  id, name, handle, description, parent_id (self-ref), ...       │
│  + 11 permission flags (members_can_*)                          │
└─────────────────────────────────────────────────────────────────┘
         │                     │
         │ has_many            │ parent_of (self-ref)
         ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                         MEMBERSHIPS                             │
│  id, group_id, user_id, role, inviter_id, accepted_at, ...      │
└─────────────────────────────────────────────────────────────────┘
         │
         │ audited_by
         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AUDIT.RECORD_VERSION                         │
│  id, record_id, op, ts, xact_id, table_oid, record, old_record  │
│  actor_id (from session var)                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Entity: Group

**Purpose**: Organizational container for collaborative decision-making. Contains members, discussions, and polls.

### Attributes

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | BIGSERIAL | NOT NULL | auto | Primary key |
| `name` | TEXT | NOT NULL | - | Display name (1-255 chars) |
| `handle` | CITEXT | NOT NULL | - | URL-safe identifier (3-100 chars, unique, lowercase) |
| `description` | TEXT | NULL | - | Optional description |
| `parent_id` | BIGINT | NULL | - | FK to groups.id for subgroups |
| `created_by_id` | BIGINT | NOT NULL | - | FK to users.id (creator) |
| `archived_at` | TIMESTAMPTZ | NULL | - | Soft delete timestamp |
| `created_at` | TIMESTAMPTZ | NOT NULL | NOW() | Creation timestamp |
| `updated_at` | TIMESTAMPTZ | NOT NULL | NOW() | Last update timestamp |

### Permission Flags (all BOOLEAN NOT NULL DEFAULT)

| Column | Default | Description |
|--------|---------|-------------|
| `members_can_add_members` | TRUE | Members can invite others |
| `members_can_add_guests` | TRUE | Members can add discussion guests |
| `members_can_start_discussions` | TRUE | Members can create discussions |
| `members_can_raise_motions` | TRUE | Members can create polls/proposals |
| `members_can_edit_discussions` | FALSE | Members can edit discussion titles |
| `members_can_edit_comments` | TRUE | Members can edit own comments |
| `members_can_delete_comments` | TRUE | Members can delete own comments |
| `members_can_announce` | FALSE | Members can send announcements |
| `members_can_create_subgroups` | FALSE | Members can create subgroups |
| `admins_can_edit_user_content` | FALSE | Admins can edit any content |
| `parent_members_can_see_discussions` | FALSE | Parent group members see subgroup content |

### Constraints

```sql
CONSTRAINT groups_name_length CHECK (LENGTH(name) BETWEEN 1 AND 255)
-- Handle validation: 3-100 chars, lowercase alphanumeric + hyphens, must start and end with alphanumeric
-- Note: Regex alone allows 2-char (e.g., "ab"); LENGTH check enforces minimum 3
CONSTRAINT groups_handle_format CHECK (
    handle ~* '^[a-z0-9][a-z0-9-]*[a-z0-9]$'
    AND LENGTH(handle) BETWEEN 3 AND 100
)
CONSTRAINT groups_parent_not_self CHECK (parent_id != id)
```

### Indexes

| Index | Type | Columns | Notes |
|-------|------|---------|-------|
| `groups_pkey` | PRIMARY KEY | `id` | - |
| `groups_handle_key` | UNIQUE | `lower(handle)` | Case-insensitive uniqueness |
| `groups_parent_id_idx` | B-TREE | `parent_id` | Subgroup queries |
| `groups_created_by_id_idx` | B-TREE | `created_by_id` | User's created groups |
| `groups_archived_at_idx` | B-TREE | `archived_at` | Filter active/archived |

### Relationships

| Relationship | Type | Target | FK Column |
|--------------|------|--------|-----------|
| Creator | belongs_to | User | `created_by_id` |
| Parent | belongs_to | Group | `parent_id` |
| Subgroups | has_many | Group | `parent_id` (inverse) |
| Memberships | has_many | Membership | `group_id` (inverse) |

---

## Entity: Membership

**Purpose**: Relationship between a User and a Group, tracking role and invitation status.

### Attributes

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | BIGSERIAL | NOT NULL | auto | Primary key |
| `group_id` | BIGINT | NOT NULL | - | FK to groups.id |
| `user_id` | BIGINT | NOT NULL | - | FK to users.id |
| `role` | TEXT | NOT NULL | 'member' | 'admin' or 'member' |
| `inviter_id` | BIGINT | NOT NULL | - | FK to users.id (who invited) |
| `accepted_at` | TIMESTAMPTZ | NULL | - | NULL = pending invitation |
| `created_at` | TIMESTAMPTZ | NOT NULL | NOW() | Invitation created |
| `updated_at` | TIMESTAMPTZ | NOT NULL | NOW() | Last update |

### Constraints

```sql
CONSTRAINT memberships_role_valid CHECK (role IN ('admin', 'member'))
CONSTRAINT memberships_unique_user_group UNIQUE (group_id, user_id)
```

### Indexes

| Index | Type | Columns | Notes |
|-------|------|---------|-------|
| `memberships_pkey` | PRIMARY KEY | `id` | - |
| `memberships_group_user_key` | UNIQUE | `group_id, user_id` | Prevent duplicates |
| `memberships_user_id_idx` | B-TREE | `user_id` | User's groups |
| `memberships_group_id_idx` | B-TREE | `group_id` | Group's members |
| `memberships_inviter_id_idx` | B-TREE | `inviter_id` | Invited by queries |
| `memberships_pending_idx` | B-TREE | `user_id, accepted_at` | WHERE `accepted_at IS NULL` |

### Relationships

| Relationship | Type | Target | FK Column |
|--------------|------|--------|-----------|
| Group | belongs_to | Group | `group_id` |
| User | belongs_to | User | `user_id` |
| Inviter | belongs_to | User | `inviter_id` |

### Triggers

| Trigger | Event | Function | Purpose |
|---------|-------|----------|---------|
| `memberships_last_admin_protection` | BEFORE UPDATE/DELETE | `prevent_last_admin_removal()` | FR-016: Prevent removing last admin |
| `memberships_audit` | AFTER INSERT/UPDATE/DELETE | `audit.insert_update_delete_trigger()` | TC-003: Audit logging |
| `memberships_updated_at` | BEFORE UPDATE | `update_updated_at()` | Timestamp maintenance |

---

## Entity: audit.record_version

**Purpose**: Immutable audit log for groups and memberships changes.

### Attributes

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| `id` | BIGSERIAL | NOT NULL | auto | Primary key |
| `record_id` | TEXT | NULL | - | Current record identifier (BIGINT cast to TEXT for BIGSERIAL PKs per TC-002) |
| `op` | audit.operation | NOT NULL | - | INSERT/UPDATE/DELETE/TRUNCATE/SNAPSHOT |
| `ts` | TIMESTAMPTZ | NOT NULL | clock_timestamp() | Change timestamp |
| `xact_id` | BIGINT | NOT NULL | txid_current() | Transaction ID for correlation |
| `table_oid` | OID | NOT NULL | - | PostgreSQL table identifier |
| `table_schema` | NAME | NOT NULL | - | Schema name |
| `table_name` | NAME | NOT NULL | - | Table name |
| `record` | JSONB | NULL | - | New row state |
| `old_record` | JSONB | NULL | - | Previous row state |
| `actor_id` | BIGINT | NULL | - | User who made change |

### Indexes

| Index | Type | Columns | Notes |
|-------|------|---------|-------|
| `record_version_pkey` | PRIMARY KEY | `id` | - |
| `record_version_record_id` | B-TREE PARTIAL | `record_id` | WHERE NOT NULL |
| `record_version_ts` | BRIN | `ts` | Time-range queries (99% smaller) |
| `record_version_table_oid` | B-TREE | `table_oid` | Filter by table |
| `record_version_xact_id` | B-TREE | `xact_id` | Transaction correlation |
| `record_version_actor_id` | B-TREE PARTIAL | `actor_id` | WHERE NOT NULL |

---

## State Transitions

### Membership States

```
                    ┌─────────────┐
                    │   INVITED   │  (accepted_at IS NULL)
                    └──────┬──────┘
                           │ accept()
                           ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   MEMBER    │◄────│   ACTIVE    │────►│    ADMIN    │
└─────────────┘     └──────┬──────┘     └─────────────┘
     role='member'         │                role='admin'
                           │ remove()
                           ▼
                    ┌─────────────┐
                    │   DELETED   │  (hard delete)
                    └─────────────┘
```

### Group States

```
┌─────────────┐                     ┌─────────────┐
│   ACTIVE    │────── archive() ───►│  ARCHIVED   │
│             │◄───── unarchive() ──│             │
└─────────────┘                     └─────────────┘
 archived_at IS NULL                 archived_at IS NOT NULL
```

---

## Validation Rules

### Group

| Field | Rule | Error |
|-------|------|-------|
| `name` | Required, 1-255 chars | "Name is required" / "Name too long" |
| `handle` | Optional on create (auto-generated), 3-100 chars, `^[a-z0-9][a-z0-9-]*[a-z0-9]$` | "Handle must be 3-100 lowercase alphanumeric characters" |
| `handle` | Unique (case-insensitive) | "Handle already taken" |
| `parent_id` | Must reference existing group | "Parent group not found" |
| `parent_id` | Cannot be self | "Group cannot be its own parent" |

### Membership

| Field | Rule | Error |
|-------|------|-------|
| `group_id` | Must reference existing group | "Group not found" |
| `user_id` | Must reference existing user | "User not found" |
| `role` | Must be 'admin' or 'member' | "Invalid role" |
| `(group_id, user_id)` | Unique | "User is already a member of this group" |
| Last admin | Cannot demote/remove | "Cannot remove the last administrator" |

---

## Query Patterns

### Common Queries (sqlc)

```sql
-- name: CreateGroup :one
INSERT INTO groups (name, handle, description, parent_id, created_by_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: GetGroupByHandle :one
SELECT * FROM groups WHERE handle = $1;

-- name: ListGroupsByUser :many
SELECT g.* FROM groups g
JOIN memberships m ON m.group_id = g.id
WHERE m.user_id = $1 AND m.accepted_at IS NOT NULL AND g.archived_at IS NULL
ORDER BY g.name;

-- name: CreateMembership :one
INSERT INTO memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMembershipByGroupAndUser :one
SELECT * FROM memberships WHERE group_id = $1 AND user_id = $2;

-- name: ListMembershipsByGroup :many
SELECT m.*, u.name as user_name, u.email as user_email
FROM memberships m
JOIN users u ON u.id = m.user_id
WHERE m.group_id = $1
ORDER BY m.role DESC, u.name;

-- name: CountAdminsByGroup :one
SELECT COUNT(*) FROM memberships
WHERE group_id = $1 AND role = 'admin' AND accepted_at IS NOT NULL;
```

### External Dependencies (Feature 001)

This feature depends on the following queries from Feature 001 (User Authentication):

```sql
-- name: GetUserByID :one (from queries/users.sql)
-- Required for: validating user_id exists before creating invitation
-- Error case: If user not found, return 404 "User not found"
SELECT * FROM users WHERE id = $1;
```

**Note**: The `inviteMember` handler (T040) must verify the invited user exists before creating the membership. Use `db.GetUserByID` and check for `db.IsNotFound(err)` to return 404.

---

## Migration Order

1. **002_create_audit_schema.sql** - Audit infrastructure (shared)
2. **003_create_groups.sql** - Groups table + triggers
3. **004_create_memberships.sql** - Memberships table + last-admin trigger
4. **005_enable_auditing.sql** - Enable audit on groups + memberships

---

## Performance Considerations

| Operation | Expected Latency | Notes |
|-----------|------------------|-------|
| CreateGroup | < 50ms | Single insert + membership insert |
| GetGroupByHandle | < 5ms | Index lookup |
| ListGroupsByUser | < 20ms | Join with memberships |
| CreateMembership | < 30ms | Unique constraint check |
| Permission check | < 5ms | Membership + group fetch |

### Scaling Notes

- BRIN index on audit.ts handles time-range queries efficiently at scale
- Partial indexes on memberships (pending invitations) reduce index size
- Group permission flags are denormalized intentionally (11 booleans vs junction table)
