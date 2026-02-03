# Feature Specification: Groups & Memberships

**Feature Branch**: `004-groups-memberships`
**Created**: 2026-02-02
**Status**: Draft
**Input**: User description: "Build organizational containers (Groups) with hierarchy support and permission-based membership system for the Loomio rewrite"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a Group (Priority: P1)

A user wants to create a new organizational group to bring together people for collaborative decision-making. They provide a name and optional description, and the system creates the group with them as the first administrator.

**Why this priority**: Creating a group is the foundational action - without groups, no other collaborative features can exist. This must work before any membership management can happen.

**Independent Test**: Can be fully tested by having a user create a group and verifying they become its admin. Delivers immediate value as the user can now invite others.

**Acceptance Scenarios**:

1. **Given** a logged-in user, **When** they create a group with name "Climate Action Team", **Then** the group is created and they are automatically added as an administrator
2. **Given** a logged-in user creating a group, **When** they provide a description and custom handle "climate-team", **Then** the group is created with those details
3. **Given** a user creating a group, **When** the handle "climate-team" is already taken, **Then** the system rejects the creation with a clear error message
4. **Given** a group creation, **When** no handle is provided, **Then** the system generates one from the group name (e.g., "climate-action-team")

---

### User Story 2 - Invite Members to a Group (Priority: P1)

A group administrator wants to invite other users to join their group. They can add existing users as members with specific roles (admin or regular member).

**Why this priority**: Groups have no value without members. Inviting members is essential to enable collaboration immediately after group creation.

**Independent Test**: Can be tested by having an admin invite a user and verifying the invitation is created. The invited user can then accept and become a member.

**Acceptance Scenarios**:

1. **Given** a group administrator, **When** they invite a user to the group, **Then** a pending membership is created for that user
2. **Given** a pending invitation, **When** the invited user accepts, **Then** they become an active member of the group
3. **Given** a pending invitation, **When** the invited user views the invitation, **Then** they see who invited them and the group details
4. **Given** a group with an existing member, **When** an admin tries to invite them again, **Then** the system prevents duplicate memberships

---

### User Story 3 - Manage Group Members (Priority: P2)

A group administrator needs to manage the members of their group - promoting members to admin, demoting admins, or removing members who should no longer have access.

**Why this priority**: After establishing a group with members, administrators need control over roles and access. This enables proper governance but is not required for initial collaboration.

**Independent Test**: Can be tested by having an admin change a member's role and verifying the permission change takes effect immediately.

**Acceptance Scenarios**:

1. **Given** a group administrator and a regular member, **When** the admin promotes the member to admin, **Then** the member gains administrative privileges
2. **Given** a group with multiple administrators, **When** one admin demotes another, **Then** that person becomes a regular member
3. **Given** a group with one administrator, **When** they try to demote themselves, **Then** the system prevents this to ensure at least one admin exists
4. **Given** a group administrator, **When** they remove a member from the group, **Then** that person loses access to the group

---

### User Story 4 - Configure Group Permissions (Priority: P2)

A group administrator wants to configure what regular members can do within the group - whether they can invite others, start discussions, create polls, and so on.

**Why this priority**: Permission configuration allows groups to adapt to different governance models. Important for mature group management but not required for basic collaboration.

**Independent Test**: Can be tested by setting a permission flag (e.g., members can invite others) and verifying that non-admin members gain or lose that capability.

**Acceptance Scenarios**:

1. **Given** a group with "members can add members" enabled, **When** a regular member invites another user, **Then** the invitation is allowed
2. **Given** a group with "members can add members" disabled, **When** a regular member tries to invite someone, **Then** the action is denied
3. **Given** a group administrator, **When** they toggle "members can start discussions", **Then** the setting is saved and enforced immediately
4. **Given** a group, **When** viewing its settings, **Then** all 11 permission flags are visible with their current values

---

### User Story 5 - Create Subgroups (Priority: P3)

A group administrator wants to create subgroups within their main group for focused work on specific topics. Subgroups inherit from their parent but can override settings.

**Why this priority**: Subgroups add organizational depth but are not essential for initial group collaboration. Many groups function well as flat structures.

