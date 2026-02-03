-- sqlc queries for groups table
-- See: data-model.md for entity definition

-- name: CreateGroup :one
-- Creates a new group with the given parameters
INSERT INTO groups (
    name, handle, description, parent_id, created_by_id,
    members_can_add_members, members_can_add_guests, members_can_start_discussions,
    members_can_raise_motions, members_can_edit_discussions, members_can_edit_comments,
    members_can_delete_comments, members_can_announce, members_can_create_subgroups,
    admins_can_edit_user_content, parent_members_can_see_discussions
) VALUES (
    $1, $2, $3, $4, $5,
    COALESCE($6, TRUE), COALESCE($7, TRUE), COALESCE($8, TRUE),
    COALESCE($9, TRUE), COALESCE($10, FALSE), COALESCE($11, TRUE),
    COALESCE($12, TRUE), COALESCE($13, FALSE), COALESCE($14, FALSE),
    COALESCE($15, FALSE), COALESCE($16, FALSE)
)
RETURNING *;

-- name: GetGroupByID :one
-- Retrieves a group by its ID
SELECT * FROM groups WHERE id = $1;

-- name: GetGroupByHandle :one
-- Retrieves a group by its URL-safe handle (case-insensitive via CITEXT)
SELECT * FROM groups WHERE handle = $1;

-- name: ListGroupsByUser :many
-- Lists all groups a user is an active member of
-- Excludes archived groups by default unless include_archived is true
SELECT g.* FROM groups g
JOIN memberships m ON m.group_id = g.id
WHERE m.user_id = $1
  AND m.accepted_at IS NOT NULL
  AND (sqlc.arg(include_archived)::boolean = TRUE OR g.archived_at IS NULL)
ORDER BY g.name;

-- name: ListSubgroupsByParent :many
-- Lists all subgroups under a parent group
SELECT * FROM groups
WHERE parent_id = $1
  AND (sqlc.arg(include_archived)::boolean = TRUE OR archived_at IS NULL)
ORDER BY name;

-- name: UpdateGroup :one
-- Updates group fields (partial update pattern)
UPDATE groups SET
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
UPDATE groups SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UnarchiveGroup :one
-- Restores an archived group
UPDATE groups SET archived_at = NULL, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountGroupMembers :one
-- Counts active members in a group
SELECT COUNT(*) AS member_count FROM memberships
WHERE group_id = $1 AND accepted_at IS NOT NULL;

-- name: CountGroupAdmins :one
-- Counts active admins in a group
SELECT COUNT(*) AS admin_count FROM memberships
WHERE group_id = $1 AND role = 'admin' AND accepted_at IS NOT NULL;

-- name: HandleExists :one
-- Checks if a handle is already taken
SELECT EXISTS(SELECT 1 FROM groups WHERE handle = $1) AS exists;
