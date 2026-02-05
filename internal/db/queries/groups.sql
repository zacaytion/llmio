-- sqlc queries for groups table
-- See: data-model.md for entity definition

-- name: CreateGroup :one
-- Creates a new group with the given parameters
-- Uses sqlc.narg for optional boolean flags with explicit ::boolean cast for pgx compatibility
INSERT INTO data.groups (
    name, handle, description, parent_id, created_by_id,
    members_can_add_members, members_can_add_guests, members_can_start_discussions,
    members_can_raise_motions, members_can_edit_discussions, members_can_edit_comments,
    members_can_delete_comments, members_can_announce, members_can_create_subgroups,
    admins_can_edit_user_content, parent_members_can_see_discussions
) VALUES (
    @name, @handle, @description, @parent_id, @created_by_id,
    COALESCE(sqlc.narg(members_can_add_members)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_add_guests)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_start_discussions)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_raise_motions)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_edit_discussions)::boolean, FALSE),
    COALESCE(sqlc.narg(members_can_edit_comments)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_delete_comments)::boolean, TRUE),
    COALESCE(sqlc.narg(members_can_announce)::boolean, FALSE),
    COALESCE(sqlc.narg(members_can_create_subgroups)::boolean, FALSE),
    COALESCE(sqlc.narg(admins_can_edit_user_content)::boolean, FALSE),
    COALESCE(sqlc.narg(parent_members_can_see_discussions)::boolean, FALSE)
)
RETURNING *;

-- name: GetGroupByID :one
-- Retrieves a group by its ID
SELECT * FROM data.groups WHERE id = $1;

-- name: GetGroupByHandle :one
-- Retrieves a group by its URL-safe handle (case-insensitive via CITEXT)
SELECT * FROM data.groups WHERE handle = $1;

-- name: ListGroupsByUser :many
-- Lists all groups a user is an active member of
-- Excludes archived groups by default unless include_archived is true
SELECT g.* FROM data.groups g
JOIN data.memberships m ON m.group_id = g.id
WHERE m.user_id = $1
  AND m.accepted_at IS NOT NULL
  AND (sqlc.arg(include_archived)::boolean = TRUE OR g.archived_at IS NULL)
ORDER BY g.name;

-- name: ListSubgroupsByParent :many
-- Lists all subgroups under a parent group
SELECT * FROM data.groups
WHERE parent_id = $1
  AND (sqlc.arg(include_archived)::boolean = TRUE OR archived_at IS NULL)
ORDER BY name;

-- name: UpdateGroup :one
-- Updates group fields (partial update pattern)
UPDATE data.groups SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    members_can_add_members = COALESCE(sqlc.narg(members_can_add_members), members_can_add_members),
    members_can_add_guests = COALESCE(sqlc.narg(members_can_add_guests), members_can_add_guests),
    members_can_start_discussions = COALESCE(sqlc.narg(members_can_start_discussions), members_can_start_discussions),
    members_can_raise_motions = COALESCE(sqlc.narg(members_can_raise_motions), members_can_raise_motions),
    members_can_edit_discussions = COALESCE(sqlc.narg(members_can_edit_discussions), members_can_edit_discussions),
    members_can_edit_comments = COALESCE(sqlc.narg(members_can_edit_comments), members_can_edit_comments),
    members_can_delete_comments = COALESCE(sqlc.narg(members_can_delete_comments), members_can_delete_comments),
    members_can_announce = COALESCE(sqlc.narg(members_can_announce), members_can_announce),
    members_can_create_subgroups = COALESCE(sqlc.narg(members_can_create_subgroups), members_can_create_subgroups),
    admins_can_edit_user_content = COALESCE(sqlc.narg(admins_can_edit_user_content), admins_can_edit_user_content),
    parent_members_can_see_discussions = COALESCE(sqlc.narg(parent_members_can_see_discussions), parent_members_can_see_discussions),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchiveGroup :one
-- Soft-deletes a group by setting archived_at
UPDATE data.groups SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UnarchiveGroup :one
-- Restores an archived group
UPDATE data.groups SET archived_at = NULL, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountGroupMembers :one
-- Counts active members in a group
SELECT COUNT(*) AS member_count FROM data.memberships
WHERE group_id = $1 AND accepted_at IS NOT NULL;

-- name: CountGroupAdmins :one
-- Counts active admins in a group
SELECT COUNT(*) AS admin_count FROM data.memberships
WHERE group_id = $1 AND role = 'admin' AND accepted_at IS NOT NULL;

-- name: CountGroupMembershipStats :one
-- T169: Combined query to get both member and admin counts in a single query
-- More efficient than two separate queries for handleGetGroup
SELECT
    COUNT(*) AS member_count,
    COUNT(*) FILTER (WHERE role = 'admin') AS admin_count
FROM data.memberships
WHERE group_id = $1 AND accepted_at IS NOT NULL;

-- name: HandleExists :one
-- Checks if a handle is already taken
SELECT EXISTS(SELECT 1 FROM data.groups WHERE handle = $1) AS exists;

-- name: ListGroupsByUserWithCounts :many
-- T190-T191: Lists all groups a user is an active member of, with member counts
-- Uses subquery to avoid N+1 queries when fetching groups list
-- Returns group data plus member_count and admin_count for each group
SELECT
    g.*,
    m.role AS current_user_role,
    (SELECT COUNT(*) FROM data.memberships sm WHERE sm.group_id = g.id AND sm.accepted_at IS NOT NULL) AS member_count,
    (SELECT COUNT(*) FROM data.memberships sm WHERE sm.group_id = g.id AND sm.role = 'admin' AND sm.accepted_at IS NOT NULL) AS admin_count
FROM data.groups g
JOIN data.memberships m ON m.group_id = g.id
WHERE m.user_id = $1
  AND m.accepted_at IS NOT NULL
  AND (sqlc.arg(include_archived)::boolean = TRUE OR g.archived_at IS NULL)
ORDER BY g.name;
