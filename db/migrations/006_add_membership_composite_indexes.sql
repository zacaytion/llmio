-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- T208-T209: Index Pattern Documentation
-- ============================================================
--
-- This migration adds composite indexes optimized for the membership
-- and groups queries. The index design follows these principles:
--
-- 1. COVERING INDEXES: Include all columns needed by the query to avoid
--    table lookups. For example, memberships_user_accepted_group_idx
--    includes both user_id and group_id so the JOIN can be satisfied
--    entirely from the index.
--
-- 2. PARTIAL INDEXES: Use WHERE clauses to exclude rows that queries
--    never need. Since most queries filter on accepted_at IS NOT NULL,
--    partial indexes exclude pending memberships from the index.
--
-- 3. COLUMN ORDERING: Put equality predicates first (user_id =, group_id =),
--    then range/filter predicates (accepted_at IS NOT NULL).
--
-- Query patterns supported:
--
--   ListGroupsByUserWithCounts:
--     - Finds memberships by user_id WHERE accepted_at IS NOT NULL
--     - JOINs to groups table
--     - Uses memberships_user_accepted_group_idx
--
--   CountGroupMembershipStats:
--     - Counts by group_id WHERE accepted_at IS NOT NULL
--     - Optionally filters by role for admin counts
--     - Uses memberships_group_stats_idx
--
-- ============================================================

-- T192: Composite index for efficient listing of groups by user
-- Covers the ListGroupsByUserWithCounts query which needs:
--   1. Find memberships by user_id where accepted_at IS NOT NULL
--   2. Get the group_id for joining
-- This index also supports the current_user_role lookup in the query
CREATE INDEX memberships_user_accepted_group_idx ON memberships(user_id, group_id)
    WHERE accepted_at IS NOT NULL;

-- T192: Composite index for membership stats queries
-- Covers CountGroupMembershipStats and the subqueries in ListGroupsByUserWithCounts
-- Both need group_id + accepted_at filtering, and optionally role
CREATE INDEX memberships_group_stats_idx ON memberships(group_id, accepted_at, role)
    WHERE accepted_at IS NOT NULL;

COMMENT ON INDEX memberships_user_accepted_group_idx IS 'Supports ListGroupsByUserWithCounts - finds accepted memberships for a user';
COMMENT ON INDEX memberships_group_stats_idx IS 'Supports membership stats queries - counts by group with role filtering';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS memberships_group_stats_idx;
DROP INDEX IF EXISTS memberships_user_accepted_group_idx;

-- +goose StatementEnd