**Independent Test**: Can be tested by creating a subgroup under a parent group and verifying the hierarchical relationship is established.

**Acceptance Scenarios**:

1. **Given** a group administrator, **When** they create a subgroup, **Then** it is linked to the parent group
2. **Given** a parent group with specific permission settings, **When** a subgroup is created, **Then** it can optionally inherit those settings
3. **Given** a subgroup, **When** viewing its details, **Then** the parent group relationship is visible
4. **Given** a parent group member, **When** "parent members can see discussions" is enabled on a subgroup, **Then** they can view subgroup content

---

### User Story 6 - Archive a Group (Priority: P3)

A group administrator wants to archive a group that is no longer active. Archived groups should be preserved for historical reference but hidden from normal views.

**Why this priority**: Archiving is a lifecycle management feature. Groups must be functional before worrying about end-of-life management.

**Independent Test**: Can be tested by archiving a group and verifying it no longer appears in active group listings but can still be accessed directly.

**Acceptance Scenarios**:

1. **Given** a group administrator, **When** they archive the group, **Then** the group is marked as archived with a timestamp
2. **Given** an archived group, **When** users list their groups, **Then** the archived group is hidden by default
3. **Given** an archived group, **When** accessed directly by URL or handle, **Then** it displays with an archived indicator
4. **Given** an archived group, **When** an admin unarchives it, **Then** the group returns to normal active status

---

### Edge Cases

