-- sqlc queries for memberships table
-- See: data-model.md for entity definition

-- name: CreateMembership :one
-- Creates a new membership (invitation if accepted_at is NULL)
INSERT INTO memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMembershipByID :one
-- Retrieves a membership by its ID
SELECT * FROM memberships WHERE id = $1;

-- name: GetMembershipByGroupAndUser :one
-- Retrieves a specific user's membership in a group
SELECT * FROM memberships WHERE group_id = $1 AND user_id = $2;

-- name: ListMembershipsByGroup :many
-- Lists all memberships in a group with optional status filter
-- status: 'all', 'active' (accepted), or 'pending' (not accepted)
SELECT m.* FROM memberships m
WHERE m.group_id = $1
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'active' AND m.accepted_at IS NOT NULL)
    OR (sqlc.arg(status)::text = 'pending' AND m.accepted_at IS NULL)
  )
ORDER BY m.role DESC, m.created_at;

-- name: ListMembershipsByUser :many
-- Lists all memberships for a user
SELECT * FROM memberships
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListPendingInvitationsByUser :many
-- Lists all pending invitations for a user (not yet accepted)
SELECT * FROM memberships
WHERE user_id = $1 AND accepted_at IS NULL
ORDER BY created_at DESC;

-- name: AcceptMembership :one
-- Accepts a pending invitation
UPDATE memberships SET accepted_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateMembershipRole :one
-- Changes the role of a membership (promote/demote)
UPDATE memberships SET role = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteMembership :exec
-- Removes a membership
DELETE FROM memberships WHERE id = $1;

-- name: CountAdminsByGroup :one
-- Counts active admins in a group (used for last-admin protection)
SELECT COUNT(*) AS admin_count FROM memberships
WHERE group_id = $1 AND role = 'admin' AND accepted_at IS NOT NULL;

-- name: IsMember :one
-- Checks if a user is an active member of a group
SELECT EXISTS(
    SELECT 1 FROM memberships
    WHERE group_id = $1 AND user_id = $2 AND accepted_at IS NOT NULL
) AS is_member;

-- name: IsAdmin :one
-- Checks if a user is an admin of a group
SELECT EXISTS(
    SELECT 1 FROM memberships
    WHERE group_id = $1 AND user_id = $2 AND role = 'admin' AND accepted_at IS NOT NULL
) AS is_admin;

-- name: GetMembershipWithUser :one
-- Gets membership with embedded user info
SELECT
    m.*,
    u.name AS user_name,
    u.username AS user_username,
    i.name AS inviter_name,
    i.username AS inviter_username
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN users i ON i.id = m.inviter_id
WHERE m.id = $1;

-- name: ListMembershipsWithUsers :many
-- Lists memberships with embedded user info
SELECT
    m.*,
    u.name AS user_name,
    u.username AS user_username,
    i.name AS inviter_name,
    i.username AS inviter_username
FROM memberships m
JOIN users u ON u.id = m.user_id
JOIN users i ON i.id = m.inviter_id
WHERE m.group_id = $1
  AND (
    sqlc.arg(status)::text = 'all'
    OR (sqlc.arg(status)::text = 'active' AND m.accepted_at IS NOT NULL)
    OR (sqlc.arg(status)::text = 'pending' AND m.accepted_at IS NULL)
  )
ORDER BY m.role DESC, u.name;

-- name: ListInvitationsWithGroups :many
-- Lists pending invitations with group and inviter info
SELECT
    m.*,
    g.name AS group_name,
    g.handle AS group_handle,
    g.description AS group_description,
    i.name AS inviter_name,
    i.username AS inviter_username
FROM memberships m
JOIN groups g ON g.id = m.group_id
JOIN users i ON i.id = m.inviter_id
WHERE m.user_id = $1 AND m.accepted_at IS NULL
ORDER BY m.created_at DESC;