**Group Lifecycle:**
- What happens when the last administrator tries to leave a group? System prevents this - group must have at least one admin. Returns HTTP 409 with message "Cannot remove or demote the last administrator".
- How are group handles validated? Must be unique, lowercase, URL-safe (alphanumeric and hyphens only), 3-100 characters. Leading and trailing hyphens are not allowed (regex: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`). Invalid format returns HTTP 422.
- What happens when a group name is very long? Names are limited to 255 characters.

**Archived Group Behavior:**
- What happens when a parent group is archived? Subgroups remain accessible but their parent relationship shows as archived (via `parent_archived` indicator).
- Can you create a subgroup under an archived parent? No - returns HTTP 409 "Cannot create subgroup under archived group".
- Can you modify an archived group? No - PATCH /groups/{id} returns HTTP 409 "Cannot modify archived group". Unarchive first.
- Can you invite members to an archived group? No - returns HTTP 409 "Cannot invite to archived group".
- Can you promote/demote members in an archived group? No - returns HTTP 409 "Cannot modify membership in archived group".
- Can you remove members from an archived group? No - returns HTTP 409 "Cannot remove member from archived group".
- What happens to pending invitations when a group is archived? They remain pending but cannot be accepted until the group is unarchived. Accept returns HTTP 409.
- Can you unarchive a group when its parent is still archived? Yes - subgroups can be unarchived independently of parent status.

**Membership Invitations:**
- How does the system handle a user being invited to the same group multiple times? Prevents duplicate memberships, returns HTTP 409 "User is already a member or has a pending invitation".
- What happens when inviting a non-existent user? System returns HTTP 404 with message "User not found".
- What happens when accepting an already-accepted invitation? Returns HTTP 409 "Invitation already accepted".
- Can a pending member view group details? No - pending members cannot access GET /groups/{id}. Returns HTTP 403.
- Can a pending member invite others? No - pending members have no group permissions. Returns HTTP 403.
- Can a non-admin invite someone with the admin role? No - only admins can invite with role=admin. Returns HTTP 403.

**Role Changes:**
- What happens when promoting an already-admin member? Returns HTTP 409 "Member is already an administrator".
- What happens when demoting an already-member (non-admin)? Returns HTTP 409 "Member is already a regular member".

**Authorization Boundaries:**
- Can a non-member view membership details via GET /memberships/{id}? No - requires membership in the group. Returns HTTP 403.
- Can a regular member remove another member? No - only admins can remove members. Returns HTTP 403.

**Concurrency:**
- What happens with concurrent demote/promote operations? Database trigger enforces last-admin protection atomically. Concurrent requests that would violate the constraint receive HTTP 409.

**Subgroup Hierarchy:**
- Is there a maximum subgroup nesting depth? No explicit limit in this feature. Deep nesting is allowed.
- Can a group be its own parent (circular reference)? No - self-referential parent_id is rejected with HTTP 422.

## Requirements *(mandatory)*

### Functional Requirements

**Group Management**:
- **FR-001**: System MUST allow authenticated users to create groups with a name (required) and description (optional)
- **FR-002**: System MUST automatically create an admin membership for the user who creates a group
- **FR-003**: System MUST support unique, URL-safe group handles (alphanumeric + hyphens, 3-100 characters, lowercase)
- **FR-004**: System MUST auto-generate a handle from the group name if none is provided
- **FR-005**: System MUST support soft-deletion of groups via archiving (archived_at timestamp)
- **FR-006**: System MUST allow archived groups to be unarchived by administrators

**Group Hierarchy**:
- **FR-007**: System MUST support parent-child relationships between groups (subgroups)
- **FR-008**: System MUST allow subgroups to optionally inherit permission settings from parent groups (triggered by `inherit_permissions: true` in the create subgroup request; see contracts/groups.yaml CreateGroupRequest)
- **FR-009**: System MUST track the parent group relationship and allow navigation between parent and subgroups

**Membership Management**:
- **FR-010**: System MUST support two membership roles: administrator and regular member
- **FR-011**: System MUST allow administrators to invite users to join a group
- **FR-012**: System MUST track pending invitations separately from accepted memberships (via accepted_at timestamp)
- **FR-013**: System MUST allow invited users to accept invitations to become active members
- **FR-014**: System MUST prevent duplicate memberships (one membership per user per group)
- **FR-015**: System MUST allow administrators to promote members to admin or demote admins to members
- **FR-016**: System MUST prevent the last administrator from being removed or demoted
- **FR-017**: System MUST allow administrators to remove members from the group
- **FR-018**: System MUST track who invited each member (inviter_id)

**Permission System**:
- **FR-019**: System MUST support the following group permission flags (all boolean):
  - `members_can_add_members` - members can invite other members
  - `members_can_add_guests` - members can add guests to discussions
  - `members_can_start_discussions` - members can create new discussions
  - `members_can_raise_motions` - members can create polls/proposals
  - `members_can_edit_discussions` - members can edit discussion titles/descriptions
  - `members_can_edit_comments` - members can edit their own comments
  - `members_can_delete_comments` - members can delete their own comments
  - `members_can_announce` - members can send announcements
  - `members_can_create_subgroups` - members can create subgroups
  - `admins_can_edit_user_content` - admins can edit any member's content
  - `parent_members_can_see_discussions` - parent group members can see subgroup discussions
- **FR-020**: System MUST allow administrators to configure all permission flags
- **FR-021**: System MUST enforce permission flags when members attempt restricted actions
  - **Note**: This feature implements enforcement for `members_can_add_members` and `members_can_create_subgroups`. Full enforcement of `members_can_start_discussions`, `members_can_raise_motions`, `members_can_edit_discussions`, `members_can_edit_comments`, `members_can_delete_comments`, and `members_can_announce` will be validated in Feature 005 (Discussions) and Feature 006 (Polls) when those capabilities are implemented.
- **FR-022**: System MUST grant administrators all capabilities regardless of permission flags

### Key Entities

- **Group**: An organizational container with a name, optional description, unique handle, optional parent group, and configurable permission flags. Can be archived for soft deletion.
- **Membership**: The relationship between a User and a Group, indicating their role (admin/member) and status (invited/accepted). Tracks the inviter for audit purposes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All API endpoints respond in < 500ms p95 under normal load (defined as: single-user testing, database with < 1000 groups, < 10000 memberships). Per-operation targets: creates < 200ms p95, reads < 100ms p95, updates < 150ms p95, deletes < 100ms p95.
- **SC-002**: Invitation workflow requires exactly 2 API calls: POST invite + POST accept (no polling required)
- **SC-003**: Permission changes take effect immediately - within the same session without refresh
- **SC-004**: Each of the 2 permission flags enforced in this feature (`members_can_add_members`, `members_can_create_subgroups`) has at least one dedicated test case verifying enforcement (deny when false, allow when true, admin bypass per FR-022). The remaining 9 flags are tested for storage/retrieval; enforcement tests will be added in Features 005/006 when the corresponding capabilities (discussions, polls, comments) are implemented. No permission bypass detected in test suite.
- **SC-005**: Archived groups remain accessible for historical reference while hidden from active listings
- **SC-006**: Every membership action (create/invite, accept, promote, demote, remove) is auditable with actor and timestamp information; audit tests explicitly verify each action type generates a corresponding audit.record_version entry

### Audit Trail Specification

The `audit.record_version` table captures the following fields for each mutation:

| Field | Source | Description |
|-------|--------|-------------|
| `id` | Auto-generated | UUID primary key |
| `table_oid` | Trigger | PostgreSQL OID of the audited table |
| `table_name` | Trigger | Name of the audited table (`groups` or `memberships`) |
| `record_id` | Trigger | Primary key of the affected record (as TEXT) |
| `operation` | Trigger | `INSERT`, `UPDATE`, or `DELETE` |
| `record` | Trigger | JSONB snapshot of new record state (NULL for DELETE) |
| `old_record` | Trigger | JSONB snapshot of old record state (NULL for INSERT) |
| `actor_id` | Session variable | User ID from `app.current_user_id` session variable |
| `ts` | Trigger | Timestamp of the operation |
| `xact_id` | Trigger | Transaction ID for correlating multi-table operations |

**JSONB Field Mapping:**
- `groups` table: All columns except `created_at`, `updated_at` (immutable metadata)
- `memberships` table: All columns including `user_id`, `group_id`, `role`, `accepted_at`, `inviter_id`

**System-Initiated Actions:**
- Migrations and schema changes: `actor_id` = NULL (no user context)
- Background jobs (future): Will use a system user ID or NULL with job metadata in separate field

**Retention Policy:**
- Audit records are retained indefinitely in this feature
- Future feature may add retention policy with configurable TTL

**Audit Access API:**
- Not included in Feature 004 scope
- Future feature may add GET /api/v1/audit endpoints for admin access

## Clarifications

### Session 2026-02-02

- Q: Should auditing use PostgreSQL triggers or extensions? → A: PostgreSQL triggers (automatic audit on INSERT/UPDATE/DELETE)
- Q: Which tables should have audit triggers? → A: Memberships + groups + shared audit_log table (supa_audit pattern)
- Q: How should actor context be passed to audit triggers? → A: Session variable (`SET LOCAL app.current_user_id`) per transaction
- Q: Where should last-admin protection be enforced? → A: Database trigger (BEFORE UPDATE/DELETE on memberships)
- Q: Should handle uniqueness be case-sensitive or case-insensitive? → A: Case-insensitive (Climate-Team = climate-team)

## Technical Constraints

**Audit Infrastructure:**
- **TC-001**: Audit logging MUST use PostgreSQL triggers (not application-layer or extensions) to ensure automatic, reliable capture of all membership changes at the database level
- **TC-002**: Audit records MUST be stored in a shared `audit.record_version` table using the supa_audit pattern: JSONB snapshots of old/new record state, BRIN index on timestamps, `xact_id` for transaction correlation. For tables with BIGSERIAL primary keys, store the ID directly in `record_id` as TEXT (cast from BIGINT) rather than converting to UUID.
- **TC-003**: Audit triggers MUST be applied to both `groups` and `memberships` tables
- **TC-004**: Actor context MUST be passed via PostgreSQL session variable (`SELECT set_config('app.current_user_id', '<user_id>', true)`) at the start of each transaction; triggers read this to populate the actor field in audit records. Note: Use `set_config()` instead of `SET LOCAL` for sqlc compatibility.

**Data Integrity:**
- **TC-005**: Last-admin protection MUST be enforced via a BEFORE UPDATE/DELETE trigger on `memberships` that prevents removing or demoting the last admin of any group. The trigger raises PostgreSQL error code P0001 with message "Cannot remove or demote the last administrator".
- **TC-006**: Handle uniqueness MUST be case-insensitive; enforce via unique index on `lower(handle)` or `citext` column type

**Error Detection:**
- **TC-008**: Unique constraint violations (PostgreSQL error code 23505) MUST be detected using `pgconn.PgError` type assertion, not string matching on error messages. This ensures reliable detection regardless of error message formatting.
- **TC-009**: Database trigger errors (PostgreSQL error code P0001) from last-admin protection MUST be caught and mapped to HTTP 409 Conflict with the trigger's error message preserved.
- **TC-010**: Handle format validation MUST occur at the application layer (before database insert) to return HTTP 422 with clear validation messages rather than relying on database constraint failures.

**Authorization Order:**
- **TC-011**: Authorization checks MUST follow this order: (1) Authentication verification, (2) Resource existence check, (3) Authorization check, (4) Business rule validation. This determines the HTTP status code returned (401 → 404 → 403 → 409/422).

**Rate Limiting:**
- **TC-007**: Invitation endpoints SHOULD be rate-limited to prevent spam (e.g., max 10 invites per minute per user). Implementation deferred to infrastructure layer; this feature documents the requirement for future enforcement.

## Assumptions

- User authentication and session management are already implemented (Feature 001)
- Email notifications for invitations will be handled by a future notifications feature - for now, users must check for pending invitations
- The "guests" concept (FR-019: `members_can_add_guests`) relates to discussion-level access, which will be fully implemented in Feature 005 (Discussions)
- Permission enforcement for discussions and polls will be validated when those features are implemented; this feature establishes the permission flag storage and basic enforcement structure

## Deferred Functionality

The following permission flags are stored and configurable in Feature 004 but enforcement is deferred to later features:

| Flag | Deferred To | Definition |
|------|-------------|------------|
| `members_can_add_guests` | Feature 005 | A "guest" is a user with discussion-level access (can view/participate in specific discussions) without full group membership. Enforcement location: POST /discussions/{id}/guests |
| `members_can_start_discussions` | Feature 005 | When false, only admins can POST /discussions. Enforcement location: createDiscussion handler |
| `members_can_raise_motions` | Feature 006 | "Motions" and "polls" are synonymous in Loomio terminology. When false, only admins can create polls/proposals. Enforcement location: createPoll handler |
| `members_can_edit_discussions` | Feature 005 | Scope: discussion title, description, and context fields. When false, only creator and admins can PATCH /discussions/{id}. Does not affect comments. |
| `members_can_edit_comments` | Feature 005 | Scope: own comments only (regardless of flag). When true, members can edit their own comments. Admins can edit any comment via `admins_can_edit_user_content`. |
| `members_can_delete_comments` | Feature 005 | Scope: own comments only (regardless of flag). When true, members can delete their own comments. Admins can delete any comment. |
| `members_can_announce` | Feature 005 | An "announcement" is a notification sent to all group members about a discussion or poll. Enforcement location: POST /discussions/{id}/announce |
| `admins_can_edit_user_content` | Feature 005 | Scope: discussions, comments, and poll options created by any member. When true, admins can PATCH/DELETE content they didn't create. |
| `parent_members_can_see_discussions` | Feature 005 | When true on a subgroup, members of the parent group can view discussions in the subgroup without being subgroup members. Enforcement location: getDiscussion authorization check |

## Cross-Feature Integration

**Feature 001 (User Authentication):**
- Session cookie `loomio_session` provides authenticated user context
- `user_id` from session is used for audit trail `actor_id`

**Feature 005 (Discussions) - Contract:**
- Will import Group entity for discussion.group_id relationship
- Will enforce 7 permission flags listed in Deferred Functionality table
- Will call `GetAuthorizationContext()` for permission checks

**Feature 006 (Polls) - Contract:**
- Will import Group entity for poll.group_id relationship
- Will enforce `members_can_raise_motions` flag
- Will call `GetAuthorizationContext()` for permission checks

**API Versioning:**
- New permission flags in future versions will be added with sensible defaults (typically `true` for backwards compatibility)
- Existing clients will continue to work as new flags default to permissive behavior
